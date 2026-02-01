# Project Documents Tool for Agents

Add a `project_docs` tool that agents can use to read/write project documents during conversations.

## Tool Definition

Add to `internal/llm/tools.go` in the `allTools` map:

```go
"project_docs": {
    Function: shared.FunctionDefinitionParam{
        Name:        "project_docs",
        Description: "Read, write, or list project documents (.waynebot/ directory).",
        Parameters: {
            "action": "read | write | append | list",
            "doc_type": "erd | prd | decisions (required for read/write/append)",
            "content": "content to write or append (required for write/append)",
        },
    },
}
```

## Tool Implementation

Create `internal/tools/project_docs.go`:

The tool needs access to the project path. Use the existing `ProjectDirFromContext(ctx)` pattern (same as `file_read.go`).

```go
func ProjectDocs(cfg *SandboxConfig) ToolFunc {
    return func(ctx context.Context, raw json.RawMessage) (string, error) {
        // Parse args: action, doc_type, content
        // Get project dir from context
        // Based on action:
        //   list: return which docs exist
        //   read: read .waynebot/{type}.md
        //   write: write .waynebot/{type}.md (not decisions)
        //   append: append to .waynebot/decisions.md with timestamp
    }
}
```

## Registration

In `internal/tools/registry.go` `RegisterDefaults()`:
```go
r.Register("project_docs", ProjectDocs(cfg))
```

## Key Files
- NEW: `internal/tools/project_docs.go`
- `internal/tools/registry.go` — add to RegisterDefaults
- `internal/llm/tools.go` — add tool definition to allTools map
