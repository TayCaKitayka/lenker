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
- config revision tracking and rollback planning skeleton

Configuration:

- `LENKER_AGENT_HTTP_ADDR`
- `LENKER_AGENT_NODE_ID`
- `LENKER_AGENT_BOOTSTRAP_TOKEN`
- `LENKER_AGENT_PANEL_URL`
- `LENKER_AGENT_STATE_DIR`
- `LENKER_AGENT_LOG_LEVEL`
- `LENKER_AGENT_HEARTBEAT_INTERVAL`
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

- `POST /api/v1/nodes/register`
- `POST /api/v1/nodes/{id}/heartbeat`

Conservative note:

`LENKER_AGENT_TLS_ENABLED` is a foundation flag only. Full mTLS bootstrap, certificate rotation, network retry policy, and signed config transport are intentionally not implemented in this step.

Not included here yet:

- real node runtime logic
- VPN configuration generation
- process supervision implementation
- real Xray process control
- real config apply executor
- real rollback engine
- full mTLS or certificate rotation
- support for protocols beyond the main MVP path
