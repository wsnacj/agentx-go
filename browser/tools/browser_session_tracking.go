package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

var defaultBrowserSessionRegistry = agentxbrowserruntime.NewBrowserSessionRegistry()
var defaultBrowserSessionStateRegistry = agentxbrowserruntime.NewBrowserSessionStateRegistry()

func DefaultBrowserSessionRegistry() *BrowserSessionRegistry {
	return defaultBrowserSessionRegistry
}

func DefaultBrowserSessionStateRegistry() *agentxbrowserruntime.BrowserSessionStateRegistry {
	return defaultBrowserSessionStateRegistry
}

func browserSessionRegistryForOptions(opts BrowserToolOptions) *BrowserSessionRegistry {
	if opts.SessionRegistry != nil {
		return opts.SessionRegistry
	}
	return defaultBrowserSessionRegistry
}

func browserSessionRoute(runtimeInfo BrowserRuntimeInfo, browserApp string, backend string) agentxbrowserruntime.BrowserSessionRoute {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	return agentxbrowserruntime.BrowserSessionRoute{
		Backend:    browserSessionTrackingBackend(runtimeInfo, backend),
		Profile:    strings.TrimSpace(runtimeInfo.Profile),
		Target:     strings.TrimSpace(runtimeInfo.Target),
		BrowserApp: strings.TrimSpace(browserApp),
	}
}

func browserSessionTrackingBackend(runtimeInfo BrowserRuntimeInfo, backend string) string {
	runtimeInfo = normalizeBrowserRuntimeInfo(runtimeInfo)
	if strings.TrimSpace(runtimeInfo.Backend) != "" {
		return strings.TrimSpace(runtimeInfo.Backend)
	}
	backend = strings.TrimSpace(backend)
	if canonical := strings.TrimSpace(browserRuntimeCanonicalBackend(backend)); canonical != "" {
		return canonical
	}
	return backend
}

func browserSessionMutationRegistry(
	watchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager,
	registry *BrowserSessionRegistry,
) *BrowserSessionRegistry {
	if registry == nil {
		registry = watchManagerProvider.SessionRegistry
	}
	return registry
}

func browserSharedMutationContext(
	watchManagerProvider agentxbrowserruntime.SharedSessionBrowserObserverManager,
	registry *BrowserSessionRegistry,
) agentxbrowserruntime.SharedSessionBrowserMutationContext {
	return agentxbrowserruntime.SharedSessionBrowserMutationContext{
		Registry:        browserSessionMutationRegistry(watchManagerProvider, registry),
		RunRegistry:     watchManagerProvider.RunRegistry,
		StateRegistry:   watchManagerProvider.StateRegistry,
		ReconnectWindow: browserRuntimeReconnectWatchdogWindow,
	}
}

func browserTargetIDForTab(ctx context.Context, registry *BrowserSessionRegistry, runtimeInfo BrowserRuntimeInfo, browserApp string, tabIndex int) string {
	if registry == nil || tabIndex <= 0 {
		return ""
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return ""
	}
	route := browserSessionRoute(runtimeInfo, browserApp, "")
	return agentxbrowserruntime.ResolveSharedSessionBrowserTabTargetID(registry, sessionID, route, tabIndex)
}

func browserCurrentTargetSelectionSnapshotForRoute(ctx context.Context, registry *BrowserSessionRegistry, runtimeInfo BrowserRuntimeInfo, browserApp string, backend string) *agentxbrowserruntime.BrowserSessionTargetSelection {
	if registry == nil {
		return nil
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return nil
	}
	return agentxbrowserruntime.SnapshotSharedSessionBrowserCurrentTargetSelection(registry, sessionID, browserSessionRoute(runtimeInfo, browserApp, backend))
}

func browserResolvedTargetURL(ctx context.Context, registry *BrowserSessionRegistry, runtimeInfo BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, browserApp string, target browserToolTarget) string {
	if tracked, ok := browserResolvedTarget(ctx, registry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target); ok {
		return strings.TrimSpace(tracked.URL)
	}
	return ""
}

func browserResolvedTarget(ctx context.Context, registry *BrowserSessionRegistry, runtimeInfo BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, browserApp string, target browserToolTarget) (BrowserSessionTarget, bool) {
	if registry == nil {
		return BrowserSessionTarget{}, false
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return BrowserSessionTarget{}, false
	}
	tracked, ok := agentxbrowserruntime.ResolveSharedSessionBrowserTarget(
		registry,
		sessionID,
		browserSessionRoute(runtimeInfo, browserApp, ""),
		target.TargetID,
		target.TabIndex,
		false,
	)
	if !ok {
		return BrowserSessionTarget{}, false
	}
	if hiddenImplicitHostDefaultBase &&
		strings.TrimSpace(target.TargetID) == "" &&
		target.TabIndex <= 0 {
		trackedTarget := strings.ToLower(strings.TrimSpace(tracked.Target))
		if trackedTarget == "" || trackedTarget == "host" {
			return BrowserSessionTarget{}, false
		}
	}
	return tracked, true
}

func browserEffectiveBrowserApp(ctx context.Context, registry *BrowserSessionRegistry, runtimeInfo BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, requested string, target browserToolTarget) string {
	requested = strings.TrimSpace(requested)
	if requested != "" {
		return requested
	}
	if tracked, ok := browserResolvedTarget(ctx, registry, runtimeInfo, hiddenImplicitHostDefaultBase, "", target); ok {
		return strings.TrimSpace(tracked.BrowserApp)
	}
	if tracked, ok := browserResolvedTarget(ctx, registry, runtimeInfo, hiddenImplicitHostDefaultBase, "", browserToolTarget{}); ok {
		return strings.TrimSpace(tracked.BrowserApp)
	}
	return ""
}

func browserRuntimeSessionTargetSelectionFromShared(target *agentxbrowserruntime.BrowserSessionTargetSelection) *browserRuntimeSessionTargetSelection {
	if target == nil {
		return nil
	}
	return &browserRuntimeSessionTargetSelection{
		ID:            strings.TrimSpace(target.ID),
		TabIndex:      target.TabIndex,
		URL:           strings.TrimSpace(target.URL),
		Title:         strings.TrimSpace(target.Title),
		Backend:       strings.TrimSpace(target.Backend),
		Profile:       strings.TrimSpace(target.Profile),
		RuntimeTarget: strings.TrimSpace(target.RuntimeTarget),
		BrowserApp:    strings.TrimSpace(target.BrowserApp),
		Source:        strings.TrimSpace(target.Source),
	}
}
