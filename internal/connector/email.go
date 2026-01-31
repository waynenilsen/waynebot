package connector

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// EmailMessage represents a parsed email from the IMAP server.
type EmailMessage struct {
	UID     uint32
	From    string
	Subject string
	Body    string
	Date    time.Time
}

// IMAPClient abstracts IMAP operations for testability.
type IMAPClient interface {
	// Connect dials the IMAP server and authenticates.
	Connect(host string, port int, user, pass string) error
	// FetchUnseen returns all UNSEEN messages.
	FetchUnseen() ([]EmailMessage, error)
	// MarkSeen marks the given UIDs as \Seen.
	MarkSeen(uids []uint32) error
	// Close disconnects from the server.
	Close() error
}

// EmailConfig holds IMAP connection settings for the email connector.
type EmailConfig struct {
	Host      string
	Port      int
	User      string
	Pass      string
	ChannelID int64
	PollEvery time.Duration
}

// EmailConnector polls an IMAP inbox and posts new emails into a channel.
type EmailConnector struct {
	cfg    EmailConfig
	client IMAPClient
	db     *db.DB
	hub    *ws.Hub
}

// NewEmailConnector creates a connector that polls an IMAP mailbox.
func NewEmailConnector(cfg EmailConfig, client IMAPClient, d *db.DB, hub *ws.Hub) *EmailConnector {
	if cfg.PollEvery == 0 {
		cfg.PollEvery = 60 * time.Second
	}
	return &EmailConnector{
		cfg:    cfg,
		client: client,
		db:     d,
		hub:    hub,
	}
}

func (e *EmailConnector) Name() string {
	return fmt.Sprintf("email(%s)", e.cfg.User)
}

// Run connects to the IMAP server and polls for unseen messages until ctx
// is cancelled. On connection failure it retries with backoff.
func (e *EmailConnector) Run(ctx context.Context) {
	for {
		if err := e.client.Connect(e.cfg.Host, e.cfg.Port, e.cfg.User, e.cfg.Pass); err != nil {
			slog.Error("email connector: connect failed", "error", err)
			if !sleepCtx(ctx, 30*time.Second) {
				return
			}
			continue
		}

		e.pollLoop(ctx)
		e.client.Close()

		if ctx.Err() != nil {
			return
		}
		// Connection dropped, retry after brief pause.
		if !sleepCtx(ctx, 5*time.Second) {
			return
		}
	}
}

func (e *EmailConnector) pollLoop(ctx context.Context) {
	for {
		msgs, err := e.client.FetchUnseen()
		if err != nil {
			slog.Error("email connector: fetch unseen", "error", err)
			return // reconnect
		}

		if len(msgs) > 0 {
			e.postMessages(msgs)

			uids := make([]uint32, len(msgs))
			for i, m := range msgs {
				uids[i] = m.UID
			}
			if err := e.client.MarkSeen(uids); err != nil {
				slog.Error("email connector: mark seen", "error", err)
				return // reconnect
			}
		}

		if !sleepCtx(ctx, e.cfg.PollEvery) {
			return
		}
	}
}

func (e *EmailConnector) postMessages(msgs []EmailMessage) {
	for _, em := range msgs {
		content := FormatEmail(em)
		msg, err := model.CreateMessage(e.db, e.cfg.ChannelID, 0, "connector", em.From, content)
		if err != nil {
			slog.Error("email connector: create message", "error", err)
			continue
		}
		e.hub.Broadcast(ws.Event{
			Type: "new_message",
			Data: msg,
		})
	}
}

// FormatEmail renders an email message as channel-friendly text.
func FormatEmail(em EmailMessage) string {
	var b strings.Builder
	fmt.Fprintf(&b, "**%s**\n", em.Subject)
	if !em.Date.IsZero() {
		fmt.Fprintf(&b, "*%s*\n", em.Date.Format(time.RFC822))
	}
	b.WriteString("\n")
	b.WriteString(strings.TrimSpace(em.Body))
	return b.String()
}

// sleepCtx sleeps for d or until ctx is cancelled. Returns false if cancelled.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
