package browserruntime

import "testing"

func TestBuildSharedSessionBrowserWorkbenchSurfaceBuildsNestedReview(t *testing.T) {
	surface := BuildSharedSessionBrowserWorkbenchSurface(
		SharedSessionBrowserWorkbenchSurfaceRequest{
			Ready:                true,
			Sections:             []string{"route", "coordination"},
			Summary:              &SharedSessionBrowserSummary{Category: "review", State: "manual_confirmation_required", SummaryCode: "popup_review_required", NextStepAlias: "tabs", ManualRetryHint: "rerun_with_force"},
			PrimaryBrowserAction: "browser action=tabs",
			PrimaryNodeAction:    "nodes action=run_status",
			NextStep:             "browser action=tabs",
			ReviewDecision:       "session_target_popup_review_required",
		},
	)

	if surface == nil || surface.Review == nil {
		t.Fatalf("expected workbench surface to include nested review, got %#v", surface)
	}
	if surface.Review.PolicyState != "popup_review_required" || surface.Review.Decision != "session_target_popup_review_required" || surface.Review.Ready {
		t.Fatalf("expected workbench review to mirror decision posture, got %#v", surface.Review)
	}
	if surface.PrimaryBrowserAction != "browser action=tabs" || surface.NextStep != "browser action=tabs" {
		t.Fatalf("expected workbench surface to preserve action plan, got %#v", surface)
	}
}

func TestBuildSharedSessionBrowserWorkbenchSurfaceCarriesRouteSurface(t *testing.T) {
	surface := BuildSharedSessionBrowserWorkbenchSurface(
		SharedSessionBrowserWorkbenchSurfaceRequest{
			Ready:               true,
			BrowserSurface:      "explicit_managed_opt_in",
			BrowserOptInTargets: []string{"node", "node"},
		},
	)
	if surface == nil {
		t.Fatalf("expected route-surface-backed workbench surface")
	}
	if surface.BrowserSurface != "explicit_managed_opt_in" {
		t.Fatalf("expected workbench browser surface label, got %#v", surface)
	}
	if len(surface.BrowserOptInTargets) != 1 || surface.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected deduplicated workbench opt-in targets, got %#v", surface.BrowserOptInTargets)
	}
}

func TestBuildSharedSessionBrowserWorkbenchSurfaceCarriesCapabilitySurface(t *testing.T) {
	surface := BuildSharedSessionBrowserWorkbenchSurface(
		SharedSessionBrowserWorkbenchSurfaceRequest{
			Ready:            true,
			BrowserTools:     []string{"browser_runtime", "browser_act", "browser_act"},
			ArtifactTools:    []string{"browser_act"},
			ArtifactKinds:    []string{"screenshot", "screenshot"},
			ArtifactContract: "artifacts+media",
			BrowserActKinds:  []string{"click", "click"},
		},
	)
	if surface == nil {
		t.Fatalf("expected capability-surface-backed workbench surface")
	}
	if len(surface.BrowserTools) != 2 || surface.BrowserTools[0] != "browser_runtime" || surface.BrowserTools[1] != "browser_act" {
		t.Fatalf("expected deduplicated workbench browser tools, got %#v", surface.BrowserTools)
	}
	if len(surface.ArtifactTools) != 1 || surface.ArtifactTools[0] != "browser_act" {
		t.Fatalf("expected workbench artifact tools, got %#v", surface.ArtifactTools)
	}
	if len(surface.ArtifactKinds) != 1 || surface.ArtifactKinds[0] != "screenshot" || surface.ArtifactContract != "artifacts+media" {
		t.Fatalf("expected workbench artifact surface, got %#v", surface)
	}
	if len(surface.BrowserActKinds) != 1 || surface.BrowserActKinds[0] != "click" {
		t.Fatalf("expected workbench browser_act kinds, got %#v", surface.BrowserActKinds)
	}
}

