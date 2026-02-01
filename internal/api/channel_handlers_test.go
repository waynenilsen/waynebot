package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func createChannel(t *testing.T, router http.Handler, token, name, description string) int64 {
	t.Helper()
	body := fmt.Sprintf(`{"name":%q,"description":%q}`, name, description)
	rec := doJSON(t, router, "POST", "/api/channels", body, "Authorization", "Bearer "+token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create channel %s: status=%d body=%s", name, rec.Code, rec.Body.String())
	}
	var resp struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp.ID
}

func TestListChannelsEmpty(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "GET", "/api/channels", "", "Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp []any
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 0 {
		t.Errorf("expected empty list, got %d", len(resp))
	}
}

func TestCreateChannel(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/channels",
		`{"name":"general","description":"Main channel"}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Name != "general" {
		t.Errorf("name = %q, want general", resp.Name)
	}
	if resp.Description != "Main channel" {
		t.Errorf("description = %q, want Main channel", resp.Description)
	}
}

func TestCreateChannelDuplicate(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	createChannel(t, router, token, "general", "first")

	rec := doJSON(t, router, "POST", "/api/channels",
		`{"name":"general","description":"second"}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
}

func TestCreateChannelValidation(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	tests := []struct {
		name string
		body string
	}{
		{"empty name", `{"name":"","description":"ok"}`},
		{"name too long", `{"name":"` + strings.Repeat("x", 101) + `","description":"ok"}`},
		{"description too long", `{"name":"ok","description":"` + strings.Repeat("x", 501) + `"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doJSON(t, router, "POST", "/api/channels", tt.body, "Authorization", "Bearer "+token)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestCreateChannelUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	rec := doJSON(t, router, "POST", "/api/channels", `{"name":"general","description":""}`)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestListChannels(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	createChannel(t, router, token, "general", "General chat")
	createChannel(t, router, token, "random", "Random stuff")

	rec := doJSON(t, router, "GET", "/api/channels", "", "Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var resp []struct {
		Name string `json:"name"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(resp))
	}
	if resp[0].Name != "general" {
		t.Errorf("first channel = %q, want general", resp[0].Name)
	}
}

func TestPostMessage(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	rec := doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/messages", chID),
		`{"content":"hello world"}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		ID         int64  `json:"id"`
		Content    string `json:"content"`
		AuthorName string `json:"author_name"`
		AuthorType string `json:"author_type"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Content != "hello world" {
		t.Errorf("content = %q, want hello world", resp.Content)
	}
	if resp.AuthorName != "alice" {
		t.Errorf("author_name = %q, want alice", resp.AuthorName)
	}
	if resp.AuthorType != "human" {
		t.Errorf("author_type = %q, want human", resp.AuthorType)
	}
}

