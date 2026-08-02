// Package planner turns profile matches and document representation metadata
// into conservative route plans for AgentX docparse.
package planner

import (
	"strings"

	"github.com/wsnacj/agentx-go/scenes/docparse/profile"
	"github.com/wsnacj/agentx-go/scenes/docparse/representation"
)

const (
	StatusReady          = "ready"
	StatusNeedsReview    = "needs_review"
	StatusNoUsableRoute  = "no_usable_route"
	RouteSpecDocparse    = "spec_docparse"
	RouteProfileProbe    = "profile_probe"
	RouteHostProfile     = "host_profile_adapter"
	RouteGenericText     = "generic_text_projection"
	RouteUnknownProposal = "unknown_profile_proposal"
	RouteFixedFormCard   = "fixed_form_card"
	RouteTableStatement  = "table_statement"
	RouteOCRXHTMLLLM     = "ocrx_html_llm"
)

// Input contains route planning signals. It intentionally avoids domain
// heuristics; profile-specific route hints must be registered explicitly.
type Input struct {
	TaskKind              string
	SpecPath              string
	Document              representation.Document
	ProfileMatch          profile.MatchResult
	HasHostProfileAdapter bool
	AllowUnknownProposal  bool
}

// Plan is an ordered list of candidate routes and the readiness boundary.
type Plan struct {
	Status         string   `json:"status"`
	ReviewRequired bool     `json:"review_required,omitempty"`
	Routes         []Route  `json:"routes,omitempty"`
	Reasons        []string `json:"reasons,omitempty"`
}

// Route describes a planned execution route. The route owner is intentionally
// explicit so generic runtime and scene/host responsibilities stay separate.
type Route struct {
	Kind      string `json:"kind"`
	Priority  int    `json:"priority"`
	Owner     string `json:"owner"`
	Enabled   bool   `json:"enabled"`
	ProfileID string `json:"profile_id,omitempty"`
	SpecPath  string `json:"spec_path,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// Planner produces deterministic plans.
type Planner struct{}

// New creates a route planner.
func New() Planner {
	return Planner{}
}

// PlanRoutes builds a conservative route plan.
func (Planner) PlanRoutes(input Input) Plan {
	routes := []Route{}
	reasons := []string{}
	specPath := strings.TrimSpace(input.SpecPath)
	profileProbe := profileProbeTask(input.TaskKind)
	if specPath == "" && input.ProfileMatch.Profile != nil {
		specPath = strings.TrimSpace(input.ProfileMatch.Profile.SpecPath)
	}
	if specPath != "" {
		routes = append(routes, Route{
			Kind:     RouteSpecDocparse,
			Priority: 10,
			Owner:    "core/docparse",
			Enabled:  true,
			SpecPath: specPath,
			Reason:   "explicit_spec_path",
		})
		reasons = append(reasons, "explicit_spec_path")
	}
	if input.HasHostProfileAdapter {
		if profileID := selectedProfileID(input.ProfileMatch); profileID != "" {
			routes = append(routes, Route{
				Kind:      RouteHostProfile,
				Priority:  20,
				Owner:     "host",
				Enabled:   true,
				ProfileID: profileID,
				Reason:    input.ProfileMatch.Status,
			})
			reasons = append(reasons, "host_profile_adapter_available")
		}
	}
	if profileProbe && specPath == "" {
		routes = append(routes, Route{
			Kind:     RouteProfileProbe,
			Priority: 30,
			Owner:    "scene/agentx_docparse/referencehost",
			Enabled:  true,
			Reason:   "profile_probe_requested",
		})
		reasons = append(reasons, "profile_probe_requested")
	}
	for _, route := range routesFromProfileHints(input.ProfileMatch) {
		routes = append(routes, route)
	}
	if representation.PagesHaveUsableText(input.Document.TextPages()) && specPath == "" && !profileProbe {
		routes = append(routes, Route{
			Kind:     RouteGenericText,
			Priority: 90,
			Owner:    "scene/agentx_docparse/referencehost",
			Enabled:  true,
			Reason:   "text_representation_available",
		})
		reasons = append(reasons, "text_representation_available")
	}
	if input.AllowUnknownProposal && input.ProfileMatch.Status == profile.MatchStatusUnknown && specPath == "" {
		routes = append(routes, Route{
			Kind:     RouteUnknownProposal,
			Priority: 100,
			Owner:    "host",
			Enabled:  true,
			Reason:   "profile_not_registered",
		})
		reasons = append(reasons, "unknown_profile_requires_review")
	}
	status := StatusReady
	if len(routes) == 0 {
		status = StatusNoUsableRoute
		reasons = append(reasons, "no_usable_route")
	}
	reviewRequired := routeKindPresent(routes, RouteUnknownProposal)
	if reviewRequired {
		status = StatusNeedsReview
	}
	return Plan{
		Status:         status,
		ReviewRequired: reviewRequired,
		Routes:         sortRoutes(routes),
		Reasons:        uniqueStrings(reasons),
	}
}

func profileProbeTask(taskKind string) bool {
	switch strings.ToLower(strings.TrimSpace(taskKind)) {
	case "profile_probe", "document.profile_probe", "document.classify", "document.classify_type", "document.detect_type":
		return true
	default:
		return false
	}
}

func routesFromProfileHints(match profile.MatchResult) []Route {
	var p *profile.ExtractionProfile
	if match.Profile != nil {
		p = match.Profile
	} else if len(match.Candidates) > 0 {
		p = &match.Candidates[0].Profile
	}
	if p == nil {
		return nil
	}
	out := []Route{}
	for _, hint := range p.RouteHints {
		switch strings.TrimSpace(hint) {
		case RouteFixedFormCard:
			out = append(out, profileHintRoute(RouteFixedFormCard, "core/card", p.ID))
		case RouteTableStatement:
			out = append(out, profileHintRoute(RouteTableStatement, "core/table", p.ID))
		case RouteOCRXHTMLLLM:
			out = append(out, profileHintRoute(RouteOCRXHTMLLLM, "scene/agentx_docparse", p.ID))
		}
	}
	return out
}

func profileHintRoute(kind string, owner string, profileID string) Route {
	return Route{
		Kind:      kind,
		Priority:  50,
		Owner:     owner,
		Enabled:   true,
		ProfileID: profileID,
		Reason:    "profile_route_hint",
	}
}

func selectedProfileID(match profile.MatchResult) string {
	if match.Profile != nil {
		return strings.TrimSpace(match.Profile.ID)
	}
	if len(match.Candidates) > 0 {
		return strings.TrimSpace(match.Candidates[0].Profile.ID)
	}
	return ""
}

func routeKindPresent(routes []Route, kind string) bool {
	for _, route := range routes {
		if route.Kind == kind {
			return true
		}
	}
	return false
}

func sortRoutes(routes []Route) []Route {
	out := append([]Route(nil), routes...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Priority < out[i].Priority || (out[j].Priority == out[i].Priority && out[j].Kind < out[i].Kind) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
