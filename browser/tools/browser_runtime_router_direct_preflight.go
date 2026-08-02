package tools

import (
	"context"
	"strings"
)

type browserRuntimeRouterDirectPreflightRequest struct {
	browserRuntimeRouterSessionPreflightRequest
	Target browserToolTarget
}

type browserRuntimeRouterDirectPreflight struct {
	browserRuntimeRouterDirectPreflightRequest
	Lane     BrowserExecutionLane
	Fallback *browserRuntimeRouterURLFallback
}

type browserRuntimeRouterDirectPreflightArgs struct {
	Params           map[string]any
	Base             BrowserRuntimeInfo
	Target           browserToolTarget
	RequestURL       string
	LocalURLFallback bool
	FallbackCheck    func(browserRuntimeRouterDirectPreflightRequest) error
}

type browserRuntimeRouterURLFallback struct {
	Reason                string
	OriginalBackend       string
	OriginalRuntimeTarget string
	SelectedBackend       string
	SelectedRuntimeTarget string
}

func (b browserRuntimeRouterBackend) buildDirectPreflightRequest(
	ctx context.Context,
	args browserRuntimeRouterDirectPreflightArgs,
) (browserRuntimeRouterDirectPreflightRequest, error) {
	executionPreview := b.executionPreview()
	request, err := b.buildSessionPreflightRequest(
		ctx,
		args.Params,
		args.Base,
		executionPreview.DefaultCandidateRoute,
		executionPreview.DefaultCandidateDescriptor,
		executionPreview.HiddenImplicitHostDefaultBase,
	)
	if err != nil {
		return browserRuntimeRouterDirectPreflightRequest{}, err
	}
	return browserRuntimeRouterDirectPreflightRequest{
		browserRuntimeRouterSessionPreflightRequest: request,
		Target: args.Target,
	}, nil
}

func (b browserRuntimeRouterBackend) resolveDirectPreflight(
	ctx context.Context,
	args browserRuntimeRouterDirectPreflightArgs,
) (browserRuntimeRouterDirectPreflight, error) {
	request, err := b.buildDirectPreflightRequest(ctx, args)
	if err != nil {
		return browserRuntimeRouterDirectPreflight{}, err
	}
	var fallbackErr error
	if args.FallbackCheck != nil {
		fallbackErr = args.FallbackCheck(request)
	}
	sessionPreflight, err := b.resolveSessionPreflight(request.browserRuntimeRouterSessionPreflightRequest)
	if err != nil {
		if fallbackErr != nil && b.shouldPreferDirectFallbackError(request.browserRuntimeRouterSessionPreflightRequest, err) {
			return browserRuntimeRouterDirectPreflight{}, fallbackErr
		}
		return browserRuntimeRouterDirectPreflight{}, err
	}
	route := sessionPreflight.Route
	var policyFallback *browserRuntimeRouterURLFallback
	if args.LocalURLFallback {
		policyFallback, err = b.localURLFallbackForRemoteRoute(request, route, args.RequestURL)
	}
	if err != nil {
		return browserRuntimeRouterDirectPreflight{}, err
	}
	if policyFallback != nil {
		hostRoute, err := b.resolveTargetBrowserRoute(BrowserRuntimeInfo{
			Profile: firstNonEmpty(strings.TrimSpace(b.hostRuntime.Profile), defaultBrowserRuntimeInfo().Profile),
			Target:  "host",
		}, "local browser route is unavailable")
		if err != nil {
			return browserRuntimeRouterDirectPreflight{}, err
		}
		route = hostRoute
		policyFallback.SelectedBackend = strings.TrimSpace(hostRoute.RuntimeInfo.Backend)
		policyFallback.SelectedRuntimeTarget = strings.TrimSpace(hostRoute.RuntimeInfo.Target)
	}
	if fallbackErr != nil && strings.EqualFold(sessionPreflight.Route.RuntimeInfo.Target, "host") {
		return browserRuntimeRouterDirectPreflight{}, fallbackErr
	}
	return browserRuntimeRouterDirectPreflight{
		browserRuntimeRouterDirectPreflightRequest: request,
		Lane:     browserRuntimeRouterExecutionLane(route),
		Fallback: policyFallback,
	}, nil
}

func (b browserRuntimeRouterBackend) shouldPreferDirectFallbackError(request browserRuntimeRouterSessionPreflightRequest, err error) bool {
	if err == nil {
		return false
	}
	if request.HiddenRequestedRuntimeTarget {
		return true
	}
	return browserImplicitLegacyHostRouteErrMatchesDefaultRequestError(
		request.HiddenImplicitHostDefaultBase,
		request.Requested,
		firstNonEmpty(b.hostRuntime.Profile, defaultBrowserRuntimeInfo().Profile),
		err,
	)
}

func (b browserRuntimeRouterBackend) localURLFallbackForRemoteRoute(
	request browserRuntimeRouterDirectPreflightRequest,
	route browserResolvedExecutionRoute,
	requestURL string,
) (*browserRuntimeRouterURLFallback, error) {
	if request.ExplicitRuntimeTarget ||
		strings.TrimSpace(request.Target.TargetID) != "" ||
		request.Target.TabIndex > 0 ||
		!browserURLLooksPrivateOrLocal(requestURL) ||
		!browserBackendRemoteTargetURLGuardEnabled(route.Backend) {
		return nil, nil
	}
	info := normalizeBrowserRuntimeInfo(route.RuntimeInfo)
	return &browserRuntimeRouterURLFallback{
		Reason:                browserRouteFallbackRemoteLocalURLReason,
		OriginalBackend:       strings.TrimSpace(info.Backend),
		OriginalRuntimeTarget: strings.TrimSpace(info.Target),
	}, nil
}

func browserAnnotateURLFallbackResult[T any](result T, fallback *browserRuntimeRouterURLFallback) T {
	if fallback == nil {
		return result
	}
	note := browserRuntimeRouterURLFallbackNote(fallback)
	switch value := any(result).(type) {
	case BrowserOpenResult:
		value.Note = appendBrowserResultNote(value.Note, note)
		return any(value).(T)
	case BrowserNavigateResult:
		value.Note = appendBrowserResultNote(value.Note, note)
		return any(value).(T)
	default:
		return result
	}
}

func browserRuntimeRouterURLFallbackNote(fallback *browserRuntimeRouterURLFallback) string {
	if fallback == nil {
		return ""
	}
	return strings.Join([]string{
		"route_fallback_reason=" + strings.TrimSpace(fallback.Reason),
		"original_backend=" + strings.TrimSpace(fallback.OriginalBackend),
		"original_runtime_target=" + strings.TrimSpace(fallback.OriginalRuntimeTarget),
		"selected_backend=" + strings.TrimSpace(fallback.SelectedBackend),
		"selected_runtime_target=" + strings.TrimSpace(fallback.SelectedRuntimeTarget),
	}, " ")
}

func appendBrowserResultNote(current string, addition string) string {
	current = strings.TrimSpace(current)
	addition = strings.TrimSpace(addition)
	if current == "" {
		return addition
	}
	if addition == "" {
		return current
	}
	return current + "; " + addition
}
