package tools

import (
	"encoding/json"
	"testing"
)

func TestMarshalBrowserClickPayloadCarriesRouteSurfaceIntoTopLevelAliases(t *testing.T) {
	out, err := marshalBrowserClickPayload(browserClickToolPayload{
		Backend:             "proxy",
		BrowserApp:          "Chromium",
		Profile:             "workbench",
		RuntimeTarget:       "node",
		BrowserSurface:      "explicit_managed_opt_in",
		BrowserOptInTargets: []string{"node"},
		Selector:            "#submit",
		Status:              "clicked",
	})
	if err != nil {
		t.Fatalf("marshalBrowserClickPayload: %v", err)
	}
	var payload struct {
		BrowserSurface      string                         `json:"browser_surface"`
		BrowserOptInTargets []string                       `json:"browser_opt_in_targets"`
		Surface             *browserTopLevelSurfaceSummary `json:"surface"`
		View                *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode click payload: %v", err)
	}
	if payload.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.BrowserOptInTargets) != 1 ||
		payload.BrowserOptInTargets[0] != "node" {
		t.Fatalf("unexpected root route surface payload: %#v", payload)
	}
	if payload.Surface == nil ||
		payload.Surface.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.Surface.BrowserOptInTargets) != 1 ||
		payload.Surface.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected click surface alias to carry route surface, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.View.BrowserOptInTargets) != 1 ||
		payload.View.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected click view alias to carry route surface, got %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasOverlaysRouteSurfaceOnExistingAliases(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["node"],"surface":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"},"view":{"kind":"workbench","category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Surface *browserTopLevelSurfaceSummary `json:"surface"`
		View    *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if payload.Surface == nil ||
		payload.Surface.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.Surface.BrowserOptInTargets) != 1 ||
		payload.Surface.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected unified surface alias to be overlaid with route surface, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.View.BrowserOptInTargets) != 1 ||
		payload.View.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected unified view alias to be overlaid with route surface, got %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasBuildsSurfaceFromViewRouteSurface(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","display":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual"},"view":{"kind":"workbench","category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["node"]}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		BrowserSurface      string                         `json:"browser_surface"`
		BrowserOptInTargets []string                       `json:"browser_opt_in_targets"`
		Surface             *browserTopLevelSurfaceSummary `json:"surface"`
		View                *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if payload.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.BrowserOptInTargets) != 1 ||
		payload.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected root payload to recover route surface from view, got %#v", payload)
	}
	if payload.Surface == nil ||
		payload.Surface.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.Surface.BrowserOptInTargets) != 1 ||
		payload.Surface.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected surface alias to recover route surface from view, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.View.BrowserOptInTargets) != 1 ||
		payload.View.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected view alias to keep route surface, got %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasOverlaysDefaultCandidateRouteOnExistingAliases(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","default_candidate_route":{"backend":"proxy","profile":"isolated","runtime_target":"node"},"summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"},"display":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"},"surface":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"},"view":{"kind":"workbench","category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		DefaultCandidate browserRuntimeRouteDescriptor  `json:"default_candidate_route"`
		Summary          *browserTopLevelSummary        `json:"summary"`
		Display          *browserTopLevelDisplaySummary `json:"display"`
		Surface          *browserTopLevelSurfaceSummary `json:"surface"`
		View             *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if payload.DefaultCandidate != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) {
		t.Fatalf("expected root payload to keep default candidate route, got %#v", payload.DefaultCandidate)
	}
	if payload.Summary == nil || payload.Summary.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected summary alias to inherit default candidate route, got %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected display alias to inherit default candidate route, got %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected surface alias to inherit default candidate route, got %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected view alias to inherit default candidate route, got %#v", payload.View)
	}
}

