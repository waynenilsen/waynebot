# Phase 9: Hardening

**Goal:** Production readiness.

## Tasks

- Session cleanup goroutine (every 15 min, also cleans expired ws_tickets)
- Rate limiting (60/min auth, 120/min other)
- Structured logging (slog.JSONHandler)
- Graceful shutdown sequence: HTTP → connectors → agents → WS hub → DB
- Health endpoint (GET /health, unauthenticated)
- CORS config via env var
- LLM call + tool execution logging to llm_calls / tool_executions tables

## Verification

Shutdown gracefully, rate limits work, logs are structured JSON.
