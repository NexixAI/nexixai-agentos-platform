# NexixAI AgentOS v1.02 Schemas Appendix

This appendix defines **implementation-ready JSON schemas and canonical payloads** for AgentOS v1.02. It is intended to eliminate interpretation drift across services, stacks, and federated nodes.

**Conventions**
- All timestamps: RFC3339 (`2025-12-19T22:15:30Z`)
- IDs: opaque strings (globally unique)
- Streaming: SSE events are JSON objects serialized as `data: <json>\n\n`
- Error model is consistent across all endpoints
- `tenant_id` is mandatory in **Auth Context** and appears in all objects’ metadata
- All payloads are JSON unless otherwise noted

---

## A) Common Types

### A.1 ID Types
```json
{
  "tenant_id": "tnt_demo",
  "agent_id": "agt_review_summarizer_v1",
  "run_id": "run_01JH9Y0Q2KQ4R7S7X6G0M3E4F8",
  "step_id": "stp_01JH9Y0Q4YQZ2R2FQJ8K9N8M2P",
  "event_id": "evt_01JH9Y0Q7T1WZ5C0J1D8K9L0MN",
  "tool_id": "tool.http.fetch",
  "policy_id": "pol_default",
  "request_id": "req_01JH9Y0Q9K7PZJH1K8M9N0P2Q3"
}
```

### A.2 Correlation + Tracing
All request/response bodies MAY include these fields; responses MUST include `correlation_id`.
```json
{
  "correlation_id": "corr_01JH9Y0QA5N2K4P8S9T0V1W2X3",
  "traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
}
```

### A.3 Standard Error Object
Returned with appropriate HTTP status codes.
```json
{
  "error": {
    "code": "quota_exceeded",
    "message": "Tenant token budget exceeded for current window.",
    "retryable": false,
    "details": {
      "tenant_id": "tnt_demo",
      "limit": 100000,
      "used": 100450,
      "window": "day"
    }
  },
  "correlation_id": "corr_01JH9Y0QA5N2K4P8S9T0V1W2X3"
}
```

Recommended `error.code` values (non-exhaustive):
- `unauthorized`, `forbidden`, `not_found`, `conflict`, `invalid_request`
- `dependency_unavailable`, `timeout`, `rate_limited`, `quota_exceeded`
- `policy_blocked`, `policy_redacted`, `internal_error`

---

## B) Auth Context (Mandatory Internal Contract)

### B.1 Auth Context Object
This object is the **required internal contract** carried through Stack A ports and Stack B requests (derived from JWT/OIDC or service token).
```json
{
  "auth": {
    "tenant_id": "tnt_demo",
    "principal": {
      "principal_id": "usr_12345",
      "principal_type": "user",
      "display_name": "Demo User"
    },
    "scopes": [
      "runs:create",
      "runs:read",
      "runs:cancel"
    ],
    "roles": [
      "tenant_user"
    ],
    "session": {
      "session_id": "ses_01JH9Y0QABCD1234",
      "ip": "203.0.113.10",
      "user_agent": "NexixAIClient/1.0"
    },
    "policy_context": {
      "data_classification": "internal",
      "region": "us-west-2",
      "purpose": "product_inference",
      "consent_tags": ["none"]
    }
  }
}
```

### B.2 Service-to-Service Auth Context
```json
{
  "auth": {
    "tenant_id": "tnt_demo",
    "principal": {
      "principal_id": "svc_orders_api",
      "principal_type": "service",
      "display_name": "Orders API"
    },
    "scopes": [
      "runs:create",
      "runs:read"
    ],
    "roles": [
      "tenant_service"
    ],
    "policy_context": {
      "data_classification": "confidential",
      "region": "us-west-2",
      "purpose": "backend_workflow"
    }
  }
}
```

---

## C) Tenant Objects + Quota Configuration

