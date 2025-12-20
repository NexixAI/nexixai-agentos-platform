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

- ✅ v1.02 PRS + Schemas Appendix committed
- ⏭️ Next: scaffold Stack A/Stack B/Federation + OpenAPI + conformance tests
