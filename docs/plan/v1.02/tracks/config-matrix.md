# Configuration Matrix (Track 02)

Key AgentOS environment variables and how they should be set per mode.

| Env var | Purpose | Default | Dev/Demo | Prod |
|---------|---------|---------|----------|------|
| `AGENTOS_PROFILE` | Selects profile (`dev`/`demo`/`prod`) | `dev` | Optional | **Required** |
| `AGENTOS_METRICS_REQUIRE_AUTH` | Require tenant auth on `/metrics` | `0` | Optional | **Must be `1`** |
| `AGENTOS_DEFAULT_TENANT` | Seed tenant for local bootstrap | empty | Optional | **Must be empty** |
| `AGENTOS_ALLOW_DEV_HEADERS` | Allow dev-only header bypass | empty | Optional | **Must be empty** |
| `AGENTOS_RUN_STORE_FILE` | Run store path (agent-orchestrator) | `data/agent-orchestrator/runs.json` | Optional | Recommended to set explicit path |
| `AGENTOS_AUDIT_SINK` | Audit sink (`stdout`/`stderr`/`file:PATH`) | `file:data/audit/<service>.audit.log` | Optional | Recommended to set explicit path |
| `AGENTOS_QUOTA_RUN_CREATE_QPS` | Run create QPS limit | `10` | Optional | Optional (set per tenant needs) |
| `AGENTOS_QUOTA_CONCURRENT_RUNS` | Concurrent run limit | `25` | Optional | Optional (set per tenant needs) |
| `AGENTOS_QUOTA_INVOKE_QPS` | Model invoke QPS limit | `20` | Optional | Optional (set per tenant needs) |
| `AGENTOS_FED_FORWARD_INDEX_FILE` | Persistent federation forward index path | `data/federation/forward-index.json` | Optional | Recommended to set explicit path |
| `AGENTOS_PEERS_FILE` | Peer registry JSON (federation) | none | Optional | **Required** |
| `AGENTOS_STACK_ID` | Local stack identifier (federation) | `stk_local` | Optional | Recommended |
| `AGENTOS_ENVIRONMENT` | Environment label | empty | Optional | Recommended |
| `AGENTOS_REGION` | Region label | empty | Optional | Recommended |
| `AGENTOS_FED_FORWARD_MAX_ATTEMPTS` | Forward retry attempts | `3` | Optional | Optional |
| `AGENTOS_FED_FORWARD_BASE_BACKOFF_MS` | Forward base backoff (ms) | `250` | Optional | Optional |
| `AGENTOS_PROM_URL` | Prometheus base URL (optional observability check) | empty | Optional | Optional |
| `AGENTOS_GRAFANA_URL` | Grafana base URL (optional observability check) | empty | Optional | Optional |

Notes:
- Prod mode must fail fast when required values are missing or unreadable (e.g., `AGENTOS_PEERS_FILE`).
- Metrics endpoints must be protected in prod (`AGENTOS_METRICS_REQUIRE_AUTH=1`).
- Do not rely on dev conveniences (`AGENTOS_DEFAULT_TENANT`, `AGENTOS_ALLOW_DEV_HEADERS`) in prod.
