# Track 01 — Local deployment parity with CI

## Goal
Make **one-command local bring-up** reliable across Windows/macOS/Linux, while preserving all Phase 0–16 invariants.

This track is complete when a new dev machine can:
- build images
- start the stack
- pass local smoke checks
- stop and clean up reliably

## What you are allowed to change
- `deploy/local/*` compose topology and docs
- `scripts/*` local helpers (PowerShell + bash allowed)
- README sections **only if needed** (minimal diffs; do not rewrite)

Do not change normative docs under `docs/product/agentos-prs/` or `docs/api/`.

## Required outcomes
1) **Port collision avoidance (Windows-friendly)**
- Default host ports must avoid Windows excluded ranges and common conflicts.
- Recommend a standard host-port block:
  - Stack A: 50081
  - Stack B: 50082
- Federation: 50083
- Prometheus: 59090
- Alertmanager: 59093
- Blackbox: 59115
- Grafana: 53000

Host ports use the 5008x block because Windows can exclude 8081/8082/8083. Internal container ports stay 8081/8082/8083, but observability and hardening stacks probe the host ports (50081/50082/50083) to match the local compose mapping.

2) **Credential helper traps**
- Document how to recover when Docker Desktop credential helper breaks.
- Provide a “clean docker config” bypass example for local builds.

3) **Reset / nuke flows**
- Provide a single command (or script) that:
  - stops containers
  - removes the compose network
  - removes the dev volumes for this stack
  - leaves Docker Desktop intact

4) **Local smoke script**
- Provide (or update) a script that:
  - hits Stack A/B health endpoints and federation health
  - uses curl.exe on Windows and forces IPv4
  - has retries with a bounded timeout
  - exits non-zero on failure

## Required gates (run locally)
- `docker compose -f deploy/local/compose.yaml up -d --build`
- Run local smoke script (documented entrypoint)
- `docker compose -f deploy/local/compose.yaml down -v`

## Deliverables
- Updated local compose files (and alerting compose if necessary)
- Updated docs: a short local deploy section describing:
  - ports
  - how to run
  - how to reset
  - how to smoke test
