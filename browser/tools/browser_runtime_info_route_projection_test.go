package tools

import (
	"context"
	"reflect"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestBrowserRuntimeProjectActionDispatchRouteResultIncludesRouteResolutionAndRoutes(t *testing.T) {
	reg := llmxtools.NewRegistry()
	ctx, _, ok := newBrowserRegistrationContext(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			capabilities:       fullBrowserCapabilities(),
		},
		EnabledTools: []string{"browser_runtime"},
	})
	if !ok {
		t.Fatalf("expected browser registration context")
	}
	payload := browserRuntimePayload{
		Action:                 "status",
		Status:                 "ok",
		RequestedProfile:       "workbench",
		RequestedRuntimeTarget: "node",
		DefaultRoute:           browserRuntimePayloadDefaultRouteDescriptor(ctx),
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	projection := browserRuntimeProjectActionDispatchRouteResult(ctx, payload, browserRuntimeActionDispatchResultPostProcess{
		ConfiguredInfo:         BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		ResolutionDefaultRoute: browserRegistrationDefaultRuntimeInfo(ctx),
		IncludeRoutes:          true,
	})

	if projection.RouteResolution == nil {
		t.Fatalf("expected route-result projection to include route resolution")
	}
	if projection.RouteResolution.ProfileSource != "explicit_request" || projection.RouteResolution.RuntimeTargetSource != "explicit_request" {
		t.Fatalf("unexpected route-result projection resolution: %#v", projection.RouteResolution)
	}
	if !browserStringSliceContains(projection.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected route-result projection to backfill configured profiles, got %#v", projection.ConfiguredProfiles)
	}
	if !projection.ApplyRoutes || len(projection.Routes) == 0 {
		t.Fatalf("expected route-result projection to include route matrix when requested, got %#v", projection)
	}
	if projection.HideSelectedRoute {
		t.Fatalf("expected route-result projection to preserve selected route for managed lane, got %#v", projection)
	}
}

func TestBrowserRuntimeFinalizeActionDispatchPayloadIncludesRouteResolutionAndRoutes(t *testing.T) {
	reg := llmxtools.NewRegistry()
	ctx, _, ok := newBrowserRegistrationContext(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			capabilities:       fullBrowserCapabilities(),
		},
		EnabledTools: []string{"browser_runtime"},
	})
	if !ok {
		t.Fatalf("expected browser registration context")
	}
	payload := browserRuntimePayload{
		Action:                 "status",
		Status:                 "ok",
		RequestedProfile:       "workbench",
		RequestedRuntimeTarget: "node",
		DefaultRoute:           browserRuntimePayloadDefaultRouteDescriptor(ctx),
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	browserRuntimeFinalizeActionDispatchPayload(ctx, &payload, browserRuntimeActionDispatchResultPostProcess{
		ConfiguredInfo:         BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		ResolutionDefaultRoute: browserRegistrationDefaultRuntimeInfo(ctx),
		IncludeRoutes:          true,
	})

	if payload.RouteResolution == nil {
		t.Fatalf("expected finalized payload to include route resolution")
	}
	if payload.RouteResolution.ProfileSource != "explicit_request" || payload.RouteResolution.RuntimeTargetSource != "explicit_request" {
		t.Fatalf("unexpected finalized route resolution: %#v", payload.RouteResolution)
	}
	if payload.SelectedRoute == nil || payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected finalized payload to preserve selected route, got %#v", payload.SelectedRoute)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected finalized payload to backfill configured profiles from selected route, got %#v", payload.ConfiguredProfiles)
	}
	if len(payload.Routes) == 0 {
		t.Fatalf("expected finalized payload to include route matrix when include_routes is requested")
	}
}

func TestBrowserRuntimeApplyActionDispatchRouteResultProjectionSkipsConfiguredProfilesWithoutSelectedRoute(t *testing.T) {
	payload := browserRuntimePayload{}

	browserRuntimeApplyActionDispatchRouteResultProjection(&payload, browserRuntimeActionDispatchRouteResultProjection{
		ConfiguredProfiles: []string{"workbench"},
	})

	if len(payload.ConfiguredProfiles) != 0 {
		t.Fatalf("expected route-result projection to avoid backfilling configured profiles without a selected route, got %#v", payload.ConfiguredProfiles)
	}
}

func TestBrowserRuntimeProjectActionDispatchRouteResultUsesLogicalDefaultRouteForResolutionWhenTopLevelDefaultHidden(t *testing.T) {
	payload := browserRuntimePayload{
		Action:       "sessions",
		Status:       "ok",
		DefaultRoute: browserRuntimeRouteDescriptor{},
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "system",
			Profile:       "default",
			RuntimeTarget: "host",
		},
		SessionTargetSelection: &browserRuntimeSessionTargetSelection{
			ID:            "host-current",
			RuntimeTarget: "host",
			Source:        "tracked_active_tab",
		},
	}

	projection := browserRuntimeProjectActionDispatchRouteResult(browserRegistrationContext{}, payload, browserRuntimeActionDispatchResultPostProcess{
		ConfiguredInfo:                BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		ResolutionDefaultRoute:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		HiddenImplicitHostDefaultBase: true,
	})

	if projection.RouteResolution == nil {
		t.Fatalf("expected route-result projection to retain route resolution when target selection is present, got %#v", projection)
	}
	if projection.RouteResolution.ProfileSource != "default_route" ||
		projection.RouteResolution.RuntimeTargetSource != "default_route" ||
		projection.RouteResolution.TargetSource != "tracked_active_tab" {
		t.Fatalf("expected route-result projection to use logical default route for resolution, got %#v", projection.RouteResolution)
	}
	if projection.HideSelectedRoute {
		t.Fatalf("expected route-result projection not to hide selected route when target selection is present, got %#v", projection)
	}
}

func TestBrowserRuntimeProjectActionDispatchRouteResultUsesDoctorRouteManagedCandidateForResolution(t *testing.T) {
	payload := browserRuntimePayload{
		Action:       "status",
		Status:       "ok",
		DefaultRoute: browserRuntimeRouteDescriptor{},
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		ConfiguredTargets: []string{"host", "node"},
	}

	diagnosticsPreview := browserRuntimeDiagnosticsPreview{
		Registration: browserDefaultRuntimePreview{
			SubstrateAssessment: browserDefaultSubstrateAssessment{
				HostRuntime:    DefaultBrowserRuntimeInfo(),
				DefaultRuntime: DefaultBrowserRuntimeInfo(),
				HostRoute: browserConcreteRouteAssessment{
					Configured:     true,
					RouteAvailable: true,
					Route: browserResolvedExecutionRoute{
						RuntimeInfo: DefaultBrowserRuntimeInfo(),
					},
				},
				NodeRoute: browserDefaultPromotionRouteAssessment{
					Configured:     true,
					RouteAvailable: true,
					Ready:          true,
					Route: browserResolvedExecutionRoute{
						RuntimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
					},
				},
			},
			SubstrateSummary: BrowserWorkbenchSubstrateSummary{
				DefaultRoute:       BrowserRuntimeInfo{},
				HostRoute:          DefaultBrowserRuntimeInfo(),
				HostRouteAvailable: true,
				ConfiguredTargets:  []string{"host", "node"},
				SelectionStrategy:  BrowserSubstrateSelectionPreferNodeOverLegacy,
				SelectionReason:    "Managed browser route is configured, but the default route is still hidden behind the implicit legacy host fallback.",
				NodeConfigured:     true,
				NodeRouteAvailable: true,
				NodePromotionReady: true,
			},
			LogicalDefaultRoute:           DefaultBrowserRuntimeInfo(),
			VisibleDefaultRoute:           BrowserRuntimeInfo{},
			HiddenImplicitHostDefaultBase: true,
		},
		DefaultRoute:      DefaultBrowserRuntimeInfo(),
		ConfiguredTargets: []string{"host", "node"},
	}

	projection := browserRuntimeProjectActionDispatchRouteResult(browserRegistrationContext{}, payload, browserRuntimeActionDispatchResultPostProcess{
		ConfiguredInfo:                DefaultBrowserRuntimeInfo(),
		ResolutionDefaultRoute:        DefaultBrowserRuntimeInfo(),
		HiddenImplicitHostDefaultBase: true,
		DiagnosticsPreview:            diagnosticsPreview,
		UseDiagnosticsPreview:         true,
	})

	if projection.RouteResolution == nil {
		t.Fatalf("expected route-result projection to retain route resolution for managed selected route, got %#v", projection)
	}
	if projection.RouteResolution.ProfileSource != "default_route" ||
		projection.RouteResolution.RuntimeTargetSource != "default_route" {
		t.Fatalf("expected route-result projection to use doctor-route managed candidate as resolution default, got %#v", projection.RouteResolution)
	}
	if projection.HideSelectedRoute {
		t.Fatalf("expected route-result projection not to hide managed selected route, got %#v", projection)
	}
}

