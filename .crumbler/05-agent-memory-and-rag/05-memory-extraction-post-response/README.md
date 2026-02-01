# Memory Extraction Post-Response

After an agent responds, extract key facts, decisions, and preferences and store them as memories for future RAG retrieval.

## What to Build

### 1. Memory Extractor (`internal/agent/memory_extractor.go`)

After the agent posts its final response, run a lightweight extraction step:
- Take the current conversation context (recent messages + agent response)
- Use a secondary LLM call with a specific extraction prompt
- The extraction prompt asks the model to identify:
  - **Facts**: concrete information stated ("the API uses REST", "deploy target is AWS")
  - **Decisions**: choices made ("we decided to use PostgreSQL", "going with approach B")
  - **Preferences**: user preferences expressed ("prefer functional style", "no ORMs")
- Output format: JSON array of `{kind, content}` objects

### 2. Extraction Prompt

```
You are a memory extraction system. Given the conversation below, extract key facts, decisions, and preferences that would be useful to remember for future conversations.

Output a JSON array of objects with "kind" (fact/decision/preference) and "content" (concise statement).
Only extract genuinely important information. Skip small talk and transient details.
If nothing important was discussed, return an empty array [].
```

### 3. Storage Pipeline

After extraction:
1. Parse the JSON response
2. For each extracted memory:
   a. Generate embedding via the embedding client
   b. Check for duplicates (cosine similarity > 0.9 against existing memories)
   c. If not duplicate, store in memories table
3. Associate with persona_id, channel_id, and project_id if applicable

### 4. Budget-Conscious Extraction

- Only run extraction if the conversation had substantive content (> N messages from humans)
- Use a cheap/fast model for extraction (e.g., a smaller model or lower max_tokens)
- Track extraction token usage separately in llm_calls (mark with a metadata flag)
- Skip extraction if persona is over budget

### 5. Tests

- Test extraction prompt produces valid JSON
- Test duplicate detection (similar memories not re-stored)
- Test that extraction runs after respond() completes
- Test extraction skipped when no substantive content
- Mock the LLM call for unit tests

## Key Files to Modify
- NEW: `internal/agent/memory_extractor.go`
- NEW: `internal/agent/memory_extractor_test.go`
- `internal/agent/actor.go` — call extractor after postMessage in respond()
