package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

var (
	// ErrRunExists signals attempts to create a run that already exists for the tenant.
	ErrRunExists = errors.New("run already exists")
	// ErrInvalidRun signals missing required run identity fields.
	ErrInvalidRun = errors.New("invalid run")
)

// RunStore is a tenant-scoped persistence port for run state.
type RunStore interface {
	Create(ctx context.Context, run types.Run) error
	Get(ctx context.Context, tenantID, runID string) (types.Run, bool, error)
	Save(ctx context.Context, run types.Run) error
}

// NewRunStoreFromEnv constructs the default run store adapter, using AGENTOS_RUN_STORE_FILE
// or falling back to ./data/agent-orchestrator/runs.json.
func NewRunStoreFromEnv() (RunStore, error) {
	path := strings.TrimSpace(os.Getenv("AGENTOS_RUN_STORE_FILE"))
	if path == "" {
		path = filepath.Join("data", "agent-orchestrator", "runs.json")
	}
	return NewFileRunStore(path)
}
