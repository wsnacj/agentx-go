package tools

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type runtimeInfoCapabilityBrowserBackend struct {
	*fakeBrowserBackend
	runtimeInfo   BrowserRuntimeInfo
	capabilities  BrowserCapabilities
	routeSource   string
	routeEndpoint string
}

func (b *runtimeInfoCapabilityBrowserBackend) BrowserRuntimeInfo() BrowserRuntimeInfo {
	return b.runtimeInfo
}

func (b *runtimeInfoCapabilityBrowserBackend) BrowserCapabilities() BrowserCapabilities {
	return b.capabilities
}

func (b *runtimeInfoCapabilityBrowserBackend) browserDoctorRouteMetadata() browserDoctorRouteMetadata {
	if b == nil {
		return browserDoctorRouteMetadata{}
	}
	return browserDoctorRouteMetadata{
		Source:   strings.TrimSpace(b.routeSource),
		Endpoint: strings.TrimSpace(b.routeEndpoint),
	}
}

type runtimeInfoCapabilityRouteResolverBrowserBackend struct {
	*runtimeInfoCapabilityBrowserBackend
	resolve func(BrowserRuntimeInfo) (BrowserRuntimeInfo, error)
}

func (b *runtimeInfoCapabilityRouteResolverBrowserBackend) ResolveBrowserRuntimeRoute(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
	if b.resolve != nil {
		return b.resolve(requested)
	}
	return requested, nil
}

type countingRuntimeInfoCapabilityRouteResolverBrowserBackend struct {
	*runtimeInfoCapabilityBrowserBackend
	resolve      func(BrowserRuntimeInfo) (BrowserRuntimeInfo, error)
	resolveCalls int
}

func (b *countingRuntimeInfoCapabilityRouteResolverBrowserBackend) ResolveBrowserRuntimeRoute(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
	b.resolveCalls++
	if b.resolve != nil {
		return b.resolve(requested)
	}
	return requested, nil
}

type countingCapabilityRouteResolverBrowserBackend struct {
	*runtimeInfoCapabilityBrowserBackend
	resolve         func(BrowserRuntimeInfo) (BrowserRuntimeInfo, error)
	resolveCalls    int
	capabilityCalls int
}

func (b *countingCapabilityRouteResolverBrowserBackend) BrowserCapabilities() BrowserCapabilities {
	b.capabilityCalls++
	return b.runtimeInfoCapabilityBrowserBackend.BrowserCapabilities()
}

func (b *countingCapabilityRouteResolverBrowserBackend) ResolveBrowserRuntimeRoute(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
	b.resolveCalls++
	if b.resolve != nil {
		return b.resolve(requested)
	}
	return requested, nil
}

type countingExecutionRouteResolverBrowserBackend struct {
	*runtimeInfoBrowserBackend
	resolveExecution func(BrowserRuntimeInfo) (browserResolvedExecutionRoute, error)
	executionCalls   int
	runtimeCalls     int
	backendCalls     int
}

func (b *countingExecutionRouteResolverBrowserBackend) ResolveBrowserExecutionRoute(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
	b.executionCalls++
	if b.resolveExecution != nil {
		return b.resolveExecution(requested)
	}
	return browserResolvedExecutionRoute{
		Backend:     b,
		RuntimeInfo: requested,
	}, nil
}

func (b *countingExecutionRouteResolverBrowserBackend) ResolveBrowserRuntimeRoute(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
	b.runtimeCalls++
	return requested, nil
}

func (b *countingExecutionRouteResolverBrowserBackend) ResolveBrowserBackend(requested BrowserRuntimeInfo) (BrowserBackend, BrowserRuntimeInfo, error) {
	b.backendCalls++
	return b, requested, nil
}

type executionRouteCapabilityBrowserBackend struct {
	*runtimeInfoCapabilityBrowserBackend
	resolveExecution func(BrowserRuntimeInfo) (browserResolvedExecutionRoute, error)
}

func (b *executionRouteCapabilityBrowserBackend) ResolveBrowserExecutionRoute(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
	if b.resolveExecution != nil {
		return b.resolveExecution(requested)
	}
	return browserResolvedExecutionRoute{
		Backend:      b,
		RuntimeInfo:  requested,
		Capabilities: b.capabilities,
	}, nil
}

func (b *executionRouteCapabilityBrowserBackend) ResolveBrowserExecutionRouteForSession(_ context.Context, _ map[string]any, base BrowserRuntimeInfo, _ bool) (browserResolvedExecutionRoute, error) {
	return b.ResolveBrowserExecutionRoute(base)
}

type countingSessionRuntimeRouteResolverBrowserBackend struct {
	*runtimeInfoCapabilityBrowserBackend
	resolveSessionRuntime func(context.Context, map[string]any, BrowserRuntimeInfo, bool) (BrowserRuntimeInfo, error)
	sessionRuntimeCalls   int
	runtimeCalls          int
}

func (b *countingSessionRuntimeRouteResolverBrowserBackend) ResolveBrowserRuntimeRouteForSession(ctx context.Context, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool) (BrowserRuntimeInfo, error) {
	b.sessionRuntimeCalls++
	if b.resolveSessionRuntime != nil {
		return b.resolveSessionRuntime(ctx, params, base, hiddenImplicitHostDefaultBase)
	}
	return base, nil
}

func (b *countingSessionRuntimeRouteResolverBrowserBackend) ResolveBrowserRuntimeRoute(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
	b.runtimeCalls++
	return requested, nil
}

func TestDefaultBrowserRuntimeInfoHelper(t *testing.T) {
	got := DefaultBrowserRuntimeInfo()
	if got.Backend != "system" || got.Profile != "default" || got.Target != "host" {
		t.Fatalf("unexpected default browser runtime info: %#v", got)
	}
}

func TestNormalizeBrowserRuntimeInfoLowercasesAndTrims(t *testing.T) {
	got := normalizeBrowserRuntimeInfo(BrowserRuntimeInfo{
		Backend: " Proxy ",
		Profile: " Workbench ",
		Target:  " Node ",
	})
	if got.Backend != "proxy" || got.Profile != "workbench" || got.Target != "node" {
		t.Fatalf("unexpected normalized browser runtime info: %#v", got)
	}
}

func TestBrowserRuntimeInfoForBackendPrefersProvider(t *testing.T) {
	backend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: " Proxy ", Profile: " Isolated ", Target: " Node "},
	}
	got := browserRuntimeInfoForBackend(BrowserToolOptions{Backend: &fakeBrowserBackend{}}, backend)
	if got.Backend != "proxy" || got.Profile != "isolated" || got.Target != "node" {
		t.Fatalf("expected provider runtime info to win, got %#v", got)
	}
}

func TestBrowserRuntimeInfoForBackendFallsBackToCustomWhenExplicitBackendProvided(t *testing.T) {
	got := browserRuntimeInfoForBackend(BrowserToolOptions{Backend: &fakeBrowserBackend{}}, &fakeBrowserBackend{})
	if got.Backend != "custom" || got.Profile != "" || got.Target != "" {
		t.Fatalf("unexpected custom runtime fallback: %#v", got)
	}
}

func TestBrowserRuntimeInfoForConcreteBackendFillsMissingFieldsFromFallback(t *testing.T) {
	backend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy"},
	}
	got := browserRuntimeInfoForConcreteBackend(backend, BrowserRuntimeInfo{
		Backend: "system",
		Profile: "default",
		Target:  "host",
	})
	if got.Backend != "proxy" || got.Profile != "default" || got.Target != "host" {
		t.Fatalf("unexpected concrete backend runtime info: %#v", got)
	}
}

func TestBrowserRuntimeInfoForConcreteBackendFallsBackWithoutProvider(t *testing.T) {
	got := browserRuntimeInfoForConcreteBackend(&fakeBrowserBackend{}, BrowserRuntimeInfo{
		Backend: " System ",
		Profile: " Default ",
		Target:  " Host ",
	})
	if got.Backend != "system" || got.Profile != "default" || got.Target != "host" {
		t.Fatalf("unexpected fallback runtime info: %#v", got)
	}
}

func TestCurrentBrowserPlatformLaneDefaultRuntimeMatchesLegacyDefault(t *testing.T) {
	lane := currentBrowserPlatformLane(BrowserToolOptions{})
	got := lane.DefaultRuntime()
	want := defaultBrowserRuntimeInfo()
	if got != want {
		t.Fatalf("platform lane default runtime = %#v, want %#v", got, want)
	}
	switch runtime.GOOS {
	case "darwin", "linux":
		if lane.Name() != runtime.GOOS {
			t.Fatalf("platform lane name = %q, want %q", lane.Name(), runtime.GOOS)
		}
	default:
		if lane.Name() != "default" {
			t.Fatalf("platform lane name = %q, want default", lane.Name())
		}
	}
}

func TestCurrentBrowserPlatformLaneDefaultCapabilitiesFailClosedWithoutBackend(t *testing.T) {
	lane := currentBrowserPlatformLane(BrowserToolOptions{})
	got := lane.DefaultCapabilities()
	want := BrowserCapabilities{}
	if got != want {
		t.Fatalf("platform lane default capabilities without backend = %#v, want %#v", got, want)
	}
}

func TestBrowserDefaultExecutionLaneForOptionsWrapsDefaultRouteOwner(t *testing.T) {
	opts := BrowserToolOptions{Root: t.TempDir()}
	lane := browserDefaultExecutionLaneForOptions(opts)
	preview := browserDefaultRuntimePreviewForToolOptions(opts)
	if lane.Runtime != preview.LogicalDefaultRoute {
		t.Fatalf("default execution lane runtime = %#v, want %#v", lane.Runtime, preview.LogicalDefaultRoute)
	}
	if fmt.Sprintf("%T", lane.Backend) != fmt.Sprintf("%T", preview.EffectiveBackend) {
		t.Fatalf("default execution lane backend mismatch: got %T want %T", lane.Backend, preview.EffectiveBackend)
	}
	if lane.Capabilities != preview.RegistrationCapabilities {
		t.Fatalf("default execution lane capabilities = %#v, want %#v", lane.Capabilities, preview.RegistrationCapabilities)
	}
	if lane.Substrate != BrowserSubstratePosture(preview.LogicalDefaultRoute.Backend, preview.LogicalDefaultRoute.Target) {
		t.Fatalf("default execution lane substrate = %q", lane.Substrate)
	}
}

func TestBrowserDefaultRuntimePreviewForToolOptionsCarriesRegistrationCapabilities(t *testing.T) {
	opts := BrowserToolOptions{Root: t.TempDir()}
	preview := browserDefaultRuntimePreviewForToolOptions(opts)
	want := browserCapabilitiesForRegistrationWithBackend(preview.EffectiveBackend, preview.SubstrateAssessment)
	if preview.RegistrationCapabilities != want {
		t.Fatalf("preview registration capabilities = %#v, want %#v", preview.RegistrationCapabilities, want)
	}
}

func TestBrowserDefaultRuntimePreviewForToolOptionsHidesImplicitLegacyHostFallback(t *testing.T) {
	preview := browserDefaultRuntimePreviewForToolOptions(BrowserToolOptions{})
	if preview.LogicalDefaultRoute != DefaultBrowserRuntimeInfo() {
		t.Fatalf("expected logical default route to remain legacy host, got %#v", preview.LogicalDefaultRoute)
	}
	if preview.VisibleDefaultRoute != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected visible default route to hide implicit legacy host fallback, got %#v", preview.VisibleDefaultRoute)
	}
	if !preview.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected hidden implicit host default base to be true")
	}
	if preview.SubstrateSummary.DefaultRoute != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected substrate summary to hide implicit legacy host default route, got %#v", preview.SubstrateSummary)
	}
}

func TestBrowserDefaultRuntimePreviewForToolOptionsHidesInjectedImplicitHostFallback(t *testing.T) {
	implicit := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        DefaultBrowserRuntimeInfo(),
		capabilities:       defaultBrowserCapabilities(),
	}
	preview := browserDefaultRuntimePreviewForToolOptions(BrowserToolOptions{ImplicitHostBackend: implicit})
	if preview.EffectiveBackend == nil {
		t.Fatal("expected injected implicit host backend to remain executable")
	}
	if preview.LogicalDefaultRoute != DefaultBrowserRuntimeInfo() {
		t.Fatalf("expected logical default route to remain legacy host, got %#v", preview.LogicalDefaultRoute)
	}
	if preview.VisibleDefaultRoute != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected visible default route to hide injected implicit host fallback, got %#v", preview.VisibleDefaultRoute)
	}
	if !preview.HiddenImplicitHostDefaultBase {
		t.Fatal("expected injected implicit host default base to remain hidden")
	}
}

func TestBrowserDefaultRuntimePreviewForToolOptionsUsesDynamicManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
	preview := browserDefaultRuntimePreviewForToolOptions(BrowserToolOptions{
		NodeBackend: node,
	})
	if node.resolveCalls != 2 {
		t.Fatalf("expected options preview to seed generic managed-default once after static assessment, got %d resolves", node.resolveCalls)
	}
	want := BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}
	if preview.LogicalDefaultRoute != want {
		t.Fatalf("expected logical default route to use dynamic managed-default, got %#v", preview.LogicalDefaultRoute)
	}
	if preview.VisibleDefaultRoute != want {
		t.Fatalf("expected visible default route to expose dynamic managed-default, got %#v", preview.VisibleDefaultRoute)
	}
	if preview.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected dynamic managed-default route to clear hidden implicit host default base")
	}
	if preview.SubstrateSummary.DefaultRoute != want {
		t.Fatalf("expected substrate summary to reuse dynamic managed-default route, got %#v", preview.SubstrateSummary)
	}
	if preview.SubstrateAssessment.DefaultRuntime != want {
		t.Fatalf("expected substrate assessment to reuse dynamic managed-default runtime, got %#v", preview.SubstrateAssessment.DefaultRuntime)
	}
}

func TestBrowserDefaultRuntimePreviewForToolOptionsUsesDynamicManagedRuntimeDefaultRouteWhenOnlyRuntimeToolEnabled(t *testing.T) {
	node := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       BrowserCapabilities{RuntimeStatus: true},
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
	preview := browserDefaultRuntimePreviewForToolOptions(BrowserToolOptions{
		NodeBackend:  node,
		EnabledTools: []string{"browser_runtime"},
	})
	want := BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}
	if preview.LogicalDefaultRoute != want {
		t.Fatalf("expected runtime-only logical default route to use managed runtime default, got %#v", preview.LogicalDefaultRoute)
	}
	if preview.VisibleDefaultRoute != want {
		t.Fatalf("expected runtime-only visible default route to expose managed runtime default, got %#v", preview.VisibleDefaultRoute)
	}
	if preview.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected runtime-only managed runtime default to clear hidden implicit host fallback")
	}
	if preview.SubstrateSummary.DefaultRoute != want {
		t.Fatalf("expected runtime-only substrate summary to reuse managed runtime default, got %#v", preview.SubstrateSummary)
	}
	if preview.SubstrateSummary.SelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected runtime-only substrate summary to prefer node over legacy host, got %#v", preview.SubstrateSummary)
	}
	if preview.SubstrateAssessment.DefaultRuntime != want {
		t.Fatalf("expected runtime-only substrate assessment to reuse managed runtime default, got %#v", preview.SubstrateAssessment.DefaultRuntime)
	}
}

func TestBrowserDefaultRuntimePreviewForDispatchOptionsUsesNormalizedDispatchOptions(t *testing.T) {
	preview := browserDefaultRuntimePreviewForDispatchOptions(BrowserToolOptions{
		TimeoutMs:         4_321,
		AllowPrivateHosts: true,
		AllowPorts:        []int{443},
	}, outboundNetworkPolicy{
		allowPrivate:  true,
		allowPortsSet: true,
		allowPorts:    map[int]bool{443: true},
		denyPorts:     map[int]bool{22: true},
	}, 4_321)
	router, ok := preview.EffectiveBackend.(browserRuntimeRouterBackend)
	if !ok {
		t.Fatalf("expected preview effective backend to keep router backend, got %T", preview.EffectiveBackend)
	}
	if router.hostTimeoutMs != 4_321 {
		t.Fatalf("expected preview to preserve timeout, got %d", router.hostTimeoutMs)
	}
	if !router.hostPolicy.allowPrivate {
		t.Fatalf("expected preview to preserve allowPrivateHosts")
	}
	if !router.hostPolicy.allowPortsSet || !router.hostPolicy.allowPorts[443] {
		t.Fatalf("expected preview to preserve allowPorts, got %#v", router.hostPolicy.allowPorts)
	}
	if !router.hostPolicy.denyPorts[22] {
		t.Fatalf("expected preview to preserve denyPorts, got %#v", router.hostPolicy.denyPorts)
	}
}

func TestResolveBrowserExecutionLaneForRegistrationPageActionWrapsDispatch(t *testing.T) {
	backend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilities{Open: true, Navigate: true, Extract: true},
	}
	ctx := browserRegistrationContext{
		opts:         BrowserToolOptions{Root: t.TempDir()},
		backend:      backend,
		capabilities: backend.capabilities,
		substrateAssessment: browserDefaultSubstrateAssessment{
			DefaultRuntime: backend.runtimeInfo,
		},
		substrateSummary: BrowserWorkbenchSubstrateSummary{
			DefaultRoute: backend.runtimeInfo,
		},
	}
	lane, dispatch, err := resolveBrowserExecutionLaneForRegistrationPageAction(
		ctx,
		context.Background(),
		map[string]any{},
		browserRegistrationPageActionDispatchOptions{},
	)
	if err != nil {
		t.Fatalf("resolveBrowserExecutionLaneForRegistrationPageAction error = %v", err)
	}
	if lane.Runtime != dispatch.RuntimeInfo {
		t.Fatalf("execution lane runtime = %#v, want dispatch runtime %#v", lane.Runtime, dispatch.RuntimeInfo)
	}
	if lane.Backend != dispatch.Backend {
		t.Fatalf("execution lane backend mismatch: got %T want %T", lane.Backend, dispatch.Backend)
	}
	if lane.Capabilities != browserCapabilitiesForConcreteBackend(dispatch.Backend) {
		t.Fatalf("execution lane capabilities = %#v", lane.Capabilities)
	}
	if lane.Substrate != BrowserSubstratePosture(dispatch.RuntimeInfo.Backend, dispatch.RuntimeInfo.Target) {
		t.Fatalf("execution lane substrate = %q", lane.Substrate)
	}
}

