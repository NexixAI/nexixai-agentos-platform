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
