package tools

import "strings"

type browserRuntimeRouteSelectionPreflight struct {
	SelectedRoute                                browserResolvedExecutionRoute
	SelectedRouteReady                           bool
	RouteErr                                     error
	CanUseManagedSessionRouteForImplicitFallback bool
}

func browserResolveRuntimeRouteSelectionPreflight(
	resolveRoute func() (browserResolvedExecutionRoute, error),
) browserRuntimeRouteSelectionPreflight {
	preflight := browserRuntimeRouteSelectionPreflight{}
	preflight.SelectedRoute, preflight.RouteErr = resolveRoute()
	if preflight.RouteErr == nil {
		preflight.SelectedRouteReady = true
		targetRoute := strings.TrimSpace(preflight.SelectedRoute.RuntimeInfo.Target)
		preflight.CanUseManagedSessionRouteForImplicitFallback =
			targetRoute != "" &&
				!strings.EqualFold(targetRoute, "host")
	}
	return preflight
}
