package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/configrender"
)

type Subscription struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	PlanID            string    `json:"plan_id"`
	Status            string    `json:"status"`
	StartsAt          time.Time `json:"starts_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	TrafficLimitBytes *int64    `json:"traffic_limit_bytes"`
	TrafficUsedBytes  int64     `json:"traffic_used_bytes"`
	DeviceLimit       int       `json:"device_limit"`
	PreferredRegion   *string   `json:"preferred_region"`
}

type SubscriptionAccess struct {
	ExportKind     string                     `json:"export_kind"`
	SubscriptionID string                     `json:"subscription_id"`
	UserID         string                     `json:"user_id"`
	UserLabel      string                     `json:"user_label"`
	PlanID         string                     `json:"plan_id"`
	PlanName       string                     `json:"plan_name"`
	Status         string                     `json:"status"`
	Protocol       string                     `json:"protocol"`
	ProtocolPath   string                     `json:"protocol_path"`
	Node           SubscriptionAccessNode     `json:"node"`
	Endpoint       SubscriptionAccessEndpoint `json:"endpoint"`
	Client         SubscriptionAccessClient   `json:"client"`
	DisplayName    string                     `json:"display_name"`
	URI            string                     `json:"uri"`
}

type SubscriptionAccessNode struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Region         string `json:"region"`
	CountryCode    string `json:"country_code"`
	Hostname       string `json:"hostname"`
	Status         string `json:"status"`
	DrainState     string `json:"drain_state"`
	ActiveRevision int    `json:"active_revision"`
}

type SubscriptionAccessEndpoint struct {
	Address     string `json:"address"`
	Port        int    `json:"port"`
	Network     string `json:"network"`
	Security    string `json:"security"`
	SNI         string `json:"sni"`
	PublicKey   string `json:"public_key"`
	ShortID     string `json:"short_id"`
	Fingerprint string `json:"fingerprint"`
	SpiderX     string `json:"spider_x"`
}

type SubscriptionAccessClient struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Flow   string `json:"flow"`
	Level  int    `json:"level"`
	PlanID string `json:"plan_id"`
}

type CreateSubscriptionInput struct {
	UserID          string
	PlanID          string
	PreferredRegion *string
}

type UpdateSubscriptionInput struct {
	Status               *string
	TrafficLimitBytes    *int64
	ClearTrafficLimit    bool
	DeviceLimit          *int
	PreferredRegion      *string
	ClearPreferredRegion bool
}

type SubscriptionsRepository interface {
	List(ctx context.Context) ([]Subscription, error)
	Create(ctx context.Context, input CreateSubscriptionInput) (Subscription, error)
	FindByID(ctx context.Context, id string) (Subscription, error)
	Access(ctx context.Context, id string) (SubscriptionAccess, error)
	Update(ctx context.Context, id string, input UpdateSubscriptionInput) (Subscription, error)
	Renew(ctx context.Context, id string, extendDays int) (Subscription, error)
}

var ErrSubscriptionAccessUnavailable = errors.New("subscription access unavailable")

type subscriptionsRepository struct {
	db *sql.DB
}

const createSubscriptionSQL = `
	INSERT INTO subscriptions (
		user_id,
		plan_id,
		status,
		starts_at,
		expires_at,
		traffic_limit_bytes,
		device_limit,
		preferred_region
	)
	SELECT u.id, p.id, 'active', $3::timestamptz, $3::timestamptz + (p.duration_days * INTERVAL '1 day'),
	       p.traffic_limit_bytes, p.device_limit, $4
	FROM users u
	JOIN plans p ON p.id = $2
	            AND p.status = 'active'
	WHERE u.id = $1
	RETURNING id::text, user_id::text, plan_id::text, status, starts_at, expires_at,
	          traffic_limit_bytes, traffic_used_bytes, device_limit, preferred_region
