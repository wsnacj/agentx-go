// Package adapters defines execution adapter interfaces for AgentX docparse.
//
// The package only owns contracts and registry mechanics. Concrete business
// adapters remain in host/project code or future narrow scene adapters.
package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/wsnacj/agentx-go/scenes/docparse/planner"
	"github.com/wsnacj/agentx-go/scenes/docparse/representation"
)

// Adapter executes one planned docparse route.
type Adapter interface {
	ID() string
	Supports(planner.Route) bool
	Execute(context.Context, Input) (Output, error)
}

// Input is the normalized adapter request.
type Input struct {
	Route    planner.Route
	Document representation.Document
	Params   map[string]any
}

// Output is the source-neutral adapter result consumed by fusion.
type Output struct {
	AdapterID      string           `json:"adapter_id,omitempty"`
	RouteKind      string           `json:"route_kind,omitempty"`
	Status         string           `json:"status"`
	Payload        map[string]any   `json:"payload,omitempty"`
	Fields         []map[string]any `json:"fields,omitempty"`
	Tables         []map[string]any `json:"tables,omitempty"`
	ReviewRequired bool             `json:"review_required,omitempty"`
	Warnings       []string         `json:"warnings,omitempty"`
	Diagnostics    map[string]any   `json:"diagnostics,omitempty"`
}

// Registry stores adapters in priority order.
type Registry struct {
	adapters []Adapter
}

// NewRegistry creates an adapter registry.
func NewRegistry(adapters ...Adapter) *Registry {
	reg := &Registry{}
	for _, adapter := range adapters {
		reg.Add(adapter)
	}
	return reg
}

// Add registers an adapter.
func (r *Registry) Add(adapter Adapter) {
	if r == nil || adapter == nil || strings.TrimSpace(adapter.ID()) == "" {
		return
	}
	r.adapters = append(r.adapters, adapter)
}

// Adapters returns registered adapter IDs.
func (r *Registry) Adapters() []string {
	if r == nil {
		return nil
	}
	out := make([]string, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		if adapter == nil {
			continue
		}
		out = append(out, adapter.ID())
	}
	return out
}

// Find returns the first adapter that supports a route.
func (r *Registry) Find(route planner.Route) Adapter {
	if r == nil {
		return nil
	}
	for _, adapter := range r.adapters {
		if adapter != nil && adapter.Supports(route) {
			return adapter
		}
	}
	return nil
}

// Execute runs the first matching adapter for a route.
func (r *Registry) Execute(ctx context.Context, input Input) (Output, bool, error) {
	adapter := r.Find(input.Route)
	if adapter == nil {
		return Output{}, false, nil
	}
	out, err := adapter.Execute(ctx, input)
	if err != nil {
		return Output{}, true, fmt.Errorf("%s: %w", adapter.ID(), err)
	}
	if strings.TrimSpace(out.AdapterID) == "" {
		out.AdapterID = adapter.ID()
	}
	if strings.TrimSpace(out.RouteKind) == "" {
		out.RouteKind = input.Route.Kind
	}
	return out, true, nil
}

// Func adapts a function to Adapter for host integration and tests.
type Func struct {
	AdapterID string
	RouteKind string
	Run       func(context.Context, Input) (Output, error)
}

// ID implements Adapter.
func (f Func) ID() string { return strings.TrimSpace(f.AdapterID) }

// Supports implements Adapter.
func (f Func) Supports(route planner.Route) bool {
	if strings.TrimSpace(f.RouteKind) == "" {
		return true
	}
	return strings.TrimSpace(route.Kind) == strings.TrimSpace(f.RouteKind)
}

// Execute implements Adapter.
func (f Func) Execute(ctx context.Context, input Input) (Output, error) {
	if f.Run == nil {
		return Output{}, fmt.Errorf("adapter function is nil")
	}
	return f.Run(ctx, input)
}
