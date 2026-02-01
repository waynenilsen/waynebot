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
	cfg := sandboxForFile(t)
	fn := ProjectDocs(cfg)

	args, _ := json.Marshal(projectDocsArgs{Action: "list"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "erd: exists=false") {
		t.Errorf("expected erd exists=false, got: %s", out)
	}
	if !strings.Contains(out, "prd: exists=false") {
		t.Errorf("expected prd exists=false, got: %s", out)
	}
	if !strings.Contains(out, "decisions: exists=false") {
		t.Errorf("expected decisions exists=false, got: %s", out)
	}
}

func TestProjectDocsListWithFiles(t *testing.T) {
	cfg := sandboxForFile(t)
	wbDir := filepath.Join(cfg.BaseDir, ".waynebot")
	os.MkdirAll(wbDir, 0o755)
	os.WriteFile(filepath.Join(wbDir, "erd.md"), []byte("# ERD"), 0o644)

	fn := ProjectDocs(cfg)
	args, _ := json.Marshal(projectDocsArgs{Action: "list"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "erd: exists=true") {
		t.Errorf("expected erd exists=true, got: %s", out)
	}
	if !strings.Contains(out, "prd: exists=false") {
		t.Errorf("expected prd exists=false, got: %s", out)
	}
}

func TestProjectDocsReadExists(t *testing.T) {
	cfg := sandboxForFile(t)
	wbDir := filepath.Join(cfg.BaseDir, ".waynebot")
	os.MkdirAll(wbDir, 0o755)
	os.WriteFile(filepath.Join(wbDir, "erd.md"), []byte("Users -> Posts"), 0o644)

	fn := ProjectDocs(cfg)
	args, _ := json.Marshal(projectDocsArgs{Action: "read", DocType: "erd"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if out != "Users -> Posts" {
		t.Errorf("got %q, want %q", out, "Users -> Posts")
	}
}

func TestProjectDocsReadNotFound(t *testing.T) {
	cfg := sandboxForFile(t)
	fn := ProjectDocs(cfg)

	args, _ := json.Marshal(projectDocsArgs{Action: "read", DocType: "erd"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for missing document")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectDocsReadInvalidType(t *testing.T) {
	cfg := sandboxForFile(t)
	fn := ProjectDocs(cfg)

	args, _ := json.Marshal(projectDocsArgs{Action: "read", DocType: "bogus"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for invalid doc type")
	}
	if !strings.Contains(err.Error(), "unknown doc_type") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectDocsWriteCreatesDir(t *testing.T) {
	cfg := sandboxForFile(t)
	fn := ProjectDocs(cfg)

	args, _ := json.Marshal(projectDocsArgs{Action: "write", DocType: "erd", Content: "# New ERD"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "wrote") {
		t.Errorf("expected success message, got: %s", out)
	}

	data, err := os.ReadFile(filepath.Join(cfg.BaseDir, ".waynebot", "erd.md"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(data) != "# New ERD" {
		t.Errorf("file content = %q, want %q", string(data), "# New ERD")
	}
}

func TestProjectDocsWriteRejectsDecisions(t *testing.T) {
	cfg := sandboxForFile(t)
	fn := ProjectDocs(cfg)

	args, _ := json.Marshal(projectDocsArgs{Action: "write", DocType: "decisions", Content: "nope"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error when writing to decisions")
	}
	if !strings.Contains(err.Error(), "append") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectDocsAppendDecisions(t *testing.T) {
	cfg := sandboxForFile(t)
	fn := ProjectDocs(cfg)

	args, _ := json.Marshal(projectDocsArgs{Action: "append", DocType: "decisions", Content: "We chose SQLite."})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "appended") {
		t.Errorf("expected success message, got: %s", out)
	}

	data, err := os.ReadFile(filepath.Join(cfg.BaseDir, ".waynebot", "decisions.md"))
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
	cfg := sandboxForFile(t)
	wbDir := filepath.Join(cfg.BaseDir, ".waynebot")
	os.MkdirAll(wbDir, 0o755)
	os.WriteFile(filepath.Join(wbDir, "decisions.md"), []byte("## First\nFirst decision."), 0o644)

	fn := ProjectDocs(cfg)
	args, _ := json.Marshal(projectDocsArgs{Action: "append", DocType: "decisions", Content: "Second decision."})
	_, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(wbDir, "decisions.md"))
	content := string(data)
	if !strings.Contains(content, "First decision.") {
		t.Error("should preserve existing content")
	}
	if !strings.Contains(content, "Second decision.") {
		t.Error("should contain new decision")
	}
}

func TestProjectDocsAppendRejectsNonDecisions(t *testing.T) {
	cfg := sandboxForFile(t)
	fn := ProjectDocs(cfg)

	args, _ := json.Marshal(projectDocsArgs{Action: "append", DocType: "erd", Content: "stuff"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error when appending to non-decisions doc")
	}
	if !strings.Contains(err.Error(), "only supported for the decisions") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProjectDocsAppendEmptyContent(t *testing.T) {
	cfg := sandboxForFile(t)
	fn := ProjectDocs(cfg)

	args, _ := json.Marshal(projectDocsArgs{Action: "append", DocType: "decisions", Content: "  "})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestProjectDocsInvalidAction(t *testing.T) {
	cfg := sandboxForFile(t)
	fn := ProjectDocs(cfg)

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
	cfg := sandboxForFile(t)
	projectDir := t.TempDir()
	wbDir := filepath.Join(projectDir, ".waynebot")
	os.MkdirAll(wbDir, 0o755)
	os.WriteFile(filepath.Join(wbDir, "prd.md"), []byte("project PRD"), 0o644)

	fn := ProjectDocs(cfg)
	ctx := WithProjectDir(context.Background(), projectDir)
	args, _ := json.Marshal(projectDocsArgs{Action: "read", DocType: "prd"})

	out, err := fn(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	if out != "project PRD" {
		t.Errorf("got %q, want %q", out, "project PRD")
	}
}
