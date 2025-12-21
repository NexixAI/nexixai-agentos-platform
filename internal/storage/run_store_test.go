package storage

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

func TestFileRunStorePersistsRunsAndEnforcesTenantScope(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "runs.json")

	store, err := NewFileRunStore(path)
	if err != nil {
		t.Fatalf("NewFileRunStore error: %v", err)
	}

	run := types.Run{
		TenantID:  "tnt_alpha",
		AgentID:   "agt_demo",
		RunID:     "run_123",
		Status:    "queued",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		EventsURL: "/v1/runs/run_123/events",
	}

	if err := store.Create(ctx, run); err != nil {
		t.Fatalf("Create run error: %v", err)
	}

	if _, ok, _ := store.Get(ctx, "tnt_other", run.RunID); ok {
		t.Fatalf("expected tenant isolation on Get")
	}

	fetched, ok, err := store.Get(ctx, run.TenantID, run.RunID)
	if err != nil {
		t.Fatalf("Get run error: %v", err)
	}
	if !ok {
		t.Fatalf("run not found after create")
	}
	if fetched.AgentID != run.AgentID {
		t.Fatalf("unexpected AgentID: %s", fetched.AgentID)
	}

	fetched.Status = "completed"
	fetched.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	if err := store.Save(ctx, fetched); err != nil {
		t.Fatalf("Save run error: %v", err)
	}

	reloaded, err := NewFileRunStore(path)
	if err != nil {
		t.Fatalf("reload store error: %v", err)
	}
	persisted, ok, err := reloaded.Get(ctx, run.TenantID, run.RunID)
	if err != nil {
		t.Fatalf("Get after reload error: %v", err)
	}
	if !ok {
		t.Fatalf("run missing after reload")
	}
	if persisted.Status != "completed" || persisted.CompletedAt == "" {
		t.Fatalf("expected persisted completion, got status=%s completed_at=%s", persisted.Status, persisted.CompletedAt)
	}

	if err := store.Create(ctx, run); !errors.Is(err, ErrRunExists) {
		t.Fatalf("expected ErrRunExists on duplicate create, got %v", err)
	}
}
