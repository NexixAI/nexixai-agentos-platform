# Phase Runner — Codex Automation (v1.02)

This runner defines the **single, repeatable** procedure Codex must follow to implement AgentOS v1.02 **phases sequentially**.

It does **not** change product requirements. It executes the plan in `docs/plan/${plan_version}/phases/phase-*.md` while obeying `SPEC_AUTHORITY.md`.

---

## Inputs (single source of truth)

Codex must read the next phase number and plan version from the **first existing** file in this order:

1) `configs/NEXT_PHASE.json` (primary)  
2) `docs/plan/NEXT_PHASE.json` (fallback)

If neither exists: **STOP and report**.

Expected format:

```json
{
  "plan_version": "v1.02",
  "next": 12
}
```

---

## Hard Guardrails (non-negotiable)

- **Do NOT modify normative specs** or anything under:
  - `docs/product/agentos-prs/`
  - `docs/api/`
- Implement **ONLY** the current phase document corresponding to `NEXT_PHASE.json`.
- Keep changes **minimal and surgical** (no drive-by refactors).
- Preserve the **Docker-in-network** invariant:
  - CI and E2E must run **inside Docker network**; do not curl localhost services from the CI host.
- **One commit per phase.**
- **One PR per phase.**
- After opening the PR: **STOP** (do not start the next phase).
- If you believe a guardrail must be violated: **STOP and report** what you need and why.

---

## Phase selection

1) Read `PHASE_NUM` from the first existing file in:
   - `configs/NEXT_PHASE.json`, else
   - `docs/plan/NEXT_PHASE.json`

2) Read `plan_version` from the same file.

3) Let:
   - `PHASE_DOC = docs/plan/${plan_version}/phases/phase-${PHASE_NUM}.md`

4) If `PHASE_DOC` does not exist: **STOP and report**.

---

## Branch + PR rules

- Ensure `origin` points to the canonical repo:
  - `https://github.com/NexixAI/nexixai-agentos-platform.git`
- Create a fresh branch from `origin/main`:
  - `phase-${PHASE_NUM}`
- Commit message:
  - `feat(phase${PHASE_NUM}): implement phase ${PHASE_NUM}`
- PR title:
  - `Phase ${PHASE_NUM}`
- Do **not** merge the PR.

---

## Workflow (must follow exactly)

### A) Preflight (clean, deterministic base)

Run:

- `git remote set-url origin https://github.com/NexixAI/nexixai-agentos-platform.git`
- `git fetch origin`
- `git checkout main`
- `git reset --hard origin/main`
- `git clean -fd`
- `git checkout -b phase-${PHASE_NUM}`

If any step fails: **STOP and report**.

---

### B) Implement ONLY what `PHASE_DOC` requires

- Read `PHASE_DOC` in full.
- Make the smallest set of changes required to satisfy it.
- If you think you need to touch forbidden directories: **STOP and report**.
- Do not edit other phase docs.

---

### C) Format + tests (required gates)

- Run `gofmt` **only on files you changed**, unless the phase doc explicitly requires broader formatting.
- Run:
  - `go test ./...`
- Run federation 2-node E2E from repo root:
  - `docker compose -f deploy/local/compose.federation-2node.yaml -f deploy/local/compose.federation-2node.ADDENDUM.yaml up --build --abort-on-container-exit federation-e2e`
  - `docker compose -f deploy/local/compose.federation-2node.yaml -f deploy/local/compose.federation-2node.ADDENDUM.yaml down -v`

If any test fails: fix only what is necessary, re-run gates, and continue.

---

### D) Report (include in final response)

Include:

- Exact `PHASE_NUM`
- Exact `PHASE_DOC`
- Changed files list
- Commands run + exit codes
- Failures encountered + what fixed them (if any)
- What remains (if anything)

---

### E) Commit + PR + STOP

- Stage only relevant changes:
  - Prefer `git add <files...>` over `git add -A`
- Create exactly **one** commit:
  - `git commit -m "feat(phase${PHASE_NUM}): implement phase ${PHASE_NUM}"`
- Push:
  - `git push -u origin phase-${PHASE_NUM}`
- Open PR (use GitHub CLI if available):
  - `gh pr create --base main --head phase-${PHASE_NUM} --title "Phase ${PHASE_NUM}" --body "Implements docs/plan/${plan_version}/phases/phase-${PHASE_NUM}.md. One phase, one commit. Tests: go test ./... + federation 2-node E2E."`

Then **STOP**.

---

## Notes on SPEC_AUTHORITY

Codex must treat `SPEC_AUTHORITY.md` as the conflict resolver.
If code/spec mismatch is discovered:
- Do not invent fields/semantics in code.
- Stop and report if the fix would require modifying normative docs (`docs/product/agentos-prs/` or `docs/api/`).

---
