package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

// mockEmbedding implements EmbeddingClient for tests.
type mockEmbedding struct {
	vector []float32
	err    error
}

func (m *mockEmbedding) Embed(_ context.Context, _ string) ([]float32, error) {
	return m.vector, m.err
}

func TestAssembleContextPriorityOrdering(t *testing.T) {
	d := openTestDB(t)

	// Create a persona with a system prompt.
	persona, err := model.CreatePersona(d, "membot", "You are helpful.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	// Create a channel and some history.
	ch, err := model.CreateChannel(d, "general", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	// Create a memory with a known embedding.
	vec := make([]float32, 1536)
	vec[0] = 1.0 // simple unit vector
	_, err = model.CreateMemory(d, persona.ID, nil, nil, "fact", "User prefers Go over Rust", vec, nil)
	if err != nil {
		t.Fatalf("create memory: %v", err)
	}

	// Create history messages.
	for i := 0; i < 3; i++ {
		_, err := model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello message")
		if err != nil {
			t.Fatalf("create message: %v", err)
		}
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{
		DB: d,
		Embedding: &mockEmbedding{
			vector: vec, // same vector = perfect cosine similarity
		},
	}

	msgs, budget := assembler.AssembleContext(context.Background(), AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		History:   history,
	})

	// First message should be system message containing both prompt and memories.
	if len(msgs) == 0 {
		t.Fatal("expected at least one message")
	}
	if msgs[0].OfSystem == nil {
		t.Fatal("expected first message to be a system message")
	}

	sysContent := msgs[0].OfSystem.Content.OfString.Value
	if !strings.Contains(sysContent, "You are helpful.") {
		t.Error("system message should contain persona prompt")
	}
	if !strings.Contains(sysContent, "## Relevant Memories") {
		t.Error("system message should contain retrieved memories")
	}
	if !strings.Contains(sysContent, "User prefers Go over Rust") {
		t.Error("system message should contain the actual memory content")
	}

	// Budget should reflect memory tokens.
	if budget.MemoryTokens == 0 {
		t.Error("expected non-zero memory tokens in budget")
	}

	// History messages should follow.
	if budget.HistoryMessages != 3 {
		t.Errorf("expected 3 history messages, got %d", budget.HistoryMessages)
	}
}

func TestAssembleContextBudgetExhaustion(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "budgetbot", "Short prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "budget-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	// Create many messages with substantial content.
	for i := 0; i < 20; i++ {
		_, err := model.CreateMessage(d, ch.ID, 999, "human", "alice", strings.Repeat("word ", 100))
		if err != nil {
			t.Fatalf("create message: %v", err)
		}
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{DB: d}

	// Use a very tight token budget.
	_, budget := assembler.AssembleContext(context.Background(), AssembleInput{
		Persona:    persona,
		ChannelID:  ch.ID,
		History:    history,
		TokenLimit: 300, // very small
	})

	if !budget.Exhausted {
		t.Error("expected budget to be exhausted with tight token limit")
	}
	if budget.HistoryMessages >= 20 {
		t.Errorf("expected fewer than 20 history messages, got %d", budget.HistoryMessages)
	}
	// Newest messages should be included (priority to recent).
	if budget.HistoryMessages == 0 {
		t.Error("expected at least some history messages")
	}
}

func TestAssembleContextNoEmbedding(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "noembedbot", "You are a bot.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "no-embed", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hi there")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	// No embedding client — memories should be skipped gracefully.
	assembler := &ContextAssembler{DB: d}

	msgs, budget := assembler.AssembleContext(context.Background(), AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		History:   history,
	})

	if budget.MemoryTokens != 0 {
		t.Errorf("expected 0 memory tokens without embedding client, got %d", budget.MemoryTokens)
	}
	if len(msgs) != 2 { // system + 1 history
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}

func TestAssembleContextWithProjectContext(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "projbot", "Base prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "proj-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	proj, err := model.CreateProject(d, "myproject", t.TempDir(), "A cool project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{DB: d}

	msgs, budget := assembler.AssembleContext(context.Background(), AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		Projects:  []model.Project{proj},
		History:   history,
	})

	if msgs[0].OfSystem == nil {
		t.Fatal("expected system message")
	}

	sysContent := msgs[0].OfSystem.Content.OfString.Value
	if !strings.Contains(sysContent, "myproject") {
		t.Error("system message should contain project name")
	}
	if !strings.Contains(sysContent, "A cool project") {
		t.Error("system message should contain project description")
	}
	if budget.ProjectTokens == 0 {
		t.Error("expected non-zero project tokens")
	}
}

func TestAssembleContextWithAgentsMd(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "agentsbot", "Base prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "agents-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	projDir := t.TempDir()
	err = os.WriteFile(filepath.Join(projDir, "AGENTS.md"), []byte("Always use gofmt.\nPrefer table-driven tests."), 0644)
	if err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	proj, err := model.CreateProject(d, "agentsproject", projDir, "A project with AGENTS.md")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{DB: d}

	msgs, budget := assembler.AssembleContext(context.Background(), AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		Projects:  []model.Project{proj},
		History:   history,
	})

	sysContent := msgs[0].OfSystem.Content.OfString.Value
	if !strings.Contains(sysContent, "## Project Instructions (AGENTS.md)") {
		t.Error("system message should contain AGENTS.md header")
	}
	if !strings.Contains(sysContent, "Always use gofmt.") {
		t.Error("system message should contain AGENTS.md content")
	}
	if budget.AgentsmdTokens == 0 {
		t.Error("expected non-zero AgentsmdTokens in budget")
	}
}

func TestAssembleContextWithoutAgentsMd(t *testing.T) {
	d := openTestDB(t)

	persona, err := model.CreatePersona(d, "noagentsbot", "Base prompt.", "test-model",
		nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "no-agents-test", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	// Project dir with no AGENTS.md
	proj, err := model.CreateProject(d, "plainproject", t.TempDir(), "No agents file")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	_, err = model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	history, _ := model.GetRecentMessages(d, ch.ID, 50)
	reverseMessages(history)

	assembler := &ContextAssembler{DB: d}

	_, budget := assembler.AssembleContext(context.Background(), AssembleInput{
		Persona:   persona,
		ChannelID: ch.ID,
		Projects:  []model.Project{proj},
		History:   history,
	})

	if budget.AgentsmdTokens != 0 {
		t.Errorf("expected 0 AgentsmdTokens without AGENTS.md, got %d", budget.AgentsmdTokens)
	}
}

func TestEstimateTokens(t *testing.T) {
	got := EstimateTokens("Hello world!") // 12 chars = 3 tokens
	if got != 3 {
		t.Errorf("expected 3 tokens, got %d", got)
	}
}
