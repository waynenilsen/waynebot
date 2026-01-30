package api

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/waynenilsen/waynebot/internal/auth"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

const sessionDuration = 30 * 24 * time.Hour // 30 days

// AuthHandler handles auth-related HTTP endpoints.
type AuthHandler struct {
	DB *db.DB
}

type registerRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	InviteCode string `json:"invite_code"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string   `json:"token"`
	User  userJSON `json:"user"`
}

type userJSON struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
}

func toUserJSON(u model.User) userJSON {
	return userJSON{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

// Register creates a new user account.
// If there are 0 users, registration is open (bootstrap).
// Otherwise, a valid invite code is required.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Username = strings.TrimSpace(req.Username)

	if err := ValidateUsername(req.Username); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := ValidatePassword(req.Password); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if this is bootstrap (first user) or invite-only.
	count, err := model.CountUsers(h.DB)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	if count > 0 {
		// Require invite code.
		if req.InviteCode == "" {
			ErrorResponse(w, http.StatusBadRequest, "invite code required")
			return
		}
	}

	// Hash password.
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Create user.
	user, err := model.CreateUser(h.DB, req.Username, hash)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			ErrorResponse(w, http.StatusConflict, "username already taken")
			return
		}
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Claim invite code if not bootstrap.
	if count > 0 {
		if _, err := model.ClaimInvite(h.DB, req.InviteCode, user.ID); err != nil {
			// Invite invalid — roll back user creation.
			model.DeleteUser(h.DB, user.ID)
			if err == sql.ErrNoRows {
				ErrorResponse(w, http.StatusBadRequest, "invalid or already used invite code")
				return
			}
			ErrorResponse(w, http.StatusInternalServerError, "internal error")
			return
		}
	}

	// Create session.
	token, err := auth.GenerateToken()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	expires := time.Now().Add(sessionDuration)
	if _, err := model.CreateSession(h.DB, token, user.ID, expires); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	auth.SetSessionCookie(w, token, expires, auth.IsProduction(r))

	WriteJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User:  toUserJSON(user),
	})
}

// Login authenticates a user and returns a session token.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := ReadJSON(r, &req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Username = strings.TrimSpace(req.Username)

	user, err := model.GetUserByUsername(h.DB, req.Username)
	if err != nil {
		ErrorResponse(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, req.Password); err != nil {
		ErrorResponse(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.GenerateToken()
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	expires := time.Now().Add(sessionDuration)
	if _, err := model.CreateSession(h.DB, token, user.ID, expires); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	auth.SetSessionCookie(w, token, expires, auth.IsProduction(r))

	WriteJSON(w, http.StatusOK, authResponse{
		Token: token,
		User:  toUserJSON(user),
	})
}

// Logout invalidates the current session.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract token to delete the session.
	token := ""
	if hdr := r.Header.Get("Authorization"); hdr != "" {
		if t, ok := strings.CutPrefix(hdr, "Bearer "); ok {
			token = t
		}
	}
	if token == "" {
		if c, err := r.Cookie("session"); err == nil {
			token = c.Value
		}
	}

	if token != "" {
		model.DeleteSession(h.DB, token)
	}

	auth.ClearSessionCookie(w)
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Me returns the currently authenticated user.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := GetUser(r)
	if user == nil {
		ErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	WriteJSON(w, http.StatusOK, toUserJSON(*user))
}