func TestBrowserRuntimeProjectActionDispatchRouteResultUsesProvidedDiagnosticsPreviewForRoutes(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
		DefaultRoute: browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	projection := browserRuntimeProjectActionDispatchRouteResult(browserRegistrationContext{}, payload, browserRuntimeActionDispatchResultPostProcess{
		ConfiguredInfo:         BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		ResolutionDefaultRoute: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		DiagnosticsPreview: browserRuntimeDiagnosticsPreview{
			Registration: browserDefaultRuntimePreview{
				SubstrateAssessment: browserDefaultSubstrateAssessment{
					NodeRoute: browserDefaultPromotionRouteAssessment{
						Configured:     true,
						RouteAvailable: true,
						Ready:          true,
						Route: browserResolvedExecutionRoute{
							RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
							Capabilities: fullBrowserCapabilities(),
						},
					},
					DefaultRuntime: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
					DefaultConcreteRoute: browserConcreteRouteAssessment{
						Configured:     true,
						RouteAvailable: true,
						Route: browserResolvedExecutionRoute{
							RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
							Capabilities: fullBrowserCapabilities(),
						},
					},
				},
				SubstrateSummary: BrowserWorkbenchSubstrateSummary{
					DefaultRoute:       BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
					ConfiguredTargets:  []string{"node"},
					SelectionStrategy:  BrowserSubstrateSelectionPreferNodeOverLegacy,
					SelectionReason:    "provided preview",
					NodeConfigured:     true,
					NodeRouteAvailable: true,
					NodePromotionReady: true,
				},
			},
			DefaultRoute:      BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			ConfiguredTargets: []string{"node"},
		},
		UseDiagnosticsPreview: true,
		IncludeRoutes:         true,
	})

	foundDefaultNode := false
	for _, route := range projection.Routes {
		if route.Status == "default" &&
			route.Backend == "proxy" &&
			route.Profile == "workbench" &&
			route.RuntimeTarget == "node" {
			foundDefaultNode = true
		}
	}
	if !foundDefaultNode {
		t.Fatalf("expected route-result projection to build routes from provided diagnostics preview, got %#v", projection.Routes)
	}
}

func TestBrowserRuntimeProjectActionDispatchRouteResultHidesImplicitLegacyHostRouteByPreviewFlag(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
		DefaultRoute: browserRuntimeRouteDescriptor{
			Backend:       "system",
			Profile:       "default",
			RuntimeTarget: "host",
		},
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "system",
			Profile:       "default",
			RuntimeTarget: "host",
		},
	}

	projection := browserRuntimeProjectActionDispatchRouteResult(browserRegistrationContext{}, payload, browserRuntimeActionDispatchResultPostProcess{
		ConfiguredInfo:                BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		ResolutionDefaultRoute:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		HiddenImplicitHostDefaultBase: true,
	})

	if projection.RouteResolution != nil {
		t.Fatalf("expected route-result projection to hide implicit legacy host route resolution when preview marks the default as hidden, got %#v", projection.RouteResolution)
	}
	if !projection.HideSelectedRoute {
		t.Fatalf("expected route-result projection to hide implicit legacy host selected route when preview marks the default as hidden, got %#v", projection)
	}
}

