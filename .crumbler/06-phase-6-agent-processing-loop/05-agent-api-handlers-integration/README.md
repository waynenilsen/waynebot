# Agent API Handlers & Integration

**Goal:** Add API endpoints for agent status/control and wire the supervisor into main.go.

## Files to Create/Modify

### New Files
- `internal/api/agent_handlers.go` — AgentHandler struct with DB + Supervisor:
  - `GET /api/agents/status` — Returns status of all persona actors (persona name, status enum, channels)
  - `POST /api/agents/start` — Start the supervisor (starts all persona actors)
  - `POST /api/agents/stop` — Stop the supervisor (stops all actors)
  All endpoints require auth.
- `internal/api/agent_handlers_test.go`

### Modified Files
- `internal/api/router.go` — Add agent routes, accept Supervisor dependency
- `cmd/waynebot/main.go` — Create LLM client, tools registry, supervisor. Start supervisor after hub. Stop supervisor before hub on shutdown.

## Route Additions

```
GET  /api/agents/status   [auth required]
POST /api/agents/start    [auth required]
POST /api/agents/stop     [auth required]
```

## Integration in main.go

```go
// After hub setup:
llmClient := llm.NewClient(cfg.OpenRouterKey)
toolsRegistry := tools.NewRegistry(tools.SandboxConfig{...})
supervisor := agent.NewSupervisor(database, hub, llmClient, toolsRegistry)
supervisor.StartAll()

// On shutdown:
supervisor.StopAll()
hub.Stop()
```

## Verification

`go test ./internal/...` passes. Full build succeeds with `go build ./...`.
