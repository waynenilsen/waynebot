package connector

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type fakeConnector struct {
	name  string
	runs  atomic.Int32
	block chan struct{}
}

func (f *fakeConnector) Name() string { return f.name }

func (f *fakeConnector) Run(ctx context.Context) {
	f.runs.Add(1)
	select {
	case <-ctx.Done():
	case <-f.block:
	}
}

func TestRegistryStartStop(t *testing.T) {
	r := NewRegistry()
	c1 := &fakeConnector{name: "a", block: make(chan struct{})}
	c2 := &fakeConnector{name: "b", block: make(chan struct{})}

	r.Register(c1)
	r.Register(c2)

	if r.Count() != 2 {
		t.Fatalf("expected 2 connectors, got %d", r.Count())
	}

	r.StartAll()

	// Wait for both goroutines to start.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if c1.runs.Load() > 0 && c2.runs.Load() > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if c1.runs.Load() == 0 {
		t.Fatal("connector a did not start")
	}
	if c2.runs.Load() == 0 {
		t.Fatal("connector b did not start")
	}

	r.StopAll()
}

func TestRegistryEmpty(t *testing.T) {
	r := NewRegistry()
	r.StartAll()
	r.StopAll() // should not panic
}
