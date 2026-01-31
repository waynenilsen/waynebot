package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func TestChatCompletionTextResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		if body["model"] != "test-model" {
			t.Fatalf("unexpected model: %v", body["model"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":      "chatcmpl-123",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "test-model",
			"choices": []map[string]any{
				{
					"index":         0,
					"finish_reason": "stop",
					"message": map[string]any{
						"role":    "assistant",
						"content": "Hello from the mock!",
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     10,
				"completion_tokens": 5,
				"total_tokens":      15,
			},
		})
	}))
	defer srv.Close()

	client := NewClientWithOptions(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(srv.URL),
	)

	resp, err := client.ChatCompletion(
		context.Background(),
		"test-model",
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Hi"),
		},
		nil,
		0.7,
		100,
	)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Content != "Hello from the mock!" {
		t.Fatalf("got content %q, want %q", resp.Content, "Hello from the mock!")
	}
	if resp.PromptTokens != 10 {
		t.Fatalf("got prompt tokens %d, want 10", resp.PromptTokens)
	}
	if resp.CompletionTokens != 5 {
		t.Fatalf("got completion tokens %d, want 5", resp.CompletionTokens)
	}
	if len(resp.ToolCalls) != 0 {
		t.Fatalf("got %d tool calls, want 0", len(resp.ToolCalls))
	}
}

func TestChatCompletionToolCallResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":      "chatcmpl-456",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "test-model",
			"choices": []map[string]any{
				{
					"index":         0,
					"finish_reason": "tool_calls",
					"message": map[string]any{
						"role":    "assistant",
						"content": "",
						"tool_calls": []map[string]any{
							{
								"id":   "call_abc",
								"type": "function",
								"function": map[string]any{
									"name":      "shell_exec",
									"arguments": `{"command":"ls","args":["-la"]}`,
								},
							},
						},
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     20,
				"completion_tokens": 15,
				"total_tokens":      35,
			},
		})
	}))
	defer srv.Close()

	client := NewClientWithOptions(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(srv.URL),
	)

	resp, err := client.ChatCompletion(
		context.Background(),
		"test-model",
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("list files"),
		},
		ToolsForPersona([]string{"shell_exec"}),
		0.7,
		100,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.ToolCalls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(resp.ToolCalls))
	}
	tc := resp.ToolCalls[0]
	if tc.ID != "call_abc" {
		t.Fatalf("got tool call ID %q, want %q", tc.ID, "call_abc")
	}
	if tc.Name != "shell_exec" {
		t.Fatalf("got tool name %q, want %q", tc.Name, "shell_exec")
	}
	if tc.Arguments != `{"command":"ls","args":["-la"]}` {
		t.Fatalf("got arguments %q", tc.Arguments)
	}
	if resp.PromptTokens != 20 {
		t.Fatalf("got prompt tokens %d, want 20", resp.PromptTokens)
	}
}

func TestChatCompletionNoChoicesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":      "chatcmpl-789",
			"object":  "chat.completion",
			"created": 1234567890,
			"model":   "test-model",
			"choices": []map[string]any{},
			"usage": map[string]any{
				"prompt_tokens":     0,
				"completion_tokens": 0,
				"total_tokens":      0,
			},
		})
	}))
	defer srv.Close()

	client := NewClientWithOptions(
		option.WithAPIKey("test-key"),
		option.WithBaseURL(srv.URL),
	)

	_, err := client.ChatCompletion(
		context.Background(),
		"test-model",
		[]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("test"),
		},
		nil,
		0.7,
		100,
	)
	if err == nil {
		t.Fatal("expected error for no choices")
	}
}
