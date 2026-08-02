package tools

import (
	"reflect"
	"strings"
	"testing"
)

func TestBrowserToolMetadataRouteHintsRoundTrip(t *testing.T) {
	meta := ApplyBrowserToolMetadataRouteHints(
		ToolMetadata{
			Type:         "browser",
			Capabilities: []string{"browser", "browser_deprecated_compat_wrapper"},
		},
		BrowserToolMetadataRouteHints{
			RuntimeInfo: BrowserRuntimeInfo{
				Backend: " Proxy ",
				Profile: " Workbench ",
				Target:  " Node ",
			},
			Source:   " Managed_Browserd ",
			Endpoint: " HTTP://127.0.0.1:43123 ",
			DefaultCandidateRoute: BrowserRuntimeInfo{
				Backend: " proxy ",
				Profile: " isolated ",
				Target:  " node ",
			},
			DefaultCandidateSource:   " Managed_Browserd ",
			DefaultCandidateEndpoint: " HTTP://127.0.0.1:43123 ",
			Surface:                  " explicit_managed_opt_in ",
			OptInTargets:             []string{" Node ", "sandbox", "node"},
			HiddenDefaultCandidate:   true,
			SubstratePosture:         " Node_Runtime ",
			SubstrateStatus:          " OK ",
			SelectionStrategy:        " Prefer_Node_Over_Legacy_Host ",
			SubstrateReason:          " Bundled browserd bootstrap is blocked because `node` is not available in PATH. ",
			RepairAvailable:          true,
		},
	)
	got := BrowserToolMetadataRouteHintsForMetadata(meta)
	want := BrowserToolMetadataRouteHints{
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "workbench",
			Target:  "node",
		},
		Source:   "managed_browserd",
		Endpoint: "http://127.0.0.1:43123",
		DefaultCandidateRoute: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		DefaultCandidateSource:   "managed_browserd",
		DefaultCandidateEndpoint: "http://127.0.0.1:43123",
		Surface:                  "explicit_managed_opt_in",
		OptInTargets:             []string{"node", "sandbox"},
		HiddenDefaultCandidate:   true,
		SubstratePosture:         "node_runtime",
		SubstrateStatus:          "ok",
		SelectionStrategy:        "prefer_node_over_legacy_host",
		SubstrateReason:          "bundled browserd bootstrap is blocked because `node` is not available in path.",
		RepairAvailable:          true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("route hints round-trip mismatch: got %#v want %#v", got, want)
	}
}

func TestResolveBrowserToolMetadataRouteHintsIncludesBrowserActMetadata(t *testing.T) {
	metadata := map[string]ToolMetadata{
		"browser_click": {
			Type:         "browser",
			Capabilities: []string{"browser", "browser_deprecated_compat_wrapper"},
		},
		"browser_runtime": {
			Type: "browser",
			Capabilities: []string{
				"browser_runtime",
				"browser_source:legacy_host",
				"browser_endpoint:http://127.0.0.1:42123",
				"browser_backend:custom-playwright",
				"browser_profile:workbench",
				"browser_target:host",
				"browser_default_candidate:hidden_managed",
				"browser_default_candidate_source:managed_browserd",
				"browser_default_candidate_endpoint:http://127.0.0.1:43123",
				"browser_default_candidate_backend:proxy",
				"browser_default_candidate_profile:isolated",
				"browser_default_candidate_target:node",
				"browser_substrate_posture:host_runtime",
				"browser_substrate_status:ok",
				"browser_substrate_selection_strategy:prefer_host_runtime",
			},
		},
		"browser_act": ApplyBrowserToolMetadataRouteHints(
			ToolMetadata{
				Type:         "browser",
				Capabilities: []string{"browser", "browser_act"},
			},
			BrowserToolMetadataRouteHints{
				Surface:      "explicit_managed_opt_in",
				OptInTargets: []string{"node"},
			},
		),
	}

	got := ResolveBrowserToolMetadataRouteHints([]string{"browser_click", "browser_runtime"}, metadata)
	want := BrowserToolMetadataRouteHints{
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "custom-playwright",
			Profile: "workbench",
			Target:  "host",
		},
		Source:   "legacy_host",
		Endpoint: "http://127.0.0.1:42123",
		DefaultCandidateRoute: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		DefaultCandidateSource:   "managed_browserd",
		DefaultCandidateEndpoint: "http://127.0.0.1:43123",
		Surface:                  "explicit_managed_opt_in",
		OptInTargets:             []string{"node"},
		HiddenDefaultCandidate:   true,
		SubstratePosture:         "host_runtime",
		SubstrateStatus:          "ok",
		SelectionStrategy:        "prefer_host_runtime",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolved route hints mismatch: got %#v want %#v", got, want)
	}
}

