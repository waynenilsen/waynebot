package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var knownDocTypes = map[string]bool{
	"erd":       true,
	"prd":       true,
	"decisions": true,
}

type projectDocsArgs struct {
	Action  string `json:"action"`
	DocType string `json:"doc_type"`
	Content string `json:"content"`
}

// ProjectDocs returns a ToolFunc that reads, writes, appends, or lists project
// documents in the .waynebot/ directory of the current project.
func ProjectDocs(baseDir string) ToolFunc {
	return func(ctx context.Context, raw json.RawMessage) (string, error) {
		var args projectDocsArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("invalid args: %w", err)
		}

		dir := baseDir
		if d := ProjectDirFromContext(ctx); d != "" {
			dir = d
		}

		switch args.Action {
		case "list":
			return projectDocsList(dir)
		case "read":
			return projectDocsRead(dir, args.DocType)
		case "write":
			return projectDocsWrite(dir, args.DocType, args.Content)
		case "append":
			return projectDocsAppend(dir, args.DocType, args.Content)
		default:
			return "", fmt.Errorf("unknown action %q: must be list, read, write, or append", args.Action)
		}
	}
}

func waynebotDocDir(baseDir string) string {
	return filepath.Join(baseDir, ".waynebot")
}

func waynebotDocPath(baseDir, docType string) string {
	return filepath.Join(waynebotDocDir(baseDir), docType+".md")
}

func projectDocsList(baseDir string) (string, error) {
	var lines []string
	for _, t := range []string{"erd", "prd", "decisions"} {
		_, err := os.Stat(waynebotDocPath(baseDir, t))
		exists := err == nil
		lines = append(lines, fmt.Sprintf("%s: exists=%v", t, exists))
	}
	return strings.Join(lines, "\n"), nil
}

func projectDocsRead(baseDir, docType string) (string, error) {
	if !knownDocTypes[docType] {
		return "", fmt.Errorf("unknown doc_type %q: must be erd, prd, or decisions", docType)
	}
	data, err := os.ReadFile(waynebotDocPath(baseDir, docType))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("document %q does not exist", docType)
		}
		return "", fmt.Errorf("read: %w", err)
	}
	return string(data), nil
}

func projectDocsWrite(baseDir, docType, content string) (string, error) {
	if !knownDocTypes[docType] {
		return "", fmt.Errorf("unknown doc_type %q: must be erd, prd, or decisions", docType)
	}
	if docType == "decisions" {
		return "", fmt.Errorf("decisions document can only be appended to, not overwritten; use action=append")
	}

	dir := waynebotDocDir(baseDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	if err := os.WriteFile(waynebotDocPath(baseDir, docType), []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}
	return fmt.Sprintf("wrote %d bytes to .waynebot/%s.md", len(content), docType), nil
}

func projectDocsAppend(baseDir, docType, content string) (string, error) {
	if docType != "decisions" {
		return "", fmt.Errorf("append is only supported for the decisions document")
	}
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("content is required for append")
	}

	dir := waynebotDocDir(baseDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	entry := fmt.Sprintf("\n## %s\n\n%s\n", time.Now().UTC().Format(time.RFC3339), content)

	f, err := os.OpenFile(waynebotDocPath(baseDir, "decisions"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return "", fmt.Errorf("append: %w", err)
	}
	return "appended decision entry to .waynebot/decisions.md", nil
}
