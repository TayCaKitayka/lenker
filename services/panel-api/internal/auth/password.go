package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

type PasswordVerifier interface {
	Verify(password string, passwordHash string) bool
}

type SHA256PasswordVerifier struct {
}

func NewPasswordVerifier() PasswordVerifier {
	return SHA256PasswordVerifier{}
}

func (SHA256PasswordVerifier) Verify(password string, passwordHash string) bool {
	hash := strings.TrimSpace(passwordHash)
	hash = strings.TrimPrefix(hash, "sha256$")
	hash = strings.TrimPrefix(hash, "sha256:")

	expected, err := hex.DecodeString(hash)
	if err != nil || len(expected) != sha256.Size {
		return false
	}

	sum := sha256.Sum256([]byte(password))
	return subtle.ConstantTimeCompare(sum[:], expected) == 1
}
