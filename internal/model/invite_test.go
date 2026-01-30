package model_test

import (
	"database/sql"
	"sync"
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestCreateInvite(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "alice", "hash")
	inv, err := model.CreateInvite(d, "ABC123", u.ID)
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if inv.Code != "ABC123" {
		t.Errorf("code = %q, want %q", inv.Code, "ABC123")
	}
	if inv.UsedBy != nil {
		t.Error("expected UsedBy to be nil")
	}
}

func TestClaimInvite(t *testing.T) {
	d := openTestDB(t)

	creator, _ := model.CreateUser(d, "alice", "hash")
	claimer, _ := model.CreateUser(d, "bob", "hash")
	model.CreateInvite(d, "CLAIM1", creator.ID)

	inv, err := model.ClaimInvite(d, "CLAIM1", claimer.ID)
	if err != nil {
		t.Fatalf("ClaimInvite: %v", err)
	}
	if inv.UsedBy == nil || *inv.UsedBy != claimer.ID {
		t.Errorf("used_by = %v, want %d", inv.UsedBy, claimer.ID)
	}
}

func TestClaimInviteAlreadyUsed(t *testing.T) {
	d := openTestDB(t)

	creator, _ := model.CreateUser(d, "alice", "hash")
	u1, _ := model.CreateUser(d, "bob", "hash")
	u2, _ := model.CreateUser(d, "carol", "hash")
	model.CreateInvite(d, "ONCE", creator.ID)

	_, err := model.ClaimInvite(d, "ONCE", u1.ID)
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}

	_, err = model.ClaimInvite(d, "ONCE", u2.ID)
	if err != sql.ErrNoRows {
		t.Errorf("second claim err = %v, want sql.ErrNoRows", err)
	}
}

func TestClaimInviteBadCode(t *testing.T) {
	d := openTestDB(t)

	model.CreateUser(d, "alice", "hash")
	_, err := model.ClaimInvite(d, "NOPE", 1)
	if err != sql.ErrNoRows {
		t.Errorf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestClaimInviteConcurrent(t *testing.T) {
	d := openTestDB(t)

	creator, _ := model.CreateUser(d, "creator", "hash")
	model.CreateInvite(d, "RACE", creator.ID)

	const n = 10
	users := make([]model.User, n)
	for i := range n {
		u, _ := model.CreateUser(d, "racer"+string(rune('a'+i)), "hash")
		users[i] = u
	}

	var (
		wg     sync.WaitGroup
		wins   int64
		winsMu sync.Mutex
	)
	wg.Add(n)
	for i := range n {
		go func(uid int64) {
			defer wg.Done()
			_, err := model.ClaimInvite(d, "RACE", uid)
			if err == nil {
				winsMu.Lock()
				wins++
				winsMu.Unlock()
			}
		}(users[i].ID)
	}
	wg.Wait()

	if wins != 1 {
		t.Errorf("wins = %d, want exactly 1", wins)
	}
}

func TestGetInvite(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "alice", "hash")
	model.CreateInvite(d, "GET1", u.ID)

	inv, err := model.GetInvite(d, "GET1")
	if err != nil {
		t.Fatalf("GetInvite: %v", err)
	}
	if inv.Code != "GET1" {
		t.Errorf("code = %q, want %q", inv.Code, "GET1")
	}
}

func TestListInvites(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "alice", "hash")
	model.CreateInvite(d, "L1", u.ID)
	model.CreateInvite(d, "L2", u.ID)

	invites, err := model.ListInvites(d)
	if err != nil {
		t.Fatalf("ListInvites: %v", err)
	}
	if len(invites) != 2 {
		t.Errorf("len = %d, want 2", len(invites))
	}
}
