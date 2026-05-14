package plans

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

func TestCreatePlanSuccess(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	repo := &fakePlansRepository{}
	handler := NewHandler(nil, repo, testAdminOnly).WithAudit(recorder)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/plans", strings.NewReader(`{
		"name": "Monthly",
		"duration_days": 30,
		"device_limit": 3
	}`))
	response := httptest.NewRecorder()

	handler.Create(response, request.WithContext(auth.WithAdmin(request.Context(), testAdmin())))

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", response.Code, response.Body.String())
	}
	if repo.created.Name != "Monthly" {
		t.Fatalf("expected plan create input")
	}
	assertAudit(t, recorder.events, audit.ActionPlanCreate, audit.OutcomeSuccess)
}

func TestCreatePlanValidationError(t *testing.T) {
	handler := NewHandler(nil, &fakePlansRepository{}, testAdminOnly)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/plans", strings.NewReader(`{"name": "Bad"}`))
	response := httptest.NewRecorder()

	handler.Create(response, request.WithContext(auth.WithAdmin(request.Context(), testAdmin())))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
	}
}

func TestArchivePlanSuccess(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	repo := &fakePlansRepository{}
	handler := NewHandler(nil, repo, testAdminOnly).WithAudit(recorder)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/plans/plan-1/archive", nil)
	request.SetPathValue("id", "plan-1")
	response := httptest.NewRecorder()

	handler.Archive(response, request.WithContext(auth.WithAdmin(request.Context(), testAdmin())))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	assertAudit(t, recorder.events, audit.ActionPlanArchive, audit.OutcomeSuccess)
}

type fakePlansRepository struct {
	created storage.CreatePlanInput
}

func (r *fakePlansRepository) List(ctx context.Context) ([]storage.Plan, error) {
	return []storage.Plan{}, nil
}

func (r *fakePlansRepository) Create(ctx context.Context, input storage.CreatePlanInput) (storage.Plan, error) {
	r.created = input
	return storage.Plan{ID: "plan-1", Name: input.Name, DurationDays: input.DurationDays, DeviceLimit: input.DeviceLimit, Status: "active"}, nil
}

func (r *fakePlansRepository) FindByID(ctx context.Context, id string) (storage.Plan, error) {
	return storage.Plan{ID: id, Name: "Monthly", DurationDays: 30, DeviceLimit: 3, Status: "active"}, nil
}

func (r *fakePlansRepository) Update(ctx context.Context, id string, input storage.UpdatePlanInput) (storage.Plan, error) {
	return storage.Plan{ID: id, Name: "Monthly", DurationDays: 30, DeviceLimit: 3, Status: "active"}, nil
}

func (r *fakePlansRepository) Archive(ctx context.Context, id string) (storage.Plan, error) {
	return storage.Plan{ID: id, Name: "Monthly", DurationDays: 30, DeviceLimit: 3, Status: "archived"}, nil
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
