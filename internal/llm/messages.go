package llm

import (
	"github.com/openai/openai-go"
	"github.com/waynenilsen/waynebot/internal/model"
)

// BuildMessages converts a persona and domain messages into openai SDK message types.
// It prepends the persona's system prompt, then maps each domain message by author type.
func BuildMessages(persona model.Persona, history []model.Message) []openai.ChatCompletionMessageParamUnion {
	msgs := make([]openai.ChatCompletionMessageParamUnion, 0, len(history)+1)

	if persona.SystemPrompt != "" {
		msgs = append(msgs, openai.SystemMessage(persona.SystemPrompt))
	}

	for _, m := range history {
		switch m.AuthorType {
		case "human":
			msgs = append(msgs, openai.UserMessage(m.AuthorName+": "+m.Content))
		case "agent":
			msgs = append(msgs, openai.AssistantMessage(m.Content))
		case "tool_call":
			msgs = append(msgs, assistantToolCallMessage(m))
		case "tool_result":
			msgs = append(msgs, openai.ToolMessage(m.Content, m.AuthorName))
		default:
			msgs = append(msgs, openai.UserMessage(m.Content))
		}
	}

	return msgs
}

// assistantToolCallMessage builds an assistant message with tool calls from a domain message.
// The Content field is expected to contain the raw JSON tool call ID, and AuthorName contains
// the tool name. The actual arguments are stored in the Content field as "name:id:arguments".
// For simplicity, we encode tool call info as: the message Content is the arguments JSON,
// AuthorName is "tool_call_id:tool_name".
func assistantToolCallMessage(m model.Message) openai.ChatCompletionMessageParamUnion {
	// For tool_call messages, we store:
	//   AuthorName = tool call ID
	//   Content    = tool_name + "\n" + arguments JSON
	// Parse them back out.
	name, args := splitToolCall(m.Content)
	return openai.ChatCompletionMessageParamUnion{
		OfAssistant: &openai.ChatCompletionAssistantMessageParam{
			ToolCalls: []openai.ChatCompletionMessageToolCallParam{
				{
					ID: m.AuthorName, // tool call ID stored in AuthorName
					Function: openai.ChatCompletionMessageToolCallFunctionParam{
						Name:      name,
						Arguments: args,
					},
				},
			},
		},
	}
}

// splitToolCall splits "tool_name\narguments_json" into name and args.
func splitToolCall(content string) (name, args string) {
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			return content[:i], content[i+1:]
		}
	}
	return content, "{}"
}
