# Agent RAG Tool

Give agents an explicit tool to search their own memories and project context on-demand.

## What to Build

### 1. `memory_search` Tool

Add a new tool that agents can invoke during conversation:

```json
{
  "name": "memory_search",
  "description": "Search your memories and project context for relevant information. Use this when you need to recall past decisions, facts, or context that may not be in your current conversation window.",
  "parameters": {
    "query": "string - what to search for",
    "kind": "string (optional) - filter by memory kind: fact, decision, preference, summary, erd, prd, decision_log, note",
    "limit": "integer (optional) - max results, default 10"
  }
}
```

### 2. Tool Implementation (`internal/tools/memory_search.go`)

When invoked:
1. Embed the query text
2. Search memories for the calling persona (use persona_id from tool context)
3. Optionally filter by kind
4. Return formatted results:
   ```
   Found 3 relevant memories:

   1. [decision] (score: 0.89) We decided to use SQLite for the storage layer
      — from #engineering, 2024-01-15

   2. [fact] (score: 0.85) The API rate limit is 1000 req/min
      — from #api-design, 2024-01-10

   3. [preference] (score: 0.82) User prefers conventional commits format
      — global, 2024-01-08
   ```

### 3. Registration

- Add to `internal/llm/tools.go` allTools map
- Add to `internal/tools/registry.go`
- The tool needs access to the embedding client and DB — pass via tool context

### 4. Auto-Enable

Consider adding `memory_search` to the default tools_enabled list for new personas, or let it be opt-in per persona.

### 5. Tests

- Test tool execution returns formatted results
- Test kind filtering works
- Test empty results case
- Test that tool context correctly identifies the calling persona

## Key Files to Modify
- NEW: `internal/tools/memory_search.go`
- NEW: `internal/tools/memory_search_test.go`
- `internal/tools/registry.go` — register memory_search
- `internal/tools/context.go` — add embedding client and DB to context if not already available
- `internal/llm/tools.go` — add tool definition
- `internal/agent/supervisor.go` — ensure tools registry has access to embedding client