func TestResolveBrowserToolMetadataRouteHintsPrefersHiddenManagedNodeCandidate(t *testing.T) {
	metadata := map[string]ToolMetadata{
		"browser_runtime": {
			Type: "browser",
			Capabilities: []string{
				"browser_runtime",
				"browser_source:legacy_host",
				"browser_endpoint:http://127.0.0.1:42123",
				"browser_backend:system",
				"browser_profile:default",
				"browser_target:host",
				"browser_default_candidate:hidden_managed",
				"browser_default_candidate_source:managed_browserd",
				"browser_default_candidate_endpoint:http://127.0.0.1:43123",
				"browser_default_candidate_backend:proxy",
				"browser_default_candidate_profile:isolated",
				"browser_default_candidate_target:node",
				"browser_substrate_posture:node_runtime",
				"browser_substrate_status:ok",
				"browser_substrate_selection_strategy:prefer_node_over_legacy_host",
			},
		},
	}

	got := ResolveBrowserToolMetadataRouteHints([]string{"browser_runtime"}, metadata)
	want := BrowserToolMetadataRouteHints{
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		Source:   "managed_browserd",
		Endpoint: "http://127.0.0.1:43123",
		DefaultCandidateRoute: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		DefaultCandidateSource:   "managed_browserd",
		DefaultCandidateEndpoint: "http://127.0.0.1:43123",
		HiddenDefaultCandidate:   true,
		SubstratePosture:         "node_runtime",
		SubstrateStatus:          "ok",
		SelectionStrategy:        "prefer_node_over_legacy_host",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resolved route hints should prefer hidden managed candidate runtime info: got %#v want %#v", got, want)
	}
}

func TestBrowserToolMetadataRouteHintsPreferredRuntimeInfoPrefersHiddenManagedNodeCandidate(t *testing.T) {
	hints := BrowserToolMetadataRouteHints{
		Source:   "legacy_host",
		Endpoint: "http://127.0.0.1:42123",
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "system",
			Profile: "default",
			Target:  "host",
		},
		DefaultCandidateSource:   "managed_browserd",
		DefaultCandidateEndpoint: "http://127.0.0.1:43123",
		DefaultCandidateRoute: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		HiddenDefaultCandidate: true,
		SelectionStrategy:      BrowserSubstrateSelectionPreferNodeOverLegacy,
		SubstratePosture:       BrowserSubstrateNodeRuntime,
	}
	if got := hints.PreferredRuntimeInfo(); !reflect.DeepEqual(got, hints.DefaultCandidateRoute) {
		t.Fatalf("expected hidden managed node candidate to become preferred runtime, got %#v", got)
	}
}

func TestBrowserToolMetadataRouteHintsPreferredRuntimeInfoKeepsVisibleHostRuntime(t *testing.T) {
	hints := BrowserToolMetadataRouteHints{
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "custom-playwright",
			Profile: "workbench",
			Target:  "host",
		},
		DefaultCandidateRoute: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		HiddenDefaultCandidate: true,
		SelectionStrategy:      BrowserSubstrateSelectionPreferHostRuntime,
		SubstratePosture:       BrowserSubstrateHostRuntime,
	}
	if got := hints.PreferredRuntimeInfo(); !reflect.DeepEqual(got, hints.RuntimeInfo) {
		t.Fatalf("expected visible host runtime to remain preferred, got %#v", got)
	}
}

