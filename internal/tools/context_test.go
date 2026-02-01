package tools

import (
	"context"
	"testing"
)

func TestWithProjectDir(t *testing.T) {
	ctx := WithProjectDir(context.Background(), "/tmp/myproject")
	got := ProjectDirFromContext(ctx)
	if got != "/tmp/myproject" {
		t.Fatalf("got %q, want %q", got, "/tmp/myproject")
	}
}

func TestProjectDirFromContextEmpty(t *testing.T) {
	got := ProjectDirFromContext(context.Background())
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}
