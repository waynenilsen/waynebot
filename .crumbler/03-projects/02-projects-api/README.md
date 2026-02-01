# Projects API

Add CRUD REST endpoints for projects.

## Endpoints

```
GET    /api/projects          — list all projects
POST   /api/projects          — create project (name, path, description)
PUT    /api/projects/{id}     — update project
DELETE /api/projects/{id}     — delete project
```

## Handler (`internal/api/projects.go`)

Follow the patterns in `channels.go` / `personas.go`:

- Parse JSON request bodies
- Call model functions
- Return JSON responses
- Proper error codes (400 for bad input, 404 for not found, etc.)

## Router

Register routes in `router.go` under the authenticated group.
