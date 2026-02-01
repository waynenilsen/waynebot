package model_test

import (
	"database/sql"
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestCreateChannel(t *testing.T) {
	d := openTestDB(t)

	ch, err := model.CreateChannel(d, "general", "General discussion", 0)
	if err != nil {
		t.Fatalf("CreateChannel: %v", err)
	}
	if ch.Name != "general" {
		t.Errorf("name = %q, want %q", ch.Name, "general")
	}
	if ch.Description != "General discussion" {
		t.Errorf("description = %q, want %q", ch.Description, "General discussion")
	}
	if ch.IsDM {
		t.Error("expected IsDM = false for regular channel")
	}
}

func TestGetChannel(t *testing.T) {
	d := openTestDB(t)

	created, _ := model.CreateChannel(d, "random", "Random stuff", 0)
	got, err := model.GetChannel(d, created.ID)
	if err != nil {
		t.Fatalf("GetChannel: %v", err)
	}
	if got.Name != "random" {
		t.Errorf("name = %q, want %q", got.Name, "random")
	}
}

func TestGetChannelByName(t *testing.T) {
	d := openTestDB(t)

	model.CreateChannel(d, "help", "Help channel", 0)
	got, err := model.GetChannelByName(d, "help")
	if err != nil {
		t.Fatalf("GetChannelByName: %v", err)
	}
	if got.Description != "Help channel" {
		t.Errorf("description = %q, want %q", got.Description, "Help channel")
	}
}

func TestUpdateChannel(t *testing.T) {
	d := openTestDB(t)

	ch, _ := model.CreateChannel(d, "old", "old desc", 0)
	if err := model.UpdateChannel(d, ch.ID, "new", "new desc"); err != nil {
		t.Fatalf("UpdateChannel: %v", err)
	}
	got, _ := model.GetChannel(d, ch.ID)
	if got.Name != "new" {
		t.Errorf("name = %q, want %q", got.Name, "new")
	}
	if got.Description != "new desc" {
		t.Errorf("description = %q, want %q", got.Description, "new desc")
	}
}

func TestDeleteChannel(t *testing.T) {
	d := openTestDB(t)

	ch, _ := model.CreateChannel(d, "doomed", "", 0)
	if err := model.DeleteChannel(d, ch.ID); err != nil {
		t.Fatalf("DeleteChannel: %v", err)
	}
	_, err := model.GetChannel(d, ch.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestListChannels(t *testing.T) {
	d := openTestDB(t)

	model.CreateChannel(d, "a", "", 0)
	model.CreateChannel(d, "b", "", 0)
	channels, err := model.ListChannels(d)
	if err != nil {
		t.Fatalf("ListChannels: %v", err)
	}
	if len(channels) != 2 {
		t.Errorf("len = %d, want 2", len(channels))
	}
}

func TestCreateChannelDuplicateName(t *testing.T) {
	d := openTestDB(t)

	model.CreateChannel(d, "unique", "", 0)
	_, err := model.CreateChannel(d, "unique", "", 0)
	if err == nil {
		t.Error("expected error for duplicate name")
	}
}

func TestListChannelsExcludesDMs(t *testing.T) {
	d := openTestDB(t)

	alice, _ := model.CreateUser(d, "alice", "hash")
	bob, _ := model.CreateUser(d, "bob", "hash")
	model.CreateChannel(d, "general", "", 0)

	p1 := model.DMParticipant{UserID: &alice.ID}
	p2 := model.DMParticipant{UserID: &bob.ID}
	_, err := model.CreateDMChannel(d, "dm-alice-bob", p1, p2, alice.ID)
	if err != nil {
		t.Fatalf("CreateDMChannel: %v", err)
	}

	channels, err := model.ListChannels(d)
	if err != nil {
		t.Fatalf("ListChannels: %v", err)
	}
	if len(channels) != 1 {
		t.Errorf("len = %d, want 1 (DM should be excluded)", len(channels))
	}
}

func TestCreateDMChannel(t *testing.T) {
	d := openTestDB(t)

	alice, _ := model.CreateUser(d, "alice", "hash")
	bob, _ := model.CreateUser(d, "bob", "hash")

	p1 := model.DMParticipant{UserID: &alice.ID}
	p2 := model.DMParticipant{UserID: &bob.ID}
	ch, err := model.CreateDMChannel(d, "dm-alice-bob", p1, p2, alice.ID)
	if err != nil {
		t.Fatalf("CreateDMChannel: %v", err)
	}
	if !ch.IsDM {
		t.Error("expected IsDM = true")
	}
	if ch.CreatedBy == nil || *ch.CreatedBy != alice.ID {
		t.Errorf("CreatedBy = %v, want %d", ch.CreatedBy, alice.ID)
	}
}

func TestCreateDMChannelWithPersonaAutoSubscribes(t *testing.T) {
	d := openTestDB(t)

	alice, _ := model.CreateUser(d, "alice", "hash")
	bot, _ := model.CreatePersona(d, "bot", "system", "model", nil, 0.7, 4096, 30, 100000)

	p1 := model.DMParticipant{UserID: &alice.ID}
	p2 := model.DMParticipant{PersonaID: &bot.ID}
	ch, err := model.CreateDMChannel(d, "dm-alice-bot", p1, p2, alice.ID)
	if err != nil {
		t.Fatalf("CreateDMChannel: %v", err)
	}

	subs, err := model.GetSubscribedChannels(d, bot.ID)
	if err != nil {
		t.Fatalf("GetSubscribedChannels: %v", err)
	}
	if len(subs) != 1 || subs[0].ID != ch.ID {
		t.Errorf("expected persona auto-subscribed to DM channel, got %v", subs)
	}
}

func TestFindDMChannel(t *testing.T) {
	d := openTestDB(t)

	alice, _ := model.CreateUser(d, "alice", "hash")
	bob, _ := model.CreateUser(d, "bob", "hash")

	p1 := model.DMParticipant{UserID: &alice.ID}
	p2 := model.DMParticipant{UserID: &bob.ID}
	created, _ := model.CreateDMChannel(d, "dm-alice-bob", p1, p2, alice.ID)

	found, err := model.FindDMChannel(d, p1, p2)
	if err != nil {
		t.Fatalf("FindDMChannel: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("found.ID = %d, want %d", found.ID, created.ID)
	}
}

func TestFindDMChannelNotFound(t *testing.T) {
	d := openTestDB(t)

	alice, _ := model.CreateUser(d, "alice", "hash")
	bob, _ := model.CreateUser(d, "bob", "hash")

	p1 := model.DMParticipant{UserID: &alice.ID}
	p2 := model.DMParticipant{UserID: &bob.ID}
	_, err := model.FindDMChannel(d, p1, p2)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestListDMsForUser(t *testing.T) {
	d := openTestDB(t)

	alice, _ := model.CreateUser(d, "alice", "hash")
	bob, _ := model.CreateUser(d, "bob", "hash")
	carol, _ := model.CreateUser(d, "carol", "hash")

	pa := model.DMParticipant{UserID: &alice.ID}
	pb := model.DMParticipant{UserID: &bob.ID}
	pc := model.DMParticipant{UserID: &carol.ID}

	model.CreateDMChannel(d, "dm-alice-bob", pa, pb, alice.ID)
	model.CreateDMChannel(d, "dm-alice-carol", pa, pc, alice.ID)
	model.CreateDMChannel(d, "dm-bob-carol", pb, pc, bob.ID)

	dms, err := model.ListDMsForUser(d, alice.ID)
	if err != nil {
		t.Fatalf("ListDMsForUser: %v", err)
	}
	if len(dms) != 2 {
		t.Errorf("len = %d, want 2", len(dms))
	}

	dms, err = model.ListDMsForUser(d, bob.ID)
	if err != nil {
		t.Fatalf("ListDMsForUser: %v", err)
	}
	if len(dms) != 2 {
		t.Errorf("len = %d, want 2", len(dms))
	}

	dms, err = model.ListDMsForUser(d, carol.ID)
	if err != nil {
		t.Fatalf("ListDMsForUser: %v", err)
	}
	if len(dms) != 2 {
		t.Errorf("len = %d, want 2", len(dms))
	}
}
