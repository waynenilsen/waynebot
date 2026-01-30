package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost    = 12
	tokenBytes    = 32 // 64 hex chars
	inviteCodeLen = 8  // 16 hex chars
)

// HashPassword hashes a plaintext password with bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
// Returns nil on success, an error otherwise.
func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateToken creates a cryptographically random session token.
func GenerateToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// GenerateInviteCode creates a cryptographically random invite code.
func GenerateInviteCode() (string, error) {
	b := make([]byte, inviteCodeLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate invite code: %w", err)
	}
	return hex.EncodeToString(b), nil
}
