# Goal: Async Workflow Orchestration Service

## What It Is

A general-purpose workflow orchestration service where users define workflows
as a graph of steps, execute them asynchronously, and track every state
transition.

## Core Concepts

- **Workflow Definition**: A DAG of steps with transitions between them.
  Defined once, instantiated many times.

- **Workflow Instance**: A running execution of a workflow definition.
  Has a current state, history of transitions, and pending steps.

- **Step**: A unit of work in the workflow. Types include:
  - **Approval**: Wait for a human to approve/reject
  - **API Call**: Make an HTTP request to an external service
  - **DB Operation**: Execute a database query/mutation
  - **Condition**: Branch based on data (amount > threshold, status == X)
  - **Parallel Fan-out**: Execute multiple steps concurrently, wait for all/any

- **State Machine**: Tracks where each instance is in the workflow graph.
  Persists state so execution can stop (waiting for async trigger) and resume.

- **Transition Log**: Every state change is recorded — who triggered it,
  when, what the data looked like before/after. Enables audit trail,
  debugging, replay.

## Architecture (Target)

```
┌─────────────────────────────────────────┐
│              API Layer                  │
│   gRPC + REST (grpc-gateway)            │
│   - Define workflows                    │
│   - Start instances                     │
│   - Trigger steps (approve, callback)   │
│   - Query state and history             │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│           Workflow Engine               │
│   - State machine execution             │
│   - Step dispatch (approval, API, DB)   │
│   - Transition validation               │
│   - Async resume on trigger             │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│            Store Layer                  │
│   - Workflow definitions                │
│   - Instance state                      │
│   - Transition history                  │
│   - Step execution records              │
│   MySQL + migrations                    │
└─────────────────────────────────────────┘
```

## Current State vs Goal

The codebase today (commit 4a209f1) is a Go library that handles
one step type (human approval) with a rigid document-centric model.
The path from here to the goal:

### Phase 1: Foundation (structural modernization)
Make the codebase ready for extension.
- ✅ Go modules (done)
- ✅ Consolidated document tables (done)
- Package restructure (internal/model, internal/store, internal/workflow)
- Migration framework (replace raw SQL with versioned migrations)

### Phase 2: Generalize the Engine
Transform from document-approval to general workflow.
- Implement JoinAll (quorum/all-must-approve)
- Add permission enforcement in workflow execution
- Add row-level locking for concurrent safety
- Add idempotency keys

### Phase 3: Service Layer
Turn the library into a deployable service.
- gRPC API with REST gateway
- Configuration management
- Structured logging
- Health/readiness endpoints
- Transactional outbox for side effects

### Phase 4: Step Types
Add step types beyond human approval.
- API call steps (HTTP callout with retry)
- DB operation steps
- Conditional branching
- Parallel fan-out (JoinAll becomes critical here)
- Timer/SLA steps

### Phase 5: Production Readiness
- Audit trail with before/after snapshots
- Observability (metrics, tracing)
- Rate limiting
- CI/CD pipeline