### C.1 Tenant Object
```json
{
  "tenant": {
    "tenant_id": "tnt_demo",
    "name": "Demo Tenant",
    "status": "active",
    "plan_tier": "demo",
    "created_at": "2025-12-19T22:00:00Z",
    "updated_at": "2025-12-19T22:00:00Z",
    "default_policy_id": "pol_default",
    "entitlements": {
      "model_tiers_allowed": ["fast", "balanced"],
      "model_backends_allowed": ["ollama", "openai"],
      "tool_ids_allowed": ["tool.http.fetch", "tool.kv.put", "tool.kv.get"]
    },
    "quotas": {
      "runs": {
        "max_concurrent_runs": 10,
        "run_create_qps": 5,
        "run_create_burst": 10
      },
      "tools": {
        "tool_calls_per_minute": 120
      },
      "models": {
        "token_budget": {
          "window": "day",
          "max_tokens": 100000,
          "max_cost_usd": 25.0
        },
        "max_tokens_per_request": 4096
      },
      "storage": {
        "max_memory_bytes": 1073741824
      }
    },
    "metadata": {
      "tags": ["seeded", "demo"],
      "billing_account": null
    }
  }
}
```

### C.2 Quota Config Schema (Canonical)
Use this schema for tenant quotas in Stack A and Stack B enforcement. (All fields optional unless otherwise stated by plan tier.)
```json
{
  "quota_config": {
    "runs": {
      "max_concurrent_runs": 10,
      "run_create_qps": 5,
      "run_create_burst": 10
    },
    "tools": {
      "tool_calls_per_minute": 120
    },
    "models": {
      "token_budget": {
        "window": "minute",
        "max_tokens": 2000,
        "max_cost_usd": 1.0
      },
      "max_tokens_per_request": 4096,
      "max_requests_per_minute": 120
    },
    "storage": {
      "max_memory_bytes": 1073741824
    }
  }
}
```

### C.3 Tenant Create Request (Admin)
```json
{
  "tenant_create": {
    "tenant_id": "tnt_acme",
    "name": "Acme Corp",
    "plan_tier": "standard",
    "default_policy_id": "pol_default",
    "entitlements": {
      "model_tiers_allowed": ["fast", "balanced", "premium"],
      "model_backends_allowed": ["openai", "anthropic", "ollama"],
      "tool_ids_allowed": ["tool.http.fetch", "tool.sql.query"]
    },
    "quotas": {
      "runs": { "max_concurrent_runs": 50, "run_create_qps": 25, "run_create_burst": 50 },
      "models": { "token_budget": { "window": "day", "max_tokens": 500000, "max_cost_usd": 250.0 } }
    }
  }
}
```

---

## D) Stack A Public API Schemas (Run + Event)

### D.1 Run Create Request (POST `/v1/agents/{agent_id}/runs`)
```json
{
  "input": {
    "type": "text",
    "text": "Summarize customer feedback from the last 24 hours and draft an action list."
  },
  "context": {
    "locale": "en-US",
    "timezone": "America/Los_Angeles",
    "channel": "web",
    "labels": {
      "product": "dashboard",
      "team": "growth"
    }
  },
  "tooling": {
    "tool_allowlist": ["tool.http.fetch", "tool.sql.query"],
    "tool_denymap": []
  },
  "run_options": {
    "priority": "normal",
    "timeout_ms": 180000,
    "max_steps": 30,
    "stream_events": true,
    "dry_run": false
  },
  "idempotency_key": "idem_8bfb6c2d-2b6d-4c7d-8f2e-1b2a0a0e7c51"
}
```


**Notes**
- `tenant_id` is derived from auth (JWT/service token). If supporting explicit headers, `X-Tenant-Id` is allowed for service-to-service calls only (policy choice), but internal `auth.tenant_id` remains authoritative.

### D.2 Run Create Response
```json
{
  "run": {
    "tenant_id": "tnt_demo",
    "agent_id": "agt_review_summarizer_v1",
    "run_id": "run_01JH9Y0Q2KQ4R7S7X6G0M3E4F8",
    "status": "queued",
    "created_at": "2025-12-19T22:15:30Z",
    "events_url": "/v1/runs/run_01JH9Y0Q2KQ4R7S7X6G0M3E4F8/events",
    "output": null,
    "error": null
  },
  "correlation_id": "corr_01JH9Y0QA5N2K4P8S9T0V1W2X3"
}
```

