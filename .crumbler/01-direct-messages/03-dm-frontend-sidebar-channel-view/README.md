# DM Frontend: Sidebar & Channel View

Add DM support to the frontend sidebar and message view.

## Types

In `frontend/src/types.ts`, add:
```typescript
interface DMChannel {
  id: number
  name: string
  is_dm: true
  created_at: string
  other_participant: {
    type: 'user' | 'persona'
    id: number
    name: string
  }
  unread_count: number
}
```

## API Client

In `frontend/src/api.ts`, add:
- `listDMs()` → `GET /api/dms`
- `createDM(targetUserId?: number, targetPersonaId?: number)` → `POST /api/dms`

## Sidebar Changes

In `Sidebar.tsx` / new `DMList.tsx`:
- Add a "Direct Messages" section below the Channels section
- List DMs with the other participant's name (no # prefix — use a person icon or just the name)
- Show unread badge same as channels
- Clicking a DM sets it as the current channel (reuse existing channel view)
- Add a "+" button to start a new DM — opens a picker to select a user or persona

## State Changes

In `AppContext.tsx`:
- Add `dms: DMChannel[]` to state
- Add actions: `SET_DMS`, `INCREMENT_DM_UNREAD`, `CLEAR_DM_UNREAD`
- When a DM is selected as current channel, reuse the same `currentChannelId` mechanism

## Hook

Create `frontend/src/hooks/useDMs.ts`:
- `dms` — list of DM channels
- `selectDM(id)` — sets as current channel
- `createDM(userId?, personaId?)` — creates DM and selects it
- Load DMs on mount

## Message View

The existing `ChatView` / message components should work as-is since DMs are just channels. Ensure the header shows the other participant's name instead of `#channel-name` when viewing a DM.
