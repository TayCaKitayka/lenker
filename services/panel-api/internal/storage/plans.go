package storage

import (
	"context"
	"database/sql"
	"errors"
)

type Plan struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	DurationDays      int    `json:"duration_days"`
	TrafficLimitBytes *int64 `json:"traffic_limit_bytes"`
	DeviceLimit       int    `json:"device_limit"`
	Status            string `json:"status"`
}

type CreatePlanInput struct {
	Name              string
	DurationDays      int
	TrafficLimitBytes *int64
	DeviceLimit       int
}

type UpdatePlanInput struct {
	Name              *string
	DurationDays      *int
	TrafficLimitBytes *int64
	ClearTrafficLimit bool
	DeviceLimit       *int
}

type PlansRepository interface {
	List(ctx context.Context) ([]Plan, error)
	Create(ctx context.Context, input CreatePlanInput) (Plan, error)
	FindByID(ctx context.Context, id string) (Plan, error)
	Update(ctx context.Context, id string, input UpdatePlanInput) (Plan, error)
	Archive(ctx context.Context, id string) (Plan, error)
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

func (r *plansRepository) Create(ctx context.Context, input CreatePlanInput) (Plan, error) {
	var plan Plan
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO plans (name, duration_days, traffic_limit_bytes, device_limit)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, name, duration_days, traffic_limit_bytes, device_limit, status
	`, input.Name, input.DurationDays, input.TrafficLimitBytes, input.DeviceLimit).Scan(
		&plan.ID,
		&plan.Name,
		&plan.DurationDays,
		&plan.TrafficLimitBytes,
		&plan.DeviceLimit,
		&plan.Status,
	)
	if err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (r *plansRepository) FindByID(ctx context.Context, id string) (Plan, error) {
	var plan Plan
	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, name, duration_days, traffic_limit_bytes, device_limit, status
		FROM plans
		WHERE id = $1
	`, id).Scan(&plan.ID, &plan.Name, &plan.DurationDays, &plan.TrafficLimitBytes, &plan.DeviceLimit, &plan.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Plan{}, ErrNotFound
		}
		return Plan{}, err
	}
	return plan, nil
}

func (r *plansRepository) Update(ctx context.Context, id string, input UpdatePlanInput) (Plan, error) {
	current, err := r.FindByID(ctx, id)
	if err != nil {
		return Plan{}, err
	}
	if input.Name != nil {
		current.Name = *input.Name
	}
	if input.DurationDays != nil {
		current.DurationDays = *input.DurationDays
	}
	if input.ClearTrafficLimit {
		current.TrafficLimitBytes = nil
	} else if input.TrafficLimitBytes != nil {
		current.TrafficLimitBytes = input.TrafficLimitBytes
	}
	if input.DeviceLimit != nil {
		current.DeviceLimit = *input.DeviceLimit
	}

	var plan Plan
	err = r.db.QueryRowContext(ctx, `
		UPDATE plans
		SET name = $2, duration_days = $3, traffic_limit_bytes = $4, device_limit = $5, updated_at = now()
		WHERE id = $1
		RETURNING id::text, name, duration_days, traffic_limit_bytes, device_limit, status
	`, id, current.Name, current.DurationDays, current.TrafficLimitBytes, current.DeviceLimit).Scan(
		&plan.ID,
		&plan.Name,
		&plan.DurationDays,
		&plan.TrafficLimitBytes,
		&plan.DeviceLimit,
		&plan.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Plan{}, ErrNotFound
		}
		return Plan{}, err
	}
	return plan, nil
}

func (r *plansRepository) Archive(ctx context.Context, id string) (Plan, error) {
	var plan Plan
	err := r.db.QueryRowContext(ctx, `
		UPDATE plans
		SET status = 'archived', updated_at = now()
		WHERE id = $1
		RETURNING id::text, name, duration_days, traffic_limit_bytes, device_limit, status
	`, id).Scan(&plan.ID, &plan.Name, &plan.DurationDays, &plan.TrafficLimitBytes, &plan.DeviceLimit, &plan.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Plan{}, ErrNotFound
		}
		return Plan{}, err
	}
	return plan, nil
}
