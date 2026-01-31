package connector

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// mockIMAP implements IMAPClient for testing.
type mockIMAP struct {
	mu         sync.Mutex
	connected  bool
	connectErr error
	fetchCalls int
	messages   []EmailMessage // returned once then cleared
	fetchErr   error
	seenUIDs   []uint32
	markErr    error
}

func (m *mockIMAP) Connect(host string, port int, user, pass string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.connectErr != nil {
		return m.connectErr
	}
	m.connected = true
	return nil
}

func (m *mockIMAP) FetchUnseen() ([]EmailMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fetchCalls++
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	msgs := m.messages
	m.messages = nil // only return once
	return msgs, nil
}

func (m *mockIMAP) MarkSeen(uids []uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.markErr != nil {
		return m.markErr
	}
	m.seenUIDs = append(m.seenUIDs, uids...)
	return nil
}

func (m *mockIMAP) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	return nil
}

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func startTestHub(t *testing.T) *ws.Hub {
	t.Helper()
	hub := ws.NewHub()
	go hub.Run()
	t.Cleanup(func() { hub.Stop() })
	return hub
}

type emailScenario struct {
	t       *testing.T
	db      *db.DB
	hub     *ws.Hub
	mock    *mockIMAP
	channel model.Channel
	conn    *EmailConnector
}

func newEmailScenario(t *testing.T) *emailScenario {
	t.Helper()
	d := openTestDB(t)
	hub := startTestHub(t)

	ch, err := model.CreateChannel(d, "email-inbox", "incoming email")
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}

	mock := &mockIMAP{}
	cfg := EmailConfig{
		Host:      "imap.test.com",
		Port:      993,
		User:      "user@test.com",
		Pass:      "secret",
		ChannelID: ch.ID,
		PollEvery: 50 * time.Millisecond,
	}

	conn := NewEmailConnector(cfg, mock, d, hub)

	return &emailScenario{
		t:       t,
		db:      d,
		hub:     hub,
		mock:    mock,
		channel: ch,
		conn:    conn,
	}
}

func TestEmailConnectorPostsMessages(t *testing.T) {
	s := newEmailScenario(t)

	s.mock.messages = []EmailMessage{
		{UID: 1, From: "alice@example.com", Subject: "Hello", Body: "Hi there", Date: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)},
		{UID: 2, From: "bob@example.com", Subject: "Meeting", Body: "Let's meet at 3pm"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		s.conn.Run(ctx)
		close(done)
	}()

	// Wait for messages to be posted.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		msgs, _ := model.GetRecentMessages(s.db, s.channel.ID, 10)
		if len(msgs) >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	<-done

	msgs, err := model.GetRecentMessages(s.db, s.channel.ID, 10)
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}

	// Messages are returned newest-first.
	if msgs[0].AuthorName != "bob@example.com" {
		t.Errorf("expected bob, got %s", msgs[0].AuthorName)
	}
	if msgs[1].AuthorName != "alice@example.com" {
		t.Errorf("expected alice, got %s", msgs[1].AuthorName)
	}
	if msgs[0].AuthorType != "connector" {
		t.Errorf("expected author_type connector, got %s", msgs[0].AuthorType)
	}

	// Verify UIDs were marked seen.
	s.mock.mu.Lock()
	seen := s.mock.seenUIDs
	s.mock.mu.Unlock()
	if len(seen) != 2 || seen[0] != 1 || seen[1] != 2 {
		t.Errorf("expected seen UIDs [1 2], got %v", seen)
	}
}

func TestEmailConnectorReconnectsOnFetchError(t *testing.T) {
	s := newEmailScenario(t)

	// First connection: fetch fails.
	s.mock.fetchErr = fmt.Errorf("connection reset")

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		s.conn.Run(ctx)
		close(done)
	}()

	// Wait for at least one fetch attempt + reconnect.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s.mock.mu.Lock()
		calls := s.mock.fetchCalls
		s.mock.mu.Unlock()
		if calls >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	<-done

	s.mock.mu.Lock()
	calls := s.mock.fetchCalls
	s.mock.mu.Unlock()
	if calls < 1 {
		t.Fatalf("expected at least 1 fetch call, got %d", calls)
	}
}

func TestEmailConnectorName(t *testing.T) {
	s := newEmailScenario(t)
	if s.conn.Name() != "email(user@test.com)" {
		t.Errorf("unexpected name: %s", s.conn.Name())
	}
}

func TestEmailConnectorCancelDuringConnect(t *testing.T) {
	s := newEmailScenario(t)
	s.mock.connectErr = fmt.Errorf("refused")

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		s.conn.Run(ctx)
		close(done)
	}()

	// Give it a moment to attempt connect, then cancel.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("connector did not stop after cancel")
	}
}

func TestFormatEmail(t *testing.T) {
	em := EmailMessage{
		From:    "alice@example.com",
		Subject: "Test Subject",
		Body:    "  Hello World  ",
		Date:    time.Date(2025, 3, 15, 14, 30, 0, 0, time.UTC),
	}
	got := FormatEmail(em)

	if !strings.Contains(got, "**Test Subject**") {
		t.Errorf("missing subject in formatted output: %s", got)
	}
	if !strings.Contains(got, "Hello World") {
		t.Errorf("missing body in formatted output: %s", got)
	}
	if strings.Contains(got, "  Hello") {
		t.Errorf("body should be trimmed: %s", got)
	}
}

func TestFormatEmailNoDate(t *testing.T) {
	em := EmailMessage{
		Subject: "No Date",
		Body:    "body",
	}
	got := FormatEmail(em)
	if strings.Contains(got, "*") && !strings.Contains(got, "**") {
		// Should not contain a date line (italic).
		t.Errorf("should not have date line: %s", got)
	}
}

func TestDefaultPollEvery(t *testing.T) {
	d := openTestDB(t)
	hub := startTestHub(t)
	ch, _ := model.CreateChannel(d, "test", "")

	conn := NewEmailConnector(EmailConfig{ChannelID: ch.ID}, &mockIMAP{}, d, hub)
	if conn.cfg.PollEvery != 60*time.Second {
		t.Errorf("expected default 60s, got %v", conn.cfg.PollEvery)
	}
}
