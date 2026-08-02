package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/wsnacj/agentx-go/scenes/docparse/planner"
)

const (
	FixedFormCardAdapterID  = "fixed_form_card"
	TableStatementAdapterID = "table_statement"
	OCRXHTMLLLMAdapterID    = "ocrx_html_llm"
)

// RouteAdapter is the narrow seam for profile-hinted specialist routes.
//
// It deliberately does not register any core/card, core/table, or OCRX+LLM
// implementation by default. Hosts or future scene adapters must provide the
// Run function explicitly, so specialist document behavior stays profile-owned.
type RouteAdapter struct {
	AdapterID string
	RouteKind string
	Backend   string
	Run       func(context.Context, Input) (Output, error)
}

// NewFixedFormCardAdapter creates the explicit fixed-form card route adapter.
func NewFixedFormCardAdapter(run func(context.Context, Input) (Output, error)) RouteAdapter {
	return NewRouteAdapter(FixedFormCardAdapterID, planner.RouteFixedFormCard, "core/card", run)
}

// NewTableStatementAdapter creates the explicit table/statement route adapter.
func NewTableStatementAdapter(run func(context.Context, Input) (Output, error)) RouteAdapter {
	return NewRouteAdapter(TableStatementAdapterID, planner.RouteTableStatement, "core/table", run)
}

// NewOCRXHTMLLLMAdapter creates the explicit OCRX HTML/table + LLM route adapter.
func NewOCRXHTMLLLMAdapter(run func(context.Context, Input) (Output, error)) RouteAdapter {
	return NewRouteAdapter(OCRXHTMLLLMAdapterID, planner.RouteOCRXHTMLLLM, "core/ocrx+llm", run)
}

// NewRouteAdapter creates a specialist route adapter for host-owned wiring.
func NewRouteAdapter(adapterID string, routeKind string, backend string, run func(context.Context, Input) (Output, error)) RouteAdapter {
	return RouteAdapter{
		AdapterID: strings.TrimSpace(adapterID),
		RouteKind: strings.TrimSpace(routeKind),
		Backend:   strings.TrimSpace(backend),
		Run:       run,
	}
}

// ID implements Adapter.
func (a RouteAdapter) ID() string { return strings.TrimSpace(a.AdapterID) }

// Supports implements Adapter.
func (a RouteAdapter) Supports(route planner.Route) bool {
	routeKind := strings.TrimSpace(a.RouteKind)
	return routeKind != "" && strings.TrimSpace(route.Kind) == routeKind
}

// Execute implements Adapter.
func (a RouteAdapter) Execute(ctx context.Context, input Input) (Output, error) {
	if a.Run == nil {
		return Output{}, fmt.Errorf("route adapter %q is not configured", a.ID())
	}
	out, err := a.Run(ctx, input)
	if err != nil {
		return Output{}, err
	}
	if out.Diagnostics == nil {
		out.Diagnostics = map[string]any{}
	}
	if backend := strings.TrimSpace(a.Backend); backend != "" {
		out.Diagnostics["backend"] = backend
	}
	if profileID := strings.TrimSpace(input.Route.ProfileID); profileID != "" {
		out.Diagnostics["profile_id"] = profileID
	}
	return out, nil
}
