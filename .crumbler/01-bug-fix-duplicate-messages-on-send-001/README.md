# 001: Multiple messages appear on send in frontend

## Symptom

Sending a single message displays it 4+ times in the chat UI.

## Root cause

The message gets added to state from **two separate paths** with no deduplication:

1. **POST response** — `frontend/src/hooks/useMessages.ts:60` calls `addMessage(msg)` immediately after the HTTP response returns.
2. **WebSocket broadcast** — the backend broadcasts `new_message` to **all** connected clients including the sender (`internal/api/channel_handlers.go:214-219`), and `frontend/src/hooks/useWebSocket.ts:26` calls `addMessage()` again.

The reducer in `frontend/src/store/AppContext.tsx:38-45` (`ADD_MESSAGE`) blindly appends to the messages array without checking if the message ID already exists.

The 4x multiplier likely comes from React StrictMode double-mounting effects creating duplicate WebSocket connections.

## Fix options (pick one)

- **A) Remove the POST-response add.** Don't call `addMessage()` after POST — let the message arrive only via WebSocket. Simple, but adds slight latency before your own message appears.
- **B) Add dedup in the reducer.** Check `msg.id` before appending. Handles both paths safely regardless of timing.
- **C) Exclude sender from broadcast.** Backend skips the WebSocket message for the connection that sent the POST. Requires threading connection identity through the hub.

Option B is the safest since it also guards against any future duplicate delivery.

## Files involved

| File | Line | Role |
|------|------|------|
| `frontend/src/hooks/useMessages.ts` | 60 | adds message from POST response |
| `frontend/src/hooks/useWebSocket.ts` | 26 | adds message from WebSocket event |
| `frontend/src/store/AppContext.tsx` | 38-45 | ADD_MESSAGE reducer (no dedup) |
| `internal/api/channel_handlers.go` | 214-219 | broadcasts to all clients |
