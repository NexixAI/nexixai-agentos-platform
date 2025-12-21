# NexixAI AgentOS Platform

AgentOS is a **spec-first, multi-tenant, federated agent platform** built by **NexixAI**.

This repository is intentionally **documentation-led**. Specs are the source of truth;
code exists to prove contracts and execution paths, not to outrun the design.

---

## What This Repo Is

- **Stack A (Agent Core)** — run orchestration, event streaming, lifecycle
- **Stack B (Model & Policy Runtime)** — model invocation, embeddings, policy
- **Federation Layer** — connect AgentOS nodes across environments
- **Operator UX** — one-command deploy, validate, redeploy, nuke
- **Spec Authority** — PRS, schemas, OpenAPI, conformance tests

---

## Design Principles

- Spec-first, code-second
- Additive-only `/v1` APIs
- Multi-tenant by default
- Federation as a first-class primitive
- CI-enforced contracts

---

## Repository Layout

```
/README.md
/SPEC_AUTHORITY.md
/AGENTS.md

/docs/
  api/
  design/
  product/
  plan/
  templates/

/cmd/agentos/
/stack-a/
/stack-b/
/federation/
/tests/
/deploy/
```

---

## Getting Started

```bash
agentos up
agentos validate
agentos status
```

---

## Status

**Version:** v1.02  
**Execution:** Phase 9 complete (hardening baseline)