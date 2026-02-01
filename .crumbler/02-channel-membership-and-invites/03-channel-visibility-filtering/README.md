# Channel Visibility Filtering

## What

Users should only see channels they're members of. Right now `GET /api/channels` returns all non-DM channels.

## Backend Changes

- Change `ListChannels` handler to use `ListChannelsForUser(db, userID)` instead of `ListChannels(db)`.
- This means `GET /api/channels` now returns only channels the user belongs to.
- Also add membership checks to `PostMessage`, `GetMessages`, and `MarkRead` — return 403 if user is not a member.

## Frontend Changes

- No structural changes needed — the channel list already renders whatever the API returns.
- The channel switcher (Cmd+K) should also only show member channels (it already uses the same state).

## Edge Case

- What about the very first channel? When there are zero channels and zero members, the first user who creates a channel gets auto-added as owner (handled in step 5).
