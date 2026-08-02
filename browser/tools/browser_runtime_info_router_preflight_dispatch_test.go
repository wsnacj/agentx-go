package tools

import (
	"context"
	"runtime"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestBrowserRuntimeRouterDirectPageActionPreflightRequestUsesManagedDispatchBaseForImplicitLegacyHost(t *testing.T) {
	router := browserRuntimeRouterBackend{
		hostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}

	got, err := router.directPageActionPreflightRequest(context.Background(), "", 0, "")
	if err != nil {
		t.Fatalf("directPageActionPreflightRequest(implicit legacy host): %v", err)
	}
	if got.Requested != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected implicit legacy-host current-page preflight to hide dispatch runtime info, got %#v", got.Requested)
	}
	if !got.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected implicit legacy-host current-page preflight to preserve hidden host fallback context")
	}
	if got.ExplicitRuntimeTarget {
		t.Fatalf("expected implicit legacy-host current-page preflight to stay targetless")
	}
}

func TestBrowserRuntimeRouterDirectPageActionPreflightRequestKeepsDefaultRuntimeOutsideImplicitLegacyHost(t *testing.T) {
	router := browserRuntimeRouterBackend{
		hostRuntime: BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
	}

	got, err := router.directPageActionPreflightRequest(context.Background(), "", 0, "")
	if err != nil {
		t.Fatalf("directPageActionPreflightRequest(explicit host runtime): %v", err)
	}
	if got.Requested != (BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"}) {
		t.Fatalf("expected non-legacy host preflight to retain default runtime info, got %#v", got.Requested)
	}
	if got.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected non-legacy host preflight to clear hidden host fallback context")
	}
	if got.ExplicitRuntimeTarget {
		t.Fatalf("expected non-legacy host preflight without params to stay implicit")
	}
}

func TestBrowserRuntimeRouterDirectURLActionPreflightRequestUsesHiddenBaseForImplicitLegacyHostURL(t *testing.T) {
	router := browserRuntimeRouterBackend{
		hostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}

	got, err := router.directURLActionPreflightRequest(context.Background(), "", 0, "https://93.184.216.34")
	if err != nil {
		t.Fatalf("directURLActionPreflightRequest(implicit legacy host URL): %v", err)
	}
	if got.Requested != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected implicit legacy-host URL preflight to hide dispatch runtime info, got %#v", got.Requested)
	}
	if !got.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected implicit legacy-host URL preflight to preserve hidden host fallback context")
	}
	if got.Target != (browserToolTarget{}) {
		t.Fatalf("expected open-style URL preflight to stay targetless, got %#v", got.Target)
	}
}

func TestBrowserRuntimeRouterDirectTabsActionPreflightRequestPrefersImplicitManagedDefaultRoute(t *testing.T) {
	nodeBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}
	router, ok := newBrowserRuntimeRouterBackendWithAssessment(
		BrowserToolOptions{NodeBackend: nodeBackend},
		&runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		browserDefaultSubstrateAssessment{
			HostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			HostRoute:   browserImplicitLegacyHostRouteAssessment(nil, BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}),
			NodeRoute: browserDefaultPromotionRouteAssessmentForConcreteRouteAssessment(browserConcreteRouteAssessment{
				Configured:     true,
				RouteAvailable: true,
				Route: browserResolvedExecutionRoute{
					Backend:     nodeBackend,
					RuntimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
				},
			}, "node"),
			DefaultRuntime:       BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			DefaultConcreteRoute: browserConcreteRouteAssessment{},
		},
	).(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected concrete browser runtime router backend")
	}

	got, err := router.directTabsActionPreflightRequest(context.Background(), BrowserTabsRequest{Action: "focus", TabIndex: 2})
	if err != nil {
		t.Fatalf("directTabsActionPreflightRequest(managed default focus): %v", err)
	}
	if got.Requested.Target != "node" {
		t.Fatalf("expected tabs preflight to prefer managed default route on implicit legacy host, got %#v", got.Requested)
	}
	if got.Target.TabIndex != 2 {
		t.Fatalf("expected tabs preflight to preserve tab target, got %#v", got.Target)
	}
}

func TestBrowserRuntimeRouterExecutionPreviewHidesImplicitLegacyHostFallback(t *testing.T) {
	router := browserRuntimeRouterBackend{
		hostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}

	preview := router.executionPreview()
	if preview.DefaultRoute != (BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}) {
		t.Fatalf("expected router execution preview logical default route to remain legacy host, got %#v", preview.DefaultRoute)
	}
	if preview.DispatchBase != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected router execution preview dispatch base to hide implicit legacy host fallback, got %#v", preview.DispatchBase)
	}
	if preview.DefaultCandidateRoute != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected router execution preview to keep default candidate route empty without managed lanes, got %#v", preview.DefaultCandidateRoute)
	}
	if !preview.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected router execution preview to mark hidden implicit host default base")
	}
	if preview.DefaultTarget != "host" {
		t.Fatalf("expected router execution preview default target to remain host, got %q", preview.DefaultTarget)
	}
}

