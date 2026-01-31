# 003 — "Failed to start agents: agents already running" while UI shows STOPPED

## Bug

Clicking **start all** on the Agents admin page returns the error _"Failed to start agents: agents already running"_ even though every agent is displayed as **STOPPED**.

## Root cause

There are two separate notions of "running" that disagree:

1. **Supervisor-level flag** (`Supervisor.running` bool in `internal/agent/supervisor.go:49-54`) — set to `true` by `StartAll()` and `false` by `StopAll()`. The start endpoint (`internal/api/agent_handlers.go:64`) rejects the request when this flag is `true`.

2. **Per-persona status** (`internal/agent/status.go`) — individual states: `idle`, `thinking`, `tool_call`, `error`, `stopped`, `budget_exceeded`. There is no `"running"` string.

The frontend (`frontend/src/pages/AgentDashboard.tsx:11,63`) checks `status === "running"` to decide what to display, but the backend never returns that string. So agents in `idle` state show as "STOPPED" in the UI, even though the supervisor considers itself running.

### Trigger sequence

1. Server starts → `supervisor.StartAll()` → `running = true`, actors go to `idle`.
2. Frontend fetches `/agents/status` → gets `"idle"` for each agent.
3. Frontend: `"idle" === "running"` → false → renders **STOPPED**.
4. User clicks **start all** → backend: `Supervisor.Running()` → `true` → 409 "agents already running".

## Files involved

| File | Lines | Role |
|------|-------|------|
| `internal/agent/supervisor.go` | 49-54, 77, 88 | `Running()` bool, set in `StartAll`/`StopAll` |
| `internal/api/agent_handlers.go` | 64 | Guards start with `Supervisor.Running()` |
| `internal/agent/status.go` | 20-37 | Status enum — no `"running"` variant |
| `frontend/src/pages/AgentDashboard.tsx` | 11, 63 | Compares against `"running"` string that backend never sends |

## Possible fixes

- **Option A**: Frontend considers `idle`, `thinking`, `tool_call` as "running" states (only `stopped`, `error`, `budget_exceeded` are non-running).
- **Option B**: Backend `/agents/status` response includes a top-level `supervisorRunning` bool so the frontend can distinguish supervisor state from per-agent state.
- **Option C**: Both — fix the frontend status mapping and expose the supervisor flag for the start/stop button logic.
