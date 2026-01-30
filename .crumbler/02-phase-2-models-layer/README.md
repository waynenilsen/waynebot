# Phase 2: Models Layer

**Goal:** All Go structs and query functions, fully tested.

## Files to Create

- `internal/model/user.go` — User struct, CRUD + `CountUsers()` for bootstrap check (~200 lines)
- `internal/model/user_test.go`
- `internal/model/session.go` — Session struct, Create, get by token, delete, cleanup expired (~180 lines)
- `internal/model/session_test.go`
- `internal/model/invite.go` — Invite struct, Create, `ClaimInvite(code, userID)` using `UPDATE invites SET used_by=? WHERE code=? AND used_by IS NULL` (atomic, returns rows-affected to detect race), list (~180 lines)
- `internal/model/invite_test.go` — Includes concurrent claim test
- `internal/model/channel.go` — Channel struct, CRUD, list (~200 lines)
- `internal/model/channel_test.go`
- `internal/model/message.go` — Message struct, Create, list paginated (cursor-based: `?before=messageId`), `GetMessagesSince(channelID, afterID)`, `GetRecentMessages(channelID, limit)` (~250 lines)
- `internal/model/message_test.go`
- `internal/model/persona.go` — Persona struct, CRUD, subscribe/unsubscribe channels, get subscribed channels (~250 lines)
- `internal/model/persona_test.go`

## Go Structs

See root README for canonical Go structs (Message, Persona, etc.)

## Testing

All model tests run against `:memory:` SQLite with migrations applied.

## Verification

`go test ./internal/model/...` passes.