func TestBuildSharedSessionBrowserSurfacePrefersReviewDisplayFallback(t *testing.T) {
	review := BuildSharedSessionBrowserReviewSurface(
		SharedSessionBrowserReviewSurfaceRequest{
			ReviewDecision: "session_target_popup_review_required",
			Explanation: &SharedSessionBrowserSummary{
				Category:        "review",
				State:           "manual_confirmation_required",
				SummaryCode:     "popup_review_required",
				NextStepAlias:   "tabs",
				ManualRetryHint: "rerun_with_force",
			},
		},
	)
	surface := BuildSharedSessionBrowserSurface(nil, review)

	if surface == nil {
		t.Fatalf("expected review-backed surface")
	}
	if surface.ReviewPolicyState != "popup_review_required" || surface.ReviewDecision != "session_target_popup_review_required" || surface.Category != "review" || surface.SummaryCode != "popup_review_required" {
		t.Fatalf("expected surface to mirror review display fallback, got %#v", surface)
	}
}

func TestBuildSharedSessionBrowserViewPrefersWorkbenchAndReviewDisplay(t *testing.T) {
	workbench := BuildSharedSessionBrowserWorkbenchSurface(
		SharedSessionBrowserWorkbenchSurfaceRequest{
			Ready:       true,
			Sections:    []string{"route", "coordination"},
			Diagnostics: &SharedSessionBrowserSummary{Category: "coordination", State: "action_plan_available", SummaryCode: "workbench_action_plan", PrimaryBrowserAction: "browser action=refresh", NextStep: "browser action=refresh"},
		},
	)
	workbenchDisplay := &SharedSessionBrowserDisplay{
		Ready:    true,
		Sections: []string{"route", "coordination"},
		SharedSessionBrowserSummary: SharedSessionBrowserSummary{
			Category:             "review",
			State:                "manual_confirmation_required",
			SummaryCode:          "popup_review_required",
			NextStepAlias:        "tabs",
			ManualRetryHint:      "rerun_with_force",
			PrimaryBrowserAction: "browser action=tabs",
		},
	}
	view := BuildSharedSessionBrowserView(workbench, workbenchDisplay, nil, workbench.Review)

	if view == nil || view.Kind != "workbench" {
		t.Fatalf("expected workbench view, got %#v", view)
	}
	if view.Category != "coordination" || view.SummaryCode != "workbench_action_plan" {
		t.Fatalf("expected workbench summary to remain primary, got %#v", view)
	}
	if view.Review != nil {
		t.Fatalf("expected workbench view without explicit review surface not to synthesize one, got %#v", view)
	}
}

func TestBuildSharedSessionBrowserDisplayFromRequestPrefersWorkbenchDisplayOverView(t *testing.T) {
	display := BuildSharedSessionBrowserDisplayFromRequest(
		SharedSessionBrowserDisplayRequest{
			WorkbenchDisplay: &SharedSessionBrowserDisplay{
				Ready:    true,
				Sections: []string{"route", "coordination"},
				SharedSessionBrowserSummary: SharedSessionBrowserSummary{
					Category:    "coordination",
					State:       "action_plan_available",
					SummaryCode: "workbench_action_plan",
				},
			},
			View: &SharedSessionBrowserView{
				Kind:     "review",
				Ready:    true,
				Sections: []string{"review"},
				SharedSessionBrowserSummary: SharedSessionBrowserSummary{
					Category:    "review",
					State:       "manual_confirmation_required",
					SummaryCode: "popup_review_required",
				},
			},
		},
	)
	if display == nil {
		t.Fatalf("expected shared display")
	}
	if display.Category != "coordination" || display.SummaryCode != "workbench_action_plan" {
		t.Fatalf("expected workbench_display to remain primary, got %#v", display)
	}
	if !display.Ready || len(display.Sections) != 2 {
		t.Fatalf("expected workbench_display shell to remain intact, got %#v", display)
	}
}

