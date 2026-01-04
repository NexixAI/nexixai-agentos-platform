# Track 06 — Run Lifecycle Completion

## Goal
Implement `POST /v1/runs/{run_id}:cancel` endpoint as specified in PRS §4.2.

## What you are allowed to change
- `agentorchestrator/server.go` — Add cancel endpoint handler
- `internal/storage/run_store.go` — Add state transition validation
- `internal/types/stacka.go` — Add "canceling" and "canceled" states if needed
- OpenAPI spec for cancel endpoint
- Unit tests for cancellation

Do not change PRS (already specifies cancel endpoint).

## Required outcomes

1) **Cancel endpoint**
- `POST /v1/runs/{run_id}:cancel`
- Requires tenant context
- Returns updated run object with status "canceled"

2) **State transitions**
- Can cancel from: "queued", "running"
- Cannot cancel from: "completed", "failed", "canceled"
- Attempting to cancel completed run returns 409 Conflict

3) **Cancellation semantics (stub mode)**
- In v1.02 stub mode: simply update status to "canceled"
- Set `completed_at` timestamp when transitioning to "canceled"
- Decrement concurrent run quota if run was counted

4) **Tenant isolation**
- User can only cancel runs in their tenant
- Attempting to cancel run in different tenant returns 404 Not Found (not 403)

5) **Audit logging**
- Log cancellation action with tenant_id, principal_id, run_id

## Required gates
- Unit test: cancel queued run succeeds
- Unit test: cancel completed run returns 409
- Unit test: cancel decrements quota
- Unit test: cannot cancel run from different tenant
- Update OpenAPI with cancel endpoint
- Update validation suite smoke test to exercise cancel

## Deliverables
- Cancel endpoint implementation
- State transition validation
- Unit tests
- Updated OpenAPI spec
