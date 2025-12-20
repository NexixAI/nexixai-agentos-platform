# NexixAI AgentOS Docs

This folder contains the authoritative product requirements and implementation contracts for **NexixAI AgentOS**.

## Documents

- **Product Requirements & Specifications (PRS)**  
  - `docs/product/agentos-prs/v1.02-prs.md`

- **Schemas Appendix (Implementation Contracts)**  
  - `docs/product/agentos-prs/v1.02-schemas-appendix.md`

## Naming (v1.02)

AgentOS is composed of two stacks per node/environment:

- **Stack A: Orchestration Stack** — agent lifecycle, run execution, tools, memory, events, federation
- **Stack B: Model Services Stack** — model access, routing, policy/governance, metering, budgets

## How to change the spec

1. Update the PRS (`v1.02-prs.md`) first.
2. Update the Schemas Appendix (`v1.02-schemas-appendix.md`) to match.
3. Bump the version (e.g., v1.03) when changes are accepted.

## Contribution convention

- Keep `/v1` APIs additive-only.
- Prefer explicit, verifiable tenant context everywhere (multi-tenancy).
- Treat federation contracts as public interfaces (versioned, negotiated capabilities).
