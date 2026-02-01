# DM Backend: API Endpoints

Add API endpoints for creating and listing DMs.

## Endpoints

### `POST /api/dms`
Create or retrieve a DM channel.

Request body:
```json
{
  "user_id": 5,       // set one of these
  "persona_id": 3     // set one of these
}
```

The authenticated user is always one side of the DM. The other side is specified by either `user_id` or `persona_id`.

Logic:
1. Validate that exactly one of `user_id` or `persona_id` is set
2. Validate the target exists
3. Check if a DM already exists between these two participants (use `FindDMChannel`)
4. If exists, return it. If not, create it with `CreateDMChannel`
5. When creating a DM with a persona, auto-subscribe the persona to the channel and notify the supervisor to restart that actor so it picks up the new subscription

Response: the channel object (same shape as regular channels, with `is_dm: true` and participant info)

### `GET /api/dms`
List all DMs for the authenticated user.

Response: array of DM channel objects, each including:
- Channel info (id, name, created_at)
- The other participant (user or persona name + type)
- Unread count (reuse read_positions logic)

### Messages

DMs reuse the existing message endpoints: `GET/POST /api/channels/{id}/messages`. No changes needed — just ensure the channel ID works regardless of `is_dm` flag.

## Router Changes

In `internal/api/router.go`, add the DM routes under the authenticated group:
```
r.Get("/api/dms", handler.ListDMs)
r.Post("/api/dms", handler.CreateDM)
```

## WebSocket

When a DM message is posted, the existing message broadcast works. No special handling needed since all connected clients get all messages and filter client-side.
