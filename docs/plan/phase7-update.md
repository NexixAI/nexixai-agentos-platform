# Phase 7 Update — Multi-tenancy + quotas + auditing

Date: 2025-12-21

This file is additive. It exists to keep the plan honest without rewriting the full execution plan doc.

## What changed in Phase 7

- Added `internal/auth` + middleware to derive and propagate AuthContext.
- Enforced tenant scoping on Stack A/Stack B/Federation (tenant required unless `AGENTOS_DEFAULT_TENANT` is set).
- Added in-memory per-tenant quota limiting (QPS + concurrent runs).
- Added audit logging (JSONL) for mutating actions.

## Local defaults

`deploy/local/compose.yaml` now sets:
- `AGENTOS_DEFAULT_TENANT=tnt_demo`
- quota env vars
- `AGENTOS_AUDIT_SINK=stdout`
