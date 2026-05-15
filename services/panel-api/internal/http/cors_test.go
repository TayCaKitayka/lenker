package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithCORSAllowsLocalPanelWebOrigin(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("Origin", localPanelWebOrigin)

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != localPanelWebOrigin {
		t.Fatalf("expected allowed origin %q, got %q", localPanelWebOrigin, got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type" {
		t.Fatalf("expected auth/content-type headers, got %q", got)
	}
}

func TestWithCORSRejectsUnknownOrigin(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("Origin", "http://localhost:3000")

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected empty allow origin, got %q", got)
	}
}

func TestWithCORSHandlesPreflight(t *testing.T) {
	handler := withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("preflight request should not call next handler")
	}))

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/admin/login", nil)
	request.Header.Set("Origin", localPanelWebOrigin)

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != localPanelWebOrigin {
		t.Fatalf("expected allowed origin %q, got %q", localPanelWebOrigin, got)
	}
}
