package federation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type forwardKey struct {
	TenantID string
	RunID    string
}

type forwardTarget struct {
	RemoteStackID   string
	RemoteEventsURL string
}

type forwardIndex struct {
	mu   sync.Mutex
	m    map[forwardKey]forwardTarget
	path string // if non-empty, persist on Set
}

type forwardIndexRecord struct {
	TenantID       string `json:"tenant_id"`
	RunID          string `json:"run_id"`
	RemoteStackID  string `json:"remote_stack_id"`
	RemoteEventsURL string `json:"remote_events_url"`
}

func newForwardIndex() *forwardIndex {
	return &forwardIndex{m: make(map[forwardKey]forwardTarget)}
}

// newForwardIndexPersistent creates a forward index that persists to a JSON file on Set().
// This is a minimal "production-ish" improvement to avoid losing forward mappings on restarts.
func newForwardIndexPersistent(path string) *forwardIndex {
	i := &forwardIndex{m: make(map[forwardKey]forwardTarget), path: path}
	_ = i.load()
	return i
}

func (i *forwardIndex) load() error {
	if i.path == "" {
		return nil
	}
	b, err := os.ReadFile(i.path)
	if err != nil {
		return nil // file doesn't exist is fine
	}
	var recs []forwardIndexRecord
	if err := json.Unmarshal(b, &recs); err != nil {
		return err
	}
	for _, r := range recs {
		i.m[forwardKey{TenantID: r.TenantID, RunID: r.RunID}] = forwardTarget{
			RemoteStackID:   r.RemoteStackID,
			RemoteEventsURL: r.RemoteEventsURL,
		}
	}
	return nil
}

func (i *forwardIndex) persistLocked() error {
	if i.path == "" {
		return nil
	}
	dir := filepath.Dir(i.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	recs := make([]forwardIndexRecord, 0, len(i.m))
	for k, v := range i.m {
		recs = append(recs, forwardIndexRecord{
			TenantID: k.TenantID, RunID: k.RunID, RemoteStackID: v.RemoteStackID, RemoteEventsURL: v.RemoteEventsURL,
		})
	}
	b, err := json.MarshalIndent(recs, "", "  ")
	if err != nil {
		return err
	}
	tmp := i.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, i.path)
}

func (i *forwardIndex) Set(tenantID, runID, remoteStackID, remoteEventsURL string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.m[forwardKey{TenantID: tenantID, RunID: runID}] = forwardTarget{
		RemoteStackID:   remoteStackID,
		RemoteEventsURL: remoteEventsURL,
	}
	_ = i.persistLocked()
}

func (i *forwardIndex) Get(tenantID, runID string) (forwardTarget, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	v, ok := i.m[forwardKey{TenantID: tenantID, RunID: runID}]
	return v, ok
}
