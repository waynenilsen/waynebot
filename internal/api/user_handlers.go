package api

import (
	"net/http"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// UserHandler handles user-related HTTP endpoints.
type UserHandler struct {
	DB *db.DB
}

// ListUsers returns all users (id + username only).
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := model.ListUsers(h.DB)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]userJSON, len(users))
	for i, u := range users {
		out[i] = toUserJSON(u)
	}

	WriteJSON(w, http.StatusOK, out)
}
