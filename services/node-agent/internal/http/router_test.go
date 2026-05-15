package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lenker/lenker/services/node-agent/internal/agent"
)

func TestHealthz(t *testing.T) {
	router := NewRouter(RouterDeps{Agent: agent.NewService(agent.Identity{})})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected body: %s", response.Body.String())
	}
}

func TestStatus(t *testing.T) {
	router := NewRouter(RouterDeps{Agent: agent.NewService(agent.Identity{
		NodeID:   "node-1",
		PanelURL: "https://panel.example.com",
	})})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/status", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"node_id":"node-1"`) {
		t.Fatalf("unexpected body: %s", response.Body.String())
	}
}
