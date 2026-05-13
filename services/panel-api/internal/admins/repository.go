package admins

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("admin not found")

type Repository interface {
	FindByEmail(ctx context.Context, email string) (Admin, error)
	FindByActiveSessionTokenHash(ctx context.Context, tokenHash string, now time.Time) (Admin, error)
	CreateSession(ctx context.Context, adminID string, tokenHash string, expiresAt time.Time) (Session, error)
}
