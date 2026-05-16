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
- admin-created one-time node bootstrap tokens
- node registration with token expiry and one-time token consumption
- node heartbeat status and `last_seen_at` updates
- config revision metadata storage with deterministic signed subscription-aware VLESS Reality Xray config skeleton payloads
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
- `GET /api/v1/nodes`
- `POST /api/v1/nodes/bootstrap-token`
- `GET /api/v1/nodes/{id}`
- `POST /api/v1/nodes/{id}/drain`
- `POST /api/v1/nodes/{id}/undrain`
- `POST /api/v1/nodes/{id}/disable`
- `POST /api/v1/nodes/{id}/enable`
- `POST /api/v1/nodes/{id}/config-revisions`
- `GET /api/v1/nodes/{id}/config-revisions`
- `GET /api/v1/nodes/{id}/config-revisions/pending`
- `GET /api/v1/nodes/{id}/config-revisions/{revisionId}`
- `POST /api/v1/nodes/{id}/config-revisions/{revisionId}/report`
- `POST /api/v1/nodes/register`
- `POST /api/v1/nodes/{id}/heartbeat`

Admin-only routes:

- all `/api/v1/users*` routes
- all `/api/v1/plans*` routes
- all `/api/v1/subscriptions*` routes
- `GET /api/v1/nodes`
- `POST /api/v1/nodes/bootstrap-token`
- `GET /api/v1/nodes/{id}`
- `POST /api/v1/nodes/{id}/drain`
- `POST /api/v1/nodes/{id}/undrain`
- `POST /api/v1/nodes/{id}/disable`
- `POST /api/v1/nodes/{id}/enable`
- `POST /api/v1/nodes/{id}/config-revisions`
- `GET /api/v1/nodes/{id}/config-revisions`
- `GET /api/v1/nodes/{id}/config-revisions/{revisionId}`

Node-agent contract routes:

- `POST /api/v1/nodes/register` accepts a one-time bootstrap token and returns a node token
- `POST /api/v1/nodes/{id}/heartbeat` accepts a node heartbeat with `Authorization: Bearer <node_token>`
- `GET /api/v1/nodes/{id}/config-revisions/pending` accepts the same node token and returns latest pending signed revision metadata for that node
- `POST /api/v1/nodes/{id}/config-revisions/{revisionId}/report` accepts the same node token and records `applied` or `failed` revision status

Use the token returned by admin login:

```http
Authorization: Bearer <session_token>
```

OpenAPI draft:

- [docs/openapi/panel-api.v1.yaml](/Users/vaceslavibraev/Desktop/vpn_service/docs/openapi/panel-api.v1.yaml)

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
- full node orchestration engine
- full mTLS or certificate rotation
- real Xray process restart/control or rollback executor
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

## Local Development Bootstrap

This flow is for a local development database only. It is not a production installer.

Install PostgreSQL using the package manager for your OS. Examples:

```sh
# macOS with Homebrew
brew install postgresql@16
brew services start postgresql@16

# Debian/Ubuntu
sudo apt-get install postgresql
sudo systemctl start postgresql
```

Create a local database and user:

```sh
createuser lenker --pwprompt
createdb -O lenker lenker
```

Export the database URL:

```sh
export LENKER_DATABASE_URL='postgres://lenker:lenker@localhost:5432/lenker?sslmode=disable'
export LENKER_DATABASE_PING=true
```

Apply migrations from the repository root:

```sh
make migrate-up
```

Create the first local admin:

```sh
ADMIN_EMAIL=owner@example.com ADMIN_PASSWORD='change-me-now' make bootstrap-admin
```

The bootstrap helper stores a bcrypt password hash. If the admin already exists, it prints a clear message and does not change the password.

Run the API:

```sh
make run-panel-api
```

Check health:

```sh
curl -i http://localhost:8080/healthz
```

Login:

```sh
curl -s http://localhost:8080/api/v1/auth/admin/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"owner@example.com","password":"change-me-now"}'
```

Copy `data.session.token` from the response and use it as a Bearer token:

```sh
export LENKER_ADMIN_TOKEN='<session_token>'
```

Create and inspect a user:

```sh
curl -s http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","display_name":"Test User"}'

curl -s http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Create and inspect a plan:

```sh
curl -s http://localhost:8080/api/v1/plans \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Monthly","duration_days":30,"device_limit":3}'

curl -s http://localhost:8080/api/v1/plans \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Create and inspect a subscription using the `id` values returned by the user and plan calls:

