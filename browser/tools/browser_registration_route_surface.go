package tools

type browserCompatRuntimeIdentity struct {
	ToolName            string
	Action              string
	EventSource         string
	Actor               string
	CapabilityMetadata  browserRuntimeCapabilityMetadata
	BrowserSurface      string
	BrowserOptInTargets []string
}

func browserRegistrationCompatPayloadRouteSurface(ctx browserRegistrationContext, name string) (string, []string) {
	targets := browserRegistrationManagedOptInProjection(ctx, ctx.enabledTools).CompatTargets(name)
	return browserTopLevelRouteSurfaceLabel("explicit_managed_opt_in", targets)
}

func browserRegistrationCompatPayloadCapabilityMetadata(
	ctx browserRegistrationContext,
	capabilities BrowserCapabilities,
	name string,
) (browserRuntimeCapabilityMetadata, string, []string) {
	browserSurface, browserOptInTargets := browserRegistrationCompatPayloadRouteSurface(ctx, name)
	return browserRegistrationPayloadCapabilityMetadata(ctx, capabilities, browserSurface, browserOptInTargets), browserSurface, browserOptInTargets
}

func browserRegistrationCompatRuntimeIdentity(
	ctx browserRegistrationContext,
	capabilities BrowserCapabilities,
	name string,
	action string,
) browserCompatRuntimeIdentity {
	capabilityMetadata, browserSurface, browserOptInTargets := browserRegistrationCompatPayloadCapabilityMetadata(ctx, capabilities, name)
	return browserCompatRuntimeIdentity{
		ToolName:            NormalizeToolName(name),
		Action:              browserNormalizeToolToken(action),
		EventSource:         browserCompatEventSource(name),
		Actor:               browserCompatActor(name, action),
		CapabilityMetadata:  capabilityMetadata,
		BrowserSurface:      browserSurface,
		BrowserOptInTargets: append([]string(nil), browserOptInTargets...),
	}
}

func (identity browserCompatRuntimeIdentity) reviewReason(state browserPendingTargetReviewState, force bool) string {
	return browserCompatPendingTargetReviewReason(identity.ToolName, identity.Action, state, force)
}

func (identity browserCompatRuntimeIdentity) legacyHostFallbackError(
	hiddenImplicitHostDefaultBase bool,
	explicitRuntimeTarget bool,
	runtimeInfo BrowserRuntimeInfo,
	target browserToolTarget,
	requestURL string,
) error {
	return browserCompatImplicitLegacyHostFallbackError(identity.ToolName, hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, target, identity.Action, requestURL)
}

func browserRegistrationActPayloadRouteSurface(ctx browserRegistrationContext, kind string) (string, []string) {
	targets := browserRegistrationManagedOptInProjection(ctx, ctx.enabledTools).actTargetsForKind(kind)
	return browserTopLevelRouteSurfaceLabel("explicit_managed_opt_in", targets)
}
