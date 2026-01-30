package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func tempDBPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "test.db")
}

func TestOpenClose(t *testing.T) {
	path := tempDBPath(t)
	d, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("db file not created: %v", err)
	}
}

func TestWALMode(t *testing.T) {
	d, err := Open(tempDBPath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	var mode string
	if err := d.SQL.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want %q", mode, "wal")
	}
}

func TestForeignKeysEnabled(t *testing.T) {
	d, err := Open(tempDBPath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	var fk int
	if err := d.SQL.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}
}

func TestWriteExec(t *testing.T) {
	d, err := Open(tempDBPath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	_, err = d.WriteExec("INSERT INTO channels (name) VALUES (?)", "general")
	if err != nil {
		t.Fatalf("WriteExec: %v", err)
	}

	var name string
	err = d.SQL.QueryRow("SELECT name FROM channels WHERE name = ?", "general").Scan(&name)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if name != "general" {
		t.Errorf("name = %q, want %q", name, "general")
	}
}

func TestWriteTx(t *testing.T) {
	d, err := Open(tempDBPath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer d.Close()

	err = d.WriteTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO channels (name) VALUES (?)", "random")
		return err
	})
	if err != nil {
		t.Fatalf("WriteTx: %v", err)
	}

	var count int
	d.SQL.QueryRow("SELECT COUNT(*) FROM channels WHERE name = ?", "random").Scan(&count)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}
