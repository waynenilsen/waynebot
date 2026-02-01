package agent

import (
	"context"
	"encoding/json"
	"strings"
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
	mu           sync.Mutex
	responses    []llm.Response
	calls        int
	lastMessages []openai.ChatCompletionMessageParamUnion
}

func (m *mockLLM) ChatCompletion(_ context.Context, _ string, msgs []openai.ChatCompletionMessageParamUnion, _ []openai.ChatCompletionToolParam, _ float64, _ int) (llm.Response, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	idx := m.calls
	if idx >= len(m.responses) {
		idx = len(m.responses) - 1
	}
	m.calls++
	m.lastMessages = msgs
	return m.responses[idx], nil
}

func (m *mockLLM) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func (m *mockLLM) getLastMessages() []openai.ChatCompletionMessageParamUnion {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastMessages
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

func TestActorProjectAwarenessEnrichesSystemPrompt(t *testing.T) {
	s := newScenario(t)

	// Create a project and associate it with the channel.
	projectDir := t.TempDir()
	proj, err := model.CreateProject(s.actor.DB, "myproject", projectDir, "A test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if err := model.SetChannelProject(s.actor.DB, s.channel.ID, proj.ID); err != nil {
		t.Fatalf("set channel project: %v", err)
	}

	s.postHumanMessage("Hi bot")
	s.runOnce(context.Background())

	if s.mock.callCount() != 1 {
		t.Fatalf("expected 1 LLM call, got %d", s.mock.callCount())
	}

	// Verify the system prompt was enriched with project info.
	msgs := s.mock.getLastMessages()
	if len(msgs) == 0 {
		t.Fatal("expected messages to be sent to LLM")
	}
	sysMsg := msgs[0]
	if sysMsg.OfSystem == nil {
		t.Fatal("expected first message to be system message")
	}
	content := sysMsg.OfSystem.Content.OfString.Value
	if !strings.Contains(content, "myproject") {
		t.Errorf("expected system prompt to contain project name, got: %s", content)
	}
	if !strings.Contains(content, "A test project") {
		t.Errorf("expected system prompt to contain project description, got: %s", content)
	}
}

func TestActorProjectAwarenessPassesProjectDirToTools(t *testing.T) {
	s := newScenario(t)

	// Create a project and associate it with the channel.
	projectDir := t.TempDir()
	proj, err := model.CreateProject(s.actor.DB, "toolproject", projectDir, "")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if err := model.SetChannelProject(s.actor.DB, s.channel.ID, proj.ID); err != nil {
		t.Fatalf("set channel project: %v", err)
	}

	// Register a tool that captures the project dir from context.
	var capturedDir string
	s.actor.Tools.Register("check_context", func(ctx context.Context, _ json.RawMessage) (string, error) {
		capturedDir = tools.ProjectDirFromContext(ctx)
		return "ok", nil
	})

	s.mock.responses = []llm.Response{
		{
			ToolCalls:    []llm.ToolCall{{ID: "call_ctx", Name: "check_context", Arguments: `{}`}},
			PromptTokens: 10, CompletionTokens: 5,
		},
		{Content: "Done!", PromptTokens: 15, CompletionTokens: 8},
	}

	s.postHumanMessage("Check context")
	s.runOnce(context.Background())

	if capturedDir != projectDir {
		t.Errorf("expected project dir %q, got %q", projectDir, capturedDir)
	}
}

func TestActorNoProjectFallsBack(t *testing.T) {
	s := newScenario(t)

	// No project associated — tool context should have empty project dir.
	var capturedDir string
	s.actor.Tools.Register("check_context", func(ctx context.Context, _ json.RawMessage) (string, error) {
		capturedDir = tools.ProjectDirFromContext(ctx)
		return "ok", nil
	})

	s.mock.responses = []llm.Response{
		{
			ToolCalls:    []llm.ToolCall{{ID: "call_ctx", Name: "check_context", Arguments: `{}`}},
			PromptTokens: 10, CompletionTokens: 5,
		},
		{Content: "Done!", PromptTokens: 15, CompletionTokens: 8},
	}

	s.postHumanMessage("Check context")
	s.runOnce(context.Background())

	if capturedDir != "" {
		t.Errorf("expected empty project dir, got %q", capturedDir)
	}
}
