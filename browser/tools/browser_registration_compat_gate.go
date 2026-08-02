package tools

func browserRegistrationSupportsExplicitManagedCompatTool(ctx browserRegistrationContext, enabled map[string]bool, name string) bool {
	return len(browserRegistrationManagedOptInProjection(ctx, enabled).CompatTargets(name)) != 0
}

func browserRegistrationExplicitManagedActKinds(ctx browserRegistrationContext) []string {
	return browserRegistrationManagedOptInProjection(ctx, ctx.enabledTools).OptInActSurface().Kinds
}

func browserRegistrationSupportsExplicitManagedActTool(ctx browserRegistrationContext, enabled map[string]bool) bool {
	if len(enabled) == 0 {
		return false
	}
	return len(browserRegistrationManagedOptInProjection(ctx, enabled).OptInActSurface().Kinds) != 0
}

func browserRegistrationManagedOptInProjection(ctx browserRegistrationContext, enabled map[string]bool) browserManagedOptInProjection {
	if ctx.managedOptInProjectionReady && browserRegistrationEnabledToolsMatch(ctx.enabledTools, enabled) {
		return ctx.managedOptInProjection
	}
	return browserManagedOptInProjectionForCapabilities(enabled, ctx.capabilities, ctx.opts.NodeBackend, ctx.opts.SandboxBackend)
}

func browserRegistrationEnabledToolsMatch(base map[string]bool, candidate map[string]bool) bool {
	if len(base) != len(candidate) {
		return false
	}
	if len(base) == 0 {
		return true
	}
	for name, allowed := range base {
		if candidate[name] != allowed {
			return false
		}
	}
	return true
}
