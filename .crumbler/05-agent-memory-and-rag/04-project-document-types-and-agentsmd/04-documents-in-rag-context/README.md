# Include Project Documents in RAG Context

When assembling context, include project documents (ERD, PRD, recent decisions) alongside AGENTS.md.

## What to Do

Modify `AssembleContext()` in `internal/agent/context.go`:

After the system prompt + project context + AGENTS.md, and before memories:

1. If projects are associated, check for `.waynebot/` directory
2. Read `erd.md` and `prd.md` if they exist — include as high-priority context
3. Read `decisions.md` — include only last ~20 entries (entries are separated by `---` or `##` headers)
4. Add these as a "## Project Documents" section in the system message
5. Track tokens in `ContextBudget` (add `DocumentTokens int` field)

## Priority Order (updated)

1. System prompt (always)
2. Project context + AGENTS.md (if project associated)
3. Project documents — erd, prd, recent decisions (if they exist)
4. Retrieved memories (semantic search)
5. Channel message history (fills remaining budget)

## Token Budget

- Cap project documents at ~8000 tokens total (32K chars)
- If over budget, prioritize: AGENTS.md > ERD > PRD > decisions
- Truncate individual documents if needed

## Key Files
- `internal/agent/context.go` — `AssembleContext()` and helpers
