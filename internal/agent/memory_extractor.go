package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/openai/openai-go"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/llm"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// extractionModel is the cheap/fast model used for memory extraction.
const extractionModel = "openai/gpt-4.1-mini"

// extractionMaxTokens caps the extraction response size.
const extractionMaxTokens = 512

// minHumanMessages is the minimum number of human messages required before
// extraction runs. Conversations below this threshold are considered too
// short to contain extractable knowledge.
const minHumanMessages = 3

// duplicateThreshold is the cosine similarity above which a candidate memory
// is considered a duplicate of an existing one and will not be stored.
const duplicateThreshold = 0.9

const extractionPrompt = `You are a memory extraction system. Given the conversation below, extract key facts, decisions, and preferences that would be useful to remember for future conversations.

Output a JSON array of objects with "kind" (fact/decision/preference) and "content" (concise statement).
Only extract genuinely important information. Skip small talk and transient details.
If nothing important was discussed, return an empty array [].`

// ExtractedMemory is a single fact/decision/preference parsed from the extraction LLM response.
type ExtractedMemory struct {
	Kind    string `json:"kind"`
	Content string `json:"content"`
}

// MemoryExtractor runs a lightweight LLM call after a response to extract
// memories from the conversation and store them for future RAG retrieval.
type MemoryExtractor struct {
	DB        *db.DB
	LLM       LLMClient
	Embedding EmbeddingClient
	Hub       *ws.Hub
}

// Extract analyses the conversation history and stores new memories.
// It is safe to call with nil Embedding (extraction will be skipped).
func (me *MemoryExtractor) Extract(ctx context.Context, persona model.Persona, channelID int64, history []model.Message, projects []model.Project) {
	if me.Embedding == nil {
		return
	}

	// Count human messages — skip extraction for trivial conversations.
	humanCount := 0
	for _, m := range history {
		if m.AuthorType == "human" {
			humanCount++
		}
	}
	if humanCount < minHumanMessages {
		return
	}

	// Check budget before spending tokens on extraction.
	ok, err := budgetCheck(me.DB, persona)
	if err != nil {
		slog.Error("memory_extractor: budget check", "persona", persona.Name, "error", err)
		return
	}
	if !ok {
		return
	}

	// Build a condensed conversation for the extraction prompt.
	conversationText := formatConversation(history)

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(extractionPrompt),
		openai.UserMessage(conversationText),
	}

	resp, err := me.LLM.ChatCompletion(ctx, extractionModel, messages, nil, 0.0, extractionMaxTokens)
	if err != nil {
		slog.Error("memory_extractor: llm call", "persona", persona.Name, "error", err)
		return
	}

	me.recordExtractionCall(persona.ID, channelID, messages, resp)

	extracted, err := parseExtraction(resp.Content)
	if err != nil {
		slog.Error("memory_extractor: parse extraction", "persona", persona.Name, "error", err)
		return
	}

	if len(extracted) == 0 {
		return
	}

	// Determine optional project ID from the channel's projects.
	var projectID *int64
	if len(projects) > 0 {
		projectID = &projects[0].ID
	}

	me.storeMemories(ctx, persona.ID, channelID, projectID, extracted)
}

// storeMemories embeds each extracted memory, checks for duplicates, and persists new ones.
func (me *MemoryExtractor) storeMemories(ctx context.Context, personaID, channelID int64, projectID *int64, extracted []ExtractedMemory) {
	for _, em := range extracted {
		if !isValidKind(em.Kind) || em.Content == "" {
			continue
		}

		embedding, err := me.Embedding.Embed(ctx, em.Content)
		if err != nil {
			slog.Error("memory_extractor: embed memory", "error", err)
			continue
		}

		// Duplicate check — search existing memories for high similarity.
		existing, err := model.SearchMemories(me.DB, personaID, embedding, 1, model.MemoryFilter{})
		if err != nil {
			slog.Error("memory_extractor: search duplicates", "error", err)
			continue
		}
		if len(existing) > 0 && existing[0].Score > duplicateThreshold {
			slog.Debug("memory_extractor: skipping duplicate memory",
				"content", em.Content,
				"existing", existing[0].Content,
				"score", existing[0].Score,
			)
			continue
		}

		chID := channelID
		_, err = model.CreateMemory(me.DB, personaID, &chID, projectID, em.Kind, em.Content, embedding, nil)
		if err != nil {
			slog.Error("memory_extractor: create memory", "error", err)
			continue
		}
	}
}

// recordExtractionCall logs the extraction LLM call with a metadata flag.
func (me *MemoryExtractor) recordExtractionCall(personaID, channelID int64, messages []openai.ChatCompletionMessageParamUnion, resp llm.Response) {
	messagesJSON, err := json.Marshal(messages)
	if err != nil {
		messagesJSON = []byte("[]")
	}
	responseJSON, err := json.Marshal(resp)
	if err != nil {
		responseJSON = []byte("{}")
	}

	res, err := me.DB.WriteExec(
		`INSERT INTO llm_calls (persona_id, channel_id, model, messages_json, response_json, prompt_tokens, completion_tokens)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		personaID, channelID, extractionModel+"#memory_extraction", string(messagesJSON), string(responseJSON), resp.PromptTokens, resp.CompletionTokens,
	)
	if err != nil {
		slog.Error("memory_extractor: record llm call", "error", err)
		return
	}

	id, _ := res.LastInsertId()
	me.Hub.Broadcast(ws.Event{
		Type: "agent_llm_call",
		Data: map[string]any{
			"id":                id,
			"persona_id":        personaID,
			"channel_id":        channelID,
			"model":             extractionModel + "#memory_extraction",
			"messages_json":     string(messagesJSON),
			"response_json":     string(responseJSON),
			"prompt_tokens":     resp.PromptTokens,
			"completion_tokens": resp.CompletionTokens,
			"created_at":        time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// parseExtraction parses the LLM's JSON response into a slice of ExtractedMemory.
func parseExtraction(content string) ([]ExtractedMemory, error) {
	content = trimJSONFences(content)
	var result []ExtractedMemory
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("unmarshal extraction: %w", err)
	}
	return result, nil
}

// trimJSONFences strips optional ```json ... ``` fences from LLM output.
func trimJSONFences(s string) string {
	// Find first '[' and last ']'.
	start := -1
	end := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '[' {
			start = i
			break
		}
	}
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == ']' {
			end = i
			break
		}
	}
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}

// formatConversation produces a simple text summary of the conversation for extraction.
func formatConversation(history []model.Message) string {
	var b []byte
	for _, m := range history {
		switch m.AuthorType {
		case "human":
			b = append(b, m.AuthorName...)
			b = append(b, ": "...)
			b = append(b, m.Content...)
			b = append(b, '\n')
		case "agent":
			b = append(b, "assistant: "...)
			b = append(b, m.Content...)
			b = append(b, '\n')
		}
	}
	return string(b)
}

func isValidKind(kind string) bool {
	return kind == "fact" || kind == "decision" || kind == "preference"
}

// budgetCheck verifies the persona hasn't exceeded its hourly token limit.
func budgetCheck(d *db.DB, persona model.Persona) (bool, error) {
	if persona.MaxTokensPerHour <= 0 {
		return true, nil
	}
	bc := NewBudgetChecker(d)
	return bc.WithinBudget(persona.ID, persona.MaxTokensPerHour)
}
