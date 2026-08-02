package browserruntime

import "strings"

var browserUnifiedToolNames = []string{"browser"}

var browserSpecialistToolNames = []string{"browser_runtime", "browser_act"}

type browserCompatToolSurface struct {
	Name          string
	ActKind       string
	RuntimeAction string
}

var browserCompatToolSurfaces = []browserCompatToolSurface{
	{Name: "browser_open", ActKind: "open", RuntimeAction: "open"},
	{Name: "browser_navigate", ActKind: "navigate", RuntimeAction: "navigate"},
	{Name: "browser_tabs", ActKind: "list_tabs", RuntimeAction: "tabs"},
	{Name: "browser_extract", ActKind: "extract", RuntimeAction: "extract"},
	{Name: "browser_screenshot", ActKind: "screenshot", RuntimeAction: "screenshot"},
	{Name: "browser_click", ActKind: "click", RuntimeAction: "click"},
	{Name: "browser_type", ActKind: "type", RuntimeAction: "type"},
	{Name: "browser_eval", ActKind: "evaluate", RuntimeAction: "evaluate"},
}

var browserCompatToolNames = browserCompatToolSurfaceNames(browserCompatToolSurfaces)

var browserCompatSurfacesByToolName = browserCompatToolSurfaceByName(browserCompatToolSurfaces)

var browserCompatToolNamesByActKind = browserCompatToolNameByActKind(browserCompatToolSurfaces)

func BrowserUnifiedToolNames() []string {
	return append([]string(nil), browserUnifiedToolNames...)
}

func BrowserSpecialistToolNames() []string {
	return append([]string(nil), browserSpecialistToolNames...)
}

func BrowserCompatToolNames() []string {
	return append([]string(nil), browserCompatToolNames...)
}

func BrowserAllToolNames() []string {
	out := make([]string, 0, len(browserUnifiedToolNames)+len(browserSpecialistToolNames)+len(browserCompatToolNames))
	out = append(out, browserUnifiedToolNames...)
	out = append(out, browserSpecialistToolNames...)
	out = append(out, browserCompatToolNames...)
	return out
}

func IsBrowserUnifiedToolName(name string) bool {
	return browserToolSurfaceContains(browserUnifiedToolNames, normalizeBrowserToolSurfaceToken(name))
}

func IsBrowserSpecialistToolName(name string) bool {
	return browserToolSurfaceContains(browserSpecialistToolNames, normalizeBrowserToolSurfaceToken(name))
}

func IsBrowserCompatToolName(name string) bool {
	_, ok := browserCompatSurfacesByToolName[normalizeBrowserToolSurfaceToken(name)]
	return ok
}

func IsBrowserToolName(name string) bool {
	normalized := normalizeBrowserToolSurfaceToken(name)
	if normalized == "" {
		return false
	}
	if IsBrowserUnifiedToolName(normalized) || IsBrowserSpecialistToolName(normalized) {
		return true
	}
	return IsBrowserCompatToolName(normalized)
}

func BrowserCompatActKindForToolName(name string) string {
	return browserCompatSurfacesByToolName[normalizeBrowserToolSurfaceToken(name)].ActKind
}

func BrowserCompatToolNameForActKind(kind string) string {
	normalized := normalizeBrowserToolSurfaceToken(kind)
	if normalized == "" {
		return ""
	}
	return browserCompatToolNamesByActKind[normalized]
}

func BrowserRuntimeActionForToolCall(toolName string, params map[string]any) string {
	switch normalizeBrowserToolSurfaceToken(toolName) {
	case "browser":
		if action := browserToolSurfaceFirstString(params, "action"); action != "" {
			return normalizeBrowserToolSurfaceToken(action)
		}
		if kind := browserToolSurfaceFirstString(params, "kind"); kind != "" {
			return normalizeBrowserToolSurfaceToken(kind)
		}
		return "browser"
	case "browser_runtime":
		if action := browserToolSurfaceFirstString(params, "action"); action != "" {
			return normalizeBrowserToolSurfaceToken(action)
		}
		return "runtime"
	case "browser_act":
		if kind := browserToolSurfaceFirstString(params, "kind"); kind != "" {
			return normalizeBrowserToolSurfaceToken(kind)
		}
		return "act"
	default:
		if surface, ok := browserCompatSurfacesByToolName[normalizeBrowserToolSurfaceToken(toolName)]; ok && surface.RuntimeAction != "" {
			return surface.RuntimeAction
		}
		return "browser"
	}
}

func browserCompatToolSurfaceNames(surfaces []browserCompatToolSurface) []string {
	names := make([]string, 0, len(surfaces))
	for _, surface := range surfaces {
		name := normalizeBrowserToolSurfaceToken(surface.Name)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return names
}

func browserCompatToolSurfaceByName(surfaces []browserCompatToolSurface) map[string]browserCompatToolSurface {
	out := make(map[string]browserCompatToolSurface, len(surfaces))
	for _, surface := range surfaces {
		surface.Name = normalizeBrowserToolSurfaceToken(surface.Name)
		surface.ActKind = normalizeBrowserToolSurfaceToken(surface.ActKind)
		surface.RuntimeAction = normalizeBrowserToolSurfaceToken(surface.RuntimeAction)
		if surface.Name == "" {
			continue
		}
		out[surface.Name] = surface
	}
	return out
}

func browserCompatToolNameByActKind(surfaces []browserCompatToolSurface) map[string]string {
	out := make(map[string]string, len(surfaces))
	for _, surface := range surfaces {
		name := normalizeBrowserToolSurfaceToken(surface.Name)
		actKind := normalizeBrowserToolSurfaceToken(surface.ActKind)
		if name == "" || actKind == "" {
			continue
		}
		out[actKind] = name
	}
	return out
}

func normalizeBrowserToolSurfaceToken(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func browserToolSurfaceContains(values []string, want string) bool {
	if want == "" {
		return false
	}
	for _, value := range values {
		if normalizeBrowserToolSurfaceToken(value) == want {
			return true
		}
	}
	return false
}

func browserToolSurfaceFirstString(params map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := params[key]
		if !ok {
			continue
		}
		text, ok := value.(string)
		if !ok {
			continue
		}
		if trimmed := strings.TrimSpace(text); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