func TestBrowserRuntimeFinalizeActionDispatchPayloadHidesImplicitLegacyHostRoute(t *testing.T) {
	payload := browserRuntimePayload{
		Action:               "sessions",
		Status:               "ok",
		CoordinationState:    "stale_state",
		CoordinationDecision: "stale_decision",
		CoordinationReady:    true,
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "system",
			Profile:       "default",
			RuntimeTarget: "host",
		},
		SessionProfileSelection: &browserRuntimeSessionProfileSelection{
			Profile:       "default",
			RuntimeTarget: "host",
			Source:        "select_profile",
		},
		SessionTargetSelection: &browserRuntimeSessionTargetSelection{
			ID:            "host-current",
			RuntimeTarget: "host",
			Source:        "tracked_active_tab",
		},
		SessionBinding: &browserRuntimeSessionBinding{
			SelectedBrowserBackend:       "system",
			SelectedBrowserApp:           "Safari",
			SelectedBrowserProfile:       "default",
			SelectedBrowserProfileSource: "select_profile",
			CurrentTargetID:              "host-current",
			SelectedBrowserTargetID:      "host-current",
			SelectedBrowserTarget:        "host-current",
			SelectedBrowserTargetSource:  "tracked_active_tab",
			BrowserProfileCount:          1,
			ActiveBrowserProfile:         "default",
			BrowserProfileStatusCounts:   map[string]int{"running": 1},
			BrowserProfiles: []browserRuntimeProfileState{{
				Profile:       "default",
				Backend:       "system",
				RuntimeTarget: "host",
				Status:        "running",
				Running:       true,
			}},
			SessionHealthState:                     "ready",
			SessionHealthReason:                    "healthy",
			SessionHealthRecoveryAction:            "refresh",
			SessionHealthReconnectHint:             "wait_for_restart",
			SessionHealthDisconnectCount:           2,
			SessionHealthCooldownRemainingMs:       900,
			SessionHealthLastRestartResult:         "cooldown",
			SessionHealthRecommendedBackoffMs:      900,
			SessionHealthResolverBlockedBy:         "multiple_candidates_filtered",
			SessionHealthResolverAmbiguityClass:    "filtered_residual",
			SessionHealthResolverCandidateKind:     "label",
			SessionHealthResolverStrength:          "medium",
			SessionHealthResolverRetryDisposition:  "manual_only",
			SessionHealthResolverManualRetryHint:   "add_ordinal",
			SessionHealthResolverNextStepAlias:     "snapshot",
			SessionHealthResolverSpecificityFields: []string{"tag", "type"},
			Coordination: &browserRuntimeCoordination{
				State: "ready",
			},
		},
	}

	browserRuntimeFinalizeActionDispatchPayload(browserRegistrationContext{}, &payload, browserRuntimeActionDispatchResultPostProcess{
		ConfiguredInfo:                BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		ResolutionDefaultRoute:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		HiddenImplicitHostDefaultBase: true,
	})

	if payload.RouteResolution != nil {
		t.Fatalf("expected finalized implicit host payload to hide top-level route resolution, got %#v", payload.RouteResolution)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected finalized implicit host payload to hide top-level selected route, got %#v", payload.SelectedRoute)
	}
	if payload.SessionProfileSelection != nil {
		t.Fatalf("expected finalized implicit host payload to hide top-level profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection != nil {
		t.Fatalf("expected finalized implicit host payload to hide top-level target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding == nil ||
		payload.SessionBinding.SelectedBrowserBackend != "" ||
		payload.SessionBinding.SelectedBrowserApp != "" ||
		payload.SessionBinding.SelectedBrowserProfile != "" ||
		payload.SessionBinding.SelectedBrowserProfileSource != "" ||
		payload.SessionBinding.CurrentTargetID != "" ||
		payload.SessionBinding.SelectedBrowserTargetID != "" ||
		payload.SessionBinding.SelectedBrowserTarget != "" ||
		payload.SessionBinding.SelectedBrowserTargetSource != "" ||
		payload.SessionBinding.BrowserProfileCount != 0 ||
		payload.SessionBinding.ActiveBrowserProfile != "" ||
		len(payload.SessionBinding.BrowserProfileStatusCounts) != 0 ||
		len(payload.SessionBinding.BrowserProfiles) != 0 ||
		payload.SessionBinding.SessionHealthState != "" ||
		payload.SessionBinding.SessionHealthReason != "" ||
		payload.SessionBinding.SessionHealthRecoveryAction != "" ||
		payload.SessionBinding.SessionHealthReconnectHint != "" ||
		payload.SessionBinding.SessionHealthDisconnectCount != 0 ||
		payload.SessionBinding.SessionHealthCooldownRemainingMs != 0 ||
		payload.SessionBinding.SessionHealthLastRestartResult != "" ||
		payload.SessionBinding.SessionHealthRecommendedBackoffMs != 0 ||
		payload.SessionBinding.SessionHealthResolverBlockedBy != "" ||
		payload.SessionBinding.SessionHealthResolverAmbiguityClass != "" ||
		payload.SessionBinding.SessionHealthResolverCandidateKind != "" ||
		payload.SessionBinding.SessionHealthResolverStrength != "" ||
		payload.SessionBinding.SessionHealthResolverRetryDisposition != "" ||
		payload.SessionBinding.SessionHealthResolverManualRetryHint != "" ||
		payload.SessionBinding.SessionHealthResolverNextStepAlias != "" ||
		len(payload.SessionBinding.SessionHealthResolverSpecificityFields) != 0 {
		t.Fatalf("expected finalized implicit host payload to scrub top-level binding selection, got %#v", payload.SessionBinding)
	}
	if payload.SessionBinding != nil && payload.SessionBinding.Coordination != nil {
		t.Fatalf("expected finalized implicit host payload to hide coordination, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.ResolverBlockedBy != "" ||
		payload.ResolverAmbiguityClass != "" ||
		payload.ResolverCandidateKind != "" ||
		payload.ResolverCandidateStrength != "" ||
		payload.ResolverRetryDisposition != "" ||
		payload.ResolverManualRetryHint != "" ||
		payload.ResolverNextStepAlias != "" ||
		len(payload.ResolverSpecificityFields) != 0 ||
		payload.Summary != nil ||
		payload.ResolverExplanation != nil ||
		payload.DiagnosticsExplanation != nil ||
		payload.WorkbenchExplanation != nil ||
		payload.WorkbenchDiagnostics != nil {
		t.Fatalf("expected finalized implicit host payload to scrub top-level resolver guidance, got %#v", payload)
	}
	if payload.CoordinationState != "" || payload.CoordinationDecision != "" || payload.CoordinationReady {
		t.Fatalf("expected finalized implicit host payload to clear stale top-level coordination summary, got %#v", payload)
	}
}

func TestBrowserRuntimeFinalizeActionDispatchPayloadCanonicalizesHiddenManagedSelectionSummaryFromDiagnosticsPreview(t *testing.T) {
	payload := browserRuntimePayload{
		Action:                     "sessions",
		Status:                     "ok",
		SubstrateSelectionStrategy: BrowserSubstrateSelectionLegacyHostDefault,
		SubstrateSelectionReason:   BrowserSubstrateSelectionReason(DefaultBrowserRuntimeInfo(), DefaultBrowserRuntimeInfo()),
	}
	diagnosticsPreview := browserRuntimeDiagnosticsPreview{
		Registration: browserDefaultRuntimePreview{
			SubstrateAssessment: browserDefaultSubstrateAssessment{
				HostRuntime: DefaultBrowserRuntimeInfo(),
				HostRoute: browserConcreteRouteAssessment{
					Configured:     true,
					RouteAvailable: true,
					Route: browserResolvedExecutionRoute{
						RuntimeInfo:  DefaultBrowserRuntimeInfo(),
						Capabilities: defaultBrowserCapabilities(),
					},
				},
				DefaultRuntime: DefaultBrowserRuntimeInfo(),
				NodeRoute: browserDefaultPromotionRouteAssessment{
					Configured:     true,
					RouteAvailable: true,
					Route: browserResolvedExecutionRoute{
						RuntimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
						Capabilities: BrowserCapabilities{
							Click: true,
						},
					},
				},
			},
			SubstrateSummary: BrowserWorkbenchSubstrateSummary{
				HostRoute:          DefaultBrowserRuntimeInfo(),
				HostRouteAvailable: true,
				ConfiguredTargets:  []string{"host", "node"},
				SelectionStrategy:  BrowserSubstrateSelectionLegacyHostDefault,
				SelectionReason:    BrowserSubstrateSelectionReason(DefaultBrowserRuntimeInfo(), DefaultBrowserRuntimeInfo()),
				NodeConfigured:     true,
				NodeRouteAvailable: true,
			},
			LogicalDefaultRoute:           DefaultBrowserRuntimeInfo(),
			VisibleDefaultRoute:           BrowserRuntimeInfo{},
			HiddenImplicitHostDefaultBase: true,
		},
		DefaultRoute:      DefaultBrowserRuntimeInfo(),
		ConfiguredTargets: []string{"host", "node"},
	}

	browserRuntimeFinalizeActionDispatchPayload(browserRegistrationContext{}, &payload, browserRuntimeActionDispatchResultPostProcess{
		ConfiguredInfo:                DefaultBrowserRuntimeInfo(),
		ResolutionDefaultRoute:        DefaultBrowserRuntimeInfo(),
		HiddenImplicitHostDefaultBase: true,
		DiagnosticsPreview:            diagnosticsPreview,
		UseDiagnosticsPreview:         true,
	})

	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected finalize to canonicalize hidden managed selection strategy from diagnostics preview, got %#v", payload)
	}
	if payload.SubstrateSelectionReason != "Managed browser route is configured, but the default route is still hidden behind the implicit legacy host fallback." {
		t.Fatalf("expected finalize to canonicalize hidden managed selection reason from diagnostics preview, got %#v", payload)
	}
}

func TestBrowserRuntimeFinalizeActionDispatchPayloadKeepsCapabilityPartialHiddenManagedReasonFromDiagnosticsPreview(t *testing.T) {
	const specificFailure = "node runtime route is configured but not the default because it does not yet advertise the required default browser capabilities; it remains available via `runtime_target=node`"
	payload := browserRuntimePayload{
		Action:                     "sessions",
		Status:                     "ok",
		SubstrateSelectionStrategy: BrowserSubstrateSelectionLegacyHostDefault,
		SubstrateSelectionReason:   specificFailure,
	}
	diagnosticsPreview := browserRuntimeDiagnosticsPreview{
		Registration: browserDefaultRuntimePreview{
			SubstrateAssessment: browserDefaultSubstrateAssessment{
				HostRuntime: DefaultBrowserRuntimeInfo(),
				HostRoute: browserConcreteRouteAssessment{
					Configured:     true,
					RouteAvailable: true,
					Route: browserResolvedExecutionRoute{
						RuntimeInfo:  DefaultBrowserRuntimeInfo(),
						Capabilities: defaultBrowserCapabilities(),
					},
				},
				DefaultRuntime: DefaultBrowserRuntimeInfo(),
				NodeRoute: browserDefaultPromotionRouteAssessment{
					Configured:     true,
					RouteAvailable: true,
					Route: browserResolvedExecutionRoute{
						RuntimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
						Capabilities: BrowserCapabilities{
							Click: true,
						},
					},
					FailureReason: specificFailure,
				},
			},
			SubstrateSummary: BrowserWorkbenchSubstrateSummary{
				HostRoute:          DefaultBrowserRuntimeInfo(),
				HostRouteAvailable: true,
				ConfiguredTargets:  []string{"host", "node"},
				SelectionStrategy:  BrowserSubstrateSelectionLegacyHostDefault,
				SelectionReason:    specificFailure,
				NodeConfigured:     true,
				NodeRouteAvailable: true,
			},
			LogicalDefaultRoute:           DefaultBrowserRuntimeInfo(),
			VisibleDefaultRoute:           BrowserRuntimeInfo{},
			HiddenImplicitHostDefaultBase: true,
		},
		DefaultRoute:      DefaultBrowserRuntimeInfo(),
		ConfiguredTargets: []string{"host", "node"},
	}

	browserRuntimeFinalizeActionDispatchPayload(browserRegistrationContext{}, &payload, browserRuntimeActionDispatchResultPostProcess{
		ConfiguredInfo:                DefaultBrowserRuntimeInfo(),
		ResolutionDefaultRoute:        DefaultBrowserRuntimeInfo(),
		HiddenImplicitHostDefaultBase: true,
		DiagnosticsPreview:            diagnosticsPreview,
		UseDiagnosticsPreview:         true,
	})

	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionLegacyHostDefault {
		t.Fatalf("expected finalize to keep explicit-lane strategy for capability-partial hidden managed reason, got %#v", payload)
	}
	if payload.SubstrateSelectionReason != specificFailure {
		t.Fatalf("expected finalize to keep capability-partial hidden managed reason, got %#v", payload)
	}
}

func TestBrowserRuntimeFinalizeActionDispatchPayloadPrefersManagedCandidateStrategyForBootstrapSpecificHiddenManagedFailureFromDiagnosticsPreview(t *testing.T) {
	const bootstrapBlockedSummary = "Bundled browserd bootstrap is blocked because `node` is not available in PATH."
	const rawFailure = `browser proxy managed_route_unavailable target=node endpoint=http://127.0.0.1:1: browserdaemon: bundled browser bootstrap requires node in PATH: exec: "node": executable file not found in $PATH`

	payload := browserRuntimePayload{
		Action:                     "sessions",
		Status:                     "ok",
		SubstrateSelectionStrategy: BrowserSubstrateSelectionLegacyHostDefault,
		SubstrateSelectionReason:   bootstrapBlockedSummary,
	}
	diagnosticsPreview := browserRuntimeDiagnosticsPreview{
		Registration: browserDefaultRuntimePreview{
			SubstrateAssessment: browserDefaultSubstrateAssessment{
				HostRuntime: DefaultBrowserRuntimeInfo(),
				HostRoute: browserConcreteRouteAssessment{
					Configured:     true,
					RouteAvailable: true,
					Route: browserResolvedExecutionRoute{
						RuntimeInfo:  DefaultBrowserRuntimeInfo(),
						Capabilities: defaultBrowserCapabilities(),
					},
				},
				DefaultRuntime: DefaultBrowserRuntimeInfo(),
				NodeRoute: browserDefaultPromotionRouteAssessment{
					Configured:    true,
					FailureReason: rawFailure,
					FailureNote:   rawFailure,
				},
			},
			SubstrateSummary: BrowserWorkbenchSubstrateSummary{
				HostRoute:                 DefaultBrowserRuntimeInfo(),
				HostRouteAvailable:        true,
				ConfiguredTargets:         []string{"host", "node"},
				SelectionStrategy:         BrowserSubstrateSelectionLegacyHostDefault,
				SelectionReason:           bootstrapBlockedSummary,
				NodeConfigured:            true,
				NodePromotionReady:        false,
				NodePromotionFailureCause: rawFailure,
			},
			LogicalDefaultRoute:           DefaultBrowserRuntimeInfo(),
			VisibleDefaultRoute:           BrowserRuntimeInfo{},
			HiddenImplicitHostDefaultBase: true,
		},
		DefaultRoute:      DefaultBrowserRuntimeInfo(),
		ConfiguredTargets: []string{"host", "node"},
	}

	browserRuntimeFinalizeActionDispatchPayload(browserRegistrationContext{}, &payload, browserRuntimeActionDispatchResultPostProcess{
		ConfiguredInfo:                DefaultBrowserRuntimeInfo(),
		ResolutionDefaultRoute:        DefaultBrowserRuntimeInfo(),
		HiddenImplicitHostDefaultBase: true,
		DiagnosticsPreview:            diagnosticsPreview,
		UseDiagnosticsPreview:         true,
	})

	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected finalize to prefer managed candidate strategy for bootstrap-specific hidden managed failure, got %#v", payload)
	}
	if payload.SubstrateSelectionReason != bootstrapBlockedSummary {
		t.Fatalf("expected finalize to keep bootstrap-specific hidden managed failure reason, got %#v", payload)
	}
}

func TestBrowserRuntimeFinalizeActionSessionPayloadBuildsSelectionsAndBinding(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-action-session-finalize-binding")
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	stateRegistry.SelectBrowserProfile("runtime-action-session-finalize-binding", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})
	stateRegistry.RecordBrowserProfileState("runtime-action-session-finalize-binding", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	tracked := sessionRegistry.TrackCurrentTarget("runtime-action-session-finalize-binding", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	ctx := browserRegistrationContext{
		sessionRegistry:      sessionRegistry,
		sessionStateRegistry: stateRegistry,
	}
	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		},
	}

	browserRuntimeFinalizeActionSessionPayload(ctx, callCtx, &payload, browserRuntimeActionSessionResultPostProcess{
		Action: "status",
	})

	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected session finalize helper to populate profile selection from registry, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != tracked.ID || payload.SessionTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected session finalize helper to populate target selection from registry, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding == nil || payload.SessionBinding.CurrentTargetID != tracked.ID || payload.SessionBinding.SelectedBrowserProfile != "workbench" || payload.SessionBinding.SelectedBrowserTargetID != tracked.ID {
		t.Fatalf("expected session finalize helper to build binding from selections, got %#v", payload.SessionBinding)
	}
	if payload.ProfileStatus == nil || !payload.ProfileStatus.Selected {
		t.Fatalf("expected session finalize helper to mark current profile selected, got %#v", payload.ProfileStatus)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected session finalize helper to refresh configured profiles from binding projection, got %#v", payload.ConfiguredProfiles)
	}
	if payload.Summary != nil || payload.Display != nil || payload.Surface != nil || payload.View != nil {
		t.Fatalf("expected status session finalize helper not to synthesize action-success shells, got summary=%#v display=%#v surface=%#v view=%#v", payload.Summary, payload.Display, payload.Surface, payload.View)
	}
}

func TestBrowserRuntimeFinalizeActionSessionPayloadMergesCurrentProfileIntoBindingProjection(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-action-session-finalize-merge-current")
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	stateRegistry.SelectBrowserProfile("runtime-action-session-finalize-merge-current", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})
	sessionRegistry.TrackCurrentTarget("runtime-action-session-finalize-merge-current", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	ctx := browserRegistrationContext{
		sessionRegistry:      sessionRegistry,
		sessionStateRegistry: stateRegistry,
	}
	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "running",
			Running:       true,
			Connected:     true,
			Note:          "fresh current status",
		},
	}

	browserRuntimeFinalizeActionSessionPayload(ctx, callCtx, &payload, browserRuntimeActionSessionResultPostProcess{
		Action: "status",
	})

	if payload.SessionBinding == nil || len(payload.SessionBinding.BrowserProfiles) != 1 {
		t.Fatalf("expected session finalize helper to build binding profiles from shared projection, got %#v", payload.SessionBinding)
	}
	got := payload.SessionBinding.BrowserProfiles[0]
	if got.Profile != "workbench" || got.Status != "running" || got.Note != "fresh current status" || !got.Selected {
		t.Fatalf("expected session finalize helper to merge current profile into shared binding projection, got %#v", got)
	}
}

