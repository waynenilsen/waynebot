package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/waynenilsen/waynebot/internal/agent"
	"github.com/waynenilsen/waynebot/internal/model"
)

func TestContextBudgetEndpoint(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	p, _ := model.CreatePersona(d, "bot", "You are a test bot.", "model", nil, 0.7, 100, 0, 0)
	ch, _ := model.CreateChannel(d, "general", "", 0)

	// Post some messages so there's history to estimate.
	model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello bot!")
	model.CreateMessage(d, ch.ID, p.ID, "agent", "bot", "Hi there!")

	rec := doJSON(t, router, "GET",
		fmt.Sprintf("/api/agents/%d/context-budget?channel_id=%d", p.ID, ch.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var budget struct {
		PersonaID       int64 `json:"persona_id"`
		ChannelID       int64 `json:"channel_id"`
		TotalTokens     int   `json:"total_tokens"`
		SystemTokens    int   `json:"system_tokens"`
		HistoryTokens   int   `json:"history_tokens"`
		HistoryMessages int   `json:"history_messages"`
		Exhausted       bool  `json:"exhausted"`
	}
	json.NewDecoder(rec.Body).Decode(&budget)

	if budget.PersonaID != p.ID {
		t.Errorf("persona_id = %d, want %d", budget.PersonaID, p.ID)
	}
	if budget.ChannelID != ch.ID {
		t.Errorf("channel_id = %d, want %d", budget.ChannelID, ch.ID)
	}
	if budget.TotalTokens != agent.DefaultContextWindow {
		t.Errorf("total_tokens = %d, want %d", budget.TotalTokens, agent.DefaultContextWindow)
	}
	if budget.HistoryMessages != 2 {
		t.Errorf("history_messages = %d, want 2", budget.HistoryMessages)
	}
	if budget.Exhausted {
		t.Error("expected exhausted = false for small history")
	}
}

func TestContextBudgetMissingChannelID(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	p, _ := model.CreatePersona(d, "bot", "prompt", "model", nil, 0.7, 100, 0, 0)

	rec := doJSON(t, router, "GET",
		fmt.Sprintf("/api/agents/%d/context-budget", p.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestResetContextEndpoint(t *testing.T) {
	d := openTestDB(t)
	router, sup := newTestRouterWithSupervisor(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	p, _ := model.CreatePersona(d, "bot", "You are a bot.", "model", nil, 0.7, 100, 0, 0)
	ch, _ := model.CreateChannel(d, "general", "", 0)

	// Post a message so there's a cursor to advance past.
	model.CreateMessage(d, ch.ID, 999, "human", "alice", "Hello!")

	// Set status to context_full.
	sup.Status.Set(p.ID, agent.StatusContextFull)

	rec := doJSON(t, router, "POST",
		fmt.Sprintf("/api/agents/%d/channels/%d/reset-context", p.ID, ch.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	// Status should be reset to idle.
	if sup.Status.Get(p.ID) != agent.StatusIdle {
		t.Errorf("expected status idle after reset, got %s", sup.Status.Get(p.ID))
	}

	// A reset message should be posted.
	msgs, _ := model.GetRecentMessages(d, ch.ID, 10)
	var found bool
	for _, m := range msgs {
		if m.AuthorType == "agent" && m.Content == "Context has been reset. I'm ready to continue." {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected context reset message to be posted")
	}
}

func TestResetContextUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)

	rec := doJSON(t, router, "POST", "/api/agents/1/channels/1/reset-context", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestContextBudgetUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router, _ := newTestRouterWithSupervisor(t, d)

	rec := doJSON(t, router, "GET", "/api/agents/1/context-budget?channel_id=1", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