func TestResolveBrowserExecutionLaneForRegistrationPageActionUsesManagedRouteCapabilities(t *testing.T) {
	routeBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		capabilities:       BrowserCapabilities{Screenshot: true},
	}
	backend := &executionRouteCapabilityBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "default", Target: "node"},
			capabilities:       BrowserCapabilities{Extract: true, Evaluate: true},
		},
		resolveExecution: func(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
			return browserResolvedExecutionRoute{
				Backend:      routeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
				Capabilities: BrowserCapabilities{Screenshot: true},
			}, nil
		},
	}
	ctx := browserRegistrationContext{
		opts:         BrowserToolOptions{Root: t.TempDir()},
		backend:      backend,
		capabilities: backend.capabilities,
		substrateAssessment: browserDefaultSubstrateAssessment{
			DefaultRuntime: backend.runtimeInfo,
		},
		substrateSummary: BrowserWorkbenchSubstrateSummary{
			DefaultRoute: backend.runtimeInfo,
		},
	}

	lane, dispatch, err := resolveBrowserExecutionLaneForRegistrationPageAction(
		ctx,
		context.Background(),
		map[string]any{},
		browserRegistrationPageActionDispatchOptions{
			UseManagedRoute:               true,
			UseManagedDefaultDispatchBase: true,
		},
	)
	if err != nil {
		t.Fatalf("resolveBrowserExecutionLaneForRegistrationPageAction(managed route) error = %v", err)
	}
	if lane.Runtime != dispatch.Route.RuntimeInfo {
		t.Fatalf("execution lane runtime = %#v, want managed route runtime %#v", lane.Runtime, dispatch.Route.RuntimeInfo)
	}
	if lane.Backend != dispatch.Route.Backend {
		t.Fatalf("execution lane backend mismatch: got %T want %T", lane.Backend, dispatch.Route.Backend)
	}
	if lane.Capabilities != dispatch.Route.Capabilities {
		t.Fatalf("execution lane capabilities = %#v, want managed route capabilities %#v", lane.Capabilities, dispatch.Route.Capabilities)
	}
	if lane.Capabilities == browserCapabilitiesForConcreteBackend(backend) {
		t.Fatalf("execution lane capabilities unexpectedly fell back to registration backend capabilities: %#v", lane.Capabilities)
	}
}

func TestResolveBrowserExecutionLaneForRegistrationDispatchWrapsDispatch(t *testing.T) {
	backend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilities{Open: true, Navigate: true, Tabs: true},
	}
	ctx := browserRegistrationContext{
		opts:         BrowserToolOptions{Root: t.TempDir()},
		backend:      backend,
		capabilities: backend.capabilities,
		substrateAssessment: browserDefaultSubstrateAssessment{
			DefaultRuntime: backend.runtimeInfo,
		},
		substrateSummary: BrowserWorkbenchSubstrateSummary{
			DefaultRoute: backend.runtimeInfo,
		},
	}
	lane, dispatch, err := resolveBrowserExecutionLaneForRegistrationDispatch(
		ctx,
		context.Background(),
		map[string]any{},
		browserRegistrationPageActionDispatchOptions{},
	)
	if err != nil {
		t.Fatalf("resolveBrowserExecutionLaneForRegistrationDispatch error = %v", err)
	}
	want := browserExecutionLaneForRegistrationDispatch(ctx, dispatch)
	if lane != want {
		t.Fatalf("execution lane = %#v, want %#v", lane, want)
	}
}

func TestResolveBrowserExecutionBackendPrefersExecutionRouteResolver(t *testing.T) {
	backend := &countingExecutionRouteResolverBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		},
	}
	backend.resolveExecution = func(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
		return browserResolvedExecutionRoute{
			Backend:     backend,
			RuntimeInfo: BrowserRuntimeInfo{Backend: "proxy"},
		}, nil
	}
	routedBackend, routedInfo, err := resolveBrowserExecutionBackend(
		map[string]any{"runtime_target": "node", "profile": "relay"},
		DefaultBrowserRuntimeInfo(),
		backend,
	)
	if err != nil {
		t.Fatalf("resolveBrowserExecutionBackend error = %v", err)
	}
	if routedBackend != backend {
		t.Fatalf("expected execution route resolver backend to be returned, got %#v", routedBackend)
	}
	if routedInfo.Backend != "proxy" || routedInfo.Profile != "relay" || routedInfo.Target != "node" {
		t.Fatalf("unexpected routed runtime info: %#v", routedInfo)
	}
	if backend.executionCalls != 1 {
		t.Fatalf("expected execution route resolver to run once, got %d", backend.executionCalls)
	}
	if backend.runtimeCalls != 0 || backend.backendCalls != 0 {
		t.Fatalf("expected execution backend resolution to skip runtime/backend resolver fallbacks, runtime=%d backend=%d", backend.runtimeCalls, backend.backendCalls)
	}
}

func TestResolveBrowserExecutionRouteForSessionPrefersExecutionRouteResolver(t *testing.T) {
	backend := &countingExecutionRouteResolverBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "default", Target: "host"},
		},
	}
	backend.resolveExecution = func(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
		if requested != (BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "node"}) {
			t.Fatalf("expected session preview request to reach execution route resolver unchanged, got %#v", requested)
		}
		return browserResolvedExecutionRoute{
			Backend:      backend,
			RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy"},
			Capabilities: BrowserCapabilitiesForActKinds([]string{"click"}),
		}, nil
	}
	route, err := resolveBrowserExecutionRouteForSession(
		context.Background(),
		nil,
		nil,
		map[string]any{"runtime_target": "node", "profile": "workbench"},
		BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "default", Target: "host"},
		false,
		backend,
	)
	if err != nil {
		t.Fatalf("resolveBrowserExecutionRouteForSession error = %v", err)
	}
	if route.Backend != backend {
		t.Fatalf("expected execution route resolver backend to win for session preview, got %#v", route.Backend)
	}
	if route.RuntimeInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("unexpected session preview execution route runtime info: %#v", route.RuntimeInfo)
	}
	if !route.Capabilities.SupportsActKind("click") || route.Capabilities.SupportsActKind("screenshot") {
		t.Fatalf("expected session preview execution route capabilities to stay concrete, got %#v", route.Capabilities)
	}
	if backend.executionCalls != 1 {
		t.Fatalf("expected execution route resolver to run once for session preview, got %d", backend.executionCalls)
	}
	if backend.runtimeCalls != 0 || backend.backendCalls != 0 {
		t.Fatalf("expected session preview execution route to skip runtime/backend resolver fallbacks, runtime=%d backend=%d", backend.runtimeCalls, backend.backendCalls)
	}
}

func TestResolveConcreteBrowserExecutionRoutePrefersExecutionRouteResolver(t *testing.T) {
	backend := &countingExecutionRouteResolverBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		},
	}
	backend.resolveExecution = func(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
		if requested != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
			t.Fatalf("expected concrete route request to reach execution route resolver unchanged, got %#v", requested)
		}
		return browserResolvedExecutionRoute{
			Backend:      backend,
			RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy"},
			Capabilities: BrowserCapabilitiesForActKinds([]string{"click"}),
		}, nil
	}
	route, err := resolveConcreteBrowserExecutionRoute(
		backend,
		BrowserRuntimeInfo{Backend: "node", Profile: "isolated", Target: "node"},
		BrowserRuntimeInfo{Profile: "workbench", Target: "node"},
	)
	if err != nil {
		t.Fatalf("resolveConcreteBrowserExecutionRoute error = %v", err)
	}
	if route.Backend != backend {
		t.Fatalf("expected concrete route execution resolver backend to win, got %#v", route.Backend)
	}
	if route.RuntimeInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("unexpected concrete route runtime info: %#v", route.RuntimeInfo)
	}
	if !route.Capabilities.SupportsActKind("click") || route.Capabilities.SupportsActKind("screenshot") {
		t.Fatalf("expected concrete route capabilities to stay concrete, got %#v", route.Capabilities)
	}
	if backend.executionCalls != 1 {
		t.Fatalf("expected concrete route execution resolver to run once, got %d", backend.executionCalls)
	}
	if backend.runtimeCalls != 0 || backend.backendCalls != 0 {
		t.Fatalf("expected concrete route resolution to skip runtime/backend resolver fallbacks, runtime=%d backend=%d", backend.runtimeCalls, backend.backendCalls)
	}
}

func TestResolveBrowserRuntimeRouteUsesRuntimeRouteResolverForDefaultRequest(t *testing.T) {
	backend := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			if requested.Backend != "system" || requested.Profile != "default" || requested.Target != "host" {
				t.Fatalf("expected default request to reach runtime route resolver unchanged, got %#v", requested)
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}

	got, err := resolveBrowserRuntimeRoute(map[string]any{}, DefaultBrowserRuntimeInfo(), backend)
	if err != nil {
		t.Fatalf("resolveBrowserRuntimeRoute(default request) error = %v", err)
	}
	if got.Backend != "proxy" || got.Profile != "workbench" || got.Target != "node" {
		t.Fatalf("expected default request to reuse runtime route resolver, got %#v", got)
	}
	if backend.resolveCalls != 1 {
		t.Fatalf("expected runtime route resolver to run once for default request, got %d", backend.resolveCalls)
	}
}

func TestResolveBrowserExecutionRouteUsesRuntimeRouteResolverForDefaultRequest(t *testing.T) {
	backend := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"extract"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			if requested.Backend != "system" || requested.Profile != "default" || requested.Target != "host" {
				t.Fatalf("expected default execution request to reach runtime route resolver unchanged, got %#v", requested)
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}

	route, err := resolveBrowserExecutionRoute(map[string]any{}, DefaultBrowserRuntimeInfo(), backend)
	if err != nil {
		t.Fatalf("resolveBrowserExecutionRoute(default request) error = %v", err)
	}
	if route.Backend != backend {
		t.Fatalf("expected runtime route fallback to keep concrete backend, got %T", route.Backend)
	}
	if route.RuntimeInfo.Backend != "proxy" || route.RuntimeInfo.Profile != "workbench" || route.RuntimeInfo.Target != "node" {
		t.Fatalf("expected default execution request to reuse runtime route resolver, got %#v", route.RuntimeInfo)
	}
	if backend.resolveCalls != 1 {
		t.Fatalf("expected runtime route resolver to run once for default execution request, got %d", backend.resolveCalls)
	}
}

func TestResolveBrowserRuntimeRouteUsesRouterDefaultRequestBaseForPromotedDefaultRoute(t *testing.T) {
	nodeBackend := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
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
	backend := newBrowserBackend(BrowserToolOptions{NodeBackend: nodeBackend}, outboundNetworkPolicy{}, 1_500)

	got, err := resolveBrowserRuntimeRoute(map[string]any{}, DefaultBrowserRuntimeInfo(), backend)
	if err != nil {
		t.Fatalf("resolveBrowserRuntimeRoute(router default request) error = %v", err)
	}
	if got != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected generic runtime route helper to reuse router default request base, got %#v", got)
	}
}

func TestResolveBrowserExecutionBackendUsesRouterDefaultRequestBaseForPromotedDefaultRoute(t *testing.T) {
	nodeBackend := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
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
	backend := newBrowserBackend(BrowserToolOptions{NodeBackend: nodeBackend}, outboundNetworkPolicy{}, 1_500)

	routedBackend, routedInfo, err := resolveBrowserExecutionBackend(map[string]any{}, DefaultBrowserRuntimeInfo(), backend)
	if err != nil {
		t.Fatalf("resolveBrowserExecutionBackend(router default request) error = %v", err)
	}
	if routedBackend != nodeBackend {
		t.Fatalf("expected generic execution backend helper to resolve concrete node backend, got %T", routedBackend)
	}
	if routedInfo != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected generic execution backend helper to reuse router default request base, got %#v", routedInfo)
	}
}

func TestResolveBrowserExecutionRouteUsesRouterDefaultRequestBaseForHiddenImplicitHostDefault(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)

	_, err := resolveBrowserExecutionRoute(map[string]any{}, DefaultBrowserRuntimeInfo(), backend)
	if err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target") {
		t.Fatalf("expected generic execution route helper to keep hidden implicit host default behind explicit runtime_target gate, got %v", err)
	}
}

func TestBrowserRuntimePreviewSessionSelectionsForExecutionUsesRouterDefaultRequestPreview(t *testing.T) {
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		capabilities:       fullBrowserCapabilities(),
	}
	nodeBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		capabilities:       fullBrowserCapabilities(),
	}
	backend := browserRuntimeRouterBackend{
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

	preview := browserRuntimePreviewSessionSelectionsForExecution(
		context.Background(),
		NewBrowserSessionStateRegistry(),
		NewBrowserSessionRegistry(),
		map[string]any{"action": "status"},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		backend,
	)
	if preview.Base != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected session execution preview to reuse router managed default dispatch base, got %#v", preview.Base)
	}
	if preview.DefaultCandidateRoute != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected session execution preview to preserve router default candidate route, got %#v", preview.DefaultCandidateRoute)
	}
	if preview.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected session execution preview to clear hidden implicit host flag for promoted managed default route")
	}
	if preview.SessionSelectionPreview.RequestedRuntimeTarget != "" {
		t.Fatalf("expected targetless default request preview not to invent runtime_target, got %#v", preview.SessionSelectionPreview)
	}
}

func TestBrowserRuntimePreviewSessionSelectionsForExecutionKeepsHiddenImplicitHostDefaultWhenManagedRouteIsUnavailable(t *testing.T) {
	backend := newBrowserBackend(BrowserToolOptions{}, outboundNetworkPolicy{}, 1_500)

	preview := browserRuntimePreviewSessionSelectionsForExecution(
		context.Background(),
		NewBrowserSessionStateRegistry(),
		NewBrowserSessionRegistry(),
		map[string]any{"action": "status"},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		backend,
	)
	if preview.Base != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected session execution preview to hide implicit legacy host default base, got %#v", preview.Base)
	}
	if preview.DefaultCandidateRoute != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected session execution preview to keep default candidate route empty without managed lanes, got %#v", preview.DefaultCandidateRoute)
	}
	if !preview.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected session execution preview to keep hidden implicit host flag when managed route is unavailable")
	}
}

func TestBrowserRuntimePreviewSessionSelectionsForExecutionSurfacesHiddenManagedDefaultCandidateRoute(t *testing.T) {
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

	preview := browserRuntimePreviewSessionSelectionsForExecution(
		context.Background(),
		NewBrowserSessionStateRegistry(),
		NewBrowserSessionRegistry(),
		map[string]any{"action": "status"},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		backend,
	)
	if preview.Base != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected session execution preview to keep hidden implicit host dispatch base, got %#v", preview.Base)
	}
	if preview.DefaultCandidateRoute != (BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}) {
		t.Fatalf("expected session execution preview to surface hidden managed default candidate route, got %#v", preview.DefaultCandidateRoute)
	}
	if preview.DefaultCandidateDescriptor != (browserRuntimeRouteDescriptor{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		Source:        "managed_browserd",
		Endpoint:      "http://127.0.0.1:43123",
	}) {
		t.Fatalf("expected session execution preview to preserve hidden managed default candidate provenance, got %#v", preview.DefaultCandidateDescriptor)
	}
	if !preview.HiddenImplicitHostDefaultBase {
		t.Fatalf("expected session execution preview to preserve hidden implicit host flag when managed lane stays hidden")
	}
}

func TestBrowserResolveRuntimeRequestFromSessionPreviewUsesManagedCurrentRoute(t *testing.T) {
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
	callCtx := WithToolSessionID(context.Background(), "resolve-runtime-request-from-session-preview-managed-current")
	sessionRegistry.TrackCurrentTarget("resolve-runtime-request-from-session-preview-managed-current", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	params := map[string]any{"action": "status"}
	preview := browserRuntimePreviewSessionSelections(
		callCtx,
		stateRegistry,
		sessionRegistry,
		params,
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
	)

	request, err := browserResolveRuntimeRequestFromSessionPreview(
		params,
		preview,
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		router,
	)
	if err != nil {
		t.Fatalf("browserResolveRuntimeRequestFromSessionPreview(managed current) error = %v", err)
	}
	if request.Requested != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected managed current route to drive requested runtime info, got %#v", request.Requested)
	}
	if request.ExplicitRuntimeTarget {
		t.Fatalf("expected managed current preview to stay implicit, got %#v", request)
	}
	if request.HiddenRequestedRuntimeTarget {
		t.Fatalf("expected resolvable managed current route not to stay hidden, got %#v", request)
	}
}

