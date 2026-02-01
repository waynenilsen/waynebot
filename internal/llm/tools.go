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
	"project_docs": {
		Function: shared.FunctionDefinitionParam{
			Name:        "project_docs",
			Description: param.NewOpt("Read, write, or list project documents (.waynebot/ directory). Use action=list to see which docs exist, action=read to read a doc, action=write to create/update erd or prd, action=append to add a timestamped entry to the decisions log."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"action": map[string]any{
						"type":        "string",
						"enum":        []string{"read", "write", "append", "list"},
						"description": "The action to perform.",
					},
					"doc_type": map[string]any{
						"type":        "string",
						"enum":        []string{"erd", "prd", "decisions"},
						"description": "The document type (required for read, write, append).",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "Content to write or append (required for write and append).",
					},
				},
				"required": []string{"action"},
			},
		},
	},
	"message_react": {
		Function: shared.FunctionDefinitionParam{
			Name:        "message_react",
			Description: param.NewOpt("Add or remove an emoji reaction on a message."),
			Parameters: shared.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"message_id": map[string]any{
						"type":        "integer",
						"description": "The ID of the message to react to.",
					},
					"emoji": map[string]any{
						"type":        "string",
						"description": "The emoji to react with (unicode).",
					},
					"remove": map[string]any{
						"type":        "boolean",
						"description": "If true, remove the reaction instead of adding it. Defaults to false.",
					},
				},
				"required": []string{"message_id", "emoji"},
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
