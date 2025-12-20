# AgentOS v1.02 Repo Rebaseline Report

**Snapshot audited:** `nexixai-agentos-platform-main (2).zip`  
**Audit timestamp:** 2025-12-20 18:19:45 (local to this environment)

This report is meant to “rebaseline” our shared understanding of what is actually in the repo right now, before proceeding to Phase 2 (Stack B).

---

## 1) What exists (key artifacts)

### Governance
- `SPEC_AUTHORITY.md` ✅
- `AGENTS.md` ✅ (includes a guardrail to avoid overwriting the execution plan template)

### Core docs
- PRS: `docs/product/agentos-prs/v1.02-prs.md` ✅
- Schemas Appendix: `docs/product/agentos-prs/v1.02-schemas-appendix.md` ✅
- Design (HLD): `docs/design/agentos-v1.02-design.md` ✅

### Execution
- Plan: `docs/plan/agentos-v1.02-execution-plan.md` ✅ (contains Phases 0–9)
- Progress log: `docs/plan/progress-log.md` ✅

### APIs
- Stack A OpenAPI: `docs/api/stack-a/openapi.yaml` ✅ (non-placeholder)
- Stack B OpenAPI: `docs/api/stack-b/openapi.yaml` ⚠️ placeholder
- Federation OpenAPI: `docs/api/federation/openapi.yaml` ⚠️ placeholder

---

## 2) Phase status (based on repo contents)

- **Phase 0:** Done (baseline + rules present)
- **Phase 1:** Done (HLD present)
- **Phase 2:** In progress
  - Stack A: mostly specified (OpenAPI present)
  - Stack B + Federation: placeholders (not yet implemented)
- Phases 3–9: not started (expected)

---

## 3) Contract drift check (Stack A vs Schemas Appendix)

**Good news:** Stack A OpenAPI reflects the core response wrapper from the appendix:
- `RunCreateResponse` includes **top-level** `correlation_id` and a `run` object, consistent with Appendix D.2.

**Remaining mismatch to resolve before we call Stack A “aligned”:**
- Appendix D.3 (“Run Object (Canonical)”) includes `correlation_id` inside the `run` object, but Stack A OpenAPI `Run` schema does **not** currently expose `run.correlation_id`.

Recommendation: add `correlation_id` as an **optional field** on `Run` (additive change), and keep the top-level `correlation_id` on responses per Appendix A.2/D.2.

---

## 4) Snapshot integrity (short hashes)

These hashes let us confirm we’re talking about the same snapshot when we refer back to this audit:

- `README.md`  `28c39be1ff03`
- `AGENTS.md`  `8745ad46f9e9`
- `SPEC_AUTHORITY.md`  `babc830234b6`
- `docs/plan/agentos-v1.02-execution-plan.md`  `5b30ffec5912`
- `docs/plan/progress-log.md`  `9aa40e18b29c`
- `docs/product/agentos-prs/v1.02-prs.md`  `72e493006397`
- `docs/product/agentos-prs/v1.02-schemas-appendix.md`  `fad338d76b41`
- `docs/design/agentos-v1.02-design.md`  `b5c12295868c`
- `docs/api/stack-a/openapi.yaml`  `2c406344ae07`
- `docs/api/stack-b/openapi.yaml`  `0c30264b3926`
- `docs/api/federation/openapi.yaml`  `0c30264b3926`

---

## 5) What we do next (no circles)

Before Phase 2 Stack B:
1. **Finish Stack A alignment** with one final additive tweak: `Run.correlation_id` (optional).
2. Then proceed to **Stack B OpenAPI** by transcribing the Schemas Appendix field-for-field.

If you want, the next step after you confirm this rebaseline is:
- a single consolidated zip that (a) adds optional `Run.correlation_id` and (b) does *nothing else*.