func TestBrowserResolveRuntimeRequestFromSessionPreviewKeepsHiddenManagedCurrentTarget(t *testing.T) {
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
	callCtx := WithToolSessionID(context.Background(), "resolve-runtime-request-from-session-preview-stale-managed-current")
	sessionRegistry.TrackCurrentTarget("resolve-runtime-request-from-session-preview-stale-managed-current", BrowserSessionTarget{
		ID:         "stale-node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	params := map[string]any{"action": "status"}
	preview := browserRuntimePreviewSessionSelections(
		callCtx,
		stateRegistry,
		sessionRegistry,
		params,
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
	)

	request, err := browserResolveRuntimeRequestFromSessionPreview(
		params,
		preview,
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		router,
	)
	if err != nil {
		t.Fatalf("browserResolveRuntimeRequestFromSessionPreview(stale managed current) error = %v", err)
	}
	if request.Requested != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected stale managed current route to keep targetless visible request, got %#v", request.Requested)
	}
	if request.ExplicitRuntimeTarget {
		t.Fatalf("expected stale managed current preview to stay implicit, got %#v", request)
	}
	if !request.HiddenRequestedRuntimeTarget {
		t.Fatalf("expected stale managed current route to stay hidden behind explicit runtime_target gate, got %#v", request)
	}
}

func TestResolveBrowserExecutionRouteForSessionDelegatesManagedCurrentRouteToRouterOwner(t *testing.T) {
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
	callCtx := WithToolSessionID(context.Background(), "resolve-browser-execution-route-session-router-owner")
	sessionRegistry.TrackCurrentTarget("resolve-browser-execution-route-session-router-owner", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	route, err := resolveBrowserExecutionRouteForSession(
		callCtx,
		stateRegistry,
		sessionRegistry,
		map[string]any{},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		router,
	)
	if err != nil {
		t.Fatalf("resolveBrowserExecutionRouteForSession(router owner) error = %v", err)
	}
	if route.Backend != node {
		t.Fatalf("expected session-aware resolve to delegate managed current route to router owner, got %T", route.Backend)
	}
	if route.RuntimeInfo.Backend != "proxy" || route.RuntimeInfo.Profile != "workbench" || route.RuntimeInfo.Target != "node" {
		t.Fatalf("unexpected session-aware router-owned route: %#v", route.RuntimeInfo)
	}
}

func TestResolveBrowserExecutionRouteForSessionKeepsImplicitHostGateForUnresolvableManagedCurrentRoute(t *testing.T) {
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
	callCtx := WithToolSessionID(context.Background(), "resolve-browser-execution-route-session-router-owner-stale-managed-current")
	sessionRegistry.TrackCurrentTarget("resolve-browser-execution-route-session-router-owner-stale-managed-current", BrowserSessionTarget{
		ID:         "stale-node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	_, err := resolveBrowserExecutionRouteForSession(
		callCtx,
		stateRegistry,
		sessionRegistry,
		map[string]any{},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		router,
	)
	if err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target") {
		t.Fatalf("expected stale managed current route to stay behind implicit host explicit runtime_target gate, got %v", err)
	}
}

func TestResolveBrowserExecutionRouteForSessionRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
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
	backend := newBrowserBackend(BrowserToolOptions{
		NodeBackend:          node,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: stateRegistry,
	}, outboundNetworkPolicy{}, 1_500)

	route, err := resolveBrowserExecutionRouteForSession(
		WithToolSessionID(context.Background(), "resolve-browser-execution-route-session-managed-default-hidden-implicit-host"),
		stateRegistry,
		sessionRegistry,
		map[string]any{"action": "status"},
		BrowserRuntimeInfo{},
		true,
		backend,
	)
	if err != nil {
		t.Fatalf("resolveBrowserExecutionRouteForSession(hidden implicit-host managed default) error = %v", err)
	}
	if route.RuntimeInfo.Backend != "proxy" || route.RuntimeInfo.Profile != "workbench" || route.RuntimeInfo.Target != "node" {
		t.Fatalf("expected session-aware execution route to reuse managed default route before implicit host fallback, got %#v", route.RuntimeInfo)
	}
}

func TestResolveBrowserRuntimeRouteForSessionDelegatesManagedCurrentRouteToRouterOwner(t *testing.T) {
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
	callCtx := WithToolSessionID(context.Background(), "resolve-browser-runtime-route-session-router-owner")
	sessionRegistry.TrackCurrentTarget("resolve-browser-runtime-route-session-router-owner", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	runtimeInfo, err := resolveBrowserRuntimeRouteForSession(
		callCtx,
		stateRegistry,
		sessionRegistry,
		map[string]any{},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		router,
	)
	if err != nil {
		t.Fatalf("resolveBrowserRuntimeRouteForSession(router owner) error = %v", err)
	}
	if runtimeInfo.Backend != "proxy" || runtimeInfo.Profile != "workbench" || runtimeInfo.Target != "node" {
		t.Fatalf("unexpected session-aware router-owned runtime route: %#v", runtimeInfo)
	}
}

func TestResolveBrowserExecutionRouteForSessionFallsBackToSessionRuntimeRouteResolver(t *testing.T) {
	backend := &countingSessionRuntimeRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"extract"}),
		},
		resolveSessionRuntime: func(_ context.Context, _ map[string]any, _ BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool) (BrowserRuntimeInfo, error) {
			if !hiddenImplicitHostDefaultBase {
				t.Fatalf("expected hidden implicit-host base to propagate into session runtime route resolver")
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}

	route, err := resolveBrowserExecutionRouteForSession(
		context.Background(),
		NewBrowserSessionStateRegistry(),
		NewBrowserSessionRegistry(),
		map[string]any{},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		backend,
	)
	if err != nil {
		t.Fatalf("resolveBrowserExecutionRouteForSession(session runtime fallback) error = %v", err)
	}
	if backend.sessionRuntimeCalls != 1 || backend.runtimeCalls != 0 {
		t.Fatalf("expected session runtime route resolver to handle fallback, got session=%d runtime=%d", backend.sessionRuntimeCalls, backend.runtimeCalls)
	}
	if route.Backend != backend {
		t.Fatalf("expected fallback route to keep concrete backend, got %T", route.Backend)
	}
	if route.RuntimeInfo.Backend != "proxy" || route.RuntimeInfo.Profile != "workbench" || route.RuntimeInfo.Target != "node" {
		t.Fatalf("unexpected session runtime fallback route: %#v", route.RuntimeInfo)
	}
}

func TestResolveBrowserRuntimeRouteForSessionKeepsVisibleBaseForUnresolvableManagedCurrentRouteWithoutSessionOwner(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	backend := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo: BrowserRuntimeInfo{
				Backend: "system",
				Profile: "default",
				Target:  "host",
			},
			capabilities: fullBrowserCapabilities(),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if requested.Target == "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return requested, nil
		},
	}
	callCtx := WithToolSessionID(context.Background(), "resolve-browser-runtime-route-session-fallback-stale-managed-current")
	sessionRegistry.TrackCurrentTarget("resolve-browser-runtime-route-session-fallback-stale-managed-current", BrowserSessionTarget{
		ID:         "stale-node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	runtimeInfo, err := resolveBrowserRuntimeRouteForSession(
		callCtx,
		stateRegistry,
		sessionRegistry,
		map[string]any{},
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		backend,
	)
	if err != nil {
		t.Fatalf("resolveBrowserRuntimeRouteForSession(stale managed current fallback) error = %v", err)
	}
	if backend.resolveCalls != 1 {
		t.Fatalf("expected stale managed current route to probe managed lane once before falling back to visible base, got %d resolves", backend.resolveCalls)
	}
	if runtimeInfo != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected unresolvable managed current route to fall back to visible hidden base, got %#v", runtimeInfo)
	}
}

func TestBrowserResolveRuntimeActionDispatchSelectionUsesManagedCurrentRouteBeforeImplicitHostDiagnosticsDegrade(t *testing.T) {
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
	callCtx := WithToolSessionID(context.Background(), "runtime-action-dispatch-selection-managed-current")
	sessionRegistry.TrackCurrentTarget("runtime-action-dispatch-selection-managed-current", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	params := map[string]any{"action": "status"}
	preview := browserRuntimePreviewSessionSelections(
		callCtx,
		stateRegistry,
		sessionRegistry,
		params,
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
	)

	selection, err := browserResolveRuntimeActionDispatchSelection(
		callCtx,
		preview,
		params,
		"status",
		"",
		"",
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		router,
	)
	if err != nil {
		t.Fatalf("browserResolveRuntimeActionDispatchSelection(managed current status) error = %v", err)
	}
	if selection.DefaultCandidateRoute != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected managed current route selection to preserve default candidate route, got %#v", selection.DefaultCandidateRoute)
	}
	if selection.Requested != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected managed current route to drive requested runtime info, got %#v", selection.Requested)
	}
	if selection.HiddenRequestedRuntimeTarget {
		t.Fatalf("expected resolvable managed current route not to stay hidden in action dispatch selection")
	}
	if !selection.SelectedRouteReady || selection.SelectedRoute.RuntimeInfo.Target != "node" {
		t.Fatalf("expected managed current route to resolve a concrete node lane, got %#v", selection)
	}
	if !selection.CanUseManagedSessionRouteForImplicitFallback {
		t.Fatalf("expected managed current route to bypass implicit-host diagnostics degrade, got %#v", selection)
	}
	if selection.UseHiddenImplicitHostDiagnosticsDegrade {
		t.Fatalf("expected status to prefer managed current route over diagnostics degrade, got %#v", selection)
	}
}

func TestBrowserResolveRuntimeActionDispatchSelectionUsesDiagnosticsDegradeForStaleManagedCurrentRoute(t *testing.T) {
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
	callCtx := WithToolSessionID(context.Background(), "runtime-action-dispatch-selection-stale-managed-current")
	sessionRegistry.TrackCurrentTarget("runtime-action-dispatch-selection-stale-managed-current", BrowserSessionTarget{
		ID:         "stale-node-current",
		TabIndex:   2,
		URL:        "https://node.example/current",
		Title:      "Node Current",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	params := map[string]any{"action": "status"}
	preview := browserRuntimePreviewSessionSelections(
		callCtx,
		stateRegistry,
		sessionRegistry,
		params,
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
	)

	selection, err := browserResolveRuntimeActionDispatchSelection(
		callCtx,
		preview,
		params,
		"status",
		"",
		"",
		BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		true,
		router,
	)
	if err != nil {
		t.Fatalf("browserResolveRuntimeActionDispatchSelection(stale managed current status) error = %v", err)
	}
	if selection.DefaultCandidateRoute != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected stale managed current diagnostics degrade to keep default candidate route empty without managed default candidate, got %#v", selection.DefaultCandidateRoute)
	}
	if selection.Requested != (BrowserRuntimeInfo{}) {
		t.Fatalf("expected stale managed current route to keep targetless visible request, got %#v", selection.Requested)
	}
	if !selection.HiddenRequestedRuntimeTarget {
		t.Fatalf("expected stale managed current route to stay hidden behind the session-route preflight contract, got %#v", selection)
	}
	if selection.SelectedRouteReady {
		t.Fatalf("expected stale managed current route not to resolve a concrete lane, got %#v", selection)
	}
	if selection.RouteErr != nil {
		t.Fatalf("expected diagnostics degrade to clear stale managed current route error, got %v", selection.RouteErr)
	}
	if selection.CanUseManagedSessionRouteForImplicitFallback {
		t.Fatalf("expected stale managed current route not to advertise managed implicit fallback, got %#v", selection)
	}
	if !selection.UseHiddenImplicitHostDiagnosticsDegrade {
		t.Fatalf("expected status to stay on diagnostics degrade when managed current route is stale, got %#v", selection)
	}
}

func TestBrowserResolveRuntimeActionDispatchSelectionKeepsManagedDefaultRouteFailureForImplicitStatusDiagnostics(t *testing.T) {
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
	params := map[string]any{"action": "status"}
	preview := browserRuntimePreviewSessionSelections(
		context.Background(),
		nil,
		nil,
		params,
		host.runtimeInfo,
		true,
	)

	selection, err := browserResolveRuntimeActionDispatchSelection(
		context.Background(),
		preview,
		params,
		"status",
		"",
		"",
		host.runtimeInfo,
		true,
		router,
	)
	if err != nil {
		t.Fatalf("browserResolveRuntimeActionDispatchSelection(targetless managed-default status failure) error = %v", err)
	}
	if selection.DefaultCandidateRoute != (BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}) {
		t.Fatalf("expected managed-default status diagnostics to preserve hidden default candidate route, got %#v", selection.DefaultCandidateRoute)
	}
	if selection.HiddenRequestedRuntimeTarget {
		t.Fatalf("expected targetless managed-default failure not to masquerade as a hidden requested runtime_target, got %#v", selection)
	}
	if !selection.UseHiddenImplicitHostDiagnosticsDegrade {
		t.Fatalf("expected targetless status to stay on diagnostics degrade, got %#v", selection)
	}
	if selection.RouteErr == nil || !strings.Contains(selection.RouteErr.Error(), "managed browserd boot failed") {
		t.Fatalf("expected diagnostics selection to preserve managed-default route failure, got %#v", selection)
	}
}

func TestBrowserResolveRuntimeActionDispatchSelectionKeepsExplicitHostLaneAgainstPromotion(t *testing.T) {
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilities{RuntimeWorkbench: true, RuntimePrepare: true},
	}
	nodeBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilities{RuntimeWorkbench: true},
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
	params := map[string]any{
		"action":         "status",
		"runtime_target": "host",
		"profile":        "workbench",
	}
	preview := browserRuntimePreviewSessionSelections(
		context.Background(),
		nil,
		nil,
		params,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		true,
	)

	selection, err := browserResolveRuntimeActionDispatchSelection(
		context.Background(),
		preview,
		params,
		"status",
		"workbench",
		"host",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		true,
		router,
	)
	if err != nil {
		t.Fatalf("browserResolveRuntimeActionDispatchSelection(explicit host) error = %v", err)
	}
	if selection.Requested != (BrowserRuntimeInfo{Backend: "chrome", Profile: "workbench", Target: "host"}) {
		t.Fatalf("expected explicit host request to stay on host lane, got %#v", selection.Requested)
	}
	if !selection.ExplicitRuntimeTarget {
		t.Fatalf("expected explicit host runtime_target to be preserved, got %#v", selection)
	}
	if selection.HiddenRequestedRuntimeTarget {
		t.Fatalf("expected explicit host runtime_target not to be treated as hidden, got %#v", selection)
	}
	if !selection.SelectedRouteReady || selection.SelectedRoute.RuntimeInfo.Target != "host" || selection.SelectedRoute.Backend != hostBackend {
		t.Fatalf("expected explicit host selection to bypass promoted node route, got %#v", selection)
	}
	if selection.UseHiddenImplicitHostDiagnosticsDegrade {
		t.Fatalf("expected explicit host selection to stay off diagnostics degrade, got %#v", selection)
	}
}

func TestBrowserResolveRuntimeActionDispatchSelectionProbesStaleManagedCurrentRouteAtMostOnce(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &countingRuntimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       fullBrowserCapabilities(),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			if strings.EqualFold(strings.TrimSpace(requested.Target), "node") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return requested, nil
		},
	}
	backend := newBrowserBackend(BrowserToolOptions{
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: stateRegistry,
	}, outboundNetworkPolicy{}, 1_500)
	callCtx := WithToolSessionID(context.Background(), "runtime-action-dispatch-selection-stale-managed-current-single-probe")
	sessionRegistry.TrackCurrentTarget("runtime-action-dispatch-selection-stale-managed-current-single-probe", BrowserSessionTarget{
		ID:         "stale-node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	params := map[string]any{"action": "clear_session"}
	preview := browserRuntimePreviewSessionSelections(
		callCtx,
		stateRegistry,
		sessionRegistry,
		params,
		BrowserRuntimeInfo{},
		true,
	)
	resolveCallsBefore := nodeBackend.resolveCalls

	_, err := browserResolveRuntimeActionDispatchSelection(
		callCtx,
		preview,
		params,
		"clear_session",
		"",
		"",
		BrowserRuntimeInfo{},
		true,
		backend,
	)
	if err == nil || !strings.Contains(err.Error(), "requires explicit runtime_target") {
		t.Fatalf("expected stale managed current clear_session to stay behind explicit runtime_target gate, got %v", err)
	}
	if got := nodeBackend.resolveCalls - resolveCallsBefore; got > 1 {
		t.Fatalf("expected stale managed current route probe to resolve at most once in action dispatch selection, got %d", got)
	}
}

func TestBrowserRuntimeSelectedRouteInfoFromBindingPayloadPrefersTargetSelectionOverProfileSelection(t *testing.T) {
	info := browserRuntimeSelectedRouteInfoFromBindingPayload(browserRuntimePayload{
		SessionBinding: &browserRuntimeSessionBinding{
			SelectedBrowserBackend:       "proxy",
			SelectedBrowserProfile:       "workbench",
			SelectedBrowserTarget:        "node",
			SelectedBrowserProfileSource: "select_profile",
			SelectedBrowserTargetID:      "tab-2",
			SelectedBrowserTargetSource:  "tracked_active_tab",
		},
		SessionProfileSelection: &browserRuntimeSessionProfileSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "select_profile",
		},
		SessionTargetSelection: &browserRuntimeSessionTargetSelection{
			ID:            "tab-2",
			Backend:       "sandbox",
			Profile:       "isolated",
			RuntimeTarget: "sandbox",
			Source:        "tracked_active_tab",
		},
	})

	if info != (BrowserRuntimeInfo{Backend: "sandbox", Profile: "isolated", Target: "sandbox"}) {
		t.Fatalf("expected selected route info to prefer target selection identity, got %#v", info)
	}
}

func TestBrowserRuntimeConfiguredInfoForBindingPayloadFallsBackToProfileStatusWithoutBinding(t *testing.T) {
	info := browserRuntimeConfiguredInfoForBindingPayload(browserRuntimePayload{
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		},
	})

	if info != (BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}) {
		t.Fatalf("expected configured info to fall back to profile status snapshot, got %#v", info)
	}
}

func TestBrowserRuntimeProjectedSelectionPtrsFromBindingPayloadPrefersStoredSelectionIdentity(t *testing.T) {
	selection, _ := browserRuntimeProjectedSelectionPtrsFromBindingPayload(browserRuntimePayload{
		SessionBinding: &browserRuntimeSessionBinding{
			SelectedBrowserBackend:       "proxy",
			SelectedBrowserApp:           "Chromium",
			SelectedBrowserProfile:       "workbench",
			SelectedBrowserTarget:        "node",
			SelectedBrowserProfileSource: "select_profile",
			BrowserProfiles: []browserRuntimeProfileState{{
				Backend:       "sandbox",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Firefox",
				Status:        "running",
				Running:       true,
				Connected:     true,
			}},
		},
	})

	if selection == nil ||
		selection.Backend != "proxy" ||
		selection.Profile != "workbench" ||
		selection.RuntimeTarget != "node" ||
		selection.BrowserApp != "Chromium" ||
		selection.Source != "select_profile" {
		t.Fatalf("expected payload-aware binding selection projection to prefer stored selection identity over mismatched profile inventory, got %#v", selection)
	}
}

func TestBrowserRuntimeProjectedSelectionPtrsFromBindingPayloadHydratesTargetBrowserAppFromSnapshot(t *testing.T) {
	_, selection := browserRuntimeProjectedSelectionPtrsFromBindingPayload(browserRuntimePayload{
		SessionBinding: &browserRuntimeSessionBinding{
			SelectedBrowserBackend:      "proxy",
			SelectedBrowserProfile:      "workbench",
			SelectedBrowserTarget:       "node",
			SelectedBrowserTargetID:     "tab-2",
			SelectedBrowserTabIndex:     2,
			SelectedBrowserTargetSource: "tracked_active_tab",
			BrowserProfiles: []browserRuntimeProfileState{{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
			}},
		},
	})

	if selection == nil ||
		selection.ID != "tab-2" ||
		selection.TabIndex != 2 ||
		selection.Backend != "proxy" ||
		selection.Profile != "workbench" ||
		selection.RuntimeTarget != "node" ||
		selection.BrowserApp != "Chromium" ||
		selection.Source != "tracked_active_tab" {
		t.Fatalf("expected payload-aware binding selection projection to hydrate browser app from shared snapshot, got %#v", selection)
	}
}

func TestBrowserRuntimeTopLevelSessionProjectionFromBindingPayloadOverlaysPayloadSelections(t *testing.T) {
	projection := browserRuntimeTopLevelSessionProjectionFromBindingPayload(browserRuntimePayload{
		SessionBinding: &browserRuntimeSessionBinding{
			SelectedBrowserBackend:       "proxy",
			SelectedBrowserProfile:       "workbench",
			SelectedBrowserTarget:        "node",
			SelectedBrowserProfileSource: "select_profile",
			RouteTargetCount:             1,
		},
		SessionProfileSelection: &browserRuntimeSessionProfileSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "select_profile",
		},
		SessionTargetSelection: &browserRuntimeSessionTargetSelection{
			ID:            "tab-2",
			TabIndex:      2,
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "tracked_active_tab",
		},
	})

	if projection == nil ||
		len(projection.Routes) != 1 ||
		projection.Routes[0].Backend != "proxy" ||
		projection.Routes[0].Profile != "workbench" ||
		projection.Routes[0].RuntimeTarget != "node" ||
		projection.Routes[0].BrowserApp != "Chromium" ||
		projection.Routes[0].CurrentTargetID != "tab-2" ||
		projection.Routes[0].CurrentTargetSource != "tracked_active_tab" ||
		projection.TargetCount != 1 {
		t.Fatalf("expected binding payload session projection to inherit payload selections through shared overlay, got %#v", projection)
	}
	if len(projection.Profiles) != 1 ||
		projection.Profiles[0].Profile != "workbench" ||
		projection.Profiles[0].RuntimeTarget != "node" ||
		projection.Profiles[0].BrowserApp != "Chromium" ||
		projection.Profiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected binding payload session projection to synthesize route-scoped profile inventory from overlay, got %#v", projection.Profiles)
	}
}

