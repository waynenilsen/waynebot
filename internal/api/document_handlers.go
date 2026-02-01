package api

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fmt"

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
	Type  string   `json:"type"`
	Files []string `json:"files"`
}

type documentResponse struct {
	Type     string `json:"type"`
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type documentWriteRequest struct {
	Content string `json:"content"`
}

func docTypeDir(project model.Project, docType string) string {
	return filepath.Join(project.Path, docType)
}

// listMarkdownFiles returns the names of all .md files in a directory.
func listMarkdownFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	return names
}

// ListDocuments returns all document categories and their files.
func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	project, ok := h.lookupProject(w, r)
	if !ok {
		return
	}

	items := make([]documentListItem, 0, len(knownDocTypes))
	for _, t := range []string{"erd", "prd", "decisions"} {
		files := listMarkdownFiles(docTypeDir(project, t))
		if files == nil {
			files = []string{}
		}
		items = append(items, documentListItem{
			Type:  t,
			Files: files,
		})
	}

	WriteJSON(w, http.StatusOK, items)
}

// ListCategoryDocuments returns the files in a specific document category.
func (h *DocumentHandler) ListCategoryDocuments(w http.ResponseWriter, r *http.Request) {
	project, ok := h.lookupProject(w, r)
	if !ok {
		return
	}

	docType := chi.URLParam(r, "type")
	if !knownDocTypes[docType] {
		ErrorResponse(w, http.StatusBadRequest, "unknown document type")
		return
	}

	files := listMarkdownFiles(docTypeDir(project, docType))
	if files == nil {
		files = []string{}
	}

	WriteJSON(w, http.StatusOK, documentListItem{
		Type:  docType,
		Files: files,
	})
}

// GetDocument reads a specific document file.
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

	filename := chi.URLParam(r, "filename")
	if filename == "" {
		ErrorResponse(w, http.StatusBadRequest, "filename is required")
		return
	}
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	data, err := os.ReadFile(filepath.Join(docTypeDir(project, docType), filename))
	if err != nil {
		if os.IsNotExist(err) {
			ErrorResponse(w, http.StatusNotFound, "document not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "failed to read document")
		return
	}

	WriteJSON(w, http.StatusOK, documentResponse{
		Type:     docType,
		Filename: filename,
		Content:  string(data),
	})
}

// PutDocument creates or updates a document file.
func (h *DocumentHandler) PutDocument(w http.ResponseWriter, r *http.Request) {
	project, ok := h.lookupProject(w, r)
	if !ok {
		return
	}

	docType := chi.URLParam(r, "type")
	if !knownDocTypes[docType] {
		ErrorResponse(w, http.StatusBadRequest, "unknown document type")
		return
	}

	filename := chi.URLParam(r, "filename")
	if filename == "" {
		ErrorResponse(w, http.StatusBadRequest, "filename is required")
		return
	}
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	var req documentWriteRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	dir := docTypeDir(project, docType)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to create document directory")
		return
	}

	if err := os.WriteFile(filepath.Join(dir, filename), []byte(req.Content), 0o644); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to write document")
		return
	}

	WriteJSON(w, http.StatusOK, documentResponse{
		Type:     docType,
		Filename: filename,
		Content:  req.Content,
	})
}

// AppendDocument appends a timestamped entry to a document file.
func (h *DocumentHandler) AppendDocument(w http.ResponseWriter, r *http.Request) {
	project, ok := h.lookupProject(w, r)
	if !ok {
		return
	}

	docType := chi.URLParam(r, "type")
	if !knownDocTypes[docType] {
		ErrorResponse(w, http.StatusBadRequest, "unknown document type")
		return
	}

	filename := chi.URLParam(r, "filename")
	if filename == "" {
		ErrorResponse(w, http.StatusBadRequest, "filename is required")
		return
	}
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
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

	dir := docTypeDir(project, docType)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to create document directory")
		return
	}

	entry := fmt.Sprintf("\n## %s\n\n%s\n", time.Now().UTC().Format(time.RFC3339), req.Content)

	f, err := os.OpenFile(filepath.Join(dir, filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to open document")
		return
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to append to document")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DeleteDocument removes a document file.
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	project, ok := h.lookupProject(w, r)
	if !ok {
		return
	}

	docType := chi.URLParam(r, "type")
	if !knownDocTypes[docType] {
		ErrorResponse(w, http.StatusBadRequest, "unknown document type")
		return
	}

	filename := chi.URLParam(r, "filename")
	if filename == "" {
		ErrorResponse(w, http.StatusBadRequest, "filename is required")
		return
	}
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	fp := filepath.Join(docTypeDir(project, docType), filename)
	if err := os.Remove(fp); err != nil {
		if os.IsNotExist(err) {
			ErrorResponse(w, http.StatusNotFound, "document not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "failed to delete document")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// lookupProject extracts the project ID from the URL, fetches it, and validates the path.
func (h *DocumentHandler) lookupProject(w http.ResponseWriter, r *http.Request) (model.Project, bool) {
	id, ok := ParseIntParam(w, r, "id")
	if !ok {
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
