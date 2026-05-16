package agent

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBuildRegistrationPayloadRequiresBootstrapToken(t *testing.T) {
	service := NewService(Identity{NodeID: "node-1"})

	_, err := service.BuildRegistrationPayload()
	if !errors.Is(err, ErrBootstrapTokenRequired) {
		t.Fatalf("expected ErrBootstrapTokenRequired, got %v", err)
	}
}

func TestBuildRegistrationPayload(t *testing.T) {
	service := NewService(Identity{NodeID: "node-1", BootstrapToken: "bootstrap", PanelURL: "https://panel.example.com"})

	payload, err := service.BuildRegistrationPayload()
	if err != nil {
		t.Fatalf("expected payload: %v", err)
	}
	if payload.NodeID != "node-1" || payload.BootstrapToken != "bootstrap" || payload.AgentVersion == "" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestRegisteredIdentityStartsActive(t *testing.T) {
	service := NewService(Identity{NodeID: "node-1"})

	status := service.Status()
	if status.Status != StatusActive || !status.Registered {
		t.Fatalf("expected active registered status, got %#v", status)
	}
}

func TestBuildHeartbeatPayloadRequiresNodeID(t *testing.T) {
	service := NewService(Identity{})

	_, err := service.BuildHeartbeatPayload(time.Now())
	if !errors.Is(err, ErrNodeIDRequired) {
		t.Fatalf("expected ErrNodeIDRequired, got %v", err)
	}
}

func TestBuildHeartbeatPayload(t *testing.T) {
	now := time.Date(2026, 5, 15, 1, 2, 3, 0, time.UTC)
	service := NewService(Identity{NodeID: "node-1"})

	payload, err := service.BuildHeartbeatPayload(now)
	if err != nil {
		t.Fatalf("expected payload: %v", err)
	}
	if payload.NodeID != "node-1" || payload.AgentVersion == "" || payload.SentAt != now {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.Status != StatusActive {
		t.Fatalf("expected active status, got %q", payload.Status)
	}
}

func TestValidateAndStoreConfigRevision(t *testing.T) {
	service := NewService(Identity{NodeID: "node-1"})
	revision := signedTestConfigRevision(t, "node-1", 2, 1)

	if err := service.ValidateAndStoreConfigRevision(revision); err != nil {
		t.Fatalf("expected revision to be stored: %v", err)
	}
	stored, ok := service.ConfigRevision(2)
	if !ok {
		t.Fatalf("expected stored revision")
	}
	if stored.BundleHash != revision.BundleHash {
		t.Fatalf("unexpected stored revision: %#v", stored)
	}
	status := service.Status()
	if status.ActiveRevision != 2 || status.LastRollbackRevision != 1 {
		t.Fatalf("expected active and rollback metadata to be tracked, got %#v", status)
	}
}

func TestValidateAndStoreConfigRevisionAcceptsPanelAPIResponseShape(t *testing.T) {
	panelResponse := signedTestConfigRevision(t, "node-1", 2, 1)
	body, err := json.Marshal(map[string]any{
		"id":                       panelResponse.ID,
		"node_id":                  panelResponse.NodeID,
		"revision_number":          panelResponse.RevisionNumber,
		"status":                   panelResponse.Status,
		"bundle_hash":              panelResponse.BundleHash,
		"signature":                panelResponse.Signature,
		"signer":                   panelResponse.Signer,
		"rollback_target_revision": panelResponse.RollbackTargetRevision,
		"bundle":                   panelResponse.Bundle,
	})
	if err != nil {
		t.Fatalf("expected json: %v", err)
	}

	var revision ConfigRevision
	if err := json.Unmarshal(body, &revision); err != nil {
		t.Fatalf("expected panel response shape: %v", err)
	}
	if revision.Bundle["payload"] != nil || revision.Bundle["signature"] != nil {
		t.Fatalf("panel response bundle must be payload only, got %#v", revision.Bundle)
	}

	service := NewService(Identity{NodeID: "node-1"})
	if err := service.ValidateAndStoreConfigRevision(revision); err != nil {
		t.Fatalf("expected panel response shape to validate: %v", err)
	}
}

func TestValidateAndStoreConfigRevisionRejectsInvalidHash(t *testing.T) {
	service := NewService(Identity{NodeID: "node-1"})
	revision := signedTestConfigRevision(t, "node-1", 2, 1)
	revision.Bundle["protocol"] = "tampered"

	err := service.ValidateAndStoreConfigRevision(revision)
	if !errors.Is(err, ErrInvalidConfigBundleHash) {
		t.Fatalf("expected ErrInvalidConfigBundleHash, got %v", err)
	}
	if _, ok := service.ConfigRevision(2); ok {
		t.Fatalf("invalid revision must not be stored")
	}
}

func TestValidateAndStoreConfigRevisionRejectsInvalidSignature(t *testing.T) {
	service := NewService(Identity{NodeID: "node-1"})
	revision := signedTestConfigRevision(t, "node-1", 2, 1)
	revision.Signature = "bad-signature"

	err := service.ValidateAndStoreConfigRevision(revision)
	if !errors.Is(err, ErrInvalidConfigSignature) {
		t.Fatalf("expected ErrInvalidConfigSignature, got %v", err)
	}
}

func TestFetchPendingConfigRevisionBuildsBearerRequest(t *testing.T) {
	expected := signedTestConfigRevision(t, "node-1", 3, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/nodes/node-1/config-revisions/pending" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer node-token" {
			t.Fatalf("unexpected authorization header")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": expected})
	}))
	defer server.Close()

	client := PanelClient{BaseURL: server.URL, HTTPClient: server.Client()}
	revision, ok, err := client.FetchPendingConfigRevision(context.Background(), "node-1", "node-token")
	if err != nil {
		t.Fatalf("expected revision: %v", err)
	}
	if !ok || revision.RevisionNumber != expected.RevisionNumber {
		t.Fatalf("unexpected revision: ok=%v revision=%#v", ok, revision)
	}
}

func TestReportConfigRevisionBuildsBearerRequest(t *testing.T) {
	var decodedReport ConfigRevisionReport
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/nodes/node-1/config-revisions/revision-1/report" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer node-token" {
			t.Fatalf("unexpected authorization header")
		}
		if err := json.NewDecoder(r.Body).Decode(&decodedReport); err != nil {
			t.Fatalf("expected report json: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"status": "applied"}})
	}))
	defer server.Close()

	client := PanelClient{BaseURL: server.URL, HTTPClient: server.Client()}
	err := client.ReportConfigRevision(context.Background(), "node-1", "node-token", "revision-1", ConfigRevisionReport{
		Status:         "applied",
		ActiveRevision: 4,
	})
	if err != nil {
		t.Fatalf("expected report success: %v", err)
	}
	if decodedReport.Status != "applied" || decodedReport.ActiveRevision != 4 {
		t.Fatalf("unexpected report body: %#v", decodedReport)
	}
}

