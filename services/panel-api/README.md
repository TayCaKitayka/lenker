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
- `GET /api/v1/plans`
- `GET /api/v1/subscriptions`

Not included here yet:

- business logic
- create/update/delete operations
- full production authentication policy
- full session lifecycle management
- RBAC permission engine
- audit persistence
- billing
- marketplace
- VPN or Xray logic

Conservative storage note:

The service opens a PostgreSQL handle at startup, but `LENKER_DATABASE_PING` defaults to `false`. This keeps the local HTTP skeleton runnable without a database while still allowing deployments and tests to opt into startup connectivity checks.

Conservative auth note:

The current verifier supports `sha256$<hex>` and `sha256:<hex>` password hashes only to make the first admin login path testable without introducing a larger auth platform. Replace this with bcrypt or Argon2id before real provider usage.
