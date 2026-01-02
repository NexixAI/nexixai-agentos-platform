# Track 04 — Optional Helm / cloud packaging

## Goal
Provide reproducible Kubernetes deployment templates as an **optional** path.

This track must preserve:
- tenancy isolation
- audit durability
- federation invariants
- stable API surfaces

## Scope
- Helm chart or Kustomize templates under a dedicated folder (e.g. `deploy/k8s/`).
- Minimal “example environment” docs (one cluster, two nodes optional).

Do not change normative docs under `docs/product/agentos-prs/` or `docs/api/`.

## Required outcomes
1) **Reproducible packaging**
- A new user can `helm install` (or apply manifests) and get a working stack.

2) **Clear configuration**
- Values file documents:
  - ports/services
  - persistence
  - secrets injection
  - observability wiring

3) **Evidence**
- Provide a short “smoke” path using port-forward or in-cluster checks.

## Required gates
- Lint/validate the chart/manifests.
- Run `go test ./...` if any Go code changes (prefer not to).

## Deliverables
- `deploy/k8s/*`
- Packaging docs in `docs/plan/v1.02/tracks/`

---

## Implementation notes (v1.02 track delivery)
- Packaging uses Kustomize (Helm alternative) under `deploy/k8s/`:
  - `base/` defines Deployments + Services for Agent Orchestrator, Model Policy, and Federation, plus a federation peers ConfigMap and metrics auth enabled.
  - `overlays/local-dev/` sets `AGENTOS_PROFILE=dev`, seeds a default tenant, and generates a placeholder `agentos-secrets` Secret from `secrets.env.example`.
- Secrets: pods expect a Secret `agentos-secrets`; `_FILE` variants are supported via `internal/secrets.Loader`. Example secret files are fake—replace with real values or external secret manager mounts.
- Persistence: defaults to `emptyDir`; patch to PVCs for production.
- Federation: peers seed file mounted at `/etc/agentos/peers/peers.seed.json`; edit/overlay to add remote peers.
- Usage docs and smoke path are in `docs/plan/v1.02/tracks/k8s-packaging.md` (includes render checks via `kubectl kustomize` and optional port-forward smoke).
