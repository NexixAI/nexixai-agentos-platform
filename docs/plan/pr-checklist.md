# PR Checklist — AgentOS v1.02 (Spec-first)

Use this checklist for any PR (or local commit series) that changes behavior, APIs, docs, or ops.

## Spec alignment
- [ ] If payloads changed: Schemas Appendix updated first (or confirmed no change)
- [ ] OpenAPI updated to match Schemas Appendix
- [ ] Canonical examples updated (if endpoint touched)
- [ ] Any new optional fields are additive-only (/v1)

## Multi-tenancy
- [ ] Tenant scoping preserved (`tenant_id` present in auth context and objects)
- [ ] No cross-tenant reads/writes/events introduced
- [ ] Quota/budget enforcement behavior documented (if affected)

## Observability
- [ ] Logs include `tenant_id`, `run_id`, `correlation_id` where relevant
- [ ] Metrics/tracing changes documented (if any)
- [ ] Alerts updated if new failure modes introduced

## Docs hygiene
- [ ] `README.md` and/or `docs/README.md` updated **without overwriting**
- [ ] Plan/progress updated:
  - [ ] `docs/plan/agentos-v1.02-execution-plan.md` current status updated OR
  - [ ] new entry added to `docs/plan/progress-log.md`

## Safety
- [ ] Destructive commands are gated (`--yes-really`) and documented (if applicable)
