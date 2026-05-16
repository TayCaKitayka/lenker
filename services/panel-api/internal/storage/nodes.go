package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/configbundle"
	"github.com/lenker/lenker/services/panel-api/internal/configrender"
)

var (
	ErrInvalidBootstrapToken = errors.New("invalid bootstrap token")
	ErrExpiredBootstrapToken = errors.New("expired bootstrap token")
	ErrBootstrapTokenUsed    = errors.New("bootstrap token already used")
	ErrInvalidNodeStatus     = errors.New("invalid node status")
	ErrInvalidNodeTransition = errors.New("invalid node transition")
)

type Node struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Region         string     `json:"region"`
	CountryCode    string     `json:"country_code"`
	Hostname       string     `json:"hostname"`
	Status         string     `json:"status"`
	DrainState     string     `json:"drain_state"`
	AgentVersion   string     `json:"agent_version"`
	XrayVersion    string     `json:"xray_version"`
	ActiveRevision int        `json:"active_revision"`
	LastHealthAt   *time.Time `json:"last_health_at"`
	LastSeenAt     *time.Time `json:"last_seen_at"`
	RegisteredAt   *time.Time `json:"registered_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateBootstrapTokenInput struct {
	Name             string
	Region           string
	CountryCode      string
	Hostname         string
	ExpiresAt        time.Time
	CreatedByAdminID string
}

type BootstrapToken struct {
	ID        string    `json:"id"`
	NodeID    string    `json:"node_id"`
	Token     string    `json:"bootstrap_token,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterNodeInput struct {
	NodeID         string
	BootstrapToken string
	AgentVersion   string
	Hostname       string
}

type RegisterNodeResult struct {
	Node      Node
	NodeToken string
}

type HeartbeatInput struct {
	NodeID         string
	NodeToken      string
	AgentVersion   string
	Status         string
	ActiveRevision int
	SentAt         time.Time
}

type ConfigRevision struct {
	ID                     string         `json:"id"`
	NodeID                 string         `json:"node_id"`
	RevisionNumber         int            `json:"revision_number"`
	BundleHash             string         `json:"bundle_hash"`
	Signature              string         `json:"signature"`
	Signer                 string         `json:"signer"`
	Status                 string         `json:"status"`
	RollbackTargetRevision int            `json:"rollback_target_revision"`
	Bundle                 map[string]any `json:"bundle"`
	CreatedAt              time.Time      `json:"created_at"`
	AppliedAt              *time.Time     `json:"applied_at"`
	FailedAt               *time.Time     `json:"failed_at"`
	RolledBackAt           *time.Time     `json:"rolled_back_at"`
	ErrorMessage           string         `json:"error_message,omitempty"`
}

type CreateDummyConfigRevisionInput struct {
	NodeID           string
	CreatedByAdminID string
}

type CreateRollbackConfigRevisionInput struct {
	NodeID           string
	RevisionID       string
	CreatedByAdminID string
}

type ReportConfigRevisionInput struct {
	NodeID       string
	NodeToken    string
	RevisionID   string
	Status       string
	AppliedAt    time.Time
	FailedAt     time.Time
	ErrorMessage string
	SentAt       time.Time
}

type NodesRepository interface {
	List(ctx context.Context) ([]Node, error)
	FindByID(ctx context.Context, id string) (Node, error)
	CreateBootstrapToken(ctx context.Context, input CreateBootstrapTokenInput) (BootstrapToken, error)
	Register(ctx context.Context, input RegisterNodeInput) (RegisterNodeResult, error)
	RecordHeartbeat(ctx context.Context, input HeartbeatInput) (Node, error)
	Drain(ctx context.Context, id string) (Node, error)
	Undrain(ctx context.Context, id string) (Node, error)
	Disable(ctx context.Context, id string) (Node, error)
	Enable(ctx context.Context, id string) (Node, error)
	CreateDummyConfigRevision(ctx context.Context, input CreateDummyConfigRevisionInput) (ConfigRevision, error)
	CreateRollbackConfigRevision(ctx context.Context, input CreateRollbackConfigRevisionInput) (ConfigRevision, error)
	ListConfigRevisions(ctx context.Context, nodeID string) ([]ConfigRevision, error)
	FindConfigRevision(ctx context.Context, nodeID string, revisionID string) (ConfigRevision, error)
	FindLatestPendingConfigRevision(ctx context.Context, nodeID string, nodeToken string) (ConfigRevision, error)
	ReportConfigRevision(ctx context.Context, input ReportConfigRevisionInput) (ConfigRevision, error)
}

