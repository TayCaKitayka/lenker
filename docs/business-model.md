# Lenker Business Model Boundary

## Purpose

This document defines the commercial boundary for Lenker.

It does not expand `MVP v0.1`, does not introduce billing or marketplace implementation work, and does not change the self-hosted open-source core.

## Principles

- The self-hosted core must remain a real, usable product.
- Paid features must not break or artificially limit the self-hosted core.
- Commercial work should focus on hosting, support, managed operations, trust, and business workflows.
- Privacy and provider trust are product requirements, not upsell surfaces.
- Marketplace and billing are explicitly outside `MVP v0.1`.

## Open-Source Core

The following areas should remain open-source core capabilities:

- self-hosted provider panel
- node agent
- users, plans, and subscriptions
- time, device, and traffic limits
- basic REST API
- basic web panel
- basic client app flow
- `VLESS + Reality + XTLS Vision` production path
- basic node registration and health checks
- config deploy and rollback foundation
- manual subscription renewal
- webhooks and public API integration points
- audit log foundation
- deployment and operations documentation

Conservative decision:

The open-source core should be useful for a small provider running Lenker independently. Paid offerings may improve operations, support, and distribution, but must not turn self-hosting into a broken demo.

## Paid, Hosted, or Enterprise Areas

### Lenker Cloud

Lenker Cloud can be a hosted provider panel with managed operations:

- hosted panel API and web panel
- managed PostgreSQL
- updates and backups
- uptime monitoring
- secret handling
- managed API endpoint
- basic operational support
- higher tiers based on users, nodes, retention, or support level

### Managed Nodes

Managed nodes can be sold as an operations service:

- node provisioning
- node-agent and VPN core updates
- monitoring and alerting
- diagnostics
- drain and maintenance operations
- emergency rollback help
- capacity planning
- multi-region setup assistance

### Paid Support

Open-source support remains community-based through public project channels. Paid support can provide:

- faster response times
- production setup help
- installation support
- diagnostics
- security review
- migration help
- private provider consultation
- update and incident assistance

### In-App Marketplace Monetization

The marketplace belongs inside the client app, but not in `MVP v0.1`.

Later monetization can include:

- provider verification
- provider onboarding
- fraud review
- dispute handling
- optional featured placement
- referral or commission models
- provider analytics

Featured placement must be clearly labeled and must not be mixed into organic ranking.

### Dynamic Branding and White-Label Later

Dynamic branding can become a paid provider feature after the core client flow is stable:

- provider profile
- logo and colors
- support links
- custom domain
- branded deeplink
- private provider mode

White-label app builds are later than dynamic branding and should not be part of `MVP v0.1`.

### Enterprise Features

Enterprise monetization can focus on governance and operational guarantees:

- SSO/SAML/OIDC for admins
- advanced RBAC
- audit log retention
- compliance exports
- dedicated deployments
- custom SLA
- private support channels
- IP allowlists
- approval workflows for sensitive actions
- disaster recovery support

Basic security must remain in the open-source core. Enterprise features should add governance, compliance, and SLA value, not remove safety from self-hosted deployments.

### Billing Plugins Later

Built-in billing is not part of `MVP v0.1`.

Later commercial work can include:

- payment provider adapters
- invoice workflows
- promo or referral mechanics
- subscription lifecycle automation
- hosted billing bridge
- custom provider billing integrations

The core should expose APIs and webhooks. Commercial billing plugins can live outside the core runtime.

### Migration Services Later

Migration can become a paid service after the core panel is stable:

- Marzban to Lenker
- 3x-ui to Lenker
- Hiddify Manager to Lenker
- custom panel to Lenker
- user, subscription, node, and config migration
- staging migration and validation

Migration tooling and services are not required for `MVP v0.1`.

## Not Monetizable

Lenker must not monetize:

- user data
- DNS history
- browsing history
- connection logs
- provider logs
- hidden telemetry
- private keys, tokens, or secrets
- pay-to-win marketplace ranking
- artificial breakage of self-hosted deployments
- basic security updates
- basic privacy protections

Pay-to-win ranking is especially excluded. Paid placement may exist only as clearly labeled advertising or featured placement, separate from organic ranking.

## MVP v0.1 Boundary

`MVP v0.1` remains focused on:

- provider panel
- node agent
- users, plans, and subscriptions
- PostgreSQL-backed control plane
- Android, Windows, and macOS client app
- `VLESS + Reality + XTLS Vision`
- manual renewal, APIs, and webhooks

The following are still outside `MVP v0.1`:

- marketplace implementation
- billing implementation
- commercial plugin implementation
- white-label builds
- enterprise SSO
- payment processing
- commission flow
- provider ranking systems
