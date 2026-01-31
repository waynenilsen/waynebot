package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/waynenilsen/waynebot/internal/api"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/ws"
)

func TestCreateWsTicket(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)
	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/ws/ticket", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Ticket    string `json:"ticket"`
		ExpiresAt string `json:"expires_at"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Ticket == "" {
		t.Error("expected non-empty ticket")
	}
	if resp.ExpiresAt == "" {
		t.Error("expected non-empty expires_at")
	}

	// Verify expiration is roughly 30 seconds from now.
	expires, err := time.Parse(time.RFC3339, resp.ExpiresAt)
	if err != nil {
		t.Fatalf("parse expires_at: %v", err)
	}
	diff := time.Until(expires)
	if diff < 25*time.Second || diff > 35*time.Second {
		t.Errorf("ticket expiry diff = %v, expected ~30s", diff)
	}
}

func TestCreateWsTicketUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	rec := doJSON(t, router, "POST", "/api/ws/ticket", "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestWebSocketUpgrade(t *testing.T) {
	d := openTestDB(t)
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	router := newTestRouterWithHub(t, d, hub)
	token := registerUser(t, router, "alice", "password123", "")

	// Get a ticket.
	rec := doJSON(t, router, "POST", "/api/ws/ticket", "",
		"Authorization", "Bearer "+token)
	if rec.Code != http.StatusCreated {
		t.Fatalf("ticket status = %d", rec.Code)
	}
	var ticketResp struct {
		Ticket string `json:"ticket"`
	}
	json.NewDecoder(rec.Body).Decode(&ticketResp)

	// Start the router as an HTTP server.
	server := httptest.NewServer(router)
	defer server.Close()

	// Upgrade to WebSocket with the ticket.
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?ticket=" + ticketResp.Ticket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Wait for the client to be registered.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hub.ClientCount() == 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if hub.ClientCount() != 1 {
		t.Fatalf("expected 1 connected client, got %d", hub.ClientCount())
	}
}

func TestWebSocketUpgradeInvalidTicket(t *testing.T) {
	d := openTestDB(t)
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	router := newTestRouterWithHub(t, d, hub)

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?ticket=bogus"
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Error("expected dial error for invalid ticket")
	}
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestWebSocketUpgradeMissingTicket(t *testing.T) {
	d := openTestDB(t)
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	router := newTestRouterWithHub(t, d, hub)

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Error("expected dial error for missing ticket")
	}
	if resp != nil && resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestWebSocketTicketSingleUse(t *testing.T) {
	d := openTestDB(t)
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	router := newTestRouterWithHub(t, d, hub)
	token := registerUser(t, router, "alice", "password123", "")

	// Get a ticket.
	rec := doJSON(t, router, "POST", "/api/ws/ticket", "",
		"Authorization", "Bearer "+token)
	var ticketResp struct {
		Ticket string `json:"ticket"`
	}
	json.NewDecoder(rec.Body).Decode(&ticketResp)

	server := httptest.NewServer(router)
	defer server.Close()

	// First use should succeed.
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?ticket=" + ticketResp.Ticket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("first dial: %v", err)
	}
	conn.Close()

	// Second use of the same ticket should fail.
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Error("expected dial error for reused ticket")
	}
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}

func TestWebSocketBroadcastOnPostMessage(t *testing.T) {
	d := openTestDB(t)
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	router := newTestRouterWithHub(t, d, hub)
	token := registerUser(t, router, "alice", "password123", "")
	chID := createChannel(t, router, token, "general", "")

	// Get a ticket and connect via WebSocket.
	rec := doJSON(t, router, "POST", "/api/ws/ticket", "",
		"Authorization", "Bearer "+token)
	var ticketResp struct {
		Ticket string `json:"ticket"`
	}
	json.NewDecoder(rec.Body).Decode(&ticketResp)

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws?ticket=" + ticketResp.Ticket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Wait for registration.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hub.ClientCount() == 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Post a message via the API.
	doJSON(t, router, "POST", fmt.Sprintf("/api/channels/%d/messages", chID),
		`{"content":"hello via ws"}`,
		"Authorization", "Bearer "+token)

	// Read the broadcast from the WebSocket.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read ws message: %v", err)
	}

	var ev struct {
		Type string `json:"type"`
		Data struct {
			Content    string `json:"content"`
			AuthorName string `json:"author_name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg, &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ev.Type != "new_message" {
		t.Errorf("event type = %q, want new_message", ev.Type)
	}
	if ev.Data.Content != "hello via ws" {
		t.Errorf("content = %q, want hello via ws", ev.Data.Content)
	}
	if ev.Data.AuthorName != "alice" {
		t.Errorf("author_name = %q, want alice", ev.Data.AuthorName)
	}
}

// newTestRouterWithHub creates a test router with a specific hub instance.
func newTestRouterWithHub(t *testing.T, d *db.DB, hub *ws.Hub) http.Handler {
	t.Helper()
	return api.NewRouter(d, []string{"*"}, hub)
}
