# 7g: Polish

**Goal:** Error handling, loading states, reconnection UX, and keyboard shortcuts.

## Files to Create/Update

### `src/components/ErrorBanner.tsx`
- Dismissable error banner at top of app
- Auto-dismiss after timeout
- Stack multiple errors

### Empty States
- Add empty state illustrations/messages for:
  - No channels yet ("Create your first channel")
  - No messages in channel ("Start the conversation")
  - No personas ("Create a persona to get started")
  - No invites ("Generate an invite to add team members")

### Loading States
- Skeleton loaders or spinners for:
  - Channel list loading
  - Message thread loading
  - Initial auth check
  - Page transitions

### WebSocket Reconnection Indicator
- Show a banner/toast when WebSocket disconnects
- "Reconnecting..." indicator with status
- "Connected" flash when reconnection succeeds

### Keyboard Shortcuts
- `Ctrl/Cmd + K` — Quick channel switcher (search/filter channels)
- `Escape` — Close modals/overlays
- `Up arrow` in empty compose — edit last message (stretch goal)
- Focus compose box when switching channels

### General Polish
- Consistent error handling across all API calls (surface errors via ErrorBanner)
- Proper focus management and tab order
- Responsive layout adjustments
- Favicon and page title updates

## Verification

`npx vitest run` passes. All polish items visible in browser. No console errors during normal usage.
