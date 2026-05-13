package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lenker/lenker/services/panel-api/internal/admins"
)

type adminsRepository struct {
	db *sql.DB
}

func NewAdminsRepository(db *sql.DB) admins.Repository {
	return &adminsRepository{db: db}
}

func (r *adminsRepository) FindByEmail(ctx context.Context, email string) (admins.Admin, error) {
	var admin admins.Admin
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, email, password_hash, status, two_factor_enabled, created_at, updated_at, last_login_at
		FROM admins
		WHERE email = $1
	`, email).Scan(
		&admin.ID,
		&admin.Email,
		&admin.PasswordHash,
		&admin.Status,
		&admin.TwoFactorEnabled,
		&admin.CreatedAt,
		&admin.UpdatedAt,
		&lastLoginAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return admins.Admin{}, admins.ErrNotFound
		}
		return admins.Admin{}, err
	}

	if lastLoginAt.Valid {
		admin.LastLoginAt = lastLoginAt.Time
	}

	return admin, nil
}

func (r *adminsRepository) CreateSession(ctx context.Context, adminID string, tokenHash string, expiresAt time.Time) (admins.Session, error) {
	var session admins.Session

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO admin_sessions (admin_id, session_token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id::text, admin_id::text, expires_at, created_at
	`, adminID, tokenHash, expiresAt).Scan(
		&session.ID,
		&session.AdminID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		return admins.Session{}, err
	}

	return session, nil
}
