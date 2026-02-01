package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// Embedder can embed a text string into a vector.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

type memorySearchArgs struct {
	Query string `json:"query"`
	Kind  string `json:"kind,omitempty"`
	Limit int    `json:"limit,omitempty"`
}

// MemorySearch returns a ToolFunc that searches the calling persona's memories
// using semantic similarity. The persona ID is extracted from context.
func MemorySearch(database *db.DB, embedder Embedder) ToolFunc {
	return func(ctx context.Context, raw json.RawMessage) (string, error) {
		personaID := PersonaIDFromContext(ctx)
		if personaID == 0 {
			return "", fmt.Errorf("persona_id not set in context")
		}

		var args memorySearchArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}
		if strings.TrimSpace(args.Query) == "" {
			return "", fmt.Errorf("query is required")
		}

		limit := args.Limit
		if limit <= 0 {
			limit = 10
		}

		queryEmbedding, err := embedder.Embed(ctx, args.Query)
		if err != nil {
			return "", fmt.Errorf("embed query: %w", err)
		}

		filter := model.MemoryFilter{}
		if args.Kind != "" {
			filter.Kind = args.Kind
		}

		results, err := model.SearchMemories(database, personaID, queryEmbedding, limit, filter)
		if err != nil {
			return "", fmt.Errorf("search memories: %w", err)
		}

		if len(results) == 0 {
			return "No relevant memories found.", nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Found %d relevant memories:\n", len(results)))
		for i, m := range results {
			sb.WriteString(fmt.Sprintf("\n%d. [%s] (score: %.2f) %s\n",
				i+1, m.Kind, m.Score, m.Content))
			sb.WriteString(fmt.Sprintf("   — %s\n", m.CreatedAt.Format("2006-01-02")))
		}
		return sb.String(), nil
	}
}