func TestBuildSharedSessionBrowserDisplayFromRequestFallsBackToReviewThenSummary(t *testing.T) {
	display := BuildSharedSessionBrowserDisplayFromRequest(
		SharedSessionBrowserDisplayRequest{
			Review: &SharedSessionBrowserReviewSurface{
				PolicyState: "popup_review_required",
				Summary: &SharedSessionBrowserSummary{
					Category:        "review",
					State:           "manual_confirmation_required",
					SummaryCode:     "popup_review_required",
					NextStepAlias:   "tabs",
					ManualRetryHint: "rerun_with_force",
				},
			},
			Summary: &SharedSessionBrowserSummary{
				Category:    "coordination",
				State:       "action_plan_available",
				SummaryCode: "workbench_action_plan",
			},
		},
	)
	if display == nil {
		t.Fatalf("expected shared display fallback")
	}
	if display.Category != "review" || display.SummaryCode != "popup_review_required" || display.NextStepAlias != "tabs" {
		t.Fatalf("expected review fallback to outrank plain summary, got %#v", display)
	}
}

func TestBuildSharedSessionBrowserReviewAliasFromRequestPrefersExplicitReview(t *testing.T) {
	review := BuildSharedSessionBrowserReviewAliasFromRequest(
		SharedSessionBrowserReviewAliasRequest{
			Review: &SharedSessionBrowserReviewSurface{
				PolicyState: "redirect_review_required",
				Decision:    "session_target_redirect_review_required",
				Summary: &SharedSessionBrowserSummary{
					Category:    "review",
					State:       "manual_confirmation_required",
					SummaryCode: "redirect_review_required",
				},
			},
			View: &SharedSessionBrowserView{
				Review: &SharedSessionBrowserReviewSurface{
					PolicyState: "popup_review_required",
					Decision:    "session_target_popup_review_required",
				},
			},
			Workbench: &SharedSessionBrowserWorkbenchSurface{
				Review: &SharedSessionBrowserReviewSurface{
					PolicyState: "download_review_required",
					Decision:    "download_review_required",
				},
			},
		},
	)
	if review == nil {
		t.Fatalf("expected shared review alias")
	}
	if review.PolicyState != "redirect_review_required" || review.Decision != "session_target_redirect_review_required" || review.Summary == nil || review.Summary.SummaryCode != "redirect_review_required" {
		t.Fatalf("expected explicit review to remain primary, got %#v", review)
	}
}

func TestBuildSharedSessionBrowserSummaryAliasFromRequestPrefersExplicitSummary(t *testing.T) {
	summary := BuildSharedSessionBrowserSummaryAliasFromRequest(
		SharedSessionBrowserSummaryAliasRequest{
			Summary: &SharedSessionBrowserSummary{
				Category:    "coordination",
				State:       "action_plan_available",
				SummaryCode: "workbench_action_plan",
			},
			Review: &SharedSessionBrowserReviewSurface{
				Summary: &SharedSessionBrowserSummary{
					Category:    "review",
					State:       "manual_confirmation_required",
					SummaryCode: "popup_review_required",
				},
			},
			Display: &SharedSessionBrowserDisplay{
				SharedSessionBrowserSummary: SharedSessionBrowserSummary{
					Category:    "resolver",
					State:       "manual_resolution_required",
					SummaryCode: "label_filtered_residual",
				},
			},
		},
	)
	if summary == nil {
		t.Fatalf("expected shared summary alias")
	}
	if summary.Category != "coordination" || summary.SummaryCode != "workbench_action_plan" {
		t.Fatalf("expected explicit summary to remain primary, got %#v", summary)
	}
}

