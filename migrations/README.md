# Migrations

This directory is reserved for PostgreSQL schema migrations for `Lenker MVP v0.1`.

## Naming Convention

Use sequential numeric prefixes with explicit direction:

```text
000001_core_identity.up.sql
000001_core_identity.down.sql
```

Rules:

- one logical change per migration pair
- `up.sql` applies the change
- `down.sql` reverts the change
- no marketplace or billing schema in `MVP v0.1`

## Migration Tool

Use `golang-migrate/migrate` as the migration CLI for the first implementation phase.

Example:

```sh
migrate -path migrations -database "$LENKER_DATABASE_URL" up
```

Repository Makefile targets:

```sh
make migrate-up
make migrate-down
VERSION=1 make migrate-force
```

Conservative choice:

The migration runner is not embedded in `panel-api` yet. Keeping migrations as an explicit CLI operation avoids coupling service startup to schema changes before deployment rules are defined.

## Admin Bootstrap

Admin `password_hash` values must be bcrypt hashes.

For local development, create the first admin after applying migrations:

```sh
ADMIN_EMAIL=owner@example.com ADMIN_PASSWORD='change-me-now' make bootstrap-admin
```

This helper is intentionally dev-only. It reads `LENKER_DATABASE_URL`, stores a bcrypt hash in `admins.password_hash`, and does not run automatically during service startup. If the admin already exists, it exits successfully and does not change the existing password.

Planned areas:

- admin and RBAC tables
- users and identities
- plans and subscriptions
- devices and subscription keys
- nodes and protocol profiles
- usage accounting
- API tokens and webhooks
- audit and config revision records

Not planned for `MVP v0.1`:

- marketplace schema
- billing schema
- payment records
