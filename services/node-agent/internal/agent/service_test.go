package agent

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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

func signedTestConfigRevision(t *testing.T, nodeID string, revisionNumber int, rollbackTarget int) ConfigRevision {
	t.Helper()
	bundle := map[string]any{
		"kind":            "dummy",
		"protocol":        "vless-reality-xtls-vision",
		"xray_runtime":    false,
		"generated_by":    "panel-api",
		"schema_version":  "config-bundle.v1alpha1",
		"revision_number": revisionNumber,
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
