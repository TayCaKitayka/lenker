package agent

import "time"

const (
	StatusBootstrapping = "bootstrapping"
	StatusPending       = "pending"
	StatusActive        = "active"
	StatusUnhealthy     = "unhealthy"
	StatusDrained       = "drained"
	StatusDisabled      = "disabled"
)

type Identity struct {
	NodeID         string `json:"node_id,omitempty"`
	BootstrapToken string `json:"-"`
	NodeToken      string `json:"-"`
	PanelURL       string `json:"panel_url,omitempty"`
	StateDir       string `json:"-"`
	XrayBin        string `json:"-"`
}

type Status struct {
	NodeID                    string    `json:"node_id,omitempty"`
	Status                    string    `json:"status"`
	Registered                bool      `json:"registered"`
	PanelURL                  string    `json:"panel_url,omitempty"`
	LastHeartbeatAt           time.Time `json:"last_heartbeat_at,omitempty"`
	ActiveRevision            int       `json:"active_revision"`
	LastAppliedRevision       int       `json:"last_applied_revision"`
	LastRollbackRevision      int       `json:"last_rollback_revision"`
	StagedRevision            int       `json:"staged_revision"`
	RollbackCandidateRevision int       `json:"rollback_candidate_revision"`
	LastValidationStatus      string    `json:"last_validation_status,omitempty"`
	LastValidationError       string    `json:"last_validation_error,omitempty"`
	LastValidationAt          time.Time `json:"last_validation_at,omitempty"`
	ConfigArtifactPath        string    `json:"config_artifact_path,omitempty"`
	MetadataArtifactPath      string    `json:"metadata_artifact_path,omitempty"`
}

type RegistrationPayload struct {
	NodeID         string `json:"node_id,omitempty"`
	BootstrapToken string `json:"bootstrap_token"`
	AgentVersion   string `json:"agent_version"`
	Hostname       string `json:"hostname"`
}

type RegistrationResponse struct {
	NodeID       string    `json:"node_id"`
	NodeToken    string    `json:"node_token"`
	Status       string    `json:"status"`
	DrainState   string    `json:"drain_state,omitempty"`
	RegisteredAt time.Time `json:"registered_at,omitempty"`
}

type HeartbeatPayload struct {
	NodeID               string    `json:"node_id"`
	AgentVersion         string    `json:"agent_version"`
	Status               string    `json:"status"`
	ActiveRevision       int       `json:"active_revision"`
	LastValidationStatus string    `json:"last_validation_status,omitempty"`
	LastValidationError  string    `json:"last_validation_error,omitempty"`
	LastValidationAt     time.Time `json:"last_validation_at,omitempty"`
	LastAppliedRevision  int       `json:"last_applied_revision,omitempty"`
	ActiveConfigPath     string    `json:"active_config_path,omitempty"`
	SentAt               time.Time `json:"sent_at"`
}

type ConfigRevision struct {
	ID                     string         `json:"id,omitempty"`
	NodeID                 string         `json:"node_id,omitempty"`
	RevisionNumber         int            `json:"revision_number"`
	Status                 string         `json:"status"`
	BundleHash             string         `json:"bundle_hash,omitempty"`
	Signature              string         `json:"signature,omitempty"`
	Signer                 string         `json:"signer,omitempty"`
	RollbackTargetRevision int            `json:"rollback_target_revision"`
	Bundle                 map[string]any `json:"bundle,omitempty"`
	CreatedAt              time.Time      `json:"created_at,omitempty"`
	AppliedAt              time.Time      `json:"applied_at,omitempty"`
}

type ConfigRevisionReport struct {
	Status               string    `json:"status"`
	AppliedAt            time.Time `json:"applied_at,omitempty"`
	FailedAt             time.Time `json:"failed_at,omitempty"`
	ErrorMessage         string    `json:"error_message,omitempty"`
	ActiveRevision       int       `json:"active_revision,omitempty"`
	LastValidationStatus string    `json:"last_validation_status,omitempty"`
	LastValidationError  string    `json:"last_validation_error,omitempty"`
	LastValidationAt     time.Time `json:"last_validation_at,omitempty"`
	LastAppliedRevision  int       `json:"last_applied_revision,omitempty"`
	ActiveConfigPath     string    `json:"active_config_path,omitempty"`
	SentAt               time.Time `json:"sent_at,omitempty"`
}

type RollbackPlan struct {
	FromRevision int    `json:"from_revision"`
	ToRevision   int    `json:"to_revision"`
	Reason       string `json:"reason,omitempty"`
}
