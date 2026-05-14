# Lenker

Lenker is an open-source VPN ecosystem for providers and users.

This repository currently contains the initial monorepo skeleton for `MVP v0.1`. The scope is intentionally narrow and follows the documents in [docs/MVP_SPEC.md](/Users/vaceslavibraev/Desktop/vpn_service/docs/MVP_SPEC.md), [docs/architecture.md](/Users/vaceslavibraev/Desktop/vpn_service/docs/architecture.md), [docs/database.md](/Users/vaceslavibraev/Desktop/vpn_service/docs/database.md), [docs/api.md](/Users/vaceslavibraev/Desktop/vpn_service/docs/api.md), and [docs/roadmap.md](/Users/vaceslavibraev/Desktop/vpn_service/docs/roadmap.md).

The project also keeps a public business boundary in [docs/business-model.md](/Users/vaceslavibraev/Desktop/vpn_service/docs/business-model.md): commercial services may exist around hosted operations, support, and managed infrastructure, but the self-hosted core remains open-source and `MVP v0.1` does not include marketplace or billing.

## MVP v0.1 Scope

Included:

- provider panel
- node agent
- users, plans, subscriptions, devices
- REST API v1
- PostgreSQL-based data model
- client app for Android, Windows, and macOS
- single production protocol path: `VLESS + Reality + XTLS Vision`

Excluded from this repository stage and from `MVP v0.1`:

- marketplace
- billing
- Telegram bot as a core module
- multi-protocol production support beyond the main path
- white-label provider builds

## Monorepo Layout

```text
.
├── apps/
│   ├── client-app/
│   └── panel-web/
├── docs/
│   ├── adr/
│   ├── MVP_SPEC.md
│   ├── api.md
│   ├── architecture.md
│   ├── business-model.md
│   ├── database.md
│   └── roadmap.md
├── migrations/
├── services/
│   ├── node-agent/
│   └── panel-api/
├── go.work
└── package.json
```

## Directory Guide

- `services/panel-api` — Go service for the provider control plane
- `services/node-agent` — Go service for managed node lifecycle
- `apps/panel-web` — React + TypeScript provider UI
- `apps/client-app` — Flutter client app for Android, Windows, and macOS
- `migrations` — database migration files for PostgreSQL
- `docs/adr` — architecture decision records

## Conservative Decisions

- The panel backend and node agent are separate Go modules.
- The web panel is prepared as a standalone React + TypeScript app.
- The client app is prepared as a standalone Flutter app shell.
- No repository area is created for marketplace or billing because they are out of scope for `MVP v0.1`.

## Status

This repository now includes the first `panel-api` backend foundation. It prepares the service entrypoint, config loading, HTTP routing, health checks, structured logging, graceful shutdown, PostgreSQL storage bootstrap, basic repository interfaces, minimal admin login foundation, and initial PostgreSQL migrations.

It does not include production business logic, billing, marketplace features, or VPN/Xray runtime logic yet.

## Backend Foundation

Run the panel API from its module:

```sh
cd services/panel-api
go run ./cmd/panel-api
```

The service exposes the first admin panel API slice:

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

`GET /healthz` is functional. `POST /api/v1/auth/admin/login` uses the initial admin auth service and session skeleton. The users, plans, and subscriptions routes are wired to PostgreSQL repositories, require the first migration to be applied, and require an admin session token in `Authorization: Bearer <session_token>`.

Conservative auth note:

Admin passwords are verified with bcrypt. Store bcrypt hashes in `admins.password_hash`.

Migration helpers are available through `make`:

```sh
make migrate-up
make migrate-down
VERSION=1 make migrate-force
```

Local development helpers are also available:

```sh
ADMIN_EMAIL=owner@example.com ADMIN_PASSWORD='change-me-now' make bootstrap-admin
make run-panel-api
make test-panel-api
```

See [services/panel-api/README.md](/Users/vaceslavibraev/Desktop/vpn_service/services/panel-api/README.md) for the full PostgreSQL, migration, first-admin, login, and protected endpoint verification flow.
