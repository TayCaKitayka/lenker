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
	"github.com/lenker/lenker/services/panel-api/internal/audit"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveAdmin      = errors.New("inactive admin")
)

type Service struct {
	admins          admins.Repository
	passwords       PasswordVerifier
	audit           audit.Recorder
	sessionDuration time.Duration
}

func NewService(admins admins.Repository, passwords PasswordVerifier) *Service {
	return &Service{
		admins:          admins,
		passwords:       passwords,
		audit:           audit.NoopRecorder{},
		sessionDuration: 12 * time.Hour,
	}
}

func (s *Service) WithAudit(recorder audit.Recorder) *Service {
	if recorder != nil {
		s.audit = recorder
	}
	return s
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
		s.recordLogin(ctx, "", audit.OutcomeFailure, "invalid_credentials")
		return LoginResult{}, ErrInvalidCredentials
	}

	admin, err := s.admins.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, admins.ErrNotFound) {
			s.recordLogin(ctx, "", audit.OutcomeFailure, "invalid_credentials")
			return LoginResult{}, ErrInvalidCredentials
		}
		s.recordLogin(ctx, "", audit.OutcomeFailure, "internal_error")
		return LoginResult{}, err
	}

	if admin.Status != "active" {
		s.recordLogin(ctx, admin.ID, audit.OutcomeFailure, "inactive_admin")
		return LoginResult{}, ErrInactiveAdmin
	}

	if !s.passwords.Verify(input.Password, admin.PasswordHash) {
		s.recordLogin(ctx, admin.ID, audit.OutcomeFailure, "invalid_credentials")
		return LoginResult{}, ErrInvalidCredentials
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		s.recordLogin(ctx, admin.ID, audit.OutcomeFailure, "internal_error")
		return LoginResult{}, err
	}

	session, err := s.admins.CreateSession(ctx, admin.ID, tokenHash, time.Now().UTC().Add(s.sessionDuration))
	if err != nil {
		s.recordLogin(ctx, admin.ID, audit.OutcomeFailure, "internal_error")
		return LoginResult{}, err
	}
	session.Token = token

	s.recordLogin(ctx, admin.ID, audit.OutcomeSuccess, "")

	return LoginResult{
		Admin:   admin,
		Session: session,
	}, nil
}

func (s *Service) recordLogin(ctx context.Context, adminID string, outcome string, reason string) {
	_ = s.audit.Record(ctx, audit.Event{
		ActorType:    "admin",
		ActorID:      adminID,
		Action:       audit.ActionAdminLogin,
		ResourceType: "admin_session",
		Outcome:      outcome,
		Reason:       reason,
	})
}

func newSessionToken() (token string, tokenHash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}

	token = hex.EncodeToString(raw)
	return token, HashSessionToken(token), nil
}

func HashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
