# AgentOS v1.02 Execution Plan

This is the **living execution plan** for NexixAI AgentOS v1.02.  
It is intended to keep implementation aligned with the v1.02 specs and prevent drift.

## Current status

**Done**
- Phase 0: Baseline + rules locked
- Phase 1: Official design doc (HLD) added

**In progress / Needs alignment**
- Phase 2: Stack A OpenAPI drafted — **must be aligned to Schemas Appendix** (request/response shapes, error model, SSE envelope)

**Next**
- Phase 2: Align Stack A OpenAPI to Schemas Appendix (fix drift)
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

### Phase 2 — OpenAPI from contracts (no schema drift)
**Goal:** machine-checkable contracts and SDK readiness.

**Deliverables**
- [ ] `docs/api/stack-a/openapi.yaml` (Runs + Events + health) — **aligned to appendix**
- [ ] `docs/api/stack-b/openapi.yaml` (Chat/Embeddings + Policy + Models)
- [ ] `docs/api/federation/openapi.yaml` (Peer + Forwarding + Events semantics)

**Order**
1. Stack A alignment (then mark complete)
2. Stack B
3. Federation