func TestFetchPendingConfigRevisionNoPending(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := PanelClient{BaseURL: server.URL, HTTPClient: server.Client()}
	_, ok, err := client.FetchPendingConfigRevision(context.Background(), "node-1", "node-token")
	if err != nil {
		t.Fatalf("expected no-op, got %v", err)
	}
	if ok {
		t.Fatalf("expected no pending revision")
	}
}

func TestFetchPendingConfigRevisionUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := PanelClient{BaseURL: server.URL, HTTPClient: server.Client()}
	_, _, err := client.FetchPendingConfigRevision(context.Background(), "node-1", "bad-token")
	if !errors.Is(err, ErrPendingRevisionAuth) {
		t.Fatalf("expected auth error, got %v", err)
	}
}

func TestFetchPendingConfigRevisionMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":`))
	}))
	defer server.Close()

	client := PanelClient{BaseURL: server.URL, HTTPClient: server.Client()}
	_, _, err := client.FetchPendingConfigRevision(context.Background(), "node-1", "node-token")
	if !errors.Is(err, ErrUnexpectedPanelResponse) {
		t.Fatalf("expected malformed response error, got %v", err)
	}
}

func TestFetchAndApplyPendingConfigRevision(t *testing.T) {
	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	client := &fakePendingConfigRevisionClient{revision: revision, ok: true}
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token"})

	applied, err := service.FetchAndApplyPendingConfigRevision(context.Background(), client)
	if err != nil {
		t.Fatalf("expected apply: %v", err)
	}
	if !applied {
		t.Fatalf("expected pending revision to be applied")
	}
	if client.nodeID != "node-1" || client.nodeToken != "node-token" {
		t.Fatalf("unexpected fetch input: %#v", client)
	}
	status := service.Status()
	if status.ActiveRevision != 4 || status.LastAppliedRevision != 4 || status.LastRollbackRevision != 3 {
		t.Fatalf("expected applied revision in status: %#v", status)
	}
	payload, err := service.BuildHeartbeatPayload(time.Now())
	if err != nil {
		t.Fatalf("expected heartbeat: %v", err)
	}
	if payload.ActiveRevision != 4 {
		t.Fatalf("expected heartbeat active revision 4, got %d", payload.ActiveRevision)
	}
}