func TestBrowserRuntimeSyncWorkbenchProjectionMirrorsResolverGuidanceSummary(t *testing.T) {
	payload := browserRuntimePayload{
		BrowserSurface: "explicit_managed_opt_in",
		BrowserOptInTargets: []string{
			"node",
		},
		SessionBinding: &browserRuntimeSessionBinding{
			SessionHealthResolverBlockedBy:         "multiple_candidates_filtered",
			SessionHealthResolverAmbiguityClass:    "filtered_residual",
			SessionHealthResolverCandidateKind:     "role_label",
			SessionHealthResolverStrength:          "strong",
			SessionHealthResolverRetryDisposition:  "manual_only",
			SessionHealthResolverManualRetryHint:   "add_ordinal",
			SessionHealthResolverNextStepAlias:     "snapshot",
			SessionHealthResolverSpecificityFields: []string{"href"},
		},
	}

	browserRuntimeSyncWorkbenchProjection(&payload, browserRuntimeWorkbenchProjectionSync{})

	if payload.ResolverBlockedBy != "multiple_candidates_filtered" ||
		payload.ResolverAmbiguityClass != "filtered_residual" ||
		payload.ResolverCandidateKind != "role_label" ||
		payload.ResolverCandidateStrength != "strong" ||
		payload.ResolverRetryDisposition != "manual_only" ||
		payload.ResolverManualRetryHint != "add_ordinal" ||
		payload.ResolverNextStepAlias != "snapshot" ||
		!reflect.DeepEqual(payload.ResolverSpecificityFields, []string{"href"}) {
		t.Fatalf("expected workbench sync to mirror resolver guidance summary, got %#v", payload)
	}
	if payload.ResolverExplanation == nil ||
		payload.ResolverExplanation.State != "manual_resolution_required" ||
		payload.ResolverExplanation.SummaryCode != "role_label_filtered_residual" ||
		payload.ResolverExplanation.NextStepAlias != "snapshot" ||
		payload.ResolverExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("expected workbench sync to build resolver explanation summary, got %#v", payload.ResolverExplanation)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "resolver" ||
		payload.DiagnosticsExplanation.State != "manual_resolution_required" ||
		payload.DiagnosticsExplanation.SummaryCode != "role_label_filtered_residual" ||
		payload.DiagnosticsExplanation.NextStepAlias != "snapshot" ||
		payload.DiagnosticsExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("expected workbench sync to build diagnostics explanation summary, got %#v", payload.DiagnosticsExplanation)
	}
	if payload.WorkbenchExplanation == nil ||
		payload.WorkbenchExplanation.Category != "resolver" ||
		payload.WorkbenchExplanation.State != "manual_resolution_required" ||
		payload.WorkbenchExplanation.SummaryCode != "role_label_filtered_residual" ||
		payload.WorkbenchExplanation.NextStepAlias != "snapshot" ||
		payload.WorkbenchExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("expected workbench sync to build workbench explanation summary, got %#v", payload.WorkbenchExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "resolver" ||
		payload.Explanation.State != "manual_resolution_required" ||
		payload.Explanation.SummaryCode != "role_label_filtered_residual" ||
		payload.Explanation.NextStepAlias != "snapshot" ||
		payload.Explanation.ManualRetryHint != "add_ordinal" ||
		payload.Explanation.ResolvedViaFallback {
		t.Fatalf("expected workbench sync to build top-level explanation alias, got %#v", payload.Explanation)
	}
	if payload.WorkbenchDiagnostics == nil ||
		payload.WorkbenchDiagnostics.Category != "resolver" ||
		payload.WorkbenchDiagnostics.State != "manual_resolution_required" ||
		payload.WorkbenchDiagnostics.SummaryCode != "role_label_filtered_residual" ||
		payload.WorkbenchDiagnostics.NextStepAlias != "snapshot" ||
		payload.WorkbenchDiagnostics.ManualRetryHint != "add_ordinal" {
		t.Fatalf("expected workbench sync to build workbench diagnostics summary, got %#v", payload.WorkbenchDiagnostics)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "resolver" ||
		payload.Diagnostics.State != "manual_resolution_required" ||
		payload.Diagnostics.SummaryCode != "role_label_filtered_residual" ||
		payload.Diagnostics.NextStepAlias != "snapshot" ||
		payload.Diagnostics.ManualRetryHint != "add_ordinal" ||
		payload.Diagnostics.ResolvedViaFallback {
		t.Fatalf("expected workbench sync to build top-level diagnostics alias, got %#v", payload.Diagnostics)
	}
	if payload.WorkbenchSummary == nil ||
		payload.WorkbenchSummary.Category != "resolver" ||
		payload.WorkbenchSummary.State != "manual_resolution_required" ||
		payload.WorkbenchSummary.SummaryCode != "role_label_filtered_residual" ||
		payload.WorkbenchSummary.NextStepAlias != "snapshot" ||
		payload.WorkbenchSummary.ManualRetryHint != "add_ordinal" ||
		payload.WorkbenchSummary.ResolvedViaFallback {
		t.Fatalf("expected workbench sync to build workbench summary alias, got %#v", payload.WorkbenchSummary)
	}
	if payload.Workbench == nil ||
		!payload.Workbench.Ready ||
		!browserStringSliceContains(payload.Workbench.Sections, "route") ||
		payload.Workbench.Explanation == nil ||
		payload.Workbench.Explanation.Category != "resolver" ||
		payload.Workbench.Explanation.State != "manual_resolution_required" ||
		payload.Workbench.Explanation.SummaryCode != "role_label_filtered_residual" ||
		payload.Workbench.Explanation.NextStepAlias != "snapshot" ||
		payload.Workbench.Explanation.ManualRetryHint != "add_ordinal" ||
		payload.Workbench.Explanation.ResolvedViaFallback ||
		payload.Workbench.Diagnostics == nil ||
		payload.Workbench.Diagnostics.Category != "resolver" ||
		payload.Workbench.Diagnostics.State != "manual_resolution_required" ||
		payload.Workbench.Diagnostics.SummaryCode != "role_label_filtered_residual" ||
		payload.Workbench.Diagnostics.NextStepAlias != "snapshot" ||
		payload.Workbench.Diagnostics.ManualRetryHint != "add_ordinal" ||
		payload.Workbench.Diagnostics.ResolvedViaFallback ||
		payload.Workbench.Summary == nil ||
		payload.Workbench.Summary.Category != "resolver" ||
		payload.Workbench.Summary.State != "manual_resolution_required" ||
		payload.Workbench.Summary.SummaryCode != "role_label_filtered_residual" ||
		payload.Workbench.Summary.NextStepAlias != "snapshot" ||
		payload.Workbench.Summary.ManualRetryHint != "add_ordinal" ||
		payload.Workbench.Summary.ResolvedViaFallback ||
		payload.Workbench.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.Workbench.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected workbench sync to build workbench surface summary, got %#v", payload.Workbench)
	}
	if payload.WorkbenchDisplay == nil ||
		!payload.WorkbenchDisplay.Ready ||
		!browserStringSliceContains(payload.WorkbenchDisplay.Sections, "route") ||
		payload.WorkbenchDisplay.Category != "resolver" ||
		payload.WorkbenchDisplay.State != "manual_resolution_required" ||
		payload.WorkbenchDisplay.SummaryCode != "role_label_filtered_residual" ||
		payload.WorkbenchDisplay.NextStepAlias != "snapshot" ||
		payload.WorkbenchDisplay.ManualRetryHint != "add_ordinal" ||
		payload.WorkbenchDisplay.ResolvedViaFallback {
		t.Fatalf("expected workbench sync to build workbench display summary, got %#v", payload.WorkbenchDisplay)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "resolver" ||
		payload.Summary.State != "manual_resolution_required" ||
		payload.Summary.SummaryCode != "role_label_filtered_residual" ||
		payload.Summary.NextStepAlias != "snapshot" ||
		payload.Summary.ManualRetryHint != "add_ordinal" ||
		payload.Summary.ResolvedViaFallback {
		t.Fatalf("expected workbench sync to build top-level summary alias, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		!payload.Display.Ready ||
		!browserStringSliceContains(payload.Display.Sections, "route") ||
		payload.Display.Category != "resolver" ||
		payload.Display.State != "manual_resolution_required" ||
		payload.Display.SummaryCode != "role_label_filtered_residual" ||
		payload.Display.NextStepAlias != "snapshot" ||
		payload.Display.ManualRetryHint != "add_ordinal" ||
		payload.Display.ResolvedViaFallback {
		t.Fatalf("expected workbench sync to build top-level display alias, got %#v", payload.Display)
	}
	if payload.Surface == nil ||
		!payload.Surface.Ready ||
		!browserStringSliceContains(payload.Surface.Sections, "route") ||
		payload.Surface.Category != "resolver" ||
		payload.Surface.State != "manual_resolution_required" ||
		payload.Surface.SummaryCode != "role_label_filtered_residual" ||
		payload.Surface.NextStepAlias != "snapshot" ||
		payload.Surface.ManualRetryHint != "add_ordinal" ||
		payload.Surface.ResolvedViaFallback ||
		payload.Surface.ReviewPolicyState != "" ||
		payload.Surface.ReviewDecision != "" ||
		payload.Surface.ReviewReady ||
		payload.Surface.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.Surface.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected workbench sync to build top-level surface alias, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "workbench" ||
		!payload.View.Ready ||
		!browserStringSliceContains(payload.View.Sections, "route") ||
		payload.View.Category != "resolver" ||
		payload.View.State != "manual_resolution_required" ||
		payload.View.SummaryCode != "role_label_filtered_residual" ||
		payload.View.NextStepAlias != "snapshot" ||
		payload.View.ManualRetryHint != "add_ordinal" ||
		payload.View.ResolvedViaFallback ||
		payload.View.Review != nil ||
		payload.View.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.View.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected workbench sync to build top-level view alias, got %#v", payload.View)
	}
}

func TestBrowserRuntimeApplyTopLevelProfileInventoryBuildsConfiguredProfiles(t *testing.T) {
	payload := browserRuntimePayload{
		RequestedProfile: "requested",
	}

	browserRuntimeApplyTopLevelProfileInventory(&payload, browserRuntimeTopLevelProfileInventoryProjection{
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		},
		ApplyProfileStatus: true,
		Profiles: []browserRuntimeProfileState{
			{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", Status: "running"},
			{Backend: "proxy", Profile: "relay", RuntimeTarget: "node", Status: "stopped"},
		},
		DiscoveredProfiles:    []string{"workbench", "relay"},
		DefaultProfile:        "workbench",
		ConfiguredInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		ApplyProfileInventory: true,
	})

	if payload.ProfileStatus == nil || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "running" {
		t.Fatalf("expected top-level profile inventory helper to populate profile status, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 2 || payload.Profiles[0].Profile != "workbench" || payload.Profiles[1].Profile != "relay" {
		t.Fatalf("expected top-level profile inventory helper to populate profiles, got %#v", payload.Profiles)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected top-level profile inventory helper to keep default profile, got %#v", payload)
	}
	for _, want := range []string{"requested", "workbench", "relay"} {
		if !browserStringSliceContains(payload.ConfiguredProfiles, want) {
			t.Fatalf("expected top-level profile inventory helper to include %q in configured profiles, got %#v", want, payload.ConfiguredProfiles)
		}
	}
}

func TestBrowserRuntimeApplyTopLevelSessionProjectionBuildsConfiguredProfilesAndMissingContextNote(t *testing.T) {
	payload := browserRuntimePayload{}

	browserRuntimeApplyTopLevelSessionProjection(&payload, browserRuntimeTopLevelSessionProjection{
		Routes: []browserRuntimeSessionRoute{
			{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", CurrentTargetID: "tab-1"},
		},
		TargetCount: 1,
		Runs: []browserRuntimeSessionRun{
			{RunID: "run-1", Status: "running"},
		},
		Profiles: []browserRuntimeProfileState{
			{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", Status: "running"},
		},
		Handoff: &agentxbrowserruntime.SharedSessionBrowserSessionHandoffSummary{
			State:         agentxbrowserruntime.SharedSessionBrowserSessionHandoffStateReady,
			NextStepAlias: "continue_current_target",
			CurrentTarget: &agentxbrowserruntime.SharedSessionBrowserSessionHandoffTargetSummary{
				ID: "tab-1",
			},
		},
		ConfiguredInfo:          BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
		ApplyConfiguredProfiles: true,
		MissingSessionIDNote:    "browser_runtime: no tool session context is available",
	})

	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].CurrentTargetID != "tab-1" || payload.SessionTargetCount != 1 {
		t.Fatalf("expected top-level session projection helper to populate session routes/count, got %#v", payload)
	}
	if len(payload.SessionRuns) != 1 || payload.SessionRuns[0].RunID != "run-1" {
		t.Fatalf("expected top-level session projection helper to populate runs, got %#v", payload.SessionRuns)
	}
	if len(payload.SessionProfiles) != 1 || payload.SessionProfiles[0].Profile != "workbench" {
		t.Fatalf("expected top-level session projection helper to populate profiles, got %#v", payload.SessionProfiles)
	}
	if payload.SessionHandoff == nil || payload.SessionHandoff.CurrentTarget == nil || payload.SessionHandoff.CurrentTarget.ID != "tab-1" {
		t.Fatalf("expected top-level session projection helper to populate handoff, got %#v", payload.SessionHandoff)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected top-level session projection helper to refresh configured profiles, got %#v", payload.ConfiguredProfiles)
	}
	if payload.Note != "browser_runtime: no tool session context is available" {
		t.Fatalf("expected top-level session projection helper to add missing-session note, got %#v", payload.Note)
	}
}

func TestBrowserRuntimeApplyTopLevelSessionProjectionPreservesExistingNote(t *testing.T) {
	payload := browserRuntimePayload{
		Note: "route preflight failed",
	}

	browserRuntimeApplyTopLevelSessionProjection(&payload, browserRuntimeTopLevelSessionProjection{
		MissingSessionIDNote: "browser_runtime: no tool session context is available",
	})

	if payload.Note != "route preflight failed" {
		t.Fatalf("expected top-level session projection helper to preserve existing note, got %#v", payload.Note)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceAppliesDirectBindingAndRefresh(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{
		ProfileInventory: &browserRuntimeTopLevelProfileInventoryProjection{
			ProfileStatus: &browserRuntimeProfileState{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				Status:        "running",
				Running:       true,
				Connected:     true,
			},
			ApplyProfileStatus: true,
			Profiles: []browserRuntimeProfileState{
				{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", Status: "running"},
			},
			DefaultProfile:        "workbench",
			ConfiguredInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			ApplyProfileInventory: true,
		},
		SessionProjection: &browserRuntimeTopLevelSessionProjection{
			Routes: []browserRuntimeSessionRoute{
				{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", CurrentTargetID: "tab-1"},
			},
			TargetCount: 1,
			Profiles: []browserRuntimeProfileState{
				{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", Status: "running"},
			},
		},
		ApplyDirectBinding: true,
		SessionProfileSelection: &browserRuntimeSessionProfileSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "select_profile",
		},
		SessionTargetSelection: &browserRuntimeSessionTargetSelection{
			ID:            "tab-1",
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "tracked_active_tab",
		},
		SessionBinding: &browserRuntimeSessionBinding{
			CurrentTargetID:         "tab-1",
			SelectedBrowserProfile:  "workbench",
			SelectedBrowserTargetID: "tab-1",
		},
		ExtraSections: []string{"coordination"},
	})

	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "workbench" {
		t.Fatalf("expected workbench session surface helper to apply direct profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != "tab-1" {
		t.Fatalf("expected workbench session surface helper to apply direct target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding == nil || payload.SessionBinding.CurrentTargetID != "tab-1" || payload.SessionBinding.SelectedBrowserTargetID != "tab-1" {
		t.Fatalf("expected workbench session surface helper to apply direct binding, got %#v", payload.SessionBinding)
	}
	if payload.ProfileStatus == nil || !payload.ProfileStatus.Selected {
		t.Fatalf("expected workbench session surface helper to mark the current profile selected for direct binding, got %#v", payload.ProfileStatus)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected workbench session surface helper to refresh configured profiles, got %#v", payload.ConfiguredProfiles)
	}
	for _, want := range []string{"route", "status", "profiles", "sessions", "coordination"} {
		if !browserStringSliceContains(payload.WorkbenchSections, want) {
			t.Fatalf("expected workbench session surface helper to include %q section, got %#v", want, payload.WorkbenchSections)
		}
	}
	if !payload.WorkbenchReady {
		t.Fatalf("expected workbench session surface helper to mark payload ready, got %#v", payload)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceKeepsExistingBindingWhenDirectBindingIsOmitted(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		SessionProfileSelection: &browserRuntimeSessionProfileSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "select_profile",
		},
		SessionTargetSelection: &browserRuntimeSessionTargetSelection{
			ID:            "tab-1",
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "tracked_active_tab",
		},
		SessionBinding: &browserRuntimeSessionBinding{
			CurrentTargetID:         "tab-1",
			SelectedBrowserProfile:  "workbench",
			SelectedBrowserTargetID: "tab-1",
		},
	}

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{
		SessionProjection: &browserRuntimeTopLevelSessionProjection{
			Routes: []browserRuntimeSessionRoute{
				{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node", CurrentTargetID: "tab-1"},
			},
			TargetCount: 1,
		},
	})

	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "workbench" {
		t.Fatalf("expected workbench session surface helper to keep existing profile selection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.ID != "tab-1" {
		t.Fatalf("expected workbench session surface helper to keep existing target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding == nil || payload.SessionBinding.CurrentTargetID != "tab-1" {
		t.Fatalf("expected workbench session surface helper to keep existing binding, got %#v", payload.SessionBinding)
	}
	if !browserStringSliceContains(payload.WorkbenchSections, "sessions") || !payload.WorkbenchReady {
		t.Fatalf("expected workbench session surface helper to refresh workbench sections without clearing existing binding, got %#v", payload)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceRefreshesConfiguredProfilesFromDirectBindingOnly(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
	}

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{
		ApplyDirectBinding: true,
		SessionProfileSelection: &browserRuntimeSessionProfileSelection{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "select_profile",
		},
		SessionTargetSelection: &browserRuntimeSessionTargetSelection{
			ID:            "tab-1",
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Source:        "tracked_active_tab",
		},
		SessionBinding: &browserRuntimeSessionBinding{
			CurrentTargetID:         "tab-1",
			SelectedBrowserProfile:  "workbench",
			SelectedBrowserTargetID: "tab-1",
		},
	})

	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected workbench direct binding helper to refresh configured profiles, got %#v", payload.ConfiguredProfiles)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected workbench direct binding helper to backfill default profile from direct binding, got %#v", payload.DefaultProfile)
	}
	if !browserStringSliceContains(payload.WorkbenchSections, "sessions") || payload.Workbench == nil || !browserStringSliceContains(payload.Workbench.Sections, "sessions") {
		t.Fatalf("expected workbench direct binding helper to surface sessions section from binding-only state, got payload=%#v workbench=%#v", payload.WorkbenchSections, payload.Workbench)
	}
	if payload.Display == nil || !browserStringSliceContains(payload.Display.Sections, "sessions") {
		t.Fatalf("expected workbench direct binding helper to carry sessions section into top-level display, got %#v", payload.Display)
	}
	if len(payload.SessionRoutes) != 1 ||
		payload.SessionRoutes[0].Backend != "proxy" ||
		payload.SessionRoutes[0].Profile != "workbench" ||
		payload.SessionRoutes[0].RuntimeTarget != "node" ||
		payload.SessionRoutes[0].CurrentTargetID != "tab-1" ||
		payload.SessionTargetCount != 1 {
		t.Fatalf("expected workbench direct binding helper to backfill minimal session projection, got routes=%#v target_count=%d", payload.SessionRoutes, payload.SessionTargetCount)
	}
	if payload.SelectedRoute == nil || payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected workbench direct binding helper to backfill managed selected route, got %#v", payload.SelectedRoute)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceRefreshesConfiguredProfilesFromBindingShellOnly(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
		SessionBinding: &browserRuntimeSessionBinding{
			SelectedBrowserProfile:       "workbench",
			SelectedBrowserTarget:        "node",
			SelectedBrowserProfileSource: "select_profile",
		},
	}

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{})

	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected binding-only workbench surface to refresh configured profiles from binding shell, got %#v", payload.ConfiguredProfiles)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected binding-only workbench surface to backfill default profile from binding shell, got %#v", payload.DefaultProfile)
	}
	if !browserStringSliceContains(payload.WorkbenchSections, "sessions") || payload.Workbench == nil || !browserStringSliceContains(payload.Workbench.Sections, "sessions") {
		t.Fatalf("expected binding-only workbench surface to expose sessions section from binding shell, got payload=%#v workbench=%#v", payload.WorkbenchSections, payload.Workbench)
	}
	if payload.Display == nil || !browserStringSliceContains(payload.Display.Sections, "sessions") {
		t.Fatalf("expected binding-only workbench surface to carry sessions section into top-level display, got %#v", payload.Display)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceBackfillsProfileInventoryFromBindingShell(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		SessionBinding: &browserRuntimeSessionBinding{
			SelectedBrowserProfile:       "workbench",
			SelectedBrowserTarget:        "node",
			SelectedBrowserProfileSource: "select_profile",
			BrowserProfiles: []browserRuntimeProfileState{{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
			}},
		},
	}

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{})

	if payload.ProfileStatus == nil ||
		payload.ProfileStatus.Profile != "workbench" ||
		payload.ProfileStatus.RuntimeTarget != "node" ||
		payload.ProfileStatus.BrowserApp != "Chromium" ||
		payload.ProfileStatus.Status != "running" {
		t.Fatalf("expected binding-only workbench surface to backfill profile status from binding shell, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 1 ||
		payload.Profiles[0].Profile != "workbench" ||
		payload.Profiles[0].RuntimeTarget != "node" ||
		payload.Profiles[0].BrowserApp != "Chromium" ||
		payload.Profiles[0].Status != "running" {
		t.Fatalf("expected binding-only workbench surface to backfill profiles from binding shell, got %#v", payload.Profiles)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceBackfillsSessionProjectionFromBindingShell(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
		SessionBinding: &browserRuntimeSessionBinding{
			CurrentTargetID:              "tab-1",
			SelectedBrowserBackend:       "proxy",
			SelectedBrowserApp:           "Chromium",
			SelectedBrowserProfile:       "workbench",
			SelectedBrowserTarget:        "node",
			SelectedBrowserProfileSource: "select_profile",
			SelectedBrowserTargetID:      "tab-1",
			SelectedBrowserTargetSource:  "tracked_active_tab",
			NodeRuns: []browserRuntimeSessionRun{{
				RunID:    "run-1",
				NodeID:   "node-1",
				Status:   "running",
				Provider: "browser",
				Action:   "nodes action=run",
			}},
			BrowserProfiles: []browserRuntimeProfileState{{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
				Selected:      true,
			}},
			RouteTargetCount: 1,
		},
	}

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{})

	if len(payload.SessionRoutes) != 1 ||
		payload.SessionRoutes[0].Backend != "proxy" ||
		payload.SessionRoutes[0].Profile != "workbench" ||
		payload.SessionRoutes[0].RuntimeTarget != "node" ||
		payload.SessionRoutes[0].BrowserApp != "Chromium" ||
		payload.SessionRoutes[0].CurrentTargetID != "tab-1" ||
		payload.SessionRoutes[0].CurrentTargetSource != "tracked_active_tab" ||
		payload.SessionTargetCount != 1 {
		t.Fatalf("expected binding-only workbench surface to backfill session route projection, got routes=%#v target_count=%d", payload.SessionRoutes, payload.SessionTargetCount)
	}
	if len(payload.SessionRuns) != 1 || payload.SessionRuns[0].RunID != "run-1" {
		t.Fatalf("expected binding-only workbench surface to backfill session runs from binding shell, got %#v", payload.SessionRuns)
	}
	if len(payload.SessionProfiles) != 1 ||
		payload.SessionProfiles[0].Profile != "workbench" ||
		payload.SessionProfiles[0].RuntimeTarget != "node" ||
		payload.SessionProfiles[0].BrowserApp != "Chromium" ||
		payload.SessionProfiles[0].Status != "running" ||
		!payload.SessionProfiles[0].Selected {
		t.Fatalf("expected binding-only workbench surface to backfill session profiles from binding shell, got %#v", payload.SessionProfiles)
	}
	if payload.SelectedRoute == nil ||
		payload.SelectedRoute.Backend != "proxy" ||
		payload.SelectedRoute.Profile != "workbench" ||
		payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected binding-only workbench surface to backfill managed selected route, got %#v", payload.SelectedRoute)
	}
	if payload.SessionProfileSelection == nil ||
		payload.SessionProfileSelection.Backend != "proxy" ||
		payload.SessionProfileSelection.Profile != "workbench" ||
		payload.SessionProfileSelection.RuntimeTarget != "node" ||
		payload.SessionProfileSelection.BrowserApp != "Chromium" ||
		payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected binding-only workbench surface to backfill profile selection from binding shell, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil ||
		payload.SessionTargetSelection.ID != "tab-1" ||
		payload.SessionTargetSelection.Backend != "proxy" ||
		payload.SessionTargetSelection.Profile != "workbench" ||
		payload.SessionTargetSelection.RuntimeTarget != "node" ||
		payload.SessionTargetSelection.BrowserApp != "Chromium" ||
		payload.SessionTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected binding-only workbench surface to backfill target selection from binding shell, got %#v", payload.SessionTargetSelection)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceBackfillsSessionProfilesFromRouteScopedBindingShell(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
		SessionBinding: &browserRuntimeSessionBinding{
			CurrentTargetID:              "tab-2",
			SelectedBrowserBackend:       "proxy",
			SelectedBrowserApp:           "Chromium",
			SelectedBrowserProfile:       "workbench",
			SelectedBrowserTarget:        "node",
			SelectedBrowserProfileSource: "select_profile",
			SelectedBrowserTargetID:      "tab-2",
			SelectedBrowserTargetSource:  "tracked_active_tab",
			RouteTargetCount:             1,
		},
	}

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{})

	if len(payload.SessionProfiles) != 1 ||
		payload.SessionProfiles[0].Backend != "proxy" ||
		payload.SessionProfiles[0].Profile != "workbench" ||
		payload.SessionProfiles[0].RuntimeTarget != "node" ||
		payload.SessionProfiles[0].BrowserApp != "Chromium" ||
		payload.SessionProfiles[0].Note != "cached route-scoped session snapshot" ||
		!payload.SessionProfiles[0].Selected {
		t.Fatalf("expected binding-shell workbench surface to synthesize route-scoped session profiles, got %#v", payload.SessionProfiles)
	}
	if len(payload.SessionRoutes) != 1 ||
		payload.SessionRoutes[0].Backend != "proxy" ||
		payload.SessionRoutes[0].Profile != "workbench" ||
		payload.SessionRoutes[0].RuntimeTarget != "node" ||
		payload.SessionRoutes[0].CurrentTargetID != "tab-2" {
		t.Fatalf("expected binding-shell workbench surface to keep synthesized session route, got %#v", payload.SessionRoutes)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceBackfillsProfileInventoryFromSessionProjection(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
	}

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{
		SessionProjection: &browserRuntimeTopLevelSessionProjection{
			Profiles: []browserRuntimeProfileState{{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
				Selected:      true,
				Note:          "cached route-scoped session snapshot",
			}},
			ConfiguredInfo:          BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			ApplyConfiguredProfiles: true,
		},
	})

	if payload.ProfileStatus == nil ||
		payload.ProfileStatus.Profile != "workbench" ||
		payload.ProfileStatus.RuntimeTarget != "node" ||
		payload.ProfileStatus.BrowserApp != "Chromium" ||
		payload.ProfileStatus.Status != "running" ||
		payload.ProfileStatus.Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected workbench surface to backfill profile status from session projection, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 1 ||
		payload.Profiles[0].Profile != "workbench" ||
		payload.Profiles[0].RuntimeTarget != "node" ||
		payload.Profiles[0].BrowserApp != "Chromium" ||
		payload.Profiles[0].Status != "running" ||
		payload.Profiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected workbench surface to backfill profiles from session projection, got %#v", payload.Profiles)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected workbench surface to backfill default profile from session projection, got %#v", payload.DefaultProfile)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected workbench surface to refresh configured profiles from session projection inventory, got %#v", payload.ConfiguredProfiles)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceBackfillsProfilesWithoutOverridingExistingStatus(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
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

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{
		SessionProjection: &browserRuntimeTopLevelSessionProjection{
			Profiles: []browserRuntimeProfileState{{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
				Selected:      true,
				Note:          "cached route-scoped session snapshot",
			}},
			ConfiguredInfo:          BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			ApplyConfiguredProfiles: true,
		},
	})

	if payload.ProfileStatus == nil || payload.ProfileStatus.Note != "fresh current status" {
		t.Fatalf("expected workbench surface to preserve existing profile status, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected workbench surface to still backfill profiles from session projection, got %#v", payload.Profiles)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected workbench surface to backfill default profile when only status already exists, got %#v", payload.DefaultProfile)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected workbench surface to refresh configured profiles when only status already exists, got %#v", payload.ConfiguredProfiles)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceBindingShellKeepsImplicitHostSelectedRouteHidden(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
		SessionBinding: &browserRuntimeSessionBinding{
			CurrentTargetID:              "tab-host",
			SelectedBrowserBackend:       "system",
			SelectedBrowserProfile:       "default",
			SelectedBrowserTarget:        "host",
			SelectedBrowserProfileSource: "remember_profile",
			SelectedBrowserTargetID:      "tab-host",
			SelectedBrowserTargetSource:  "tracked_active_tab",
			RouteTargetCount:             1,
		},
	}

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{})

	if payload.SelectedRoute != nil {
		t.Fatalf("expected binding-only workbench surface to keep implicit host selected route hidden, got %#v", payload.SelectedRoute)
	}
	if payload.SessionProfileSelection != nil || payload.SessionTargetSelection != nil {
		t.Fatalf("expected binding-only workbench surface not to backfill hidden host selections, got profile=%#v target=%#v", payload.SessionProfileSelection, payload.SessionTargetSelection)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceDerivesSelectionsFromBindingShellOnly(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
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
		},
	}

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{
		ApplyDirectBinding: true,
		SessionBinding: &browserRuntimeSessionBinding{
			SelectedBrowserProfile:       "workbench",
			SelectedBrowserTarget:        "node",
			SelectedBrowserProfileSource: "select_profile",
			SelectedBrowserTargetID:      "tab-1",
			SelectedBrowserTabIndex:      1,
			SelectedBrowserTargetSource:  "tracked_active_tab",
		},
	})

	if payload.SessionProfileSelection == nil ||
		payload.SessionProfileSelection.Backend != "proxy" ||
		payload.SessionProfileSelection.Profile != "workbench" ||
		payload.SessionProfileSelection.RuntimeTarget != "node" ||
		payload.SessionProfileSelection.BrowserApp != "Chromium" ||
		payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected binding-only direct binding helper to derive profile selection from binding shell, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil ||
		payload.SessionTargetSelection.ID != "tab-1" ||
		payload.SessionTargetSelection.TabIndex != 1 ||
		payload.SessionTargetSelection.Backend != "proxy" ||
		payload.SessionTargetSelection.Profile != "workbench" ||
		payload.SessionTargetSelection.RuntimeTarget != "node" ||
		payload.SessionTargetSelection.BrowserApp != "Chromium" ||
		payload.SessionTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected binding-only direct binding helper to derive target selection from binding shell, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding == nil ||
		payload.SessionBinding.SelectedBrowserBackend != "proxy" ||
		payload.SessionBinding.SelectedBrowserProfile != "workbench" ||
		payload.SessionBinding.SelectedBrowserTarget != "node" ||
		payload.SessionBinding.SelectedBrowserApp != "Chromium" ||
		payload.SessionBinding.SelectedBrowserTargetID != "tab-1" ||
		payload.SessionBinding.SelectedBrowserTabIndex != 1 ||
		payload.SessionBinding.SelectedBrowserProfileSource != "select_profile" ||
		payload.SessionBinding.SelectedBrowserTargetSource != "tracked_active_tab" {
		t.Fatalf("expected binding-only direct binding helper to normalize binding identity from selected route and selections, got %#v", payload.SessionBinding)
	}
	if payload.ProfileStatus == nil || !payload.ProfileStatus.Selected {
		t.Fatalf("expected binding-only direct binding helper to mark current profile selected, got %#v", payload.ProfileStatus)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceDerivesSelectionsFromBindingShellWithoutSelectedRoute(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
		Profiles: []browserRuntimeProfileState{
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
	}

	browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{
		ApplyDirectBinding: true,
		SessionBinding: &browserRuntimeSessionBinding{
			SelectedBrowserBackend:       "proxy",
			SelectedBrowserProfile:       "isolated",
			SelectedBrowserTarget:        "node",
			SelectedBrowserProfileSource: "select_profile",
			SelectedBrowserTargetID:      "tab-2",
			SelectedBrowserTabIndex:      2,
			SelectedBrowserTargetSource:  "tracked_active_tab",
		},
	})

	if payload.SessionProfileSelection == nil ||
		payload.SessionProfileSelection.Backend != "proxy" ||
		payload.SessionProfileSelection.Profile != "isolated" ||
		payload.SessionProfileSelection.RuntimeTarget != "node" ||
		payload.SessionProfileSelection.BrowserApp != "Chromium" ||
		payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected route-less direct binding helper to derive profile selection from shared projection, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection == nil ||
		payload.SessionTargetSelection.ID != "tab-2" ||
		payload.SessionTargetSelection.TabIndex != 2 ||
		payload.SessionTargetSelection.Backend != "proxy" ||
		payload.SessionTargetSelection.Profile != "isolated" ||
		payload.SessionTargetSelection.RuntimeTarget != "node" ||
		payload.SessionTargetSelection.BrowserApp != "Chromium" ||
		payload.SessionTargetSelection.Source != "tracked_active_tab" {
		t.Fatalf("expected route-less direct binding helper to derive target selection from shared projection, got %#v", payload.SessionTargetSelection)
	}
}

func TestBrowserRuntimeApplyWorkbenchSessionSurfaceSyncsCoordinationSurface(t *testing.T) {
	t.Run("adds coordination section and plan from binding", func(t *testing.T) {
		payload := browserRuntimePayload{
			Action:               "workbench",
			Status:               "ok",
			CoordinationState:    "stale_state",
			CoordinationDecision: "stale_decision",
			CoordinationReady:    true,
			SessionBinding: &browserRuntimeSessionBinding{
				Coordination: &browserRuntimeCoordination{
					State:                     "browser_ready",
					PrimaryBrowserAction:      "browser action=refresh",
					PrimaryNodeAction:         "nodes action=run",
					NextStep:                  "browser action=refresh",
					RecommendedBrowserActions: []string{"browser action=refresh"},
					RecommendedNodeActions:    []string{"nodes action=run"},
				},
			},
		}

		browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{
			SyncCoordinationSurface: true,
		})

		if !browserStringSliceContains(payload.WorkbenchSections, "coordination") {
			t.Fatalf("expected workbench session surface helper to add coordination section, got %#v", payload.WorkbenchSections)
		}
		if payload.WorkbenchPrimaryBrowserAction != "browser action=refresh" || payload.WorkbenchPrimaryNodeAction != "nodes action=run" || payload.WorkbenchNextStep != "browser action=refresh" {
			t.Fatalf("expected workbench session surface helper to refresh action plan from binding coordination, got %#v", payload)
		}
		if payload.CoordinationState != "browser_ready" || payload.CoordinationDecision != "" || payload.CoordinationReady {
			t.Fatalf("expected workbench session surface helper to sync coordination summary from binding without preserving stale lifecycle decision, got %#v", payload)
		}
		if payload.WorkbenchDiagnostics == nil ||
			payload.WorkbenchDiagnostics.Category != "coordination" ||
			payload.WorkbenchDiagnostics.State != "action_plan_available" ||
			payload.WorkbenchDiagnostics.SummaryCode != "workbench_action_plan" ||
			payload.WorkbenchDiagnostics.PrimaryBrowserAction != "browser action=refresh" ||
			payload.WorkbenchDiagnostics.PrimaryNodeAction != "nodes action=run" ||
			payload.WorkbenchDiagnostics.NextStep != "browser action=refresh" {
			t.Fatalf("expected workbench session surface helper to build coordination diagnostics summary, got %#v", payload.WorkbenchDiagnostics)
		}
		if payload.Explanation != nil {
			t.Fatalf("expected workbench coordination surface not to invent explanation alias, got %#v", payload.Explanation)
		}
		if payload.Diagnostics == nil ||
			payload.Diagnostics.Category != "coordination" ||
			payload.Diagnostics.State != "action_plan_available" ||
			payload.Diagnostics.SummaryCode != "workbench_action_plan" ||
			payload.Diagnostics.PrimaryBrowserAction != "browser action=refresh" ||
			payload.Diagnostics.PrimaryNodeAction != "nodes action=run" ||
			payload.Diagnostics.NextStep != "browser action=refresh" ||
			payload.Diagnostics.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to build top-level coordination diagnostics, got %#v", payload.Diagnostics)
		}
		if payload.WorkbenchSummary == nil ||
			payload.WorkbenchSummary.Category != "coordination" ||
			payload.WorkbenchSummary.State != "action_plan_available" ||
			payload.WorkbenchSummary.SummaryCode != "workbench_action_plan" ||
			payload.WorkbenchSummary.PrimaryBrowserAction != "browser action=refresh" ||
			payload.WorkbenchSummary.PrimaryNodeAction != "nodes action=run" ||
			payload.WorkbenchSummary.NextStep != "browser action=refresh" ||
			payload.WorkbenchSummary.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to build workbench coordination summary, got %#v", payload.WorkbenchSummary)
		}
		if payload.Workbench == nil ||
			!payload.Workbench.Ready ||
			!browserStringSliceContains(payload.Workbench.Sections, "coordination") ||
			payload.Workbench.Explanation != nil ||
			payload.Workbench.Diagnostics == nil ||
			payload.Workbench.Diagnostics.Category != "coordination" ||
			payload.Workbench.Diagnostics.State != "action_plan_available" ||
			payload.Workbench.Diagnostics.SummaryCode != "workbench_action_plan" ||
			payload.Workbench.Diagnostics.PrimaryBrowserAction != "browser action=refresh" ||
			payload.Workbench.Diagnostics.PrimaryNodeAction != "nodes action=run" ||
			payload.Workbench.Diagnostics.NextStep != "browser action=refresh" ||
			payload.Workbench.Summary == nil ||
			payload.Workbench.Summary.Category != "coordination" ||
			payload.Workbench.Summary.State != "action_plan_available" ||
			payload.Workbench.Summary.SummaryCode != "workbench_action_plan" ||
			payload.Workbench.PrimaryBrowserAction != "browser action=refresh" ||
			payload.Workbench.PrimaryNodeAction != "nodes action=run" ||
			payload.Workbench.NextStep != "browser action=refresh" ||
			!browserStringSliceContains(payload.Workbench.RecommendedBrowserActions, "browser action=refresh") ||
			!browserStringSliceContains(payload.Workbench.RecommendedNodeActions, "nodes action=run") {
			t.Fatalf("expected workbench session surface helper to build coordination workbench surface, got %#v", payload.Workbench)
		}
		if payload.WorkbenchDisplay == nil ||
			!payload.WorkbenchDisplay.Ready ||
			!browserStringSliceContains(payload.WorkbenchDisplay.Sections, "coordination") ||
			payload.WorkbenchDisplay.Category != "coordination" ||
			payload.WorkbenchDisplay.State != "action_plan_available" ||
			payload.WorkbenchDisplay.SummaryCode != "workbench_action_plan" ||
			payload.WorkbenchDisplay.PrimaryBrowserAction != "browser action=refresh" ||
			payload.WorkbenchDisplay.PrimaryNodeAction != "nodes action=run" ||
			payload.WorkbenchDisplay.NextStep != "browser action=refresh" ||
			payload.WorkbenchDisplay.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to build coordination workbench display, got %#v", payload.WorkbenchDisplay)
		}
		if payload.Summary == nil ||
			payload.Summary.Category != "coordination" ||
			payload.Summary.State != "action_plan_available" ||
			payload.Summary.SummaryCode != "workbench_action_plan" ||
			payload.Summary.PrimaryBrowserAction != "browser action=refresh" ||
			payload.Summary.PrimaryNodeAction != "nodes action=run" ||
			payload.Summary.NextStep != "browser action=refresh" ||
			payload.Summary.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to build top-level coordination summary, got %#v", payload.Summary)
		}
		if payload.Display == nil ||
			!payload.Display.Ready ||
			!browserStringSliceContains(payload.Display.Sections, "coordination") ||
			payload.Display.Category != "coordination" ||
			payload.Display.State != "action_plan_available" ||
			payload.Display.SummaryCode != "workbench_action_plan" ||
			payload.Display.PrimaryBrowserAction != "browser action=refresh" ||
			payload.Display.PrimaryNodeAction != "nodes action=run" ||
			payload.Display.NextStep != "browser action=refresh" ||
			payload.Display.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to build top-level coordination display, got %#v", payload.Display)
		}
	})

	t.Run("refreshes configured profiles from coordination-only binding", func(t *testing.T) {
		payload := browserRuntimePayload{
			Action: "workbench",
			Status: "ok",
			SelectedRoute: &browserRuntimeRouteDescriptor{
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
			},
			SessionBinding: &browserRuntimeSessionBinding{
				Coordination: &browserRuntimeCoordination{
					State:                "browser_ready",
					PrimaryBrowserAction: "browser action=refresh",
				},
			},
		}

		browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{
			SyncCoordinationSurface: true,
		})

		if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
			t.Fatalf("expected coordination-only workbench surface to refresh configured profiles from selected route, got %#v", payload.ConfiguredProfiles)
		}
		if payload.DefaultProfile != "workbench" {
			t.Fatalf("expected coordination-only workbench surface to backfill default profile from selected route, got %#v", payload.DefaultProfile)
		}
		if payload.ProfileStatus != nil || len(payload.Profiles) != 0 {
			t.Fatalf("expected coordination-only workbench surface not to invent profile inventory from binding, got status=%#v profiles=%#v", payload.ProfileStatus, payload.Profiles)
		}
		if browserStringSliceContains(payload.WorkbenchSections, "sessions") {
			t.Fatalf("expected coordination-only workbench surface not to invent sessions section, got %#v", payload.WorkbenchSections)
		}
	})

	t.Run("mirrors resolver guidance into workbench explanation", func(t *testing.T) {
		payload := browserRuntimePayload{
			Action: "workbench",
			Status: "ok",
			SessionBinding: &browserRuntimeSessionBinding{
				SessionHealthResolverBlockedBy:         "multiple_candidates_filtered",
				SessionHealthResolverAmbiguityClass:    "filtered_residual",
				SessionHealthResolverCandidateKind:     "label",
				SessionHealthResolverStrength:          "medium",
				SessionHealthResolverRetryDisposition:  "manual_only",
				SessionHealthResolverManualRetryHint:   "add_ordinal",
				SessionHealthResolverNextStepAlias:     "snapshot",
				SessionHealthResolverSpecificityFields: []string{"tag", "type"},
			},
		}

		browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{})

		if payload.WorkbenchExplanation == nil ||
			payload.WorkbenchExplanation.Category != "resolver" ||
			payload.WorkbenchExplanation.State != "manual_resolution_required" ||
			payload.WorkbenchExplanation.SummaryCode != "label_filtered_residual" ||
			payload.WorkbenchExplanation.NextStepAlias != "snapshot" ||
			payload.WorkbenchExplanation.ManualRetryHint != "add_ordinal" {
			t.Fatalf("expected workbench session surface helper to mirror resolver guidance into workbench explanation, got %#v", payload.WorkbenchExplanation)
		}
		if payload.Explanation == nil ||
			payload.Explanation.Category != "resolver" ||
			payload.Explanation.State != "manual_resolution_required" ||
			payload.Explanation.SummaryCode != "label_filtered_residual" ||
			payload.Explanation.NextStepAlias != "snapshot" ||
			payload.Explanation.ManualRetryHint != "add_ordinal" ||
			payload.Explanation.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to mirror resolver guidance into top-level explanation, got %#v", payload.Explanation)
		}
		if payload.WorkbenchDiagnostics == nil ||
			payload.WorkbenchDiagnostics.Category != "resolver" ||
			payload.WorkbenchDiagnostics.State != "manual_resolution_required" ||
			payload.WorkbenchDiagnostics.SummaryCode != "label_filtered_residual" ||
			payload.WorkbenchDiagnostics.NextStepAlias != "snapshot" ||
			payload.WorkbenchDiagnostics.ManualRetryHint != "add_ordinal" {
			t.Fatalf("expected workbench session surface helper to mirror resolver guidance into workbench diagnostics, got %#v", payload.WorkbenchDiagnostics)
		}
		if payload.Diagnostics == nil ||
			payload.Diagnostics.Category != "resolver" ||
			payload.Diagnostics.State != "manual_resolution_required" ||
			payload.Diagnostics.SummaryCode != "label_filtered_residual" ||
			payload.Diagnostics.NextStepAlias != "snapshot" ||
			payload.Diagnostics.ManualRetryHint != "add_ordinal" ||
			payload.Diagnostics.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to mirror resolver guidance into top-level diagnostics, got %#v", payload.Diagnostics)
		}
		if payload.WorkbenchSummary == nil ||
			payload.WorkbenchSummary.Category != "resolver" ||
			payload.WorkbenchSummary.State != "manual_resolution_required" ||
			payload.WorkbenchSummary.SummaryCode != "label_filtered_residual" ||
			payload.WorkbenchSummary.NextStepAlias != "snapshot" ||
			payload.WorkbenchSummary.ManualRetryHint != "add_ordinal" ||
			payload.WorkbenchSummary.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to mirror resolver guidance into workbench summary, got %#v", payload.WorkbenchSummary)
		}
		if payload.Workbench == nil ||
			!payload.Workbench.Ready ||
			payload.Workbench.Explanation == nil ||
			payload.Workbench.Explanation.Category != "resolver" ||
			payload.Workbench.Explanation.State != "manual_resolution_required" ||
			payload.Workbench.Explanation.SummaryCode != "label_filtered_residual" ||
			payload.Workbench.Explanation.NextStepAlias != "snapshot" ||
			payload.Workbench.Explanation.ManualRetryHint != "add_ordinal" ||
			payload.Workbench.Explanation.ResolvedViaFallback ||
			payload.Workbench.Diagnostics == nil ||
			payload.Workbench.Diagnostics.Category != "resolver" ||
			payload.Workbench.Diagnostics.State != "manual_resolution_required" ||
			payload.Workbench.Diagnostics.SummaryCode != "label_filtered_residual" ||
			payload.Workbench.Diagnostics.NextStepAlias != "snapshot" ||
			payload.Workbench.Diagnostics.ManualRetryHint != "add_ordinal" ||
			payload.Workbench.Diagnostics.ResolvedViaFallback ||
			payload.Workbench.Summary == nil ||
			payload.Workbench.Summary.Category != "resolver" ||
			payload.Workbench.Summary.State != "manual_resolution_required" ||
			payload.Workbench.Summary.SummaryCode != "label_filtered_residual" ||
			payload.Workbench.Summary.NextStepAlias != "snapshot" ||
			payload.Workbench.Summary.ManualRetryHint != "add_ordinal" ||
			payload.Workbench.Summary.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to mirror resolver guidance into workbench surface, got %#v", payload.Workbench)
		}
		if payload.WorkbenchDisplay == nil ||
			!payload.WorkbenchDisplay.Ready ||
			payload.WorkbenchDisplay.Category != "resolver" ||
			payload.WorkbenchDisplay.State != "manual_resolution_required" ||
			payload.WorkbenchDisplay.SummaryCode != "label_filtered_residual" ||
			payload.WorkbenchDisplay.NextStepAlias != "snapshot" ||
			payload.WorkbenchDisplay.ManualRetryHint != "add_ordinal" ||
			payload.WorkbenchDisplay.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to mirror resolver guidance into workbench display, got %#v", payload.WorkbenchDisplay)
		}
		if payload.Summary == nil ||
			payload.Summary.Category != "resolver" ||
			payload.Summary.State != "manual_resolution_required" ||
			payload.Summary.SummaryCode != "label_filtered_residual" ||
			payload.Summary.NextStepAlias != "snapshot" ||
			payload.Summary.ManualRetryHint != "add_ordinal" ||
			payload.Summary.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to mirror resolver guidance into top-level summary, got %#v", payload.Summary)
		}
		if payload.Display == nil ||
			!payload.Display.Ready ||
			payload.Display.Category != "resolver" ||
			payload.Display.State != "manual_resolution_required" ||
			payload.Display.SummaryCode != "label_filtered_residual" ||
			payload.Display.NextStepAlias != "snapshot" ||
			payload.Display.ManualRetryHint != "add_ordinal" ||
			payload.Display.ResolvedViaFallback {
			t.Fatalf("expected workbench session surface helper to mirror resolver guidance into top-level display, got %#v", payload.Display)
		}
	})

	t.Run("clears stale action plan when coordination is absent", func(t *testing.T) {
		payload := browserRuntimePayload{
			Action:                             "workbench",
			Status:                             "ok",
			CoordinationState:                  "stale_state",
			CoordinationDecision:               "stale_decision",
			CoordinationReady:                  true,
			WorkbenchSections:                  []string{"route", "coordination"},
			WorkbenchPrimaryBrowserAction:      "browser action=refresh",
			WorkbenchPrimaryNodeAction:         "nodes action=run",
			WorkbenchNextStep:                  "browser action=refresh",
			WorkbenchRecommendedBrowserActions: []string{"browser action=refresh"},
			WorkbenchRecommendedNodeActions:    []string{"nodes action=run"},
		}

		browserRuntimeApplyWorkbenchSessionSurface(context.Background(), &payload, browserRuntimeWorkbenchSessionSurface{
			SyncCoordinationSurface: true,
		})

		if browserStringSliceContains(payload.WorkbenchSections, "coordination") {
			t.Fatalf("expected workbench session surface helper to drop stale coordination section, got %#v", payload.WorkbenchSections)
		}
		if payload.WorkbenchPrimaryBrowserAction != "" || payload.WorkbenchPrimaryNodeAction != "" || payload.WorkbenchNextStep != "" {
			t.Fatalf("expected workbench session surface helper to clear stale action plan, got %#v", payload)
		}
		if len(payload.WorkbenchRecommendedBrowserActions) != 0 || len(payload.WorkbenchRecommendedNodeActions) != 0 {
			t.Fatalf("expected workbench session surface helper to clear stale recommended actions, got %#v", payload)
		}
		if payload.CoordinationState != "" || payload.CoordinationDecision != "" || payload.CoordinationReady {
			t.Fatalf("expected workbench session surface helper to clear stale coordination summary, got %#v", payload)
		}
	})
}

