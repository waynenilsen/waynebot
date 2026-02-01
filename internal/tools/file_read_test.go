package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileReadSuccess(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("world"), 0o644)

	fn := FileRead(dir)
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
	fn := FileRead(t.TempDir())

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
	fn := FileRead(t.TempDir())

	args, _ := json.Marshal(fileReadArgs{})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestFileReadDirectory(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "subdir"), 0o755)
	fn := FileRead(dir)

	args, _ := json.Marshal(fileReadArgs{Path: "subdir"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for directory")
	}
}

func TestFileReadTooLarge(t *testing.T) {
	dir := t.TempDir()
	big := make([]byte, maxFileReadSize+1)
	os.WriteFile(filepath.Join(dir, "big.bin"), big, 0o644)
	fn := FileRead(dir)

	args, _ := json.Marshal(fileReadArgs{Path: "big.bin"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Fatalf("unexpected error: %v", err)
	}
}
