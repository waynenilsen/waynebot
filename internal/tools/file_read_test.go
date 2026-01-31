package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sandboxForFile(t *testing.T) *SandboxConfig {
	t.Helper()
	return &SandboxConfig{
		BaseDir:      t.TempDir(),
		BlockedHosts: DefaultBlockedHosts(),
	}
}

func TestFileReadSuccess(t *testing.T) {
	cfg := sandboxForFile(t)
	os.WriteFile(filepath.Join(cfg.BaseDir, "hello.txt"), []byte("world"), 0o644)

	fn := FileRead(cfg)
	args, _ := json.Marshal(fileReadArgs{Path: "hello.txt"})

	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if out != "world" {
		t.Fatalf("got %q, want %q", out, "world")
	}
}

func TestFileReadPathTraversal(t *testing.T) {
	cfg := sandboxForFile(t)
	fn := FileRead(cfg)

	args, _ := json.Marshal(fileReadArgs{Path: "../../etc/passwd"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
	if !strings.Contains(err.Error(), "escape") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileReadEmptyPath(t *testing.T) {
	cfg := sandboxForFile(t)
	fn := FileRead(cfg)

	args, _ := json.Marshal(fileReadArgs{})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestFileReadDirectory(t *testing.T) {
	cfg := sandboxForFile(t)
	os.MkdirAll(filepath.Join(cfg.BaseDir, "subdir"), 0o755)
	fn := FileRead(cfg)

	args, _ := json.Marshal(fileReadArgs{Path: "subdir"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for directory")
	}
}

func TestFileReadTooLarge(t *testing.T) {
	cfg := sandboxForFile(t)
	// Create a file just over 1MB.
	big := make([]byte, maxFileReadSize+1)
	os.WriteFile(filepath.Join(cfg.BaseDir, "big.bin"), big, 0o644)
	fn := FileRead(cfg)

	args, _ := json.Marshal(fileReadArgs{Path: "big.bin"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Fatalf("unexpected error: %v", err)
	}
}
