package model

import (
	"database/sql"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

type Message struct {
	ID         int64
	ChannelID  int64
	AuthorID   int64
	AuthorType string
	AuthorName string
	Content    string
	CreatedAt  time.Time
}

func CreateMessage(d *db.DB, channelID, authorID int64, authorType, authorName, content string) (Message, error) {
	var m Message
	err := d.WriteTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(
			"INSERT INTO messages (channel_id, author_id, author_type, author_name, content) VALUES (?, ?, ?, ?, ?)",
			channelID, authorID, authorType, authorName, content,
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		return tx.QueryRow(
			"SELECT id, channel_id, author_id, author_type, author_name, content, created_at FROM messages WHERE id = ?", id,
		).Scan(&m.ID, &m.ChannelID, &m.AuthorID, &m.AuthorType, &m.AuthorName, &m.Content, &m.CreatedAt)
	})
	return m, err
}

// GetMessagesBefore returns up to `limit` messages in a channel with id < beforeID, ordered newest-first.
// This implements cursor-based pagination: ?before=messageId
func GetMessagesBefore(d *db.DB, channelID, beforeID int64, limit int) ([]Message, error) {
	rows, err := d.SQL.Query(
		"SELECT id, channel_id, author_id, author_type, author_name, content, created_at FROM messages WHERE channel_id = ? AND id < ? ORDER BY id DESC LIMIT ?",
		channelID, beforeID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

// GetMessagesSince returns messages in a channel with id > afterID, ordered oldest-first.
func GetMessagesSince(d *db.DB, channelID, afterID int64) ([]Message, error) {
	rows, err := d.SQL.Query(
		"SELECT id, channel_id, author_id, author_type, author_name, content, created_at FROM messages WHERE channel_id = ? AND id > ? ORDER BY id ASC",
		channelID, afterID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

// GetRecentMessages returns the most recent `limit` messages in a channel, ordered newest-first.
func GetRecentMessages(d *db.DB, channelID int64, limit int) ([]Message, error) {
	rows, err := d.SQL.Query(
		"SELECT id, channel_id, author_id, author_type, author_name, content, created_at FROM messages WHERE channel_id = ? ORDER BY id DESC LIMIT ?",
		channelID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func scanMessages(rows *sql.Rows) ([]Message, error) {
	var msgs []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ChannelID, &m.AuthorID, &m.AuthorType, &m.AuthorName, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}
