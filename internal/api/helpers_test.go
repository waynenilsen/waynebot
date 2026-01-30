package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/waynenilsen/waynebot/internal/api"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	api.WriteJSON(rec, http.StatusOK, map[string]string{"hello": "world"})

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}
	if body := rec.Body.String(); body != "{\"hello\":\"world\"}\n" {
		t.Errorf("body = %q", body)
	}
}

func TestReadJSON(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"alice"}`)
	req := httptest.NewRequest("POST", "/", body)

	var v struct{ Name string }
	if err := api.ReadJSON(req, &v); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if v.Name != "alice" {
		t.Errorf("name = %q, want alice", v.Name)
	}
}

func TestReadJSONInvalid(t *testing.T) {
	body := bytes.NewBufferString(`{bad json}`)
	req := httptest.NewRequest("POST", "/", body)

	var v struct{ Name string }
	if err := api.ReadJSON(req, &v); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestErrorResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	api.ErrorResponse(rec, http.StatusBadRequest, "oops")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"alice", true},
		{"alice_bob", true},
		{"A1_b2", true},
		{"", false},
		{"ab", true},
		{" ", false},
		{"alice bob", false},
		{"alice@bob", false},
		{string(make([]byte, 51)), false},
	}
	for _, tt := range tests {
		err := api.ValidateUsername(tt.input)
		if tt.valid && err != nil {
			t.Errorf("ValidateUsername(%q) unexpected error: %v", tt.input, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("ValidateUsername(%q) expected error", tt.input)
		}
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"12345678", true},
		{"short", false},
		{string(make([]byte, 128)), true},
		{string(make([]byte, 129)), false},
	}
	for _, tt := range tests {
		err := api.ValidatePassword(tt.input)
		if tt.valid && err != nil {
			t.Errorf("ValidatePassword(len=%d) unexpected error: %v", len(tt.input), err)
		}
		if !tt.valid && err == nil {
			t.Errorf("ValidatePassword(len=%d) expected error", len(tt.input))
		}
	}
}