func TestBrowserRuntimeApplyDefaultCandidateRouteToPayloadShellsOverlaysNestedWorkbenchAndReviewSummaries(t *testing.T) {
	route := browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}
	payload := &browserRuntimePayload{
		DefaultCandidateRoute: route,
		Explanation: &browserTopLevelSummary{
			SummaryCode: "top_level_explanation",
		},
		Diagnostics: &browserTopLevelSummary{
			SummaryCode: "top_level_diagnostics",
		},
		Review: &browserReviewSurfaceSummary{
			PolicyState: "popup_review_required",
			Summary: &browserTopLevelSummary{
				SummaryCode: "popup_review_required",
			},
			Display: &browserTopLevelDisplaySummary{
				SummaryCode: "popup_review_required",
			},
		},
		View: &browserTopLevelViewSummary{
			Kind: "review",
			Review: &browserReviewSurfaceSummary{
				PolicyState: "popup_review_required",
				Summary: &browserTopLevelSummary{
					SummaryCode: "popup_review_required",
				},
				Display: &browserTopLevelDisplaySummary{
					SummaryCode: "popup_review_required",
				},
			},
		},
		Workbench: &browserRuntimeWorkbenchSurfaceSummary{
			Ready: true,
			Sections: []string{
				"route",
			},
			Explanation: &browserTopLevelSummary{
				SummaryCode: "workbench_explanation",
			},
			Diagnostics: &browserTopLevelSummary{
				SummaryCode: "workbench_diagnostics",
			},
			Summary: &browserTopLevelSummary{
				SummaryCode: "workbench_summary",
			},
			Review: &browserReviewSurfaceSummary{
				PolicyState: "popup_review_required",
				Summary: &browserTopLevelSummary{
					SummaryCode: "popup_review_required",
				},
				Display: &browserTopLevelDisplaySummary{
					SummaryCode: "popup_review_required",
				},
			},
		},
	}

	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)

	if payload.Explanation == nil ||
		payload.Explanation.DefaultCandidateRoute != route ||
		payload.Diagnostics == nil ||
		payload.Diagnostics.DefaultCandidateRoute != route ||
		payload.Review == nil ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.DefaultCandidateRoute != route ||
		payload.Review.Display == nil ||
		payload.Review.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected top-level review and summary shells to inherit default candidate route, got payload=%#v", payload)
	}
	if payload.View == nil ||
		payload.View.DefaultCandidateRoute != route ||
		payload.View.Review == nil ||
		payload.View.Review.Summary == nil ||
		payload.View.Review.Summary.DefaultCandidateRoute != route ||
		payload.View.Review.Display == nil ||
		payload.View.Review.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected view review shells to inherit default candidate route, got %#v", payload.View)
	}
	if payload.Workbench == nil ||
		payload.Workbench.DefaultCandidateRoute != route ||
		payload.Workbench.Explanation == nil ||
		payload.Workbench.Explanation.DefaultCandidateRoute != route ||
		payload.Workbench.Diagnostics == nil ||
		payload.Workbench.Diagnostics.DefaultCandidateRoute != route ||
		payload.Workbench.Summary == nil ||
		payload.Workbench.Summary.DefaultCandidateRoute != route ||
		payload.Workbench.Review == nil ||
		payload.Workbench.Review.Summary == nil ||
		payload.Workbench.Review.Summary.DefaultCandidateRoute != route ||
		payload.Workbench.Review.Display == nil ||
		payload.Workbench.Review.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected workbench nested shells to inherit default candidate route, got %#v", payload.Workbench)
	}
}

