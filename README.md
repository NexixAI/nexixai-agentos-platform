
# NexixAI AgentOS Platform

AgentOS is a **spec-first, multi-tenant, federated agent platform** built by **NexixAI**.

This repository is intentionally **documentation-led**. Specs are the source of truth;  
code exists to prove contracts and execution paths, not to outrun the design.

---

## What This Repo Is

- **Agent Orchestrator (Agent Core)** — run orchestration, event streaming, lifecycle
- **Model Policy (Model & Policy Runtime)** — model invocation, embeddings, policy
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
- Docker-network invariants for multi-service tests (no localhost from CI host)

---

## Control vs Worker Plane

- Model A baseline: one shared Agent Orchestrator + one shared Model Policy with logical tenancy enforced via `tenant_id`, policy/entitlements, quotas, scoped storage, and audit.
- Model B (per-tenant pairs) is optional deployment only.
- Control plane remains compute-light; workers run heavy models and emit events back to Agent Orchestrator. Workers never make orchestration or policy decisions.

```
Box 1 — Control Plane
  [Agent Orchestrator] [Model Policy] [Queue/Event Bus]
  [UI/API Gateway] [Shared artifact store (/shared)]
  [Optional small local LLM for planning/orchestration]

Box 2/3 — Worker Nodes (GPU)
  [Worker executors + tools + storage] -> emits events to Agent Orchestrator
  [GPU-pinned model endpoints: vLLM/TGI/Triton]
```

## Repository Layout

/README.md
/SPEC_AUTHORITY.md
/AGENTS.md

/docs/
/docs/api/ (OpenAPI specs)
/docs/design/v1.02/ (canonical design)
/docs/product/agentos-prs/v1.02/ (canonical PRS & schemas)
/docs/plan/v1.02/ (canonical plan docs: index, phases, tracks)
/docs/templates/

/cmd/agentos/
/agentorchestrator/
/modelpolicy/
/federation/
/tests/
/deploy/

---

## Getting Started (Local, Single Node)

### Prerequisites
- Docker + Docker Compose v2
- Go toolchain (only required to build the CLI)

### Build the CLI
```bash
go build -o agentos ./cmd/agentos
# Windows:
go build -o agentos.exe ./cmd/agentos

Bring up the platform

./agentos up
./agentos validate
./agentos status

	Default health endpoints
		•	Agent Orchestrator: http://127.0.0.1:50081/v1/health
		•	Model Policy: http://127.0.0.1:50082/v1/health
		•	Federation: http://127.0.0.1:50083/v1/federation/health
		•	Windows note: use `curl.exe -4 http://127.0.0.1:PORT/...` (PowerShell `curl` is Invoke-WebRequest). Ports in the 808x range can be reserved/excluded on Windows; local files use 5008x to avoid that.

	Optional workers (GPU-pinned placeholders; control plane stays CPU-bound):

```
COMPOSE_PROFILES=workers docker compose -f deploy/local/compose.yaml up -d --build
```

Tear down

./agentos nuke
./agentos nuke --hard   # also removes volumes (destructive)


⸻

Federation (2-Node Local)

Run the 2-node federation stack (keep running)

docker compose \
  -f deploy/local/compose.federation-2node.yaml \
  -f deploy/local/compose.federation-2node.ADDENDUM.yaml \
  up -d --build

Run federation E2E (CI-equivalent)

docker compose \
  -f deploy/local/compose.federation-2node.yaml \
  -f deploy/local/compose.federation-2node.ADDENDUM.yaml \
  up --build --abort-on-container-exit federation-e2e

Invariant:
All multi-service tests run inside Docker and use Docker DNS names
(e.g. nodea-federation, nodeb-federation).
CI must never curl localhost for live services.

Tear down federation

docker compose \
  -f deploy/local/compose.federation-2node.yaml \
  -f deploy/local/compose.federation-2node.ADDENDUM.yaml \
  down -v


⸻

CI Meaning

When CI is green:
	•	OpenAPI + schema conformance has passed
	•	Federation E2E has passed in Docker
	•	The platform is deployable in its current form

CI green does not imply production hardening beyond Phase 9.

⸻

Status

Version: v1.02
Execution: Phase 9 complete (hardening baseline)
CI: Conformance + Federation E2E passing
