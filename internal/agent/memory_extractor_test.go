package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/waynenilsen/waynebot/internal/llm"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/ws"
)

func TestParseExtraction(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "valid array",
			input: `[{"kind":"fact","content":"The API uses REST"}]`,
			want:  1,
		},
		{
			name:  "empty array",
			input: `[]`,
			want:  0,
		},
		{
			name:  "wrapped in code fences",
			input: "```json\n[{\"kind\":\"decision\",\"content\":\"Using PostgreSQL\"}]\n```",
			want:  1,
		},
		{
			name:  "multiple items",
			input: `[{"kind":"fact","content":"Deploy to AWS"},{"kind":"preference","content":"Prefer functional style"}]`,
			want:  2,
		},
		{
			name:    "invalid json",
			input:   `not json at all`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExtraction(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.want {
				t.Errorf("expected %d items, got %d", tt.want, len(got))
			}
		})
	}
}

func TestFormatConversation(t *testing.T) {
	history := []model.Message{
		{AuthorType: "human", AuthorName: "alice", Content: "Hello"},
		{AuthorType: "agent", AuthorName: "bot", Content: "Hi there!"},
		{AuthorType: "human", AuthorName: "alice", Content: "What's the API?"},
	}

	result := formatConversation(history)

	if result == "" {
		t.Fatal("expected non-empty conversation text")
	}
	if got := result; got != "alice: Hello\nassistant: Hi there!\nalice: What's the API?\n" {
		t.Errorf("unexpected format: %q", got)
	}
}

func TestIsValidKind(t *testing.T) {
	for _, kind := range []string{"fact", "decision", "preference"} {
		if !isValidKind(kind) {
			t.Errorf("expected %q to be valid", kind)
		}
	}
	for _, kind := range []string{"", "opinion", "random"} {
		if isValidKind(kind) {
			t.Errorf("expected %q to be invalid", kind)
		}
	}
}

// extractorScenario sets up fixtures for memory extractor tests.
type extractorScenario struct {
	t         *testing.T
	extractor *MemoryExtractor
	mock      *mockLLM
	embedding *mockEmbedding
	persona   model.Persona
	channel   model.Channel
}

