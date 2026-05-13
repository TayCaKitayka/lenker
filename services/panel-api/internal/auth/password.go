package auth

import (
	"golang.org/x/crypto/bcrypt"
)

type PasswordVerifier interface {
	Verify(password string, passwordHash string) bool
}

type BcryptPasswordVerifier struct {
}

func NewPasswordVerifier() PasswordVerifier {
	return BcryptPasswordVerifier{}
}

func (BcryptPasswordVerifier) Verify(password string, passwordHash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}
