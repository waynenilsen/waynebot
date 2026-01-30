package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// ChannelHandler handles channel and message HTTP endpoints.
type ChannelHandler struct {
	DB *db.DB
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
	ID         int64  `json:"id"`
	ChannelID  int64  `json:"channel_id"`
	AuthorID   int64  `json:"author_id"`
	AuthorType string `json:"author_type"`
	AuthorName string `json:"author_name"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
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

// ListChannels returns all channels.
func (h *ChannelHandler) ListChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := model.ListChannels(h.DB)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]channelJSON, len(channels))
	for i, ch := range channels {
		out[i] = toChannelJSON(ch)
	}
	WriteJSON(w, http.StatusOK, out)
}

// CreateChannel creates a new channel.
func (h *ChannelHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
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

	ch, err := model.CreateChannel(h.DB, req.Name, req.Description)
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

// GetMessages returns paginated messages for a channel.
func (h *ChannelHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	channelID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid channel id")
		return
	}

	// Verify channel exists.
	if _, err := model.GetChannel(h.DB, channelID); err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "channel not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
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

	var messages []model.Message
	if before := r.URL.Query().Get("before"); before != "" {
		beforeID, err := strconv.ParseInt(before, 10, 64)
		if err != nil {
			ErrorResponse(w, http.StatusBadRequest, "invalid before parameter")
			return
		}
		messages, err = model.GetMessagesBefore(h.DB, channelID, beforeID, limit)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
			return
		}
	} else {
		messages, err = model.GetRecentMessages(h.DB, channelID, limit)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
			return
		}
	}

	out := make([]messageJSON, len(messages))
	for i, m := range messages {
		out[i] = toMessageJSON(m)
	}
	WriteJSON(w, http.StatusOK, out)
}

// PostMessage sends a message to a channel.
func (h *ChannelHandler) PostMessage(w http.ResponseWriter, r *http.Request) {
	channelID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid channel id")
		return
	}

	// Verify channel exists.
	if _, err := model.GetChannel(h.DB, channelID); err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "channel not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	user := GetUser(r)
	if user == nil {
		ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

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

	WriteJSON(w, http.StatusCreated, toMessageJSON(msg))
}
