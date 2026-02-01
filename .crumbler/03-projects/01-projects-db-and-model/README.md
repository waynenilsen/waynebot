# Projects DB and Model

Add the `projects` table and Go model layer.

## Migration (version 8)

```sql
CREATE TABLE projects (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  path TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Model layer (`internal/model/project.go`)

Follow the same pattern as `channel.go`:

- `CreateProject(db, name, path, description)` — validate path exists and is a directory, then insert.
- `ListProjects(db)` — return all projects ordered by name.
- `GetProject(db, id)` — get single project by ID.
- `UpdateProject(db, id, name, path, description)` — validate path, update.
- `DeleteProject(db, id)` — delete project row.

## Tests (`internal/model/project_test.go`)

Cover CRUD operations and path validation (rejects non-existent paths, rejects files that aren't directories).
