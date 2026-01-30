package model_test

import (
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestCreateChannel(t *testing.T) {
	d := openTestDB(t)

	ch, err := model.CreateChannel(d, "general", "General discussion")
	if err != nil {
		t.Fatalf("CreateChannel: %v", err)
	}
	if ch.Name != "general" {
		t.Errorf("name = %q, want %q", ch.Name, "general")
	}
	if ch.Description != "General discussion" {
		t.Errorf("description = %q, want %q", ch.Description, "General discussion")
	}
}

func TestGetChannel(t *testing.T) {
	d := openTestDB(t)

	created, _ := model.CreateChannel(d, "random", "Random stuff")
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

	model.CreateChannel(d, "help", "Help channel")
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

	ch, _ := model.CreateChannel(d, "old", "old desc")
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

	ch, _ := model.CreateChannel(d, "doomed", "")
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

	model.CreateChannel(d, "a", "")
	model.CreateChannel(d, "b", "")
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

	model.CreateChannel(d, "unique", "")
	_, err := model.CreateChannel(d, "unique", "")
	if err == nil {
		t.Error("expected error for duplicate name")
	}
}
