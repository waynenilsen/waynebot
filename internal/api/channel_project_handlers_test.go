package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func createProject(t *testing.T, router http.Handler, token, name, path, description string) int64 {
	t.Helper()
	body := fmt.Sprintf(`{"name":%q,"path":%q,"description":%q}`, name, path, description)
	rec := doJSON(t, router, "POST", "/api/projects", body, "Authorization", "Bearer "+token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create project %s: status=%d body=%s", name, rec.Code, rec.Body.String())
	}
	var resp struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp.ID
}

func TestListChannelProjectsEmpty(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	rec := doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/projects", chID), "",
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var resp []any
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d", len(resp))
	}
}

func TestAddAndListChannelProject(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")
	projID := createProject(t, router, token, "myproj", t.TempDir(), "desc")

	// Add project to channel.
	body := fmt.Sprintf(`{"project_id":%d}`, projID)
	rec := doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/projects", chID), body,
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("add: status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}

	// List projects for channel.
	rec = doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/projects", chID), "",
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: status = %d, want 200", rec.Code)
	}

	var projects []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	json.NewDecoder(rec.Body).Decode(&projects)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].Name != "myproj" {
		t.Errorf("name = %q, want myproj", projects[0].Name)
	}
}

func TestAddChannelProjectIdempotent(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")
	projID := createProject(t, router, token, "myproj", t.TempDir(), "")

	body := fmt.Sprintf(`{"project_id":%d}`, projID)
	doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/projects", chID), body,
		"Authorization", "Bearer "+token)

	// Adding again should still succeed.
	rec := doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/projects", chID), body,
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("duplicate add: status = %d, want 201", rec.Code)
	}

	// Should still have only one association.
	rec = doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/projects", chID), "",
		"Authorization", "Bearer "+token)
	var projects []any
	json.NewDecoder(rec.Body).Decode(&projects)
	if len(projects) != 1 {
		t.Errorf("expected 1 project after duplicate add, got %d", len(projects))
	}
}

func TestRemoveChannelProject(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")
	projID := createProject(t, router, token, "myproj", t.TempDir(), "")

	body := fmt.Sprintf(`{"project_id":%d}`, projID)
	doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/projects", chID), body,
		"Authorization", "Bearer "+token)

	rec := doJSON(t, router, "DELETE",
		fmt.Sprintf("/api/channels/%d/projects/%d", chID, projID), "",
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("remove: status = %d, want 204, body: %s", rec.Code, rec.Body.String())
	}

	// Should be empty now.
	rec = doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/projects", chID), "",
		"Authorization", "Bearer "+token)
	var projects []any
	json.NewDecoder(rec.Body).Decode(&projects)
	if len(projects) != 0 {
		t.Errorf("expected 0 projects after removal, got %d", len(projects))
	}
}

func TestAddChannelProjectNotFoundChannel(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/channels/999/projects", `{"project_id":1}`,
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestAddChannelProjectNotFoundProject(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	rec := doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/projects", chID),
		`{"project_id":999}`, "Authorization", "Bearer "+token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestAddChannelProjectMissingProjectID(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	rec := doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/projects", chID),
		`{}`, "Authorization", "Bearer "+token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestChannelProjectsUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	rec := doJSON(t, router, "GET", "/api/channels/1/projects", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
