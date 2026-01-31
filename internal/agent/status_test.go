package agent

import (
	"sync"
	"testing"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		s    Status
		want string
	}{
		{StatusIdle, "idle"},
		{StatusThinking, "thinking"},
		{StatusToolCall, "tool_call"},
		{StatusError, "error"},
		{StatusStopped, "stopped"},
		{StatusBudgetExceeded, "budget_exceeded"},
		{Status(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("Status(%d).String() = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestStatusTrackerDefaultIdle(t *testing.T) {
	st := NewStatusTracker()
	if got := st.Get(42); got != StatusIdle {
		t.Errorf("Get(42) = %v, want StatusIdle", got)
	}
}

func TestStatusTrackerSetGet(t *testing.T) {
	st := NewStatusTracker()
	st.Set(1, StatusThinking)
	st.Set(2, StatusToolCall)

	if got := st.Get(1); got != StatusThinking {
		t.Errorf("Get(1) = %v, want StatusThinking", got)
	}
	if got := st.Get(2); got != StatusToolCall {
		t.Errorf("Get(2) = %v, want StatusToolCall", got)
	}

	st.Set(1, StatusError)
	if got := st.Get(1); got != StatusError {
		t.Errorf("after update Get(1) = %v, want StatusError", got)
	}
}

func TestStatusTrackerConcurrent(t *testing.T) {
	st := NewStatusTracker()
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := int64(i)
			st.Set(id%5, Status(id%6))
			_ = st.Get(id % 5)
		}()
	}
	wg.Wait()
}
