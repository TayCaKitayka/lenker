package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/admins"
	"github.com/lenker/lenker/services/panel-api/internal/audit"
)

func TestServiceLoginSuccess(t *testing.T) {
	repo := &fakeAdminsRepository{
		admin: admins.Admin{
			ID:           "admin-1",
			Email:        "owner@example.com",
			PasswordHash: "ok",
			Status:       "active",
		},
	}
	recorder := &fakeAuditRecorder{}
	service := NewService(repo, fakePasswordVerifier{valid: true}).WithAudit(recorder)

	result, err := service.Login(context.Background(), LoginInput{
		Email:    " Owner@Example.com ",
		Password: "secret",
	})
	if err != nil {
		t.Fatalf("expected login success, got error: %v", err)
	}
	if result.Admin.ID != "admin-1" {
		t.Fatalf("expected admin identity in result")
	}
	if result.Session.Token == "" {
		t.Fatalf("expected session token")
	}
	if repo.createdSessionTokenHash == "" {
		t.Fatalf("expected session token hash to be persisted")
	}
	if repo.createdSessionTokenHash == result.Session.Token {
		t.Fatalf("expected persisted token hash, not raw token")
	}
	assertAudit(t, recorder.events, audit.ActionAdminLogin, audit.OutcomeSuccess, "")
}

func TestServiceLoginInvalidCredentials(t *testing.T) {
	tests := []struct {
		name     string
		repo     *fakeAdminsRepository
		password PasswordVerifier
		input    LoginInput
	}{
		{
			name:     "empty payload",
			repo:     &fakeAdminsRepository{},
			password: fakePasswordVerifier{valid: true},
			input:    LoginInput{},
		},
		{
			name: "missing admin",
			repo: &fakeAdminsRepository{
				findByEmailErr: admins.ErrNotFound,
			},
			password: fakePasswordVerifier{valid: true},
			input:    LoginInput{Email: "owner@example.com", Password: "secret"},
		},
		{
			name: "bad password",
			repo: &fakeAdminsRepository{
				admin: admins.Admin{ID: "admin-1", Email: "owner@example.com", Status: "active"},
			},
			password: fakePasswordVerifier{valid: false},
			input:    LoginInput{Email: "owner@example.com", Password: "wrong"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := &fakeAuditRecorder{}
			service := NewService(tt.repo, tt.password).WithAudit(recorder)

			_, err := service.Login(context.Background(), tt.input)
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("expected ErrInvalidCredentials, got %v", err)
			}
			if tt.repo.createdSessionTokenHash != "" {
				t.Fatalf("did not expect session creation")
			}
			assertAudit(t, recorder.events, audit.ActionAdminLogin, audit.OutcomeFailure, "invalid_credentials")
		})
	}
}

func TestServiceLoginInactiveAdmin(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	service := NewService(&fakeAdminsRepository{
		admin: admins.Admin{
			ID:     "admin-1",
			Email:  "owner@example.com",
			Status: "suspended",
		},
	}, fakePasswordVerifier{valid: true}).WithAudit(recorder)

	_, err := service.Login(context.Background(), LoginInput{
		Email:    "owner@example.com",
		Password: "secret",
	})
	if !errors.Is(err, ErrInactiveAdmin) {
		t.Fatalf("expected ErrInactiveAdmin, got %v", err)
	}
	assertAudit(t, recorder.events, audit.ActionAdminLogin, audit.OutcomeFailure, "inactive_admin")
}

type fakePasswordVerifier struct {
	valid bool
}

func (v fakePasswordVerifier) Verify(password string, passwordHash string) bool {
	return v.valid
}

type fakeAdminsRepository struct {
	admin                   admins.Admin
	findByEmailErr          error
	activeSessionAdmin      admins.Admin
	activeSessionErr        error
	activeSessionTokenHash  string
	createdSessionTokenHash string
}

func (r *fakeAdminsRepository) FindByEmail(ctx context.Context, email string) (admins.Admin, error) {
	if r.findByEmailErr != nil {
		return admins.Admin{}, r.findByEmailErr
	}
	return r.admin, nil
}

func (r *fakeAdminsRepository) FindByActiveSessionTokenHash(ctx context.Context, tokenHash string, now time.Time) (admins.Admin, error) {
	r.activeSessionTokenHash = tokenHash
	if r.activeSessionErr != nil {
		return admins.Admin{}, r.activeSessionErr
	}
	return r.activeSessionAdmin, nil
}

func (r *fakeAdminsRepository) CreateSession(ctx context.Context, adminID string, tokenHash string, expiresAt time.Time) (admins.Session, error) {
	r.createdSessionTokenHash = tokenHash
	return admins.Session{
		ID:        "session-1",
		AdminID:   adminID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}, nil
}

type fakeAuditRecorder struct {
	events []audit.Event
}

func (r *fakeAuditRecorder) Record(ctx context.Context, event audit.Event) error {
	r.events = append(r.events, event)
	return nil
}

func assertAudit(t *testing.T, events []audit.Event, action string, outcome string, reason string) {
	t.Helper()
	if len(events) == 0 {
		t.Fatalf("expected audit event")
	}
	event := events[len(events)-1]
	if event.Action != action || event.Outcome != outcome || event.Reason != reason {
		t.Fatalf("unexpected audit event: %#v", event)
	}
}
