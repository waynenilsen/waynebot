# AGENTS.md Support in Context Assembly

When a channel has associated projects, read `AGENTS.md` from the project root and include it in the system prompt.

## What to Do

Modify `formatProjectContext()` in `internal/agent/context.go` to:

1. For each project in `input.Projects`, check if `{project.Path}/AGENTS.md` exists
2. If it exists, read the file contents
3. Append to the system prompt as a high-priority section after the project context block
4. Truncate if too large — cap at ~4000 tokens (16K chars) to avoid blowing the budget
5. Track AGENTS.md tokens in the `ContextBudget` struct (add `AgentsmdTokens int` field)

## Key Code

- `internal/agent/context.go` — `formatProjectContext()` function (line ~197)
- `ContextBudget` struct — add `AgentsmdTokens` field

## Example Output in System Prompt

```
## Project Context
This channel is associated with the project **myproject**.
...

## Project Instructions (AGENTS.md)
<contents of AGENTS.md>
```

## Notes

- Use `os.ReadFile` — if the file doesn't exist or can't be read, skip silently
- Only read from the first project (projects[0]) since that's how `formatProjectContext` works today
