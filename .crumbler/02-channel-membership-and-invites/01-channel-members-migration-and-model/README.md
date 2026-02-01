# channel_members Migration and Model

## What

Add a `channel_members` table for human membership in non-DM channels. Agents already have `persona_channels` — this is the human equivalent.

## Schema (migration 6)

```sql
CREATE TABLE channel_members (
    channel_id INTEGER NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('owner', 'member')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (channel_id, user_id)
);
CREATE INDEX idx_channel_members_user ON channel_members(user_id);
```

## Model functions (internal/model/membership.go)

- `AddChannelMember(db, channelID, userID, role)` — INSERT OR IGNORE
- `RemoveChannelMember(db, channelID, userID)`
- `GetChannelMembers(db, channelID) -> []ChannelMember` — returns user_id, username, role
- `IsChannelMember(db, channelID, userID) -> bool`
- `ListChannelsForUser(db, userID) -> []Channel` — all non-DM channels where user is a member

## Notes

- Keep `persona_channels` as-is for agents. The two tables serve the same purpose for different entity types.
- The `role` column enables owner-only actions later (like removing members).
- Don't change any API or frontend yet — just the migration + model layer.
