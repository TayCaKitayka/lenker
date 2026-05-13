# ADR-005: PostgreSQL as Primary Database

## Status
Accepted

## Context
Lenker needs a single source of truth for provider state, users, subscriptions, nodes, devices, audit records, and config history.

## Decision
Use `PostgreSQL` as the primary database for `MVP v0.1`.

## Consequences
- Relational modeling fits the current domain well.
- Migrations can be versioned alongside the monorepo.
- Marketplace and billing tables remain out of scope for the initial schema.
