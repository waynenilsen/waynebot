package model

import (
	"database/sql"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

// WsTicket represents a single-use WebSocket connection ticket.
type WsTicket struct {
	Ticket    string
	UserID    int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

// CreateWsTicket creates a new WebSocket ticket that expires at the given time.
func CreateWsTicket(d *db.DB, ticket string, userID int64, expiresAt time.Time) (WsTicket, error) {
	var t WsTicket
	err := d.WriteTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			"INSERT INTO ws_tickets (ticket, user_id, expires_at) VALUES (?, ?, ?)",
			ticket, userID, expiresAt,
		)
		if err != nil {
			return err
		}
		return tx.QueryRow(
			"SELECT ticket, user_id, created_at, expires_at FROM ws_tickets WHERE ticket = ?", ticket,
		).Scan(&t.Ticket, &t.UserID, &t.CreatedAt, &t.ExpiresAt)
	})
	return t, err
}

// ClaimWsTicket atomically retrieves and deletes a valid (non-expired) ticket.
// Returns sql.ErrNoRows if the ticket doesn't exist or is expired.
func ClaimWsTicket(d *db.DB, ticket string) (WsTicket, error) {
	var t WsTicket
	err := d.WriteTx(func(tx *sql.Tx) error {
		err := tx.QueryRow(
			"SELECT ticket, user_id, created_at, expires_at FROM ws_tickets WHERE ticket = ? AND expires_at > ?",
			ticket, time.Now(),
		).Scan(&t.Ticket, &t.UserID, &t.CreatedAt, &t.ExpiresAt)
		if err != nil {
			return err
		}
		_, err = tx.Exec("DELETE FROM ws_tickets WHERE ticket = ?", ticket)
		return err
	})
	return t, err
}

// CleanupExpiredWsTickets removes all expired tickets.
func CleanupExpiredWsTickets(d *db.DB) (int64, error) {
	res, err := d.WriteExec("DELETE FROM ws_tickets WHERE expires_at < ?", time.Now())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
