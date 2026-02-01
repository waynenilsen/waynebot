package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileWriteSuccess(t *testing.T) {
	dir := t.TempDir()
	fn := FileWrite(dir)

	args, _ := json.Marshal(fileWriteArgs{Path: "out.txt", Content: "hello"})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "5 bytes") {
		t.Fatalf("unexpected result: %s", out)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "out.txt"))
	if string(data) != "hello" {
		t.Fatalf("file content = %q, want %q", data, "hello")
	}
}

func TestFileWriteCreatesSubdirs(t *testing.T) {
	dir := t.TempDir()
	fn := FileWrite(dir)

	args, _ := json.Marshal(fileWriteArgs{Path: "a/b/c.txt", Content: "nested"})
	_, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, "a", "b", "c.txt"))
	if string(data) != "nested" {
		t.Fatalf("file content = %q", data)
	}
}

func TestFileWritePathTraversal(t *testing.T) {
	fn := FileWrite(t.TempDir())

	args, _ := json.Marshal(fileWriteArgs{Path: "../../evil.txt", Content: "bad"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestFileWriteTooLarge(t *testing.T) {
	fn := FileWrite(t.TempDir())

	big := strings.Repeat("x", maxFileWriteSize+1)
	args, _ := json.Marshal(fileWriteArgs{Path: "big.txt", Content: big})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for oversized content")
	}
}

func TestFileWriteEmptyPath(t *testing.T) {
	fn := FileWrite(t.TempDir())

	args, _ := json.Marshal(fileWriteArgs{Content: "hello"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}
