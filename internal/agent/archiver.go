package agent

import (
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

const (
	maxEventsPerAgent = 10_000
	purgeInterval     = 5 * time.Minute
)

// Archiver purges old llm_calls and tool_executions rows, compressing them to
// gzip archives on disk before deletion.
type Archiver struct {
	DB         *db.DB
	ArchiveDir string
}

// Run starts the purge loop. It blocks until ctx is cancelled.
func (a *Archiver) Run(ctx context.Context) {
	ticker := time.NewTicker(purgeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.purgeAll()
		}
	}
}

func (a *Archiver) purgeAll() {
	personaIDs, err := a.personaIDsWithExcess("llm_calls")
	if err != nil {
		slog.Error("archiver: list personas for llm_calls", "error", err)
	}
	for _, pid := range personaIDs {
		if err := a.purgeTable(pid, "llm_calls"); err != nil {
			slog.Error("archiver: purge llm_calls", "persona_id", pid, "error", err)
		}
	}

	personaIDs, err = a.personaIDsWithExcess("tool_executions")
	if err != nil {
		slog.Error("archiver: list personas for tool_executions", "error", err)
	}
	for _, pid := range personaIDs {
		if err := a.purgeTable(pid, "tool_executions"); err != nil {
			slog.Error("archiver: purge tool_executions", "persona_id", pid, "error", err)
		}
	}
}

// personaIDsWithExcess returns persona IDs that have more than maxEventsPerAgent rows.
func (a *Archiver) personaIDsWithExcess(table string) ([]int64, error) {
	query := fmt.Sprintf(
		`SELECT persona_id FROM %s GROUP BY persona_id HAVING COUNT(*) > ?`, table,
	)
	rows, err := a.DB.SQL.Query(query, maxEventsPerAgent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// purgeTable archives and deletes excess rows for a single persona in a single table.
func (a *Archiver) purgeTable(personaID int64, table string) error {
	// Find the cutoff ID: keep the newest maxEventsPerAgent rows.
	var cutoffID int64
	query := fmt.Sprintf(
		`SELECT id FROM %s WHERE persona_id = ? ORDER BY id DESC LIMIT 1 OFFSET ?`, table,
	)
	err := a.DB.SQL.QueryRow(query, personaID, maxEventsPerAgent).Scan(&cutoffID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("find cutoff: %w", err)
	}

	// Read rows to archive.
	selectQuery := fmt.Sprintf(
		`SELECT * FROM %s WHERE persona_id = ? AND id <= ? ORDER BY id`, table,
	)
	rows, err := a.DB.SQL.Query(selectQuery, personaID, cutoffID)
	if err != nil {
		return fmt.Errorf("select rows: %w", err)
	}

	archivePath, err := a.writeArchive(rows, table, personaID)
	rows.Close()
	if err != nil {
		return fmt.Errorf("write archive: %w", err)
	}

	// Delete archived rows inside the write mutex.
	deleteQuery := fmt.Sprintf(
		`DELETE FROM %s WHERE persona_id = ? AND id <= ?`, table,
	)
	result, err := a.DB.WriteExec(deleteQuery, personaID, cutoffID)
	if err != nil {
		return fmt.Errorf("delete archived rows: %w", err)
	}

	deleted, _ := result.RowsAffected()
	slog.Info("archiver: purged rows", "table", table, "persona_id", personaID, "deleted", deleted, "archive", archivePath)
	return nil
}

// writeArchive writes rows as jsonl.gz and returns the archive path.
func (a *Archiver) writeArchive(rows *sql.Rows, table string, personaID int64) (string, error) {
	if err := os.MkdirAll(a.ArchiveDir, 0o755); err != nil {
		return "", fmt.Errorf("create archive dir: %w", err)
	}

	ts := time.Now().UTC().Format("20060102T150405Z")
	filename := fmt.Sprintf("%s_%d_%s.jsonl.gz", table, personaID, ts)
	path := filepath.Join(a.ArchiveDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	enc := json.NewEncoder(gz)

	cols, err := rows.Columns()
	if err != nil {
		os.Remove(path)
		return "", fmt.Errorf("get columns: %w", err)
	}

	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			os.Remove(path)
			return "", fmt.Errorf("scan row: %w", err)
		}

		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = values[i]
		}

		if err := enc.Encode(row); err != nil {
			os.Remove(path)
			return "", fmt.Errorf("encode row: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		os.Remove(path)
		return "", fmt.Errorf("rows iteration: %w", err)
	}

	return path, nil
}
