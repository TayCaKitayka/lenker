# Lenker Docker local dev

This directory contains the local Docker profile for developer smoke checks.
It is not a production deployment guide.

## Services

- `postgres` - PostgreSQL 16 for local development.
- `migrate` - applies SQL migrations from `migrations/`.
- `panel-api` - Lenker provider API on `http://localhost:8080`.
- `bootstrap-admin` - one-shot helper for creating a local admin.
- `node-agent` - Lenker node-agent foundation on `http://localhost:8090`.

## Run

From the repository root:

```sh
make docker-build
make docker-up
make docker-bootstrap-admin
make docker-smoke
```

Default local admin credentials created by `make docker-bootstrap-admin`:

```text
email: owner@example.com
password: change-me-now
```

Login:

```sh
curl -s http://localhost:8080/api/v1/auth/admin/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"owner@example.com","password":"change-me-now"}'
```

Health checks:

```sh
curl -fsS http://localhost:8080/healthz
curl -fsS http://localhost:8090/healthz
```

Stop containers:

```sh
make docker-down
```

Remove local database and agent state volumes:

```sh
docker compose -f deploy/docker/docker-compose.local.yml down -v
```
