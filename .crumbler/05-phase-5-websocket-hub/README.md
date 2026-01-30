# Phase 5: WebSocket Hub

**Goal:** Real-time message broadcast to connected clients, with notification channel for agents.

## Files to Create

- `internal/ws/hub.go` — Hub with register/unregister/broadcast. Exposes `NotifyChan` that agents can select on to wake immediately when a new message is posted. (~280 lines)
- `internal/ws/hub_test.go`
- `internal/ws/client.go` — Per-connection read/write pumps with ping/pong keepalive (~200 lines)
- `internal/ws/client_test.go`
- `internal/api/ws_handler.go` — `POST /api/ws/ticket` to mint ticket, `GET /ws?ticket=...` to upgrade. Ticket is single-use, expires in 30s. (~180 lines)
- `internal/api/ws_handler_test.go`

## Integration

`PostMessage` handler broadcasts via hub after insert AND sends on `NotifyChan`. Start hub goroutine in `main.go`.

## WebSocket Auth

Instead of passing the session token in the WebSocket URL query string (which leaks into logs), the client first calls `POST /api/ws/ticket` to exchange its session token for a single-use, 30-second ticket. The ticket is then passed as `GET /ws?ticket=...`. The ticket is deleted on use or expiration.

## Verification

`go test ./internal/ws/... ./internal/api/...` passes.