type nodesRepository struct {
	db *sql.DB
}

func NewNodesRepository(db *sql.DB) NodesRepository {
	return &nodesRepository{db: db}
}

func (r *nodesRepository) List(ctx context.Context) ([]Node, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, name, region, country_code, hostname, status, drain_state, agent_version, xray_version, active_revision, last_health_at, last_seen_at, registered_at, updated_at
		FROM nodes
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := []Node{}
	for rows.Next() {
		node, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *nodesRepository) FindByID(ctx context.Context, id string) (Node, error) {
	node, err := scanNode(r.db.QueryRowContext(ctx, `
		SELECT id::text, name, region, country_code, hostname, status, drain_state, agent_version, xray_version, active_revision, last_health_at, last_seen_at, registered_at, updated_at
		FROM nodes
		WHERE id = $1
	`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Node{}, ErrNotFound
		}
		return Node{}, err
	}
	return node, nil
}

func (r *nodesRepository) CreateBootstrapToken(ctx context.Context, input CreateBootstrapTokenInput) (BootstrapToken, error) {
	token, err := newBootstrapToken()
	if err != nil {
		return BootstrapToken{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return BootstrapToken{}, err
	}
	defer tx.Rollback()

	var nodeID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO nodes (name, region, country_code, hostname, status, drain_state)
		VALUES ($1, $2, $3, $4, 'pending', 'active')
		RETURNING id::text
	`, input.Name, input.Region, input.CountryCode, input.Hostname).Scan(&nodeID)
	if err != nil {
		return BootstrapToken{}, err
	}

	var result BootstrapToken
	var createdBy sql.NullString
	if input.CreatedByAdminID != "" {
		createdBy = sql.NullString{String: input.CreatedByAdminID, Valid: true}
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO node_bootstrap_tokens (node_id, token_hash, expires_at, created_by_admin_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, node_id::text, expires_at, created_at
	`, nodeID, HashBootstrapToken(token), input.ExpiresAt, createdBy).Scan(
		&result.ID,
		&result.NodeID,
		&result.ExpiresAt,
		&result.CreatedAt,
	)
	if err != nil {
		return BootstrapToken{}, err
	}

	if err := tx.Commit(); err != nil {
		return BootstrapToken{}, err
	}

	result.Token = token
	return result, nil
}

func (r *nodesRepository) Register(ctx context.Context, input RegisterNodeInput) (RegisterNodeResult, error) {
	nodeToken, err := newNodeToken()
	if err != nil {
		return RegisterNodeResult{}, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return RegisterNodeResult{}, err
	}
	defer tx.Rollback()

	tokenHash := HashBootstrapToken(input.BootstrapToken)
	now := time.Now().UTC()

	var bootstrapID string
	var tokenNodeID string
	var tokenStatus string
	var expiresAt time.Time
	var usedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT id::text, node_id::text, status, expires_at, used_at
		FROM node_bootstrap_tokens
		WHERE token_hash = $1
		FOR UPDATE
	`, tokenHash).Scan(&bootstrapID, &tokenNodeID, &tokenStatus, &expiresAt, &usedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RegisterNodeResult{}, ErrInvalidBootstrapToken
		}
		return RegisterNodeResult{}, err
	}
	if tokenStatus == "used" || usedAt.Valid {
		return RegisterNodeResult{}, ErrBootstrapTokenUsed
	}
	if tokenStatus != "active" {
		return RegisterNodeResult{}, ErrInvalidBootstrapToken
	}
	if !expiresAt.After(now) {
		return RegisterNodeResult{}, ErrExpiredBootstrapToken
	}
	if input.NodeID != "" && input.NodeID != tokenNodeID {
		return RegisterNodeResult{}, ErrInvalidBootstrapToken
	}

	var node Node
	var lastHealthAt sql.NullTime
	var lastSeenAt sql.NullTime
	var registeredAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		UPDATE nodes
		SET hostname = CASE WHEN $2 = '' THEN hostname ELSE $2 END,
		    agent_version = $3,
		    auth_token_hash = $4,
		    status = 'active',
		    registered_at = COALESCE(registered_at, $5),
		    last_seen_at = $5,
		    last_health_at = $5,
		    updated_at = $5
		WHERE id = $1
		  AND status IN ('pending', 'active', 'unhealthy', 'drained')
		RETURNING id::text, name, region, country_code, hostname, status, drain_state, agent_version, xray_version, active_revision, last_health_at, last_seen_at, registered_at, updated_at
	`, tokenNodeID, input.Hostname, input.AgentVersion, HashNodeToken(nodeToken), now).Scan(
		&node.ID,
		&node.Name,
		&node.Region,
		&node.CountryCode,
		&node.Hostname,
		&node.Status,
		&node.DrainState,
		&node.AgentVersion,
		&node.XrayVersion,
		&node.ActiveRevision,
		&lastHealthAt,
		&lastSeenAt,
		&registeredAt,
		&node.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RegisterNodeResult{}, ErrInvalidBootstrapToken
		}
		return RegisterNodeResult{}, err
	}

	if lastHealthAt.Valid {
		node.LastHealthAt = &lastHealthAt.Time
	}
	if lastSeenAt.Valid {
		node.LastSeenAt = &lastSeenAt.Time
	}
	if registeredAt.Valid {
		node.RegisteredAt = &registeredAt.Time
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE node_bootstrap_tokens
		SET status = 'used',
		    used_at = $2,
		    updated_at = $2
		WHERE id = $1
	`, bootstrapID, now)
	if err != nil {
		return RegisterNodeResult{}, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO node_registrations (node_id, bootstrap_token_hash, registration_status)
		VALUES ($1, $2, 'completed')
	`, node.ID, tokenHash)
	if err != nil {
		return RegisterNodeResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return RegisterNodeResult{}, err
	}

	return RegisterNodeResult{Node: node, NodeToken: nodeToken}, nil
}