func TestBuildSharedSessionBrowserSummaryAliasFromRequestNormalizesResolverFallback(t *testing.T) {
	summary := BuildSharedSessionBrowserSummaryAliasFromRequest(
		SharedSessionBrowserSummaryAliasRequest{
			Summary: &SharedSessionBrowserSummary{
				State:           "resolved_via_fallback",
				SummaryCode:     "label_filtered_residual",
				ManualRetryHint: "add_ordinal",
			},
		},
	)
	if summary == nil {
		t.Fatalf("expected shared summary alias")
	}
	if !summary.ResolvedViaFallback || summary.Category != "resolver_fallback" {
		t.Fatalf("expected resolver fallback summary normalization, got %#v", summary)
	}
}

func TestBuildSharedSessionBrowserSummaryAliasFromRequestNormalizesResolverCategory(t *testing.T) {
	summary := BuildSharedSessionBrowserSummaryAliasFromRequest(
		SharedSessionBrowserSummaryAliasRequest{
			Summary: &SharedSessionBrowserSummary{
				State:         "manual_resolution_required",
				SummaryCode:   "label_filtered_residual",
				NextStepAlias: "snapshot",
			},
		},
	)
	if summary == nil {
		t.Fatalf("expected shared summary alias")
	}
	if summary.Category != "resolver" || summary.ResolvedViaFallback {
		t.Fatalf("expected resolver summary normalization, got %#v", summary)
	}
}

func TestBuildSharedSessionBrowserTopLevelAliasProjectionThreadsDerivedChain(t *testing.T) {
	projection := BuildSharedSessionBrowserTopLevelAliasProjection(
		SharedSessionBrowserTopLevelAliasProjectionRequest{
			Review: SharedSessionBrowserReviewAliasRequest{
				View: &SharedSessionBrowserView{
					Review: &SharedSessionBrowserReviewSurface{
						PolicyState: "popup_review_required",
						Decision:    "session_target_popup_review_required",
						Summary: &SharedSessionBrowserSummary{
							Category:      "review",
							State:         "manual_confirmation_required",
							SummaryCode:   "popup_review_required",
							NextStepAlias: "tabs",
						},
					},
				},
			},
			Surface: SharedSessionBrowserSurfaceAliasRequest{
				BrowserSurface:      "explicit_managed_opt_in",
				BrowserOptInTargets: []string{"node"},
			},
			View: SharedSessionBrowserViewAliasRequest{
				BrowserSurface:      "explicit_managed_opt_in",
				BrowserOptInTargets: []string{"node"},
			},
		},
	)

	if projection.Review == nil || projection.Review.Summary == nil || projection.Review.Summary.SummaryCode != "popup_review_required" {
		t.Fatalf("expected review projection to derive from nested view review, got %#v", projection.Review)
	}
	if projection.Summary == nil || projection.Summary.SummaryCode != "popup_review_required" {
		t.Fatalf("expected summary projection to derive from review, got %#v", projection.Summary)
	}
	if projection.Display == nil || projection.Display.SummaryCode != "popup_review_required" {
		t.Fatalf("expected display projection to derive from summary/review, got %#v", projection.Display)
	}
	if projection.Surface == nil || projection.Surface.BrowserSurface != "explicit_managed_opt_in" || len(projection.Surface.BrowserOptInTargets) != 1 || projection.Surface.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected surface projection to retain route surface, got %#v", projection.Surface)
	}
	if projection.View == nil || projection.View.BrowserSurface != "explicit_managed_opt_in" || len(projection.View.BrowserOptInTargets) != 1 || projection.View.BrowserOptInTargets[0] != "node" || projection.View.SummaryCode != "popup_review_required" {
		t.Fatalf("expected view projection to derive from threaded top-level chain, got %#v", projection.View)
	}
}

