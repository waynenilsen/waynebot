package model_test

import (
	"testing"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestCreateUser(t *testing.T) {
	d := openTestDB(t)

	u, err := model.CreateUser(d, "alice", "hash123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if u.Username != "alice" {
		t.Errorf("username = %q, want %q", u.Username, "alice")
	}
	if u.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestGetUser(t *testing.T) {
	d := openTestDB(t)

	created, _ := model.CreateUser(d, "bob", "hash")
	got, err := model.GetUser(d, created.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got.Username != "bob" {
		t.Errorf("username = %q, want %q", got.Username, "bob")
	}
}

func TestGetUserByUsername(t *testing.T) {
	d := openTestDB(t)

	model.CreateUser(d, "carol", "hash")
	got, err := model.GetUserByUsername(d, "carol")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if got.Username != "carol" {
		t.Errorf("username = %q, want %q", got.Username, "carol")
	}
}

func TestUpdateUser(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "dave", "old")
	if err := model.UpdateUser(d, u.ID, "dave2", "new"); err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	got, _ := model.GetUser(d, u.ID)
	if got.Username != "dave2" {
		t.Errorf("username = %q, want %q", got.Username, "dave2")
	}
	if got.PasswordHash != "new" {
		t.Errorf("password_hash = %q, want %q", got.PasswordHash, "new")
	}
}

func TestDeleteUser(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "eve", "hash")
	if err := model.DeleteUser(d, u.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	_, err := model.GetUser(d, u.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestListUsers(t *testing.T) {
	d := openTestDB(t)

	model.CreateUser(d, "a", "h")
	model.CreateUser(d, "b", "h")
	users, err := model.ListUsers(d)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("len = %d, want 2", len(users))
	}
}

func TestCountUsers(t *testing.T) {
	d := openTestDB(t)

	count, _ := model.CountUsers(d)
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	model.CreateUser(d, "x", "h")
	count, _ = model.CountUsers(d)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestCreateUserDuplicateUsername(t *testing.T) {
	d := openTestDB(t)

	model.CreateUser(d, "dupe", "h")
	_, err := model.CreateUser(d, "dupe", "h2")
	if err == nil {
		t.Error("expected error for duplicate username")
	}
}