### D.3 Run Object (Canonical)
```json
{
  "run": {
    "tenant_id": "tnt_demo",
    "agent_id": "agt_review_summarizer_v1",
    "run_id": "run_01JH9Y0Q2KQ4R7S7X6G0M3E4F8",
    "status": "running",
    "created_at": "2025-12-19T22:15:30Z",
    "started_at": "2025-12-19T22:15:32Z",
    "completed_at": null,
    "run_options": {
      "priority": "normal",
      "timeout_ms": 180000,
      "max_steps": 30
    },
    "summary": {
      "steps_total": 7,
      "tool_calls_total": 3,
      "model_calls_total": 4
    },
    "output": {
      "type": "text",
      "text": null
    },
    "error": null,
    "correlation_id": "corr_01JH9Y0QA5N2K4P8S9T0V1W2X3"
  }
}
```

### D.4 Event Envelope (Canonical)
All events are tenant-scoped and must include these top-level fields.
```json
{
  "event": {
    "event_id": "evt_01JH9Y0Q7T1WZ5C0J1D8K9L0MN",
    "sequence": 12,
    "time": "2025-12-19T22:15:36Z",
    "type": "agentos.run.step.started",
    "tenant_id": "tnt_demo",
    "agent_id": "agt_review_summarizer_v1",
    "run_id": "run_01JH9Y0Q2KQ4R7S7X6G0M3E4F8",
    "step_id": "stp_01JH9Y0Q4YQZ2R2FQJ8K9N8M2P",
    "trace": {
      "traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
      "span_id": "00f067aa0ba902b7"
    },
    "payload": {
      "name": "plan",
      "message": "Planning next actions"
    }
  }
}
```

### D.5 Required Event Types + Payloads

#### D.5.1 Run Created
```json
{
  "event": {
    "event_id": "evt_...",
    "sequence": 1,
    "time": "2025-12-19T22:15:30Z",
    "type": "agentos.run.created",
    "tenant_id": "tnt_demo",
    "agent_id": "agt_review_summarizer_v1",
    "run_id": "run_...",
    "step_id": null,
    "trace": { "traceparent": "00-...-...-01", "span_id": "..." },
    "payload": {
      "status": "queued"
    }
  }
}
```

#### D.5.2 Step Started
```json
{
  "event": {
    "event_id": "evt_...",
    "sequence": 5,
    "time": "2025-12-19T22:15:33Z",
    "type": "agentos.run.step.started",
    "tenant_id": "tnt_demo",
    "agent_id": "agt_...",
    "run_id": "run_...",
    "step_id": "stp_...",
    "trace": { "traceparent": "00-...-...-01", "span_id": "..." },
    "payload": {
      "step_kind": "model",
      "name": "draft_response"
    }
  }
}
```

#### D.5.3 Tool Call Requested
```json
{
  "event": {
    "event_id": "evt_...",
    "sequence": 8,
    "time": "2025-12-19T22:15:35Z",
    "type": "agentos.tool.call.requested",
    "tenant_id": "tnt_demo",
    "agent_id": "agt_...",
    "run_id": "run_...",
    "step_id": "stp_...",
    "trace": { "traceparent": "00-...-...-01", "span_id": "..." },
    "payload": {
      "tool_call_id": "tcall_01JH9Y0QZZZ...",
      "tool_id": "tool.sql.query",
      "input": { "query": "SELECT ...", "params": { "since": "2025-12-18T22:15:30Z" } },
      "timeout_ms": 10000
    }
  }
}
```

#### D.5.4 Tool Call Completed
```json
{
  "event": {
    "event_id": "evt_...",
    "sequence": 9,
    "time": "2025-12-19T22:15:35Z",
    "type": "agentos.tool.call.completed",
    "tenant_id": "tnt_demo",
    "agent_id": "agt_...",
    "run_id": "run_...",
    "step_id": "stp_...",
    "trace": { "traceparent": "00-...-...-01", "span_id": "..." },
    "payload": {
      "tool_call_id": "tcall_01JH9Y0QZZZ...",
      "tool_id": "tool.sql.query",
      "status": "ok",
      "output": { "rows": 120, "sample": [] },
      "duration_ms": 231,
      "error": null
    }
  }
}
```

#### D.5.5 Model Call Requested
```json
{
  "event": {
    "event_id": "evt_...",
    "sequence": 10,
    "time": "2025-12-19T22:15:36Z",
    "type": "agentos.model.requested",
    "tenant_id": "tnt_demo",
    "agent_id": "agt_...",
    "run_id": "run_...",
    "step_id": "stp_...",
    "trace": { "traceparent": "00-...-...-01", "span_id": "..." },
    "payload": {
      "model_ref": "tier:balanced",
      "operation": "chat",
      "requirements": { "max_latency_ms": 2500, "max_cost_usd": 0.05 },
      "input_summary": { "messages": 6, "has_tools": true }
    }
  }
}
```

