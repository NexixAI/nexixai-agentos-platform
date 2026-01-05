package types

// NOTE: These are minimal structs sufficient for stub handlers.
// They should evolve carefully, guided by the Schemas Appendix + OpenAPI.

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type RunCreateRequest struct {
	Input          RunInput   `json:"input"`
	Context        RunContext `json:"context"`
	Tooling        Tooling    `json:"tooling"`
	RunOptions     RunOptions `json:"run_options"`
	IdempotencyKey string     `json:"idempotency_key"`
}

type RunCreateResponse struct {
	Run           Run    `json:"run"`
	CorrelationID string `json:"correlation_id"`
}

type RunGetResponse struct {
	Run           Run    `json:"run"`
	CorrelationID string `json:"correlation_id"`
}

type Run struct {
	TenantID       string     `json:"tenant_id"`
	AgentID        string     `json:"agent_id"`
	RunID          string     `json:"run_id"`
	Status         string     `json:"status"`
	CreatedAt      string     `json:"created_at"`
	StartedAt      string     `json:"started_at,omitempty"`
	CompletedAt    string     `json:"completed_at,omitempty"`
	EventsURL      string     `json:"events_url"`
	RunOptions     RunOptions `json:"run_options,omitempty"`
	Output         *RunOutput `json:"output,omitempty"`
	Error          *RunError  `json:"error,omitempty"`
	IdempotencyKey string     `json:"idempotency_key,omitempty"`
}

type RunInput struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type RunContext struct {
	Locale    string         `json:"locale"`
	Timezone  string         `json:"timezone"`
	Channel   string         `json:"channel"`
	UserID    string         `json:"user_id,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type Tooling struct {
	Tools []ToolDescriptor `json:"tools"`
}

type ToolDescriptor struct {
	Name        string         `json:"name"`
	Kind        string         `json:"kind"`
	Description string         `json:"description,omitempty"`
	Config      map[string]any `json:"config,omitempty"`
	InputSchema map[string]any `json:"input_schema,omitempty"`
}

type RunOptions struct {
	Priority     string `json:"priority"`
	TimeoutMs    int    `json:"timeout_ms"`
	MaxSteps     int    `json:"max_steps"`
	StreamEvents bool   `json:"stream_events,omitempty"`
}

type RunOutput struct {
	Type      string           `json:"type,omitempty"`
	Text      string           `json:"text,omitempty"`
	Artifacts []map[string]any `json:"artifacts,omitempty"`
}

type RunError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type TraceContext struct {
	Traceparent string `json:"traceparent"`
	SpanID      string `json:"span_id,omitempty"`
	Tracestate  string `json:"tracestate,omitempty"`
}

type EventEnvelope struct {
	Event Event `json:"event"`
}

type Event struct {
	EventID  string         `json:"event_id"`
	Sequence int            `json:"sequence"`
	Time     string         `json:"time"`
	Type     string         `json:"type"`
	TenantID string         `json:"tenant_id"`
	AgentID  string         `json:"agent_id"`
	RunID    string         `json:"run_id"`
	StepID   string         `json:"step_id,omitempty"`
	Trace    TraceContext   `json:"trace"`
	Payload  map[string]any `json:"payload"`
}
