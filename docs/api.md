# Lenker REST API v1 for MVP v0.1

## Purpose

This document defines the first REST API surface for `Lenker MVP v0.1`.

The API is intentionally limited to the first release scope:

- provider panel operations
- node lifecycle operations
- user subscription operations
- client application operations
- webhook integration

Marketplace and billing APIs are not part of this version.

## API Principles

- versioned REST API
- JSON request and response bodies
- OpenAPI as the canonical machine-readable contract
- scoped API tokens
- audit logging for sensitive actions
- idempotency for retry-prone write operations

Base path:

```text
/api/v1
```

## Authentication Model

### Admin API

- session-based auth for panel operators
- optional token-based auth for automation

### User App API

- email-first authentication
- app session token after successful sign-in

Current implementation note:

Full email-first user authentication is still future work. The current
consumer-facing foundation is deliberately narrower: a provider admin can issue
a subscription access token for one active subscription, and a consumer can use
that token only to read the redacted access export for the single MVP protocol
path.

### Node API

- `HTTPS + mTLS`
- bootstrap token only for first registration

## Resource Groups

### Auth

#### `POST /auth/admin/login`

Authenticate a panel admin.

#### `POST /auth/admin/logout`

Terminate the current admin session.

#### `GET /auth/admin/me`

Return the current admin identity and permissions.

#### `POST /auth/user/request-magic-link`

Start email-first sign-in for a user.

#### `POST /auth/user/verify-magic-link`

Complete user sign-in and return an app session.

#### `GET /auth/user/me`

Return the current user identity and subscription summary.

Conservative note:

For `MVP v0.1`, email magic link is the preferred user auth flow.

### Users

#### `GET /users`

List users with filters by status, plan, and expiration.

#### `POST /users`

Create a user.

#### `GET /users/{userId}`

Return user details.

#### `PATCH /users/{userId}`

Update editable user fields.

#### `POST /users/{userId}/suspend`

Suspend user access.

#### `POST /users/{userId}/activate`

Re-enable user access.

### Plans

#### `GET /plans`

List plans.

#### `POST /plans`

Create a plan.

#### `GET /plans/{planId}`

Return plan details.

#### `PATCH /plans/{planId}`

Update a plan.

#### `POST /plans/{planId}/archive`

Archive a plan from future assignment.

### Subscriptions

#### `GET /subscriptions`

List subscriptions with filters.

#### `POST /subscriptions`

Create a subscription for a user and plan.

#### `GET /subscriptions/{subscriptionId}`

Return subscription details.

#### `PATCH /subscriptions/{subscriptionId}`

Update editable subscription fields.

#### `POST /subscriptions/{subscriptionId}/renew`

Manually renew a subscription.

#### `POST /subscriptions/{subscriptionId}/rotate-key`

Issue a new key version and make it active.

#### `POST /subscriptions/{subscriptionId}/revoke-active-key`

Revoke the currently active key.

#### `POST /subscriptions/{subscriptionId}/reset-devices`

Clear or revoke bound devices for the subscription.

#### `GET /subscriptions/{subscriptionId}/usage`

Return current and recent usage data.

#### `GET /subscriptions/{subscriptionId}/access`

Return provider-side access export metadata for the supported client path.

#### `POST /subscriptions/{subscriptionId}/access-token`

Issue a plaintext subscription access token for the active subscription.

Current implementation note:

This admin-only endpoint stores only a SHA-256 token hash and returns the
plaintext token once in the response. The token expires with the subscription
and is accepted only by the consumer-facing access read endpoint.

#### `GET /client/subscription-access`

Return a redacted subscription access export using
`Authorization: Bearer <subscription_access_token>`.

Current implementation note:

The implemented first product-layer slice exposes an admin-only read model for
an active subscription. It derives a deterministic
`subscription_access.v1alpha1` payload for the single MVP path,
`VLESS + Reality + XTLS Vision`, using the existing subscription, user, plan,
and node tables. The MVP node-selection rule chooses the first active,
non-draining node with a hostname that matches the subscription
`preferred_region` when one is set; otherwise it uses stable ordering by region,
name, and id. The response includes a structured endpoint/client payload and a
minimal VLESS URI. It is not an end-user auth flow, device-management flow,
marketplace export, or multi-protocol delivery system.

The client read endpoint uses the same deterministic export derivation but does
not require or accept an admin session. It returns the same endpoint/client URI
material without provider-internal user id, user label, or plan id fields.