func TestBrowserToolMetadataRouteHintsWithPreferredRuntimeInfoPromotesHiddenManagedNodeCandidate(t *testing.T) {
	hints := BrowserToolMetadataRouteHints{
		Source:   "legacy_host",
		Endpoint: "http://127.0.0.1:42123",
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "system",
			Profile: "default",
			Target:  "host",
		},
		DefaultCandidateSource:   "managed_browserd",
		DefaultCandidateEndpoint: "http://127.0.0.1:43123",
		DefaultCandidateRoute: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		HiddenDefaultCandidate: true,
		SelectionStrategy:      BrowserSubstrateSelectionPreferNodeOverLegacy,
		SubstratePosture:       BrowserSubstrateNodeRuntime,
	}
	got := hints.WithPreferredRuntimeInfo()
	if !reflect.DeepEqual(got.RuntimeInfo, hints.DefaultCandidateRoute) {
		t.Fatalf("expected preferred runtime projection to promote hidden managed node candidate, got %#v", got)
	}
	if got.Source != hints.DefaultCandidateSource || got.Endpoint != hints.DefaultCandidateEndpoint {
		t.Fatalf("expected preferred runtime projection to promote hidden managed candidate provenance, got %#v", got)
	}
}

func TestBrowserToolMetadataRouteHintsEmptyRecognizesZeroAndNonZeroHints(t *testing.T) {
	if !(BrowserToolMetadataRouteHints{}).Empty() {
		t.Fatalf("expected zero browser route hints to be empty")
	}
	if (BrowserToolMetadataRouteHints{
		Source:                 "managed_browserd",
		Endpoint:               "http://127.0.0.1:43123",
		DefaultCandidateRoute:  BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		HiddenDefaultCandidate: true,
	}).Empty() {
		t.Fatalf("expected browser route hints with managed candidate to be non-empty")
	}
}

func TestBrowserToolMetadataRouteHintsDynamicGroupSelectorsFromHiddenManagedCandidate(t *testing.T) {
	hints := BrowserToolMetadataRouteHints{
		Source:                 "managed_browserd",
		Endpoint:               "http://127.0.0.1:43123",
		RuntimeInfo:            BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		DefaultCandidateRoute:  BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		HiddenDefaultCandidate: true,
		SubstratePosture:       "node_runtime",
		SubstrateStatus:        "ok",
		SelectionStrategy:      "prefer_node_over_legacy_host",
		RepairAvailable:        true,
	}
	got := hints.DynamicGroupSelectors()
	want := []string{
		"group:browser-bootstrap-repairable",
		"group:browser-default-candidate:hidden_managed",
		"group:browser-default-selection:prefer_node_over_legacy_host",
		"group:browser-substrate-status:ok",
		"group:browser-substrate:node_runtime",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected browser dynamic group selectors: got=%#v want=%#v", got, want)
	}
}

func TestBrowserToolMetadataRouteHintsDefaultCandidateLabel(t *testing.T) {
	if got := (BrowserToolMetadataRouteHints{}).DefaultCandidateLabel(); got != "" {
		t.Fatalf("expected empty default candidate label for zero hints, got %q", got)
	}
	if got := (BrowserToolMetadataRouteHints{HiddenDefaultCandidate: true}).DefaultCandidateLabel(); got != "hidden_managed" {
		t.Fatalf("expected hidden managed default candidate label, got %q", got)
	}
}

func TestBrowserToolMetadataRouteHintsEffectiveDefaultCandidateRouteFallsBackToVisibleManagedRuntime(t *testing.T) {
	hints := BrowserToolMetadataRouteHints{
		Source:   "managed_browserd",
		Endpoint: "http://127.0.0.1:43123",
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "workbench",
			Target:  "node",
		},
		HiddenDefaultCandidate: true,
	}
	if got := hints.EffectiveDefaultCandidateRoute(); !reflect.DeepEqual(got, hints.RuntimeInfo) {
		t.Fatalf("expected label-only hidden managed candidate to fall back to visible managed runtime, got %#v", got)
	}
	if canonical := hints.Canonicalized(); canonical.DefaultCandidateSource != "managed_browserd" || canonical.DefaultCandidateEndpoint != "http://127.0.0.1:43123" {
		t.Fatalf("expected label-only hidden managed candidate to preserve visible runtime provenance, got %#v", canonical)
	}
}

