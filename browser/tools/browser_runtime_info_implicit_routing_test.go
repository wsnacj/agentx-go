package tools

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestBrowserRuntimeRouterBackendResolveDefaultProfileRouteRequiresExplicitRuntimeTargetWithoutDefaultConcreteRoute(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	if _, ok := browserRuntimeRouterCachedDefaultConcreteRouteAssessment(router); ok {
		t.Fatalf("expected implicit legacy host to stay out of router-owned default concrete route cache")
	}
	if _, err := router.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{Profile: "default"}); err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target") {
		t.Fatalf("expected implicit legacy default-profile route to require explicit runtime_target, got %v", err)
	}
	if _, ok := browserRuntimeRouterCachedRouteAssessment(router, BrowserRuntimeInfo{Profile: "default"}); ok {
		t.Fatalf("expected implicit legacy default-profile route to stay out of generic targetless cached owner")
	}
}

func TestBrowserRuntimeRouterBackendResolveImplicitHostDefaultRouteRequiresExplicitRuntimeTargetEvenWhenExplicitHostFails(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	router.hostRoute = browserConcreteRouteAssessment{
		Configured:    true,
		FailureReason: "cached explicit host failure",
		FailureNote:   "cached explicit host failure",
	}
	if _, err := router.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{}); err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target") {
		t.Fatalf("expected implicit legacy host default route to require explicit runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendResolveImplicitHostDefaultRouteRequiresExplicitRuntimeTargetWithoutFallbackRoute(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	if _, err := router.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{}); err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target") {
		t.Fatalf("expected implicit legacy host default route to require explicit runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendTabsListRequiresExplicitHostForImplicitFallback(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	_, err := router.Tabs(context.Background(), BrowserTabsRequest{Action: "list"})
	if err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target=host") {
		t.Fatalf("expected implicit host tabs list to require explicit host runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendTabsListRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			tabsResult: BrowserTabsResult{Backend: "node-tabs", Action: "list", Status: "ok", BrowserApp: "Chromium"},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend:          node,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: NewBrowserSessionStateRegistry(),
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
	callCtx := WithToolSessionID(context.Background(), "router-tabs-list-managed-current-hidden-implicit-host")
	sessionRegistry.TrackCurrentTarget("router-tabs-list-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	result, err := router.Tabs(callCtx, BrowserTabsRequest{Action: "list"})
	if err != nil {
		t.Fatalf("Tabs(list managed current route) error = %v", err)
	}
	if len(host.tabsReqs) != 0 {
		t.Fatalf("expected managed current route to avoid host tabs backend, got %#v", host.tabsReqs)
	}
	if len(node.tabsReqs) != 1 || node.tabsReqs[0].Action != "list" || node.tabsReqs[0].BrowserApp != "Chromium" || node.tabsReqs[0].TabIndex != 0 {
		t.Fatalf("expected managed current route to dispatch tabs list through node backend, got %#v", node.tabsReqs)
	}
	if result.Backend != "node-tabs" || result.Action != "list" || result.Status != "ok" || result.BrowserApp != "Chromium" {
		t.Fatalf("unexpected managed current route direct tabs result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendNavigateTabTargetRequiresExplicitHostForImplicitFallback(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	_, err := router.Navigate(context.Background(), BrowserNavigateRequest{
		URL:      "https://93.184.216.34",
		TabIndex: 2,
	})
	if err == nil || !strings.Contains(err.Error(), `target "tab:2" requires explicit runtime_target=host`) {
		t.Fatalf("expected implicit host navigate tab target to require explicit host runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendOpenRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	_, err := router.Open(context.Background(), BrowserOpenRequest{URL: "https://93.184.216.34"})
	if err == nil || !strings.Contains(err.Error(), "browser backend open requires explicit runtime_target because the default browser route falls back to the legacy system host path") {
		t.Fatalf("expected implicit host open to require explicit runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendOpenSurfacesManagedDefaultRouteFailureBeforeImplicitHostGate(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       BrowserCapabilities{Open: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{}, errors.New("managed browserd boot failed")
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend: node,
	}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	_, err := router.Open(context.Background(), BrowserOpenRequest{URL: "https://93.184.216.34"})
	if err == nil {
		t.Fatal("expected targetless managed-default launch failure")
	}
	if !strings.Contains(err.Error(), "managed browserd boot failed") ||
		strings.Contains(err.Error(), "requires explicit runtime_target") {
		t.Fatalf("expected targetless open to surface managed-default launch failure before implicit-host gate, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendOpenRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				openResult: BrowserOpenResult{Backend: "node-open", Status: "opened"},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{Open: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{NodeBackend: node}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	result, err := router.Open(context.Background(), BrowserOpenRequest{URL: "https://93.184.216.34"})
	if err != nil {
		t.Fatalf("Open(hidden implicit-host managed default) error = %v", err)
	}
	if len(host.openReqs) != 0 {
		t.Fatalf("expected managed default open to avoid host backend, got %#v", host.openReqs)
	}
	if len(node.openReqs) != 1 || node.openReqs[0].URL != "https://93.184.216.34" {
		t.Fatalf("expected managed default open to dispatch through node backend, got %#v", node.openReqs)
	}
	if result.Backend != "node-open" || result.Status != "opened" {
		t.Fatalf("unexpected managed default direct open result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendOpenKeepsImplicitHostGateForUnresolvableManagedCurrentRoute(t *testing.T) {
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
	callCtx := WithToolSessionID(context.Background(), "router-open-stale-managed-current-hidden-implicit-host")
	sessionRegistry.TrackCurrentTarget("router-open-stale-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "stale-node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	_, err := router.Open(callCtx, BrowserOpenRequest{URL: "https://93.184.216.34"})
	if err == nil || !strings.Contains(err.Error(), "browser backend open requires explicit runtime_target because the default browser route falls back to the legacy system host path") {
		t.Fatalf("expected stale managed current route to keep direct open behind implicit host explicit runtime_target gate, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendNavigateRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	_, err := router.Navigate(context.Background(), BrowserNavigateRequest{URL: "https://93.184.216.34"})
	if err == nil || !strings.Contains(err.Error(), "browser backend navigate requires explicit runtime_target because the default browser route falls back to the legacy system host path") {
		t.Fatalf("expected implicit host navigate to require explicit runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendNavigateRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				navigateResult: BrowserNavigateResult{Backend: "node-navigate", Status: "navigated"},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{Navigate: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{NodeBackend: node}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	result, err := router.Navigate(context.Background(), BrowserNavigateRequest{URL: "https://93.184.216.34"})
	if err != nil {
		t.Fatalf("Navigate(hidden implicit-host managed default) error = %v", err)
	}
	if len(host.navigateReqs) != 0 {
		t.Fatalf("expected managed default navigate to avoid host backend, got %#v", host.navigateReqs)
	}
	if len(node.navigateReqs) != 1 || node.navigateReqs[0].URL != "https://93.184.216.34" || node.navigateReqs[0].TabIndex != 0 {
		t.Fatalf("expected managed default navigate to dispatch through node backend, got %#v", node.navigateReqs)
	}
	if result.Backend != "node-navigate" || result.Status != "navigated" {
		t.Fatalf("unexpected managed default direct navigate result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendTabsFocusRequiresExplicitHostForImplicitFallback(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	_, err := router.Tabs(context.Background(), BrowserTabsRequest{
		Action:   "focus",
		TabIndex: 2,
	})
	if err == nil || !strings.Contains(err.Error(), `target "tab:2" requires explicit runtime_target=host`) {
		t.Fatalf("expected implicit host tabs focus target to require explicit host runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendTabsListRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				tabsResult: BrowserTabsResult{Backend: "node-tabs", Action: "list", Status: "listed"},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{Tabs: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{NodeBackend: node}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	result, err := router.Tabs(context.Background(), BrowserTabsRequest{Action: "list"})
	if err != nil {
		t.Fatalf("Tabs(list hidden implicit-host managed default) error = %v", err)
	}
	if len(host.tabsReqs) != 0 {
		t.Fatalf("expected managed default tabs list to avoid host backend, got %#v", host.tabsReqs)
	}
	if len(node.tabsReqs) != 1 || node.tabsReqs[0].Action != "list" || node.tabsReqs[0].TabIndex != 0 {
		t.Fatalf("expected managed default tabs list to dispatch through node backend, got %#v", node.tabsReqs)
	}
	if result.Backend != "node-tabs" || result.Action != "list" || result.Status != "listed" {
		t.Fatalf("unexpected managed default direct tabs list result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendTabsFocusRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				tabsResult: BrowserTabsResult{Backend: "node-tabs", Action: "focus", Status: "focused"},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{Tabs: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{NodeBackend: node}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	result, err := router.Tabs(context.Background(), BrowserTabsRequest{Action: "focus", TabIndex: 2, WaitMs: 120})
	if err != nil {
		t.Fatalf("Tabs(focus hidden implicit-host managed default) error = %v", err)
	}
	if len(host.tabsReqs) != 0 {
		t.Fatalf("expected managed default tabs focus to avoid host backend, got %#v", host.tabsReqs)
	}
	if len(node.tabsReqs) != 1 || node.tabsReqs[0].Action != "focus" || node.tabsReqs[0].TabIndex != 2 || node.tabsReqs[0].WaitMs != 120 {
		t.Fatalf("expected managed default tabs focus to dispatch through node backend, got %#v", node.tabsReqs)
	}
	if result.Backend != "node-tabs" || result.Action != "focus" || result.Status != "focused" {
		t.Fatalf("unexpected managed default direct tabs focus result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendTabsCloseRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				tabsResult: BrowserTabsResult{Backend: "node-tabs", Action: "close", Status: "closed"},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{Tabs: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{NodeBackend: node}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	result, err := router.Tabs(context.Background(), BrowserTabsRequest{Action: "close", TabIndex: 2})
	if err != nil {
		t.Fatalf("Tabs(close hidden implicit-host managed default) error = %v", err)
	}
	if len(host.tabsReqs) != 0 {
		t.Fatalf("expected managed default tabs close to avoid host backend, got %#v", host.tabsReqs)
	}
	if len(node.tabsReqs) != 1 || node.tabsReqs[0].Action != "close" || node.tabsReqs[0].TabIndex != 2 {
		t.Fatalf("expected managed default tabs close to dispatch through node backend, got %#v", node.tabsReqs)
	}
	if result.Backend != "node-tabs" || result.Action != "close" || result.Status != "closed" {
		t.Fatalf("unexpected managed default direct tabs close result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendExtractRequiresExplicitHostCurrentPageForImplicitFallback(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	_, err := router.Extract(context.Background(), BrowserExtractRequest{MaxChars: 128})
	if err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target=host or an explicit target/url") {
		t.Fatalf("expected implicit host extract to require explicit host runtime_target or target/url, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendExtractRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			extractResult: BrowserExtractResult{Backend: "node-extract", BrowserApp: "Chromium", FinalURL: "https://node.example/current"},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend:          node,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: NewBrowserSessionStateRegistry(),
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
	callCtx := WithToolSessionID(context.Background(), "router-extract-managed-current-hidden-implicit-host")
	sessionRegistry.TrackCurrentTarget("router-extract-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	result, err := router.Extract(callCtx, BrowserExtractRequest{MaxChars: 128})
	if err != nil {
		t.Fatalf("Extract(managed current route) error = %v", err)
	}
	if len(host.extractReqs) != 0 {
		t.Fatalf("expected managed current route to avoid host extract backend, got %#v", host.extractReqs)
	}
	if len(node.extractReqs) != 1 || node.extractReqs[0].BrowserApp != "Chromium" || node.extractReqs[0].URL != "" || node.extractReqs[0].TabIndex != 0 || node.extractReqs[0].MaxChars != 128 {
		t.Fatalf("expected managed current route to dispatch extract through node backend, got %#v", node.extractReqs)
	}
	if result.Backend != "node-extract" || result.BrowserApp != "Chromium" || result.FinalURL != "https://node.example/current" {
		t.Fatalf("unexpected managed current route direct extract result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendExtractRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				extractResult: BrowserExtractResult{
					Backend:     "node-extract",
					BrowserApp:  "Chromium",
					FinalURL:    "https://node.example/default",
					Content:     "managed default content",
					ContentType: "text/plain",
				},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{Extract: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{NodeBackend: node}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	result, err := router.Extract(context.Background(), BrowserExtractRequest{MaxChars: 128})
	if err != nil {
		t.Fatalf("Extract(hidden implicit-host managed default) error = %v", err)
	}
	if len(host.extractReqs) != 0 {
		t.Fatalf("expected managed default extract to avoid host backend, got %#v", host.extractReqs)
	}
	if len(node.extractReqs) != 1 || node.extractReqs[0].TabIndex != 0 || node.extractReqs[0].URL != "" || node.extractReqs[0].MaxChars != 128 {
		t.Fatalf("expected managed default extract to dispatch through node backend, got %#v", node.extractReqs)
	}
	if result.Backend != "node-extract" || result.BrowserApp != "Chromium" || result.FinalURL != "https://node.example/default" {
		t.Fatalf("unexpected managed default direct extract result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendSnapshotRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				snapshotResult: BrowserSnapshotResult{
					Backend:    "node-snapshot",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/default",
					Title:      "Workbench",
					Snapshot:   "managed default snapshot",
					Format:     "ai",
				},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{Snapshot: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{NodeBackend: node}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	result, err := router.Snapshot(context.Background(), BrowserSnapshotRequest{MaxChars: 64, MaxElements: 8})
	if err != nil {
		t.Fatalf("Snapshot(hidden implicit-host managed default) error = %v", err)
	}
	if len(host.snapshotReqs) != 0 {
		t.Fatalf("expected managed default snapshot to avoid host backend, got %#v", host.snapshotReqs)
	}
	if len(node.snapshotReqs) != 1 || node.snapshotReqs[0].TabIndex != 0 || node.snapshotReqs[0].URL != "" || node.snapshotReqs[0].MaxChars != 64 || node.snapshotReqs[0].MaxElements != 8 {
		t.Fatalf("expected managed default snapshot to dispatch through node backend, got %#v", node.snapshotReqs)
	}
	if result.Backend != "node-snapshot" || result.BrowserApp != "Chromium" || result.FinalURL != "https://node.example/default" || result.Snapshot != "managed default snapshot" {
		t.Fatalf("unexpected managed default direct snapshot result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendClickRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				clickResult: BrowserClickResult{
					Backend:    "node-click",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/after-click",
					Title:      "Clicked",
					Status:     "clicked",
				},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{Click: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{NodeBackend: node}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	result, err := router.Click(context.Background(), BrowserClickRequest{Selector: "button.buy", PostWaitMs: 500})
	if err != nil {
		t.Fatalf("Click(hidden implicit-host managed default) error = %v", err)
	}
	if len(host.clickReqs) != 0 {
		t.Fatalf("expected managed default click to avoid host backend, got %#v", host.clickReqs)
	}
	if len(node.clickReqs) != 1 || node.clickReqs[0].TabIndex != 0 || node.clickReqs[0].URL != "" || node.clickReqs[0].Selector != "button.buy" || node.clickReqs[0].PostWaitMs != 500 || node.clickReqs[0].WaitMs != 0 {
		t.Fatalf("expected managed default click to dispatch through node backend, got %#v", node.clickReqs)
	}
	if result.Backend != "node-click" || result.BrowserApp != "Chromium" || result.FinalURL != "https://node.example/after-click" || result.Status != "clicked" {
		t.Fatalf("unexpected managed default direct click result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendScreenshotRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	root := t.TempDir()
	outputPath := root + "/managed-default.png"
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				screenshotResult: BrowserScreenshotResult{
					Backend:       "node-screenshot",
					BrowserApp:    "Chromium",
					FinalURL:      "https://node.example/default",
					Title:         "Workbench",
					CaptureScope:  "viewport",
					CaptureWidth:  1280,
					CaptureHeight: 720,
					Status:        "captured",
				},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{Screenshot: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{Root: root, NodeBackend: node}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	result, err := router.Screenshot(context.Background(), BrowserScreenshotRequest{OutputPath: outputPath})
	if err != nil {
		t.Fatalf("Screenshot(hidden implicit-host managed default) error = %v", err)
	}
	if len(host.screenshotReqs) != 0 {
		t.Fatalf("expected managed default screenshot to avoid host backend, got %#v", host.screenshotReqs)
	}
	if len(node.screenshotReqs) != 1 || node.screenshotReqs[0].TabIndex != 0 || node.screenshotReqs[0].URL != "" || node.screenshotReqs[0].WaitMs != 0 || node.screenshotReqs[0].OutputPath != outputPath {
		t.Fatalf("expected managed default screenshot to dispatch through node backend, got %#v", node.screenshotReqs)
	}
	if result.Backend != "node-screenshot" || result.BrowserApp != "Chromium" || result.FinalURL != "https://node.example/default" || result.Path != outputPath || result.Status != "captured" {
		t.Fatalf("unexpected managed default direct screenshot result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendTypeRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				typeResult: BrowserTypeResult{
					Backend:    "node-type",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/form",
					Title:      "Form",
					Value:      "agentx",
					Status:     "typed",
					Submitted:  true,
				},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{TypeText: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{NodeBackend: node}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	result, err := router.Type(context.Background(), BrowserTypeRequest{Selector: "input[name=q]", Text: "agentx", Submit: true, PostWaitMs: 250})
	if err != nil {
		t.Fatalf("Type(hidden implicit-host managed default) error = %v", err)
	}
	if len(host.typeReqs) != 0 {
		t.Fatalf("expected managed default type to avoid host backend, got %#v", host.typeReqs)
	}
	if len(node.typeReqs) != 1 || node.typeReqs[0].TabIndex != 0 || node.typeReqs[0].URL != "" || node.typeReqs[0].WaitMs != 0 || node.typeReqs[0].PostWaitMs != 250 || node.typeReqs[0].Selector != "input[name=q]" || node.typeReqs[0].Text != "agentx" || !node.typeReqs[0].Submit {
		t.Fatalf("expected managed default type to dispatch through node backend, got %#v", node.typeReqs)
	}
	if result.Backend != "node-type" || result.BrowserApp != "Chromium" || result.FinalURL != "https://node.example/form" || result.Value != "agentx" || !result.Submitted || result.Status != "typed" {
		t.Fatalf("unexpected managed default direct type result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendEvalRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				evalResult: BrowserEvalResult{
					Backend:    "node-eval",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/default",
					Title:      "Workbench",
					Result:     "managed default eval",
					Status:     "evaluated",
				},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{Evaluate: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{NodeBackend: node}, host, browserDefaultSubstrateAssessment{
		HostRuntime: host.runtimeInfo,
		HostRoute: browserConcreteRouteAssessment{
			Configured:     true,
			RouteAvailable: true,
			Route:          browserResolvedExecutionRoute{Backend: host, RuntimeInfo: host.runtimeInfo, Capabilities: browserCapabilitiesForConcreteBackend(host)},
		},
		DefaultRuntime: host.runtimeInfo,
	}).(browserRuntimeRouterBackend)

	result, err := router.Eval(context.Background(), BrowserEvalRequest{Script: "document.title", MaxChars: 24})
	if err != nil {
		t.Fatalf("Eval(hidden implicit-host managed default) error = %v", err)
	}
	if len(host.evalReqs) != 0 {
		t.Fatalf("expected managed default eval to avoid host backend, got %#v", host.evalReqs)
	}
	if len(node.evalReqs) != 1 || node.evalReqs[0].TabIndex != 0 || node.evalReqs[0].URL != "" || node.evalReqs[0].WaitMs != 0 || node.evalReqs[0].Script != "document.title" || node.evalReqs[0].MaxChars != 24 {
		t.Fatalf("expected managed default eval to dispatch through node backend, got %#v", node.evalReqs)
	}
	if result.Backend != "node-eval" || result.BrowserApp != "Chromium" || result.FinalURL != "https://node.example/default" || result.Result != "managed default eval" || result.Status != "evaluated" {
		t.Fatalf("unexpected managed default direct eval result: %#v", result)
	}
}

func TestBrowserRuntimeRouterBackendExtractURLRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	_, err := router.Extract(context.Background(), BrowserExtractRequest{URL: "https://93.184.216.34", MaxChars: 128})
	if err == nil || !strings.Contains(err.Error(), "browser backend extract requires explicit runtime_target because the default browser route falls back to the legacy system host path") {
		t.Fatalf("expected explicit-url implicit host extract to require explicit runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendClickTabTargetRequiresExplicitHostForImplicitFallback(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	_, err := router.Click(context.Background(), BrowserClickRequest{
		TabIndex: 2,
		Selector: "#submit",
	})
	if err == nil || !strings.Contains(err.Error(), `target "tab:2" requires explicit runtime_target=host`) {
		t.Fatalf("expected implicit host click tab target to require explicit host runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendClickURLRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	_, err := router.Click(context.Background(), BrowserClickRequest{
		URL:      "https://93.184.216.34",
		Selector: "#submit",
	})
	if err == nil || !strings.Contains(err.Error(), "browser backend click requires explicit runtime_target because the default browser route falls back to the legacy system host path") {
		t.Fatalf("expected explicit-url implicit host click to require explicit runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendEvalRequiresExplicitHostCurrentPageForImplicitFallback(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	_, err := router.Eval(context.Background(), BrowserEvalRequest{Script: "1+1"})
	if err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target=host or an explicit target/url") {
		t.Fatalf("expected implicit host eval to require explicit host runtime_target or target/url, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendResolveImplicitHostDefaultProfileRequiresExplicitRuntimeTargetEvenWhenExplicitHostFails(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	router.hostRoute = browserConcreteRouteAssessment{
		Configured:    true,
		FailureReason: "cached explicit host failure",
		FailureNote:   "cached explicit host failure",
	}
	if _, err := router.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{Profile: "default"}); err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target") {
		t.Fatalf("expected implicit legacy default-profile route to require explicit runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendResolveImplicitHostDefaultRouteRoutesManagedNodeWhenAvailable(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       BrowserCapabilities{},
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
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend: node,
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
		DefaultRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}).(browserRuntimeRouterBackend)

	route, err := router.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{})
	if err != nil {
		t.Fatalf("ResolveBrowserExecutionRoute(hidden implicit-host default request) error = %v", err)
	}
	if route.Backend != node || route.RuntimeInfo.Backend != "proxy" || route.RuntimeInfo.Profile != "workbench" || route.RuntimeInfo.Target != "node" {
		t.Fatalf("expected implicit-host default request to route through managed node lane, got %#v", route)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected first implicit-host default request to probe node once, got %d resolves", node.resolveCalls)
	}
	route, err = router.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{Profile: "default"})
	if err != nil {
		t.Fatalf("ResolveBrowserExecutionRoute(hidden implicit-host default-profile request) error = %v", err)
	}
	if route.Backend != node || route.RuntimeInfo.Profile != "workbench" || route.RuntimeInfo.Target != "node" {
		t.Fatalf("expected implicit-host default-profile request to reuse managed node lane, got %#v", route)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected repeated implicit-host default request to reuse managed-default cache, got %d resolves", node.resolveCalls)
	}
}

func TestBrowserRuntimeRouterBackendResolveImplicitHostDefaultRuntimeRouteRoutesManagedNodeWhenAvailable(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
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
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend: node,
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
		DefaultRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}).(browserRuntimeRouterBackend)

	info, err := router.ResolveBrowserRuntimeRoute(BrowserRuntimeInfo{})
	if err != nil {
		t.Fatalf("ResolveBrowserRuntimeRoute(hidden implicit-host default request) error = %v", err)
	}
	if info.Backend != "proxy" || info.Profile != "workbench" || info.Target != "node" {
		t.Fatalf("expected implicit-host default runtime-route request to route through managed node lane, got %#v", info)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected first implicit-host default runtime-route request to probe node once, got %d resolves", node.resolveCalls)
	}
	info, err = router.ResolveBrowserRuntimeRoute(BrowserRuntimeInfo{Profile: "default"})
	if err != nil {
		t.Fatalf("ResolveBrowserRuntimeRoute(hidden implicit-host default-profile request) error = %v", err)
	}
	if info.Backend != "proxy" || info.Profile != "workbench" || info.Target != "node" {
		t.Fatalf("expected implicit-host default-profile runtime-route request to reuse managed node lane, got %#v", info)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected repeated implicit-host default runtime-route request to reuse managed-default cache, got %d resolves", node.resolveCalls)
	}
}

func TestBrowserRuntimeRouterBackendResolveImplicitHostDefaultBackendRoutesManagedNodeWhenAvailable(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
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
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend: node,
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
		DefaultRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}).(browserRuntimeRouterBackend)

	backend, info, err := router.ResolveBrowserBackend(BrowserRuntimeInfo{})
	if err != nil {
		t.Fatalf("ResolveBrowserBackend(hidden implicit-host default request) error = %v", err)
	}
	if backend != node || info.Backend != "proxy" || info.Profile != "workbench" || info.Target != "node" {
		t.Fatalf("expected implicit-host default backend request to route through managed node lane, got backend=%T info=%#v", backend, info)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected first implicit-host default backend request to probe node once, got %d resolves", node.resolveCalls)
	}
	backend, info, err = router.ResolveBrowserBackend(BrowserRuntimeInfo{Profile: "default"})
	if err != nil {
		t.Fatalf("ResolveBrowserBackend(hidden implicit-host default-profile request) error = %v", err)
	}
	if backend != node || info.Backend != "proxy" || info.Profile != "workbench" || info.Target != "node" {
		t.Fatalf("expected implicit-host default-profile backend request to reuse managed node lane, got backend=%T info=%#v", backend, info)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected repeated implicit-host default backend request to reuse managed-default cache, got %d resolves", node.resolveCalls)
	}
}

func TestBrowserRuntimeRouterBackendBrowserRuntimeInfoUsesDynamicManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
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
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend: node,
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
		DefaultRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}).(browserRuntimeRouterBackend)

	if got := router.BrowserRuntimeInfo(); got.Backend != "proxy" || got.Profile != "workbench" || got.Target != "node" {
		t.Fatalf("expected router default runtime info to seed dynamic managed-default route, got %#v", got)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected router BrowserRuntimeInfo to seed generic managed-default once, got %d resolves", node.resolveCalls)
	}
	if got := router.BrowserRuntimeInfo(); got.Backend != "proxy" || got.Profile != "workbench" || got.Target != "node" {
		t.Fatalf("expected router default runtime info follow-up to reuse cached managed-default route, got %#v", got)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected router BrowserRuntimeInfo follow-up to reuse cached managed-default route, got %d resolves", node.resolveCalls)
	}
}

func TestBrowserRuntimeRouterBackendDefaultRuntimeInfoUsesDynamicManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
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
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend: node,
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
		DefaultRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}).(browserRuntimeRouterBackend)

	if got := router.defaultRuntimeInfo(); got.Backend != "proxy" || got.Profile != "workbench" || got.Target != "node" {
		t.Fatalf("expected router internal default runtime info to seed dynamic managed-default route, got %#v", got)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected router internal default runtime info to seed generic managed-default once, got %d resolves", node.resolveCalls)
	}
	if got := router.defaultRuntimeInfo(); got.Backend != "proxy" || got.Profile != "workbench" || got.Target != "node" {
		t.Fatalf("expected router internal default runtime info follow-up to reuse cached managed-default route, got %#v", got)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected router internal default runtime info follow-up to reuse cached managed-default route, got %d resolves", node.resolveCalls)
	}
}

func TestBrowserRuntimeRouterBackendDynamicRouteCacheInfoUsesManagedDefaultTargetBeforeHiddenImplicitHostFallback(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
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
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend: node,
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
		DefaultRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}).(browserRuntimeRouterBackend)

	if info, ok := router.dynamicRouteCacheInfo(BrowserRuntimeInfo{Profile: "relay"}); !ok || info.Target != "node" || info.Profile != "relay" {
		t.Fatalf("expected targetless profile cache to seed managed-default target, got %#v ok=%v", info, ok)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected targetless profile cache to seed generic managed-default once, got %d resolves", node.resolveCalls)
	}
	if info, ok := router.dynamicRouteCacheInfo(BrowserRuntimeInfo{Profile: "relay"}); !ok || info.Target != "node" || info.Profile != "relay" {
		t.Fatalf("expected targetless profile cache follow-up to reuse managed-default target, got %#v ok=%v", info, ok)
	}
}

func TestBrowserRuntimeRouterCachedRouteAssessmentUsesCachedImplicitManagedDefaultRoute(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
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
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend: node,
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
		DefaultRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}).(browserRuntimeRouterBackend)

	if _, ok := browserRuntimeRouterCachedRouteAssessment(router, BrowserRuntimeInfo{}); ok {
		t.Fatalf("expected generic cached route owner to stay empty before implicit managed-default resolve")
	}
	if _, err := router.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{}); err != nil {
		t.Fatalf("ResolveBrowserExecutionRoute(hidden implicit-host default request) error = %v", err)
	}
	assessment, ok := browserRuntimeRouterCachedRouteAssessment(router, BrowserRuntimeInfo{})
	if !ok || !assessment.RouteAvailable {
		t.Fatalf("expected generic cached route owner to reuse cached implicit managed-default route, got %#v ok=%v", assessment, ok)
	}
	if assessment.Route.RuntimeInfo.Backend != "proxy" || assessment.Route.RuntimeInfo.Profile != "workbench" || assessment.Route.RuntimeInfo.Target != "node" {
		t.Fatalf("unexpected cached implicit managed-default route assessment: %#v", assessment)
	}
	if _, ok := browserRuntimeRouterCachedRouteAssessment(router, BrowserRuntimeInfo{Profile: "default"}); ok {
		t.Fatalf("expected generic cached route owner to keep implicit managed-default out of default-profile follow-up cache")
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected generic cached route owner to reuse cached implicit managed-default route without extra resolve, got %d", node.resolveCalls)
	}
}

func TestBrowserRuntimeRouterCachedRouteAssessmentUsesCachedImplicitManagedDefaultNodeTargetLane(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
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
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend: node,
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
		DefaultRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}).(browserRuntimeRouterBackend)

	if _, err := router.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{}); err != nil {
		t.Fatalf("ResolveBrowserExecutionRoute(hidden implicit-host default request) error = %v", err)
	}
	assessment, ok := browserRuntimeRouterCachedRouteAssessment(router, BrowserRuntimeInfo{Target: "node"})
	if !ok || !assessment.RouteAvailable {
		t.Fatalf("expected target=node cached route owner to reuse cached implicit managed-default route, got %#v ok=%v", assessment, ok)
	}
	if assessment.Route.RuntimeInfo.Backend != "proxy" || assessment.Route.RuntimeInfo.Profile != "workbench" || assessment.Route.RuntimeInfo.Target != "node" {
		t.Fatalf("unexpected cached implicit managed-default node lane assessment: %#v", assessment)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected target=node cached route owner to reuse cached implicit managed-default route without extra resolve, got %d", node.resolveCalls)
	}
}

func TestBrowserRuntimeRouterBackendResolveImplicitHostAlternateProfileRequiresExplicitTarget(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)
	router, ok := backend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected backend factory to keep router owner active, got %T", backend)
	}
	_, err := router.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{Profile: "relay"})
	if err == nil || !strings.Contains(err.Error(), "requires an explicit runtime_target") {
		t.Fatalf("expected alternate implicit-host profile to require explicit runtime_target, got %v", err)
	}
}

func TestBrowserRuntimeRouterBackendResolveImplicitHostAlternateProfileRoutesManagedNodeWhenAvailable(t *testing.T) {
	host := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       BrowserCapabilities{},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			requested.Backend = "proxy"
			return requested, nil
		},
	}
	router := newBrowserRuntimeRouterBackendWithAssessment(BrowserToolOptions{
		NodeBackend: node,
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
		DefaultRuntime: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}).(browserRuntimeRouterBackend)

	route, err := router.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{Profile: "relay"})
	if err != nil {
		t.Fatalf("ResolveBrowserExecutionRoute(hidden implicit-host alternate profile) error = %v", err)
	}
	if route.Backend != node || route.RuntimeInfo.Backend != "proxy" || route.RuntimeInfo.Profile != "relay" || route.RuntimeInfo.Target != "node" {
		t.Fatalf("expected implicit-host alternate profile to route through managed node lane, got %#v", route)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected first hidden implicit-host alternate profile probe to resolve node once, got %d resolves", node.resolveCalls)
	}
	route, err = router.ResolveBrowserExecutionRoute(BrowserRuntimeInfo{Profile: "relay"})
	if err != nil {
		t.Fatalf("ResolveBrowserExecutionRoute(hidden implicit-host alternate profile repeat) error = %v", err)
	}
	if route.Backend != node || route.RuntimeInfo.Profile != "relay" || route.RuntimeInfo.Target != "node" {
		t.Fatalf("expected repeated hidden implicit-host alternate profile to reuse managed node route, got %#v", route)
	}
	if node.resolveCalls != 1 {
		t.Fatalf("expected repeated hidden implicit-host alternate profile to reuse cached node route, got %d resolves", node.resolveCalls)
	}
}
