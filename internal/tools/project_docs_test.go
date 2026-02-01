package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectDocsListEmpty(t *testing.T) {
	dir := t.TempDir()
	fn := ProjectDocs(dir)

	args, _ := json.Marshal(projectDocsArgs{Action: "list"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "erd: (empty)") {
		t.Errorf("expected erd empty, got: %s", out)
	}
	if !strings.Contains(out, "prd: (empty)") {
		t.Errorf("expected prd empty, got: %s", out)
	}
	if !strings.Contains(out, "decisions: (empty)") {
		t.Errorf("expected decisions empty, got: %s", out)
	}
}

func TestProjectDocsListWithFiles(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "erd"), 0o755)
	os.WriteFile(filepath.Join(dir, "erd", "main.md"), []byte("# ERD"), 0o644)

	fn := ProjectDocs(dir)
	args, _ := json.Marshal(projectDocsArgs{Action: "list"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "erd: main.md") {
		t.Errorf("expected erd to list main.md, got: %s", out)
	}
	if !strings.Contains(out, "prd: (empty)") {
		t.Errorf("expected prd empty, got: %s", out)
	}
}

func TestProjectDocsListByCategory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "erd"), 0o755)
	os.WriteFile(filepath.Join(dir, "erd", "main.md"), []byte("# ERD"), 0o644)
	os.WriteFile(filepath.Join(dir, "erd", "v2.md"), []byte("# ERD v2"), 0o644)

	fn := ProjectDocs(dir)
	args, _ := json.Marshal(projectDocsArgs{Action: "list", DocType: "erd"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "main.md") {
		t.Errorf("expected main.md in output, got: %s", out)
	}
	if !strings.Contains(out, "v2.md") {
		t.Errorf("expected v2.md in output, got: %s", out)
	}
}

func TestProjectDocsReadExists(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "erd"), 0o755)
	os.WriteFile(filepath.Join(dir, "erd", "main.md"), []byte("Users -> Posts"), 0o644)

	fn := ProjectDocs(dir)
	args, _ := json.Marshal(projectDocsArgs{Action: "read", DocType: "erd", Filename: "main"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if out != "Users -> Posts" {
		t.Errorf("got %q, want %q", out, "Users -> Posts")
	}
}

func TestProjectDocsReadNotFound(t *testing.T) {
	fn := ProjectDocs(t.TempDir())

	args, _ := json.Marshal(projectDocsArgs{Action: "read", DocType: "erd", Filename: "main"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for missing document")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectDocsReadInvalidType(t *testing.T) {
	fn := ProjectDocs(t.TempDir())

	args, _ := json.Marshal(projectDocsArgs{Action: "read", DocType: "bogus", Filename: "main"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for invalid doc type")
	}
	if !strings.Contains(err.Error(), "unknown doc_type") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectDocsReadRequiresFilename(t *testing.T) {
	fn := ProjectDocs(t.TempDir())

	args, _ := json.Marshal(projectDocsArgs{Action: "read", DocType: "erd"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for missing filename")
	}
	if !strings.Contains(err.Error(), "filename is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectDocsWriteCreatesDir(t *testing.T) {
	dir := t.TempDir()
	fn := ProjectDocs(dir)

	args, _ := json.Marshal(projectDocsArgs{Action: "write", DocType: "erd", Filename: "main", Content: "# New ERD"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "wrote") {
		t.Errorf("expected success message, got: %s", out)
	}

	data, err := os.ReadFile(filepath.Join(dir, "erd", "main.md"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(data) != "# New ERD" {
		t.Errorf("file content = %q, want %q", string(data), "# New ERD")
	}
}

func TestProjectDocsAppend(t *testing.T) {
	dir := t.TempDir()
	fn := ProjectDocs(dir)

	args, _ := json.Marshal(projectDocsArgs{Action: "append", DocType: "decisions", Filename: "log", Content: "We chose SQLite."})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "appended") {
		t.Errorf("expected success message, got: %s", out)
	}

	data, err := os.ReadFile(filepath.Join(dir, "decisions", "log.md"))
	if err != nil {
		t.Fatalf("read decisions: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "We chose SQLite.") {
		t.Error("decisions should contain appended content")
	}
	if !strings.Contains(content, "## 20") {
		t.Error("decisions should contain timestamp header")
	}
}

func TestProjectDocsAppendToExisting(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "decisions"), 0o755)
	os.WriteFile(filepath.Join(dir, "decisions", "log.md"), []byte("## First\nFirst decision."), 0o644)

	fn := ProjectDocs(dir)
	args, _ := json.Marshal(projectDocsArgs{Action: "append", DocType: "decisions", Filename: "log", Content: "Second decision."})
	_, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "decisions", "log.md"))
	content := string(data)
	if !strings.Contains(content, "First decision.") {
		t.Error("should preserve existing content")
	}
	if !strings.Contains(content, "Second decision.") {
		t.Error("should contain new decision")
	}
}

func TestProjectDocsAppendEmptyContent(t *testing.T) {
	fn := ProjectDocs(t.TempDir())

	args, _ := json.Marshal(projectDocsArgs{Action: "append", DocType: "decisions", Filename: "log", Content: "  "})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestProjectDocsInvalidAction(t *testing.T) {
	fn := ProjectDocs(t.TempDir())

	args, _ := json.Marshal(projectDocsArgs{Action: "delete"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
	if !strings.Contains(err.Error(), "unknown action") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectDocsUsesProjectDir(t *testing.T) {
	dir := t.TempDir()
	projectDir := t.TempDir()
	os.MkdirAll(filepath.Join(projectDir, "prd"), 0o755)
	os.WriteFile(filepath.Join(projectDir, "prd", "main.md"), []byte("project PRD"), 0o644)

	fn := ProjectDocs(dir)
	ctx := WithProjectDir(context.Background(), projectDir)
	args, _ := json.Marshal(projectDocsArgs{Action: "read", DocType: "prd", Filename: "main"})

	out, err := fn(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	if out != "project PRD" {
		t.Errorf("got %q, want %q", out, "project PRD")
	}
}
