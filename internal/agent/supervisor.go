package agent

import (
	"context"
	"log/slog"
	"sync"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/tools"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// Supervisor manages actor goroutines, one per persona.
type Supervisor struct {
	DB        *db.DB
	Hub       *ws.Hub
	LLM       LLMClient
	Embedding EmbeddingClient
	Tools     *tools.Registry
	Status    *StatusTracker
	Cursors   *CursorStore
	Decision  *DecisionMaker
	Budget    *BudgetChecker

	mu      sync.Mutex
	actors  map[int64]actorHandle
	wg      sync.WaitGroup
	running bool
}

type actorHandle struct {
	cancel context.CancelFunc
}

// NewSupervisor creates a Supervisor with all required dependencies.
func NewSupervisor(database *db.DB, hub *ws.Hub, llmClient LLMClient, embeddingClient EmbeddingClient, toolsRegistry *tools.Registry) *Supervisor {
	return &Supervisor{
		DB:        database,
		Hub:       hub,
		LLM:       llmClient,
		Embedding: embeddingClient,
		Tools:     toolsRegistry,
		Status:    NewStatusTracker(),
		Cursors:   NewCursorStore(database),
		Decision:  NewDecisionMaker(),
		Budget:    NewBudgetChecker(database),
	}
}

// Running returns true if the supervisor has been started.
func (s *Supervisor) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// StartAll lists all personas from the DB and starts an actor goroutine for each.
func (s *Supervisor) StartAll() error {
	personas, err := model.ListPersonas(s.DB)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.actors == nil {
		s.actors = make(map[int64]actorHandle)
	}

	for _, p := range personas {
		if _, exists := s.actors[p.ID]; exists {
			continue
		}
		s.startActorLocked(p)
	}

	s.running = true
	return nil
}

// StopAll cancels all actor contexts and waits for goroutines to finish.
func (s *Supervisor) StopAll() {
	s.mu.Lock()
	for id, h := range s.actors {
		h.cancel()
		delete(s.actors, id)
	}
	s.running = false
	s.mu.Unlock()

	s.wg.Wait()
}

// RestartActor stops and restarts a single actor by persona ID.
func (s *Supervisor) RestartActor(personaID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop existing actor if running.
	if h, ok := s.actors[personaID]; ok {
		h.cancel()
		delete(s.actors, personaID)
	}

	persona, err := model.GetPersona(s.DB, personaID)
	if err != nil {
		return err
	}

	s.startActorLocked(persona)
	return nil
}

// startActorLocked starts a goroutine for the given persona. Must be called with s.mu held.
func (s *Supervisor) startActorLocked(p model.Persona) {
	ctx, cancel := context.WithCancel(context.Background())
	s.actors[p.ID] = actorHandle{cancel: cancel}

	actor := &Actor{
		Persona:   p,
		DB:        s.DB,
		Hub:       s.Hub,
		LLM:       s.LLM,
		Embedding: s.Embedding,
		Tools:     s.Tools,
		Status:    s.Status,
		Cursors:   s.Cursors,
		Decision:  s.Decision,
		Budget:    s.Budget,
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		slog.Info("supervisor: starting actor", "persona", p.Name, "persona_id", p.ID)
		actor.Run(ctx)
		slog.Info("supervisor: actor stopped", "persona", p.Name, "persona_id", p.ID)
	}()
}
