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

#### `GET /subscriptions/{subscriptionId}/export`

Return export metadata for the supported client path.

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
last health timestamp. Unknown nodes return `not_found`. It does not register
new nodes, and it does not store metrics, logs, config blobs, or traffic
accounting.

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

Current Stage C implementation note:

The implemented foundation uses `POST /nodes/{nodeId}/config-revisions` to
create signed dummy bundle metadata only. It stores revision number, status,
bundle hash, signature, signer, rollback target metadata, and timestamps. It
does not generate real Xray config, deliver config to an agent, apply config,
restart processes, or execute rollback.

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
pending dummy signed bundle metadata for that node. If the node is unknown, the
token does not match, the node is disabled, or no pending revision exists, it
returns `not_found`. It does not apply config, generate Xray JSON, restart
processes, or execute rollback.

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
