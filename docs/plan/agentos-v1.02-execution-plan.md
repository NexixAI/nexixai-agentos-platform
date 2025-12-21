# AgentOS v1.02 Execution Plan

This is the **living execution plan** for NexixAI AgentOS v1.02.
It keeps us honest: we move fast **without drifting**.

> **Rule:** This plan is process guidance (non-normative). Contract truth remains in the Schemas Appendix and OpenAPI per `SPEC_AUTHORITY.md`.

## Current status
**Done**
- Phase 0: Baseline + rules locked
- Phase 1: Official design doc (HLD) added
- Phase 2: Stack A/B/Federation OpenAPI populated and aligned to Schemas Appendix examples
- Phase 3: Canonical examples populated (Stack A/B/Federation)
- Phase 4: Conformance tests + GitHub Actions workflow added (drift gates)

**Next**
- Phase 5: Minimal Go scaffolding that matches the spec (servers + handlers stubs)

**Blocked**
- None
## Change control rules (no drift, no thrash)

1. **Schemas Appendix wins** for payload truth.
2. If anything is ambiguous: **update the spec first**, then code.
3. **/v1 is additive-only**: new optional fields OK; no semantic rewrites.
4. **No piecemeal patches:** when we discover drift, we ship a single consolidated zip that restores alignment (plan + contracts + logs together).
5. **Minimal diffs:** never rewrite READMEs or governance docs wholesale.

## Phase 0 — Lock the baseline and rules
**Goal:** make the repo “spec-first” with zero ambiguity about what’s authoritative.

**Deliverables**
- [x] `docs/product/agentos-prs/v1.02-prs.md`
- [x] `docs/product/agentos-prs/v1.02-schemas-appendix.md`
- [x] `SPEC_AUTHORITY.md`
- [x] Root `README.md` + `docs/README.md`
- [x] `AGENTS.md` with Codex rules (no README overwrites, minimal diffs, spec authority precedence)

**Exit criteria**
- New contributor can answer “what’s authoritative?” in <30 seconds.

## Phase 1 — Official design doc (HLD)
**Goal:** bridge PRS + schemas into implementable architecture.

**Deliverables**
- [x] `docs/design/agentos-v1.02-design.md`

**Exit criteria**
- You can implement from it without asking “where does this logic go?”

## Phase 2 — Turn contracts into OpenAPI (without changing schemas)
**Goal:** make the API machine-checkable and SDK-ready.

**Deliverables**
- [ ] `docs/api/stack-a/openapi.yaml`
- [ ] `docs/api/stack-b/openapi.yaml`
- [ ] `docs/api/federation/openapi.yaml`

**Order**
1. Stack A: Runs + Events (SSE documented) + health
2. Stack B: Chat/Embeddings + Policy + Models
3. Federation: peer info/capabilities + runs:forward (+ optional events ingest)

**Rules**
- OpenAPI must match Schemas Appendix field-for-field.

**Exit criteria**
- OpenAPI covers required v1 endpoints in PRS AND payload shapes match appendix.

## Phase 3 — Canonical examples (“golden payloads”)
**Goal:** make drift impossible and testing easy.

**Deliverables**
- [ ] Populate `docs/api/**/examples/*.json` with the exact canonical examples from the Schemas Appendix (not “similar” ones).

**Exit criteria**
- Every required endpoint has request/response examples; event envelope example exists.

## Phase 4 — Conformance tests + drift gates (CI)
**Goal:** enforce the spec automatically.

**Deliverables**
- [ ] `tests/conformance/` validates examples against schemas/OpenAPI
- [ ] additive-only guards for `/v1` contracts
- [ ] GitHub Actions workflow runs conformance checks (OpenAPI validation + tests)

**Exit criteria**
- A PR that drifts contracts fails CI.

## Phase 5 — Minimal Go scaffolding that matches the spec
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

## Phase 6 — Deployment UX v1 (“one command”)
**Goal:** deliver the operator experience: one command deploy, validate gates, and reports.

**Deliverables**
- [ ] `agentos up|redeploy|validate|status|nuke`
- [ ] real-time phase output (informative progress)
- [ ] reports (md + json) after each command
- [ ] pre-seeded dev/demo environment (default tenant/admin/sample)

**Exit criteria**
- `agentos up` yields a working system + printed summary + reports + validate gates.

## Phase 7 — Multi-tenancy + quotas + auditing (enforced)
**Goal:** tenants are real and safe.

**Deliverables**
- [ ] tenant-scoped persistence keys everywhere
- [ ] auth context propagation on all internal calls
- [ ] quota enforcement:
  - Stack A: run concurrency + create QPS
  - Stack B: token/cost budgets + entitlements
- [ ] audit log baseline (always includes `tenant_id`)

**Exit criteria**
- Cross-tenant isolation test passes.
- Quota exceed returns consistent `quota_exceeded` errors.

## Phase 8 — Federation v1 (two nodes working)
**Goal:** prove the “connect nodes” story.

**Deliverables**
- [ ] peer discovery / capabilities
- [ ] runs:forward works end-to-end
- [ ] events stream works from remote run
- [ ] dedupe semantics by event_id + ordering by sequence

**Exit criteria**
- Node A forwards a run to Node B and Node A client streams events + receives final output.

## Phase 9 — Hardening pass
**Goal:** close the biggest gaps safely (enterprise touches without boiling the ocean).

**Deliverables**
- [ ] mTLS guidance or built-in option
- [ ] rate limiting / basic abuse protection at the edge
- [ ] structured logs + correlation IDs everywhere
- [ ] baseline alerting (SMTP email test)
- [ ] image/version pinning strategy

**Exit criteria**
- Production-ish defaults exist and are documented.
