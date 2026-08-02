package planner

import (
	"testing"

	"github.com/wsnacj/agentx-go/scenes/docparse/profile"
	"github.com/wsnacj/agentx-go/scenes/docparse/representation"
)

func TestPlannerPrefersExplicitSpecPath(t *testing.T) {
	doc, err := representation.FromTextPages("doc.txt", []string{"hello world from document"}, "text")
	if err != nil {
		t.Fatal(err)
	}
	got := New().PlanRoutes(Input{
		SpecPath:             "specs/report.yaml",
		Document:             doc,
		ProfileMatch:         profile.MatchResult{Status: profile.MatchStatusUnknown},
		AllowUnknownProposal: true,
	})
	if got.Status != StatusReady || len(got.Routes) == 0 || got.Routes[0].Kind != RouteSpecDocparse {
		t.Fatalf("expected spec route first: %#v", got)
	}
	if got.ReviewRequired {
		t.Fatalf("explicit spec route should not require review by itself: %#v", got)
	}
	if hasRoute(got, RouteUnknownProposal) {
		t.Fatalf("explicit spec route should not add unknown proposal: %#v", got)
	}
}

func TestPlannerDoesNotEnableSpecializedRoutesWithoutProfileHints(t *testing.T) {
	doc, err := representation.FromTextPages("doc.txt", []string{"some extracted text"}, "text")
	if err != nil {
		t.Fatal(err)
	}
	got := New().PlanRoutes(Input{
		Document:             doc,
		AllowUnknownProposal: true,
		ProfileMatch:         profile.MatchResult{Status: profile.MatchStatusUnknown},
	})
	if len(got.Routes) != 2 || got.Routes[0].Kind != RouteGenericText || got.Routes[1].Kind != RouteUnknownProposal {
		t.Fatalf("expected only generic text and proposal routes: %#v", got)
	}
	if got.Status != StatusNeedsReview || !got.ReviewRequired {
		t.Fatalf("unknown profile should require review: %#v", got)
	}
}

func TestPlannerUsesProfileProbeRouteForClassifyOnlyTask(t *testing.T) {
	doc, err := representation.FromTextPages("doc.txt", []string{"some extracted text"}, "text")
	if err != nil {
		t.Fatal(err)
	}
	got := New().PlanRoutes(Input{
		TaskKind:             "document.profile_probe",
		Document:             doc,
		AllowUnknownProposal: true,
		ProfileMatch:         profile.MatchResult{Status: profile.MatchStatusUnknown},
	})
	if len(got.Routes) != 2 || got.Routes[0].Kind != RouteProfileProbe || got.Routes[1].Kind != RouteUnknownProposal {
		t.Fatalf("expected profile probe and proposal routes: %#v", got)
	}
	if hasRoute(got, RouteGenericText) {
		t.Fatalf("classify-only task should not plan generic extraction: %#v", got)
	}
	if got.Status != StatusNeedsReview || !got.ReviewRequired {
		t.Fatalf("profile probe should remain review-required: %#v", got)
	}
}

func TestPlannerUsesExplicitProfileRouteHints(t *testing.T) {
	doc, err := representation.FromTextPages("doc.txt", []string{"table statement text with enough characters"}, "text")
	if err != nil {
		t.Fatal(err)
	}
	got := New().PlanRoutes(Input{
		Document: doc,
		ProfileMatch: profile.MatchResult{
			Status: profile.MatchStatusVerified,
			Profile: &profile.ExtractionProfile{
				ID:         "statement-v1",
				RouteHints: []string{RouteTableStatement},
			},
		},
		HasHostProfileAdapter: true,
	})
	if !hasRoute(got, RouteHostProfile) || !hasRoute(got, RouteTableStatement) {
		t.Fatalf("expected host profile and hinted table routes: %#v", got)
	}
}

func hasRoute(plan Plan, kind string) bool {
	for _, route := range plan.Routes {
		if route.Kind == kind {
			return true
		}
	}
	return false
}
