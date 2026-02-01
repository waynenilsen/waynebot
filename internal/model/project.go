package model

import (
	"fmt"
	"os"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

type Project struct {
	ID          int64
	Name        string
	Path        string
	Description string
	CreatedAt   time.Time
}

func scanProject(s interface{ Scan(...any) error }) (Project, error) {
	var p Project
	err := s.Scan(&p.ID, &p.Name, &p.Path, &p.Description, &p.CreatedAt)
	return p, err
}

const projectCols = "id, name, path, description, created_at"

// validateProjectPath checks that path exists and is a directory.
func validateProjectPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path %q: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory", path)
	}
	return nil
}

// CreateProject inserts a new project after validating that path is an existing directory.
func CreateProject(d *db.DB, name, path, description string) (Project, error) {
	if err := validateProjectPath(path); err != nil {
		return Project{}, err
	}
	res, err := d.WriteExec(
		"INSERT INTO projects (name, path, description) VALUES (?, ?, ?)",
		name, path, description,
	)
	if err != nil {
		return Project{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Project{}, err
	}
	return scanProject(d.SQL.QueryRow(
		"SELECT "+projectCols+" FROM projects WHERE id = ?", id,
	))
}

func GetProject(d *db.DB, id int64) (Project, error) {
	return scanProject(d.SQL.QueryRow(
		"SELECT "+projectCols+" FROM projects WHERE id = ?", id,
	))
}

func ListProjects(d *db.DB) ([]Project, error) {
	rows, err := d.SQL.Query(
		"SELECT " + projectCols + " FROM projects ORDER BY name",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func UpdateProject(d *db.DB, id int64, name, path, description string) error {
	if err := validateProjectPath(path); err != nil {
		return err
	}
	_, err := d.WriteExec(
		"UPDATE projects SET name = ?, path = ?, description = ? WHERE id = ?",
		name, path, description, id,
	)
	return err
}

func DeleteProject(d *db.DB, id int64) error {
	_, err := d.WriteExec("DELETE FROM projects WHERE id = ?", id)
	return err
}
