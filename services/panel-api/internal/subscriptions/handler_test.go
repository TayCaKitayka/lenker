package subscriptions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/admins"
	"github.com/lenker/lenker/services/panel-api/internal/audit"
	"github.com/lenker/lenker/services/panel-api/internal/auth"
	"github.com/lenker/lenker/services/panel-api/internal/storage"
)

func TestCreateSubscriptionSuccess(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	repo := &fakeSubscriptionsRepository{}
	handler := NewHandler(nil, repo, testAdminOnly).WithAudit(recorder)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", strings.NewReader(`{
		"user_id": "user-1",
		"plan_id": "plan-1",
		"preferred_region": "nl"
	}`))
	response := httptest.NewRecorder()

	handler.Create(response, request.WithContext(auth.WithAdmin(request.Context(), testAdmin())))

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", response.Code, response.Body.String())
	}
	if repo.created.UserID != "user-1" || repo.created.PlanID != "plan-1" {
		t.Fatalf("expected subscription create input")
	}
	assertAudit(t, recorder.events, audit.ActionSubscriptionCreate, audit.OutcomeSuccess)
}

func TestCreateSubscriptionNotFound(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	handler := NewHandler(nil, &fakeSubscriptionsRepository{createErr: storage.ErrNotFound}, testAdminOnly).WithAudit(recorder)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", strings.NewReader(`{
		"user_id": "user-1",
		"plan_id": "missing"
	}`))
	response := httptest.NewRecorder()

	handler.Create(response, request.WithContext(auth.WithAdmin(request.Context(), testAdmin())))

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", response.Code, response.Body.String())
	}
	assertAudit(t, recorder.events, audit.ActionSubscriptionCreate, audit.OutcomeFailure)
}

func TestRenewSubscriptionSuccess(t *testing.T) {
	recorder := &fakeAuditRecorder{}
	repo := &fakeSubscriptionsRepository{}
	handler := NewHandler(nil, repo, testAdminOnly).WithAudit(recorder)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/sub-1/renew", strings.NewReader(`{"extend_days": 30}`))
	request.SetPathValue("id", "sub-1")
	response := httptest.NewRecorder()

	handler.Renew(response, request.WithContext(auth.WithAdmin(request.Context(), testAdmin())))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if repo.extendDays != 30 {
		t.Fatalf("expected extend_days to reach repository")
	}
	assertAudit(t, recorder.events, audit.ActionSubscriptionRenew, audit.OutcomeSuccess)
}

func TestRenewSubscriptionValidationError(t *testing.T) {
	handler := NewHandler(nil, &fakeSubscriptionsRepository{}, testAdminOnly)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/sub-1/renew", strings.NewReader(`{"extend_days": 0}`))
	request.SetPathValue("id", "sub-1")
	response := httptest.NewRecorder()

	handler.Renew(response, request.WithContext(auth.WithAdmin(request.Context(), testAdmin())))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
	}
}

type fakeSubscriptionsRepository struct {
	created    storage.CreateSubscriptionInput
	extendDays int
	createErr  error
}

func (r *fakeSubscriptionsRepository) List(ctx context.Context) ([]storage.Subscription, error) {
	return []storage.Subscription{}, nil
}

func (r *fakeSubscriptionsRepository) Create(ctx context.Context, input storage.CreateSubscriptionInput) (storage.Subscription, error) {
	if r.createErr != nil {
		return storage.Subscription{}, r.createErr
	}
	r.created = input
	return testSubscription("sub-1"), nil
}

func (r *fakeSubscriptionsRepository) FindByID(ctx context.Context, id string) (storage.Subscription, error) {
	return testSubscription(id), nil
}

func (r *fakeSubscriptionsRepository) Update(ctx context.Context, id string, input storage.UpdateSubscriptionInput) (storage.Subscription, error) {
	return testSubscription(id), nil
}

func (r *fakeSubscriptionsRepository) Renew(ctx context.Context, id string, extendDays int) (storage.Subscription, error) {
	r.extendDays = extendDays
	return testSubscription(id), nil
}

func testSubscription(id string) storage.Subscription {
	now := time.Now().UTC()
	return storage.Subscription{
		ID:          id,
		UserID:      "user-1",
		PlanID:      "plan-1",
		Status:      "active",
		StartsAt:    now,
		ExpiresAt:   now.Add(30 * 24 * time.Hour),
		DeviceLimit: 3,
	}
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
