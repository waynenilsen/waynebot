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
	"github.com/waynenilsen/waynebot/internal/tools"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// LLMClient is the interface used by Actor for LLM calls, enabling test mocks.
type LLMClient interface {
	ChatCompletion(ctx context.Context, model string, messages []openai.ChatCompletionMessageParamUnion, tools []openai.ChatCompletionToolParam, temperature float64, maxTokens int) (llm.Response, error)
}

const (
	maxToolRounds  = 10
	fallbackTicker = 60 * time.Second
)

// Actor is the per-persona processing loop. It listens for notifications from the
// WebSocket hub and processes new messages in subscribed channels.
type Actor struct {
	Persona  model.Persona
	DB       *db.DB
	Hub      *ws.Hub
	LLM      LLMClient
	Tools    *tools.Registry
	Status   *StatusTracker
	Cursors  *CursorStore
	Decision *DecisionMaker
	Budget   *BudgetChecker
}

// Run starts the actor's processing loop. It blocks until ctx is cancelled.
func (a *Actor) Run(ctx context.Context) {
	ticker := time.NewTicker(fallbackTicker)
	defer ticker.Stop()

	a.Status.Set(a.Persona.ID, StatusIdle)

	for {
		select {
		case <-ctx.Done():
			a.Status.Set(a.Persona.ID, StatusStopped)
			return
		case <-a.Hub.NotifyChan:
			a.processChannels(ctx)
		case <-ticker.C:
			a.processChannels(ctx)
		}
	}
}

// processChannels iterates all subscribed channels and processes new messages.
func (a *Actor) processChannels(ctx context.Context) {
	channels, err := model.GetSubscribedChannels(a.DB, a.Persona.ID)
	if err != nil {
		slog.Error("actor: get subscribed channels", "persona", a.Persona.Name, "error", err)
		return
	}

	for _, ch := range channels {
		if ctx.Err() != nil {
			return
		}
		a.processChannel(ctx, ch)
	}
}

// processChannel checks for new messages in a channel and responds if appropriate.
func (a *Actor) processChannel(ctx context.Context, ch model.Channel) {
	cursor, err := a.Cursors.Get(a.Persona.ID, ch.ID)
	if err != nil {
		slog.Error("actor: get cursor", "persona", a.Persona.Name, "channel_id", ch.ID, "error", err)
		return
	}

	newMessages, err := model.GetMessagesSince(a.DB, ch.ID, cursor)
	if err != nil {
		slog.Error("actor: get messages since cursor", "persona", a.Persona.Name, "cursor", cursor, "channel_id", ch.ID, "error", err)
		return
	}

	if len(newMessages) == 0 {
		return
	}

	// Update cursor to latest message regardless of whether we respond.
	latestID := newMessages[len(newMessages)-1].ID
	defer func() {
		if err := a.Cursors.Set(a.Persona.ID, ch.ID, latestID); err != nil {
			slog.Error("actor: set cursor", "persona", a.Persona.Name, "error", err)
		}
	}()

	if !a.Decision.ShouldRespond(a.Persona, ch.ID, newMessages) {
		return
	}

	ok, err := a.Budget.WithinBudget(a.Persona.ID, a.Persona.MaxTokensPerHour)
	if err != nil {
		slog.Error("actor: budget check", "persona", a.Persona.Name, "error", err)
		return
	}
	if !ok {
		a.Status.Set(a.Persona.ID, StatusBudgetExceeded)
		a.broadcastStatus(ch.ID, StatusBudgetExceeded)
		return
	}

	a.respond(ctx, ch)
}