func (r *nodesRepository) RecordHeartbeat(ctx context.Context, input HeartbeatInput) (Node, error) {
	if !isValidNodeStatus(input.Status) {
		return Node{}, ErrInvalidNodeStatus
	}

	var node Node
	var lastHealthAt sql.NullTime
	var lastSeenAt sql.NullTime
	var registeredAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		UPDATE nodes
		SET status = CASE WHEN drain_state = 'drained' THEN 'drained' ELSE $3 END,
		    agent_version = $4,
		    active_revision = $5,
		    last_health_at = $6,
		    last_seen_at = $6,
		    updated_at = now()
		WHERE id = $1
		  AND auth_token_hash = $2
		  AND registered_at IS NOT NULL
		  AND status != 'disabled'
		RETURNING id::text, name, region, country_code, hostname, status, drain_state, agent_version, xray_version, active_revision, last_health_at, last_seen_at, registered_at, updated_at
	`, input.NodeID, HashNodeToken(input.NodeToken), input.Status, input.AgentVersion, input.ActiveRevision, input.SentAt).Scan(
		&node.ID,
		&node.Name,
		&node.Region,
		&node.CountryCode,
		&node.Hostname,
		&node.Status,
		&node.DrainState,
		&node.AgentVersion,
		&node.XrayVersion,
		&node.ActiveRevision,
		&lastHealthAt,
		&lastSeenAt,
		&registeredAt,
		&node.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Node{}, ErrNotFound
		}
		return Node{}, err
	}
	if lastHealthAt.Valid {
		node.LastHealthAt = &lastHealthAt.Time
	}
	if lastSeenAt.Valid {
		node.LastSeenAt = &lastSeenAt.Time
	}
	if registeredAt.Valid {
		node.RegisteredAt = &registeredAt.Time
	}
	return node, nil
}

func (r *nodesRepository) Drain(ctx context.Context, id string) (Node, error) {
	return r.transition(ctx, id, func(node Node) (nodeTransition, error) {
		if node.Status == "disabled" {
			return nodeTransition{}, ErrInvalidNodeTransition
		}
		if node.DrainState == "draining" {
			return nodeTransition{}, ErrInvalidNodeTransition
		}
		return nodeTransition{DrainState: "draining"}, nil
	})
}

func (r *nodesRepository) Undrain(ctx context.Context, id string) (Node, error) {
	return r.transition(ctx, id, func(node Node) (nodeTransition, error) {
		if node.Status == "disabled" {
			return nodeTransition{}, ErrInvalidNodeTransition
		}
		if node.DrainState == "active" {
			return nodeTransition{}, ErrInvalidNodeTransition
		}
		return nodeTransition{DrainState: "active"}, nil
	})
}

func (r *nodesRepository) Disable(ctx context.Context, id string) (Node, error) {
	return r.transition(ctx, id, func(node Node) (nodeTransition, error) {
		if node.Status == "disabled" {
			return nodeTransition{}, ErrInvalidNodeTransition
		}
		return nodeTransition{Status: "disabled"}, nil
	})
}

func (r *nodesRepository) Enable(ctx context.Context, id string) (Node, error) {
	return r.transition(ctx, id, func(node Node) (nodeTransition, error) {
		if node.Status != "disabled" {
			return nodeTransition{}, ErrInvalidNodeTransition
		}
		return nodeTransition{Status: "unhealthy"}, nil
	})
}

func (r *nodesRepository) CreateDummyConfigRevision(ctx context.Context, input CreateDummyConfigRevisionInput) (ConfigRevision, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ConfigRevision{}, err
	}
	defer tx.Rollback()

	var currentRevision int
	var status string
	var drainState string
	err = tx.QueryRowContext(ctx, `
		SELECT active_revision, status, drain_state
		FROM nodes
		WHERE id = $1
		FOR UPDATE
	`, input.NodeID).Scan(&currentRevision, &status, &drainState)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ConfigRevision{}, ErrNotFound
		}
		return ConfigRevision{}, err
	}
	if status == "disabled" || drainState != "active" {
		return ConfigRevision{}, ErrInvalidNodeTransition
	}

	var nextRevision int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(revision_number), 0) + 1
		FROM config_revisions
		WHERE node_id = $1
	`, input.NodeID).Scan(&nextRevision)
	if err != nil {
		return ConfigRevision{}, err
	}

	var hostname string
	var region string
	var countryCode string
	err = tx.QueryRowContext(ctx, `
		SELECT hostname, region, country_code
		FROM nodes
		WHERE id = $1
	`, input.NodeID).Scan(&hostname, &region, &countryCode)
	if err != nil {
		return ConfigRevision{}, err
	}
	subscriptionInputs, err := listConfigSubscriptionInputs(ctx, tx, region)
	if err != nil {
		return ConfigRevision{}, err
	}

	payload := configrender.RenderVLESSRealityPayload(configrender.RenderInput{
		NodeID:                 input.NodeID,
		RevisionNumber:         nextRevision,
		Hostname:               hostname,
		Region:                 region,
		CountryCode:            countryCode,
		RollbackTargetRevision: currentRevision,
		SubscriptionInputs:     subscriptionInputs,
	})
	if err := configrender.ValidateVLESSRealityPayload(payload); err != nil {
		return ConfigRevision{}, fmt.Errorf("%w: %v", ErrInvalidNodeTransition, err)
	}
	bundleHash, err := configbundle.HashPayload(payload)
	if err != nil {
		return ConfigRevision{}, err
	}
	bundle := configbundle.Bundle{
		NodeID:                 input.NodeID,
		RevisionNumber:         nextRevision,
		Status:                 "pending",
		BundleHash:             bundleHash,
		Signer:                 configbundle.DefaultSigner,
		RollbackTargetRevision: currentRevision,
		Payload:                payload,
	}
	signature, err := configbundle.NewDevSigner().Sign(bundle)
	if err != nil {
		return ConfigRevision{}, err
	}
	bundle.Signature = signature

	bundleJSON, err := json.Marshal(bundle)
	if err != nil {
		return ConfigRevision{}, err
	}

	var createdBy sql.NullString
	if input.CreatedByAdminID != "" {
		createdBy = sql.NullString{String: input.CreatedByAdminID, Valid: true}
	}

	revision, err := scanConfigRevision(tx.QueryRowContext(ctx, `
		INSERT INTO config_revisions (
		    node_id,
		    revision_number,
		    bundle_hash,
		    signature,
		    signer,
		    status,
		    rollback_target_revision,
		    bundle_json,
		    created_by_admin_id
		)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6, $7, $8)
		RETURNING id::text, node_id::text, revision_number, bundle_hash, signature, signer, status, rollback_target_revision, bundle_json, created_at, applied_at, failed_at, rolled_back_at, error_message
	`, input.NodeID, nextRevision, bundleHash, signature, configbundle.DefaultSigner, currentRevision, bundleJSON, createdBy))
	if err != nil {
		return ConfigRevision{}, err
	}
	if err := tx.Commit(); err != nil {
		return ConfigRevision{}, err
	}
	return revision, nil
}

