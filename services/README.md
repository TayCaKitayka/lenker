# Services

This directory contains backend runtime services for `Lenker MVP v0.1`.

Included services:

- `panel-api` — provider control plane in Go
- `node-agent` — managed node agent in Go

Out of scope at this stage:

- billing service
- marketplace service
- standalone auth service

The initial repository skeleton keeps services separate to match the control-plane and node-plane split defined in the architecture docs.
