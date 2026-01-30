package model_test

import (
	"testing"
	"time"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestCreateSession(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "alice", "hash")
	s, err := model.CreateSession(d, "tok123", u.ID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if s.Token != "tok123" {
		t.Errorf("token = %q, want %q", s.Token, "tok123")
	}
	if s.UserID != u.ID {
		t.Errorf("user_id = %d, want %d", s.UserID, u.ID)
	}
}

func TestGetSessionByToken(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "bob", "hash")
	model.CreateSession(d, "findme", u.ID, time.Now().Add(time.Hour))

	s, err := model.GetSessionByToken(d, "findme")
	if err != nil {
		t.Fatalf("GetSessionByToken: %v", err)
	}
	if s.UserID != u.ID {
		t.Errorf("user_id = %d, want %d", s.UserID, u.ID)
	}
}

func TestDeleteSession(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "carol", "hash")
	model.CreateSession(d, "del", u.ID, time.Now().Add(time.Hour))

	if err := model.DeleteSession(d, "del"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	_, err := model.GetSessionByToken(d, "del")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestDeleteUserSessions(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "dave", "hash")
	model.CreateSession(d, "s1", u.ID, time.Now().Add(time.Hour))
	model.CreateSession(d, "s2", u.ID, time.Now().Add(time.Hour))

	if err := model.DeleteUserSessions(d, u.ID); err != nil {
		t.Fatalf("DeleteUserSessions: %v", err)
	}
	_, err := model.GetSessionByToken(d, "s1")
	if err == nil {
		t.Error("expected s1 deleted")
	}
	_, err = model.GetSessionByToken(d, "s2")
	if err == nil {
		t.Error("expected s2 deleted")
	}
}

func TestCleanupExpiredSessions(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "eve", "hash")
	model.CreateSession(d, "expired", u.ID, time.Now().Add(-time.Hour))
	model.CreateSession(d, "valid", u.ID, time.Now().Add(time.Hour))

	n, err := model.CleanupExpiredSessions(d)
	if err != nil {
		t.Fatalf("CleanupExpiredSessions: %v", err)
	}
	if n != 1 {
		t.Errorf("deleted = %d, want 1", n)
	}

	_, err = model.GetSessionByToken(d, "valid")
	if err != nil {
		t.Errorf("valid session should still exist: %v", err)
	}
}
