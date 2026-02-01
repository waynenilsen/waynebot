package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileReadUsesProjectDir(t *testing.T) {
	baseDir := t.TempDir()
	projectDir := t.TempDir()
	os.WriteFile(filepath.Join(projectDir, "proj.txt"), []byte("project content"), 0o644)

	fn := FileRead(baseDir)
	ctx := WithProjectDir(context.Background(), projectDir)
	args, _ := json.Marshal(fileReadArgs{Path: "proj.txt"})

	out, err := fn(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	if out != "project content" {
		t.Fatalf("got %q, want %q", out, "project content")
	}
}

func TestFileReadFallsBackToBaseDir(t *testing.T) {
	baseDir := t.TempDir()
	os.WriteFile(filepath.Join(baseDir, "base.txt"), []byte("base content"), 0o644)

	fn := FileRead(baseDir)
	args, _ := json.Marshal(fileReadArgs{Path: "base.txt"})

	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if out != "base content" {
		t.Fatalf("got %q, want %q", out, "base content")
	}
}

func TestFileWriteUsesProjectDir(t *testing.T) {
	baseDir := t.TempDir()
	projectDir := t.TempDir()

	fn := FileWrite(baseDir)
	ctx := WithProjectDir(context.Background(), projectDir)
	args, _ := json.Marshal(fileWriteArgs{Path: "out.txt", Content: "hello project"})

	_, err := fn(ctx, args)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, "out.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello project" {
		t.Fatalf("got %q, want %q", string(data), "hello project")
	}
}

func TestFileWriteFallsBackToBaseDir(t *testing.T) {
	baseDir := t.TempDir()

	fn := FileWrite(baseDir)
	args, _ := json.Marshal(fileWriteArgs{Path: "out.txt", Content: "hello base"})

	_, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(baseDir, "out.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello base" {
		t.Fatalf("got %q, want %q", string(data), "hello base")
	}
}

func TestShellExecUsesProjectDir(t *testing.T) {
	baseDir := t.TempDir()
	projectDir := t.TempDir()

	fn := ShellExec(baseDir)
	ctx := WithProjectDir(context.Background(), projectDir)
	args, _ := json.Marshal(shellExecArgs{Command: "pwd"})

	out, err := fn(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != projectDir {
		t.Fatalf("got %q, want %q", strings.TrimSpace(out), projectDir)
	}
}

func TestShellExecFallsBackToBaseDir(t *testing.T) {
	baseDir := t.TempDir()

	fn := ShellExec(baseDir)
	args, _ := json.Marshal(shellExecArgs{Command: "pwd"})

	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != baseDir {
		t.Fatalf("got %q, want %q", strings.TrimSpace(out), baseDir)
	}
}

func TestFileReadProjectDirPathTraversal(t *testing.T) {
	baseDir := t.TempDir()
	projectDir := t.TempDir()

	fn := FileRead(baseDir)
	ctx := WithProjectDir(context.Background(), projectDir)
	args, _ := json.Marshal(fileReadArgs{Path: "../../etc/passwd"})

	_, err := fn(ctx, args)
	if err == nil {
		t.Fatal("expected error for path traversal with project dir")
	}
	if !strings.Contains(err.Error(), "escape") {
		t.Fatalf("unexpected error: %v", err)
	}
}
