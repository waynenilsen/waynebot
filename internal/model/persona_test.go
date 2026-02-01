package model_test

import (
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestCreatePersona(t *testing.T) {
	d := openTestDB(t)

	tools := []string{"shell_exec", "file_read"}
	p, err := model.CreatePersona(d, "helper", "You help.", "gpt-4", tools, 0.7, 4096, 30, 100000)
	if err != nil {
		t.Fatalf("CreatePersona: %v", err)
	}
	if p.Name != "helper" {
		t.Errorf("name = %q, want %q", p.Name, "helper")
	}
	if len(p.ToolsEnabled) != 2 {
		t.Errorf("tools = %v, want 2 items", p.ToolsEnabled)
	}
	if p.Temperature != 0.7 {
		t.Errorf("temperature = %f, want 0.7", p.Temperature)
	}
}

func TestGetPersona(t *testing.T) {
	d := openTestDB(t)

	created, _ := model.CreatePersona(d, "bot", "prompt", "model", []string{"a"}, 0.5, 1024, 10, 50000)
	got, err := model.GetPersona(d, created.ID)
	if err != nil {
		t.Fatalf("GetPersona: %v", err)
	}
	if got.Name != "bot" {
		t.Errorf("name = %q, want %q", got.Name, "bot")
	}
	if got.ToolsEnabled[0] != "a" {
		t.Errorf("tools[0] = %q, want %q", got.ToolsEnabled[0], "a")
	}
}

func TestGetPersonaByName(t *testing.T) {
	d := openTestDB(t)

	model.CreatePersona(d, "finder", "p", "m", []string{}, 0.5, 1024, 10, 50000)
	got, err := model.GetPersonaByName(d, "finder")
	if err != nil {
		t.Fatalf("GetPersonaByName: %v", err)
	}
	if got.Name != "finder" {
		t.Errorf("name = %q, want %q", got.Name, "finder")
	}
}

func TestUpdatePersona(t *testing.T) {
	d := openTestDB(t)

	p, _ := model.CreatePersona(d, "old", "p", "m", []string{"x"}, 0.5, 1024, 10, 50000)
	err := model.UpdatePersona(d, p.ID, "new", "p2", "m2", []string{"y", "z"}, 0.9, 2048, 60, 200000)
	if err != nil {
		t.Fatalf("UpdatePersona: %v", err)
	}
	got, _ := model.GetPersona(d, p.ID)
	if got.Name != "new" {
		t.Errorf("name = %q, want %q", got.Name, "new")
	}
	if len(got.ToolsEnabled) != 2 {
		t.Errorf("tools = %v, want 2 items", got.ToolsEnabled)
	}
	if got.MaxTokens != 2048 {
		t.Errorf("max_tokens = %d, want 2048", got.MaxTokens)
	}
}

func TestDeletePersona(t *testing.T) {
	d := openTestDB(t)

	p, _ := model.CreatePersona(d, "doomed", "p", "m", []string{}, 0.5, 1024, 10, 50000)
	if err := model.DeletePersona(d, p.ID); err != nil {
		t.Fatalf("DeletePersona: %v", err)
	}
	_, err := model.GetPersona(d, p.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestListPersonas(t *testing.T) {
	d := openTestDB(t)

	model.CreatePersona(d, "a", "p", "m", []string{}, 0.5, 1024, 10, 50000)
	model.CreatePersona(d, "b", "p", "m", []string{}, 0.5, 1024, 10, 50000)
	personas, err := model.ListPersonas(d)
	if err != nil {
		t.Fatalf("ListPersonas: %v", err)
	}
	if len(personas) != 2 {
		t.Errorf("len = %d, want 2", len(personas))
	}
}

func TestSubscribeUnsubscribeChannel(t *testing.T) {
	d := openTestDB(t)

	p, _ := model.CreatePersona(d, "bot", "p", "m", []string{}, 0.5, 1024, 10, 50000)
	ch1, _ := model.CreateChannel(d, "general", "", 0)
	ch2, _ := model.CreateChannel(d, "random", "", 0)

	model.SubscribeChannel(d, p.ID, ch1.ID)
	model.SubscribeChannel(d, p.ID, ch2.ID)

	channels, err := model.GetSubscribedChannels(d, p.ID)
	if err != nil {
		t.Fatalf("GetSubscribedChannels: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("len = %d, want 2", len(channels))
	}

	model.UnsubscribeChannel(d, p.ID, ch1.ID)
	channels, _ = model.GetSubscribedChannels(d, p.ID)
	if len(channels) != 1 {
		t.Errorf("len after unsub = %d, want 1", len(channels))
	}
	if channels[0].Name != "random" {
		t.Errorf("remaining channel = %q, want %q", channels[0].Name, "random")
	}
}

func TestSubscribeChannelIdempotent(t *testing.T) {
	d := openTestDB(t)

	p, _ := model.CreatePersona(d, "bot", "p", "m", []string{}, 0.5, 1024, 10, 50000)
	ch, _ := model.CreateChannel(d, "general", "", 0)

	model.SubscribeChannel(d, p.ID, ch.ID)
	err := model.SubscribeChannel(d, p.ID, ch.ID)
	if err != nil {
		t.Fatalf("duplicate subscribe should not error: %v", err)
	}

	channels, _ := model.GetSubscribedChannels(d, p.ID)
	if len(channels) != 1 {
		t.Errorf("len = %d, want 1", len(channels))
	}
}

func TestCreatePersonaDuplicateName(t *testing.T) {
	d := openTestDB(t)

	model.CreatePersona(d, "unique", "p", "m", []string{}, 0.5, 1024, 10, 50000)
	_, err := model.CreatePersona(d, "unique", "p2", "m2", []string{}, 0.5, 1024, 10, 50000)
	if err == nil {
		t.Error("expected error for duplicate name")
	}
}
