
# NexixAI AgentOS Platform

AgentOS is a **spec-first, multi-tenant, federated agent platform** built by **NexixAI**.

This repository is intentionally **documentation-led**. Specs are the source of truth;
code exists to prove contracts and execution paths, not to outrun the design.

---

## What This Repo Is

- **Agent Orchestrator (Orchestration)** — run orchestration, event streaming, lifecycle
- **Model Policy (Model Services)** — model invocation, embeddings, policy
- **Federation Layer** — connect AgentOS nodes across environments
- **Operator UX** — one-command deploy, validate, redeploy, nuke
- **Spec Authority** — PRS, schemas, OpenAPI, conformance tests
- **Shared Internal Packages** — auth, quota, storage, metrics, audit

---

## Design Principles

- Spec-first, code-second
- Additive-only `/v1` APIs
- Multi-tenant by default
- Federation as a first-class primitive
- CI-enforced contracts
- Docker-network invariants for multi-service tests (no localhost from CI host)
- RFC-0001 AI JCL discipline for AI-assisted development

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

```
/README.md
/SPEC_AUTHORITY.md
/AI_CONTRACT.md
/AGENTS.md

/docs/
/docs/api/                              (OpenAPI specs)
/docs/design/v1.02/                     (canonical design)
/docs/product/agentos-prs/v1.02/        (canonical PRS & schemas)
/docs/plan/v1.02/                       (canonical plan docs: index, phases, tracks)
/docs/rfc/                              (RFCs including AI JCL)

/cmd/agentos/                           (CLI entry point)
/agentorchestrator/                     (Agent Orchestrator service)
/modelpolicy/                           (Model Policy service)
/federation/                            (Federation service)
/internal/                              (shared packages: auth, quota, storage, metrics, audit, etc.)
/tests/                                 (conformance + E2E tests)
/deploy/                                (Docker Compose configs)
/configs/                               (phase tracking)
/.github/                               (CI workflows)
```

---

## Key Documentation

- **Product Requirements**: `docs/product/agentos-prs/v1.02/prs.md`
- **JSON Schemas**: `docs/product/agentos-prs/v1.02/schemas-appendix.md`
- **Design**: `docs/design/v1.02/agentos-design.md`
- **Spec Authority**: `SPEC_AUTHORITY.md`
- **OpenAPI Specs**: `docs/api/*/openapi.yaml`

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
```

### Bring Up the Platform

```bash
./agentos up
./agentos validate
./agentos status
```

### Default Health Endpoints

- Agent Orchestrator: http://127.0.0.1:50081/v1/health
- Model Policy: http://127.0.0.1:50082/v1/health
- Federation: http://127.0.0.1:50083/v1/federation/health

**Windows note**: use `curl.exe -4 http://127.0.0.1:PORT/...` (PowerShell `curl` is Invoke-WebRequest). Ports in the 808x range can be reserved/excluded on Windows; local files use 5008x to avoid that.

### Quick Smoke Test

**Automated (recommended):**

```bash
# Unix/macOS
./scripts/smoke-local.sh

# Windows (PowerShell)
.\scripts\smoke-local.ps1
```

The smoke scripts include retries with bounded timeouts and force IPv4 for reliability.

**Manual:**

```bash
# Create a run (uses default tenant tnt_demo)
curl -X POST http://127.0.0.1:50081/v1/agents/my-agent/runs \
  -H "X-Tenant-Id: tnt_demo" \
  -H "Content-Type: application/json" \
  -d '{"input": "hello"}'

# Check health
curl http://127.0.0.1:50081/v1/health

# List models
curl -H "X-Tenant-Id: tnt_demo" http://127.0.0.1:50082/v1/models
```

### Optional Workers

GPU-pinned placeholders; control plane stays CPU-bound:

```bash
COMPOSE_PROFILES=workers docker compose -f deploy/local/compose.yaml up -d --build
```

### Tear Down

```bash
./agentos nuke
./agentos nuke --hard   # also removes volumes (destructive)
```

---

## Federation (2-Node Local)

### Run the 2-node federation stack

Keep running:

```bash
docker compose \
  -f deploy/local/compose.federation-2node.yaml \
  -f deploy/local/compose.federation-2node.ADDENDUM.yaml \
  up -d --build
```

### Run federation E2E (CI-equivalent)

```bash
docker compose \
  -f deploy/local/compose.federation-2node.yaml \
  -f deploy/local/compose.federation-2node.ADDENDUM.yaml \
  up --build --abort-on-container-exit federation-e2e
```

**Invariant**: All multi-service tests run inside Docker and use Docker DNS names (e.g. nodea-federation, nodeb-federation). CI must never curl localhost for live services.

### Tear down federation

```bash
docker compose \
  -f deploy/local/compose.federation-2node.yaml \
  -f deploy/local/compose.federation-2node.ADDENDUM.yaml \
  down -v
```

### Federation Security (mTLS + JWT)

Federation supports mTLS and JWT verification for production deployments.

**Local Development (generate self-signed certs):**

```bash
./scripts/generate-federation-certs.sh
```

This generates CA, server, and client certificates in `./certs/federation/`.

**Enable mTLS:**

