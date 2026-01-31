package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
		log.Printf("actor %s: get subscribed channels: %v", a.Persona.Name, err)
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
		log.Printf("actor %s: get cursor for channel %d: %v", a.Persona.Name, ch.ID, err)
		return
	}

	newMessages, err := model.GetMessagesSince(a.DB, ch.ID, cursor)
	if err != nil {
		log.Printf("actor %s: get messages since %d in channel %d: %v", a.Persona.Name, cursor, ch.ID, err)
		return
	}

	if len(newMessages) == 0 {
		return
	}

	// Update cursor to latest message regardless of whether we respond.
	latestID := newMessages[len(newMessages)-1].ID
	defer func() {
		if err := a.Cursors.Set(a.Persona.ID, ch.ID, latestID); err != nil {
			log.Printf("actor %s: set cursor: %v", a.Persona.Name, err)
		}
	}()

	if !a.Decision.ShouldRespond(a.Persona, ch.ID, newMessages) {
		return
	}

	ok, err := a.Budget.WithinBudget(a.Persona.ID, a.Persona.MaxTokensPerHour)
	if err != nil {
		log.Printf("actor %s: budget check: %v", a.Persona.Name, err)
		return
	}
	if !ok {
		a.Status.Set(a.Persona.ID, StatusBudgetExceeded)
		return
	}

	a.respond(ctx, ch)
}

// respond builds history, calls the LLM (with tool call loop), and posts the final response.
func (a *Actor) respond(ctx context.Context, ch model.Channel) {
	a.Status.Set(a.Persona.ID, StatusThinking)
	defer a.Status.Set(a.Persona.ID, StatusIdle)

	history, err := model.GetRecentMessages(a.DB, ch.ID, 50)
	if err != nil {
		log.Printf("actor %s: get history: %v", a.Persona.Name, err)
		a.Status.Set(a.Persona.ID, StatusError)
		return
	}

	// GetRecentMessages returns newest-first; reverse for chronological order.
	reverseMessages(history)

	messages := llm.BuildMessages(a.Persona, history)
	toolDefs := llm.ToolsForPersona(a.Persona.ToolsEnabled)

	for round := 0; round < maxToolRounds; round++ {
		if ctx.Err() != nil {
			return
		}

		resp, err := a.LLM.ChatCompletion(ctx, a.Persona.Model, messages, toolDefs, a.Persona.Temperature, a.Persona.MaxTokens)
		if err != nil {
			log.Printf("actor %s: llm call: %v", a.Persona.Name, err)
			a.Status.Set(a.Persona.ID, StatusError)
			return
		}

		a.recordLLMCall(ch.ID, resp)

		if len(resp.ToolCalls) == 0 {
			if resp.Content != "" {
				a.postMessage(ch, resp.Content)
			}
			a.Decision.RecordResponse(a.Persona.ID, ch.ID)
			return
		}

		// Process tool calls.
		a.Status.Set(a.Persona.ID, StatusToolCall)
		messages = a.executeToolCalls(messages, resp)
	}

	log.Printf("actor %s: hit max tool rounds (%d) in channel %d", a.Persona.Name, maxToolRounds, ch.ID)
}

// executeToolCalls runs each tool call, appends assistant + tool result messages for the
// next LLM round, and returns the updated messages slice.
func (a *Actor) executeToolCalls(messages []openai.ChatCompletionMessageParamUnion, resp llm.Response) []openai.ChatCompletionMessageParamUnion {
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
		result, err := a.Tools.Call(context.Background(), tc.Name, json.RawMessage(tc.Arguments))
		if err != nil {
			result = fmt.Sprintf("error: %v", err)
		}
		messages = append(messages, openai.ToolMessage(tc.ID, result))
	}

	return messages
}

// postMessage creates a message in the DB and broadcasts it via the hub.
func (a *Actor) postMessage(ch model.Channel, content string) {
	msg, err := model.CreateMessage(a.DB, ch.ID, a.Persona.ID, "agent", a.Persona.Name, content)
	if err != nil {
		log.Printf("actor %s: post message: %v", a.Persona.Name, err)
		return
	}

	a.Hub.Broadcast(ws.Event{
		Type: "new_message",
		Data: msg,
	})
}

// recordLLMCall logs token usage to the llm_calls table.
func (a *Actor) recordLLMCall(channelID int64, resp llm.Response) {
	_, err := a.DB.WriteExec(
		`INSERT INTO llm_calls (persona_id, channel_id, model, messages_json, response_json, prompt_tokens, completion_tokens)
		 VALUES (?, ?, ?, '[]', '{}', ?, ?)`,
		a.Persona.ID, channelID, a.Persona.Model, resp.PromptTokens, resp.CompletionTokens,
	)
	if err != nil {
		log.Printf("actor %s: record llm call: %v", a.Persona.Name, err)
	}
}

// reverseMessages reverses a slice of messages in place.
func reverseMessages(msgs []model.Message) {
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
}
