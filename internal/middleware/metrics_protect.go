\
package middleware

import (
	"net/http"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/auth"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/httpx"
	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/metrics"
)

// ProtectMetrics enforces tenant auth on /metrics if AGENTOS_METRICS_REQUIRE_AUTH=1.
func ProtectMetrics(next http.Handler) http.Handler {
	if !metrics.MetricsRequireAuth() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac, _ := auth.Get(r.Context())
		if _, ok := auth.RequireTenant(ac); !ok {
			httpx.Error(w, http.StatusUnauthorized, "unauthorized", "tenant_id required", httpx.CorrelationID(r), false)
			return
		}
		next.ServeHTTP(w, r)
	})
}
