package model

import (
	"github.com/waynenilsen/waynebot/internal/db"
)

// SetChannelProject associates a project with a channel. If the association
// already exists the call is a no-op (INSERT OR IGNORE).
func SetChannelProject(d *db.DB, channelID, projectID int64) error {
	_, err := d.WriteExec(
		"INSERT OR IGNORE INTO channel_projects (channel_id, project_id) VALUES (?, ?)",
		channelID, projectID,
	)
	return err
}

// RemoveChannelProject removes the association between a channel and a project.
func RemoveChannelProject(d *db.DB, channelID, projectID int64) error {
	_, err := d.WriteExec(
		"DELETE FROM channel_projects WHERE channel_id = ? AND project_id = ?",
		channelID, projectID,
	)
	return err
}

// ListChannelProjects returns all projects associated with a channel.
func ListChannelProjects(d *db.DB, channelID int64) ([]Project, error) {
	rows, err := d.SQL.Query(
		`SELECT p.id, p.name, p.path, p.description, p.created_at
		 FROM projects p
		 JOIN channel_projects cp ON cp.project_id = p.id
		 WHERE cp.channel_id = ?
		 ORDER BY p.name`,
		channelID,
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

// ListProjectChannels returns all channels associated with a project.
func ListProjectChannels(d *db.DB, projectID int64) ([]Channel, error) {
	rows, err := d.SQL.Query(
		`SELECT c.id, c.name, c.description, c.is_dm, c.created_by, c.created_at
		 FROM channels c
		 JOIN channel_projects cp ON cp.channel_id = c.id
		 WHERE cp.project_id = ?
		 ORDER BY c.id`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}
