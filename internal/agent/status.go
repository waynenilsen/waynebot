package agent

import (
	"maps"
	"sync"
)

// Status represents the current state of an agent persona.
type Status int

const (
	StatusIdle Status = iota
	StatusThinking
	StatusToolCall
	StatusError
	StatusStopped
	StatusBudgetExceeded
	StatusContextFull
)

func (s Status) String() string {
	switch s {
	case StatusIdle:
		return "idle"
	case StatusThinking:
		return "thinking"
	case StatusToolCall:
		return "tool_call"
	case StatusError:
		return "error"
	case StatusStopped:
		return "stopped"
	case StatusBudgetExceeded:
		return "budget_exceeded"
	case StatusContextFull:
		return "context_full"
	default:
		return "unknown"
	}
}

// StatusTracker tracks the current status of each persona. Goroutine-safe.
type StatusTracker struct {
	mu       sync.RWMutex
	statuses map[int64]Status
}

// NewStatusTracker creates a new StatusTracker.
func NewStatusTracker() *StatusTracker {
	return &StatusTracker{statuses: make(map[int64]Status)}
}

// Get returns the current status for a persona. Returns StatusIdle if not set.
func (st *StatusTracker) Get(personaID int64) Status {
	st.mu.RLock()
	defer st.mu.RUnlock()
	s, ok := st.statuses[personaID]
	if !ok {
		return StatusIdle
	}
	return s
}

// Set updates the status for a persona.
func (st *StatusTracker) Set(personaID int64, s Status) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.statuses[personaID] = s
}

// All returns a snapshot of all persona statuses.
func (st *StatusTracker) All() map[int64]Status {
	st.mu.RLock()
	defer st.mu.RUnlock()
	out := make(map[int64]Status, len(st.statuses))
	maps.Copy(out, st.statuses)
	return out
}
