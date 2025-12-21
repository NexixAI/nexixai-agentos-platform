# Phase 8 Update — Federation v1

Date: 2025-12-21

## What changed in Phase 8

- Added a 2-node local compose to exercise federation on one machine.
- Implemented real forwarding:
  - Node A federation calls Node B Stack A run-create.
  - Node A federation SSE endpoint proxies Node B run events.
- Implemented push-mode ingest:
  - Stores ingested events and enforces dedupe by `event_id` and monotonic `sequence`.
- Added a GitHub Actions federation E2E workflow.

## Exit criteria (Phase 8)

- Node A forwards a run to Node B.
- Client streams events via Node A federation.
- Duplicate event ingest is rejected.
