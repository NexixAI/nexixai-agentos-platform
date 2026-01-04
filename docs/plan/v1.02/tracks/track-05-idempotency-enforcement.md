# Track 05 — Idempotency Enforcement

## Goal
Enforce idempotency key deduplication for run creation as specified in PRS §4.3.

## What you are allowed to change
- `internal/storage/run_store.go` — Add idempotency key lookup
- `agentorchestrator/server.go` — Check idempotency before creating run
- `internal/types/stacka.go` — Ensure IdempotencyKey field present in storage
- Unit tests for idempotency logic
- OpenAPI examples demonstrating idempotency behavior

Do not change normative PRS (already updated) or Schemas Appendix (idempotency_key already defined).

## Required outcomes

1) **Idempotency key storage**
- Store mapping: `tenant/{tenant_id}/idempotency/{idempotency_key}` → `run_id`
- Key format: `idem_{ulid}` or user-provided string (max 256 chars)
- TTL: 24 hours (configurable via `AGENTOS_IDEMPOTENCY_TTL_HOURS`)

2) **Deduplication logic**
- If `idempotency_key` provided in `RunCreateRequest`:
  - Check if key exists for tenant
  - If exists: return existing run (200 OK, not 201 Created)
  - If not exists: create run and store key mapping
- If `idempotency_key` not provided: proceed normally (no deduplication)

3) **Tenant scoping**
- Idempotency keys scoped per tenant
- Same key for different tenants creates separate runs

4) **Error handling**
- If idempotency key maps to run in different state than expected: return existing run anyway
- If storage lookup fails: log error and proceed with creation (fail-open for availability)

## Required gates
- Unit tests: idempotency key enforced across retries
- Unit tests: different tenants can use same idempotency key
- Smoke test: create run twice with same key returns same run_id
- Update validation suite to test idempotency

## Deliverables
- Updated `RunStore` interface with `GetByIdempotencyKey` method
- Updated file-based storage implementation
- Unit tests
- OpenAPI example showing idempotency behavior
