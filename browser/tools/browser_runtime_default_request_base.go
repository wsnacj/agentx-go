package tools

import "strings"

type browserDefaultRequestBaseResolver interface {
	ResolveBrowserDefaultRequestBase(map[string]any, BrowserRuntimeInfo) BrowserRuntimeInfo
}

type browserDefaultRequestPreviewResolver interface {
	ResolveBrowserDefaultRequestPreview(map[string]any, BrowserRuntimeInfo, bool) (BrowserRuntimeInfo, bool)
}

func browserResolveDefaultRequestPreview(params map[string]any, base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, backend BrowserBackend) (BrowserRuntimeInfo, bool) {
	base = normalizeBrowserRuntimeInfo(base)
	if resolver, ok := backend.(browserDefaultRequestPreviewResolver); ok {
		adjustedBase, adjustedHiddenImplicitHostDefaultBase := resolver.ResolveBrowserDefaultRequestPreview(params, base, hiddenImplicitHostDefaultBase)
		return normalizeBrowserRuntimeInfo(adjustedBase), adjustedHiddenImplicitHostDefaultBase
	}
	if resolver, ok := backend.(browserDefaultRequestBaseResolver); ok {
		adjustedBase := normalizeBrowserRuntimeInfo(resolver.ResolveBrowserDefaultRequestBase(params, base))
		if adjustedBase != base && adjustedBase != (BrowserRuntimeInfo{}) {
			targetRoute := strings.ToLower(strings.TrimSpace(adjustedBase.Target))
			if targetRoute != "" && targetRoute != "host" {
				hiddenImplicitHostDefaultBase = false
			}
		}
		return adjustedBase, hiddenImplicitHostDefaultBase
	}
	return base, hiddenImplicitHostDefaultBase
}

func browserRequestedRuntimeInfoForDefaultRequestBase(params map[string]any, base BrowserRuntimeInfo, backend BrowserBackend) (BrowserRuntimeInfo, bool, bool, error) {
	requested, explicitProfile, explicitTarget, err := browserRequestedRuntimeInfo(params, base)
	if err != nil || explicitProfile || explicitTarget {
		return requested, explicitProfile, explicitTarget, err
	}
	adjustedBase, _ := browserResolveDefaultRequestPreview(params, base, false, backend)
	if adjustedBase == normalizeBrowserRuntimeInfo(base) {
		return requested, explicitProfile, explicitTarget, nil
	}
	return browserRequestedRuntimeInfo(params, adjustedBase)
}
