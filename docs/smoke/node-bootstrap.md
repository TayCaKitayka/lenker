# Node Bootstrap Smoke Checklist

This checklist verifies the local node bootstrap, registration, heartbeat,
admin node lifecycle, and config revision report flow. It is for local
development only.

It does not verify mTLS, Xray runtime process control, process restart/reload,
metrics, or traffic handling.

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

## 9. Create, Fetch, And Report Config Revision Metadata

Create signed config revision metadata as an admin:

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/config-revisions \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Expected result:

- response contains `data.revision_number`;
- response contains `data.bundle_hash`, `data.signature`, `data.signer`, and
  `data.bundle`;
- `data.bundle` contains a deterministic subscription-aware VLESS Reality Xray-
  compatible skeleton payload for the single MVP path;
- `data.bundle.config` contains Xray-like `log`, `policy`, `stats`, `inbounds`,
  `outbounds`, and `routing` sections;
- `data.bundle.subscription_inputs` and `data.bundle.access_entries` are arrays,
  empty if there are no active eligible subscriptions for this node region;
- `data.rollback_target_revision` points at the node's active revision, or `0`
  if none has been applied yet.

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
validated, checked against the single-path Xray compatibility gate, optionally
checked through a configured Xray binary dry-run, serialized to local config
artifacts, stored in memory, and reflected in the heartbeat active revision
payload. Set `LENKER_AGENT_XRAY_BIN=/path/to/xray` to enable the optional
one-shot `xray run -test -config <candidate>` boundary. Leave it unset to use
the default internal validation path.

For Docker local development, the default profile leaves Xray dry-run disabled:

```sh
unset LENKER_AGENT_XRAY_BIN
make docker-up
```

If a local Xray binary is already installed, opt in explicitly without
downloading or baking it into the image:

```sh
export LENKER_LOCAL_XRAY_DIR="$(dirname "$(command -v xray)")"
export LENKER_AGENT_XRAY_BIN=/opt/lenker/xray/xray
make docker-up
```

The compose file bind-mounts `$LENKER_LOCAL_XRAY_DIR` to `/opt/lenker/xray`.
`GET http://localhost:8090/status` should show
`"xray_dry_run_enabled":true`. A successful pending revision apply then proves
the candidate passed `xray run -test -config <candidate>` before staged ->
active. To verify the failure path, set `LENKER_AGENT_XRAY_BIN` to a missing
path and create a new revision; the revision should report `failed` with
`xray_dry_run_failed:xray_binary_not_found`, while the previous active config
remains unchanged.

The node detail response also exposes read-only runtime readiness metadata after
apply/failure: `last_validation_status`, `last_validation_error`,
`last_validation_at`, `last_applied_revision`, and `active_config_path`.

Report the pending revision as applied:

```sh
export LENKER_REVISION_ID='<revision_id>'

curl -s -X POST http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/config-revisions/$LENKER_REVISION_ID/report \
  -H "Authorization: Bearer $LENKER_NODE_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"applied","active_revision":1}'
```

Expected result:

- `data.status` is `applied`;
- `data.applied_at` is set;
- a later admin config revision fetch shows the applied status.

Validation failures can be reported as failed:

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/config-revisions/$LENKER_REVISION_ID/report \
  -H "Authorization: Bearer $LENKER_NODE_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"status":"failed","error_message":"invalid_xray_config:missing_stream_settings"}'
```

Optional Xray binary dry-run failures use the same failed report path with a
compact reason such as `xray_dry_run_failed:invalid_config`.

Expected result:

- `data.status` is `failed`;
- `data.failed_at` is set;
- `data.error_message` contains the compact failure reason;
- node `active_revision` does not advance.

This smoke path still does not restart processes or control Xray. The node-agent
compatibility and serialization foundation writes local artifacts only under its
state directory after hash/signature validation, internal Xray compatibility
validation, and optional Xray binary dry-run validation all pass:

```text
revisions/<revision_number>/config.json
revisions/<revision_number>/metadata.json
staged/config.json
staged/metadata.json
active/config.json
active/metadata.json
state.json
```

Rollback is requested by creating a new pending revision from an applied source:

```sh
curl -s -X POST http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/config-revisions/$LENKER_REVISION_ID/rollback \
  -H "Authorization: Bearer $LENKER_ADMIN_TOKEN"
```

Expected result:

- response contains a new pending revision;
- `data.bundle.operation_kind` is `rollback`;
- `data.bundle.source_revision_number` points at the applied source revision;
- the agent later applies it with the same staged -> active file switch path;
- no Xray process is restarted or controlled.

After metadata apply in the agent skeleton, heartbeat can report the applied
revision number and runtime readiness metadata:

```sh
curl -s http://localhost:8080/api/v1/nodes/$LENKER_NODE_ID/heartbeat \
  -H "Authorization: Bearer $LENKER_NODE_TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"node_id\":\"$LENKER_NODE_ID\",\"agent_version\":\"0.1.0-dev\",\"status\":\"active\",\"active_revision\":1,\"last_validation_status\":\"applied\",\"last_validation_at\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",\"last_applied_revision\":1,\"active_config_path\":\"/var/lib/lenker/node-agent/active/config.json\"}"
```

Expected result:

- `data.active_revision` matches the reported metadata revision.
- `data.last_validation_status` is `applied`.

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