func TestBrowserRuntimeApplyLifecycleResultProjectionMapsActionDecisionFields(t *testing.T) {
	t.Run("coordinate_restart_uses_restart_fields", func(t *testing.T) {
		payload := browserRuntimePayload{}

		browserRuntimeApplyLifecycleResultProjection(&payload, agentxbrowserruntime.SharedSessionBrowserLifecycleActionOutcome{
			PreparedProfile: "workbench",
			RestartDecision: "session_restarted",
			RestartReady:    true,
		})

		if payload.PreparedProfile != "workbench" {
			t.Fatalf("expected lifecycle result projection to keep prepared profile, got %#v", payload)
		}
		if payload.RestartDecision != "session_restarted" || !payload.RestartReady {
			t.Fatalf("expected coordinate restart projection to populate restart fields, got %#v", payload)
		}
		if payload.PrepareDecision != "" || payload.PrepareReady {
			t.Fatalf("expected coordinate restart projection not to populate prepare fields, got %#v", payload)
		}
	})

	t.Run("coordinate_teardown_uses_prepare_fields", func(t *testing.T) {
		payload := browserRuntimePayload{}

		browserRuntimeApplyLifecycleResultProjection(&payload, agentxbrowserruntime.SharedSessionBrowserLifecycleActionOutcome{
			PreparedProfile: "workbench",
			PrepareDecision: "session_stopped",
			PrepareReady:    true,
		})

		if payload.PreparedProfile != "workbench" {
			t.Fatalf("expected lifecycle result projection to keep prepared profile, got %#v", payload)
		}
		if payload.PrepareDecision != "session_stopped" || !payload.PrepareReady {
			t.Fatalf("expected coordinate teardown projection to populate prepare fields, got %#v", payload)
		}
		if payload.RestartDecision != "" || payload.RestartReady {
			t.Fatalf("expected coordinate teardown projection not to populate restart fields, got %#v", payload)
		}
	})
}

