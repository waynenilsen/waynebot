# Phase 8: Email Connector

**Goal:** IMAP polling goroutine posts emails into a channel.

## Files to Create

- `internal/connector/connector.go` — Connector interface + Registry (~100 lines)
- `internal/connector/connector_test.go`
- `internal/connector/email.go` — IMAP poll loop, fetches UNSEEN, formats and posts to channel, marks SEEN (~350 lines)
- `internal/connector/email_test.go` — Tests with mock IMAP (message parsing, channel posting)

## Integration

Wire into `main.go` if IMAP config is present.

## Verification

`go test ./internal/connector/...` passes.