// respond builds history, calls the LLM (with tool call loop), and posts the final response.
func (a *Actor) respond(ctx context.Context, ch model.Channel) {
	a.Status.Set(a.Persona.ID, StatusThinking)
	a.broadcastStatus(ch.ID, StatusThinking)
	defer func() {
		// Don't override terminal states like context_full or error.
		if s := a.Status.Get(a.Persona.ID); s == StatusThinking || s == StatusToolCall {
			a.Status.Set(a.Persona.ID, StatusIdle)
			a.broadcastStatus(ch.ID, StatusIdle)
		}
	}()

	history, err := model.GetRecentMessages(a.DB, ch.ID, 50)
	if err != nil {
		slog.Error("actor: get history", "persona", a.Persona.Name, "error", err)
		a.Status.Set(a.Persona.ID, StatusError)
		return
	}

	// GetRecentMessages returns newest-first; reverse for chronological order.
	reverseMessages(history)

	// Look up associated projects for system prompt enrichment and tool scoping.
	projects, err := model.ListChannelProjects(a.DB, ch.ID)
	if err != nil {
		slog.Error("actor: list channel projects", "persona", a.Persona.Name, "channel_id", ch.ID, "error", err)
	}

	assembler := &ContextAssembler{}
	messages, budget := assembler.AssembleContext(AssembleInput{
		Persona:   a.Persona,
		ChannelID: ch.ID,
		Projects:  projects,
		History:   history,
	})

	if budget.Exhausted && budget.HistoryMessages == 0 {
		slog.Warn("actor: context window full, cannot process messages",
			"persona", a.Persona.Name,
			"channel_id", ch.ID,
		)
		a.Status.Set(a.Persona.ID, StatusContextFull)
		a.broadcastStatus(ch.ID, StatusContextFull)
		a.postMessage(ch, "My context window is full. I cannot process new messages until context is reset. Please use `/reset-context` or start a new conversation thread.")
		a.broadcastContextBudget(ch.ID, budget)
		return
	}

	if budget.Exhausted {
		slog.Info("actor: context budget partially exhausted",
			"persona", a.Persona.Name,
			"channel_id", ch.ID,
			"history_messages", budget.HistoryMessages,
		)
	}

	toolDefs := llm.ToolsForPersona(a.Persona.ToolsEnabled)

	for round := 0; round < maxToolRounds; round++ {
		if ctx.Err() != nil {
			return
		}

		resp, err := a.LLM.ChatCompletion(ctx, a.Persona.Model, messages, toolDefs, a.Persona.Temperature, a.Persona.MaxTokens)
		if err != nil {
			slog.Error("actor: llm call", "persona", a.Persona.Name, "error", err)
			a.Status.Set(a.Persona.ID, StatusError)
			a.broadcastStatus(ch.ID, StatusError)
			return
		}

		a.recordLLMCall(ch.ID, messages, resp)

		if len(resp.ToolCalls) == 0 {
			if resp.Content != "" {
				a.postMessage(ch, resp.Content)
			}
			a.Decision.RecordResponse(a.Persona.ID, ch.ID)
			a.broadcastContextBudget(ch.ID, budget)
			return
		}

		// Process tool calls.
		a.Status.Set(a.Persona.ID, StatusToolCall)
		a.broadcastStatus(ch.ID, StatusToolCall)
		messages = a.executeToolCalls(messages, resp, projects)
	}

	slog.Warn("actor: hit max tool rounds", "persona", a.Persona.Name, "max_rounds", maxToolRounds, "channel_id", ch.ID)
}

// executeToolCalls runs each tool call, appends assistant + tool result messages for the
// next LLM round, and returns the updated messages slice.
func (a *Actor) executeToolCalls(messages []openai.ChatCompletionMessageParamUnion, resp llm.Response, projects []model.Project) []openai.ChatCompletionMessageParamUnion {
	// Build assistant message containing the tool calls.
	toolCalls := make([]openai.ChatCompletionMessageToolCallParam, len(resp.ToolCalls))
	for i, tc := range resp.ToolCalls {
		toolCalls[i] = openai.ChatCompletionMessageToolCallParam{
			ID: tc.ID,
			Function: openai.ChatCompletionMessageToolCallFunctionParam{
				Name:      tc.Name,
				Arguments: tc.Arguments,
			},
		}
	}
	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfAssistant: &openai.ChatCompletionAssistantMessageParam{
			ToolCalls: toolCalls,
		},
	})

	// Execute each tool and append the result.
	for _, tc := range resp.ToolCalls {
		start := time.Now()
		toolCtx := tools.WithPersonaID(context.Background(), a.Persona.ID)
		if len(projects) > 0 {
			toolCtx = tools.WithProjectDir(toolCtx, projects[0].Path)
		}
		result, err := a.Tools.Call(toolCtx, tc.Name, json.RawMessage(tc.Arguments))
		duration := time.Since(start)

		errText := ""
		if err != nil {
			errText = err.Error()
			result = fmt.Sprintf("error: %v", err)
		}

		a.recordToolExecution(tc.Name, tc.Arguments, result, errText, duration)
		messages = append(messages, openai.ToolMessage(tc.ID, result))
	}

	return messages
}

