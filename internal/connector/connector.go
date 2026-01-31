package connector

import (
	"context"
	"log/slog"
	"sync"
)

// Connector represents a long-running background service that bridges
// an external system (email, RSS, etc.) into waynebot channels.
type Connector interface {
	// Name returns a human-readable identifier for logging.
	Name() string
	// Run starts the connector's poll loop. It blocks until ctx is cancelled.
	Run(ctx context.Context)
}

// Registry manages a set of connectors, starting and stopping them together.
type Registry struct {
	mu         sync.Mutex
	connectors []Connector
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewRegistry creates an empty connector registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a connector. Must be called before StartAll.
func (r *Registry) Register(c Connector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connectors = append(r.connectors, c)
}

// StartAll launches every registered connector in its own goroutine.
func (r *Registry) StartAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	for _, c := range r.connectors {
		r.wg.Add(1)
		go func(c Connector) {
			defer r.wg.Done()
			slog.Info("connector started", "name", c.Name())
			c.Run(ctx)
			slog.Info("connector stopped", "name", c.Name())
		}(c)
	}
}

// StopAll cancels all running connectors and waits for them to finish.
func (r *Registry) StopAll() {
	r.mu.Lock()
	cancel := r.cancel
	r.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	r.wg.Wait()
}

// Count returns the number of registered connectors.
func (r *Registry) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.connectors)
}
