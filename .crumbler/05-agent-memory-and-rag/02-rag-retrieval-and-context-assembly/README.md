# RAG Retrieval and Context Assembly

Before an agent responds, it should retrieve relevant memories and assemble context deterministically with clear priority ordering.

## What to Build

### 1. Context Assembler (`internal/agent/context.go`)

Create a context assembler that builds the LLM message array with priority:

1. **System prompt** (always included, non-negotiable)
2. **Project context** — AGENTS.md file contents if project is associated (see crumb 04)
3. **Retrieved memories** — semantic search results from memories table
4. **Recent channel messages** — the conversation history (this is where we flex)

The assembler should:
- Accept a token budget (from persona.MaxTokens or a context window config)
- Estimate tokens for each section (rough: 1 token ≈ 4 chars for English)
- Fill sections in priority order, allocating remaining budget to channel messages
- Return the assembled `[]openai.ChatCompletionMessageParamUnion` and a `ContextBudget` struct showing what was included

```go
type ContextBudget struct {
    TotalTokens      int
    SystemTokens     int
    MemoryTokens     int
    ProjectTokens    int
    HistoryTokens    int
    HistoryMessages  int  // how many messages fit
    Exhausted        bool // true if context is full
}
```

### 2. RAG Retrieval Step

In `actor.go`'s `respond()` method, before calling the LLM:

1. Take the last few messages (current trigger) and embed them as a query
2. Search memories for this persona (scoped to channel + global)
3. Format retrieved memories as a system-level context block:
   ```
   ## Relevant Memories
   - [fact] User prefers Go over Rust for this project
   - [decision] We agreed to use SQLite for storage
   ```
4. Pass to the context assembler

### 3. Modify `actor.go` respond()

Replace the current simple approach:
```go
// OLD: history, err := model.GetRecentMessages(a.DB, ch.ID, 50)
// NEW: use context assembler
```

The actor needs access to the embedding client and memory search. Add `Embedding` field to Actor struct.

### 4. Token Estimation

Simple heuristic is fine — don't need tiktoken:
- `EstimateTokens(text string) int` — `len(text) / 4`
- Model context windows: store as constant map or persona config
- Default context window: 200k for Claude models, configurable

### 5. Tests

- Context assembler: test priority ordering (memories before history)
- Context assembler: test budget exhaustion (short budget = fewer history messages)
- Test that retrieved memories appear in assembled context

## Key Files to Modify
- NEW: `internal/agent/context.go`
- NEW: `internal/agent/context_test.go`
- `internal/agent/actor.go` — modify respond() to use context assembler
- `internal/agent/actor.go` — add Embedding field to Actor struct
- `internal/agent/supervisor.go` — pass embedding client to actors
