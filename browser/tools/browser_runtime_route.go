package tools

import (
	"context"
	"fmt"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

var supportedBrowserRuntimeTargets = map[string]bool{
	"host":    true,
	"sandbox": true,
	"node":    true,
}

type browserExecutionRouteResolver interface {
	ResolveBrowserExecutionRoute(BrowserRuntimeInfo) (browserResolvedExecutionRoute, error)
}

type browserSessionExecutionRouteResolver interface {
	ResolveBrowserExecutionRouteForSession(context.Context, map[string]any, BrowserRuntimeInfo, bool) (browserResolvedExecutionRoute, error)
}

type browserSessionRuntimeRouteResolver interface {
	ResolveBrowserRuntimeRouteForSession(context.Context, map[string]any, BrowserRuntimeInfo, bool) (BrowserRuntimeInfo, error)
}

type browserRuntimeSessionSelectionPreview struct {
	Applied                map[string]any
	RequestedRuntimeTarget string
	RequestedBrowserApp    string
}

func browserRequestedRuntimeInfo(params map[string]any, base BrowserRuntimeInfo) (BrowserRuntimeInfo, bool, bool, error) {
	base = normalizeBrowserRuntimeInfo(base)
	profileRaw := strings.ToLower(strings.TrimSpace(firstString(params, "profile", "browser_profile", "runtime_profile")))
	targetRaw := browserRequestedRuntimeTarget(params)
	explicitProfile := profileRaw != ""
	explicitTarget := targetRaw != ""

	if explicitProfile && profileRaw == "" {
		return BrowserRuntimeInfo{}, false, false, fmt.Errorf("profile must be a non-empty string")
	}
	if explicitTarget && !supportedBrowserRuntimeTargets[targetRaw] {
		return BrowserRuntimeInfo{}, false, false, fmt.Errorf("runtime_target must be one of host, sandbox, node")
	}

	requested := base
	if explicitProfile {
		requested.Profile = profileRaw
	}
	if explicitTarget {
		if !explicitProfile && targetRaw != "" && !strings.EqualFold(targetRaw, base.Target) {
			// Switching runtime target without an explicit profile should let the
			// destination backend resolve its own default profile instead of
			// inheriting the source route's profile (for example host=default ->
			// node=isolated).
			requested.Profile = ""
		}
		requested.Target = targetRaw
	}
	requested = normalizeBrowserRuntimeInfo(requested)
	return requested, explicitProfile, explicitTarget, nil
}

func resolveBrowserRuntimeRoute(params map[string]any, base BrowserRuntimeInfo, backend BrowserBackend) (BrowserRuntimeInfo, error) {
	requested, explicitProfile, explicitTarget, err := browserRequestedRuntimeInfoForDefaultRequestBase(params, base, backend)
	if err != nil {
		return BrowserRuntimeInfo{}, err
	}
	return resolveBrowserRuntimeRouteInfo(requested, explicitProfile, explicitTarget, base, backend)
}

func resolveBrowserRuntimeRouteInfo(requested BrowserRuntimeInfo, explicitProfile bool, explicitTarget bool, base BrowserRuntimeInfo, backend BrowserBackend) (BrowserRuntimeInfo, error) {
	if resolver, ok := backend.(BrowserRuntimeRouteResolver); ok {
		resolved, err := resolver.ResolveBrowserRuntimeRoute(requested)
		if err != nil {
			return BrowserRuntimeInfo{}, err
		}
		resolved = normalizeBrowserRuntimeInfo(resolved)
		if resolved.Backend == "" {
			resolved.Backend = requested.Backend
		}
		if explicitProfile && resolved.Profile == "" {
			resolved.Profile = requested.Profile
		}
		if explicitTarget && resolved.Target == "" {
			resolved.Target = requested.Target
		}
		return resolved, nil
	}
	if !explicitProfile && !explicitTarget {
		return requested, nil
	}
	if explicitProfile {
		if base.Profile == "" {
			return BrowserRuntimeInfo{}, fmt.Errorf("profile %q requires a backend that advertises runtime profiles", requested.Profile)
		}
		if requested.Profile != base.Profile {
			return BrowserRuntimeInfo{}, fmt.Errorf("profile %q is unsupported for backend %q (configured profile=%q)", requested.Profile, browserRuntimeBackendName(base), base.Profile)
		}
	}
	if explicitTarget {
		if base.Target == "" {
			return BrowserRuntimeInfo{}, fmt.Errorf("runtime_target %q requires a backend that advertises runtime targets", requested.Target)
		}
		if requested.Target != base.Target {
			return BrowserRuntimeInfo{}, fmt.Errorf("runtime_target %q is unsupported for backend %q (configured target=%q)", requested.Target, browserRuntimeBackendName(base), base.Target)
		}
	}
	return requested, nil
}

func browserResolveRuntimeRouteFromSessionPreview(params map[string]any, preview browserRuntimeSessionSelectionPreview, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, backend BrowserBackend) (BrowserRuntimeInfo, bool, error) {
	explicitRuntimeTarget := browserHasExplicitRuntimeTarget(params)
	runtimeInfo, err := resolveBrowserRuntimeRoute(preview.Applied, base, backend)
	if err == nil || !hiddenImplicitHostDefaultBase || explicitRuntimeTarget {
		return runtimeInfo, explicitRuntimeTarget, err
	}
	runtimeInfo, _, _, err = browserRequestedRuntimeInfo(params, BrowserRuntimeInfo{})
	return runtimeInfo, explicitRuntimeTarget, err
}

func browserResolveExecutionRouteFromSessionPreview(params map[string]any, preview browserRuntimeSessionSelectionPreview, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, backend BrowserBackend) (browserResolvedExecutionRoute, error) {
	runtimeInfo, _, err := browserResolveRuntimeRouteFromSessionPreview(params, preview, base, hiddenImplicitHostDefaultBase, backend)
	if err != nil {
		return browserResolvedExecutionRoute{}, err
	}
	routedBackend, routedInfo, err := resolveBrowserExecutionBackendForRoute(runtimeInfo, backend)
	if err != nil {
		return browserResolvedExecutionRoute{}, err
	}
	return browserNormalizeResolvedExecutionRoute(browserResolvedExecutionRoute{
		Backend:      routedBackend,
		RuntimeInfo:  routedInfo,
		Capabilities: browserCapabilitiesForConcreteBackend(routedBackend),
	}, runtimeInfo, false, false), nil
}

type browserRuntimeActionDispatchSelection struct {
	DefaultCandidateRoute                        BrowserRuntimeInfo
	DefaultCandidateDescriptor                   browserRuntimeRouteDescriptor
	Requested                                    BrowserRuntimeInfo
	ExplicitRuntimeTarget                        bool
	HiddenRequestedRuntimeTarget                 bool
	SelectedRoute                                browserResolvedExecutionRoute
	SelectedRouteReady                           bool
	RouteErr                                     error
	CanUseManagedSessionRouteForImplicitFallback bool
	UseHiddenImplicitHostDiagnosticsDegrade      bool
}

type browserResolvedRuntimeRequestFromSessionPreview struct {
	DefaultCandidateRoute        BrowserRuntimeInfo
	DefaultCandidateDescriptor   browserRuntimeRouteDescriptor
	Requested                    BrowserRuntimeInfo
	ExplicitRuntimeTarget        bool
	HiddenRequestedRuntimeTarget bool
}

func browserResolveRuntimeRequestFromSessionPreview(
	params map[string]any,
	preview browserRuntimeSessionSelectionPreview,
	base BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
	backend BrowserBackend,
) (browserResolvedRuntimeRequestFromSessionPreview, error) {
	defaultCandidateRoute := browserRuntimeDefaultCandidateRouteForSessionPreview(base, hiddenImplicitHostDefaultBase, backend)
	defaultCandidateDescriptor := browserRuntimeDefaultCandidateRouteDescriptorForSessionPreview(base, hiddenImplicitHostDefaultBase, backend)
	requested, explicitRuntimeTarget, err := browserResolveRuntimeRouteFromSessionPreview(
		params,
		preview,
		base,
		hiddenImplicitHostDefaultBase,
		backend,
	)
	if err != nil {
		requested, _, explicitRuntimeTarget, err = browserRequestedRuntimeInfo(preview.Applied, base)
		if err != nil {
			return browserResolvedRuntimeRequestFromSessionPreview{}, err
		}
	}
	return browserResolvedRuntimeRequestFromSessionPreview{
		DefaultCandidateRoute:      defaultCandidateRoute,
		DefaultCandidateDescriptor: defaultCandidateDescriptor,
		Requested:                  requested,
		ExplicitRuntimeTarget:      explicitRuntimeTarget,
		HiddenRequestedRuntimeTarget: hiddenImplicitHostDefaultBase &&
			!explicitRuntimeTarget &&
			strings.TrimSpace(preview.RequestedRuntimeTarget) != "" &&
			strings.TrimSpace(requested.Target) == "",
	}, nil
}

type browserRuntimeActionDispatchRoutePreflight struct {
	DefaultCandidateRoute                        BrowserRuntimeInfo
	DefaultCandidateDescriptor                   browserRuntimeRouteDescriptor
	Requested                                    BrowserRuntimeInfo
	ExplicitRuntimeTarget                        bool
	HiddenRequestedRuntimeTarget                 bool
	SelectedRoute                                browserResolvedExecutionRoute
	SelectedRouteReady                           bool
	RouteErr                                     error
	CanUseManagedSessionRouteForImplicitFallback bool
}

func browserRuntimeCanUseManagedImplicitFallback(selection browserRuntimeActionDispatchSelection) bool {
	if selection.CanUseManagedSessionRouteForImplicitFallback {
		return true
	}
	if !selection.SelectedRouteReady {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(selection.SelectedRoute.RuntimeInfo.Target)) {
	case "node", "sandbox":
		return true
	default:
		return false
	}
}

func browserResolveRuntimeActionDispatchRoutePreflight(
	ctx context.Context,
	preview browserRuntimeSessionSelectionPreview,
	params map[string]any,
	requestedTarget string,
	base BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
	backend BrowserBackend,
) (browserRuntimeActionDispatchRoutePreflight, error) {
	requestedTarget = strings.TrimSpace(requestedTarget)
	requestPreview, err := browserResolveRuntimeRequestFromSessionPreview(
		params,
		preview,
		base,
		hiddenImplicitHostDefaultBase,
		backend,
	)
	if err != nil {
		return browserRuntimeActionDispatchRoutePreflight{}, err
	}
	preflight := browserRuntimeActionDispatchRoutePreflight{
		DefaultCandidateRoute:        requestPreview.DefaultCandidateRoute,
		DefaultCandidateDescriptor:   requestPreview.DefaultCandidateDescriptor,
		Requested:                    requestPreview.Requested,
		ExplicitRuntimeTarget:        requestPreview.ExplicitRuntimeTarget,
		HiddenRequestedRuntimeTarget: requestPreview.HiddenRequestedRuntimeTarget,
	}
	if hiddenImplicitHostDefaultBase && requestedTarget == "" {
		selection := browserResolveRuntimeRouteSelectionPreflight(func() (browserResolvedExecutionRoute, error) {
			return resolveBrowserExecutionRouteForSessionPreview(
				ctx,
				preview,
				params,
				base,
				hiddenImplicitHostDefaultBase,
				backend,
			)
		})
		preflight.SelectedRoute = selection.SelectedRoute
		preflight.SelectedRouteReady = selection.SelectedRouteReady
		preflight.RouteErr = selection.RouteErr
		preflight.CanUseManagedSessionRouteForImplicitFallback = selection.CanUseManagedSessionRouteForImplicitFallback
	}
	return preflight, nil
}

func browserResolveRuntimeActionDispatchSelection(ctx context.Context, preview browserRuntimeSessionSelectionPreview, params map[string]any, action string, requestedProfile string, requestedTarget string, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, backend BrowserBackend) (browserRuntimeActionDispatchSelection, error) {
	selection := browserRuntimeActionDispatchSelection{}
	requestedTarget = strings.TrimSpace(requestedTarget)
	preflight, err := browserResolveRuntimeActionDispatchRoutePreflight(
		ctx,
		preview,
		params,
		requestedTarget,
		base,
		hiddenImplicitHostDefaultBase,
		backend,
	)
	if err != nil {
		return browserRuntimeActionDispatchSelection{}, err
	}
	selection.DefaultCandidateRoute = preflight.DefaultCandidateRoute
	selection.DefaultCandidateDescriptor = preflight.DefaultCandidateDescriptor
	selection.Requested = preflight.Requested
	selection.ExplicitRuntimeTarget = preflight.ExplicitRuntimeTarget
	selection.HiddenRequestedRuntimeTarget = preflight.HiddenRequestedRuntimeTarget
	selection.SelectedRoute = preflight.SelectedRoute
	selection.SelectedRouteReady = preflight.SelectedRouteReady
	selection.RouteErr = preflight.RouteErr
	selection.CanUseManagedSessionRouteForImplicitFallback = preflight.CanUseManagedSessionRouteForImplicitFallback
	canUseManagedImplicitFallback := browserRuntimeCanUseManagedImplicitFallback(selection)
	defaultProfile := strings.TrimSpace(firstNonEmpty(base.Profile, defaultBrowserRuntimeInfo().Profile))
	routeErrIsImplicitHostGate := browserImplicitLegacyHostRouteErrMatchesDefaultRequestError(
		hiddenImplicitHostDefaultBase,
		selection.Requested,
		defaultProfile,
		selection.RouteErr,
	)
	if hiddenImplicitHostDefaultBase &&
		requestedTarget == "" &&
		browserRuntimeActionRequiresExplicitRuntimeTargetForImplicitFallback(action) &&
		!canUseManagedImplicitFallback {
		if selection.RouteErr == nil || selection.HiddenRequestedRuntimeTarget || routeErrIsImplicitHostGate {
			return selection, browserImplicitLegacyHostRuntimeActionRequiresExplicitRuntimeTargetError(action)
		}
		return selection, nil
	}
	selection.UseHiddenImplicitHostDiagnosticsDegrade = browserRuntimeUsesImplicitLegacyHostDiagnosticsDegradePath(
		action,
		hiddenImplicitHostDefaultBase,
		requestedProfile,
		requestedTarget,
		canUseManagedImplicitFallback,
	)
	if selection.UseHiddenImplicitHostDiagnosticsDegrade && (selection.HiddenRequestedRuntimeTarget || routeErrIsImplicitHostGate) {
		selection.RouteErr = nil
	}
	if !selection.UseHiddenImplicitHostDiagnosticsDegrade && !selection.SelectedRouteReady {
		routeSelection := browserResolveRuntimeRouteSelectionPreflight(func() (browserResolvedExecutionRoute, error) {
			return resolveBrowserExecutionRouteForSessionPreview(
				ctx,
				preview,
				params,
				base,
				hiddenImplicitHostDefaultBase,
				backend,
			)
		})
		selection.SelectedRoute = routeSelection.SelectedRoute
		selection.SelectedRouteReady = routeSelection.SelectedRouteReady
		selection.RouteErr = routeSelection.RouteErr
	}
	return selection, nil
}

func browserRuntimeDefaultCandidateRouteForSessionPreview(
	base BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
	backend BrowserBackend,
) BrowserRuntimeInfo {
	switch routed := backend.(type) {
	case browserRuntimeRouterBackend:
		return normalizeBrowserRuntimeInfo(
			routed.sessionExecutionPreview(base, hiddenImplicitHostDefaultBase).DefaultCandidateRoute,
		)
	case *browserRuntimeRouterBackend:
		if routed != nil {
			return normalizeBrowserRuntimeInfo(
				routed.sessionExecutionPreview(base, hiddenImplicitHostDefaultBase).DefaultCandidateRoute,
			)
		}
	}
	return BrowserRuntimeInfo{}
}

func browserRuntimeDefaultCandidateRouteDescriptorForSessionPreview(
	base BrowserRuntimeInfo,
	hiddenImplicitHostDefaultBase bool,
	backend BrowserBackend,
) browserRuntimeRouteDescriptor {
	switch routed := backend.(type) {
	case browserRuntimeRouterBackend:
		return routed.sessionExecutionPreview(base, hiddenImplicitHostDefaultBase).DefaultCandidateDescriptor
	case *browserRuntimeRouterBackend:
		if routed != nil {
			return routed.sessionExecutionPreview(base, hiddenImplicitHostDefaultBase).DefaultCandidateDescriptor
		}
	}
	return browserRuntimeRouteDescriptor{}
}

func resolveBrowserExecutionBackendForRoute(runtimeInfo BrowserRuntimeInfo, backend BrowserBackend) (BrowserBackend, BrowserRuntimeInfo, error) {
	if router, ok := backend.(BrowserRuntimeBackendRouter); ok {
		routedBackend, routedInfo, err := router.ResolveBrowserBackend(runtimeInfo)
		if err != nil {
			return nil, BrowserRuntimeInfo{}, err
		}
		return routedBackend, normalizeBrowserRuntimeInfo(routedInfo), nil
	}
	return backend, runtimeInfo, nil
}

func resolveBrowserRuntimeRouteForSession(ctx context.Context, stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry, sessionRegistry *BrowserSessionRegistry, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, backend BrowserBackend) (BrowserRuntimeInfo, error) {
	sessionExecutionPreview := browserRuntimePreviewSessionSelectionsForExecution(
		ctx,
		stateRegistry,
		sessionRegistry,
		params,
		base,
		hiddenImplicitHostDefaultBase,
		backend,
	)
	if resolver, ok := backend.(browserSessionRuntimeRouteResolver); ok {
		return resolver.ResolveBrowserRuntimeRouteForSession(
			ctx,
			params,
			sessionExecutionPreview.Base,
			sessionExecutionPreview.HiddenImplicitHostDefaultBase,
		)
	}
	runtimeInfo, _, err := browserResolveRuntimeRouteFromSessionPreview(
		params,
		sessionExecutionPreview.SessionSelectionPreview,
		sessionExecutionPreview.Base,
		sessionExecutionPreview.HiddenImplicitHostDefaultBase,
		backend,
	)
	return runtimeInfo, err
}

func resolveBrowserExecutionBackend(params map[string]any, base BrowserRuntimeInfo, backend BrowserBackend) (BrowserBackend, BrowserRuntimeInfo, error) {
	requested, explicitProfile, explicitTarget, err := browserRequestedRuntimeInfoForDefaultRequestBase(params, base, backend)
	if err != nil {
		return nil, BrowserRuntimeInfo{}, err
	}
	if resolver, ok := backend.(browserExecutionRouteResolver); ok {
		route, err := resolver.ResolveBrowserExecutionRoute(requested)
		if err != nil {
			return nil, BrowserRuntimeInfo{}, err
		}
		route.RuntimeInfo = normalizeBrowserRuntimeInfo(route.RuntimeInfo)
		if route.RuntimeInfo.Backend == "" {
			route.RuntimeInfo.Backend = requested.Backend
		}
		if explicitProfile && route.RuntimeInfo.Profile == "" {
			route.RuntimeInfo.Profile = requested.Profile
		}
		if explicitTarget && route.RuntimeInfo.Target == "" {
			route.RuntimeInfo.Target = requested.Target
		}
		return route.Backend, normalizeBrowserRuntimeInfo(route.RuntimeInfo), nil
	}
	runtimeInfo, err := resolveBrowserRuntimeRoute(params, base, backend)
	if err != nil {
		return nil, BrowserRuntimeInfo{}, err
	}
	return resolveBrowserExecutionBackendForRoute(runtimeInfo, backend)
}

func resolveBrowserExecutionRoute(params map[string]any, base BrowserRuntimeInfo, backend BrowserBackend) (browserResolvedExecutionRoute, error) {
	requested, explicitProfile, explicitTarget, err := browserRequestedRuntimeInfoForDefaultRequestBase(params, base, backend)
	if err != nil {
		return browserResolvedExecutionRoute{}, err
	}
	if resolver, ok := backend.(browserExecutionRouteResolver); ok {
		route, err := resolver.ResolveBrowserExecutionRoute(requested)
		if err != nil {
			return browserResolvedExecutionRoute{}, err
		}
		return browserNormalizeResolvedExecutionRoute(route, requested, explicitProfile, explicitTarget), nil
	}
	runtimeInfo, err := resolveBrowserRuntimeRoute(params, base, backend)
	if err != nil {
		return browserResolvedExecutionRoute{}, err
	}
	routedBackend, routedInfo, err := resolveBrowserExecutionBackendForRoute(runtimeInfo, backend)
	if err != nil {
		return browserResolvedExecutionRoute{}, err
	}
	return browserNormalizeResolvedExecutionRoute(browserResolvedExecutionRoute{
		Backend:      routedBackend,
		RuntimeInfo:  routedInfo,
		Capabilities: browserCapabilitiesForConcreteBackend(routedBackend),
	}, runtimeInfo, false, false), nil
}

func browserNormalizeResolvedExecutionRoute(route browserResolvedExecutionRoute, requested BrowserRuntimeInfo, explicitProfile bool, explicitTarget bool) browserResolvedExecutionRoute {
	route.RuntimeInfo = normalizeBrowserRuntimeInfo(route.RuntimeInfo)
	if route.RuntimeInfo.Backend == "" {
		route.RuntimeInfo.Backend = requested.Backend
	}
	if explicitProfile && route.RuntimeInfo.Profile == "" {
		route.RuntimeInfo.Profile = requested.Profile
	}
	if explicitTarget && route.RuntimeInfo.Target == "" {
		route.RuntimeInfo.Target = requested.Target
	}
	route.RuntimeInfo = normalizeBrowserRuntimeInfo(route.RuntimeInfo)
	if route.Capabilities == (BrowserCapabilities{}) {
		route.Capabilities = browserCapabilitiesForConcreteBackend(route.Backend)
	}
	if metadata := browserDoctorRouteMetadataForBackend(route.Backend); metadata != (browserDoctorRouteMetadata{}) {
		if strings.TrimSpace(route.Source) == "" {
			route.Source = strings.TrimSpace(metadata.Source)
		}
		if strings.TrimSpace(route.Endpoint) == "" {
			route.Endpoint = strings.TrimSpace(metadata.Endpoint)
		}
	}
	return route
}

func resolveBrowserExecutionRouteForSessionPreview(ctx context.Context, preview browserRuntimeSessionSelectionPreview, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, backend BrowserBackend) (browserResolvedExecutionRoute, error) {
	if resolver, ok := backend.(browserSessionExecutionRouteResolver); ok {
		route, err := resolver.ResolveBrowserExecutionRouteForSession(ctx, params, base, hiddenImplicitHostDefaultBase)
		if err != nil {
			return browserResolvedExecutionRoute{}, err
		}
		return browserNormalizeResolvedExecutionRoute(route, base, false, false), nil
	}
	requested, explicitProfile, explicitTarget, err := browserRequestedRuntimeInfo(preview.Applied, base)
	if err != nil {
		return browserResolvedExecutionRoute{}, err
	}
	if resolver, ok := backend.(browserExecutionRouteResolver); ok {
		route, err := resolver.ResolveBrowserExecutionRoute(requested)
		if err != nil {
			return browserResolvedExecutionRoute{}, err
		}
		return browserNormalizeResolvedExecutionRoute(route, requested, explicitProfile, explicitTarget), nil
	}
	if resolver, ok := backend.(browserSessionRuntimeRouteResolver); ok {
		runtimeInfo, err := resolver.ResolveBrowserRuntimeRouteForSession(ctx, params, base, hiddenImplicitHostDefaultBase)
		if err != nil {
			return browserResolvedExecutionRoute{}, err
		}
		routedBackend, routedInfo, err := resolveBrowserExecutionBackendForRoute(runtimeInfo, backend)
		if err != nil {
			return browserResolvedExecutionRoute{}, err
		}
		return browserNormalizeResolvedExecutionRoute(browserResolvedExecutionRoute{
			Backend:      routedBackend,
			RuntimeInfo:  routedInfo,
			Capabilities: browserCapabilitiesForConcreteBackend(routedBackend),
		}, runtimeInfo, false, false), nil
	}
	return browserResolveExecutionRouteFromSessionPreview(params, preview, base, hiddenImplicitHostDefaultBase, backend)
}

func resolveBrowserExecutionBackendForSession(ctx context.Context, stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry, sessionRegistry *BrowserSessionRegistry, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, backend BrowserBackend) (BrowserBackend, BrowserRuntimeInfo, error) {
	sessionExecutionPreview := browserRuntimePreviewSessionSelectionsForExecution(
		ctx,
		stateRegistry,
		sessionRegistry,
		params,
		base,
		hiddenImplicitHostDefaultBase,
		backend,
	)
	route, err := resolveBrowserExecutionRouteForSessionPreview(
		ctx,
		sessionExecutionPreview.SessionSelectionPreview,
		params,
		sessionExecutionPreview.Base,
		sessionExecutionPreview.HiddenImplicitHostDefaultBase,
		backend,
	)
	if err != nil {
		return nil, BrowserRuntimeInfo{}, err
	}
	return route.Backend, normalizeBrowserRuntimeInfo(route.RuntimeInfo), nil
}

func resolveBrowserExecutionRouteForSession(ctx context.Context, stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry, sessionRegistry *BrowserSessionRegistry, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, backend BrowserBackend) (browserResolvedExecutionRoute, error) {
	sessionExecutionPreview := browserRuntimePreviewSessionSelectionsForExecution(
		ctx,
		stateRegistry,
		sessionRegistry,
		params,
		base,
		hiddenImplicitHostDefaultBase,
		backend,
	)
	return resolveBrowserExecutionRouteForSessionPreview(
		ctx,
		sessionExecutionPreview.SessionSelectionPreview,
		params,
		sessionExecutionPreview.Base,
		sessionExecutionPreview.HiddenImplicitHostDefaultBase,
		backend,
	)
}

type browserResolvedExecutionRoute struct {
	Backend      BrowserBackend
	RuntimeInfo  BrowserRuntimeInfo
	Capabilities BrowserCapabilities
	Source       string
	Endpoint     string

	hiddenImplicitHostDefaultBase bool
	sessionRegistry               *BrowserSessionRegistry
	sessionStateRegistry          agentxbrowserruntime.SharedSessionBrowserStateRegistry
	watchManagerProvider          agentxbrowserruntime.SharedSessionBrowserObserverManager
	params                        map[string]any
	maxChars                      int
}

type browserManagedRouteExecutionArgs struct {
	URL              string
	BrowserApp       string
	WaitMs           int
	TabIndex         int
	Force            bool
	FinalURL         string
	Title            string
	ResultBrowserApp string
	ResultBackend    string
}

func resolveBrowserManagedRouteForSession(ctx context.Context, stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry, sessionRegistry *BrowserSessionRegistry, watchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, backend BrowserBackend, maxChars int) (browserResolvedExecutionRoute, error) {
	route, err := resolveBrowserExecutionRouteForSession(ctx, stateRegistry, sessionRegistry, params, base, hiddenImplicitHostDefaultBase, backend)
	if err != nil {
		return browserResolvedExecutionRoute{}, err
	}
	return route.withManagedRuntime(sessionRegistry, stateRegistry, watchManagerProvider, params, maxChars, hiddenImplicitHostDefaultBase), nil
}

func (route browserResolvedExecutionRoute) withManagedRuntime(sessionRegistry *BrowserSessionRegistry, sessionStateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry, watchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager, params map[string]any, maxChars int, hiddenImplicitHostDefaultBase bool) browserResolvedExecutionRoute {
	route.RuntimeInfo = normalizeBrowserRuntimeInfo(route.RuntimeInfo)
	route.hiddenImplicitHostDefaultBase = hiddenImplicitHostDefaultBase
	route.sessionRegistry = sessionRegistry
	route.sessionStateRegistry = sessionStateRegistry
	route.watchManagerProvider = watchManagerProvider
	route.params = params
	route.maxChars = maxChars
	return route
}

func browserResolvedExecutionRouteExecuteManaged[T any](
	ctx context.Context,
	route browserResolvedExecutionRoute,
	invoke func(BrowserBackend) (T, error),
	args browserManagedRouteExecutionArgs,
	fromPolicy func(browserManagedResolverFailurePolicyResult) T,
) (browserManagedResolverExecutionResult[T], error) {
	return browserManagedResolverExecute(
		ctx,
		func() (T, error) {
			return invoke(route.Backend)
		},
		route.managedResolverFailurePolicyArgs(args),
		fromPolicy,
	)
}

func (route browserResolvedExecutionRoute) managedFinalURL(ctx context.Context, browserApp string, target browserToolTarget, reqURL string) string {
	return firstNonEmpty(strings.TrimSpace(reqURL), browserResolvedTargetURL(ctx, route.sessionRegistry, route.RuntimeInfo, route.hiddenImplicitHostDefaultBase, browserApp, target))
}

func (route browserResolvedExecutionRoute) managedResolverFailurePolicyArgs(args browserManagedRouteExecutionArgs) browserManagedResolverFailurePolicyArgs {
	return browserManagedResolverFailurePolicyArgs{
		Route: route,
		Request: browserManagedRouteExecutionArgs{
			URL:              strings.TrimSpace(args.URL),
			BrowserApp:       strings.TrimSpace(args.BrowserApp),
			WaitMs:           args.WaitMs,
			TabIndex:         args.TabIndex,
			Force:            args.Force,
			FinalURL:         strings.TrimSpace(args.FinalURL),
			Title:            strings.TrimSpace(args.Title),
			ResultBrowserApp: strings.TrimSpace(args.ResultBrowserApp),
			ResultBackend:    strings.TrimSpace(args.ResultBackend),
		},
	}
}

func (route browserResolvedExecutionRoute) managedResolverRecovery(
	ctx context.Context,
	req browserManagedRouteExecutionArgs,
	outcome *agentxbrowserruntime.BrowserElementResolverOutcome,
	finalURL string,
	title string,
	resultBrowserApp string,
	resultBackend string,
	note string,
) browserManagedResolverRecoveryResult {
	recovery := browserManagedResolverRecoveryResult{
		FinalURL:   strings.TrimSpace(finalURL),
		Title:      strings.TrimSpace(title),
		BrowserApp: strings.TrimSpace(resultBrowserApp),
		Backend:    strings.TrimSpace(resultBackend),
		Note:       strings.TrimSpace(note),
	}
	snapshot, recovered := route.managedResolverSnapshotRecovery(ctx, req, outcome)
	if recovered {
		recovery.Snapshot = snapshot
		recovery.SnapshotRecovered = true
		recovery.SnapshotText, recovery.SnapshotTruncated = route.managedResolverSnapshotText(snapshot)
		recovery.FinalURL = firstNonEmpty(recovery.FinalURL, strings.TrimSpace(snapshot.FinalURL))
		recovery.Title = firstNonEmpty(recovery.Title, strings.TrimSpace(snapshot.Title))
		recovery.BrowserApp = firstNonEmpty(recovery.BrowserApp, strings.TrimSpace(snapshot.BrowserApp))
		recovery.Backend = firstNonEmpty(recovery.Backend, strings.TrimSpace(snapshot.Backend))
	}
	refresh, attempted, err := route.managedResolverRefreshRecovery(ctx, req, outcome)
	if attempted {
		recovery.Note = browserManagedResolverRefreshRecoveryNote(recovery.Note, refresh, err)
		recovery.InvalidateSessionTargets = refresh.InvalidateSessionTargets
	}
	return recovery
}

func (route browserResolvedExecutionRoute) managedResolverSnapshotRecoveryRequest(req browserManagedRouteExecutionArgs) BrowserSnapshotRequest {
	requestMaxChars := firstInt(route.params, "max_chars")
	if requestMaxChars <= 0 || requestMaxChars > route.maxChars {
		requestMaxChars = route.maxChars
	}
	requestMaxElements := firstInt(route.params, "max_elements")
	if requestMaxElements <= 0 {
		requestMaxElements = 16
	}
	if requestMaxElements > 24 {
		requestMaxElements = 24
	}
	return BrowserSnapshotRequest{
		URL:         strings.TrimSpace(req.URL),
		BrowserApp:  strings.TrimSpace(req.BrowserApp),
		WaitMs:      req.WaitMs,
		MaxChars:    requestMaxChars,
		MaxElements: requestMaxElements,
		TabIndex:    req.TabIndex,
		Refs:        "role",
		Interactive: true,
	}
}

func (route browserResolvedExecutionRoute) managedResolverSnapshotRecovery(ctx context.Context, req browserManagedRouteExecutionArgs, outcome *agentxbrowserruntime.BrowserElementResolverOutcome) (BrowserSnapshotResult, bool) {
	if !route.Capabilities.Snapshot {
		return BrowserSnapshotResult{}, false
	}
	if strings.EqualFold(strings.TrimSpace(route.RuntimeInfo.Target), "host") {
		return BrowserSnapshotResult{}, false
	}
	if browserResolverRecoveryAction(outcome) != "browser action=snapshot" {
		return BrowserSnapshotResult{}, false
	}
	result, err := route.Backend.Snapshot(ctx, route.managedResolverSnapshotRecoveryRequest(req))
	if err != nil {
		return BrowserSnapshotResult{}, false
	}
	result.Elements = browserNormalizeSnapshotElements(result.Elements, firstNonEmpty(strings.TrimSpace(result.FinalURL), strings.TrimSpace(req.URL)), strings.TrimSpace(result.Title))
	return result, true
}

func (route browserResolvedExecutionRoute) managedResolverSnapshotText(snapshot BrowserSnapshotResult) (string, bool) {
	requestMaxChars := firstInt(route.params, "max_chars")
	if requestMaxChars <= 0 || requestMaxChars > route.maxChars {
		requestMaxChars = route.maxChars
	}
	value := strings.TrimSpace(snapshot.Snapshot)
	truncated := snapshot.Truncated
	if trimmed, changed := trimToMaxChars(value, requestMaxChars); changed {
		value = trimmed
		truncated = true
	}
	return value, truncated
}

func (route browserResolvedExecutionRoute) managedResolverRefreshRecovery(ctx context.Context, req browserManagedRouteExecutionArgs, outcome *agentxbrowserruntime.BrowserElementResolverOutcome) (browserRuntimePrepareResult, bool, error) {
	if browserResolverRecoveryAction(outcome) != "browser action=refresh" {
		return browserRuntimePrepareResult{}, false, nil
	}
	if strings.EqualFold(strings.TrimSpace(route.RuntimeInfo.Target), "host") {
		return browserRuntimePrepareResult{}, false, nil
	}
	control, ok := route.Backend.(BrowserRuntimeControlBackend)
	if !ok || control == nil {
		return browserRuntimePrepareResult{}, false, nil
	}
	selectedRoute := browserRuntimeRouteDescriptorPtr(route.RuntimeInfo)
	binding := browserRuntimeBuildSessionBinding(ctx, route.sessionRegistry, nil, route.sessionStateRegistry, selectedRoute, nil, nil)
	result, err := browserRuntimeRestartProfile(ctx, route.sessionStateRegistry, control, route.RuntimeInfo.Profile, route.RuntimeInfo, binding, req.Force)
	browserRuntimeApplyPrepareResult(ctx, &browserRuntimePayload{SelectedRoute: selectedRoute}, route.watchManagerProvider, route.sessionRegistry, route.sessionStateRegistry, route.RuntimeInfo, result)
	return result, true, err
}

func browserRuntimeBackendName(info BrowserRuntimeInfo) string {
	info = normalizeBrowserRuntimeInfo(info)
	if info.Backend != "" {
		return info.Backend
	}
	return "browser"
}

func browserRuntimeApplySessionProfileSelection(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry, params map[string]any, hiddenImplicitHostDefaultBase bool) map[string]any {
	if registry == nil {
		return params
	}
	concrete, ok := registry.(*BrowserSessionStateRegistry)
	if !ok || concrete == nil {
		return params
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return params
	}
	if strings.TrimSpace(firstString(params, "profile", "browser_profile", "runtime_profile")) != "" {
		return params
	}
	target := browserRequestedRuntimeTarget(params)
	selection, ok := concrete.SelectedBrowserProfile(sessionID, target)
	if !ok || strings.TrimSpace(selection.Profile) == "" {
		return params
	}
	if hiddenImplicitHostDefaultBase {
		selectedTarget := strings.ToLower(strings.TrimSpace(selection.RuntimeTarget))
		if selectedTarget == "" || selectedTarget == "host" {
			return params
		}
	}
	cloned := make(map[string]any, len(params)+3)
	for key, value := range params {
		cloned[key] = value
	}
	if !browserHasExplicitRuntimeTarget(cloned) && strings.TrimSpace(selection.RuntimeTarget) != "" {
		cloned["runtime_target"] = selection.RuntimeTarget
	}
	cloned["profile"] = selection.Profile
	if strings.TrimSpace(firstString(cloned, "browser", "browser_app", "app")) == "" && strings.TrimSpace(selection.BrowserApp) != "" {
		cloned["browser_app"] = selection.BrowserApp
	}
	return cloned
}

func browserRuntimeApplySessionTargetRouteSelection(ctx context.Context, registry *BrowserSessionRegistry, params map[string]any, hiddenImplicitHostDefaultBase bool) map[string]any {
	if registry == nil {
		return params
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return params
	}
	if browserHasExplicitRuntimeTarget(params) {
		return params
	}
	if strings.TrimSpace(firstString(params, "target")) != "" || firstInt(params, "tab_index") > 0 {
		return params
	}
	target, ok := agentxbrowserruntime.ResolveSharedSessionBrowserCurrentTarget(
		registry,
		sessionID,
		agentxbrowserruntime.BrowserSessionRoute{},
		false,
	)
	if !ok {
		return params
	}
	runtimeTarget := strings.TrimSpace(target.Target)
	backend := strings.TrimSpace(target.Backend)
	profile := strings.TrimSpace(target.Profile)
	if runtimeTarget == "" && backend == "" && profile == "" {
		return params
	}
	if hiddenImplicitHostDefaultBase {
		if runtimeTarget == "" || strings.EqualFold(runtimeTarget, "host") {
			return params
		}
	}
	cloned := make(map[string]any, len(params)+4)
	for key, value := range params {
		cloned[key] = value
	}
	if runtimeTarget != "" {
		cloned["runtime_target"] = runtimeTarget
	}
	if strings.TrimSpace(firstString(cloned, "profile", "browser_profile", "runtime_profile")) == "" && profile != "" {
		cloned["profile"] = profile
	}
	if strings.TrimSpace(firstString(cloned, "browser", "browser_app", "app")) == "" && strings.TrimSpace(target.BrowserApp) != "" {
		cloned["browser_app"] = target.BrowserApp
	}
	return cloned
}

func browserRuntimeApplyPageBoundElementRouteSelection(ctx context.Context, registry *BrowserSessionRegistry, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool) map[string]any {
	if registry == nil || !hiddenImplicitHostDefaultBase {
		return params
	}
	if browserHasExplicitRuntimeTarget(params) {
		return params
	}
	if strings.TrimSpace(firstString(params, "target")) != "" || firstInt(params, "tab_index", "index") > 0 {
		return params
	}
	if strings.TrimSpace(firstString(params, "url")) != "" {
		return params
	}
	elementTarget, ok := browserPageBoundElementTargetForRouteSelection(params)
	if !ok {
		return params
	}
	tracked, ok := browserResolveTrackedTargetForElementBinding(
		ctx,
		registry,
		base,
		hiddenImplicitHostDefaultBase,
		strings.TrimSpace(firstString(params, "browser", "browser_app", "app")),
		browserToolTarget{},
	)
	if !ok || !browserElementRefMatchesCurrentPage(elementTarget.Payload, tracked.URL, tracked.Title) {
		return params
	}
	return browserRuntimeApplyTrackedTargetRouteSelection(params, base, hiddenImplicitHostDefaultBase, tracked)
}

func browserRuntimeApplyTargetHandleRouteSelection(ctx context.Context, registry *BrowserSessionRegistry, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool) map[string]any {
	if registry == nil {
		return params
	}
	base = normalizeBrowserRuntimeInfo(base)
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return params
	}
	rawTarget := strings.TrimSpace(firstString(params, "target"))
	tabIndex := firstInt(params, "tab_index", "index")
	if rawTarget == "" && tabIndex <= 0 {
		return params
	}
	target := browserToolTarget{}
	if rawTarget != "" {
		parsed, err := parseBrowserToolTarget(rawTarget)
		if err != nil {
			return params
		}
		target = parsed
	}
	if tabIndex > 0 && target.TabIndex > 0 && target.TabIndex != tabIndex {
		return params
	}
	if tabIndex > 0 {
		target.TabIndex = tabIndex
		target.Value = fmt.Sprintf("tab:%d", tabIndex)
		target.TargetID = ""
	}
	var (
		tracked BrowserSessionTarget
		ok      bool
	)
	switch {
	case target.Value == "current" && strings.TrimSpace(target.TargetID) == "" && target.TabIndex <= 0:
		tracked, ok = agentxbrowserruntime.ResolveSharedSessionBrowserCurrentTarget(
			registry,
			sessionID,
			agentxbrowserruntime.BrowserSessionRoute{},
			false,
		)
	case strings.TrimSpace(target.TargetID) != "":
		tracked, ok = agentxbrowserruntime.ResolveSharedSessionBrowserTarget(
			registry,
			sessionID,
			agentxbrowserruntime.BrowserSessionRoute{},
			target.TargetID,
			0,
			false,
		)
	case target.TabIndex > 0:
		tracked, ok = agentxbrowserruntime.ResolveSharedSessionBrowserCurrentTarget(
			registry,
			sessionID,
			agentxbrowserruntime.BrowserSessionRoute{},
			false,
		)
	default:
		return params
	}
	if !ok {
		return params
	}
	return browserRuntimeApplyTrackedTargetRouteSelection(params, base, hiddenImplicitHostDefaultBase, tracked)
}

