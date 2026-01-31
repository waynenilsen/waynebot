package llm

import (
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestBuildMessagesSystemPrompt(t *testing.T) {
	persona := model.Persona{
		SystemPrompt: "You are a helpful bot.",
	}
	msgs := BuildMessages(persona, nil)

	if len(msgs) != 1 {
		t.Fatalf("got %d messages, want 1", len(msgs))
	}
	if msgs[0].OfSystem == nil {
		t.Fatal("expected system message")
	}
}

func TestBuildMessagesEmptySystemPrompt(t *testing.T) {
	persona := model.Persona{SystemPrompt: ""}
	msgs := BuildMessages(persona, nil)

	if len(msgs) != 0 {
		t.Fatalf("got %d messages, want 0 (no system prompt)", len(msgs))
	}
}

func TestBuildMessagesHumanMessages(t *testing.T) {
	persona := model.Persona{SystemPrompt: "system"}
	history := []model.Message{
		{AuthorType: "human", AuthorName: "alice", Content: "hello"},
		{AuthorType: "human", AuthorName: "bob", Content: "world"},
	}
	msgs := BuildMessages(persona, history)

	if len(msgs) != 3 {
		t.Fatalf("got %d messages, want 3", len(msgs))
	}
	// First is system
	if msgs[0].OfSystem == nil {
		t.Fatal("expected system message first")
	}
	// Second and third are user messages
	if msgs[1].OfUser == nil {
		t.Fatal("expected user message")
	}
	if msgs[2].OfUser == nil {
		t.Fatal("expected user message")
	}
}

func TestBuildMessagesAgentMessages(t *testing.T) {
	persona := model.Persona{SystemPrompt: "system"}
	history := []model.Message{
		{AuthorType: "agent", AuthorName: "bot", Content: "I can help with that."},
	}
	msgs := BuildMessages(persona, history)

	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2", len(msgs))
	}
	if msgs[1].OfAssistant == nil {
		t.Fatal("expected assistant message")
	}
}

func TestBuildMessagesToolCallAndResult(t *testing.T) {
	persona := model.Persona{SystemPrompt: "system"}
	history := []model.Message{
		{AuthorType: "tool_call", AuthorName: "call_123", Content: "shell_exec\n{\"command\":\"ls\"}"},
		{AuthorType: "tool_result", AuthorName: "call_123", Content: "file1.txt\nfile2.txt"},
	}
	msgs := BuildMessages(persona, history)

	if len(msgs) != 3 {
		t.Fatalf("got %d messages, want 3", len(msgs))
	}

	// Tool call becomes assistant message with tool calls
	if msgs[1].OfAssistant == nil {
		t.Fatal("expected assistant message for tool_call")
	}
	if len(msgs[1].OfAssistant.ToolCalls) != 1 {
		t.Fatalf("got %d tool calls, want 1", len(msgs[1].OfAssistant.ToolCalls))
	}
	tc := msgs[1].OfAssistant.ToolCalls[0]
	if tc.ID != "call_123" {
		t.Fatalf("got tool call ID %q, want %q", tc.ID, "call_123")
	}
	if tc.Function.Name != "shell_exec" {
		t.Fatalf("got tool name %q, want %q", tc.Function.Name, "shell_exec")
	}
	if tc.Function.Arguments != `{"command":"ls"}` {
		t.Fatalf("got arguments %q", tc.Function.Arguments)
	}

	// Tool result becomes tool message
	if msgs[2].OfTool == nil {
		t.Fatal("expected tool message for tool_result")
	}
}

func TestBuildMessagesDefaultFallback(t *testing.T) {
	persona := model.Persona{SystemPrompt: "system"}
	history := []model.Message{
		{AuthorType: "unknown", Content: "some content"},
	}
	msgs := BuildMessages(persona, history)

	if len(msgs) != 2 {
		t.Fatalf("got %d messages, want 2", len(msgs))
	}
	if msgs[1].OfUser == nil {
		t.Fatal("expected user message for unknown author type")
	}
}

func TestSplitToolCall(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantArgs string
	}{
		{"shell_exec\n{\"command\":\"ls\"}", "shell_exec", "{\"command\":\"ls\"}"},
		{"file_read\n{\"path\":\"foo.txt\"}", "file_read", "{\"path\":\"foo.txt\"}"},
		{"no_newline", "no_newline", "{}"},
		{"\nempty_name", "", "empty_name"},
	}

	for _, tt := range tests {
		name, args := splitToolCall(tt.input)
		if name != tt.wantName {
			t.Errorf("splitToolCall(%q) name = %q, want %q", tt.input, name, tt.wantName)
		}
		if args != tt.wantArgs {
			t.Errorf("splitToolCall(%q) args = %q, want %q", tt.input, args, tt.wantArgs)
		}
	}
}