#### D.5.6 Model Call Completed
```json
{
  "event": {
    "event_id": "evt_...",
    "sequence": 11,
    "time": "2025-12-19T22:15:37Z",
    "type": "agentos.model.completed",
    "tenant_id": "tnt_demo",
    "agent_id": "agt_...",
    "run_id": "run_...",
    "step_id": "stp_...",
    "trace": { "traceparent": "00-...-...-01", "span_id": "..." },
    "payload": {
      "model_ref": "tier:balanced",
      "operation": "chat",
      "status": "ok",
      "usage": { "input_tokens": 1200, "output_tokens": 220, "total_tokens": 1420 },
      "latency_ms": 890,
      "policy_outcome": { "decision": "allowed", "reasons": [], "redactions": [] },
      "route": { "backend": "ollama", "model": "qwen2.5:7b", "node": "gpu-node-1" }
    }
  }
}
```

#### D.5.7 Run Completed
```json
{
  "event": {
    "event_id": "evt_...",
    "sequence": 20,
    "time": "2025-12-19T22:15:45Z",
    "type": "agentos.run.completed",
    "tenant_id": "tnt_demo",
    "agent_id": "agt_...",
    "run_id": "run_...",
    "step_id": null,
    "trace": { "traceparent": "00-...-...-01", "span_id": "..." },
    "payload": {
      "status": "completed",
      "output": { "type": "text", "text": "Summary...\nActions..." }
    }
  }
}
```

#### D.5.8 Run Failed
```json
{
  "event": {
    "event_id": "evt_...",
    "sequence": 20,
    "time": "2025-12-19T22:15:45Z",
    "type": "agentos.run.failed",
    "tenant_id": "tnt_demo",
    "agent_id": "agt_...",
    "run_id": "run_...",
    "step_id": null,
    "trace": { "traceparent": "00-...-...-01", "span_id": "..." },
    "payload": {
      "status": "failed",
      "error": {
        "code": "dependency_unavailable",
        "message": "Model Services Stack unavailable",
        "retryable": true,
        "details": { "service": "stack-b" }
      }
    }
  }
}
```

---

## E) Stack B (Model Services Stack) Schemas

Stack B provides governed, vendor-agnostic model access for Stack A and (optionally) other trusted internal callers.

**Contract notes**
- OpenAI-compat surfaces are allowed, but the **AgentOS canonical envelope** is used for Stack A ↔ Stack B calls.
- Stack B always returns `correlation_id` and may return `traceparent` for end-to-end tracing.

### E.1 Model Invoke Request (Chat)
(OpenAI-compat surface allowed; this is the **AgentOS canonical enrichment**.)

```json
{
  "auth": {
    "tenant_id": "tnt_demo",
    "principal": { "principal_id": "svc_stack_a", "principal_type": "service" },
    "scopes": ["models:invoke"],
    "policy_context": { "data_classification": "internal", "region": "us-west-2", "purpose": "agent_run" }
  },
  "request": {
    "request_id": "req_01JH9Y0Q9K7PZJH1K8M9N0P2Q3",
    "operation": "chat",
    "model_ref": "tier:balanced",
    "requirements": {
      "max_latency_ms": 2500,
      "max_cost_usd": 0.05,
      "min_context_tokens": 0
    },
    "input": {
      "messages": [
        { "role": "system", "content": "You are an assistant." },
        { "role": "user", "content": "Summarize the top themes in these reviews." }
      ],
      "tools": [
        {
          "tool_id": "tool.sql.query",
          "name": "sql_query",
          "input_schema": {
            "type": "object",
            "properties": { "query": { "type": "string" } },
            "required": ["query"]
          }
        }
      ],
      "tool_choice": "auto"
    },
    "generation": {
      "temperature": 0.2,
      "top_p": 0.9,
      "max_output_tokens": 600
    }
  },
  "traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
}
```

