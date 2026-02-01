# DM Backend: Migration & Models

Add database support for DMs by extending the channels table and adding model functions.

## Migration (Migration 5)

Add to the `channels` table:
- `is_dm BOOLEAN NOT NULL DEFAULT 0` — distinguishes DMs from regular channels
- `created_by INTEGER` — the user who initiated the DM (nullable, only for DMs)

Add a new table `dm_participants`:
```sql
CREATE TABLE dm_participants (
  channel_id INTEGER NOT NULL REFERENCES channels(id),
  user_id    INTEGER,
  persona_id INTEGER,
  PRIMARY KEY (channel_id, COALESCE(user_id, 0), COALESCE(persona_id, 0))
);
CREATE INDEX idx_dm_participants_user ON dm_participants(user_id);
CREATE INDEX idx_dm_participants_persona ON dm_participants(persona_id);
```

Each DM channel has exactly 2 rows in `dm_participants`. A participant is either a user (user_id set, persona_id NULL) or a persona/agent (persona_id set, user_id NULL).

## Model Changes

In `internal/model/channel.go`:
- Update `Channel` struct to include `IsDM bool` and `CreatedBy *int64`
- `CreateDMChannel(name, participant1, participant2)` — creates channel with `is_dm=1`, inserts 2 dm_participants rows, and if either participant is a persona, auto-subscribes it via `persona_channels`
- `FindDMChannel(participant1, participant2)` — returns existing DM channel between two participants (to avoid duplicates)
- `ListDMsForUser(userID)` — returns all DM channels where user is a participant, with the other participant's info
- Update `ListChannels` to only return non-DM channels (add `WHERE is_dm = 0`)

## Participant Type

Use a simple struct to represent a DM participant:
```go
type DMParticipant struct {
  UserID    *int64
  PersonaID *int64
}
```
