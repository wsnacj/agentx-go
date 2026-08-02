package tools

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestBrowserRuntimePrepareActionDispatchControlPlaneUsesSelectedRoute(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	selectedRoute := browserResolvedExecutionRoute{
		Backend: &runtimeControlBrowserBackend{
			runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
				fakeBrowserBackend: &fakeBrowserBackend{},
				runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			},
		},
		RuntimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		Capabilities: BrowserCapabilities{
			Open:       true,
			Screenshot: true,
			Click:      true,
		},
	}
	ctx := browserRegistrationContext{
		sessionRegistry:      sessionRegistry,
		sessionStateRegistry: stateRegistry,
		enabledTools: map[string]bool{
			"browser_runtime":    true,
			"browser_act":        true,
			"browser_screenshot": true,
		},
		substrateAssessment: browserDefaultSubstrateAssessment{
			DefaultRuntime: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		},
		substrateSummary: BrowserWorkbenchSubstrateSummary{
			DefaultRoute:       BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			SelectionStrategy:  BrowserSubstrateSelectionPreferNodeOverLegacy,
			SelectionReason:    "managed default route",
			ConfiguredTargets:  []string{"node", "host"},
			NodeConfigured:     true,
			NodeRouteAvailable: true,
		},
	}
	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
	}

	prepared := browserRuntimePrepareActionDispatchControlPlane(
		ctx,
		context.Background(),
		&payload,
		"status",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		ctx.substrateAssessment,
		browserRuntimeActionDispatchSelection{
			SelectedRoute:      selectedRoute,
			SelectedRouteReady: true,
		},
	)
	if prepared.Handled {
		t.Fatalf("expected selected route control-plane preflight to stay on the routed path, got %#v", prepared)
	}
	if prepared.SelectedInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected selected route info to flow through control-plane preflight, got %#v", prepared.SelectedInfo)
	}
	if prepared.ConfiguredInfo != prepared.SelectedInfo {
		t.Fatalf("expected configured info to follow selected route, got %#v want %#v", prepared.ConfiguredInfo, prepared.SelectedInfo)
	}
	if payload.SelectedRoute == nil || payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected payload selected route to be populated from dispatch control-plane preflight, got %#v", payload.SelectedRoute)
	}
	if !prepared.Capabilities.RuntimeStatus || !prepared.Capabilities.RuntimeWorkbench {
		t.Fatalf("expected control-plane preflight to expose runtime control capabilities from the selected route, got %#v", prepared.Capabilities)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_runtime") || !browserStringSliceContains(payload.BrowserTools, "browser_act") {
		t.Fatalf("expected selected-route control-plane preflight to project route-scoped browser tools, got %#v", payload.BrowserTools)
	}
	if !browserStringSliceContains(payload.ArtifactTools, "browser_screenshot") {
		t.Fatalf("expected selected-route control-plane preflight to project route-scoped artifact tools, got %#v", payload.ArtifactTools)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected control-plane preflight to refresh default route metadata from substrate context, got %#v", payload.DefaultRoute)
	}
}

func TestBrowserRuntimePrepareActionDispatchControlPlaneUsesBasePreviewOwnerWithoutActionProbe(t *testing.T) {
	nodeBackend := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo: BrowserRuntimeInfo{
				Backend: "proxy",
				Profile: "isolated",
				Target:  "node",
			},
			capabilities:  fullBrowserCapabilities(),
			routeSource:   "managed_browserd",
			routeEndpoint: "http://127.0.0.1:43123",
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	reg := llmxtools.NewRegistry()
	ctx, _, ok := newBrowserRegistrationContext(reg, BrowserToolOptions{
		Root:        t.TempDir(),
		NodeBackend: nodeBackend,
		EnabledTools: []string{
			"browser_runtime",
			"browser_act",
			"browser_click",
		},
	})
	if !ok {
		t.Fatalf("expected browser registration context")
	}
	payload := browserRuntimePayload{Action: "status", Status: "ok"}
	selectedRoute := browserResolvedExecutionRoute{
		Backend: nodeBackend,
		RuntimeInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "workbench",
			Target:  "node",
		},
		Capabilities: BrowserCapabilitiesForActKinds([]string{"click"}),
	}

	resolveCallsBeforePrepare := nodeBackend.resolveCalls
	prepared := browserRuntimePrepareActionDispatchControlPlane(
		ctx,
		context.Background(),
		&payload,
		"status",
		browserRegistrationDefaultRuntimeInfo(ctx),
		browserRegistrationSubstrateAssessment(ctx),
		browserRuntimeActionDispatchSelection{
			SelectedRoute:      selectedRoute,
			SelectedRouteReady: true,
		},
	)
	if nodeBackend.resolveCalls != resolveCallsBeforePrepare {
		t.Fatalf("expected control-plane preflight to reuse base preview owner without extra action probe, before=%d after=%d", resolveCallsBeforePrepare, nodeBackend.resolveCalls)
	}
	if prepared.Handled || payload.SelectedRoute == nil || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected selected route control-plane preflight to keep routed node path, got prepared=%#v payload=%#v", prepared, payload)
	}
}

func TestBrowserRuntimePrepareActionDispatchControlPlaneAppliesDiagnosticsDegrade(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-action-dispatch-control-plane-diagnostics")
	stateRegistry := NewBrowserSessionStateRegistry()
	stateRegistry.RecordBrowserProfileState("runtime-action-dispatch-control-plane-diagnostics", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached diagnostics snapshot",
	})
	defaultRoute := BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}
	substrateAssessment := browserDefaultSubstrateAssessmentForHostBackend(BrowserToolOptions{}, nil)
	ctx := browserRegistrationContext{
		sessionStateRegistry: stateRegistry,
		enabledTools: map[string]bool{
			"browser_runtime": true,
		},
		substrateAssessment: substrateAssessment,
		substrateSummary: BrowserWorkbenchSubstrateSummary{
			DefaultRoute:       defaultRoute,
			SelectionStrategy:  BrowserSubstrateSelectionPreferHostRuntime,
			SelectionReason:    "implicit host diagnostics",
			ConfiguredTargets:  []string{"host"},
			HostRoute:          defaultRoute,
			HostRouteAvailable: true,
		},
	}
	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
	}

	diagnosticsPreview := browserRuntimeDiagnosticsPreview{
		Registration: browserDefaultRuntimePreview{
			SubstrateSummary: BrowserWorkbenchSubstrateSummary{
				DefaultRoute:       defaultRoute,
				SelectionStrategy:  BrowserSubstrateSelectionPreferHostRuntime,
				SelectionReason:    "implicit host diagnostics",
				ConfiguredTargets:  []string{"host"},
				HostRoute:          defaultRoute,
				HostRouteAvailable: true,
			},
		},
		DefaultRoute:           defaultRoute,
		DefaultRouteDescriptor: browserRuntimeRouteDescriptor{Backend: "system", Profile: "default", RuntimeTarget: "host"},
		ConfiguredTargets:      []string{"host"},
	}

	prepared := browserRuntimePrepareActionDispatchControlPlaneWithPreview(
		ctx,
		callCtx,
		&payload,
		"status",
		defaultRoute,
		substrateAssessment,
		diagnosticsPreview,
		browserRuntimeActionDispatchSelection{
			UseHiddenImplicitHostDiagnosticsDegrade: true,
		},
	)
	if !prepared.Handled {
		t.Fatalf("expected diagnostics degrade control-plane preflight to be fully handled, got %#v", prepared)
	}
	if prepared.ConfiguredInfo != defaultRoute {
		t.Fatalf("expected diagnostics degrade to retain default-route configured info, got %#v", prepared.ConfiguredInfo)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected diagnostics degrade payload to remain ok, got %#v", payload)
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Profile != "default" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("expected diagnostics degrade payload to reuse cached default-route status snapshot, got %#v", payload.ProfileStatus)
	}
	if payload.DefaultProfile != "default" {
		t.Fatalf("expected diagnostics degrade payload to retain the default profile, got %#v", payload)
	}
	if payload.DefaultRoute != (browserRuntimeRouteDescriptor{Backend: "system", Profile: "default", RuntimeTarget: "host"}) {
		t.Fatalf("expected diagnostics degrade payload to project default route metadata, got %#v", payload.DefaultRoute)
	}
	if payload.Note != "" {
		t.Fatalf("expected host-only diagnostics degrade payload to stay note-free, got %#v", payload.Note)
	}
	if len(payload.ConfiguredTargets) != 1 || payload.ConfiguredTargets[0] != "host" {
		t.Fatalf("expected diagnostics degrade payload to project configured targets, got %#v", payload.ConfiguredTargets)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "status") {
		t.Fatalf("expected diagnostics degrade payload to advertise runtime status capability, got %#v", payload.RuntimeActions)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_runtime") {
		t.Fatalf("expected diagnostics degrade payload to advertise runtime diagnostics surface, got %#v", payload.BrowserTools)
	}
	if payload.SelectedRoute != nil {
		t.Fatalf("expected diagnostics degrade payload not to set a concrete selected route, got %#v", payload.SelectedRoute)
	}
}

