package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/openai/openai-go"
	"github.com/waynenilsen/waynebot/internal/agent"
	"github.com/waynenilsen/waynebot/internal/api"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/llm"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/tools"
	"github.com/waynenilsen/waynebot/internal/ws"
)

type idleLLM struct{}

func (idleLLM) ChatCompletion(_ context.Context, _ string, _ []openai.ChatCompletionMessageParamUnion, _ []openai.ChatCompletionToolParam, _ float64, _ int) (llm.Response, error) {
	return llm.Response{}, nil
}

func newTestRouterWithSupervisor(t *testing.T, d *db.DB) (http.Handler, *agent.Supervisor) {
	t.Helper()
	hub := ws.NewHub()
	go hub.Run()
	t.Cleanup(func() { hub.Stop() })

	sup := agent.NewSupervisor(d, hub, idleLLM{}, nil, tools.NewRegistry())
	router := api.NewRouter(d, []string{"*"}, hub, sup)
	return router, sup
}

func TestAgentStatusUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)

	rec := doJSON(t, router, "GET", "/api/agents/status", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAgentStatusEmpty(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)

	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "GET", "/api/agents/status", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		SupervisorRunning bool              `json:"supervisor_running"`
		Agents            []json.RawMessage `json:"agents"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Agents) != 0 {
		t.Errorf("expected 0 entries, got %d", len(resp.Agents))
	}
}

func TestAgentStatusWithPersonas(t *testing.T) {
	d := openTestDB(t)
	router, sup := newTestRouterWithSupervisor(t, d)

	token := registerUser(t, router, "alice", "password123", "")

	p, err := model.CreatePersona(d, "testbot", "prompt", "model", nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	sup.Status.Set(p.ID, agent.StatusIdle)

	rec := doJSON(t, router, "GET", "/api/agents/status", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		SupervisorRunning bool `json:"supervisor_running"`
		Agents            []struct {
			PersonaID   int64    `json:"persona_id"`
			PersonaName string   `json:"persona_name"`
			Status      string   `json:"status"`
			Channels    []string `json:"channels"`
		} `json:"agents"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp.Agents) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(resp.Agents))
	}
	if resp.Agents[0].PersonaName != "testbot" {
		t.Errorf("persona_name = %q, want testbot", resp.Agents[0].PersonaName)
	}
	if resp.Agents[0].Status != "idle" {
		t.Errorf("status = %q, want idle", resp.Agents[0].Status)
	}
}

