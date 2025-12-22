package stackb

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/eyoshidagorgonia/nexixai-agentos-platform/internal/types"
)

type provider interface {
	Invoke(req types.ModelInvokeRequest) (map[string]any, map[string]any, error)
}

type providerEntry struct {
	model    types.Model
	impl     provider
	defaults bool
}

type registry struct {
	mu    sync.RWMutex
	model map[string]providerEntry
}

func newRegistry() *registry {
	r := &registry{
		model: make(map[string]providerEntry),
	}
	r.register(types.Model{
		ModelID:      "local-stub-llm",
		Provider:     "stub",
		DisplayName:  "Local Stub LLM",
		Capabilities: map[string]any{"chat": true},
	}, stubProvider{}, true)
	return r
}

func (r *registry) register(model types.Model, impl provider, defaults bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.model[model.ModelID] = providerEntry{model: model, impl: impl, defaults: defaults}
}

func (r *registry) Models() []types.Model {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]types.Model, 0, len(r.model))
	for _, entry := range r.model {
		out = append(out, entry.model)
	}
	return out
}

func (r *registry) Resolve(modelID string) (provider, types.Model, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if entry, ok := r.model[modelID]; ok {
		return entry.impl, entry.model, true
	}
	// fall back to default provider if requested model missing
	for _, entry := range r.model {
		if entry.defaults {
			return entry.impl, entry.model, true
		}
	}
	return nil, types.Model{}, false
}

type stubProvider struct{}

func (stubProvider) Invoke(req types.ModelInvokeRequest) (map[string]any, map[string]any, error) {
	text := "stub response"
	if inputText, ok := req.Input["text"].(string); ok && strings.TrimSpace(inputText) != "" {
		text = fmt.Sprintf("stub: %s", inputText)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	output := map[string]any{
		"type": "text",
		"text": text,
		"echo": req.Input,
		"ts":   now,
	}

	usage := map[string]any{
		"prompt_tokens":     tokenEstimate(req.Input),
		"completion_tokens": 32,
		"total_tokens":      tokenEstimate(req.Input) + 32,
		"model_id":          req.ModelID,
		"provider":          "stub",
		"timestamp":         now,
	}

	return output, usage, nil
}

func tokenEstimate(input map[string]any) int {
	if input == nil {
		return 8
	}
	if txt, ok := input["text"].(string); ok {
		n := len(strings.Fields(txt))
		if n < 4 {
			return 8
		}
		return n * 4
	}
	return 12
}
