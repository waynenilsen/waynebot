package model

import (
	"database/sql"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

type Invite struct {
	ID        int64
	Code      string
	CreatedBy int64
	UsedBy    *int64
	CreatedAt time.Time
}

func CreateInvite(d *db.DB, code string, createdBy int64) (Invite, error) {
	var inv Invite
	err := d.WriteTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(
			"INSERT INTO invites (code, created_by) VALUES (?, ?)",
			code, createdBy,
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		return tx.QueryRow(
			"SELECT id, code, created_by, used_by, created_at FROM invites WHERE id = ?", id,
		).Scan(&inv.ID, &inv.Code, &inv.CreatedBy, &inv.UsedBy, &inv.CreatedAt)
	})
	return inv, err
}

// ClaimInvite atomically claims an invite for a user.
// Returns the invite if claimed, or sql.ErrNoRows if the code doesn't exist or is already used.
func ClaimInvite(d *db.DB, code string, userID int64) (Invite, error) {
	var inv Invite
	err := d.WriteTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(
			"UPDATE invites SET used_by = ? WHERE code = ? AND used_by IS NULL",
			userID, code,
		)
		if err != nil {
			return err
		}
		rows, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return sql.ErrNoRows
		}
		return tx.QueryRow(
			"SELECT id, code, created_by, used_by, created_at FROM invites WHERE code = ?", code,
		).Scan(&inv.ID, &inv.Code, &inv.CreatedBy, &inv.UsedBy, &inv.CreatedAt)
	})
	return inv, err
}

func GetInvite(d *db.DB, code string) (Invite, error) {
	var inv Invite
	err := d.SQL.QueryRow(
		"SELECT id, code, created_by, used_by, created_at FROM invites WHERE code = ?", code,
	).Scan(&inv.ID, &inv.Code, &inv.CreatedBy, &inv.UsedBy, &inv.CreatedAt)
	return inv, err
}

func ListInvites(d *db.DB) ([]Invite, error) {
	rows, err := d.SQL.Query(
		"SELECT id, code, created_by, used_by, created_at FROM invites ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []Invite
	for rows.Next() {
		var inv Invite
		if err := rows.Scan(&inv.ID, &inv.Code, &inv.CreatedBy, &inv.UsedBy, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invites = append(invites, inv)
	}
	return invites, rows.Err()
}
