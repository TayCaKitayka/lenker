package agent

import (
	"context"
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
	ErrInvalidConfigPayload    = errors.New("invalid config payload")
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
	if err := validateRenderedConfigPayload(revision); err != nil {
		return err
	}
	s.configRevisions[revision.RevisionNumber] = revision
	s.status.ActiveRevision = revision.RevisionNumber
	s.status.LastRollbackRevision = revision.RollbackTargetRevision
	return nil
}

func (s *Service) ApplyConfigRevisionMetadata(revision ConfigRevision) error {
	if err := s.ValidateAndStoreConfigRevision(revision); err != nil {
		return err
	}
	s.TrackAppliedRevision(revision)
	return nil
}

func (s *Service) FetchAndApplyPendingConfigRevision(ctx context.Context, client PendingConfigRevisionClient) (bool, error) {
	if client == nil {
		return false, ErrUnexpectedPanelResponse
	}
	if s.identity.NodeID == "" {
		return false, ErrNodeIDRequired
	}
	if s.identity.NodeToken == "" {
		return false, ErrNodeTokenRequired
	}

	revision, ok, err := client.FetchPendingConfigRevision(ctx, s.identity.NodeID, s.identity.NodeToken)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	if err := s.ApplyConfigRevisionMetadata(revision); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) PollPendingConfigRevision(ctx context.Context, client PendingConfigRevisionClient, now time.Time) (bool, error) {
	if client == nil {
		return false, ErrUnexpectedPanelResponse
	}
	if s.identity.NodeID == "" {
		return false, ErrNodeIDRequired
	}
	if s.identity.NodeToken == "" {
		return false, ErrNodeTokenRequired
	}

	revision, ok, err := client.FetchPendingConfigRevision(ctx, s.identity.NodeID, s.identity.NodeToken)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	reportTime := now.UTC()
	if reportTime.IsZero() {
		reportTime = time.Now().UTC()
	}
	if err := s.ApplyConfigRevisionMetadata(revision); err != nil {
		reportErr := client.ReportConfigRevision(ctx, s.identity.NodeID, s.identity.NodeToken, revision.ID, ConfigRevisionReport{
			Status:       "failed",
			FailedAt:     reportTime,
			ErrorMessage: configRevisionErrorMessage(err),
			SentAt:       reportTime,
		})
		if reportErr != nil {
			return false, reportErr
		}
		return false, err
	}

	if err := client.ReportConfigRevision(ctx, s.identity.NodeID, s.identity.NodeToken, revision.ID, ConfigRevisionReport{
		Status:         "applied",
		AppliedAt:      reportTime,
		ActiveRevision: revision.RevisionNumber,
		SentAt:         reportTime,
	}); err != nil {
		return false, err
	}
	return true, nil
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

func validateRenderedConfigPayload(revision ConfigRevision) error {
	if revision.Bundle == nil {
		return ErrInvalidConfigPayload
	}
	requiredStrings := map[string]string{
		"schema_version": "config-bundle.v1alpha1",
		"generated_by":   "panel-api",
		"protocol":       "vless-reality-xtls-vision",
		"core_type":      "xray",
		"config_kind":    "xray-config-skeleton",
	}
	for key, expected := range requiredStrings {
		value, ok := revision.Bundle[key].(string)
		if !ok || value != expected {
			return ErrInvalidConfigPayload
		}
	}
	if number, ok := numberAsInt(revision.Bundle["revision_number"]); !ok || number != revision.RevisionNumber {
		return ErrInvalidConfigPayload
	}
	if _, ok := revision.Bundle["node"].(map[string]any); !ok {
		return ErrInvalidConfigPayload
	}
	if _, ok := revision.Bundle["transport"].(map[string]any); !ok {
		return ErrInvalidConfigPayload
	}
	config, ok := revision.Bundle["config"].(map[string]any)
	if !ok {
		return ErrInvalidConfigPayload
	}
	if _, ok := config["inbounds"].([]any); !ok {
		return ErrInvalidConfigPayload
	}
	if _, ok := config["outbounds"].([]any); !ok {
		return ErrInvalidConfigPayload
	}
	if _, ok := config["routing"].(map[string]any); !ok {
		return ErrInvalidConfigPayload
	}
	return nil
}

func numberAsInt(value any) (int, bool) {
	switch typedValue := value.(type) {
	case int:
		return typedValue, true
	case int64:
		return int(typedValue), true
	case float64:
		if typedValue != float64(int(typedValue)) {
			return 0, false
		}
		return int(typedValue), true
	default:
		return 0, false
	}
}

func configRevisionErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrInvalidConfigBundleHash):
		return "invalid config bundle hash"
	case errors.Is(err, ErrInvalidConfigSignature):
		return "invalid config bundle signature"
	case errors.Is(err, ErrInvalidConfigPayload):
		return "invalid config payload"
	case errors.Is(err, ErrInvalidConfigRevision):
		return "invalid config revision"
	default:
		return "config revision apply failed"
	}
}
