# 006 — Add emoji reactions to messages

## Summary

Users and agents should be able to react to messages with emoji. This is a common chat feature that enables lightweight feedback without requiring a full reply — useful for acknowledgements, voting, sentiment, etc.

## Requirements

### Core

- Any user or agent can add an emoji reaction to any message in a channel they belong to
- Multiple distinct emoji reactions per message (e.g. both a thumbs-up and a heart)
- Multiple users/agents can react with the same emoji — show a count
- A user/agent can toggle their own reaction off (unreact)
- Reactions are visible in real time via WebSocket to all channel participants

### Data model

New table (follows existing patterns — `int64` PKs, `author_id`/`author_type` like `messages`):

```sql
CREATE TABLE reactions (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  message_id  INTEGER NOT NULL REFERENCES messages(id),
  author_id   INTEGER NOT NULL,
  author_type TEXT    NOT NULL,  -- "human" or "agent" (same as messages.author_type)
  emoji       TEXT    NOT NULL,  -- unicode emoji
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(message_id, author_id, author_type, emoji)
);
```

This mirrors the `messages` table convention where `author_id` + `author_type` identify who did it, rather than separate nullable FK columns.

### API

Routes nested under the existing `/api/channels/{id}/messages` group. Emoji is passed in the JSON body to avoid URL-encoding issues with unicode.

| Method | Path                                                | Body               | Description                     |
| ------ | --------------------------------------------------- | ------------------ | ------------------------------- |
| PUT    | `/api/channels/{id}/messages/{messageID}/reactions` | `{"emoji": "..."}` | Toggle reaction on (idempotent) |
| DELETE | `/api/channels/{id}/messages/{messageID}/reactions` | `{"emoji": "..."}` | Toggle reaction off             |

Path params extracted with `chi.URLParam(r, "id")` and `chi.URLParam(r, "messageID")`, parsed via `strconv.ParseInt` — same as existing handlers.

Reactions should be included when fetching messages via `GET /api/channels/{id}/messages`. Add a `reactions` field to the message JSON response:

```json
{
  "id": 42,
  "channel_id": 5,
  "author_id": 2,
  "author_type": "human",
  "author_name": "alice",
  "content": "Hello",
  "created_at": "2025-01-30T12:34:56Z",
  "reactions": [
    { "emoji": "\ud83d\udc4d", "count": 3, "reacted": true },
    { "emoji": "\u2764\ufe0f", "count": 1, "reacted": false }
  ]
}
```

`reacted` indicates whether the requesting user has this reaction. Handler struct follows existing pattern:

```go
type ReactionHandler struct {
    DB  *db.DB
    Hub *ws.Hub
}
```

### WebSocket events

Follow existing naming convention (`new_message` style):

- `new_reaction` — broadcast when a reaction is added
- `remove_reaction` — broadcast when a reaction is removed

Payload:

```json
{
  "message_id": 42,
  "channel_id": 5,
  "emoji": "\ud83d\udc4d",
  "author_id": 2,
  "author_type": "human",
  "counts": [
    { "emoji": "\ud83d\udc4d", "count": 3 },
    { "emoji": "\u2764\ufe0f", "count": 1 }
  ]
}
```

Broadcast via `h.Hub.Broadcast(ws.Event{Type: "new_reaction", Data: ...})` — same as `new_message`.

### Frontend

- Show reaction pills below each message (emoji + count)
- Click an existing reaction pill to toggle your own reaction on/off
- Emoji picker (can be simple — a small curated set is fine for v1, full picker later)
- Agents' reactions should be visually distinguishable (e.g. subtle bot icon overlay)

### Agent integration

- Agents can react via tool calls (new `message_react` tool) or directly through the API
- Useful for agents to signal understanding, flag issues, etc. without cluttering the channel with messages

## Files likely involved

| Area                | Files                      |
| ------------------- | -------------------------- |
| DB migration        | `internal/db/migrations/`  |
| Model               | `internal/db/models/`      |
| Store               | `internal/db/`             |
| API handlers        | `internal/api/`            |
| WebSocket hub       | `internal/ws/`             |
| Frontend components | `frontend/src/components/` |
| Agent tools         | `internal/tools/`          |

## Out of scope (for now)

- Custom emoji / image reactions
- Reaction-based workflows or triggers
- Reaction permissions (e.g. restricting which emoji are allowed)
