# NexixAI AgentOS v1.02 Product Requirements and Specifications

This document is **AgentOS v1.02 PRS** for **NexixAI**. It is **v1.01 exactly as-is**, with **additions and necessary changes to support multi-tenancy**. The two stacks continue to use **generic names**:
- **Stack A: Orchestration Stack**
- **Stack B: Model Services Stack**

---

## 1) Product vision
**AgentOS** is a portable, federatable, **multi-tenant** agent platform that:
- Runs agents safely and repeatably (**Stack A**)
- Provides governed, vendor-agnostic model access (**Stack B**)
- Can be deployed out-of-the-box with one command, fully observable, and operator-friendly
- Can connect multiple nodes/environments into a coherent system without tight coupling
- Supports multiple tenants with isolation, quotas, and auditability

## 2) Primary user personas
1) **Platform Operator**: one-command deploy, status/validate, reports, safe nuke/redeploy flows  
2) **Product Integrator (internal product team)**: stable APIs/SDKs to invoke agents and stream results  
3) **Platform Developer**: clean internal contracts so adding/replacing components is low-lift  
4) **Security/Compliance**: audit logs, authZ, policy, controlled model usage  
5) **Tenant Admin** (new): manages tenant config, entitlements, quotas, keys/tokens, and visibility into tenant health/cost  

---

# 3) System architecture

## 3.1 Two-stack architecture (per node/environment)

### Stack A: Orchestration Stack
**Responsibilities**
- Agent lifecycle (deployable agents), run lifecycle (invocations)
- Planning/execution loop, tool calling, retries/timeouts
- State persistence for runs (replay/debug), event emission
- AuthN/AuthZ enforcement for runs + tools + memory access
- Federation client/server for run forwarding and event streaming
- **Tenant isolation and enforcement** for agents/runs/events/tools/memory/state (new)

### Stack B: Model Services Stack
**Responsibilities**
- Stable model interface (chat, embeddings) independent of vendor/backend
- Policy & governance gates (PII redaction, allow/deny, tool-use permissions)
- Routing/fallback/canary across model backends
- Usage metering (tokens/cost), per-tenant quotas/budgets
- Audit log for model requests, policy decisions, routing decisions
- **Tenant entitlements + budgets + policy enforcement** on every model request (new)

## 3.2 Federation (multi-node / multi-environment connectivity)
**Goal**: connect “AgentOS nodes” without leaking internals or requiring coordinated deploys.

**Federation capabilities**
- Peer discovery & capability negotiation
- Secure invocation forwarding (runs/tools/model services)
- Event replication/streaming
- Version skew tolerance via negotiation
- **Tenant context propagation and enforcement across nodes** (new)

---

# 4) External product-facing API (Stack A)

## 4.1 Public objects
- **Tenant** (new): organizational boundary for isolation, quotas, and policy  
- **Agent**: named capability configuration with versioning and permissions  
- **Run**: invocation record with status, outputs, and references  
- **Event**: streamed timeline of a run (steps, tool calls, model calls)  
- **Tool** (optional external integration): callable function hosted by AgentOS or by the product via webhook  

## 4.2 Required endpoints (v1)
**Run lifecycle**
- `POST /v1/agents/{agent_id}/runs`  
  - Inputs: user input, context, **tenant**, optional tool allowlist, idempotency key  
  - Output: `run_id`, `status`, `events_url`
- `GET /v1/runs/{run_id}`
- `POST /v1/runs/{run_id}:cancel`
- `GET /v1/runs/{run_id}/events` (SSE; required)  
  - Event types: `run.created`, `step.started`, `tool.requested`, `tool.completed`, `model.requested`, `model.completed`, `run.completed`, `run.failed`

**Agent metadata**
- `GET /v1/agents`
- `GET /v1/agents/{agent_id}`

**Health**
- `GET /healthz` (liveness)
- `GET /readyz` (readiness: deps reachable)

## 4.3 External API requirements
- **Multi-tenancy** (new):
  - All requests are executed within a single `tenant_id` derived from auth context (preferred) or an explicit header for service-to-service calls.
  - All responses include `tenant_id` in metadata.
- **Versioning**: `/v1` additive-only. Breaking changes require `/v2`.
- **Idempotency**: required on `POST /runs` and any mutation endpoints.
- **Errors**: consistent error model (`code`, `message`, `retryable`, `details`, `correlation_id`)
- **Tracing**: accept/emit `traceparent` and return `correlation_id` in every response.
- **Auth**: JWT/OIDC for user-context; service-to-service tokens for backend integrations.
- **Auditability**: every run has an immutable audit record (who/what/when/**tenant**/policy).

