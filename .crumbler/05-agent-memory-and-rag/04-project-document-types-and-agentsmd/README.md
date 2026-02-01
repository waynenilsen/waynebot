# Project Document Types and AGENTS.md

Add structured, file-based memory for projects using software engineering document conventions, and support AGENTS.md for agent instructions.

## What to Build

### 1. AGENTS.md Support

When a channel has an associated project, the agent should look for an `AGENTS.md` file in the project root directory:
- Read `{project.Path}/AGENTS.md` if it exists
- Include its contents in the system prompt context (high priority, after system prompt)
- This file contains project-specific instructions for agents

Modify `enrichPersonaPrompt()` in `actor.go` to:
1. Check for `AGENTS.md` in the project path
2. Read and include it
3. Keep it concise — estimate tokens and truncate if too large (preserve first N tokens)

### 2. Project Document Types

Projects can have structured documents stored as files in a `.waynebot/` directory within the project:
- `.waynebot/erd.md` — Engineering Requirements Document
- `.waynebot/prd.md` — Product Requirements Document
- `.waynebot/decisions.md` — Decision Log (append-only)
- `.waynebot/notes/` — Free-form notes directory

These are file-based (not DB) so they can be version-controlled with the project.

### 3. Document CRUD API

Add API endpoints for managing project documents:
- `GET /api/projects/{id}/documents` — list available documents
- `GET /api/projects/{id}/documents/{type}` — read a document (type = erd, prd, decisions, etc.)
- `PUT /api/projects/{id}/documents/{type}` — create/update a document
- `POST /api/projects/{id}/documents/decisions` — append to decision log

Validate that the project path exists and is writable. Documents are stored as markdown files.

### 4. Document Tool for Agents

Add a new tool `project_docs` that agents can use to read/write project documents:
- `project_docs read erd` — read the ERD
- `project_docs write prd "content..."` — write the PRD
- `project_docs append decisions "Decision: Use SQLite for..."` — append to decision log
- `project_docs list` — list available documents

Register in `internal/tools/` and `internal/llm/tools.go`.

### 5. Include Documents in RAG Context

When assembling context (crumb 02), if a project is associated:
1. Read AGENTS.md → include in system prompt section
2. Read active project documents (erd, prd) → include as memory-priority context
3. Decision log → include recent entries (last 20 or so)

### 6. Frontend: Project Documents Panel

Add a documents section to the project view:
- List documents with edit capability
- Markdown editor for each document type
- Decision log with append-only UI

### 7. Tests

- Test AGENTS.md reading and inclusion
- Test document CRUD via API
- Test document tool execution
- Test that project docs appear in assembled context

## Key Files to Modify
- `internal/agent/actor.go` — modify enrichPersonaPrompt to read AGENTS.md
- NEW: `internal/tools/project_docs.go` — project documents tool
- `internal/tools/registry.go` — register project_docs
- `internal/llm/tools.go` — add project_docs definition
- NEW: `internal/api/document_handlers.go` — document CRUD endpoints
- `internal/api/router.go` — register document routes
- NEW: `frontend/src/components/ProjectDocuments.tsx`
- `frontend/src/pages/ProjectsPage.tsx` — integrate documents panel
