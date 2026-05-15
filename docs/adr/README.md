# Architecture Decision Records

This directory stores Architecture Decision Records for `Lenker MVP v0.1` and later phases.

Recommended use:

- one file per decision
- sequential numbering
- clear statement of context, decision, and consequences

Current ADRs:

- ADR-001 monorepo structure
- ADR-002 Go for panel backend and node agent
- ADR-003 React + TypeScript for panel web
- ADR-004 Flutter for client app
- ADR-005 PostgreSQL as primary database
- ADR-006 HTTPS + mTLS for panel-node transport
- ADR-007 open-source core and commercial services
- ADR-008 licensing and project governance

Conservative note:

ADRs in the first stage should not expand implementation beyond the fixed `MVP v0.1` scope. Post-MVP marketplace, billing, and commercial-service decisions may be documented as boundaries, but should not introduce code work into the first release.
