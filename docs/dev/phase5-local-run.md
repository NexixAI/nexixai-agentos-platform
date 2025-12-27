# Phase 5 — Local run (Go scaffold)

This repo includes minimal Go servers for Agent Orchestrator, Model Policy, and Federation.

## Build / run

From repo root:

```bash
go run ./cmd/agentos serve agent-orchestrator --addr :50081
go run ./cmd/agentos serve model-policy --addr :50082
go run ./cmd/agentos serve federation --addr :50083
```

## Quick smoke

Agent Orchestrator:
```bash
curl -s http://127.0.0.1:50081/v1/health | jq .
curl -s -X POST http://127.0.0.1:50081/v1/agents/agt_123/runs       -H 'Content-Type: application/json'       -d @docs/api/agent-orchestrator/examples/runs-create.request.json | jq .
```

Model Policy:
```bash
curl -s http://127.0.0.1:50082/v1/health | jq .
curl -s http://127.0.0.1:50082/v1/models | jq .
curl -s -X POST http://127.0.0.1:50082/v1/models:invoke       -H 'Content-Type: application/json'       -d @docs/api/model-policy/examples/chat.request.json | jq .
```

Federation:
```bash
curl -s http://127.0.0.1:50083/v1/federation/health | jq .
curl -s http://127.0.0.1:50083/v1/federation/peers | jq .
curl -N http://127.0.0.1:50083/v1/federation/runs/run_demo/events
```
