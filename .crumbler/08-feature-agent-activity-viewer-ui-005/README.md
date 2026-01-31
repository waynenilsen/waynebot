# 005: Agent activity viewer in frontend

## Depends on

- 004 (full token capture in DB)

## Problem

There's no way to see what an agent is doing from the UI. The Agent Dashboard only shows running/stopped status and channel subscriptions. Users need to be able to inspect LLM calls, tool executions, token usage, and errors to understand and debug agent behavior.

## Requirements

### Backend — new API endpoints

- `GET /api/agents/{persona_id}/llm-calls?limit=50&offset=0` — paginated list of LLM calls with full messages/response JSON, token counts, timestamps
- `GET /api/agents/{persona_id}/tool-executions?limit=50&offset=0` — paginated list of tool executions with args, output, errors, duration
- `GET /api/agents/{persona_id}/stats` — summary: total calls in last hour, total tokens in last hour, error count, average response time

### Frontend — agent activity page

Accessible from the Agent Dashboard by clicking on an agent. Shows:

- **Status header** — current status (thinking/idle/tool_call/error), uptime, subscribed channels
- **Stats bar** — tokens used in last hour vs budget, call count, error count
- **Activity feed** — reverse-chronological list of LLM calls and tool executions interleaved by timestamp
  - LLM call entries: model, token counts, expandable to show full messages/response
  - Tool execution entries: tool name, duration, expandable to show args/output/error
- **Pagination** — load more button or infinite scroll, don't load 10k rows at once

### Live updates

- When viewing an agent's activity page and the agent is running, new events should appear at the top via WebSocket (new event type `agent_activity` or poll on interval)

## Files likely involved

| Area | Files |
|------|-------|
| Backend API | `internal/api/agent_handlers.go`, `internal/api/router.go` |
| Backend DB queries | `internal/db/` or inline in handlers |
| Frontend page | new `AgentActivityPage.tsx` |
| Frontend routing | wherever page routing is configured |
| Frontend types | `frontend/src/types.ts` |
| Frontend API | `frontend/src/api.ts` |