func TestPostMessageValidation(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	tests := []struct {
		name string
		body string
	}{
		{"empty content", `{"content":""}`},
		{"content too long", `{"content":"` + strings.Repeat("x", 10001) + `"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/messages", chID),
				tt.body, "Authorization", "Bearer "+token)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestPostMessageToNonexistentChannel(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/channels/999/messages",
		`{"content":"hello"}`,
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestGetMessages(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	// Post a few messages.
	for i := range 3 {
		doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/messages", chID),
			fmt.Sprintf(`{"content":"msg %d"}`, i),
			"Authorization", "Bearer "+token)
	}

	rec := doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/messages", chID), "",
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var msgs []struct {
		ID      int64  `json:"id"`
		Content string `json:"content"`
	}
	json.NewDecoder(rec.Body).Decode(&msgs)
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	// Newest first.
	if msgs[0].Content != "msg 2" {
		t.Errorf("first message = %q, want msg 2", msgs[0].Content)
	}
}

func TestGetMessagesPagination(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	// Post 5 messages.
	for i := range 5 {
		doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/messages", chID),
			fmt.Sprintf(`{"content":"msg %d"}`, i),
			"Authorization", "Bearer "+token)
	}

	// Get first page with limit=2.
	rec := doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/messages?limit=2", chID), "",
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var page1 []struct {
		ID      int64  `json:"id"`
		Content string `json:"content"`
	}
	json.NewDecoder(rec.Body).Decode(&page1)
	if len(page1) != 2 {
		t.Fatalf("page1: expected 2, got %d", len(page1))
	}

	// Get second page using before cursor.
	lastID := page1[len(page1)-1].ID
	rec = doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/messages?limit=2&before=%d", chID, lastID), "",
		"Authorization", "Bearer "+token)

	var page2 []struct {
		ID      int64  `json:"id"`
		Content string `json:"content"`
	}
	json.NewDecoder(rec.Body).Decode(&page2)
	if len(page2) != 2 {
		t.Fatalf("page2: expected 2, got %d", len(page2))
	}

	// Pages should not overlap.
	if page1[0].ID == page2[0].ID {
		t.Error("pages should not overlap")
	}
}

func TestGetMessagesNonexistentChannel(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "GET", "/api/channels/999/messages", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestGetMessagesEmptyChannel(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "empty", "")

	rec := doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/messages", chID), "",
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var msgs []any
	json.NewDecoder(rec.Body).Decode(&msgs)
	if len(msgs) != 0 {
		t.Errorf("expected empty list, got %d", len(msgs))
	}
}

func TestListChannelsOnlyShowsMemberChannels(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	aliceToken := registerUser(t, router, "alice", "password123", "")

	// Use an invite code from alice to register bob.
	rec := doJSON(t, router, "POST", "/api/invites", `{}`, "Authorization", "Bearer "+aliceToken)
	var inv struct {
		Code string `json:"code"`
	}
	json.NewDecoder(rec.Body).Decode(&inv)
	bobToken := registerUser(t, router, "bob", "password123", inv.Code)

	// Alice creates two channels (she's auto-added as owner).
	createChannel(t, router, aliceToken, "general", "")
	createChannel(t, router, aliceToken, "random", "")

	// Bob should see zero channels.
	rec = doJSON(t, router, "GET", "/api/channels", "", "Authorization", "Bearer "+bobToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var bobChannels []any
	json.NewDecoder(rec.Body).Decode(&bobChannels)
	if len(bobChannels) != 0 {
		t.Errorf("bob expected 0 channels, got %d", len(bobChannels))
	}

	// Alice should see two channels.
	rec = doJSON(t, router, "GET", "/api/channels", "", "Authorization", "Bearer "+aliceToken)
	var aliceChannels []any
	json.NewDecoder(rec.Body).Decode(&aliceChannels)
	if len(aliceChannels) != 2 {
		t.Errorf("alice expected 2 channels, got %d", len(aliceChannels))
	}
}

func TestNonMemberGetMessagesForbidden(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	aliceToken := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/invites", `{}`, "Authorization", "Bearer "+aliceToken)
	var inv struct {
		Code string `json:"code"`
	}
	json.NewDecoder(rec.Body).Decode(&inv)
	bobToken := registerUser(t, router, "bob", "password123", inv.Code)

	chID := createChannel(t, router, aliceToken, "general", "")

	rec = doJSON(t, router, "GET", fmt.Sprintf("/api/channels/%d/messages", chID), "",
		"Authorization", "Bearer "+bobToken)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestNonMemberPostMessageForbidden(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	aliceToken := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/invites", `{}`, "Authorization", "Bearer "+aliceToken)
	var inv struct {
		Code string `json:"code"`
	}
	json.NewDecoder(rec.Body).Decode(&inv)
	bobToken := registerUser(t, router, "bob", "password123", inv.Code)

	chID := createChannel(t, router, aliceToken, "general", "")

	rec = doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/messages", chID),
		`{"content":"sneaky"}`,
		"Authorization", "Bearer "+bobToken)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestNonMemberMarkReadForbidden(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	aliceToken := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/invites", `{}`, "Authorization", "Bearer "+aliceToken)
	var inv struct {
		Code string `json:"code"`
	}
	json.NewDecoder(rec.Body).Decode(&inv)
	bobToken := registerUser(t, router, "bob", "password123", inv.Code)

	chID := createChannel(t, router, aliceToken, "general", "")

	rec = doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/read", chID), "",
		"Authorization", "Bearer "+bobToken)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}
