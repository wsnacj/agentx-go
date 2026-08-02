package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

const (
	BrowserSurfaceUnified    = "browser_unified"
	BrowserSurfaceSpecialist = "browser_specialist"
	BrowserSurfaceCompat     = "browser_compat"
)

func BrowserUnifiedToolNames() []string {
	return agentxbrowserruntime.BrowserUnifiedToolNames()
}

func BrowserSpecialistToolNames() []string {
	return agentxbrowserruntime.BrowserSpecialistToolNames()
}

func BrowserCompatToolNames() []string {
	return agentxbrowserruntime.BrowserCompatToolNames()
}

func BrowserCompatToolNameForActKind(kind string) string {
	return agentxbrowserruntime.BrowserCompatToolNameForActKind(kind)
}

func BrowserAllToolNames() []string {
	return agentxbrowserruntime.BrowserAllToolNames()
}

func IsBrowserReadOnlyToolName(name string) bool {
	normalized := NormalizeToolName(name)
	return browserCompatIsReadOnly(normalized)
}

func IsBrowserLikelySideEffectToolName(name string) bool {
	normalized := NormalizeToolName(name)
	switch {
	case normalized == "", IsBrowserUnifiedToolName(normalized), IsBrowserReadOnlyToolName(normalized):
		return false
	case normalized == "browser_runtime", normalized == "browser_act":
		return true
	case IsBrowserCompatToolName(normalized):
		return browserCompatIsLikelySideEffect(normalized)
	default:
		return false
	}
}

func BrowserBuiltinRiskLevel(name string) (RiskLevel, bool) {
	normalized := NormalizeToolName(name)
	switch {
	case normalized == "":
		return RiskUnknown, false
	case IsBrowserUnifiedToolName(normalized), normalized == "browser_runtime":
		return RiskLow, true
	case normalized == "browser_act":
		return RiskHigh, true
	case IsBrowserCompatToolName(normalized):
		return browserCompatBuiltinRiskLevel(normalized)
	default:
		return RiskUnknown, false
	}
}

func IsBrowserUnifiedToolName(name string) bool {
	return agentxbrowserruntime.IsBrowserUnifiedToolName(name)
}

func IsBrowserSpecialistToolName(name string) bool {
	return agentxbrowserruntime.IsBrowserSpecialistToolName(name)
}

func IsBrowserCompatToolName(name string) bool {
	normalized := NormalizeToolName(name)
	if !agentxbrowserruntime.IsBrowserCompatToolName(normalized) {
		return false
	}
	_, ok := browserCompatDescriptorForTool(normalized)
	return ok
}

func BrowserSurfaceForToolName(name string) string {
	switch {
	case IsBrowserUnifiedToolName(name):
		return BrowserSurfaceUnified
	case IsBrowserSpecialistToolName(name):
		return BrowserSurfaceSpecialist
	case IsBrowserCompatToolName(name):
		return BrowserSurfaceCompat
	default:
		return ""
	}
}

// BrowserArtifactKindForToolName returns the stable artifact kind owned by the
// browser tool surface metadata. It only reports kinds with explicit browser
// owner facts; callers should keep path/content sniffing as a fallback.
func BrowserArtifactKindForToolName(name string) string {
	return browserCompatArtifactKind(name)
}

func BrowserSurfaceStatus(unifiedVisible bool, specialistVisible, compatVisible []string) string {
	switch {
	case unifiedVisible:
		return "ok"
	case len(specialistVisible) > 0 || len(compatVisible) > 0:
		return "warn"
	default:
		return "error"
	}
}

func NormalizeBrowserSurface(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case BrowserSurfaceUnified:
		return BrowserSurfaceUnified
	case BrowserSurfaceSpecialist:
		return BrowserSurfaceSpecialist
	case BrowserSurfaceCompat:
		return BrowserSurfaceCompat
	default:
		return ""
	}
}

func BrowserSurfaceFallbackEntrypoint(surface string) string {
	switch NormalizeBrowserSurface(surface) {
	case BrowserSurfaceUnified:
		return "browser"
	case BrowserSurfaceSpecialist:
		return "browser_runtime"
	case BrowserSurfaceCompat:
		compatNames := BrowserCompatToolNames()
		if len(compatNames) == 0 {
			return ""
		}
		return compatNames[0]
	default:
		return ""
	}
}

func ResolveBrowserSurface(surface string, entrypoint string) (string, string) {
	normalizedSurface := NormalizeBrowserSurface(surface)
	normalizedEntrypoint := strings.TrimSpace(entrypoint)
	if normalizedSurface != "" {
		if normalizedEntrypoint == "" {
			normalizedEntrypoint = BrowserSurfaceFallbackEntrypoint(normalizedSurface)
		}
		return normalizedSurface, normalizedEntrypoint
	}
	switch normalizedEntrypoint {
	case "browser":
		return BrowserSurfaceUnified, normalizedEntrypoint
	case "browser_runtime", "browser_act":
		return BrowserSurfaceSpecialist, normalizedEntrypoint
	default:
		if IsBrowserCompatToolName(normalizedEntrypoint) {
			return BrowserSurfaceCompat, normalizedEntrypoint
		}
		return "", normalizedEntrypoint
	}
}

func BrowserDefaultEntrypoint(unifiedVisible bool, specialistVisible, compatVisible []string) string {
	switch {
	case unifiedVisible:
		return "browser"
	case len(specialistVisible) > 0:
		return specialistVisible[0]
	case len(compatVisible) > 0:
		return compatVisible[0]
	default:
		return ""
	}
}

func BrowserDefaultSurface(unifiedVisible bool, specialistVisible, compatVisible []string) string {
	switch {
	case unifiedVisible:
		return BrowserSurfaceUnified
	case len(specialistVisible) > 0:
		return BrowserSurfaceSpecialist
	case len(compatVisible) > 0:
		return BrowserSurfaceCompat
	default:
		return ""
	}
}

func BrowserVisibleSurfaceLabels(specialistVisible, compatVisible []string) []string {
	labels := make([]string, 0, 2)
	if len(specialistVisible) > 0 {
		labels = append(labels, "specialist="+strings.Join(specialistVisible, "/"))
	}
	if len(compatVisible) > 0 {
		labels = append(labels, "deprecated_compat="+strings.Join(compatVisible, "/"))
	}
	if len(labels) == 0 {
		return nil
	}
	return labels
}
