package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func createPersona(t *testing.T, router http.Handler, token, name, systemPrompt string) int64 {
	t.Helper()
	body := fmt.Sprintf(`{"name":%q,"system_prompt":%q,"model":"gpt-4","tools_enabled":[],"temperature":0.7,"max_tokens":1000,"cooldown_secs":5,"max_tokens_per_hour":10000}`, name, systemPrompt)
	rec := doJSON(t, router, "POST", "/api/personas", body, "Authorization", "Bearer "+token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create persona %s: status=%d body=%s", name, rec.Code, rec.Body.String())
	}
	var resp struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp.ID
}

func TestListPersonasEmpty(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "GET", "/api/personas", "", "Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp []any
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d", len(resp))
	}
}

func TestCreatePersona(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/personas",
		`{"name":"helper","system_prompt":"You are a helpful assistant.","model":"gpt-4","tools_enabled":["search"],"temperature":0.7,"max_tokens":1000,"cooldown_secs":5,"max_tokens_per_hour":10000}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ID           int64    `json:"id"`
		Name         string   `json:"name"`
		SystemPrompt string   `json:"system_prompt"`
		ToolsEnabled []string `json:"tools_enabled"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Name != "helper" {
		t.Errorf("name = %q, want helper", resp.Name)
	}
	if resp.SystemPrompt != "You are a helpful assistant." {
		t.Errorf("system_prompt = %q", resp.SystemPrompt)
	}
	if len(resp.ToolsEnabled) != 1 || resp.ToolsEnabled[0] != "search" {
		t.Errorf("tools_enabled = %v, want [search]", resp.ToolsEnabled)
	}
}

func TestCreatePersonaDuplicate(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	createPersona(t, router, token, "helper", "You help.")

	rec := doJSON(t, router, "POST", "/api/personas",
		`{"name":"helper","system_prompt":"Different prompt.","model":"gpt-4","tools_enabled":[],"temperature":0.7,"max_tokens":1000,"cooldown_secs":5,"max_tokens_per_hour":10000}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
}

func TestCreatePersonaValidation(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	tests := []struct {
		name string
		body string
	}{
		{"empty name", `{"name":"","system_prompt":"valid","model":"gpt-4","tools_enabled":[],"temperature":0.7,"max_tokens":1000,"cooldown_secs":5,"max_tokens_per_hour":10000}`},
		{"name too long", `{"name":"` + strings.Repeat("x", 101) + `","system_prompt":"valid","model":"gpt-4","tools_enabled":[],"temperature":0.7,"max_tokens":1000,"cooldown_secs":5,"max_tokens_per_hour":10000}`},
		{"empty system_prompt", `{"name":"valid","system_prompt":"","model":"gpt-4","tools_enabled":[],"temperature":0.7,"max_tokens":1000,"cooldown_secs":5,"max_tokens_per_hour":10000}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doJSON(t, router, "POST", "/api/personas", tt.body, "Authorization", "Bearer "+token)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestCreatePersonaUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	rec := doJSON(t, router, "POST", "/api/personas",
		`{"name":"helper","system_prompt":"You help.","model":"gpt-4","tools_enabled":[],"temperature":0.7,"max_tokens":1000,"cooldown_secs":5,"max_tokens_per_hour":10000}`)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestUpdatePersona(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	id := createPersona(t, router, token, "helper", "You help.")

	rec := doJSON(t, router, "PUT", fmt.Sprintf("/api/personas/%d", id),
		`{"name":"updated","system_prompt":"New prompt.","model":"gpt-4o","tools_enabled":["code"],"temperature":0.5,"max_tokens":2000,"cooldown_secs":10,"max_tokens_per_hour":5000}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Name         string `json:"name"`
		SystemPrompt string `json:"system_prompt"`
		Model        string `json:"model"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Name != "updated" {
		t.Errorf("name = %q, want updated", resp.Name)
	}
	if resp.Model != "gpt-4o" {
		t.Errorf("model = %q, want gpt-4o", resp.Model)
	}
}

func TestUpdatePersonaNotFound(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "PUT", "/api/personas/999",
		`{"name":"x","system_prompt":"x","model":"gpt-4","tools_enabled":[],"temperature":0.7,"max_tokens":1000,"cooldown_secs":5,"max_tokens_per_hour":10000}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestDeletePersona(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	id := createPersona(t, router, token, "helper", "You help.")

	rec := doJSON(t, router, "DELETE", fmt.Sprintf("/api/personas/%d", id), "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	// Verify it's gone.
	rec = doJSON(t, router, "GET", "/api/personas", "", "Authorization", "Bearer "+token)
	var resp []any
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty list after delete, got %d", len(resp))
	}
}

func TestDeletePersonaNotFound(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "DELETE", "/api/personas/999", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestListPersonas(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	createPersona(t, router, token, "helper", "You help.")
	createPersona(t, router, token, "coder", "You code.")

	rec := doJSON(t, router, "GET", "/api/personas", "", "Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp []struct {
		Name string `json:"name"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Fatalf("expected 2 personas, got %d", len(resp))
	}
}
