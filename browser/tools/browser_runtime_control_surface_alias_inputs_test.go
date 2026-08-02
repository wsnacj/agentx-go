package tools

import (
	"encoding/json"
	"testing"
)

func TestBrowserRuntimeTopLevelAliasInputsFromPayloadDecodesAliasFields(t *testing.T) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(`{
		"workbench_explanation":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot"},
		"review":{"policy_state":"popup_review_required","decision":"session_target_popup_review_required","summary":{"category":"review","state":"manual_confirmation_required","summary_code":"popup_review_required"}},
		"display":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan","primary_browser_action":"browser action=refresh"},
		"default_candidate_route":{"backend":"proxy","profile":"isolated","runtime_target":"node"},
		"browser_tools":["browser_runtime","browser_runtime","browser_act"],
		"artifact_contract":" artifacts+media ",
		"browser_act_kinds":["click","click"],
		"workbench":{"ready":true,"sections":["coordination"],"browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["node"],"summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"}}
	}`), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	inputs := browserRuntimeTopLevelAliasInputsFromPayload(payload)

	if inputs.workbenchExplanation == nil ||
		inputs.workbenchExplanation.Category != "resolver" ||
		inputs.workbenchExplanation.SummaryCode != "label_filtered_residual" {
		t.Fatalf("expected workbench explanation to decode into top-level summary, got %#v", inputs.workbenchExplanation)
	}
	if inputs.review == nil ||
		inputs.review.PolicyState != "popup_review_required" ||
		inputs.review.Decision != "session_target_popup_review_required" {
		t.Fatalf("expected review shell to decode, got %#v", inputs.review)
	}
	if inputs.display == nil ||
		inputs.display.Category != "coordination" ||
		inputs.display.PrimaryBrowserAction != "browser action=refresh" {
		t.Fatalf("expected display shell to decode, got %#v", inputs.display)
	}
	if len(inputs.browserTools) != 2 ||
		inputs.browserTools[0] != "browser_runtime" ||
		inputs.browserTools[1] != "browser_act" {
		t.Fatalf("expected browser tools to decode and dedupe, got %#v", inputs.browserTools)
	}
	if inputs.artifactContract != browserArtifactContract {
		t.Fatalf("expected artifact contract to trim and decode, got %#v", inputs.artifactContract)
	}
	if len(inputs.browserActKinds) != 1 || inputs.browserActKinds[0] != "click" {
		t.Fatalf("expected browser_act kinds to decode and dedupe, got %#v", inputs.browserActKinds)
	}
	if inputs.workbench == nil ||
		inputs.workbench.BrowserSurface != "explicit_managed_opt_in" ||
		len(inputs.workbench.BrowserOptInTargets) != 1 ||
		inputs.workbench.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected workbench shell to decode route surface, got %#v", inputs.workbench)
	}
	if inputs.defaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) {
		t.Fatalf("expected default candidate route to decode, got %#v", inputs.defaultCandidateRoute)
	}
	if inputs.browserSurface != "explicit_managed_opt_in" ||
		len(inputs.browserOptInTargets) != 1 ||
		inputs.browserOptInTargets[0] != "node" {
		t.Fatalf("expected root route surface to inherit from decoded workbench, got %#v", inputs)
	}
}

