package storage

import (
	"context"
	"database/sql"
)

type Plan struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	DurationDays      int    `json:"duration_days"`
	TrafficLimitBytes *int64 `json:"traffic_limit_bytes"`
	DeviceLimit       int    `json:"device_limit"`
	Status            string `json:"status"`
}

type PlansRepository interface {
	List(ctx context.Context) ([]Plan, error)
}

type plansRepository struct {
	db *sql.DB
}

func NewPlansRepository(db *sql.DB) PlansRepository {
	return &plansRepository{db: db}
}

func (r *plansRepository) List(ctx context.Context) ([]Plan, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, name, duration_days, traffic_limit_bytes, device_limit, status
		FROM plans
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Plan
	for rows.Next() {
		var plan Plan
		if err := rows.Scan(&plan.ID, &plan.Name, &plan.DurationDays, &plan.TrafficLimitBytes, &plan.DeviceLimit, &plan.Status); err != nil {
			return nil, err
		}
		result = append(result, plan)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
