package api

import (
	"net/http"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// MentionHandler handles mention-related endpoints.
type MentionHandler struct {
	DB *db.DB
}

type mentionTargetJSON struct {
	Type string `json:"type"`
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ListMentionTargets returns all users and personas that can be @mentioned.
func (h *MentionHandler) ListMentionTargets(w http.ResponseWriter, r *http.Request) {
	users, err := model.ListUsers(h.DB)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	personas, err := model.ListPersonas(h.DB)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]mentionTargetJSON, 0, len(users)+len(personas))
	for _, u := range users {
		out = append(out, mentionTargetJSON{
			Type: "user",
			ID:   u.ID,
			Name: u.Username,
		})
	}
	for _, p := range personas {
		out = append(out, mentionTargetJSON{
			Type: "persona",
			ID:   p.ID,
			Name: p.Name,
		})
	}

	WriteJSON(w, http.StatusOK, out)
}
