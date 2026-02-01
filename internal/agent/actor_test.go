package agent

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/openai/openai-go"
	"github.com/waynenilsen/waynebot/internal/llm"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/tools"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// mockLLM implements LLMClient for tests.
type mockLLM struct {
	mu        sync.Mutex
	responses []llm.Response
	calls     int
}

func (m *mockLLM) ChatCompletion(_ context.Context, _ string, _ []openai.ChatCompletionMessageParamUnion, _ []openai.ChatCompletionToolParam, _ float64, _ int) (llm.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	idx := m.calls
	if idx >= len(m.responses) {
		idx = len(m.responses) - 1
	}
	m.calls++
	return m.responses[idx], nil
}

func (m *mockLLM) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

// scenario sets up the common test fixtures for actor tests.
type scenario struct {
	t       *testing.T
	actor   *Actor
	hub     *ws.Hub
	mock    *mockLLM
	persona model.Persona
	channel model.Channel
	stopHub func()
}

func newScenario(t *testing.T) *scenario {
	t.Helper()
	d := openTestDB(t)
	hub := ws.NewHub()
	done := make(chan struct{})
	go func() {
		hub.Run()
		close(done)
	}()

	persona, err := model.CreatePersona(d, "testbot", "You are a test bot.", "test-model",
		[]string{"shell_exec"}, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	ch, err := model.CreateChannel(d, "general", "", 0)
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	if err := model.SubscribeChannel(d, persona.ID, ch.ID); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	mock := &mockLLM{
		responses: []llm.Response{{Content: "Hello!", PromptTokens: 10, CompletionTokens: 5}},
	}

	registry := tools.NewRegistry()
	registry.Register("shell_exec", func(_ context.Context, _ json.RawMessage) (string, error) {
		return "ok", nil
	})

	actor := &Actor{
		Persona:  persona,
		DB:       d,
		Hub:      hub,
		LLM:      mock,
		Tools:    registry,
		Status:   NewStatusTracker(),
		Cursors:  NewCursorStore(d),
		Decision: NewDecisionMaker(),
		Budget:   NewBudgetChecker(d),
	}

	t.Cleanup(func() {
		hub.Stop()
		<-done
	})

	return &scenario{
		t:       t,
		actor:   actor,
		hub:     hub,
		mock:    mock,
		persona: persona,
		channel: ch,
	}
}

func (s *scenario) postHumanMessage(content string) model.Message {
	s.t.Helper()
	msg, err := model.CreateMessage(s.actor.DB, s.channel.ID, 999, "human", "alice", content)
	if err != nil {
		s.t.Fatalf("post human message: %v", err)
	}
	return msg
}

func (s *scenario) runOnce(ctx context.Context) {
	s.t.Helper()
	s.actor.processChannels(ctx)
}

func TestActorNormalResponse(t *testing.T) {
	s := newScenario(t)
	s.postHumanMessage("Hi bot")

	s.runOnce(context.Background())

	if s.mock.callCount() != 1 {
		t.Fatalf("expected 1 LLM call, got %d", s.mock.callCount())
	}

	// Check that a response message was created.
	msgs, err := model.GetRecentMessages(s.actor.DB, s.channel.ID, 10)
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}

	var found bool
	for _, m := range msgs {
		if m.AuthorType == "agent" && m.Content == "Hello!" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected agent response message 'Hello!' in channel")
	}

	// Cursor should be updated.
	cursor, _ := s.actor.Cursors.Get(s.persona.ID, s.channel.ID)
	if cursor == 0 {
		t.Error("expected cursor to be updated")
	}

	// Status should be idle after responding.
	if s.actor.Status.Get(s.persona.ID) != StatusIdle {
		t.Errorf("expected status idle, got %s", s.actor.Status.Get(s.persona.ID))
	}
}