func TestBuildSharedSessionBrowserRouteAliasFromRequestPrefersExplicitRootRoute(t *testing.T) {
	route := BuildSharedSessionBrowserRouteAliasFromRequest(
		SharedSessionBrowserRouteAliasRequest{
			BrowserSurface:      "explicit_managed_opt_in",
			BrowserOptInTargets: []string{"sandbox"},
			Surface: &SharedSessionBrowserSurface{
				BrowserSurface:      "explicit_managed_opt_in",
				BrowserOptInTargets: []string{"node"},
			},
			View: &SharedSessionBrowserView{
				BrowserSurface:      "explicit_managed_opt_in",
				BrowserOptInTargets: []string{"host"},
			},
			Workbench: &SharedSessionBrowserWorkbenchSurface{
				BrowserSurface:      "explicit_managed_opt_in",
				BrowserOptInTargets: []string{"node"},
			},
		},
	)
	if route == nil {
		t.Fatalf("expected shared route alias")
	}
	if route.BrowserSurface != "explicit_managed_opt_in" || len(route.BrowserOptInTargets) != 1 || route.BrowserOptInTargets[0] != "sandbox" {
		t.Fatalf("expected explicit root route to remain primary, got %#v", route)
	}
}

func TestBuildSharedSessionBrowserRouteAliasFromRequestFallsBackToWorkbench(t *testing.T) {
	route := BuildSharedSessionBrowserRouteAliasFromRequest(
		SharedSessionBrowserRouteAliasRequest{
			Workbench: &SharedSessionBrowserWorkbenchSurface{
				BrowserSurface:      "explicit_managed_opt_in",
				BrowserOptInTargets: []string{"node", "node"},
			},
		},
	)
	if route == nil {
		t.Fatalf("expected shared route alias")
	}
	if route.BrowserSurface != "explicit_managed_opt_in" || len(route.BrowserOptInTargets) != 1 || route.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected workbench route fallback, got %#v", route)
	}
}

func TestBuildSharedSessionBrowserCapabilityAliasFromRequestMergesExplicitAndNestedFields(t *testing.T) {
	capability := BuildSharedSessionBrowserCapabilityAliasFromRequest(
		SharedSessionBrowserCapabilityAliasRequest{
			BrowserTools: []string{"browser_runtime"},
			Surface: &SharedSessionBrowserSurface{
				ArtifactTools: []string{"browser_act"},
			},
			View: &SharedSessionBrowserView{
				ArtifactKinds: []string{"screenshot"},
			},
			Workbench: &SharedSessionBrowserWorkbenchSurface{
				ArtifactContract: "artifacts+media",
				BrowserActKinds:  []string{"click"},
			},
		},
	)
	if capability == nil {
		t.Fatalf("expected shared capability alias")
	}
	if len(capability.BrowserTools) != 1 || capability.BrowserTools[0] != "browser_runtime" {
		t.Fatalf("expected explicit browser tools to remain primary, got %#v", capability)
	}
	if len(capability.ArtifactTools) != 1 || capability.ArtifactTools[0] != "browser_act" {
		t.Fatalf("expected surface artifact tools fallback, got %#v", capability)
	}
	if len(capability.ArtifactKinds) != 1 || capability.ArtifactKinds[0] != "screenshot" {
		t.Fatalf("expected view artifact kinds fallback, got %#v", capability)
	}
	if capability.ArtifactContract != "artifacts+media" {
		t.Fatalf("expected workbench artifact contract fallback, got %#v", capability)
	}
	if len(capability.BrowserActKinds) != 1 || capability.BrowserActKinds[0] != "click" {
		t.Fatalf("expected workbench browser_act kinds fallback, got %#v", capability)
	}
}

func TestBuildSharedSessionBrowserCapabilityAliasFromRequestFallsBackToWorkbench(t *testing.T) {
	capability := BuildSharedSessionBrowserCapabilityAliasFromRequest(
		SharedSessionBrowserCapabilityAliasRequest{
			Workbench: &SharedSessionBrowserWorkbenchSurface{
				BrowserTools:    []string{"browser_runtime", "browser_act"},
				ArtifactTools:   []string{"browser_act"},
				BrowserActKinds: []string{"click"},
			},
		},
	)
	if capability == nil {
		t.Fatalf("expected shared capability alias")
	}
	if len(capability.BrowserTools) != 2 || capability.BrowserTools[0] != "browser_runtime" || capability.BrowserTools[1] != "browser_act" {
		t.Fatalf("expected workbench capability fallback, got %#v", capability)
	}
	if len(capability.ArtifactTools) != 1 || capability.ArtifactTools[0] != "browser_act" || len(capability.BrowserActKinds) != 1 || capability.BrowserActKinds[0] != "click" {
		t.Fatalf("expected workbench capability metadata, got %#v", capability)
	}
}

