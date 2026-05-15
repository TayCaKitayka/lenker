package nodes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/admins"
	"github.com/lenker/lenker/services/panel-api/internal/auth"
	httpapi "github.com/lenker/lenker/services/panel-api/internal/http"
	"github.com/lenker/lenker/services/panel-api/internal/storage"
)

func TestCreateBootstrapTokenSuccess(t *testing.T) {
	repo := &fakeNodesRepository{}
	handler := NewHandler(nil, repo, testAdminOnly)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/bootstrap-token", strings.NewReader(`{
		"name": "node-a",
		"region": "eu",
		"country_code": "fi",
		"hostname": "node-a.example.com",
		"expires_in_minutes": 15
	}`))
	request.Header.Set("Authorization", "Bearer admin-token")
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", response.Code, response.Body.String())
	}
	if repo.bootstrap.Name != "node-a" || repo.bootstrap.CountryCode != "FI" {
		t.Fatalf("unexpected bootstrap input: %#v", repo.bootstrap)
	}
	if repo.bootstrap.ExpiresAt.IsZero() {
		t.Fatalf("expected expiry to be set")
	}
	if !strings.Contains(response.Body.String(), `"bootstrap_token"`) {
		t.Fatalf("expected plaintext bootstrap token in one-time response: %s", response.Body.String())
	}
}

func TestCreateBootstrapTokenRequiresAuth(t *testing.T) {
	repo := &fakeNodesRepository{}
	handler := NewHandler(nil, repo, testAdminOnly)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/bootstrap-token", strings.NewReader(`{}`))
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", response.Code, response.Body.String())
	}
	if repo.bootstrapCalled {
		t.Fatalf("bootstrap token should not be created without admin auth")
	}
}

func TestListNodesSuccess(t *testing.T) {
	repo := &fakeNodesRepository{nodes: []storage.Node{testNode("node-1")}}
	handler := NewHandler(nil, repo, testAdminOnly)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/nodes", nil)
	request.Header.Set("Authorization", "Bearer admin-token")
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"active_revision_id":3`) {
		t.Fatalf("expected node summary response: %s", response.Body.String())
	}
}

func TestListNodesRequiresAuth(t *testing.T) {
	repo := &fakeNodesRepository{}
	handler := NewHandler(nil, repo, testAdminOnly)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/nodes", nil)
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", response.Code, response.Body.String())
	}
	if repo.listCalled {
		t.Fatalf("list should not be called without admin auth")
	}
}

func TestGetNodeSuccess(t *testing.T) {
	repo := &fakeNodesRepository{node: testNode("node-1")}
	handler := NewHandler(nil, repo, testAdminOnly)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/node-1", nil)
	request.Header.Set("Authorization", "Bearer admin-token")
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if repo.foundID != "node-1" {
		t.Fatalf("unexpected find id: %q", repo.foundID)
	}
	if !strings.Contains(response.Body.String(), `"hostname":"node-1.example.com"`) {
		t.Fatalf("expected node detail response: %s", response.Body.String())
	}
}

func TestGetNodeNotFound(t *testing.T) {
	repo := &fakeNodesRepository{findErr: storage.ErrNotFound}
	handler := NewHandler(nil, repo, testAdminOnly)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/missing", nil)
	request.Header.Set("Authorization", "Bearer admin-token")
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", response.Code, response.Body.String())
	}
}

func TestNodeLifecycleActions(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		actionName string
	}{
		{name: "drain", path: "/api/v1/nodes/node-1/drain", actionName: "drain"},
		{name: "undrain", path: "/api/v1/nodes/node-1/undrain", actionName: "undrain"},
		{name: "disable", path: "/api/v1/nodes/node-1/disable", actionName: "disable"},
		{name: "enable", path: "/api/v1/nodes/node-1/enable", actionName: "enable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeNodesRepository{node: testNode("node-1")}
			handler := NewHandler(nil, repo, testAdminOnly)
			mux := http.NewServeMux()
			handler.RegisterRoutes(mux)

			request := httptest.NewRequest(http.MethodPost, tt.path, nil)
			request.Header.Set("Authorization", "Bearer admin-token")
			response := httptest.NewRecorder()

			mux.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
			}
			if repo.lifecycleAction != tt.actionName || repo.lifecycleID != "node-1" {
				t.Fatalf("unexpected lifecycle call: action=%q id=%q", repo.lifecycleAction, repo.lifecycleID)
			}
		})
	}
}

