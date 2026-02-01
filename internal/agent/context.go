package agent

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/openai/openai-go"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// DefaultContextWindow is the default token budget when not specified.
const DefaultContextWindow = 200_000

// EmbeddingClient is the interface for embedding text, enabling test mocks.
type EmbeddingClient interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// ContextBudget tracks how the token budget was allocated across sections.
type ContextBudget struct {
	TotalTokens     int
	SystemTokens    int
	ProjectTokens   int
	MemoryTokens    int
	HistoryTokens   int
	HistoryMessages int
	Exhausted       bool
}

// ContextAssembler builds the LLM message array with priority-ordered sections.
type ContextAssembler struct {
	DB        *db.DB
	Embedding EmbeddingClient
}

// AssembleInput holds everything needed to assemble context.
type AssembleInput struct {
	Persona    model.Persona
	ChannelID  int64
	Projects   []model.Project
	History    []model.Message // chronological order
	TokenLimit int             // 0 = DefaultContextWindow
}

// EstimateTokens gives a rough token count for a string (1 token ≈ 4 chars).
func EstimateTokens(text string) int {
	return len(text) / 4
}

// AssembleContext builds the message array with priority ordering:
// 1. System prompt (always)
// 2. Project context (if any)
// 3. Retrieved memories (semantic search)
// 4. Channel message history (fills remaining budget)
func (ca *ContextAssembler) AssembleContext(ctx context.Context, input AssembleInput) ([]openai.ChatCompletionMessageParamUnion, ContextBudget) {
	budget := ContextBudget{}
	tokenLimit := input.TokenLimit
	if tokenLimit <= 0 {
		tokenLimit = DefaultContextWindow
	}
	budget.TotalTokens = tokenLimit

	remaining := tokenLimit

	// 1. System prompt — always included.
	systemPrompt := input.Persona.SystemPrompt
	if len(input.Projects) > 0 {
		systemPrompt += formatProjectContext(input.Projects)
	}
	budget.SystemTokens = EstimateTokens(systemPrompt)
	budget.ProjectTokens = 0
	if len(input.Projects) > 0 {
		budget.ProjectTokens = EstimateTokens(formatProjectContext(input.Projects))
	}
	remaining -= budget.SystemTokens

	// 2. Retrieve memories via semantic search.
	memoriesBlock := ca.retrieveMemories(ctx, input)
	memTokens := EstimateTokens(memoriesBlock)
	if memTokens > remaining {
		memoriesBlock = ""
		memTokens = 0
	}
	budget.MemoryTokens = memTokens
	remaining -= memTokens

	// Combine system prompt + memories into a single system message.
	fullSystem := systemPrompt
	if memoriesBlock != "" {
		fullSystem += "\n\n" + memoriesBlock
	}

	msgs := make([]openai.ChatCompletionMessageParamUnion, 0, len(input.History)+1)
	msgs = append(msgs, openai.SystemMessage(fullSystem))

	// 3. Fill remaining budget with history messages (newest have priority).
	// Walk from newest to oldest, accumulating tokens, then reverse.
	type histEntry struct {
		msg    openai.ChatCompletionMessageParamUnion
		tokens int
	}

	var selected []histEntry
	historyUsed := 0

	for i := len(input.History) - 1; i >= 0; i-- {
		m := input.History[i]
		oaiMsg := buildSingleMessage(m)
		msgText := messageText(m)
		t := EstimateTokens(msgText)
		if historyUsed+t > remaining {
			budget.Exhausted = true
			break
		}
		selected = append(selected, histEntry{msg: oaiMsg, tokens: t})
		historyUsed += t
	}

	budget.HistoryTokens = historyUsed
	budget.HistoryMessages = len(selected)

	// Reverse selected to restore chronological order.
	for i, j := 0, len(selected)-1; i < j; i, j = i+1, j-1 {
		selected[i], selected[j] = selected[j], selected[i]
	}

	for _, e := range selected {
		msgs = append(msgs, e.msg)
	}

	return msgs, budget
}

// retrieveMemories embeds recent messages as a query and searches for relevant memories.
func (ca *ContextAssembler) retrieveMemories(ctx context.Context, input AssembleInput) string {
	if ca.Embedding == nil {
		return ""
	}

	// Build query from last few messages (up to 5).
	queryMessages := input.History
	if len(queryMessages) > 5 {
		queryMessages = queryMessages[len(queryMessages)-5:]
	}
	if len(queryMessages) == 0 {
		return ""
	}

	var queryParts []string
	for _, m := range queryMessages {
		queryParts = append(queryParts, m.Content)
	}
	query := strings.Join(queryParts, "\n")

	queryEmbedding, err := ca.Embedding.Embed(ctx, query)
	if err != nil {
		slog.Error("context: embed query for memory search", "error", err)
		return ""
	}

	// Search memories scoped to this persona. Include both channel-specific and global (nil channel).
	filter := model.MemoryFilter{}
	memories, err := model.SearchMemories(ca.DB, input.Persona.ID, queryEmbedding, 10, filter)
	if err != nil {
		slog.Error("context: search memories", "error", err)
		return ""
	}

	if len(memories) == 0 {
		return ""
	}

	// Filter to a reasonable similarity threshold.
	var relevant []model.ScoredMemory
	for _, m := range memories {
		if m.Score >= 0.3 {
			relevant = append(relevant, m)
		}
	}

	if len(relevant) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Relevant Memories\n")
	for _, m := range relevant {
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", m.Kind, m.Content))
	}
	return sb.String()
}

// formatProjectContext builds the project context string for the system prompt.
func formatProjectContext(projects []model.Project) string {
	var sb strings.Builder
	sb.WriteString("\n\n## Project Context\n")
	sb.WriteString(fmt.Sprintf("This channel is associated with the project **%s**.", projects[0].Name))
	if projects[0].Description != "" {
		sb.WriteString("\nDescription: " + projects[0].Description)
	}
	sb.WriteString("\nFile tools (file_read, file_write, shell_exec) are scoped to the project directory.")
	return sb.String()
}

// buildSingleMessage converts a single domain message to an OpenAI message param.
func buildSingleMessage(m model.Message) openai.ChatCompletionMessageParamUnion {
	switch m.AuthorType {
	case "human":
		return openai.UserMessage(m.AuthorName + ": " + m.Content)
	case "agent":
		return openai.AssistantMessage(m.Content)
	case "tool_call":
		name, args := splitToolCallContent(m.Content)
		return openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				ToolCalls: []openai.ChatCompletionMessageToolCallParam{
					{
						ID: m.AuthorName,
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      name,
							Arguments: args,
						},
					},
				},
			},
		}
	case "tool_result":
		return openai.ToolMessage(m.Content, m.AuthorName)
	default:
		return openai.UserMessage(m.Content)
	}
}

// splitToolCallContent splits "tool_name\narguments_json" into name and args.
func splitToolCallContent(content string) (name, args string) {
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			return content[:i], content[i+1:]
		}
	}
	return content, "{}"
}

// messageText returns the text content of a message for token estimation.
func messageText(m model.Message) string {
	switch m.AuthorType {
	case "human":
		return m.AuthorName + ": " + m.Content
	default:
		return m.Content
	}
}
