# AgentOS v1.02 Execution Plan

This is the **living execution plan** for NexixAI AgentOS v1.02.  
It is intended to keep implementation aligned with the v1.02 specs and prevent drift.

## Current status

**Done**
- Phase 0: Baseline + rules locked
- Phase 1: Official design doc (HLD) added
- Phase 2 (in progress): Stack A OpenAPI drafted; alignment to Schemas Appendix is in progress (see progress log)

**Next**
- Phase 2: Finish Stack A OpenAPI alignment (Appendix → OpenAPI, field-for-field)
- Phase 2: Populate Stack B OpenAPI from Schemas Appendix
- Phase 2: Populate Federation OpenAPI from Schemas Appendix

**Blocked**
- None

**Notes**
- We follow `SPEC_AUTHORITY.md` for conflict resolution and “spec-first” discipline.
- When in doubt: update specs first, then code.
- This plan is **not** normative. Payload truth remains in the Schemas Appendix.

---

## Change control rules (stay on rails)

1. **Schemas Appendix wins** for payload truth.
2. If implementation discovers ambiguity: **update spec first**, then code.
3. **/v1 is additive-only**: new optional fields OK; no semantic rewrites.
4. Codex must make **minimal diffs**; never overwrite README/docs wholesale.
5. Every phase has exit criteria; we don’t skip phases unless there’s a clear reason.

---

## Phase checklist

### Phase 0 — Lock the baseline and rules
**Goal:** make the repo “spec-first” with zero ambiguity about what’s authoritative.

**Deliverables**
- [x] `docs/product/agentos-prs/v1.02-prs.md`
- [x] `docs/product/agentos-prs/v1.02-schemas-appendix.md`
- [x] `SPEC_AUTHORITY.md`
- [x] Root `README.md` + `docs/README.md`
- [x] `AGENTS.md` with Codex rules:
  - don’t overwrite README
  - minimal diffs
  - spec authority precedence
  - if unclear, update spec—not code

**Exit criteria**
- A new contributor (or Codex) can answer “What’s the source of truth?” in 30 seconds.

---

### Phase 1 — Write the official design doc (HLD)
**Goal:** bridge PRS + schemas into an implementable architecture.

**Deliverable**
- [x] `docs/design/agentos-v1.02-design.md`

**Exit criteria**
- You can implement from it without asking “where does this logic go?”

---

### Phase 2 — Turn contracts into OpenAPI (without changing schemas)
**Goal:** make the API machine-checkable and SDK-ready.

**Deliverables**
- [ ] `docs/api/stack-a/openapi.yaml`
- [ ] `docs/api/stack-b/openapi.yaml`
- [ ] `docs/api/federation/openapi.yaml`

**Order of implementation**
1. Stack A: Runs + Events (SSE documented) + health
2. Stack B: Chat/Embeddings + Policy + Models
3. Federation: peer info/capabilities + runs:forward (+ optional events ingest)

**Rules**
- OpenAPI must match Schemas Appendix field-for-field (schemas remain king).

**Exit criteria**
- OpenAPI covers all required v1 endpoints in PRS and payload shapes match appendix.

---

### Phase 3 — Canonical examples (“golden payloads”)
**Goal:** make drift impossible and testing easy.

**Deliverables**
- [ ] Populate `docs/api/**/examples/*.json` with the exact canonical examples from the Schemas Appendix (not “similar” ones).

**Exit criteria**
- Every required endpoint has request/response examples; event envelope example exists.

---

### Phase 4 — Conformance tests + drift gates (CI)
**Goal:** enforce the spec automatically.

**Deliverables**
- [ ] `tests/conformance/` with tests that:
  - validate examples conform to schemas/OpenAPI
  - enforce additive-only changes for `/v1`
  - enforce required event envelope fields
  - enforce tenant scoping presence in required places
- [ ] GitHub Actions workflow:
  - validate OpenAPI syntax
  - run conformance tests

**Exit criteria**
- A PR that drifts contracts fails CI.

---

### Phase 5 — Minimal Go scaffolding that matches the spec
**Goal:** start implementing without premature complexity.

**Deliverables (scaffold only)**
- [ ] `cmd/agentos/` (CLI)
- [ ] `stack-a/` (HTTP server skeleton + handlers stubs)
- [ ] `stack-b/` (HTTP server skeleton + provider interface)
- [ ] `federation/` (peer endpoints skeleton)
- [ ] `internal/ports/` + `internal/adapters/`
- [ ] `internal/types/` structs aligned to schemas

**Exit criteria**
- Servers run locally; endpoints return stubbed responses that match OpenAPI (even if not fully functional).

---

### Phase 6 — Deployment UX v1 (“one command”)
**Goal:** deliver operator experience: one command deploy, validate gates, and reports.

**Deliverables**
- [ ] `agentos up|redeploy|validate|status|nuke` (CLI commands)
- [ ] Local compose/dev wiring (or local-run instructions)
- [ ] Report output artifacts (md + json)
- [ ] Seed behavior (default tenant in dev/demo)

**Exit criteria**
- `agentos up` produces a live system + printed summary + reports + validate gates.

---

### Phase 7 — Multi-tenancy + quotas + auditing (enforced, not aspirational)
**Goal:** tenants are real and safe.

**Deliverables**
- [ ] Tenant-scoped persistence keys
- [ ] Auth context propagation everywhere internally
- [ ] Quota/budget enforcement points:
  - Stack A: run concurrency + create QPS
  - Stack B: token/cost budgets + entitlements
- [ ] Audit log baseline (tenant_id always included)

**Exit criteria**
- Cross-tenant isolation test passes.
- Quota exceed returns consistent `quota_exceeded` errors.

---

### Phase 8 — Federation v1 (two nodes working)
**Goal:** prove the “connect nodes” story.

**Deliverables**
- [ ] peer discovery / capabilities
- [ ] runs:forward works end-to-end
- [ ] events stream works from remote run
- [ ] dedupe semantics by event_id + ordering by sequence

**Exit criteria**
- Node A forwards a run to Node B, and Node A client can stream events + receive final output.

---

### Phase 9 — Hardening pass (“enterprise touches” without boiling ocean)
**Goal:** close the biggest gaps safely.

**Deliverables**
- [ ] mTLS guidance or built-in option
- [ ] rate limiting and basic abuse protection at the gateway
- [ ] structured logs + correlation IDs everywhere
- [ ] baseline alerting (SMTP email test)
- [ ] image/version pinning strategy (even if minimal)

**Exit criteria**
- “Production-ish” defaults exist and are documented.
