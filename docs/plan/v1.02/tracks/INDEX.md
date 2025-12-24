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
1. Track 01 — Local deployment parity with CI
2. Track 02 — Production-grade configuration validation
3. Track 03 — Secrets management integration (prod path)
4. Track 04 — Optional Helm / cloud packaging
