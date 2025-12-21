package federation

import "sync"

type forwardKey struct {
	TenantID string
	RunID    string
}

type forwardTarget struct {
	RemoteStackID   string
	RemoteEventsURL string
}

type forwardIndex struct {
	mu sync.Mutex
	m  map[forwardKey]forwardTarget
}

func newForwardIndex() *forwardIndex {
	return &forwardIndex{m: make(map[forwardKey]forwardTarget)}
}

func (i *forwardIndex) Put(tenantID, runID, remoteStackID, remoteEventsURL string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.m[forwardKey{TenantID: tenantID, RunID: runID}] = forwardTarget{
		RemoteStackID:   remoteStackID,
		RemoteEventsURL: remoteEventsURL,
	}
}

func (i *forwardIndex) Get(tenantID, runID string) (forwardTarget, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	v, ok := i.m[forwardKey{TenantID: tenantID, RunID: runID}]
	return v, ok
}