func TestBrowserRuntimeFinalizeActionSessionPayloadBuildsBindingWithoutSelectedRouteFromBindingEvaluation(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-action-session-finalize-route-less-binding")
	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
	}

	browserRuntimeFinalizeActionSessionPayload(browserRegistrationContext{}, callCtx, &payload, browserRuntimeActionSessionResultPostProcess{
		Action: "status",
		BindingEvaluation: &agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
			Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
				CurrentTargetID: "tab-2",
				SelectedProfileSelection: &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "remember_profile",
				},
				SelectedTargetSelection: &agentxbrowserruntime.BrowserSessionTargetSelection{
					ID:            "tab-2",
					TabIndex:      2,
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "tracked_active_tab",
				},
				Profiles: []agentxbrowserruntime.SharedSessionBrowserProfileState{{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Status:        "running",
					Running:       true,
					Connected:     true,
					Note:          "cached current status",
				}},
				Summary: agentxbrowserruntime.SharedSessionBrowserBindingSummary{
					CurrentTargetID:      "tab-2",
					RouteTargetCount:     1,
					BrowserProfileCount:  1,
					ActiveBrowserProfile: "workbench",
				},
			},
		},
	})

	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.Source != "remember_profile" {
		t.Fatalf("expected route-less session finalize helper to preserve provided profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != "tab-2" || payload.SessionTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected route-less session finalize helper to preserve provided target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding == nil || payload.SessionBinding.CurrentTargetID != "tab-2" || payload.SessionBinding.SelectedBrowserBackend != "proxy" || payload.SessionBinding.SelectedBrowserProfile != "workbench" || payload.SessionBinding.SelectedBrowserTarget != "node" || payload.SessionBinding.SelectedBrowserTargetID != "tab-2" {
		t.Fatalf("expected route-less session finalize helper to build binding from provided evaluation, got %#v", payload.SessionBinding)
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Selected {
		t.Fatalf("expected route-less session finalize helper to refresh profile status from binding evaluation, got %#v", payload.ProfileStatus)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") || payload.DefaultProfile != "workbench" {
		t.Fatalf("expected route-less session finalize helper to refresh configured/default profiles from binding evaluation, got payload=%#v configured=%#v", payload.DefaultProfile, payload.ConfiguredProfiles)
	}
	if payload.SelectedRoute == nil || payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected route-less session finalize helper to backfill selected route from binding evaluation, got %#v", payload.SelectedRoute)
	}
}

