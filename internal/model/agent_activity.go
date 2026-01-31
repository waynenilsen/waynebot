package model

import (
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
)

// LLMCall represents a recorded LLM call from the llm_calls table.
type LLMCall struct {
	ID               int64
	PersonaID        int64
	ChannelID        int64
	Model            string
	MessagesJSON     string
	ResponseJSON     string
	PromptTokens     int
	CompletionTokens int
	CreatedAt        time.Time
}

// ToolExecution represents a recorded tool execution from the tool_executions table.
type ToolExecution struct {
	ID         int64
	PersonaID  int64
	ToolName   string
	ArgsJSON   string
	OutputText string
	ErrorText  string
	DurationMs int64
	CreatedAt  time.Time
}

// AgentStats holds summary statistics for an agent.
type AgentStats struct {
	TotalCallsLastHour  int64
	TotalTokensLastHour int64
	ErrorCountLastHour  int64
	AvgResponseMs       float64
}

// ListLLMCalls returns paginated LLM calls for a persona, newest first.
func ListLLMCalls(d *db.DB, personaID int64, limit, offset int) ([]LLMCall, error) {
	rows, err := d.SQL.Query(
		`SELECT id, persona_id, channel_id, model, messages_json, response_json, prompt_tokens, completion_tokens, created_at
		 FROM llm_calls
		 WHERE persona_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		personaID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var calls []LLMCall
	for rows.Next() {
		var c LLMCall
		if err := rows.Scan(&c.ID, &c.PersonaID, &c.ChannelID, &c.Model, &c.MessagesJSON, &c.ResponseJSON, &c.PromptTokens, &c.CompletionTokens, &c.CreatedAt); err != nil {
			return nil, err
		}
		calls = append(calls, c)
	}
	return calls, rows.Err()
}

// ListToolExecutions returns paginated tool executions for a persona, newest first.
func ListToolExecutions(d *db.DB, personaID int64, limit, offset int) ([]ToolExecution, error) {
	rows, err := d.SQL.Query(
		`SELECT id, persona_id, tool_name, args_json, COALESCE(output_text, ''), COALESCE(error_text, ''), COALESCE(duration_ms, 0), created_at
		 FROM tool_executions
		 WHERE persona_id = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		personaID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var execs []ToolExecution
	for rows.Next() {
		var e ToolExecution
		if err := rows.Scan(&e.ID, &e.PersonaID, &e.ToolName, &e.ArgsJSON, &e.OutputText, &e.ErrorText, &e.DurationMs, &e.CreatedAt); err != nil {
			return nil, err
		}
		execs = append(execs, e)
	}
	return execs, rows.Err()
}

// GetAgentStats returns summary statistics for a persona over the last hour.
func GetAgentStats(d *db.DB, personaID int64) (AgentStats, error) {
	var stats AgentStats

	err := d.SQL.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(prompt_tokens + completion_tokens), 0)
		 FROM llm_calls
		 WHERE persona_id = ? AND created_at >= datetime('now', '-1 hour')`,
		personaID,
	).Scan(&stats.TotalCallsLastHour, &stats.TotalTokensLastHour)
	if err != nil {
		return stats, err
	}

	err = d.SQL.QueryRow(
		`SELECT COUNT(*)
		 FROM tool_executions
		 WHERE persona_id = ? AND error_text != '' AND created_at >= datetime('now', '-1 hour')`,
		personaID,
	).Scan(&stats.ErrorCountLastHour)
	if err != nil {
		return stats, err
	}

	err = d.SQL.QueryRow(
		`SELECT COALESCE(AVG(duration_ms), 0)
		 FROM tool_executions
		 WHERE persona_id = ? AND created_at >= datetime('now', '-1 hour')`,
		personaID,
	).Scan(&stats.AvgResponseMs)
	if err != nil {
		return stats, err
	}

	return stats, nil
}
