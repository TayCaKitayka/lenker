# ADR-006: HTTPS and mTLS for Panel-Node Transport

## Status
Accepted

## Context
The panel must communicate with managed node agents over an authenticated channel that supports registration, config delivery, health reporting, and rollback workflows.

## Decision
Use `HTTPS + mTLS` for panel-node transport in `MVP v0.1`.

## Consequences
- Transport is strong enough for the required control-plane operations.
- Debugging remains simpler than introducing gRPC in the first release.
- Bootstrap trust and certificate rotation must be designed carefully.
