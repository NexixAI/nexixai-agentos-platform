# Access (Local)

After `agentos up`, use:

- Stack A: http://127.0.0.1:50081
  - Health: `/v1/health`
  - Create run: `POST /v1/agents/{agent_id}/runs`
  - Events: `GET /v1/runs/{run_id}/events`

- Stack B: http://127.0.0.1:50082
  - Health: `/v1/health`
  - Models: `/v1/models`
  - Invoke: `POST /v1/models:invoke`
  - Policy check: `POST /v1/policy:check`

- Federation: http://127.0.0.1:50083
  - Health: `/v1/federation/health`
  - Peers: `/v1/federation/peers`
  - Forward: `POST /v1/federation/runs:forward`
  - Events ingest: `POST /v1/federation/events:ingest`
  - Events SSE: `GET /v1/federation/runs/{run_id}/events`

Reports are written to `reports/<timestamp>/`.
