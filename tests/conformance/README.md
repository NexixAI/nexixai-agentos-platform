# Conformance Tests (Phase 4)

These tests enforce "spec-first" discipline by validating canonical JSON examples against the OpenAPI schemas.

## Run locally

From repo root:

```bash
python -m pip install -r tests/conformance/requirements.txt
python tests/conformance/run_conformance.py
```

## What it checks

- Parses OpenAPI specs:
  - docs/api/stack-a/openapi.yaml
  - docs/api/stack-b/openapi.yaml
  - docs/api/federation/openapi.yaml
- Validates canonical example JSON payloads under docs/api/**/examples/ against component schemas.
- Resolves both internal and external `$ref` pointers (including federation refs back to stack-a schemas).

If any example no longer matches the contract, the run exits non-zero (CI will fail).
