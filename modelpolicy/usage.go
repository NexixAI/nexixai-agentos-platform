package modelpolicy

import (
	"sync"
	"time"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

// usageKey represents a time-bucketed usage key for tracking per-hour and per-day usage.
type usageKey struct {
	TenantID string
	Hour     string // YYYY-MM-DD-HH
	Day      string // YYYY-MM-DD
}

type usageMeter struct {
	mu        sync.Mutex
	perTenant map[string]map[string]int
	// Time-bucketed usage tracking
	hourlyUsage map[string]int // key: "tenantID:YYYY-MM-DD-HH"
	dailyUsage  map[string]int // key: "tenantID:YYYY-MM-DD"
}

func newUsageMeter() *usageMeter {
	return &usageMeter{
		perTenant:   make(map[string]map[string]int),
		hourlyUsage: make(map[string]int),
		dailyUsage:  make(map[string]int),
	}
}

// Record tallies usage metrics per tenant; this is a minimal in-memory accumulator.
func (m *usageMeter) Record(tenantID string, usage map[string]any) {
	if tenantID == "" || usage == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := m.perTenant[tenantID]
	if dst == nil {
		dst = make(map[string]int)
		m.perTenant[tenantID] = dst
	}

	// Extract total_tokens for budget tracking
	totalTokens := 0
	for k, v := range usage {
		iv, ok := toInt(v)
		if !ok {
			continue
		}
		dst[k] += iv
		if k == "total_tokens" {
			totalTokens = iv
		}
	}

	// Track time-bucketed usage for budget enforcement
	if totalTokens > 0 {
		now := time.Now().UTC()
		hourKey := tenantID + ":" + now.Format("2006-01-02-15")
		dayKey := tenantID + ":" + now.Format("2006-01-02")
		m.hourlyUsage[hourKey] += totalTokens
		m.dailyUsage[dayKey] += totalTokens
	}
}

// CheckBudget verifies if the tenant is within their token budget.
// Returns (allowed, reason) where reason explains any denial.
func (m *usageMeter) CheckBudget(tenantID string, budget *types.TokenBudget) (bool, string) {
	if budget == nil {
		return true, ""
	}
	if budget.MaxTokensPerHour <= 0 && budget.MaxTokensPerDay <= 0 {
		return true, ""
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	hourKey := tenantID + ":" + now.Format("2006-01-02-15")
	dayKey := tenantID + ":" + now.Format("2006-01-02")

	// Check hourly budget
	if budget.MaxTokensPerHour > 0 {
		hourlyUsed := m.hourlyUsage[hourKey]
		if hourlyUsed >= budget.MaxTokensPerHour {
			return false, "hourly_token_budget_exceeded"
		}
	}

	// Check daily budget
	if budget.MaxTokensPerDay > 0 {
		dailyUsed := m.dailyUsage[dayKey]
		if dailyUsed >= budget.MaxTokensPerDay {
			return false, "daily_token_budget_exceeded"
		}
	}

	return true, ""
}

// GetUsage returns current hourly and daily usage for a tenant.
func (m *usageMeter) GetUsage(tenantID string) (hourly int, daily int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	hourKey := tenantID + ":" + now.Format("2006-01-02-15")
	dayKey := tenantID + ":" + now.Format("2006-01-02")

	return m.hourlyUsage[hourKey], m.dailyUsage[dayKey]
}

func toInt(v any) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int64:
		return int(t), true
	case float64:
		return int(t), true
	}
	return 0, false
}
