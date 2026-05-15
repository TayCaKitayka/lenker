# ADR-008: Licensing and Project Governance

## Status
Accepted

## Context
Lenker is an early-stage open-source VPN ecosystem with a self-hosted provider
control plane, managed node-agent, and future commercial services around the
open-source core.

The project includes network-facing backend and agent components. A permissive
license would make it easier to create closed hosted forks of the core control
plane without sharing modifications back with users or the community.

The project also needs a basic governance layer because VPN infrastructure is
security-sensitive and because code licensing does not cover brand identity or
responsible disclosure expectations.

## Decision
The root repository license is `AGPL-3.0-only`.

This is a conservative choice:

- AGPL matches network-service software better than GPL-only licensing.
- `only` avoids automatically opting into future license versions before the
  project deliberately reviews them.
- The self-hosted core remains open-source.
- Closed SaaS forks of the backend/control-plane core are discouraged.

The Lenker name and future logo are handled separately through
[TRADEMARK.md](../../TRADEMARK.md). The code license allows forking under its
terms, but it does not allow forks to present themselves as official Lenker.

The project maintains a basic [SECURITY.md](../../SECURITY.md) because auth,
node registration, bootstrap trust material, config signing, secrets, logs, and
subscription keys are sensitive areas.

Marketplace, billing, commercial plugin implementation, white-label builds, and
enterprise features remain future scope and do not change `MVP v0.1`.

Future SDKs, protocol specifications, or integration libraries may receive a
more permissive license such as `Apache-2.0` later, but no such split is active
now.

## Consequences
- The repository has a clear default license from the first public stage.
- Hosted services and managed operations remain possible without closing the
  self-hosted core.
- Contributors and forks have clearer boundaries around brand use and security
  disclosure.
- Any future license split must be documented explicitly instead of assumed.
