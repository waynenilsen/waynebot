package api_test

import (
	"context"
	"encoding/json"
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

	sup := agent.NewSupervisor(d, hub, idleLLM{}, tools.NewRegistry())
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