func TestResolveBrowserToolMetadataRouteHintsPromotesLabelOnlyHiddenManagedRuntimeCandidate(t *testing.T) {
	metadata := map[string]ToolMetadata{
		"browser_runtime": {
			Type: "browser",
			Capabilities: []string{
				"browser_runtime",
				"browser_source:managed_browserd",
				"browser_endpoint:http://127.0.0.1:43123",
				"browser_backend:proxy",
				"browser_profile:workbench",
				"browser_target:node",
				"browser_default_candidate:hidden_managed",
				"browser_substrate_posture:node_runtime",
				"browser_substrate_selection_strategy:prefer_node_runtime",
			},
		},
	}

	got := ResolveBrowserToolMetadataRouteHints([]string{"browser_runtime"}, metadata)
	want := BrowserToolMetadataRouteHints{
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "workbench",
			Target:  "node",
		},
		Source:   "managed_browserd",
		Endpoint: "http://127.0.0.1:43123",
		DefaultCandidateRoute: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "workbench",
			Target:  "node",
		},
		DefaultCandidateSource:   "managed_browserd",
		DefaultCandidateEndpoint: "http://127.0.0.1:43123",
		HiddenDefaultCandidate:   true,
		SubstratePosture:         "node_runtime",
		SelectionStrategy:        "prefer_node_runtime",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected label-only hidden managed candidate to be canonicalized, got %#v want %#v", got, want)
	}
}

