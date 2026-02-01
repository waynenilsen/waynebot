# Invite UX (Backend + Frontend)

## What

Users need a way to invite other users and agents into channels. This is the main UX for channel membership management.

## Approach: Member Management Panel

Add a members panel accessible from the channel header. Shows current members and allows adding new ones.

### Frontend

- Add a "members" button/icon in the channel header area (or message thread header).
- Clicking it opens a slide-out panel or modal showing:
  - Current members list (from `GET /api/channels/{id}/members`)
  - An "Add member" section with a searchable dropdown of all users and personas
  - Remove button next to each member (except owner, or only for owner to use)
- Use existing `GET /api/personas` for persona list. Need a new `GET /api/users` endpoint for user list.

### Backend

- Add `GET /api/users` endpoint (returns id + username for all users, auth required).
- Add `ListUsers(db) -> []User` model function.
- The add/remove member endpoints from step 2 are used here.

### Agent Integration

- When a persona is added to a channel via the members panel, the supervisor should restart that actor so it picks up the new channel immediately (same pattern as DM creation in `dm_handlers.go`).