func TestBrowserRuntimeApplyTopLevelBindingProjectionBuildsBindingAndMarksCurrentProfile(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-top-level-binding-apply")
	payload := browserRuntimePayload{
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		},
	}

	browserRuntimeApplyTopLevelBindingProjection(callCtx, &payload, agentxbrowserruntime.SharedSessionBrowserTopLevelBindingProjection{
		ProfileSelection: &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "select_profile",
		},
		TargetSelection: &agentxbrowserruntime.BrowserSessionTargetSelection{
			ID:            "tab-2",
			TabIndex:      2,
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "tracked_active_tab",
		},
		Evaluation: agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
			Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
				CurrentTargetID: "tab-2",
				SelectedProfileSelection: &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "select_profile",
				},
				SelectedTargetSelection: &agentxbrowserruntime.BrowserSessionTargetSelection{
					ID:            "tab-2",
					TabIndex:      2,
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "tracked_active_tab",
				},
				Profiles: []agentxbrowserruntime.SharedSessionBrowserProfileState{{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Status:        "running",
					Running:       true,
					Connected:     true,
				}},
				Summary: agentxbrowserruntime.SharedSessionBrowserBindingSummary{
					CurrentTargetID:      "tab-2",
					RouteTargetCount:     1,
					BrowserProfileCount:  1,
					ActiveBrowserProfile: "workbench",
				},
			},
			Health: agentxbrowserruntime.SharedSessionBrowserHealthEvaluation{
				Summary: &agentxbrowserruntime.SharedSessionBrowserHealthSummary{
					State:                       "healthy",
					Reason:                      "resolver guidance is available",
					RecoveryAction:              "browser action=snapshot",
					ReconnectHint:               "wait_for_restart",
					DisconnectCount:             3,
					DisconnectBurstCount:        2,
					DisconnectBurstWindowMs:     30000,
					CooldownRemainingMs:         900,
					RetryBackoffRemainingMs:     450,
					RestartAttemptCount:         4,
					RestartFailureCount:         1,
					LastDisconnectUnixMilli:     111,
					LastReconnectUnixMilli:      222,
					LastRestartAttemptUnixMilli: 333,
					LastRestartResult:           "restarted",
					LastRestartError:            "transport closed",
					RecommendedBackoffMs:        1200,
					ResolverBlockedBy:           "multiple_candidates_filtered",
					AmbiguityClass:              "filtered_residual",
					CandidateKind:               "label",
					CandidateStrength:           "medium",
					RetryDisposition:            "manual_only",
					ManualRetryHint:             "add_ordinal",
					NextStepAlias:               "snapshot",
					SpecificityFields:           []string{"tag", "type"},
				},
			},
		},
	})

	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected top-level binding apply helper to populate profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != "tab-2" || payload.SessionTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected top-level binding apply helper to populate target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding == nil || payload.SessionBinding.CurrentTargetID != "tab-2" || payload.SessionBinding.SelectedBrowserTargetID != "tab-2" {
		t.Fatalf("expected top-level binding apply helper to build binding from shared evaluation, got %#v", payload.SessionBinding)
	}
	if payload.SessionHandoff == nil ||
		payload.SessionHandoff.State != agentxbrowserruntime.SharedSessionBrowserSessionHandoffStateReady ||
		payload.SessionHandoff.CurrentTarget == nil ||
		payload.SessionHandoff.CurrentTarget.ID != "tab-2" ||
		payload.SessionBinding.SessionHandoff == nil ||
		payload.SessionBinding.SessionHandoff.SelectedProfile != "workbench" {
		t.Fatalf("expected top-level binding apply helper to build session handoff, got payload=%#v binding=%#v", payload.SessionHandoff, payload.SessionBinding)
	}
	if payload.SessionBinding == nil ||
		payload.SessionBinding.SessionHealthReconnectHint != "wait_for_restart" ||
		payload.SessionBinding.SessionHealthDisconnectCount != 3 ||
		payload.SessionBinding.SessionHealthDisconnectBurstCount != 2 ||
		payload.SessionBinding.SessionHealthDisconnectBurstWindowMs != 30000 ||
		payload.SessionBinding.SessionHealthCooldownRemainingMs != 900 ||
		payload.SessionBinding.SessionHealthRetryBackoffRemainingMs != 450 ||
		payload.SessionBinding.SessionHealthRestartAttemptCount != 4 ||
		payload.SessionBinding.SessionHealthRestartFailureCount != 1 ||
		payload.SessionBinding.SessionHealthLastDisconnectUnixMilli != 111 ||
		payload.SessionBinding.SessionHealthLastReconnectUnixMilli != 222 ||
		payload.SessionBinding.SessionHealthLastRestartAttemptUnixMilli != 333 ||
		payload.SessionBinding.SessionHealthLastRestartResult != "restarted" ||
		payload.SessionBinding.SessionHealthLastRestartError != "transport closed" ||
		payload.SessionBinding.SessionHealthRecommendedBackoffMs != 1200 ||
		payload.SessionBinding.SessionHealthResolverBlockedBy != "multiple_candidates_filtered" ||
		payload.SessionBinding.SessionHealthResolverAmbiguityClass != "filtered_residual" ||
		payload.SessionBinding.SessionHealthResolverCandidateKind != "label" ||
		payload.SessionBinding.SessionHealthResolverStrength != "medium" ||
		payload.SessionBinding.SessionHealthResolverRetryDisposition != "manual_only" ||
		payload.SessionBinding.SessionHealthResolverManualRetryHint != "add_ordinal" ||
		payload.SessionBinding.SessionHealthResolverNextStepAlias != "snapshot" ||
		!reflect.DeepEqual(payload.SessionBinding.SessionHealthResolverSpecificityFields, []string{"tag", "type"}) {
		t.Fatalf("expected top-level binding apply helper to keep resolver guidance aliases, got %#v", payload.SessionBinding)
	}
	if payload.ResolverBlockedBy != "multiple_candidates_filtered" ||
		payload.ResolverAmbiguityClass != "filtered_residual" ||
		payload.ResolverCandidateKind != "label" ||
		payload.ResolverCandidateStrength != "medium" ||
		payload.ResolverRetryDisposition != "manual_only" ||
		payload.ResolverManualRetryHint != "add_ordinal" ||
		payload.ResolverNextStepAlias != "snapshot" ||
		!reflect.DeepEqual(payload.ResolverSpecificityFields, []string{"tag", "type"}) {
		t.Fatalf("expected top-level binding apply helper to mirror resolver guidance summary, got %#v", payload)
	}
	if payload.ResolverExplanation == nil ||
		payload.ResolverExplanation.State != "manual_resolution_required" ||
		payload.ResolverExplanation.SummaryCode != "label_filtered_residual" ||
		payload.ResolverExplanation.NextStepAlias != "snapshot" ||
		payload.ResolverExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("expected top-level binding apply helper to build resolver explanation summary, got %#v", payload.ResolverExplanation)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "resolver" ||
		payload.DiagnosticsExplanation.State != "manual_resolution_required" ||
		payload.DiagnosticsExplanation.SummaryCode != "label_filtered_residual" ||
		payload.DiagnosticsExplanation.NextStepAlias != "snapshot" ||
		payload.DiagnosticsExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("expected top-level binding apply helper to build diagnostics explanation summary, got %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "resolver" ||
		payload.Summary.State != "manual_resolution_required" ||
		payload.Summary.SummaryCode != "label_filtered_residual" ||
		payload.Summary.NextStepAlias != "snapshot" ||
		payload.Summary.ManualRetryHint != "add_ordinal" ||
		payload.Summary.ResolvedViaFallback {
		t.Fatalf("expected top-level binding apply helper to build summary alias, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Ready ||
		len(payload.Display.Sections) != 0 ||
		payload.Display.Category != "resolver" ||
		payload.Display.State != "manual_resolution_required" ||
		payload.Display.SummaryCode != "label_filtered_residual" ||
		payload.Display.NextStepAlias != "snapshot" ||
		payload.Display.ManualRetryHint != "add_ordinal" ||
		payload.Display.ResolvedViaFallback {
		t.Fatalf("expected top-level binding apply helper to build display alias, got %#v", payload.Display)
	}
	if payload.ProfileStatus == nil || !payload.ProfileStatus.Selected {
		t.Fatalf("expected top-level binding apply helper to mark current profile selected, got %#v", payload.ProfileStatus)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected top-level binding apply helper to refresh configured profiles from binding route, got %#v", payload.ConfiguredProfiles)
	}
}

func TestBrowserRuntimeApplyTopLevelBindingProjectionHydratesSelectionBrowserAppFromBindingSnapshot(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-top-level-binding-hydrate-browser-app")
	payload := browserRuntimePayload{
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	browserRuntimeApplyTopLevelBindingProjection(callCtx, &payload, agentxbrowserruntime.SharedSessionBrowserTopLevelBindingProjection{
		ProfileSelection: &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "select_profile",
		},
		TargetSelection: &agentxbrowserruntime.BrowserSessionTargetSelection{
			ID:            "tab-2",
			TabIndex:      2,
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "tracked_active_tab",
		},
		Evaluation: agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
			Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
				Profiles: []agentxbrowserruntime.SharedSessionBrowserProfileState{{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Status:        "running",
					Running:       true,
					Connected:     true,
				}},
			},
		},
	})

	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.BrowserApp != "Chromium" {
		t.Fatalf("expected top-level binding apply helper to hydrate profile browser app from binding snapshot, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.BrowserApp != "Chromium" {
		t.Fatalf("expected top-level binding apply helper to hydrate target browser app from binding snapshot, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding == nil || payload.SessionBinding.SelectedBrowserApp != "Chromium" {
		t.Fatalf("expected top-level binding apply helper to keep binding browser app in sync with shared projection, got %#v", payload.SessionBinding)
	}
}

func TestBrowserRuntimeApplyTopLevelBindingProjectionDerivesSelectionsFromBindingSnapshot(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-top-level-binding-derive-selections")
	payload := browserRuntimePayload{
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		},
	}

	browserRuntimeApplyTopLevelBindingProjection(callCtx, &payload, agentxbrowserruntime.SharedSessionBrowserTopLevelBindingProjection{
		Evaluation: agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
			Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
				CurrentTargetID: "tab-2",
				SelectedProfileSelection: &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "select_profile",
				},
				SelectedTargetSelection: &agentxbrowserruntime.BrowserSessionTargetSelection{
					ID:            "tab-2",
					TabIndex:      2,
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "tracked_active_tab",
				},
				Profiles: []agentxbrowserruntime.SharedSessionBrowserProfileState{{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Status:        "running",
					Running:       true,
					Connected:     true,
				}},
				Summary: agentxbrowserruntime.SharedSessionBrowserBindingSummary{
					CurrentTargetID:      "tab-2",
					RouteTargetCount:     1,
					BrowserProfileCount:  1,
					ActiveBrowserProfile: "workbench",
				},
			},
		},
	})

	if payload.SessionProfileSelection == nil ||
		payload.SessionProfileSelection.Backend != "proxy" ||
		payload.SessionProfileSelection.Profile != "workbench" ||
		payload.SessionProfileSelection.RuntimeTarget != "node" ||
		payload.SessionProfileSelection.BrowserApp != "Chromium" ||
		payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected top-level binding apply helper to derive profile selection from binding snapshot, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil ||
		payload.SessionTargetSelection.ID != "tab-2" ||
		payload.SessionTargetSelection.TabIndex != 2 ||
		payload.SessionTargetSelection.Backend != "proxy" ||
		payload.SessionTargetSelection.Profile != "workbench" ||
		payload.SessionTargetSelection.RuntimeTarget != "node" ||
		payload.SessionTargetSelection.BrowserApp != "Chromium" ||
		payload.SessionTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected top-level binding apply helper to derive target selection from binding snapshot, got %#v", payload.SessionTargetSelection)
	}
	if payload.ProfileStatus == nil || !payload.ProfileStatus.Selected {
		t.Fatalf("expected top-level binding apply helper to mark current profile selected after deriving selections, got %#v", payload.ProfileStatus)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected top-level binding apply helper to backfill default profile from binding snapshot, got %#v", payload.DefaultProfile)
	}
}

