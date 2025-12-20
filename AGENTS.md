# AGENTS — Implementation Rules (Codex / Automation)

This repo is **spec-first**. The goal is to implement AgentOS without interpretation drift.

## Authority (must follow)
Read and follow `SPEC_AUTHORITY.md`. If specs conflict, follow the precedence order.

## Output format (critical)
- Prefer producing **zip archives** of file changes when working outside a PR workflow.
- When editing in-repo, make **minimal diffs** and keep changes reviewable.

## README rules (DO NOT OVERWRITE)
- Never rewrite `README.md` or `docs/README.md` from scratch.
- Preserve tone/sections; edit only impacted sections.
- Keep the “Specs (authoritative)” links intact.

## Spec discipline
- `/v1` APIs are **additive-only**.
- Do not invent new fields or semantics in code.
- If something is ambiguous, update the spec docs first (Appendix/OpenAPI/PRS/Design), then implement.

## Multi-tenancy invariants (non-negotiable)
- Every request executes within exactly one `tenant_id`.
- `auth.tenant_id` is required on all internal port calls.
- No cross-tenant reads/writes/events; federation must preserve and enforce tenant context.

## Implementation boundaries
- Stack A owns: Runs, Events, Tools/Memory ports, Scheduling, Federation client/server.
- Stack B owns: Model providers, Routing, Policy checks, Usage metering, Budgets/Quotas for model calls.

## “Minimal diff” checklist for any change
- Update only the necessary files.
- If APIs change: update OpenAPI + examples + docs together.
- Include: short notes on what changed and why.
## Execution plan editing rule (DO NOT BREAK)
- Never overwrite `docs/plan/agentos-v1.02-execution-plan.md` with a shortened template.
- Only update the **Current status** section, and append entries to `docs/plan/progress-log.md`.