func TestBrowserRuntimeApplyLifecycleActionOutcomeSetsCoordinationDecisionOnError(t *testing.T) {
	payload := browserRuntimePayload{}

	browserRuntimeApplyLifecycleActionOutcome(
		context.Background(),
		"",
		&payload,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		agentxbrowserruntime.SharedSessionBrowserLifecycleActionOutcome{
			Action:                           "coordinate",
			CoordinationGoal:                 "restart",
			Result:                           browserRuntimePrepareResult{Profile: "workbench", Decision: "restart_started", Ready: false},
			PreparedProfile:                  "workbench",
			RestartDecision:                  "restart_started",
			Err:                              errors.New("restart failed"),
			ApplyCoordinationDecisionOnError: true,
		},
	)

	if payload.PreparedProfile != "workbench" {
		t.Fatalf("expected lifecycle action outcome helper to keep prepared profile on error, got %#v", payload)
	}
	if payload.RestartDecision != "restart_started" {
		t.Fatalf("expected lifecycle action outcome helper to populate restart decision on error, got %#v", payload)
	}
	if payload.Status != "error" || payload.Note != "restart failed" {
		t.Fatalf("expected lifecycle action outcome helper to surface error status and note, got %#v", payload)
	}
	if payload.CoordinationDecision != "restart_started" {
		t.Fatalf("expected lifecycle action outcome helper to preserve coordination decision on error, got %#v", payload)
	}
}

