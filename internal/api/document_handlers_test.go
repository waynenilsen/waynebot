package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestListDocumentsEmpty(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "GET",
		fmt.Sprintf("/api/projects/%d/documents", proj.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var items []struct {
		Type  string   `json:"type"`
		Files []string `json:"files"`
	}
	json.NewDecoder(rec.Body).Decode(&items)

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	for _, item := range items {
		if len(item.Files) != 0 {
			t.Errorf("expected %s to have no files, got %v", item.Type, item.Files)
		}
	}
}

func TestListDocumentsWithSomeDocs(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	os.MkdirAll(filepath.Join(projDir, "erd"), 0o755)
	os.MkdirAll(filepath.Join(projDir, "decisions"), 0o755)
	os.WriteFile(filepath.Join(projDir, "erd", "main.md"), []byte("# ERD"), 0o644)
	os.WriteFile(filepath.Join(projDir, "decisions", "log.md"), []byte("# Decisions"), 0o644)

	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "GET",
		fmt.Sprintf("/api/projects/%d/documents", proj.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var items []struct {
		Type  string   `json:"type"`
		Files []string `json:"files"`
	}
	json.NewDecoder(rec.Body).Decode(&items)

	expect := map[string]int{"erd": 1, "prd": 0, "decisions": 1}
	for _, item := range items {
		if len(item.Files) != expect[item.Type] {
			t.Errorf("%s: files count = %d, want %d", item.Type, len(item.Files), expect[item.Type])
		}
	}
}

func TestGetDocumentExists(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	os.MkdirAll(filepath.Join(projDir, "erd"), 0o755)
	os.WriteFile(filepath.Join(projDir, "erd", "main.md"), []byte("Users -> Posts"), 0o644)

	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "GET",
		fmt.Sprintf("/api/projects/%d/documents/erd/main.md", proj.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Type     string `json:"type"`
		Filename string `json:"filename"`
		Content  string `json:"content"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Type != "erd" {
		t.Errorf("type = %q, want erd", resp.Type)
	}
	if resp.Filename != "main.md" {
		t.Errorf("filename = %q, want main.md", resp.Filename)
	}
	if resp.Content != "Users -> Posts" {
		t.Errorf("content = %q, want %q", resp.Content, "Users -> Posts")
	}
}

func TestGetDocumentNotFound(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "GET",
		fmt.Sprintf("/api/projects/%d/documents/erd/main.md", proj.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestGetDocumentInvalidType(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "GET",
		fmt.Sprintf("/api/projects/%d/documents/bogus/main.md", proj.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestPutDocumentCreate(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "PUT",
		fmt.Sprintf("/api/projects/%d/documents/erd/main.md", proj.ID),
		`{"content":"# New ERD\nUsers table"}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	data, err := os.ReadFile(filepath.Join(projDir, "erd", "main.md"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(data) != "# New ERD\nUsers table" {
		t.Errorf("file content = %q, want %q", string(data), "# New ERD\nUsers table")
	}
}

func TestPutDocumentUpdate(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	os.MkdirAll(filepath.Join(projDir, "prd"), 0o755)
	os.WriteFile(filepath.Join(projDir, "prd", "main.md"), []byte("old content"), 0o644)

	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "PUT",
		fmt.Sprintf("/api/projects/%d/documents/prd/main.md", proj.ID),
		`{"content":"new content"}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	data, _ := os.ReadFile(filepath.Join(projDir, "prd", "main.md"))
	if string(data) != "new content" {
		t.Errorf("file content = %q, want %q", string(data), "new content")
	}
}

func TestAppendDocumentNew(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "POST",
		fmt.Sprintf("/api/projects/%d/documents/decisions/log.md", proj.ID),
		`{"content":"We chose Go for the backend."}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	data, err := os.ReadFile(filepath.Join(projDir, "decisions", "log.md"))
	if err != nil {
		t.Fatalf("read decisions: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "We chose Go for the backend.") {
		t.Error("decisions should contain the appended content")
	}
	if !strings.Contains(content, "## 20") {
		t.Error("decisions should contain a timestamp header")
	}
}

func TestAppendDocumentExisting(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	os.MkdirAll(filepath.Join(projDir, "decisions"), 0o755)
	os.WriteFile(filepath.Join(projDir, "decisions", "log.md"), []byte("## Existing\nFirst decision."), 0o644)

	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "POST",
		fmt.Sprintf("/api/projects/%d/documents/decisions/log.md", proj.ID),
		`{"content":"Second decision."}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	data, _ := os.ReadFile(filepath.Join(projDir, "decisions", "log.md"))
	content := string(data)
	if !strings.Contains(content, "First decision.") {
		t.Error("should preserve existing content")
	}
	if !strings.Contains(content, "Second decision.") {
		t.Error("should contain new decision")
	}
}

func TestAppendDocumentEmptyContent(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "POST",
		fmt.Sprintf("/api/projects/%d/documents/decisions/log.md", proj.ID),
		`{"content":"  "}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestDeleteDocument(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	os.MkdirAll(filepath.Join(projDir, "erd"), 0o755)
	os.WriteFile(filepath.Join(projDir, "erd", "main.md"), []byte("# ERD"), 0o644)

	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "DELETE",
		fmt.Sprintf("/api/projects/%d/documents/erd/main.md", proj.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	if _, err := os.Stat(filepath.Join(projDir, "erd", "main.md")); !os.IsNotExist(err) {
		t.Error("file should have been deleted")
	}
}

func TestDeleteDocumentNotFound(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "DELETE",
		fmt.Sprintf("/api/projects/%d/documents/erd/nonexistent.md", proj.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestDocumentsMissingProject(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "GET", "/api/projects/99999/documents", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestDocumentsRequireAuth(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	projDir := t.TempDir()
	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "GET",
		fmt.Sprintf("/api/projects/%d/documents", proj.ID), "")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestListCategoryDocuments(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	os.MkdirAll(filepath.Join(projDir, "erd"), 0o755)
	os.WriteFile(filepath.Join(projDir, "erd", "main.md"), []byte("# ERD"), 0o644)
	os.WriteFile(filepath.Join(projDir, "erd", "v2.md"), []byte("# ERD v2"), 0o644)

	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "GET",
		fmt.Sprintf("/api/projects/%d/documents/erd", proj.ID), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Type  string   `json:"type"`
		Files []string `json:"files"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Type != "erd" {
		t.Errorf("type = %q, want erd", resp.Type)
	}
	if len(resp.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(resp.Files))
	}
}
