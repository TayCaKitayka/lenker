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

Conservative choice:

The migration runner is not embedded in `panel-api` yet. Keeping migrations as an explicit CLI operation avoids coupling service startup to schema changes before deployment rules are defined.

## Admin Password Hashes

The current foundation auth verifier accepts admin `password_hash` values in one of these formats:

```text
sha256$<hex>
sha256:<hex>
```

This is a conservative foundation placeholder for early backend wiring. Before real provider usage, replace it with a stronger password hashing policy and migrate stored admin hashes.

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
