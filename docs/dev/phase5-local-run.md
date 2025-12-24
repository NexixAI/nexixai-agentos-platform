# Phase 5 — Local run (Go scaffold)

This repo includes minimal Go servers for Stack A, Stack B, and Federation.

## Build / run

From repo root:

```bash
go run ./cmd/agentos serve stack-a --addr :50081
go run ./cmd/agentos serve stack-b --addr :50082
go run ./cmd/agentos serve federation --addr :50083
```

## Quick smoke

Stack A:
```bash
curl -s http://127.0.0.1:50081/v1/health | jq .
curl -s -X POST http://127.0.0.1:50081/v1/agents/agt_123/runs       -H 'Content-Type: application/json'       -d @docs/api/stack-a/examples/runs-create.request.json | jq .
```

Stack B:
```bash
curl -s http://127.0.0.1:50082/v1/health | jq .
curl -s http://127.0.0.1:50082/v1/models | jq .
curl -s -X POST http://127.0.0.1:50082/v1/models:invoke       -H 'Content-Type: application/json'       -d @docs/api/stack-b/examples/chat.request.json | jq .
```

Federation:
```bash
curl -s http://127.0.0.1:50083/v1/federation/health | jq .
curl -s http://127.0.0.1:50083/v1/federation/peers | jq .
curl -N http://127.0.0.1:50083/v1/federation/runs/run_demo/events
```
