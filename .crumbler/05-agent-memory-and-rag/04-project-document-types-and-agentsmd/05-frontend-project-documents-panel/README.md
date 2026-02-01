# Frontend: Project Documents Panel

Add a documents section to the project view for managing project documents.

## What to Build

### ProjectDocuments Component

Create `frontend/src/components/ProjectDocuments.tsx`:

- Shows a list of document types (ERD, PRD, Decisions) with their status (exists/empty)
- Click a document to open it in an editor
- Markdown textarea editor for ERD and PRD (full replace on save)
- Decision log with append-only UI — text input + "Add Decision" button
- Each decision gets a timestamp automatically

### Integration

Add to `frontend/src/pages/ProjectsPage.tsx`:
- When a project is selected/expanded, show the documents panel
- Or add as a tab/section within project detail view

### UI Design

Keep it simple:
- Tab bar or accordion: ERD | PRD | Decisions
- Each tab shows:
  - A textarea for editing (monospace font for markdown)
  - Save button
  - For decisions: read-only display of existing + append input at bottom

### API Integration

Use the API functions from crumb 02:
- `getProjectDocuments(projectId)` to list
- `getProjectDocument(projectId, type)` to read
- `updateProjectDocument(projectId, type, content)` to save
- `appendDecision(projectId, content)` to add decision

## Key Files
- NEW: `frontend/src/components/ProjectDocuments.tsx`
- `frontend/src/pages/ProjectsPage.tsx` — integrate panel
- `frontend/src/api.ts` — API functions (from crumb 02)
- `frontend/src/types.ts` — types (from crumb 02)
