package api

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/waynenilsen/waynebot/internal/agent"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// DMHandler handles DM-related HTTP endpoints.
type DMHandler struct {
	DB         *db.DB
	Hub        *ws.Hub
	Supervisor *agent.Supervisor
}

type createDMRequest struct {
	UserID    *int64 `json:"user_id"`
	PersonaID *int64 `json:"persona_id"`
}

type dmParticipantJSON struct {
	UserID      *int64  `json:"user_id,omitempty"`
	UserName    *string `json:"user_name,omitempty"`
	PersonaID   *int64  `json:"persona_id,omitempty"`
	PersonaName *string `json:"persona_name,omitempty"`
}

type dmChannelJSON struct {
	ID               int64             `json:"id"`
	Name             string            `json:"name"`
	IsDM             bool              `json:"is_dm"`
	CreatedAt        string            `json:"created_at"`
	OtherParticipant dmParticipantJSON `json:"other_participant"`
}

type dmChannelWithUnreadJSON struct {
	dmChannelJSON
	UnreadCount int64 `json:"unread_count"`
}

// CreateDM creates or retrieves a DM channel between the authenticated user and another user or persona.
func (h *DMHandler) CreateDM(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createDMRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if (req.UserID == nil) == (req.PersonaID == nil) {
		ErrorResponse(w, http.StatusBadRequest, "exactly one of user_id or persona_id must be set")
		return
	}

	// Build the two participants.
	self := model.DMParticipant{UserID: &user.ID}
	var other model.DMParticipant
	var otherParticipant dmParticipantJSON

	if req.UserID != nil {
		if *req.UserID == user.ID {
			ErrorResponse(w, http.StatusBadRequest, "cannot DM yourself")
			return
		}
		target, err := model.GetUser(h.DB, *req.UserID)
		if err != nil {
			if err == sql.ErrNoRows {
				ErrorResponse(w, http.StatusNotFound, "user not found")
				return
			}
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
			return
		}
		other = model.DMParticipant{UserID: req.UserID}
		otherParticipant = dmParticipantJSON{UserID: &target.ID, UserName: &target.Username}
	} else {
		target, err := model.GetPersona(h.DB, *req.PersonaID)
		if err != nil {
			if err == sql.ErrNoRows {
				ErrorResponse(w, http.StatusNotFound, "persona not found")
				return
			}
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
			return
		}
		other = model.DMParticipant{PersonaID: req.PersonaID}
		otherParticipant = dmParticipantJSON{PersonaID: &target.ID, PersonaName: &target.Name}
	}

	// Check for existing DM.
	ch, err := model.FindDMChannel(h.DB, self, other)
	if err != nil && err != sql.ErrNoRows {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	created := false
	if err == sql.ErrNoRows {
		name := fmt.Sprintf("dm-%d", user.ID)
		if req.UserID != nil {
			name = fmt.Sprintf("dm-%d-%d", user.ID, *req.UserID)
		} else {
			name = fmt.Sprintf("dm-%d-p%d", user.ID, *req.PersonaID)
		}
		ch, err = model.CreateDMChannel(h.DB, name, self, other, user.ID)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
			return
		}
		created = true

		// If we created a DM with a persona, restart the actor so it picks up the subscription.
		if req.PersonaID != nil && h.Supervisor != nil && h.Supervisor.Running() {
			if err := h.Supervisor.RestartActor(*req.PersonaID); err != nil {
				slog.Error("dm: failed to restart actor", "persona_id", *req.PersonaID, "err", err)
			}
		}
	}

	resp := dmChannelJSON{
		ID:               ch.ID,
		Name:             ch.Name,
		IsDM:             ch.IsDM,
		CreatedAt:        ch.CreatedAt.Format(time.RFC3339),
		OtherParticipant: otherParticipant,
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	WriteJSON(w, status, resp)
}

// ListDMs returns all DM channels for the authenticated user.
func (h *DMHandler) ListDMs(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	dms, err := model.ListDMsForUser(h.DB, user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	var counts map[int64]int64
	counts, _ = model.GetUnreadCounts(h.DB, user.ID)

	out := make([]dmChannelWithUnreadJSON, len(dms))
	for i, dm := range dms {
		out[i] = dmChannelWithUnreadJSON{
			dmChannelJSON: dmChannelJSON{
				ID:        dm.Channel.ID,
				Name:      dm.Channel.Name,
				IsDM:      dm.Channel.IsDM,
				CreatedAt: dm.Channel.CreatedAt.Format(time.RFC3339),
				OtherParticipant: dmParticipantJSON{
					UserID:      dm.OtherUserID,
					UserName:    dm.OtherUserName,
					PersonaID:   dm.OtherPersonaID,
					PersonaName: dm.OtherPersonaName,
				},
			},
			UnreadCount: counts[dm.Channel.ID],
		}
	}
	WriteJSON(w, http.StatusOK, out)
}
