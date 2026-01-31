package ws_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/waynenilsen/waynebot/internal/ws"
)

func TestHubRegisterUnregister(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	c := ws.NewTestClient(hub, 1)
	hub.Register(c)
	waitFor(t, func() bool { return hub.ClientCount() == 1 })

	hub.Unregister(c)
	waitFor(t, func() bool { return hub.ClientCount() == 0 })
}

func TestHubBroadcast(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	c1 := ws.NewTestClient(hub, 1)
	c2 := ws.NewTestClient(hub, 2)
	hub.Register(c1)
	hub.Register(c2)
	waitFor(t, func() bool { return hub.ClientCount() == 2 })

	hub.Broadcast(ws.Event{Type: "test", Data: "hello"})

	msg1 := recvFrom(t, c1)
	msg2 := recvFrom(t, c2)

	var ev1, ev2 ws.Event
	json.Unmarshal(msg1, &ev1)
	json.Unmarshal(msg2, &ev2)

	if ev1.Type != "test" {
		t.Errorf("client1 event type = %q, want test", ev1.Type)
	}
	if ev2.Type != "test" {
		t.Errorf("client2 event type = %q, want test", ev2.Type)
	}
}

func TestHubNotifyChan(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	// Register a client so broadcast has a recipient.
	c := ws.NewTestClient(hub, 1)
	hub.Register(c)
	waitFor(t, func() bool { return hub.ClientCount() == 1 })

	// Drain any existing signals.
	select {
	case <-hub.NotifyChan:
	default:
	}

	hub.Broadcast(ws.Event{Type: "msg", Data: "hi"})

	select {
	case <-hub.NotifyChan:
		// Expected.
	case <-time.After(time.Second):
		t.Error("expected signal on NotifyChan")
	}

	// Drain the client send channel.
	recvFrom(t, c)
}

func TestHubClientCount(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	if hub.ClientCount() != 0 {
		t.Errorf("initial client count = %d, want 0", hub.ClientCount())
	}

	c1 := ws.NewTestClient(hub, 1)
	c2 := ws.NewTestClient(hub, 2)
	c3 := ws.NewTestClient(hub, 3)

	hub.Register(c1)
	hub.Register(c2)
	hub.Register(c3)
	waitFor(t, func() bool { return hub.ClientCount() == 3 })

	hub.Unregister(c2)
	waitFor(t, func() bool { return hub.ClientCount() == 2 })
}

func TestHubStop(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()

	c := ws.NewTestClient(hub, 1)
	hub.Register(c)
	waitFor(t, func() bool { return hub.ClientCount() == 1 })

	hub.Stop()

	// After stop, client send channel should be closed.
	select {
	case _, ok := <-c.SendChan():
		if ok {
			t.Error("expected send channel to be closed after hub stop")
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for send channel close")
	}
}

// waitFor polls until cond returns true or times out.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("waitFor timed out")
}

// recvFrom reads a message from a test client's send channel.
func recvFrom(t *testing.T, c *ws.Client) []byte {
	t.Helper()
	select {
	case msg := <-c.SendChan():
		return msg
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for message")
		return nil
	}
}
