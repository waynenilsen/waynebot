# Agent Core (status, cursor, decision, budget)

**Goal:** Create the foundational components of `internal/agent/` — status tracking, cursor store, decision logic, and budget checking.

## Files to Create

- `internal/agent/status.go` — Status enum (idle, thinking, tool_call, error, stopped, budget_exceeded). Goroutine-safe StatusTracker per persona with Get/Set methods using sync.RWMutex.
- `internal/agent/status_test.go`
- `internal/agent/cursor.go` — CursorStore with Get(personaID, channelID) and Set(personaID, channelID, messageID). Uses existing `actor_cursors` table via `*db.DB`.
- `internal/agent/cursor_test.go`
- `internal/agent/decision.go` — `ShouldRespond(persona, channel, messages) bool`: skip if all new messages are from self (by author_type="agent" and author_id=persona.ID), check per-channel cooldown timer, check @mentions.
- `internal/agent/decision_test.go`
- `internal/agent/budget.go` — BudgetChecker: queries `llm_calls` table for SUM(prompt_tokens + completion_tokens) by persona_id in the last hour. Compare against persona.MaxTokensPerHour. Method: `WithinBudget(personaID) (bool, error)`.
- `internal/agent/budget_test.go`

## Key Patterns

- All DB operations follow existing pattern: functions take `*db.DB`, writes use `WriteExec`/`WriteTx`, reads use `d.SQL.Query()`
- Tests use in-memory SQLite via `db.Open(":memory:")`
- StatusTracker is purely in-memory, no DB

## Verification

`go test ./internal/agent/...` passes.