```bash
# Server configuration
export AGENTOS_FED_REQUIRE_MTLS=true
export AGENTOS_FED_SERVER_CERT_FILE=./certs/federation/server-cert.pem
export AGENTOS_FED_SERVER_KEY_FILE=./certs/federation/server-key.pem
export AGENTOS_FED_CA_CERT_FILE=./certs/federation/ca-cert.pem

# Client configuration (for forwarder)
export AGENTOS_FED_CLIENT_CERT_FILE=./certs/federation/client-cert.pem
export AGENTOS_FED_CLIENT_KEY_FILE=./certs/federation/client-key.pem
```

**Enable JWT Verification:**

```bash
# Generate RSA key pair for JWT signing
ssh-keygen -t rsa -b 4096 -m PEM -f ./certs/federation/jwt-key -N ""
openssl rsa -in ./certs/federation/jwt-key -pubout -out ./certs/federation/jwt-key.pub

# Configure JWT verification
export AGENTOS_FED_JWT_PUBLIC_KEY_FILE=./certs/federation/jwt-key.pub
```

Federation will verify JWT signatures and extract `tenant_id` and `principal_id` from claims.

**JWT Claims (required):**
- `tenant_id` or `tid` - Tenant identifier
- `principal_id` or `sub` - Principal identifier

**Production:**
- Use proper certificate management (e.g., cert-manager, AWS ACM, Let's Encrypt)
- Rotate certificates regularly
- Use a proper JWT signing service (e.g., Auth0, Keycloak)
- Store private keys in secret managers (e.g., AWS Secrets Manager, HashiCorp Vault)

---

## Multi-Tenancy

All requests execute within exactly one `tenant_id`. Pass tenant context via:

- **Header**: `X-Tenant-Id: tnt_your_tenant`
- **JWT claim**: `tenant_id` or `tid`

Default tenant `tnt_demo` is seeded automatically for local development.

Per-tenant enforcement includes:
- Rate limiting (QPS caps)
- Concurrent run limits
- Scoped storage isolation
- Audit logging with tenant context

---

## Environment Variables

Key configuration options:

| Variable | Description | Default |
|----------|-------------|---------|
| `AGENTOS_DEFAULT_TENANT` | Seeded default tenant | `tnt_demo` |
| `AGENTOS_QUOTA_RUN_CREATE_QPS` | Run creation rate limit | `10` |
| `AGENTOS_QUOTA_CONCURRENT_RUNS` | Max concurrent runs per tenant | `25` |
| `AGENTOS_QUOTA_INVOKE_QPS` | Model invocation rate limit | `20` |
| `AGENTOS_AUDIT_SINK` | Audit destination (`stdout`, `stderr`, `file:PATH`) | `file:data/audit/...` |

---

## Troubleshooting

### Docker Credential Helper Issues

If Docker Desktop's credential helper breaks (common on macOS/Windows), you may see errors like:

```
error getting credentials - err: exec: "docker-credential-desktop": executable file not found
```

**Quick fix (bypass credential helper for local builds):**

1. Edit `~/.docker/config.json`
2. Remove or comment out the `credsStore` line:
   ```json
   {
     "credsStore": "desktop"  // <-- remove this line
   }
   ```
3. Retry `docker compose up --build`

**Alternative (reset Docker Desktop):**

1. Docker Desktop → Troubleshoot → Reset to factory defaults
2. Re-authenticate with `docker login` if needed

### Port Already in Use

If you see "port already in use" errors, another process may be binding to 5008x ports:

```bash
# Find what's using the port (Unix/macOS)
lsof -i :50081

# Windows (PowerShell)
netstat -ano | findstr :50081
```

Stop the conflicting process or change the compose port mappings.

### Services Not Responding

If health checks fail after `./agentos up`:

1. Check container status: `docker compose -f deploy/local/compose.yaml ps`
2. Check logs: `docker compose -f deploy/local/compose.yaml logs`
3. Ensure you're using 127.0.0.1 (not localhost) to avoid IPv6 issues
4. On Windows, use `curl.exe -4` to force IPv4

---

## Known Limitations

This is a **scaffold/proof-of-concept** implementation. The following are intentionally stubbed:

**Stubbed execution layers (v1.02 scope):**
- **Agent execution**: Runs auto-progress through `queued` → `running` → `completed` without real execution
- **Model providers**: Returns stub responses; no real LLM integration
- **Tool execution**: Not implemented
- **Memory/KV store**: Not implemented
- **Persistence**: File-based JSON storage (not production-grade)

**Missing enforcement (tracked for completion via v1.02 gap-closure tracks):**
- **Idempotency key enforcement**: API accepts `idempotency_key` but doesn't enforce deduplication → Track 05
- **Run cancellation**: `POST /v1/runs/{run_id}:cancel` not implemented → Track 06
- **Agent metadata endpoints**: `GET /v1/agents` not implemented → Track 07
- **Policy engine logic**: Policy check framework present but returns allow-all → Track 08
- **Federation mTLS**: Uses HTTP with bearer tokens; mTLS not configured → Track 09

These stubs demonstrate the architecture and API contracts. Production implementations would replace the stub adapters. Gap-closure tracks (5-9) implement missing enforcement of v1.02 PRS requirements.

---

## CI Meaning

When CI is green:
- OpenAPI + schema conformance has passed
- Federation E2E has passed in Docker
- The platform is deployable in its current form

CI green does not imply production hardening beyond Phase 15.

---

## Status

**Version**: v1.02
**Execution**: Phase 16 complete (release readiness)
**CI**: Conformance + Federation E2E passing
