# Phase 9 — Hardening pass (production-ish defaults)

Phase 9 adds **opt-in** hardening components that improve operator safety and “enterprise feel” without changing core contracts.

## What ships in this phase

### 1) Baseline alerting (works out of the box)
- Prometheus + Blackbox Exporter probes health endpoints (no `/metrics` required).
- Alertmanager routes alerts to a default local sink (prints alerts to stdout).
- Optional SMTP email config is provided in-place.

Files:
- `deploy/observability/compose.alerting.yaml`
- `deploy/observability/prometheus/*`
- `deploy/observability/blackbox/*`
- `deploy/observability/alertmanager/alertmanager.yml`

Run:
```bash
docker compose -f deploy/observability/compose.alerting.yaml up -d
```
Access:
- Prometheus: http://127.0.0.1:59090
- Alertmanager: http://127.0.0.1:59093

### 2) Edge rate limiting (opt-in)
- Nginx sits in front of each service and applies basic rate limiting and connection limits.
- This is **not** the final gateway story; it is a safety layer for local/dev and early demos.

Files:
- `deploy/hardening/compose.edge-ratelimit.yaml`
- `deploy/hardening/nginx/*.conf`

Run:
```bash
docker compose -f deploy/hardening/compose.edge-ratelimit.yaml up -d
```
Access:
- Stack A via edge: http://localhost:9081
- Stack B via edge: http://localhost:9082
- Federation via edge: http://localhost:9083

### 3) mTLS guidance (documented)
Phase 9 does **not** force mTLS into the core servers yet. It sets the stage for Phase 10+ by defining:
- where to terminate TLS (edge/gateway),
- what should be internal-only,
- and how to evolve to mutual TLS between nodes.

Recommended next step:
- terminate TLS at the gateway/edge,
- use a private network/VPN between stacks,
- enforce mTLS for federation first (highest risk boundary).

### 4) Version pinning strategy (documented)
The new compose files pin image tags to known working versions. Future upgrades should be:
- explicit (one PR),
- validated (`agentos validate` + federation e2e),
- and rolled back by reverting the pin.

## SMTP email setup (optional)
Edit `deploy/observability/alertmanager/alertmanager.yml`:
- uncomment the SMTP `global:` block,
- create an `email` receiver,
- set `route.receiver: email`.

Tip: use an app-password for Gmail/Google Workspace.

Updated: 2025-12-21