`

func NewSubscriptionsRepository(db *sql.DB) SubscriptionsRepository {
	return &subscriptionsRepository{db: db}
}

func (r *subscriptionsRepository) List(ctx context.Context) ([]Subscription, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, user_id::text, plan_id::text, status, starts_at, expires_at,
		       traffic_limit_bytes, traffic_used_bytes, device_limit, preferred_region
		FROM subscriptions
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Subscription
	for rows.Next() {
		var subscription Subscription
		if err := rows.Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.PlanID,
			&subscription.Status,
			&subscription.StartsAt,
			&subscription.ExpiresAt,
			&subscription.TrafficLimitBytes,
			&subscription.TrafficUsedBytes,
			&subscription.DeviceLimit,
			&subscription.PreferredRegion,
		); err != nil {
			return nil, err
		}
		result = append(result, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *subscriptionsRepository) Create(ctx context.Context, input CreateSubscriptionInput) (Subscription, error) {
	var subscription Subscription
	now := time.Now().UTC()
	err := r.db.QueryRowContext(ctx, createSubscriptionSQL, input.UserID, input.PlanID, now, input.PreferredRegion).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.PlanID,
		&subscription.Status,
		&subscription.StartsAt,
		&subscription.ExpiresAt,
		&subscription.TrafficLimitBytes,
		&subscription.TrafficUsedBytes,
		&subscription.DeviceLimit,
		&subscription.PreferredRegion,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Subscription{}, ErrNotFound
		}
		return Subscription{}, err
	}
	return subscription, nil
}

func (r *subscriptionsRepository) FindByID(ctx context.Context, id string) (Subscription, error) {
	var subscription Subscription
	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, user_id::text, plan_id::text, status, starts_at, expires_at,
		       traffic_limit_bytes, traffic_used_bytes, device_limit, preferred_region
		FROM subscriptions
		WHERE id = $1
	`, id).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.PlanID,
		&subscription.Status,
		&subscription.StartsAt,
		&subscription.ExpiresAt,
		&subscription.TrafficLimitBytes,
		&subscription.TrafficUsedBytes,
		&subscription.DeviceLimit,
		&subscription.PreferredRegion,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Subscription{}, ErrNotFound
		}
		return Subscription{}, err
	}
	return subscription, nil
}

