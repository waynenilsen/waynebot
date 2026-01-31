package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

// Client wraps the openai-go SDK configured for OpenRouter.
type Client struct {
	client openai.Client
}

// Response holds the result of a chat completion call.
type Response struct {
	Content          string
	ToolCalls        []ToolCall
	PromptTokens     int
	CompletionTokens int
}

// ToolCall represents a single tool invocation requested by the model.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string // raw JSON
}

// NewClient creates an LLM client pointed at OpenRouter.
func NewClient(apiKey string) *Client {
	c := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://openrouter.ai/api/v1"),
	)
	return &Client{client: c}
}

// NewClientWithOptions creates an LLM client with custom options, useful for testing.
func NewClientWithOptions(opts ...option.RequestOption) *Client {
	c := openai.NewClient(opts...)
	return &Client{client: c}
}

// ChatCompletion sends a chat completion request and returns a structured response.
func (c *Client) ChatCompletion(
	ctx context.Context,
	model string,
	messages []openai.ChatCompletionMessageParamUnion,
	tools []openai.ChatCompletionToolParam,
	temperature float64,
	maxTokens int,
) (Response, error) {
	params := openai.ChatCompletionNewParams{
		Model:       shared.ChatModel(model),
		Messages:    messages,
		Temperature: param.NewOpt(temperature),
		MaxTokens:   param.NewOpt(int64(maxTokens)),
	}
	if len(tools) > 0 {
		params.Tools = tools
	}

	completion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return Response{}, fmt.Errorf("chat completion: %w", err)
	}

	if len(completion.Choices) == 0 {
		return Response{}, fmt.Errorf("chat completion: no choices returned")
	}

	choice := completion.Choices[0]
	resp := Response{
		Content:          choice.Message.Content,
		PromptTokens:     int(completion.Usage.PromptTokens),
		CompletionTokens: int(completion.Usage.CompletionTokens),
	}

	for _, tc := range choice.Message.ToolCalls {
		resp.ToolCalls = append(resp.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return resp, nil
}
