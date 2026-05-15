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
	PanelURL       string `json:"panel_url,omitempty"`
}

type Status struct {
	NodeID               string    `json:"node_id,omitempty"`
	Status               string    `json:"status"`
	Registered           bool      `json:"registered"`
	PanelURL             string    `json:"panel_url,omitempty"`
	LastHeartbeatAt      time.Time `json:"last_heartbeat_at,omitempty"`
	ActiveRevision       int       `json:"active_revision"`
	LastAppliedRevision  int       `json:"last_applied_revision"`
	LastRollbackRevision int       `json:"last_rollback_revision"`
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
	NodeID         string    `json:"node_id"`
	AgentVersion   string    `json:"agent_version"`
	Status         string    `json:"status"`
	ActiveRevision int       `json:"active_revision"`
	SentAt         time.Time `json:"sent_at"`
}

type ConfigRevision struct {
	RevisionNumber int       `json:"revision_number"`
	BundleHash     string    `json:"bundle_hash,omitempty"`
	Status         string    `json:"status"`
	AppliedAt      time.Time `json:"applied_at,omitempty"`
}

type RollbackPlan struct {
	FromRevision int    `json:"from_revision"`
	ToRevision   int    `json:"to_revision"`
	Reason       string `json:"reason,omitempty"`
}