func (r *subscriptionsRepository) Access(ctx context.Context, id string) (SubscriptionAccess, error) {
	var row struct {
		subscription Subscription
		userEmail    string
		displayName  string
		userStatus   string
		planName     string
		nodeID       sql.NullString
		nodeName     sql.NullString
		nodeRegion   sql.NullString
		countryCode  sql.NullString
		hostname     sql.NullString
		nodeStatus   sql.NullString
		drainState   sql.NullString
		activeRev    sql.NullInt64
	}
	err := r.db.QueryRowContext(ctx, `
		SELECT s.id::text,
		       s.user_id::text,
		       s.plan_id::text,
		       s.status,
		       s.starts_at,
		       s.expires_at,
		       s.traffic_limit_bytes,
		       s.traffic_used_bytes,
		       s.device_limit,
		       s.preferred_region,
		       u.email,
		       u.display_name,
		       u.status,
		       p.name,
		       n.id::text,
		       n.name,
		       n.region,
		       n.country_code,
		       n.hostname,
		       n.status,
		       n.drain_state,
		       n.active_revision
		FROM subscriptions s
		JOIN users u ON u.id = s.user_id
		JOIN plans p ON p.id = s.plan_id
		LEFT JOIN LATERAL (
			SELECT id, name, region, country_code, hostname, status, drain_state, active_revision
			FROM nodes
			WHERE status = 'active'
			  AND drain_state = 'active'
			  AND hostname <> ''
			  AND (s.preferred_region IS NULL OR s.preferred_region = '' OR region = s.preferred_region)
			ORDER BY CASE WHEN s.preferred_region IS NOT NULL AND s.preferred_region <> '' AND region = s.preferred_region THEN 0 ELSE 1 END,
			         region ASC,
			         name ASC,
			         id::text ASC
			LIMIT 1
		) n ON true
		WHERE s.id = $1
	`, id).Scan(
		&row.subscription.ID,
		&row.subscription.UserID,
		&row.subscription.PlanID,
		&row.subscription.Status,
		&row.subscription.StartsAt,
		&row.subscription.ExpiresAt,
		&row.subscription.TrafficLimitBytes,
		&row.subscription.TrafficUsedBytes,
		&row.subscription.DeviceLimit,
		&row.subscription.PreferredRegion,
		&row.userEmail,
		&row.displayName,
		&row.userStatus,
		&row.planName,
		&row.nodeID,
		&row.nodeName,
		&row.nodeRegion,
		&row.countryCode,
		&row.hostname,
		&row.nodeStatus,
		&row.drainState,
		&row.activeRev,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SubscriptionAccess{}, ErrNotFound
		}
		return SubscriptionAccess{}, err
	}
	if row.subscription.Status != "active" || row.userStatus != "active" || !row.subscription.ExpiresAt.After(time.Now().UTC()) {
		return SubscriptionAccess{}, ErrSubscriptionAccessUnavailable
	}
	if !row.nodeID.Valid || !row.hostname.Valid {
		return SubscriptionAccess{}, ErrSubscriptionAccessUnavailable
	}

	preferredRegion := ""
	if row.subscription.PreferredRegion != nil {
		preferredRegion = *row.subscription.PreferredRegion
	}
	accessEntry := configrender.BuildAccessEntry(configrender.SubscriptionInput{
		SubscriptionID:     row.subscription.ID,
		UserID:             row.subscription.UserID,
		PlanID:             row.subscription.PlanID,
		UserStatus:         row.userStatus,
		SubscriptionStatus: row.subscription.Status,
		PreferredRegion:    preferredRegion,
		PlanName:           row.planName,
		DeviceLimit:        row.subscription.DeviceLimit,
		TrafficLimitBytes:  row.subscription.TrafficLimitBytes,
		StartsAt:           row.subscription.StartsAt.UTC().Format(time.RFC3339),
		ExpiresAt:          row.subscription.ExpiresAt.UTC().Format(time.RFC3339),
	})
	node := SubscriptionAccessNode{
		ID:             row.nodeID.String,
		Name:           row.nodeName.String,
		Region:         row.nodeRegion.String,
		CountryCode:    row.countryCode.String,
		Hostname:       row.hostname.String,
		Status:         row.nodeStatus.String,
		DrainState:     row.drainState.String,
		ActiveRevision: int(row.activeRev.Int64),
	}
	endpoint := SubscriptionAccessEndpoint{
		Address:     node.Hostname,
		Port:        configrender.DefaultVLESSPort,
		Network:     "tcp",
		Security:    "reality",
		SNI:         configrender.DefaultRealitySNI,
		PublicKey:   configrender.DefaultRealityPublic,
		ShortID:     configrender.DefaultRealityShortID,
		Fingerprint: configrender.DefaultFingerprint,
		SpiderX:     configrender.DefaultSpiderX,
	}
	userLabel := row.userEmail
	if strings.TrimSpace(row.displayName) != "" {
		userLabel = row.displayName
	}
	displayName := fmt.Sprintf("Lenker %s %s", node.Region, row.planName)
	client := SubscriptionAccessClient{
		ID:     accessEntry.VLESSClientID,
		Email:  accessEntry.Email,
		Flow:   accessEntry.Flow,
		Level:  0,
		PlanID: accessEntry.PlanID,
	}
	return SubscriptionAccess{
		ExportKind:     "subscription_access.v1alpha1",
		SubscriptionID: row.subscription.ID,
		UserID:         row.subscription.UserID,
		UserLabel:      userLabel,
		PlanID:         row.subscription.PlanID,
		PlanName:       row.planName,
		Status:         row.subscription.Status,
		Protocol:       configrender.ProtocolVLESS,
		ProtocolPath:   configrender.ProtocolVLESS,
		Node:           node,
		Endpoint:       endpoint,
		Client:         client,
		DisplayName:    displayName,
		URI:            buildVLESSRealityURI(endpoint, client, displayName),
	}, nil
}

