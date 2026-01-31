package llm

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

// allTools maps tool names to their OpenAI function definitions.
var allTools = map[string]openai.ChatCompletionToolParam{
	"shell_exec": {
		Function: shared.FunctionDefinitionParam{
			Name:        "shell_exec",
			Description: param.NewOpt("Execute a shell command with arguments inside the sandbox."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "The command to execute.",
					},
					"args": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Arguments to pass to the command.",
					},
				},
				"required": []string{"command"},
			},
		},
	},
	"file_read": {
		Function: shared.FunctionDefinitionParam{
			Name:        "file_read",
			Description: param.NewOpt("Read the contents of a file inside the sandbox."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Relative path to the file to read.",
					},
				},
				"required": []string{"path"},
			},
		},
	},
	"file_write": {
		Function: shared.FunctionDefinitionParam{
			Name:        "file_write",
			Description: param.NewOpt("Write content to a file inside the sandbox."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Relative path to the file to write.",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "Content to write to the file.",
					},
				},
				"required": []string{"path", "content"},
			},
		},
	},
	"http_fetch": {
		Function: shared.FunctionDefinitionParam{
			Name:        "http_fetch",
			Description: param.NewOpt("Fetch a URL via HTTP."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "The URL to fetch.",
					},
					"method": map[string]any{
						"type":        "string",
						"description": "HTTP method (GET, POST, etc.). Defaults to GET.",
					},
					"header": map[string]any{
						"type":        "object",
						"description": "HTTP headers as key-value pairs.",
						"additionalProperties": map[string]any{
							"type": "string",
						},
					},
				},
				"required": []string{"url"},
			},
		},
	},
}

// ToolsForPersona returns the openai tool params for tools enabled on the given persona.
// Only tools present in the persona's ToolsEnabled list are included.
func ToolsForPersona(enabled []string) []openai.ChatCompletionToolParam {
	if len(enabled) == 0 {
		return nil
	}
	tools := make([]openai.ChatCompletionToolParam, 0, len(enabled))
	for _, name := range enabled {
		if t, ok := allTools[name]; ok {
			tools = append(tools, t)
		}
	}
	return tools
}

// AllToolNames returns the names of all available tools.
func AllToolNames() []string {
	names := make([]string, 0, len(allTools))
	for name := range allTools {
		names = append(names, name)
	}
	return names
}
