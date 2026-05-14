# panel-api

`panel-api` is the Go service for the Lenker provider control plane.

Planned responsibilities for `MVP v0.1`:

- admin authentication
- RBAC
- users, plans, subscriptions, devices
- node registration and lifecycle
- protocol profile management
- config revision and deployment coordination
- API tokens and webhooks
- audit logging

Current foundation:

- application entrypoint at `cmd/panel-api`
- environment-based config loading
- structured JSON logging through Go `slog`
- HTTP server with graceful shutdown
- router and response envelope helpers
- health endpoint
- PostgreSQL storage bootstrap through `database/sql` and the `pgx` stdlib driver
- repository interfaces and initial query implementations for admins, users, plans, and subscriptions
- minimal admin login service with password hash verification, inactive admin check, and session creation
- admin session validation middleware using `Authorization: Bearer <session_token>`
- bcrypt password verification for admin accounts
- minimal admin CRUD slice for users, plans, and subscriptions
- RBAC and audit package-level contracts without a full permission engine
- package placeholders for the MVP control-plane domains

Local run:

```sh
go run ./cmd/panel-api
```

Configuration:

- `LENKER_APP_ENV`
- `LENKER_HTTP_ADDR`
- `LENKER_LOG_LEVEL`
- `LENKER_SHUTDOWN_TIMEOUT_SECONDS`
- `LENKER_DATABASE_URL`
- `LENKER_DATABASE_PING`

Implemented foundation routes:

- `GET /healthz`
- `POST /api/v1/auth/admin/login`
- `GET /api/v1/users`
- `POST /api/v1/users`
- `GET /api/v1/users/{id}`
- `PATCH /api/v1/users/{id}`
- `POST /api/v1/users/{id}/suspend`
- `POST /api/v1/users/{id}/activate`
- `GET /api/v1/plans`
- `POST /api/v1/plans`
- `GET /api/v1/plans/{id}`
- `PATCH /api/v1/plans/{id}`
- `POST /api/v1/plans/{id}/archive`
- `GET /api/v1/subscriptions`
- `POST /api/v1/subscriptions`
- `GET /api/v1/subscriptions/{id}`
- `PATCH /api/v1/subscriptions/{id}`
- `POST /api/v1/subscriptions/{id}/renew`

Admin-only routes:

- all `/api/v1/users*` routes
- all `/api/v1/plans*` routes
- all `/api/v1/subscriptions*` routes

Use the token returned by admin login:

```http
Authorization: Bearer <session_token>
```

Not included here yet:

- delete operations
- advanced business rules
- full production authentication policy
- logout and full session lifecycle management
- refresh tokens
- 2FA
- RBAC permission engine
- audit persistence
- devices, key rotation, and export flows
- billing
- marketplace
- VPN or Xray logic

Conservative storage note:

The service opens a PostgreSQL handle at startup, but `LENKER_DATABASE_PING` defaults to `false`. This keeps the local HTTP skeleton runnable without a database while still allowing deployments and tests to opt into startup connectivity checks.

Conservative auth note:

Admin password hashes must use bcrypt. This keeps the first auth path stronger than the earlier foundation placeholder without adding a larger auth platform, 2FA, refresh tokens, OAuth, or phone auth.

Migration workflow:

```sh
make migrate-up
make migrate-down
VERSION=1 make migrate-force
```
