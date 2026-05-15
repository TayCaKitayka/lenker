package nodes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/storage"
)

func TestRegisterSuccess(t *testing.T) {
	repo := &fakeNodesRepository{}
	handler := NewHandler(nil, repo)

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

func TestRegisterValidation(t *testing.T) {
	handler := NewHandler(nil, &fakeNodesRepository{})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/register", strings.NewReader(`{"agent_version":"0.1.0-dev"}`))
	response := httptest.NewRecorder()

	handler.Register(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", response.Code, response.Body.String())
	}
}

func TestHeartbeatSuccess(t *testing.T) {
	repo := &fakeNodesRepository{}
	handler := NewHandler(nil, repo)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/node-1/heartbeat", strings.NewReader(`{
		"node_id": "node-1",
		"agent_version": "0.1.0-dev",
		"status": "healthy",
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
	handler := NewHandler(nil, &fakeNodesRepository{})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/node-1/heartbeat", strings.NewReader(`{}`))
	request.SetPathValue("id", "node-1")
	response := httptest.NewRecorder()

	handler.Heartbeat(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", response.Code, response.Body.String())
	}
}

func TestHeartbeatNotFound(t *testing.T) {
	handler := NewHandler(nil, &fakeNodesRepository{heartbeatErr: storage.ErrNotFound})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/nodes/node-1/heartbeat", strings.NewReader(`{
		"agent_version": "0.1.0-dev",
		"status": "healthy"
	}`))
	request.SetPathValue("id", "node-1")
	request.Header.Set("Authorization", "Bearer node-token")
	response := httptest.NewRecorder()

	handler.Heartbeat(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", response.Code, response.Body.String())
	}
}

type fakeNodesRepository struct {
	registered   storage.RegisterNodeInput
	heartbeat    storage.HeartbeatInput
	heartbeatErr error
}

func (r *fakeNodesRepository) Register(ctx context.Context, input storage.RegisterNodeInput) (storage.RegisterNodeResult, error) {
	r.registered = input
	return storage.RegisterNodeResult{
		Node:      storage.Node{ID: "node-1", Status: "registered", DrainState: "active"},
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
