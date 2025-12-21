# Phase 6 — Deployment UX v1

Phase 6 introduces a "one command" operator workflow for local deployments using Docker Compose.

## Commands

From repo root:

```bash
# start (build + run) and validate
go run ./cmd/agentos up

# non-destructive validate (health + smoke using canonical examples)
go run ./cmd/agentos validate

# show containers
go run ./cmd/agentos status

# rebuild/restart
go run ./cmd/agentos redeploy

# destructive teardown (keeps volumes by default)
go run ./cmd/agentos nuke

# destructive teardown including volumes
go run ./cmd/agentos nuke --hard
```

## Outputs

Each `up` and `validate` writes a report to:

- `reports/<timestamp>/summary.md`
- `reports/<timestamp>/summary.json`

## Defaults

- Stack A: http://localhost:8081
- Stack B: http://localhost:8082
- Federation: http://localhost:8083

Compose file: `deploy/local/compose.yaml`