func TestBrowserRuntimeRouterExecutionPreviewUsesDynamicManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       fullBrowserCapabilities(),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			if strings.EqualFold(strings.TrimSpace(requested.Profile), "isolated") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(
		BrowserToolOptions{NodeBackend: node},
		&runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		browserDefaultSubstrateAssessment{
			HostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			HostRoute:   browserImplicitLegacyHostRouteAssessment(nil, BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}),
			NodeRoute: browserDefaultPromotionRouteAssessment{
				Configured: true,
			},
			DefaultRuntime:       BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			DefaultConcreteRoute: browserConcreteRouteAssessment{},
		},
	).(browserRuntimeRouterBackend)

	preview := router.executionPreview()
	want := BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}
	if preview.DefaultRoute != want || preview.DispatchBase != want {
		t.Fatalf("expected router execution preview to seed dynamic managed-default route, got %#v", preview)
	}
	if preview.DefaultCandidateRoute != want {
		t.Fatalf("expected router execution preview to keep managed default candidate route, got %#v want %#v", preview.DefaultCandidateRoute, want)
	}
	if preview.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected router execution preview to stop hiding host fallback after seeding managed default")
	}
	if preview.DefaultTarget != "node" {
		t.Fatalf("expected router execution preview default target to switch to node, got %q", preview.DefaultTarget)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected router execution preview to seed generic managed-default once, got %d resolves", node.resolveCalls)
	}
}

func TestBrowserRuntimeRouterExecutionPreviewKeepsPromotedConcreteDefaultRoute(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}
	router := newBrowserRuntimeRouterBackend(BrowserToolOptions{NodeBackend: node}, host).(browserRuntimeRouterBackend)

	preview := router.executionPreview()
	want := BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}
	if preview.DefaultRoute != want || preview.DispatchBase != want {
		t.Fatalf("expected router execution preview to keep promoted concrete default route, got %#v", preview)
	}
	if preview.DefaultCandidateRoute != want {
		t.Fatalf("expected router execution preview to keep promoted default candidate route, got %#v want %#v", preview.DefaultCandidateRoute, want)
	}
	if preview.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected router execution preview hidden implicit host default base to stay false for promoted default route")
	}
	if preview.DefaultTarget != "node" {
		t.Fatalf("expected router execution preview default target to track promoted route, got %q", preview.DefaultTarget)
	}
}

func TestBrowserRuntimeRouterExecutionPreviewSurfacesHiddenManagedDefaultCandidateRoute(t *testing.T) {
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       fullBrowserCapabilities(),
			routeSource:        "managed_browserd",
			routeEndpoint:      "http://127.0.0.1:43123",
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(
		BrowserToolOptions{NodeBackend: node},
		&runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		browserDefaultSubstrateAssessment{
			HostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			HostRoute:   browserImplicitLegacyHostRouteAssessment(nil, BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}),
			NodeRoute: browserDefaultPromotionRouteAssessment{
				Configured: true,
			},
			DefaultRuntime:       BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			DefaultConcreteRoute: browserConcreteRouteAssessment{},
		},
	).(browserRuntimeRouterBackend)

	preview := router.executionPreview()
	if preview.DefaultRoute != DefaultBrowserRuntimeInfo() {
		t.Fatalf("expected router execution preview logical default route to remain host when managed lane stays hidden, got %#v", preview.DefaultRoute)
	}
	if preview.DispatchBase != (BrowserRuntimeInfo{}) || !preview.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected router execution preview to keep hidden implicit host dispatch base, got %#v", preview)
	}
	if preview.DefaultCandidateRoute != (BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}) {
		t.Fatalf("expected router execution preview to expose hidden managed default candidate route, got %#v", preview.DefaultCandidateRoute)
	}
	if preview.DefaultCandidateDescriptor != (browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Source:        "managed_browserd",
		Endpoint:      "http://127.0.0.1:43123",
	}) {
		t.Fatalf("expected router execution preview to preserve hidden managed default candidate provenance, got %#v", preview.DefaultCandidateDescriptor)
	}
}

func TestBrowserRuntimeRouterExecutionLaneWrapsResolvedRoute(t *testing.T) {
	backend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilities{Open: true, Navigate: true, Tabs: true},
	}
	route := browserResolvedExecutionRoute{
		Backend:      backend,
		RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		Capabilities: backend.capabilities,
	}
	lane := browserRuntimeRouterExecutionLane(route)
	if lane.Runtime != route.RuntimeInfo {
		t.Fatalf("execution lane runtime = %#v, want %#v", lane.Runtime, route.RuntimeInfo)
	}
	if lane.Backend != route.Backend {
		t.Fatalf("execution lane backend mismatch: got %T want %T", lane.Backend, route.Backend)
	}
	if lane.Capabilities != route.Capabilities {
		t.Fatalf("execution lane capabilities = %#v, want %#v", lane.Capabilities, route.Capabilities)
	}
	if lane.Substrate != BrowserSubstratePosture(route.RuntimeInfo.Backend, route.RuntimeInfo.Target) {
		t.Fatalf("execution lane substrate = %q", lane.Substrate)
	}
}

