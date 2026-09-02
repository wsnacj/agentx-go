// Package tools provides a provider-neutral tool catalog and portable tools.
package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
)

// Handler executes one registered tool.
type Handler = toolcontract.Handler

// Executor executes tool calls.
type Executor = toolcontract.Executor

// Registry is a concurrency-safe catalog and executor.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]entry
	version uint64
}

type entry struct {
	definition toolcontract.Definition
	handler    Handler
}

// NewRegistry constructs an empty Registry.
func NewRegistry() *Registry { return &Registry{entries: make(map[string]entry)} }

// Register associates a function name with its handler. Empty or nil values are ignored.
func (r *Registry) Register(definition toolcontract.Definition, handler Handler) {
	if r == nil || definition.Function.Name == "" || handler == nil {
		return
	}
	r.mu.Lock()
	r.entries[definition.Function.Name] = entry{definition: definition, handler: handler}
	r.version++
	r.mu.Unlock()
}

// Execute invokes a registered handler, including deterministic name repair.
func (r *Registry) Execute(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
	if r == nil {
		return "", fmt.Errorf("llmx: tool registry not initialised")
	}
	r.mu.RLock()
	item, ok := r.entries[call.Name]
	repair := nameRepairMatch{}
	if !ok {
		repair = resolveNameRepair(r.entries, call.Name)
		if repair.ok {
			item = r.entries[repair.name]
			call.Name = repair.name
			ok = true
		}
	}
	r.mu.RUnlock()
	if !ok {
		return "", NewToolNameError(call.Name, ToolNameRepairResolution{
			Requested: strings.TrimSpace(call.Name), NormalizedKey: strings.TrimSpace(canonicalNameKey(call.Name)),
			Candidates: repair.candidates, Ambiguous: repair.ambiguous,
			Reason: firstNonEmpty(repair.reason, "not_registered"),
		})
	}
	return item.handler(ctx, call)
}

// ValidateCall validates one registered definition without executing its
// handler. It applies the same deterministic name repair as Execute and is
// safe to call before authorization or sandbox admission.
func (r *Registry) ValidateCall(call toolcontract.Call, bindings BindingContext) (ArgumentValidation, error) {
	if r == nil {
		return ArgumentValidation{}, fmt.Errorf("llmx: tool registry not initialised")
	}
	r.mu.RLock()
	item, ok := r.entries[call.Name]
	repair := nameRepairMatch{}
	if !ok {
		repair = resolveNameRepair(r.entries, call.Name)
		if repair.ok {
			item = r.entries[repair.name]
			call.Name = repair.name
			ok = true
		}
	}
	r.mu.RUnlock()
	if !ok {
		return ArgumentValidation{}, NewToolNameError(call.Name, ToolNameRepairResolution{
			Requested: strings.TrimSpace(call.Name), NormalizedKey: strings.TrimSpace(canonicalNameKey(call.Name)),
			Candidates: repair.candidates, Ambiguous: repair.ambiguous,
			Reason: firstNonEmpty(repair.reason, "not_registered"),
		})
	}
	return ValidateCallArguments(item.definition, call.Arguments, bindings)
}

// Ensure returns exec or a new empty Registry when exec is nil.
func Ensure(exec Executor) Executor {
	if exec != nil {
		return exec
	}
	return NewRegistry()
}

// Reset removes all entries and advances the mutation version.
func (r *Registry) Reset() {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.entries = make(map[string]entry)
	r.version++
	r.mu.Unlock()
}

// Version reports the catalog mutation version.
func (r *Registry) Version() uint64 {
	if r == nil {
		return 0
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.version
}

// Definitions returns a stable, name-sorted snapshot.
func (r *Registry) Definitions() []toolcontract.Definition {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.entries) == 0 {
		return nil
	}
	names := make([]string, 0, len(r.entries))
	for name := range r.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	definitions := make([]toolcontract.Definition, 0, len(names))
	for _, name := range names {
		definitions = append(definitions, r.entries[name].definition)
	}
	return definitions
}
