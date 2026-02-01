# Agent Project Awareness

When an agent processes messages in a channel with an associated project, scope file tools to the project directory.

## Changes

### 1. Actor channel processing (`internal/agent/actor.go`)

When processing a channel, look up associated projects via `ListChannelProjects`. If a channel has project(s), pass the first project's path into the tool context (or all paths if we want multi-project support — start with single/first for simplicity).

### 2. Tool context (`internal/tools/context.go`)

Add `WithProjectDir(ctx, path)` and `ProjectDirFromContext(ctx)` following the same pattern as `WithPersonaID`.

### 3. File tools update

In `file_read.go`, `file_write.go`, and `shell_exec.go`: if `ProjectDirFromContext(ctx)` returns a non-empty path, use that as the base directory instead of the sandbox default.

### 4. System prompt enrichment

When building messages for the LLM, if the channel has project(s), append project info to the system prompt so the agent knows:
- Project name and description
- That file tools are scoped to the project directory

## Tests

- Test that tools use project dir when set in context
- Test that tools fall back to sandbox when no project dir
