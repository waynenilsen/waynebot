package model_test

import (
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestSetChannelProject(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()

	ch, _ := model.CreateChannel(d, "general", "", 0)
	p, _ := model.CreateProject(d, "proj", dir, "desc")

	if err := model.SetChannelProject(d, ch.ID, p.ID); err != nil {
		t.Fatalf("SetChannelProject: %v", err)
	}

	projects, err := model.ListChannelProjects(d, ch.ID)
	if err != nil {
		t.Fatalf("ListChannelProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("len = %d, want 1", len(projects))
	}
	if projects[0].ID != p.ID {
		t.Errorf("project id = %d, want %d", projects[0].ID, p.ID)
	}
}

func TestSetChannelProjectIdempotent(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()

	ch, _ := model.CreateChannel(d, "general", "", 0)
	p, _ := model.CreateProject(d, "proj", dir, "")

	model.SetChannelProject(d, ch.ID, p.ID)
	if err := model.SetChannelProject(d, ch.ID, p.ID); err != nil {
		t.Fatalf("duplicate SetChannelProject should not error: %v", err)
	}

	projects, _ := model.ListChannelProjects(d, ch.ID)
	if len(projects) != 1 {
		t.Errorf("len = %d, want 1", len(projects))
	}
}

func TestRemoveChannelProject(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()

	ch, _ := model.CreateChannel(d, "general", "", 0)
	p, _ := model.CreateProject(d, "proj", dir, "")

	model.SetChannelProject(d, ch.ID, p.ID)
	if err := model.RemoveChannelProject(d, ch.ID, p.ID); err != nil {
		t.Fatalf("RemoveChannelProject: %v", err)
	}

	projects, _ := model.ListChannelProjects(d, ch.ID)
	if len(projects) != 0 {
		t.Errorf("len = %d, want 0 after removal", len(projects))
	}
}

func TestListChannelProjectsEmpty(t *testing.T) {
	d := openTestDB(t)

	ch, _ := model.CreateChannel(d, "general", "", 0)

	projects, err := model.ListChannelProjects(d, ch.ID)
	if err != nil {
		t.Fatalf("ListChannelProjects: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("len = %d, want 0", len(projects))
	}
}

func TestListChannelProjectsMultiple(t *testing.T) {
	d := openTestDB(t)
	dirA := t.TempDir()
	dirB := t.TempDir()

	ch, _ := model.CreateChannel(d, "general", "", 0)
	pA, _ := model.CreateProject(d, "alpha", dirA, "")
	pB, _ := model.CreateProject(d, "bravo", dirB, "")

	model.SetChannelProject(d, ch.ID, pB.ID)
	model.SetChannelProject(d, ch.ID, pA.ID)

	projects, err := model.ListChannelProjects(d, ch.ID)
	if err != nil {
		t.Fatalf("ListChannelProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("len = %d, want 2", len(projects))
	}
	if projects[0].Name != "alpha" {
		t.Errorf("first = %q, want alpha (ordered by name)", projects[0].Name)
	}
}

func TestListProjectChannels(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()

	ch1, _ := model.CreateChannel(d, "general", "", 0)
	ch2, _ := model.CreateChannel(d, "random", "", 0)
	p, _ := model.CreateProject(d, "proj", dir, "")

	model.SetChannelProject(d, ch1.ID, p.ID)
	model.SetChannelProject(d, ch2.ID, p.ID)

	channels, err := model.ListProjectChannels(d, p.ID)
	if err != nil {
		t.Fatalf("ListProjectChannels: %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("len = %d, want 2", len(channels))
	}
}

func TestChannelProjectCascadeOnChannelDelete(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()

	ch, _ := model.CreateChannel(d, "doomed", "", 0)
	p, _ := model.CreateProject(d, "proj", dir, "")

	model.SetChannelProject(d, ch.ID, p.ID)
	model.DeleteChannel(d, ch.ID)

	channels, err := model.ListProjectChannels(d, p.ID)
	if err != nil {
		t.Fatalf("ListProjectChannels: %v", err)
	}
	if len(channels) != 0 {
		t.Errorf("len = %d, want 0 after channel delete", len(channels))
	}
}

func TestChannelProjectCascadeOnProjectDelete(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()

	ch, _ := model.CreateChannel(d, "general", "", 0)
	p, _ := model.CreateProject(d, "doomed", dir, "")

	model.SetChannelProject(d, ch.ID, p.ID)
	model.DeleteProject(d, p.ID)

	projects, err := model.ListChannelProjects(d, ch.ID)
	if err != nil {
		t.Fatalf("ListChannelProjects: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("len = %d, want 0 after project delete", len(projects))
	}
}