func TestBrowserRuntimeRouterPlatformLaneMatchesCurrentPlatform(t *testing.T) {
	lane := browserRuntimeRouterPlatformLane()
	current := currentBrowserPlatformLane(BrowserToolOptions{})
	if lane.Name() != current.Name() {
		t.Fatalf("router platform lane = %q, want %q", lane.Name(), current.Name())
	}
}

func TestBrowserRuntimeRouterResolveBrowserExecutionLaneWrapsRouteResolver(t *testing.T) {
	backend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilities{Open: true, Navigate: true, Tabs: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: backend,
		hostRuntime: BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		hostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      backend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
				Capabilities: backend.capabilities,
			},
		},
	}
	lane, err := router.resolveBrowserExecutionLane(BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"})
	if err != nil {
		t.Fatalf("resolveBrowserExecutionLane error = %v", err)
	}
	if lane.Runtime != (BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"}) {
		t.Fatalf("execution lane runtime = %#v", lane.Runtime)
	}
	if lane.Backend != backend {
		t.Fatalf("execution lane backend mismatch: got %T want %T", lane.Backend, backend)
	}
	if lane.Capabilities != backend.capabilities {
		t.Fatalf("execution lane capabilities = %#v, want %#v", lane.Capabilities, backend.capabilities)
	}
}

func TestBrowserRuntimeRouterResolveDirectURLActionPreflightWrapsRequestAndLane(t *testing.T) {
	backend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilities{Open: true, Navigate: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: backend,
		hostRuntime: BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		hostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      backend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
				Capabilities: backend.capabilities,
			},
		},
	}
	preflight, err := router.resolveDirectURLActionPreflight(context.Background(), "browser backend open", "", 0, "https://93.184.216.34")
	if err != nil {
		t.Fatalf("resolveDirectURLActionPreflight error = %v", err)
	}
	if preflight.Requested != (BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"}) {
		t.Fatalf("direct URL preflight requested = %#v", preflight.Requested)
	}
	if preflight.Lane.Backend != backend || preflight.Lane.Runtime != preflight.Requested {
		t.Fatalf("direct URL preflight = %#v", preflight)
	}
}

func TestBrowserRuntimeRouterResolveDirectTabsActionPreflightWrapsRequestAndLane(t *testing.T) {
	backend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilities{Tabs: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: backend,
		hostRuntime: BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		hostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      backend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
				Capabilities: backend.capabilities,
			},
		},
	}
	preflight, err := router.resolveDirectTabsActionPreflight(context.Background(), "browser backend tabs", BrowserTabsRequest{Action: "focus", TabIndex: 2})
	if err != nil {
		t.Fatalf("resolveDirectTabsActionPreflight error = %v", err)
	}
	if preflight.Target.TabIndex != 2 {
		t.Fatalf("direct tabs preflight target = %#v", preflight.Target)
	}
	if preflight.Lane.Backend != backend || preflight.Lane.Runtime != preflight.Requested {
		t.Fatalf("direct tabs preflight = %#v", preflight)
	}
}

func TestBrowserRuntimeRouterResolveDirectPageActionPreflightWrapsRequestAndLane(t *testing.T) {
	backend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilities{Extract: true, Snapshot: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: backend,
		hostRuntime: BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		hostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      backend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
				Capabilities: backend.capabilities,
			},
		},
	}
	preflight, err := router.resolveDirectPageActionPreflight(context.Background(), "browser backend extract", "", 3, "")
	if err != nil {
		t.Fatalf("resolveDirectPageActionPreflight error = %v", err)
	}
	if preflight.Requested != (BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"}) {
		t.Fatalf("direct page preflight requested = %#v", preflight.Requested)
	}
	if preflight.Target.TabIndex != 3 {
		t.Fatalf("direct page preflight target = %#v", preflight.Target)
	}
	if preflight.Lane.Backend != backend || preflight.Lane.Runtime != preflight.Requested {
		t.Fatalf("direct page preflight = %#v", preflight)
	}
}

func TestBrowserRuntimeRouterResolveDirectPreflightKeepsExplicitHostLane(t *testing.T) {
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilities{Open: true, Navigate: true},
	}
	nodeBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilities{Open: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: hostBackend,
		hostRuntime: BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		hostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      hostBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
				Capabilities: hostBackend.capabilities,
			},
		},
		nodeBackend: nodeBackend,
		nodeRoute: browserDefaultPromotionRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Ready:          true,
			Route: browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
				Capabilities: nodeBackend.capabilities,
			},
		},
		defaultRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
				Capabilities: nodeBackend.capabilities,
			},
		},
	}
	preflight, err := router.resolveDirectPreflight(context.Background(), browserRuntimeRouterDirectPreflightArgs{
		Params: map[string]any{
			"runtime_target": "host",
			"profile":        "workbench",
		},
		Base: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	})
	if err != nil {
		t.Fatalf("resolveDirectPreflight(explicit host) error = %v", err)
	}
	if !preflight.ExplicitRuntimeTarget {
		t.Fatalf("expected explicit runtime_target to be preserved, got %#v", preflight)
	}
	if preflight.Requested.Target != "host" || preflight.Requested.Profile != "workbench" {
		t.Fatalf("expected preflight request to stay on explicit host lane, got %#v", preflight.Requested)
	}
	if preflight.Lane.Backend != hostBackend || preflight.Lane.Runtime.Target != "host" {
		t.Fatalf("expected explicit host preflight to bypass promoted node lane, got %#v", preflight)
	}
}