func newExtractorScenario(t *testing.T) *extractorScenario {
	t.Helper()
	d := openTestDB(t)
	hub := ws.NewHub()
	done := make(chan struct{})
	go func() {
		hub.Run()
		close(done)
	}()
	t.Cleanup(func() {
		hub.Stop()
		<-done
	})

	persona, err := model.CreatePersona(d, "membot", "Test bot.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "extraction-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	mock := &mockLLM{
		responses: []llm.Response{{Content: "[]", PromptTokens: 10, CompletionTokens: 5}},
	}

	emb := &mockEmbedding{vector: make([]float32, 1536)}

	extractor := &MemoryExtractor{
		DB:        d,
		LLM:       mock,
		Embedding: emb,
		Hub:       hub,
	}

	return &extractorScenario{
		t:         t,
		extractor: extractor,
		mock:      mock,
		embedding: emb,
		persona:   persona,
		channel:   ch,
	}
}

func (es *extractorScenario) makeHistory(humanCount int) []model.Message {
	var history []model.Message
	for i := 0; i < humanCount; i++ {
		history = append(history, model.Message{AuthorType: "human", AuthorName: "alice", Content: "message " + string(rune('A'+i))})
		history = append(history, model.Message{AuthorType: "agent", AuthorName: "bot", Content: "response " + string(rune('A'+i))})
	}
	return history
}

func TestExtractSkippedWhenTooFewHumanMessages(t *testing.T) {
	es := newExtractorScenario(t)

	// Only 2 human messages — below minHumanMessages threshold.
	history := es.makeHistory(2)

	es.extractor.Extract(context.Background(), es.persona, es.channel.ID, history, nil)

	if es.mock.callCount() != 0 {
		t.Errorf("expected 0 LLM calls for short conversation, got %d", es.mock.callCount())
	}
}

func TestExtractSkippedWhenNoEmbeddingClient(t *testing.T) {
	es := newExtractorScenario(t)
	es.extractor.Embedding = nil

	history := es.makeHistory(5)

	es.extractor.Extract(context.Background(), es.persona, es.channel.ID, history, nil)

	if es.mock.callCount() != 0 {
		t.Errorf("expected 0 LLM calls with nil embedding, got %d", es.mock.callCount())
	}
}

func TestExtractRunsAndStoresMemories(t *testing.T) {
	es := newExtractorScenario(t)

	extracted := []ExtractedMemory{
		{Kind: "fact", Content: "The API uses REST"},
		{Kind: "decision", Content: "Using PostgreSQL for storage"},
	}
	extractedJSON, _ := json.Marshal(extracted)
	es.mock.responses = []llm.Response{{Content: string(extractedJSON), PromptTokens: 20, CompletionTokens: 15}}

	// Give each memory a slightly different embedding so they don't deduplicate against each other.
	callNum := 0
	es.embedding.embedFunc = func(_ context.Context, text string) ([]float32, error) {
		v := make([]float32, 1536)
		v[0] = float32(callNum)
		callNum++
		return v, nil
	}

	history := es.makeHistory(4)

	es.extractor.Extract(context.Background(), es.persona, es.channel.ID, history, nil)

	if es.mock.callCount() != 1 {
		t.Fatalf("expected 1 LLM call, got %d", es.mock.callCount())
	}

	// Verify memories were stored.
	memories, err := model.ListMemories(es.extractor.DB, es.persona.ID, model.MemoryFilter{})
	if err != nil {
		t.Fatalf("list memories: %v", err)
	}
	if len(memories) != 2 {
		t.Errorf("expected 2 memories stored, got %d", len(memories))
	}
}

func TestExtractDeduplicatesMemories(t *testing.T) {
	es := newExtractorScenario(t)

	// Pre-store a memory.
	vec := make([]float32, 1536)
	vec[0] = 1.0
	_, err := model.CreateMemory(es.extractor.DB, es.persona.ID, &es.channel.ID, nil, "fact", "The API uses REST", vec, nil)
	if err != nil {
		t.Fatalf("create existing memory: %v", err)
	}

	// LLM returns the same fact.
	extracted := []ExtractedMemory{{Kind: "fact", Content: "The API uses REST"}}
	extractedJSON, _ := json.Marshal(extracted)
	es.mock.responses = []llm.Response{{Content: string(extractedJSON), PromptTokens: 20, CompletionTokens: 15}}

	// Embedding returns the same vector — will match the existing memory.
	es.embedding.vector = vec

	history := es.makeHistory(4)
	es.extractor.Extract(context.Background(), es.persona, es.channel.ID, history, nil)

	// Should still have only 1 memory (no duplicate stored).
	memories, err := model.ListMemories(es.extractor.DB, es.persona.ID, model.MemoryFilter{})
	if err != nil {
		t.Fatalf("list memories: %v", err)
	}
	if len(memories) != 1 {
		t.Errorf("expected 1 memory (no duplicate), got %d", len(memories))
	}
}

func TestExtractSkippedWhenBudgetExceeded(t *testing.T) {
	es := newExtractorScenario(t)

	// Set budget limit and exhaust it.
	es.persona.MaxTokensPerHour = 100
	_, err := es.extractor.DB.WriteExec(
		`INSERT INTO llm_calls (persona_id, channel_id, model, messages_json, response_json, prompt_tokens, completion_tokens)
		 VALUES (?, ?, 'test-model', '[]', '{}', 60, 60)`,
		es.persona.ID, es.channel.ID,
	)
	if err != nil {
		t.Fatalf("insert llm_call: %v", err)
	}

	history := es.makeHistory(5)
	es.extractor.Extract(context.Background(), es.persona, es.channel.ID, history, nil)

	if es.mock.callCount() != 0 {
		t.Errorf("expected 0 LLM calls when budget exceeded, got %d", es.mock.callCount())
	}
}

func TestExtractRecordsLLMCallWithExtractionTag(t *testing.T) {
	es := newExtractorScenario(t)
	es.mock.responses = []llm.Response{{Content: "[]", PromptTokens: 10, CompletionTokens: 5}}

	history := es.makeHistory(4)
	es.extractor.Extract(context.Background(), es.persona, es.channel.ID, history, nil)

	// Verify the extraction call was recorded with the model tag.
	var recordedModel string
	err := es.extractor.DB.SQL.QueryRow(
		"SELECT model FROM llm_calls WHERE persona_id = ? ORDER BY id DESC LIMIT 1",
		es.persona.ID,
	).Scan(&recordedModel)
	if err != nil {
		t.Fatalf("query llm_calls: %v", err)
	}
	if recordedModel != extractionModel+"#memory_extraction" {
		t.Errorf("expected model %q, got %q", extractionModel+"#memory_extraction", recordedModel)
	}
}
