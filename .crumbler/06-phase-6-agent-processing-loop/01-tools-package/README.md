# Tools Package

**Goal:** Create `internal/tools/` with sandboxed tool implementations that agents can invoke.

## Files to Create

- `internal/tools/sandbox.go` — SandboxConfig: allowed commands list, allowed base directory, blocked hosts
- `internal/tools/sandbox_test.go`
- `internal/tools/registry.go` — Registry maps tool names to ToolFunc type. ToolFunc signature: `func(ctx context.Context, args json.RawMessage) (string, error)`
- `internal/tools/registry_test.go`
- `internal/tools/shell_exec.go` — Sandboxed shell command execution. Allowed-command allowlist, 30s timeout, 10KB output cap
- `internal/tools/shell_exec_test.go`
- `internal/tools/file_read.go` — Path-restricted file read. Restricted to sandbox dir, no path traversal, 1MB max
- `internal/tools/file_read_test.go`
- `internal/tools/file_write.go` — Path-restricted file write. Same restrictions as file_read
- `internal/tools/file_write_test.go`
- `internal/tools/http_fetch.go` — HTTP fetch with denied hosts list (localhost, 127.0.0.1, etc), 15s timeout, 50KB max response
- `internal/tools/http_fetch_test.go`

## Sandbox Rules

- shell_exec: Configurable allowed-command allowlist, 30s timeout, 10KB output
- file_read: Restricted to sandbox dir, no path traversal (`..`), 1MB max
- file_write: Same as file_read
- http_fetch: Blocked hosts list (localhost, 127.0.0.1, 0.0.0.0, 169.254.169.254, ::1), 15s timeout, 50KB max

## Key Types

```go
type SandboxConfig struct {
    AllowedCommands []string
    BaseDir         string
    BlockedHosts    []string
}

type ToolFunc func(ctx context.Context, args json.RawMessage) (string, error)

type Registry struct {
    tools map[string]ToolFunc
}
```

## Verification

`go test ./internal/tools/...` passes.
