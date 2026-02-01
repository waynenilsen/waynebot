package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestListMembersEmpty(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	rec := doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/members", chID), "",
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

func TestAddAndListUserMember(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	// Get alice's user ID from /auth/me
	meRec := doJSON(t, router, "GET", "/api/auth/me", "", "Authorization", "Bearer "+token)
	var me struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(meRec.Body).Decode(&me)

	// Add alice as member
	body := fmt.Sprintf(`{"user_id":%d}`, me.ID)
	rec := doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/members", chID), body,
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("add member: status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}

	// List members
	rec = doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/members", chID), "",
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: status = %d, want 200", rec.Code)
	}

	var members []struct {
		Type string `json:"type"`
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Role string `json:"role"`
	}
	json.NewDecoder(rec.Body).Decode(&members)
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].Type != "user" {
		t.Errorf("type = %q, want user", members[0].Type)
	}
	if members[0].Name != "alice" {
		t.Errorf("name = %q, want alice", members[0].Name)
	}
	if members[0].Role != "member" {
		t.Errorf("role = %q, want member", members[0].Role)
	}
}

func TestAddPersonaMember(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")
	personaID := createPersona(t, router, token, "helper", "You help.")

	body := fmt.Sprintf(`{"persona_id":%d}`, personaID)
	rec := doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/members", chID), body,
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("add persona: status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}

	// List should show the persona
	rec = doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/members", chID), "",
		"Authorization", "Bearer "+token)
	var members []struct {
		Type string `json:"type"`
		Name string `json:"name"`
		Role string `json:"role"`
	}
	json.NewDecoder(rec.Body).Decode(&members)
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].Type != "persona" {
		t.Errorf("type = %q, want persona", members[0].Type)
	}
	if members[0].Name != "helper" {
		t.Errorf("name = %q, want helper", members[0].Name)
	}
}

func TestRemoveUserMember(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	meRec := doJSON(t, router, "GET", "/api/auth/me", "", "Authorization", "Bearer "+token)
	var me struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(meRec.Body).Decode(&me)

	// Add then remove
	doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/members", chID),
		fmt.Sprintf(`{"user_id":%d}`, me.ID), "Authorization", "Bearer "+token)

	rec := doJSON(t, router, "DELETE", fmt.Sprintf("/api/channels/%d/members", chID),
		fmt.Sprintf(`{"user_id":%d}`, me.ID), "Authorization", "Bearer "+token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("remove: status = %d, want 204, body: %s", rec.Code, rec.Body.String())
	}

	// List should be empty
	rec = doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/members", chID), "",
		"Authorization", "Bearer "+token)
	var members []any
	json.NewDecoder(rec.Body).Decode(&members)
	if len(members) != 0 {
		t.Errorf("expected 0 members after remove, got %d", len(members))
	}
}

func TestRemovePersonaMember(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")
	personaID := createPersona(t, router, token, "helper", "You help.")

	// Add then remove
	doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/members", chID),
		fmt.Sprintf(`{"persona_id":%d}`, personaID), "Authorization", "Bearer "+token)

	rec := doJSON(t, router, "DELETE", fmt.Sprintf("/api/channels/%d/members", chID),
		fmt.Sprintf(`{"persona_id":%d}`, personaID), "Authorization", "Bearer "+token)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("remove persona: status = %d, want 204", rec.Code)
	}

	rec = doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/members", chID), "",
		"Authorization", "Bearer "+token)
	var members []any
	json.NewDecoder(rec.Body).Decode(&members)
	if len(members) != 0 {
		t.Errorf("expected 0 members after remove, got %d", len(members))
	}
}

func TestAddMemberBadRequest(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	tests := []struct {
		name string
		body string
	}{
		{"neither id", `{}`},
		{"both ids", `{"user_id":1,"persona_id":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/members", chID),
				tt.body, "Authorization", "Bearer "+token)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestMembersChannelNotFound(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "GET", "/api/channels/999/members", "",
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestMembersUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	rec := doJSON(t, router, "GET", "/api/channels/1/members", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
