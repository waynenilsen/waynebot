package api

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

// PersonaHandler handles persona HTTP endpoints.
type PersonaHandler struct {
	DB *db.DB
}

type createPersonaRequest struct {
	Name             string   `json:"name"`
	SystemPrompt     string   `json:"system_prompt"`
	Model            string   `json:"model"`
	ToolsEnabled     []string `json:"tools_enabled"`
	Temperature      float64  `json:"temperature"`
	MaxTokens        int      `json:"max_tokens"`
	CooldownSecs     int      `json:"cooldown_secs"`
	MaxTokensPerHour int      `json:"max_tokens_per_hour"`
}

type personaJSON struct {
	ID               int64    `json:"id"`
	Name             string   `json:"name"`
	SystemPrompt     string   `json:"system_prompt"`
	Model            string   `json:"model"`
	ToolsEnabled     []string `json:"tools_enabled"`
	Temperature      float64  `json:"temperature"`
	MaxTokens        int      `json:"max_tokens"`
	CooldownSecs     int      `json:"cooldown_secs"`
	MaxTokensPerHour int      `json:"max_tokens_per_hour"`
	CreatedAt        string   `json:"created_at"`
}

func toPersonaJSON(p model.Persona) personaJSON {
	tools := p.ToolsEnabled
	if tools == nil {
		tools = []string{}
	}
	return personaJSON{
		ID:               p.ID,
		Name:             p.Name,
		SystemPrompt:     p.SystemPrompt,
		Model:            p.Model,
		ToolsEnabled:     tools,
		Temperature:      p.Temperature,
		MaxTokens:        p.MaxTokens,
		CooldownSecs:     p.CooldownSecs,
		MaxTokensPerHour: p.MaxTokensPerHour,
		CreatedAt:        p.CreatedAt.Format(time.RFC3339),
	}
}

func validatePersonaRequest(name, systemPrompt string) (string, error) {
	name = strings.TrimSpace(name)
	if len(name) < 1 || len(name) > 100 {
		return "", &validationError{"name must be 1-100 characters"}
	}
	if len(systemPrompt) < 1 || len(systemPrompt) > 50000 {
		return "", &validationError{"system_prompt must be 1-50000 characters"}
	}
	return name, nil
}

type validationError struct {
	msg string
}

func (e *validationError) Error() string { return e.msg }

// ListTemplates returns the built-in persona templates.
func (h *PersonaHandler) ListTemplates(w http.ResponseWriter, _ *http.Request) {
	WriteJSON(w, http.StatusOK, model.PersonaTemplates())
}

// ListPersonas returns all personas.
func (h *PersonaHandler) ListPersonas(w http.ResponseWriter, r *http.Request) {
	personas, err := model.ListPersonas(h.DB)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]personaJSON, len(personas))
	for i, p := range personas {
		out[i] = toPersonaJSON(p)
	}
	WriteJSON(w, http.StatusOK, out)
}

// CreatePersona creates a new persona.
func (h *PersonaHandler) CreatePersona(w http.ResponseWriter, r *http.Request) {
	var req createPersonaRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	name, err := validatePersonaRequest(req.Name, req.SystemPrompt)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.ToolsEnabled == nil {
		req.ToolsEnabled = []string{}
	}

	p, err := model.CreatePersona(h.DB, name, req.SystemPrompt, req.Model, req.ToolsEnabled, req.Temperature, req.MaxTokens, req.CooldownSecs, req.MaxTokensPerHour)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			ErrorResponse(w, http.StatusConflict, "persona name already taken")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	WriteJSON(w, http.StatusCreated, toPersonaJSON(p))
}

// UpdatePersona updates an existing persona.
func (h *PersonaHandler) UpdatePersona(w http.ResponseWriter, r *http.Request) {
	id, ok := ParseIntParam(w, r, "id")
	if !ok {
		return
	}

	// Verify persona exists.
	if _, err := model.GetPersona(h.DB, id); err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "persona not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	var req createPersonaRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	name, err := validatePersonaRequest(req.Name, req.SystemPrompt)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.ToolsEnabled == nil {
		req.ToolsEnabled = []string{}
	}

	if err := model.UpdatePersona(h.DB, id, name, req.SystemPrompt, req.Model, req.ToolsEnabled, req.Temperature, req.MaxTokens, req.CooldownSecs, req.MaxTokensPerHour); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			ErrorResponse(w, http.StatusConflict, "persona name already taken")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	p, err := model.GetPersona(h.DB, id)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	WriteJSON(w, http.StatusOK, toPersonaJSON(p))
}

// DeletePersona deletes a persona.
func (h *PersonaHandler) DeletePersona(w http.ResponseWriter, r *http.Request) {
	id, ok := ParseIntParam(w, r, "id")
	if !ok {
		return
	}

	if _, err := model.GetPersona(h.DB, id); err != nil {
		if err == sql.ErrNoRows {
			ErrorResponse(w, http.StatusNotFound, "persona not found")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := model.DeletePersona(h.DB, id); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
