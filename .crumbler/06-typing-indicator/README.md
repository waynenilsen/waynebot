# Typing Indicator Support

## Problem
Need "typing..." indicator support so users can see when an agent is composing a response.

## What to do
1. Backend: The agent status tracker already has states (idle, thinking, tool_call, etc.). We need to broadcast status changes via WebSocket
2. Check `internal/agent/status.go` - ensure status changes emit WS events
3. Add a new WS event type like `agent_status` or `agent_typing` that includes persona_id, channel_id, and status
4. Frontend: Show a typing indicator in the message thread when an agent's status is "thinking" for the current channel
5. The indicator should appear below the last message, showing which agent(s) are typing
6. Use the existing `useAgentActivity` or `useWebSocket` hooks to listen for these events

## Key files
- `internal/agent/status.go`
- `internal/agent/actor.go` (where status transitions happen)
- `internal/ws/hub.go`
- `frontend/src/hooks/useWebSocket.ts`
- `frontend/src/components/MessageThread.tsx`