func TestBrowserRuntimeRouterResolveDirectPageActionPreflightUsesConcreteLaneCapabilities(t *testing.T) {
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilities{Extract: true, Click: true},
	}
	nodeBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilities{Extract: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: hostBackend,
		hostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		hostRoute:   browserImplicitLegacyHostRouteAssessment(nil, BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}),
		nodeBackend: nodeBackend,
		nodeRoute: browserDefaultPromotionRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Ready:          true,
			Route: browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
				Capabilities: nodeBackend.capabilities,
			},
		},
	}
	preflight, err := router.resolveDirectPageActionPreflight(context.Background(), "browser backend extract", "", 0, "")
	if err != nil {
		t.Fatalf("resolveDirectPageActionPreflight(concrete capabilities) error = %v", err)
	}
	if preflight.Lane.Runtime.Target != "node" || preflight.Lane.Runtime.Backend != "proxy" {
		t.Fatalf("expected managed default page-action preflight to resolve node lane, got %#v", preflight)
	}
	if !preflight.Lane.Capabilities.Extract {
		t.Fatalf("expected concrete node lane to keep extract capability, got %#v", preflight.Lane.Capabilities)
	}
	if preflight.Lane.Capabilities.Click {
		t.Fatalf("expected concrete node lane not to inherit host-only click capability, got %#v", preflight.Lane.Capabilities)
	}
}

func TestBrowserRuntimeRouterResolveSessionPreflightRejectsHiddenManagedCurrentTarget(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	backend := newBrowserBackend(BrowserToolOptions{
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: stateRegistry,
	}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected router backend wrapper")
	}
	callCtx := WithToolSessionID(context.Background(), "router-session-preflight-hidden-managed-current")
	sessionRegistry.TrackCurrentTarget("router-session-preflight-hidden-managed-current", BrowserSessionTarget{
		ID:         "stale-node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	request, err := router.buildSessionPreflightRequest(
		callCtx,
		map[string]any{},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		BrowserRuntimeInfo{},
		browserRuntimeRouteDescriptor{},
		true,
	)
	if err != nil {
		t.Fatalf("buildSessionPreflightRequest(hidden managed current) error = %v", err)
	}
	if !request.HiddenRequestedRuntimeTarget {
		t.Fatalf("expected hidden managed current target to remain visible to shared session preflight guard, got %#v", request)
	}
	if request.Preview.RequestedRuntimeTarget != "node" {
		t.Fatalf("expected preview requested runtime_target=node, got %#v", request.Preview)
	}
	if request.Requested != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected hidden managed current target to keep visible request targetless, got %#v", request.Requested)
	}

	_, err = router.resolveSessionPreflight(request)
	if err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target") {
		t.Fatalf("expected shared session preflight to keep hidden managed current target behind explicit runtime_target gate, got %v", err)
	}
}

func TestBrowserRuntimeRouterResolveSessionPreflightUsesManagedCurrentRouteForImplicitFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend:          node,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: stateRegistry,
	}, host, browserDefaultSubstrateAssessment{
		HostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      host,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
				Capabilities: browserCapabilitiesForConcreteBackend(host),
			},
		},
		NodeRoute: browserDefaultPromotionRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Ready:          true,
			Route: browserResolvedExecutionRoute{
				Backend:      node,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
				Capabilities: browserCapabilitiesForConcreteBackend(node),
			},
		},
		DefaultRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}).(browserRuntimeRouterBackend)
	callCtx := WithToolSessionID(context.Background(), "router-session-preflight-managed-current")
	sessionRegistry.TrackCurrentTarget("router-session-preflight-managed-current", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	request, err := router.buildSessionPreflightRequest(
		callCtx,
		map[string]any{},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node"},
		true,
	)
	if err != nil {
		t.Fatalf("buildSessionPreflightRequest(managed current) error = %v", err)
	}
	if request.DefaultCandidateRoute != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected managed current session preflight request to preserve default candidate route, got %#v", request.DefaultCandidateRoute)
	}
	preflight, err := router.resolveSessionPreflight(request)
	if err != nil {
		t.Fatalf("resolveSessionPreflight(managed current) error = %v", err)
	}
	if preflight.DefaultCandidateRoute != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected managed current session preflight to preserve default candidate route, got %#v", preflight.DefaultCandidateRoute)
	}
	if preflight.Route.RuntimeInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected managed current session preflight to resolve concrete node lane, got %#v", preflight.Route.RuntimeInfo)
	}
	if !preflight.CanUseManagedSessionRouteForImplicitFallback {
		t.Fatalf("expected managed current session preflight to keep implicit managed fallback signal, got %#v", preflight)
	}
}

