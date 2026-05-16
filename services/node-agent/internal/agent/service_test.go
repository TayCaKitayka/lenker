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
	"os"
	"path/filepath"
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
	if status.ActiveRevision != 0 || status.LastRollbackRevision != 1 {
		t.Fatalf("expected validation to track rollback metadata without advancing active revision, got %#v", status)
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
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token", StateDir: t.TempDir()})
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
	if service.Status().ConfigArtifactPath == "" {
		t.Fatalf("expected config artifact path in status")
	}
	if _, err := os.Stat(service.Status().ConfigArtifactPath); err != nil {
		t.Fatalf("expected config artifact: %v", err)
	}
}

func TestPollPendingConfigRevisionReportsFailedForTamperedRevision(t *testing.T) {
	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	revision.Bundle["protocol"] = "tampered"
	client := &fakePendingConfigRevisionClient{revision: revision, ok: true}
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token", StateDir: t.TempDir()})

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

func TestPollPendingConfigRevisionReportsFailedForInvalidXrayConfig(t *testing.T) {
	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	config := revision.Bundle["config"].(map[string]any)
	inbound := config["inbounds"].([]any)[0].(map[string]any)
	delete(inbound, "streamSettings")
	resignTestConfigRevision(t, &revision)
	client := &fakePendingConfigRevisionClient{revision: revision, ok: true}
	stateDir := t.TempDir()
	activeDir := filepath.Join(stateDir, "active")
	if err := os.MkdirAll(activeDir, 0o700); err != nil {
		t.Fatalf("expected active dir: %v", err)
	}
	activeConfigPath := filepath.Join(activeDir, "config.json")
	if err := os.WriteFile(activeConfigPath, []byte(`{"revision":"old"}`+"\n"), 0o600); err != nil {
		t.Fatalf("expected old active config: %v", err)
	}
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token", StateDir: stateDir})

	applied, err := service.PollPendingConfigRevision(context.Background(), client, time.Now())
	if !errors.Is(err, ErrInvalidXrayConfig) {
		t.Fatalf("expected invalid xray config, got %v", err)
	}
	if applied {
		t.Fatalf("invalid xray config must not be applied")
	}
	if service.Status().ActiveRevision != 0 || service.Status().LastAppliedRevision != 0 {
		t.Fatalf("invalid xray config must not advance status: %#v", service.Status())
	}
	body, err := os.ReadFile(activeConfigPath)
	if err != nil {
		t.Fatalf("expected active config to remain readable: %v", err)
	}
	if string(body) != `{"revision":"old"}`+"\n" {
		t.Fatalf("validation failure changed active config: %s", string(body))
	}
	if !client.reported || client.report.Status != "failed" || client.report.ErrorMessage != "invalid_xray_config:missing_stream_settings" {
		t.Fatalf("expected failed validation report, got %#v", client.report)
	}
}

func TestPollPendingConfigRevisionReportFailureHandled(t *testing.T) {
	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	client := &fakePendingConfigRevisionClient{revision: revision, ok: true, reportErr: ErrUnexpectedPanelResponse}
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token", StateDir: t.TempDir()})

	_, err := service.PollPendingConfigRevision(context.Background(), client, time.Now())
	if !errors.Is(err, ErrUnexpectedPanelResponse) {
		t.Fatalf("expected report failure, got %v", err)
	}
}

func TestPollPendingConfigRevisionWriteFailureReportsFailed(t *testing.T) {
	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	client := &fakePendingConfigRevisionClient{revision: revision, ok: true}
	stateFile := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(stateFile, []byte("file"), 0o600); err != nil {
		t.Fatalf("expected state file fixture: %v", err)
	}
	service := NewService(Identity{NodeID: "node-1", NodeToken: "node-token", StateDir: stateFile})

	applied, err := service.PollPendingConfigRevision(context.Background(), client, time.Now())
	if !errors.Is(err, ErrConfigArtifactWrite) {
		t.Fatalf("expected artifact write error, got %v", err)
	}
	if applied {
		t.Fatalf("write failure must not apply revision")
	}
	if service.Status().ActiveRevision != 0 || service.Status().LastAppliedRevision != 0 {
		t.Fatalf("write failure must not advance status: %#v", service.Status())
	}
	if !client.reported || client.report.Status != "failed" || client.report.ErrorMessage != "config artifact write failed" {
		t.Fatalf("expected failed report for write failure, got %#v", client.report)
	}
}

func TestApplyConfigRevisionWritesLocalArtifacts(t *testing.T) {
	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	service := NewService(Identity{NodeID: "node-1", StateDir: t.TempDir()})

	if err := service.ApplyConfigRevision(revision); err != nil {
		t.Fatalf("expected local apply: %v", err)
	}
	status := service.Status()
	if status.ActiveRevision != 4 || status.LastRollbackRevision != 3 {
		t.Fatalf("expected status after local apply: %#v", status)
	}
	if status.StagedRevision != 4 || status.RollbackCandidateRevision != 3 {
		t.Fatalf("expected staged rollback metadata after local apply: %#v", status)
	}
	configBody, err := os.ReadFile(status.ConfigArtifactPath)
	if err != nil {
		t.Fatalf("expected active config artifact: %v", err)
	}
	if !json.Valid(configBody) {
		t.Fatalf("expected valid config json: %s", string(configBody))
	}
	metadataBody, err := os.ReadFile(status.MetadataArtifactPath)
	if err != nil {
		t.Fatalf("expected active metadata artifact: %v", err)
	}
	if !json.Valid(metadataBody) {
		t.Fatalf("expected valid metadata json: %s", string(metadataBody))
	}
	revisionConfigPath := filepath.Join(filepath.Dir(filepath.Dir(status.ConfigArtifactPath)), "revisions", "4", "config.json")
	if _, err := os.Stat(revisionConfigPath); err != nil {
		t.Fatalf("expected revision config artifact: %v", err)
	}
	stagedConfigPath := filepath.Join(filepath.Dir(filepath.Dir(status.ConfigArtifactPath)), "staged", "config.json")
	if _, err := os.Stat(stagedConfigPath); err != nil {
		t.Fatalf("expected staged config artifact: %v", err)
	}
	statePath := filepath.Join(filepath.Dir(filepath.Dir(status.ConfigArtifactPath)), "state.json")
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("expected state artifact: %v", err)
	}
}

