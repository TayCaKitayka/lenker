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
	"path/filepath"
	"strings"
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
	ErrStateDirRequired        = errors.New("state dir is required")
	ErrConfigArtifactWrite     = errors.New("config artifact write failed")
)

type Service struct {
	identity        Identity
	status          Status
	configRevisions map[int]ConfigRevision
	xrayDryRun      XrayDryRunValidator
}

type ServiceOption func(*Service)

func WithXrayDryRunValidator(validator XrayDryRunValidator) ServiceOption {
	return func(s *Service) {
		s.xrayDryRun = validator
	}
}

func NewService(identity Identity, options ...ServiceOption) *Service {
	registered := identity.NodeID != ""
	status := StatusBootstrapping
	if registered {
		status = StatusActive
	}

	service := &Service{
		identity: identity,
		status: Status{
			NodeID:     identity.NodeID,
			Status:     status,
			Registered: registered,
			PanelURL:   identity.PanelURL,
		},
		configRevisions: make(map[int]ConfigRevision),
	}
	if strings.TrimSpace(identity.XrayBin) != "" {
		service.xrayDryRun = CommandXrayDryRunValidator{Binary: identity.XrayBin}
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	return service
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

func (s *Service) ApplyConfigRevision(revision ConfigRevision) error {
	return s.ApplyConfigRevisionWithContext(context.Background(), revision)
}

func (s *Service) ApplyConfigRevisionWithContext(ctx context.Context, revision ConfigRevision) error {
	if err := s.ValidateAndStoreConfigRevision(revision); err != nil {
		return err
	}
	if err := s.ValidateXrayDryRun(ctx, revision); err != nil {
		return err
	}
	artifact, err := s.SerializeConfigRevision(revision)
	if err != nil {
		return err
	}
	s.TrackAppliedRevision(revision)
	s.status.ConfigArtifactPath = artifact.ConfigPath
	s.status.MetadataArtifactPath = artifact.MetadataPath
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
	if err := s.ApplyConfigRevisionWithContext(ctx, revision); err != nil {
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

type ConfigArtifact struct {
	ConfigPath           string
	MetadataPath         string
	RevisionConfigPath   string
	RevisionMetadataPath string
	StagedConfigPath     string
	StagedMetadataPath   string
	StatePath            string
}

func (s *Service) SerializeConfigRevision(revision ConfigRevision) (ConfigArtifact, error) {
	staged, err := s.StageConfigRevision(revision)
	if err != nil {
		return ConfigArtifact{}, err
	}
	return s.ActivateStagedConfigRevision(revision, staged)
}

func (s *Service) ValidateXrayDryRun(ctx context.Context, revision ConfigRevision) error {
	if s.xrayDryRun == nil {
		return nil
	}
	stateDir := strings.TrimSpace(s.identity.StateDir)
	if stateDir == "" {
		return ErrStateDirRequired
	}
	config, ok := revision.Bundle["config"].(map[string]any)
	if !ok {
		return ErrInvalidConfigPayload
	}

	candidateDir := filepath.Join(stateDir, "candidates")
	if err := os.MkdirAll(candidateDir, 0o700); err != nil {
		return fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}
	candidate, err := os.CreateTemp(candidateDir, "candidate-*.json")
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}
	candidatePath := candidate.Name()
	defer os.Remove(candidatePath)

	configBody, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		_ = candidate.Close()
		return err
	}
	configBody = append(configBody, '\n')
	if _, err := candidate.Write(configBody); err != nil {
		_ = candidate.Close()
		return fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}
	if err := candidate.Chmod(0o600); err != nil {
		_ = candidate.Close()
		return fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}
	if err := candidate.Close(); err != nil {
		return fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}

	return s.xrayDryRun.Validate(ctx, candidatePath)
}

func (s *Service) StageConfigRevision(revision ConfigRevision) (ConfigArtifact, error) {
	stateDir := strings.TrimSpace(s.identity.StateDir)
	if stateDir == "" {
		return ConfigArtifact{}, ErrStateDirRequired
	}

	config, ok := revision.Bundle["config"].(map[string]any)
	if !ok {
		return ConfigArtifact{}, ErrInvalidConfigPayload
	}

	revisionDir := filepath.Join(stateDir, "revisions", fmt.Sprintf("%d", revision.RevisionNumber))
	stagedDir := filepath.Join(stateDir, "staged")
	activeDir := filepath.Join(stateDir, "active")
	if err := os.MkdirAll(revisionDir, 0o700); err != nil {
		return ConfigArtifact{}, fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}
	if err := os.MkdirAll(stagedDir, 0o700); err != nil {
		return ConfigArtifact{}, fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}

	configBody, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return ConfigArtifact{}, err
	}
	configBody = append(configBody, '\n')

	configPath := filepath.Join(revisionDir, "config.json")
	metadataPath := filepath.Join(revisionDir, "metadata.json")
	stagedConfigPath := filepath.Join(stagedDir, "config.json")
	stagedMetadataPath := filepath.Join(stagedDir, "metadata.json")
	activeConfigPath := filepath.Join(activeDir, "config.json")
	activeMetadataPath := filepath.Join(activeDir, "metadata.json")
	statePath := filepath.Join(stateDir, "state.json")

	metadata := map[string]any{
		"revision_id":              revision.ID,
		"node_id":                  revision.NodeID,
		"revision_number":          revision.RevisionNumber,
		"bundle_hash":              revision.BundleHash,
		"signer":                   revision.Signer,
		"rollback_target_revision": revision.RollbackTargetRevision,
		"operation_kind":           stringFromBundle(revision.Bundle, "operation_kind"),
		"source_revision_id":       stringFromBundle(revision.Bundle, "source_revision_id"),
		"source_revision_number":   numberFromBundle(revision.Bundle, "source_revision_number"),
		"config_path":              configPath,
		"staged_config_path":       stagedConfigPath,
		"active_config_path":       activeConfigPath,
		"apply_mode":               "staged-active-file-switch",
	}
	metadataBody, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return ConfigArtifact{}, err
	}
	metadataBody = append(metadataBody, '\n')

	for _, write := range []struct {
		path string
		body []byte
	}{
		{path: configPath, body: configBody},
		{path: metadataPath, body: metadataBody},
		{path: stagedConfigPath, body: configBody},
		{path: stagedMetadataPath, body: metadataBody},
	} {
		if err := writeFileAtomic(write.path, write.body, 0o600); err != nil {
			return ConfigArtifact{}, fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
		}
	}

	s.status.StagedRevision = revision.RevisionNumber
	s.status.RollbackCandidateRevision = revision.RollbackTargetRevision

	return ConfigArtifact{
		ConfigPath:           activeConfigPath,
		MetadataPath:         activeMetadataPath,
		RevisionConfigPath:   configPath,
		RevisionMetadataPath: metadataPath,
		StagedConfigPath:     stagedConfigPath,
		StagedMetadataPath:   stagedMetadataPath,
		StatePath:            statePath,
	}, nil
}

