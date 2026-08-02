package tools

import "testing"

func TestBrowserTopLevelViewWithCapabilitySurfaceSummaryPromotesDefaultCandidateRouteToNestedReview(t *testing.T) {
	route := browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}
	view := browserTopLevelViewWithCapabilitySurfaceSummary(
		&browserTopLevelViewSummary{
			Kind:                  "review",
			DefaultCandidateRoute: route,
			Review: &browserReviewSurfaceSummary{
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
		},
		browserTopLevelCapabilitySurface{BrowserTools: []string{"browser_runtime"}},
	)
	if view == nil ||
		!browserStringSliceContains(view.BrowserTools, "browser_runtime") ||
		view.Review == nil ||
		view.Review.Summary == nil ||
		view.Review.Summary.DefaultCandidateRoute != route ||
		view.Review.Display == nil ||
		view.Review.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected capability projection to preserve default candidate route across view review shells, got %#v", view)
	}
}

func TestBrowserWorkbenchWithCapabilitySurfaceSummaryPromotesDefaultCandidateRouteToNestedReview(t *testing.T) {
	route := browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}
	workbench := browserWorkbenchWithCapabilitySurfaceSummary(
		&browserRuntimeWorkbenchSurfaceSummary{
			Ready:                 true,
			DefaultCandidateRoute: route,
			Review: &browserReviewSurfaceSummary{
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
		},
		browserTopLevelCapabilitySurface{BrowserTools: []string{"browser_runtime"}},
	)
	if workbench == nil ||
		!browserStringSliceContains(workbench.BrowserTools, "browser_runtime") ||
		workbench.Review == nil ||
		workbench.Review.Summary == nil ||
		workbench.Review.Summary.DefaultCandidateRoute != route ||
		workbench.Review.Display == nil ||
		workbench.Review.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected capability projection to preserve default candidate route across workbench review shells, got %#v", workbench)
	}
}