func TestBrowserRuntimeTopLevelReviewFromInputsPromotesDefaultCandidateRoute(t *testing.T) {
	inputs := browserRuntimeTopLevelAliasInputs{
		defaultCandidateRoute: browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"},
		review: &browserReviewSurfaceSummary{
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
	}

	review, ok := browserRuntimeTopLevelReviewFromInputs(inputs)

	if !ok {
		t.Fatalf("expected review alias to build from inputs")
	}
	if review.Summary == nil ||
		review.Summary.DefaultCandidateRoute != inputs.defaultCandidateRoute ||
		review.Display == nil ||
		review.Display.DefaultCandidateRoute != inputs.defaultCandidateRoute {
		t.Fatalf("expected review alias to inherit default candidate route, got %#v", review)
	}
}

func TestBrowserRuntimeTopLevelExplanationAndDiagnosticsAliasFromInputsPromoteDefaultCandidateRoute(t *testing.T) {
	inputs := browserRuntimeTopLevelAliasInputs{
		defaultCandidateRoute: browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"},
		workbenchExplanation: &browserTopLevelSummary{
			Category:    "coordination",
			State:       "action_plan_available",
			SummaryCode: "workbench_action_plan",
		},
		workbenchDiagnostics: &browserTopLevelSummary{
			Category:    "resolver",
			State:       "manual_resolution_required",
			SummaryCode: "label_filtered_residual",
		},
	}

	explanation, ok := browserRuntimeTopLevelExplanationAliasFromInputs(inputs)
	if !ok {
		t.Fatalf("expected explanation alias to build from inputs")
	}
	if explanation.DefaultCandidateRoute != inputs.defaultCandidateRoute {
		t.Fatalf("expected explanation alias to inherit default candidate route, got %#v", explanation)
	}

	diagnostics, ok := browserRuntimeTopLevelDiagnosticsAliasFromInputs(inputs)
	if !ok {
		t.Fatalf("expected diagnostics alias to build from inputs")
	}
	if diagnostics.DefaultCandidateRoute != inputs.defaultCandidateRoute {
		t.Fatalf("expected diagnostics alias to inherit default candidate route, got %#v", diagnostics)
	}
}

func TestBrowserRuntimeApplyTopLevelRouteCapabilityAliasProjectionBuildsFallbackFields(t *testing.T) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(`{"status":"ok"}`), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	inputs := browserRuntimeTopLevelAliasInputs{
		workbench: &browserRuntimeWorkbenchSurfaceSummary{
			Ready:               true,
			Sections:            []string{"coordination"},
			BrowserSurface:      "explicit_managed_opt_in",
			BrowserOptInTargets: []string{"node"},
			BrowserTools:        []string{"browser_runtime", "browser_act"},
			ArtifactTools:       []string{"browser.artifact.resolve"},
			ArtifactKinds:       []string{"download"},
			ArtifactContract:    browserArtifactContract,
			BrowserActKinds:     []string{"open"},
		},
	}

	mutated, err := browserRuntimeApplyTopLevelRouteCapabilityAliasProjection(payload, &inputs)
	if err != nil {
		t.Fatalf("apply route/capability overlay: %v", err)
	}
	if !mutated {
		t.Fatalf("expected route/capability overlay to mutate payload")
	}
	var got struct {
		BrowserSurface      string   `json:"browser_surface"`
		BrowserOptInTargets []string `json:"browser_opt_in_targets"`
		BrowserTools        []string `json:"browser_tools"`
		ArtifactTools       []string `json:"artifact_tools"`
		ArtifactKinds       []string `json:"artifact_kinds"`
		ArtifactContract    string   `json:"artifact_contract"`
		BrowserActKinds     []string `json:"browser_act_kinds"`
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := json.Unmarshal(blob, &got); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if got.BrowserSurface != "explicit_managed_opt_in" || len(got.BrowserOptInTargets) != 1 || got.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected route overlay to be recovered from workbench, got %#v", got)
	}
	if len(got.BrowserTools) != 2 || got.BrowserTools[0] != "browser_runtime" || got.BrowserTools[1] != "browser_act" {
		t.Fatalf("expected capability overlay to recover browser tools, got %#v", got.BrowserTools)
	}
	if len(got.ArtifactTools) != 1 || got.ArtifactTools[0] != "browser.artifact.resolve" {
		t.Fatalf("expected artifact tools overlay, got %#v", got.ArtifactTools)
	}
	if len(got.ArtifactKinds) != 1 || got.ArtifactKinds[0] != "download" || got.ArtifactContract != browserArtifactContract {
		t.Fatalf("expected artifact surface overlay, got %#v", got)
	}
	if len(got.BrowserActKinds) != 1 || got.BrowserActKinds[0] != "open" {
		t.Fatalf("expected browser_act kinds overlay, got %#v", got.BrowserActKinds)
	}
}

func TestBrowserRuntimeApplyTopLevelRouteCapabilityAliasProjectionPreservesExistingFields(t *testing.T) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(`{
		"browser_surface":"explicit_managed_opt_in",
		"browser_opt_in_targets":["sandbox"],
		"browser_tools":["browser_runtime"],
		"artifact_tools":["browser_screenshot"],
		"artifact_kinds":["screenshot"],
		"artifact_contract":"artifacts+media",
		"browser_act_kinds":["click"]
	}`), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	inputs := browserRuntimeTopLevelAliasInputs{
		workbench: &browserRuntimeWorkbenchSurfaceSummary{
			Ready:               true,
			Sections:            []string{"coordination"},
			BrowserSurface:      "explicit_managed_opt_in",
			BrowserOptInTargets: []string{"node"},
			BrowserTools:        []string{"browser_runtime", "browser_act"},
			ArtifactTools:       []string{"browser.artifact.resolve"},
			ArtifactKinds:       []string{"download"},
			ArtifactContract:    browserArtifactContract,
			BrowserActKinds:     []string{"open"},
		},
	}

	mutated, err := browserRuntimeApplyTopLevelRouteCapabilityAliasProjection(payload, &inputs)
	if err != nil {
		t.Fatalf("apply route/capability overlay: %v", err)
	}
	if mutated {
		t.Fatalf("expected existing top-level route/capability fields to remain unchanged")
	}
}

