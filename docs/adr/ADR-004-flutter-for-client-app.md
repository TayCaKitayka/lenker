# ADR-004: Flutter for Client App

## Status
Accepted

## Context
`MVP v0.1` targets `Android`, `Windows`, and `macOS` from a single product surface.

## Decision
Use `Flutter` for `apps/client-app`.

## Consequences
- The app shell can target all required MVP platforms from one codebase.
- Platform-specific VPN core integration will still require careful bridge design.
- iOS and Linux remain out of scope for the first release.
