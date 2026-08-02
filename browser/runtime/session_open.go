package browserruntime

import (
	"strings"
	"time"
)

// SharedSessionBrowserOpenResultEventRequest carries the route-scoped runtime
// result from an open action so it can share one browserruntime-owned contract
// with the broader resolved-target seam.
type SharedSessionBrowserOpenResultEventRequest struct {
	SessionID string
	Route     BrowserSessionRoute
	URL       string
	Title     string
	Source    string
}

// SharedSessionBrowserOpenResultEventResult captures the tracked target handle
// produced by an open runtime result.
type SharedSessionBrowserOpenResultEventResult struct {
	TargetID string
}

// ApplySharedSessionBrowserOpenResultEvent applies an open runtime result
// through the shared resolved-target seam so primary and sibling providers can
// refresh from the same source-time writeback event.
func (m SharedSessionBrowserObserverManager) ApplyOpenResultEvent(
	req SharedSessionBrowserOpenResultEventRequest,
) SharedSessionBrowserOpenResultEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.URL = strings.TrimSpace(req.URL)
	req.Title = strings.TrimSpace(req.Title)
	req.Source = strings.TrimSpace(req.Source)

	if m.SessionRegistry == nil || req.SessionID == "" {
		return SharedSessionBrowserOpenResultEventResult{}
	}

	resolved := m.ApplyResolvedTargetEvent(
		SharedSessionBrowserResolvedTargetEventRequest{
			SessionID: req.SessionID,
			Route:     req.Route,
			URL:       req.URL,
			Title:     req.Title,
			Source:    req.Source,
		},
	)
	return SharedSessionBrowserOpenResultEventResult{
		TargetID: strings.TrimSpace(resolved.TargetID),
	}
}

// ApplySharedSessionBrowserOpenResultEvent applies an open runtime result
// through the shared resolved-target seam so primary and sibling providers can
// refresh from the same source-time writeback event.
func ApplySharedSessionBrowserOpenResultEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserOpenResultEventRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserOpenResultEventResult {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ApplyOpenResultEvent(req)
}

// ApplySharedSessionBrowserOpenResultWithContext applies an open runtime
// result and routes the write through the top-level mutation seam when manager
// dependencies are available.
func ApplySharedSessionBrowserOpenResultWithContext(
	ctx SharedSessionBrowserMutationContext,
	req SharedSessionBrowserOpenResultEventRequest,
) SharedSessionBrowserOpenResultEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.URL = strings.TrimSpace(req.URL)
	req.Title = strings.TrimSpace(req.Title)
	req.Source = strings.TrimSpace(req.Source)

	if ctx.Registry == nil || req.SessionID == "" {
		return SharedSessionBrowserOpenResultEventResult{}
	}
	if ctx.usesWatchManagerEventSeam() {
		return ApplySharedSessionBrowserOpenResultEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			req,
			ctx.ReconnectWindow,
		)
	}

	resolved := ApplySharedSessionBrowserResolvedTargetWithContext(
		ctx,
		SharedSessionBrowserResolvedTargetEventRequest{
			SessionID: req.SessionID,
			Route:     req.Route,
			URL:       req.URL,
			Title:     req.Title,
			Source:    req.Source,
		},
	)
	return SharedSessionBrowserOpenResultEventResult{
		TargetID: strings.TrimSpace(resolved.TargetID),
	}
}

// ApplySharedSessionBrowserOpenResult applies an open runtime result through
// the shared open-result contract.
func ApplySharedSessionBrowserOpenResult(
	registry *BrowserSessionRegistry,
	req SharedSessionBrowserOpenResultEventRequest,
) SharedSessionBrowserOpenResultEventResult {
	return ApplySharedSessionBrowserOpenResultWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		req,
	)
}
