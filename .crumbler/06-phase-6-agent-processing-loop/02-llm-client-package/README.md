# LLM Client Package

**Goal:** Create `internal/llm/` with an OpenRouter-compatible LLM client using the openai-go SDK.

## Dependencies

Add `github.com/openai/openai-go` to go.mod. This SDK supports setting a custom base URL for OpenRouter.

## Files to Create

- `internal/llm/client.go` — Client struct wrapping openai-go with OpenRouter base URL (`https://openrouter.ai/api/v1`). Method: `ChatCompletion(ctx, model, messages, tools, temperature, maxTokens) (Response, error)`. Response includes content, tool_calls, prompt_tokens, completion_tokens.
- `internal/llm/client_test.go` — Tests with httptest mock server that returns canned responses
- `internal/llm/messages.go` — Convert domain `model.Message` slice to openai SDK message types. Build system message from persona.SystemPrompt. Handle tool call/result messages in the conversation.
- `internal/llm/messages_test.go`
- `internal/llm/tools.go` — Tool JSON schema definitions for shell_exec, file_read, file_write, http_fetch. Returns `[]openai.ChatCompletionToolParam` filtered by persona's tools_enabled list.
- `internal/llm/tools_test.go`

## Key Types

```go
type Client struct {
    client *openai.Client
}

type Response struct {
    Content          string
    ToolCalls        []ToolCall
    PromptTokens     int
    CompletionTokens int
}

type ToolCall struct {
    ID        string
    Name      string
    Arguments string // raw JSON
}
```

## Config

Uses `WAYNEBOT_OPENROUTER_KEY` from config (already exists in config.go).

## Verification

`go test ./internal/llm/...` passes.