### E.2 Model Invoke Response (Chat)
```json
{
  "response": {
    "request_id": "req_01JH9Y0Q9K7PZJH1K8M9N0P2Q3",
    "status": "ok",
    "output": {
      "type": "chat",
      "message": { "role": "assistant", "content": "Here are the actions..." },
      "tool_calls": []
    },
    "usage": { "input_tokens": 1200, "output_tokens": 220, "total_tokens": 1420 },
    "latency_ms": 890,
    "policy_outcome": {
      "decision": "allowed",
      "reasons": [],
      "redactions": []
    },
    "route": {
      "backend": "ollama",
      "model": "qwen2.5:7b",
      "node": "gpu-node-1",
      "fallback_used": false
    }
  },
  "correlation_id": "corr_01JH9Y0QA5N2K4P8S9T0V1W2X3",
  "traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
}
```

### E.3 Policy Check Request / Response
**Request**
```json
{
  "auth": {
    "tenant_id": "tnt_demo",
    "principal": { "principal_id": "svc_stack_a", "principal_type": "service" },
    "scopes": ["policy:check"],
    "policy_context": { "data_classification": "internal", "region": "us-west-2", "purpose": "agent_run" }
  },
  "policy_check": {
    "request_id": "req_...",
    "content": {
      "type": "text",
      "text": "User input that might contain sensitive data..."
    },
    "ruleset": "pol_default"
  }
}
```

**Response**
```json
{
  "policy_result": {
    "request_id": "req_...",
    "decision": "allowed",
    "reasons": [],
    "redactions": []
  },
  "correlation_id": "corr_..."
}
```

### E.4 Model Invoke Request / Response (Embeddings)
**Request**
```json
{
  "auth": {
    "tenant_id": "tnt_demo",
    "principal": { "principal_id": "svc_stack_a", "principal_type": "service" },
    "scopes": ["models:invoke"],
    "policy_context": { "data_classification": "internal", "region": "us-west-2", "purpose": "agent_run" }
  },
  "request": {
    "request_id": "req_...",
    "operation": "embed",
    "model_ref": "tier:balanced",
    "requirements": {
      "max_latency_ms": 2500,
      "max_cost_usd": 0.02,
      "min_context_tokens": 0
    },
    "input": {
      "texts": ["This movie was surprisingly heartfelt.", "The pacing felt uneven in the middle."],
      "encoding_format": "float"
    }
  },
  "traceparent": "00-...-...-01"
}
```

**Response**
```json
{
  "response": {
    "request_id": "req_...",
    "status": "ok",
    "output": {
      "type": "embedding",
      "embeddings": [
        { "index": 0, "vector": [0.0123, -0.0456, 0.0789] },
        { "index": 1, "vector": [-0.0011, 0.0222, -0.0333] }
      ]
    },
    "usage": { "input_tokens": 40, "output_tokens": 0, "total_tokens": 40 },
    "latency_ms": 120,
    "policy_outcome": { "decision": "allowed", "reasons": [], "redactions": [] },
    "route": { "backend": "ollama", "model": "nomic-embed-text", "node": "gpu-node-1", "fallback_used": false }
  },
  "correlation_id": "corr_...",
  "traceparent": "00-...-...-01"
}
```

### E.5 List Models (Capabilities) Request / Response
**Request**
```json
{
  "auth": {
    "tenant_id": "tnt_demo",
    "principal": { "principal_id": "svc_stack_a", "principal_type": "service" },
    "scopes": ["models:list"],
    "policy_context": { "data_classification": "internal", "region": "us-west-2", "purpose": "agent_run" }
  }
}
```

**Response**
```json
{
  "models": [
    {
      "model_id": "qwen2.5:7b",
      "provider": "ollama",
      "capabilities": ["chat"],
      "max_output_tokens": 2048,
      "context_window_tokens": 32768
    },
    {
      "model_id": "nomic-embed-text",
      "provider": "ollama",
      "capabilities": ["embed"]
    }
  ],
  "correlation_id": "corr_..."
}
```

---

## F) Federation Payloads (Run Forwarding + Events)

### F.1 Peer Info
```json
{
  "peer": {
    "stack_id": "stk_usw2_node_01",
    "environment": "dev",
    "region": "us-west-2",
    "api_versions": ["v1"],
    "endpoints": {
      "stack_a_base_url": "https://node01.dev.nexixai.local",
      "stack_b_base_url": "https://node01.dev.nexixai.local/bedrock"
    },
    "build": { "version": "1.0.2", "git_sha": "abc123", "timestamp": "2025-12-19T21:55:00Z" }
  }
}
```

