# Projects

Add a project system where a project is a folder on the filesystem.

## Key Requirements

- **DB table**: `projects` table with columns like `id`, `name`, `path` (absolute path to folder on disk), `description`, `created_at`.
- **API**: CRUD endpoints for projects. Validation that the path exists and is a directory.
- **Association**: Projects can be associated with channels (or personas) so agents know which project context to use. A `channel_projects` or `persona_projects` join table.
- **Agent awareness**: When an agent is processing messages in a channel associated with a project, it should have the project path available. The agent's file tools (`file_read`, `file_write`, `shell_exec`) should be scoped to the project directory instead of `/tmp/waynebot-sandbox`.
- **Frontend**: UI to create/manage projects, associate them with channels. Show project info in channel header or sidebar.