func TestBrowserRuntimeRouterResolveSessionPreflightSurfacesHiddenManagedDefaultCandidateRoute(t *testing.T) {
	nodeBackend := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       fullBrowserCapabilities(),
			routeSource:        "managed_browserd",
			routeEndpoint:      "http://127.0.0.1:43123",
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	backend := newBrowserBackend(BrowserToolOptions{NodeBackend: nodeBackend}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected router backend wrapper")
	}

	executionPreview := router.sessionExecutionPreview(
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
	)
	request, err := router.buildSessionPreflightRequest(
		context.Background(),
		map[string]any{},
		executionPreview.Base,
		executionPreview.DefaultCandidateRoute,
		executionPreview.DefaultCandidateDescriptor,
		executionPreview.HiddenImplicitHostDefaultBase,
	)
	if err != nil {
		t.Fatalf("buildSessionPreflightRequest(hidden managed candidate) error = %v", err)
	}
	if request.DefaultCandidateRoute != (BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}) {
		t.Fatalf("expected hidden managed session preflight request to preserve default candidate route, got %#v", request.DefaultCandidateRoute)
	}
	if request.DefaultCandidateDescriptor != (browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Source:        "managed_browserd",
		Endpoint:      "http://127.0.0.1:43123",
	}) {
		t.Fatalf("expected hidden managed session preflight request to preserve default candidate provenance, got %#v", request.DefaultCandidateDescriptor)
	}
	if !request.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected hidden managed session preflight request to keep hidden implicit host default flag, got %#v", request)
	}
	_, err = router.resolveSessionPreflight(request)
	if err == nil {
		t.Fatalf("expected hidden managed session preflight to fail without a visible default route")
	}
}

func TestBrowserRuntimeRouterResolveBrowserExecutionRouteForSessionKeepsExplicitHostLane(t *testing.T) {
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilities{Open: true, Navigate: true},
	}
	nodeBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilities{Open: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: hostBackend,
		hostRuntime: BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		hostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      hostBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
				Capabilities: hostBackend.capabilities,
			},
		},
		nodeBackend: nodeBackend,
		nodeRoute: browserDefaultPromotionRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Ready:          true,
			Route: browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
				Capabilities: nodeBackend.capabilities,
			},
		},
		defaultRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
				Capabilities: nodeBackend.capabilities,
			},
		},
	}
	route, err := router.ResolveBrowserExecutionRouteForSession(context.Background(), map[string]any{
		"runtime_target": "host",
		"profile":        "workbench",
	}, BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}, true)
	if err != nil {
		t.Fatalf("ResolveBrowserExecutionRouteForSession(explicit host) error = %v", err)
	}
	if route.Backend != hostBackend || route.RuntimeInfo != (BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"}) {
		t.Fatalf("expected session-aware explicit host route to bypass promoted node lane, got %#v", route)
	}
}

func TestBrowserRuntimeRouterResolveBrowserExecutionRouteForSessionUsesExecutionPreviewDispatchBase(t *testing.T) {
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		capabilities:       BrowserCapabilities{Open: true, Navigate: true},
	}
	nodeBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilities{Open: true, Navigate: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: hostBackend,
		hostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		hostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      hostBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
				Capabilities: hostBackend.capabilities,
			},
		},
		nodeBackend: nodeBackend,
		nodeRoute: browserDefaultPromotionRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Ready:          true,
			Route: browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
				Capabilities: nodeBackend.capabilities,
			},
		},
		defaultRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
				Capabilities: nodeBackend.capabilities,
			},
		},
	}

	route, err := router.ResolveBrowserExecutionRouteForSession(
		context.Background(),
		map[string]any{},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
	)
	if err != nil {
		t.Fatalf("ResolveBrowserExecutionRouteForSession(stale logical host base) error = %v", err)
	}
	if route.Backend != nodeBackend || route.RuntimeInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected session-aware execution route to reuse router execution preview dispatch base, got %#v", route)
	}
}

func TestBrowserRuntimeRouterResolveBrowserRuntimeRouteForSessionUsesExecutionPreviewDispatchBase(t *testing.T) {
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		capabilities:       BrowserCapabilities{Open: true, Navigate: true},
	}
	nodeBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilities{Open: true, Navigate: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: hostBackend,
		hostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		hostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      hostBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
				Capabilities: hostBackend.capabilities,
			},
		},
		nodeBackend: nodeBackend,
		nodeRoute: browserDefaultPromotionRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Ready:          true,
			Route: browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
				Capabilities: nodeBackend.capabilities,
			},
		},
		defaultRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
				Capabilities: nodeBackend.capabilities,
			},
		},
	}

	runtimeInfo, err := router.ResolveBrowserRuntimeRouteForSession(
		context.Background(),
		map[string]any{},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
	)
	if err != nil {
		t.Fatalf("ResolveBrowserRuntimeRouteForSession(stale logical host base) error = %v", err)
	}
	if runtimeInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected session-aware runtime route to reuse router execution preview dispatch base, got %#v", runtimeInfo)
	}
}

