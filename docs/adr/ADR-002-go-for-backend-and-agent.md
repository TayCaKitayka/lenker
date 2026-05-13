# ADR-002: Go for Panel Backend and Node Agent

## Status
Accepted

## Context
The control plane and node agent both need long-running services, predictable operational behavior, and shared implementation patterns around transport, config delivery, and health handling.

## Decision
Use `Go` for both `services/panel-api` and `services/node-agent`.

## Consequences
- One systems language is used for both server-side services.
- Shared libraries for auth, config signing, transport, and audit can evolve in one ecosystem.
- The implementation avoids a mixed Go/Python operational model in `MVP v0.1`.
