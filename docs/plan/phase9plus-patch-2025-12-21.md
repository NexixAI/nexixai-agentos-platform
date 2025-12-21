# Phase 9+ Patch — Timeline clarity + native metrics + federation hardening

Date: 2025-12-21

This patch addresses three audit items:

1. **Progress-entry filename dates**
   - Adds `docs/plan/progress-entries/INDEX.md` explaining conventions.
   - Adds `scripts/normalize_progress_entry_dates.sh` to print `git mv` commands to align filenames to internal header dates.
   - Adds a short NOTE to existing entries explaining the above.

2. **Alerting beyond blackbox**
   - Adds native Prometheus `/metrics` endpoints to Stack A / Stack B / Federation.
   - Updates alerting Prometheus config to scrape `/metrics` and adds alert rules for:
     - 5xx ratio, p95 latency, quota denial spikes, federation forward failures.

3. **Federation production semantics gap**
   - Adds persistent forward index (file-backed) to avoid losing mappings after restart.
   - Adds retry/backoff to forward calls and basic retry on SSE proxy connect.
   - Adds additive `from_sequence` query param for federation event streaming filtering.
   - Adds `docs/design/federation-production-notes.md` describing remaining production work.