func TestBrowserRuntimeApplyTopLevelBindingProjectionWithoutSelectedRouteBuildsBindingFromEvaluation(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-top-level-binding-route-less")
	payload := browserRuntimePayload{}

	browserRuntimeApplyTopLevelBindingProjection(callCtx, &payload, agentxbrowserruntime.SharedSessionBrowserTopLevelBindingProjection{
		Evaluation: agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
			Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
				CurrentTargetID: "tab-2",
				SelectedProfileSelection: &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "select_profile",
				},
				SelectedTargetSelection: &agentxbrowserruntime.BrowserSessionTargetSelection{
					ID:            "tab-2",
					TabIndex:      2,
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "tracked_active_tab",
				},
				Profiles: []agentxbrowserruntime.SharedSessionBrowserProfileState{{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Status:        "running",
					Running:       true,
					Connected:     true,
				}},
				Summary: agentxbrowserruntime.SharedSessionBrowserBindingSummary{
					CurrentTargetID:      "tab-2",
					RouteTargetCount:     1,
					BrowserProfileCount:  1,
					ActiveBrowserProfile: "workbench",
				},
			},
		},
	})

	if payload.SessionBinding == nil || payload.SessionBinding.CurrentTargetID != "tab-2" || payload.SessionBinding.SelectedBrowserBackend != "proxy" || payload.SessionBinding.SelectedBrowserProfile != "workbench" || payload.SessionBinding.SelectedBrowserTarget != "node" || payload.SessionBinding.SelectedBrowserTargetID != "tab-2" {
		t.Fatalf("expected route-less top-level binding helper to build binding from shared evaluation, got %#v", payload.SessionBinding)
	}
	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Backend != "proxy" || payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.RuntimeTarget != "node" {
		t.Fatalf("expected route-less top-level binding helper to derive profile selection from shared evaluation, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != "tab-2" || payload.SessionTargetSelection.Backend != "proxy" || payload.SessionTargetSelection.Profile != "workbench" || payload.SessionTargetSelection.RuntimeTarget != "node" {
		t.Fatalf("expected route-less top-level binding helper to derive target selection from shared evaluation, got %#v", payload.SessionTargetSelection)
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Selected {
		t.Fatalf("expected route-less top-level binding helper to hydrate selected profile status, got %#v", payload.ProfileStatus)
	}
	if payload.DefaultProfile != "workbench" || !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected route-less top-level binding helper to refresh default/configured profiles, got default=%q configured=%#v", payload.DefaultProfile, payload.ConfiguredProfiles)
	}
	if payload.SelectedRoute == nil || payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected route-less top-level binding helper to backfill selected route from shared evaluation, got %#v", payload.SelectedRoute)
	}
}