func TestBrowserRuntimeFinalizeActionSessionPayloadAppliesWorkbenchAndCoordination(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-action-session-finalize-workbench")
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		WorkbenchSections: []string{"route"},
	}
	bindingEvaluation := &agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
		Coordination: agentxbrowserruntime.SharedSessionBrowserCoordinationEvaluation{
			Plan: agentxbrowserruntime.SharedSessionBrowserCoordinationPlan{
				State:                    "browser_ready",
				BrowserOnNode:            true,
				HasRunningBrowserProfile: true,
				PrimaryNodeAction:        "nodes action=run",
				RecommendedNodeActions:   []string{"nodes action=run"},
			},
			RestartAction: "browser action=refresh",
			Guidance: agentxbrowserruntime.SharedSessionBrowserCoordinationGuidance{
				PrimaryAction:      "browser action=refresh",
				NextStep:           "browser action=refresh",
				RecommendedActions: []string{"browser action=refresh"},
			},
		},
	}

	browserRuntimeFinalizeActionSessionPayload(browserRegistrationContext{}, callCtx, &payload, browserRuntimeActionSessionResultPostProcess{
		Action:            "workbench",
		BindingEvaluation: bindingEvaluation,
	})
	if !browserStringSliceContains(payload.WorkbenchSections, "coordination") {
		t.Fatalf("expected session finalize helper to add coordination section for workbench, got %#v", payload.WorkbenchSections)
	}
	if !payload.WorkbenchReady || payload.WorkbenchPrimaryBrowserAction != "browser action=refresh" || payload.WorkbenchPrimaryNodeAction != "nodes action=run" || payload.WorkbenchNextStep != "browser action=refresh" {
		t.Fatalf("expected session finalize helper to apply workbench action plan from binding evaluation, got %#v", payload)
	}

	refreshPayload := browserRuntimePayload{
		Action: "refresh",
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		RestartDecision: "session_restarted",
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		},
	}
	browserRuntimeFinalizeActionSessionPayload(browserRegistrationContext{}, callCtx, &refreshPayload, browserRuntimeActionSessionResultPostProcess{
		Action:            "refresh",
		CoordinationGoal:  "prepare",
		BindingEvaluation: bindingEvaluation,
	})
	expected := browserRuntimeCoordinationStatus("browser_ready", "restart", refreshPayload.ProfileStatus, false, "session_restarted")
	if refreshPayload.CoordinationState != "browser_ready" {
		t.Fatalf("expected refresh session finalize helper to reuse binding coordination state, got %#v", refreshPayload)
	}
	if refreshPayload.CoordinationDecision != expected.Decision || refreshPayload.CoordinationReady != expected.Ready {
		t.Fatalf("expected refresh session finalize helper to evaluate restart coordination status, got decision=%q ready=%v want %#v", refreshPayload.CoordinationDecision, refreshPayload.CoordinationReady, expected)
	}

	coordinatePayload := browserRuntimePayload{
		Action: "coordinate",
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		PrepareDecision:     "started",
		SyncSessionDecision: "session_route_synced",
		SyncSessionReady:    true,
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		},
	}
	browserRuntimeFinalizeActionSessionPayload(browserRegistrationContext{}, callCtx, &coordinatePayload, browserRuntimeActionSessionResultPostProcess{
		Action:            "coordinate",
		CoordinationGoal:  "sync",
		BindingEvaluation: bindingEvaluation,
	})
	expectedSync := browserRuntimeCoordinationStatus("browser_ready", "sync", coordinatePayload.ProfileStatus, true, "started")
	if coordinatePayload.CoordinationState != "browser_ready" {
		t.Fatalf("expected coordinate session finalize helper to reuse binding coordination state, got %#v", coordinatePayload)
	}
	if coordinatePayload.CoordinationDecision != expectedSync.Decision || coordinatePayload.CoordinationReady != expectedSync.Ready {
		t.Fatalf("expected coordinate session finalize helper to keep lifecycle coordination summary separate from sync-session decision, got decision=%q ready=%v want %#v", coordinatePayload.CoordinationDecision, coordinatePayload.CoordinationReady, expectedSync)
	}
}

func TestBrowserRuntimeFinalizeActionSessionPayloadProjectsWorkbenchSessionInventoryFromBindingEvaluation(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-action-session-finalize-workbench-session-projection")
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
	}
	bindingEvaluation := &agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
		Routes: []agentxbrowserruntime.SharedSessionBrowserRouteSnapshot{{
			Backend:             "proxy",
			Profile:             "workbench",
			RuntimeTarget:       "node",
			BrowserApp:          "Chromium",
			CurrentTargetID:     "tab-2",
			CurrentTargetSource: "tracked_active_tab",
		}},
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
			Runs: []agentxbrowserruntime.SharedSessionRunInfo{{
				RunID:    "run-1",
				NodeID:   "node-1",
				Status:   "running",
				Provider: "browser",
				Action:   "nodes action=run",
			}},
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
				NodeRunCount:         1,
				BrowserProfileCount:  1,
				ActiveBrowserProfile: "workbench",
			},
		},
		Coordination: agentxbrowserruntime.SharedSessionBrowserCoordinationEvaluation{
			Plan: agentxbrowserruntime.SharedSessionBrowserCoordinationPlan{
				State:                    "browser_ready",
				BrowserOnNode:            true,
				HasRunningBrowserProfile: true,
				PrimaryBrowserAction:     "browser action=refresh",
				PrimaryNodeAction:        "nodes action=run",
				NextStep:                 "browser action=refresh",
				RecommendedBrowserActions: []string{
					"browser action=refresh",
				},
				RecommendedNodeActions: []string{
					"nodes action=run",
				},
			},
		},
	}

	browserRuntimeFinalizeActionSessionPayload(browserRegistrationContext{}, callCtx, &payload, browserRuntimeActionSessionResultPostProcess{
		Action:            "workbench",
		BindingEvaluation: bindingEvaluation,
	})

	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].CurrentTargetID != "tab-2" || payload.SessionTargetCount != 1 {
		t.Fatalf("expected workbench session finalize helper to project session routes/count from binding evaluation, got %#v", payload)
	}
	if len(payload.SessionRuns) != 1 || payload.SessionRuns[0].RunID != "run-1" {
		t.Fatalf("expected workbench session finalize helper to project shared runs, got %#v", payload.SessionRuns)
	}
	if len(payload.SessionProfiles) != 1 || payload.SessionProfiles[0].Profile != "workbench" || !payload.SessionProfiles[0].Selected {
		t.Fatalf("expected workbench session finalize helper to project session profiles, got %#v", payload.SessionProfiles)
	}
	if !browserStringSliceContains(payload.WorkbenchSections, "sessions") {
		t.Fatalf("expected workbench session finalize helper to include sessions section when binding evaluation carries session inventory, got %#v", payload.WorkbenchSections)
	}
}

func TestBrowserRuntimeFinalizeActionSessionPayloadBackfillsWorkbenchProfileInventoryFromSessionProjection(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-action-session-finalize-workbench-route-scoped-profiles")
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
	}
	bindingEvaluation := &agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
		Routes: []agentxbrowserruntime.SharedSessionBrowserRouteSnapshot{{
			Backend:             "proxy",
			Profile:             "workbench",
			RuntimeTarget:       "node",
			BrowserApp:          "Chromium",
			CurrentTargetID:     "tab-2",
			CurrentTargetSource: "tracked_active_tab",
		}},
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
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Source:        "tracked_active_tab",
			},
			Summary: agentxbrowserruntime.SharedSessionBrowserBindingSummary{
				CurrentTargetID:      "tab-2",
				RouteTargetCount:     1,
				BrowserProfileCount:  1,
				ActiveBrowserProfile: "workbench",
			},
		},
	}

	browserRuntimeFinalizeActionSessionPayload(browserRegistrationContext{}, callCtx, &payload, browserRuntimeActionSessionResultPostProcess{
		Action:            "workbench",
		BindingEvaluation: bindingEvaluation,
	})

	if len(payload.SessionProfiles) != 1 ||
		payload.SessionProfiles[0].Profile != "workbench" ||
		payload.SessionProfiles[0].RuntimeTarget != "node" ||
		payload.SessionProfiles[0].BrowserApp != "Chromium" ||
		payload.SessionProfiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected workbench finalize helper to synthesize route-scoped session profiles, got %#v", payload.SessionProfiles)
	}
	if payload.ProfileStatus == nil ||
		payload.ProfileStatus.Profile != "workbench" ||
		payload.ProfileStatus.RuntimeTarget != "node" ||
		payload.ProfileStatus.BrowserApp != "Chromium" ||
		payload.ProfileStatus.Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected workbench finalize helper to backfill top-level profile status from session projection, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 1 ||
		payload.Profiles[0].Profile != "workbench" ||
		payload.Profiles[0].RuntimeTarget != "node" ||
		payload.Profiles[0].BrowserApp != "Chromium" ||
		payload.Profiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected workbench finalize helper to backfill top-level profiles from session projection, got %#v", payload.Profiles)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected workbench finalize helper to backfill default profile from session projection, got %#v", payload.DefaultProfile)
	}
}

func TestBrowserRuntimeFinalizeActionSessionPayloadBackfillsProfilesWithoutOverridingExistingStatus(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-action-session-finalize-workbench-existing-status")
	payload := browserRuntimePayload{
		Action: "workbench",
		Status: "ok",
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
	bindingEvaluation := &agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
		Routes: []agentxbrowserruntime.SharedSessionBrowserRouteSnapshot{{
			Backend:             "proxy",
			Profile:             "workbench",
			RuntimeTarget:       "node",
			BrowserApp:          "Chromium",
			CurrentTargetID:     "tab-2",
			CurrentTargetSource: "tracked_active_tab",
		}},
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
				Backend:       "proxy",
				Profile:       "workbench",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Source:        "tracked_active_tab",
			},
			Summary: agentxbrowserruntime.SharedSessionBrowserBindingSummary{
				CurrentTargetID:      "tab-2",
				RouteTargetCount:     1,
				BrowserProfileCount:  1,
				ActiveBrowserProfile: "workbench",
			},
		},
	}

	browserRuntimeFinalizeActionSessionPayload(browserRegistrationContext{}, callCtx, &payload, browserRuntimeActionSessionResultPostProcess{
		Action:            "workbench",
		BindingEvaluation: bindingEvaluation,
	})

	if payload.ProfileStatus == nil || payload.ProfileStatus.Note != "fresh current status" {
		t.Fatalf("expected workbench finalize helper to preserve existing profile status, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Note != "cached route-scoped session snapshot" {
		t.Fatalf("expected workbench finalize helper to still backfill top-level profiles from session projection, got %#v", payload.Profiles)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected workbench finalize helper to backfill default profile when current status already exists, got %#v", payload.DefaultProfile)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected workbench finalize helper to refresh configured profiles when current status already exists, got %#v", payload.ConfiguredProfiles)
	}
}

func TestBrowserRuntimeRefreshCoordinationSummaryUsesActionDecisionAndGoal(t *testing.T) {
	payload := browserRuntimePayload{
		PrepareDecision:     "started",
		RestartDecision:     "restart_started",
		SyncSessionDecision: "session_route_synced",
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		},
		SessionBinding: &browserRuntimeSessionBinding{
			Coordination: &browserRuntimeCoordination{
				State: "browser_ready",
			},
		},
	}

	browserRuntimeRefreshCoordinationSummary(&payload, "refresh", "prepare")
	expectedRefresh := browserRuntimeCoordinationStatus("browser_ready", "restart", payload.ProfileStatus, false, "restart_started")
	if payload.CoordinationDecision != expectedRefresh.Decision || payload.CoordinationReady != expectedRefresh.Ready {
		t.Fatalf("expected coordination summary helper to normalize refresh goal to restart, got decision=%q ready=%v want %#v", payload.CoordinationDecision, payload.CoordinationReady, expectedRefresh)
	}

	payload.CoordinationDecision = ""
	payload.CoordinationReady = false
	payload.CoordinationState = ""
	browserRuntimeRefreshCoordinationSummary(&payload, "coordinate", "sync")
	expectedSync := browserRuntimeCoordinationStatus("browser_ready", "sync", payload.ProfileStatus, false, "started")
	if payload.CoordinationDecision != expectedSync.Decision || payload.CoordinationReady != expectedSync.Ready {
		t.Fatalf("expected coordination summary helper to keep coordinate summary on lifecycle decision while sync result stays separate, got decision=%q ready=%v want %#v", payload.CoordinationDecision, payload.CoordinationReady, expectedSync)
	}
}

func TestBrowserRuntimeApplyCoordinationSummaryProjectionClearsAndWritesPayload(t *testing.T) {
	payload := browserRuntimePayload{
		CoordinationState:    "browser_ready",
		CoordinationDecision: "started",
		CoordinationReady:    true,
	}

	browserRuntimeApplyCoordinationSummaryProjection(
		&payload,
		browserRuntimeCoordinationSummaryProjection{Clear: true},
	)
	if payload.CoordinationState != "" || payload.CoordinationDecision != "" || payload.CoordinationReady {
		t.Fatalf("expected clear projection to reset coordination summary, got %#v", payload)
	}

	browserRuntimeApplyCoordinationSummaryProjection(
		&payload,
		browserRuntimeCoordinationSummaryProjection{
			Summary: &agentxbrowserruntime.SharedSessionBrowserCoordinationSummary{
				State:    "browser_ready",
				Decision: "restart_ready",
				Ready:    true,
			},
		},
	)
	if payload.CoordinationState != "browser_ready" || payload.CoordinationDecision != "restart_ready" || !payload.CoordinationReady {
		t.Fatalf("expected summary projection to write coordination fields, got %#v", payload)
	}
}

func TestBrowserRuntimeRefreshCoordinationSummaryPrefersTypedHealthBlockers(t *testing.T) {
	payload := browserRuntimePayload{
		PrepareDecision: "started",
		RestartDecision: "restart_started",
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		},
		SessionBinding: &browserRuntimeSessionBinding{
			SessionHealthState:          "cooldown_active",
			SessionHealthRecoveryAction: "browser action=wait",
			Coordination: &browserRuntimeCoordination{
				State: "browser_ready",
			},
		},
	}

	browserRuntimeRefreshCoordinationSummary(&payload, "refresh", "prepare")
	if payload.CoordinationState != "browser_ready" {
		t.Fatalf("expected health-blocked coordination summary to preserve shared coordination state, got %#v", payload)
	}
	if payload.CoordinationDecision != "cooldown_active" || payload.CoordinationReady {
		t.Fatalf("expected cooldown_active to override refresh coordination summary, got decision=%q ready=%v", payload.CoordinationDecision, payload.CoordinationReady)
	}

	payload.CoordinationDecision = ""
	payload.CoordinationReady = false
	payload.CoordinationState = ""
	payload.SessionBinding.SessionHealthState = "restart_failed_permanent"
	payload.SessionBinding.SessionHealthRecoveryAction = "browser action=start"
	browserRuntimeRefreshCoordinationSummary(&payload, "coordinate", "sync")
	if payload.CoordinationState != "browser_ready" {
		t.Fatalf("expected permanent restart failure to preserve coordination state, got %#v", payload)
	}
	if payload.CoordinationDecision != "restart_failed_permanent" || payload.CoordinationReady {
		t.Fatalf("expected restart_failed_permanent to override coordinate summary, got decision=%q ready=%v", payload.CoordinationDecision, payload.CoordinationReady)
	}
}

func TestBrowserRuntimeFinalizeActionSessionPayloadUsesTypedHealthBlockersForCoordinationSummary(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-action-session-finalize-health-blocker")
	payload := browserRuntimePayload{
		Action: "refresh",
		Status: "ok",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
		RestartDecision: "session_restarted",
		ProfileStatus: &browserRuntimeProfileState{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
			Status:        "running",
			Running:       true,
			Connected:     true,
		},
	}
	bindingEvaluation := &agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
		Health: agentxbrowserruntime.SharedSessionBrowserHealthEvaluation{
			Summary: &agentxbrowserruntime.SharedSessionBrowserHealthSummary{
				State:          "restart_failed_permanent",
				Reason:         "browser relaunch failed twice",
				RecoveryAction: "browser action=start",
			},
		},
		Coordination: agentxbrowserruntime.SharedSessionBrowserCoordinationEvaluation{
			Plan: agentxbrowserruntime.SharedSessionBrowserCoordinationPlan{
				State:                    "browser_ready",
				BrowserOnNode:            true,
				HasRunningBrowserProfile: true,
			},
			RestartAction: "browser action=start",
			Guidance: agentxbrowserruntime.SharedSessionBrowserCoordinationGuidance{
				PrimaryAction:      "browser action=start",
				NextStep:           "browser action=start",
				RecommendedActions: []string{"browser action=start", "browser action=ensure"},
			},
		},
	}

	browserRuntimeFinalizeActionSessionPayload(browserRegistrationContext{}, callCtx, &payload, browserRuntimeActionSessionResultPostProcess{
		Action:            "refresh",
		BindingEvaluation: bindingEvaluation,
	})
	if payload.CoordinationState != "browser_ready" {
		t.Fatalf("expected finalize helper to keep shared coordination state while health blocks restart, got %#v", payload)
	}
	if payload.CoordinationDecision != "restart_failed_permanent" || payload.CoordinationReady {
		t.Fatalf("expected finalize helper to surface typed health blocker instead of restart_ready, got decision=%q ready=%v", payload.CoordinationDecision, payload.CoordinationReady)
	}
	if payload.SessionBinding == nil || payload.SessionBinding.Coordination == nil || payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser action=start" {
		t.Fatalf("expected finalize helper to preserve shared restart guidance, got %#v", payload.SessionBinding)
	}
}

func TestBrowserRuntimeApplyInspectionActionWorkbenchPopulatesPayload(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-inspection-workbench")
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	backend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
				PlaywrightCache: &agentxbrowserruntime.BrowserPlaywrightCacheSummary{
					HostOS:                     "darwin",
					HostArch:                   "arm64",
					NodeVersion:                "24.2.0",
					PlaywrightPackageVersion:   "1.55.0",
					RuntimeSummaryGeneration:   "runtime-workbench",
					RuntimeBaselineReady:       true,
					SelectedLaunchSource:       "runtime_observed",
					SelectedLaunchDelivery:     "delivery-workbench",
					SelectedLaunchReady:        true,
					SelectedLaunchExecutableOK: true,
					LaunchReady:                true,
				},
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}}
	sessionRegistry.TrackCurrentTarget("runtime-inspection-workbench", BrowserSessionTarget{
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
		watchManagerProvider: agentxbrowserruntime.SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, browserRuntimeReconnectWatchdogWindow),
	}
	payload := browserRuntimePayload{
		Action:    "workbench",
		Status:    "ok",
		SessionID: "runtime-inspection-workbench",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}
	bindingEvaluation := browserRuntimeApplyInspectionAction(
		ctx,
		callCtx,
		&payload,
		ctx.watchManagerProvider.Bind(backend),
		browserRuntimeInspectionActionOptions{
			Action:           "workbench",
			SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			EffectiveProfile: "workbench",
			Capabilities: BrowserCapabilities{
				RuntimeStatus:    true,
				RuntimeWorkbench: true,
				RuntimeList:      true,
				RuntimeSessions:  true,
			},
		},
	)
	if bindingEvaluation == nil {
		t.Fatalf("expected workbench inspection helper to return binding evaluation")
	}
	for _, want := range []string{"route", "status", "profiles", "sessions"} {
		if !browserStringSliceContains(payload.WorkbenchSections, want) {
			t.Fatalf("expected workbench inspection helper to include %q section, got %#v", want, payload.WorkbenchSections)
		}
	}
	if !payload.WorkbenchReady {
		t.Fatalf("expected workbench inspection helper to mark payload ready, got %#v", payload)
	}
	if payload.ProfileStatus == nil || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "running" {
		t.Fatalf("expected workbench inspection helper to populate profile status, got %#v", payload.ProfileStatus)
	}
	if payload.LaunchDiagnostics == nil ||
		payload.LaunchDiagnostics.Source != "runtime_status" ||
		payload.LaunchDiagnostics.RuntimeTarget != "node" ||
		payload.LaunchDiagnostics.RuntimeSummaryGeneration != "runtime-workbench" ||
		payload.LaunchDiagnostics.LaunchReady == nil || !*payload.LaunchDiagnostics.LaunchReady {
		t.Fatalf("expected workbench inspection helper to project launch diagnostics, got %#v", payload.LaunchDiagnostics)
	}
	if payload.WorkbenchLaunchDiagnostics == nil ||
		payload.WorkbenchLaunchDiagnostics.SelectedLaunchSource != "runtime_observed" ||
		payload.WorkbenchLaunchDiagnostics.SelectedLaunchDeliveryGeneration != "delivery-workbench" ||
		payload.WorkbenchLaunchDiagnostics.SelectedLaunchReady == nil || !*payload.WorkbenchLaunchDiagnostics.SelectedLaunchReady {
		t.Fatalf("expected workbench inspection helper to project workbench launch diagnostics, got %#v", payload.WorkbenchLaunchDiagnostics)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Profile != "workbench" {
		t.Fatalf("expected workbench inspection helper to populate profiles, got %#v", payload.Profiles)
	}
	if payload.DefaultProfile != "workbench" || !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected workbench inspection helper to populate default/configured profiles, got %#v", payload)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionTargetCount != 1 {
		t.Fatalf("expected workbench inspection helper to populate session view, got routes=%#v target_count=%d", payload.SessionRoutes, payload.SessionTargetCount)
	}
}

