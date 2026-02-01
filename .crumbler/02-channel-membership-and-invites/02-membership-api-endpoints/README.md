# Membership API Endpoints

## What

Expose endpoints for managing channel membership (both humans and agents).

## Endpoints

### `GET /api/channels/{id}/members`
Returns list of members (both human and agent) for a channel.

Response:
```json
[
  {"type": "user", "id": 1, "name": "wayne", "role": "owner"},
  {"type": "persona", "id": 3, "name": "helper-bot", "role": "member"}
]
```

### `POST /api/channels/{id}/members`
Add a member to a channel. Body:
```json
{"user_id": 2}
// or
{"persona_id": 3}
```
- If `persona_id` is provided, calls `SubscribeChannel` (existing function).
- If `user_id` is provided, calls `AddChannelMember` (new function from step 1).
- Returns 201 on success, 409 if already a member.

### `DELETE /api/channels/{id}/members`
Remove a member from a channel. Body:
```json
{"user_id": 2}
// or
{"persona_id": 3}
```

## Implementation

- Add a new `MemberHandler` in `internal/api/member_handlers.go`
- Wire routes in `router.go`
- For the GET endpoint, query both `channel_members` and `persona_channels` to return a unified list.
