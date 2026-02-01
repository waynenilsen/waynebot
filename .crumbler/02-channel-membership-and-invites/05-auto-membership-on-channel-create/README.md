# Auto-Membership on Channel Create

## What

When a user creates a channel, they should automatically become a member with the "owner" role. Without this, creating a channel would result in an invisible channel.

## Changes

- In `CreateChannel` handler (or model function), after inserting the channel, also insert a `channel_members` row with role='owner' for the creating user.
- This requires knowing the user ID in `CreateChannel` — update the model function signature: `CreateChannel(db, name, description, createdByUserID)`.
- Update the `CreateChannel` handler to pass the authenticated user's ID.

## Backfill

- Add a backfill step in the migration (or a separate migration) that adds all existing users as members of all existing channels. This prevents existing channels from disappearing after the membership filter is applied. Use role='owner' for simplicity since there's no creator tracking on old channels.