func (s *Service) ActivateStagedConfigRevision(revision ConfigRevision, artifact ConfigArtifact) (ConfigArtifact, error) {
	stateDir := strings.TrimSpace(s.identity.StateDir)
	if stateDir == "" {
		return ConfigArtifact{}, ErrStateDirRequired
	}

	activeDir := filepath.Join(stateDir, "active")
	if err := os.MkdirAll(activeDir, 0o700); err != nil {
		return ConfigArtifact{}, fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}

	configBody, err := os.ReadFile(artifact.StagedConfigPath)
	if err != nil {
		return ConfigArtifact{}, fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}
	if !json.Valid(configBody) {
		return ConfigArtifact{}, ErrInvalidConfigPayload
	}
	metadataBody, err := os.ReadFile(artifact.StagedMetadataPath)
	if err != nil {
		return ConfigArtifact{}, fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}
	if !json.Valid(metadataBody) {
		return ConfigArtifact{}, ErrInvalidConfigPayload
	}

	if artifact.ConfigPath == "" {
		artifact.ConfigPath = filepath.Join(activeDir, "config.json")
	}
	if artifact.MetadataPath == "" {
		artifact.MetadataPath = filepath.Join(activeDir, "metadata.json")
	}
	if artifact.StatePath == "" {
		artifact.StatePath = filepath.Join(stateDir, "state.json")
	}

	if err := writeFileAtomic(artifact.MetadataPath, metadataBody, 0o600); err != nil {
		return ConfigArtifact{}, fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}
	if err := writeFileAtomic(artifact.ConfigPath, configBody, 0o600); err != nil {
		return ConfigArtifact{}, fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}

	state := map[string]any{
		"active_revision":             revision.RevisionNumber,
		"staged_revision":             revision.RevisionNumber,
		"last_applied_revision":       revision.RevisionNumber,
		"rollback_candidate_revision": revision.RollbackTargetRevision,
		"config_artifact_path":        artifact.ConfigPath,
		"metadata_artifact_path":      artifact.MetadataPath,
		"revision_config_path":        artifact.RevisionConfigPath,
		"revision_metadata_path":      artifact.RevisionMetadataPath,
		"operation_kind":              stringFromBundle(revision.Bundle, "operation_kind"),
		"source_revision_id":          stringFromBundle(revision.Bundle, "source_revision_id"),
		"source_revision_number":      numberFromBundle(revision.Bundle, "source_revision_number"),
	}
	stateBody, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return ConfigArtifact{}, err
	}
	stateBody = append(stateBody, '\n')
	if err := writeFileAtomic(artifact.StatePath, stateBody, 0o600); err != nil {
		return ConfigArtifact{}, fmt.Errorf("%w: %v", ErrConfigArtifactWrite, err)
	}

	return artifact, nil
}

