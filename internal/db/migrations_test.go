package db

import (
	"testing"
)

func TestMigrationsApplied(t *testing.T) {
	d, err := Open(tempDBPath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	v, err := d.SchemaVersion()
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if v != len(migrations) {
		t.Errorf("schema version = %d, want %d", v, len(migrations))
	}
}

func TestMigrationsIdempotent(t *testing.T) {
	path := tempDBPath(t)

	d1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	d1.Close()

	d2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer d2.Close()

	v, err := d2.SchemaVersion()
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if v != len(migrations) {
		t.Errorf("schema version = %d, want %d", v, len(migrations))
	}
}

func TestTablesExist(t *testing.T) {
	d, err := Open(tempDBPath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	tables := []string{
		"users", "sessions", "ws_tickets", "invites",
		"channels", "messages", "personas", "persona_channels",
		"actor_cursors", "llm_calls", "tool_executions",
	}

	for _, table := range tables {
		var name string
		err := d.SQL.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}
