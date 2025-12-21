package metrics

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	registry = prometheus.NewRegistry()

	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agentos_http_requests_total",
			Help: "Total HTTP requests served by AgentOS services.",
		},
		[]string{"service", "method", "code"},
	)

	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "agentos_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "code"},
	)

	quotaDenied = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agentos_quota_denied_total",
			Help: "Total quota denials.",
		},
		[]string{"service", "kind"},
	)

	fedForwardFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "agentos_federation_forward_failures_total",
			Help: "Federation forward failures (remote create run failures).",
		},
		[]string{"service", "reason"},
	)
)

func init() {
	_ = registry.Register(collectors.NewGoCollector())
	_ = registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	_ = registry.Register(httpRequests)
	_ = registry.Register(httpDuration)
	_ = registry.Register(quotaDenied)
	_ = registry.Register(fedForwardFailures)
}

// Handler returns a Prometheus scrape handler for the AgentOS registry.
func Handler() http.Handler {
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}

type statusCapturingWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusCapturingWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusCapturingWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Instrument wraps an HTTP handler and records basic request counters and latency.
func Instrument(service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusCapturingWriter{ResponseWriter: w, status: 200}
		start := time.Now()
		next.ServeHTTP(sw, r)
		code := strconv.Itoa(sw.status)
		httpRequests.WithLabelValues(service, r.Method, code).Inc()
		httpDuration.WithLabelValues(service, r.Method, code).Observe(time.Since(start).Seconds())
	})
}

func IncQuotaDenied(service, kind string) {
	quotaDenied.WithLabelValues(service, kind).Inc()
}

func IncFederationForwardFailure(service, reason string) {
	fedForwardFailures.WithLabelValues(service, reason).Inc()
}

// MetricsRequireAuth returns whether metrics endpoints should require auth.
// Default is false for local/dev; set AGENTOS_METRICS_REQUIRE_AUTH=1 to enable.
func MetricsRequireAuth() bool {
	v := os.Getenv("AGENTOS_METRICS_REQUIRE_AUTH")
	if v == "" {
		return false
	}
	i, _ := strconv.Atoi(v)
	return i == 1
}
