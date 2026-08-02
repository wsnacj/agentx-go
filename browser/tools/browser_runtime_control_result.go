package tools

import "strings"

type browserRuntimeActionDispatchResultPostProcess struct {
	ConfiguredInfo                BrowserRuntimeInfo
	ResolutionDefaultRoute        BrowserRuntimeInfo
	HiddenImplicitHostDefaultBase bool
	DiagnosticsPreview            browserRuntimeDiagnosticsPreview
	UseDiagnosticsPreview         bool
	RouteErr                      error
	IncludeRoutes                 bool
}

type browserRuntimeActionDispatchRouteResultProjection struct {
	RouteResolution    *browserRuntimeRouteResolution
	ConfiguredProfiles []string
	Routes             []browserRuntimeRouteStatus
	ApplyRoutes        bool
	HideSelectedRoute  bool
}

func browserRuntimeProjectActionDispatchRouteResult(
	ctx browserRegistrationContext,
	payload browserRuntimePayload,
	options browserRuntimeActionDispatchResultPostProcess,
) browserRuntimeActionDispatchRouteResultProjection {
	configuredProfiles := browserRuntimeConfiguredProfilesProjectionForPayload(payload, options.ConfiguredInfo)
	resolutionDefaultRoute := browserRuntimeRouteResolutionDefaultRouteForProjection(ctx, payload, options)
	projection := browserRuntimeActionDispatchRouteResultProjection{
		RouteResolution: browserRuntimeRouteResolutionPtr(
			payload,
			resolutionDefaultRoute,
			options.HiddenImplicitHostDefaultBase,
		),
		ConfiguredProfiles: configuredProfiles.Profiles,
		HideSelectedRoute: browserRuntimeShouldHideImplicitLegacyHostRouteResolution(
			payload,
			options.HiddenImplicitHostDefaultBase,
		),
	}
	if options.IncludeRoutes || payload.Status != "ok" || options.RouteErr != nil {
		if options.UseDiagnosticsPreview {
			projection.Routes = browserRuntimeRouteMatrixWithPreview(
				ctx,
				options.DiagnosticsPreview,
				projection.ConfiguredProfiles,
			)
		} else {
			projection.Routes = browserRuntimeRouteMatrix(ctx, projection.ConfiguredProfiles)
		}
		projection.ApplyRoutes = true
	}
	return projection
}

func browserRuntimeRouteResolutionDefaultRouteForProjection(
	ctx browserRegistrationContext,
	payload browserRuntimePayload,
	options browserRuntimeActionDispatchResultPostProcess,
) BrowserRuntimeInfo {
	defaultRoute := normalizeBrowserRuntimeInfo(options.ResolutionDefaultRoute)
	if !options.UseDiagnosticsPreview {
		return defaultRoute
	}
	if strings.TrimSpace(payload.RequestedProfile) != "" || strings.TrimSpace(payload.RequestedRuntimeTarget) != "" {
		return defaultRoute
	}
	if !browserRuntimeUsesDoctorRouteInspectionSummary(payload.Action) {
		return defaultRoute
	}
	route := browserRuntimeDoctorRouteSummary(&payload, options.DiagnosticsPreview)
	if route == nil {
		return defaultRoute
	}
	if !browserRuntimeShouldPromoteDoctorRouteInspectionSummary(ctx, &payload, payload.Action, options.DiagnosticsPreview, route) {
		return defaultRoute
	}
	info := normalizeBrowserRuntimeInfo(BrowserRuntimeInfo{
		Backend: route.Backend,
		Profile: route.Profile,
		Target:  route.RuntimeTarget,
	})
	if info == (BrowserRuntimeInfo{}) {
		return defaultRoute
	}
	return info
}

func browserRuntimeApplyActionDispatchRouteResultProjection(
	payload *browserRuntimePayload,
	projection browserRuntimeActionDispatchRouteResultProjection,
) {
	if payload == nil {
		return
	}
	payload.RouteResolution = projection.RouteResolution
	browserRuntimeApplyConfiguredProfilesProjection(payload, browserRuntimeConfiguredProfilesProjection{
		Profiles:             projection.ConfiguredProfiles,
		PreserveExisting:     true,
		RequireSelectedRoute: true,
	})
	if projection.ApplyRoutes {
		payload.Routes = projection.Routes
	}
	if projection.HideSelectedRoute {
		payload.SelectedRoute = nil
	}
}

func browserRuntimeFinalizeActionDispatchPayload(
	ctx browserRegistrationContext,
	payload *browserRuntimePayload,
	options browserRuntimeActionDispatchResultPostProcess,
) {
	if payload == nil {
		return
	}
	browserRuntimeHideImplicitLegacyHostFallbackSelections(payload, options.HiddenImplicitHostDefaultBase)
	browserRuntimeApplyActionDispatchRouteResultProjection(
		payload,
		browserRuntimeProjectActionDispatchRouteResult(ctx, *payload, options),
	)
	if options.UseDiagnosticsPreview {
		browserCanonicalizeHiddenManagedPayloadSelectionSummary(payload, options.DiagnosticsPreview)
	}
}
