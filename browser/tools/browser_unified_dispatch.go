package tools

func browserUnifiedDelegableActKinds(ctx browserRegistrationContext) []string {
	kinds := append([]string(nil), ctx.capabilities.SupportedActKinds()...)
	kinds = append(kinds, browserManagedAdditionalActSurfaceForCapabilities(ctx.enabledTools, ctx.capabilities, ctx.opts.NodeBackend, ctx.opts.SandboxBackend).Kinds...)
	return mergeToolMetadataStrings(nil, kinds)
}

func browserUnifiedCanDelegateActAction(
	ctx browserRegistrationContext,
	runtimeActions []string,
	params map[string]any,
	action string,
) bool {
	action = browserNormalizeToolToken(action)
	kinds := browserUnifiedDelegableActKinds(ctx)
	if action == "" || len(kinds) == 0 {
		return false
	}
	if _, ok := browserUnifiedRuntimeActionAliases[action]; ok || containsString(runtimeActions, action) {
		return false
	}
	kind := browserNormalizeToolToken(browserUnifiedActKind(params, action))
	if kind == "" {
		return false
	}
	return containsString(kinds, kind)
}
