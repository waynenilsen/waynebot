package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/waynenilsen/waynebot/internal/api"
	"github.com/waynenilsen/waynebot/internal/auth"
	"github.com/waynenilsen/waynebot/internal/db"
	"github.com/waynenilsen/waynebot/internal/model"
	"github.com/waynenilsen/waynebot/internal/ws"
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

func newTestRouter(t *testing.T, d *db.DB) http.Handler {
	t.Helper()
	hub := ws.NewHub()
	go hub.Run()
	t.Cleanup(func() { hub.Stop() })
	return api.NewRouter(d, []string{"*"}, hub)
}

func doJSON(t *testing.T, router http.Handler, method, path, body string, headers ...string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func registerUser(t *testing.T, router http.Handler, username, password, inviteCode string) (token string) {
	t.Helper()
	body := `{"username":"` + username + `","password":"` + password + `"`
	if inviteCode != "" {
		body += `,"invite_code":"` + inviteCode + `"`
	}
	body += `}`
	rec := doJSON(t, router, "POST", "/api/auth/register", body)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register %s: status=%d body=%s", username, rec.Code, rec.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp.Token
}

// TestRegisterBootstrap verifies the first user can register without an invite code.
func TestRegisterBootstrap(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	rec := doJSON(t, router, "POST", "/api/auth/register",
		`{"username":"alice","password":"password123"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Token string `json:"token"`
		User  struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
		} `json:"user"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Username != "alice" {
		t.Errorf("username = %q, want alice", resp.User.Username)
	}

	// Session cookie should be set.
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session" {
			found = true
			if c.Value != resp.Token {
				t.Errorf("cookie value = %q, want %q", c.Value, resp.Token)
			}
		}
	}
	if !found {
		t.Error("expected session cookie")
	}
}

// TestRegisterRequiresInviteAfterBootstrap verifies invite code is required for second user.
func TestRegisterRequiresInviteAfterBootstrap(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/auth/register",
		`{"username":"bob","password":"password123"}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// TestRegisterWithInviteCode verifies a second user can register with a valid invite.
func TestRegisterWithInviteCode(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	registerUser(t, router, "alice", "password123", "")

	// Create invite as alice.
	alice, _ := model.GetUserByUsername(d, "alice")
	code, _ := auth.GenerateInviteCode()
	model.CreateInvite(d, code, alice.ID)

	rec := doJSON(t, router, "POST", "/api/auth/register",
		`{"username":"bob","password":"password123","invite_code":"`+code+`"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201, body: %s", rec.Code, rec.Body.String())
	}
}

// TestRegisterInvalidInviteCode verifies registration fails with bad invite.
func TestRegisterInvalidInviteCode(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/auth/register",
		`{"username":"bob","password":"password123","invite_code":"bogus"}`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400, body: %s", rec.Code, rec.Body.String())
	}
}

// TestRegisterDuplicateUsername verifies duplicate username is rejected.
func TestRegisterDuplicateUsername(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	registerUser(t, router, "alice", "password123", "")

	// Try same username with invite.
	alice, _ := model.GetUserByUsername(d, "alice")
	code, _ := auth.GenerateInviteCode()
	model.CreateInvite(d, code, alice.ID)

	rec := doJSON(t, router, "POST", "/api/auth/register",
		`{"username":"alice","password":"password123","invite_code":"`+code+`"}`)

	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
}

// TestRegisterValidation verifies input validation on register.
func TestRegisterValidation(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	tests := []struct {
		name string
		body string
	}{
		{"bad username", `{"username":"","password":"password123"}`},
		{"short password", `{"username":"alice","password":"short"}`},
		{"username with spaces", `{"username":"alice bob","password":"password123"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doJSON(t, router, "POST", "/api/auth/register", tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

// TestLogin verifies successful login.
func TestLogin(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/auth/login",
		`{"username":"alice","password":"password123"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Token string `json:"token"`
		User  struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Username != "alice" {
		t.Errorf("username = %q, want alice", resp.User.Username)
	}
}

// TestLoginWrongPassword verifies login fails with wrong password.
func TestLoginWrongPassword(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/auth/login",
		`{"username":"alice","password":"wrongpassword"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

// TestLoginUnknownUser verifies login fails for non-existent user.
func TestLoginUnknownUser(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	rec := doJSON(t, router, "POST", "/api/auth/login",
		`{"username":"nobody","password":"password123"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

// TestLogout verifies logout invalidates session.
func TestLogout(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "POST", "/api/auth/logout", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	// Token should no longer work for /me.
	rec = doJSON(t, router, "GET", "/api/auth/me", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("after logout: status = %d, want 401", rec.Code)
	}
}

// TestMe verifies the me endpoint returns the authenticated user.
func TestMe(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	token := registerUser(t, router, "alice", "password123", "")

	rec := doJSON(t, router, "GET", "/api/auth/me", "",
		"Authorization", "Bearer "+token)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Username string `json:"username"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp.Username != "alice" {
		t.Errorf("username = %q, want alice", resp.Username)
	}
}

// TestMeUnauthenticated verifies me fails without auth.
func TestMeUnauthenticated(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	rec := doJSON(t, router, "GET", "/api/auth/me", "")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

// TestHealthz verifies the health check endpoint.
func TestHealthz(t *testing.T) {
	d := openTestDB(t)
	router := newTestRouter(t, d)

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
