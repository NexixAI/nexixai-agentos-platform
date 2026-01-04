# Track 07 — Agent Registry

## Goal
Implement agent metadata endpoints as specified in PRS §4.2.

## What you are allowed to change
- `agentorchestrator/server.go` — Add agent list/get endpoints
- `internal/storage/` — Add agent store (file-based for v1.02)
- `internal/types/stacka.go` — Add Agent storage type
- OpenAPI spec for agent endpoints
- Seed demo agent in default tenant

Do not change PRS (already specifies agent endpoints).

## Required outcomes

1) **Agent endpoints**
- `GET /v1/agents` — List agents for tenant (tenant-scoped)
- `GET /v1/agents/{agent_id}` — Get agent metadata

2) **Agent object** (minimal for v1.02)
```json
{
  "agent_id": "agt_demo",
  "tenant_id": "tnt_demo",
  "name": "Demo Agent",
  "description": "Sample agent for validation",
  "version": "1.0",
  "status": "active",
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

3) **Storage**
- File-based: `data/agents/{tenant_id}/{agent_id}.json`
- Tenant isolation enforced

4) **Seeding**
- `agentos up` seeds one demo agent: `agt_demo` in `tnt_demo`
- Update reports to show seeded agent IDs

5) **Behavior**
- `GET /v1/agents` returns array of agent objects for requesting tenant
- `GET /v1/agents/{agent_id}` returns 404 if agent doesn't exist or belongs to different tenant
- Run creation validates that agent_id exists (optional: can defer to future track)

## Required gates
- Unit test: list agents returns only tenant's agents
- Unit test: get agent from different tenant returns 404
- Smoke test: `GET /v1/agents` returns seeded demo agent
- Update OpenAPI with agent endpoints

## Deliverables
- Agent store interface and file implementation
- Agent endpoints
- Demo agent seeding
- Unit tests
- Updated OpenAPI
