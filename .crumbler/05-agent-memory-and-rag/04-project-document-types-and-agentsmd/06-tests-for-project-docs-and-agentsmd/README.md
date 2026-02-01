# Tests for Project Docs and AGENTS.md

Write tests covering all the new functionality.

## Tests to Write

### 1. AGENTS.md in Context (`internal/agent/context_test.go`)

- `TestAssembleContextWithAgentsmd` — create a temp dir with AGENTS.md, create a project pointing there, verify system message contains AGENTS.md content
- `TestAssembleContextAgentsmdMissing` — project without AGENTS.md, verify no error and normal behavior
- `TestAssembleContextAgentsmdTruncation` — very large AGENTS.md, verify it gets truncated

### 2. Document CRUD API (`internal/api/document_handlers_test.go`)

- Test list documents (empty project, project with some docs)
- Test get document (exists, doesn't exist)
- Test put document (create new, update existing)
- Test append decision (empty log, existing log)
- Test validation (invalid type, missing project)

### 3. Project Docs Tool (`internal/tools/project_docs_test.go`)

- Test list action
- Test read action (doc exists, doesn't exist)
- Test write action (creates .waynebot dir, writes file)
- Test append action (to decisions)
- Test invalid action

### 4. Documents in RAG Context (`internal/agent/context_test.go`)

- `TestAssembleContextWithProjectDocs` — verify ERD/PRD content appears in system message
- `TestAssembleContextDocsBudget` — verify documents respect token budget

## Test Pattern

Follow existing fluent test pattern:
```go
func TestXxx(t *testing.T) {
    d := openTestDB(t)
    // setup persona, channel, project with temp dir
    // create .waynebot/ files in temp dir
    // run the thing
    // assert
}
```

Use `t.TempDir()` for project paths so cleanup is automatic.

## Key Files
- `internal/agent/context_test.go` — add context tests
- NEW: `internal/api/document_handlers_test.go`
- NEW: `internal/tools/project_docs_test.go`
