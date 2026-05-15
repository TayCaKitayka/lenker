package configbundle

import (
	"errors"
	"testing"
)

func TestHashPayloadStable(t *testing.T) {
	first, err := HashPayload(map[string]any{
		"schema_version":  "config-bundle.v1alpha1",
		"revision_number": 1,
		"protocol":        "vless-reality-xtls-vision",
	})
	if err != nil {
		t.Fatalf("expected hash: %v", err)
	}
	second, err := HashPayload(map[string]any{
		"protocol":        "vless-reality-xtls-vision",
		"revision_number": 1,
		"schema_version":  "config-bundle.v1alpha1",
	})
	if err != nil {
		t.Fatalf("expected hash: %v", err)
	}
	if first != second {
		t.Fatalf("expected stable hash, got %q and %q", first, second)
	}
}

func TestDevSignerVerifiesBundle(t *testing.T) {
	payload := map[string]any{"mode": "dummy"}
	hash, err := HashPayload(payload)
	if err != nil {
		t.Fatalf("expected hash: %v", err)
	}

	bundle := Bundle{
		NodeID:                 "node-1",
		RevisionNumber:         1,
		BundleHash:             hash,
		Signer:                 DefaultSigner,
		RollbackTargetRevision: 0,
		Payload:                payload,
	}
	signer := NewDevSigner()
	signature, err := signer.Sign(bundle)
	if err != nil {
		t.Fatalf("expected signature: %v", err)
	}
	bundle.Signature = signature

	if err := signer.Verify(bundle); err != nil {
		t.Fatalf("expected valid signature: %v", err)
	}
	bundle.BundleHash = "tampered"
	if err := signer.Verify(bundle); !errors.Is(err, ErrInvalidBundleHash) {
		t.Fatalf("expected ErrInvalidBundleHash, got %v", err)
	}
}

func TestDevSignerRejectsTamperedSignature(t *testing.T) {
	payload := map[string]any{"mode": "dummy"}
	hash, err := HashPayload(payload)
	if err != nil {
		t.Fatalf("expected hash: %v", err)
	}
	bundle := Bundle{
		NodeID:                 "node-1",
		RevisionNumber:         1,
		BundleHash:             hash,
		Signer:                 DefaultSigner,
		RollbackTargetRevision: 0,
		Payload:                payload,
		Signature:              "bad-signature",
	}
	if err := NewDevSigner().Verify(bundle); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature, got %v", err)
	}
}
