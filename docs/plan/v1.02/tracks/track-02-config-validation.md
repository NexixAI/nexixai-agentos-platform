# Track 02 — Production-grade configuration validation

## Goal
Fail fast on misconfiguration and make diagnosis obvious.
No silent defaults that reduce safety.

## Scope
- Add validation around runtime config/env parsing.
- Add tests that cover invalid/missing configuration.
- Add a documented configuration matrix.

Do not change normative docs under `docs/product/agentos-prs/` or `docs/api/`.

## Required outcomes
1) **Fail-fast**
- If a required configuration is missing in `prod` mode, the service must refuse to start.
- Error messages must name the exact variable/key and expected format.

2) **No unsafe defaults**
- Avoid defaulting to “allow” when a policy gate is missing.
- Avoid defaulting to weak auth when auth is expected.

3) **Configuration matrix doc**
- Add a doc listing key env vars, their meaning, and which modes require them (e.g., `docs/plan/v1.02/tracks/config-matrix.md`).

## Required gates
- `go test ./...`
- If config changes affect compose: bring up the stack and run the local smoke script.

## Deliverables
- Validation logic + tests
- `docs/plan/v1.02/tracks/config-matrix.md` describing the config surface
