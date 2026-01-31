# Agent Actor & Supervisor

**Goal:** Create the Actor (per-persona processing loop) and Supervisor (manages all actors).

## Dependencies

Requires: tools package, llm package, agent core (status, cursor, decision, budget).

## Files to Create

- `internal/agent/actor.go` — Actor struct with Run(ctx) loop:
  1. Select on ctx.Done(), hub.NotifyChan, or fallback ticker (e.g. 60s)
  2. Get subscribed channels for persona
  3. For each channel: check cursor for new messages
  4. Call ShouldRespond() — skip if no, continue if yes
  5. Build message history from DB
  6. Check budget — stop if exceeded
  7. Call LLM (loop up to 10 rounds for tool calls)
  8. Execute tool calls via tools.Registry
  9. Post response as message (author_type="agent", author_id=persona.ID)
  10. Update cursor
  11. Broadcast via hub
- `internal/agent/actor_test.go` — Tests with mock LLM client (interface). Test: normal response, tool call loop, budget exceeded, self-message skip, max rounds.

- `internal/agent/supervisor.go` — Supervisor manages actor goroutines:
  - StartAll(): list personas from DB, start an actor goroutine for each
  - StopAll(): cancel all contexts, wait for goroutines
  - RestartActor(personaID): stop and restart a single actor
  - Uses sync.WaitGroup for clean shutdown
- `internal/agent/supervisor_test.go`

## Anti-Loop Rules (enforced in Actor)

1. Skip channel if all new messages from self
2. Per-channel cooldown (persona.CooldownSecs)
3. Max 10 tool-call rounds per LLM invocation
4. Token budget per persona per hour

## Key Interface

```go
// LLMClient interface for testability
type LLMClient interface {
    ChatCompletion(ctx context.Context, model string, messages []llm.Message, tools []llm.ToolDef, temperature float64, maxTokens int) (llm.Response, error)
}
```

## Verification

`go test ./internal/agent/...` passes.
