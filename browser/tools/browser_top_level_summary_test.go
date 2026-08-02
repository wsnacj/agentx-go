package tools

import "testing"

func TestBrowserTopLevelDisplayFromViewPromotesDefaultCandidateRouteFromReview(t *testing.T) {
	route := browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Source:        "managed_browserd",
		Endpoint:      "http://127.0.0.1:43123",
	}
	display := browserTopLevelDisplayFromView(&browserTopLevelViewSummary{
		Kind:        "review",
		Category:    "review",
		State:       "manual_confirmation_required",
		SummaryCode: "popup_review_required",
		Review: &browserReviewSurfaceSummary{
			Summary: &browserTopLevelSummary{
				Category:              "review",
				State:                 "manual_confirmation_required",
				SummaryCode:           "popup_review_required",
				DefaultCandidateRoute: route,
			},
		},
	})
	if display == nil || display.DefaultCandidateRoute != route {
		t.Fatalf("expected display to promote default candidate route from nested review, got %#v", display)
	}
}

func TestBrowserTopLevelSummaryFromViewPromotesDefaultCandidateRouteFromReview(t *testing.T) {
	route := browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Source:        "managed_browserd",
		Endpoint:      "http://127.0.0.1:43123",
	}
	summary := browserTopLevelSummaryFromView(&browserTopLevelViewSummary{
		Kind:        "review",
		Category:    "review",
		State:       "manual_confirmation_required",
		SummaryCode: "popup_review_required",
		Review: &browserReviewSurfaceSummary{
			Display: &browserTopLevelDisplaySummary{
				Category:              "review",
				State:                 "manual_confirmation_required",
				SummaryCode:           "popup_review_required",
				DefaultCandidateRoute: route,
			},
		},
	})
	if summary == nil || summary.DefaultCandidateRoute != route {
		t.Fatalf("expected summary to promote default candidate route from nested review, got %#v", summary)
	}
}

func TestBrowserTopLevelSummaryFromReviewPromotesDefaultCandidateRouteFromNestedDisplay(t *testing.T) {
	route := browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Source:        "managed_browserd",
		Endpoint:      "http://127.0.0.1:43123",
	}
	summary := browserTopLevelSummaryFromReview(&browserReviewSurfaceSummary{
		PolicyState: "popup_review_required",
		Summary: &browserTopLevelSummary{
			Category:    "review",
			State:       "manual_confirmation_required",
			SummaryCode: "popup_review_required",
		},
		Display: &browserTopLevelDisplaySummary{
			Category:              "review",
			State:                 "manual_confirmation_required",
			SummaryCode:           "popup_review_required",
			DefaultCandidateRoute: route,
		},
	})
	if summary == nil || summary.DefaultCandidateRoute != route {
		t.Fatalf("expected review summary to promote default candidate route from nested display, got %#v", summary)
	}
}

func TestBrowserTopLevelSummaryFromWorkbenchPromotesDefaultCandidateRouteFromNestedReview(t *testing.T) {
	route := browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Source:        "managed_browserd",
		Endpoint:      "http://127.0.0.1:43123",
	}
	summary := browserTopLevelSummaryFromWorkbench(&browserRuntimeWorkbenchSurfaceSummary{
		Ready: true,
		Sections: []string{
			"route",
		},
		Summary: &browserTopLevelSummary{
			Category:    "coordination",
			State:       "managed_route_pending_default",
			SummaryCode: "managed_route_hidden_by_legacy_host_default",
		},
		Review: &browserReviewSurfaceSummary{
			Display: &browserTopLevelDisplaySummary{
				Category:              "review",
				State:                 "manual_confirmation_required",
				SummaryCode:           "popup_review_required",
				DefaultCandidateRoute: route,
			},
		},
	})
	if summary == nil || summary.DefaultCandidateRoute != route {
		t.Fatalf("expected workbench summary to promote default candidate route from nested review, got %#v", summary)
	}
}

func TestBrowserReviewSurfaceSummaryFromViewPromotesDefaultCandidateRouteToNestedShells(t *testing.T) {
	route := browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Source:        "managed_browserd",
		Endpoint:      "http://127.0.0.1:43123",
	}
	review := browserReviewSurfaceSummaryFromView(&browserTopLevelViewSummary{
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
	})
	if review == nil ||
		review.Summary == nil ||
		review.Summary.DefaultCandidateRoute != route ||
		review.Display == nil ||
		review.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected review alias from view to overlay default candidate route onto nested shells, got %#v", review)
	}
}

func TestBrowserReviewSurfaceSummaryFromWorkbenchPromotesDefaultCandidateRouteToNestedShells(t *testing.T) {
	route := browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Source:        "managed_browserd",
		Endpoint:      "http://127.0.0.1:43123",
	}
	review := browserReviewSurfaceSummaryFromWorkbench(&browserRuntimeWorkbenchSurfaceSummary{
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
	})
	if review == nil ||
		review.Summary == nil ||
		review.Summary.DefaultCandidateRoute != route ||
		review.Display == nil ||
		review.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected review alias from workbench to overlay default candidate route onto nested shells, got %#v", review)
	}
}
