package modelpolicy

import "sync"

type usageMeter struct {
	mu        sync.Mutex
	perTenant map[string]map[string]int
}

func newUsageMeter() *usageMeter {
	return &usageMeter{perTenant: make(map[string]map[string]int)}
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
	for k, v := range usage {
		iv, ok := toInt(v)
		if !ok {
			continue
		}
		dst[k] += iv
	}
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
