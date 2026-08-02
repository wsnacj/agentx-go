package browserruntime

import "strings"

// ResolveSharedSessionBrowserCurrentTarget resolves the current selected target
// for a scoped session route after pruning stale route state. When
// allowDefaultFallback is true, it falls back to the default route selection if
// no current target exists on the scoped route.
func ResolveSharedSessionBrowserCurrentTarget(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, allowDefaultFallback bool) (BrowserSessionTarget, bool) {
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	if registry == nil || sessionID == "" {
		return BrowserSessionTarget{}, false
	}
	registry.PruneStaleRouteState(sessionID, route)
	target, ok := registry.CurrentTargetForRoute(sessionID, route)
	if ok && strings.TrimSpace(target.ID) != "" {
		return target, true
	}
	if !allowDefaultFallback {
		return BrowserSessionTarget{}, false
	}
	defaultRoute := BrowserSessionRoute{}
	registry.PruneStaleRouteState(sessionID, defaultRoute)
	target, ok = registry.CurrentTargetForRoute(sessionID, defaultRoute)
	if !ok || strings.TrimSpace(target.ID) == "" {
		return BrowserSessionTarget{}, false
	}
	return target, true
}

// ResolveSharedSessionBrowserTarget resolves an explicit target handle, a
// route-scoped tab index, or the current selected target for a scoped session
// route. Current-target lookup can optionally fall back to the default route.
func ResolveSharedSessionBrowserTarget(registry *BrowserSessionRegistry, sessionID string, route BrowserSessionRoute, targetID string, tabIndex int, allowDefaultCurrentFallback bool) (BrowserSessionTarget, bool) {
	sessionID = strings.TrimSpace(sessionID)
	route = normalizeBrowserSessionRoute(route)
	targetID = strings.TrimSpace(targetID)
	if registry == nil || sessionID == "" {
		return BrowserSessionTarget{}, false
	}
	if targetID != "" {
		return registry.ResolveTarget(sessionID, targetID)
	}
	if tabIndex > 0 {
		registry.PruneStaleRouteState(sessionID, route)
		target, ok := registry.ResolveTabForRoute(sessionID, route, tabIndex)
		if !ok || strings.TrimSpace(target.ID) == "" {
			return BrowserSessionTarget{}, false
		}
		return target, true
	}
	return ResolveSharedSessionBrowserCurrentTarget(registry, sessionID, route, allowDefaultCurrentFallback)
}