func TestFetchAndApplyPendingConfigRevisionNoPending(t *testing.T) {
	client := &fakePendingConfigRevisionClient{ok: false}
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token"})

	applied, err := service.FetchAndApplyPendingConfigRevision(context.Background(), client)
	if err != nil {
		t.Fatalf("expected no-op: %v", err)
	}
	if applied {
		t.Fatalf("expected no pending revision")
	}
}

func TestFetchAndApplyPendingConfigRevisionRejectsTamperedRevision(t *testing.T) {
	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	revision.Bundle["protocol"] = "tampered"
	client := &fakePendingConfigRevisionClient{revision: revision, ok: true}
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token"})

	applied, err := service.FetchAndApplyPendingConfigRevision(context.Background(), client)
	if !errors.Is(err, ErrInvalidConfigBundleHash) {
		t.Fatalf("expected invalid bundle hash, got %v", err)
	}
	if applied {
		t.Fatalf("tampered revision must not be applied")
	}
	if _, ok := service.ConfigRevision(4); ok {
		t.Fatalf("tampered revision must not be stored")
	}
}

func TestPollPendingConfigRevisionNoPendingDoesNotReport(t *testing.T) {
	client := &fakePendingConfigRevisionClient{ok: false}
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token"})

	applied, err := service.PollPendingConfigRevision(context.Background(), client, time.Now())
	if err != nil {
		t.Fatalf("expected no-op: %v", err)
	}
	if applied {
		t.Fatalf("expected no pending revision")
	}
	if client.reported {
		t.Fatalf("no pending revision must not report status")
	}
}

func TestPollPendingConfigRevisionReportsApplied(t *testing.T) {
	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	client := &fakePendingConfigRevisionClient{revision: revision, ok: true}
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token"})
	now := time.Date(2026, 5, 16, 1, 2, 3, 0, time.UTC)

	applied, err := service.PollPendingConfigRevision(context.Background(), client, now)
	if err != nil {
		t.Fatalf("expected applied report: %v", err)
	}
	if !applied {
		t.Fatalf("expected revision applied")
	}
	if !client.reported || client.report.Status != "applied" || client.report.ActiveRevision != 4 {
		t.Fatalf("expected applied report, got %#v", client.report)
	}
	if service.Status().ActiveRevision != 4 || service.Status().LastAppliedRevision != 4 {
		t.Fatalf("expected active revision in status: %#v", service.Status())
	}
}

func TestPollPendingConfigRevisionReportsFailedForTamperedRevision(t *testing.T) {
	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	revision.Bundle["protocol"] = "tampered"
	client := &fakePendingConfigRevisionClient{revision: revision, ok: true}
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token"})

	applied, err := service.PollPendingConfigRevision(context.Background(), client, time.Now())
	if !errors.Is(err, ErrInvalidConfigBundleHash) {
		t.Fatalf("expected invalid hash, got %v", err)
	}
	if applied {
		t.Fatalf("tampered revision must not be applied")
	}
	if !client.reported || client.report.Status != "failed" || client.report.ErrorMessage != "invalid config bundle hash" {
		t.Fatalf("expected failed report, got %#v", client.report)
	}
}

