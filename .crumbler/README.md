# Waynebot — Waterfall Implementation Plan

## Architecture Overview

Go HTTP server (port 59731) + SQLite backend, React/Vite frontend (port 53461), multi-agent processing loop. OpenRouter for LLM via `github.com/openai/openai-go` (base URL override). No file exceeds 3k lines.

All backend packages must have tests (`*_test.go`). Frontend components and hooks should be tested where reasonably possible using Vitest + React Testing Library. Frontend UI must be built using the `/frontend-design` skill for high design quality.

---

## Directory Layout

```
waynebot/
  cmd/waynebot/
    main.go                           # Entrypoint (~120 lines)
  internal/
    config/config.go                  # Env-based config (~140 lines)
    db/
      db.go                           # Open, WAL, pragmas, single-writer mutex (~130 lines)
      db_test.go                      # DB open/close, WAL mode verification
      migrations.go                   # Versioned migrations (~300 lines)
      migrations_test.go              # Migration idempotency, version tracking
    model/
      user.go                         # User struct + queries (~200 lines)
      user_test.go
      session.go                      # Session struct + queries (~180 lines)
      session_test.go
      invite.go                       # Invite struct + queries, atomic claim (~180 lines)
      invite_test.go
      channel.go                      # Channel struct + queries (~200 lines)
      channel_test.go
      message.go                      # Message struct + queries (~250 lines)
      message_test.go
      persona.go                      # Persona struct + queries (~250 lines)
      persona_test.go
    auth/
      auth.go                         # Bcrypt, token gen, CSRF token gen (~180 lines)
      auth_test.go
      middleware.go                    # Chi auth middleware, SameSite cookies (~150 lines)
      middleware_test.go
    api/
      router.go                       # Route assembly (~120 lines)
      helpers.go                      # JSON helpers, input validation (~130 lines)
      helpers_test.go
      auth_handlers.go                # Register/login/logout/me (~280 lines)
      auth_handlers_test.go
      channel_handlers.go             # Channel CRUD + messages (~300 lines)
      channel_handlers_test.go
      persona_handlers.go             # Persona CRUD (~220 lines)
      persona_handlers_test.go
      agent_handlers.go               # Agent status/start/stop (~200 lines)
      agent_handlers_test.go
      invite_handlers.go              # Invite create/list (~180 lines)
      invite_handlers_test.go
      ws_handler.go                   # WebSocket upgrade via ticket (~180 lines)
      ws_handler_test.go
    ws/
      hub.go                          # Broadcast hub + notify channel for agents (~280 lines)
      hub_test.go
      client.go                       # Per-connection r/w (~200 lines)
      client_test.go
    agent/
      actor.go                        # Single actor outer loop (~450 lines)
      actor_test.go
      status.go                       # Status enum + tracker (~120 lines)
      status_test.go
      decision.go                     # Should-respond logic (~200 lines)
      decision_test.go
      cursor.go                       # Per-actor message cursor (~120 lines)
      cursor_test.go
      supervisor.go                   # Start/stop/monitor actors (~350 lines)
      supervisor_test.go
      budget.go                       # Per-persona token/cost budget enforcement (~150 lines)
      budget_test.go
    llm/
      client.go                       # OpenRouter via openai-go (~250 lines)
      client_test.go                  # Tests with mock HTTP server
      messages.go                     # Domain→SDK message conversion (~200 lines)
      messages_test.go
      tools.go                        # Tool JSON schema defs (~150 lines)
      tools_test.go
    tools/
      registry.go                     # Tool dispatch (~120 lines)
      registry_test.go
      sandbox.go                      # Sandbox config: allowed dirs, blocked commands (~150 lines)
      sandbox_test.go
      shell_exec.go                   # Sandboxed shell command (~200 lines)
      shell_exec_test.go
      file_read.go                    # Path-restricted file read (~120 lines)
      file_read_test.go
      file_write.go                   # Path-restricted file write (~120 lines)
      file_write_test.go
      http_fetch.go                   # HTTP fetch with denied hosts (~130 lines)
      http_fetch_test.go
    connector/
      connector.go                    # Interface + registry (~100 lines)
      connector_test.go
      email.go                        # IMAP polling (~350 lines)
      email_test.go
  frontend/
    index.html
    package.json
    vite.config.ts
    vitest.config.ts
    tsconfig.json
    tailwind.config.ts
    postcss.config.js
    src/
      main.tsx                        # Entry (~10 lines)
      App.tsx                         # Router + auth gate (~50 lines)
      App.test.tsx
      types.ts                        # All TS interfaces (~100 lines)
      api.ts                          # HTTP fetch wrappers (~120 lines)
      api.test.ts
      ws.ts                           # WebSocket manager (~70 lines)
      ws.test.ts
      store/
        AppContext.tsx                 # Global state context (~80 lines)
        AppContext.test.tsx
      hooks/
        useAuth.ts                    # Auth state (~60 lines)
        useAuth.test.ts
        useChannels.ts                # Channel list (~50 lines)
        useChannels.test.ts
        useMessages.ts                # Messages + send (~60 lines)
        useMessages.test.ts
        usePersonas.ts                # Persona CRUD (~60 lines)
        useAgents.ts                  # Agent status (~50 lines)
        useInvites.ts                 # Invite codes (~40 lines)
        useWebSocket.ts               # WS connection (~50 lines)
      components/
        Layout.tsx                    # Sidebar + main (~40 lines)
        Sidebar.tsx                   # Nav + channels (~70 lines)
        ChannelList.tsx               # Channel items (~70 lines)
        MessageThread.tsx             # Message list + compose (~100 lines)
        MessageThread.test.tsx
        MessageItem.tsx               # Single message (~80 lines)
        MessageItem.test.tsx
        MessageCompose.tsx            # Text input (~60 lines)
        MessageCompose.test.tsx
        MarkdownRenderer.tsx          # react-markdown wrap (~30 lines)
        LoginPage.tsx                 # Login/register form (~110 lines)
        LoginPage.test.tsx
        PersonaPage.tsx               # Persona list + form (~100 lines)
        PersonaForm.tsx               # Persona fields (~110 lines)
        AgentDashboard.tsx            # Agent cards (~100 lines)
        InvitePage.tsx                # Invite list (~70 lines)
        ErrorBanner.tsx               # Global error/offline banner (~40 lines)
      utils/
        markdown.ts                   # remark config (~15 lines)
        dateFormat.ts                 # Timestamp helpers (~20 lines)
```

