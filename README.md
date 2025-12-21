# Federation E2E CI Fix Bundle

This bundle contains two updated files to fix the `federation-e2e` GitHub Actions failure where containers start but `go` is reported as missing.

## What was broken
- The compose file used `bash -lc ...` to run services. The `-l` (login shell) causes `/etc/profile` to reset `PATH`, dropping `/usr/local/go/bin` on Debian-based images, so `go` becomes `command not found`.
- The compose file mounted `../..:/workspace`. Docker Compose resolves relative paths from the **project directory** (usually the repo root), so in CI this mounted the wrong host directory.

## What changed
- `deploy/local/compose.federation-2node.yaml`
  - Uses `command: ["go", ...]` (no login shell)
  - Uses `volumes: [".:/workspace"]` so it mounts the repo root correctly in CI and locally (when run from repo root).
- `.github/workflows/federation-e2e.yml`
  - Improves health waiting and prints logs on failure.

## Apply
Copy the two files into your repo at the same paths, commit, and re-run the workflow.
