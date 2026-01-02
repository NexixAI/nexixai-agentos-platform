# Kubernetes Packaging (Track 04)

This doc outlines the optional Kubernetes packaging path delivered in Track 04. It uses Kustomize (via `kubectl kustomize`) and keeps secrets external.

## Layout
- `deploy/k8s/base/` — core manifests for Agent Orchestrator, Model Policy, Federation, services, and peer registry ConfigMap.
- `deploy/k8s/overlays/local-dev/` — example overlay that sets `AGENTOS_PROFILE=dev`, seeds a default tenant, and generates a placeholder secrets Secret from `secrets.env.example`.

## Image & ports
- Default image: `ghcr.io/nexixai/agentos:latest` (override via `kustomize edit set image ghcr.io/nexixai/agentos:latest=myimage:tag`).
- Services expose cluster ports 50081/50082/50083 mapped to container ports 8081/8082/8083.

## Secrets injection
- Base expects a Secret named `agentos-secrets` (keys: `AGENTOS_SERVICE_TOKEN`, `AGENTOS_FED_SHARED_SECRET`, `AGENTOS_SIGNING_KEY`, etc.).
- Local overlay generates this Secret from `secrets.env.example`. Replace with real values or use external secret manager projections/CSI drivers.
- Pods also accept `_FILE` variants for secret paths (see `internal/secrets.Loader`).

## Persistence
- Defaults use `emptyDir` for data/audit/forward-index paths. For production, swap to PVCs by patching the `data` volume mounts.

## Observability
- Metrics endpoints are enabled and require tenant auth (`AGENTOS_METRICS_REQUIRE_AUTH=1` in base).
- To scrape, add a ServiceMonitor/PodMonitor in your cluster (not included here).

## Peers / federation
- `peers-configmap.yaml` mounts a seed file at `/etc/agentos/peers/peers.seed.json` with a single local stack. Add remote peers by editing the ConfigMap or patching with an overlay.

## Smoke path (example)
```bash
# Render manifests
kubectl kustomize deploy/k8s/overlays/local-dev > /tmp/agentos.yaml

# Apply (optional; requires a cluster)
# kubectl apply -f /tmp/agentos.yaml

# Port-forward (if applied)
# kubectl -n agentos port-forward svc/agentos-agent-orchestrator 50081:50081 &
# curl -H \"X-Tenant-Id: tnt_demo\" http://127.0.0.1:50081/v1/health
```

## Rotation guidance
- Update Secret values or `_FILE` mounts and re-apply; restart Deployments to pick up new credentials.
- Federation shared secret lives in `agentos-secrets` (key `AGENTOS_FED_SHARED_SECRET`); update both nodes consistently.

## Validation
- Render check: `kubectl kustomize deploy/k8s/base` and `kubectl kustomize deploy/k8s/overlays/local-dev` (no cluster required).
