# Tasks

See GOAL.md for the full vision: async workflow orchestration service.

## Phase 1: Modernize the Foundation

**Why:** Can't build a general-purpose workflow service on a flat single-package library with raw SQL scripts.

- [x] Introduce `go.mod`, remove `dep` (`Gopkg.toml`, `Gopkg.lock`) — **commit `4a209f1`** (SWE-bench task: workflow-as-service-table-consolidation)
- [x] Replace dynamic table creation (`doctype.go:96-121` creates `wf_documents_NNN` per DocType) with a single `wf_documents` table — **commit `4a209f1`**
- [x] Restructure into layered package layout (`internal/model`, `internal/store`, `internal/workflow`) with store interface — **oracle built locally** (SWE-bench task: pending submission)
- [ ] Introduce a migration framework to replace raw DDL scripts in `sql/`
- [ ] Add tests for `ApplyEvent` — the core workflow execution path is completely untested
- [ ] Rename Document-centric naming to generic workflow naming (Document → WorkItem, DocType → WorkItemType, DocEvent → StepEvent, etc.) — ad-hoc cleanup, not a SWE-bench task

## Phase 2: Generalize the Engine

**Why:** The engine currently only handles human approvals with a rigid document-centric model. Must generalize for multiple step types.

- [ ] Implement `NodeTypeJoinAll` for quorum / all-must-approve patterns (currently `// TODO(js)` in node.go)
- [ ] Add row-level locking in `applyEvent` to prevent concurrent approval races
- [ ] Enforce permissions inside `ApplyEvent` instead of trusting the caller
- [ ] Add idempotency keys to `DocEvent` to prevent duplicate event creation

## Phase 3: Service Layer

**Why:** The repo is a library today; it needs a network API and async execution model to become a deployable service.

- [ ] Add gRPC API layer with REST gateway (grpc-gateway)
- [ ] Add configuration management (env vars / config file)
- [ ] Add structured logging
- [ ] Add health and readiness endpoints
- [ ] Add transactional outbox for async side effects (NodeFunc currently runs inside the DB transaction)

## Phase 4: Step Types

**Why:** General workflow orchestration needs more than human approvals.

- [ ] API call steps — HTTP callout with retry, timeout, callback URL for async response
- [ ] DB operation steps — execute parameterized queries as workflow steps
- [ ] Conditional routing — branch based on data (amount thresholds, rule-based)
- [ ] Parallel fan-out — execute multiple steps concurrently, JoinAll/JoinAny for convergence
- [ ] Timer/SLA steps — escalation for stuck steps, deadline-based transitions

## Phase 5: Production Readiness

**Why:** Hardening before real workflows run through it.

- [ ] Immutable audit trail with actor, timestamps, before/after state snapshots
- [ ] Comprehensive test suite (unit, integration, concurrency stress tests)
- [ ] Observability (metrics, tracing)
- [ ] Rate limiting and input validation at API boundary
- [ ] CI/CD pipeline (linting, race detection, migration verification)
- [ ] Webhook / notification push (email, Slack, webhooks)
