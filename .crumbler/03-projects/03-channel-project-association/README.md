# Channel-Project Association

Allow channels to be associated with a project so agents know which project context to use.

## Migration (add to the same migration 8 if not yet applied, otherwise migration 9)

```sql
CREATE TABLE channel_projects (
  channel_id INTEGER NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
  project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  PRIMARY KEY (channel_id, project_id)
);
```

A channel can have multiple projects (or zero). This is a simple join table.

## Model (`internal/model/channel_project.go`)

- `SetChannelProject(db, channelID, projectID)` — associate a project with a channel (insert or ignore).
- `RemoveChannelProject(db, channelID, projectID)` — remove association.
- `ListChannelProjects(db, channelID)` — return all projects for a channel.
- `ListProjectChannels(db, projectID)` — return all channels for a project.

## API

```
GET    /api/channels/{id}/projects          — list projects for channel
POST   /api/channels/{id}/projects          — add project to channel (body: {project_id})
DELETE /api/channels/{id}/projects/{projectID} — remove project from channel
```

## Tests

Cover association CRUD and cascade behavior (deleting a project removes associations).