func TestNodeLifecycleInvalidTransition(t *testing.T) {
	repo := &fakeNodesRepository{lifecycleErr: storage.ErrInvalidNodeTransition}
	handler := NewHandler(nil, repo, testAdminOnly)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/node-1/enable", nil)
	request.Header.Set("Authorization", "Bearer admin-token")
	response := httptest.NewRecorder()

	mux.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "validation_error") {
		t.Fatalf("expected validation_error response: %s", response.Body.String())
	}
}

func TestRegisterSuccess(t *testing.T) {
	repo := &fakeNodesRepository{}
	handler := NewHandler(nil, repo, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/register", strings.NewReader(`{
		"bootstrap_token": "bootstrap",
		"agent_version": "0.1.0-dev",
		"hostname": "node-a"
	}`))
	response := httptest.NewRecorder()

	handler.Register(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", response.Code, response.Body.String())
	}
	if repo.registered.BootstrapToken != "bootstrap" || repo.registered.AgentVersion == "" {
		t.Fatalf("unexpected registration input: %#v", repo.registered)
	}
	if !strings.Contains(response.Body.String(), `"node_token"`) {
		t.Fatalf("expected node token response: %s", response.Body.String())
	}
}

func TestRegisterInvalidToken(t *testing.T) {
	testRegisterTokenError(t, storage.ErrInvalidBootstrapToken, http.StatusUnauthorized, "invalid_bootstrap_token")
}

func TestRegisterExpiredToken(t *testing.T) {
	testRegisterTokenError(t, storage.ErrExpiredBootstrapToken, http.StatusUnauthorized, "expired_bootstrap_token")
}

func TestRegisterReusedToken(t *testing.T) {
	testRegisterTokenError(t, storage.ErrBootstrapTokenUsed, http.StatusUnauthorized, "bootstrap_token_used")
}

func TestRegisterValidation(t *testing.T) {
	handler := NewHandler(nil, &fakeNodesRepository{}, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/register", strings.NewReader(`{"agent_version":"0.1.0-dev"}`))
	response := httptest.NewRecorder()

	handler.Register(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
	}
}

func TestHeartbeatSuccess(t *testing.T) {
	repo := &fakeNodesRepository{}
	handler := NewHandler(nil, repo, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/node-1/heartbeat", strings.NewReader(`{
		"node_id": "node-1",
		"agent_version": "0.1.0-dev",
		"status": "active",
		"active_revision": 7
	}`))
	request.SetPathValue("id", "node-1")
	request.Header.Set("Authorization", "Bearer node-token")
	response := httptest.NewRecorder()

	handler.Heartbeat(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", response.Code, response.Body.String())
	}
	if repo.heartbeat.NodeToken != "node-token" || repo.heartbeat.ActiveRevision != 7 {
		t.Fatalf("unexpected heartbeat input: %#v", repo.heartbeat)
	}
}

func TestHeartbeatUnauthorized(t *testing.T) {
	handler := NewHandler(nil, &fakeNodesRepository{}, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/node-1/heartbeat", strings.NewReader(`{}`))
	request.SetPathValue("id", "node-1")
	response := httptest.NewRecorder()

	handler.Heartbeat(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", response.Code, response.Body.String())
	}
}

func TestHeartbeatNotFound(t *testing.T) {
	handler := NewHandler(nil, &fakeNodesRepository{heartbeatErr: storage.ErrNotFound}, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/node-1/heartbeat", strings.NewReader(`{
		"agent_version": "0.1.0-dev",
		"status": "active"
	}`))
	request.SetPathValue("id", "node-1")
	request.Header.Set("Authorization", "Bearer node-token")
	response := httptest.NewRecorder()

	handler.Heartbeat(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", response.Code, response.Body.String())
	}
}

func TestHeartbeatValidation(t *testing.T) {
	handler := NewHandler(nil, &fakeNodesRepository{}, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/node-1/heartbeat", strings.NewReader(`{
		"agent_version": "0.1.0-dev",
		"status": "healthy"
	}`))
	request.SetPathValue("id", "node-1")
	request.Header.Set("Authorization", "Bearer node-token")
	response := httptest.NewRecorder()

	handler.Heartbeat(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "validation_error") {
		t.Fatalf("expected validation_error response: %s", response.Body.String())
	}
}

func testRegisterTokenError(t *testing.T, err error, expectedStatus int, expectedCode string) {
	t.Helper()

	handler := NewHandler(nil, &fakeNodesRepository{registerErr: err}, nil)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/register", strings.NewReader(`{
		"bootstrap_token": "bootstrap",
		"agent_version": "0.1.0-dev"
	}`))
	response := httptest.NewRecorder()

	handler.Register(response, request)

	if response.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d: %s", expectedStatus, response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), expectedCode) {
		t.Fatalf("expected %s response: %s", expectedCode, response.Body.String())
	}
}

type fakeNodesRepository struct {
	bootstrap       storage.CreateBootstrapTokenInput
	bootstrapCalled bool
	nodes           []storage.Node
	node            storage.Node
	listCalled      bool
	foundID         string
	findErr         error
	lifecycleAction string
	lifecycleID     string
	lifecycleErr    error
	registered      storage.RegisterNodeInput
	registerErr     error
	heartbeat       storage.HeartbeatInput
	heartbeatErr    error
}

func (r *fakeNodesRepository) List(ctx context.Context) ([]storage.Node, error) {
	r.listCalled = true
	return r.nodes, nil
}

func (r *fakeNodesRepository) FindByID(ctx context.Context, id string) (storage.Node, error) {
	r.foundID = id
	if r.findErr != nil {
		return storage.Node{}, r.findErr
	}
	return r.node, nil
}

func (r *fakeNodesRepository) CreateBootstrapToken(ctx context.Context, input storage.CreateBootstrapTokenInput) (storage.BootstrapToken, error) {
	r.bootstrapCalled = true
	r.bootstrap = input
	return storage.BootstrapToken{
		ID:        "token-1",
		NodeID:    "node-1",
		Token:     "lnkbt_plaintext",
		ExpiresAt: input.ExpiresAt,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (r *fakeNodesRepository) Register(ctx context.Context, input storage.RegisterNodeInput) (storage.RegisterNodeResult, error) {
	r.registered = input
	if r.registerErr != nil {
		return storage.RegisterNodeResult{}, r.registerErr
	}
	return storage.RegisterNodeResult{
		Node:      storage.Node{ID: "node-1", Status: "active", DrainState: "active"},
		NodeToken: "node-token",
	}, nil
}

func (r *fakeNodesRepository) RecordHeartbeat(ctx context.Context, input storage.HeartbeatInput) (storage.Node, error) {
	r.heartbeat = input
	if r.heartbeatErr != nil {
		return storage.Node{}, r.heartbeatErr
	}
	now := time.Now().UTC()
	return storage.Node{
		ID:             input.NodeID,
		Status:         input.Status,
		DrainState:     "active",
		AgentVersion:   input.AgentVersion,
		ActiveRevision: input.ActiveRevision,
		LastHealthAt:   &now,
	}, nil
}

func (r *fakeNodesRepository) Drain(ctx context.Context, id string) (storage.Node, error) {
	return r.lifecycle("drain", id, func(node storage.Node) storage.Node {
		node.DrainState = "draining"
		return node
	})
}

func (r *fakeNodesRepository) Undrain(ctx context.Context, id string) (storage.Node, error) {
	return r.lifecycle("undrain", id, func(node storage.Node) storage.Node {
		node.DrainState = "active"
		return node
	})
}

func (r *fakeNodesRepository) Disable(ctx context.Context, id string) (storage.Node, error) {
	return r.lifecycle("disable", id, func(node storage.Node) storage.Node {
		node.Status = "disabled"
		return node
	})
}

func (r *fakeNodesRepository) Enable(ctx context.Context, id string) (storage.Node, error) {
	return r.lifecycle("enable", id, func(node storage.Node) storage.Node {
		node.Status = "unhealthy"
		return node
	})
}

func (r *fakeNodesRepository) lifecycle(action string, id string, update func(storage.Node) storage.Node) (storage.Node, error) {
	r.lifecycleAction = action
	r.lifecycleID = id
	if r.lifecycleErr != nil {
		return storage.Node{}, r.lifecycleErr
	}
	node := r.node
	if node.ID == "" {
		node = testNode(id)
	}
	return update(node), nil
}

func testNode(id string) storage.Node {
	now := time.Date(2026, 5, 15, 1, 2, 3, 0, time.UTC)
	return storage.Node{
		ID:             id,
		Name:           "finland-1",
		Region:         "eu",
		CountryCode:    "FI",
		Hostname:       id + ".example.com",
		Status:         "active",
		DrainState:     "active",
		AgentVersion:   "0.1.0-dev",
		ActiveRevision: 3,
		LastSeenAt:     &now,
		RegisteredAt:   &now,
		UpdatedAt:      now,
	}
}

func testAdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer admin-token" {
			httpapi.WriteUnauthorized(w)
			return
		}
		admin := admins.Admin{ID: "admin-1", Email: "owner@example.com", Status: "active"}
		next.ServeHTTP(w, r.WithContext(auth.WithAdmin(r.Context(), admin)))
	})
}
