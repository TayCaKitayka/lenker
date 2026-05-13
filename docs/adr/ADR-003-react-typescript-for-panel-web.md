# ADR-003: React and TypeScript for Panel Web

## Status
Accepted

## Context
The provider panel needs a standalone operational UI that consumes the REST API exposed by the panel backend.

## Decision
Use `React + TypeScript` for `apps/panel-web`.

## Consequences
- The panel web can evolve independently from backend services.
- TypeScript improves API-facing UI maintainability.
- The repository keeps a conventional web app setup for future contributors.
