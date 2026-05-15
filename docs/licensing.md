# Lenker Licensing

## Current License

Lenker is currently licensed under the root [LICENSE](../LICENSE):

- `AGPL-3.0-only`

Unless a file or directory explicitly states otherwise, this license applies to
the entire repository.

## Why AGPL

Lenker includes network-facing control-plane software: `panel-api`,
`node-agent`, API contracts, and operational tooling. AGPL is a conservative
fit for this shape because it keeps modified network-service versions tied back
to source availability.

This supports the project boundary defined in
[business-model.md](business-model.md):

- the self-hosted core remains open-source and usable;
- commercial work can happen around hosting, support, managed operations, and
  enterprise services;
- closed SaaS forks of the core control plane are discouraged.

## Future License Boundaries

Future SDKs, protocol specifications, or client integration libraries may use a
more permissive license such as `Apache-2.0` if that helps adoption. No such
split exists today.

Commercial plugins, Lenker Cloud, managed services, or hosted support offerings
may have separate terms later. The current MVP does not contain a commercial
plugin implementation, billing implementation, or marketplace implementation.

## Not Legal Advice

This document explains the current project intent. It is not legal advice. If
you need legal certainty for production, redistribution, app-store publishing,
or commercial use, consult qualified counsel.
