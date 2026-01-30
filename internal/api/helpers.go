package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

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
