package audit

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	Time          string         `json:"time"`
	TenantID      string         `json:"tenant_id"`
	PrincipalID   string         `json:"principal_id,omitempty"`
	Action        string         `json:"action"`
	Resource      string         `json:"resource"`
	Outcome       string         `json:"outcome"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	RequestID     string         `json:"request_id,omitempty"`
	Meta          map[string]any `json:"meta,omitempty"`
}

type Logger interface {
	Log(e Entry)
}

type jsonLineLogger struct {
	mu sync.Mutex
	w  io.Writer
}

func (l *jsonLineLogger) Log(e Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e.Time == "" {
		e.Time = time.Now().UTC().Format(time.RFC3339)
	}
	b, _ := json.Marshal(e)
	_, _ = l.w.Write(append(b, '\n'))
}

// NewFromEnv creates an audit logger.
//
// AGENTOS_AUDIT_SINK:
//   - "stdout"
//   - "stderr"
//   - "file:/path/to/audit.log" (default; falls back to data/audit/{service}.audit.log)
func NewFromEnv() Logger {
	sink := strings.TrimSpace(os.Getenv("AGENTOS_AUDIT_SINK"))
	if sink == "" {
		svc := strings.TrimSpace(os.Getenv("AGENTOS_SERVICE"))
		name := "audit.log"
		if svc != "" {
			name = svc + ".audit.log"
		}
		sink = "file:" + filepath.Join("data", "audit", name)
	}
	if sink == "" || sink == "stdout" {
		return &jsonLineLogger{w: os.Stdout}
	}
	if sink == "stderr" {
		return &jsonLineLogger{w: os.Stderr}
	}
	if strings.HasPrefix(sink, "file:") {
		p := strings.TrimPrefix(sink, "file:")
		_ = os.MkdirAll(dir(p), 0o755)
		f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return &jsonLineLogger{w: os.Stdout}
		}
		return &jsonLineLogger{w: f}
	}
	return &jsonLineLogger{w: os.Stdout}
}

func dir(p string) string {
	if p == "" {
		return "."
	}
	return filepath.Dir(p)
}