func browserRuntimeApplyTrackedTargetRouteSelection(params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, tracked BrowserSessionTarget) map[string]any {
	explicitProfile := strings.TrimSpace(firstString(params, "profile", "browser_profile", "runtime_profile")) != ""
	explicitTarget := browserHasExplicitRuntimeTarget(params)
	explicitBrowserApp := strings.TrimSpace(firstString(params, "browser", "browser_app", "app")) != ""
	if explicitProfile && explicitTarget && explicitBrowserApp {
		return params
	}
	if hiddenImplicitHostDefaultBase {
		trackedTarget := strings.ToLower(strings.TrimSpace(tracked.Target))
		if trackedTarget == "" || trackedTarget == "host" {
			return params
		}
	}
	cloned := make(map[string]any, len(params)+3)
	for key, value := range params {
		cloned[key] = value
	}
	injectedTarget := false
	if !explicitTarget &&
		strings.TrimSpace(tracked.Target) != "" &&
		!(base.Target == "" && strings.EqualFold(strings.TrimSpace(tracked.Target), "host")) &&
		!strings.EqualFold(strings.TrimSpace(tracked.Target), base.Target) {
		cloned["runtime_target"] = tracked.Target
		injectedTarget = true
	}
	if !explicitProfile &&
		!strings.EqualFold(strings.TrimSpace(tracked.Target), "host") &&
		strings.TrimSpace(tracked.Profile) != "" &&
		(injectedTarget || !strings.EqualFold(strings.TrimSpace(tracked.Profile), base.Profile)) {
		cloned["profile"] = tracked.Profile
	}
	if !explicitBrowserApp && strings.TrimSpace(tracked.BrowserApp) != "" {
		cloned["browser_app"] = tracked.BrowserApp
	}
	return cloned
}

