
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
- Docker-network invariants for multi-service tests (no localhost from CI host)

---

## Repository Layout

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
	•	Stack A: http://localhost:8081/v1/health
	•	Stack B: http://localhost:8082/v1/health
	•	Federation: http://localhost:8083/v1/federation/health

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
