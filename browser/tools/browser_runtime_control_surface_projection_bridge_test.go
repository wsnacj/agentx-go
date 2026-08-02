package tools

import "testing"

func TestBrowserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadataPromotesDefaultCandidateRoute(t *testing.T) {
	route := browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}
	surface, view := browserRuntimeTopLevelSurfaceAndViewFromDisplayAndReviewWithMetadata(
		&browserTopLevelDisplaySummary{
			Ready:                 true,
			Sections:              []string{"summary"},
			Category:              "coordination",
			State:                 "managed_route_pending_default",
			SummaryCode:           "managed_route_hidden_by_legacy_host_default",
			DefaultCandidateRoute: route,
		},
		&browserReviewSurfaceSummary{
			PolicyState: "popup_review_required",
			Summary: &browserTopLevelSummary{
				Category:    "review",
				State:       "manual_confirmation_required",
				SummaryCode: "popup_review_required",
			},
			Display: &browserTopLevelDisplaySummary{
				Category:    "review",
				State:       "manual_confirmation_required",
				SummaryCode: "popup_review_required",
			},
		},
		browserTopLevelCapabilitySurface{BrowserTools: []string{"browser_runtime"}},
		"",
		nil,
	)
	if surface == nil || surface.DefaultCandidateRoute != route {
		t.Fatalf("expected projected surface to preserve default candidate route, got %#v", surface)
	}
	if view == nil ||
		view.DefaultCandidateRoute != route ||
		view.Review == nil ||
		view.Review.Summary == nil ||
		view.Review.Summary.DefaultCandidateRoute != route ||
		view.Review.Display == nil ||
		view.Review.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected projected view to preserve default candidate route across nested review shells, got %#v", view)
	}
}

func TestBrowserRuntimeReviewSurfaceSummaryForTopLevelPromotesDefaultCandidateRoute(t *testing.T) {
	route := browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}
	review := browserRuntimeReviewSurfaceSummaryForTopLevel(
		"session_target_popup_review_required",
		false,
		&browserTopLevelSummary{
			Category:              "review",
			State:                 "manual_confirmation_required",
			SummaryCode:           "popup_review_required",
			DefaultCandidateRoute: route,
		},
		nil,
		nil,
		&browserTopLevelDisplaySummary{
			Category:    "review",
			State:       "manual_confirmation_required",
			SummaryCode: "popup_review_required",
		},
	)
	if review == nil ||
		review.Explanation == nil ||
		review.Explanation.DefaultCandidateRoute != route ||
		review.Display == nil ||
		review.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected review projection to preserve default candidate route across nested shells, got %#v", review)
	}
}

func TestBrowserRuntimeProjectTopLevelSurfaceBuildsReviewSurfaceAndView(t *testing.T) {
	projection := browserRuntimeProjectTopLevelSurface(browserRuntimePayload{
		Action:              "workbench",
		Status:              "ok",
		PrepareDecision:     "session_target_popup_review_required",
		PrepareReady:        false,
		Display:             &browserTopLevelDisplaySummary{Category: "coordination", State: "action_plan_available", SummaryCode: "workbench_action_plan"},
		WorkbenchReady:      true,
		WorkbenchSections:   []string{"coordination"},
		WorkbenchDisplay:    &browserRuntimeWorkbenchDisplaySummary{Ready: true, Sections: []string{"coordination"}, Category: "coordination", State: "action_plan_available", SummaryCode: "workbench_action_plan"},
		BrowserTools:        []string{"browser_runtime"},
		ArtifactTools:       []string{"browser.artifact.resolve"},
		ArtifactKinds:       []string{"download"},
		ArtifactContract:    "artifacts+media",
		BrowserActKinds:     []string{"open"},
		BrowserSurface:      "explicit_managed_opt_in",
		BrowserOptInTargets: []string{"node"},
	})
	if projection.Review == nil {
		t.Fatalf("expected review projection")
	}
	if projection.Surface == nil ||
		projection.Surface.BrowserSurface != "explicit_managed_opt_in" ||
		len(projection.Surface.BrowserTools) != 1 ||
		projection.Surface.BrowserTools[0] != "browser_runtime" {
		t.Fatalf("unexpected surface projection: %#v", projection.Surface)
	}
	if projection.View == nil ||
		projection.View.Review == nil ||
		projection.View.BrowserSurface != "explicit_managed_opt_in" ||
		len(projection.View.BrowserOptInTargets) != 1 ||
		projection.View.BrowserOptInTargets[0] != "node" {
		t.Fatalf("unexpected view projection: %#v", projection.View)
	}
}

func TestBrowserRuntimeApplyTopLevelSurfaceProjectionWritesPayloadShells(t *testing.T) {
	payload := browserRuntimePayload{}
	projection := browserRuntimeTopLevelSurfaceProjection{
		Review:  &browserReviewSurfaceSummary{PolicyState: "popup_review_required"},
		Surface: &browserTopLevelSurfaceSummary{Category: "coordination", SummaryCode: "workbench_action_plan"},
		View:    &browserTopLevelViewSummary{Kind: "review", SummaryCode: "popup_review_required"},
	}

	browserRuntimeApplyTopLevelSurfaceProjection(&payload, projection)

	if payload.Review == nil || payload.Review.PolicyState != "popup_review_required" {
		t.Fatalf("expected payload review shell to be updated, got %#v", payload.Review)
	}
	if payload.Surface == nil || payload.Surface.SummaryCode != "workbench_action_plan" {
		t.Fatalf("expected payload surface shell to be updated, got %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.SummaryCode != "popup_review_required" {
		t.Fatalf("expected payload view shell to be updated, got %#v", payload.View)
	}
}