func TestBrowserRuntimePrepareActionDispatchControlPlaneUsesManagedPendingDefaultNoteForHiddenDiagnosticsDegrade(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-action-dispatch-control-plane-hidden-managed-diagnostics")
	stateRegistry := NewBrowserSessionStateRegistry()
	stateRegistry.RecordBrowserProfileState("runtime-action-dispatch-control-plane-hidden-managed-diagnostics", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached diagnostics snapshot",
	})
	defaultRoute := BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}
	substrateAssessment := browserDefaultSubstrateAssessment{
		HostRuntime: defaultRoute,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				RuntimeInfo:  defaultRoute,
				Capabilities: defaultBrowserCapabilities(),
			},
		},
		DefaultRuntime:       defaultRoute,
		DefaultConcreteRoute: browserConcreteRouteAssessment{},
		NodeRoute: browserDefaultPromotionRouteAssessment{
			Configured:     true,
			RouteAvailable: false,
			FailureReason:  "node runtime route is configured but not the default yet",
			FailureNote:    "node runtime route is configured but not the default yet",
		},
	}
	ctx := browserRegistrationContext{
		sessionStateRegistry: stateRegistry,
		enabledTools: map[string]bool{
			"browser_runtime": true,
		},
		substrateAssessment: substrateAssessment,
		substrateSummary: BrowserWorkbenchSubstrateSummary{
			DefaultRoute:              defaultRoute,
			SelectionStrategy:         BrowserSubstrateSelectionLegacyHostDefault,
			SelectionReason:           "implicit host diagnostics",
			ConfiguredTargets:         []string{"host", "node"},
			HostRoute:                 defaultRoute,
			HostRouteAvailable:        true,
			NodeConfigured:            true,
			NodePromotionFailureCause: "node runtime route is configured but not the default yet",
		},
	}
	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
	}
	diagnosticsPreview := browserRuntimeDiagnosticsPreview{
		Registration: browserDefaultRuntimePreview{
			SubstrateAssessment: substrateAssessment,
			SubstrateSummary: BrowserWorkbenchSubstrateSummary{
				DefaultRoute:              defaultRoute,
				SelectionStrategy:         BrowserSubstrateSelectionLegacyHostDefault,
				SelectionReason:           "implicit host diagnostics",
				ConfiguredTargets:         []string{"host", "node"},
				HostRoute:                 defaultRoute,
				HostRouteAvailable:        true,
				NodeConfigured:            true,
				NodePromotionFailureCause: "node runtime route is configured but not the default yet",
			},
		},
		DefaultRoute:           defaultRoute,
		DefaultRouteDescriptor: browserRuntimeRouteDescriptor{},
		ConfiguredTargets:      []string{"host", "node"},
	}

	prepared := browserRuntimePrepareActionDispatchControlPlaneWithPreview(
		ctx,
		callCtx,
		&payload,
		"status",
		defaultRoute,
		substrateAssessment,
		diagnosticsPreview,
		browserRuntimeActionDispatchSelection{
			UseHiddenImplicitHostDiagnosticsDegrade: true,
		},
	)
	if !prepared.Handled {
		t.Fatalf("expected hidden managed diagnostics degrade control-plane preflight to be fully handled, got %#v", prepared)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected hidden managed diagnostics degrade payload to remain ok, got %#v", payload)
	}
	if payload.Note != "Managed browser route is configured, but the default route is still hidden behind the implicit legacy host fallback." {
		t.Fatalf("expected hidden managed diagnostics degrade to surface managed pending-default note, got %#v", payload.Note)
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Profile != "default" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("expected hidden managed diagnostics degrade payload to reuse cached default-route status snapshot, got %#v", payload.ProfileStatus)
	}
	if payload.DefaultRoute != (browserRuntimeRouteDescriptor{}) {
		t.Fatalf("expected hidden managed diagnostics degrade payload to keep implicit host default_route hidden before doctor gate, got %#v", payload.DefaultRoute)
	}
}

func TestBrowserRuntimePrepareActionDispatchControlPlanePreservesRouteErrorNoteOnUnsupported(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "prepare",
		Status: "ok",
	}
	defaultRoute := BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}
	substrateAssessment := browserDefaultSubstrateAssessmentForHostBackend(BrowserToolOptions{}, nil)
	routeErr := context.DeadlineExceeded

	prepared := browserRuntimePrepareActionDispatchControlPlaneWithPreview(
		browserRegistrationContext{},
		context.Background(),
		&payload,
		"prepare",
		defaultRoute,
		substrateAssessment,
		browserRuntimeDiagnosticsPreview{},
		browserRuntimeActionDispatchSelection{
			RouteErr: routeErr,
		},
	)
	if !prepared.Handled {
		t.Fatalf("expected unsupported route error to be handled in control-plane preflight, got %#v", prepared)
	}
	if payload.Status != "unsupported" || payload.Note != routeErr.Error() {
		t.Fatalf("expected unsupported control-plane preflight to preserve route error note, got %#v", payload)
	}
}

func TestBrowserRuntimePrepareActionDispatchControlPlaneSkipsDiagnosticsDegradeForManagedLaunchFailure(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
	}
	defaultRoute := BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}
	managedRouteErr := errors.New("browser proxy managed_route_unavailable target=node endpoint=http://127.0.0.1:1: managed browserd boot failed")
	preview := browserRuntimeDiagnosticsPreview{
		Registration: browserDefaultRuntimePreview{
			SubstrateAssessment: browserDefaultSubstrateAssessment{
				HostRuntime: defaultRoute,
				HostRoute: browserConcreteRouteAssessment{
					Configured:     true,
					RouteAvailable: true,
					Route: browserResolvedExecutionRoute{
						RuntimeInfo:  defaultRoute,
						Capabilities: defaultBrowserCapabilities(),
					},
				},
				DefaultRuntime: defaultRoute,
				NodeRoute: browserDefaultPromotionRouteAssessment{
					Configured:     true,
					FailureReason:  "node runtime route is configured but not the default because its concrete default route could not be resolved: " + managedRouteErr.Error(),
					FailureNote:    managedRouteErr.Error(),
					RouteAvailable: false,
				},
			},
			SubstrateSummary: BrowserWorkbenchSubstrateSummary{
				DefaultRoute:              defaultRoute,
				DefaultCandidateRoute:     BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
				DefaultCandidateEndpoint:  "http://127.0.0.1:1",
				SelectionStrategy:         BrowserSubstrateSelectionLegacyHostDefault,
				SelectionReason:           "managed default route failed; implicit host fallback remains hidden",
				ConfiguredTargets:         []string{"host", "node"},
				HostRoute:                 defaultRoute,
				HostRouteAvailable:        true,
				NodeConfigured:            true,
				NodePromotionFailureCause: managedRouteErr.Error(),
			},
		},
		DefaultRoute:           defaultRoute,
		DefaultRouteDescriptor: browserRuntimeRouteDescriptor{Backend: "system", Profile: "default", RuntimeTarget: "host"},
		ConfiguredTargets:      []string{"host", "node"},
	}

	prepared := browserRuntimePrepareActionDispatchControlPlaneWithPreview(
		browserRegistrationContext{
			enabledTools: map[string]bool{
				"browser_runtime": true,
			},
		},
		context.Background(),
		&payload,
		"status",
		defaultRoute,
		preview.Registration.SubstrateAssessment,
		preview,
		browserRuntimeActionDispatchSelection{
			DefaultCandidateRoute:                   BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			RouteErr:                                managedRouteErr,
			UseHiddenImplicitHostDiagnosticsDegrade: true,
		},
	)
	if !prepared.Handled {
		t.Fatalf("expected managed launch failure control-plane preflight to be handled, got %#v", prepared)
	}
	if payload.Status != "unsupported" || payload.Note != managedRouteErr.Error() {
		t.Fatalf("expected managed launch failure to stay unsupported instead of degrading, got %#v", payload)
	}
	if prepared.DefaultCandidateRoute != (BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}) {
		t.Fatalf("expected managed launch failure control-plane preflight to preserve selection default candidate route, got %#v", prepared.DefaultCandidateRoute)
	}
	if payload.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Endpoint:      "http://127.0.0.1:1",
	}) {
		t.Fatalf("expected managed launch failure control-plane preflight to preserve selection default candidate route, got %#v", payload.DefaultCandidateRoute)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "status") {
		t.Fatalf("expected managed launch failure control-plane preflight to retain diagnostics status action, got %#v", payload.RuntimeActions)
	}
}

func TestBrowserRuntimePrepareActionDispatchControlPlaneWithPreviewUsesProvidedSubstrateSnapshot(t *testing.T) {
	selectedRoute := browserResolvedExecutionRoute{
		RuntimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		Capabilities: BrowserCapabilities{
			RuntimeStatus:    true,
			RuntimeWorkbench: true,
		},
	}
	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
	}
	diagnosticsPreview := browserRuntimeDiagnosticsPreview{
		Registration: browserDefaultRuntimePreview{
			SubstrateSummary: BrowserWorkbenchSubstrateSummary{
				DefaultCandidateRoute:    BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
				DefaultCandidateSource:   "managed_browserd",
				DefaultCandidateEndpoint: "http://127.0.0.1:43123",
				SelectionStrategy:        BrowserSubstrateSelectionPreferNodeOverLegacy,
				SelectionReason:          "shared execution preview",
			},
		},
		DefaultRoute:           BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		DefaultRouteDescriptor: browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node"},
		ConfiguredTargets:      []string{"node", "host"},
	}

	prepared := browserRuntimePrepareActionDispatchControlPlaneWithPreview(
		browserRegistrationContext{},
		context.Background(),
		&payload,
		"status",
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		browserDefaultSubstrateAssessment{},
		diagnosticsPreview,
		browserRuntimeActionDispatchSelection{
			DefaultCandidateRoute: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			SelectedRoute:         selectedRoute,
			SelectedRouteReady:    true,
		},
	)
	if prepared.Handled {
		t.Fatalf("expected selected route control-plane preflight to stay on the routed path, got %#v", prepared)
	}
	if prepared.DefaultCandidateRoute != (BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}) {
		t.Fatalf("expected selected route control-plane preflight to preserve selection default candidate route, got %#v", prepared.DefaultCandidateRoute)
	}
	if payload.DefaultRoute != diagnosticsPreview.DefaultRouteDescriptor {
		t.Fatalf("expected payload default route to reuse provided execution preview snapshot, got %#v want %#v", payload.DefaultRoute, diagnosticsPreview.DefaultRouteDescriptor)
	}
	if payload.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Source:        "managed_browserd",
		Endpoint:      "http://127.0.0.1:43123",
	}) {
		t.Fatalf("expected payload default candidate route to backfill from selection when preview lacks one, got %#v", payload.DefaultCandidateRoute)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy || payload.SubstrateSelectionReason != "shared execution preview" {
		t.Fatalf("expected payload substrate summary to reuse provided execution preview snapshot, got %#v", payload)
	}
	if len(payload.ConfiguredTargets) != 2 || payload.ConfiguredTargets[0] != "node" || payload.ConfiguredTargets[1] != "host" {
		t.Fatalf("expected payload configured targets to reuse provided execution preview snapshot, got %#v", payload.ConfiguredTargets)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "status") {
		t.Fatalf("expected payload runtime actions to reuse shared route projection, got %#v", payload.RuntimeActions)
	}
}

