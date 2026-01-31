package model

import (
	"database/sql"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

type Reaction struct {
	ID         int64
	MessageID  int64
	AuthorID   int64
	AuthorType string
	Emoji      string
	CreatedAt  time.Time
}

// ReactionCount is an aggregated emoji count for a message.
type ReactionCount struct {
	Emoji   string `json:"emoji"`
	Count   int    `json:"count"`
	Reacted bool   `json:"reacted"`
}

// AddReaction inserts a reaction. Returns true if a new row was inserted, false if it already existed.
func AddReaction(d *db.DB, messageID, authorID int64, authorType, emoji string) (bool, error) {
	res, err := d.WriteExec(
		`INSERT OR IGNORE INTO reactions (message_id, author_id, author_type, emoji) VALUES (?, ?, ?, ?)`,
		messageID, authorID, authorType, emoji,
	)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// RemoveReaction deletes a reaction. Returns true if a row was deleted.
func RemoveReaction(d *db.DB, messageID, authorID int64, authorType, emoji string) (bool, error) {
	res, err := d.WriteExec(
		`DELETE FROM reactions WHERE message_id = ? AND author_id = ? AND author_type = ? AND emoji = ?`,
		messageID, authorID, authorType, emoji,
	)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// GetReactionCounts returns aggregated reaction counts for a single message.
// The reacted field indicates whether the given viewer has reacted with that emoji.
func GetReactionCounts(d *db.DB, messageID, viewerID int64, viewerType string) ([]ReactionCount, error) {
	rows, err := d.SQL.Query(
		`SELECT r.emoji,
		        COUNT(*) AS cnt,
		        MAX(CASE WHEN r.author_id = ? AND r.author_type = ? THEN 1 ELSE 0 END) AS reacted
		 FROM reactions r
		 WHERE r.message_id = ?
		 GROUP BY r.emoji
		 ORDER BY MIN(r.id)`,
		viewerID, viewerType, messageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReactionCounts(rows)
}

// GetReactionCountsBatch returns aggregated reaction counts for multiple messages.
// Returns a map of messageID -> []ReactionCount.
func GetReactionCountsBatch(d *db.DB, messageIDs []int64, viewerID int64, viewerType string) (map[int64][]ReactionCount, error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}

	// Build placeholders for the IN clause.
	placeholders := ""
	args := make([]any, 0, len(messageIDs)+2)
	args = append(args, viewerID, viewerType)
	for i, id := range messageIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}

	rows, err := d.SQL.Query(
		`SELECT r.message_id, r.emoji,
		        COUNT(*) AS cnt,
		        MAX(CASE WHEN r.author_id = ? AND r.author_type = ? THEN 1 ELSE 0 END) AS reacted
		 FROM reactions r
		 WHERE r.message_id IN (`+placeholders+`)
		 GROUP BY r.message_id, r.emoji
		 ORDER BY r.message_id, MIN(r.id)`,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]ReactionCount)
	for rows.Next() {
		var msgID int64
		var rc ReactionCount
		var reacted int
		if err := rows.Scan(&msgID, &rc.Emoji, &rc.Count, &reacted); err != nil {
			return nil, err
		}
		rc.Reacted = reacted == 1
		result[msgID] = append(result[msgID], rc)
	}
	return result, rows.Err()
}

// GetMessageChannelID returns the channel_id for a given message, or sql.ErrNoRows if not found.
func GetMessageChannelID(d *db.DB, messageID int64) (int64, error) {
	var channelID int64
	err := d.SQL.QueryRow("SELECT channel_id FROM messages WHERE id = ?", messageID).Scan(&channelID)
	return channelID, err
}

func scanReactionCounts(rows *sql.Rows) ([]ReactionCount, error) {
	var counts []ReactionCount
	for rows.Next() {
		var rc ReactionCount
		var reacted int
		if err := rows.Scan(&rc.Emoji, &rc.Count, &reacted); err != nil {
			return nil, err
		}
		rc.Reacted = reacted == 1
		counts = append(counts, rc)
	}
	return counts, rows.Err()
}
