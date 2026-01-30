package model

import (
	"database/sql"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

type Channel struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}

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
			"SELECT id, name, description, created_at FROM channels WHERE id = ?", id,
		).Scan(&ch.ID, &ch.Name, &ch.Description, &ch.CreatedAt)
	})
	return ch, err
}

func GetChannel(d *db.DB, id int64) (Channel, error) {
	var ch Channel
	err := d.SQL.QueryRow(
		"SELECT id, name, description, created_at FROM channels WHERE id = ?", id,
	).Scan(&ch.ID, &ch.Name, &ch.Description, &ch.CreatedAt)
	return ch, err
}

func GetChannelByName(d *db.DB, name string) (Channel, error) {
	var ch Channel
	err := d.SQL.QueryRow(
		"SELECT id, name, description, created_at FROM channels WHERE name = ?", name,
	).Scan(&ch.ID, &ch.Name, &ch.Description, &ch.CreatedAt)
	return ch, err
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

func ListChannels(d *db.DB) ([]Channel, error) {
	rows, err := d.SQL.Query(
		"SELECT id, name, description, created_at FROM channels ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var ch Channel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Description, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}