Conservative note:

For `MVP v0.1`, the export path should prioritize the native Lenker client workflow. External export formats can remain minimal.

### Devices

#### `GET /devices`

List devices.

#### `GET /devices/{deviceId}`

Return device details.

#### `POST /devices/{deviceId}/revoke`

Revoke a device from future use.

### Nodes

#### `GET /nodes`

List managed nodes.

#### `POST /nodes/bootstrap-token`

Create a one-time node bootstrap token.

Current implementation note:

This admin-protected endpoint creates a pending node and returns the plaintext
bootstrap token only once. The database stores only a token hash, expiry, and
used timestamp.

#### `POST /nodes/register`

Complete node registration. Intended for the node agent.

Current implementation note:

The current backend slice validates an active, unexpired, unused bootstrap
token, consumes it on success, activates the pending node, and returns a node
token for subsequent heartbeats. Invalid, expired, and reused bootstrap tokens
are rejected with explicit error codes. Full mTLS bootstrap and certificate
rotation remain future work inside `MVP v0.1`.

#### `POST /nodes/{nodeId}/heartbeat`

Record a node-agent heartbeat.

Current implementation note:

The current backend slice accepts `Authorization: Bearer <node_token>` and
updates basic node status, agent version, active revision, `last_seen_at`, and
last health timestamp. It also accepts read-only runtime readiness metadata:
`last_validation_status`, `last_validation_error`, `last_validation_at`,
`last_applied_revision`, `active_config_path`, runtime preparation fields, and
the explicit `runtime_process_mode` / `runtime_process_state` process runner
gate. Heartbeats may also include a compact `runtime_events` slice; panel-api
keeps only the newest bounded recent events per node. Unknown nodes return
`not_found`. It does not register new nodes, and it does not store metrics,
logs, config blobs, or traffic accounting.

#### `GET /nodes/{nodeId}`

Return node details.

#### `PATCH /nodes/{nodeId}`

Update node metadata.

#### `POST /nodes/{nodeId}/drain`

Put the node into drain mode.

Current implementation note:

The current backend sets `drain_state` to `draining`. This records provider
intent without pretending that traffic has already moved away from the node.
Heartbeat continues to work while a node is draining.

#### `POST /nodes/{nodeId}/undrain`

Return the node to active service.

Current implementation note:

The current backend sets `drain_state` back to `active`. Disabled nodes cannot
be silently activated by undrain.

#### `POST /nodes/{nodeId}/disable`

Disable a node.

Current implementation note:

Disabled nodes do not accept heartbeat updates.

#### `POST /nodes/{nodeId}/enable`

Enable a disabled node.

Current implementation note:

The current backend returns enabled nodes to `unhealthy` until the next
successful heartbeat proves the node is active.

#### `POST /nodes/{nodeId}/deploy-config`

Generate and deploy the next signed config revision.

Current implementation note:

The implemented foundation uses `POST /nodes/{nodeId}/config-revisions` to
create deterministic signed subscription-aware VLESS Reality Xray-compatible
skeleton payloads for the single MVP path. It derives `subscription_inputs` and
`access_entries` from active subscriptions, active users, plans, and target node
region. The rendered `config` object follows an Xray-like shape with `log`,
`policy`, `stats`, VLESS Reality `inbounds`, `outbounds`, and `routing`. It
stores revision number, status, bundle hash, signature, signer, rollback target
metadata, and timestamps. Panel-api runs a lightweight renderer precheck before
signing; node-agent enforces the authoritative compatibility gate before staged
files become active. If `LENKER_AGENT_XRAY_BIN` is configured, node-agent also
runs a one-shot Xray binary dry-run before the staged -> active switch. It does
not restart processes.

#### `GET /nodes/{nodeId}/config-revisions`

List stored config revision metadata for a node.

#### `GET /nodes/{nodeId}/config-revisions/{revisionId}`

Return one config revision metadata record and verify it belongs to the node in
the path.

#### `GET /nodes/{nodeId}/config-revisions/pending`

Return the latest pending signed config revision metadata for the node agent.

Current implementation note:

This node-facing endpoint requires `Authorization: Bearer <node_token>`, checks
that the token belongs to the node in the path, and returns only the latest
pending signed config skeleton payload metadata for that node. If the node is
unknown, the token does not match, the node is disabled, or no pending revision
exists, it returns `not_found`. The node-agent validates the payload and
serializes local config artifacts before reporting `applied`. Optional Xray
binary dry-run validation runs here when configured on the agent.

#### `POST /nodes/{nodeId}/config-revisions/{revisionId}/report`

Report a node-agent config revision status transition.

Current implementation note:

This node-facing endpoint requires `Authorization: Bearer <node_token>`, checks
that the token belongs to the node in the path, and allows the node to report
only its own revision as `applied` or `failed`. Applied reports set the revision
`applied_at` timestamp and update the node active revision. Failed reports set
`failed_at` and persist a concise `error_message`; node active revision is not
advanced on failed validation. Xray compatibility failures use stable summaries
such as `invalid_xray_config:invalid_routing_outbound_reference`; optional Xray
binary dry-run failures use `xray_dry_run_failed:<reason>`. Reports also update
the node read-only runtime readiness fields shown in admin node detail and may
carry the same bounded `runtime_events` trail as heartbeat. It does not execute
rollback, restart processes, or control Xray. The optional
`runtime_process_mode=local` value is only a node-agent local skeleton signal;
it does not mean the panel starts or supervises a daemon.

#### `POST /nodes/{nodeId}/config-revisions/{revisionId}/rollback`

Create a pending rollback revision from an applied config revision.

Current implementation note:

This admin endpoint verifies that the target revision belongs to the node and is
applied, then creates a new pending revision whose signed payload preserves the
target rendered config object and adds rollback source metadata. The new
revision carries `rollback_target_revision` set to the node's current active
revision. Node-agent later applies it through the normal polling path. Panel-api
does not push to nodes, mutate local files, restart processes, or execute a
runtime rollback.

#### `POST /nodes/{nodeId}/rollback`

Roll back to the last known good config revision.

#### `GET /nodes/{nodeId}/health`

Return the latest node health snapshot.

#### `GET /nodes/{nodeId}/logs`

Return filtered node logs subject to RBAC.

### Protocol Profiles

#### `GET /protocol-profiles`

List protocol profiles.

#### `POST /protocol-profiles`

Create a protocol profile.

#### `GET /protocol-profiles/{profileId}`

Return protocol profile details.

#### `PATCH /protocol-profiles/{profileId}`

Update a protocol profile.

Conservative note:

The first release requires only one production-ready profile: `VLESS + Reality + XTLS Vision`.

### API Tokens

#### `GET /api-tokens`

List API tokens.

#### `POST /api-tokens`

Create a scoped API token.

#### `DELETE /api-tokens/{tokenId}`

Revoke an API token.

### Webhooks

#### `GET /webhooks`

List webhook configurations.

#### `POST /webhooks`

Create a webhook target.

#### `PATCH /webhooks/{webhookId}`

Update a webhook target.

#### `POST /webhooks/{webhookId}/test`

Send a test delivery.

### Incoming External Events

#### `POST /webhooks/subscription-renewed`

Accept an external event that renews a subscription.

#### `POST /webhooks/subscription-suspended`

Accept an external event that suspends a subscription.

Conservative note:

For `MVP v0.1`, webhook-driven subscription lifecycle is enough. Full billing resource APIs are intentionally excluded.

### Client App

#### `GET /app/me`

Return user profile and provider branding data.

#### `GET /app/subscription`

Return the active subscription for the signed-in user.

#### `GET /app/regions`

Return available regions or node choices for the subscription.

#### `POST /app/connect-preferences`

Save preferred region or auto-selection mode.

#### `POST /app/rotate-key`

Rotate the active subscription key for the signed-in user.

#### `GET /app/diagnostics`

Return basic app-facing subscription and service diagnostics.

## Cross-Cutting Behavior

### Idempotency

The following write operations should accept idempotency keys:

- subscription renew
- key rotation
- webhook ingest
- node config deploy

### Audit Requirements

The following actions must be audit logged:

- admin login
- user suspension
- subscription renew
- key rotation
- node drain and undrain
- config deploy and rollback
- API token creation and revocation

### Error Model

Responses should use stable error codes and machine-readable payloads.

Suggested fields:

- `code`
- `message`
- `details`
- `request_id`

## Out of Scope for API v1

- marketplace resources
- billing resources
- payment intent resources
- provider review resources
- Telegram bot APIs
- white-label build APIs
