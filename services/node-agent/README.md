# node-agent

`node-agent` is the future Go service that runs on managed Lenker nodes.

Planned responsibilities for `MVP v0.1`:

- one-time bootstrap registration
- `HTTPS + mTLS` trust establishment
- node health reporting
- basic metrics reporting
- signed config bundle retrieval
- atomic config apply and rollback
- drain mode support
- local Xray process control for the main protocol path

Not included here yet:

- real node runtime logic
- VPN configuration generation
- process supervision implementation
- support for protocols beyond the main MVP path