func TestBrowserRuntimeRouterResolveBrowserExecutionRouteForSessionUsesExecutionPreviewHiddenImplicitHostBase(t *testing.T) {
	router := browserRuntimeRouterBackend{
		hostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}

	_, err := router.ResolveBrowserExecutionRouteForSession(
		context.Background(),
		map[string]any{},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
	)
	if err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target") {
		t.Fatalf("expected session-aware execution route to keep implicit legacy host behind explicit runtime_target gate when execution preview hides host base, got %v", err)
	}
}

func TestBrowserRuntimeRouterCheckDirectURLActionFallbackWrapsLegacyURLGate(t *testing.T) {
	router := browserRuntimeRouterBackend{
		hostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	err := router.checkDirectURLActionFallback("browser backend open", browserRuntimeRouterDirectPreflightRequest{
		browserRuntimeRouterSessionPreflightRequest: browserRuntimeRouterSessionPreflightRequest{
			HiddenImplicitHostDefaultBase: true,
		},
	}, "https://93.184.216.34")
	if err == nil {
		t.Fatalf("expected legacy host URL fallback gate to reject implicit URL dispatch")
	}
}

func TestBrowserRuntimeRouterCheckDirectTabsActionFallbackWrapsLegacyTabsGate(t *testing.T) {
	router := browserRuntimeRouterBackend{
		hostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	err := router.checkDirectTabsActionFallback("browser backend tabs", browserRuntimeRouterDirectPreflightRequest{
		browserRuntimeRouterSessionPreflightRequest: browserRuntimeRouterSessionPreflightRequest{
			Requested:                     BrowserRuntimeInfo{},
			HiddenImplicitHostDefaultBase: true,
		},
		Target: browserToolTarget{TabIndex: 2},
	}, "focus")
	if err == nil {
		t.Fatalf("expected legacy host tabs fallback gate to reject implicit focus dispatch")
	}
}

func TestBrowserRuntimeRouterCheckDirectPageActionFallbackWrapsLegacyPageGate(t *testing.T) {
	router := browserRuntimeRouterBackend{
		hostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	err := router.checkDirectPageActionFallback("browser backend extract", browserRuntimeRouterDirectPreflightRequest{
		browserRuntimeRouterSessionPreflightRequest: browserRuntimeRouterSessionPreflightRequest{
			Requested:                     BrowserRuntimeInfo{},
			HiddenImplicitHostDefaultBase: true,
		},
		Target: browserToolTarget{TabIndex: 1},
	}, "")
	if err == nil {
		t.Fatalf("expected legacy host page fallback gate to reject implicit page dispatch")
	}
}

func TestBrowserRuntimeRouterInvokeDirectURLActionPreparesBrowserApp(t *testing.T) {
	backend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilities{Open: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: backend,
		hostRuntime: BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		hostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      backend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
				Capabilities: backend.capabilities,
			},
		},
	}
	got, err := invokeDirectURLAction(context.Background(), router, "browser backend open", "", 0, "https://93.184.216.34", func(_ BrowserBackend, browserApp string) (string, error) {
		return browserApp, nil
	})
	if err != nil {
		t.Fatalf("invokeDirectURLAction error = %v", err)
	}
	if got != "" {
		t.Fatalf("expected direct URL invoke to preserve empty browser app when no preview app exists, got %q", got)
	}
}

func TestBrowserRuntimeRouterInvokeDirectTabsActionPreparesBrowserApp(t *testing.T) {
	backend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilities{Tabs: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: backend,
		hostRuntime: BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		hostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      backend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
				Capabilities: backend.capabilities,
			},
		},
	}
	got, err := invokeDirectTabsAction(context.Background(), router, "browser backend tabs", BrowserTabsRequest{Action: "list", BrowserApp: "chrome"}, func(_ BrowserBackend, req BrowserTabsRequest) (string, error) {
		return req.BrowserApp, nil
	})
	if err != nil {
		t.Fatalf("invokeDirectTabsAction error = %v", err)
	}
	if got != "chrome" {
		t.Fatalf("expected direct tabs invoke to preserve browser app, got %q", got)
	}
}

func TestBrowserRuntimeRouterInvokeDirectPageActionPreparesBrowserApp(t *testing.T) {
	backend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilities{Extract: true},
	}
	router := browserRuntimeRouterBackend{
		hostBackend: backend,
		hostRuntime: BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		hostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      backend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
				Capabilities: backend.capabilities,
			},
		},
	}
	got, err := invokeDirectPageAction(context.Background(), router, "browser backend extract", "chrome", 0, "", func(_ BrowserBackend, browserApp string) (string, error) {
		return browserApp, nil
	})
	if err != nil {
		t.Fatalf("invokeDirectPageAction error = %v", err)
	}
	if got != "chrome" {
		t.Fatalf("expected direct page invoke to preserve browser app, got %q", got)
	}
}