func TestBrowserRuntimeApplyDefaultCandidateRouteToPayloadShellsPromotesRootDefaultCandidateRouteFromNestedReview(t *testing.T) {
	route := browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}
	payload := &browserRuntimePayload{
		Explanation: &browserTopLevelSummary{
			SummaryCode: "top_level_explanation",
		},
		Display: &browserTopLevelDisplaySummary{
			SummaryCode: "top_level_display",
		},
		Review: &browserReviewSurfaceSummary{
			PolicyState: "popup_review_required",
			Summary: &browserTopLevelSummary{
				SummaryCode:           "popup_review_required",
				DefaultCandidateRoute: route,
			},
		},
	}

	browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload)

	if payload.DefaultCandidateRoute != route {
		t.Fatalf("expected payload root to promote default candidate route from nested review, got %#v", payload.DefaultCandidateRoute)
	}
	if payload.Explanation == nil || payload.Explanation.DefaultCandidateRoute != route {
		t.Fatalf("expected explanation shell to inherit promoted default candidate route, got %#v", payload.Explanation)
	}
	if payload.Display == nil || payload.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected display shell to inherit promoted default candidate route, got %#v", payload.Display)
	}
}

func TestBrowserUnifiedApplyExplanationAliasOverlaysDefaultCandidateRouteOnExistingExplanationReviewAndWorkbenchAliases(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"review_required","default_candidate_route":{"backend":"proxy","profile":"isolated","runtime_target":"node"},"explanation":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required"},"diagnostics":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required"},"review":{"policy_state":"popup_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required"},"display":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required"}},"workbench":{"ready":true,"sections":["route"],"explanation":{"category":"coordination","state":"managed_route_pending_default","summary_code":"managed_route_hidden_by_legacy_host_default"},"diagnostics":{"category":"coordination","state":"managed_route_pending_default","summary_code":"managed_route_hidden_by_legacy_host_default"},"summary":{"category":"coordination","state":"managed_route_pending_default","summary_code":"managed_route_hidden_by_legacy_host_default"},"review":{"policy_state":"popup_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required"},"display":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required"}}}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		DefaultCandidate browserRuntimeRouteDescriptor          `json:"default_candidate_route"`
		Explanation      *browserTopLevelSummary                `json:"explanation"`
		Diagnostics      *browserTopLevelSummary                `json:"diagnostics"`
		Review           *browserReviewSurfaceSummary           `json:"review"`
		Workbench        *browserRuntimeWorkbenchSurfaceSummary `json:"workbench"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if payload.DefaultCandidate != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) {
		t.Fatalf("expected root payload to keep default candidate route, got %#v", payload.DefaultCandidate)
	}
	if payload.Explanation == nil || payload.Explanation.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected explanation alias to inherit default candidate route, got %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil || payload.Diagnostics.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected diagnostics alias to inherit default candidate route, got %#v", payload.Diagnostics)
	}
	if payload.Review == nil ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.DefaultCandidateRoute != payload.DefaultCandidate ||
		payload.Review.Display == nil ||
		payload.Review.Display.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected review alias to inherit default candidate route, got %#v", payload.Review)
	}
	if payload.Workbench == nil ||
		payload.Workbench.DefaultCandidateRoute != payload.DefaultCandidate ||
		payload.Workbench.Explanation == nil ||
		payload.Workbench.Explanation.DefaultCandidateRoute != payload.DefaultCandidate ||
		payload.Workbench.Diagnostics == nil ||
		payload.Workbench.Diagnostics.DefaultCandidateRoute != payload.DefaultCandidate ||
		payload.Workbench.Summary == nil ||
		payload.Workbench.Summary.DefaultCandidateRoute != payload.DefaultCandidate ||
		payload.Workbench.Review == nil ||
		payload.Workbench.Review.Summary == nil ||
		payload.Workbench.Review.Summary.DefaultCandidateRoute != payload.DefaultCandidate ||
		payload.Workbench.Review.Display == nil ||
		payload.Workbench.Review.Display.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected workbench alias to inherit default candidate route, got %#v", payload.Workbench)
	}
}

func TestBrowserUnifiedApplyExplanationAliasPromotesDefaultCandidateRouteFromWorkbenchDisplayWhenRootMissing(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","workbench_display":{"ready":true,"sections":["route"],"category":"coordination","state":"managed_route_pending_default","summary_code":"managed_route_hidden_by_legacy_host_default","default_candidate_route":{"backend":"proxy","profile":"isolated","runtime_target":"node"},"primary_browser_action":"browser action=ready","next_step":"browser action=ready"}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		DefaultCandidate browserRuntimeRouteDescriptor          `json:"default_candidate_route"`
		Summary          *browserTopLevelSummary                `json:"summary"`
		Display          *browserTopLevelDisplaySummary         `json:"display"`
		View             *browserTopLevelViewSummary            `json:"view"`
		WorkbenchDisplay *browserRuntimeWorkbenchDisplaySummary `json:"workbench_display"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if payload.DefaultCandidate != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) {
		t.Fatalf("expected root payload to promote default candidate route from workbench_display, got %#v", payload.DefaultCandidate)
	}
	if payload.WorkbenchDisplay == nil || payload.WorkbenchDisplay.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected workbench_display to keep default candidate route, got %#v", payload.WorkbenchDisplay)
	}
	if payload.Summary == nil || payload.Summary.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected summary alias to inherit promoted default candidate route, got %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected display alias to inherit promoted default candidate route, got %#v", payload.Display)
	}
	if payload.View == nil || payload.View.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected view alias to inherit promoted default candidate route, got %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasPromotesDefaultCandidateRouteFromReviewWhenRootMissing(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"review_required","display":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","next_step_alias":"tabs","manual_retry_hint":"rerun_with_force","primary_browser_action":"browser action=tabs","next_step":"browser action=tabs"},"review":{"policy_state":"popup_review_required","decision":"session_target_popup_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","default_candidate_route":{"backend":"proxy","profile":"isolated","runtime_target":"node"},"next_step_alias":"tabs","manual_retry_hint":"rerun_with_force"},"display":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","default_candidate_route":{"backend":"proxy","profile":"isolated","runtime_target":"node"},"next_step_alias":"tabs","manual_retry_hint":"rerun_with_force","primary_browser_action":"browser action=tabs","next_step":"browser action=tabs"}}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		DefaultCandidate browserRuntimeRouteDescriptor  `json:"default_candidate_route"`
		Review           *browserReviewSurfaceSummary   `json:"review"`
		Surface          *browserTopLevelSurfaceSummary `json:"surface"`
		View             *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if payload.DefaultCandidate != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) {
		t.Fatalf("expected root payload to promote default candidate route from review, got %#v", payload.DefaultCandidate)
	}
	if payload.Review == nil ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.DefaultCandidateRoute != payload.DefaultCandidate ||
		payload.Review.Display == nil ||
		payload.Review.Display.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected review shell to keep promoted default candidate route, got %#v", payload.Review)
	}
	if payload.Surface == nil || payload.Surface.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected surface alias to inherit promoted default candidate route, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.DefaultCandidateRoute != payload.DefaultCandidate ||
		payload.View.Review == nil ||
		payload.View.Review.Summary == nil ||
		payload.View.Review.Summary.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected view alias to inherit promoted default candidate route, got %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasBuildsCapabilitySurfaceFromView(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","display":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual"},"view":{"kind":"workbench","category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","browser_tools":["browser_runtime","browser_act","browser_screenshot"],"artifact_tools":["browser_screenshot","browser_act"],"artifact_kinds":["screenshot"],"artifact_contract":"artifacts+media","browser_act_kinds":["click","screenshot"]}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		BrowserTools     []string                       `json:"browser_tools"`
		ArtifactTools    []string                       `json:"artifact_tools"`
		ArtifactKinds    []string                       `json:"artifact_kinds"`
		ArtifactContract string                         `json:"artifact_contract"`
		BrowserActKinds  []string                       `json:"browser_act_kinds"`
		Surface          *browserTopLevelSurfaceSummary `json:"surface"`
		View             *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_screenshot") ||
		!browserStringSliceContains(payload.ArtifactTools, "browser_act") ||
		!browserStringSliceContains(payload.ArtifactKinds, "screenshot") ||
		payload.ArtifactContract != browserArtifactContract ||
		!browserStringSliceContains(payload.BrowserActKinds, "click") {
		t.Fatalf("expected root payload to recover capability surface from view, got %#v", payload)
	}
	if payload.Surface == nil ||
		!browserStringSliceContains(payload.Surface.BrowserTools, "browser_screenshot") ||
		!browserStringSliceContains(payload.Surface.ArtifactTools, "browser_act") ||
		!browserStringSliceContains(payload.Surface.ArtifactKinds, "screenshot") ||
		payload.Surface.ArtifactContract != browserArtifactContract ||
		!browserStringSliceContains(payload.Surface.BrowserActKinds, "click") {
		t.Fatalf("expected surface alias to recover capability surface from view, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		!browserStringSliceContains(payload.View.BrowserTools, "browser_screenshot") ||
		!browserStringSliceContains(payload.View.ArtifactTools, "browser_act") ||
		!browserStringSliceContains(payload.View.ArtifactKinds, "screenshot") ||
		payload.View.ArtifactContract != browserArtifactContract ||
		!browserStringSliceContains(payload.View.BrowserActKinds, "click") {
		t.Fatalf("expected view alias to keep capability surface, got %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasBuildsViewFromSurfaceRouteSurface(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","display":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual"},"surface":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["sandbox"]}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Surface *browserTopLevelSurfaceSummary `json:"surface"`
		View    *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if payload.View == nil ||
		payload.View.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.View.BrowserOptInTargets) != 1 ||
		payload.View.BrowserOptInTargets[0] != "sandbox" {
		t.Fatalf("expected view alias to recover route surface from surface, got %#v", payload.View)
	}
	if payload.Surface == nil ||
		payload.Surface.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.Surface.BrowserOptInTargets) != 1 ||
		payload.Surface.BrowserOptInTargets[0] != "sandbox" {
		t.Fatalf("expected surface alias to keep route surface, got %#v", payload.Surface)
	}
}

func TestBrowserUnifiedApplyExplanationAliasPrefersSurfaceRouteOverExplicitViewRoute(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","surface":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["node"]},"view":{"kind":"review","category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required","browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["sandbox"]}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		Surface *browserTopLevelSurfaceSummary `json:"surface"`
		View    *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if payload.View == nil ||
		payload.View.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.View.BrowserOptInTargets) != 1 ||
		payload.View.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected explicit view alias to adopt surface route precedence, got %#v", payload.View)
	}
	if payload.Surface == nil ||
		payload.Surface.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.Surface.BrowserOptInTargets) != 1 ||
		payload.Surface.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected surface alias to keep primary route surface, got %#v", payload.Surface)
	}
}