func browserRuntimeApplySessionSelections(ctx context.Context, stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry, sessionRegistry *BrowserSessionRegistry, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool) map[string]any {
	params = browserRuntimeApplyTargetHandleRouteSelection(ctx, sessionRegistry, params, base, hiddenImplicitHostDefaultBase)
	params = browserRuntimeApplyPageBoundElementRouteSelection(ctx, sessionRegistry, params, base, hiddenImplicitHostDefaultBase)
	params = browserRuntimeApplySessionProfileSelection(ctx, stateRegistry, params, hiddenImplicitHostDefaultBase)
	return browserRuntimeApplySessionTargetRouteSelection(ctx, sessionRegistry, params, hiddenImplicitHostDefaultBase)
}

func browserRuntimePreviewSessionSelections(ctx context.Context, stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry, sessionRegistry *BrowserSessionRegistry, params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool) browserRuntimeSessionSelectionPreview {
	applied := browserRuntimeApplySessionSelections(ctx, stateRegistry, sessionRegistry, params, base, hiddenImplicitHostDefaultBase)
	return browserRuntimeSessionSelectionPreview{
		Applied:                applied,
		RequestedRuntimeTarget: browserRequestedRuntimeTarget(applied),
		RequestedBrowserApp:    strings.TrimSpace(firstString(applied, "browser_app", "browser")),
	}
}
