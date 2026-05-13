package rbac

import "context"

type Permission string

const (
	PermissionReadUsers         Permission = "users:read"
	PermissionReadPlans         Permission = "plans:read"
	PermissionReadSubscriptions Permission = "subscriptions:read"
)

type Authorizer interface {
	Authorize(ctx context.Context, actorID string, permission Permission) error
}

type NoopAuthorizer struct {
}

func (NoopAuthorizer) Authorize(ctx context.Context, actorID string, permission Permission) error {
	return nil
}
