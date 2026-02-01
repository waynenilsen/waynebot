package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// DocumentHandler handles project document HTTP endpoints.
type DocumentHandler struct {
	DB *db.DB
}

var knownDocTypes = map[string]bool{
	"erd":       true,
	"prd":       true,
	"decisions": true,
}

type documentListItem struct {
	Type   string `json:"type"`
	Exists bool   `json:"exists"`
}

type documentResponse struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type documentWriteRequest struct {
	Content string `json:"content"`
}

func waynebotDir(project model.Project) string {
	return filepath.Join(project.Path, ".waynebot")
}

func docPath(project model.Project, docType string) string {
	return filepath.Join(waynebotDir(project), docType+".md")
}

// ListDocuments returns which document types exist for a project.
func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	project, ok := h.lookupProject(w, r)
	if !ok {
		return
	}

	items := make([]documentListItem, 0, len(knownDocTypes))
	for _, t := range []string{"erd", "prd", "decisions"} {
		_, err := os.Stat(docPath(project, t))
		items = append(items, documentListItem{
			Type:   t,
			Exists: err == nil,
		})
	}

	WriteJSON(w, http.StatusOK, items)
}

// GetDocument reads a single project document.
func (h *DocumentHandler) GetDocument(w http.ResponseWriter, r *http.Request) {
	project, ok := h.lookupProject(w, r)
	if !ok {
		return
	}

	docType := chi.URLParam(r, "type")
	if !knownDocTypes[docType] {
		ErrorResponse(w, http.StatusBadRequest, "unknown document type")
		return
	}

	data, err := os.ReadFile(docPath(project, docType))
	if err != nil {
		if os.IsNotExist(err) {
			ErrorResponse(w, http.StatusNotFound, "document not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "failed to read document")
		return
	}

	WriteJSON(w, http.StatusOK, documentResponse{
		Type:    docType,
		Content: string(data),
	})
}

// PutDocument creates or updates a project document (erd or prd only).
func (h *DocumentHandler) PutDocument(w http.ResponseWriter, r *http.Request) {
	project, ok := h.lookupProject(w, r)
	if !ok {
		return
	}

	docType := chi.URLParam(r, "type")
	if docType != "erd" && docType != "prd" {
		ErrorResponse(w, http.StatusBadRequest, "only erd and prd documents can be updated via PUT")
		return
	}

	var req documentWriteRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dir := waynebotDir(project)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to create .waynebot directory")
		return
	}

	if err := os.WriteFile(docPath(project, docType), []byte(req.Content), 0o644); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to write document")
		return
	}

	WriteJSON(w, http.StatusOK, documentResponse{
		Type:    docType,
		Content: req.Content,
	})
}

// AppendDecision appends an entry to the decisions log with a timestamp header.
func (h *DocumentHandler) AppendDecision(w http.ResponseWriter, r *http.Request) {
	project, ok := h.lookupProject(w, r)
	if !ok {
		return
	}

	var req documentWriteRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		ErrorResponse(w, http.StatusBadRequest, "content is required")
		return
	}

	dir := waynebotDir(project)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to create .waynebot directory")
		return
	}

	entry := fmt.Sprintf("\n## %s\n\n%s\n", time.Now().UTC().Format(time.RFC3339), req.Content)

	f, err := os.OpenFile(docPath(project, "decisions"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to open decisions log")
		return
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to append decision")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// lookupProject extracts the project ID from the URL, fetches it, and validates the path.
func (h *DocumentHandler) lookupProject(w http.ResponseWriter, r *http.Request) (model.Project, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid project id")
		return model.Project{}, false
	}

	project, err := model.GetProject(h.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "project not found")
			return model.Project{}, false
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return model.Project{}, false
	}

	info, err := os.Stat(project.Path)
	if err != nil || !info.IsDir() {
		ErrorResponse(w, http.StatusBadRequest, "project path is not a valid directory")
		return model.Project{}, false
	}

	return project, true
}
