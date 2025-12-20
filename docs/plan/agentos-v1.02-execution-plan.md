# AgentOS v1.02 Execution Plan

This is the **living execution plan** for NexixAI AgentOS v1.02.  
It is intended to keep implementation aligned with the v1.02 specs and prevent drift.

## Current status

**Done**
- Phase 0: Baseline + rules locked
- Phase 1: Official design doc (HLD) added

**Next**
- Phase 2: Populate OpenAPI stubs from the Schemas Appendix (Stack A → Stack B → Federation)

**Blocked**
- None

**Notes**
- We follow `SPEC_AUTHORITY.md` for conflict resolution and “spec-first” discipline.
- When in doubt: update specs first, then code.

---

## Change control rules (stay on rails)

- **/v1 is additive-only.** New optional fields allowed; no semantic rewrites.
- **Do not invent fields/semantics in code.** If ambiguous: update the spec docs first.
- **Minimal diffs.** Avoid “rewrite from scratch” changes, especially in READMEs.
- **Schemas Appendix wins** for payload truth; OpenAPI must match it.
- Federation contracts are treated as public APIs: versioned, negotiated, tenant-safe.

---

## Phase checklist

### Phase 0 — Lock the baseline and rules
**Goal:** make “spec-first” unambiguous for humans and Codex.

- [x] `docs/product/agentos-prs/v1.02-prs.md` exists
- [x] `docs/product/agentos-prs/v1.02-schemas-appendix.md` exists
- [x] Root `README.md` + `docs/README.md`
- [x] `SPEC_AUTHORITY.md` present
- [x] `AGENTS.md` present (Codex rules, minimal diffs, no README overwrites)

**Exit criteria:** a new contributor can find the source of truth in <30 seconds.

---

### Phase 1 — Official design doc (HLD)
**Goal:** bridge PRS + schemas into implementable architecture.

- [x] `docs/design/agentos-v1.02-design.md` exists

**Exit criteria:** clear module boundaries, persistence/tenancy/federation mechanics, deployment UX shape.

---

### Phase 2 — OpenAPI from contracts (no schema drift)
**Goal:** machine-checkable contracts and SDK readiness.

**Deliverables**
- [ ] `docs/api/stack-a/openapi.yaml` (Runs + Events + health)
- [ ] `docs/api/stack-b/openapi.yaml` (Chat/Embeddings + Policy + Models)
- [ ] `docs/api/federation/openapi.yaml` (Peer + Forwarding + Events semantics)

**Order**
1. Stack A: `POST /v1/agents/{agent_id}/runs`, `GET /v1/runs/{run_id}`, `GET /v1/runs/{run_id}/events`, health
2. Stack B: chat/embeddings/policy/models
3. Federation: peer info/capabilities + runs:forward (+ optional events ingest)

**Exit criteria:** OpenAPI covers required v1 endpoints and matches the Schemas Appendix shapes.

---

### Phase 3 — Canonical examples (“golden payloads”)
**Goal:** prevent drift and enable testing.

- [ ] Populate `docs/api/**/examples/*.json` with canonical payloads from the Schemas Appendix.

**Exit criteria:** every required endpoint has request/response examples; event envelope example exists.

---

### Phase 4 — Conformance tests + drift gates (CI)
**Goal:** enforce the spec automatically.

- [ ] `tests/conformance/` validates examples against schemas/OpenAPI
- [ ] additive-only guard for `/v1`
- [ ] CI workflow runs conformance checks

**Exit criteria:** drift fails CI.

---

### Phase 5 — Minimal Go scaffolding (matches spec)
**Goal:** start implementing without premature complexity.

- [ ] `cmd/agentos/` CLI scaffold
- [ ] Stack A service scaffold
- [ ] Stack B service scaffold
- [ ] Federation endpoints scaffold
- [ ] Ports/adapters layout + types aligned to schemas

**Exit criteria:** services run locally and return schema-correct stubs.

---

### Phase 6 — Deployment UX v1 (“one command”)
**Goal:** operator experience: up/redeploy/validate/status/nuke + reports.

- [ ] `agentos up|redeploy|validate|status|nuke`
- [ ] real-time phase output
- [ ] reports (md + json)
- [ ] default seeded tenant/admin/sample agent/tool (dev/demo)

**Exit criteria:** `agentos up` yields a working system + printed summary + reports + validate gates.

---

### Phase 7 — Multi-tenancy + quotas + auditing (enforced)
**Goal:** tenants are real and safe.

- [ ] tenant-scoped persistence keys everywhere
- [ ] enforcement: Stack A run limits; Stack B token/cost budgets + entitlements
- [ ] audit log baseline (tenant_id always included)

**Exit criteria:** cross-tenant isolation tests pass; quotas return consistent errors.

---

### Phase 8 — Federation v1 (two nodes)
**Goal:** prove the “connect nodes” story.

- [ ] peer discovery + capabilities
- [ ] runs:forward end-to-end
- [ ] remote event streaming works (dedupe/order semantics)

**Exit criteria:** Node A forwards run to Node B and streams events back with tenant enforcement.

---

### Phase 9 — Hardening pass
**Goal:** close key enterprise gaps safely.

- [ ] mTLS guidance/option
- [ ] rate limiting at edge
- [ ] structured logs + correlation everywhere
- [ ] alert test via SMTP
- [ ] image/version pinning strategy

**Exit criteria:** production-ish defaults exist and are documented.
