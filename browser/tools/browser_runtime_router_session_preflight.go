package tools

import (
	"context"
	"strings"
)

type browserRuntimeRouterSessionPreflightRequest struct {
	Requested                     BrowserRuntimeInfo
	DefaultCandidateRoute         BrowserRuntimeInfo
	DefaultCandidateDescriptor    browserRuntimeRouteDescriptor
	Preview                       browserRuntimeSessionSelectionPreview
	ExplicitRuntimeTarget         bool
	HiddenImplicitHostDefaultBase bool
	HiddenRequestedRuntimeTarget  bool
}

type browserRuntimeRouterSessionPreflight struct {
	browserRuntimeRouterSessionPreflightRequest
	Route                                        browserResolvedExecutionRoute
	CanUseManagedSessionRouteForImplicitFallback bool
}

func (b browserRuntimeRouterBackend) buildSessionPreflightRequest(
	ctx context.Context,
	params map[string]any,
	base BrowserRuntimeInfo,
	defaultCandidateRoute BrowserRuntimeInfo,
	defaultCandidateDescriptor browserRuntimeRouteDescriptor,
	hiddenImplicitHostDefaultBase bool,
) (browserRuntimeRouterSessionPreflightRequest, error) {
	requested, preview, explicitRuntimeTarget, err := b.directRequestRuntimeInfo(
		ctx,
		params,
		base,
		hiddenImplicitHostDefaultBase,
	)
	if err != nil {
		return browserRuntimeRouterSessionPreflightRequest{}, err
	}
	return browserRuntimeRouterSessionPreflightRequest{
		Requested:                     requested,
		DefaultCandidateRoute:         normalizeBrowserRuntimeInfo(defaultCandidateRoute),
		DefaultCandidateDescriptor:    defaultCandidateDescriptor,
		Preview:                       preview,
		ExplicitRuntimeTarget:         explicitRuntimeTarget,
		HiddenImplicitHostDefaultBase: hiddenImplicitHostDefaultBase,
		HiddenRequestedRuntimeTarget: hiddenImplicitHostDefaultBase &&
			!explicitRuntimeTarget &&
			strings.TrimSpace(preview.RequestedRuntimeTarget) != "" &&
			strings.TrimSpace(requested.Target) == "",
	}, nil
}

func (b browserRuntimeRouterBackend) resolveSessionPreflight(
	request browserRuntimeRouterSessionPreflightRequest,
) (browserRuntimeRouterSessionPreflight, error) {
	if request.HiddenRequestedRuntimeTarget {
		return browserRuntimeRouterSessionPreflight{}, browserImplicitLegacyHostDefaultRequestError(
			request.HiddenImplicitHostDefaultBase,
			request.Requested,
			firstNonEmpty(b.hostRuntime.Profile, defaultBrowserRuntimeInfo().Profile),
		)
	}
	selection := browserResolveRuntimeRouteSelectionPreflight(func() (browserResolvedExecutionRoute, error) {
		return b.resolveBrowserRoute(request.Requested)
	})
	if selection.RouteErr != nil {
		return browserRuntimeRouterSessionPreflight{}, selection.RouteErr
	}
	return browserRuntimeRouterSessionPreflight{
		browserRuntimeRouterSessionPreflightRequest: request,
		Route: selection.SelectedRoute,
		CanUseManagedSessionRouteForImplicitFallback: selection.CanUseManagedSessionRouteForImplicitFallback,
	}, nil
}