---

## Unified Data Models

### SQLite Schema (in `internal/db/migrations.go`, versioned)

Migrations are tracked in a `schema_version` table. Each migration has a version number and is applied exactly once. New migrations are appended as functions; existing ones are never modified.

```sql
-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- v1: Initial schema

-- Users
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Sessions (SameSite=Lax cookies, Bearer token header)
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL
);

-- Short-lived WebSocket tickets (replace token-in-URL)
CREATE TABLE ws_tickets (
    ticket TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL  -- 30 seconds from creation
);

-- Invites (used_by claim is atomic via UPDATE ... WHERE used_by IS NULL)
CREATE TABLE invites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT UNIQUE NOT NULL,
    created_by INTEGER NOT NULL REFERENCES users(id),
    used_by INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Channels
CREATE TABLE channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Messages (unified: human and agent messages in one table)
-- NOTE: author_id refers to users.id when author_type='human',
-- or personas.id when author_type='agent'. Always filter on
-- (author_id, author_type) together — never author_id alone.
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id INTEGER NOT NULL REFERENCES channels(id),
    author_id INTEGER NOT NULL,
    author_type TEXT NOT NULL CHECK(author_type IN ('human', 'agent')),
    author_name TEXT NOT NULL,         -- denormalized for display (intentional: preserves name at time of posting)
    content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_messages_channel_id ON messages(channel_id, id);

-- Personas
CREATE TABLE personas (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    system_prompt TEXT NOT NULL,
    model TEXT NOT NULL DEFAULT 'anthropic/claude-sonnet-4-20250514',
    tools_enabled TEXT NOT NULL DEFAULT '["shell_exec","file_read","file_write","http_fetch"]', -- JSON array
    temperature REAL NOT NULL DEFAULT 0.7,
    max_tokens INTEGER NOT NULL DEFAULT 4096,
    cooldown_secs INTEGER NOT NULL DEFAULT 30,
    max_tokens_per_hour INTEGER NOT NULL DEFAULT 100000,  -- token budget (0 = unlimited)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Persona-channel subscriptions
CREATE TABLE persona_channels (
    persona_id INTEGER NOT NULL REFERENCES personas(id) ON DELETE CASCADE,
    channel_id INTEGER NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    PRIMARY KEY (persona_id, channel_id)
);

-- Actor cursors (tracks last-seen message per actor per channel)
CREATE TABLE actor_cursors (
    persona_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    last_seen_message_id INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(persona_id, channel_id)
);

-- LLM call log (also used for budget enforcement)
CREATE TABLE llm_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    persona_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    model TEXT NOT NULL,
    messages_json TEXT NOT NULL,
    response_json TEXT NOT NULL,
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_llm_calls_persona_time ON llm_calls(persona_id, created_at);

-- Tool execution log
CREATE TABLE tool_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    persona_id INTEGER NOT NULL,
    tool_name TEXT NOT NULL,
    args_json TEXT NOT NULL,
    output_text TEXT,
    error_text TEXT,
    duration_ms INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Go Structs (canonical)

```go
// model/message.go
type Message struct {
    ID         int64
    ChannelID  int64
    AuthorID   int64
    AuthorType string // "human" or "agent"
    AuthorName string
    Content    string
    CreatedAt  time.Time
}

