package api

import (
	"net/http"
	"strconv"

	"github.com/waynenilsen/waynebot/internal/agent"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/ws"
)

// ContextHandler handles context budget and reset endpoints.
type ContextHandler struct {
	DB         *db.DB
	Hub        *ws.Hub
	Supervisor *agent.Supervisor
}

type contextBudgetJSON struct {
	PersonaID       int64 `json:"persona_id"`
	ChannelID       int64 `json:"channel_id"`
	TotalTokens     int   `json:"total_tokens"`
	SystemTokens    int   `json:"system_tokens"`
	ProjectTokens   int   `json:"project_tokens"`
	HistoryTokens   int   `json:"history_tokens"`
	HistoryMessages int   `json:"history_messages"`
	Exhausted       bool  `json:"exhausted"`
}

// ContextBudget returns the current context budget estimate for a persona+channel.
func (h *ContextHandler) ContextBudget(w http.ResponseWriter, r *http.Request) {
	personaID, ok := ParseIntParam(w, r, "persona_id")
	if !ok {
		return
	}

	channelIDStr := r.URL.Query().Get("channel_id")
	if channelIDStr == "" {
		ErrorResponse(w, http.StatusBadRequest, "channel_id is required")
		return
	}
	channelID, err := strconv.ParseInt(channelIDStr, 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid channel_id")
		return
	}

	persona, err := model.GetPersona(h.DB, personaID)
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, "persona not found")
		return
	}

	history, err := model.GetRecentMessages(h.DB, channelID, 50)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Reverse to chronological order (GetRecentMessages returns newest-first).
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	projects, err := model.ListChannelProjects(h.DB, channelID)
	if err != nil {
		projects = nil
	}

	assembler := &agent.ContextAssembler{}
	_, budget := assembler.AssembleContext(agent.AssembleInput{
		Persona:   persona,
		ChannelID: channelID,
		Projects:  projects,
		History:   history,
	})

	WriteJSON(w, http.StatusOK, contextBudgetJSON{
		PersonaID:       personaID,
		ChannelID:       channelID,
		TotalTokens:     budget.TotalTokens,
		SystemTokens:    budget.SystemTokens,
		ProjectTokens:   budget.ProjectTokens,
		HistoryTokens:   budget.HistoryTokens,
		HistoryMessages: budget.HistoryMessages,
		Exhausted:       budget.Exhausted,
	})
}

// ResetContext resets an agent's context for a channel by advancing the cursor.
func (h *ContextHandler) ResetContext(w http.ResponseWriter, r *http.Request) {
	personaID, ok := ParseIntParam(w, r, "persona_id")
	if !ok {
		return
	}

	channelID, ok := ParseIntParam(w, r, "channel_id")
	if !ok {
		return
	}

	// Verify persona exists.
	persona, err := model.GetPersona(h.DB, personaID)
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, "persona not found")
		return
	}

	// Get the latest message ID in the channel to advance cursor.
	latestID, err := model.GetLatestMessageID(h.DB, channelID)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Advance cursor to latest message.
	if err := h.Supervisor.Cursors.Set(personaID, channelID, latestID); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Reset status back to idle.
	h.Supervisor.Status.Set(personaID, agent.StatusIdle)

	// Post a system message.
	ch := model.Channel{ID: channelID}
	msg, err := model.CreateMessage(h.DB, ch.ID, personaID, "agent", persona.Name, "Context has been reset. I'm ready to continue.")
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.Hub.Broadcast(ws.Event{
		Type: "new_message",
		Data: msg,
	})

	WriteJSON(w, http.StatusOK, map[string]string{"status": "reset"})
}
