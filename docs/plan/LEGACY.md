# Legacy Plan Docs (v1.02 migration)

Canonical plan docs now live under `docs/plan/v1.02/`.

The files left at `docs/plan/` and `docs/plan/tracks/` are **compatibility stubs** only. They exist to keep old links working while everything migrates to the versioned topology.

## What is canonical
- Plan index: `docs/plan/v1.02/INDEX.md`
- Phases: `docs/plan/v1.02/phases/phase-*.md`
- Tracks: `docs/plan/v1.02/tracks/track-*.md`

## Legacy surfaces (stubs)
- `docs/plan/INDEX.md`
- `docs/plan/phase-0.md` ... `docs/plan/phase-16.md`
- `docs/plan/tracks/INDEX.md`
- `docs/plan/tracks/track-01-local-deploy-parity.md`
- `docs/plan/tracks/track-02-config-validation.md`
- `docs/plan/tracks/track-03-secrets-prod-path.md`
- `docs/plan/tracks/track-04-helm-cloud-packaging.md`

## Linking rule
- New links MUST point to the canonical v1.02 paths above.
- Stubs should not be referenced in new docs or code.
- If you find a missing canonical file, stop and fix the versioned path instead of leaning on the stub.

## Why keep stubs
- Avoid breaking existing links while teams update bookmarks and scripts.
- Provide a clear pointer to the canonical location during the migration window.

## Maintenance
- Any new phase/track docs should be added only under `docs/plan/v1.02/`.
- Keep stub text consistent; no extra content belongs in legacy paths.
- Run `scripts/docs/check-canonical-links.ps1` before publishing docs changes to catch legacy path regressions.
