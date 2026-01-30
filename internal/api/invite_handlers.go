package api

import (
	"net/http"
	"time"

	"github.com/waynenilsen/waynebot/internal/auth"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// InviteHandler handles invite HTTP endpoints.
type InviteHandler struct {
	DB *db.DB
}

type inviteJSON struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	CreatedBy int64  `json:"created_by"`
	UsedBy    *int64 `json:"used_by"`
	CreatedAt string `json:"created_at"`
}

func toInviteJSON(inv model.Invite) inviteJSON {
	return inviteJSON{
		ID:        inv.ID,
		Code:      inv.Code,
		CreatedBy: inv.CreatedBy,
		UsedBy:    inv.UsedBy,
		CreatedAt: inv.CreatedAt.Format(time.RFC3339),
	}
}

// CreateInvite generates a new invite code.
func (h *InviteHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	code, err := auth.GenerateInviteCode()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	inv, err := model.CreateInvite(h.DB, code, user.ID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	WriteJSON(w, http.StatusCreated, toInviteJSON(inv))
}

// ListInvites returns all invites.
func (h *InviteHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	invites, err := model.ListInvites(h.DB)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]inviteJSON, len(invites))
	for i, inv := range invites {
		out[i] = toInviteJSON(inv)
	}
	WriteJSON(w, http.StatusOK, out)
}