func TestAgentStartAndStop(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)

	token := registerUser(t, router, "alice", "password123", "")

	// Start with no personas should succeed.
	rec := doJSON(t, router, "POST", "/api/agents/start", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("start: status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	// Starting again should conflict.
	rec = doJSON(t, router, "POST", "/api/agents/start", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusConflict {
		t.Errorf("start again: status = %d, want 409", rec.Code)
	}

	// Stop should succeed.
	rec = doJSON(t, router, "POST", "/api/agents/stop", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("stop: status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	// Stopping again should conflict.
	rec = doJSON(t, router, "POST", "/api/agents/stop", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusConflict {
		t.Errorf("stop again: status = %d, want 409", rec.Code)
	}
}

func TestAgentStartUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)

	rec := doJSON(t, router, "POST", "/api/agents/start", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAgentStopUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)

	rec := doJSON(t, router, "POST", "/api/agents/stop", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func seedLLMCall(t *testing.T, d *db.DB, personaID, channelID int64) {
	t.Helper()
	_, err := d.WriteExec(
		`INSERT INTO llm_calls (persona_id, channel_id, model, messages_json, response_json, prompt_tokens, completion_tokens)
		 VALUES (?, ?, 'test-model', '[]', '{}', 100, 50)`,
		personaID, channelID,
	)
	if err != nil {
		t.Fatalf("seed llm call: %v", err)
	}
}

func seedToolExecution(t *testing.T, d *db.DB, personaID int64, errText string) {
	t.Helper()
	_, err := d.WriteExec(
		`INSERT INTO tool_executions (persona_id, tool_name, args_json, output_text, error_text, duration_ms)
		 VALUES (?, 'shell_exec', '{"cmd":"ls"}', 'file.txt', ?, 42)`,
		personaID, errText,
	)
	if err != nil {
		t.Fatalf("seed tool execution: %v", err)
	}
}

func TestAgentLLMCallsEmpty(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	p, _ := model.CreatePersona(d, "bot", "prompt", "model", nil, 0.7, 100, 0, 0)

	rec := doJSON(t, router, "GET", fmt.Sprintf("/api/agents/%d/llm-calls", p.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var calls []json.RawMessage
	json.NewDecoder(rec.Body).Decode(&calls)
	if len(calls) != 0 {
		t.Errorf("expected 0 calls, got %d", len(calls))
	}
}

func TestAgentLLMCallsWithData(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	p, _ := model.CreatePersona(d, "bot", "prompt", "model", nil, 0.7, 100, 0, 0)
	ch, _ := model.CreateChannel(d, "general", "", 0)

	seedLLMCall(t, d, p.ID, ch.ID)
	seedLLMCall(t, d, p.ID, ch.ID)

	rec := doJSON(t, router, "GET", fmt.Sprintf("/api/agents/%d/llm-calls?limit=1", p.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var calls []struct {
		ID           int64  `json:"id"`
		Model        string `json:"model"`
		PromptTokens int    `json:"prompt_tokens"`
	}
	json.NewDecoder(rec.Body).Decode(&calls)
	if len(calls) != 1 {
		t.Fatalf("expected 1 call (limit=1), got %d", len(calls))
	}
	if calls[0].Model != "test-model" {
		t.Errorf("model = %q, want test-model", calls[0].Model)
	}
	if calls[0].PromptTokens != 100 {
		t.Errorf("prompt_tokens = %d, want 100", calls[0].PromptTokens)
	}
}

func TestAgentToolExecutionsWithData(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	p, _ := model.CreatePersona(d, "bot", "prompt", "model", nil, 0.7, 100, 0, 0)

	seedToolExecution(t, d, p.ID, "")
	seedToolExecution(t, d, p.ID, "command failed")

	rec := doJSON(t, router, "GET", fmt.Sprintf("/api/agents/%d/tool-executions", p.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var execs []struct {
		ToolName   string `json:"tool_name"`
		DurationMs int64  `json:"duration_ms"`
		ErrorText  string `json:"error_text"`
	}
	json.NewDecoder(rec.Body).Decode(&execs)
	if len(execs) != 2 {
		t.Fatalf("expected 2 execs, got %d", len(execs))
	}
	hasError := false
	for _, e := range execs {
		if e.ErrorText == "command failed" {
			hasError = true
		}
		if e.ToolName != "shell_exec" {
			t.Errorf("tool_name = %q, want shell_exec", e.ToolName)
		}
	}
	if !hasError {
		t.Error("expected one exec with error_text='command failed'")
	}
}

func TestAgentStats(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	p, _ := model.CreatePersona(d, "bot", "prompt", "model", nil, 0.7, 100, 0, 0)
	ch, _ := model.CreateChannel(d, "general", "", 0)

	seedLLMCall(t, d, p.ID, ch.ID)
	seedLLMCall(t, d, p.ID, ch.ID)
	seedToolExecution(t, d, p.ID, "")
	seedToolExecution(t, d, p.ID, "oops")

	rec := doJSON(t, router, "GET", fmt.Sprintf("/api/agents/%d/stats", p.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var stats struct {
		TotalCallsLastHour  int64   `json:"total_calls_last_hour"`
		TotalTokensLastHour int64   `json:"total_tokens_last_hour"`
		ErrorCountLastHour  int64   `json:"error_count_last_hour"`
		AvgResponseMs       float64 `json:"avg_response_ms"`
	}
	json.NewDecoder(rec.Body).Decode(&stats)

	if stats.TotalCallsLastHour != 2 {
		t.Errorf("total_calls = %d, want 2", stats.TotalCallsLastHour)
	}
	if stats.TotalTokensLastHour != 300 { // 2 * (100+50)
		t.Errorf("total_tokens = %d, want 300", stats.TotalTokensLastHour)
	}
	if stats.ErrorCountLastHour != 1 {
		t.Errorf("error_count = %d, want 1", stats.ErrorCountLastHour)
	}
	if stats.AvgResponseMs != 42 {
		t.Errorf("avg_response_ms = %f, want 42", stats.AvgResponseMs)
	}
}

func TestAgentLLMCallsUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)

	rec := doJSON(t, router, "GET", "/api/agents/1/llm-calls", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
