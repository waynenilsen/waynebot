package model

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

type Memory struct {
	ID              int64
	PersonaID       int64
	ChannelID       *int64
	ProjectID       *int64
	Kind            string
	Content         string
	Embedding       []float32
	SourceMessageID *int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ScoredMemory struct {
	Memory
	Score float64
}

// MemoryFilter controls which memories are returned by ListMemories and SearchMemories.
type MemoryFilter struct {
	ChannelID *int64
	ProjectID *int64
	Kind      string
}

const memoryCols = "id, persona_id, channel_id, project_id, kind, content, embedding, source_message_id, created_at, updated_at"

func scanMemory(s interface{ Scan(...any) error }) (Memory, error) {
	var m Memory
	var embBlob []byte
	err := s.Scan(&m.ID, &m.PersonaID, &m.ChannelID, &m.ProjectID, &m.Kind, &m.Content, &embBlob, &m.SourceMessageID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return m, err
	}
	m.Embedding = DeserializeEmbedding(embBlob)
	return m, nil
}

func CreateMemory(d *db.DB, personaID int64, channelID, projectID *int64, kind, content string, embedding []float32, sourceMessageID *int64) (Memory, error) {
	embBlob := SerializeEmbedding(embedding)
	res, err := d.WriteExec(
		"INSERT INTO memories (persona_id, channel_id, project_id, kind, content, embedding, source_message_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		personaID, channelID, projectID, kind, content, embBlob, sourceMessageID,
	)
	if err != nil {
		return Memory{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Memory{}, err
	}
	return scanMemory(d.SQL.QueryRow(
		"SELECT "+memoryCols+" FROM memories WHERE id = ?", id,
	))
}

func GetMemory(d *db.DB, id int64) (Memory, error) {
	return scanMemory(d.SQL.QueryRow(
		"SELECT "+memoryCols+" FROM memories WHERE id = ?", id,
	))
}

func ListMemories(d *db.DB, personaID int64, filter MemoryFilter) ([]Memory, error) {
	query := "SELECT " + memoryCols + " FROM memories WHERE persona_id = ?"
	args := []any{personaID}

	if filter.ChannelID != nil {
		query += " AND channel_id = ?"
		args = append(args, *filter.ChannelID)
	}
	if filter.ProjectID != nil {
		query += " AND project_id = ?"
		args = append(args, *filter.ProjectID)
	}
	if filter.Kind != "" {
		query += " AND kind = ?"
		args = append(args, filter.Kind)
	}

	query += " ORDER BY updated_at DESC"

	rows, err := d.SQL.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		m, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, m)
	}
	return memories, rows.Err()
}

func UpdateMemory(d *db.DB, id int64, content string, embedding []float32) error {
	embBlob := SerializeEmbedding(embedding)
	_, err := d.WriteExec(
		"UPDATE memories SET content = ?, embedding = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		content, embBlob, id,
	)
	return err
}

func DeleteMemory(d *db.DB, id int64) error {
	_, err := d.WriteExec("DELETE FROM memories WHERE id = ?", id)
	return err
}

// SearchMemories loads candidate memories from the DB, computes cosine similarity
// against queryEmbedding in Go, and returns the top-N results sorted by score.
func SearchMemories(d *db.DB, personaID int64, queryEmbedding []float32, limit int, filter MemoryFilter) ([]ScoredMemory, error) {
	candidates, err := ListMemories(d, personaID, filter)
	if err != nil {
		return nil, fmt.Errorf("list memories: %w", err)
	}

	var scored []ScoredMemory
	for _, m := range candidates {
		if len(m.Embedding) == 0 {
			continue
		}
		score := CosineSimilarity(queryEmbedding, m.Embedding)
		scored = append(scored, ScoredMemory{Memory: m, Score: score})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if limit > 0 && len(scored) > limit {
		scored = scored[:limit]
	}
	return scored, nil
}

// SerializeEmbedding converts a float32 slice to bytes for BLOB storage.
func SerializeEmbedding(v []float32) []byte {
	if v == nil {
		return nil
	}
	buf := make([]byte, len(v)*4)
	for i, f := range v {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf
}

// DeserializeEmbedding converts bytes back to a float32 slice.
func DeserializeEmbedding(b []byte) []float32 {
	if b == nil || len(b) == 0 {
		return nil
	}
	n := len(b) / 4
	v := make([]float32, n)
	for i := range n {
		v[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return v
}

// CosineSimilarity computes the cosine similarity between two vectors.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		ai, bi := float64(a[i]), float64(b[i])
		dot += ai * bi
		normA += ai * ai
		normB += bi * bi
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
