# 002 — Chat input box scrolls away with messages

## Bug

When a channel has more than a screenful of messages, the message compose input scrolls off the bottom of the viewport instead of staying pinned. The user must scroll down past all messages to find the input box.

## Root cause

The main content column in `App.tsx:121` is:

```tsx
<div className="flex-1 flex flex-col">
```

It is missing `min-h-0`. In a nested flex-column layout, a flex child defaults to `min-height: auto`, which means it refuses to shrink below its content size. When `MessageThread` fills with enough messages, the parent grows beyond the viewport height, pushing `MessageCompose` off-screen.

Adding `min-h-0` allows the flex child to shrink, so `MessageThread`'s `overflow-y-auto` kicks in and the message list scrolls internally while the input stays put.

## Files to change

| File | Line | What |
|------|------|------|
| `frontend/src/App.tsx` | 121 | Add `min-h-0` → `className="flex-1 flex flex-col min-h-0"` |

## Verification

1. Open a channel with many messages (or send enough to overflow the viewport).
2. Confirm the input box stays visible at the bottom at all times.
3. Confirm the message list scrolls independently.