func TestActorToolCallLoop(t *testing.T) {
	s := newScenario(t)
	s.mock.responses = []llm.Response{
		{
			ToolCalls: []llm.ToolCall{
				{ID: "call_1", Name: "shell_exec", Arguments: `{"command":"ls"}`},
			},
			PromptTokens: 10, CompletionTokens: 5,
		},
		{Content: "Done!", PromptTokens: 15, CompletionTokens: 8},
	}

	s.postHumanMessage("Run ls please")

	s.runOnce(context.Background())

	if s.mock.callCount() != 2 {
		t.Fatalf("expected 2 LLM calls (tool + final), got %d", s.mock.callCount())
	}

	msgs, _ := model.GetRecentMessages(s.actor.DB, s.channel.ID, 10)
	var found bool
	for _, m := range msgs {
		if m.AuthorType == "agent" && m.Content == "Done!" {
			found = true
		}
	}
	if !found {
		t.Error("expected agent response 'Done!' after tool call loop")
	}
}

func TestActorBudgetExceeded(t *testing.T) {
	s := newScenario(t)
	// Set a budget limit and exhaust it.
	s.actor.Persona.MaxTokensPerHour = 100
	_, err := s.actor.DB.WriteExec(
		`INSERT INTO llm_calls (persona_id, channel_id, model, messages_json, response_json, prompt_tokens, completion_tokens)
		 VALUES (?, ?, 'test-model', '[]', '{}', 60, 60)`,
		s.persona.ID, s.channel.ID,
	)
	if err != nil {
		t.Fatalf("insert llm_call: %v", err)
	}

	s.postHumanMessage("Hi")

	s.runOnce(context.Background())

	if s.mock.callCount() != 0 {
		t.Errorf("expected 0 LLM calls when budget exceeded, got %d", s.mock.callCount())
	}

	if s.actor.Status.Get(s.persona.ID) != StatusBudgetExceeded {
		t.Errorf("expected status budget_exceeded, got %s", s.actor.Status.Get(s.persona.ID))
	}
}

func TestActorSkipsSelfMessages(t *testing.T) {
	s := newScenario(t)

	// Post a message from the persona itself.
	_, err := model.CreateMessage(s.actor.DB, s.channel.ID, s.persona.ID, "agent", s.persona.Name, "I said something")
	if err != nil {
		t.Fatalf("create self message: %v", err)
	}

	s.runOnce(context.Background())

	if s.mock.callCount() != 0 {
		t.Errorf("expected 0 LLM calls for self-only messages, got %d", s.mock.callCount())
	}
}

func TestActorMaxToolRounds(t *testing.T) {
	s := newScenario(t)

	// Return tool calls forever — actor should stop after maxToolRounds.
	toolResp := llm.Response{
		ToolCalls:    []llm.ToolCall{{ID: "call_loop", Name: "shell_exec", Arguments: `{"command":"echo"}`}},
		PromptTokens: 5, CompletionTokens: 5,
	}
	s.mock.responses = []llm.Response{toolResp}

	s.postHumanMessage("Loop forever")

	s.runOnce(context.Background())

	if s.mock.callCount() != maxToolRounds {
		t.Errorf("expected %d LLM calls (max rounds), got %d", maxToolRounds, s.mock.callCount())
	}
}

func TestActorRunStopsOnCancel(t *testing.T) {
	s := newScenario(t)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		s.actor.Run(ctx)
		close(done)
	}()

	// Let it start.
	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("actor.Run did not return after context cancel")
	}

	if s.actor.Status.Get(s.persona.ID) != StatusStopped {
		t.Errorf("expected status stopped, got %s", s.actor.Status.Get(s.persona.ID))
	}
}

func TestActorCursorUpdatedEvenWhenSkipped(t *testing.T) {
	s := newScenario(t)

	// Post a self-message — should be skipped but cursor still updated.
	msg, err := model.CreateMessage(s.actor.DB, s.channel.ID, s.persona.ID, "agent", s.persona.Name, "self")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	s.runOnce(context.Background())

	cursor, _ := s.actor.Cursors.Get(s.persona.ID, s.channel.ID)
	if cursor != msg.ID {
		t.Errorf("expected cursor = %d, got %d", msg.ID, cursor)
	}
}
