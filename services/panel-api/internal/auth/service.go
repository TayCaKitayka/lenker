package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/admins"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveAdmin      = errors.New("inactive admin")
)

type Service struct {
	admins          admins.Repository
	passwords       PasswordVerifier
	sessionDuration time.Duration
}

func NewService(admins admins.Repository, passwords PasswordVerifier) *Service {
	return &Service{
		admins:          admins,
		passwords:       passwords,
		sessionDuration: 12 * time.Hour,
	}
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	Admin   admins.Admin   `json:"admin"`
	Session admins.Session `json:"session"`
}

func (s *Service) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" || input.Password == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	admin, err := s.admins.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, admins.ErrNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if admin.Status != "active" {
		return LoginResult{}, ErrInactiveAdmin
	}

	if !s.passwords.Verify(input.Password, admin.PasswordHash) {
		return LoginResult{}, ErrInvalidCredentials
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return LoginResult{}, err
	}

	session, err := s.admins.CreateSession(ctx, admin.ID, tokenHash, time.Now().UTC().Add(s.sessionDuration))
	if err != nil {
		return LoginResult{}, err
	}
	session.Token = token

	return LoginResult{
		Admin:   admin,
		Session: session,
	}, nil
}

func newSessionToken() (token string, tokenHash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}

	token = hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}
