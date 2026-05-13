# Lenker MVP v0.1 Specification

## Overview

Lenker is an open-source VPN ecosystem for providers and users. The first public milestone, `MVP v0.1`, is intentionally narrow. It focuses on the provider control plane, managed VPN nodes, user subscriptions, and a basic client application for `Android`, `Windows`, and `macOS`.

This release is designed to prove the core operational path:

1. A provider deploys the panel.
2. The provider adds nodes through a managed agent.
3. The provider creates plans and subscriptions.
4. The system issues a working subscription based on `VLESS + Reality + XTLS Vision`.
5. The user signs in with email, syncs their subscription, and connects from the client app.

## Product Goals

- Deliver a self-hosted provider panel for VPN operations.
- Make nodes a first-class managed resource.
- Support users, plans, subscriptions, devices, and key rotation.
- Provide a minimal but usable client application.
- Expose a public REST API from the first version.
- Keep security and operational safety ahead of feature breadth.

## In Scope

### Provider Panel

- Admin authentication
- Basic RBAC
- User management
- Plan management
- Subscription management
- Device list and device reset/revoke actions
- Node inventory and node lifecycle operations
- Protocol preset management for the single MVP protocol path
- API tokens and webhook configuration
- Audit log
- Basic dashboard with health and status summaries

### Node Management

- Node bootstrap via one-time registration flow
- Node registration over `HTTPS + mTLS`
- Node health reporting
- Basic system metrics reporting
- Signed config bundle delivery
- Atomic config apply
- Rollback after failed deploy
- Drain mode
- Xray process control required for the main protocol path

### Protocol Support

Only one production protocol path is included in `MVP v0.1`:

- `VLESS + Reality + XTLS Vision`

This is the only required deployable provider path and the only required client path in the first release.

### Client Application

Supported platforms:

- Android
- Windows
- macOS

Included capabilities:

- Email-first user authentication
- Provider deeplink or provider code onboarding
- Subscription sync
- Connect and disconnect
- Region selection
- Auto-select best node using a simple strategy
- Subscription status display
- Traffic and device summary
- Key rotation
- Basic diagnostics
- Encrypted local storage

### Public API

- REST API v1 for panel and client operations
- OpenAPI specification
- API tokens with scopes
- Webhook endpoints for external subscription lifecycle events

## Explicitly Out of Scope

The following items are not part of `MVP v0.1`:

- Marketplace flows
- Provider catalog inside the app
- Public provider rating site
- Built-in billing
- Payment processing
- Invoices, refunds, promo codes, and full financial records
- Telegram bot as a core component
- Migration from Marzban, 3x-ui, Hiddify Manager, or other panels
- White-label builds per provider
- iOS client
- Linux client
- Multi-protocol production support beyond the main protocol path
- Advanced routing UI
- Split tunneling UI
- Full analytics suite
- Support ticketing system
- Public status page

## Product Constraints

- `panel backend`: Go
- `node agent`: Go
- `panel web`: React + TypeScript
- `database`: PostgreSQL
- `panel-node transport`: HTTPS + mTLS
- `user auth`: email-first
- `client platforms`: Android, Windows, macOS only

## Core User Flows

### Provider Flow

1. Admin signs in to the provider panel.
2. Admin creates plans.
3. Admin creates users and subscriptions.
4. Admin bootstraps a node.
5. Node registers with the panel and receives trust material.
6. Panel deploys `VLESS + Reality + XTLS Vision` config.
7. Admin issues a subscription to the user.

### User Flow

1. User opens the client app.
2. User signs in via email-first flow.
3. User joins a provider via deeplink or provider code.
4. App fetches the subscription and available regions.
5. User taps connect.
6. App uses the assigned configuration and selected region.
7. If needed, user rotates the subscription key from the app.

## Conservative Decisions

### Conservative choice: one protocol path only

`MVP v0.1` supports only `VLESS + Reality + XTLS Vision`.

Reason:

- It keeps backend, agent, and client complexity under control.
- It reduces test matrix size.
- It avoids building a protocol abstraction layer that is broader than the first release needs.

### Conservative choice: email-first auth

The first release uses email as the primary authentication method for users.

Reason:

- It is simpler than mixing phone, Telegram, and OAuth in the first release.
- It keeps account recovery and identity flows more predictable.
- It avoids coupling MVP scope to messaging platform dependencies.

### Conservative choice: no marketplace in the first release

Marketplace is intentionally deferred.

Reason:

- It introduces moderation, ranking, verification, and trust problems that are separate from the core panel-and-connect path.
- It would slow down the delivery of the operational foundation.

### Conservative choice: no built-in billing

`MVP v0.1` supports manual renewal, API calls, and webhooks only.

Reason:

- Billing increases regulatory and operational scope.
- Providers can integrate billing later without blocking the first release.
