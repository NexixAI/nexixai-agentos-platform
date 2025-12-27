package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

type fileRunStore struct {
	mu   sync.Mutex
	path string
	runs map[string]types.Run
}

// NewFileRunStore returns a file-backed RunStore persisted as JSON.
func NewFileRunStore(path string) (RunStore, error) {
	if path == "" {
		path = filepath.Join("data", "agent-orchestrator", "runs.json")
	}
	s := &fileRunStore{
		path: path,
		runs: make(map[string]types.Run),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *fileRunStore) Create(ctx context.Context, run types.Run) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if err := validateRun(run); err != nil {
		return err
	}
	key := storageKey(run.TenantID, run.RunID)

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.runs[key]; exists {
		return ErrRunExists
	}
	s.runs[key] = run
	return s.persistLocked()
}

func (s *fileRunStore) Get(ctx context.Context, tenantID, runID string) (types.Run, bool, error) {
	if err := ctxErr(ctx); err != nil {
		return types.Run{}, false, err
	}
	if tenantID == "" || runID == "" {
		return types.Run{}, false, nil
	}
	key := storageKey(tenantID, runID)

	s.mu.Lock()
	defer s.mu.Unlock()
	run, ok := s.runs[key]
	return run, ok, nil
}

func (s *fileRunStore) Save(ctx context.Context, run types.Run) error {
	if err := ctxErr(ctx); err != nil {
		return err
	}
	if err := validateRun(run); err != nil {
		return err
	}
	key := storageKey(run.TenantID, run.RunID)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs[key] = run
	return s.persistLocked()
}

func (s *fileRunStore) load() error {
	if s.path == "" {
		return nil
	}
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(b) == 0 {
		return nil
	}
	var persisted map[string]types.Run
	if err := json.Unmarshal(b, &persisted); err != nil {
		return err
	}
	for k, v := range persisted {
		s.runs[k] = v
	}
	return nil
}

func (s *fileRunStore) persistLocked() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.runs, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func validateRun(run types.Run) error {
	if run.TenantID == "" || run.RunID == "" {
		return ErrInvalidRun
	}
	return nil
}

func storageKey(tenantID, runID string) string {
	return filepath.ToSlash(filepath.Join("tenant", tenantID, "runs", runID))
}

func ctxErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}
