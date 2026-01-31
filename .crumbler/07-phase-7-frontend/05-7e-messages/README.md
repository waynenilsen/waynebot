# 7e: Messages

**Goal:** Message thread display, individual messages, and compose box. Use `/frontend-design` skill for all UI components.

## Files to Create

### `src/hooks/useMessages.ts`
- Fetch messages for current channel via `getMessages(channelId, { limit: 50 })`
- Support cursor-based pagination: load older messages when user scrolls to top (pass `before` param)
- Append new messages from WebSocket events (already wired in AppContext)
- Expose `messages`, `loading`, `loadMore()`, `sendMessage(content)`

### `src/components/MessageThread.tsx`
**Use `/frontend-design` skill.**

- Scrollable message list for the current channel
- Auto-scroll to bottom on new messages (unless user has scrolled up)
- "Load more" trigger at top (infinite scroll or button)
- Empty state when no messages

### `src/components/MessageItem.tsx`
**Use `/frontend-design` skill.**

- Display single message with: avatar/icon, author name, timestamp, content
- Differentiate human vs persona messages visually (e.g. bot badge, different avatar style)
- Render message content as markdown (via MarkdownRenderer)

### `src/components/MarkdownRenderer.tsx`
- Simple markdown rendering for message content
- Support basics: bold, italic, code blocks, inline code, links, lists
- Can use a lightweight lib like `react-markdown` or hand-roll basics
- Sanitize output (no raw HTML)

### `src/components/MessageCompose.tsx`
**Use `/frontend-design` skill.**

- Text input area (textarea, auto-growing)
- Send on Enter (Shift+Enter for newline)
- Send button
- Disabled state while sending
- Character count indicator near 10,000 char limit

## Tests

- `src/hooks/useMessages.test.ts` — mock api, test fetch, pagination, send
- `src/components/MessageThread.test.tsx` — render with messages, test scroll behavior
- `src/components/MessageCompose.test.tsx` — type message, submit, assert callback

## Verification

`npx vitest run` passes. Can send and receive messages in browser.
