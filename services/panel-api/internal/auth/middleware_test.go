package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lenker/lenker/services/panel-api/internal/admins"
	"github.com/lenker/lenker/services/panel-api/internal/audit"
)

func TestSessionMiddlewareSuccess(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	repo := &fakeAdminsRepository{
		activeSessionAdmin: admins.Admin{
			ID:     "admin-1",
			Email:  "owner@example.com",
			Status: "active",
		},
	}
	middleware := NewSessionMiddleware(slog.Default(), repo).WithAudit(recorder)

	var gotAdmin admins.Admin
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		gotAdmin, ok = AdminFromContext(r.Context())
		if !ok {
			t.Fatalf("expected admin in request context")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	request.Header.Set("Authorization", "Bearer valid-token")
	response := httptest.NewRecorder()

	middleware.RequireAdmin(next).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", response.Code)
	}
	if gotAdmin.ID != "admin-1" {
		t.Fatalf("unexpected admin identity: %#v", gotAdmin)
	}
	if repo.activeSessionTokenHash == "" || repo.activeSessionTokenHash == "valid-token" {
		t.Fatalf("expected middleware to pass hashed token to repository")
	}
	assertAudit(t, recorder.events, audit.ActionAdminSessionValidation, audit.OutcomeSuccess, "")
}

func TestSessionMiddlewareMissingToken(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	middleware := NewSessionMiddleware(slog.Default(), &fakeAdminsRepository{}).WithAudit(recorder)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

	middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("next handler should not be called")
	})).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
	assertAudit(t, recorder.events, audit.ActionAdminSessionValidation, audit.OutcomeFailure, "missing_token")
}

func TestSessionMiddlewareInvalidToken(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	middleware := NewSessionMiddleware(slog.Default(), &fakeAdminsRepository{
		activeSessionErr: admins.ErrNotFound,
	}).WithAudit(recorder)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	request.Header.Set("Authorization", "Bearer invalid-token")

	middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("next handler should not be called")
	})).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
	assertAudit(t, recorder.events, audit.ActionAdminSessionValidation, audit.OutcomeFailure, "invalid_session")
}

func TestSessionMiddlewareExpiredOrInactiveSession(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	middleware := NewSessionMiddleware(slog.Default(), &fakeAdminsRepository{
		activeSessionErr: admins.ErrNotFound,
	}).WithAudit(recorder)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	request.Header.Set("Authorization", "Bearer expired-token")

	middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("next handler should not be called")
	})).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
	assertAudit(t, recorder.events, audit.ActionAdminSessionValidation, audit.OutcomeFailure, "invalid_session")
}

func TestSessionMiddlewareInternalError(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	middleware := NewSessionMiddleware(slog.Default(), &fakeAdminsRepository{
		activeSessionErr: errors.New("storage failed"),
	}).WithAudit(recorder)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	request.Header.Set("Authorization", "Bearer valid-token")

	middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("next handler should not be called")
	})).ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.Code)
	}
	assertAudit(t, recorder.events, audit.ActionAdminSessionValidation, audit.OutcomeFailure, "internal_error")
}

func TestHashSessionTokenIsStable(t *testing.T) {
	token := "session-token"
	first := HashSessionToken(token)
	second := HashSessionToken(token)
	if first == "" || first != second {
		t.Fatalf("expected stable token hash")
	}
	if first == token {
		t.Fatalf("expected hash to differ from raw token")
	}
}

var _ admins.Repository = (*fakeAdminsRepository)(nil)
var _ audit.Recorder = (*fakeAuditRecorder)(nil)