func (r *nodesRepository) CreateRollbackConfigRevision(ctx context.Context, input CreateRollbackConfigRevisionInput) (ConfigRevision, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ConfigRevision{}, err
	}
	defer tx.Rollback()

	var currentRevision int
	var status string
	var drainState string
	err = tx.QueryRowContext(ctx, `
		SELECT active_revision, status, drain_state
		FROM nodes
		WHERE id = $1
		FOR UPDATE
	`, input.NodeID).Scan(&currentRevision, &status, &drainState)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ConfigRevision{}, ErrNotFound
		}
		return ConfigRevision{}, err
	}
	if status == "disabled" || drainState != "active" {
		return ConfigRevision{}, ErrInvalidNodeTransition
	}

	var target ConfigRevision
	target, err = scanConfigRevision(tx.QueryRowContext(ctx, `
		SELECT id::text, node_id::text, revision_number, bundle_hash, signature, signer, status, rollback_target_revision, bundle_json, created_at, applied_at, failed_at, rolled_back_at, error_message
		FROM config_revisions
		WHERE node_id = $1
		  AND id = $2
		FOR UPDATE
	`, input.NodeID, input.RevisionID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ConfigRevision{}, ErrNotFound
		}
		return ConfigRevision{}, err
	}
	if target.Status != "applied" {
		return ConfigRevision{}, ErrInvalidNodeTransition
	}

	var nextRevision int
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(revision_number), 0) + 1
		FROM config_revisions
		WHERE node_id = $1
	`, input.NodeID).Scan(&nextRevision)
	if err != nil {
		return ConfigRevision{}, err
	}

	payload, err := configrender.RenderRollbackPayload(target.Bundle, configrender.RollbackInput{
		RevisionNumber:         nextRevision,
		RollbackTargetRevision: currentRevision,
		SourceRevisionID:       target.ID,
		SourceRevisionNumber:   target.RevisionNumber,
	})
	if err != nil {
		return ConfigRevision{}, err
	}
	if err := configrender.ValidateVLESSRealityPayload(payload); err != nil {
		return ConfigRevision{}, fmt.Errorf("%w: %v", ErrInvalidNodeTransition, err)
	}
	bundleHash, err := configbundle.HashPayload(payload)
	if err != nil {
		return ConfigRevision{}, err
	}
	bundle := configbundle.Bundle{
		NodeID:                 input.NodeID,
		RevisionNumber:         nextRevision,
		Status:                 "pending",
		BundleHash:             bundleHash,
		Signer:                 configbundle.DefaultSigner,
		RollbackTargetRevision: currentRevision,
		Payload:                payload,
	}
	signature, err := configbundle.NewDevSigner().Sign(bundle)
	if err != nil {
		return ConfigRevision{}, err
	}
	bundle.Signature = signature

	bundleJSON, err := json.Marshal(bundle)
	if err != nil {
		return ConfigRevision{}, err
	}

	var createdBy sql.NullString
	if input.CreatedByAdminID != "" {
		createdBy = sql.NullString{String: input.CreatedByAdminID, Valid: true}
	}

	revision, err := scanConfigRevision(tx.QueryRowContext(ctx, `
		INSERT INTO config_revisions (
		    node_id,
		    revision_number,
		    bundle_hash,
		    signature,
		    signer,
		    status,
		    rollback_target_revision,
		    bundle_json,
		    created_by_admin_id
		)
		VALUES ($1, $2, $3, $4, $5, 'pending', $6, $7, $8)
		RETURNING id::text, node_id::text, revision_number, bundle_hash, signature, signer, status, rollback_target_revision, bundle_json, created_at, applied_at, failed_at, rolled_back_at, error_message
	`, input.NodeID, nextRevision, bundleHash, signature, configbundle.DefaultSigner, currentRevision, bundleJSON, createdBy))
	if err != nil {
		return ConfigRevision{}, err
	}
	if err := tx.Commit(); err != nil {
		return ConfigRevision{}, err
	}
	return revision, nil
}

func listConfigSubscriptionInputs(ctx context.Context, tx *sql.Tx, nodeRegion string) ([]configrender.SubscriptionInput, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT s.id::text,
		       s.user_id::text,
		       s.plan_id::text,
		       u.status,
		       s.status,
		       COALESCE(s.preferred_region, ''),
		       p.name,
		       s.device_limit,
		       s.traffic_limit_bytes,
		       s.starts_at,
		       s.expires_at
		FROM subscriptions s
		JOIN users u ON u.id = s.user_id
		JOIN plans p ON p.id = s.plan_id
		WHERE s.status = 'active'
		  AND u.status = 'active'
		  AND s.expires_at > now()
		  AND (s.preferred_region IS NULL OR s.preferred_region = '' OR s.preferred_region = $1)
		ORDER BY s.id::text ASC
	`, nodeRegion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []configrender.SubscriptionInput
	for rows.Next() {
		var input configrender.SubscriptionInput
		var trafficLimit sql.NullInt64
		var startsAt time.Time
		var expiresAt time.Time
		if err := rows.Scan(
			&input.SubscriptionID,
			&input.UserID,
			&input.PlanID,
			&input.UserStatus,
			&input.SubscriptionStatus,
			&input.PreferredRegion,
			&input.PlanName,
			&input.DeviceLimit,
			&trafficLimit,
			&startsAt,
			&expiresAt,
		); err != nil {
			return nil, err
		}
		if trafficLimit.Valid {
			value := trafficLimit.Int64
			input.TrafficLimitBytes = &value
		}
		input.StartsAt = startsAt.UTC().Format(time.RFC3339)
		input.ExpiresAt = expiresAt.UTC().Format(time.RFC3339)
		result = append(result, input)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *nodesRepository) ListConfigRevisions(ctx context.Context, nodeID string) ([]ConfigRevision, error) {
	if _, err := r.FindByID(ctx, nodeID); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, node_id::text, revision_number, bundle_hash, signature, signer, status, rollback_target_revision, bundle_json, created_at, applied_at, failed_at, rolled_back_at, error_message
		FROM config_revisions
		WHERE node_id = $1
		ORDER BY revision_number DESC
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	revisions := []ConfigRevision{}
	for rows.Next() {
		revision, err := scanConfigRevision(rows)
		if err != nil {
			return nil, err
		}
		revisions = append(revisions, revision)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return revisions, nil
}

func (r *nodesRepository) FindConfigRevision(ctx context.Context, nodeID string, revisionID string) (ConfigRevision, error) {
	revision, err := scanConfigRevision(r.db.QueryRowContext(ctx, `
		SELECT id::text, node_id::text, revision_number, bundle_hash, signature, signer, status, rollback_target_revision, bundle_json, created_at, applied_at, failed_at, rolled_back_at, error_message
		FROM config_revisions
		WHERE node_id = $1
		  AND id = $2
	`, nodeID, revisionID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ConfigRevision{}, ErrNotFound
		}
		return ConfigRevision{}, err
	}
	return revision, nil
}

func (r *nodesRepository) FindLatestPendingConfigRevision(ctx context.Context, nodeID string, nodeToken string) (ConfigRevision, error) {
	revision, err := scanConfigRevision(r.db.QueryRowContext(ctx, `
		SELECT cr.id::text, cr.node_id::text, cr.revision_number, cr.bundle_hash, cr.signature, cr.signer, cr.status, cr.rollback_target_revision, cr.bundle_json, cr.created_at, cr.applied_at, cr.failed_at, cr.rolled_back_at, cr.error_message
		FROM config_revisions cr
		JOIN nodes n ON n.id = cr.node_id
		WHERE cr.node_id = $1
		  AND n.auth_token_hash = $2
		  AND n.registered_at IS NOT NULL
		  AND n.status != 'disabled'
		  AND cr.status = 'pending'
		ORDER BY cr.revision_number DESC
		LIMIT 1
	`, nodeID, HashNodeToken(nodeToken)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ConfigRevision{}, ErrNotFound
		}
		return ConfigRevision{}, err
	}
	return revision, nil
}

func (r *nodesRepository) ReportConfigRevision(ctx context.Context, input ReportConfigRevisionInput) (ConfigRevision, error) {
	if input.Status != "applied" && input.Status != "failed" {
		return ConfigRevision{}, ErrInvalidNodeTransition
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ConfigRevision{}, err
	}
	defer tx.Rollback()

	var reportedAt time.Time
	if input.Status == "applied" {
		reportedAt = input.AppliedAt
	} else {
		reportedAt = input.FailedAt
	}
	if reportedAt.IsZero() {
		reportedAt = input.SentAt
	}
	if reportedAt.IsZero() {
		reportedAt = time.Now().UTC()
	}
	reportedAt = reportedAt.UTC()

	var revision ConfigRevision
	if input.Status == "applied" {
		revision, err = scanConfigRevision(tx.QueryRowContext(ctx, `
			UPDATE config_revisions cr
			SET status = 'applied',
			    applied_at = $4,
			    failed_at = NULL,
			    error_message = NULL,
			    updated_at = $4
			FROM nodes n
			WHERE cr.node_id = n.id
			  AND cr.node_id = $1
			  AND cr.id = $2
			  AND n.auth_token_hash = $3
			  AND n.registered_at IS NOT NULL
			  AND n.status != 'disabled'
			  AND cr.status = 'pending'
			RETURNING cr.id::text, cr.node_id::text, cr.revision_number, cr.bundle_hash, cr.signature, cr.signer, cr.status, cr.rollback_target_revision, cr.bundle_json, cr.created_at, cr.applied_at, cr.failed_at, cr.rolled_back_at, cr.error_message
		`, input.NodeID, input.RevisionID, HashNodeToken(input.NodeToken), reportedAt))
	} else {
		revision, err = scanConfigRevision(tx.QueryRowContext(ctx, `
			UPDATE config_revisions cr
			SET status = 'failed',
			    failed_at = $4,
			    error_message = NULLIF($5, ''),
			    updated_at = $4
			FROM nodes n
			WHERE cr.node_id = n.id
			  AND cr.node_id = $1
			  AND cr.id = $2
			  AND n.auth_token_hash = $3
			  AND n.registered_at IS NOT NULL
			  AND n.status != 'disabled'
			  AND cr.status = 'pending'
			RETURNING cr.id::text, cr.node_id::text, cr.revision_number, cr.bundle_hash, cr.signature, cr.signer, cr.status, cr.rollback_target_revision, cr.bundle_json, cr.created_at, cr.applied_at, cr.failed_at, cr.rolled_back_at, cr.error_message
		`, input.NodeID, input.RevisionID, HashNodeToken(input.NodeToken), reportedAt, input.ErrorMessage))
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ConfigRevision{}, ErrNotFound
		}
		return ConfigRevision{}, err
	}

	if input.Status == "applied" {
		if _, err := tx.ExecContext(ctx, `
			UPDATE nodes
			SET active_revision = $2,
			    updated_at = $3
			WHERE id = $1
			  AND auth_token_hash = $4
			  AND registered_at IS NOT NULL
			  AND status != 'disabled'
		`, input.NodeID, revision.RevisionNumber, reportedAt, HashNodeToken(input.NodeToken)); err != nil {
			return ConfigRevision{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return ConfigRevision{}, err
	}
	return revision, nil
}

func newBootstrapToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "lnkbt_" + hex.EncodeToString(raw), nil
}

func newNodeToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func HashBootstrapToken(token string) string {
	return hashString(token)
}

func HashNodeToken(token string) string {
	return hashString(token)
}

func isValidNodeStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "pending", "active", "unhealthy", "drained", "disabled":
		return true
	default:
		return false
	}
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanNode(row rowScanner) (Node, error) {
	var node Node
	var lastHealthAt sql.NullTime
	var lastSeenAt sql.NullTime
	var registeredAt sql.NullTime
	err := row.Scan(
		&node.ID,
		&node.Name,
		&node.Region,
		&node.CountryCode,
		&node.Hostname,
		&node.Status,
		&node.DrainState,
		&node.AgentVersion,
		&node.XrayVersion,
		&node.ActiveRevision,
		&lastHealthAt,
		&lastSeenAt,
		&registeredAt,
		&node.UpdatedAt,
	)
	if err != nil {
		return Node{}, err
	}
	if lastHealthAt.Valid {
		node.LastHealthAt = &lastHealthAt.Time
	}
	if lastSeenAt.Valid {
		node.LastSeenAt = &lastSeenAt.Time
	}
	if registeredAt.Valid {
		node.RegisteredAt = &registeredAt.Time
	}
	return node, nil
}

