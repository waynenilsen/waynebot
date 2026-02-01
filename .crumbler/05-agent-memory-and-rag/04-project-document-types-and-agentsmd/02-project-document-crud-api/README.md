# Project Document CRUD API

Add API endpoints for managing project documents stored as files in `.waynebot/` within the project directory.

## Document Types

Files stored in `{project.Path}/.waynebot/`:
- `erd.md` — Engineering Requirements Document
- `prd.md` — Product Requirements Document
- `decisions.md` — Decision Log (append-only)
- `notes/` — Free-form notes directory (future, skip for now)

## Endpoints

Add to `internal/api/router.go` under the projects route group:

```
GET    /api/projects/{id}/documents          — list available documents
GET    /api/projects/{id}/documents/{type}   — read a document
PUT    /api/projects/{id}/documents/{type}   — create/update a document
POST   /api/projects/{id}/documents/decisions — append to decision log
```

## Implementation

Create `internal/api/document_handlers.go`:

```go
type DocumentHandler struct {
    DB *db.DB
}
```

### ListDocuments
- Look up the project by ID
- Check which `.waynebot/*.md` files exist
- Return `[{type: "erd", exists: true}, {type: "prd", exists: false}, ...]`

### GetDocument
- Validate `{type}` is one of: erd, prd, decisions
- Read `{project.Path}/.waynebot/{type}.md`
- Return `{type, content}` or 404 if not found

### PutDocument
- Accept `{content}` body
- Validate type is erd or prd (NOT decisions — that's append-only)
- Create `.waynebot/` dir if needed (`os.MkdirAll`)
- Write `{project.Path}/.waynebot/{type}.md`

### AppendDecision
- Accept `{content}` body
- Create `.waynebot/` dir if needed
- Append to `decisions.md` with timestamp header

## Frontend API

Add to `frontend/src/api.ts`:
- `getProjectDocuments(projectId)`
- `getProjectDocument(projectId, type)`
- `updateProjectDocument(projectId, type, content)`
- `appendDecision(projectId, content)`

Add types to `frontend/src/types.ts`:
- `ProjectDocument { type: string; content: string; exists: boolean }`

## Validation
- Project must exist
- Project path must be a valid directory
- Type must be one of the known document types
- Reject path traversal in type parameter
