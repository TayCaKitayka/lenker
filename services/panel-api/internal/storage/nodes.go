package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidBootstrapToken = errors.New("invalid bootstrap token")
	ErrExpiredBootstrapToken = errors.New("expired bootstrap token")
	ErrBootstrapTokenUsed    = errors.New("bootstrap token already used")
	ErrInvalidNodeStatus     = errors.New("invalid node status")
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

type NodesRepository interface {
	CreateBootstrapToken(ctx context.Context, input CreateBootstrapTokenInput) (BootstrapToken, error)
	Register(ctx context.Context, input RegisterNodeInput) (RegisterNodeResult, error)
	RecordHeartbeat(ctx context.Context, input HeartbeatInput) (Node, error)
}

type nodesRepository struct {
	db *sql.DB
}

func NewNodesRepository(db *sql.DB) NodesRepository {
	return &nodesRepository{db: db}
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
