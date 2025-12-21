package quota

import (
	"os"
	"strconv"
	"sync"
	"time"
)

// Limiter is an in-memory per-tenant quota gate.
// Phase 7 is intentionally simple; later phases can swap this for Redis / distributed rate limiting.
type Limiter struct {
	mu sync.Mutex

	buckets    map[string]*bucket // per-tenant token bucket
	concurrent map[string]int     // per-tenant concurrent counter

	qps           int
	maxConcurrent int
}

type bucket struct {
	tokens float64
	last   time.Time
}

func NewFromEnv(qpsEnv, concurrentEnv string, defaultQPS, defaultConcurrent int) *Limiter {
	qps := envInt(qpsEnv, defaultQPS)
	conc := envInt(concurrentEnv, defaultConcurrent)
	return &Limiter{
		buckets:       make(map[string]*bucket),
		concurrent:    make(map[string]int),
		qps:           qps,
		maxConcurrent: conc,
	}
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil || i <= 0 {
		return def
	}
	return i
}

// AllowQPS returns true if a tenant is within its QPS budget.
func (l *Limiter) AllowQPS(tenant string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b := l.buckets[tenant]
	if b == nil {
		b = &bucket{tokens: float64(l.qps), last: time.Now()}
		l.buckets[tenant] = b
	}
	now := time.Now()
	dt := now.Sub(b.last).Seconds()
	b.last = now

	// refill
	b.tokens += dt * float64(l.qps)
	if b.tokens > float64(l.qps) {
		b.tokens = float64(l.qps)
	}

	if b.tokens < 1.0 {
		return false
	}
	b.tokens -= 1.0
	return true
}

// TryIncConcurrent increments concurrent count if under max.
func (l *Limiter) TryIncConcurrent(tenant string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cur := l.concurrent[tenant]
	if cur >= l.maxConcurrent {
		return false
	}
	l.concurrent[tenant] = cur + 1
	return true
}

// DecConcurrent decrements concurrent count (floor at 0).
func (l *Limiter) DecConcurrent(tenant string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	cur := l.concurrent[tenant]
	cur--
	if cur < 0 {
		cur = 0
	}
	l.concurrent[tenant] = cur
}
