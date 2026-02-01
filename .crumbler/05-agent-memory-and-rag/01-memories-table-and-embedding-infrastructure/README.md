# Memories Table and Embedding Infrastructure

Set up the foundational data layer for agent memory: the `memories` table with vector embeddings, and the embedding generation client.

## What to Build

### 1. Database Migration (Version 10)

Add a `memories` table to `internal/db/migrations.go`:

```sql
CREATE TABLE memories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    persona_id INTEGER NOT NULL REFERENCES personas(id) ON DELETE CASCADE,
    channel_id INTEGER,
    project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL,
    kind TEXT NOT NULL CHECK(kind IN ('fact', 'decision', 'preference', 'summary', 'erd', 'prd', 'decision_log', 'note')),
    content TEXT NOT NULL,
    embedding BLOB,  -- float32 array serialized as bytes
    source_message_id INTEGER REFERENCES messages(id) ON DELETE SET NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_memories_persona ON memories(persona_id);
CREATE INDEX idx_memories_channel ON memories(channel_id);
CREATE INDEX idx_memories_project ON memories(project_id);
CREATE INDEX idx_memories_kind ON memories(kind);
```

### 2. Memory Model (`internal/model/memory.go`)

CRUD operations:
- `CreateMemory(db, personaID, channelID, projectID, kind, content, embedding, sourceMessageID) (Memory, error)`
- `GetMemory(db, id) (Memory, error)`
- `ListMemories(db, personaID, opts) ([]Memory, error)` — filter by channel, project, kind
- `UpdateMemory(db, id, content, embedding) error`
- `DeleteMemory(db, id) error`

The Memory struct:
```go
type Memory struct {
    ID              int64
    PersonaID       int64
    ChannelID       *int64
    ProjectID       *int64
    Kind            string
    Content         string
    Embedding       []float32
    SourceMessageID *int64
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### 3. Embedding Client (`internal/embedding/client.go`)

Create a new package `internal/embedding` with:
- A client that calls OpenRouter/OpenAI embeddings API (model: `openai/text-embedding-3-small`, 1536 dims)
- Use env var `WAYNEBOT_OPENROUTER_KEY` (same key as LLM)
- Interface: `Embed(ctx, text string) ([]float32, error)` and `EmbedBatch(ctx, texts []string) ([][]float32, error)`
- Serialize/deserialize helpers: `[]float32 <-> []byte` for BLOB storage

### 4. Cosine Similarity Search

Since SQLite doesn't have native vector ops, implement in-application:
- `SearchMemories(db, personaID, queryEmbedding []float32, limit int, filters) ([]ScoredMemory, error)`
- Load candidate memories from DB (filtered by persona/channel/project)
- Compute cosine similarity in Go
- Return top-N sorted by score

This is fine for the expected scale (hundreds to low thousands of memories per persona).

### 5. Tests

- Migration test: ensure version 10 applies cleanly
- Model CRUD tests for memories
- Embedding serialization round-trip test
- Cosine similarity correctness test (use known vectors)

## Key Files to Modify
- `internal/db/migrations.go` — add migration 10
- NEW: `internal/model/memory.go`
- NEW: `internal/model/memory_test.go`
- NEW: `internal/embedding/client.go`
- NEW: `internal/embedding/client_test.go`
- NEW: `internal/embedding/similarity.go`
- `cmd/waynebot/main.go` — initialize embedding client, pass to supervisor
