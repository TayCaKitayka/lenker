package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
)

type Node struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Hostname       string     `json:"hostname"`
	Status         string     `json:"status"`
	DrainState     string     `json:"drain_state"`
	AgentVersion   string     `json:"agent_version"`
	XrayVersion    string     `json:"xray_version"`
	ActiveRevision int        `json:"active_revision"`
	LastHealthAt   *time.Time `json:"last_health_at"`
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
	Register(ctx context.Context, input RegisterNodeInput) (RegisterNodeResult, error)
	RecordHeartbeat(ctx context.Context, input HeartbeatInput) (Node, error)
}

type nodesRepository struct {
	db *sql.DB
}

func NewNodesRepository(db *sql.DB) NodesRepository {
	return &nodesRepository{db: db}
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

	var node Node
	var lastHealthAt sql.NullTime
	if input.NodeID == "" {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO nodes (hostname, agent_version, auth_token_hash)
			VALUES ($1, $2, $3)
			RETURNING id::text, name, hostname, status, drain_state, agent_version, xray_version, active_revision, last_health_at
		`, input.Hostname, input.AgentVersion, HashNodeToken(nodeToken)).Scan(
			&node.ID,
			&node.Name,
			&node.Hostname,
			&node.Status,
			&node.DrainState,
			&node.AgentVersion,
			&node.XrayVersion,
			&node.ActiveRevision,
			&lastHealthAt,
		)
	} else {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO nodes (id, hostname, agent_version, auth_token_hash)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE
			SET hostname = EXCLUDED.hostname,
			    agent_version = EXCLUDED.agent_version,
			    auth_token_hash = EXCLUDED.auth_token_hash,
			    status = 'registered',
			    updated_at = now()
			RETURNING id::text, name, hostname, status, drain_state, agent_version, xray_version, active_revision, last_health_at
		`, input.NodeID, input.Hostname, input.AgentVersion, HashNodeToken(nodeToken)).Scan(
			&node.ID,
			&node.Name,
			&node.Hostname,
			&node.Status,
			&node.DrainState,
			&node.AgentVersion,
			&node.XrayVersion,
			&node.ActiveRevision,
			&lastHealthAt,
		)
	}
	if err != nil {
		return RegisterNodeResult{}, err
	}

	if lastHealthAt.Valid {
		node.LastHealthAt = &lastHealthAt.Time
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO node_registrations (node_id, bootstrap_token_hash, registration_status)
		VALUES ($1, $2, 'completed')
	`, node.ID, hashString(input.BootstrapToken))
	if err != nil {
		return RegisterNodeResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return RegisterNodeResult{}, err
	}

	return RegisterNodeResult{Node: node, NodeToken: nodeToken}, nil
}

func (r *nodesRepository) RecordHeartbeat(ctx context.Context, input HeartbeatInput) (Node, error) {
	var node Node
	var lastHealthAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		UPDATE nodes
		SET status = $3,
		    agent_version = $4,
		    active_revision = $5,
		    last_health_at = $6,
		    updated_at = now()
		WHERE id = $1
		  AND auth_token_hash = $2
		RETURNING id::text, name, hostname, status, drain_state, agent_version, xray_version, active_revision, last_health_at
	`, input.NodeID, HashNodeToken(input.NodeToken), input.Status, input.AgentVersion, input.ActiveRevision, input.SentAt).Scan(
		&node.ID,
		&node.Name,
		&node.Hostname,
		&node.Status,
		&node.DrainState,
		&node.AgentVersion,
		&node.XrayVersion,
		&node.ActiveRevision,
		&lastHealthAt,
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
	return node, nil
}

func newNodeToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func HashNodeToken(token string) string {
	return hashString(token)
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
