# 005 — Agent tools defined but never registered into the tool registry

## Bug

The tool implementations (`shell_exec`, `file_read`, `file_write`, `http_fetch`) exist in `internal/tools/` but are never registered into the `Registry` that gets passed to the supervisor. When an LLM response includes tool calls, `Registry.Call()` will fail to find any tool.

## Root cause

In `cmd/waynebot/main.go`, the registry is created empty and handed straight to the supervisor:

```go
toolsRegistry := tools.NewRegistry()
supervisor := agent.NewSupervisor(database, hub, llmClient, toolsRegistry)
```

No `toolsRegistry.Register(...)` calls follow. The individual tool files define their logic but nothing wires them in.

## Files involved

| File | Lines | Role |
|------|-------|------|
| `cmd/waynebot/main.go` | ~46 | Creates empty registry, never registers tools |
| `internal/tools/registry.go` | 13-56 | `Registry` with `Register()` / `Call()` methods |
| `internal/tools/shell_exec.go` | — | `shell_exec` tool implementation |
| `internal/tools/file_read.go` | — | `file_read` tool implementation |
| `internal/tools/file_write.go` | — | `file_write` tool implementation |
| `internal/tools/http_fetch.go` | — | `http_fetch` tool implementation |

## Fix

Add registration calls in `main.go` after creating the registry, or add a `RegisterDefaults()` helper on `Registry` that registers all built-in tools, then call it from `main.go`.