func scanConfigRevision(row rowScanner) (ConfigRevision, error) {
	var revision ConfigRevision
	var rollbackTarget sql.NullInt64
	var bundleJSON []byte
	var appliedAt sql.NullTime
	var failedAt sql.NullTime
	var rolledBackAt sql.NullTime
	var errorMessage sql.NullString
	err := row.Scan(
		&revision.ID,
		&revision.NodeID,
		&revision.RevisionNumber,
		&revision.BundleHash,
		&revision.Signature,
		&revision.Signer,
		&revision.Status,
		&rollbackTarget,
		&bundleJSON,
		&revision.CreatedAt,
		&appliedAt,
		&failedAt,
		&rolledBackAt,
		&errorMessage,
	)
	if err != nil {
		return ConfigRevision{}, err
	}
	if rollbackTarget.Valid {
		revision.RollbackTargetRevision = int(rollbackTarget.Int64)
	}
	if appliedAt.Valid {
		revision.AppliedAt = &appliedAt.Time
	}
	if failedAt.Valid {
		revision.FailedAt = &failedAt.Time
	}
	if rolledBackAt.Valid {
		revision.RolledBackAt = &rolledBackAt.Time
	}
	if errorMessage.Valid {
		revision.ErrorMessage = errorMessage.String
	}
	if len(bundleJSON) > 0 {
		var storedBundle configbundle.Bundle
		if err := json.Unmarshal(bundleJSON, &storedBundle); err != nil {
			return ConfigRevision{}, err
		}
		revision.Bundle = storedBundle.Payload
	}
	return revision, nil
}

