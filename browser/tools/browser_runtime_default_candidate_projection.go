package tools

func browserTopLevelSummaryWithDefaultCandidateRoute(
	summary *browserTopLevelSummary,
	route browserRuntimeRouteDescriptor,
) *browserTopLevelSummary {
	if summary == nil {
		return nil
	}
	out := browserCloneTopLevelSummary(summary)
	out.DefaultCandidateRoute = route
	if browserUnifiedSummaryEmpty(*out) {
		return nil
	}
	return out
}

func browserTopLevelDisplayWithDefaultCandidateRoute(
	display *browserTopLevelDisplaySummary,
	route browserRuntimeRouteDescriptor,
) *browserTopLevelDisplaySummary {
	if display == nil {
		return nil
	}
	out := browserCloneTopLevelDisplaySummary(display)
	out.DefaultCandidateRoute = route
	if browserTopLevelDisplayEmpty(*out) {
		return nil
	}
	return out
}

func browserTopLevelSurfaceWithDefaultCandidateRoute(
	surface *browserTopLevelSurfaceSummary,
	route browserRuntimeRouteDescriptor,
) *browserTopLevelSurfaceSummary {
	if surface == nil {
		return nil
	}
	out := browserCloneTopLevelSurfaceSummary(surface)
	out.DefaultCandidateRoute = route
	if browserTopLevelSurfaceEmpty(*out) {
		return nil
	}
	return out
}

func browserTopLevelViewWithDefaultCandidateRoute(
	view *browserTopLevelViewSummary,
	route browserRuntimeRouteDescriptor,
) *browserTopLevelViewSummary {
	if view == nil {
		return nil
	}
	out := browserCloneTopLevelViewSummary(view)
	out.DefaultCandidateRoute = route
	out.Review = browserReviewSurfaceSummaryWithDefaultCandidateRoute(out.Review, route)
	if browserTopLevelViewEmpty(*out) {
		return nil
	}
	return out
}

func browserReviewSurfaceSummaryWithDefaultCandidateRoute(
	summary *browserReviewSurfaceSummary,
	route browserRuntimeRouteDescriptor,
) *browserReviewSurfaceSummary {
	if summary == nil {
		return nil
	}
	out := browserCloneReviewSurfaceSummary(summary)
	out.Explanation = browserTopLevelSummaryWithDefaultCandidateRoute(out.Explanation, route)
	out.Diagnostics = browserTopLevelSummaryWithDefaultCandidateRoute(out.Diagnostics, route)
	out.Summary = browserTopLevelSummaryWithDefaultCandidateRoute(out.Summary, route)
	out.Display = browserTopLevelDisplayWithDefaultCandidateRoute(out.Display, route)
	if browserReviewSurfaceSummaryEmpty(*out) {
		return nil
	}
	return out
}

func browserWorkbenchDisplayWithDefaultCandidateRoute(
	display *browserRuntimeWorkbenchDisplaySummary,
	route browserRuntimeRouteDescriptor,
) *browserRuntimeWorkbenchDisplaySummary {
	if display == nil {
		return nil
	}
	cloned := *display
	if len(display.Sections) > 0 {
		cloned.Sections = append([]string(nil), display.Sections...)
	}
	display = &cloned
	display.DefaultCandidateRoute = route
	if browserUnifiedWorkbenchDisplayEmpty(*display) {
		return nil
	}
	return display
}

