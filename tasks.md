# Tasks

## Phase 1: Modernize the Foundation

**Why:** Can't build on a foundation that uses deprecated tooling and has an untested core path.

- [ ] Introduce `go.mod`, remove `dep` (`Gopkg.toml`, `Gopkg.lock`)
- [ ] Restructure into layered package layout (`internal/model`, `internal/workflow`, `internal/store`, `cmd/server`)
- [ ] Replace dynamic table creation (`doctype.go:96-121` creates `wf_documents_NNN` per DocType) with a single `wf_documents` table
- [ ] Introduce a migration framework to replace raw DDL scripts in `sql/`
- [ ] Add tests for `ApplyEvent` — the core workflow execution path is completely untested

## Phase 2: Fix Correctness Gaps

**Why:** The engine must be correct before it's exposed — race conditions and missing quorum logic are showstoppers for payments.

- [ ] Add row-level locking in `applyEvent` to prevent concurrent approval races
- [ ] Implement `NodeTypeJoinAll` for quorum / all-must-approve patterns
- [ ] Enforce permissions inside `ApplyEvent` instead of trusting the caller
- [ ] Add idempotency keys to `DocEvent` to prevent duplicate event creation

## Phase 3: Make It a Service

**Why:** The repo is a library today; it needs a network API and safe trigger mechanism to become a deployable service.

- [ ] Add gRPC API layer with REST gateway (grpc-gateway)
- [ ] Add configuration management (env vars / config file)
- [ ] Add structured logging
- [ ] Add health and readiness endpoints
- [ ] Add post-commit side-effect mechanism (transactional outbox) — `NodeFunc` currently runs inside the DB transaction

## Phase 4: Payments-Domain Features

**Why:** Generic workflow engine → payments-specific approval service.

- [ ] Immutable audit trail with actor, timestamps, before/after state snapshots
- [ ] SLA timers and escalation for stuck approvals
- [ ] Conditional routing (amount thresholds, rule-based branching)
- [ ] Recall / cancel by requester with first-class support
- [ ] Webhook / notification push (email, Slack, webhooks beyond the in-app mailbox)

## Phase 5: Production Readiness

**Why:** Hardening before real money flows through it.

- [ ] Comprehensive test suite (unit, integration, concurrency stress tests)
- [ ] Observability (metrics, tracing)
- [ ] Rate limiting and input validation at API boundary
- [ ] CI/CD pipeline (linting, race detection, migration verification)