// postMessage creates a message in the DB and broadcasts it via the hub.
func (a *Actor) postMessage(ch model.Channel, content string) {
	msg, err := model.CreateMessage(a.DB, ch.ID, a.Persona.ID, "agent", a.Persona.Name, content)
	if err != nil {
		slog.Error("actor: post message", "persona", a.Persona.Name, "error", err)
		return
	}

	a.Hub.Broadcast(ws.Event{
		Type: "new_message",
		Data: map[string]any{
			"id":          msg.ID,
			"channel_id":  msg.ChannelID,
			"author_id":   msg.AuthorID,
			"author_type": msg.AuthorType,
			"author_name": msg.AuthorName,
			"content":     msg.Content,
			"created_at":  msg.CreatedAt.Format(time.RFC3339),
			"reactions":   []any{},
		},
	})
}

// recordLLMCall logs the full request messages and response to the llm_calls table.
func (a *Actor) recordLLMCall(channelID int64, messages []openai.ChatCompletionMessageParamUnion, resp llm.Response) {
	messagesJSON, err := json.Marshal(messages)
	if err != nil {
		slog.Error("actor: marshal messages", "persona", a.Persona.Name, "error", err)
		messagesJSON = []byte("[]")
	}

	responseJSON, err := json.Marshal(resp)
	if err != nil {
		slog.Error("actor: marshal response", "persona", a.Persona.Name, "error", err)
		responseJSON = []byte("{}")
	}

	res, err := a.DB.WriteExec(
		`INSERT INTO llm_calls (persona_id, channel_id, model, messages_json, response_json, prompt_tokens, completion_tokens)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		a.Persona.ID, channelID, a.Persona.Model, string(messagesJSON), string(responseJSON), resp.PromptTokens, resp.CompletionTokens,
	)
	if err != nil {
		slog.Error("actor: record llm call", "persona", a.Persona.Name, "error", err)
		return
	}

	id, _ := res.LastInsertId()
	a.Hub.Broadcast(ws.Event{
		Type: "agent_llm_call",
		Data: map[string]any{
			"id":                id,
			"persona_id":        a.Persona.ID,
			"channel_id":        channelID,
			"model":             a.Persona.Model,
			"messages_json":     string(messagesJSON),
			"response_json":     string(responseJSON),
			"prompt_tokens":     resp.PromptTokens,
			"completion_tokens": resp.CompletionTokens,
			"created_at":        time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// recordToolExecution logs a tool invocation to the tool_executions table.
func (a *Actor) recordToolExecution(toolName, argsJSON, output, errText string, duration time.Duration) {
	res, err := a.DB.WriteExec(
		`INSERT INTO tool_executions (persona_id, tool_name, args_json, output_text, error_text, duration_ms)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		a.Persona.ID, toolName, argsJSON, output, errText, duration.Milliseconds(),
	)
	if err != nil {
		slog.Error("actor: record tool execution", "persona", a.Persona.Name, "tool", toolName, "error", err)
		return
	}

	id, _ := res.LastInsertId()
	a.Hub.Broadcast(ws.Event{
		Type: "agent_tool_execution",
		Data: map[string]any{
			"id":          id,
			"persona_id":  a.Persona.ID,
			"tool_name":   toolName,
			"args_json":   argsJSON,
			"output_text": output,
			"error_text":  errText,
			"duration_ms": duration.Milliseconds(),
			"created_at":  time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// broadcastStatus sends an agent_status event via the WebSocket hub.
func (a *Actor) broadcastStatus(channelID int64, status Status) {
	a.Hub.Broadcast(ws.Event{
		Type: "agent_status",
		Data: map[string]any{
			"persona_id":   a.Persona.ID,
			"persona_name": a.Persona.Name,
			"channel_id":   channelID,
			"status":       status.String(),
		},
	})
}

// broadcastContextBudget sends context budget info via the WebSocket hub.
func (a *Actor) broadcastContextBudget(channelID int64, budget ContextBudget) {
	a.Hub.Broadcast(ws.Event{
		Type: "agent_context_budget",
		Data: map[string]any{
			"persona_id":       a.Persona.ID,
			"channel_id":       channelID,
			"total_tokens":     budget.TotalTokens,
			"system_tokens":    budget.SystemTokens,
			"project_tokens":   budget.ProjectTokens,
			"history_tokens":   budget.HistoryTokens,
			"history_messages": budget.HistoryMessages,
			"exhausted":        budget.Exhausted,
		},
	})
}

// reverseMessages reverses a slice of messages in place.
func reverseMessages(msgs []model.Message) {
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
}
