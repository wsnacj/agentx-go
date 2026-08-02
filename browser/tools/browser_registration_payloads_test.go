package tools

import (
	"encoding/json"
	"testing"
)

func TestMarshalBrowserScreenshotPayloadAppliesSelectedRouteCapabilityMetadata(t *testing.T) {
	out, err := marshalBrowserScreenshotPayload(browserScreenshotToolPayload{
		browserRegistrationRouteCapabilityPayload: browserRegistrationRouteCapabilityPayload{
			CapabilityMetadata: browserRuntimeCapabilityMetadata{
				BrowserTools:        []string{"browser_runtime", "browser_act", "browser_screenshot"},
				ArtifactTools:       []string{"browser_screenshot", "browser_act"},
				ArtifactKinds:       []string{"screenshot"},
				ArtifactContract:    browserArtifactContract,
				BrowserActKinds:     []string{"screenshot", "click"},
				BrowserSurface:      "explicit_managed_opt_in",
				BrowserOptInTargets: []string{"node"},
			},
		},
		Backend:       "proxy",
		BrowserApp:    "Chromium",
		Profile:       "workbench",
		RuntimeTarget: "node",
		Path:          ".agentx/browser/screenshot-test.png",
		Bytes:         123,
		Status:        "captured",
	})
	if err != nil {
		t.Fatalf("marshalBrowserScreenshotPayload: %v", err)
	}

	var payload struct {
		BrowserTools        []string                       `json:"browser_tools"`
		ArtifactTools       []string                       `json:"artifact_tools"`
		ArtifactKinds       []string                       `json:"artifact_kinds"`
		ArtifactContract    string                         `json:"artifact_contract"`
		BrowserActKinds     []string                       `json:"browser_act_kinds"`
		BrowserSurface      string                         `json:"browser_surface"`
		BrowserOptInTargets []string                       `json:"browser_opt_in_targets"`
		Surface             *browserTopLevelSurfaceSummary `json:"surface"`
		View                *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_screenshot") ||
		!browserStringSliceContains(payload.BrowserTools, "browser_act") ||
		!browserStringSliceContains(payload.ArtifactTools, "browser_screenshot") ||
		!browserStringSliceContains(payload.ArtifactTools, "browser_act") ||
		!browserStringSliceContains(payload.ArtifactKinds, "screenshot") ||
		payload.ArtifactContract != browserArtifactContract ||
		!browserStringSliceContains(payload.BrowserActKinds, "screenshot") ||
		payload.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.BrowserOptInTargets) != 1 ||
		payload.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected screenshot payload to expose selected-route capability metadata, got %#v", payload)
	}
	if payload.Surface == nil ||
		!browserStringSliceContains(payload.Surface.BrowserTools, "browser_screenshot") ||
		!browserStringSliceContains(payload.Surface.ArtifactTools, "browser_act") ||
		!browserStringSliceContains(payload.Surface.ArtifactKinds, "screenshot") ||
		payload.Surface.ArtifactContract != browserArtifactContract ||
		!browserStringSliceContains(payload.Surface.BrowserActKinds, "click") {
		t.Fatalf("expected screenshot surface alias to carry selected-route capability metadata, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		!browserStringSliceContains(payload.View.BrowserTools, "browser_screenshot") ||
		!browserStringSliceContains(payload.View.ArtifactTools, "browser_act") ||
		!browserStringSliceContains(payload.View.ArtifactKinds, "screenshot") ||
		payload.View.ArtifactContract != browserArtifactContract ||
		!browserStringSliceContains(payload.View.BrowserActKinds, "click") {
		t.Fatalf("expected screenshot view alias to carry selected-route capability metadata, got %#v", payload.View)
	}
}

func TestMarshalBrowserOpenPayloadPreservesDefaultCandidateRouteInProjectedSurfaceAndView(t *testing.T) {
	route := browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}
	out, err := marshalBrowserOpenPayload(browserOpenToolPayload{
		URL:     "https://93.184.216.34",
		Backend: "proxy-open",
		Status:  "opened",
		Display: &browserTopLevelDisplaySummary{
			Category:              "navigation",
			State:                 "completed",
			SummaryCode:           "open_completed",
			DefaultCandidateRoute: route,
		},
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
	if err != nil {
		t.Fatalf("marshalBrowserOpenPayload: %v", err)
	}

	var payload struct {
		DefaultCandidate browserRuntimeRouteDescriptor  `json:"default_candidate_route"`
		Explanation      *browserTopLevelSummary        `json:"explanation"`
		Diagnostics      *browserTopLevelSummary        `json:"diagnostics"`
		Display          *browserTopLevelDisplaySummary `json:"display"`
		Review           *browserReviewSurfaceSummary   `json:"review"`
		Surface          *browserTopLevelSurfaceSummary `json:"surface"`
		View             *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.DefaultCandidate != route {
		t.Fatalf("expected open payload root default candidate route, got %#v", payload.DefaultCandidate)
	}
	if payload.Explanation == nil || payload.Explanation.DefaultCandidateRoute != route {
		t.Fatalf("expected open payload explanation to keep default candidate route, got %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil || payload.Diagnostics.DefaultCandidateRoute != route {
		t.Fatalf("expected open payload diagnostics to keep default candidate route, got %#v", payload.Diagnostics)
	}
	if payload.Display == nil || payload.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected open payload display to keep default candidate route, got %#v", payload.Display)
	}
	if payload.Review == nil ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.DefaultCandidateRoute != route ||
		payload.Review.Display == nil ||
		payload.Review.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected open payload review to keep default candidate route, got %#v", payload.Review)
	}
	if payload.Surface == nil || payload.Surface.DefaultCandidateRoute != route {
		t.Fatalf("expected projected surface to preserve default candidate route, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.DefaultCandidateRoute != route ||
		payload.View.Review == nil ||
		payload.View.Review.Summary == nil ||
		payload.View.Review.Summary.DefaultCandidateRoute != route ||
		payload.View.Review.Display == nil ||
		payload.View.Review.Display.DefaultCandidateRoute != route {
		t.Fatalf("expected projected view to preserve default candidate route across nested review shells, got %#v", payload.View)
	}
}
