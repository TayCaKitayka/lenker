# Contributing to Lenker

Lenker is early-stage. Contributions are welcome when they fit the current MVP
boundary and keep changes small enough to review carefully.

## Current Scope

`MVP v0.1` is focused on:

- provider panel backend;
- node agent foundation;
- users, plans, and subscriptions;
- node registration and heartbeat;
- PostgreSQL-backed control plane state;
- one protocol path: `VLESS + Reality + XTLS Vision`;
- Android, Windows, and macOS client targets later in the MVP.

Do not add the following without prior discussion:

- marketplace implementation;
- billing or payment processing;
- extra production protocols;
- broad node orchestration;
- production VPN runtime shortcuts;
- Telegram bot as a core module;
- enterprise-only systems.

## Checks

Run the main local check before opening a PR:

```sh
make test
```

This runs the current backend, node-agent, and OpenAPI checks.

## Issues

Good issues include:

- the problem being solved;
- the component affected;
- expected behavior;
- current behavior;
- reproduction steps or design context;
- whether the issue touches privacy, secrets, or security.

## Pull Requests

Keep PRs focused. A good PR should:

- stay inside the documented MVP scope;
- avoid unrelated refactors;
- include or update tests when behavior changes;
- update docs when public behavior changes;
- call out security or privacy impact explicitly.

Privacy and security changes need careful review. Avoid logging secrets,
subscription tokens, session tokens, DNS history, browsing history, or provider
private data.
