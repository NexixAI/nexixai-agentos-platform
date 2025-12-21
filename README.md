# NexixAI AgentOS Platform

**AgentOS** is NexixAI’s federated, multi-tenant agent platform with a two-stack architecture:

- **Stack A: Orchestration Stack** — agent lifecycle, run execution, tools/memory/events, federation
- **Stack B: Model Services Stack** — model access, routing, policy/governance, metering/budgets

This repo is **spec-first** right now: the authoritative requirements and contracts live under `docs/`.

## Specs (authoritative)

- **AgentOS v1.02 Product Requirements and Specifications (PRS)**  
  `docs/product/agentos-prs/v1.02-prs.md`

- **AgentOS v1.02 Schemas Appendix (Implementation Contracts)**  
  `docs/product/agentos-prs/v1.02-schemas-appendix.md`

## Working agreement

- `/v1` APIs are **additive-only**.
- Tenant context is explicit and enforced end-to-end.
- Federation contracts are treated as public, versioned interfaces.

## Status

- ✅ Phase 0–1: v1.02 PRS + Schemas Appendix + Design doc
- ✅ Phase 2: OpenAPI for Stack A, Stack B, Federation
- ✅ Phase 3: Canonical JSON examples under `docs/api/**/examples/`
- ✅ Phase 4: Conformance runner + GitHub Actions workflow
- ✅ Phase 5: Minimal Go servers + handler stubs (schema-conformant responses)
- ✅ Phase 6: `agentos` CLI (`up|redeploy|validate|status|nuke`)
- ✅ Phase 7: Multi-tenancy + quotas + audit logging (baseline)
- ✅ Phase 8: Federation v1 (2-node compose + SSE proxy + e2e workflow)
- ✅ Phase 9: Hardening add-ons (alerting + edge rate limiting)

