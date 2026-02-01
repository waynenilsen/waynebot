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
	cfg := sandboxForFile(t)
	projectDir := t.TempDir()
	os.WriteFile(filepath.Join(projectDir, "proj.txt"), []byte("project content"), 0o644)

	fn := FileRead(cfg)
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

func TestFileReadFallsBackToSandbox(t *testing.T) {
	cfg := sandboxForFile(t)
	os.WriteFile(filepath.Join(cfg.BaseDir, "sandbox.txt"), []byte("sandbox content"), 0o644)

	fn := FileRead(cfg)
	args, _ := json.Marshal(fileReadArgs{Path: "sandbox.txt"})

	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if out != "sandbox content" {
		t.Fatalf("got %q, want %q", out, "sandbox content")
	}
}

func TestFileWriteUsesProjectDir(t *testing.T) {
	cfg := sandboxForFile(t)
	projectDir := t.TempDir()

	fn := FileWrite(cfg)
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

func TestFileWriteFallsBackToSandbox(t *testing.T) {
	cfg := sandboxForFile(t)

	fn := FileWrite(cfg)
	args, _ := json.Marshal(fileWriteArgs{Path: "out.txt", Content: "hello sandbox"})

	_, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(cfg.BaseDir, "out.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello sandbox" {
		t.Fatalf("got %q, want %q", string(data), "hello sandbox")
	}
}

func TestShellExecUsesProjectDir(t *testing.T) {
	cfg := sandboxForShell(t)
	cfg.AllowedCommands = append(cfg.AllowedCommands, "pwd")
	projectDir := t.TempDir()

	fn := ShellExec(cfg)
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

func TestShellExecFallsBackToSandbox(t *testing.T) {
	cfg := sandboxForShell(t)
	cfg.AllowedCommands = append(cfg.AllowedCommands, "pwd")

	fn := ShellExec(cfg)
	args, _ := json.Marshal(shellExecArgs{Command: "pwd"})

	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != cfg.BaseDir {
		t.Fatalf("got %q, want %q", strings.TrimSpace(out), cfg.BaseDir)
	}
}

func TestFileReadProjectDirPathTraversal(t *testing.T) {
	cfg := sandboxForFile(t)
	projectDir := t.TempDir()

	fn := FileRead(cfg)
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