func TestBrowserToolMetadataRouteHintsDetailTextCanonicalizesLabelOnlyHiddenManagedCandidate(t *testing.T) {
	got := BrowserToolMetadataRouteHintsDetailText(BrowserToolMetadataRouteHints{
		Source:   "managed_browserd",
		Endpoint: "http://127.0.0.1:43123",
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "workbench",
			Target:  "node",
		},
		HiddenDefaultCandidate: true,
		SelectionStrategy:      BrowserSubstrateSelectionPreferNodeRuntime,
		SubstrateReason:        "managed browser route is ready",
	})
	for _, want := range []string{
		"source=managed_browserd",
		"endpoint=http://127.0.0.1:43123",
		"backend=proxy",
		"profile=workbench",
		"target=node",
		"default_candidate=hidden_managed",
		"default_candidate_source=managed_browserd",
		"default_candidate_endpoint=http://127.0.0.1:43123",
		"default_candidate_backend=proxy",
		"default_candidate_profile=workbench",
		"default_candidate_target=node",
		"selection_strategy=prefer_node_runtime",
		`substrate_reason="managed browser route is ready"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected route hints detail text to contain %q, got %q", want, got)
		}
	}
}

func TestBrowserToolMetadataHiddenDefaultCandidateTextUsesSharedNarration(t *testing.T) {
	got := BrowserToolMetadataHiddenDefaultCandidateText(
		BrowserToolMetadataRouteHints{
			RuntimeInfo: BrowserRuntimeInfo{
				Backend: "system",
				Profile: "default",
				Target:  "host",
			},
			DefaultCandidateRoute: BrowserRuntimeInfo{
				Backend: "proxy",
				Profile: "isolated",
				Target:  "node",
			},
			HiddenDefaultCandidate: true,
			SelectionStrategy:      BrowserSubstrateSelectionPreferNodeOverLegacy,
			SubstrateReason:        "Bundled browserd bootstrap is blocked because Chromium executable is not installed.",
			RepairAvailable:        true,
		},
		BrowserToolMetadataRouteHintActions{
			DoctorAction: "browser action=doctor",
			RepairAction: "browser action=repair",
			ReadyAction:  "browser action=ready",
		},
	)
	for _, want := range []string{
		"hidden managed browserd default candidate on `proxy/isolated/node`",
		"`selection_strategy=prefer_node_over_legacy_host`",
		"Chromium executable is not installed",
		"prefer `browser action=doctor`",
		"run `browser action=repair` if browserd bootstrap is blocked",
		"then use `browser action=ready` before falling back to host-only browser flows",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected hidden default candidate text to contain %q, got %q", want, got)
		}
	}
	if !strings.HasPrefix(got, "Browser route note:") || !strings.HasSuffix(got, ".") {
		t.Fatalf("expected hidden default candidate text to be a full sentence, got %q", got)
	}
}

func TestBrowserToolMetadataHiddenDefaultCandidateTextOmitsRepairWhenUnavailable(t *testing.T) {
	got := BrowserToolMetadataHiddenDefaultCandidateText(
		BrowserToolMetadataRouteHints{
			RuntimeInfo: BrowserRuntimeInfo{
				Backend: "proxy",
				Profile: "workbench",
				Target:  "node",
			},
			HiddenDefaultCandidate: true,
			SelectionStrategy:      BrowserSubstrateSelectionPreferNodeRuntime,
		},
		BrowserToolMetadataRouteHintActions{
			DoctorAction: "browser_runtime action=doctor",
			RepairAction: "browser_runtime action=repair",
			ReadyAction:  "browser_runtime action=prepare",
		},
	)
	if strings.Contains(got, "browser_runtime action=repair") {
		t.Fatalf("expected repair action to be omitted when repair is unavailable, got %q", got)
	}
	if !strings.Contains(got, "browser_runtime action=doctor") ||
		!strings.Contains(got, "browser_runtime action=prepare") {
		t.Fatalf("expected doctor/ready actions to remain visible, got %q", got)
	}
}

func TestBrowserToolMetadataRouteHintsDisplayTextPrefersHiddenManagedNarration(t *testing.T) {
	got := BrowserToolMetadataRouteHintsDisplayText(BrowserToolMetadataRouteHints{
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "system",
			Profile: "default",
			Target:  "host",
		},
		DefaultCandidateRoute: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		HiddenDefaultCandidate: true,
		SelectionStrategy:      BrowserSubstrateSelectionPreferNodeOverLegacy,
		SubstrateReason:        "Bundled browserd bootstrap is blocked because Chromium executable is not installed.",
	})
	if !strings.Contains(got, "hidden managed browserd default candidate on `proxy/isolated/node`") ||
		!strings.Contains(got, "`selection_strategy=prefer_node_over_legacy_host`") ||
		!strings.Contains(got, "Chromium executable is not installed") {
		t.Fatalf("expected shared display text to use hidden managed narration, got %q", got)
	}
	if strings.Contains(got, "browser action=") {
		t.Fatalf("expected shared display text to omit action suggestions, got %q", got)
	}
}

func TestBrowserToolMetadataRouteHintsDisplayTextFallsBackToDetailFields(t *testing.T) {
	got := BrowserToolMetadataRouteHintsDisplayText(BrowserToolMetadataRouteHints{
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "workbench",
			Target:  "node",
		},
		Source:            "managed_browserd",
		SelectionStrategy: BrowserSubstrateSelectionPreferNodeRuntime,
	})
	for _, want := range []string{
		"source=managed_browserd",
		"backend=proxy",
		"profile=workbench",
		"target=node",
		"selection_strategy=prefer_node_runtime",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected shared display text to preserve detail fields %q, got %q", want, got)
		}
	}
}

func TestBrowserSubstrateMetadataRouteHintsSurfaceHiddenManagedCandidateWhenVisibleDefaultRemainsLegacyHost(t *testing.T) {
	hints := browserSubstrateMetadataRouteHints(browserDefaultRuntimePreview{
		VisibleDefaultRoute: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		SubstrateSummary: BrowserWorkbenchSubstrateSummary{
			DefaultRoute:             DefaultBrowserRuntimeInfo(),
			SubstrateSource:          "legacy_host",
			SubstrateEndpoint:        "http://127.0.0.1:42123",
			DefaultCandidateRoute:    BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			DefaultCandidateSource:   "managed_browserd",
			DefaultCandidateEndpoint: "http://127.0.0.1:43123",
			SelectionStrategy:        BrowserSubstrateSelectionPreferNodeOverLegacy,
			SubstratePosture:         BrowserSubstrateNodeRuntime,
			SubstrateStatus:          "ok",
			SubstrateReason:          "managed browser route is configured but not the default yet",
		},
	})
	if !hints.HiddenDefaultCandidate {
		t.Fatalf("expected substrate metadata route hints to keep hidden managed candidate flag, got %#v", hints)
	}
	if !reflect.DeepEqual(hints.DefaultCandidateRoute, BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}) {
		t.Fatalf("expected substrate metadata route hints to expose managed default candidate route, got %#v", hints)
	}
	if hints.Source != "legacy_host" || hints.Endpoint != "http://127.0.0.1:42123" {
		t.Fatalf("expected substrate metadata route hints to preserve current substrate provenance before canonicalization, got %#v", hints)
	}
	if hints.DefaultCandidateSource != "managed_browserd" || hints.DefaultCandidateEndpoint != "http://127.0.0.1:43123" {
		t.Fatalf("expected substrate metadata route hints to preserve managed default candidate provenance, got %#v", hints)
	}
	if canonical := hints.Canonicalized(); canonical.Source != "managed_browserd" || canonical.Endpoint != "http://127.0.0.1:43123" {
		t.Fatalf("expected canonicalized substrate metadata route hints to promote managed candidate provenance, got %#v", canonical)
	}
}
