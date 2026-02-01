package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// MemberHandler handles channel membership endpoints.
type MemberHandler struct {
	DB *db.DB
}

type memberJSON struct {
	Type string `json:"type"`
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type addMemberRequest struct {
	UserID    *int64 `json:"user_id"`
	PersonaID *int64 `json:"persona_id"`
}

type removeMemberRequest struct {
	UserID    *int64 `json:"user_id"`
	PersonaID *int64 `json:"persona_id"`
}

func (h *MemberHandler) parseChannelID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	channelID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid channel id")
		return 0, false
	}
	if _, err := model.GetChannel(h.DB, channelID); err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "channel not found")
		} else {
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
		}
		return 0, false
	}
	return channelID, true
}

// ListMembers returns all members (users and personas) of a channel.
func (h *MemberHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	channelID, ok := h.parseChannelID(w, r)
	if !ok {
		return
	}

	users, err := model.GetChannelMembers(h.DB, channelID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	personas, err := model.GetChannelPersonas(h.DB, channelID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]memberJSON, 0, len(users)+len(personas))
	for _, u := range users {
		out = append(out, memberJSON{
			Type: "user",
			ID:   u.UserID,
			Name: u.Username,
			Role: u.Role,
		})
	}
	for _, p := range personas {
		out = append(out, memberJSON{
			Type: "persona",
			ID:   p.PersonaID,
			Name: p.Name,
			Role: "member",
		})
	}

	WriteJSON(w, http.StatusOK, out)
}

// AddMember adds a user or persona to a channel.
func (h *MemberHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	channelID, ok := h.parseChannelID(w, r)
	if !ok {
		return
	}

	var req addMemberRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.UserID == nil && req.PersonaID == nil {
		ErrorResponse(w, http.StatusBadRequest, "user_id or persona_id is required")
		return
	}
	if req.UserID != nil && req.PersonaID != nil {
		ErrorResponse(w, http.StatusBadRequest, "provide only one of user_id or persona_id")
		return
	}

	if req.PersonaID != nil {
		if err := model.SubscribeChannel(h.DB, *req.PersonaID, channelID); err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}

	if err := model.AddChannelMember(h.DB, channelID, *req.UserID, "member"); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// RemoveMember removes a user or persona from a channel.
func (h *MemberHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	channelID, ok := h.parseChannelID(w, r)
	if !ok {
		return
	}

	var req removeMemberRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.UserID == nil && req.PersonaID == nil {
		ErrorResponse(w, http.StatusBadRequest, "user_id or persona_id is required")
		return
	}
	if req.UserID != nil && req.PersonaID != nil {
		ErrorResponse(w, http.StatusBadRequest, "provide only one of user_id or persona_id")
		return
	}

	if req.PersonaID != nil {
		if err := model.UnsubscribeChannel(h.DB, *req.PersonaID, channelID); err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := model.RemoveChannelMember(h.DB, channelID, *req.UserID); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
