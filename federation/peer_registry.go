package federation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PeerInfo struct {
	StackID     string    `json:"stack_id"`
	Environment string    `json:"environment"`
	Region      string    `json:"region"`
	APIVersions []string  `json:"api_versions"`
	Endpoints   Endpoints `json:"endpoints"`
	Build       Build     `json:"build"`
}

type Endpoints struct {
	StackABaseURL string `json:"stack_a_base_url"`
	StackBBaseURL string `json:"stack_b_base_url"`
}

type Build struct {
	Version   string `json:"version"`
	GitSHA    string `json:"git_sha"`
	Timestamp string `json:"timestamp"`
}

type PeersFile struct {
	Local PeerInfo   `json:"local"`
	Peers []PeerInfo `json:"peers"`
}

// Registry provides a minimal peer lookup by stack_id.
type Registry struct {
	Local PeerInfo
	peers map[string]PeerInfo
}

func LoadRegistryFromEnv() (*Registry, error) {
	path := strings.TrimSpace(os.Getenv("AGENTOS_PEERS_FILE"))
	if path == "" {
		return nil, fmt.Errorf("AGENTOS_PEERS_FILE not set")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var pf PeersFile
	if err := json.Unmarshal(b, &pf); err != nil {
		return nil, err
	}

	// Allow overriding local identity via env (useful for compose multi-node).
	stackID := strings.TrimSpace(os.Getenv("AGENTOS_STACK_ID"))
	env := strings.TrimSpace(os.Getenv("AGENTOS_ENVIRONMENT"))
	region := strings.TrimSpace(os.Getenv("AGENTOS_REGION"))
	if stackID != "" {
		pf.Local.StackID = stackID
	}
	if env != "" {
		pf.Local.Environment = env
	}
	if region != "" {
		pf.Local.Region = region
	}

	peers := make(map[string]PeerInfo, len(pf.Peers))
	for _, p := range pf.Peers {
		if p.StackID == "" {
			continue
		}
		peers[p.StackID] = p
	}

	return &Registry{Local: pf.Local, peers: peers}, nil
}

func (r *Registry) Get(stackID string) (PeerInfo, bool) {
	p, ok := r.peers[stackID]
	return p, ok
}
