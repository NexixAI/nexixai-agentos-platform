# Execution Tracks (post-v1.02)

Tracks are **execution workstreams** that improve deployability and operability without changing the normative product surface.
They are executed as PRs, but they are **not** PRS phases.

Guardrails:
- Do not change normative docs (`docs/product/agentos-prs/`, `docs/api/`) unless a change is explicitly required by `SPEC_AUTHORITY.md`.
- Prefer docs + scripts + deploy topology fixes over product behavior changes.
- Preserve Phase 0–16 invariants: tenancy enforcement, audit durability, federation invariants, CI docker-network invariant.

Single knob:
- `configs/NEXT_TRACK.json` controls which track runs next.

## Track list

### Deployment & Operations (Tracks 1-4)
1. Track 01 — Local deployment parity with CI
2. Track 02 — Production-grade configuration validation
3. Track 03 — Secrets management integration (prod path)
4. Track 04 — Optional Helm / cloud packaging

### Gap Closure (Tracks 5-9)
These tracks implement missing features identified in gap analysis 2026-01-03:

5. Track 05 — Idempotency enforcement (PRS §4.3)
6. Track 06 — Run lifecycle completion (PRS §4.2 cancel endpoint)
7. Track 07 — Agent registry (PRS §4.2 metadata endpoints)
8. Track 08 — Policy engine implementation (PRS §6.3 enforcement)
9. Track 09 — Federation security hardening (PRS §7.4 mTLS/JWT)
