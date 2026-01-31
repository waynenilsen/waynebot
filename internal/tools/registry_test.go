package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func echoTool(_ context.Context, args json.RawMessage) (string, error) {
	return string(args), nil
}

func TestRegistryRegisterAndCall(t *testing.T) {
	r := NewRegistry()
	if err := r.Register("echo", echoTool); err != nil {
		t.Fatal(err)
	}

	out, err := r.Call(context.Background(), "echo", json.RawMessage(`"hello"`))
	if err != nil {
		t.Fatal(err)
	}
	if out != `"hello"` {
		t.Fatalf("got %q, want %q", out, `"hello"`)
	}
}

func TestRegistryDuplicateRegister(t *testing.T) {
	r := NewRegistry()
	if err := r.Register("echo", echoTool); err != nil {
		t.Fatal(err)
	}
	if err := r.Register("echo", echoTool); err == nil {
		t.Fatal("expected error on duplicate register")
	}
}

func TestRegistryCallUnknown(t *testing.T) {
	r := NewRegistry()
	_, err := r.Call(context.Background(), "nope", nil)
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestRegistryNames(t *testing.T) {
	r := NewRegistry()
	_ = r.Register("beta", echoTool)
	_ = r.Register("alpha", echoTool)

	names := r.Names()
	if len(names) != 2 {
		t.Fatalf("got %d names, want 2", len(names))
	}
}
