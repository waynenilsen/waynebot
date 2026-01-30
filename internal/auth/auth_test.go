package auth_test

import (
	"testing"

	"github.com/waynenilsen/waynebot/internal/auth"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if err := auth.CheckPassword(hash, "secret123"); err != nil {
		t.Errorf("CheckPassword with correct password: %v", err)
	}
	if err := auth.CheckPassword(hash, "wrong"); err == nil {
		t.Error("CheckPassword with wrong password should fail")
	}
}

func TestHashPasswordDifferentHashes(t *testing.T) {
	h1, _ := auth.HashPassword("same")
	h2, _ := auth.HashPassword("same")
	if h1 == h2 {
		t.Error("expected different hashes for same password (bcrypt uses random salt)")
	}
}

func TestGenerateToken(t *testing.T) {
	tok, err := auth.GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if len(tok) != 64 {
		t.Errorf("token length = %d, want 64", len(tok))
	}

	tok2, _ := auth.GenerateToken()
	if tok == tok2 {
		t.Error("two tokens should not be equal")
	}
}

func TestGenerateInviteCode(t *testing.T) {
	code, err := auth.GenerateInviteCode()
	if err != nil {
		t.Fatalf("GenerateInviteCode: %v", err)
	}
	if len(code) != 16 {
		t.Errorf("invite code length = %d, want 16", len(code))
	}

	code2, _ := auth.GenerateInviteCode()
	if code == code2 {
		t.Error("two invite codes should not be equal")
	}
}
