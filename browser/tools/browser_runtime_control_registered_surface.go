package tools

type browserRuntimeRegisteredSurface struct {
	BrowserTools    []string
	ArtifactTools   []string
	BrowserActKinds []string
}

func browserRuntimeRegisteredOrEnabledTool(ctx browserRegistrationContext, name string) bool {
	name = NormalizeToolName(name)
	if name == "" {
		return false
	}
	if registered := browserRuntimeRegisteredToolSet(ctx); len(registered) != 0 {
		return registered[name]
	}
	if toolEnabled(ctx.enabledTools, name) {
		return true
	}
	return toolEnabled(ctx.enabledTools, "browser") && (name == "browser_runtime" || name == "browser_act")
}

func browserRuntimeRegisteredSurfaceForCapabilities(ctx browserRegistrationContext, capabilities BrowserCapabilities) browserRuntimeRegisteredSurface {
	if ctx.derivedCache != nil {
		return ctx.derivedCache.registeredSurfaceForCapabilities(ctx, capabilities)
	}
	return browserRuntimeRegisteredSurfaceForCapabilitiesUncached(ctx, capabilities)
}

func browserRuntimeRegisteredSurfaceForCapabilitiesUncached(ctx browserRegistrationContext, capabilities BrowserCapabilities) browserRuntimeRegisteredSurface {
	return browserRuntimeRegisteredSurface{
		BrowserTools:    browserRuntimeSupportedToolNames(ctx, capabilities),
		ArtifactTools:   browserRuntimeArtifactTools(ctx, capabilities),
		BrowserActKinds: browserRuntimeSupportedActKinds(ctx, capabilities),
	}
}

func browserRuntimeMergeRegisteredSurfaces(surfaces ...browserRuntimeRegisteredSurface) browserRuntimeRegisteredSurface {
	out := browserRuntimeRegisteredSurface{}
	for _, surface := range surfaces {
		out.BrowserTools = mergeToolMetadataStrings(out.BrowserTools, surface.BrowserTools)
		out.ArtifactTools = mergeToolMetadataStrings(out.ArtifactTools, surface.ArtifactTools)
		out.BrowserActKinds = mergeToolMetadataStrings(out.BrowserActKinds, surface.BrowserActKinds)
	}
	return out
}