func TestBuildSharedSessionBrowserSurfaceAliasFromRequestUsesViewRouteWhenRebuildingSurface(t *testing.T) {
	surface := BuildSharedSessionBrowserSurfaceAliasFromRequest(
		SharedSessionBrowserSurfaceAliasRequest{
			Display: &SharedSessionBrowserDisplay{
				SharedSessionBrowserSummary: SharedSessionBrowserSummary{
					Category:    "coordination",
					State:       "action_plan_available",
					SummaryCode: "workbench_action_plan",
				},
			},
			View: &SharedSessionBrowserView{
				Kind:                "workbench",
				BrowserSurface:      "explicit_managed_opt_in",
				BrowserOptInTargets: []string{"node"},
			},
		},
	)
	if surface == nil {
		t.Fatalf("expected shared surface alias")
	}
	if surface.BrowserSurface != "explicit_managed_opt_in" || len(surface.BrowserOptInTargets) != 1 || surface.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected rebuilt surface to inherit route surface from view, got %#v", surface)
	}
	if surface.Category != "coordination" || surface.SummaryCode != "workbench_action_plan" {
		t.Fatalf("expected rebuilt surface to preserve display summary, got %#v", surface)
	}
}

func TestBuildSharedSessionBrowserSurfaceAliasFromRequestUsesViewCapabilityWhenRebuildingSurface(t *testing.T) {
	surface := BuildSharedSessionBrowserSurfaceAliasFromRequest(
		SharedSessionBrowserSurfaceAliasRequest{
			Display: &SharedSessionBrowserDisplay{
				SharedSessionBrowserSummary: SharedSessionBrowserSummary{
					Category:    "coordination",
					State:       "action_plan_available",
					SummaryCode: "workbench_action_plan",
				},
			},
			View: &SharedSessionBrowserView{
				Kind:            "workbench",
				BrowserTools:    []string{"browser_runtime", "browser_act"},
				ArtifactTools:   []string{"browser_act"},
				BrowserActKinds: []string{"click"},
			},
		},
	)
	if surface == nil {
		t.Fatalf("expected shared surface alias")
	}
	if len(surface.BrowserTools) != 2 || surface.BrowserTools[0] != "browser_runtime" || surface.BrowserTools[1] != "browser_act" {
		t.Fatalf("expected rebuilt surface to inherit capability surface from view, got %#v", surface)
	}
	if len(surface.ArtifactTools) != 1 || surface.ArtifactTools[0] != "browser_act" || len(surface.BrowserActKinds) != 1 || surface.BrowserActKinds[0] != "click" {
		t.Fatalf("expected rebuilt surface to preserve view capability metadata, got %#v", surface)
	}
}

func TestBuildSharedSessionBrowserViewAliasFromRequestPrefersSurfaceRouteOverExplicitViewRoute(t *testing.T) {
	view := BuildSharedSessionBrowserViewAliasFromRequest(
		SharedSessionBrowserViewAliasRequest{
			View: &SharedSessionBrowserView{
				Kind:                "review",
				BrowserSurface:      "explicit_managed_opt_in",
				BrowserOptInTargets: []string{"sandbox"},
				SharedSessionBrowserSummary: SharedSessionBrowserSummary{
					Category:    "review",
					State:       "manual_confirmation_required",
					SummaryCode: "popup_review_required",
				},
			},
			Surface: &SharedSessionBrowserSurface{
				BrowserSurface:      "explicit_managed_opt_in",
				BrowserOptInTargets: []string{"node"},
			},
		},
	)
	if view == nil {
		t.Fatalf("expected shared view alias")
	}
	if view.BrowserSurface != "explicit_managed_opt_in" || len(view.BrowserOptInTargets) != 1 || view.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected explicit view alias to adopt surface route precedence, got %#v", view)
	}
	if view.Kind != "review" || view.SummaryCode != "popup_review_required" {
		t.Fatalf("expected explicit view shell to remain primary, got %#v", view)
	}
}

