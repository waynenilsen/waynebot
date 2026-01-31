package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/ws"
)

type contextKey string

const personaIDKey contextKey = "persona_id"

// WithPersonaID returns a context carrying the given persona ID.
func WithPersonaID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, personaIDKey, id)
}

// PersonaIDFromContext retrieves the persona ID from a context, or 0 if not set.
func PersonaIDFromContext(ctx context.Context) int64 {
	id, _ := ctx.Value(personaIDKey).(int64)
	return id
}

type messageReactArgs struct {
	MessageID int64  `json:"message_id"`
	Emoji     string `json:"emoji"`
	Remove    bool   `json:"remove"`
}

// MessageReact returns a ToolFunc that adds or removes an emoji reaction on a message.
// The persona ID is extracted from the context via WithPersonaID.
func MessageReact(database *db.DB, hub *ws.Hub) ToolFunc {
	return func(ctx context.Context, raw json.RawMessage) (string, error) {
		personaID := PersonaIDFromContext(ctx)
		if personaID == 0 {
			return "", fmt.Errorf("persona_id not set in context")
		}

		var args messageReactArgs
		if err := json.Unmarshal(raw, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}
		if args.MessageID == 0 {
			return "", fmt.Errorf("message_id is required")
		}
		if args.Emoji == "" {
			return "", fmt.Errorf("emoji is required")
		}

		channelID, err := model.GetMessageChannelID(database, args.MessageID)
		if err != nil {
			return "", fmt.Errorf("message not found: %w", err)
		}

		if args.Remove {
			removed, err := model.RemoveReaction(database, args.MessageID, personaID, "agent", args.Emoji)
			if err != nil {
				return "", fmt.Errorf("remove reaction: %w", err)
			}
			if !removed {
				return "no reaction to remove", nil
			}
			counts, _ := model.GetReactionCounts(database, args.MessageID, personaID, "agent")
			if hub != nil {
				hub.Broadcast(ws.Event{
					Type: "remove_reaction",
					Data: map[string]any{
						"message_id":  args.MessageID,
						"channel_id":  channelID,
						"emoji":       args.Emoji,
						"author_id":   personaID,
						"author_type": "agent",
						"counts":      counts,
					},
				})
			}
			return "reaction removed", nil
		}

		added, err := model.AddReaction(database, args.MessageID, personaID, "agent", args.Emoji)
		if err != nil {
			return "", fmt.Errorf("add reaction: %w", err)
		}
		if !added {
			return "already reacted", nil
		}
		counts, _ := model.GetReactionCounts(database, args.MessageID, personaID, "agent")
		if hub != nil {
			hub.Broadcast(ws.Event{
				Type: "new_reaction",
				Data: map[string]any{
					"message_id":  args.MessageID,
					"channel_id":  channelID,
					"emoji":       args.Emoji,
					"author_id":   personaID,
					"author_type": "agent",
					"counts":      counts,
				},
			})
		}
		return "reaction added", nil
	}
}
