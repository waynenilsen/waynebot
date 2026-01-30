package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateInvite(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/invites", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ID        int64  `json:"id"`
		Code      string `json:"code"`
		CreatedBy int64  `json:"created_by"`
		UsedBy    *int64 `json:"used_by"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Code == "" {
		t.Error("expected non-empty invite code")
	}
	if resp.UsedBy != nil {
		t.Errorf("used_by should be nil, got %v", resp.UsedBy)
	}
}

func TestCreateInviteUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	rec := doJSON(t, router, "POST", "/api/invites", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestListInvites(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	// Create two invites.
	doJSON(t, router, "POST", "/api/invites", "", "Authorization", "Bearer "+token)
	doJSON(t, router, "POST", "/api/invites", "", "Authorization", "Bearer "+token)

	rec := doJSON(t, router, "GET", "/api/invites", "", "Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp []struct {
		Code string `json:"code"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Errorf("expected 2 invites, got %d", len(resp))
	}
}

func TestListInvitesEmpty(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "GET", "/api/invites", "", "Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp []any
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d", len(resp))
	}
}

func TestInviteFlowEndToEnd(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	aliceToken := registerUser(t, router, "alice", "password123", "")

	// Alice creates an invite.
	rec := doJSON(t, router, "POST", "/api/invites", "", "Authorization", "Bearer "+aliceToken)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create invite: status=%d", rec.Code)
	}

	var invite struct {
		Code string `json:"code"`
	}
	json.NewDecoder(rec.Body).Decode(&invite)

	// Bob registers using the invite code.
	rec = doJSON(t, router, "POST", "/api/auth/register",
		`{"username":"bob","password":"password123","invite_code":"`+invite.Code+`"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register bob: status=%d body=%s", rec.Code, rec.Body.String())
	}
}
