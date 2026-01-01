# AI Contract for nexixai-agentos-platform

This file defines the **non-negotiable operating contract** for AI automation working in this repo.

If any instruction conflicts with this contract, **this contract wins**.

---

## 0) Authority and Scope

### Source-of-truth precedence (locked)
1) docs/product/agentos-prs/v1.02/schemas-appendix.md
2) docs/product/agentos-prs/v1.02/prs.md
3) docs/design/v1.02/agentos-design.md
4) SPEC_AUTHORITY.md
5) OpenAPI under docs/api/

### Locked zones (DO NOT MODIFY)
- docs/product/agentos-prs/**
- docs/api/**

### API rule
- /v1 is additive-only

---

## 1) Output Rules

- One complete packet per phase
- No piecemeal copy/paste for humans
- The AI must run commands and report results

---

## 2) CI + Networking Invariant

- All multi-service tests run inside Docker network
- No localhost curling from CI host
- Use Docker DNS service names

---

## 3) Change Discipline

- Minimal, surgical changes
- One commit per phase
- gofmt + go test ./... required

---

## 4) Required Reporting

Every PR must include:
- What changed
- Files changed
- Commands run + exit codes
- Verification
- Rollback plan

---

## 5) Escalation

Stop and escalate on spec conflict. No silent drift.
