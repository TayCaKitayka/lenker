# Lenker Database Model for MVP v0.1

## Purpose

This document describes the logical database model for `Lenker MVP v0.1`. It focuses only on entities required for the first release.

## Database Choice

- `PostgreSQL`

PostgreSQL is the primary source of truth for panel state and provider data.

## Design Scope

Included:

- admins and RBAC
- users and identities
- plans and subscriptions
- devices and subscription keys
- nodes and protocol profiles
- usage accounting
- API tokens and webhooks
- audit and config revision records

Excluded:

- marketplace provider registry
- reviews and ratings
- payments, invoices, refunds, promo codes
- support tickets

## Core Entities

### admins

Provider-side operators of the panel.

Suggested fields:

- `id`
- `email`
- `password_hash`
- `status`
- `two_factor_enabled`
- `created_at`
- `updated_at`
- `last_login_at`

### admin_roles

Defines named roles such as:

- owner
- admin
- support
- node_operator
- auditor

Suggested fields:

- `id`
- `name`
- `description`

### admin_role_bindings

Maps admins to roles.

Suggested fields:

- `id`
- `admin_id`
- `role_id`

### admin_sessions

Tracks active admin sessions.

Suggested fields:

- `id`
- `admin_id`
- `session_token_hash`
- `expires_at`
- `created_at`
- `last_seen_at`

### users

End users of the provider service.

Suggested fields:

- `id`
- `email`
- `status`
- `display_name`
- `created_at`
- `updated_at`

### user_identities

Stores authentication identity data for email-first auth.

Suggested fields:

- `id`
- `user_id`
- `type`
- `email`
- `email_verified_at`
- `magic_link_nonce_hash`
- `last_auth_at`

Conservative note:

For `MVP v0.1`, the only required identity type is `email`.

### plans

Commercial and operational subscription templates.

Suggested fields:

- `id`
- `name`
- `duration_days`
- `traffic_limit_bytes`
- `device_limit`
- `region_policy`
- `status`
- `created_at`
- `updated_at`

### subscriptions

Operational subscription records assigned to users.

Suggested fields:

- `id`
- `user_id`
- `plan_id`
- `status`
- `starts_at`
- `expires_at`
- `traffic_limit_bytes`
- `traffic_used_bytes`
- `device_limit`
- `preferred_region`
- `last_key_rotated_at`
- `created_at`
- `updated_at`

### subscription_keys

Tracks active and historical subscription keys.

Suggested fields:

- `id`
- `subscription_id`
- `key_version`
- `key_material_encrypted`
- `status`
- `issued_at`
- `revoked_at`

### devices

Tracks user devices bound to a subscription.

Suggested fields:

- `id`
- `subscription_id`
- `device_label`
- `platform`
- `device_token_hash`
- `status`
- `first_seen_at`
- `last_seen_at`

### nodes

Managed infrastructure nodes.

Suggested fields:

- `id`
- `name`
- `region`
- `country_code`
- `hostname`
- `public_ipv4`
- `public_ipv6`
- `status`
- `drain_state`
- `agent_version`
- `xray_version`
- `last_health_at`
- `last_seen_at`
- `registered_at`
- `created_at`
- `updated_at`

### node_bootstrap_tokens

One-time bootstrap credentials created by an admin before a node-agent registers.
The plaintext token is returned only once and only its hash is stored.

Suggested fields:

- `id`
- `node_id`
- `token_hash`
- `status`
- `expires_at`
- `used_at`
- `created_by_admin_id`
- `created_at`
- `updated_at`

Conservative note:

Bootstrap tokens are separate from long-lived node heartbeat tokens. Tokens are
one-time use and expire; registration consumes the token and stores only hashes.

### node_groups

Logical groups for assignment and operations.

Suggested fields:

- `id`
- `name`
- `description`

### node_group_members

Maps nodes to groups.

Suggested fields:

- `id`
- `node_group_id`
- `node_id`

### protocol_profiles

Provider-defined protocol presets.

Suggested fields:

- `id`
- `name`
- `type`
- `transport`
- `port`
- `reality_enabled`
- `xtls_enabled`
- `status`
- `created_at`
- `updated_at`

Conservative note:

For `MVP v0.1`, only one production profile type is required: `VLESS + Reality + XTLS Vision`.

### subscription_node_assignments

Associates subscriptions with eligible or preferred nodes.

Suggested fields:

- `id`
- `subscription_id`
- `node_id`
- `assignment_type`
- `priority`

### usage_counters

Stores near-current usage counters.

Suggested fields:

- `id`
- `subscription_id`
- `node_id`
- `bytes_up`
- `bytes_down`
- `measured_at`

### traffic_rollups_daily

Stores daily aggregated usage.

Suggested fields:

- `id`
- `subscription_id`
- `date`
- `bytes_up`
- `bytes_down`

### api_tokens

Provider API tokens for automation.

Suggested fields:

- `id`
- `name`
- `token_hash`
- `scope`
- `status`
- `expires_at`
- `created_at`

### webhooks

Outbound or configured webhook targets.

Suggested fields:

- `id`
- `name`
- `target_url`
- `signing_secret_encrypted`
- `event_mask`
- `status`
- `created_at`

### webhook_deliveries

Webhook delivery attempts and results.

Suggested fields:

- `id`
- `webhook_id`
- `event_type`
- `delivery_status`
- `attempt_count`
- `last_attempt_at`

### audit_logs

Immutable records of sensitive panel actions.

Suggested fields:

- `id`
- `actor_type`
- `actor_id`
- `action`
- `resource_type`
- `resource_id`
- `metadata_json`
- `created_at`

### system_events

Operational events generated by the panel or nodes.

Suggested fields:

- `id`
- `source`
- `severity`
- `event_type`
- `message`
- `created_at`

### config_revisions

Version history for node configuration bundles.

Suggested fields:

- `id`
- `node_id`
- `revision_number`
- `bundle_hash`
- `signature`
- `status`
- `created_at`
- `applied_at`

### node_registrations

Tracks bootstrap and trust establishment for node agents.

Suggested fields:

- `id`
- `node_id`
- `bootstrap_token_hash`
- `registration_status`
- `certificate_fingerprint`
- `registered_at`

### provider_branding

Stores provider presentation data used by the client app.

Suggested fields:

- `id`
- `display_name`
- `logo_url`
- `primary_color`
- `support_url`
- `updated_at`

## Key Relationships

- `admins` many-to-many `admin_roles` through `admin_role_bindings`
- `users` one-to-many `user_identities`
- `users` one-to-many `subscriptions`
- `plans` one-to-many `subscriptions`
- `subscriptions` one-to-many `subscription_keys`
- `subscriptions` one-to-many `devices`
- `subscriptions` one-to-many `usage_counters`
- `subscriptions` one-to-many `traffic_rollups_daily`
- `nodes` many-to-many `node_groups` through `node_group_members`
- `subscriptions` many-to-many `nodes` through `subscription_node_assignments`
- `nodes` one-to-many `config_revisions`

## Data Handling Notes

- secrets must be stored encrypted at rest where applicable
- audit logs should be append-only
- key material must be versioned
- usage data should support both current counters and daily rollups

## Conservative Decisions

### Conservative choice: keep billing entities out of the schema

Billing tables are excluded from `MVP v0.1`.

Reason:

- They would force premature financial workflow design.
- Manual renewals and external webhook events are sufficient for the first release.

### Conservative choice: keep marketplace entities out of the schema

Marketplace provider and review records are excluded from the first release schema.

Reason:

- They introduce moderation and trust logic that is unrelated to the core provider operation path.
