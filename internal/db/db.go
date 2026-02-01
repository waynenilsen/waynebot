package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

// DB wraps a *sql.DB with a write mutex to serialize all writes.
type DB struct {
	SQL     *sql.DB
	writeMu sync.Mutex
}

// Open opens (or creates) a SQLite database at path, enables WAL mode
// and sets busy_timeout. It then runs all pending migrations.
func Open(path string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// SQLite pragmas
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := sqlDB.Exec(p); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("exec %q: %w", p, err)
		}
	}

	d := &DB{SQL: sqlDB}

	if err := d.migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return d, nil
}

// Close closes the underlying database.
func (d *DB) Close() error {
	return d.SQL.Close()
}

// WriteExec executes a write statement under the write mutex.
func (d *DB) WriteExec(query string, args ...any) (sql.Result, error) {
	d.writeMu.Lock()
	defer d.writeMu.Unlock()
	return d.SQL.Exec(query, args...)
}

// WriteTx runs fn inside a transaction under the write mutex.
func (d *DB) WriteTx(fn func(tx *sql.Tx) error) error {
	d.writeMu.Lock()
	defer d.writeMu.Unlock()

	tx, err := d.SQL.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
