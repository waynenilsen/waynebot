package api

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/waynenilsen/waynebot/internal/agent"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// ChannelHandler handles channel and message HTTP endpoints.
type ChannelHandler struct {
	DB         *db.DB
	Hub        *ws.Hub
	Supervisor *agent.Supervisor
}

type createChannelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type channelJSON struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type channelWithUnreadJSON struct {
	channelJSON
	UnreadCount int64 `json:"unread_count"`
}

func toChannelJSON(ch model.Channel) channelJSON {
	return channelJSON{
		ID:          ch.ID,
		Name:        ch.Name,
		Description: ch.Description,
		CreatedAt:   ch.CreatedAt.Format(time.RFC3339),
	}
}

type postMessageRequest struct {
	Content string `json:"content"`
}

type messageJSON struct {
	ID         int64                 `json:"id"`
	ChannelID  int64                 `json:"channel_id"`
	AuthorID   int64                 `json:"author_id"`
	AuthorType string                `json:"author_type"`
	AuthorName string                `json:"author_name"`
	Content    string                `json:"content"`
	CreatedAt  string                `json:"created_at"`
	Reactions  []model.ReactionCount `json:"reactions"`
}

func toMessageJSON(m model.Message) messageJSON {
	return messageJSON{
		ID:         m.ID,
		ChannelID:  m.ChannelID,
		AuthorID:   m.AuthorID,
		AuthorType: m.AuthorType,
		AuthorName: m.AuthorName,
		Content:    m.Content,
		CreatedAt:  m.CreatedAt.Format(time.RFC3339),
	}
}

// ListChannels returns channels the authenticated user is a member of, with unread counts.
func (h *ChannelHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	channels, err := model.ListChannelsForUser(h.DB, user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	counts, _ := model.GetUnreadCounts(h.DB, user.ID)

	out := make([]channelWithUnreadJSON, len(channels))
	for i, ch := range channels {
		out[i] = channelWithUnreadJSON{
			channelJSON: toChannelJSON(ch),
			UnreadCount: counts[ch.ID],
		}
	}
	WriteJSON(w, http.StatusOK, out)
}

// CreateChannel creates a new channel and adds the creator as owner.
func (h *ChannelHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createChannelRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)

	if len(req.Name) < 1 || len(req.Name) > 100 {
		ErrorResponse(w, http.StatusBadRequest, "name must be 1-100 characters")
		return
	}
	if len(req.Description) > 500 {
		ErrorResponse(w, http.StatusBadRequest, "description must be 0-500 characters")
		return
	}

	ch, err := model.CreateChannel(h.DB, req.Name, req.Description, user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			ErrorResponse(w, http.StatusConflict, "channel name already taken")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	WriteJSON(w, http.StatusCreated, toChannelJSON(ch))
}

// requireChannelMember parses the channel ID, verifies the channel exists,
// and checks that the authenticated user is a member. For DM channels it
// checks dm_participants; for regular channels it checks channel_members.
// Returns the channelID and true on success, or writes an error response
// and returns false.
func (h *ChannelHandler) requireChannelMember(w http.ResponseWriter, r *http.Request) (int64, bool) {
	channelID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid channel id")
		return 0, false
	}

	ch, err := model.GetChannel(h.DB, channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "channel not found")
			return 0, false
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return 0, false
	}

	user := GetUser(r)
	if user == nil {
		ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return 0, false
	}

	if ch.IsDM {
		isMember, err := model.IsDMParticipant(h.DB, channelID, user.ID)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
			return 0, false
		}
		if !isMember {
			ErrorResponse(w, http.StatusForbidden, "not a channel member")
			return 0, false
		}
	} else {
		isMember, err := model.IsChannelMember(h.DB, channelID, user.ID)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
			return 0, false
		}
		if !isMember {
			ErrorResponse(w, http.StatusForbidden, "not a channel member")
			return 0, false
		}
	}

	return channelID, true
}

