# NexixAI AgentOS Docs

This folder contains the authoritative product requirements and implementation contracts for **NexixAI AgentOS**.

## Documents

- **Product Requirements & Specifications (PRS)**  
  - `docs/product/agentos-prs/v1.02/prs.md`

- **Schemas Appendix (Implementation Contracts)**  
  - `docs/product/agentos-prs/v1.02/schemas-appendix.md`

## Naming (v1.02)

AgentOS control plane is composed of two services per environment:

- **Agent Orchestrator** — agent lifecycle, run execution, tools, memory, events, federation
- **Model Policy** — model access, routing, policy/governance, metering, budgets

Baseline tenancy model: Model A (shared control-plane services with logical isolation). Model B (per-tenant pairs) is optional deployment only.

## Change process

1. Update the PRS (`docs/product/agentos-prs/v1.02/prs.md`) first.
2. Update the Schemas Appendix (`docs/product/agentos-prs/v1.02/schemas-appendix.md`) to match.
3. Bump the version (e.g., v1.03) when changes are accepted.
