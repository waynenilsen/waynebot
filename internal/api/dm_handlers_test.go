package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCreateDMWithUser(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	aliceToken := registerUser(t, router, "alice", "password123", "")

	// Create invite so bob can register.
	rec := doJSON(t, router, "POST", "/api/invites", `{}`, "Authorization", "Bearer "+aliceToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create invite: status=%d body=%s", rec.Code, rec.Body.String())
	}
	var invite struct {
		Code string `json:"code"`
	}
	json.NewDecoder(rec.Body).Decode(&invite)

	bobToken := registerUser(t, router, "bob", "password123", invite.Code)

	// Get bob's user ID.
	rec = doJSON(t, router, "GET", "/api/auth/me", "", "Authorization", "Bearer "+bobToken)
	var bobMe struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(rec.Body).Decode(&bobMe)

	// Alice creates a DM with bob.
	rec = doJSON(t, router, "POST", "/api/dms",
		fmt.Sprintf(`{"user_id":%d}`, bobMe.ID),
		"Authorization", "Bearer "+aliceToken)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ID               int64 `json:"id"`
		IsDM             bool  `json:"is_dm"`
		OtherParticipant struct {
			UserID   *int64  `json:"user_id"`
			UserName *string `json:"user_name"`
		} `json:"other_participant"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if !resp.IsDM {
		t.Error("expected is_dm = true")
	}
	if resp.OtherParticipant.UserID == nil || *resp.OtherParticipant.UserID != bobMe.ID {
		t.Errorf("other_participant.user_id = %v, want %d", resp.OtherParticipant.UserID, bobMe.ID)
	}
	if resp.OtherParticipant.UserName == nil || *resp.OtherParticipant.UserName != "bob" {
		t.Errorf("other_participant.user_name = %v, want bob", resp.OtherParticipant.UserName)
	}
}

func TestCreateDMWithUserIdempotent(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	aliceToken := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/invites", `{}`, "Authorization", "Bearer "+aliceToken)
	var invite struct {
		Code string `json:"code"`
	}
	json.NewDecoder(rec.Body).Decode(&invite)
	bobToken := registerUser(t, router, "bob", "password123", invite.Code)

	rec = doJSON(t, router, "GET", "/api/auth/me", "", "Authorization", "Bearer "+bobToken)
	var bobMe struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(rec.Body).Decode(&bobMe)

	body := fmt.Sprintf(`{"user_id":%d}`, bobMe.ID)

	// First call creates.
	rec = doJSON(t, router, "POST", "/api/dms", body, "Authorization", "Bearer "+aliceToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first call: status=%d, want 201", rec.Code)
	}
	var first struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(rec.Body).Decode(&first)

	// Second call returns existing.
	rec = doJSON(t, router, "POST", "/api/dms", body, "Authorization", "Bearer "+aliceToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("second call: status=%d, want 200", rec.Code)
	}
	var second struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(rec.Body).Decode(&second)

	if first.ID != second.ID {
		t.Errorf("idempotent check failed: first=%d second=%d", first.ID, second.ID)
	}
}

func TestCreateDMWithPersona(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	personaID := createPersona(t, router, token, "helper", "You help")

	rec := doJSON(t, router, "POST", "/api/dms",
		fmt.Sprintf(`{"persona_id":%d}`, personaID),
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ID               int64 `json:"id"`
		IsDM             bool  `json:"is_dm"`
		OtherParticipant struct {
			PersonaID   *int64  `json:"persona_id"`
			PersonaName *string `json:"persona_name"`
		} `json:"other_participant"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if !resp.IsDM {
		t.Error("expected is_dm = true")
	}
	if resp.OtherParticipant.PersonaID == nil || *resp.OtherParticipant.PersonaID != personaID {
		t.Errorf("other_participant.persona_id = %v, want %d", resp.OtherParticipant.PersonaID, personaID)
	}
	if resp.OtherParticipant.PersonaName == nil || *resp.OtherParticipant.PersonaName != "helper" {
		t.Errorf("other_participant.persona_name = %v, want helper", resp.OtherParticipant.PersonaName)
	}
}

func TestCreateDMValidation(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	tests := []struct {
		name   string
		body   string
		status int
	}{
		{"neither set", `{}`, http.StatusBadRequest},
		{"both set", `{"user_id":1,"persona_id":1}`, http.StatusBadRequest},
		{"self DM", `{"user_id":1}`, http.StatusBadRequest},
		{"nonexistent user", `{"user_id":999}`, http.StatusNotFound},
		{"nonexistent persona", `{"persona_id":999}`, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doJSON(t, router, "POST", "/api/dms", tt.body, "Authorization", "Bearer "+token)
			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d, body: %s", rec.Code, tt.status, rec.Body.String())
			}
		})
	}
}

func TestCreateDMUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	rec := doJSON(t, router, "POST", "/api/dms", `{"user_id":1}`)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestListDMsEmpty(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "GET", "/api/dms", "", "Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp []any
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d", len(resp))
	}
}

func TestListDMs(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	aliceToken := registerUser(t, router, "alice", "password123", "")

	// Create a persona DM.
	personaID := createPersona(t, router, aliceToken, "helper", "You help")
	rec := doJSON(t, router, "POST", "/api/dms",
		fmt.Sprintf(`{"persona_id":%d}`, personaID),
		"Authorization", "Bearer "+aliceToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create DM: status=%d body=%s", rec.Code, rec.Body.String())
	}

	// List DMs.
	rec = doJSON(t, router, "GET", "/api/dms", "", "Authorization", "Bearer "+aliceToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("list DMs: status=%d", rec.Code)
	}

	var dms []struct {
		ID               int64 `json:"id"`
		IsDM             bool  `json:"is_dm"`
		UnreadCount      int64 `json:"unread_count"`
		OtherParticipant struct {
			PersonaID   *int64  `json:"persona_id"`
			PersonaName *string `json:"persona_name"`
		} `json:"other_participant"`
	}
	json.NewDecoder(rec.Body).Decode(&dms)

	if len(dms) != 1 {
		t.Fatalf("expected 1 DM, got %d", len(dms))
	}
	if !dms[0].IsDM {
		t.Error("expected is_dm = true")
	}
	if dms[0].OtherParticipant.PersonaName == nil || *dms[0].OtherParticipant.PersonaName != "helper" {
		t.Errorf("other_participant.persona_name = %v, want helper", dms[0].OtherParticipant.PersonaName)
	}
}

func TestDMMessagesUseExistingEndpoints(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	personaID := createPersona(t, router, token, "helper", "You help")
	rec := doJSON(t, router, "POST", "/api/dms",
		fmt.Sprintf(`{"persona_id":%d}`, personaID),
		"Authorization", "Bearer "+token)
	var dm struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(rec.Body).Decode(&dm)

	// Post a message to the DM channel using existing message endpoint.
	rec = doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/messages", dm.ID),
		`{"content":"hello persona"}`,
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("post message: status=%d body=%s", rec.Code, rec.Body.String())
	}

	// Read messages back.
	rec = doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/messages", dm.ID), "",
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("get messages: status=%d", rec.Code)
	}

	var msgs []struct {
		Content string `json:"content"`
	}
	json.NewDecoder(rec.Body).Decode(&msgs)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "hello persona" {
		t.Errorf("content = %q, want hello persona", msgs[0].Content)
	}
}