func TestBrowserRuntimePrepareActionDispatchControlPlaneWithPreviewPrefersSelectionCandidateDescriptor(t *testing.T) {
	selectedRoute := browserResolvedExecutionRoute{
		RuntimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		Capabilities: BrowserCapabilities{
			RuntimeStatus:    true,
			RuntimeWorkbench: true,
		},
	}
	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
	}
	diagnosticsPreview := browserRuntimeDiagnosticsPreview{
		Registration: browserDefaultRuntimePreview{
			SubstrateSummary: BrowserWorkbenchSubstrateSummary{
				SelectionStrategy: BrowserSubstrateSelectionPreferNodeOverLegacy,
				SelectionReason:   "session execution preview",
			},
		},
		DefaultRoute:           BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		DefaultRouteDescriptor: browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node"},
		ConfiguredTargets:      []string{"node", "host"},
	}

	prepared := browserRuntimePrepareActionDispatchControlPlaneWithPreview(
		browserRegistrationContext{},
		context.Background(),
		&payload,
		"status",
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		browserDefaultSubstrateAssessment{},
		diagnosticsPreview,
		browserRuntimeActionDispatchSelection{
			DefaultCandidateRoute: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			DefaultCandidateDescriptor: browserRuntimeRouteDescriptor{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Source:        "managed_browserd",
				Endpoint:      "http://127.0.0.1:43123",
			},
			SelectedRoute:      selectedRoute,
			SelectedRouteReady: true,
		},
	)
	if prepared.Handled {
		t.Fatalf("expected selected route control-plane preflight to stay on the routed path, got %#v", prepared)
	}
	if prepared.DefaultCandidateDescriptor != (browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Source:        "managed_browserd",
		Endpoint:      "http://127.0.0.1:43123",
	}) {
		t.Fatalf("expected selected route control-plane preflight to preserve selection candidate descriptor, got %#v", prepared.DefaultCandidateDescriptor)
	}
	if payload.DefaultCandidateRoute != prepared.DefaultCandidateDescriptor {
		t.Fatalf("expected payload default candidate route to prefer selection candidate descriptor provenance, got %#v want %#v", payload.DefaultCandidateRoute, prepared.DefaultCandidateDescriptor)
	}
}

func TestBrowserRuntimeDegradedRouteProjectionFromSnapshotBuildsWorkbenchProjection(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-degraded-route-projection")
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	stateRegistry.SelectBrowserProfile("runtime-degraded-route-projection", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})
	stateRegistry.RecordBrowserProfileState("runtime-degraded-route-projection", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached diagnostics state",
	})
	tracked := sessionRegistry.TrackCurrentTarget("runtime-degraded-route-projection", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	projection := browserRuntimeDegradedRouteProjectionFromSnapshot(
		callCtx,
		sessionRegistry,
		nil,
		stateRegistry,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		"",
	)

	if projection.DefaultProfile != "workbench" {
		t.Fatalf("expected degraded route projection to keep default profile, got %#v", projection)
	}
	if projection.ProfileStatus == nil || projection.ProfileStatus.Profile != "workbench" || projection.ProfileStatus.Status != "running" || !projection.ProfileStatus.Running || !projection.ProfileStatus.Connected {
		t.Fatalf("expected degraded route projection to reuse cached profile status, got %#v", projection.ProfileStatus)
	}
	if len(projection.Profiles) != 1 || projection.Profiles[0].Profile != "workbench" || projection.Profiles[0].Status != "running" {
		t.Fatalf("expected degraded route projection to reuse cached profiles, got %#v", projection.Profiles)
	}
	if len(projection.SessionRoutes) != 1 || projection.SessionRoutes[0].Backend != "proxy" || projection.SessionRoutes[0].Profile != "workbench" || projection.SessionRoutes[0].RuntimeTarget != "node" || projection.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected degraded route projection to populate session routes from route snapshot, got %#v", projection.SessionRoutes)
	}
	if projection.SessionTargetCount != 1 {
		t.Fatalf("expected degraded route projection to populate session target count, got %#v", projection)
	}
	if len(projection.SessionProfiles) != 1 || projection.SessionProfiles[0].Profile != "workbench" {
		t.Fatalf("expected degraded route projection to populate session profiles, got %#v", projection.SessionProfiles)
	}
	if projection.SessionProfileSelection == nil || projection.SessionProfileSelection.Profile != "workbench" || projection.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected degraded route projection to reuse session profile selection, got %#v", projection.SessionProfileSelection)
	}
	if projection.SessionTargetSelection == nil || projection.SessionTargetSelection.ID != tracked.ID || projection.SessionTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected degraded route projection to reuse current target selection, got %#v", projection.SessionTargetSelection)
	}
	if projection.SessionBinding == nil || projection.SessionBinding.CurrentTargetID != tracked.ID || projection.SessionBinding.SelectedBrowserProfile != "workbench" || projection.SessionBinding.SelectedBrowserTargetID != tracked.ID {
		t.Fatalf("expected degraded route projection to build a session binding, got %#v", projection.SessionBinding)
	}
}

func TestBrowserRuntimeApplyDegradedActionDispatchPayloadWorkbenchUsesSharedProjection(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-degraded-dispatch-workbench")
	sessionRegistry := NewBrowserSessionRegistry()
	tracked := sessionRegistry.TrackCurrentTarget("runtime-degraded-dispatch-workbench", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://93.184.216.34/dashboard",
		Title:      "Dashboard",
		BrowserApp: "Chrome",
		Backend:    "custom-playwright",
		Profile:    "workbench",
		Target:     "host",
	}, "tracked_active_tab")
	ctx := browserRegistrationContext{
		sessionRegistry: sessionRegistry,
	}
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
	}

	browserRuntimeApplyDegradedActionDispatchPayload(
		ctx,
		callCtx,
		&payload,
		"workbench",
		BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
		"workbench",
	)

	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected degraded workbench payload to keep default profile, got %#v", payload)
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected degraded workbench payload to fall back to cached route-scoped status, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Profile != "workbench" || payload.Profiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected degraded workbench payload to fall back to cached route-scoped profiles, got %#v", payload.Profiles)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].Backend != "custom-playwright" || payload.SessionRoutes[0].Profile != "workbench" || payload.SessionRoutes[0].RuntimeTarget != "host" || payload.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected degraded workbench payload to reuse route-scoped session snapshot, got %#v", payload.SessionRoutes)
	}
	if payload.SessionTargetCount != 1 {
		t.Fatalf("expected degraded workbench payload to surface session target count, got %#v", payload)
	}
	if len(payload.SessionProfiles) != 1 || payload.SessionProfiles[0].Profile != "workbench" || payload.SessionProfiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected degraded workbench payload to reuse fallback session profiles, got %#v", payload.SessionProfiles)
	}
	if payload.SessionProfileSelection != nil {
		t.Fatalf("expected degraded workbench payload without state registry not to synthesize profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != tracked.ID || payload.SessionTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected degraded workbench payload to surface current target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding == nil || payload.SessionBinding.CurrentTargetID != tracked.ID || payload.SessionBinding.SelectedBrowserTargetID != tracked.ID {
		t.Fatalf("expected degraded workbench payload to build session binding from shared projection, got %#v", payload.SessionBinding)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected degraded workbench payload to keep configured profiles, got %#v", payload.ConfiguredProfiles)
	}
	for _, want := range []string{"route", "status", "profiles", "sessions"} {
		if !browserStringSliceContains(payload.WorkbenchSections, want) {
			t.Fatalf("expected degraded workbench payload to include %q section, got %#v", want, payload.WorkbenchSections)
		}
	}
}

func TestBrowserRuntimeDegradedRouteProjectionFromSnapshotFallsBackToRouteScopedSessionProfilesWithoutStateRegistry(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-degraded-route-projection-no-state")
	sessionRegistry := NewBrowserSessionRegistry()
	tracked := sessionRegistry.TrackCurrentTarget("runtime-degraded-route-projection-no-state", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://93.184.216.34/dashboard",
		Title:      "Dashboard",
		BrowserApp: "Chrome",
		Backend:    "custom-playwright",
		Profile:    "workbench",
		Target:     "host",
	}, "tracked_active_tab")

	projection := browserRuntimeDegradedRouteProjectionFromSnapshot(
		callCtx,
		sessionRegistry,
		nil,
		nil,
		BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
		"workbench",
	)

	if len(projection.SessionRoutes) != 1 ||
		projection.SessionRoutes[0].Backend != "custom-playwright" ||
		projection.SessionRoutes[0].Profile != "workbench" ||
		projection.SessionRoutes[0].RuntimeTarget != "host" ||
		projection.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected degraded route projection without state registry to keep route-scoped session snapshot, got %#v", projection.SessionRoutes)
	}
	if len(projection.SessionProfiles) != 1 ||
		projection.SessionProfiles[0].Backend != "custom-playwright" ||
		projection.SessionProfiles[0].Profile != "workbench" ||
		projection.SessionProfiles[0].RuntimeTarget != "host" ||
		projection.SessionProfiles[0].BrowserApp != "chrome" ||
		projection.SessionProfiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected degraded route projection without state registry to synthesize route-scoped session profiles, got %#v", projection.SessionProfiles)
	}
}

