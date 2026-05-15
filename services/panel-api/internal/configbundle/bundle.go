package configbundle

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	DefaultSigner = "lenker-dev-hmac-sha256"
	defaultKey    = "lenker-dev-config-bundle-signing-key"
)

var ErrInvalidSignature = errors.New("invalid config bundle signature")
var ErrInvalidBundleHash = errors.New("invalid config bundle hash")

type Bundle struct {
	NodeID                 string         `json:"node_id"`
	RevisionNumber         int            `json:"revision_number"`
	Status                 string         `json:"status"`
	BundleHash             string         `json:"bundle_hash"`
	Signature              string         `json:"signature"`
	Signer                 string         `json:"signer"`
	RollbackTargetRevision int            `json:"rollback_target_revision"`
	Payload                map[string]any `json:"payload"`
}

type Signer interface {
	Sign(bundle Bundle) (string, error)
}

type Verifier interface {
	Verify(bundle Bundle) error
}

type HMACSigner struct {
	key []byte
}

func NewDevSigner() HMACSigner {
	return HMACSigner{key: []byte(defaultKey)}
}

func (s HMACSigner) Sign(bundle Bundle) (string, error) {
	mac := hmac.New(sha256.New, s.key)
	if _, err := mac.Write([]byte(signingPayload(bundle))); err != nil {
		return "", err
	}
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func (s HMACSigner) Verify(bundle Bundle) error {
	if err := VerifyBundleHash(bundle); err != nil {
		return err
	}
	expected, err := s.Sign(bundle)
	if err != nil {
		return err
	}
	actual, err := hex.DecodeString(bundle.Signature)
	if err != nil {
		return ErrInvalidSignature
	}
	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return err
	}
	if !hmac.Equal(actual, expectedBytes) {
		return ErrInvalidSignature
	}
	return nil
}

func VerifyBundleHash(bundle Bundle) error {
	expected, err := HashPayload(bundle.Payload)
	if err != nil {
		return err
	}
	if bundle.BundleHash != expected {
		return ErrInvalidBundleHash
	}
	return nil
}

func HashPayload(payload map[string]any) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func signingPayload(bundle Bundle) string {
	return fmt.Sprintf("%s\n%d\n%s\n%d", bundle.NodeID, bundle.RevisionNumber, bundle.BundleHash, bundle.RollbackTargetRevision)
}
