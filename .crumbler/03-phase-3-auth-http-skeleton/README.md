# Phase 3: Auth + HTTP Skeleton

**Goal:** Server starts, auth works (register/login/logout/me), with tests.

## Files to Create

- `internal/auth/auth.go` — bcrypt hash/check, token gen (crypto/rand), invite code gen (~180 lines)
- `internal/auth/auth_test.go`
- `internal/auth/middleware.go` — Chi middleware: reads `Authorization: Bearer <token>` header or `session` cookie (SameSite=Lax, HttpOnly, Secure in production). Sets user on context. (~150 lines)
- `internal/auth/middleware_test.go`
- `internal/api/helpers.go` — WriteJSON, ReadJSON, ErrorResponse, GetUser, input validation helpers (~130 lines)
- `internal/api/helpers_test.go`
- `internal/api/router.go` — Chi router, CORS, rate limiting middleware, route groups (~120 lines)
- `internal/api/auth_handlers.go` — Register (open if 0 users, else requires valid invite_code), login (returns token + sets cookie), logout, me (~280 lines)
- `internal/api/auth_handlers_test.go` — httptest-based handler tests

## Integration

Update `cmd/waynebot/main.go` to wire HTTP server with graceful shutdown.

## Input Validation

- `username`: 1–50 chars, alphanumeric + underscore only
- `password`: 8–128 chars

## Verification

`go test ./internal/auth/... ./internal/api/...` passes. Manual curl smoke test against running server.