func TestStageFailureLeavesActiveArtifactUntouched(t *testing.T) {
	stateDir := t.TempDir()
	activeDir := filepath.Join(stateDir, "active")
	if err := os.MkdirAll(activeDir, 0o700); err != nil {
		t.Fatalf("expected active dir: %v", err)
	}
	activeConfigPath := filepath.Join(activeDir, "config.json")
	if err := os.WriteFile(activeConfigPath, []byte(`{"revision":"old"}`+"\n"), 0o600); err != nil {
		t.Fatalf("expected old active config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "staged"), []byte("not-a-dir"), 0o600); err != nil {
		t.Fatalf("expected staged file fixture: %v", err)
	}

	service := NewService(Identity{NodeID: "node-1", StateDir: stateDir})
	err := service.ApplyConfigRevision(signedTestConfigRevision(t, "node-1", 4, 3))
	if !errors.Is(err, ErrConfigArtifactWrite) {
		t.Fatalf("expected artifact write error, got %v", err)
	}
	body, err := os.ReadFile(activeConfigPath)
	if err != nil {
		t.Fatalf("expected active config to remain readable: %v", err)
	}
	if string(body) != `{"revision":"old"}`+"\n" {
		t.Fatalf("stage failure changed active config: %s", string(body))
	}
	if service.Status().ActiveRevision != 0 {
		t.Fatalf("stage failure must not advance active revision: %#v", service.Status())
	}
}

func TestActivateFailureLeavesActiveArtifactUntouched(t *testing.T) {
	stateDir := t.TempDir()
	activeDir := filepath.Join(stateDir, "active")
	if err := os.MkdirAll(activeDir, 0o700); err != nil {
		t.Fatalf("expected active dir: %v", err)
	}
	activeConfigPath := filepath.Join(activeDir, "config.json")
	if err := os.WriteFile(activeConfigPath, []byte(`{"revision":"old"}`+"\n"), 0o600); err != nil {
		t.Fatalf("expected old active config: %v", err)
	}

	revision := signedTestConfigRevision(t, "node-1", 4, 3)
	service := NewService(Identity{NodeID: "node-1", StateDir: stateDir})
	staged, err := service.StageConfigRevision(revision)
	if err != nil {
		t.Fatalf("expected staged revision: %v", err)
	}
	if err := os.Remove(staged.StagedConfigPath); err != nil {
		t.Fatalf("expected staged config removal: %v", err)
	}

	_, err = service.ActivateStagedConfigRevision(revision, staged)
	if !errors.Is(err, ErrConfigArtifactWrite) {
		t.Fatalf("expected artifact write error, got %v", err)
	}
	body, err := os.ReadFile(activeConfigPath)
	if err != nil {
		t.Fatalf("expected active config to remain readable: %v", err)
	}
	if string(body) != `{"revision":"old"}`+"\n" {
		t.Fatalf("activation failure changed active config: %s", string(body))
	}
	if service.Status().ActiveRevision != 0 {
		t.Fatalf("activation failure must not advance active revision: %#v", service.Status())
	}
}

func TestApplyRollbackRevisionSwitchesActiveConfigToSourceSemantics(t *testing.T) {
	stateDir := t.TempDir()
	service := NewService(Identity{NodeID: "node-1", StateDir: stateDir})
	revisionA := signedTestConfigRevision(t, "node-1", 1, 0)
	revisionB := signedTestConfigRevision(t, "node-1", 2, 1)
	revisionB.Bundle["config"].(map[string]any)["log"] = map[string]any{"loglevel": "error"}
	resignTestConfigRevision(t, &revisionB)
	rollbackRevision := signedTestConfigRevision(t, "node-1", 3, 2)
	rollbackRevision.Bundle["config"] = revisionA.Bundle["config"]
	rollbackRevision.Bundle["operation_kind"] = "rollback"
	rollbackRevision.Bundle["source_revision_id"] = revisionA.ID
	rollbackRevision.Bundle["source_revision_number"] = revisionA.RevisionNumber
	resignTestConfigRevision(t, &rollbackRevision)

	if err := service.ApplyConfigRevision(revisionA); err != nil {
		t.Fatalf("expected revision A apply: %v", err)
	}
	revisionAConfig, err := os.ReadFile(service.Status().ConfigArtifactPath)
	if err != nil {
		t.Fatalf("expected revision A config: %v", err)
	}
	if err := service.ApplyConfigRevision(revisionB); err != nil {
		t.Fatalf("expected revision B apply: %v", err)
	}
	if err := service.ApplyConfigRevision(rollbackRevision); err != nil {
		t.Fatalf("expected rollback revision apply: %v", err)
	}
	rollbackConfig, err := os.ReadFile(service.Status().ConfigArtifactPath)
	if err != nil {
		t.Fatalf("expected rollback config: %v", err)
	}
	if string(rollbackConfig) != string(revisionAConfig) {
		t.Fatalf("rollback active config must match source config:\n%s\n---\n%s", string(rollbackConfig), string(revisionAConfig))
	}
	if service.Status().ActiveRevision != 3 || service.Status().RollbackCandidateRevision != 2 {
		t.Fatalf("unexpected rollback status: %#v", service.Status())
	}
}

func TestInvalidRollbackRevisionReportsFailedAndKeepsPreviousActive(t *testing.T) {
	stateDir := t.TempDir()
	service := NewService(Identity{NodeID: "node-1", StateDir: stateDir})
	revisionA := signedTestConfigRevision(t, "node-1", 1, 0)
	if err := service.ApplyConfigRevision(revisionA); err != nil {
		t.Fatalf("expected revision A apply: %v", err)
	}
	previousActive, err := os.ReadFile(service.Status().ConfigArtifactPath)
	if err != nil {
		t.Fatalf("expected previous active config: %v", err)
	}

	rollbackRevision := signedTestConfigRevision(t, "node-1", 2, 1)
	rollbackRevision.Bundle["operation_kind"] = "rollback"
	rollbackRevision.Bundle["source_revision_id"] = revisionA.ID
	rollbackRevision.Bundle["source_revision_number"] = revisionA.RevisionNumber
	config := rollbackRevision.Bundle["config"].(map[string]any)
	routing := config["routing"].(map[string]any)
	rules := routing["rules"].([]any)
	rule := rules[0].(map[string]any)
	rule["outboundTag"] = "missing"
	resignTestConfigRevision(t, &rollbackRevision)
	client := &fakePendingConfigRevisionClient{revision: rollbackRevision, ok: true}
	pollingService := NewService(Identity{NodeID: "node-1", NodeToken: "node-token", StateDir: stateDir})
	pollingService.status.ActiveRevision = 1
	pollingService.status.LastAppliedRevision = 1

	applied, err := pollingService.PollPendingConfigRevision(context.Background(), client, time.Now())
	if !errors.Is(err, ErrInvalidXrayConfig) {
		t.Fatalf("expected invalid xray config, got %v", err)
	}
	if applied {
		t.Fatalf("invalid rollback revision must not be applied")
	}
	currentActive, err := os.ReadFile(filepath.Join(stateDir, "active", "config.json"))
	if err != nil {
		t.Fatalf("expected active config after failed rollback: %v", err)
	}
	if string(currentActive) != string(previousActive) {
		t.Fatalf("failed rollback changed active config:\n%s\n---\n%s", string(currentActive), string(previousActive))
	}
	if pollingService.Status().ActiveRevision != 1 || pollingService.Status().LastAppliedRevision != 1 {
		t.Fatalf("failed rollback must keep active revision: %#v", pollingService.Status())
	}
	if !client.reported || client.report.Status != "failed" || client.report.ErrorMessage != "invalid_xray_config:invalid_routing_outbound_reference" {
		t.Fatalf("expected failed rollback validation report, got %#v", client.report)
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
		"config_kind":     "xray-config-compatible-skeleton",
		"operation_kind":  "deploy",
		"node": map[string]any{
			"id": nodeID,
		},
		"transport": map[string]any{
			"network":  "tcp",
			"security": "reality",
			"xtls":     "vision",
		},
		"config": map[string]any{
			"log": map[string]any{
				"loglevel": "warning",
			},
			"inbounds": []any{
				map[string]any{
					"tag":      "vless-reality-in",
					"listen":   "0.0.0.0",
					"port":     443,
					"protocol": "vless",
					"settings": map[string]any{
						"clients": []any{
							map[string]any{
								"id":    "00000000-0000-0000-0000-000000000004",
								"email": "subscription:00000000-0000-0000-0000-000000000004",
								"flow":  "xtls-rprx-vision",
								"level": 0,
							},
						},
						"decryption": "none",
						"fallbacks":  []any{},
					},
					"streamSettings": map[string]any{
						"network":  "tcp",
						"security": "reality",
						"realitySettings": map[string]any{
							"show":         false,
							"dest":         "www.cloudflare.com:443",
							"xver":         0,
							"serverNames":  []any{"www.cloudflare.com"},
							"privateKey":   "lenker-placeholder-reality-private-key",
							"shortIds":     []any{"lenker00"},
							"minClientVer": "",
							"maxClientVer": "",
							"maxTimeDiff":  0,
						},
					},
				},
			},
			"outbounds": []any{
				map[string]any{
					"tag":      "direct",
					"protocol": "freedom",
				},
			},
			"routing": map[string]any{
				"domainStrategy": "AsIs",
				"rules": []any{
					map[string]any{
						"type":        "field",
						"inboundTag":  []any{"vless-reality-in"},
						"outboundTag": "direct",
					},
				},
			},
			"policy": map[string]any{
				"levels": map[string]any{
					"0": map[string]any{
						"handshake":         4,
						"connIdle":          300,
						"uplinkOnly":        2,
						"downlinkOnly":      5,
						"statsUserUplink":   true,
						"statsUserDownlink": true,
					},
				},
				"system": map[string]any{
					"statsInboundUplink":    true,
					"statsInboundDownlink":  true,
					"statsOutboundUplink":   true,
					"statsOutboundDownlink": true,
				},
			},
			"stats": map[string]any{},
		},
		"subscription_inputs": []any{
			map[string]any{
				"subscription_id":     "00000000-0000-0000-0000-000000000004",
				"user_id":             "00000000-0000-0000-0000-000000000001",
				"plan_id":             "00000000-0000-0000-0000-000000000002",
				"subscription_status": "active",
				"user_status":         "active",
				"preferred_region":    "eu",
				"device_limit":        2,
			},
		},
		"access_entries": []any{
			map[string]any{
				"subscription_id": "00000000-0000-0000-0000-000000000004",
				"vless_client_id": "00000000-0000-0000-0000-000000000004",
			},
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