```sh
curl -s http://localhost:8080/api/v1/subscriptions \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"<user_id>","plan_id":"<plan_id>"}'

curl -s http://localhost:8080/api/v1/subscriptions \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Create a one-time node bootstrap token:

```sh
curl -s http://localhost:8080/api/v1/nodes/bootstrap-token \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"finland-1","region":"eu","country_code":"FI","hostname":"node-fi-1","expires_in_minutes":30}'
```

Copy `data.bootstrap_token` and `data.node_id` from the response. The plaintext
bootstrap token is shown only once; only its hash is stored.

Register the node-agent:

```sh
curl -s http://localhost:8080/api/v1/nodes/register \
  -H 'Content-Type: application/json' \
  -d '{"node_id":"<node_id>","bootstrap_token":"<bootstrap_token>","agent_version":"0.1.0-dev","hostname":"node-fi-1"}'
```

Copy `data.node_token` from the registration response and use it for heartbeat:

```sh
curl -s http://localhost:8080/api/v1/nodes/<node_id>/heartbeat \
  -H "Authorization: Bearer <node_token>" \
  -H 'Content-Type: application/json' \
  -d '{"node_id":"<node_id>","agent_version":"0.1.0-dev","status":"active","active_revision":0}'
```

Registration rejects invalid, expired, and already used bootstrap tokens. A
heartbeat for an unknown node returns `not_found`; heartbeat does not create
nodes.

## Node Management

List nodes:

```sh
curl -s http://localhost:8080/api/v1/nodes \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Get node details:

```sh
curl -s http://localhost:8080/api/v1/nodes/<node_id> \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Drain and undrain a node:

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/<node_id>/drain \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"

curl -s -X POST http://localhost:8080/api/v1/nodes/<node_id>/undrain \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Disable and enable a node:

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/<node_id>/disable \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"

curl -s -X POST http://localhost:8080/api/v1/nodes/<node_id>/enable \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Conservative lifecycle behavior:

- `drain` sets `drain_state` to `draining`; it does not stop heartbeat.
- `undrain` sets `drain_state` to `active`; disabled nodes stay disabled.
- `disable` sets `status` to `disabled`; disabled nodes do not accept heartbeat updates.
- `enable` returns the node to `unhealthy` until the next successful heartbeat.

## Config Revision Metadata

The config foundation creates deterministic subscription-aware VLESS Reality
Xray config skeleton payloads for the single MVP protocol path and records
revision number, status, bundle hash, signature, signer, rollback target
metadata, and timestamps. The renderer uses real active users, plans, and
subscriptions as `subscription_inputs` and turns eligible subscriptions into
deterministic `access_entries`. Eligibility is intentionally simple for this
stage: active subscriptions for active users whose preferred region is empty or
matches the target node region.

The payload is structured and hash/signature stable. The panel still does not
write runtime Xray config files, restart or control Xray, or execute rollback.

Create a config revision metadata record:

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/<node_id>/config-revisions \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

List and inspect revisions:

```sh
curl -s http://localhost:8080/api/v1/nodes/<node_id>/config-revisions \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"

curl -s http://localhost:8080/api/v1/nodes/<node_id>/config-revisions/<revision_id> \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Revision creation is rejected for disabled nodes and for nodes in `draining` or
`drained` drain state.

Node-agent fetches the latest pending revision metadata through the node-facing
endpoint:

```sh
curl -s http://localhost:8080/api/v1/nodes/<node_id>/config-revisions/pending \
  -H "Authorization: Bearer <node_token>"
```

The endpoint returns `not_found` when there is no pending revision, the token
does not match the node, or the node is disabled.

Node-agent reports metadata apply status through the node-facing report
endpoint:

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/<node_id>/config-revisions/<revision_id>/report \
  -H "Authorization: Bearer <node_token>" \
  -H 'Content-Type: application/json' \
  -d '{"status":"applied","active_revision":1}'
```

Failed metadata validation can be reported as:

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/<node_id>/config-revisions/<revision_id>/report \
  -H "Authorization: Bearer <node_token>" \
  -H 'Content-Type: application/json' \
  -d '{"status":"failed","error_message":"invalid config payload"}'
```

Applied reports update revision status and node `active_revision`. Failed
reports persist `failed_at` and `error_message`. The report path does not
execute rollback, restart processes, or control Xray.

Rollback planning is metadata-only at this stage. New revisions keep
`rollback_target_revision` pointed at the node's current `active_revision`, and
previous applied revisions remain queryable through the config revision list and
detail endpoints. There is no rollback executor yet.

Manual smoke checklist:

- [docs/smoke/node-bootstrap.md](/Users/vaceslavibraev/Desktop/vpn_service/docs/smoke/node-bootstrap.md)

Useful local targets from the repository root:

```sh
make migrate-up
make migrate-down
VERSION=1 make migrate-force
ADMIN_EMAIL=owner@example.com ADMIN_PASSWORD='change-me-now' make bootstrap-admin
make run-panel-api
make test-panel-api
```
