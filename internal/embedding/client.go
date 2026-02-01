package embedding

import (
	"context"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const (
	Model = "openai/text-embedding-3-small"
	Dims  = 1536
)

// Client calls an OpenAI-compatible embeddings API via OpenRouter.
type Client struct {
	oa openai.Client
}

// NewClient creates an embedding client pointed at OpenRouter.
func NewClient(apiKey string) *Client {
	c := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://openrouter.ai/api/v1"),
	)
	return &Client{oa: c}
}

// Embed returns the embedding vector for a single text.
func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	batch, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return batch[0], nil
}

// EmbedBatch returns embedding vectors for multiple texts.
func (c *Client) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	resp, err := c.oa.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: Model,
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: texts,
		},
	})
	if err != nil {
		return nil, err
	}

	results := make([][]float32, len(texts))
	for _, item := range resp.Data {
		vec := make([]float32, len(item.Embedding))
		for j, v := range item.Embedding {
			vec[j] = float32(v)
		}
		results[item.Index] = vec
	}
	return results, nil
}