---

# 5) Internal platform contracts (ports/adapters)

## 5.1 Non-negotiable rule
**Modules do not call each other via concrete implementations.** They call **ports** (interfaces). Implementations are adapters.

## 5.2 Required ports (Stack A)
1) **Model Port** (Stack A → Stack B)  
   - `chat(request) -> response`  
   - `embed(request) -> response`  
   - `policy_check(request) -> decision`  
   - `list_models() -> capabilities`
2) **Tool Port**  
   - `invoke(tool_id, input, auth_ctx, timeout_ms) -> output`
3) **Memory Port**  
   - `put(ns, key, value, metadata)`  
   - `get(ns, key)`  
   - `search(ns, query, filters)` (vector/keyword)
4) **Run State Port**  
   - persistent run graph, step transitions, idempotency keys, replay support
5) **Event Port**  
   - `emit(event_envelope)` + optional streaming sink
6) **Queue/Scheduler Port**  
   - enqueue run, concurrency limits, backpressure

**Port contract requirement (new):** every port call carries an **Auth Context** that includes at minimum:
- `tenant_id`
- `principal_id` (user/service)
- `scopes/roles`
- `policy_context` (data classification, region constraints if any)

## 5.3 Extensibility requirements
Adding a new component (new DB, vector store, queue, model backend) should be:
1) implement the port  
2) declare capabilities  
3) pass conformance tests  
4) roll out via config + canary  

---

# 6) Stack B API + governance

## 6.1 Stack B endpoints
Prefer OpenAI-compat where possible:
- `POST /v1/chat/completions`
- `POST /v1/embeddings`
- `POST /v1/policy/check` (or moderation-like)
- `GET /v1/models`

## 6.2 Stack B request/response requirements
Every response must include:
- `usage`: tokens, latency  
- `policy_outcome`: allowed/blocked/redacted + reasons  
- `route`: backend served (for audit/debug)  
- `correlation_id` + `traceparent` propagation  

**Multi-tenancy requirement (new):**
- Every request must include/derive `tenant_id` and must be enforced for:
  - model entitlements (allowed model tiers/backends)
  - budgets/quotas (tokens/cost caps)
  - policy gates (block/redact rules)

## 6.3 Stack B capabilities
- Backend plugins: `ollama`, `vllm`, `tgi`, `openai`, `anthropic`, `aws-bedrock`, etc.
- Policy gates:
  - PII detection/redaction (baseline)
  - allow/deny per tenant
  - tool-use permission enforcement
  - max tokens, budget enforcement
- Routing:
  - choose model tier by `quality/cost/latency` requirements
  - fallback chains
  - canary by tenant/agent percentage

---

# 7) Federation specification

## 7.1 Federation goals
- Connect nodes in separate environments safely
- Don’t require shared internal networks
- Keep product-facing API stable (federation is transparent to integrators)

## 7.2 Peer endpoints (minimum)
- `GET /v1/peer/info` → stack identity, versions, endpoints  
- `GET /v1/peer/capabilities` → supported features/models/tools/limits  

## 7.3 Forwarding endpoints (minimum)
- `POST /v1/federation/runs:forward`  
  - Inputs: run_request + routing selector (stack_id/region/capability)  
  - Output: remote `run_id` + `events_url`
- `GET /v1/federation/runs/{run_id}/events` (or returned URL) for cross-stack streaming
- `POST /v1/federation/events:ingest` (optional for replication)

## 7.4 Federation requirements
- **Security**: mTLS between stacks + JWT identity propagation.  
- **Global IDs**: run_id/event_id unique across federation.  
- **Delivery semantics**: at-least-once events; dedupe by `event_id`; ordering by `sequence`.  
- **Version skew**: negotiation via capabilities; fail gracefully if unsupported.  

**Multi-tenancy requirements (new):**
- Forwarded requests must carry `tenant_id` and principal identity (signed/verifiable).
- Receiving peer must enforce tenant isolation and entitlements locally.
- Event streaming/replication must never cross tenant boundaries (tenant-scoped streams and filters).

---

# 8) Deployment UX and operations spec

## 8.1 Single-command CLI
Ship one entrypoint: `agentos`

### Required commands
- `agentos up` (idempotent deploy)
- `agentos redeploy` (non-destructive reapply + restart)
- `agentos validate` (health + smoke tests; non-destructive)
- `agentos status`
- `agentos logs <service>`
- `agentos nuke --yes-really`
- `agentos nuke --hard --yes-really`

