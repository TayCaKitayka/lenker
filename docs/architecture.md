# Lenker Architecture for MVP v0.1

## Purpose

This document describes the system architecture for `Lenker MVP v0.1`. It is limited to the first release scope and excludes marketplace and billing subsystems.

## Technology Stack

- `panel backend`: Go
- `node agent`: Go
- `panel web`: React + TypeScript
- `client app`: Flutter
- `database`: PostgreSQL
- `panel-node transport`: HTTPS + mTLS

## System Components

### 1. Panel Backend

The panel backend is the control plane of the system.

Responsibilities:

- admin authentication and authorization
- RBAC enforcement
- users, plans, subscriptions, devices, and key lifecycle
- node registration and node lifecycle
- protocol preset storage
- config generation for Xray
- signed config bundle creation
- webhook ingest and delivery
- audit logging
- client-facing REST endpoints

The backend is the source of truth for provider configuration and subscription state.

### 2. Panel Web

The web panel is the operational UI for providers.

Responsibilities:

- display dashboard state
- manage users and subscriptions
- manage plans
- manage nodes
- manage protocol preset selection
- manage API tokens and webhooks
- review audit logs

The panel web talks only to the panel backend REST API.

### 3. Node Agent

The node agent runs on each managed VPN node.

Responsibilities:

- register with the panel using a one-time bootstrap flow
- maintain authenticated transport with the panel
- report node health and basic metrics
- receive signed config bundles
- validate and apply configs atomically
- restart Xray as required
- roll back to the last working config on failure
- expose local state required for operations
- support drain mode

The node agent must not require the panel to store SSH passwords after bootstrap.

### 4. VPN Core on Node

For `MVP v0.1`, the only required deployed core is:

- `Xray`

The deployed protocol path is:

- `VLESS + Reality + XTLS Vision`

The node agent manages local Xray configuration and lifecycle.

### 5. Client Application

The client app is the end-user surface.

Supported platforms:

- Android
- Windows
- macOS

Responsibilities:

- email-first sign-in
- provider onboarding via deeplink or provider code
- subscription sync
- connect and disconnect
- node or region selection
- basic diagnostics
- key rotation
- secure local storage

The client app consumes provider APIs exposed by the panel backend.

### 6. PostgreSQL Database

PostgreSQL stores:

- provider operational data
- users and subscriptions
- nodes and protocol metadata
- device associations
- audit records
- API tokens and webhook configuration
- config revision history

## High-Level Interaction Model

### Provider Control Flow

1. Admin uses the web panel.
2. Panel web calls the panel backend REST API.
3. Panel backend reads and writes state in PostgreSQL.
4. Panel backend generates signed config bundles.
5. Panel backend delivers config bundles to node agents over `HTTPS + mTLS`.
6. Node agents apply configs and report status back.

### User Connection Flow

1. User signs in from the client app using email-first auth.
2. Client app retrieves the user subscription and available node or region choices.
3. Client app requests the active subscription configuration metadata.
4. Client app connects using the provider-issued configuration.
5. Client app reports only the minimum app state required for subscription lifecycle operations.

## Trust Boundaries

### Boundary 1: Admin to Panel

- protected by admin authentication and RBAC
- sensitive actions require audit logging

### Boundary 2: Panel to Node Agent

- protected by `HTTPS + mTLS`
- bootstrap uses one-time registration material
- config bundles are signed

### Boundary 3: User App to Panel

- protected by standard HTTPS
- user auth is email-first
- session tokens must be scoped to app use

### Boundary 4: Secrets and Persistent State

- subscription secrets and node trust material must not be stored in plaintext where avoidable
- configuration revisions must be versioned

## Architectural Principles

- nodes are first-class resources
- one operational protocol path for the first release
- API-first design
- self-hosted provider model
- minimal logging by default
- rollback must be built into node config deployment

## Deployment Shape

### Minimum Supported Topology

`MVP v0.1` assumes a small-provider deployment:

- 1 panel backend instance
- 1 panel web deployment
- 1 PostgreSQL instance
- 1 or more node agents
- 1 or more Xray nodes

This is the minimum supported operational shape for the first release.

## Conservative Decisions

### Conservative choice: REST between panel web and backend

The UI uses REST only.

Reason:

- simpler to document and maintain in an open-source repository
- easier external API reuse

### Conservative choice: HTTPS + mTLS instead of gRPC for panel-node transport

`MVP v0.1` uses `HTTPS + mTLS`.

Reason:

- enough for the required operational flows
- easier transport debugging during early development
- lower integration overhead for the first release

### Conservative choice: single-node-core path

Only the Xray path is required in the first release.

Reason:

- avoids premature multi-core orchestration
- keeps rollback and health semantics tractable
