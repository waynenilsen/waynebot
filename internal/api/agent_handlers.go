package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/waynenilsen/waynebot/internal/agent"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// AgentHandler handles agent status and control endpoints.
type AgentHandler struct {
	DB         *db.DB
	Supervisor *agent.Supervisor
}

type agentStatusEntry struct {
	PersonaID   int64    `json:"persona_id"`
	PersonaName string   `json:"persona_name"`
	Status      string   `json:"status"`
	Channels    []string `json:"channels"`
}

type agentStatusResponse struct {
	SupervisorRunning bool               `json:"supervisor_running"`
	Agents            []agentStatusEntry `json:"agents"`
}

// Status returns the status of all persona actors.
func (h *AgentHandler) Status(w http.ResponseWriter, r *http.Request) {
	personas, err := model.ListPersonas(h.DB)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	statuses := h.Supervisor.Status.All()

	entries := make([]agentStatusEntry, 0, len(personas))
	for _, p := range personas {
		status, ok := statuses[p.ID]
		if !ok {
			status = agent.StatusStopped
		}

		channels, err := model.GetSubscribedChannels(h.DB, p.ID)
		if err != nil {
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
			return
		}
		names := make([]string, len(channels))
		for i, ch := range channels {
			names[i] = ch.Name
		}

		entries = append(entries, agentStatusEntry{
			PersonaID:   p.ID,
			PersonaName: p.Name,
			Status:      status.String(),
			Channels:    names,
		})
	}

	WriteJSON(w, http.StatusOK, agentStatusResponse{
		SupervisorRunning: h.Supervisor.Running(),
		Agents:            entries,
	})
}

// Start starts the supervisor and all persona actors.
func (h *AgentHandler) Start(w http.ResponseWriter, r *http.Request) {
	if h.Supervisor.Running() {
		ErrorResponse(w, http.StatusConflict, "agents already running")
		return
	}

	if err := h.Supervisor.StartAll(); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "failed to start agents")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// Stop stops the supervisor and all persona actors.
func (h *AgentHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if !h.Supervisor.Running() {
		ErrorResponse(w, http.StatusConflict, "agents not running")
		return
	}

	h.Supervisor.StopAll()

	WriteJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}

type llmCallJSON struct {
	ID               int64  `json:"id"`
	PersonaID        int64  `json:"persona_id"`
	ChannelID        int64  `json:"channel_id"`
	Model            string `json:"model"`
	MessagesJSON     string `json:"messages_json"`
	ResponseJSON     string `json:"response_json"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	CreatedAt        string `json:"created_at"`
}

func toLLMCallJSON(c model.LLMCall) llmCallJSON {
	return llmCallJSON{
		ID:               c.ID,
		PersonaID:        c.PersonaID,
		ChannelID:        c.ChannelID,
		Model:            c.Model,
		MessagesJSON:     c.MessagesJSON,
		ResponseJSON:     c.ResponseJSON,
		PromptTokens:     c.PromptTokens,
		CompletionTokens: c.CompletionTokens,
		CreatedAt:        c.CreatedAt.Format(time.RFC3339),
	}
}

type toolExecJSON struct {
	ID         int64  `json:"id"`
	PersonaID  int64  `json:"persona_id"`
	ToolName   string `json:"tool_name"`
	ArgsJSON   string `json:"args_json"`
	OutputText string `json:"output_text"`
	ErrorText  string `json:"error_text"`
	DurationMs int64  `json:"duration_ms"`
	CreatedAt  string `json:"created_at"`
}

func toToolExecJSON(e model.ToolExecution) toolExecJSON {
	return toolExecJSON{
		ID:         e.ID,
		PersonaID:  e.PersonaID,
		ToolName:   e.ToolName,
		ArgsJSON:   e.ArgsJSON,
		OutputText: e.OutputText,
		ErrorText:  e.ErrorText,
		DurationMs: e.DurationMs,
		CreatedAt:  e.CreatedAt.Format(time.RFC3339),
	}
}

type agentStatsJSON struct {
	TotalCallsLastHour  int64   `json:"total_calls_last_hour"`
	TotalTokensLastHour int64   `json:"total_tokens_last_hour"`
	ErrorCountLastHour  int64   `json:"error_count_last_hour"`
	AvgResponseMs       float64 `json:"avg_response_ms"`
}

// LLMCalls returns paginated LLM calls for a persona.
func (h *AgentHandler) LLMCalls(w http.ResponseWriter, r *http.Request) {
	personaID, ok := ParseIntParam(w, r, "persona_id")
	if !ok {
		return
	}

	limit, offset := parsePagination(r, 50, 200)

	calls, err := model.ListLLMCalls(h.DB, personaID, limit, offset)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]llmCallJSON, len(calls))
	for i, c := range calls {
		out[i] = toLLMCallJSON(c)
	}
	WriteJSON(w, http.StatusOK, out)
}

// ToolExecutions returns paginated tool executions for a persona.
func (h *AgentHandler) ToolExecutions(w http.ResponseWriter, r *http.Request) {
	personaID, ok := ParseIntParam(w, r, "persona_id")
	if !ok {
		return
	}

	limit, offset := parsePagination(r, 50, 200)

	execs, err := model.ListToolExecutions(h.DB, personaID, limit, offset)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	out := make([]toolExecJSON, len(execs))
	for i, e := range execs {
		out[i] = toToolExecJSON(e)
	}
	WriteJSON(w, http.StatusOK, out)
}

// Stats returns summary statistics for a persona.
func (h *AgentHandler) Stats(w http.ResponseWriter, r *http.Request) {
	personaID, ok := ParseIntParam(w, r, "persona_id")
	if !ok {
		return
	}

	stats, err := model.GetAgentStats(h.DB, personaID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	WriteJSON(w, http.StatusOK, agentStatsJSON{
		TotalCallsLastHour:  stats.TotalCallsLastHour,
		TotalTokensLastHour: stats.TotalTokensLastHour,
		ErrorCountLastHour:  stats.ErrorCountLastHour,
		AvgResponseMs:       stats.AvgResponseMs,
	})
}

// parsePagination extracts limit and offset from query params with defaults and max bounds.
func parsePagination(r *http.Request, defaultLimit, maxLimit int) (int, int) {
	limit := defaultLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= maxLimit {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
