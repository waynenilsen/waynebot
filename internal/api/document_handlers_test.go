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
		Type   string `json:"type"`
		Exists bool   `json:"exists"`
	}
	json.NewDecoder(rec.Body).Decode(&items)

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	for _, item := range items {
		if item.Exists {
			t.Errorf("expected %s to not exist", item.Type)
		}
	}
}

func TestListDocumentsWithSomeDocs(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	wbDir := filepath.Join(projDir, ".waynebot")
	os.MkdirAll(wbDir, 0o755)
	os.WriteFile(filepath.Join(wbDir, "erd.md"), []byte("# ERD"), 0o644)
	os.WriteFile(filepath.Join(wbDir, "decisions.md"), []byte("# Decisions"), 0o644)

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
		Type   string `json:"type"`
		Exists bool   `json:"exists"`
	}
	json.NewDecoder(rec.Body).Decode(&items)

	expect := map[string]bool{"erd": true, "prd": false, "decisions": true}
	for _, item := range items {
		if item.Exists != expect[item.Type] {
			t.Errorf("%s: exists = %v, want %v", item.Type, item.Exists, expect[item.Type])
		}
	}
}

func TestGetDocumentExists(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	wbDir := filepath.Join(projDir, ".waynebot")
	os.MkdirAll(wbDir, 0o755)
	os.WriteFile(filepath.Join(wbDir, "erd.md"), []byte("Users -> Posts"), 0o644)

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
		Type    string `json:"type"`
		Content string `json:"content"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Type != "erd" {
		t.Errorf("type = %q, want erd", resp.Type)
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
		fmt.Sprintf("/api/projects/%d/documents/erd", proj.ID), "",
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
		fmt.Sprintf("/api/projects/%d/documents/bogus", proj.ID), "",
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
		fmt.Sprintf("/api/projects/%d/documents/erd", proj.ID),
		`{"content":"# New ERD\nUsers table"}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	// Verify file was written.
	data, err := os.ReadFile(filepath.Join(projDir, ".waynebot", "erd.md"))
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
	wbDir := filepath.Join(projDir, ".waynebot")
	os.MkdirAll(wbDir, 0o755)
	os.WriteFile(filepath.Join(wbDir, "prd.md"), []byte("old content"), 0o644)

	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "PUT",
		fmt.Sprintf("/api/projects/%d/documents/prd", proj.ID),
		`{"content":"new content"}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	data, _ := os.ReadFile(filepath.Join(wbDir, "prd.md"))
	if string(data) != "new content" {
		t.Errorf("file content = %q, want %q", string(data), "new content")
	}
}

func TestPutDocumentRejectsDecisions(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "PUT",
		fmt.Sprintf("/api/projects/%d/documents/decisions", proj.ID),
		`{"content":"should fail"}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestAppendDecisionNew(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "POST",
		fmt.Sprintf("/api/projects/%d/documents/decisions", proj.ID),
		`{"content":"We chose Go for the backend."}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	data, err := os.ReadFile(filepath.Join(projDir, ".waynebot", "decisions.md"))
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

func TestAppendDecisionExisting(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	wbDir := filepath.Join(projDir, ".waynebot")
	os.MkdirAll(wbDir, 0o755)
	os.WriteFile(filepath.Join(wbDir, "decisions.md"), []byte("## Existing\nFirst decision."), 0o644)

	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "POST",
		fmt.Sprintf("/api/projects/%d/documents/decisions", proj.ID),
		`{"content":"Second decision."}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	data, _ := os.ReadFile(filepath.Join(wbDir, "decisions.md"))
	content := string(data)
	if !strings.Contains(content, "First decision.") {
		t.Error("should preserve existing content")
	}
	if !strings.Contains(content, "Second decision.") {
		t.Error("should contain new decision")
	}
}

func TestAppendDecisionEmptyContent(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	projDir := t.TempDir()
	proj, err := model.CreateProject(d, "testproj", projDir, "test project")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}

	rec := doJSON(t, router, "POST",
		fmt.Sprintf("/api/projects/%d/documents/decisions", proj.ID),
		`{"content":"  "}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
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
