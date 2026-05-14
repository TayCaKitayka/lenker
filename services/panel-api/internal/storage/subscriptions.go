package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"
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
	Update(ctx context.Context, id string, input UpdateSubscriptionInput) (Subscription, error)
	Renew(ctx context.Context, id string, extendDays int) (Subscription, error)
}

type subscriptionsRepository struct {
	db *sql.DB
}

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
	err := r.db.QueryRowContext(ctx, `
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
		SELECT u.id, p.id, 'active', $3, $3 + (p.duration_days * INTERVAL '1 day'),
		       p.traffic_limit_bytes, p.device_limit, $4
		FROM users u
		JOIN plans p ON p.id = $2
		            AND p.status = 'active'
		WHERE u.id = $1
		RETURNING id::text, user_id::text, plan_id::text, status, starts_at, expires_at,
		          traffic_limit_bytes, traffic_used_bytes, device_limit, preferred_region
	`, input.UserID, input.PlanID, now, input.PreferredRegion).Scan(
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
