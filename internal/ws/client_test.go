package ws_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/waynenilsen/waynebot/internal/ws"
)

func TestClientReadWritePumps(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	// Start a test HTTP server that upgrades to WebSocket.
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade: %v", err)
			return
		}
		client := ws.NewClient(hub, conn, 42)
		hub.Register(client)
		go client.WritePump()
		go client.ReadPump()
	}))
	defer server.Close()

	// Connect a websocket client.
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Wait for registration.
	waitFor(t, func() bool { return hub.ClientCount() == 1 })

	// Broadcast an event and verify the client receives it.
	hub.Broadcast(ws.Event{Type: "test_event", Data: "payload"})

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var ev ws.Event
	if err := json.Unmarshal(msg, &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.Type != "test_event" {
		t.Errorf("event type = %q, want test_event", ev.Type)
	}
}

func TestClientDisconnectUnregisters(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := ws.NewClient(hub, conn, 1)
		hub.Register(client)
		go client.WritePump()
		go client.ReadPump()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	waitFor(t, func() bool { return hub.ClientCount() == 1 })

	// Close the client connection.
	conn.Close()

	// The hub should eventually unregister the client.
	waitFor(t, func() bool { return hub.ClientCount() == 0 })
}
