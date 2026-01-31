package agent

import (
	"context"
	"testing"
	"time"

	"github.com/openai/openai-go"
	"github.com/waynenilsen/waynebot/internal/llm"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/tools"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// idleLLM never gets called — returns empty response if it does.
type idleLLM struct{}

func (idleLLM) ChatCompletion(_ context.Context, _ string, _ []openai.ChatCompletionMessageParamUnion, _ []openai.ChatCompletionToolParam, _ float64, _ int) (llm.Response, error) {
	return llm.Response{}, nil
}

func newSupervisor(t *testing.T) (*Supervisor, *ws.Hub) {
	t.Helper()
	d := openTestDB(t)
	hub := ws.NewHub()
	done := make(chan struct{})
	go func() {
		hub.Run()
		close(done)
	}()
	t.Cleanup(func() {
		hub.Stop()
		<-done
	})

	return &Supervisor{
		DB:       d,
		Hub:      hub,
		LLM:      idleLLM{},
		Tools:    tools.NewRegistry(),
		Status:   NewStatusTracker(),
		Cursors:  NewCursorStore(d),
		Decision: NewDecisionMaker(),
		Budget:   NewBudgetChecker(d),
	}, hub
}

func TestSupervisorStartAllAndStopAll(t *testing.T) {
	sup, _ := newSupervisor(t)

	// Create two personas.
	p1, err := model.CreatePersona(sup.DB, "bot1", "prompt", "model", nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona 1: %v", err)
	}
	p2, err := model.CreatePersona(sup.DB, "bot2", "prompt", "model", nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona 2: %v", err)
	}

	if err := sup.StartAll(); err != nil {
		t.Fatalf("StartAll: %v", err)
	}

	// Wait for actors to register themselves as idle.
	waitFor(t, func() bool {
		return sup.Status.Get(p1.ID) == StatusIdle && sup.Status.Get(p2.ID) == StatusIdle
	})

	sup.StopAll()

	// After stop, statuses should be stopped.
	if sup.Status.Get(p1.ID) != StatusStopped {
		t.Errorf("persona 1: expected stopped, got %s", sup.Status.Get(p1.ID))
	}
	if sup.Status.Get(p2.ID) != StatusStopped {
		t.Errorf("persona 2: expected stopped, got %s", sup.Status.Get(p2.ID))
	}
}

func TestSupervisorRestartActor(t *testing.T) {
	sup, _ := newSupervisor(t)

	p, err := model.CreatePersona(sup.DB, "bot", "prompt", "model", nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	if err := sup.StartAll(); err != nil {
		t.Fatalf("StartAll: %v", err)
	}

	waitFor(t, func() bool {
		return sup.Status.Get(p.ID) == StatusIdle
	})

	// Restart should not error.
	if err := sup.RestartActor(p.ID); err != nil {
		t.Fatalf("RestartActor: %v", err)
	}

	// Actor should come back to idle.
	waitFor(t, func() bool {
		return sup.Status.Get(p.ID) == StatusIdle
	})

	sup.StopAll()
}

func TestSupervisorStartAllIdempotent(t *testing.T) {
	sup, _ := newSupervisor(t)

	_, err := model.CreatePersona(sup.DB, "bot", "prompt", "model", nil, 0.7, 100, 0, 0)
	if err != nil {
		t.Fatalf("create persona: %v", err)
	}

	if err := sup.StartAll(); err != nil {
		t.Fatalf("StartAll 1: %v", err)
	}

	// Calling again should not start duplicate actors.
	if err := sup.StartAll(); err != nil {
		t.Fatalf("StartAll 2: %v", err)
	}

	sup.mu.Lock()
	count := len(sup.actors)
	sup.mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 actor, got %d", count)
	}

	sup.StopAll()
}

func TestSupervisorStopAllWithNoActors(t *testing.T) {
	sup, _ := newSupervisor(t)

	// Should not panic or block.
	sup.StopAll()
}

// waitFor polls a condition with a timeout.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("waitFor timed out")
}
