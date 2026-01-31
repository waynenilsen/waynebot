# 7d: Layout + Channels

**Goal:** Main app layout with sidebar and channel list. Use `/frontend-design` skill for all UI components.

## Files to Create

### `src/hooks/useChannels.ts`
- Fetch channels on mount via `getChannels()`
- Store in AppContext
- Expose `channels`, `currentChannel`, `selectChannel(id)`, `createChannel(name, desc)`

### `src/hooks/useWebSocket.ts`
- Connect WebSocket when user is authenticated
- Pass incoming events to AppContext (e.g. `new_message` → `addMessage`)
- Expose `connected: boolean` for UI indicator
- Disconnect on logout

### `src/components/Layout.tsx`
**Use `/frontend-design` skill.**

Slack-clone layout:
- Left sidebar (fixed width ~240px) with channel list, user info, logout button
- Main content area (flex-1) showing current channel's messages
- Responsive but desktop-first

### `src/components/Sidebar.tsx`
**Use `/frontend-design` skill.**

- App name/logo at top
- Channel list section
- User info at bottom (username, logout button)
- Navigation links to Personas, Agents, Invites pages

### `src/components/ChannelList.tsx`
**Use `/frontend-design` skill.**

- List of channels with # prefix
- Highlight current/active channel
- "New Channel" button/input
- Create channel inline form

### Update `src/App.tsx`
- After auth gate, render `<Layout>` with WebSocket connection active
- Route between channel view and admin pages (can use simple state-based routing or install react-router)

## Tests

- `src/hooks/useChannels.test.ts` — mock api, test fetch and select
- `src/hooks/useWebSocket.test.ts` — mock WebSocket, test connect/disconnect/event handling
- `src/components/ChannelList.test.tsx` — render with channels, test selection, test create

## Verification

`npx vitest run` passes. Sidebar shows channels, clicking switches active channel.
