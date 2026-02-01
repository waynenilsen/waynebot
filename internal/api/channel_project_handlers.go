package api

import (
	"database/sql"
	"net/http"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// ChannelProjectHandler handles channel-project association endpoints.
type ChannelProjectHandler struct {
	DB *db.DB
}

type addChannelProjectRequest struct {
	ProjectID int64 `json:"project_id"`
}

// ListChannelProjects returns all projects associated with a channel.
func (h *ChannelProjectHandler) ListChannelProjects(w http.ResponseWriter, r *http.Request) {
	channelID, ok := ParseIntParam(w, r, "id")
	if !ok {
		return
	}

	if _, err := model.GetChannel(h.DB, channelID); err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "channel not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	projects, err := model.ListChannelProjects(h.DB, channelID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]projectJSON, len(projects))
	for i, p := range projects {
		out[i] = toProjectJSON(p)
	}
	WriteJSON(w, http.StatusOK, out)
}

// AddChannelProject associates a project with a channel.
func (h *ChannelProjectHandler) AddChannelProject(w http.ResponseWriter, r *http.Request) {
	channelID, ok := ParseIntParam(w, r, "id")
	if !ok {
		return
	}

	if _, err := model.GetChannel(h.DB, channelID); err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "channel not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	var req addChannelProjectRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.ProjectID == 0 {
		ErrorResponse(w, http.StatusBadRequest, "project_id is required")
		return
	}

	if _, err := model.GetProject(h.DB, req.ProjectID); err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "project not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := model.SetChannelProject(h.DB, channelID, req.ProjectID); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveChannelProject removes a project association from a channel.
func (h *ChannelProjectHandler) RemoveChannelProject(w http.ResponseWriter, r *http.Request) {
	channelID, ok := ParseIntParam(w, r, "id")
	if !ok {
		return
	}

	projectID, ok := ParseIntParam(w, r, "projectID")
	if !ok {
		return
	}

	if err := model.RemoveChannelProject(h.DB, channelID, projectID); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
