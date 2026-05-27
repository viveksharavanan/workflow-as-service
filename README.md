# Workflow as Service

Payments approval workflow engine — request, N levels of approval, trigger a downstream action. Go, MySQL, gRPC + REST.

## Background

Seeded from [js-ojus/flow](https://github.com/js-ojus/flow) (a minimal Apache 2-licensed Go workflow library with a state-machine engine, RBAC, and mailbox notifications — archived, single author, ~6K LOC, no HTTP layer).

What's changing from js-ojus/flow (nearly everything):
- **Library → Service** — adding gRPC API with REST gateway; the original is a Go library with no network layer
- **Foundation modernized** — `dep` → Go modules; flat single-package → layered package layout; dynamic per-DocType tables → single documents table; raw DDL scripts → versioned migrations
- **Correctness fixed** — adding row-level locking (original has no concurrency protection), implementing `NodeTypeJoinAll` (declared but stubbed as TODO), enforcing permissions inside `ApplyEvent` (original trusts the caller), adding idempotency keys
- **Payments domain added** — immutable audit trail, SLA timers and escalation, conditional routing (amount thresholds), recall/cancel, webhook notifications
- **Test coverage built** — original tests cover only CRUD; the core workflow path (`ApplyEvent`) is completely untested

## Setup

Prerequisites: Go 1.22+, MySQL 8.0+

```bash
# TODO — will be updated as the service layer is built
```

## License

Apache 2.0 — see [LICENSE](LICENSE).
