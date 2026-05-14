# ADR-007: Open-Source Core and Commercial Services

## Status
Accepted

## Context
Lenker is intended to be an open-source VPN ecosystem for providers and users. The project can support commercial activity later, but the first release depends on provider trust, privacy, and a usable self-hosted core.

Commercial features around VPN infrastructure are especially sensitive because the system handles provider operations, subscription state, and user connection metadata. Monetization must not create incentives to sell private data, hide logs, manipulate marketplace ranking, or degrade the self-hosted version.

## Decision
Lenker monetization will be built around hosted services, managed operations, paid support, marketplace trust services, migration services, and enterprise governance.

The self-hosted core remains open-source and must stay usable for small providers. Paid offerings may reduce operational burden or provide SLA, governance, verification, and integration value, but they must not close the core control plane or make basic security and privacy paid-only.

`MVP v0.1` does not include marketplace, billing, commercial plugin implementation, white-label builds, enterprise SSO, or payment processing.

## Consequences
- Lenker Cloud can be designed later as a hosted operations product without changing the self-hosted MVP scope.
- Managed nodes and paid support can become commercial offerings without adding billing logic to the core runtime.
- Marketplace monetization must avoid pay-to-win ranking and keep paid placement clearly labeled.
- Billing plugins can be designed after the API and webhook boundaries are stable.
- Basic security, privacy protections, and self-host operations remain part of the open-source core.
- Product docs can discuss commercial direction while implementation remains limited to `MVP v0.1`.
