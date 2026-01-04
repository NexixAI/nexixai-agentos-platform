# Track 09 — Federation Security Hardening

## Goal
Add mTLS and JWT verification to federation as recommended in PRS §7.4.

## What you are allowed to change
- `federation/forwarder.go` — Add mTLS client config
- `federation/server.go` — Add JWT verification middleware
- `deploy/local/` — Add cert generation for local dev
- Configuration for cert paths
- Documentation for production cert management

Do not change PRS (already recommends mTLS).

## Required outcomes

1) **mTLS client (forwarder)**
- Load client cert/key from env: `AGENTOS_FED_CLIENT_CERT`, `AGENTOS_FED_CLIENT_KEY`
- Verify server cert with CA: `AGENTOS_FED_CA_CERT`
- If certs not configured: fall back to HTTP (dev mode)

2) **mTLS server (listener)**
- Require client cert if `AGENTOS_FED_REQUIRE_MTLS=true`
- Verify client cert with CA
- Extract peer identity from cert (CN or SAN)

3) **JWT verification**
- Verify bearer token signature using `AGENTOS_FED_JWT_PUBLIC_KEY`
- Extract tenant_id and principal_id from claims
- If verification fails: return 401 Unauthorized
- If JWT public key not configured: skip verification (dev mode)

4) **Local dev cert generation**
- Provide script: `scripts/generate-federation-certs.sh`
- Generates self-signed CA + peer certs for local testing
- Updates `.env` with cert paths

5) **Documentation**
- Add federation security section to README
- Document cert generation for production
- Document JWT signing key setup

## Required gates
- Local federation E2E passes with mTLS enabled
- Local federation E2E passes with mTLS disabled (backward compat)
- Unit tests for JWT verification

## Deliverables
- mTLS configuration
- JWT verification middleware
- Cert generation script
- Documentation updates
