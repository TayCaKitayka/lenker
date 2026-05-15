# Lenker

English | [Русский](README.ru.md)

Lenker is an early-stage open-source VPN ecosystem for providers and users.

It is not a ready-to-run VPN service yet. The current repository focuses on the backend foundation for a provider control plane, managed nodes, subscriptions, and one MVP protocol path: `VLESS + Reality + XTLS Vision`.

## What Is Lenker

Lenker is intended to become a self-hosted VPN operations stack:

```text
provider panel -> node agent -> subscriptions -> client app -> future marketplace
```

The first product milestone, `MVP v0.1`, is deliberately narrow. It is about proving that a provider can manage users, plans, subscriptions, and nodes before the project grows into billing, marketplace, multi-protocol support, or production client distribution.

## Current Status

Lenker is under active foundation development.

Current repository state:

- `panel-api` foundation exists.
- Admin auth uses bcrypt password verification.
- Admin session middleware uses `Authorization: Bearer <session_token>`.
- Admin CRUD slice exists for users, plans, and subscriptions.
- Local development bootstrap can create the first admin.
- PostgreSQL migrations exist for identity, subscriptions, and node foundation tables.
- OpenAPI draft and lightweight validation are in place.
- GitHub Actions runs backend and OpenAPI checks.
- `node-agent` foundation exists.
- `panel-api` and `node-agent` have a bootstrap token, registration, and heartbeat contract.

Not ready yet:

- production VPN runtime
- real Xray process control
- signed config deployment
- real rollback executor
- full mTLS/certificate lifecycle
- production client app

## MVP v0.1 Scope

Included in `MVP v0.1`:

- provider panel backend
- node agent
- users, plans, and subscriptions
- node registration and heartbeat
- PostgreSQL-backed state
- REST API and OpenAPI draft
- manual renewal, API, and webhook foundation
- Android, Windows, and macOS client app target
- one production protocol path: `VLESS + Reality + XTLS Vision`

Explicitly not included in `MVP v0.1`:

- marketplace implementation
- built-in billing or payment processing
- provider ranking, reviews, or commission flow
- Telegram bot as a core module
- iOS or Linux client
- production multi-protocol support
- white-label builds
- enterprise SSO
- migration tools from other panels
- full analytics or support ticketing

## What Works Today

Backend foundation:

- `GET /healthz`
- `POST /api/v1/auth/admin/login`
- admin-protected users API
- admin-protected plans API
- admin-protected subscriptions API
- `POST /api/v1/nodes/register`
- `POST /api/v1/nodes/bootstrap-token`
- admin-protected node list/detail and lifecycle actions
- `POST /api/v1/nodes/{id}/heartbeat`

Node-agent foundation:

- `GET /healthz`
- `GET /status`
- env-based config loading
- registration payload builder
- heartbeat payload builder
- config revision and rollback placeholder models

Local tooling:

- PostgreSQL migrations through `golang-migrate/migrate`
- first-admin bootstrap CLI
- OpenAPI validation
- unit and contract tests
- GitHub Actions CI

## Repository Layout

```text
.
├── apps/
│   ├── client-app/
│   └── panel-web/
├── docs/
│   ├── adr/
│   ├── openapi/
│   ├── MVP_SPEC.md
│   ├── api.md
│   ├── architecture.md
│   ├── business-model.md
│   ├── database.md
│   └── roadmap.md
├── migrations/
├── scripts/
├── services/
│   ├── node-agent/
│   └── panel-api/
├── Makefile
├── README.md
├── README.ru.md
├── go.work
└── package.json
```

## Quick Start

Prerequisites for current backend work:

- Go 1.22+
- Ruby, for the lightweight OpenAPI validator
- PostgreSQL
- `golang-migrate/migrate`

Set a local database URL:

```sh
export LENKER_DATABASE_URL='postgres://lenker:lenker@localhost:5432/lenker?sslmode=disable'
export LENKER_DATABASE_PING=true
```

