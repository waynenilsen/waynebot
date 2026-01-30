package auth

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

type contextKey string

const userContextKey contextKey = "user"

// UserFromContext returns the authenticated user from the request context, or nil.
func UserFromContext(ctx context.Context) *model.User {
	u, _ := ctx.Value(userContextKey).(*model.User)
	return u
}

// Middleware returns an HTTP middleware that authenticates requests via
// Authorization: Bearer <token> header or session cookie. If a valid session
// is found, the corresponding user is placed on the request context.
// Unauthenticated requests pass through without a user (use RequireAuth to block them).
func Middleware(database *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			sess, err := model.GetSessionByToken(database, token)
			if err != nil {
				// Invalid or unknown token — continue unauthenticated.
				next.ServeHTTP(w, r)
				return
			}

			if sess.ExpiresAt.Before(time.Now()) {
				next.ServeHTTP(w, r)
				return
			}

			user, err := model.GetUser(database, sess.UserID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth returns middleware that responds 401 if no user is on the context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFromContext(r.Context()) == nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// extractToken reads the token from the Authorization header or session cookie.
func extractToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		if tok, ok := strings.CutPrefix(auth, "Bearer "); ok {
			return tok
		}
	}
	if c, err := r.Cookie("session"); err == nil {
		return c.Value
	}
	return ""
}

// SetSessionCookie sets a session cookie on the response.
func SetSessionCookie(w http.ResponseWriter, token string, expires time.Time, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie removes the session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// IsProduction returns true if the request appears to be over HTTPS.
func IsProduction(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

// ErrUnauthorized is returned when no valid session is found.
var ErrUnauthorized = sql.ErrNoRows
