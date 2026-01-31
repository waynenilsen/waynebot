# 004: Agent transparency logging — full token capture with autopurge

## Problem

Agent activity is opaque. The `llm_calls` table stores empty placeholders for messages/responses (`actor.go:226` writes `'[]'` and `'{}'`). Tool executions are recorded but not queryable from the UI. There's no way to see what agents are actually doing.

## Requirements

### Full token capture

- Store the complete LLM request messages and response in `llm_calls` (currently placeholders)
- `actor.go:226` needs to serialize the actual `messages_json` and `response_json` instead of empty values
- This is the core transparency fix — every token in and out gets persisted

### Autopurge with archival to gzip

- Keep a rolling window of **10,000 events per agent** in the live `llm_calls` table
- When count exceeds 10k for a given `persona_id`, purge oldest rows
- Before deleting, compress purged rows to gzip text files on disk
- Archive path: configurable, e.g. `WAYNEBOT_ARCHIVE_DIR` defaulting to `./archives/`
- Archive file naming: `llm_calls_{persona_id}_{timestamp}.jsonl.gz`
- Same strategy for `tool_executions` table

### Purge mechanism

- Background goroutine on a timer (e.g. every 5 minutes) or triggered after each LLM call
- Per-agent check: `SELECT COUNT(*) FROM llm_calls WHERE persona_id = ?`
- If > 10,000: select oldest rows beyond the limit, write to gzip archive, then delete
- Must be transactional — don't delete rows that failed to archive

### No API/UI changes in this ticket

This is backend-only. Exposing logs in the UI is a separate concern.

## Files involved

| File | Line | What to change |
|------|------|----------------|
| `internal/agent/actor.go` | 222-232 | `recordLLMCall` — serialize actual messages and response JSON |
| `internal/agent/actor.go` | 234-244 | `recordToolExecution` — already captures data, just needs purge |
| `internal/db/migrations.go` | 93-115 | Tables already exist, no schema change needed |
| `internal/agent/` | new file | Archiver/purge goroutine |
| `internal/config/config.go` | | Add `WAYNEBOT_ARCHIVE_DIR` config |

## Notes

- The `messages_json` column could get large (full conversation context per call). Consider storing only the delta (new messages since last call) or accept the storage cost since autopurge keeps it bounded.
- Budget checking in `budget.go` queries `llm_calls` — purging old rows won't affect it since budget only looks at the last hour.
