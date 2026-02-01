# DM Frontend: Agent Activity Integration

Enhance the agent activity page so users can DM an agent inline while viewing its activity.

## Changes to AgentActivityPage

In `frontend/src/pages/AgentActivityPage.tsx`:

Add a split view or panel:
- Left/main: existing activity feed (LLM calls, tool executions, stats)
- Right/bottom: DM conversation with this agent

When the activity page loads for a persona:
1. Call `createDM(personaId)` to get-or-create the DM channel
2. Load messages for that DM channel
3. Show a message input at the bottom of the DM panel

This lets users inspect what an agent is doing AND talk to it in the same view.

## Layout

Use a two-column layout on desktop:
- Left column (60-65%): activity feed (existing content)
- Right column (35-40%): DM conversation panel

On mobile/narrow screens, use tabs: "Activity" | "Chat"

## DM Panel Component

Create `frontend/src/components/AgentDMPanel.tsx`:
- Receives `personaId` as prop
- Auto-creates/fetches DM channel on mount
- Shows messages using existing message display components
- Has message input at bottom
- Subscribes to WebSocket for real-time updates
- Auto-scrolls on new messages

## Integration Points

- Reuse `useMessages` hook for the DM channel
- Reuse existing `MessageBubble` / message rendering components
- The DM panel should feel like a lightweight chat window embedded in the activity page
