package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestShellExecEcho(t *testing.T) {
	fn := ShellExec(t.TempDir())

	args, _ := json.Marshal(shellExecArgs{Command: "echo", Args: []string{"hello"}})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "hello" {
		t.Fatalf("got %q, want %q", out, "hello")
	}
}

func TestShellExecEmptyCommand(t *testing.T) {
	fn := ShellExec(t.TempDir())

	args, _ := json.Marshal(shellExecArgs{})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestShellExecOutputCap(t *testing.T) {
	fn := ShellExec(t.TempDir())

	// Generate output larger than 10KB.
	args, _ := json.Marshal(shellExecArgs{
		Command: "dd",
		Args:    []string{"if=/dev/zero", "bs=1024", "count=20", "status=none"},
	})
	out, _ := fn(context.Background(), args)
	if len(out) > shellOutputCap+100 {
		t.Fatalf("output not capped: got %d bytes", len(out))
	}
}

func TestShellExecFullPath(t *testing.T) {
	fn := ShellExec(t.TempDir())

	args, _ := json.Marshal(shellExecArgs{Command: "/bin/echo", Args: []string{"hi"}})
	out, err := fn(context.Background(), args)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "hi" {
		t.Fatalf("got %q", out)
	}
}
