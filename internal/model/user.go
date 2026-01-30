package model

import (
	"database/sql"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

func CreateUser(d *db.DB, username, passwordHash string) (User, error) {
	var u User
	err := d.WriteTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(
			"INSERT INTO users (username, password_hash) VALUES (?, ?)",
			username, passwordHash,
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		return tx.QueryRow(
			"SELECT id, username, password_hash, created_at FROM users WHERE id = ?", id,
		).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	})
	return u, err
}

func GetUser(d *db.DB, id int64) (User, error) {
	var u User
	err := d.SQL.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE id = ?", id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func GetUserByUsername(d *db.DB, username string) (User, error) {
	var u User
	err := d.SQL.QueryRow(
		"SELECT id, username, password_hash, created_at FROM users WHERE username = ?", username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	return u, err
}

func UpdateUser(d *db.DB, id int64, username, passwordHash string) error {
	_, err := d.WriteExec(
		"UPDATE users SET username = ?, password_hash = ? WHERE id = ?",
		username, passwordHash, id,
	)
	return err
}

func DeleteUser(d *db.DB, id int64) error {
	_, err := d.WriteExec("DELETE FROM users WHERE id = ?", id)
	return err
}

func ListUsers(d *db.DB) ([]User, error) {
	rows, err := d.SQL.Query(
		"SELECT id, username, password_hash, created_at FROM users ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func CountUsers(d *db.DB) (int64, error) {
	var count int64
	err := d.SQL.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}