func TestBuildSharedSessionBrowserViewAliasFromRequestPrefersSurfaceCapabilityOverExplicitViewCapability(t *testing.T) {
	view := BuildSharedSessionBrowserViewAliasFromRequest(
		SharedSessionBrowserViewAliasRequest{
			View: &SharedSessionBrowserView{
				Kind:            "review",
				BrowserTools:    []string{"browser_runtime"},
				ArtifactTools:   []string{"browser_screenshot"},
				BrowserActKinds: []string{"snapshot"},
				SharedSessionBrowserSummary: SharedSessionBrowserSummary{
					Category:    "review",
					State:       "manual_confirmation_required",
					SummaryCode: "popup_review_required",
				},
			},
			Surface: &SharedSessionBrowserSurface{
				BrowserTools:    []string{"browser_runtime", "browser_act"},
				ArtifactTools:   []string{"browser_act"},
				BrowserActKinds: []string{"click"},
			},
		},
	)
	if view == nil {
		t.Fatalf("expected shared view alias")
	}
	if len(view.BrowserTools) != 2 || view.BrowserTools[0] != "browser_runtime" || view.BrowserTools[1] != "browser_act" {
		t.Fatalf("expected explicit view alias to adopt surface capability precedence, got %#v", view)
	}
	if len(view.ArtifactTools) != 1 || view.ArtifactTools[0] != "browser_act" || len(view.BrowserActKinds) != 1 || view.BrowserActKinds[0] != "click" {
		t.Fatalf("expected explicit view alias to keep primary capability surface, got %#v", view)
	}
}

func TestBuildSharedSessionBrowserSurfaceFromRequestCarriesRouteSurface(t *testing.T) {
	surface := BuildSharedSessionBrowserSurfaceFromRequest(
		SharedSessionBrowserSurfaceRequest{
			BrowserSurface:      "explicit_managed_opt_in",
			BrowserOptInTargets: []string{"node", "node"},
		},
	)
	if surface == nil {
		t.Fatalf("expected route-surface-backed shared surface")
	}
	if surface.BrowserSurface != "explicit_managed_opt_in" {
		t.Fatalf("expected browser surface label, got %#v", surface)
	}
	if len(surface.BrowserOptInTargets) != 1 || surface.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected deduplicated opt-in targets, got %#v", surface.BrowserOptInTargets)
	}
}

func TestBuildSharedSessionBrowserSurfaceFromRequestCarriesCapabilitySurface(t *testing.T) {
	surface := BuildSharedSessionBrowserSurfaceFromRequest(
		SharedSessionBrowserSurfaceRequest{
			Display: &SharedSessionBrowserDisplay{
				SharedSessionBrowserSummary: SharedSessionBrowserSummary{
					Category:    "coordination",
					State:       "action_plan_available",
					SummaryCode: "workbench_action_plan",
				},
			},
			BrowserTools:     []string{"browser_runtime", "browser_act"},
			ArtifactTools:    []string{"browser_act"},
			ArtifactKinds:    []string{"screenshot"},
			ArtifactContract: "artifacts+media",
			BrowserActKinds:  []string{"click"},
		},
	)
	if surface == nil {
		t.Fatalf("expected capability-surface-backed shared surface")
	}
	if len(surface.BrowserTools) != 2 || surface.BrowserTools[0] != "browser_runtime" || surface.BrowserTools[1] != "browser_act" {
		t.Fatalf("expected shared surface browser tools, got %#v", surface.BrowserTools)
	}
	if len(surface.ArtifactTools) != 1 || surface.ArtifactTools[0] != "browser_act" {
		t.Fatalf("expected shared surface artifact tools, got %#v", surface.ArtifactTools)
	}
	if len(surface.ArtifactKinds) != 1 || surface.ArtifactKinds[0] != "screenshot" || surface.ArtifactContract != "artifacts+media" {
		t.Fatalf("expected shared surface artifact surface, got %#v", surface)
	}
	if len(surface.BrowserActKinds) != 1 || surface.BrowserActKinds[0] != "click" {
		t.Fatalf("expected shared surface browser_act kinds, got %#v", surface.BrowserActKinds)
	}
}

