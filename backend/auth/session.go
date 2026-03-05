package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateToken creates a cryptographically random 32-byte session token
// returned as a 64-character hex string.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
