# Realtime Agent Responses

## Problem
User has to refresh the page to see agent responses in the channel. They should appear in real-time via WebSocket.

## Investigation
The WebSocket hub broadcasts `new_message` events and the frontend listens via `useWebSocket`. The agent posts messages via `model.CreateMessage()` but we need to verify the agent actor actually broadcasts the message through the hub after posting.

## What to fix
1. Check `internal/agent/actor.go` - after the agent posts a response message, ensure it broadcasts via `hub.Broadcast()` with a `new_message` event containing the full message data
2. Check the frontend `useWebSocket` hook and `useMessages` hook - ensure `new_message` events from agents are properly added to the message list for the current channel
3. Verify the message format matches what the frontend expects (id, channel_id, author_id, author_type, author_name, content, created_at, reactions)

## Key files
- `internal/agent/actor.go` - where agent posts responses
- `internal/ws/hub.go` - broadcast mechanism
- `internal/api/channel_handlers.go` - how human messages are broadcast (reference pattern)
- `frontend/src/hooks/useWebSocket.ts` - WS event handling
- `frontend/src/hooks/useMessages.ts` - message state management
- `frontend/src/store/AppContext.tsx` - global state
