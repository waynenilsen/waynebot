package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/waynenilsen/waynebot/internal/auth"
	"github.com/waynenilsen/waynebot/internal/model"
)

// WriteJSON writes v as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// ReadJSON decodes the request body into v. Returns an error message if decoding fails.
func ReadJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// ErrorResponse writes a JSON error response.
func ErrorResponse(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// GetUser returns the authenticated user from the request context, or writes a 401.
// Returns nil if no user is present (response already written).
func GetUser(r *http.Request) *model.User {
	return auth.UserFromContext(r.Context())
}

// ParseIntParam extracts a named URL parameter as int64, writing a 400 error on failure.
// Returns the parsed value and true on success, or 0 and false if the response was written.
func ParseIntParam(w http.ResponseWriter, r *http.Request, param string) (int64, bool) {
	v, err := strconv.ParseInt(chi.URLParam(r, param), 10, 64)
	if err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid "+param)
		return 0, false
	}
	return v, true
}

// Validation helpers

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// ValidateUsername checks username constraints: 1-50 chars, alphanumeric + underscore.
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if len(username) < 1 || len(username) > 50 {
		return fmt.Errorf("username must be 1-50 characters")
	}
	if !usernameRe.MatchString(username) {
		return fmt.Errorf("username must contain only letters, numbers, and underscores")
	}
	return nil
}

// ValidatePassword checks password constraints: 8-128 chars.
func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 128 {
		return fmt.Errorf("password must be 8-128 characters")
	}
	return nil
}
