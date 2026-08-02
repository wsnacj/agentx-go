// Package understanding orchestrates document representation, profile matching,
// route planning, adapter execution, and fusion for AgentX docparse.
package understanding

import (
	"context"
	"fmt"
	"strings"

	"github.com/wsnacj/agentx-go/scenes/docparse/adapters"
	"github.com/wsnacj/agentx-go/scenes/docparse/fusion"
	"github.com/wsnacj/agentx-go/scenes/docparse/planner"
	"github.com/wsnacj/agentx-go/scenes/docparse/profile"
	"github.com/wsnacj/agentx-go/scenes/docparse/representation"
)

// Engine is the document understanding orchestration surface. It is safe to
// use as plan-only when no adapters are registered.
type Engine struct {
	matcher  *profile.Matcher
	planner  planner.Planner
	adapters *adapters.Registry
}

// Options configures Engine.
type Options struct {
	Profiles *profile.Registry
	Adapters *adapters.Registry
	Planner  planner.Planner
}

// New creates an Engine.
func New(opts Options) *Engine {
	pl := opts.Planner
	if pl == (planner.Planner{}) {
		pl = planner.New()
	}
	return &Engine{
		matcher:  profile.NewMatcher(opts.Profiles),
		planner:  pl,
		adapters: opts.Adapters,
	}
}

// Input is a single document understanding request.
type Input struct {
	TaskKind              string
	Params                map[string]any
	Document              representation.Document
	HasHostProfileAdapter bool
	AllowUnknownProposal  bool
}

// Output contains plan, optional fused result, and diagnostics.
type Output struct {
	Profile         profile.MatchResult `json:"profile"`
	Plan            planner.Plan        `json:"plan"`
	AdapterOutputs  []adapters.Output   `json:"adapter_outputs,omitempty"`
	UnhandledRoutes []planner.Route     `json:"unhandled_routes,omitempty"`
	Fusion          *fusion.Result      `json:"fusion,omitempty"`
}

// PlanOnly matches and plans without executing adapters.
func (e *Engine) PlanOnly(input Input) Output {
	match := e.matchProfile(input)
	plan := e.planner.PlanRoutes(planner.Input{
		TaskKind:              strings.TrimSpace(input.TaskKind),
		SpecPath:              firstString(input.Params, "spec_path"),
		Document:              input.Document,
		ProfileMatch:          match,
		HasHostProfileAdapter: input.HasHostProfileAdapter,
		AllowUnknownProposal:  input.AllowUnknownProposal,
	})
	return Output{Profile: match, Plan: plan}
}

// Run executes registered adapters for the planned routes, if any.
func (e *Engine) Run(ctx context.Context, input Input) (Output, error) {
	out := e.PlanOnly(input)
	if e.adapters == nil || len(out.Plan.Routes) == 0 {
		return out, nil
	}
	adapterOutputs := []adapters.Output{}
	unhandledRoutes := []planner.Route{}
	for _, route := range out.Plan.Routes {
		if !route.Enabled {
			continue
		}
		adapterOut, handled, err := e.adapters.Execute(ctx, adapters.Input{
			Route:    route,
			Document: input.Document,
			Params:   cloneMap(input.Params),
		})
		if err != nil {
			return out, err
		}
		if handled {
			adapterOutputs = append(adapterOutputs, adapterOut)
		} else {
			unhandledRoutes = append(unhandledRoutes, route)
		}
	}
	if len(adapterOutputs) > 0 {
		out.AdapterOutputs = adapterOutputs
		merged := fusion.Merge(adapterOutputs)
		out.Fusion = &merged
	}
	if len(unhandledRoutes) > 0 {
		out.UnhandledRoutes = unhandledRoutes
	}
	return out, nil
}

func (e *Engine) matchProfile(input Input) profile.MatchResult {
	return e.matcher.Match(profile.MatchInput{
		ExplicitProfileID:    firstString(input.Params, "profile_id", "parser_profile_id"),
		ExplicitDocumentType: firstString(input.Params, "document_type", "expected_document_type", "actual_document_type", "detected_document_type"),
		SpecPath:             firstString(input.Params, "spec_path"),
		RequestedFields:      requestedFields(input.Params),
	})
}

func requestedFields(params map[string]any) []string {
	for _, key := range []string{"requested_fields", "required_fields"} {
		if values := stringSlice(params[key]); len(values) > 0 {
			return values
		}
	}
	return nil
}

func firstString(params map[string]any, keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(toString(params[key]))
		if value != "" {
			return value
		}
	}
	return ""
}

func stringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(toString(item))
			if text != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func toString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "<nil>" {
		return ""
	}
	return text
}

func cloneMap(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
