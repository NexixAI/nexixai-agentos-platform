# Phase 7 — Multi-tenancy, quotas, auditing

Phase 7 makes multi-tenancy *enforced* (not a comment in a doc).

## Multi-tenancy enforcement

- Requests derive an **AuthContext** from headers.
- A **tenant_id** is required for tenant-scoped endpoints.
  - For local/dev “works out of the box”, `AGENTOS_DEFAULT_TENANT` may supply a default.
- Stack A run storage is keyed as `(tenant_id, run_id)`.
  - Cross-tenant access does not leak existence (returns 404).

## Quotas

In-memory per-tenant gates (Phase 7 scaffold):

- Stack A:
  - `AGENTOS_QUOTA_RUN_CREATE_QPS` (default: 10)
  - `AGENTOS_QUOTA_CONCURRENT_RUNS` (default: 25)
- Stack B:
  - `AGENTOS_QUOTA_INVOKE_QPS` (default: 20)

Exceeding a quota returns:

- HTTP 429
- `error.code = quota_exceeded`

## Audit logging

Each mutating action emits an audit JSON line:

- `runs.create`
- `models.invoke`
- `policy.check`
- federation `runs.forward` / `events.ingest`

Configure sink:

- `AGENTOS_AUDIT_SINK=stdout` (default)
- `AGENTOS_AUDIT_SINK=stderr`
- `AGENTOS_AUDIT_SINK=file:/path/to/audit.log`

## Operator flow

```bash
go run ./cmd/agentos up --tenant tnt_demo --principal prn_local
go run ./cmd/agentos validate --tenant tnt_demo --principal prn_local
```

Updated: 2025-12-21
