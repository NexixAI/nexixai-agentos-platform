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

- Agent Orchestrator: http://127.0.0.1:50081
- Model Policy: http://127.0.0.1:50082
- Federation: http://127.0.0.1:50083

Compose file: `deploy/local/compose.yaml`
