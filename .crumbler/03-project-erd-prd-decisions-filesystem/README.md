# Project ERD/PRD/Decisions - Filesystem Based

## Problem
The current implementation has project documents (ERD, PRD, decisions) stored wrong. They should be simple files on disk inside the git project directory, not backed by any DB or special `.waynebot/` directory structure.

## How it should work
- Each project has a git repo directory (the project path)
- Inside the project path there are 3 folders: `erd/`, `prd/`, `decisions/`
- Each folder contains markdown documents
- State is dictated by what's physically on disk
- Nothing in the DB about these documents
- The API should read/write these files directly from disk
- The frontend should list and display these documents

## What to do
1. Update `internal/tools/project_docs.go` - change to read/write from `{project_path}/erd/`, `{project_path}/prd/`, `{project_path}/decisions/` directories instead of `.waynebot/`
2. Update `internal/api/document_handlers.go` - change API to list/read/write markdown files in these directories
3. Update the frontend ProjectsPage and any document components to work with multiple documents per category
4. Remove any DB-backed document storage (if any exists in memories table with erd/prd/decision_log kinds)
5. Update the context assembler to read documents from these directories instead of `.waynebot/`

## Key files
- `internal/tools/project_docs.go`
- `internal/api/document_handlers.go`
- `internal/agent/context.go`
- `frontend/src/pages/ProjectsPage.tsx`
- `frontend/src/api.ts`
