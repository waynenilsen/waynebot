package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const maxFileWriteSize = 1 << 20 // 1MB

type fileWriteArgs struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// FileWrite returns a ToolFunc that writes files restricted to the sandbox base
// directory. Path traversal is rejected and content larger than 1MB is refused.
func FileWrite(cfg *SandboxConfig) ToolFunc {
	return func(_ context.Context, raw json.RawMessage) (string, error) {
		var args fileWriteArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("invalid args: %w", err)
		}
		if args.Path == "" {
			return "", fmt.Errorf("path is required")
		}
		if len(args.Content) > maxFileWriteSize {
			return "", fmt.Errorf("content too large: %d bytes (max %d)", len(args.Content), maxFileWriteSize)
		}

		resolved, err := securePath(cfg.BaseDir, args.Path)
		if err != nil {
			return "", err
		}

		dir := filepath.Dir(resolved)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("mkdir: %w", err)
		}

		if err := os.WriteFile(resolved, []byte(args.Content), 0o644); err != nil {
			return "", fmt.Errorf("write: %w", err)
		}
		return fmt.Sprintf("wrote %d bytes to %s", len(args.Content), args.Path), nil
	}
}