### Modes
- `--mode demo|dev|prod` (changes behaviors explicitly)
- `--profile local|federated|gpu-node` (optional)

### Config precedence
1) CLI flags  
2) `.agentos.env` (generated but editable)  
3) defaults baked in repo  

## 8.2 Real-time deploy output
Deploy must stream:
- Phase progress (preflight, boot, seed, wire, validate, report)
- Real milestone counts (e.g., services healthy 7/9)
- On failure: failing check + last relevant logs + suggested next command

## 8.3 Post-success summary (required)
Print and write into report:
- Product URLs + Admin URLs
- Observability URLs (Grafana/Prometheus/logs/traces)
- Credentials location + rotation guidance
- Health check results + smoke test timings
- Alert status + “send test alert” instructions
- Paths to report artifacts

**Multi-tenancy summary additions (new):**
- Default seeded `tenant_id` (e.g., `tnt_demo` in demo mode)
- How to create a new tenant (command/API)
- Tenant admin credential/token location (where applicable)
- Default quotas and how to override them

## 8.4 Reports (required artifacts)
On every `up/redeploy/nuke/validate`:
- `reports/<action>-YYYYMMDD-HHMMSS.md` (human)
- `reports/<action>-...json` (machine)

Must include:
- host info, images/digests, ports/endpoints
- seeded objects (**tenant/admin/sample agent IDs**)
- health checks + smoke tests
- alert configuration status
- remediation tips

## 8.5 “Works out of the box” deployment requirements
- prewired networks and service discovery
- preprovisioned dashboards + datasources
- preconfigured alerting (SMTP email supported)
- preseeded tenant + admin + sample agent/tool
- health checks and smoke tests are required gates

**Multi-tenant seeding (new):**
- `demo/dev`: create a default tenant and tenant admin principal (and optionally a platform admin)
- `prod`: require explicit tenant creation or bootstrap flow, but still support a one-command guided bootstrap

---

# 9) Observability and alerting spec

## 9.1 Minimum observability stack (OOTB)
- Metrics: Prometheus scraping Stack A, Stack B, gateway, system exporters
- Dashboards: Grafana provisioned with core dashboards
- Logs: structured JSON logs with `run_id`, `trace_id`, `tenant_id`
- Traces: OpenTelemetry Collector + backend (Tempo/Jaeger)

## 9.2 Required metrics
**Stack A**
- run_count, run_success, run_failure
- run_latency_ms (p50/p95)
- step_latency_ms, tool_latency_ms, tool_error_rate
- queue_depth, active_runs

**Stack B**
- model_request_count, model_error_rate
- token_usage, cost_estimate
- policy_block_count, redaction_count
- backend_health, route_distribution

**Gateway**
- 4xx/5xx rates, auth failures

**Multi-tenancy metrics requirements (new):**
- Metrics should be attributable to `tenant_id` **where safe**, with controls to avoid runaway cardinality:
  - Provide aggregated “top tenants” views and/or sampling
  - Always support per-tenant dashboards via filtering or pre-aggregation

## 9.3 Required alerts (baseline)
- service down / scrape missing
- high 5xx rate
- p95 latency high
- model backend unreachable
- queue depth/backlog high
- disk nearly full
- database unavailable

**Multi-tenancy alerting additions (new):**
- Tenant quota/budget exceeded (tokens/cost)
- Tenant run concurrency limit reached (rate-limited) sustained over threshold
- Optional: “noisy neighbor” detection (one tenant dominating capacity)

## 9.4 Alert delivery
- Alertmanager SMTP email required OOTB (configurable via `.agentos.env`)
- Provide `agentos test-alert` (or as part of `validate`) to prove delivery

---

# 10) Security requirements

## 10.1 AuthN/AuthZ
- Tenant-scoped authorization model
- Scopes/permissions for:
  - runs:create/read/cancel
  - agents:read
  - tools:invoke (and tool-hosting if enabled)
  - admin operations
- Separate admin surfaces; locked down in prod

**Multi-tenancy enforcement (new):**
- `tenant_id` must be present in auth context for every request.
- Authorization checks must include tenant scope (no cross-tenant reads/writes).
- Provide separation of roles:
  - platform operator/admin
  - tenant admin
  - tenant user/service

## 10.2 Secrets
- Dev/demo: generated local secrets allowed
- Prod: secret manager integration required (or mandatory rotation + no plaintext env secrets)
- All secrets access audited

## 10.3 Audit logging
- Runs: who invoked what agent, what policy context
- Stack B: model request metadata, policy outcomes, routing decisions
- Federation: peer identity, forwarded runs, event replication actions