### F.2 Peer Capabilities
```json
{
  "capabilities": {
    "supports": {
      "runs_forward": true,
      "events_stream": true,
      "tenant_isolation": true,
      "tool_hosting": false
    },
    "limits": {
      "max_concurrent_runs": 200,
      "max_event_streams": 1000
    },
    "models": {
      "tiers": ["fast", "balanced", "premium"],
      "backends": ["ollama", "vllm"],
      "max_context_tokens": 32768
    }
  }
}
```

### F.3 Run Forward Request (POST `/v1/federation/runs:forward`)
```json
{
  "forward": {
    "target_selector": {
      "stack_id": "stk_usw2_node_02",
      "region": "us-west-2",
      "required_capabilities": ["events_stream", "tenant_isolation"],
      "preferred_model_tier": "balanced"
    },
    "auth": {
      "tenant_id": "tnt_demo",
      "principal": { "principal_id": "svc_stack_a", "principal_type": "service" },
      "scopes": ["runs:forward"],
      "policy_context": { "data_classification": "internal", "region": "us-west-2", "purpose": "agent_run" }
    },
    "run_request": {
      "agent_id": "agt_review_summarizer_v1",
      "input": { "type": "text", "text": "Summarize customer feedback..." },
      "context": { "locale": "en-US", "timezone": "America/Los_Angeles" },
      "tooling": { "tool_allowlist": ["tool.sql.query"] },
      "run_options": { "timeout_ms": 180000, "max_steps": 30, "stream_events": true },
      "idempotency_key": "idem_..."
    },
    "traceparent": "00-...-...-01"
  }
}
```

### F.4 Run Forward Response
```json
{
  "forwarded": {
    "tenant_id": "tnt_demo",
    "remote_stack_id": "stk_usw2_node_02",
    "remote_run_id": "run_01JH9YZZZZZZZZZZZZZZZZZZZZ",
    "remote_events_url": "https://node02.dev.nexixai.local/v1/runs/run_01JH9YZZZZZZZZZZZZZZZZZZZZ/events",
    "status": "queued"
  },
  "correlation_id": "corr_..."
}
```

### F.5 Event Replication Ingest (Optional) (POST `/v1/federation/events:ingest`)
```json
{
  "events": [
    {
      "event_id": "evt_...",
      "sequence": 12,
      "time": "2025-12-19T22:15:36Z",
      "type": "agentos.model.completed",
      "tenant_id": "tnt_demo",
      "agent_id": "agt_...",
      "run_id": "run_...",
      "step_id": "stp_...",
      "trace": { "traceparent": "00-...-...-01", "span_id": "..." },
      "payload": { "status": "ok" }
    }
  ],
  "ingest": {
    "source_stack_id": "stk_usw2_node_02",
    "dedupe": { "mode": "event_id" }
  }
}
```

---

## G) Run/Event Storage Keys (Normative Guidance)
To ensure isolation and avoid collisions, implementations MUST scope persisted keys by tenant:

- Runs: `tenant/{tenant_id}/runs/{run_id}`
- Events: `tenant/{tenant_id}/runs/{run_id}/events/{sequence}`
- Agent configs: `tenant/{tenant_id}/agents/{agent_id}`
- Tool registry: `tenant/{tenant_id}/tools/{tool_id}`
- Budgets/usage: `tenant/{tenant_id}/usage/{window}/{date}`

---

## H) Required Deployment Seed Outputs (for reports)
When `agentos up` succeeds, the report MUST include at minimum:

```json
{
  "seeded": {
    "default_tenant_id": "tnt_demo",
    "platform_admin_principal_id": "usr_platform_admin",
    "tenant_admin_principal_id": "usr_tnt_demo_admin",
    "sample_agent_ids": ["agt_sample_echo_v1"],
    "sample_tool_ids": ["tool.echo"]
  }
}
```

---

## I) Compatibility Guarantees (Normative)
- `/v1` payloads are **additive-only**. New fields may be added; existing fields will not change meaning.
- Event envelope fields in **D.4** are mandatory and stable; `payload` varies by `type`.
- Federation forwarding payload **must** include verifiable tenant+principal identity (mTLS + JWT recommended).