// model/persona.go
type Persona struct {
    ID               int64
    Name             string
    SystemPrompt     string
    Model            string
    ToolsEnabled     []string // JSON-encoded in DB
    Temperature      float64
    MaxTokens        int
    CooldownSecs     int
    MaxTokensPerHour int
    CreatedAt        time.Time
}
```

### TypeScript Interfaces (canonical)

```typescript
interface Message {
  id: number;
  channelId: number;
  authorId: number;
  authorType: "human" | "agent";
  authorName: string;
  content: string;
  createdAt: string;
}

interface Persona {
  id: number;
  name: string;
  systemPrompt: string;
  model: string;
  toolsEnabled: string[];
  temperature: number;
  maxTokens: number;
  cooldownSecs: number;
  maxTokensPerHour: number;
  createdAt: string;
}
```

---

## Unified API Endpoints

```
POST   /api/register                 # Open when 0 users, else requires invite_code
POST   /api/login                    # Returns { token, user }, sets SameSite=Lax cookie
POST   /api/logout                   # Deletes session
GET    /api/me                       # Returns current user (for token validation)

GET    /api/channels                 # List all channels
POST   /api/channels                 # Create { name, description }
GET    /api/channels/:id/messages    # Paginated: ?limit=50&before=<messageId>
POST   /api/channels/:id/messages    # Send { content }

GET    /api/personas                 # List all
POST   /api/personas                 # Create
PUT    /api/personas/:id             # Update
DELETE /api/personas/:id             # Delete

GET    /api/agents/status            # Map of persona_id -> { status, detail, updatedAt, tokensUsedThisHour }
POST   /api/agents/:id/start         # Start actor for persona
POST   /api/agents/:id/stop          # Stop actor

POST   /api/invites                  # Generate invite code
GET    /api/invites                  # List invites