func browserWorkbenchWithDefaultCandidateRoute(
	workbench *browserRuntimeWorkbenchSurfaceSummary,
	route browserRuntimeRouteDescriptor,
) *browserRuntimeWorkbenchSurfaceSummary {
	if workbench == nil {
		return nil
	}
	cloned := *workbench
	if len(workbench.Sections) > 0 {
		cloned.Sections = append([]string(nil), workbench.Sections...)
	}
	cloned.Review = browserCloneReviewSurfaceSummary(workbench.Review)
	if len(workbench.BrowserTools) > 0 {
		cloned.BrowserTools = append([]string(nil), workbench.BrowserTools...)
	}
	if len(workbench.ArtifactTools) > 0 {
		cloned.ArtifactTools = append([]string(nil), workbench.ArtifactTools...)
	}
	if len(workbench.ArtifactKinds) > 0 {
		cloned.ArtifactKinds = append([]string(nil), workbench.ArtifactKinds...)
	}
	if len(workbench.BrowserActKinds) > 0 {
		cloned.BrowserActKinds = append([]string(nil), workbench.BrowserActKinds...)
	}
	if len(workbench.BrowserOptInTargets) > 0 {
		cloned.BrowserOptInTargets = append([]string(nil), workbench.BrowserOptInTargets...)
	}
	cloned.Explanation = browserTopLevelSummaryWithDefaultCandidateRoute(workbench.Explanation, route)
	cloned.Diagnostics = browserTopLevelSummaryWithDefaultCandidateRoute(workbench.Diagnostics, route)
	cloned.Summary = browserTopLevelSummaryWithDefaultCandidateRoute(workbench.Summary, route)
	if len(workbench.RecommendedBrowserActions) > 0 {
		cloned.RecommendedBrowserActions = append([]string(nil), workbench.RecommendedBrowserActions...)
	}
	if len(workbench.RecommendedNodeActions) > 0 {
		cloned.RecommendedNodeActions = append([]string(nil), workbench.RecommendedNodeActions...)
	}
	workbench = &cloned
	workbench.Review = browserReviewSurfaceSummaryWithDefaultCandidateRoute(workbench.Review, route)
	workbench.DefaultCandidateRoute = route
	if browserUnifiedWorkbenchEmpty(*workbench) {
		return nil
	}
	return workbench
}

func browserRuntimeApplyDefaultCandidateRouteToPayloadShells(payload *browserRuntimePayload) {
	if payload == nil {
		return
	}
	route := payload.DefaultCandidateRoute
	if route == (browserRuntimeRouteDescriptor{}) {
		route = browserRuntimeDefaultCandidateRouteFromPayloadShells(payload)
		payload.DefaultCandidateRoute = route
	}
	payload.Explanation = browserTopLevelSummaryWithDefaultCandidateRoute(payload.Explanation, route)
	payload.Diagnostics = browserTopLevelSummaryWithDefaultCandidateRoute(payload.Diagnostics, route)
	payload.Summary = browserTopLevelSummaryWithDefaultCandidateRoute(payload.Summary, route)
	payload.Display = browserTopLevelDisplayWithDefaultCandidateRoute(payload.Display, route)
	payload.Review = browserReviewSurfaceSummaryWithDefaultCandidateRoute(payload.Review, route)
	payload.WorkbenchSummary = browserTopLevelSummaryWithDefaultCandidateRoute(payload.WorkbenchSummary, route)
	payload.WorkbenchDisplay = browserWorkbenchDisplayWithDefaultCandidateRoute(payload.WorkbenchDisplay, route)
	payload.Surface = browserTopLevelSurfaceWithDefaultCandidateRoute(payload.Surface, route)
	payload.View = browserTopLevelViewWithDefaultCandidateRoute(payload.View, route)
	payload.Workbench = browserWorkbenchWithDefaultCandidateRoute(payload.Workbench, route)
}

func browserRuntimeDefaultCandidateRouteFromPayloadShells(payload *browserRuntimePayload) browserRuntimeRouteDescriptor {
	if payload == nil {
		return browserRuntimeRouteDescriptor{}
	}
	for _, route := range []browserRuntimeRouteDescriptor{
		browserRuntimeRouteDescriptorFromTopLevelSummary(payload.Explanation),
		browserRuntimeRouteDescriptorFromTopLevelSummary(payload.Diagnostics),
		browserRuntimeRouteDescriptorFromTopLevelSummary(payload.Summary),
		browserRuntimeRouteDescriptorFromTopLevelDisplay(payload.Display),
		browserRuntimeRouteDescriptorFromReviewSurface(payload.Review),
		browserRuntimeRouteDescriptorFromTopLevelSummary(payload.WorkbenchSummary),
		browserRuntimeRouteDescriptorFromWorkbenchDisplay(payload.WorkbenchDisplay),
		browserRuntimeRouteDescriptorFromTopLevelSurface(payload.Surface),
		browserRuntimeRouteDescriptorFromTopLevelView(payload.View),
		browserRuntimeRouteDescriptorFromWorkbenchSurface(payload.Workbench),
	} {
		if route != (browserRuntimeRouteDescriptor{}) {
			return route
		}
	}
	return browserRuntimeRouteDescriptor{}
}
