# Phase 6: Agent Processing Loop

**Goal:** Personas run as background goroutines, call OpenRouter, use sandboxed tools, post responses.

## Files to Create

### Agent Package
- `internal/agent/status.go` — Status enum (idle, thinking, tool_call, error, stopped, budget_exceeded), goroutine-safe tracker (~120 lines)
- `internal/agent/status_test.go`
- `internal/agent/cursor.go` — CursorStore (SQLite-backed last_seen_message_id per persona per channel) (~120 lines)
- `internal/agent/cursor_test.go`
- `internal/agent/decision.go` — `shouldRespond()`: skip self-messages, check cooldown, check @mentions (~200 lines)
- `internal/agent/decision_test.go`
- `internal/agent/budget.go` — BudgetChecker: queries llm_calls for total tokens used by persona in last hour, compares to max_tokens_per_hour (~150 lines)
- `internal/agent/budget_test.go`
- `internal/agent/actor.go` — Actor struct, Run(ctx) loop: select on ctx.Done(), hub.NotifyChan, or fallback ticker → check cursor → decide → build history → check budget → LLM loop → post response → broadcast (~450 lines)
- `internal/agent/actor_test.go` — Tests with mock LLM client
- `internal/agent/supervisor.go` — Start/stop/monitor actors as goroutines with context cancellation. StopAll() for shutdown. (~350 lines)
- `internal/agent/supervisor_test.go`

### LLM Package
- `internal/llm/client.go` — openai-go client with OpenRouter base URL. ChatCompletion(). Returns token counts. (~250 lines)
- `internal/llm/client_test.go` — Tests with mock HTTP server
- `internal/llm/messages.go` — Convert domain Message to SDK message types (~200 lines)
- `internal/llm/messages_test.go`
- `internal/llm/tools.go` — Tool JSON schema definitions (~150 lines)
- `internal/llm/tools_test.go`

### Tools Package
- `internal/tools/sandbox.go` — SandboxConfig: allowed commands list, allowed base directory, blocked hosts (~150 lines)
- `internal/tools/sandbox_test.go`
- `internal/tools/registry.go` — Registry maps tool names to ToolFunc (~120 lines)
- `internal/tools/registry_test.go`
- `internal/tools/shell_exec.go` — Sandboxed shell command (~200 lines)
- `internal/tools/shell_exec_test.go`
- `internal/tools/file_read.go` — Path-restricted file read (~120 lines)
- `internal/tools/file_read_test.go`
- `internal/tools/file_write.go` — Path-restricted file write (~120 lines)
- `internal/tools/file_write_test.go`
- `internal/tools/http_fetch.go` — HTTP fetch with denied hosts (~130 lines)
- `internal/tools/http_fetch_test.go`

### API Handlers
- `internal/api/agent_handlers.go` — GET status, POST start, POST stop (~200 lines)
- `internal/api/agent_handlers_test.go`

## Anti-Loop Strategy

1. Skip if all new msgs from self
2. Per-channel cooldown timer (configurable per persona, default 30s)
3. Max 10 tool-call rounds per LLM invocation
4. Token budget per persona per hour

## Tool Sandboxing

- shell_exec: Configurable allowed-command allowlist, 30s timeout, 10KB output
- file_read: Restricted to sandbox dir, no path traversal, 1MB max
- file_write: Same as file_read
- http_fetch: Blocked hosts list (localhost, etc), 15s timeout, 50KB max

## Verification

`go test ./internal/agent/... ./internal/llm/... ./internal/tools/...` passes.
