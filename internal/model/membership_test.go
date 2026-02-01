package model_test

import (
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestAddChannelMember(t *testing.T) {
	d := openTestDB(t)

	ch, _ := model.CreateChannel(d, "general", "")
	alice, _ := model.CreateUser(d, "alice", "hash")

	if err := model.AddChannelMember(d, ch.ID, alice.ID, "owner"); err != nil {
		t.Fatalf("AddChannelMember: %v", err)
	}

	ok, err := model.IsChannelMember(d, ch.ID, alice.ID)
	if err != nil {
		t.Fatalf("IsChannelMember: %v", err)
	}
	if !ok {
		t.Error("expected alice to be a member")
	}
}

func TestAddChannelMemberIdempotent(t *testing.T) {
	d := openTestDB(t)

	ch, _ := model.CreateChannel(d, "general", "")
	alice, _ := model.CreateUser(d, "alice", "hash")

	model.AddChannelMember(d, ch.ID, alice.ID, "owner")
	if err := model.AddChannelMember(d, ch.ID, alice.ID, "member"); err != nil {
		t.Fatalf("duplicate AddChannelMember should not error: %v", err)
	}

	members, _ := model.GetChannelMembers(d, ch.ID)
	if len(members) != 1 {
		t.Errorf("len = %d, want 1", len(members))
	}
	if members[0].Role != "owner" {
		t.Errorf("role = %q, want %q (original should be kept)", members[0].Role, "owner")
	}
}

func TestRemoveChannelMember(t *testing.T) {
	d := openTestDB(t)

	ch, _ := model.CreateChannel(d, "general", "")
	alice, _ := model.CreateUser(d, "alice", "hash")

	model.AddChannelMember(d, ch.ID, alice.ID, "member")
	if err := model.RemoveChannelMember(d, ch.ID, alice.ID); err != nil {
		t.Fatalf("RemoveChannelMember: %v", err)
	}

	ok, _ := model.IsChannelMember(d, ch.ID, alice.ID)
	if ok {
		t.Error("expected alice to not be a member after removal")
	}
}

func TestGetChannelMembers(t *testing.T) {
	d := openTestDB(t)

	ch, _ := model.CreateChannel(d, "general", "")
	alice, _ := model.CreateUser(d, "alice", "hash")
	bob, _ := model.CreateUser(d, "bob", "hash")

	model.AddChannelMember(d, ch.ID, alice.ID, "owner")
	model.AddChannelMember(d, ch.ID, bob.ID, "member")

	members, err := model.GetChannelMembers(d, ch.ID)
	if err != nil {
		t.Fatalf("GetChannelMembers: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("len = %d, want 2", len(members))
	}

	if members[0].Username != "alice" || members[0].Role != "owner" {
		t.Errorf("member[0] = {%s, %s}, want {alice, owner}", members[0].Username, members[0].Role)
	}
	if members[1].Username != "bob" || members[1].Role != "member" {
		t.Errorf("member[1] = {%s, %s}, want {bob, member}", members[1].Username, members[1].Role)
	}
}

func TestIsChannelMemberNotMember(t *testing.T) {
	d := openTestDB(t)

	ch, _ := model.CreateChannel(d, "general", "")
	alice, _ := model.CreateUser(d, "alice", "hash")

	ok, err := model.IsChannelMember(d, ch.ID, alice.ID)
	if err != nil {
		t.Fatalf("IsChannelMember: %v", err)
	}
	if ok {
		t.Error("expected false for non-member")
	}
}

func TestListChannelsForUser(t *testing.T) {
	d := openTestDB(t)

	alice, _ := model.CreateUser(d, "alice", "hash")
	bob, _ := model.CreateUser(d, "bob", "hash")

	ch1, _ := model.CreateChannel(d, "general", "")
	ch2, _ := model.CreateChannel(d, "random", "")
	model.CreateChannel(d, "secret", "")

	model.AddChannelMember(d, ch1.ID, alice.ID, "member")
	model.AddChannelMember(d, ch2.ID, alice.ID, "member")
	model.AddChannelMember(d, ch1.ID, bob.ID, "member")

	channels, err := model.ListChannelsForUser(d, alice.ID)
	if err != nil {
		t.Fatalf("ListChannelsForUser: %v", err)
	}
	if len(channels) != 2 {
		t.Errorf("alice: len = %d, want 2", len(channels))
	}

	channels, err = model.ListChannelsForUser(d, bob.ID)
	if err != nil {
		t.Fatalf("ListChannelsForUser: %v", err)
	}
	if len(channels) != 1 {
		t.Errorf("bob: len = %d, want 1", len(channels))
	}
}

func TestListChannelsForUserExcludesDMs(t *testing.T) {
	d := openTestDB(t)

	alice, _ := model.CreateUser(d, "alice", "hash")
	bob, _ := model.CreateUser(d, "bob", "hash")

	ch, _ := model.CreateChannel(d, "general", "")
	model.AddChannelMember(d, ch.ID, alice.ID, "member")

	p1 := model.DMParticipant{UserID: &alice.ID}
	p2 := model.DMParticipant{UserID: &bob.ID}
	model.CreateDMChannel(d, "dm-alice-bob", p1, p2, alice.ID)

	channels, err := model.ListChannelsForUser(d, alice.ID)
	if err != nil {
		t.Fatalf("ListChannelsForUser: %v", err)
	}
	if len(channels) != 1 {
		t.Errorf("len = %d, want 1 (DMs should be excluded)", len(channels))
	}
}

func TestChannelMembersCascadeOnChannelDelete(t *testing.T) {
	d := openTestDB(t)

	ch, _ := model.CreateChannel(d, "doomed", "")
	alice, _ := model.CreateUser(d, "alice", "hash")
	model.AddChannelMember(d, ch.ID, alice.ID, "member")

	model.DeleteChannel(d, ch.ID)

	channels, err := model.ListChannelsForUser(d, alice.ID)
	if err != nil {
		t.Fatalf("ListChannelsForUser: %v", err)
	}
	if len(channels) != 0 {
		t.Errorf("len = %d, want 0 after channel delete", len(channels))
	}
}