POST   /api/ws/ticket                # Exchange session token for short-lived WS ticket
GET    /ws?ticket=...                # WebSocket (real-time messages, agent status)
GET    /health                       # Unauthenticated health check
```

---

## Go Dependencies

```
github.com/go-chi/chi/v5            # HTTP router
github.com/go-chi/cors              # CORS middleware
github.com/go-chi/httprate          # Rate limiting
github.com/coder/websocket          # WebSocket
github.com/openai/openai-go         # LLM client (OpenRouter via base URL override)
github.com/emersion/go-imap/v2      # IMAP connector
golang.org/x/crypto                 # bcrypt
modernc.org/sqlite                  # Pure-Go SQLite
```

---

## Cross-Cutting Concerns

### SQLite Write Contention

All database writes go through a single `*sync.Mutex` (the "write mutex") in `db.go`. This serializes writes and avoids `SQLITE_BUSY` errors from concurrent goroutines (HTTP handlers, agent actors, connectors). Reads are not serialized. The connection sets `PRAGMA busy_timeout = 5000` as a safety net. The write mutex is the primary contention control.

### Input Validation

All API handlers validate input lengths before processing:
- `username`: 1–50 chars, alphanumeric + underscore only
- `password`: 8–128 chars
- `channel.name`: 1–100 chars
- `channel.description`: 0–500 chars
- `message.content`: 1–10000 chars
- `persona.name`: 1–100 chars
- `persona.system_prompt`: 1–50000 chars

Validation helpers live in `api/helpers.go`.

### Tool Sandboxing

Agent tools are restricted to prevent damage:

- **shell_exec**: Configurable allowed-command allowlist (default: common read-only commands like `ls`, `cat`, `grep`, `curl`, `python`, `node`). Commands not on the list are rejected. Execution timeout of 30 seconds. Output truncated to 10KB.
- **file_read**: Restricted to a configurable base directory (env `WAYNEBOT_SANDBOX_DIR`, default: `./sandbox`). Path traversal (`..`) rejected. Max file size 1MB.
- **file_write**: Same directory restriction as file_read. Max file size 1MB.
- **http_fetch**: Blocked hosts list (default: `localhost`, `127.0.0.1`, `169.254.169.254`, `[::1]`) to prevent SSRF. Timeout 15 seconds. Response body truncated to 50KB.

### Agent Anti-Loop Strategy

Agent-to-agent ping-pong is controlled via **prompt engineering**: each persona's system prompt instructs it to use emoji reactions (e.g., thumbs-up) to acknowledge messages rather than generating verbose replies, and to avoid responding to pure acknowledgment messages. The system enforces:

1. Skip if all new messages since cursor are from self
2. Per-channel cooldown timer (configurable per persona, default 30s)
3. Max 10 tool-call rounds per LLM invocation
4. Token budget per persona per hour (enforced before each LLM call)

### WebSocket Auth

Instead of passing the session token in the WebSocket URL query string (which leaks into logs), the client first calls `POST /api/ws/ticket` to exchange its session token for a single-use, 30-second ticket. The ticket is then passed as `GET /ws?ticket=...`. The ticket is deleted on use or expiration.

### Author ID Disambiguation

`messages.author_id` refers to `users.id` when `author_type='human'` and `personas.id` when `author_type='agent'`. **All queries involving author_id must filter on author_type as well.** This is documented in the schema and enforced by code review. The tradeoff (vs. a unified actors table) is simplicity — users and personas have very different schemas and lifecycles.

### Denormalized Author Name

`messages.author_name` is intentionally denormalized. It preserves the display name at the time of posting. If a user or persona is renamed, historical messages keep their original name. This is the desired behavior for chat history integrity.

---

## Waterfall Phases

### Phase 1: Foundation

**Goal:** Binary starts, opens SQLite, runs versioned migrations, exits on signal.

**Note:** Backend HTTP server must run on port 59731.

Files:

- `cmd/waynebot/main.go`
- `internal/config/config.go` — Env-based config: port (59731), DB path, sandbox dir, CORS origins, IMAP settings, OpenRouter key. All ports and paths configurable via env vars.
- `internal/db/db.go` — Open DB, WAL mode, `busy_timeout=5000`, write mutex
- `internal/db/db_test.go`
- `internal/db/migrations.go` — Versioned migrations with `schema_version` table
- `internal/db/migrations_test.go`

Verify: `go run ./cmd/waynebot` creates `waynebot.db` with all tables. `go test ./internal/db/...` passes.

---

### Phase 2: Models Layer

**Goal:** All Go structs and query functions, fully tested.

Files:

- `internal/model/user.go` — CRUD + `CountUsers()` for bootstrap check
- `internal/model/user_test.go`
- `internal/model/session.go` — Create, get by token, delete, cleanup expired
- `internal/model/session_test.go`
- `internal/model/invite.go` — Create, `ClaimInvite(code, userID)` using `UPDATE invites SET used_by=? WHERE code=? AND used_by IS NULL` (atomic, returns rows-affected to detect race), list
- `internal/model/invite_test.go` — Includes concurrent claim test
- `internal/model/channel.go` — CRUD, list
- `internal/model/channel_test.go`
- `internal/model/message.go` — Create, list paginated (cursor-based: `?before=messageId`), `GetMessagesSince(channelID, afterID)`, `GetRecentMessages(channelID, limit)`
- `internal/model/message_test.go`
- `internal/model/persona.go` — CRUD, subscribe/unsubscribe channels, get subscribed channels
- `internal/model/persona_test.go`

All model tests run against `:memory:` SQLite with migrations applied.

Verify: `go test ./internal/model/...` passes.

---

### Phase 3: Auth + HTTP Skeleton

**Goal:** Server starts, auth works (register/login/logout/me), with tests.

Files:

- `internal/auth/auth.go` — bcrypt hash/check, token gen (crypto/rand), invite code gen
- `internal/auth/auth_test.go`
- `internal/auth/middleware.go` — Chi middleware: reads `Authorization: Bearer <token>` header or `session` cookie (SameSite=Lax, HttpOnly, Secure in production). Sets user on context.
- `internal/auth/middleware_test.go`
- `internal/api/helpers.go` — WriteJSON, ReadJSON, ErrorResponse, GetUser, input validation helpers
- `internal/api/helpers_test.go`
- `internal/api/router.go` — Chi router, CORS, rate limiting middleware, route groups
- `internal/api/auth_handlers.go` — Register (open if 0 users, else requires valid invite_code), login (returns token + sets cookie), logout, me
- `internal/api/auth_handlers_test.go` — httptest-based handler tests

Update `main.go` to wire HTTP server with graceful shutdown.

Verify: `go test ./internal/auth/... ./internal/api/...` passes. Manual curl smoke test against running server.

---

### Phase 4: Channel + Message + Persona + Invite APIs

**Goal:** Full REST CRUD, with tests.

Files:

- `internal/api/channel_handlers.go` — List, create, get messages (cursor-based pagination), post message
- `internal/api/channel_handlers_test.go`
- `internal/api/persona_handlers.go` — List, create, update, delete (cascades subscriptions)
- `internal/api/persona_handlers_test.go`
- `internal/api/invite_handlers.go` — Generate, list
- `internal/api/invite_handlers_test.go`

Verify: `go test ./internal/api/...` passes.

---

### Phase 5: WebSocket Hub

**Goal:** Real-time message broadcast to connected clients, with notification channel for agents.

Files:

- `internal/ws/hub.go` — Hub with register/unregister/broadcast. Exposes `NotifyChan` that agents can select on to wake immediately when a new message is posted (replaces pure polling).
- `internal/ws/hub_test.go`
- `internal/ws/client.go` — Per-connection read/write pumps with ping/pong keepalive
- `internal/ws/client_test.go`
- `internal/api/ws_handler.go` — `POST /api/ws/ticket` to mint ticket, `GET /ws?ticket=...` to upgrade. Ticket is single-use, expires in 30s.
- `internal/api/ws_handler_test.go`

Integration: `PostMessage` handler broadcasts via hub after insert AND sends on `NotifyChan`. Start hub goroutine in `main.go`.

Verify: `go test ./internal/ws/... ./internal/api/...` passes.

---

### Phase 6: Agent Processing Loop

**Goal:** Personas run as background goroutines, call OpenRouter, use sandboxed tools, post responses.

Files:

- `internal/agent/status.go` — Status enum (`idle`, `thinking`, `tool_call`, `error`, `stopped`, `budget_exceeded`), goroutine-safe tracker
- `internal/agent/status_test.go`
- `internal/agent/cursor.go` — CursorStore (SQLite-backed last_seen_message_id per persona per channel)
- `internal/agent/cursor_test.go`
- `internal/agent/decision.go` — `shouldRespond()`: skip self-messages, check cooldown, check @mentions
- `internal/agent/decision_test.go`
- `internal/agent/budget.go` — `BudgetChecker`: queries `llm_calls` for total tokens used by persona in last hour, compares to `max_tokens_per_hour`. Rejects call if over budget, sets actor status to `budget_exceeded`.
- `internal/agent/budget_test.go`
- `internal/agent/actor.go` — Actor struct, `Run(ctx)` loop: `select` on `ctx.Done()`, `hub.NotifyChan`, or fallback ticker → check cursor → decide → build history → check budget → LLM loop → post response → broadcast. Interruptible via context cancellation.
- `internal/agent/actor_test.go` — Tests with mock LLM client
- `internal/agent/supervisor.go` — Start/stop/monitor actors as goroutines with context cancellation. `StopAll()` for shutdown.
- `internal/agent/supervisor_test.go`
- `internal/llm/client.go` — `openai-go` client with OpenRouter base URL. `ChatCompletion(ctx, model, messages, tools, temp, maxTokens)`. Returns token counts from response.
- `internal/llm/client_test.go` — Tests with mock HTTP server returning canned OpenAI-format responses
- `internal/llm/messages.go` — Convert domain `Message` to SDK message types. Human messages prefixed with author name.
- `internal/llm/messages_test.go`
- `internal/llm/tools.go` — Tool JSON schema definitions (shell_exec, file_read, file_write, http_fetch).
- `internal/llm/tools_test.go`
- `internal/tools/sandbox.go` — `SandboxConfig`: allowed commands list, allowed base directory, blocked hosts. Loaded from config/env.
- `internal/tools/sandbox_test.go`
- `internal/tools/registry.go` — `Registry` maps tool names to `ToolFunc(ctx, args) (string, error)`.
- `internal/tools/registry_test.go`
- `internal/tools/shell_exec.go` — Validates command against allowlist, `exec.CommandContext` with 30s timeout, captures stdout/stderr, truncates output to 10KB.
- `internal/tools/shell_exec_test.go` — Tests allowlist enforcement, timeout, output truncation
- `internal/tools/file_read.go` — Validates path within sandbox dir (no `..` traversal), `os.ReadFile` with 1MB limit.
- `internal/tools/file_read_test.go` — Tests path traversal rejection, size limit
- `internal/tools/file_write.go` — Same path validation as file_read. `os.WriteFile` with 1MB limit.
- `internal/tools/file_write_test.go`
- `internal/tools/http_fetch.go` — Validates host not in blocked list, `http.NewRequestWithContext` with 15s timeout, body truncated to 50KB.
- `internal/tools/http_fetch_test.go` — Tests blocked host rejection
- `internal/api/agent_handlers.go` — GET status (includes tokensUsedThisHour), POST start, POST stop.
- `internal/api/agent_handlers_test.go`

Anti-loop: (1) skip if all new msgs from self, (2) per-channel cooldown timer, (3) max 10 tool rounds, (4) token budget per hour. Agent-to-agent ping-pong is additionally mitigated via prompt engineering — personas are instructed to use emoji reactions for acknowledgment and avoid responding to pure-ack messages.

LLM tool loop: call LLM → if tool_calls, execute tools (respecting sandbox), append results, re-call → repeat until text response or max rounds.

Wire supervisor into `main.go`, stop all agents on shutdown.

Verify: `go test ./internal/agent/... ./internal/llm/... ./internal/tools/...` passes.

---

### Phase 7: Frontend

**Goal:** Working Slack-clone UI with high design quality. **All UI components must be built using the `/frontend-design` skill.**

**Phase 7a: Scaffolding** — Vite + React + TS + Tailwind project under `frontend/`. Dev server runs on port 53461 and proxies `/api` and `/ws` to Go backend (port 59731). Install Vitest + React Testing Library + jsdom for testing. Configure `vitest.config.ts`.

**Phase 7b: Types + API + WebSocket + Store** — `types.ts`, `api.ts` (with error handling that surfaces errors to UI), `ws.ts` (with auto-reconnect + exponential backoff), `store/AppContext.tsx` (React context providing shared state: auth, channels, messages, agents — WebSocket events dispatch into this context so all hooks stay in sync), `utils/`. Tests for `api.ts`, `ws.ts`, `AppContext.tsx`.

**Phase 7c: Auth** — `useAuth.ts`, `LoginPage.tsx` (use `/frontend-design`), auth gate in `App.tsx`. Tests for `useAuth.ts`, `LoginPage.tsx`, `App.tsx`.

**Phase 7d: Layout + Channels** — `Layout.tsx`, `Sidebar.tsx`, `ChannelList.tsx`, `useChannels.ts`, `useWebSocket.ts` (use `/frontend-design`). Routes wired in `App.tsx`. Tests for `useChannels.ts`.

**Phase 7e: Messages** — `MessageThread.tsx`, `MessageItem.tsx`, `MessageCompose.tsx`, `MarkdownRenderer.tsx`, `useMessages.ts` (use `/frontend-design`). Human vs agent visual treatment (agent gets colored left border + bot badge). Tests for `MessageThread.tsx`, `MessageItem.tsx`, `MessageCompose.tsx`, `useMessages.ts`.

**Phase 7f: Persona + Agent + Invite pages** — `PersonaPage.tsx`, `PersonaForm.tsx`, `AgentDashboard.tsx` (shows token budget usage), `InvitePage.tsx` + hooks (use `/frontend-design`).

**Phase 7g: Polish** — `ErrorBanner.tsx` (global banner for API errors / offline state), empty states, loading spinners, WebSocket reconnection indicator, keyboard shortcuts (use `/frontend-design`).

Verify: `npx vitest run` passes. Full flow in browser — login, create channel, send message, start agent, see agent reply in real time.

---

### Phase 8: Email Connector

**Goal:** IMAP polling goroutine posts emails into a channel.

Files:

- `internal/connector/connector.go` — `Connector` interface + `Registry`
- `internal/connector/connector_test.go`
- `internal/connector/email.go` — IMAP poll loop, fetches UNSEEN, formats and posts to channel, marks SEEN
- `internal/connector/email_test.go` — Tests with mock IMAP (message parsing, channel posting)

Wire into `main.go` if IMAP config is present.

Verify: `go test ./internal/connector/...` passes.

---

### Phase 9: Hardening

- Session cleanup goroutine (every 15 min, also cleans expired ws_tickets)
- Rate limiting (60/min auth, 120/min other)
- Structured logging (`slog.JSONHandler`)
- Graceful shutdown sequence: HTTP → connectors → agents → WS hub → DB
- Health endpoint
- CORS config via env var
- LLM call + tool execution logging to `llm_calls` / `tool_executions` tables

---

## Verification Plan

1. **Phase 1-2:** `go build ./cmd/waynebot` compiles. `go test ./internal/db/... ./internal/model/...` passes.
2. **Phase 3:** `go test ./internal/auth/... ./internal/api/...` passes. Manual curl smoke test.
3. **Phase 4:** `go test ./internal/api/...` passes.
4. **Phase 5:** `go test ./internal/ws/...` passes.
5. **Phase 6:** `go test ./internal/agent/... ./internal/llm/... ./internal/tools/...` passes.
6. **Phase 7:** `npx vitest run` passes. Browser smoke test of full UI flow.
7. **Phase 8:** `go test ./internal/connector/...` passes.
8. **Phase 9:** Shutdown gracefully, rate limits work, logs are structured JSON.
9. **Full suite:** `go test ./...` from repo root passes. `cd frontend && npx vitest run` passes.
