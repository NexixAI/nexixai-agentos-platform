# AgentOS v1.02 Execution Plan

This is the **living execution plan** for NexixAI AgentOS v1.02.  
It is intended to keep implementation aligned with the v1.02 specs and prevent drift.

## Current status

**Done**
- Phase 0: Baseline + rules locked
- Phase 1: Official design doc (HLD) added
- Phase 2 (partial): Stack A OpenAPI populated (Runs + Events + Health)

**Next**
- Phase 2: Populate Stack B OpenAPI from Schemas Appendix
- Phase 2: Populate Federation OpenAPI from Schemas Appendix

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
- [x] `docs/api/stack-a/openapi.yaml` (Runs + Events + health)
- [ ] `docs/api/stack-b/openapi.yaml` (Chat/Embeddings + Policy + Models)
- [ ] `docs/api/federation/openapi.yaml` (Peer + Forwarding + Events semantics)

**Order**
1. Stack A: `POST /v1/agents/{agent_id}/runs`, `GET /v1/runs/{run_id}`, `GET /v1/runs/{run_id}/events`, health
2. Stack B: chat/embeddings/policy/models
3. Federation: peer info/capabilities + runs:forward (+ optional events ingest)

**Exit criteria:** OpenAPI covers required v1 endpoints and matches the Schemas Appendix shapes.