func TestBrowserRuntimeApplyDegradedActionDispatchPayloadWorkbenchRefreshesTargetSpecificRouteSurface(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	ctx, enabled, ok := newBrowserRegistrationContext(reg, BrowserToolOptions{
		Root: t.TempDir(),
		Backend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
		},
		NodeBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"click"}),
		},
		SandboxBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "sandbox", Target: "sandbox"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"screenshot"}),
		},
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime", "browser_act", "browser_click", "browser_screenshot"},
	})
	if !ok {
		t.Fatalf("expected browser registration context")
	}
	registerEnabledBrowserTools(ctx, enabled)

	callCtx := WithToolSessionID(context.Background(), "runtime-degraded-route-surface-workbench")
	sessionRegistry.TrackCurrentTarget("runtime-degraded-route-surface-workbench", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	payload := browserRuntimePayload{
		Action:              "workbench",
		Status:              "ok",
		BrowserSurface:      "explicit_managed_opt_in",
		BrowserOptInTargets: []string{"sandbox"},
	}

	browserRuntimeApplyDegradedActionDispatchPayload(
		ctx,
		callCtx,
		&payload,
		"workbench",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		"workbench",
	)

	if payload.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected degraded workbench payload to replace stale root route surface with node target, got %#v", payload)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_runtime") ||
		!browserStringSliceContains(payload.BrowserTools, "browser_act") ||
		!browserStringSliceContains(payload.BrowserTools, "browser_click") ||
		browserStringSliceContains(payload.BrowserTools, "browser_screenshot") {
		t.Fatalf("expected degraded workbench payload to reuse node-specific managed tool metadata, got %#v", payload.BrowserTools)
	}
	if !reflect.DeepEqual(payload.BrowserActKinds, []string{"click"}) {
		t.Fatalf("expected degraded workbench payload to reuse node-specific managed act surface, got %#v", payload.BrowserActKinds)
	}
	if payload.Workbench == nil ||
		payload.Workbench.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.Workbench.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected degraded workbench surface to inherit refreshed node route surface, got %#v", payload.Workbench)
	}
	if payload.Surface == nil ||
		payload.Surface.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.Surface.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected degraded top-level surface to keep refreshed node route surface, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.View.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected degraded top-level view to keep refreshed node route surface, got %#v", payload.View)
	}
}

func TestBrowserRuntimeApplyDegradedActionDispatchPayloadWithPreviewDoesNotLeakCurrentManagedOptInSurface(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
	}
	nodeBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"click"}),
	}
	opts := BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_runtime", "browser_act", "browser_click"},
	}
	RegisterBrowserTools(reg, opts)
	ctx, _, ok := newBrowserRegistrationContext(reg, opts)
	if !ok {
		t.Fatalf("expected browser registration context")
	}

	registration := browserDefaultRuntimePreviewForDispatchOptions(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	preview := browserRuntimeDiagnosticsPreview{
		Registration:           registration,
		DefaultRoute:           registration.LogicalDefaultRoute,
		DefaultRouteDescriptor: browserRuntimeRouteDescriptor{},
		ConfiguredTargets:      append([]string(nil), registration.SubstrateSummary.ConfiguredTargets...),
	}

	payload := browserRuntimePayload{
		Action: "status",
		Status: "ok",
	}
	browserRuntimeRefreshSubstrateContextWithPreview(ctx, &payload, preview)
	if browserStringSliceContains(payload.BrowserTools, "browser_click") || browserStringSliceContains(payload.BrowserTools, "browser_act") {
		t.Fatalf("expected preview-owned status payload to start without current managed opt-in tools, got %#v", payload.BrowserTools)
	}

	browserRuntimeApplyDegradedActionDispatchPayloadWithPreview(
		ctx,
		WithToolSessionID(context.Background(), "runtime-degraded-preview-owned-status"),
		&payload,
		"status",
		preview.DefaultRoute,
		"",
		preview,
	)

	if payload.BrowserSurface != "" || len(payload.BrowserOptInTargets) != 0 {
		t.Fatalf("expected degraded payload with provided hidden implicit host preview not to leak managed opt-in route hints, got %#v", payload)
	}
	if browserStringSliceContains(payload.BrowserTools, "browser_click") || browserStringSliceContains(payload.BrowserTools, "browser_act") {
		t.Fatalf("expected degraded payload with provided hidden implicit host preview not to leak current managed opt-in tools, got %#v", payload.BrowserTools)
	}
}

