# Phase 9 Update — Hardening pass

Date: 2025-12-21

## What Phase 9 adds

- Alerting stack that works without app `/metrics` (Prometheus + Blackbox + Alertmanager).
- Default alerts route to a local webhook sink (stdout) so you get signal immediately.
- Optional SMTP email wiring is included as a documented config change.
- Opt-in edge rate limiting (Nginx) with per-service request limiting and connection caps.
- Version pinning in the added deployment manifests.

## Exit criteria

- Alerting stack runs locally and fires a test “service down” alert if you stop a service.
- Edge rate-limit stack can be enabled without breaking existing endpoints.
- Docs clearly explain how to enable email alerts.
