package types

// Tenant represents a minimal tenant admin record for Phase 14.
type Tenant struct {
	TenantID     string         `json:"tenant_id"`
	Name         string         `json:"name,omitempty"`
	Status       string         `json:"status,omitempty"`
	PlanTier     string         `json:"plan_tier,omitempty"`
	Entitlements map[string]any `json:"entitlements,omitempty"`
	Quotas       map[string]any `json:"quotas,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Policy       *TenantPolicy  `json:"policy,omitempty"`
	CreatedAt    string         `json:"created_at,omitempty"`
	UpdatedAt    string         `json:"updated_at,omitempty"`
}

// TenantPolicy defines per-tenant model access and usage policies.
type TenantPolicy struct {
	AllowedModels []string     `json:"allowed_models,omitempty"`
	DeniedModels  []string     `json:"denied_models,omitempty"`
	TokenBudget   *TokenBudget `json:"token_budget,omitempty"`
}

// TokenBudget defines token usage limits per tenant.
type TokenBudget struct {
	MaxTokensPerHour int `json:"max_tokens_per_hour,omitempty"`
	MaxTokensPerDay  int `json:"max_tokens_per_day,omitempty"`
}