**Multi-tenancy audit additions (new):**
- Every audit record includes `tenant_id` and principal identity.
- Audit queries must be tenant-filtered by default, with platform-admin override controls.

---

# 11) Testing requirements

## 11.1 Deployment gates (required)
- Health checks: all services ready + dependency connectivity
- Smoke tests:
  - start a sample agent run
  - stream events and complete
  - execute at least one tool call
  - execute at least one model call through Stack B with policy check
  - verify metrics changed (within 60 seconds)

**Multi-tenancy deployment tests (new):**
- Validate default tenant exists (demo/dev)
- Run invocation works under the default tenant
- Cross-tenant isolation test (minimum):
  - create Tenant A run
  - ensure Tenant B cannot read/list Tenant A run or events

## 11.2 Conformance tests (required for extensibility)
For each port (memory/tool/model/queue/event):
- behavior correctness
- standardized error mapping
- timeouts and retry semantics
- capability reporting correctness
- version compatibility checks

**Multi-tenant conformance additions (new):**
- Ports must correctly scope reads/writes by `tenant_id`
- Error mapping must include “forbidden/unauthorized” vs “not found” semantics per policy

---

# 12) Operational semantics (destructive vs non-destructive)

## 12.1 Non-destructive operations
- `status`, `validate`, `logs`, targeted `restart`, `redeploy`

## 12.2 Destructive operations
- `nuke` removes containers/networks
- `nuke --hard` removes volumes/caches/models
- destructive commands require `--yes-really` and must generate reports

**Multi-tenancy operations (new):**
- `redeploy` must preserve tenant data by default.
- Any “tenant reset” or “delete tenant” operation (if introduced) must be explicit, audited, and gated.

---

# 13) Acceptance criteria for v1
A v1 release is “done” when:
1) `agentos up` works on a new machine with no edits  
2) Deploy provides real-time progress + final summary + report artifacts  
3) Dashboards live within 2 minutes  
4) Email alert test succeeds  
5) `agentos validate` passes on clean deploy and fails when deps are broken  
6) `redeploy` is non-destructive and idempotent  
7) Federation forwards a run to a second node and streams events back  
8) Swapping a component requires only adapter + conformance + config (no core rewrite)

**Multi-tenancy acceptance additions (new):**
9) Tenant A cannot read/list Tenant B agents/runs/memory/events  
10) Per-tenant quotas (runs/concurrency and model budgets) are enforced with clear errors  
11) Federation preserves and enforces tenant context end-to-end  

---

# 14) Suggested repo spec layout (v1)
- `agentos/` (CLI + phase runner)
- `deploy/` (defaults, compose/helm, seed, health, reports)
- `stack-a/` (public API, runtime, ports, adapters)
- `stack-b/` (API, providers, policy, routing)
- `federation/` (shared schemas + peer endpoints)
- `observability/` (dashboards, rules, collector config)
- `runbooks/` (operator docs)

---

# 15) Multi-tenancy specification (new)

## 15.1 Tenancy model
- **Tenant**: top-level isolation boundary identified by `tenant_id` (globally unique)
- **Principal**: user or service identity operating within a tenant (`principal_id`)
- **Policy Context**: request metadata affecting governance (classification, region constraints, etc.)

## 15.2 Isolation requirements (v1 default)
**Isolation level:** logical isolation (tenant-scoped keys/tables/namespaces) with an upgrade path to stronger isolation.

Tenant-scoped partitions:
- Stack A: agents, runs, run events, tool registry/permissions, memory namespaces, run state
- Stack B: entitlements, budgets, policy rules, usage records, audit records

## 15.3 Quotas and budgets
Per tenant enforce at minimum:
- max concurrent runs (Stack A)
- run creation QPS / rate limit (Stack A)
- token and/or cost budget per time window (Stack B)
- allowed model tiers/backends (Stack B)

## 15.4 Admin controls (minimum capability)
Tenant management must exist as either:
- a minimal admin API surface, and/or
- CLI commands that call internal APIs

Required capabilities:
- create tenant
- configure tenant entitlements and quotas
- issue/rotate tenant tokens/keys
- enable/disable tenant

## 15.5 Observability for tenants
- Logs/traces must include `tenant_id`
- Metrics must support tenant attribution with safeguards against high cardinality
- Provide at least one “tenant health” dashboard pattern (filterable or pre-aggregated)

## 15.6 Federation tenancy rules
- Tenant identity must be propagated (signed/verifiable)
- Receiving peer must enforce tenant policies locally
- Event streams must be tenant-scoped; no cross-tenant leakage