func writeFileAtomic(path string, body []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	if _, err := temp.Write(body); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Chmod(perm); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, path)
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
		"config_kind":    "xray-config-compatible-skeleton",
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
	if operationKind, ok := revision.Bundle["operation_kind"].(string); !ok || (operationKind != "deploy" && operationKind != "rollback") {
		return ErrInvalidConfigPayload
	}
	if _, ok := revision.Bundle["subscription_inputs"].([]any); !ok {
		return ErrInvalidConfigPayload
	}
	if _, ok := revision.Bundle["access_entries"].([]any); !ok {
		return ErrInvalidConfigPayload
	}
	return ValidateXrayConfigArtifact(config)
}

func stringFromBundle(bundle map[string]any, key string) string {
	value, _ := bundle[key].(string)
	return value
}

func numberFromBundle(bundle map[string]any, key string) int {
	value, _ := numberAsInt(bundle[key])
	return value
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
	case errors.Is(err, ErrInvalidXrayConfig):
		var validationErr ConfigValidationError
		if errors.As(err, &validationErr) && validationErr.Reason != "" {
			return "invalid_xray_config:" + validationErr.Reason
		}
		return "invalid_xray_config"
	case errors.Is(err, ErrXrayDryRunFailed):
		var dryRunErr XrayDryRunError
		if errors.As(err, &dryRunErr) && dryRunErr.Reason != "" {
			return "xray_dry_run_failed:" + dryRunErr.Reason
		}
		return "xray_dry_run_failed"
	case errors.Is(err, ErrInvalidConfigPayload):
		return "invalid config payload"
	case errors.Is(err, ErrInvalidConfigRevision):
		return "invalid config revision"
	case errors.Is(err, ErrStateDirRequired):
		return "state dir is required"
	case errors.Is(err, ErrConfigArtifactWrite):
		return "config artifact write failed"
	default:
		return "config revision apply failed"
	}
}