func TestBuildSharedSessionBrowserViewFromRequestInheritsRouteSurfaceFromSharedSurface(t *testing.T) {
	surface := BuildSharedSessionBrowserSurfaceFromRequest(
		SharedSessionBrowserSurfaceRequest{
			Display: &SharedSessionBrowserDisplay{
				SharedSessionBrowserSummary: SharedSessionBrowserSummary{
					Category:    "coordination",
					State:       "action_plan_available",
					SummaryCode: "workbench_action_plan",
				},
			},
			BrowserSurface:      "explicit_managed_opt_in",
			BrowserOptInTargets: []string{"node"},
		},
	)
	view := BuildSharedSessionBrowserViewFromRequest(
		SharedSessionBrowserViewRequest{
			Surface: surface,
		},
	)
	if view == nil {
		t.Fatalf("expected result view")
	}
	if view.BrowserSurface != "explicit_managed_opt_in" {
		t.Fatalf("expected view to inherit browser surface, got %#v", view)
	}
	if len(view.BrowserOptInTargets) != 1 || view.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected view to inherit opt-in targets, got %#v", view.BrowserOptInTargets)
	}
}

func TestBuildSharedSessionBrowserViewFromRequestInheritsRouteSurfaceFromWorkbenchSurface(t *testing.T) {
	workbench := BuildSharedSessionBrowserWorkbenchSurface(
		SharedSessionBrowserWorkbenchSurfaceRequest{
			Ready:               true,
			Sections:            []string{"coordination"},
			BrowserSurface:      "explicit_managed_opt_in",
			BrowserOptInTargets: []string{"node"},
			Diagnostics: &SharedSessionBrowserSummary{
				Category:    "coordination",
				State:       "action_plan_available",
				SummaryCode: "workbench_action_plan",
			},
		},
	)
	view := BuildSharedSessionBrowserViewFromRequest(
		SharedSessionBrowserViewRequest{
			Workbench: workbench,
		},
	)
	if view == nil {
		t.Fatalf("expected workbench view")
	}
	if view.BrowserSurface != "explicit_managed_opt_in" {
		t.Fatalf("expected workbench view to inherit browser surface, got %#v", view)
	}
	if len(view.BrowserOptInTargets) != 1 || view.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected workbench view to inherit opt-in targets, got %#v", view.BrowserOptInTargets)
	}
}

func TestBuildSharedSessionBrowserViewFromRequestSupportsRouteSurfaceWithoutOtherAliases(t *testing.T) {
	view := BuildSharedSessionBrowserViewFromRequest(
		SharedSessionBrowserViewRequest{
			BrowserSurface:      "explicit_managed_opt_in",
			BrowserOptInTargets: []string{"node"},
		},
	)
	if view == nil {
		t.Fatalf("expected route-surface-backed view")
	}
	if view.Kind != "" || view.BrowserSurface != "explicit_managed_opt_in" {
		t.Fatalf("unexpected route-only view payload: %#v", view)
	}
	if len(view.BrowserOptInTargets) != 1 || view.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected route-only view to preserve targets, got %#v", view.BrowserOptInTargets)
	}
}
