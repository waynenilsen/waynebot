package model

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

type Persona struct {
	ID               int64
	Name             string
	SystemPrompt     string
	Model            string
	ToolsEnabled     []string
	Temperature      float64
	MaxTokens        int
	CooldownSecs     int
	MaxTokensPerHour int
	CreatedAt        time.Time
}

func CreatePersona(d *db.DB, name, systemPrompt, model string, toolsEnabled []string, temperature float64, maxTokens, cooldownSecs, maxTokensPerHour int) (Persona, error) {
	toolsJSON, err := json.Marshal(toolsEnabled)
	if err != nil {
		return Persona{}, err
	}

	var p Persona
	err = d.WriteTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(
			`INSERT INTO personas (name, system_prompt, model, tools_enabled, temperature, max_tokens, cooldown_secs, max_tokens_per_hour)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			name, systemPrompt, model, string(toolsJSON), temperature, maxTokens, cooldownSecs, maxTokensPerHour,
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		return scanPersona(tx.QueryRow(
			"SELECT id, name, system_prompt, model, tools_enabled, temperature, max_tokens, cooldown_secs, max_tokens_per_hour, created_at FROM personas WHERE id = ?", id,
		), &p)
	})
	return p, err
}

func GetPersona(d *db.DB, id int64) (Persona, error) {
	var p Persona
	err := scanPersona(d.SQL.QueryRow(
		"SELECT id, name, system_prompt, model, tools_enabled, temperature, max_tokens, cooldown_secs, max_tokens_per_hour, created_at FROM personas WHERE id = ?", id,
	), &p)
	return p, err
}

func GetPersonaByName(d *db.DB, name string) (Persona, error) {
	var p Persona
	err := scanPersona(d.SQL.QueryRow(
		"SELECT id, name, system_prompt, model, tools_enabled, temperature, max_tokens, cooldown_secs, max_tokens_per_hour, created_at FROM personas WHERE name = ?", name,
	), &p)
	return p, err
}

func UpdatePersona(d *db.DB, id int64, name, systemPrompt, model string, toolsEnabled []string, temperature float64, maxTokens, cooldownSecs, maxTokensPerHour int) error {
	toolsJSON, err := json.Marshal(toolsEnabled)
	if err != nil {
		return err
	}
	_, err = d.WriteExec(
		`UPDATE personas SET name = ?, system_prompt = ?, model = ?, tools_enabled = ?, temperature = ?, max_tokens = ?, cooldown_secs = ?, max_tokens_per_hour = ? WHERE id = ?`,
		name, systemPrompt, model, string(toolsJSON), temperature, maxTokens, cooldownSecs, maxTokensPerHour, id,
	)
	return err
}

func DeletePersona(d *db.DB, id int64) error {
	_, err := d.WriteExec("DELETE FROM personas WHERE id = ?", id)
	return err
}

func ListPersonas(d *db.DB) ([]Persona, error) {
	rows, err := d.SQL.Query(
		"SELECT id, name, system_prompt, model, tools_enabled, temperature, max_tokens, cooldown_secs, max_tokens_per_hour, created_at FROM personas ORDER BY id",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var personas []Persona
	for rows.Next() {
		var p Persona
		var toolsJSON string
		if err := rows.Scan(&p.ID, &p.Name, &p.SystemPrompt, &p.Model, &toolsJSON, &p.Temperature, &p.MaxTokens, &p.CooldownSecs, &p.MaxTokensPerHour, &p.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(toolsJSON), &p.ToolsEnabled); err != nil {
			return nil, err
		}
		personas = append(personas, p)
	}
	return personas, rows.Err()
}

// SubscribeChannel subscribes a persona to a channel.
func SubscribeChannel(d *db.DB, personaID, channelID int64) error {
	_, err := d.WriteExec(
		"INSERT OR IGNORE INTO persona_channels (persona_id, channel_id) VALUES (?, ?)",
		personaID, channelID,
	)
	return err
}

// UnsubscribeChannel removes a persona's subscription to a channel.
func UnsubscribeChannel(d *db.DB, personaID, channelID int64) error {
	_, err := d.WriteExec(
		"DELETE FROM persona_channels WHERE persona_id = ? AND channel_id = ?",
		personaID, channelID,
	)
	return err
}

// GetSubscribedChannels returns all channels a persona is subscribed to.
func GetSubscribedChannels(d *db.DB, personaID int64) ([]Channel, error) {
	rows, err := d.SQL.Query(
		`SELECT c.id, c.name, c.description, c.created_at
		 FROM channels c
		 INNER JOIN persona_channels pc ON pc.channel_id = c.id
		 WHERE pc.persona_id = ?
		 ORDER BY c.id`,
		personaID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var ch Channel
		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Description, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

// scanPersona scans a single persona row, handling JSON deserialization of tools_enabled.
func scanPersona(row *sql.Row, p *Persona) error {
	var toolsJSON string
	if err := row.Scan(&p.ID, &p.Name, &p.SystemPrompt, &p.Model, &toolsJSON, &p.Temperature, &p.MaxTokens, &p.CooldownSecs, &p.MaxTokensPerHour, &p.CreatedAt); err != nil {
		return err
	}
	return json.Unmarshal([]byte(toolsJSON), &p.ToolsEnabled)
}
