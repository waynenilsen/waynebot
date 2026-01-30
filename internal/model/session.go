package model

import (
	"database/sql"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

type Session struct {
	ID        int64
	Token     string
	UserID    int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

func CreateSession(d *db.DB, token string, userID int64, expiresAt time.Time) (Session, error) {
	var s Session
	err := d.WriteTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(
			"INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)",
			token, userID, expiresAt,
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		return tx.QueryRow(
			"SELECT id, token, user_id, created_at, expires_at FROM sessions WHERE id = ?", id,
		).Scan(&s.ID, &s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	})
	return s, err
}

func GetSessionByToken(d *db.DB, token string) (Session, error) {
	var s Session
	err := d.SQL.QueryRow(
		"SELECT id, token, user_id, created_at, expires_at FROM sessions WHERE token = ?", token,
	).Scan(&s.ID, &s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	return s, err
}

func DeleteSession(d *db.DB, token string) error {
	_, err := d.WriteExec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func DeleteUserSessions(d *db.DB, userID int64) error {
	_, err := d.WriteExec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}

func CleanupExpiredSessions(d *db.DB) (int64, error) {
	res, err := d.WriteExec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
