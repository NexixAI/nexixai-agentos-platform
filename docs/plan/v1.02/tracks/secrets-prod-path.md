# Secrets — Production Path (Track 03)

This doc describes the production secrets path for AgentOS v1.02. It is non-normative and complements the track doc.

## Secret loading (code)
- Use `internal/secrets.Loader`:
  - Env var first: `SECRET_NAME`
  - File fallback: `SECRET_NAME_FILE` (Docker/Kubernetes secret mounts)
  - Optional external provider hook via `secrets.WithExternal(func(name) (string, error) { ... })`
- Loader never logs secret values. Optional logger can record source (env/file/external) for audit.
- Use `Load` for optional secrets; `Require` when startup must fail if missing.

## Compose / runtime injection
- `deploy/local/compose.yaml` and `deploy/local/compose.federation-2node.yaml` include `env_file: ./secrets.example.env` placeholders.
- Replace `deploy/local/secrets.example.env` with environment-specific secret files or mount secrets at runtime (e.g., `SECRET_NAME_FILE=/run/secrets/...`).
- Do not commit real secrets; example values are intentionally fake.

## Rotation guidance
- **Service tokens** (e.g., API keys):
  1. Create new token in secret manager.
  2. Update `SECRET_NAME` or `SECRET_NAME_FILE` to point to the new version.
  3. Roll restart services; verify via health + smoke.
  4. Remove old token after cutover.
- **Federation credentials** (shared secret or token):
  1. Write new credential to secret store/file used by both nodes.
  2. Deploy node B first, then node A to avoid split-brain.
  3. Confirm federation forward + events ingest succeeds, then revoke old credential.
- **Signing keys** (JWT/signature material):
  1. Store key material in a secret file mount or secret manager.
  2. Provide versioned filenames or secret versions; roll restart.
  3. Audit logs to confirm no key material is emitted.

## No secret leakage
- Tests under `internal/secrets/loader_test.go` ensure loader logs never include secret values.
- Keep application logs at startup free of secrets; only log source or key name when needed.
