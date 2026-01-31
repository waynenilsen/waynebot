package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// ReactionHandler handles emoji reaction HTTP endpoints.
type ReactionHandler struct {
	DB  *db.DB
	Hub *ws.Hub
}

type reactionRequest struct {
	Emoji string `json:"emoji"`
}

type reactionEventPayload struct {
	MessageID  int64                 `json:"message_id"`
	ChannelID  int64                 `json:"channel_id"`
	Emoji      string                `json:"emoji"`
	AuthorID   int64                 `json:"author_id"`
	AuthorType string                `json:"author_type"`
	Counts     []model.ReactionCount `json:"counts"`
}

// AddReaction handles PUT /api/channels/{id}/messages/{messageID}/reactions
func (h *ReactionHandler) AddReaction(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	channelID, messageID, ok := h.parseParams(w, r)
	if !ok {
		return
	}

	var req reactionRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Emoji = strings.TrimSpace(req.Emoji)
	if req.Emoji == "" {
		ErrorResponse(w, http.StatusBadRequest, "emoji is required")
		return
	}

	// Verify message belongs to this channel.
	msgChannelID, err := model.GetMessageChannelID(h.DB, messageID)
	if err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "message not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}
	if msgChannelID != channelID {
		ErrorResponse(w, http.StatusNotFound, "message not found in this channel")
		return
	}

	added, err := model.AddReaction(h.DB, messageID, user.ID, "human", req.Emoji)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	counts, _ := model.GetReactionCounts(h.DB, messageID, user.ID, "human")

	if added && h.Hub != nil {
		h.Hub.Broadcast(ws.Event{
			Type: "new_reaction",
			Data: reactionEventPayload{
				MessageID:  messageID,
				ChannelID:  channelID,
				Emoji:      req.Emoji,
				AuthorID:   user.ID,
				AuthorType: "human",
				Counts:     counts,
			},
		})
	}

	WriteJSON(w, http.StatusOK, counts)
}

// RemoveReaction handles DELETE /api/channels/{id}/messages/{messageID}/reactions
func (h *ReactionHandler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	channelID, messageID, ok := h.parseParams(w, r)
	if !ok {
		return
	}

	var req reactionRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Emoji = strings.TrimSpace(req.Emoji)
	if req.Emoji == "" {
		ErrorResponse(w, http.StatusBadRequest, "emoji is required")
		return
	}

	removed, err := model.RemoveReaction(h.DB, messageID, user.ID, "human", req.Emoji)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	counts, _ := model.GetReactionCounts(h.DB, messageID, user.ID, "human")

	if removed && h.Hub != nil {
		h.Hub.Broadcast(ws.Event{
			Type: "remove_reaction",
			Data: reactionEventPayload{
				MessageID:  messageID,
				ChannelID:  channelID,
				Emoji:      req.Emoji,
				AuthorID:   user.ID,
				AuthorType: "human",
				Counts:     counts,
			},
		})
	}

	WriteJSON(w, http.StatusOK, counts)
}

func (h *ReactionHandler) parseParams(w http.ResponseWriter, r *http.Request) (channelID, messageID int64, ok bool) {
	var err error
	channelID, err = strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid channel id")
		return 0, 0, false
	}
	messageID, err = strconv.ParseInt(chi.URLParam(r, "messageID"), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid message id")
		return 0, 0, false
	}
	return channelID, messageID, true
}