func buildVLESSRealityURI(endpoint SubscriptionAccessEndpoint, client SubscriptionAccessClient, displayName string) string {
	values := url.Values{}
	values.Set("encryption", "none")
	values.Set("flow", client.Flow)
	values.Set("fp", endpoint.Fingerprint)
	values.Set("pbk", endpoint.PublicKey)
	values.Set("security", endpoint.Security)
	values.Set("sid", endpoint.ShortID)
	values.Set("sni", endpoint.SNI)
	values.Set("spx", endpoint.SpiderX)
	values.Set("type", endpoint.Network)

	uri := url.URL{
		Scheme:   "vless",
		User:     url.User(client.ID),
		Host:     fmt.Sprintf("%s:%d", endpoint.Address, endpoint.Port),
		RawQuery: values.Encode(),
		Fragment: displayName,
	}
	return uri.String()
}

func (r *subscriptionsRepository) Update(ctx context.Context, id string, input UpdateSubscriptionInput) (Subscription, error) {
	current, err := r.FindByID(ctx, id)
	if err != nil {
		return Subscription{}, err
	}
	if input.Status != nil {
		current.Status = *input.Status
	}
	if input.ClearTrafficLimit {
		current.TrafficLimitBytes = nil
	} else if input.TrafficLimitBytes != nil {
		current.TrafficLimitBytes = input.TrafficLimitBytes
	}
	if input.DeviceLimit != nil {
		current.DeviceLimit = *input.DeviceLimit
	}
	if input.ClearPreferredRegion {
		current.PreferredRegion = nil
	} else if input.PreferredRegion != nil {
		current.PreferredRegion = input.PreferredRegion
	}

	var subscription Subscription
	err = r.db.QueryRowContext(ctx, `
		UPDATE subscriptions
		SET status = $2,
		    traffic_limit_bytes = $3,
		    device_limit = $4,
		    preferred_region = $5,
		    updated_at = now()
		WHERE id = $1
		RETURNING id::text, user_id::text, plan_id::text, status, starts_at, expires_at,
		          traffic_limit_bytes, traffic_used_bytes, device_limit, preferred_region
	`, id, current.Status, current.TrafficLimitBytes, current.DeviceLimit, current.PreferredRegion).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.PlanID,
		&subscription.Status,
		&subscription.StartsAt,
		&subscription.ExpiresAt,
		&subscription.TrafficLimitBytes,
		&subscription.TrafficUsedBytes,
		&subscription.DeviceLimit,
		&subscription.PreferredRegion,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Subscription{}, ErrNotFound
		}
		return Subscription{}, err
	}
	return subscription, nil
}

func (r *subscriptionsRepository) Renew(ctx context.Context, id string, extendDays int) (Subscription, error) {
	var subscription Subscription
	err := r.db.QueryRowContext(ctx, `
		UPDATE subscriptions
		SET status = 'active',
		    expires_at = GREATEST(expires_at, now()) + ($2 * INTERVAL '1 day'),
		    updated_at = now()
		WHERE id = $1
		RETURNING id::text, user_id::text, plan_id::text, status, starts_at, expires_at,
		          traffic_limit_bytes, traffic_used_bytes, device_limit, preferred_region
	`, id, extendDays).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.PlanID,
		&subscription.Status,
		&subscription.StartsAt,
		&subscription.ExpiresAt,
		&subscription.TrafficLimitBytes,
		&subscription.TrafficUsedBytes,
		&subscription.DeviceLimit,
		&subscription.PreferredRegion,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Subscription{}, ErrNotFound
		}
		return Subscription{}, err
	}
	return subscription, nil
}