func TestBrowserRuntimeApplyInspectionActionProfilesErrorPopulatesPayload(t *testing.T) {
	callCtx := WithToolSessionID(context.Background(), "runtime-inspection-profiles-error")
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	backend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesErr: errors.New("profiles unavailable"),
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}}
	ctx := browserRegistrationContext{
		sessionRegistry:      sessionRegistry,
		sessionStateRegistry: stateRegistry,
		watchManagerProvider: agentxbrowserruntime.SharedSessionBrowserObserverManagerFor(sessionRegistry, nil, stateRegistry, browserRuntimeReconnectWatchdogWindow),
	}
	payload := browserRuntimePayload{
		Action:    "profiles",
		Status:    "ok",
		SessionID: "runtime-inspection-profiles-error",
		SelectedRoute: &browserRuntimeRouteDescriptor{
			Backend:       "proxy",
			Profile:       "workbench",
			RuntimeTarget: "node",
		},
	}

	bindingEvaluation := browserRuntimeApplyInspectionAction(
		ctx,
		callCtx,
		&payload,
		ctx.watchManagerProvider.Bind(backend),
		browserRuntimeInspectionActionOptions{
			Action:           "profiles",
			SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			EffectiveProfile: "workbench",
			Capabilities: BrowserCapabilities{
				RuntimeStatus: true,
				RuntimeList:   true,
			},
		},
	)
	if bindingEvaluation == nil {
		t.Fatalf("expected profiles inspection helper to return binding evaluation")
	}
	if payload.Status != "error" || payload.Note != "profiles unavailable" {
		t.Fatalf("expected profiles inspection helper to apply shared terminal status/note, got %#v", payload)
	}
	if len(payload.Profiles) != 0 || payload.ProfileStatus != nil {
		t.Fatalf("expected profiles inspection helper not to synthesize inventory on error, got %#v", payload)
	}
}

func TestBrowserRuntimeRefreshWorkbenchProjectionIncludesSectionsAndExtras(t *testing.T) {
	payload := browserRuntimePayload{
		Profiles: []browserRuntimeProfileState{
			{Profile: "workbench", Backend: "proxy", RuntimeTarget: "node"},
		},
		ProfileStatus: &browserRuntimeProfileState{
			Profile:       "workbench",
			Backend:       "proxy",
			RuntimeTarget: "node",
			Status:        "running",
		},
		SessionRoutes: []browserRuntimeSessionRoute{
			{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node"},
		},
	}

	browserRuntimeRefreshWorkbenchProjection(&payload, "coordination")

	for _, want := range []string{"route", "status", "profiles", "sessions", "coordination"} {
		if !browserStringSliceContains(payload.WorkbenchSections, want) {
			t.Fatalf("expected workbench projection helper to include %q, got %#v", want, payload.WorkbenchSections)
		}
	}
	if !payload.WorkbenchReady {
		t.Fatalf("expected workbench projection helper to mark payload ready, got %#v", payload)
	}
}

func TestBrowserRuntimeSyncWorkbenchProjectionClearsStaleActionPlanWithoutCoordination(t *testing.T) {
	payload := browserRuntimePayload{
		SessionRoutes: []browserRuntimeSessionRoute{
			{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node"},
		},
		SessionTargetCount:                 1,
		CoordinationState:                  "stale_state",
		CoordinationDecision:               "stale_decision",
		CoordinationReady:                  true,
		WorkbenchSections:                  []string{"route", "sessions", "coordination"},
		WorkbenchReady:                     true,
		WorkbenchPrimaryBrowserAction:      "browser action=refresh",
		WorkbenchPrimaryNodeAction:         "nodes action=run",
		WorkbenchNextStep:                  "browser action=refresh",
		WorkbenchRecommendedBrowserActions: []string{"browser action=refresh"},
		WorkbenchRecommendedNodeActions:    []string{"nodes action=run"},
	}

	browserRuntimeSyncWorkbenchProjection(&payload, browserRuntimeWorkbenchProjectionSync{
		ClearActionPlan: true,
	})

	if browserStringSliceContains(payload.WorkbenchSections, "coordination") {
		t.Fatalf("expected workbench projection sync helper to drop stale coordination section, got %#v", payload.WorkbenchSections)
	}
	if !browserStringSliceContains(payload.WorkbenchSections, "route") || !browserStringSliceContains(payload.WorkbenchSections, "sessions") || !payload.WorkbenchReady {
		t.Fatalf("expected workbench projection sync helper to preserve remaining route/session sections, got %#v", payload)
	}
	if payload.WorkbenchPrimaryBrowserAction != "" || payload.WorkbenchPrimaryNodeAction != "" || payload.WorkbenchNextStep != "" {
		t.Fatalf("expected workbench projection sync helper to clear stale action plan, got %#v", payload)
	}
	if len(payload.WorkbenchRecommendedBrowserActions) != 0 || len(payload.WorkbenchRecommendedNodeActions) != 0 {
		t.Fatalf("expected workbench projection sync helper to clear stale recommended actions, got %#v", payload)
	}
	if payload.CoordinationState != "" || payload.CoordinationDecision != "" || payload.CoordinationReady {
		t.Fatalf("expected workbench projection sync helper to clear stale coordination summary, got %#v", payload)
	}
}

func TestBrowserRuntimeSyncWorkbenchProjectionPrefersPendingReviewDiagnostics(t *testing.T) {
	payload := browserRuntimePayload{
		Action: "workbench",
		SessionBinding: &browserRuntimeSessionBinding{
			Coordination: &browserRuntimeCoordination{
				State:                     "coordinated",
				PrimaryBrowserAction:      "browser action=tabs",
				PrimaryNodeAction:         "nodes action=run_status",
				NextStep:                  "browser action=tabs",
				RecommendedBrowserActions: []string{"browser action=tabs", "browser action=focus", "browser action=close", "browser action=pin_target"},
				RecommendedNodeActions:    []string{"nodes action=run_status"},
			},
		},
		SessionRoutes: []browserRuntimeSessionRoute{
			{
				Backend:            "proxy",
				Profile:            "workbench",
				RuntimeTarget:      "node",
				FollowPolicyState:  "popup_review_required",
				FollowPolicyReason: "browser session has a pending popup target \"Offer\"; rerun with force=true before following or adopting it",
			},
		},
	}

	browserRuntimeSyncWorkbenchProjection(&payload, browserRuntimeWorkbenchProjectionSync{
		SyncCoordinationSurface: true,
	})

	if payload.WorkbenchExplanation == nil ||
		payload.WorkbenchExplanation.Category != "review" ||
		payload.WorkbenchExplanation.State != "manual_confirmation_required" ||
		payload.WorkbenchExplanation.SummaryCode != "popup_review_required" ||
		payload.WorkbenchExplanation.NextStepAlias != "tabs" ||
		payload.WorkbenchExplanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("expected workbench explanation to surface pending review blocker, got %#v", payload.WorkbenchExplanation)
	}
	if payload.WorkbenchDiagnostics == nil ||
		payload.WorkbenchDiagnostics.Category != "review" ||
		payload.WorkbenchDiagnostics.State != "manual_confirmation_required" ||
		payload.WorkbenchDiagnostics.SummaryCode != "popup_review_required" ||
		payload.WorkbenchDiagnostics.NextStepAlias != "tabs" ||
		payload.WorkbenchDiagnostics.ManualRetryHint != "rerun_with_force" ||
		payload.WorkbenchDiagnostics.PrimaryBrowserAction != "browser action=tabs" ||
		payload.WorkbenchDiagnostics.PrimaryNodeAction != "nodes action=run_status" ||
		payload.WorkbenchDiagnostics.NextStep != "browser action=tabs" {
		t.Fatalf("expected workbench diagnostics to reuse review blocker plus action plan, got %#v", payload.WorkbenchDiagnostics)
	}
	if payload.WorkbenchSummary == nil ||
		payload.WorkbenchSummary.Category != "review" ||
		payload.WorkbenchSummary.State != "manual_confirmation_required" ||
		payload.WorkbenchSummary.SummaryCode != "popup_review_required" ||
		payload.WorkbenchSummary.NextStepAlias != "tabs" ||
		payload.WorkbenchSummary.ManualRetryHint != "rerun_with_force" ||
		payload.WorkbenchSummary.PrimaryBrowserAction != "browser action=tabs" ||
		payload.WorkbenchSummary.NextStep != "browser action=tabs" {
		t.Fatalf("expected workbench summary to mirror review diagnostics, got %#v", payload.WorkbenchSummary)
	}
	if payload.WorkbenchDisplay == nil ||
		!payload.WorkbenchDisplay.Ready ||
		!browserStringSliceContains(payload.WorkbenchDisplay.Sections, "coordination") ||
		payload.WorkbenchDisplay.Category != "review" ||
		payload.WorkbenchDisplay.State != "manual_confirmation_required" ||
		payload.WorkbenchDisplay.SummaryCode != "popup_review_required" ||
		payload.WorkbenchDisplay.NextStepAlias != "tabs" ||
		payload.WorkbenchDisplay.ManualRetryHint != "rerun_with_force" ||
		payload.WorkbenchDisplay.PrimaryBrowserAction != "browser action=tabs" ||
		payload.WorkbenchDisplay.NextStep != "browser action=tabs" {
		t.Fatalf("expected workbench display to mirror review diagnostics, got %#v", payload.WorkbenchDisplay)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "review" ||
		payload.Explanation.State != "manual_confirmation_required" ||
		payload.Explanation.SummaryCode != "popup_review_required" ||
		payload.Explanation.NextStepAlias != "tabs" ||
		payload.Explanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("expected top-level explanation alias to mirror review diagnostics, got %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "review" ||
		payload.Diagnostics.State != "manual_confirmation_required" ||
		payload.Diagnostics.SummaryCode != "popup_review_required" ||
		payload.Diagnostics.NextStepAlias != "tabs" ||
		payload.Diagnostics.ManualRetryHint != "rerun_with_force" ||
		payload.Diagnostics.PrimaryBrowserAction != "browser action=tabs" ||
		payload.Diagnostics.NextStep != "browser action=tabs" {
		t.Fatalf("expected top-level diagnostics alias to mirror review diagnostics, got %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "review" ||
		payload.Summary.State != "manual_confirmation_required" ||
		payload.Summary.SummaryCode != "popup_review_required" ||
		payload.Summary.NextStepAlias != "tabs" ||
		payload.Summary.ManualRetryHint != "rerun_with_force" ||
		payload.Summary.PrimaryBrowserAction != "browser action=tabs" ||
		payload.Summary.NextStep != "browser action=tabs" {
		t.Fatalf("expected top-level summary to mirror review diagnostics, got %#v", payload.Summary)
	}
	if payload.Display == nil ||
		!payload.Display.Ready ||
		!browserStringSliceContains(payload.Display.Sections, "coordination") ||
		payload.Display.Category != "review" ||
		payload.Display.State != "manual_confirmation_required" ||
		payload.Display.SummaryCode != "popup_review_required" ||
		payload.Display.NextStepAlias != "tabs" ||
		payload.Display.ManualRetryHint != "rerun_with_force" ||
		payload.Display.PrimaryBrowserAction != "browser action=tabs" ||
		payload.Display.NextStep != "browser action=tabs" {
		t.Fatalf("expected top-level display to mirror review diagnostics, got %#v", payload.Display)
	}
	if payload.Workbench == nil ||
		payload.Workbench.Review == nil ||
		payload.Workbench.Review.PolicyState != "popup_review_required" ||
		payload.Workbench.Review.Decision != "session_target_popup_review_required" ||
		payload.Workbench.Review.Summary == nil ||
		payload.Workbench.Review.Summary.SummaryCode != "popup_review_required" ||
		payload.Workbench.Review.Summary.ManualRetryHint != "rerun_with_force" ||
		payload.Workbench.Review.Display == nil ||
		payload.Workbench.Review.Display.SummaryCode != "popup_review_required" ||
		payload.Workbench.Review.Display.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("expected workbench review surface to mirror review diagnostics, got %#v", payload.Workbench)
	}
	if payload.Review == nil ||
		payload.Review.PolicyState != "popup_review_required" ||
		payload.Review.Decision != "session_target_popup_review_required" ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.SummaryCode != "popup_review_required" ||
		payload.Review.Summary.ManualRetryHint != "rerun_with_force" ||
		payload.Review.Display == nil ||
		payload.Review.Display.SummaryCode != "popup_review_required" ||
		payload.Review.Display.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("expected top-level review surface to mirror review diagnostics, got %#v", payload.Review)
	}
	if payload.View == nil ||
		payload.View.Kind != "workbench" ||
		!payload.View.Ready ||
		!browserStringSliceContains(payload.View.Sections, "coordination") ||
		payload.View.Category != "review" ||
		payload.View.State != "manual_confirmation_required" ||
		payload.View.SummaryCode != "popup_review_required" ||
		payload.View.NextStepAlias != "tabs" ||
		payload.View.ManualRetryHint != "rerun_with_force" ||
		payload.View.PrimaryBrowserAction != "browser action=tabs" ||
		payload.View.NextStep != "browser action=tabs" ||
		payload.View.Review == nil ||
		payload.View.Review.PolicyState != "popup_review_required" ||
		payload.View.Review.Decision != "session_target_popup_review_required" ||
		payload.View.Review.Display == nil ||
		payload.View.Review.Display.ManualRetryHint != "rerun_with_force" ||
		payload.View.Review.Ready {
		t.Fatalf("expected workbench projection to build top-level view alias, got %#v", payload.View)
	}
}

func TestBrowserRuntimeHideInspectionProjectionClearsStatusProjection(t *testing.T) {
	payload := browserRuntimePayload{
		Action:         "status",
		DefaultProfile: "workbench",
		ProfileStatus: &browserRuntimeProfileState{
			Profile:       "workbench",
			Backend:       "proxy",
			RuntimeTarget: "node",
			Status:        "running",
		},
	}

	browserRuntimeHideInspectionProjection(&payload, "status")

	if payload.ProfileStatus != nil || payload.DefaultProfile != "" {
		t.Fatalf("expected inspection projection hide helper to clear status projection, got %#v", payload)
	}
}

func TestBrowserRuntimeHideInspectionProjectionClearsProfilesProjection(t *testing.T) {
	payload := browserRuntimePayload{
		Action:             "profiles",
		DefaultProfile:     "workbench",
		discoveredProfiles: []string{"workbench", "relay"},
		Profiles: []browserRuntimeProfileState{
			{Profile: "workbench", Backend: "proxy", RuntimeTarget: "node"},
			{Profile: "relay", Backend: "proxy", RuntimeTarget: "node"},
		},
	}

	browserRuntimeHideInspectionProjection(&payload, "profiles")

	if len(payload.Profiles) != 0 || payload.DefaultProfile != "" {
		t.Fatalf("expected inspection projection hide helper to clear profiles projection, got %#v", payload)
	}
	if len(payload.discoveredProfiles) != 0 {
		t.Fatalf("expected inspection projection hide helper to clear discovered profiles with hidden profile inventory, got %#v", payload.discoveredProfiles)
	}
}

func TestBrowserRuntimeHideInspectionProjectionClearsWorkbenchProjectionAndPlan(t *testing.T) {
	payload := browserRuntimePayload{
		Action:         "workbench",
		DefaultProfile: "workbench",
		discoveredProfiles: []string{
			"workbench",
			"relay",
		},
		Profiles: []browserRuntimeProfileState{
			{Profile: "workbench", Backend: "proxy", RuntimeTarget: "node"},
		},
		ProfileStatus: &browserRuntimeProfileState{
			Profile:       "workbench",
			Backend:       "proxy",
			RuntimeTarget: "node",
			Status:        "running",
		},
		SessionRoutes: []browserRuntimeSessionRoute{
			{Backend: "proxy", Profile: "workbench", RuntimeTarget: "node"},
		},
		SessionTargetCount:                 1,
		WorkbenchSections:                  []string{"route", "status", "profiles", "sessions", "coordination"},
		WorkbenchReady:                     true,
		WorkbenchPrimaryBrowserAction:      "browser action=refresh",
		WorkbenchPrimaryNodeAction:         "nodes action=run",
		WorkbenchNextStep:                  "browser action=refresh",
		WorkbenchRecommendedBrowserActions: []string{"browser action=refresh"},
		WorkbenchRecommendedNodeActions:    []string{"nodes action=run"},
	}

	browserRuntimeHideInspectionProjection(&payload, "workbench")

	if payload.ProfileStatus != nil || len(payload.Profiles) != 0 || payload.DefaultProfile != "" {
		t.Fatalf("expected inspection projection hide helper to clear workbench top-level status/profile projection, got %#v", payload)
	}
	if len(payload.discoveredProfiles) != 0 {
		t.Fatalf("expected inspection projection hide helper to clear hidden workbench discovered profiles, got %#v", payload.discoveredProfiles)
	}
	if !browserStringSliceContains(payload.WorkbenchSections, "route") || !browserStringSliceContains(payload.WorkbenchSections, "sessions") {
		t.Fatalf("expected inspection projection hide helper to keep route/sessions sections, got %#v", payload.WorkbenchSections)
	}
	for _, hidden := range []string{"status", "profiles", "coordination"} {
		if browserStringSliceContains(payload.WorkbenchSections, hidden) {
			t.Fatalf("expected inspection projection hide helper to remove %q section, got %#v", hidden, payload.WorkbenchSections)
		}
	}
	if !payload.WorkbenchReady {
		t.Fatalf("expected inspection projection hide helper to keep workbench ready when route/session summary remain, got %#v", payload)
	}
	if payload.WorkbenchPrimaryBrowserAction != "" || payload.WorkbenchPrimaryNodeAction != "" || payload.WorkbenchNextStep != "" {
		t.Fatalf("expected inspection projection hide helper to clear workbench plan fields, got %#v", payload)
	}
	if len(payload.WorkbenchRecommendedBrowserActions) != 0 || len(payload.WorkbenchRecommendedNodeActions) != 0 {
		t.Fatalf("expected inspection projection hide helper to clear workbench recommended actions, got %#v", payload)
	}
}
