package model

import (
	"github.com/waynenilsen/waynebot/internal/db"
)

// UpdateReadPosition sets the user's last-read message ID for a channel.
// Uses INSERT OR REPLACE (upsert) to create or update the position.
func UpdateReadPosition(d *db.DB, userID, channelID, messageID int64) error {
	_, err := d.WriteExec(
		`INSERT INTO read_positions (user_id, channel_id, last_read_message_id, updated_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(user_id, channel_id)
		 DO UPDATE SET last_read_message_id = excluded.last_read_message_id,
		               updated_at = CURRENT_TIMESTAMP`,
		userID, channelID, messageID,
	)
	return err
}

// UnreadCount holds the unread count for a single channel.
type UnreadCount struct {
	ChannelID int64
	Count     int64
}

// GetUnreadCounts returns the unread message count for each channel for a user.
func GetUnreadCounts(d *db.DB, userID int64) (map[int64]int64, error) {
	rows, err := d.SQL.Query(
		`SELECT c.id,
		        COUNT(m.id) as unread_count
		 FROM channels c
		 LEFT JOIN read_positions rp ON rp.channel_id = c.id AND rp.user_id = ?
		 LEFT JOIN messages m ON m.channel_id = c.id AND m.id > COALESCE(rp.last_read_message_id, 0)
		 GROUP BY c.id`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[int64]int64)
	for rows.Next() {
		var channelID, count int64
		if err := rows.Scan(&channelID, &count); err != nil {
			return nil, err
		}
		if count > 0 {
			counts[channelID] = count
		}
	}
	return counts, rows.Err()
}

// GetLatestMessageID returns the ID of the latest message in a channel, or 0 if none.
func GetLatestMessageID(d *db.DB, channelID int64) (int64, error) {
	var id int64
	err := d.SQL.QueryRow(
		"SELECT COALESCE(MAX(id), 0) FROM messages WHERE channel_id = ?",
		channelID,
	).Scan(&id)
	return id, err
}
