# Phase 8 — Federation v1 (two nodes working)

This phase proves the **“connect stacks”** story end-to-end: Node A forwards a Run to Node B and the client can stream events via Node A.

## What this adds

- A 2-node local compose file: `deploy/local/compose.federation-2node.yaml`
- Peer registry seed file: `deploy/local/peers.seed.json`
- Federation implementation:
  - `POST /v1/federation/runs:forward` calls **remote Stack A** `POST /v1/agents/{agent_id}/runs`
  - `GET /v1/federation/runs/{run_id}/events` acts as an **SSE proxy** to the remote run events stream
  - `POST /v1/federation/events:ingest` stores events and enforces **dedupe by event_id** + monotonic `sequence`
  - `GET /v1/federation/peer` + `/peer/capabilities` implemented per OpenAPI

## Run it (local)

```bash
docker compose -f deploy/local/compose.federation-2node.yaml up -d
docker compose -f deploy/local/compose.federation-2node.yaml ps
```

Health:

```bash
curl -s http://127.0.0.1:50083/v1/federation/health
curl -s http://127.0.0.1:50084/v1/health
```

Forward a run (Node A -> Node B):

```bash
curl -sS -X POST http://127.0.0.1:50083/v1/federation/runs:forward \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: tnt_demo' \
  -d @docs/api/federation/examples/runs-forward.request.json
```

Stream proxied events (Node A proxies Node B):

```bash
# Replace RUN_ID with forwarded.remote_run_id from the forward response
curl -N http://127.0.0.1:50083/v1/federation/runs/RUN_ID/events -H 'X-Tenant-Id: tnt_demo'
```

## Notes

- Peer registry is file-based in Phase 8: set `AGENTOS_PEERS_FILE`.
- Two-node compose uses the official `golang` image and `go run` for convenience.
- Phase 9 can swap registry storage + add mTLS/rate limits.

Updated: 2025-12-21
