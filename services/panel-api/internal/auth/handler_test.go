package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/admins"
	"github.com/lenker/lenker/services/panel-api/internal/audit"
	"golang.org/x/crypto/bcrypt"
)

func TestAdminLoginHandlerSuccess(t *testing.T) {
	passwordHash := mustBcryptHash(t, "secret")
	recorder := &fakeAuditRecorder{}
	handler := NewHandler(nil, NewService(&fakeAdminsRepository{
		admin: admins.Admin{
			ID:           "admin-1",
			Email:        "owner@example.com",
			PasswordHash: passwordHash,
			Status:       "active",
		},
	}, NewPasswordVerifier()).WithAudit(recorder))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/admin/login", strings.NewReader(`{
		"email": "owner@example.com",
		"password": "secret"
	}`))
	response := httptest.NewRecorder()

	handler.AdminLogin(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"token"`) {
		t.Fatalf("expected session token in response: %s", response.Body.String())
	}
	assertAudit(t, recorder.events, audit.ActionAdminLogin, audit.OutcomeSuccess, "")
}

func TestAdminLoginHandlerInvalidCredentials(t *testing.T) {
	passwordHash := mustBcryptHash(t, "secret")
	recorder := &fakeAuditRecorder{}
	handler := NewHandler(nil, NewService(&fakeAdminsRepository{
		admin: admins.Admin{
			ID:           "admin-1",
			Email:        "owner@example.com",
			PasswordHash: passwordHash,
			Status:       "active",
		},
	}, NewPasswordVerifier()).WithAudit(recorder))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/admin/login", strings.NewReader(`{
		"email": "owner@example.com",
		"password": "wrong"
	}`))
	response := httptest.NewRecorder()

	handler.AdminLogin(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"invalid_credentials"`) {
		t.Fatalf("expected invalid_credentials response: %s", response.Body.String())
	}
	assertAudit(t, recorder.events, audit.ActionAdminLogin, audit.OutcomeFailure, "invalid_credentials")
}

func TestAdminLoginHandlerInactiveAdmin(t *testing.T) {
	passwordHash := mustBcryptHash(t, "secret")
	recorder := &fakeAuditRecorder{}
	handler := NewHandler(nil, NewService(&fakeAdminsRepository{
		admin: admins.Admin{
			ID:           "admin-1",
			Email:        "owner@example.com",
			PasswordHash: passwordHash,
			Status:       "suspended",
		},
	}, NewPasswordVerifier()).WithAudit(recorder))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/admin/login", strings.NewReader(`{
		"email": "owner@example.com",
		"password": "secret"
	}`))
	response := httptest.NewRecorder()

	handler.AdminLogin(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"inactive_admin"`) {
		t.Fatalf("expected inactive_admin response: %s", response.Body.String())
	}
	assertAudit(t, recorder.events, audit.ActionAdminLogin, audit.OutcomeFailure, "inactive_admin")
}

func TestAdminLoginHandlerMalformedJSON(t *testing.T) {
	handler := NewHandler(nil, NewService(&fakeAdminsRepository{}, NewPasswordVerifier()))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/admin/login", strings.NewReader(`{`))
	response := httptest.NewRecorder()

	handler.AdminLogin(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"bad_request"`) {
		t.Fatalf("expected bad_request response: %s", response.Body.String())
	}
}

func TestBcryptPasswordVerifier(t *testing.T) {
	passwordHash := mustBcryptHash(t, "secret")
	verifier := NewPasswordVerifier()

	if !verifier.Verify("secret", passwordHash) {
		t.Fatalf("expected bcrypt password verification to pass")
	}
	if verifier.Verify("wrong", passwordHash) {
		t.Fatalf("expected bcrypt password verification to fail")
	}
	if verifier.Verify("secret", "sha256:legacy-placeholder") {
		t.Fatalf("expected legacy sha256 placeholder hash to fail")
	}
}

func mustBcryptHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("failed to generate bcrypt hash: %v", err)
	}
	return string(hash)
}

var _ admins.Repository = (*handlerFakeAdminsRepository)(nil)

type handlerFakeAdminsRepository struct {
	admin admins.Admin
}

func (r *handlerFakeAdminsRepository) FindByEmail(ctx context.Context, email string) (admins.Admin, error) {
	return r.admin, nil
}

func (r *handlerFakeAdminsRepository) FindByActiveSessionTokenHash(ctx context.Context, tokenHash string, now time.Time) (admins.Admin, error) {
	return admins.Admin{}, admins.ErrNotFound
}

func (r *handlerFakeAdminsRepository) CreateSession(ctx context.Context, adminID string, tokenHash string, expiresAt time.Time) (admins.Session, error) {
	return admins.Session{
		ID:        "session-1",
		AdminID:   adminID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}, nil
}
