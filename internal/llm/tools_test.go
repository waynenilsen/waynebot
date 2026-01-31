package llm

import (
	"sort"
	"testing"
)

func TestToolsForPersonaAllTools(t *testing.T) {
	tools := ToolsForPersona([]string{"shell_exec", "file_read", "file_write", "http_fetch"})
	if len(tools) != 4 {
		t.Fatalf("got %d tools, want 4", len(tools))
	}
}

func TestToolsForPersonaSubset(t *testing.T) {
	tools := ToolsForPersona([]string{"shell_exec", "file_read"})
	if len(tools) != 2 {
		t.Fatalf("got %d tools, want 2", len(tools))
	}

	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Function.Name
	}
	sort.Strings(names)

	if names[0] != "file_read" || names[1] != "shell_exec" {
		t.Fatalf("got tools %v, want [file_read, shell_exec]", names)
	}
}

func TestToolsForPersonaEmpty(t *testing.T) {
	tools := ToolsForPersona(nil)
	if tools != nil {
		t.Fatalf("got %v, want nil", tools)
	}

	tools = ToolsForPersona([]string{})
	if tools != nil {
		t.Fatalf("got %v, want nil", tools)
	}
}

func TestToolsForPersonaUnknownTool(t *testing.T) {
	tools := ToolsForPersona([]string{"nonexistent", "shell_exec"})
	if len(tools) != 1 {
		t.Fatalf("got %d tools, want 1 (unknown tool skipped)", len(tools))
	}
	if tools[0].Function.Name != "shell_exec" {
		t.Fatalf("got tool %q, want shell_exec", tools[0].Function.Name)
	}
}

func TestAllToolNames(t *testing.T) {
	names := AllToolNames()
	if len(names) != 5 {
		t.Fatalf("got %d tool names, want 5", len(names))
	}

	sort.Strings(names)
	expected := []string{"file_read", "file_write", "http_fetch", "message_react", "shell_exec"}
	for i, name := range names {
		if name != expected[i] {
			t.Fatalf("got name %q at index %d, want %q", name, i, expected[i])
		}
	}
}

func TestToolDefinitionsHaveRequiredFields(t *testing.T) {
	for name, tool := range allTools {
		if tool.Function.Name == "" {
			t.Errorf("tool %q has empty function name", name)
		}
		if tool.Function.Parameters == nil {
			t.Errorf("tool %q has nil parameters", name)
		}
	}
}
