package api

import (
	"net/http"

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

	WriteJSON(w, http.StatusOK, entries)
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
