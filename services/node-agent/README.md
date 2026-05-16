# node-agent

`node-agent` is the Go service foundation that runs on managed Lenker nodes.

Planned responsibilities for `MVP v0.1`:

- one-time bootstrap registration
- `HTTPS + mTLS` trust establishment
- node health reporting
- basic metrics reporting
- signed config bundle retrieval
- atomic config apply and rollback
- drain mode support
- local Xray process control for the main protocol path

Current foundation:

- application entrypoint at `cmd/node-agent`
- environment-based config loading
- structured JSON logging through Go `slog`
- HTTP server with graceful shutdown
- local `GET /healthz`
- local `GET /status`
- agent identity, registration payload, heartbeat payload, status, config revision, and rollback placeholder models
- registration and heartbeat request builders
- signed config revision validation with in-memory metadata storage and local config artifact serialization
- config revision tracking and rollback planning skeleton

Configuration:

- `LENKER_AGENT_HTTP_ADDR`
- `LENKER_AGENT_NODE_ID`
- `LENKER_AGENT_BOOTSTRAP_TOKEN`
- `LENKER_AGENT_NODE_TOKEN`
- `LENKER_AGENT_PANEL_URL`
- `LENKER_AGENT_STATE_DIR`
- `LENKER_AGENT_LOG_LEVEL`
- `LENKER_AGENT_HEARTBEAT_INTERVAL`
- `LENKER_AGENT_CONFIG_POLL_INTERVAL`
- `LENKER_AGENT_TLS_ENABLED`

Local run:

```sh
go run ./cmd/node-agent
```

From the repository root:

```sh
make run-node-agent
make test-node-agent
```

Local HTTP surface:

- `GET /healthz`
- `GET /status`

Panel contract currently implemented:

- `POST /api/v1/nodes/bootstrap-token`
- `POST /api/v1/nodes/register`
- `POST /api/v1/nodes/{id}/heartbeat`
- `GET /api/v1/nodes/{id}/config-revisions/pending`
- `POST /api/v1/nodes/{id}/config-revisions/{revisionId}/report`

Registration payload:

```json
{
  "node_id": "<node_id-from-bootstrap-token-response>",
  "bootstrap_token": "<plaintext-bootstrap-token>",
  "agent_version": "0.1.0-dev",
  "hostname": "node-hostname"
}
```

Heartbeat payload:

```json
{
  "node_id": "<registered-node-id>",
  "agent_version": "0.1.0-dev",
  "status": "active",
  "active_revision": 0,
  "sent_at": "2026-05-15T00:00:00Z"
}
```

Current node lifecycle statuses are `pending`, `active`, `unhealthy`,
`drained`, and `disabled`. The node-agent foundation builds payloads only; it
does not implement a retrying network client or mTLS certificate lifecycle yet.

Drain and disable operations are controlled by `panel-api` admin endpoints. The
agent continues to build heartbeat payloads; disabled nodes are rejected by the
panel until an admin enables them again.

Conservative note:

`LENKER_AGENT_TLS_ENABLED` is a foundation flag only. Full mTLS bootstrap,
certificate rotation, and production network retry policy are intentionally not
implemented in this step.

Config delivery/apply foundation:

The agent has a small panel client for fetching the latest pending signed
revision metadata with `Authorization: Bearer <node_token>`. A polling loop runs
on `LENKER_AGENT_CONFIG_POLL_INTERVAL`, treats `404 not_found` as no-op, rejects
unauthorized or malformed responses, validates the bundle hash and deterministic
dev signature, verifies the deterministic subscription-aware VLESS Reality Xray
config skeleton payload shape, stores metadata in memory, serializes local
config artifacts, and updates the active/applied revision in status and
heartbeat payloads after serialization succeeds.

After validation, the agent reports `applied` to panel-api. Validation failures
such as bad hash, bad signature, malformed payload, or local artifact write
failure are reported as `failed` with a concise `error_message`.

Local artifact layout under `LENKER_AGENT_STATE_DIR`:

```text
revisions/<revision_number>/config.json
revisions/<revision_number>/metadata.json
active/config.json
active/metadata.json
```

Writes use a temp-file then rename pattern. `metadata.json` includes the
revision id, bundle hash, signer, rollback target revision, and config path
references. This apply step prepares local files only; it does not start,
restart, reload, or supervise Xray, and it does not execute rollback.

Not included here yet:

- real node runtime logic
- VPN configuration generation
- process supervision implementation
- real Xray process control
- real config apply executor beyond local serialization
- real rollback engine
- full mTLS or certificate rotation
- support for protocols beyond the main MVP path
