# Track 08 — Policy Engine Implementation

## Goal
Implement basic policy enforcement as specified in PRS §6.3.

## What you are allowed to change
- `modelpolicy/policy.go` — Implement policy evaluation logic
- `modelpolicy/usage.go` — Add budget enforcement
- `internal/tenants/store.go` — Store per-tenant policy config
- Unit tests for policy engine

Do not change PRS (already specifies policy requirements).

## Required outcomes

1) **Policy gates implemented**

**Baseline for v1.02:**
- Per-tenant allow/deny model list
- Per-tenant token budget enforcement (tokens/day, tokens/hour)
- Blocked requests return 403 with `policy_blocked` error code

**Out of scope for v1.02 (defer to v1.03):**
- PII detection/redaction (requires NLP/regex rules)
- Tool-use permissions (tools not implemented)

2) **Tenant policy configuration**

Add to tenant object:
```json
{
  "policy": {
    "allowed_models": ["qwen2.5:7b", "nomic-embed-text"],
    "denied_models": [],
    "token_budget": {
      "max_tokens_per_hour": 10000,
      "max_tokens_per_day": 100000
    }
  }
}
```

3) **Enforcement points**
- `POST /v1/models:invoke` checks:
  1. Is model in allowed_models? (if list non-empty)
  2. Is model in denied_models? (deny takes precedence)
  3. Has tenant exceeded token budget?
- If any check fails: return 403 with reasons in error details

4) **Budget tracking**
- Track token usage per tenant per hour/day
- Store in `data/usage/{tenant_id}/{YYYY-MM-DD-HH}.json`
- Reset hourly/daily counters automatically

5) **Default policy**
- Default tenant gets allow-all policy (no restrictions)
- New tenants created via admin API can specify policy

## Required gates
- Unit test: denied model returns 403
- Unit test: exceeding token budget returns 403
- Unit test: allowed model with budget remaining succeeds
- Smoke test: invoke with budget exhaustion

## Deliverables
- Policy evaluation logic
- Budget tracking
- Tenant policy configuration
- Unit tests
