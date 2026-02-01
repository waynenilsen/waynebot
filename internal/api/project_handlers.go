package api

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// ProjectHandler handles project HTTP endpoints.
type ProjectHandler struct {
	DB *db.DB
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type projectJSON struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

func toProjectJSON(p model.Project) projectJSON {
	return projectJSON{
		ID:          p.ID,
		Name:        p.Name,
		Path:        p.Path,
		Description: p.Description,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
	}
}

// ListProjects returns all projects.
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := model.ListProjects(h.DB)
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

// CreateProject creates a new project.
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 1 || len(req.Name) > 100 {
		ErrorResponse(w, http.StatusBadRequest, "name must be 1-100 characters")
		return
	}

	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" {
		ErrorResponse(w, http.StatusBadRequest, "path is required")
		return
	}

	p, err := model.CreateProject(h.DB, req.Name, req.Path, req.Description)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			ErrorResponse(w, http.StatusConflict, "project name already taken")
			return
		}
		if strings.Contains(err.Error(), "not a directory") || strings.Contains(err.Error(), "no such file") {
			ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	WriteJSON(w, http.StatusCreated, toProjectJSON(p))
}

// UpdateProject updates an existing project.
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id, ok := ParseIntParam(w, r, "id")
	if !ok {
		return
	}

	if _, err := model.GetProject(h.DB, id); err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "project not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	var req createProjectRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 1 || len(req.Name) > 100 {
		ErrorResponse(w, http.StatusBadRequest, "name must be 1-100 characters")
		return
	}

	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" {
		ErrorResponse(w, http.StatusBadRequest, "path is required")
		return
	}

	if err := model.UpdateProject(h.DB, id, req.Name, req.Path, req.Description); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			ErrorResponse(w, http.StatusConflict, "project name already taken")
			return
		}
		if strings.Contains(err.Error(), "not a directory") || strings.Contains(err.Error(), "no such file") {
			ErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	p, err := model.GetProject(h.DB, id)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	WriteJSON(w, http.StatusOK, toProjectJSON(p))
}

// DeleteProject deletes a project.
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := ParseIntParam(w, r, "id")
	if !ok {
		return
	}

	if _, err := model.GetProject(h.DB, id); err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "project not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := model.DeleteProject(h.DB, id); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
