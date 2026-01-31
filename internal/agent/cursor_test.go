package agent

import (
	"testing"

	"github.com/waynenilsen/waynebot/internal/db"
)

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestCursorGetDefault(t *testing.T) {
	d := openTestDB(t)
	cs := NewCursorStore(d)

	got, err := cs.Get(1, 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != 0 {
		t.Errorf("Get(1,1) = %d, want 0", got)
	}
}

func TestCursorSetAndGet(t *testing.T) {
	d := openTestDB(t)
	cs := NewCursorStore(d)

	if err := cs.Set(1, 1, 42); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := cs.Get(1, 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != 42 {
		t.Errorf("Get(1,1) = %d, want 42", got)
	}
}

func TestCursorUpsert(t *testing.T) {
	d := openTestDB(t)
	cs := NewCursorStore(d)

	if err := cs.Set(1, 1, 10); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := cs.Set(1, 1, 20); err != nil {
		t.Fatalf("Set update: %v", err)
	}

	got, err := cs.Get(1, 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != 20 {
		t.Errorf("Get(1,1) = %d, want 20", got)
	}
}

func TestCursorIsolation(t *testing.T) {
	d := openTestDB(t)
	cs := NewCursorStore(d)

	if err := cs.Set(1, 1, 100); err != nil {
		t.Fatalf("Set(1,1): %v", err)
	}
	if err := cs.Set(1, 2, 200); err != nil {
		t.Fatalf("Set(1,2): %v", err)
	}
	if err := cs.Set(2, 1, 300); err != nil {
		t.Fatalf("Set(2,1): %v", err)
	}

	tests := []struct {
		persona, channel, want int64
	}{
		{1, 1, 100},
		{1, 2, 200},
		{2, 1, 300},
		{2, 2, 0},
	}
	for _, tt := range tests {
		got, err := cs.Get(tt.persona, tt.channel)
		if err != nil {
			t.Fatalf("Get(%d,%d): %v", tt.persona, tt.channel, err)
		}
		if got != tt.want {
			t.Errorf("Get(%d,%d) = %d, want %d", tt.persona, tt.channel, got, tt.want)
		}
	}
}
