package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/waynenilsen/waynebot/internal/auth"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
)

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func createTestUserWithSession(t *testing.T, d *db.DB) (model.User, model.Session) {
	t.Helper()
	hash, _ := auth.HashPassword("password")
	u, err := model.CreateUser(d, "testuser", hash)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	tok, _ := auth.GenerateToken()
	s, err := model.CreateSession(d, tok, u.ID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	return u, s
}

func TestMiddlewareBearerToken(t *testing.T) {
	d := openTestDB(t)
	u, s := createTestUserWithSession(t, d)

	handler := auth.Middleware(d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := auth.UserFromContext(r.Context())
		if got == nil {
			t.Fatal("expected user in context")
		}
		if got.ID != u.ID {
			t.Errorf("user ID = %d, want %d", got.ID, u.ID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+s.Token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestMiddlewareCookie(t *testing.T) {
	d := openTestDB(t)
	u, s := createTestUserWithSession(t, d)

	handler := auth.Middleware(d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := auth.UserFromContext(r.Context())
		if got == nil {
			t.Fatal("expected user in context")
		}
		if got.ID != u.ID {
			t.Errorf("user ID = %d, want %d", got.ID, u.ID)
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: s.Token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestMiddlewareNoAuth(t *testing.T) {
	d := openTestDB(t)

	handler := auth.Middleware(d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.UserFromContext(r.Context()) != nil {
			t.Fatal("expected no user in context")
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestMiddlewareExpiredSession(t *testing.T) {
	d := openTestDB(t)
	hash, _ := auth.HashPassword("password")
	u, _ := model.CreateUser(d, "expired", hash)
	tok, _ := auth.GenerateToken()
	s, _ := model.CreateSession(d, tok, u.ID, time.Now().Add(-time.Hour))

	handler := auth.Middleware(d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.UserFromContext(r.Context()) != nil {
			t.Fatal("expected no user for expired session")
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+s.Token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestMiddlewareInvalidToken(t *testing.T) {
	d := openTestDB(t)

	handler := auth.Middleware(d)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.UserFromContext(r.Context()) != nil {
			t.Fatal("expected no user for invalid token")
		}
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestRequireAuth(t *testing.T) {
	handler := auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
