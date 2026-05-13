package storage

import (
	"context"
	"database/sql"
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

type SubscriptionsRepository interface {
	List(ctx context.Context) ([]Subscription, error)
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
