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
	Action   string `json:"action"`
	DocType  string `json:"doc_type"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// ProjectDocs returns a ToolFunc that reads, writes, appends, or lists project
// documents in the erd/, prd/, decisions/ directories of the current project.
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
			return projectDocsList(dir, args.DocType)
		case "read":
			return projectDocsRead(dir, args.DocType, args.Filename)
		case "write":
			return projectDocsWrite(dir, args.DocType, args.Filename, args.Content)
		case "append":
			return projectDocsAppend(dir, args.DocType, args.Filename, args.Content)
		default:
			return "", fmt.Errorf("unknown action %q: must be list, read, write, or append", args.Action)
		}
	}
}

func docTypeDir(baseDir, docType string) string {
	return filepath.Join(baseDir, docType)
}

func docFilePath(baseDir, docType, filename string) string {
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}
	return filepath.Join(docTypeDir(baseDir, docType), filename)
}

// listMarkdownFiles returns the names of all .md files in a directory.
func listMarkdownFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func projectDocsList(baseDir, docType string) (string, error) {
	// If no doc_type specified, list all categories and their files.
	if docType == "" {
		var lines []string
		for _, t := range []string{"erd", "prd", "decisions"} {
			files, _ := listMarkdownFiles(docTypeDir(baseDir, t))
			if len(files) == 0 {
				lines = append(lines, fmt.Sprintf("%s: (empty)", t))
			} else {
				lines = append(lines, fmt.Sprintf("%s: %s", t, strings.Join(files, ", ")))
			}
		}
		return strings.Join(lines, "\n"), nil
	}

	if !knownDocTypes[docType] {
		return "", fmt.Errorf("unknown doc_type %q: must be erd, prd, or decisions", docType)
	}

	files, err := listMarkdownFiles(docTypeDir(baseDir, docType))
	if err != nil {
		return "", fmt.Errorf("list %s: %w", docType, err)
	}
	if len(files) == 0 {
		return fmt.Sprintf("%s: (empty)", docType), nil
	}
	return fmt.Sprintf("%s: %s", docType, strings.Join(files, ", ")), nil
}

func projectDocsRead(baseDir, docType, filename string) (string, error) {
	if !knownDocTypes[docType] {
		return "", fmt.Errorf("unknown doc_type %q: must be erd, prd, or decisions", docType)
	}
	if filename == "" {
		return "", fmt.Errorf("filename is required for read")
	}
	data, err := os.ReadFile(docFilePath(baseDir, docType, filename))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("document %s/%s does not exist", docType, filename)
		}
		return "", fmt.Errorf("read: %w", err)
	}
	return string(data), nil
}

func projectDocsWrite(baseDir, docType, filename, content string) (string, error) {
	if !knownDocTypes[docType] {
		return "", fmt.Errorf("unknown doc_type %q: must be erd, prd, or decisions", docType)
	}
	if filename == "" {
		return "", fmt.Errorf("filename is required for write")
	}

	dir := docTypeDir(baseDir, docType)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	fp := docFilePath(baseDir, docType, filename)
	if err := os.WriteFile(fp, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}
	return fmt.Sprintf("wrote %d bytes to %s/%s", len(content), docType, filepath.Base(fp)), nil
}

func projectDocsAppend(baseDir, docType, filename, content string) (string, error) {
	if !knownDocTypes[docType] {
		return "", fmt.Errorf("unknown doc_type %q: must be erd, prd, or decisions", docType)
	}
	if filename == "" {
		return "", fmt.Errorf("filename is required for append")
	}
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("content is required for append")
	}

	dir := docTypeDir(baseDir, docType)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	entry := fmt.Sprintf("\n## %s\n\n%s\n", time.Now().UTC().Format(time.RFC3339), content)

	fp := docFilePath(baseDir, docType, filename)
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return "", fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return "", fmt.Errorf("append: %w", err)
	}
	return fmt.Sprintf("appended entry to %s/%s", docType, filepath.Base(fp)), nil
}