// GetMessages returns paginated messages for a channel.
func (h *ChannelHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	channelID, ok := h.requireChannelMember(w, r)
	if !ok {
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed < 1 || parsed > 200 {
			ErrorResponse(w, http.StatusBadRequest, "limit must be 1-200")
			return
		}
		limit = parsed
	}

	var (
		messages []model.Message
		err      error
	)
	if before := r.URL.Query().Get("before"); before != "" {
		beforeID, parseErr := strconv.ParseInt(before, 10, 64)
		if parseErr != nil {
			ErrorResponse(w, http.StatusBadRequest, "invalid before parameter")
			return
		}
		messages, err = model.GetMessagesBefore(h.DB, channelID, beforeID, limit)
	} else {
		messages, err = model.GetRecentMessages(h.DB, channelID, limit)
	}
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Fetch reactions for all messages in one batch query.
	var reactionMap map[int64][]model.ReactionCount
	user := GetUser(r)
	if user != nil && len(messages) > 0 {
		ids := make([]int64, len(messages))
		for i, m := range messages {
			ids[i] = m.ID
		}
		reactionMap, _ = model.GetReactionCountsBatch(h.DB, ids, user.ID, "human")
	}

	out := make([]messageJSON, len(messages))
	for i, m := range messages {
		mj := toMessageJSON(m)
		if reactionMap != nil {
			mj.Reactions = reactionMap[m.ID]
		}
		out[i] = mj
	}
	WriteJSON(w, http.StatusOK, out)
}

// PostMessage sends a message to a channel.
func (h *ChannelHandler) PostMessage(w http.ResponseWriter, r *http.Request) {
	channelID, ok := h.requireChannelMember(w, r)
	if !ok {
		return
	}

	user := GetUser(r)

	var req postMessageRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if len(req.Content) < 1 || len(req.Content) > 10000 {
		ErrorResponse(w, http.StatusBadRequest, "content must be 1-10000 characters")
		return
	}

	msg, err := model.CreateMessage(h.DB, channelID, user.ID, "human", user.Username, req.Content)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	if h.Hub != nil {
		h.Hub.Broadcast(ws.Event{
			Type: "new_message",
			Data: toMessageJSON(msg),
		})
	}

	// Auto-subscribe mentioned personas that aren't already in this channel.
	h.autoSubscribeMentionedPersonas(channelID, req.Content)

	WriteJSON(w, http.StatusCreated, toMessageJSON(msg))
}

// MarkRead updates the user's read position for a channel to the latest message.
func (h *ChannelHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	channelID, ok := h.requireChannelMember(w, r)
	if !ok {
		return
	}

	user := GetUser(r)

	latestID, err := model.GetLatestMessageID(h.DB, channelID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := model.UpdateReadPosition(h.DB, user.ID, channelID, latestID); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]int64{"last_read_message_id": latestID})
}

// autoSubscribeMentionedPersonas parses @mentions from content and subscribes
// any mentioned personas that aren't already subscribed to the channel.
func (h *ChannelHandler) autoSubscribeMentionedPersonas(channelID int64, content string) {
	personas, err := model.ListPersonas(h.DB)
	if err != nil {
		slog.Error("channel: list personas for mention check", "error", err)
		return
	}

	lower := strings.ToLower(content)
	for _, p := range personas {
		mention := "@" + strings.ToLower(p.Name)
		if !strings.Contains(lower, mention) {
			continue
		}

		// Check if already subscribed.
		channels, err := model.GetSubscribedChannels(h.DB, p.ID)
		if err != nil {
			slog.Error("channel: get subscribed channels", "persona", p.Name, "error", err)
			continue
		}
		alreadySubscribed := false
		for _, ch := range channels {
			if ch.ID == channelID {
				alreadySubscribed = true
				break
			}
		}
		if alreadySubscribed {
			continue
		}

		if err := model.SubscribeChannel(h.DB, p.ID, channelID); err != nil {
			slog.Error("channel: auto-subscribe persona on mention", "persona", p.Name, "channel_id", channelID, "error", err)
			continue
		}
		slog.Info("channel: auto-subscribed persona via @mention", "persona", p.Name, "channel_id", channelID)

		// Restart actor so it picks up the new subscription.
		if h.Supervisor != nil && h.Supervisor.Running() {
			if err := h.Supervisor.RestartActor(p.ID); err != nil {
				slog.Error("channel: restart actor after mention subscribe", "persona", p.Name, "error", err)
			}
		}
	}
}