func TestBrowserRuntimeApplyTopLevelBindingProjectionWithoutSelectedRouteKeepsImplicitHostHidden(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-top-level-binding-route-less-implicit-host")
	payload := browserRuntimePayload{}

	browserRuntimeApplyTopLevelBindingProjection(callCtx, &payload, agentxbrowserruntime.SharedSessionBrowserTopLevelBindingProjection{
		Evaluation: agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
			Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
				CurrentTargetID: "tab-host",
				SelectedProfileSelection: &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
					Backend:       "system",
					Profile:       "default",
					RuntimeTarget: "host",
					Source:        "remember_profile",
				},
				SelectedTargetSelection: &agentxbrowserruntime.BrowserSessionTargetSelection{
					ID:            "tab-host",
					Backend:       "system",
					Profile:       "default",
					RuntimeTarget: "host",
					Source:        "tracked_active_tab",
				},
			},
		},
	})

	if payload.SelectedRoute != nil {
		t.Fatalf("expected route-less top-level binding helper to keep implicit legacy host hidden, got %#v", payload.SelectedRoute)
	}
}

func TestBrowserRuntimeApplyTopLevelBindingProjectionBackfillsProfileInventoryFromBindingSnapshot(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-top-level-binding-profile-inventory")
	payload := browserRuntimePayload{
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	browserRuntimeApplyTopLevelBindingProjection(callCtx, &payload, agentxbrowserruntime.SharedSessionBrowserTopLevelBindingProjection{
		Evaluation: agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
			Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
				SelectedProfileSelection: &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Source:        "select_profile",
				},
				Profiles: []agentxbrowserruntime.SharedSessionBrowserProfileState{{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
					BrowserApp:    "Chromium",
					Status:        "running",
					Running:       true,
					Connected:     true,
				}},
				Summary: agentxbrowserruntime.SharedSessionBrowserBindingSummary{
					BrowserProfileCount:  1,
					ActiveBrowserProfile: "workbench",
				},
			},
		},
	})

	if payload.ProfileStatus == nil ||
		payload.ProfileStatus.Profile != "workbench" ||
		payload.ProfileStatus.RuntimeTarget != "node" ||
		payload.ProfileStatus.BrowserApp != "Chromium" ||
		payload.ProfileStatus.Status != "running" ||
		!payload.ProfileStatus.Running ||
		!payload.ProfileStatus.Connected ||
		!payload.ProfileStatus.Selected {
		t.Fatalf("expected top-level binding apply helper to backfill profile status from binding snapshot, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 1 ||
		payload.Profiles[0].Profile != "workbench" ||
		payload.Profiles[0].RuntimeTarget != "node" ||
		payload.Profiles[0].BrowserApp != "Chromium" ||
		payload.Profiles[0].Status != "running" {
		t.Fatalf("expected top-level binding apply helper to backfill profiles from binding snapshot, got %#v", payload.Profiles)
	}
}

func TestBrowserRuntimeTopLevelProfileInventoryProjectionFromBindingPayloadUsesPayloadProfilesForStatusFallback(t *testing.T) {
	projection := browserRuntimeTopLevelProfileInventoryProjectionFromBindingPayload(browserRuntimePayload{
		SessionBinding: &browserRuntimeSessionBinding{
			SelectedBrowserBackend:       "proxy",
			SelectedBrowserProfile:       "default",
			SelectedBrowserTarget:        "node",
			SelectedBrowserProfileSource: "select_profile",
		},
		SessionProfileSelection: &browserRuntimeSessionProfileSelection{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "select_profile",
		},
		Profiles: []browserRuntimeProfileState{
			{
				Backend:       "proxy",
				Profile:       "default",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "stopped",
			},
			{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
			},
		},
	})

	if projection == nil || !projection.ApplyProfileStatus || projection.ApplyProfileInventory {
		t.Fatalf("expected binding payload inventory projection, got %#v", projection)
	}
	if projection.ConfiguredInfo.Profile != "isolated" {
		t.Fatalf("expected payload selection overlay to drive configured profile, got %#v", projection)
	}
	if projection.ProfileStatus == nil ||
		projection.ProfileStatus.Profile != "isolated" ||
		projection.ProfileStatus.Status != "running" ||
		!projection.ProfileStatus.Running ||
		!projection.ProfileStatus.Connected {
		t.Fatalf("expected payload profiles to backfill status from shared binding inventory projection, got %#v", projection)
	}
}

func TestBrowserRuntimeSharedBindingEvaluationPreservesSessionHealthReliabilityMetadata(t *testing.T) {
	evaluation := browserRuntimeSharedBindingEvaluation(browserRuntimeSessionBinding{
		SessionHealthState:                       "cooldown_active",
		SessionHealthReason:                      "browser restart cooldown active",
		SessionHealthRecoveryAction:              "browser action=wait",
		SessionHealthReconnectHint:               "wait_for_restart",
		SessionHealthDisconnectCount:             2,
		SessionHealthDisconnectBurstCount:        1,
		SessionHealthDisconnectBurstWindowMs:     30000,
		SessionHealthCooldownRemainingMs:         900,
		SessionHealthRetryBackoffRemainingMs:     450,
		SessionHealthRestartAttemptCount:         4,
		SessionHealthRestartFailureCount:         1,
		SessionHealthLastDisconnectUnixMilli:     111,
		SessionHealthLastReconnectUnixMilli:      222,
		SessionHealthLastRestartAttemptUnixMilli: 333,
		SessionHealthLastRestartResult:           "restarted",
		SessionHealthLastRestartError:            "transport closed",
		SessionHealthRecommendedBackoffMs:        1200,
	}, nil)

	if evaluation.Health.Summary == nil ||
		evaluation.Health.Summary.State != "cooldown_active" ||
		evaluation.Health.Summary.ReconnectHint != "wait_for_restart" ||
		evaluation.Health.Summary.DisconnectCount != 2 ||
		evaluation.Health.Summary.DisconnectBurstCount != 1 ||
		evaluation.Health.Summary.DisconnectBurstWindowMs != 30000 ||
		evaluation.Health.Summary.CooldownRemainingMs != 900 ||
		evaluation.Health.Summary.RetryBackoffRemainingMs != 450 ||
		evaluation.Health.Summary.RestartAttemptCount != 4 ||
		evaluation.Health.Summary.RestartFailureCount != 1 ||
		evaluation.Health.Summary.LastDisconnectUnixMilli != 111 ||
		evaluation.Health.Summary.LastReconnectUnixMilli != 222 ||
		evaluation.Health.Summary.LastRestartAttemptUnixMilli != 333 ||
		evaluation.Health.Summary.LastRestartResult != "restarted" ||
		evaluation.Health.Summary.LastRestartError != "transport closed" ||
		evaluation.Health.Summary.RecommendedBackoffMs != 1200 {
		t.Fatalf("expected shared binding evaluation to preserve session health reliability metadata, got %#v", evaluation.Health.Summary)
	}
}

func TestBrowserRuntimeSessionCoordinationRestartPendingFallsBackToWaitGuidance(t *testing.T) {
	coordination := browserRuntimeSessionCoordination(context.Background(), nil, browserRuntimeSessionBinding{
		SessionHealthState:                   "restart_pending",
		SessionHealthReason:                  "browser relaunch backoff active for 450ms after restart failure",
		SessionHealthReconnectHint:           "retry_after_backoff",
		SessionHealthRetryBackoffRemainingMs: 450,
		SessionHealthRecommendedBackoffMs:    1200,
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
		BrowserProfileCount:        1,
		ActiveBrowserProfile:       "isolated",
		SelectedBrowserProfile:     "isolated",
		SelectedBrowserTarget:      "node",
		SelectedBrowserTargetID:    "target-1",
		RouteTargetCount:           1,
		BrowserProfileStatusCounts: map[string]int{"running": 1},
	}, &browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
	}, nil)
	if coordination == nil {
		t.Fatalf("expected restart_pending session coordination")
	}
	if coordination.State != "browser_ready" {
		t.Fatalf("expected restart_pending to preserve browser_ready state, got %#v", coordination)
	}
	if coordination.RestartBrowserAction != "" {
		t.Fatalf("expected restart_pending to clear restart browser action, got %#v", coordination)
	}
	if coordination.PrimaryBrowserAction != "browser action=wait" || coordination.NextStep != "browser action=wait" {
		t.Fatalf("expected restart_pending to fall back to wait guidance, got %#v", coordination)
	}
	if len(coordination.RecommendedBrowserActions) == 0 || coordination.RecommendedBrowserActions[0] != "browser action=wait" {
		t.Fatalf("expected restart_pending coordination to prepend wait guidance, got %#v", coordination)
	}
}

