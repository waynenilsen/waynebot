package model_test

import (
	"database/sql"
	"math"
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func createTestPersona(t *testing.T, d interface {
	WriteExec(string, ...any) (sql.Result, error)
}) int64 {
	t.Helper()
	type writeExecer interface {
		WriteExec(string, ...any) (sql.Result, error)
	}
	res, err := d.(writeExecer).WriteExec(
		"INSERT INTO personas (name, system_prompt) VALUES (?, ?)",
		"testbot", "you are a test bot",
	)
	if err != nil {
		t.Fatalf("create test persona: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func TestCreateMemory(t *testing.T) {
	d := openTestDB(t)
	pid := createTestPersona(t, d)

	emb := []float32{0.1, 0.2, 0.3}
	m, err := model.CreateMemory(d, pid, nil, nil, "fact", "Go is great", emb, nil)
	if err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}
	if m.PersonaID != pid {
		t.Errorf("persona_id = %d, want %d", m.PersonaID, pid)
	}
	if m.Kind != "fact" {
		t.Errorf("kind = %q, want %q", m.Kind, "fact")
	}
	if m.Content != "Go is great" {
		t.Errorf("content = %q, want %q", m.Content, "Go is great")
	}
	if len(m.Embedding) != 3 {
		t.Errorf("embedding len = %d, want 3", len(m.Embedding))
	}
}

func TestCreateMemoryWithOptionalFields(t *testing.T) {
	d := openTestDB(t)
	pid := createTestPersona(t, d)

	chID := int64(42)
	m, err := model.CreateMemory(d, pid, &chID, nil, "preference", "prefers dark mode", nil, nil)
	if err != nil {
		t.Fatalf("CreateMemory: %v", err)
	}
	if m.ChannelID == nil || *m.ChannelID != 42 {
		t.Errorf("channel_id = %v, want 42", m.ChannelID)
	}
	if m.Embedding != nil {
		t.Errorf("embedding = %v, want nil", m.Embedding)
	}
}

func TestGetMemory(t *testing.T) {
	d := openTestDB(t)
	pid := createTestPersona(t, d)

	created, _ := model.CreateMemory(d, pid, nil, nil, "fact", "hello", []float32{1, 2}, nil)
	got, err := model.GetMemory(d, created.ID)
	if err != nil {
		t.Fatalf("GetMemory: %v", err)
	}
	if got.Content != "hello" {
		t.Errorf("content = %q, want %q", got.Content, "hello")
	}
}

func TestGetMemoryNotFound(t *testing.T) {
	d := openTestDB(t)
	_, err := model.GetMemory(d, 9999)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestListMemories(t *testing.T) {
	d := openTestDB(t)
	pid := createTestPersona(t, d)

	model.CreateMemory(d, pid, nil, nil, "fact", "one", nil, nil)
	model.CreateMemory(d, pid, nil, nil, "decision", "two", nil, nil)

	all, err := model.ListMemories(d, pid, model.MemoryFilter{})
	if err != nil {
		t.Fatalf("ListMemories: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("len = %d, want 2", len(all))
	}

	// Filter by kind
	facts, err := model.ListMemories(d, pid, model.MemoryFilter{Kind: "fact"})
	if err != nil {
		t.Fatalf("ListMemories(kind=fact): %v", err)
	}
	if len(facts) != 1 {
		t.Errorf("facts len = %d, want 1", len(facts))
	}
}

func TestListMemoriesFilterByChannel(t *testing.T) {
	d := openTestDB(t)
	pid := createTestPersona(t, d)

	ch := int64(10)
	model.CreateMemory(d, pid, &ch, nil, "fact", "in channel", nil, nil)
	model.CreateMemory(d, pid, nil, nil, "fact", "global", nil, nil)

	filtered, err := model.ListMemories(d, pid, model.MemoryFilter{ChannelID: &ch})
	if err != nil {
		t.Fatalf("ListMemories: %v", err)
	}
	if len(filtered) != 1 {
		t.Errorf("len = %d, want 1", len(filtered))
	}
	if filtered[0].Content != "in channel" {
		t.Errorf("content = %q, want %q", filtered[0].Content, "in channel")
	}
}

func TestUpdateMemory(t *testing.T) {
	d := openTestDB(t)
	pid := createTestPersona(t, d)

	m, _ := model.CreateMemory(d, pid, nil, nil, "fact", "old", []float32{1}, nil)
	if err := model.UpdateMemory(d, m.ID, "new", []float32{2, 3}); err != nil {
		t.Fatalf("UpdateMemory: %v", err)
	}
	got, _ := model.GetMemory(d, m.ID)
	if got.Content != "new" {
		t.Errorf("content = %q, want %q", got.Content, "new")
	}
	if len(got.Embedding) != 2 {
		t.Errorf("embedding len = %d, want 2", len(got.Embedding))
	}
}

func TestDeleteMemory(t *testing.T) {
	d := openTestDB(t)
	pid := createTestPersona(t, d)

	m, _ := model.CreateMemory(d, pid, nil, nil, "fact", "doomed", nil, nil)
	if err := model.DeleteMemory(d, m.ID); err != nil {
		t.Fatalf("DeleteMemory: %v", err)
	}
	_, err := model.GetMemory(d, m.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestEmbeddingSerializationRoundTrip(t *testing.T) {
	original := []float32{0.0, 1.0, -1.0, 0.5, math.MaxFloat32, math.SmallestNonzeroFloat32}
	blob := model.SerializeEmbedding(original)
	restored := model.DeserializeEmbedding(blob)

	if len(restored) != len(original) {
		t.Fatalf("len = %d, want %d", len(restored), len(original))
	}
	for i := range original {
		if restored[i] != original[i] {
			t.Errorf("[%d] = %v, want %v", i, restored[i], original[i])
		}
	}
}

func TestEmbeddingSerializationNil(t *testing.T) {
	if b := model.SerializeEmbedding(nil); b != nil {
		t.Errorf("expected nil for nil input, got %v", b)
	}
	if v := model.DeserializeEmbedding(nil); v != nil {
		t.Errorf("expected nil for nil input, got %v", v)
	}
}

func TestCosineSimilarity(t *testing.T) {
	// Identical vectors should give similarity of 1.0
	a := []float32{1, 2, 3}
	if s := model.CosineSimilarity(a, a); math.Abs(s-1.0) > 1e-6 {
		t.Errorf("identical vectors: similarity = %f, want 1.0", s)
	}

	// Orthogonal vectors should give similarity of 0.0
	x := []float32{1, 0, 0}
	y := []float32{0, 1, 0}
	if s := model.CosineSimilarity(x, y); math.Abs(s) > 1e-6 {
		t.Errorf("orthogonal vectors: similarity = %f, want 0.0", s)
	}

	// Opposite vectors should give similarity of -1.0
	neg := []float32{-1, -2, -3}
	if s := model.CosineSimilarity(a, neg); math.Abs(s+1.0) > 1e-6 {
		t.Errorf("opposite vectors: similarity = %f, want -1.0", s)
	}

	// Different length vectors should return 0
	if s := model.CosineSimilarity([]float32{1, 2}, []float32{1}); s != 0 {
		t.Errorf("different length: similarity = %f, want 0", s)
	}
}

func TestSearchMemories(t *testing.T) {
	d := openTestDB(t)
	pid := createTestPersona(t, d)

	// Create memories with known embeddings
	model.CreateMemory(d, pid, nil, nil, "fact", "close match", []float32{1, 0, 0}, nil)
	model.CreateMemory(d, pid, nil, nil, "fact", "medium match", []float32{0.7, 0.7, 0}, nil)
	model.CreateMemory(d, pid, nil, nil, "fact", "far match", []float32{0, 0, 1}, nil)
	model.CreateMemory(d, pid, nil, nil, "fact", "no embedding", nil, nil)

	query := []float32{1, 0, 0}
	results, err := model.SearchMemories(d, pid, query, 2, model.MemoryFilter{})
	if err != nil {
		t.Fatalf("SearchMemories: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len = %d, want 2", len(results))
	}
	if results[0].Content != "close match" {
		t.Errorf("first result = %q, want %q", results[0].Content, "close match")
	}
	if results[0].Score < results[1].Score {
		t.Error("results not sorted by score descending")
	}
}

func TestMigrationVersion10(t *testing.T) {
	d := openTestDB(t)
	v, err := d.SchemaVersion()
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if v < 10 {
		t.Errorf("schema version = %d, want >= 10", v)
	}

	// Verify the table exists by inserting directly
	pid := createTestPersona(t, d)
	_, err = d.WriteExec(
		"INSERT INTO memories (persona_id, kind, content) VALUES (?, 'fact', 'test')",
		pid,
	)
	if err != nil {
		t.Fatalf("direct insert into memories: %v", err)
	}
}
