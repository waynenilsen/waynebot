package model_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/waynenilsen/waynebot/internal/model"
)

func TestCreateProject(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()

	p, err := model.CreateProject(d, "myproject", dir, "A test project")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if p.Name != "myproject" {
		t.Errorf("name = %q, want %q", p.Name, "myproject")
	}
	if p.Path != dir {
		t.Errorf("path = %q, want %q", p.Path, dir)
	}
	if p.Description != "A test project" {
		t.Errorf("description = %q, want %q", p.Description, "A test project")
	}
}

func TestCreateProjectRejectsNonExistentPath(t *testing.T) {
	d := openTestDB(t)

	_, err := model.CreateProject(d, "bad", "/no/such/path/anywhere", "")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestCreateProjectRejectsFile(t *testing.T) {
	d := openTestDB(t)
	f := filepath.Join(t.TempDir(), "afile.txt")
	os.WriteFile(f, []byte("hello"), 0644)

	_, err := model.CreateProject(d, "bad", f, "")
	if err == nil {
		t.Error("expected error for file path (not a directory)")
	}
}

func TestCreateProjectDuplicatePath(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()

	_, err := model.CreateProject(d, "first", dir, "")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	_, err = model.CreateProject(d, "second", dir, "")
	if err == nil {
		t.Error("expected error for duplicate path")
	}
}

func TestGetProject(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()

	created, _ := model.CreateProject(d, "proj", dir, "desc")
	got, err := model.GetProject(d, created.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got.Name != "proj" {
		t.Errorf("name = %q, want %q", got.Name, "proj")
	}
}

func TestGetProjectNotFound(t *testing.T) {
	d := openTestDB(t)

	_, err := model.GetProject(d, 9999)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestListProjects(t *testing.T) {
	d := openTestDB(t)
	dirB := t.TempDir()
	dirA := t.TempDir()

	model.CreateProject(d, "bravo", dirB, "")
	model.CreateProject(d, "alpha", dirA, "")

	projects, err := model.ListProjects(d)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("len = %d, want 2", len(projects))
	}
	if projects[0].Name != "alpha" {
		t.Errorf("first project = %q, want %q (ordered by name)", projects[0].Name, "alpha")
	}
}

func TestUpdateProject(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()
	dir2 := t.TempDir()

	p, _ := model.CreateProject(d, "old", dir, "old desc")
	if err := model.UpdateProject(d, p.ID, "new", dir2, "new desc"); err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	got, _ := model.GetProject(d, p.ID)
	if got.Name != "new" {
		t.Errorf("name = %q, want %q", got.Name, "new")
	}
	if got.Path != dir2 {
		t.Errorf("path = %q, want %q", got.Path, dir2)
	}
	if got.Description != "new desc" {
		t.Errorf("description = %q, want %q", got.Description, "new desc")
	}
}

func TestUpdateProjectRejectsInvalidPath(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()

	p, _ := model.CreateProject(d, "proj", dir, "")
	err := model.UpdateProject(d, p.ID, "proj", "/no/such/path", "")
	if err == nil {
		t.Error("expected error for non-existent path on update")
	}
}

func TestDeleteProject(t *testing.T) {
	d := openTestDB(t)
	dir := t.TempDir()

	p, _ := model.CreateProject(d, "doomed", dir, "")
	if err := model.DeleteProject(d, p.ID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
	_, err := model.GetProject(d, p.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}
