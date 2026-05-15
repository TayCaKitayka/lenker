package agent

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

const AgentVersion = "0.1.0-dev"
const DevConfigBundleSigner = "lenker-dev-hmac-sha256"
const devConfigBundleKey = "lenker-dev-config-bundle-signing-key"

var (
	ErrBootstrapTokenRequired  = errors.New("bootstrap token is required")
	ErrNodeIDRequired          = errors.New("node id is required")
	ErrInvalidConfigRevision   = errors.New("invalid config revision")
	ErrInvalidConfigBundleHash = errors.New("invalid config bundle hash")
	ErrInvalidConfigSignature  = errors.New("invalid config bundle signature")
)

type Service struct {
	identity        Identity
	status          Status
	configRevisions map[int]ConfigRevision
}

func NewService(identity Identity) *Service {
	registered := identity.NodeID != ""
	status := StatusBootstrapping
	if registered {
		status = StatusActive
	}

	return &Service{
		identity: identity,
		status: Status{
			NodeID:     identity.NodeID,
			Status:     status,
			Registered: registered,
			PanelURL:   identity.PanelURL,
		},
		configRevisions: make(map[int]ConfigRevision),
	}
}

func (s *Service) Status() Status {
	return s.status
}

func (s *Service) BuildRegistrationPayload() (RegistrationPayload, error) {
	if s.identity.BootstrapToken == "" {
		return RegistrationPayload{}, ErrBootstrapTokenRequired
	}

	hostname, _ := os.Hostname()
	return RegistrationPayload{
		NodeID:         s.identity.NodeID,
		BootstrapToken: s.identity.BootstrapToken,
		AgentVersion:   AgentVersion,
		Hostname:       hostname,
	}, nil
}

func (s *Service) BuildHeartbeatPayload(now time.Time) (HeartbeatPayload, error) {
	if s.identity.NodeID == "" {
		return HeartbeatPayload{}, ErrNodeIDRequired
	}

	return HeartbeatPayload{
		NodeID:         s.identity.NodeID,
		AgentVersion:   AgentVersion,
		Status:         s.status.Status,
		ActiveRevision: s.status.ActiveRevision,
		SentAt:         now.UTC(),
	}, nil
}

func (s *Service) MarkHeartbeatSent(at time.Time) {
	s.status.LastHeartbeatAt = at.UTC()
	if s.status.Status == StatusPending {
		s.status.Status = StatusActive
	}
}

func (s *Service) TrackAppliedRevision(revision ConfigRevision) {
	s.status.ActiveRevision = revision.RevisionNumber
	s.status.LastAppliedRevision = revision.RevisionNumber
}

func (s *Service) ValidateAndStoreConfigRevision(revision ConfigRevision) error {
	if revision.NodeID == "" || revision.RevisionNumber <= 0 || revision.BundleHash == "" || revision.Signature == "" || revision.Signer == "" {
		return ErrInvalidConfigRevision
	}
	if s.identity.NodeID != "" && revision.NodeID != s.identity.NodeID {
		return ErrInvalidConfigRevision
	}
	if revision.Signer != DevConfigBundleSigner {
		return ErrInvalidConfigSignature
	}
	if err := verifyConfigBundleHash(revision); err != nil {
		return err
	}
	if err := verifyConfigSignature(revision); err != nil {
		return err
	}
	s.configRevisions[revision.RevisionNumber] = revision
	s.status.ActiveRevision = revision.RevisionNumber
	s.status.LastRollbackRevision = revision.RollbackTargetRevision
	return nil
}

func (s *Service) ConfigRevision(revisionNumber int) (ConfigRevision, bool) {
	revision, ok := s.configRevisions[revisionNumber]
	return revision, ok
}

func (s *Service) PlanRollback(toRevision int, reason string) RollbackPlan {
	return RollbackPlan{
		FromRevision: s.status.ActiveRevision,
		ToRevision:   toRevision,
		Reason:       reason,
	}
}

func verifyConfigBundleHash(revision ConfigRevision) error {
	body, err := json.Marshal(revision.Bundle)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(body)
	if revision.BundleHash != hex.EncodeToString(sum[:]) {
		return ErrInvalidConfigBundleHash
	}
	return nil
}

func verifyConfigSignature(revision ConfigRevision) error {
	mac := hmac.New(sha256.New, []byte(devConfigBundleKey))
	if _, err := mac.Write([]byte(configSigningPayload(revision))); err != nil {
		return err
	}
	expected := mac.Sum(nil)
	actual, err := hex.DecodeString(revision.Signature)
	if err != nil {
		return ErrInvalidConfigSignature
	}
	if !hmac.Equal(actual, expected) {
		return ErrInvalidConfigSignature
	}
	return nil
}

func configSigningPayload(revision ConfigRevision) string {
	return fmt.Sprintf("%s\n%d\n%s\n%d", revision.NodeID, revision.RevisionNumber, revision.BundleHash, revision.RollbackTargetRevision)
}
