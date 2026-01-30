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

```
go run ./cmd/waynebot
```
