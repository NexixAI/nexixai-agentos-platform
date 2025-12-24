# Track 03 — Secrets management integration (prod path)

## Goal
Provide a production path that avoids plaintext secrets in repo and supports rotation.

## Scope
- Document and implement a minimal secret-loading interface:
  - env var references (default)
  - file-based secrets (Docker/Kubernetes friendly)
  - optional external secret manager integration hooks (documented)
- Ensure secrets do not leak into logs.

Do not change normative docs under `docs/product/agentos-prs/` or `docs/api/`.

## Required outcomes
1) **No plaintext secrets committed**
- Compose files must use placeholders and/or file mounts.
- Example secrets should be clearly fake.

2) **Rotation guidance**
- Document how to rotate:
  - service tokens
  - federation credentials
  - any signing keys

3) **No secret leakage**
- Add tests (or log-scrape checks) proving secrets are not emitted in logs at startup.

## Required gates
- `go test ./...`
- If a docker-compose example is updated: verify the stack still starts and smoke passes.

## Deliverables
- Secret-loading mechanism (minimal)
- Docs under `docs/plan/v1.02/tracks/` explaining the prod path