func TestBrowserUnifiedApplyExplanationAliasBuildsRouteSurfaceFromWorkbenchRouteSurface(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","display":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual"},"workbench":{"ready":true,"sections":["coordination"],"browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["node"],"summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"}}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		BrowserSurface      string                                 `json:"browser_surface"`
		BrowserOptInTargets []string                               `json:"browser_opt_in_targets"`
		Workbench           *browserRuntimeWorkbenchSurfaceSummary `json:"workbench"`
		Surface             *browserTopLevelSurfaceSummary         `json:"surface"`
		View                *browserTopLevelViewSummary            `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if payload.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.BrowserOptInTargets) != 1 ||
		payload.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected root payload to recover route surface from workbench, got %#v", payload)
	}
	if payload.Workbench == nil ||
		payload.Workbench.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.Workbench.BrowserOptInTargets) != 1 ||
		payload.Workbench.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected workbench shell to keep route surface, got %#v", payload.Workbench)
	}
	if payload.Surface == nil ||
		payload.Surface.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.Surface.BrowserOptInTargets) != 1 ||
		payload.Surface.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected surface alias to recover route surface from workbench, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.View.BrowserOptInTargets) != 1 ||
		payload.View.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected view alias to recover route surface from workbench, got %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasBuildsCapabilitySurfaceFromWorkbench(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","display":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual"},"workbench":{"ready":true,"sections":["coordination"],"browser_tools":["browser_runtime","browser_act"],"artifact_tools":["browser_act"],"browser_act_kinds":["click"],"summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"}}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		BrowserTools    []string                               `json:"browser_tools"`
		ArtifactTools   []string                               `json:"artifact_tools"`
		BrowserActKinds []string                               `json:"browser_act_kinds"`
		Workbench       *browserRuntimeWorkbenchSurfaceSummary `json:"workbench"`
		Surface         *browserTopLevelSurfaceSummary         `json:"surface"`
		View            *browserTopLevelViewSummary            `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_runtime") ||
		!browserStringSliceContains(payload.ArtifactTools, "browser_act") ||
		!browserStringSliceContains(payload.BrowserActKinds, "click") {
		t.Fatalf("expected root payload to recover capability surface from workbench, got %#v", payload)
	}
	if payload.Workbench == nil ||
		!browserStringSliceContains(payload.Workbench.BrowserTools, "browser_runtime") ||
		!browserStringSliceContains(payload.Workbench.ArtifactTools, "browser_act") ||
		!browserStringSliceContains(payload.Workbench.BrowserActKinds, "click") {
		t.Fatalf("expected workbench shell to keep capability surface, got %#v", payload.Workbench)
	}
	if payload.Surface == nil ||
		!browserStringSliceContains(payload.Surface.BrowserTools, "browser_runtime") ||
		!browserStringSliceContains(payload.Surface.ArtifactTools, "browser_act") ||
		!browserStringSliceContains(payload.Surface.BrowserActKinds, "click") {
		t.Fatalf("expected surface alias to recover capability surface from workbench, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		!browserStringSliceContains(payload.View.BrowserTools, "browser_runtime") ||
		!browserStringSliceContains(payload.View.ArtifactTools, "browser_act") ||
		!browserStringSliceContains(payload.View.BrowserActKinds, "click") {
		t.Fatalf("expected view alias to recover capability surface from workbench, got %#v", payload.View)
	}
}

func TestBrowserUnifiedApplyExplanationAliasPrefersRootRouteOverNestedRouteSurface(t *testing.T) {
	out, err := browserUnifiedApplyExplanationAlias(`{"status":"ok","browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["sandbox"],"surface":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["node"]},"view":{"kind":"workbench","category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["node"]},"workbench":{"ready":true,"sections":["coordination"],"browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["node"],"summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"}}}`)
	if err != nil {
		t.Fatalf("browserUnifiedApplyExplanationAlias: %v", err)
	}
	var payload struct {
		BrowserSurface      string                                 `json:"browser_surface"`
		BrowserOptInTargets []string                               `json:"browser_opt_in_targets"`
		Surface             *browserTopLevelSurfaceSummary         `json:"surface"`
		View                *browserTopLevelViewSummary            `json:"view"`
		Workbench           *browserRuntimeWorkbenchSurfaceSummary `json:"workbench"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified aliased payload: %v", err)
	}
	if payload.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.BrowserOptInTargets) != 1 ||
		payload.BrowserOptInTargets[0] != "sandbox" {
		t.Fatalf("expected root payload to keep explicit route surface, got %#v", payload)
	}
	if payload.Surface == nil ||
		payload.Surface.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.Surface.BrowserOptInTargets) != 1 ||
		payload.Surface.BrowserOptInTargets[0] != "sandbox" {
		t.Fatalf("expected surface alias to adopt root route precedence, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.View.BrowserOptInTargets) != 1 ||
		payload.View.BrowserOptInTargets[0] != "sandbox" {
		t.Fatalf("expected view alias to adopt root route precedence, got %#v", payload.View)
	}
	if payload.Workbench == nil ||
		payload.Workbench.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.Workbench.BrowserOptInTargets) != 1 ||
		payload.Workbench.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected workbench shell to keep its original route surface, got %#v", payload.Workbench)
	}
}