func TestResolveBrowserActDispatchUsesManagedDispatchBaseForImplicitLegacyHostPageAction(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "extract"}),
	}
	nodeBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"click"}),
	}
	router, ok := newBrowserRuntimeRouterBackendWithAssessment(
		BrowserToolOptions{
			Backend:              hostBackend,
			NodeBackend:          nodeBackend,
			SessionRegistry:      sessionRegistry,
			SessionStateRegistry: stateRegistry,
		},
		hostBackend,
		browserDefaultSubstrateAssessment{
			HostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			HostRoute:   browserImplicitLegacyHostRouteAssessment(nil, BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}),
			NodeRoute: browserDefaultPromotionRouteAssessmentForConcreteRouteAssessment(browserConcreteRouteAssessment{
				Configured:     true,
				RouteAvailable: true,
				Route: browserResolvedExecutionRoute{
					Backend:      nodeBackend,
					RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
					Capabilities: BrowserCapabilitiesForActKinds([]string{"click"}),
				},
			}, "node"),
			DefaultRuntime:       BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			DefaultConcreteRoute: browserConcreteRouteAssessment{},
		},
	).(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected concrete browser runtime router backend")
	}
	dispatch, err := resolveBrowserActDispatch(
		context.Background(),
		router,
		sessionRegistry,
		stateRegistry,
		agentxbrowserruntime.SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, browserRuntimeReconnectWatchdogWindow),
		DefaultBrowserRuntimeInfo(),
		true,
		BrowserToolOptions{DefaultBrowserApp: "chrome"},
		1024,
		map[string]any{"kind": "click"},
		"click",
	)
	if err != nil {
		t.Fatalf("resolveBrowserActDispatch(implicit legacy host click): %v", err)
	}
	if dispatch.RuntimeInfo.Backend != "proxy" || dispatch.RuntimeInfo.Profile != "isolated" || dispatch.RuntimeInfo.Target != "node" {
		t.Fatalf("expected click preflight to promote implicit legacy-host route onto managed node runtime, got %#v", dispatch.RuntimeInfo)
	}
	if dispatch.Target != (browserToolTarget{}) {
		t.Fatalf("expected click preflight without url/target to stay targetless, got %#v", dispatch.Target)
	}
	if dispatch.ExplicitRuntimeTarget {
		t.Fatalf("expected click preflight without runtime_target to stay implicit")
	}
}

func TestResolveBrowserActDispatchKeepsDefaultRuntimeForOpenOutsideImplicitLegacyHost(t *testing.T) {
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	dispatch, err := resolveBrowserActDispatch(
		context.Background(),
		hostBackend,
		sessionRegistry,
		stateRegistry,
		agentxbrowserruntime.SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, browserRuntimeReconnectWatchdogWindow),
		BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		false,
		BrowserToolOptions{DefaultBrowserApp: "chrome"},
		1024,
		map[string]any{"kind": "open"},
		"open",
	)
	if err != nil {
		t.Fatalf("resolveBrowserActDispatch(non-legacy open): %v", err)
	}
	if dispatch.RuntimeInfo != (BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"}) {
		t.Fatalf("expected open preflight outside implicit legacy-host fallback to retain host runtime info, got %#v", dispatch.RuntimeInfo)
	}
	if dispatch.BrowserApp != "chrome" {
		t.Fatalf("expected default browser app to flow through act dispatch, got %q", dispatch.BrowserApp)
	}
}

func TestResolveBrowserRegistrationPageActionDispatchUsesManagedRouteForImplicitLegacyHostPageAction(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"open", "extract"}),
	}
	nodeBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"click"}),
	}
	substrateAssessment := browserDefaultSubstrateAssessment{
		HostRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		HostRoute:   browserImplicitLegacyHostRouteAssessment(nil, BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}),
		NodeRoute: browserDefaultPromotionRouteAssessmentForConcreteRouteAssessment(browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route: browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
				Capabilities: BrowserCapabilitiesForActKinds([]string{"click"}),
			},
		}, "node"),
		DefaultRuntime:       BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		DefaultConcreteRoute: browserConcreteRouteAssessment{},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(
		BrowserToolOptions{
			Backend:              hostBackend,
			NodeBackend:          nodeBackend,
			SessionRegistry:      sessionRegistry,
			SessionStateRegistry: stateRegistry,
		},
		hostBackend,
		substrateAssessment,
	)
	registrationCtx := browserRegistrationContext{
		opts: BrowserToolOptions{
			Root:                 t.TempDir(),
			Backend:              hostBackend,
			NodeBackend:          nodeBackend,
			SessionRegistry:      sessionRegistry,
			SessionStateRegistry: stateRegistry,
		},
		sessionRegistry:      sessionRegistry,
		sessionStateRegistry: stateRegistry,
		watchManagerProvider: agentxbrowserruntime.SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, browserRuntimeReconnectWatchdogWindow),
		backend:              router,
		maxChars:             1024,
		substrateAssessment:  substrateAssessment,
		substrateSummary: BrowserWorkbenchSubstrateSummary{
			DefaultRoute: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		},
	}
	dispatch, err := resolveBrowserRegistrationPageActionDispatch(
		registrationCtx,
		context.Background(),
		map[string]any{},
		browserRegistrationPageActionDispatchOptions{UseManagedRoute: true, UseManagedDefaultDispatchBase: true},
	)
	if err != nil {
		t.Fatalf("resolveBrowserRegistrationPageActionDispatch(implicit legacy host click path): %v", err)
	}
	if dispatch.RuntimeInfo.Backend != "proxy" || dispatch.RuntimeInfo.Profile != "isolated" || dispatch.RuntimeInfo.Target != "node" {
		t.Fatalf("expected registration page-action dispatch to promote implicit legacy-host route onto managed node runtime, got %#v", dispatch.RuntimeInfo)
	}
}

