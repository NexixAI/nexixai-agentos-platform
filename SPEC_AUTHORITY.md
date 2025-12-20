# SPEC_AUTHORITY — NexixAI AgentOS

This file defines the **source-of-truth documents** and **conflict resolution rules** for implementing AgentOS.
It exists to prevent “interpretation drift” between specs, code, and generated artifacts.

## Normative documents (v1.02)

**Primary**
1. `docs/product/agentos-prs/v1.02-schemas-appendix.md`  
   Source of truth for **JSON payload shapes**, required fields, event envelopes, auth context, federation payloads.

2. OpenAPI specs (once populated; must match the Schemas Appendix)
   - `docs/api/stack-a/openapi.yaml`
   - `docs/api/stack-b/openapi.yaml`
   - `docs/api/federation/openapi.yaml`

**Secondary**
3. `docs/product/agentos-prs/v1.02-prs.md`  
   Source of truth for **behavioral requirements**, quality attributes, deployment UX, observability, and acceptance criteria.

4. `docs/design/agentos-v1.02-design.md`  
   Source of truth for **internal architecture decisions**: module boundaries, ports/adapters, persistence strategy, multi-tenancy enforcement points, and federation mechanics.

5. Examples + conformance tests (once added)
   - `docs/api/*/examples/*.json`
   - `tests/conformance/*`

## Conflict resolution (strict)

If documents disagree, the precedence order is:
1) Schemas Appendix  
2) OpenAPI specs  
3) PRS  
4) Design doc  
5) Examples / tests  
6) Everything else

If a conflict is found:
- **Do not invent new fields or semantics in code.**
- Update the higher-precedence document first (usually Schemas Appendix), then update the lower-precedence docs to match.

## Versioning rules

- `/v1` APIs are **additive-only**. New optional fields may be added; existing field meaning must not change.
- Breaking changes require creating `/v2` documents and endpoints.

## Multi-tenancy invariants (non-negotiable)

- Every request is executed within exactly one `tenant_id`.
- `tenant_id` must be present in Auth Context for all internal calls.
- No cross-tenant reads/writes/events; federation must preserve and enforce tenant context end-to-end.

## README update rules (DO NOT OVERWRITE)

When a change affects usage, APIs, docs layout, build, deploy, or runbooks:
- Update `README.md` and/or `docs/README.md`.

Strict rules:
1. **Do not rewrite READMEs from scratch.** Preserve structure/tone/sections unless explicitly instructed.
2. **Make minimal diffs** focused on impacted sections only.
3. **Do not delete user-authored text** unless it is clearly incorrect.
4. Keep spec links stable:
   - `docs/product/agentos-prs/v1.02-prs.md`
   - `docs/product/agentos-prs/v1.02-schemas-appendix.md`
5. If adding new docs, append new bullets—do not reorganize the whole file.

## Implementation guidance for automation/Codex

Automation must:
- Treat `SPEC_AUTHORITY.md` as the decision rule.
- Prefer updating specs over “guessing” in code.
- Keep changes small and reversible; update docs alongside code where relevant.
