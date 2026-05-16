package storage

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/configbundle"
	"github.com/lenker/lenker/services/panel-api/internal/configrender"
)

func TestScanConfigRevisionReturnsSignedPayloadAsBundle(t *testing.T) {
	payload := configrender.RenderVLESSRealityPayload(configrender.RenderInput{
		NodeID:         "node-1",
		RevisionNumber: 1,
		Hostname:       "node-1.example.com",
		Region:         "eu",
		CountryCode:    "FI",
	})
	hash, err := configbundle.HashPayload(payload)
	if err != nil {
		t.Fatalf("expected payload hash: %v", err)
	}
	bundle := configbundle.Bundle{
		NodeID:                 "node-1",
		RevisionNumber:         1,
		Status:                 "pending",
		BundleHash:             hash,
		Signature:              "signature",
		Signer:                 configbundle.DefaultSigner,
		RollbackTargetRevision: 0,
		Payload:                payload,
	}
	bundleJSON, err := json.Marshal(bundle)
	if err != nil {
		t.Fatalf("expected bundle json: %v", err)
	}

	createdAt := time.Date(2026, 5, 15, 1, 2, 3, 0, time.UTC)
	revision, err := scanConfigRevision(fakeRow{
		"id-1",
		"node-1",
		1,
		hash,
		"signature",
		configbundle.DefaultSigner,
		"pending",
		nil,
		bundleJSON,
		createdAt,
		nil,
		nil,
		nil,
		nil,
	})
	if err != nil {
		t.Fatalf("expected revision: %v", err)
	}
	if revision.BundleHash != hash {
		t.Fatalf("unexpected hash: %q", revision.BundleHash)
	}
	if revision.Bundle["payload"] != nil || revision.Bundle["signature"] != nil {
		t.Fatalf("response bundle must be payload only, got %#v", revision.Bundle)
	}
	if revision.Bundle["protocol"] != "vless-reality-xtls-vision" {
		t.Fatalf("expected config payload in response bundle, got %#v", revision.Bundle)
	}
	if revision.Bundle["config_kind"] != "xray-config-skeleton" {
		t.Fatalf("expected xray skeleton payload in response bundle, got %#v", revision.Bundle)
	}
	responseHash, err := configbundle.HashPayload(revision.Bundle)
	if err != nil {
		t.Fatalf("expected response bundle hash: %v", err)
	}
	if responseHash != revision.BundleHash {
		t.Fatalf("response bundle hash mismatch: %q != %q", responseHash, revision.BundleHash)
	}
}

type fakeRow []any

func (r fakeRow) Scan(dest ...any) error {
	for i := range dest {
		switch target := dest[i].(type) {
		case *string:
			if value, ok := r[i].(string); ok {
				*target = value
			}
		case *int:
			if value, ok := r[i].(int); ok {
				*target = value
			}
		case *[]byte:
			if value, ok := r[i].([]byte); ok {
				*target = value
			}
		case *time.Time:
			if value, ok := r[i].(time.Time); ok {
				*target = value
			}
		default:
			// Nullable SQL fields stay invalid for this focused contract test.
		}
	}
	return nil
}
