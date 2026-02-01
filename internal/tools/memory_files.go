package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type memorySaveArgs struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// MemorySave returns a ToolFunc that saves a memory to a markdown file
// in the project's ./memories/ directory.
func MemorySave() ToolFunc {
	return func(ctx context.Context, raw json.RawMessage) (string, error) {
		projectDir := ProjectDirFromContext(ctx)
		if projectDir == "" {
			return "", fmt.Errorf("no project directory in context")
		}

		var args memorySaveArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}
		if strings.TrimSpace(args.Title) == "" {
			return "", fmt.Errorf("title is required")
		}
		if strings.TrimSpace(args.Content) == "" {
			return "", fmt.Errorf("content is required")
		}

		memDir := filepath.Join(projectDir, "memories")
		if err := os.MkdirAll(memDir, 0755); err != nil {
			return "", fmt.Errorf("create memories dir: %w", err)
		}

		filename := fmt.Sprintf("%s-%s.md",
			time.Now().Format("2006-01-02-15-04"),
			args.Title,
		)
		path := filepath.Join(memDir, filename)

		if err := os.WriteFile(path, []byte(args.Content), 0644); err != nil {
			return "", fmt.Errorf("write memory file: %w", err)
		}

		return fmt.Sprintf("Memory saved to memories/%s", filename), nil
	}
}

type memorySearchFilesArgs struct {
	Query string `json:"query"`
}

// MemorySearchFiles returns a ToolFunc that searches memory files using grep.
func MemorySearchFiles() ToolFunc {
	return func(ctx context.Context, raw json.RawMessage) (string, error) {
		projectDir := ProjectDirFromContext(ctx)
		if projectDir == "" {
			return "", fmt.Errorf("no project directory in context")
		}

		var args memorySearchFilesArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}
		if strings.TrimSpace(args.Query) == "" {
			return "", fmt.Errorf("query is required")
		}

		memDir := filepath.Join(projectDir, "memories")
		if _, err := os.Stat(memDir); os.IsNotExist(err) {
			return "No memories directory found. No memories have been saved yet.", nil
		}

		cmd := exec.CommandContext(ctx, "grep", "-r", "-i", "-l", args.Query, memDir)
		matchingFiles, err := cmd.Output()
		if err != nil {
			// grep returns exit code 1 when no matches found
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return "No matching memories found.", nil
			}
			return "", fmt.Errorf("grep: %w", err)
		}

		if strings.TrimSpace(string(matchingFiles)) == "" {
			return "No matching memories found.", nil
		}

		// Now get context from matching files
		cmd = exec.CommandContext(ctx, "grep", "-r", "-i", "-n", "--include=*.md", "-C", "2", args.Query, memDir)
		output, err := cmd.Output()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return "No matching memories found.", nil
			}
			return "", fmt.Errorf("grep: %w", err)
		}

		result := strings.TrimSpace(string(output))
		// Make paths relative to project dir for readability
		result = strings.ReplaceAll(result, memDir+"/", "memories/")

		if len(result) > 4000 {
			result = result[:4000] + "\n... (truncated)"
		}

		return result, nil
	}
}
