package pack

import (
	"errors"
	"fmt"
	"sync"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type Registry interface {
	Register(def Definition) error
	Get(id string) (Definition, bool)
	List() []Definition
	ResolveWorkflow(packID string, caseType string) (agentxworkflow.Spec, bool)
	ResolveMaterializedWorkflow(packID string, caseType string) (agentxworkflow.Spec, bool, error)
}

type MemoryRegistry struct {
	coordinator *Coordinator
	mu          sync.RWMutex
	defs        map[string]Definition
}

// NewMemoryRegistry构造一个使用指定 Coordinator进行校验与物化的内存注册表。
func NewMemoryRegistry(coordinator *Coordinator) (*MemoryRegistry, error) {
	if _, err := coordinator.workflowValidator(); err != nil {
		return nil, err
	}
	if _, err := coordinator.toolArgumentLowerer(); err != nil {
		return nil, err
	}
	return &MemoryRegistry{coordinator: coordinator, defs: map[string]Definition{}}, nil
}

func (r *MemoryRegistry) Register(def Definition) error {
	if r == nil || r.coordinator == nil {
		return errors.New("agentx pack: registry is required")
	}
	if err := r.coordinator.ValidateDefinition(def); err != nil {
		return err
	}
	id := def.Manifest.ID
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.defs[id]; exists {
		return fmt.Errorf("pack: definition %q already registered", id)
	}
	r.defs[id] = cloneDefinition(def)
	return nil
}

func (r *MemoryRegistry) Get(id string) (Definition, bool) {
	if r == nil {
		return Definition{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	def, ok := r.defs[id]
	if !ok {
		return Definition{}, false
	}
	return cloneDefinition(def), true
}

func (r *MemoryRegistry) List() []Definition {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Definition, 0, len(r.defs))
	for _, def := range r.defs {
		out = append(out, cloneDefinition(def))
	}
	sortDefinitions(out)
	return out
}

func (r *MemoryRegistry) ResolveWorkflow(packID string, caseType string) (agentxworkflow.Spec, bool) {
	def, ok := r.Get(packID)
	if !ok {
		return agentxworkflow.Spec{}, false
	}
	if !def.Manifest.SupportsCaseType(caseType) {
		return agentxworkflow.Spec{}, false
	}
	spec, err := def.ResolveWorkflowForCaseType(caseType, "")
	if err != nil {
		return agentxworkflow.Spec{}, false
	}
	return spec, true
}
