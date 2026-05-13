# ADR-001: Monorepo Structure

## Status
Accepted

## Context
Lenker `MVP v0.1` includes a provider control plane, a node agent, a web panel, and a client application. These parts share product scope, API contracts, and release coordination.

## Decision
Use a single monorepo with top-level directories for `apps/`, `services/`, `docs/`, and `migrations/`.

## Consequences
- Shared documentation and contracts stay close to implementation.
- Cross-component changes are easier to coordinate.
- Repository tooling can stay simple in the first release.
- Team boundaries will need discipline as the codebase grows.
