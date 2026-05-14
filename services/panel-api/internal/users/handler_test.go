package users

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenker/lenker/services/panel-api/internal/admins"
	"github.com/lenker/lenker/services/panel-api/internal/audit"
	"github.com/lenker/lenker/services/panel-api/internal/auth"
	"github.com/lenker/lenker/services/panel-api/internal/storage"
)

func TestCreateUserSuccess(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	repo := &fakeUsersRepository{}
	handler := NewHandler(nil, repo, testAdminOnly).WithAudit(recorder)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{
		"email": "User@Example.com",
		"display_name": "User One"
	}`))
	response := httptest.NewRecorder()

	handler.Create(response, request.WithContext(auth.WithAdmin(request.Context(), testAdmin())))

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", response.Code, response.Body.String())
	}
	if repo.created.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %#v", repo.created)
	}
	assertAudit(t, recorder.events, audit.ActionUserCreate, audit.OutcomeSuccess)
}

func TestCreateUserValidationError(t *testing.T) {
	handler := NewHandler(nil, &fakeUsersRepository{}, testAdminOnly)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{"email": ""}`))
	response := httptest.NewRecorder()

	handler.Create(response, request.WithContext(auth.WithAdmin(request.Context(), testAdmin())))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
	}
}

func TestGetUserNotFound(t *testing.T) {
	handler := NewHandler(nil, &fakeUsersRepository{findErr: storage.ErrNotFound}, testAdminOnly)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/users/user-404", nil)
	request.SetPathValue("id", "user-404")
	response := httptest.NewRecorder()

	handler.Get(response, request.WithContext(auth.WithAdmin(request.Context(), testAdmin())))

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", response.Code, response.Body.String())
	}
}

func TestSuspendUserSuccess(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	repo := &fakeUsersRepository{}
	handler := NewHandler(nil, repo, testAdminOnly).WithAudit(recorder)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/users/user-1/suspend", nil)
	request.SetPathValue("id", "user-1")
	response := httptest.NewRecorder()

	handler.Suspend(response, request.WithContext(auth.WithAdmin(request.Context(), testAdmin())))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if repo.status != "suspended" {
		t.Fatalf("expected suspended status update")
	}
	assertAudit(t, recorder.events, audit.ActionUserSuspend, audit.OutcomeSuccess)
}

type fakeUsersRepository struct {
	created storage.CreateUserInput
	status  string
	findErr error
}

func (r *fakeUsersRepository) List(ctx context.Context) ([]storage.User, error) {
	return []storage.User{}, nil
}

func (r *fakeUsersRepository) Create(ctx context.Context, input storage.CreateUserInput) (storage.User, error) {
	r.created = input
	return storage.User{ID: "user-1", Email: input.Email, Status: "active", DisplayName: input.DisplayName}, nil
}

func (r *fakeUsersRepository) FindByID(ctx context.Context, id string) (storage.User, error) {
	if r.findErr != nil {
		return storage.User{}, r.findErr
	}
	return storage.User{ID: id, Email: "user@example.com", Status: "active"}, nil
}

func (r *fakeUsersRepository) Update(ctx context.Context, id string, input storage.UpdateUserInput) (storage.User, error) {
	return storage.User{ID: id, Email: "user@example.com", Status: "active"}, nil
}

func (r *fakeUsersRepository) SetStatus(ctx context.Context, id string, status string) (storage.User, error) {
	r.status = status
	return storage.User{ID: id, Email: "user@example.com", Status: status}, nil
}

type fakeAuditRecorder struct {
	events []audit.Event
}

func (r *fakeAuditRecorder) Record(ctx context.Context, event audit.Event) error {
	r.events = append(r.events, event)
	return nil
}

func testAdminOnly(next http.Handler) http.Handler {
	return next
}

func testAdmin() admins.Admin {
	return admins.Admin{ID: "admin-1", Email: "owner@example.com", Status: "active"}
}

func assertAudit(t *testing.T, events []audit.Event, action string, outcome string) {
	t.Helper()
	if len(events) == 0 {
		t.Fatalf("expected audit event")
	}
	event := events[len(events)-1]
	if event.Action != action || event.Outcome != outcome {
		t.Fatalf("unexpected audit event: %#v", event)
	}
}
