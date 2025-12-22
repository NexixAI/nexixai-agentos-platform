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
	CreatedAt    string         `json:"created_at,omitempty"`
	UpdatedAt    string         `json:"updated_at,omitempty"`
}
