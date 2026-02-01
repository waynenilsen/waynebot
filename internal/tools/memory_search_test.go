package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// stubEmbedder returns a fixed vector for any input.
type stubEmbedder struct {
	vec []float32
	err error
}

func (s *stubEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	return s.vec, s.err
}

func openMemoryTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func createMemoryTestPersona(t *testing.T, d *db.DB) int64 {
	t.Helper()
	res, err := d.WriteExec(
		"INSERT INTO personas (name, system_prompt) VALUES (?, ?)",
		"testbot", "you are a test bot",
	)
	if err != nil {
		t.Fatalf("create test persona: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func TestMemorySearchReturnsResults(t *testing.T) {
	d := openMemoryTestDB(t)
	pid := createMemoryTestPersona(t, d)

	model.CreateMemory(d, pid, nil, nil, "fact", "Go is great", []float32{1, 0, 0}, nil)
	model.CreateMemory(d, pid, nil, nil, "decision", "Use SQLite", []float32{0.9, 0.1, 0}, nil)
	model.CreateMemory(d, pid, nil, nil, "preference", "Dark mode", []float32{0, 0, 1}, nil)

	embedder := &stubEmbedder{vec: []float32{1, 0, 0}}
	fn := MemorySearch(d, embedder)

	ctx := WithPersonaID(context.Background(), pid)
	args, _ := json.Marshal(memorySearchArgs{Query: "golang"})
	out, err := fn(ctx, args)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "Found") {
		t.Errorf("expected Found header, got: %s", out)
	}
	if !strings.Contains(out, "Go is great") {
		t.Errorf("expected top result 'Go is great', got: %s", out)
	}
	if !strings.Contains(out, "[fact]") {
		t.Errorf("expected kind label, got: %s", out)
	}
	if !strings.Contains(out, "score:") {
		t.Errorf("expected score, got: %s", out)
	}
}

func TestMemorySearchKindFilter(t *testing.T) {
	d := openMemoryTestDB(t)
	pid := createMemoryTestPersona(t, d)

	model.CreateMemory(d, pid, nil, nil, "fact", "Go is great", []float32{1, 0, 0}, nil)
	model.CreateMemory(d, pid, nil, nil, "decision", "Use SQLite", []float32{1, 0, 0}, nil)

	embedder := &stubEmbedder{vec: []float32{1, 0, 0}}
	fn := MemorySearch(d, embedder)

	ctx := WithPersonaID(context.Background(), pid)
	args, _ := json.Marshal(memorySearchArgs{Query: "anything", Kind: "decision"})
	out, err := fn(ctx, args)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "Use SQLite") {
		t.Errorf("expected decision memory, got: %s", out)
	}
	if strings.Contains(out, "Go is great") {
		t.Errorf("fact should be filtered out, got: %s", out)
	}
}

func TestMemorySearchEmptyResults(t *testing.T) {
	d := openMemoryTestDB(t)
	pid := createMemoryTestPersona(t, d)

	embedder := &stubEmbedder{vec: []float32{1, 0, 0}}
	fn := MemorySearch(d, embedder)

	ctx := WithPersonaID(context.Background(), pid)
	args, _ := json.Marshal(memorySearchArgs{Query: "anything"})
	out, err := fn(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	if out != "No relevant memories found." {
		t.Errorf("expected no results message, got: %s", out)
	}
}

func TestMemorySearchNoPersonaID(t *testing.T) {
	d := openMemoryTestDB(t)
	embedder := &stubEmbedder{vec: []float32{1, 0, 0}}
	fn := MemorySearch(d, embedder)

	args, _ := json.Marshal(memorySearchArgs{Query: "anything"})
	_, err := fn(context.Background(), args)
	if err == nil {
		t.Fatal("expected error without persona_id")
	}
	if !strings.Contains(err.Error(), "persona_id") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMemorySearchEmptyQuery(t *testing.T) {
	d := openMemoryTestDB(t)
	embedder := &stubEmbedder{vec: []float32{1, 0, 0}}
	fn := MemorySearch(d, embedder)

	ctx := WithPersonaID(context.Background(), 1)
	args, _ := json.Marshal(memorySearchArgs{Query: "  "})
	_, err := fn(ctx, args)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
	if !strings.Contains(err.Error(), "query is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMemorySearchLimit(t *testing.T) {
	d := openMemoryTestDB(t)
	pid := createMemoryTestPersona(t, d)

	for i := range 5 {
		vec := []float32{1, float32(i) * 0.01, 0}
		model.CreateMemory(d, pid, nil, nil, "fact", "memory "+string(rune('A'+i)), vec, nil)
	}

	embedder := &stubEmbedder{vec: []float32{1, 0, 0}}
	fn := MemorySearch(d, embedder)

	ctx := WithPersonaID(context.Background(), pid)
	args, _ := json.Marshal(memorySearchArgs{Query: "test", Limit: 2})
	out, err := fn(ctx, args)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Found 2") {
		t.Errorf("expected 2 results, got: %s", out)
	}
}

func TestMemorySearchEmbedError(t *testing.T) {
	d := openMemoryTestDB(t)
	pid := createMemoryTestPersona(t, d)

	embedder := &stubEmbedder{err: sql.ErrConnDone}
	fn := MemorySearch(d, embedder)

	ctx := WithPersonaID(context.Background(), pid)
	args, _ := json.Marshal(memorySearchArgs{Query: "test"})
	_, err := fn(ctx, args)
	if err == nil {
		t.Fatal("expected error from embedder")
	}
	if !strings.Contains(err.Error(), "embed query") {
		t.Errorf("unexpected error: %v", err)
	}
}
