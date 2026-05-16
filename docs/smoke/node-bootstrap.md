# Node Bootstrap Smoke Checklist

This checklist verifies the local node bootstrap, registration, heartbeat, and
admin node lifecycle flow. It is for local development only.

It does not verify mTLS, Xray runtime, config deployment, rollback, metrics, or
traffic handling.

## Prerequisites

- Go 1.22+
- PostgreSQL
- `golang-migrate/migrate`
- `curl`

From the repository root:

```sh
export LENKER_DATABASE_URL='postgres://lenker:lenker@localhost:5432/lenker?sslmode=disable'
export LENKER_DATABASE_PING=true
```

## 1. Apply Migrations

```sh
make migrate-up
```

Expected result:

- migrations finish successfully;
- `nodes`, `node_bootstrap_tokens`, and `node_registrations` tables exist.

## 2. Create First Admin

```sh
ADMIN_EMAIL=owner@example.com ADMIN_PASSWORD='change-me-now' make bootstrap-admin
```

Expected result:

- admin exists;
- password hash is stored as bcrypt.

## 3. Run panel-api

```sh
make run-panel-api
```

In another terminal:

```sh
curl -i http://localhost:8080/healthz
```

Expected response:

```json
{"data":{"status":"ok"}}
```

## 4. Login As Admin

```sh
curl -s http://localhost:8080/api/v1/auth/admin/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"owner@example.com","password":"change-me-now"}'
```

Expected result:

- response contains `data.session.token`.

Export it:

```sh
export LENKER_ADMIN_TOKEN='<session_token>'
```

## 5. Create Bootstrap Token

```sh
curl -s http://localhost:8080/api/v1/nodes/bootstrap-token \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"name":"finland-1","region":"eu","country_code":"FI","hostname":"node-fi-1","expires_in_minutes":30}'
```

Expected result:

- response contains `data.node_id`;
- response contains `data.bootstrap_token`;
- plaintext bootstrap token is shown only once.

Export returned values:

```sh
export LENKER_NODE_ID='<node_id>'
export LENKER_NODE_BOOTSTRAP_TOKEN='<bootstrap_token>'
```

## 6. Register Node

```sh
curl -s http://localhost:8080/api/v1/nodes/register \
  -H 'Content-Type: application/json' \
  -d "{\"node_id\":\"$LENKER_NODE_ID\",\"bootstrap_token\":\"$LENKER_NODE_BOOTSTRAP_TOKEN\",\"agent_version\":\"0.1.0-dev\",\"hostname\":\"node-fi-1\"}"
```

Expected result:

- response contains `data.node_token`;
- `data.status` is `active`;
- bootstrap token is consumed and cannot be reused.

Export the node token:

```sh
export LENKER_NODE_TOKEN='<node_token>'
```

Reusing the same bootstrap token should return `bootstrap_token_used`.

## 7. Send Heartbeat

```sh
curl -s http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/heartbeat \
  -H "Authorization: Bearer $LENKER_NODE_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"node_id\":\"$LENKER_NODE_ID\",\"agent_version\":\"0.1.0-dev\",\"status\":\"active\",\"active_revision\":0}"
```

Expected result:

- `data.status` is `active`;
- `data.drain_state` is `active`;
- `data.last_seen_at` is set.

## 8. List And Inspect Nodes

```sh
curl -s http://localhost:8080/api/v1/nodes \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"

curl -s http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Expected result:

- list includes the registered node;
- details include `status`, `drain_state`, `last_seen_at`, `registered_at`,
  `agent_version`, and `active_revision_id`.

## 9. Create And Fetch Pending Config Revision Metadata

Create signed dummy config revision metadata as an admin:

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/config-revisions \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Expected result:

- response contains `data.revision_number`;
- response contains `data.bundle_hash`, `data.signature`, `data.signer`, and
  `data.bundle`;
- this is dummy metadata only, not real Xray config.

Fetch the latest pending revision as the node-agent:

```sh
curl -s http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/config-revisions/pending \
  -H "Authorization: Bearer $LENKER_NODE_TOKEN"
```

Expected result:

- response contains the latest pending revision for this node;
- using another node token or a missing token does not return this revision;
- if no pending revision exists, the endpoint returns `not_found`.

The node-agent unit tests verify that the fetched metadata can be hash/signature
validated, stored in memory, and reflected in the heartbeat active revision
payload. This smoke path still does not write Xray config files, restart
processes, or execute rollback.

After metadata apply in the agent skeleton, heartbeat can report the applied
revision number:

```sh
curl -s http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/heartbeat \
  -H "Authorization: Bearer $LENKER_NODE_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"node_id\":\"$LENKER_NODE_ID\",\"agent_version\":\"0.1.0-dev\",\"status\":\"active\",\"active_revision\":1}"
```

Expected result:

- `data.active_revision` matches the reported metadata revision.

## 10. Drain And Undrain

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/drain \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Expected result:

- `data.drain_state` is `draining`;
- heartbeat is still accepted.

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/undrain \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Expected result:

- `data.drain_state` is `active`.

## 11. Disable And Enable

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/disable \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Expected result:

- `data.status` is `disabled`;
- heartbeat for this node no longer updates node state.

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/enable \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Expected result:

- `data.status` is `unhealthy`;
- the next successful heartbeat can move it back to `active`.
