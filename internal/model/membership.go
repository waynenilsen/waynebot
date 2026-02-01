package model

import (
	"database/sql"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

// ChannelMember represents a human user's membership in a non-DM channel.
type ChannelMember struct {
	ChannelID int64
	UserID    int64
	Username  string
	Role      string
	CreatedAt time.Time
}

// AddChannelMember adds a user to a channel. If the user is already a member,
// the call is a no-op (INSERT OR IGNORE).
func AddChannelMember(d *db.DB, channelID, userID int64, role string) error {
	_, err := d.WriteExec(
		"INSERT OR IGNORE INTO channel_members (channel_id, user_id, role) VALUES (?, ?, ?)",
		channelID, userID, role,
	)
	return err
}

// RemoveChannelMember removes a user from a channel.
func RemoveChannelMember(d *db.DB, channelID, userID int64) error {
	_, err := d.WriteExec(
		"DELETE FROM channel_members WHERE channel_id = ? AND user_id = ?",
		channelID, userID,
	)
	return err
}

// GetChannelMembers returns all members of a channel with their usernames.
func GetChannelMembers(d *db.DB, channelID int64) ([]ChannelMember, error) {
	rows, err := d.SQL.Query(
		`SELECT cm.channel_id, cm.user_id, u.username, cm.role, cm.created_at
		 FROM channel_members cm
		 JOIN users u ON u.id = cm.user_id
		 WHERE cm.channel_id = ?
		 ORDER BY cm.created_at`,
		channelID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []ChannelMember
	for rows.Next() {
		var m ChannelMember
		if err := rows.Scan(&m.ChannelID, &m.UserID, &m.Username, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// IsChannelMember checks whether a user is a member of a channel.
func IsChannelMember(d *db.DB, channelID, userID int64) (bool, error) {
	var count int
	err := d.SQL.QueryRow(
		"SELECT COUNT(*) FROM channel_members WHERE channel_id = ? AND user_id = ?",
		channelID, userID,
	).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return count > 0, nil
}

// ListChannelsForUser returns all non-DM channels where the user is a member.
func ListChannelsForUser(d *db.DB, userID int64) ([]Channel, error) {
	rows, err := d.SQL.Query(
		`SELECT c.id, c.name, c.description, c.is_dm, c.created_by, c.created_at
		 FROM channels c
		 JOIN channel_members cm ON cm.channel_id = c.id
		 WHERE cm.user_id = ? AND c.is_dm = 0
		 ORDER BY c.id`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}