func TestResolveBrowserRegistrationPageActionDispatchKeepsDefaultRuntimeOutsideImplicitLegacyHost(t *testing.T) {
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
	}
	registrationCtx := browserRegistrationContext{
		opts: BrowserToolOptions{
			Root:    t.TempDir(),
			Backend: hostBackend,
		},
		backend:  hostBackend,
		maxChars: 1024,
		substrateSummary: BrowserWorkbenchSubstrateSummary{
			DefaultRoute: BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		},
	}
	dispatch, err := resolveBrowserRegistrationPageActionDispatch(
		registrationCtx,
		context.Background(),
		map[string]any{},
		browserRegistrationPageActionDispatchOptions{},
	)
	if err != nil {
		t.Fatalf("resolveBrowserRegistrationPageActionDispatch(non-legacy open path): %v", err)
	}
	if dispatch.RuntimeInfo != (BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"}) {
		t.Fatalf("expected registration dispatch outside implicit legacy-host fallback to retain host runtime info, got %#v", dispatch.RuntimeInfo)
	}
}

func TestResolveBrowserActDispatchKeepsDarwinInteractiveHostRuntime(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-only guardrail")
	}
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		capabilities:       fullBrowserCapabilities(),
	}
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	dispatch, err := resolveBrowserActDispatch(
		context.Background(),
		hostBackend,
		sessionRegistry,
		stateRegistry,
		agentxbrowserruntime.SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, browserRuntimeReconnectWatchdogWindow),
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		false,
		BrowserToolOptions{DefaultBrowserApp: "safari"},
		1024,
		map[string]any{"kind": "click"},
		"click",
	)
	if err != nil {
		t.Fatalf("resolveBrowserActDispatch(darwin default click): %v", err)
	}
	if dispatch.RuntimeInfo != (BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}) {
		t.Fatalf("expected darwin default interactive click path to stay on host runtime, got %#v", dispatch.RuntimeInfo)
	}
	if dispatch.BrowserApp != "safari" {
		t.Fatalf("expected darwin default browser app to remain safari, got %q", dispatch.BrowserApp)
	}
}

func TestResolveBrowserExecutionLaneForActDispatchUsesResolvedConcreteCapabilities(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"click"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			if !strings.EqualFold(strings.TrimSpace(requested.Profile), "workbench") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	ctx, _, ok := newBrowserRegistrationContext(reg, BrowserToolOptions{
		Root:        t.TempDir(),
		Backend:     hostBackend,
		NodeBackend: nodeBackend,
	})
	if !ok {
		t.Fatalf("expected browser registration context")
	}
	if ctx.capabilities.SupportsActKind("click") {
		t.Fatalf("expected static registration capability surface to stay conservative, got %#v", ctx.capabilities)
	}
	dispatchBase, hiddenImplicitHostDefaultBase := browserRegistrationDispatchRuntime(ctx)
	lane, dispatch, err := resolveBrowserExecutionLaneForActDispatch(
		context.Background(),
		ctx.backend,
		ctx.sessionRegistry,
		ctx.sessionStateRegistry,
		ctx.watchManagerProvider,
		dispatchBase,
		hiddenImplicitHostDefaultBase,
		ctx.opts,
		ctx.maxChars,
		map[string]any{
			"kind":           "click",
			"profile":        "workbench",
			"runtime_target": "node",
		},
		"click",
	)
	if err != nil {
		t.Fatalf("resolveBrowserExecutionLaneForActDispatch(explicit managed click) error = %v", err)
	}
	want := BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}
	if lane.Runtime != want || dispatch.RuntimeInfo != want {
		t.Fatalf("expected act dispatch lane/runtime to resolve explicit managed route, lane=%#v dispatch=%#v", lane, dispatch)
	}
	if lane.Backend != nodeBackend {
		t.Fatalf("expected act dispatch lane backend to use explicit node backend, got %T", lane.Backend)
	}
	if !lane.Capabilities.Click || lane.Capabilities.Open {
		t.Fatalf("expected act dispatch lane to use concrete node capabilities, got %#v", lane.Capabilities)
	}
}
