package tools

type browserRegistrationExecutionPreview struct {
	Registration                  browserDefaultRuntimePreview
	DefaultRoute                  BrowserRuntimeInfo
	DefaultCandidateRoute         BrowserRuntimeInfo
	DispatchBase                  BrowserRuntimeInfo
	HiddenImplicitHostDefaultBase bool
	EffectiveBackend              BrowserBackend
}

func browserRegistrationExecutionPreviewDefaultCandidateRoute(registration browserDefaultRuntimePreview) BrowserRuntimeInfo {
	if info := normalizeBrowserRuntimeInfo(registration.SubstrateSummary.DefaultCandidateRoute); info != (BrowserRuntimeInfo{}) {
		return info
	}
	return browserVisibleDefaultRuntimeInfoForPreview(registration)
}

func browserRegistrationExecutionPreviewForContext(ctx browserRegistrationContext) browserRegistrationExecutionPreview {
	registration := browserRegistrationDefaultRuntimePreview(ctx)
	return browserRegistrationExecutionPreview{
		Registration:                  registration,
		DefaultRoute:                  normalizeBrowserRuntimeInfo(registration.LogicalDefaultRoute),
		DefaultCandidateRoute:         browserRegistrationExecutionPreviewDefaultCandidateRoute(registration),
		DispatchBase:                  normalizeBrowserRuntimeInfo(registration.VisibleDefaultRoute),
		HiddenImplicitHostDefaultBase: registration.HiddenImplicitHostDefaultBase,
		EffectiveBackend:              registration.EffectiveBackend,
	}
}