type nodeTransition struct {
	Status     string
	DrainState string
}

func (r *nodesRepository) transition(ctx context.Context, id string, decide func(Node) (nodeTransition, error)) (Node, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Node{}, err
	}
	defer tx.Rollback()

	node, err := scanNode(tx.QueryRowContext(ctx, `
		SELECT id::text, name, region, country_code, hostname, status, drain_state, agent_version, xray_version, active_revision, last_health_at, last_seen_at, registered_at, updated_at
		FROM nodes
		WHERE id = $1
		FOR UPDATE
	`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Node{}, ErrNotFound
		}
		return Node{}, err
	}

	next, err := decide(node)
	if err != nil {
		return Node{}, err
	}
	if next.Status == "" {
		next.Status = node.Status
	}
	if next.DrainState == "" {
		next.DrainState = node.DrainState
	}

	updated, err := scanNode(tx.QueryRowContext(ctx, `
		UPDATE nodes
		SET status = $2,
		    drain_state = $3,
		    updated_at = now()
		WHERE id = $1
		RETURNING id::text, name, region, country_code, hostname, status, drain_state, agent_version, xray_version, active_revision, last_health_at, last_seen_at, registered_at, updated_at
	`, id, next.Status, next.DrainState))
	if err != nil {
		return Node{}, err
	}
	if err := tx.Commit(); err != nil {
		return Node{}, err
	}
	return updated, nil
}