func TestBrowserRuntimeApplySessionSelectionProjectionKeepsRouteStableWithoutTargetRouteApply(t *testing.T) {
	payload := browserRuntimePayload{
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	browserRuntimeApplySessionSelectionProjection(&payload, browserRuntimeSessionSelectionProjection{
		ProfileSelection: &browserRuntimeSessionProfileSelection{
			Backend:       "proxy",
			Profile:       "alternate",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "sync_session",
		},
		TargetSelection: &browserRuntimeSessionTargetSelection{
			ID:            "target-2",
			TabIndex:      2,
			Backend:       "proxy",
			Profile:       "alternate",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "sync_session",
		},
	})

	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "alternate" || payload.SessionProfileSelection.Source != "sync_session" {
		t.Fatalf("expected selection projection helper to populate profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != "target-2" || payload.SessionTargetSelection.Source != "sync_session" {
		t.Fatalf("expected selection projection helper to populate target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SelectedRoute == nil || payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected selection projection helper without route-apply to keep selected route stable, got %#v", payload.SelectedRoute)
	}
}

func TestBrowserRuntimeApplySessionSelectionProjectionAppliesTargetToRoute(t *testing.T) {
	payload := browserRuntimePayload{
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	browserRuntimeApplySessionSelectionProjection(&payload, browserRuntimeSessionSelectionProjection{
		ProfileSelection: &browserRuntimeSessionProfileSelection{
			Backend:       "proxy",
			Profile:       "alternate",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "select_target",
		},
		TargetSelection: &browserRuntimeSessionTargetSelection{
			ID:            "target-2",
			TabIndex:      2,
			Backend:       "proxy",
			Profile:       "alternate",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "select_target",
		},
		ApplyTargetToRoute: true,
	})

	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "alternate" || payload.SessionProfileSelection.Source != "select_target" {
		t.Fatalf("expected selection projection helper to populate promoted profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != "target-2" || payload.SessionTargetSelection.Source != "select_target" {
		t.Fatalf("expected selection projection helper to populate target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SelectedRoute == nil || payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "alternate" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected selection projection helper to apply target-backed route mutation, got %#v", payload.SelectedRoute)
	}
}

func TestBrowserRuntimeApplySessionActionOutcomeAppliesSelectTargetProjectionAndNote(t *testing.T) {
	payload := browserRuntimePayload{
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	browserRuntimeApplySessionActionOutcome(&payload, agentxbrowserruntime.SharedSessionBrowserActionOutcome{
		ApplyDecision:        true,
		Action:               "select_target",
		Decision:             "session_target_selected",
		Ready:                true,
		SelectTargetDecision: "session_target_selected",
		SelectTargetReady:    true,
		Note:                 "target selected",
		SelectionProjection: &agentxbrowserruntime.SharedSessionBrowserSelectionProjection{
			ProfileSelection: &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
				Backend:       "proxy",
				Profile:       "alternate",
				RuntimeTarget: "node",
				Source:        "select_target",
			},
			TargetSelection: &agentxbrowserruntime.BrowserSessionTargetSelection{
				ID:            "target-2",
				Backend:       "proxy",
				Profile:       "alternate",
				RuntimeTarget: "node",
				Source:        "select_target",
			},
			ApplyTargetToRoute: true,
		},
	})

	if payload.SelectTargetDecision != "session_target_selected" || !payload.SelectTargetReady {
		t.Fatalf("expected session action outcome helper to populate select_target decision/ready, got %#v", payload)
	}
	if payload.Note != "target selected" {
		t.Fatalf("expected session action outcome helper to preserve note, got %#v", payload)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != "target-2" {
		t.Fatalf("expected session action outcome helper to apply target selection projection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SelectedRoute == nil || payload.SelectedRoute.Profile != "alternate" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected session action outcome helper to apply target-backed route mutation, got %#v", payload.SelectedRoute)
	}
}

func TestBrowserRuntimeApplySessionActionOutcomeAppliesCoordinateSyncErrorProjection(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	browserRuntimeApplySessionActionOutcome(&payload, agentxbrowserruntime.SharedSessionBrowserActionOutcome{
		ApplyDecision:                    true,
		Action:                           "coordinate",
		CoordinationGoal:                 "sync",
		Decision:                         "session_target_sync_failed",
		SyncSessionDecision:              "session_target_sync_failed",
		Err:                              errors.New("sync failed"),
		ApplyCoordinationDecisionOnError: true,
		SelectionProjection: &agentxbrowserruntime.SharedSessionBrowserSelectionProjection{
			TargetSelection: &agentxbrowserruntime.BrowserSessionTargetSelection{
				ID:            "target-1",
				RuntimeTarget: "node",
				Source:        "sync_session",
			},
			ApplyTargetToRoute: true,
		},
	})

	if payload.SyncSessionDecision != "session_target_sync_failed" || payload.SyncSessionReady {
		t.Fatalf("expected coordinate sync outcome to reuse sync_session decision fields, got %#v", payload)
	}
	if payload.Status != "error" || payload.Note != "sync failed" {
		t.Fatalf("expected coordinate sync outcome to surface error status/note, got %#v", payload)
	}
	if payload.CoordinationDecision != "session_target_sync_failed" {
		t.Fatalf("expected coordinate sync outcome to preserve coordination decision on error, got %#v", payload)
	}
	if payload.SessionTargetSelection != nil {
		t.Fatalf("expected coordinate sync outcome to skip selection projection on error, got %#v", payload.SessionTargetSelection)
	}
}

func TestBrowserRuntimeDispatchSessionSyncActionAppliesUnsupportedSyncStatus(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	handled := browserRuntimeDispatchSessionSyncAction(
		context.Background(),
		&payload,
		browserRuntimeSessionSyncDispatchOptions{
			Action:       "sync_session",
			Capabilities: BrowserCapabilities{},
		},
	)

	if !handled {
		t.Fatalf("expected sync dispatch helper to handle sync_session action")
	}
	if payload.Status != "unsupported" || payload.Note != "browser_runtime: selected route does not support action sync_session" {
		t.Fatalf("expected sync dispatch helper to apply unsupported status/note, got %#v", payload)
	}
}

func TestBrowserRuntimeDispatchSessionSyncActionRunsCoordinateSyncResult(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-sync-dispatch-helper")
	tracked := sessionRegistry.TrackTabs("browser-runtime-coordinate-sync-dispatch-helper", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://node.example/workbench",
			Title:      "Workbench",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
	}, 1)
	if len(tracked) != 1 || strings.TrimSpace(tracked[0].ID) == "" {
		t.Fatalf("expected one tracked target for coordinate sync helper, got %#v", tracked)
	}

	payload := browserRuntimePayload{Status: "ok"}
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
		},
	}

	handled := browserRuntimeDispatchSessionSyncAction(
		callCtx,
		&payload,
		browserRuntimeSessionSyncDispatchOptions{
			Action:               "coordinate",
			CoordinationGoal:     "sync",
			Capabilities:         BrowserCapabilities{RuntimeCoordinate: true},
			SelectedBackend:      control,
			WatchManagerProvider: agentxbrowserruntime.SharedSessionBrowserObserverManager{SessionRegistry: sessionRegistry},
			SelectedInfo: BrowserRuntimeInfo{
				Backend: "proxy",
				Profile: "workbench",
				Target:  "node",
			},
			SelectedRoute: &browserRuntimeRouteDescriptor{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
			},
			RequestedBrowserApp: "Chromium",
		},
	)

	if !handled {
		t.Fatalf("expected sync dispatch helper to handle coordinate(sync)")
	}
	if payload.Status != "ok" || payload.SyncSessionDecision != "session_target_synced" || !payload.SyncSessionReady {
		t.Fatalf("expected sync dispatch helper to apply coordinate(sync) outcome, got %#v", payload)
	}
	if payload.SessionProfileSelection != nil {
		t.Fatalf("expected coordinate sync helper to preserve target-first contract without profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != tracked[0].ID || payload.SessionTargetSelection.TabIndex != 1 || payload.SessionTargetSelection.Source != "sync_session" {
		t.Fatalf("expected coordinate sync helper to project tracked target selection, got %#v", payload.SessionTargetSelection)
	}
}

func TestBrowserRuntimeDispatchSessionSelectionActionAppliesUnsupportedSelectProfileStatus(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	handled := browserRuntimeDispatchSessionSelectionAction(
		context.Background(),
		&payload,
		browserRuntimeSessionSelectionDispatchOptions{
			Action:       "select_profile",
			Capabilities: BrowserCapabilities{},
		},
	)

	if !handled {
		t.Fatalf("expected selection dispatch helper to handle select_profile action")
	}
	if payload.Status != "unsupported" || payload.Note != "browser_runtime: selected route does not support action select_profile" {
		t.Fatalf("expected selection dispatch helper to apply unsupported status/note, got %#v", payload)
	}
}

func TestBrowserRuntimeDispatchSessionSelectionActionRunsSelectTargetPromotion(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-target-dispatch-helper")
	tracked := sessionRegistry.TrackTabs("browser-runtime-select-target-dispatch-helper", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   2,
			URL:        "https://node.example/workbench",
			Title:      "Workbench",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
	}, 2)
	if len(tracked) != 1 || strings.TrimSpace(tracked[0].ID) == "" {
		t.Fatalf("expected one tracked target for select_target helper, got %#v", tracked)
	}

	payload := browserRuntimePayload{
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	handled := browserRuntimeDispatchSessionSelectionAction(
		callCtx,
		&payload,
		browserRuntimeSessionSelectionDispatchOptions{
			Action:               "select_target",
			Capabilities:         BrowserCapabilities{RuntimeSelectTarget: true},
			WatchManagerProvider: agentxbrowserruntime.SharedSessionBrowserObserverManager{SessionRegistry: sessionRegistry, StateRegistry: stateRegistry},
			StateRegistry:        stateRegistry,
			SelectedInfo: BrowserRuntimeInfo{
				Backend: "proxy",
				Profile: "workbench",
				Target:  "node",
			},
			SelectedRoute:       payload.SelectedRoute,
			RequestedBrowserApp: "Chromium",
			Params: map[string]any{
				"target": "tab:2",
			},
		},
	)

	if !handled {
		t.Fatalf("expected selection dispatch helper to handle select_target action")
	}
	if payload.Status != "ok" || payload.SelectTargetDecision != "session_target_already_selected" || !payload.SelectTargetReady {
		t.Fatalf("expected selection dispatch helper to apply select_target decision/ready, got %#v", payload)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != tracked[0].ID || payload.SessionTargetSelection.Source != "select_target" {
		t.Fatalf("expected selection dispatch helper to project selected target, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.RuntimeTarget != "node" || payload.SessionProfileSelection.Source != "select_target" {
		t.Fatalf("expected selection dispatch helper to promote matching profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SelectedRoute == nil || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected selection dispatch helper to keep target-backed route mutation, got %#v", payload.SelectedRoute)
	}
	if stored, ok := stateRegistry.SelectedBrowserProfile("browser-runtime-select-target-dispatch-helper", "node"); !ok || stored.Profile != "workbench" || stored.Source != "select_target" {
		t.Fatalf("expected selection dispatch helper to persist promoted profile selection, got %#v %#v", stored, ok)
	}
}

func TestBrowserRuntimeApplySessionActionOutcomeClearProfileResetsStatusBeforeInventoryProjection(t *testing.T) {
	payload := browserRuntimePayload{}

	browserRuntimeApplySessionActionOutcome(&payload, agentxbrowserruntime.SharedSessionBrowserActionOutcome{
		ApplyDecision:      true,
		Action:             "clear_profile",
		Decision:           "session_profile_cleared",
		Ready:              true,
		ClearDecision:      "session_profile_cleared",
		ClearReady:         true,
		ClearProfileStatus: true,
	})

	if payload.ClearDecision != "session_profile_cleared" || !payload.ClearReady {
		t.Fatalf("expected clear action outcome to populate clear_profile decision/ready, got %#v", payload)
	}
	if payload.ProfileStatus != nil {
		t.Fatalf("expected clear action outcome to clear stale profile status before inventory apply, got %#v", payload.ProfileStatus)
	}
	if payload.ClearSessionDecision != "" || payload.ClearTargetDecision != "" {
		t.Fatalf("expected clear_profile outcome not to leak sibling clear decisions, got %#v", payload)
	}
}

func TestBrowserRuntimeApplySessionActionOutcomeClearSessionKeepsCleanupCounts(t *testing.T) {
	payload := browserRuntimePayload{}

	browserRuntimeApplySessionActionOutcome(&payload, agentxbrowserruntime.SharedSessionBrowserActionOutcome{
		ApplyDecision:          true,
		Action:                 "clear_session",
		Decision:               "session_route_cleared",
		Ready:                  true,
		ClearSessionDecision:   "session_route_cleared",
		ClearSessionReady:      true,
		ClearProfileStatus:     true,
		ClearedSessionProfiles: 2,
		ClearedSessionTargets:  3,
	})

	if payload.ClearSessionDecision != "session_route_cleared" || !payload.ClearSessionReady {
		t.Fatalf("expected clear action outcome to populate clear_session decision/ready, got %#v", payload)
	}
	if payload.ClearedSessionProfiles != 2 || payload.ClearedSessionTargets != 3 {
		t.Fatalf("expected clear action outcome to preserve clear_session cleanup counts, got %#v", payload)
	}
	if payload.ProfileStatus != nil {
		t.Fatalf("expected clear action outcome not to apply raw profile status directly, got %#v", payload.ProfileStatus)
	}
	if payload.ClearDecision != "" || payload.ClearTargetDecision != "" {
		t.Fatalf("expected clear_session outcome not to leak sibling clear decisions, got %#v", payload)
	}
}

func TestBrowserRuntimeApplySessionActionOutcomeDelegatesClearResultProjection(t *testing.T) {
	payload := browserRuntimePayload{}

	browserRuntimeApplySessionActionOutcome(&payload, agentxbrowserruntime.SharedSessionBrowserActionOutcome{
		ApplyDecision:        true,
		Action:               "clear_session",
		ClearSessionDecision: "session_route_cleared",
		ClearSessionReady:    true,
		ProfileInventoryProjection: &agentxbrowserruntime.SharedSessionBrowserTopLevelProfileInventoryProjection{
			ProfileStatus: agentxbrowserruntime.SharedSessionBrowserProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Status:        "stopped",
				Running:       false,
				Connected:     false,
			},
			HasProfileStatus: true,
		},
		Decision:               "session_route_cleared",
		Ready:                  true,
		ClearProfileStatus:     true,
		ClearedSessionProfiles: 2,
		ClearedSessionTargets:  1,
	})

	if payload.ClearSessionDecision != "session_route_cleared" || !payload.ClearSessionReady {
		t.Fatalf("expected session action outcome helper to delegate clear_session decision/ready, got %#v", payload)
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.RuntimeTarget != "node" || payload.ProfileStatus.Status != "stopped" {
		t.Fatalf("expected session action outcome helper to apply clear-result profile inventory projection, got %#v", payload.ProfileStatus)
	}
	if payload.ClearedSessionProfiles != 2 || payload.ClearedSessionTargets != 1 {
		t.Fatalf("expected session action outcome helper to delegate clear_session counts, got %#v", payload)
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.RuntimeTarget != "node" {
		t.Fatalf("expected session action outcome helper to delegate clear-session profile projection, got %#v", payload.ProfileStatus)
	}
}

func TestBrowserRuntimeDispatchClearSessionActionAppliesUnsupportedStatus(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	handled := browserRuntimeDispatchClearSessionAction(
		context.Background(),
		&payload,
		browserRuntimeClearDispatchOptions{
			Action:       "clear_target",
			Capabilities: BrowserCapabilities{},
		},
	)

	if !handled {
		t.Fatalf("expected clear dispatch helper to handle clear_target action")
	}
	if payload.Status != "unsupported" || payload.Note != "browser_runtime: selected route does not support action clear_target" {
		t.Fatalf("expected clear dispatch helper to apply unsupported status/note, got %#v", payload)
	}
}

func TestBrowserRuntimeDispatchClearSessionActionRunsClearSessionResult(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-dispatch-helper")
	sessionRegistry.TrackCurrentTarget("browser-runtime-clear-dispatch-helper", BrowserSessionTarget{
		ID:       "target-1",
		TabIndex: 1,
		URL:      "https://93.184.216.34/workbench",
		Title:    "Workbench",
		Backend:  "proxy",
		Profile:  "workbench",
		Target:   "node",
	}, "tracked_active_tab")
	stateRegistry.RecordBrowserProfileState("browser-runtime-clear-dispatch-helper", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	payload := browserRuntimePayload{Status: "ok"}

	handled := browserRuntimeDispatchClearSessionAction(
		callCtx,
		&payload,
		browserRuntimeClearDispatchOptions{
			Action:          "clear_session",
			Capabilities:    BrowserCapabilities{RuntimeClearSession: true},
			SessionRegistry: sessionRegistry,
			StateRegistry:   stateRegistry,
			SelectedInfo: BrowserRuntimeInfo{
				Backend: "proxy",
				Profile: "workbench",
				Target:  "node",
			},
			SelectedRoute: &browserRuntimeRouteDescriptor{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
			},
		},
	)

	if !handled {
		t.Fatalf("expected clear dispatch helper to handle clear_session action")
	}
	if payload.Status != "ok" || payload.ClearSessionDecision != "session_route_cleared" || !payload.ClearSessionReady {
		t.Fatalf("expected clear dispatch helper to apply clear_session result, got %#v", payload)
	}
	if payload.ClearedSessionProfiles != 1 || payload.ClearedSessionTargets != 1 {
		t.Fatalf("expected clear dispatch helper to preserve session cleanup counts, got %#v", payload)
	}
}

func TestBrowserRuntimeApplyActionSurfaceAppliesInspectionSessionProjection(t *testing.T) {
	payload := browserRuntimePayload{Action: "sessions", Status: "ok"}

	browserRuntimeApplySharedActionSurface(
		context.Background(),
		&payload,
		browserRuntimeInspectionActionSurface(
			"sessions",
			agentxbrowserruntime.SharedSessionBrowserInspectionProjection{
				Note:           "missing tool session context",
				HasSessionView: true,
				SessionProjection: agentxbrowserruntime.SharedSessionBrowserTopLevelSessionProjection{
					Routes: []agentxbrowserruntime.SharedSessionBrowserRouteSnapshot{
						{
							Backend:       "proxy",
							Profile:       "workbench",
							RuntimeTarget: "node",
						},
					},
					TargetCount: 1,
				},
			},
			BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		),
	)

	if payload.Note != "missing tool session context" {
		t.Fatalf("expected action surface helper to preserve inspection note, got %#v", payload.Note)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].Backend != "proxy" || payload.SessionRoutes[0].Profile != "workbench" || payload.SessionRoutes[0].RuntimeTarget != "node" {
		t.Fatalf("expected action surface helper to apply inspection session routes, got %#v", payload.SessionRoutes)
	}
	if payload.SessionTargetCount != 1 {
		t.Fatalf("expected action surface helper to apply inspection session target count, got %#v", payload.SessionTargetCount)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected action surface helper to apply inspection configured profiles, got %#v", payload.ConfiguredProfiles)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "inspection" ||
		payload.Summary.SummaryCode != "sessions_completed" ||
		payload.Summary.PrimaryBrowserAction != "browser action=sessions" ||
		payload.Summary.NextStep != "browser action=sessions" {
		t.Fatalf("expected action surface helper to build session inspection summary, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "inspection" ||
		payload.Display.SummaryCode != "sessions_completed" ||
		payload.Display.PrimaryBrowserAction != "browser action=sessions" ||
		payload.Display.NextStep != "browser action=sessions" {
		t.Fatalf("expected action surface helper to build session inspection display, got %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "inspection" ||
		payload.Surface.SummaryCode != "sessions_completed" ||
		payload.Surface.PrimaryBrowserAction != "browser action=sessions" ||
		payload.Surface.NextStep != "browser action=sessions" {
		t.Fatalf("expected action surface helper to build session inspection surface, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Category != "inspection" ||
		payload.View.SummaryCode != "sessions_completed" ||
		payload.View.PrimaryBrowserAction != "browser action=sessions" ||
		payload.View.NextStep != "browser action=sessions" {
		t.Fatalf("expected action surface helper to build session inspection view, got %#v", payload.View)
	}
}

func TestBrowserRuntimeApplyActionSurfaceAppliesDegradedSessionProjection(t *testing.T) {
	payload := browserRuntimePayload{}

	browserRuntimeApplySharedActionSurface(
		context.Background(),
		&payload,
		browserRuntimeDegradedActionSurface("sessions", browserRuntimeDegradedRouteProjection{
			Route: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			SessionRoutes: []browserRuntimeSessionRoute{
				{
					Backend:       "proxy",
					Profile:       "workbench",
					RuntimeTarget: "node",
				},
			},
			SessionTargetCount: 2,
			SessionRuns: []browserRuntimeSessionRun{
				{RunID: "run-1", Status: "running"},
			},
			SessionProfiles: []browserRuntimeProfileState{
				{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", Status: "running"},
			},
		}),
	)

	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].Backend != "proxy" || payload.SessionRoutes[0].Profile != "workbench" || payload.SessionRoutes[0].RuntimeTarget != "node" {
		t.Fatalf("expected action surface helper to apply degraded session routes, got %#v", payload.SessionRoutes)
	}
	if payload.SessionTargetCount != 2 || len(payload.SessionRuns) != 1 || payload.SessionRuns[0].RunID != "run-1" {
		t.Fatalf("expected action surface helper to apply degraded session inventory, got %#v", payload)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected action surface helper to apply degraded configured profiles, got %#v", payload.ConfiguredProfiles)
	}
}

func TestBrowserRuntimeApplyActionSurfaceAppliesWorkbenchInspectionConfiguredProfilesWithoutProfileInventory(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
	}

	browserRuntimeApplySharedActionSurface(
		context.Background(),
		&payload,
		browserRuntimeInspectionActionSurface(
			"workbench",
			agentxbrowserruntime.SharedSessionBrowserInspectionProjection{
				HasSessionView: true,
				SessionProjection: agentxbrowserruntime.SharedSessionBrowserTopLevelSessionProjection{
					Routes: []agentxbrowserruntime.SharedSessionBrowserRouteSnapshot{
						{
							Backend:       "proxy",
							Profile:       "workbench",
							RuntimeTarget: "node",
						},
					},
					TargetCount: 1,
					Profiles: []agentxbrowserruntime.SharedSessionBrowserProjectedProfileState{{
						State: agentxbrowserruntime.SharedSessionBrowserProfileState{
							Backend:       "proxy",
							Profile:       "workbench",
							RuntimeTarget: "node",
							BrowserApp:    "Chromium",
							Status:        "running",
							Running:       true,
							Connected:     true,
							Note:          "cached route-scoped session snapshot",
						},
						Selected: true,
					}},
				},
			},
			BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		),
	)

	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected workbench inspection surface to refresh configured profiles from selected route, got %#v", payload.ConfiguredProfiles)
	}
	if payload.ProfileStatus == nil ||
		payload.ProfileStatus.Profile != "workbench" ||
		payload.ProfileStatus.RuntimeTarget != "node" ||
		payload.ProfileStatus.BrowserApp != "Chromium" ||
		payload.ProfileStatus.Status != "running" ||
		payload.ProfileStatus.Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected workbench inspection surface to backfill profile status from session projection fallback, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 1 ||
		payload.Profiles[0].Profile != "workbench" ||
		payload.Profiles[0].RuntimeTarget != "node" ||
		payload.Profiles[0].BrowserApp != "Chromium" ||
		payload.Profiles[0].Status != "running" ||
		payload.Profiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected workbench inspection surface to backfill profiles from session projection fallback, got %#v", payload.Profiles)
	}
}

func TestBrowserRuntimeApplyPrepareResultSurfaceAppliesInventoryAndCleanup(t *testing.T) {
	payload := browserRuntimePayload{Action: "prepare", Status: "ok"}

	browserRuntimeApplyPrepareResultSurface(
		context.Background(),
		&payload,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		agentxbrowserruntime.BuildSharedSessionBrowserExecutionSurfaceProjection(
			BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			browserRuntimePrepareResult{
				Profiles: &BrowserProfilesResult{
					DefaultProfile: "workbench",
					Profiles: []BrowserProfileInfo{
						{Profile: "workbench", Status: "running"},
						{Profile: "relay", Status: "stopped"},
					},
					Note: "profiles synced",
				},
			},
			agentxbrowserruntime.SharedSessionBrowserExecutionApplication{
				Resolution: agentxbrowserruntime.SharedSessionBrowserExecutionResolution{
					SyncedState: agentxbrowserruntime.SharedSessionBrowserProfileState{
						Backend:       "proxy",
						Profile:       "workbench",
						RuntimeTarget: "node",
						Status:        "running",
						Running:       true,
						Connected:     true,
					},
					HasSyncedState: true,
				},
				Cleanup: agentxbrowserruntime.SharedSessionBrowserExecutionCleanup{
					ClearedSessionTargets: 2,
				},
				ProjectedProfiles: []agentxbrowserruntime.SharedSessionBrowserProjectedProfileState{
					{State: agentxbrowserruntime.SharedSessionBrowserProfileState{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", Status: "running"}},
					{State: agentxbrowserruntime.SharedSessionBrowserProfileState{Backend: "proxy", Profile: "relay", RuntimeTarget: "node", Status: "stopped"}},
				},
			},
		),
	)

	if payload.Note != "profiles synced" {
		t.Fatalf("expected prepare-result surface helper to preserve profiles note, got %#v", payload.Note)
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.RuntimeTarget != "node" || payload.ProfileStatus.Status != "running" {
		t.Fatalf("expected prepare-result surface helper to apply synced profile status, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 2 || payload.Profiles[0].Profile != "workbench" || payload.Profiles[1].Profile != "relay" {
		t.Fatalf("expected prepare-result surface helper to apply projected profiles, got %#v", payload.Profiles)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected prepare-result surface helper to preserve default profile, got %#v", payload.DefaultProfile)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") || !browserStringSliceContains(payload.ConfiguredProfiles, "relay") {
		t.Fatalf("expected prepare-result surface helper to refresh configured profiles, got %#v", payload.ConfiguredProfiles)
	}
	if payload.ClearedSessionTargets != 2 {
		t.Fatalf("expected prepare-result surface helper to preserve cleanup counts, got %#v", payload.ClearedSessionTargets)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "coordination" ||
		payload.Summary.SummaryCode != "prepare_completed" ||
		payload.Summary.PrimaryBrowserAction != "browser action=prepare" ||
		payload.Summary.NextStep != "browser action=prepare" {
		t.Fatalf("expected prepare-result surface helper to build stable summary, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "coordination" ||
		payload.Display.SummaryCode != "prepare_completed" ||
		payload.Display.PrimaryBrowserAction != "browser action=prepare" ||
		payload.Display.NextStep != "browser action=prepare" {
		t.Fatalf("expected prepare-result surface helper to build stable display, got %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "coordination" ||
		payload.Surface.SummaryCode != "prepare_completed" ||
		payload.Surface.PrimaryBrowserAction != "browser action=prepare" ||
		payload.Surface.NextStep != "browser action=prepare" {
		t.Fatalf("expected prepare-result surface helper to build stable surface, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Category != "coordination" ||
		payload.View.SummaryCode != "prepare_completed" ||
		payload.View.PrimaryBrowserAction != "browser action=prepare" ||
		payload.View.NextStep != "browser action=prepare" {
		t.Fatalf("expected prepare-result surface helper to build stable view, got %#v", payload.View)
	}
}

func TestBrowserRuntimeApplySessionActionOutcomeAppliesRememberProjection(t *testing.T) {
	payload := browserRuntimePayload{
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	browserRuntimeApplySessionActionOutcome(
		&payload,
		agentxbrowserruntime.BuildSharedSessionBrowserRememberActionOutcome(
			agentxbrowserruntime.SharedSessionBrowserRememberProfileResult{
				Decision: "session_profile_remembered",
				Ready:    true,
				SelectionProjection: &agentxbrowserruntime.SharedSessionBrowserSelectionProjection{
					ProfileSelection: &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
						Backend:       "proxy",
						Profile:       "alternate",
						RuntimeTarget: "node",
						Source:        "remember_profile",
					},
					TargetSelection: &agentxbrowserruntime.BrowserSessionTargetSelection{
						ID:            "target-2",
						TabIndex:      2,
						Backend:       "proxy",
						Profile:       "alternate",
						RuntimeTarget: "node",
						Source:        "remember_profile",
					},
					ApplyTargetToRoute: true,
				},
			},
		),
	)

	if payload.RememberDecision != "session_profile_remembered" || !payload.RememberReady {
		t.Fatalf("expected shared remember action outcome to populate remember decision/ready, got %#v", payload)
	}
	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "alternate" || payload.SessionProfileSelection.Source != "remember_profile" {
		t.Fatalf("expected shared remember action outcome to apply profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != "target-2" || payload.SessionTargetSelection.Source != "remember_profile" {
		t.Fatalf("expected shared remember action outcome to apply target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SelectedRoute == nil || payload.SelectedRoute.Profile != "alternate" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected shared remember action outcome to apply target-backed route mutation, got %#v", payload.SelectedRoute)
	}
}

func TestBrowserRuntimeApplyLifecycleActionOutcomeAppliesRememberOutcome(t *testing.T) {
	payload := browserRuntimePayload{
		PreparedProfile: "workbench",
	}

	browserRuntimeApplyLifecycleActionOutcome(
		context.Background(),
		"",
		&payload,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		agentxbrowserruntime.SharedSessionBrowserLifecycleActionOutcome{
			Action:          "restart",
			PreparedProfile: "workbench",
			RestartDecision: "restart_started",
			Result: browserRuntimePrepareResult{
				Profile:  "workbench",
				Decision: "restart_started",
				Ready:    false,
			},
			RememberOutcome: func() *agentxbrowserruntime.SharedSessionBrowserActionOutcome {
				outcome := agentxbrowserruntime.BuildSharedSessionBrowserRememberActionOutcome(
					agentxbrowserruntime.SharedSessionBrowserRememberProfileResult{
						Decision: "session_profile_remembered",
						Ready:    true,
						SelectionProjection: &agentxbrowserruntime.SharedSessionBrowserSelectionProjection{
							ProfileSelection: &agentxbrowserruntime.SharedSessionBrowserProfileSelection{
								Backend:       "proxy",
								Profile:       "workbench",
								RuntimeTarget: "node",
								BrowserApp:    "Chromium",
								Source:        "remember_profile",
							},
							TargetSelection: &agentxbrowserruntime.BrowserSessionTargetSelection{
								ID:            "target-1",
								Backend:       "proxy",
								Profile:       "workbench",
								RuntimeTarget: "node",
								BrowserApp:    "Chromium",
								Source:        "remember_profile",
							},
							ApplyTargetToRoute: true,
						},
					},
				)
				return &outcome
			}(),
		},
	)

	if payload.RememberDecision != "session_profile_remembered" || !payload.RememberReady {
		t.Fatalf("expected lifecycle action outcome to apply shared remember decision/ready, got %#v", payload)
	}
	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.Source != "remember_profile" {
		t.Fatalf("expected lifecycle action outcome to apply session profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != "target-1" || payload.SessionTargetSelection.Source != "remember_profile" {
		t.Fatalf("expected lifecycle action outcome to apply remembered target selection, got %#v", payload.SessionTargetSelection)
	}
}

func TestApplyBrowserRememberTargetSelectionUsesSharedDispatch(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-remember-target-apply-helper")
	tracked := sessionRegistry.TrackTabs("browser-remember-target-apply-helper", []agentxbrowserruntime.BrowserSessionTarget{{
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}}, 2)
	if len(tracked) != 1 || strings.TrimSpace(tracked[0].ID) == "" {
		t.Fatalf("expected one tracked target for remember-target apply helper, got %#v", tracked)
	}

	result := applyBrowserRememberTargetSelection(
		callCtx,
		browserRegistrationContext{
			sessionRegistry:      sessionRegistry,
			sessionStateRegistry: stateRegistry,
			watchManagerProvider: agentxbrowserruntime.SharedSessionBrowserObserverManager{
				SessionRegistry: sessionRegistry,
				StateRegistry:   stateRegistry,
			},
		},
		browserRememberTargetApplyOptions{
			Route: browserSessionRoute(
				BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
				"Chromium",
				"proxy",
			),
			TargetID: tracked[0].ID,
		},
	)
	if result.Selection == nil || result.Selection.ID != tracked[0].ID || !result.Ready {
		t.Fatalf("expected remember-target apply helper to remember target selection, got %#v", result)
	}
	if result.Decision != "session_target_remembered" && result.Decision != "session_target_already_remembered" {
		t.Fatalf("unexpected remember-target apply helper decision: %#v", result)
	}
	if result.ProfileSelection == nil || result.ProfileSelection.Profile != "workbench" || result.ProfileSelection.RuntimeTarget != "node" || result.ProfileSelection.Source != "remember_target" {
		t.Fatalf("expected remember-target apply helper to promote profile selection, got %#v", result.ProfileSelection)
	}
}

func TestBrowserRuntimeApplySessionActionOutcomeAppliesReviewRequiredStatus(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	browserRuntimeApplySessionActionOutcome(&payload, agentxbrowserruntime.SharedSessionBrowserActionOutcome{
		ApplyDecision:        true,
		Action:               "select_target",
		Decision:             "session_target_popup_review_required",
		SelectTargetDecision: "session_target_popup_review_required",
		Status:               "review_required",
		Note:                 "pending popup target",
	})

	if payload.SelectTargetDecision != "session_target_popup_review_required" || payload.SelectTargetReady {
		t.Fatalf("expected session action outcome helper to keep review-required decision and ready=false, got %#v", payload)
	}
	if payload.Status != "review_required" || payload.Note != "pending popup target" {
		t.Fatalf("expected session action outcome helper to apply review-required status/note, got %#v", payload)
	}
}

func TestBrowserRuntimeApplyMissingSessionActionInput(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	browserRuntimeApplyMissingSessionActionInput(
		&payload,
		"select_profile",
		"session_profile_required",
	)

	if payload.SelectDecision != "session_profile_required" || payload.SelectReady {
		t.Fatalf("expected missing-session-input helper to preserve decision and ready=false, got %#v", payload)
	}
	if payload.Status != "error" || payload.Note != "browser_runtime: profile is required for action select_profile" {
		t.Fatalf("expected missing-session-input helper to apply error status/note, got %#v", payload)
	}
}

func TestBrowserRuntimeApplyMissingSelectTargetSelection(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	browserRuntimeApplyMissingSelectTargetSelection(&payload, "session_target_popup_review_required", "pending popup target")

	if payload.SelectTargetDecision != "session_target_popup_review_required" || payload.SelectTargetReady {
		t.Fatalf("expected missing-select-target helper to preserve review-required decision and ready=false, got %#v", payload)
	}
	if payload.Status != "review_required" || payload.Note != "pending popup target" {
		t.Fatalf("expected missing-select-target helper to apply review-required status/note, got %#v", payload)
	}
}

func TestBrowserRuntimeDispatchSessionSelectionActionAppliesInvalidSelectTargetDecision(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-invalid-select-target")
	payload := browserRuntimePayload{
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	handled := browserRuntimeDispatchSessionSelectionAction(
		callCtx,
		&payload,
		browserRuntimeSessionSelectionDispatchOptions{
			Action:       "select_target",
			Capabilities: BrowserCapabilities{RuntimeSelectTarget: true},
			SelectedInfo: BrowserRuntimeInfo{
				Backend: "proxy",
				Profile: "workbench",
				Target:  "node",
			},
			SelectedRoute: payload.SelectedRoute,
			Params: map[string]any{
				"target": "tab:not-a-number",
			},
		},
	)

	if !handled {
		t.Fatalf("expected selection dispatch helper to handle invalid select_target input")
	}
	if payload.SelectTargetDecision != "session_target_invalid" || payload.SelectTargetReady {
		t.Fatalf("expected selection dispatch helper to apply canonical invalid-target decision, got %#v", payload)
	}
	if payload.Status != "error" || !strings.Contains(payload.Note, "target must be") {
		t.Fatalf("expected selection dispatch helper to apply shared invalid-target error terminal, got %#v", payload)
	}
}

func TestBrowserRuntimeApplyUnsupportedActionOutcome_Inspection(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	browserRuntimeApplyUnsupportedActionOutcome(&payload, "profiles")

	if payload.Status != "unsupported" || payload.Note != "browser_runtime: selected route does not support action profiles" {
		t.Fatalf("expected unsupported action helper to populate status/note, got %#v", payload)
	}
}

func TestBrowserRuntimeDispatchInspectionActionAppliesUnsupportedProfilesStatus(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	bindingEvaluation := browserRuntimeDispatchInspectionAction(
		browserRegistrationContext{},
		context.Background(),
		&payload,
		&fakeBrowserBackend{},
		agentxbrowserruntime.SharedSessionBrowserWatchManager{},
		browserRuntimeInspectionActionOptions{
			Action:       "profiles",
			Capabilities: BrowserCapabilities{},
		},
	)

	if bindingEvaluation != nil {
		t.Fatalf("expected unsupported inspection dispatch helper not to return binding evaluation, got %#v", bindingEvaluation)
	}
	if payload.Status != "unsupported" || payload.Note != "browser_runtime: selected route does not support action profiles" {
		t.Fatalf("expected inspection dispatch helper to apply unsupported status/note, got %#v", payload)
	}
}

func TestBrowserRuntimeApplyUnsupportedActionOutcome_Lifecycle(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	browserRuntimeApplyUnsupportedActionOutcome(&payload, "refresh")

	if payload.Status != "unsupported" || payload.Note != "browser_runtime: selected route does not support action refresh" {
		t.Fatalf("expected unsupported action helper to populate status/note, got %#v", payload)
	}
}

func TestBrowserRuntimeDispatchLifecycleActionAppliesUnsupportedStartStatus(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	handled := browserRuntimeDispatchLifecycleAction(
		context.Background(),
		&payload,
		browserRuntimeLifecycleDispatchOptions{
			Action:          "start",
			Capabilities:    BrowserCapabilities{},
			SelectedBackend: &fakeBrowserBackend{},
		},
	)

	if !handled {
		t.Fatalf("expected lifecycle dispatch helper to handle start action")
	}
	if payload.Status != "unsupported" || payload.Note != "browser_runtime: selected route does not support action start" {
		t.Fatalf("expected lifecycle dispatch helper to apply unsupported status/note, got %#v", payload)
	}
}

func TestBrowserRuntimeDispatchLifecycleActionRunsStartResult(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStartResult: BrowserProfileStatusResult{
					Backend:   "proxy",
					Profile:   "isolated",
					Status:    "started",
					Running:   true,
					Connected: false,
				},
			},
		},
	}

	handled := browserRuntimeDispatchLifecycleAction(
		context.Background(),
		&payload,
		browserRuntimeLifecycleDispatchOptions{
			Action:           "start",
			Capabilities:     BrowserCapabilities{RuntimeStart: true},
			SelectedBackend:  control,
			EffectiveProfile: "isolated",
			SelectedInfo: BrowserRuntimeInfo{
				Backend: "proxy",
				Profile: "isolated",
				Target:  "node",
			},
		},
	)

	if !handled {
		t.Fatalf("expected lifecycle dispatch helper to handle start action")
	}
	if payload.Status != "ok" {
		t.Fatalf("expected lifecycle dispatch helper to preserve ok status on successful start, got %#v", payload)
	}
	if payload.PrepareDecision != "" || payload.PrepareReady {
		t.Fatalf("expected start helper to preserve existing start contract without prepare decision fields, got %#v", payload)
	}
	if payload.PreparedProfile != "isolated" {
		t.Fatalf("expected lifecycle dispatch helper to populate prepared profile, got %#v", payload)
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "starting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected lifecycle dispatch helper to apply lifecycle-owned profile status, got %#v", payload.ProfileStatus)
	}
}

func TestBrowserRuntimeFinalizeActionSessionPayloadProjectsRuntimeActionSuccessAliases(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-action-session-finalize-success-aliases")

	selectTargetPayload := browserRuntimePayload{
		Action: "select_target",
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		SessionTargetSelection: &browserRuntimeSessionTargetSelection{
			ID:            "target-1",
			RuntimeTarget: "node",
			Source:        "select_target",
		},
	}
	browserRuntimeFinalizeActionSessionPayload(browserRegistrationContext{}, callCtx, &selectTargetPayload, browserRuntimeActionSessionResultPostProcess{
		Action: "select_target",
	})
	if selectTargetPayload.Summary == nil ||
		selectTargetPayload.Summary.Category != "coordination" ||
		selectTargetPayload.Summary.State != "completed" ||
		selectTargetPayload.Summary.SummaryCode != "select_target_completed" ||
		selectTargetPayload.Summary.PrimaryBrowserAction != "browser action=select_target" ||
		selectTargetPayload.Summary.NextStep != "browser action=select_target" {
		t.Fatalf("expected session finalize helper to project select_target success summary, got %#v", selectTargetPayload.Summary)
	}
	if selectTargetPayload.Display == nil ||
		selectTargetPayload.Display.Category != "coordination" ||
		selectTargetPayload.Display.SummaryCode != "select_target_completed" {
		t.Fatalf("expected session finalize helper to project select_target display, got %#v", selectTargetPayload.Display)
	}
	if selectTargetPayload.Surface == nil ||
		selectTargetPayload.Surface.Category != "coordination" ||
		selectTargetPayload.Surface.SummaryCode != "select_target_completed" {
		t.Fatalf("expected session finalize helper to project select_target surface, got %#v", selectTargetPayload.Surface)
	}
	if selectTargetPayload.View == nil ||
		selectTargetPayload.View.Category != "coordination" ||
		selectTargetPayload.View.SummaryCode != "select_target_completed" {
		t.Fatalf("expected session finalize helper to project select_target view, got %#v", selectTargetPayload.View)
	}

	startPayload := browserRuntimePayload{
		Action: "start",
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
		},
		PreparedProfile: "isolated",
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			Status:        "starting",
			Running:       true,
			Connected:     false,
		},
	}
	browserRuntimeFinalizeActionSessionPayload(browserRegistrationContext{}, callCtx, &startPayload, browserRuntimeActionSessionResultPostProcess{
		Action: "start",
	})
	if startPayload.Summary == nil ||
		startPayload.Summary.Category != "coordination" ||
		startPayload.Summary.State != "completed" ||
		startPayload.Summary.SummaryCode != "start_completed" ||
		startPayload.Summary.PrimaryBrowserAction != "browser action=start" ||
		startPayload.Summary.NextStep != "browser action=start" {
		t.Fatalf("expected lifecycle finalize helper to project start success summary, got %#v", startPayload.Summary)
	}
	if startPayload.Display == nil ||
		startPayload.Display.Category != "coordination" ||
		startPayload.Display.SummaryCode != "start_completed" {
		t.Fatalf("expected lifecycle finalize helper to project start display, got %#v", startPayload.Display)
	}
	if startPayload.Surface == nil ||
		startPayload.Surface.Category != "coordination" ||
		startPayload.Surface.SummaryCode != "start_completed" {
		t.Fatalf("expected lifecycle finalize helper to project start surface, got %#v", startPayload.Surface)
	}
	if startPayload.View == nil ||
		startPayload.View.Category != "coordination" ||
		startPayload.View.SummaryCode != "start_completed" {
		t.Fatalf("expected lifecycle finalize helper to project start view, got %#v", startPayload.View)
	}
}

func TestBrowserRuntimeDispatchCoordinateActionAppliesUnsupportedCoordinateStatus(t *testing.T) {
	payload := browserRuntimePayload{Status: "ok"}

	handled := browserRuntimeDispatchCoordinateAction(
		context.Background(),
		&payload,
		browserRuntimeCoordinateDispatchOptions{
			Action:           "coordinate",
			CoordinationGoal: "restart",
			Capabilities:     BrowserCapabilities{},
			SelectedBackend:  &fakeBrowserBackend{},
		},
	)

	if !handled {
		t.Fatalf("expected coordinate dispatch helper to handle coordinate action")
	}
	if payload.Status != "unsupported" || payload.Note != "browser_runtime: selected route does not support action coordinate" {
		t.Fatalf("expected coordinate dispatch helper to apply unsupported status/note, got %#v", payload)
	}
}

func TestBrowserRuntimeDispatchCoordinateActionRunsCoordinateSyncResult(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-dispatch-helper")
	tracked := sessionRegistry.TrackTabs("browser-runtime-coordinate-dispatch-helper", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://node.example/workbench",
			Title:      "Workbench",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
	}, 1)
	if len(tracked) != 1 || strings.TrimSpace(tracked[0].ID) == "" {
		t.Fatalf("expected one tracked target for coordinate dispatch helper, got %#v", tracked)
	}

	payload := browserRuntimePayload{
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStartResult: BrowserProfileStatusResult{
					Backend:    "proxy",
					Profile:    "workbench",
					Status:     "started",
					Running:    true,
					Connected:  true,
					BrowserApp: "Chromium",
				},
			},
		},
	}

	handled := browserRuntimeDispatchCoordinateAction(
		callCtx,
		&payload,
		browserRuntimeCoordinateDispatchOptions{
			Action:               "coordinate",
			CoordinationGoal:     "sync",
			Capabilities:         BrowserCapabilities{RuntimeCoordinate: true},
			SelectedBackend:      control,
			EffectiveProfile:     "workbench",
			SelectedInfo:         BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			SelectedRoute:        payload.SelectedRoute,
			WatchManagerProvider: agentxbrowserruntime.SharedSessionBrowserObserverManager{SessionRegistry: sessionRegistry},
			SessionRegistry:      sessionRegistry,
			RequestedBrowserApp:  "Chromium",
		},
	)

	if !handled {
		t.Fatalf("expected coordinate dispatch helper to handle coordinate action")
	}
	if payload.Status != "ok" || payload.PrepareDecision != "started" || !payload.PrepareReady {
		t.Fatalf("expected coordinate dispatch helper to apply lifecycle prepare result before sync projection, got %#v", payload)
	}
	if payload.SyncSessionDecision != "session_target_synced" || !payload.SyncSessionReady {
		t.Fatalf("expected coordinate dispatch helper to apply sync result, got %#v", payload)
	}
	if payload.SessionProfileSelection != nil {
		t.Fatalf("expected coordinate dispatch helper to preserve target-first sync contract, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != tracked[0].ID || payload.SessionTargetSelection.Source != "sync_session" {
		t.Fatalf("expected coordinate dispatch helper to project tracked target selection, got %#v", payload.SessionTargetSelection)
	}
}
