# Phase 4: Channel + Message + Persona + Invite APIs

**Goal:** Full REST CRUD, with tests.

## Files to Create

- `internal/api/channel_handlers.go` — List, create, get messages (cursor-based pagination), post message (~300 lines)
- `internal/api/channel_handlers_test.go`
- `internal/api/persona_handlers.go` — List, create, update, delete (cascades subscriptions) (~220 lines)
- `internal/api/persona_handlers_test.go`
- `internal/api/invite_handlers.go` — Generate, list (~180 lines)
- `internal/api/invite_handlers_test.go`

## API Endpoints

```
GET    /api/channels
POST   /api/channels                 # Create { name, description }
GET    /api/channels/:id/messages    # Paginated: ?limit=50&before=<messageId>
POST   /api/channels/:id/messages    # Send { content }

GET    /api/personas
POST   /api/personas
PUT    /api/personas/:id
DELETE /api/personas/:id

POST   /api/invites
GET    /api/invites
```

## Input Validation

- `channel.name`: 1–100 chars
- `channel.description`: 0–500 chars
- `message.content`: 1–10000 chars
- `persona.name`: 1–100 chars
- `persona.system_prompt`: 1–50000 chars

## Verification

`go test ./internal/api/...` passes.