func TestPollPendingConfigRevisionReportFailureHandled(t *testing.T) {
	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	client := &fakePendingConfigRevisionClient{revision: revision, ok: true, reportErr: ErrUnexpectedPanelResponse}
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token"})

	_, err := service.PollPendingConfigRevision(context.Background(), client, time.Now())
	if !errors.Is(err, ErrUnexpectedPanelResponse) {
		t.Fatalf("expected report failure, got %v", err)
	}
}

func TestValidateAndStoreConfigRevisionRejectsInvalidRenderedPayload(t *testing.T) {
	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	delete(revision.Bundle, "config")
	resignTestConfigRevision(t, &revision)
	service := NewService(Identity{NodeID: "node-1"})

	err := service.ValidateAndStoreConfigRevision(revision)
	if !errors.Is(err, ErrInvalidConfigPayload) {
		t.Fatalf("expected invalid payload, got %v", err)
	}
}

func signedTestConfigRevision(t *testing.T, nodeID string, revisionNumber int, rollbackTarget int) ConfigRevision {
	t.Helper()
	bundle := map[string]any{
		"schema_version":  "config-bundle.v1alpha1",
		"generated_by":    "panel-api",
		"protocol":        "vless-reality-xtls-vision",
		"revision_number": revisionNumber,
		"core_type":       "xray",
		"config_kind":     "xray-config-skeleton",
		"node": map[string]any{
			"id": nodeID,
		},
		"transport": map[string]any{
			"network":  "tcp",
			"security": "reality",
			"xtls":     "vision",
		},
		"config": map[string]any{
			"inbounds":  []any{},
			"outbounds": []any{},
			"routing":   map[string]any{},
		},
		"config_text": "lenker xray vless reality skeleton",
	}
	body, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("expected bundle hash: %v", err)
	}
	sum := sha256.Sum256(body)
	hash := hex.EncodeToString(sum[:])
	revision := ConfigRevision{
		ID:                     "revision-1",
		NodeID:                 nodeID,
		RevisionNumber:         revisionNumber,
		Status:                 "pending",
		BundleHash:             hash,
		Signer:                 DevConfigBundleSigner,
		RollbackTargetRevision: rollbackTarget,
		Bundle:                 bundle,
	}
	mac := hmac.New(sha256.New, []byte(devConfigBundleKey))
	if _, err := mac.Write([]byte(configSigningPayload(revision))); err != nil {
		t.Fatalf("expected signature: %v", err)
	}
	revision.Signature = hex.EncodeToString(mac.Sum(nil))
	return revision
}

func resignTestConfigRevision(t *testing.T, revision *ConfigRevision) {
	t.Helper()
	body, err := json.Marshal(revision.Bundle)
	if err != nil {
		t.Fatalf("expected bundle hash: %v", err)
	}
	sum := sha256.Sum256(body)
	revision.BundleHash = hex.EncodeToString(sum[:])
	mac := hmac.New(sha256.New, []byte(devConfigBundleKey))
	if _, err := mac.Write([]byte(configSigningPayload(*revision))); err != nil {
		t.Fatalf("expected signature: %v", err)
	}
	revision.Signature = hex.EncodeToString(mac.Sum(nil))
}

type fakePendingConfigRevisionClient struct {
	revision  ConfigRevision
	ok        bool
	err       error
	nodeID    string
	nodeToken string
	reported  bool
	reportID  string
	report    ConfigRevisionReport
	reportErr error
}

func (c *fakePendingConfigRevisionClient) FetchPendingConfigRevision(ctx context.Context, nodeID string, nodeToken string) (ConfigRevision, bool, error) {
	c.nodeID = nodeID
	c.nodeToken = nodeToken
	return c.revision, c.ok, c.err
}

func (c *fakePendingConfigRevisionClient) ReportConfigRevision(ctx context.Context, nodeID string, nodeToken string, revisionID string, report ConfigRevisionReport) error {
	c.nodeID = nodeID
	c.nodeToken = nodeToken
	c.reported = true
	c.reportID = revisionID
	c.report = report
	return c.reportErr
}
