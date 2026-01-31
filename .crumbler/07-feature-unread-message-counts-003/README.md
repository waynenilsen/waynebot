# 003: Add unread message counts per channel

## Feature

When another user (or agent) sends a message in a channel, the current user should see an unread count badge next to that channel in the sidebar. This is table-stakes for any chat app — without it you have no idea where new activity is.

## Current behavior

No unread tracking exists. The channel list in the sidebar shows channel names with no indication of new messages.

## Desired behavior

- Each channel in the sidebar shows a badge with the number of unread messages
- Viewing a channel marks it as read (resets the count to 0)
- The count updates in real-time via WebSocket as new messages arrive
- Bold channel name or visual highlight when unread count > 0

## Implementation sketch

### Backend

- New table `read_positions` with columns `(user_id, channel_id, last_read_message_id, updated_at)`
- New endpoint `POST /api/channels/{id}/read` — updates the user's read position to the latest message
- New endpoint or extend `GET /api/channels` — return `unread_count` per channel (count of messages with ID > last_read_message_id)
- Optionally broadcast `read_position_updated` via WebSocket so other tabs stay in sync

### Frontend

- Call mark-as-read when the user views/focuses a channel
- Display unread badge in the sidebar channel list
- Update counts in real-time from `new_message` WebSocket events (increment for non-active channels)

## Files likely involved

| Area | Files |
|------|-------|
| DB migration | `internal/db/migrations.go` |
| Backend API | `internal/api/channel_handlers.go`, `internal/api/router.go` |
| Backend model | `internal/model/` (new read_position model) |
| Frontend sidebar | `frontend/src/components/Sidebar.tsx` or equivalent |
| Frontend state | `frontend/src/store/AppContext.tsx` |
| Frontend WS | `frontend/src/hooks/useWebSocket.ts` |
