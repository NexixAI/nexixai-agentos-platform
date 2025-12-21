package federation

import (
	"sync"
)

type storeKey struct {
	TenantID string
	RunID    string
}

type storedEvents struct {
	seenEventIDs map[string]struct{}
	lastSequence int
	events       []map[string]any // raw event envelopes
}

type eventStore struct {
	mu sync.Mutex
	m  map[storeKey]*storedEvents
}

func newEventStore() *eventStore {
	return &eventStore{m: make(map[storeKey]*storedEvents)}
}

func (s *eventStore) Ingest(tenantID, runID string, envelopes []map[string]any) (accepted int, rejected int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := storeKey{TenantID: tenantID, RunID: runID}
	st := s.m[k]
	if st == nil {
		st = &storedEvents{seenEventIDs: make(map[string]struct{})}
		s.m[k] = st
	}

	for _, env := range envelopes {
		eventObj, ok := env["event"].(map[string]any)
		if !ok {
			rejected++
			continue
		}
		eid, _ := eventObj["event_id"].(string)
		if eid == "" {
			rejected++
			continue
		}

		seq := 0
		if v, ok := eventObj["sequence"].(float64); ok {
			seq = int(v)
		}
		if seq != 0 && seq <= st.lastSequence {
			// enforce monotonic ordering
			rejected++
			continue
		}
		if _, exists := st.seenEventIDs[eid]; exists {
			rejected++
			continue
		}

		st.seenEventIDs[eid] = struct{}{}
		if seq > st.lastSequence {
			st.lastSequence = seq
		}
		st.events = append(st.events, env)
		accepted++
	}

	return accepted, rejected
}

func (s *eventStore) List(tenantID, runID string) ([]map[string]any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.m[storeKey{TenantID: tenantID, RunID: runID}]
	if st == nil {
		return nil, false
	}
	// copy
	out := make([]map[string]any, 0, len(st.events))
	out = append(out, st.events...)
	return out, true
}
