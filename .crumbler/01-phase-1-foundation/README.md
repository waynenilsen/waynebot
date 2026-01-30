# Phase 1: Foundation

**Goal:** Binary starts, opens SQLite, runs versioned migrations, exits on signal. Backend HTTP server must run on port 59731.

## Files to Create

- `cmd/waynebot/main.go` — Entrypoint (~120 lines)
- `internal/config/config.go` — Env-based config: port (59731), DB path, sandbox dir, CORS origins, IMAP settings, OpenRouter key. All ports and paths configurable via env vars. (~140 lines)
- `internal/db/db.go` — Open DB, WAL mode, `busy_timeout=5000`, write mutex (~130 lines)
- `internal/db/db_test.go` — DB open/close, WAL mode verification
- `internal/db/migrations.go` — Versioned migrations with `schema_version` table (~300 lines)
- `internal/db/migrations_test.go` — Migration idempotency, version tracking

## Go Dependencies

```
github.com/go-chi/chi/v5
github.com/go-chi/cors
github.com/go-chi/httprate
github.com/coder/websocket
github.com/openai/openai-go
github.com/emersion/go-imap/v2
golang.org/x/crypto
modernc.org/sqlite
```

## SQLite Schema

See root README for full schema. Migrations are tracked in a `schema_version` table. Each migration has a version number and is applied exactly once.

## Cross-Cutting: SQLite Write Contention

All database writes go through a single `*sync.Mutex` (the "write mutex") in `db.go`. This serializes writes and avoids `SQLITE_BUSY` errors. Reads are not serialized. The connection sets `PRAGMA busy_timeout = 5000` as a safety net.

## Verification

- `go build ./cmd/waynebot` compiles
- `go run ./cmd/waynebot` creates `waynebot.db` with all tables
- `go test ./internal/db/...` passes

## Also

- Create a `.gitignore` for the project (Go binaries, `*.db`, node_modules, dist, etc.)
- Initialize `go.mod` with `module github.com/waynenilsen/waynebot`
