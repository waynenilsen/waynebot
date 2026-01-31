package agent

import (
	"github.com/waynenilsen/waynebot/internal/db"
)

// CursorStore manages per-persona per-channel message cursors via the actor_cursors table.
type CursorStore struct {
	DB *db.DB
}

// NewCursorStore creates a CursorStore.
func NewCursorStore(d *db.DB) *CursorStore {
	return &CursorStore{DB: d}
}

// Get returns the last seen message ID for a persona in a channel. Returns 0 if no cursor exists.
func (cs *CursorStore) Get(personaID, channelID int64) (int64, error) {
	var msgID int64
	err := cs.DB.SQL.QueryRow(
		"SELECT last_seen_message_id FROM actor_cursors WHERE persona_id = ? AND channel_id = ?",
		personaID, channelID,
	).Scan(&msgID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0, nil
		}
		return 0, err
	}
	return msgID, nil
}

// Set upserts the last seen message ID for a persona in a channel.
func (cs *CursorStore) Set(personaID, channelID, messageID int64) error {
	_, err := cs.DB.WriteExec(
		`INSERT INTO actor_cursors (persona_id, channel_id, last_seen_message_id, updated_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(persona_id, channel_id) DO UPDATE SET last_seen_message_id = excluded.last_seen_message_id, updated_at = CURRENT_TIMESTAMP`,
		personaID, channelID, messageID,
	)
	return err
}