func TestBrowserRuntimeApplyTopLevelDefaultCandidateAliasProjectionOverlaysNestedShells(t *testing.T) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(`{
		"default_candidate_route":{"backend":"proxy","profile":"isolated","runtime_target":"node"},
		"explanation":{"summary_code":"top_level_explanation"},
		"review":{"policy_state":"popup_review_required","summary":{"summary_code":"popup_review_required"}},
		"workbench":{"ready":true,"sections":["route"],"summary":{"summary_code":"workbench_summary"}}
	}`), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	inputs := browserRuntimeTopLevelAliasInputsFromPayload(payload)

	mutated, err := browserRuntimeApplyTopLevelDefaultCandidateAliasProjection(payload, &inputs)
	if err != nil {
		t.Fatalf("apply default-candidate overlay: %v", err)
	}
	if !mutated {
		t.Fatalf("expected default-candidate overlay to mutate payload")
	}

	var got struct {
		DefaultCandidate browserRuntimeRouteDescriptor          `json:"default_candidate_route"`
		Explanation      *browserTopLevelSummary                `json:"explanation"`
		Review           *browserReviewSurfaceSummary           `json:"review"`
		Workbench        *browserRuntimeWorkbenchSurfaceSummary `json:"workbench"`
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if err := json.Unmarshal(blob, &got); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if got.DefaultCandidate != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) {
		t.Fatalf("expected root default candidate route to remain intact, got %#v", got.DefaultCandidate)
	}
	if got.Explanation == nil || got.Explanation.DefaultCandidateRoute != got.DefaultCandidate {
		t.Fatalf("expected explanation shell to inherit default candidate route, got %#v", got.Explanation)
	}
	if got.Review == nil || got.Review.Summary == nil || got.Review.Summary.DefaultCandidateRoute != got.DefaultCandidate {
		t.Fatalf("expected review shell to inherit default candidate route, got %#v", got.Review)
	}
	if got.Workbench == nil || got.Workbench.DefaultCandidateRoute != got.DefaultCandidate || got.Workbench.Summary == nil || got.Workbench.Summary.DefaultCandidateRoute != got.DefaultCandidate {
		t.Fatalf("expected workbench shell to inherit default candidate route, got %#v", got.Workbench)
	}
}

func TestBrowserRuntimeApplyTopLevelAliasProjectionBuildsAliasChain(t *testing.T) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal([]byte(`{
		"status":"ok",
		"workbench_explanation":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"},
		"diagnostics_explanation":{"category":"resolver","state":"manual_resolution_required","summary_code":"label_filtered_residual","next_step_alias":"snapshot","manual_retry_hint":"add_ordinal"},
		"workbench":{"ready":true,"sections":["coordination"],"browser_surface":"explicit_managed_opt_in","browser_opt_in_targets":["node"],"summary":{"category":"coordination","state":"action_plan_available","summary_code":"workbench_action_plan"}},
		"browser_tools":["browser_runtime"],
		"artifact_tools":["browser.artifact.resolve"],
		"artifact_kinds":["download"],
		"artifact_contract":"artifacts+media",
		"browser_act_kinds":["open"]
	}`), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	inputs := browserRuntimeTopLevelAliasInputsFromPayload(payload)

	mutated, err := browserRuntimeApplyTopLevelAliasProjection(payload, &inputs)
	if err != nil {
		t.Fatalf("apply top-level alias projection: %v", err)
	}
	if !mutated {
		t.Fatalf("expected top-level alias projection to mutate payload")
	}
	if !browserUnifiedHasNonNullJSONField(payload, "explanation") ||
		!browserUnifiedHasNonNullJSONField(payload, "diagnostics") ||
		!browserUnifiedHasNonNullJSONField(payload, "summary") ||
		!browserUnifiedHasNonNullJSONField(payload, "display") ||
		!browserUnifiedHasNonNullJSONField(payload, "surface") ||
		!browserUnifiedHasNonNullJSONField(payload, "view") {
		t.Fatalf("expected alias chain to populate explanation/diagnostics/summary/display/surface/view, got %#v", payload)
	}
}
