# Rip Out Sandbox & Allowed Commands

## Problem
The sandbox feature and allowed commands list are not needed. This app will run inside a sandbox at a higher level. The agent should have unrestricted shell access within its project directory.

## What to do
1. Remove `internal/tools/sandbox.go` - the SandboxConfig, AllowedCommands, BlockedHosts
2. Update `internal/tools/shell_exec.go` - remove allowed command checking, let any command run. Keep the timeout and output cap. The working directory should be the project's directory (from context)
3. Update `internal/tools/registry.go` - remove sandbox config parameter from RegisterDefaults
4. Update `internal/tools/http_fetch.go` - remove blocked hosts checking (or keep SSRF protection - use judgment)
5. Update `cmd/waynebot/main.go` - remove sandbox config initialization
6. Update `internal/tools/file_read.go` and `file_write.go` - remove sandbox path restrictions, use project dir as base
7. Remove any tests that test sandbox/allowed commands behavior, or update them
8. The agent's working directory should be the project path from the associated project

## Key files
- `internal/tools/sandbox.go`
- `internal/tools/shell_exec.go`
- `internal/tools/registry.go`
- `internal/tools/http_fetch.go`
- `internal/tools/file_read.go`
- `internal/tools/file_write.go`
- `cmd/waynebot/main.go`
