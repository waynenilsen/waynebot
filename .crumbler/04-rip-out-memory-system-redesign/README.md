# Rip Out Memory System & Redesign

## Problem
The current memory system is too aggressive and the implementation is not liked. It uses embeddings, a memories DB table, and auto-extraction. All of that needs to go.

## New design
- Agent will be told (in system prompt) that memories are stored in chat history and can be searched via keywords
- Memories will also be stored in the project folder in `./memories/yyyy-mm-dd-hh-mm-title-in-kebab.md` files
- Use grep to look through these memories in the project dir

## What to do
1. Remove `internal/agent/memory_extractor.go` - no more auto-extraction
2. Remove `internal/embedding/` package entirely - no more vector embeddings
3. Remove `internal/tools/memory_search.go` - no more semantic search tool
4. Remove the `memories` table from the DB (add a migration to drop it)
5. Remove memory-related model code from `internal/model/memory.go`
6. Update `internal/agent/context.go` - remove memory/RAG section from context assembly
7. Update `internal/agent/actor.go` - remove memory extraction calls
8. Add instructions to the agent's system prompt about how memories work now (search chat history, use grep on project ./memories/ dir)
9. Consider adding a tool for the agent to create/search memory files (simple file operations in ./memories/)
10. Update `cmd/waynebot/main.go` - remove embedding client initialization
11. Clean up any frontend references to memories if they exist

## Key files
- `internal/agent/memory_extractor.go` (delete)
- `internal/embedding/client.go` (delete)
- `internal/tools/memory_search.go` (delete)
- `internal/model/memory.go` (delete or gut)
- `internal/agent/context.go`
- `internal/agent/actor.go`
- `cmd/waynebot/main.go`
- `internal/db/migrations.go`