func TestBrowserRuntimeSharedBindingEvaluationPreservesSelectedBindingMetadata(t *testing.T) {
	evaluation := browserRuntimeSharedBindingEvaluation(browserRuntimeSessionBinding{
		SelectedBrowserBackend:       "proxy",
		SelectedBrowserApp:           "Chromium",
		SelectedBrowserProfile:       "workbench",
		SelectedBrowserProfileSource: "select_profile",
		SelectedBrowserTarget:        "node",
		SelectedBrowserTargetID:      "tab-2",
		SelectedBrowserTabIndex:      2,
		SelectedBrowserTargetSource:  "tracked_active_tab",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
	}, nil)

	if evaluation.Snapshot.SelectedProfileSelection == nil ||
		evaluation.Snapshot.SelectedProfileSelection.Backend != "proxy" ||
		evaluation.Snapshot.SelectedProfileSelection.Profile != "workbench" ||
		evaluation.Snapshot.SelectedProfileSelection.RuntimeTarget != "node" ||
		evaluation.Snapshot.SelectedProfileSelection.BrowserApp != "Chromium" ||
		evaluation.Snapshot.SelectedProfileSelection.Source != "select_profile" {
		t.Fatalf("expected shared binding evaluation to preserve selected profile metadata, got %#v", evaluation.Snapshot.SelectedProfileSelection)
	}
	if evaluation.Snapshot.SelectedTargetSelection == nil ||
		evaluation.Snapshot.SelectedTargetSelection.ID != "tab-2" ||
		evaluation.Snapshot.SelectedTargetSelection.TabIndex != 2 ||
		evaluation.Snapshot.SelectedTargetSelection.Backend != "proxy" ||
		evaluation.Snapshot.SelectedTargetSelection.Profile != "workbench" ||
		evaluation.Snapshot.SelectedTargetSelection.RuntimeTarget != "node" ||
		evaluation.Snapshot.SelectedTargetSelection.BrowserApp != "Chromium" ||
		evaluation.Snapshot.SelectedTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected shared binding evaluation to preserve selected target metadata, got %#v", evaluation.Snapshot.SelectedTargetSelection)
	}
}

func TestBrowserRuntimeSharedBindingEvaluationPreservesSelectedBindingMetadataWithoutProfileSnapshot(t *testing.T) {
	evaluation := browserRuntimeSharedBindingEvaluation(browserRuntimeSessionBinding{
		SelectedBrowserBackend:       "proxy",
		SelectedBrowserApp:           "Chromium",
		SelectedBrowserProfile:       "workbench",
		SelectedBrowserProfileSource: "select_profile",
		SelectedBrowserTarget:        "node",
		SelectedBrowserTargetID:      "tab-2",
		SelectedBrowserTabIndex:      2,
		SelectedBrowserTargetSource:  "tracked_active_tab",
	}, nil)

	if evaluation.Snapshot.SelectedProfileSelection == nil ||
		evaluation.Snapshot.SelectedProfileSelection.Backend != "proxy" ||
		evaluation.Snapshot.SelectedProfileSelection.Profile != "workbench" ||
		evaluation.Snapshot.SelectedProfileSelection.RuntimeTarget != "node" ||
		evaluation.Snapshot.SelectedProfileSelection.BrowserApp != "Chromium" ||
		evaluation.Snapshot.SelectedProfileSelection.Source != "select_profile" {
		t.Fatalf("expected shared binding evaluation to preserve selected profile metadata without profile snapshot, got %#v", evaluation.Snapshot.SelectedProfileSelection)
	}
	if evaluation.Snapshot.SelectedTargetSelection == nil ||
		evaluation.Snapshot.SelectedTargetSelection.ID != "tab-2" ||
		evaluation.Snapshot.SelectedTargetSelection.TabIndex != 2 ||
		evaluation.Snapshot.SelectedTargetSelection.Backend != "proxy" ||
		evaluation.Snapshot.SelectedTargetSelection.Profile != "workbench" ||
		evaluation.Snapshot.SelectedTargetSelection.RuntimeTarget != "node" ||
		evaluation.Snapshot.SelectedTargetSelection.BrowserApp != "Chromium" ||
		evaluation.Snapshot.SelectedTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected shared binding evaluation to preserve selected target metadata without profile snapshot, got %#v", evaluation.Snapshot.SelectedTargetSelection)
	}
}
