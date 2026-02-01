# Context Window Management and Budget UX

The agent must go out-of-service when context becomes full. The user must be aware of remaining context and take decisive action.

## What to Build

### 1. Context Exhaustion Detection

After the context assembler runs (crumb 02), if `ContextBudget.Exhausted` is true (no room for meaningful history), the agent should:
- Post a message to the channel: "My context window is full. I cannot process new messages until context is reset. Please use `/reset-context` or start a new conversation thread."
- Set status to `StatusContextFull` (new status constant)
- NOT call the LLM
- Record the event

Add `StatusContextFull` to the status tracker constants.

### 2. Context Reset Mechanism

Add a way for users to reset an agent's context for a channel:
- API endpoint: `POST /api/agents/{persona_id}/channels/{channel_id}/reset-context`
- This should:
  - Reset the actor's cursor for that channel to the current latest message
  - Clear any cached context state
  - Set status back to idle
  - Post a system message: "Context has been reset. I'm ready to continue."

### 3. Frontend Context Budget Display

On the channel view, when an agent is a member, show a context budget indicator:
- API endpoint: `GET /api/agents/{persona_id}/context-budget?channel_id={id}`
- Returns the current ContextBudget estimate for this persona+channel
- Frontend shows a small bar/indicator: "Context: 45k/200k tokens used"
- When nearing exhaustion (>80%), show warning color
- When exhausted, show clear "Context Full" state with reset button

### 4. Agent Status Updates

Extend the agent status WebSocket events to include context budget info:
- When an agent responds, broadcast updated context usage
- Frontend `AgentDashboard` should show per-channel context usage

### 5. Tests

- Test that context exhaustion prevents LLM call
- Test that reset-context API works
- Test status transitions: idle -> thinking -> idle, and idle -> context_full -> reset -> idle

## Key Files to Modify
- `internal/agent/actor.go` — context exhaustion check
- `internal/agent/status.go` — add StatusContextFull
- NEW: `internal/api/context_handlers.go` — reset-context and budget endpoints
- `internal/api/router.go` — register new routes
- `frontend/src/components/ContextBudget.tsx` — new component
- `frontend/src/components/MessageThread.tsx` — show budget indicator
- `frontend/src/types.ts` — ContextBudget type