Apply migrations:

```sh
make migrate-up
```

Create the first local admin:

```sh
ADMIN_EMAIL=owner@example.com ADMIN_PASSWORD='change-me-now' make bootstrap-admin
```

Run the panel API:

```sh
make run-panel-api
```

Run the node agent foundation:

```sh
make run-node-agent
```

See [services/panel-api/README.md](services/panel-api/README.md) for a fuller local flow with curl examples.

## Checks

Run all current checks:

```sh
make test
```

This runs:

- `go test ./...` in `services/panel-api`
- `go test ./...` in `services/node-agent`
- OpenAPI validation for `docs/openapi/panel-api.v1.yaml`

Focused commands:

```sh
make test-panel-api
make test-node-agent
make openapi-lint
```

GitHub Actions runs `make test` on push and pull requests.

## Documentation

- [MVP spec](docs/MVP_SPEC.md)
- [Architecture](docs/architecture.md)
- [Database model](docs/database.md)
- [REST API plan](docs/api.md)
- [OpenAPI draft](docs/openapi/panel-api.v1.yaml)
- [OpenAPI notes](docs/openapi/README.md)
- [Roadmap](docs/roadmap.md)
- [Business model boundary](docs/business-model.md)
- [Node bootstrap smoke checklist](docs/smoke/node-bootstrap.md)
- [Licensing notes](docs/licensing.md)
- [Architecture decision records](docs/adr/README.md)
- [panel-api README](services/panel-api/README.md)
- [node-agent README](services/node-agent/README.md)

## Governance

- [License](LICENSE)
- [Licensing notes](docs/licensing.md)
- [Security policy](SECURITY.md)
- [Contributing guide](CONTRIBUTING.md)
- [Trademark policy](TRADEMARK.md)
- [Code of conduct](CODE_OF_CONDUCT.md)

## Business Model Boundary

Lenker is planned as an open-source core with commercial services around it, not as a crippled self-host demo.

The self-hosted core should remain useful for small providers. Future commercial work may include Lenker Cloud, managed nodes, paid support, enterprise governance, billing plugins, migration services, and marketplace trust services.

The project must not monetize user data, DNS history, browsing history, connection logs, hidden telemetry, provider logs, or pay-to-win marketplace ranking.

Marketplace and billing are not part of `MVP v0.1`.

## Security And Privacy Stance

Lenker should be privacy-first by default:

- minimal logging by default
- no sale of user data or traffic history
- no hidden telemetry
- no billing or marketplace tables in the MVP schema
- session and node tokens are stored as hashes where implemented
- full mTLS and certificate rotation are planned but not complete yet

This repository is not production-hardened yet. Do not treat it as a complete secure VPN platform.

## Roadmap

Current direction:

1. Finish backend foundation for provider operations.
2. Build node registration, heartbeat, config revision, apply, and rollback foundations.
3. Add panel web flows for admins.
4. Add the client app flow for Android, Windows, and macOS.
5. Harden release packaging, deployment docs, security policy, backup, and recovery.

Post-MVP topics such as marketplace, provider verification, billing adapters, Lenker Cloud, paid support, and enterprise features are tracked as later work.

## Contributing Status

The project is not ready for broad external contribution yet. Early issues, architecture feedback, security concerns, and focused backend review are useful.

Before opening large PRs, align with the fixed `MVP v0.1` scope and avoid adding marketplace, billing, multi-protocol runtime, or production VPN logic prematurely.

## License Note

Lenker is currently licensed under `AGPL-3.0-only` via the root
[LICENSE](LICENSE).

This choice is intended to keep the self-hosted core open-source and discourage
closed hosted forks of the control plane. Future SDKs or specs may receive a
more permissive license later, but no such split exists today.

See [docs/licensing.md](docs/licensing.md) and [TRADEMARK.md](TRADEMARK.md) for
the current project boundary.

Do not assume final licensing until a `LICENSE` file and license ADR are added.
