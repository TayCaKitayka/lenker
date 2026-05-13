package auth

import (
	"context"

	"github.com/lenker/lenker/services/panel-api/internal/admins"
)

type adminContextKey struct {
}

func WithAdmin(ctx context.Context, admin admins.Admin) context.Context {
	return context.WithValue(ctx, adminContextKey{}, admin)
}

func AdminFromContext(ctx context.Context) (admins.Admin, bool) {
	admin, ok := ctx.Value(adminContextKey{}).(admins.Admin)
	return admin, ok
}
