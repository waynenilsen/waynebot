# waynebot

A personal multi-agent system. Agents with different personas collaborate in a built-in chat workspace (Slack clone), and you participate alongside them as a peer.

## Architecture

- **Go HTTP server daemon** backed by SQLite
- **Multi-agent processing loop** — each agent (actor) runs in parallel with an interruptible outer loop
- **Built-in chat UI** — a Slack-like web interface where agents and you converse in shared channels
- **Connectors** — external sources (email, notifications, etc.) pipe into chat channels so agents can see and act on them

## Auth

Invite-only with username/password authentication. Registration requires an invite from an existing user.

## Getting Started

You need two terminals — one for the backend, one for the frontend.

### Backend

```
go run ./cmd/waynebot
```

Starts the Go server on port **59731**. Configure with environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `WAYNEBOT_PORT` | 59731 | HTTP server port |
| `WAYNEBOT_DB_PATH` | waynebot.db | SQLite database path |
| `WAYNEBOT_CORS_ORIGINS` | http://localhost:5173 | Allowed CORS origins |
| `WAYNEBOT_OPENROUTER_KEY` | | LLM API key (OpenRouter) |

### Frontend

```
cd frontend
npm install
npm run dev
```

Starts the Vite dev server on port **53461**. The dev server proxies `/api/*` and `/ws` requests to the backend automatically, so everything works from a single URL.

### Open the app

Visit **http://localhost:53461**. The first user can register without an invite code (bootstrap mode). After that, registration requires an invite from an existing user.
