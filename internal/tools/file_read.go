package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxFileReadSize = 1 << 20 // 1MB

type fileReadArgs struct {
	Path string `json:"path"`
}

// FileRead returns a ToolFunc that reads files within the project directory.
// Path traversal is rejected and files larger than 1MB are refused.
func FileRead(baseDir string) ToolFunc {
	return func(ctx context.Context, raw json.RawMessage) (string, error) {
		var args fileReadArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("invalid args: %w", err)
		}
		if args.Path == "" {
			return "", fmt.Errorf("path is required")
		}

		dir := baseDir
		if d := ProjectDirFromContext(ctx); d != "" {
			dir = d
		}

		resolved, err := securePath(dir, args.Path)
		if err != nil {
			return "", err
		}

		info, err := os.Stat(resolved)
		if err != nil {
			return "", fmt.Errorf("stat: %w", err)
		}
		if info.IsDir() {
			return "", fmt.Errorf("path is a directory")
		}
		if info.Size() > maxFileReadSize {
			return "", fmt.Errorf("file too large: %d bytes (max %d)", info.Size(), maxFileReadSize)
		}

		data, err := os.ReadFile(resolved)
		if err != nil {
			return "", fmt.Errorf("read: %w", err)
		}
		return string(data), nil
	}
}

// securePath resolves path under baseDir and ensures it doesn't escape.
func securePath(baseDir, path string) (string, error) {
	cleaned := filepath.Clean(filepath.Join(baseDir, path))
	base := filepath.Clean(baseDir)

	rel, err := filepath.Rel(base, cleaned)
	if err != nil {
		return "", fmt.Errorf("path not allowed: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("path escapes sandbox: %q", path)
	}
	return cleaned, nil
}
