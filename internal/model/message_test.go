package model_test

import (
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestCreateMessage(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "alice", "hash")
	ch, _ := model.CreateChannel(d, "general", "", 0)

	m, err := model.CreateMessage(d, ch.ID, u.ID, "human", "alice", "hello world")
	if err != nil {
		t.Fatalf("CreateMessage: %v", err)
	}
	if m.Content != "hello world" {
		t.Errorf("content = %q, want %q", m.Content, "hello world")
	}
	if m.AuthorType != "human" {
		t.Errorf("author_type = %q, want %q", m.AuthorType, "human")
	}
}

func TestGetRecentMessages(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "alice", "hash")
	ch, _ := model.CreateChannel(d, "general", "", 0)

	for i := range 5 {
		model.CreateMessage(d, ch.ID, u.ID, "human", "alice", "msg"+string(rune('0'+i)))
	}

	msgs, err := model.GetRecentMessages(d, ch.ID, 3)
	if err != nil {
		t.Fatalf("GetRecentMessages: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("len = %d, want 3", len(msgs))
	}
	// newest first
	if msgs[0].ID <= msgs[1].ID {
		t.Error("expected newest first ordering")
	}
}

func TestGetMessagesBefore(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "alice", "hash")
	ch, _ := model.CreateChannel(d, "general", "", 0)

	var ids []int64
	for i := range 5 {
		m, _ := model.CreateMessage(d, ch.ID, u.ID, "human", "alice", "msg"+string(rune('0'+i)))
		ids = append(ids, m.ID)
	}

	msgs, err := model.GetMessagesBefore(d, ch.ID, ids[4], 10)
	if err != nil {
		t.Fatalf("GetMessagesBefore: %v", err)
	}
	if len(msgs) != 4 {
		t.Fatalf("len = %d, want 4", len(msgs))
	}
	for _, m := range msgs {
		if m.ID >= ids[4] {
			t.Errorf("message id %d should be < %d", m.ID, ids[4])
		}
	}
}

func TestGetMessagesSince(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "alice", "hash")
	ch, _ := model.CreateChannel(d, "general", "", 0)

	var ids []int64
	for i := range 5 {
		m, _ := model.CreateMessage(d, ch.ID, u.ID, "human", "alice", "msg"+string(rune('0'+i)))
		ids = append(ids, m.ID)
	}

	msgs, err := model.GetMessagesSince(d, ch.ID, ids[1])
	if err != nil {
		t.Fatalf("GetMessagesSince: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("len = %d, want 3", len(msgs))
	}
	// oldest first
	if msgs[0].ID >= msgs[1].ID {
		t.Error("expected oldest first ordering")
	}
	for _, m := range msgs {
		if m.ID <= ids[1] {
			t.Errorf("message id %d should be > %d", m.ID, ids[1])
		}
	}
}

func TestMessagesIsolatedByChannel(t *testing.T) {
	d := openTestDB(t)

	u, _ := model.CreateUser(d, "alice", "hash")
	ch1, _ := model.CreateChannel(d, "ch1", "", 0)
	ch2, _ := model.CreateChannel(d, "ch2", "", 0)

	model.CreateMessage(d, ch1.ID, u.ID, "human", "alice", "in ch1")
	model.CreateMessage(d, ch2.ID, u.ID, "human", "alice", "in ch2")

	msgs, _ := model.GetRecentMessages(d, ch1.ID, 10)
	if len(msgs) != 1 {
		t.Errorf("ch1 messages = %d, want 1", len(msgs))
	}
}
