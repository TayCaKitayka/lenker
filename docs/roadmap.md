# Lenker MVP v0.1 Roadmap

## Purpose

This roadmap defines the delivery plan for `Lenker MVP v0.1`. It is intentionally constrained to the first release and excludes marketplace and billing delivery.

## Phase 1. Foundation and Control Plane

Goal:

Establish the core provider control plane and the first stable data model.

Scope:

- repository structure and contribution docs
- architecture decision records
- threat model draft
- PostgreSQL schema for MVP entities
- panel backend skeleton in Go
- panel web skeleton in React + TypeScript
- admin authentication
- basic RBAC
- users API
- plans API
- subscriptions API
- audit log
- API tokens
- webhook configuration

Exit criteria:

- admin can sign in
- provider can create users, plans, and subscriptions
- audit logging exists for sensitive panel actions
- API surface is documented for implemented resources

## Phase 2. Nodes and Configuration Delivery

Goal:

Make nodes manageable and able to receive deployable VPN configuration safely.

Scope:

- node agent skeleton in Go
- one-time bootstrap token flow
- node registration over `HTTPS + mTLS`
- node inventory model
- health reporting
- basic node metrics
- Xray config generation for `VLESS + Reality + XTLS Vision`
- signed config bundle delivery
- atomic apply
- rollback on failed deploy
- drain mode
- node logs access with RBAC limits

Exit criteria:

- provider can register a node
- provider can deploy a working config
- failed deploy can roll back automatically
- node status is visible in the panel

## Phase 3. Client App MVP

Goal:

Provide the first end-user application flow for provider login and connection.

Scope:

- Flutter app shell
- Android, Windows, and macOS targets
- email-first user auth
- provider deeplink or provider code onboarding
- provider branding fetch
- subscription sync
- connect and disconnect flow
- region selection
- simple auto-best node mode
- traffic and subscription summary
- key rotation
- basic diagnostics
- encrypted local storage

Exit criteria:

- user can sign in with email
- user can fetch an active subscription
- user can connect through the main protocol path
- user can rotate a key without provider support

## Phase 4. Hardening and Release Readiness

Goal:

Stabilize the system for the first public release.

Scope:

- backup and restore path
- release packaging for Android, Windows, macOS
- config revision visibility in the panel
- webhook delivery retries
- usage accounting validation
- structured warning and error logging defaults
- deployment documentation
- security policy
- responsible disclosure process
- dynamic branding validation for provider mode

Exit criteria:

- system can be deployed and operated by an early provider
- release artifacts exist for all target client platforms
- operational recovery basics are documented
- default logs stay within the privacy-first MVP boundary

## Deferred After MVP v0.1

These topics are intentionally outside the four MVP phases:

- in-app marketplace
- provider verification workflow
- ratings and reviews
- public metrics and ranking
- built-in billing
- payment adapters
- Telegram bot plugin
- iOS support
- additional production protocol paths

## Post-MVP Business and Monetization Milestones

These milestones are business planning items for after the self-hosted core is working. They must not be added to `MVP v0.1` implementation scope.

- Lenker Cloud architecture draft for hosted panel operations
- managed nodes operations model
- paid support policy and support boundaries
- marketplace verification and governance policy
- provider manifest signing model for future marketplace trust
- dynamic branding model before any white-label build work
- enterprise governance scope, including advanced RBAC, audit retention, SSO, and SLA boundaries
- billing plugin interface draft based on existing API and webhook boundaries
- migration services plan for Marzban, 3x-ui, Hiddify Manager, and custom panels
- trademark and naming policy for official Lenker services and builds

Conservative note:

Commercial work should start around hosting, support, and managed operations after the open-source self-host core is usable. Marketplace ranking, payment processing, commission flow, white-label builds, enterprise SSO, and full billing remain outside the nearest implementation scope.

## Conservative Decisions

### Conservative choice: no marketplace phase inside MVP v0.1

Marketplace is deferred until after the first release.

Reason:

- It is a separate trust and moderation problem.
- It does not block the core provider-to-user connection path.

### Conservative choice: no billing phase inside MVP v0.1

Billing is deferred until after the first release.

Reason:

- External integrations and manual renewals are enough to validate the core platform.

### Conservative choice: one protocol path through all phases

All phases assume the same production path:

- `VLESS + Reality + XTLS Vision`

Reason:

- It keeps operational complexity bounded.
- It avoids multiplying backend, agent, and client integration work before the base system is stable.
