package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

type Channel struct {
	ID          int64
	Name        string
	Description string
	IsDM        bool
	CreatedBy   *int64
	CreatedAt   time.Time
}

// DMParticipant represents one side of a DM conversation.
// Exactly one of UserID or PersonaID is set.
type DMParticipant struct {
	UserID    *int64
	PersonaID *int64
}

func scanChannel(s interface{ Scan(...any) error }) (Channel, error) {
	var ch Channel
	err := s.Scan(&ch.ID, &ch.Name, &ch.Description, &ch.IsDM, &ch.CreatedBy, &ch.CreatedAt)
	return ch, err
}

const channelCols = "id, name, description, is_dm, created_by, created_at"

func CreateChannel(d *db.DB, name, description string) (Channel, error) {
	var ch Channel
	err := d.WriteTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(
			"INSERT INTO channels (name, description) VALUES (?, ?)",
			name, description,
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		return tx.QueryRow(
			"SELECT "+channelCols+" FROM channels WHERE id = ?", id,
		).Scan(&ch.ID, &ch.Name, &ch.Description, &ch.IsDM, &ch.CreatedBy, &ch.CreatedAt)
	})
	return ch, err
}

func GetChannel(d *db.DB, id int64) (Channel, error) {
	return scanChannel(d.SQL.QueryRow(
		"SELECT "+channelCols+" FROM channels WHERE id = ?", id,
	))
}

func GetChannelByName(d *db.DB, name string) (Channel, error) {
	return scanChannel(d.SQL.QueryRow(
		"SELECT "+channelCols+" FROM channels WHERE name = ?", name,
	))
}

func UpdateChannel(d *db.DB, id int64, name, description string) error {
	_, err := d.WriteExec(
		"UPDATE channels SET name = ?, description = ? WHERE id = ?",
		name, description, id,
	)
	return err
}

func DeleteChannel(d *db.DB, id int64) error {
	_, err := d.WriteExec("DELETE FROM channels WHERE id = ?", id)
	return err
}

// ListChannels returns all non-DM channels.
func ListChannels(d *db.DB) ([]Channel, error) {
	rows, err := d.SQL.Query(
		"SELECT " + channelCols + " FROM channels WHERE is_dm = 0 ORDER BY id",
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

// CreateDMChannel creates a DM channel between two participants.
// If either participant is a persona, it is auto-subscribed via persona_channels.
func CreateDMChannel(d *db.DB, name string, p1, p2 DMParticipant, createdBy int64) (Channel, error) {
	var ch Channel
	err := d.WriteTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(
			"INSERT INTO channels (name, description, is_dm, created_by) VALUES (?, '', 1, ?)",
			name, createdBy,
		)
		if err != nil {
			return fmt.Errorf("insert channel: %w", err)
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}

		for _, p := range []DMParticipant{p1, p2} {
			if _, err := tx.Exec(
				"INSERT INTO dm_participants (channel_id, user_id, persona_id) VALUES (?, ?, ?)",
				id, p.UserID, p.PersonaID,
			); err != nil {
				return fmt.Errorf("insert dm_participant: %w", err)
			}
			// Auto-subscribe personas to the channel.
			if p.PersonaID != nil {
				if _, err := tx.Exec(
					"INSERT OR IGNORE INTO persona_channels (persona_id, channel_id) VALUES (?, ?)",
					*p.PersonaID, id,
				); err != nil {
					return fmt.Errorf("auto-subscribe persona: %w", err)
				}
			}
		}

		return tx.QueryRow(
			"SELECT "+channelCols+" FROM channels WHERE id = ?", id,
		).Scan(&ch.ID, &ch.Name, &ch.Description, &ch.IsDM, &ch.CreatedBy, &ch.CreatedAt)
	})
	return ch, err
}

// FindDMChannel returns the existing DM channel between two participants, if any.
func FindDMChannel(d *db.DB, p1, p2 DMParticipant) (Channel, error) {
	return scanChannel(d.SQL.QueryRow(
		`SELECT c.`+channelCols+`
		 FROM channels c
		 JOIN dm_participants d1 ON d1.channel_id = c.id
		 JOIN dm_participants d2 ON d2.channel_id = c.id
		 WHERE c.is_dm = 1
		   AND d1.user_id IS ? AND d1.persona_id IS ?
		   AND d2.user_id IS ? AND d2.persona_id IS ?`,
		p1.UserID, p1.PersonaID, p2.UserID, p2.PersonaID,
	))
}

// ListDMsForUser returns all DM channels where the given user is a participant.
func ListDMsForUser(d *db.DB, userID int64) ([]Channel, error) {
	rows, err := d.SQL.Query(
		`SELECT c.`+channelCols+`
		 FROM channels c
		 JOIN dm_participants dp ON dp.channel_id = c.id
		 WHERE c.is_dm = 1 AND dp.user_id = ?
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
