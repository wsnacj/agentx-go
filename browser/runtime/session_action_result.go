package browserruntime

import (
	"strings"
	"time"
)

// SharedSessionBrowserActionResultEventRequest carries a generic route-scoped
// runtime result that only needs target tracking plus optional tool-facing
// review posture.
type SharedSessionBrowserActionResultEventRequest struct {
	SessionID         string
	Route             BrowserSessionRoute
	PreferredTargetID string
	TabIndex          int
	TrackCurrent      bool
	URL               string
	Title             string
	Source            string
	SetCurrent        bool
	ReviewDecision    string
	ReviewReady       bool
	Note              string
}

// SharedSessionBrowserActionResultEventResult captures the tracked target and
// tool-facing review posture projected from a generic runtime result.
type SharedSessionBrowserActionResultEventResult struct {
	TargetID       string
	ReviewDecision string
	ReviewReady    bool
	Note           string
}

func sharedSessionBrowserActionResult(
	preferredTargetID string,
	reviewDecision string,
	reviewReady bool,
	note string,
) SharedSessionBrowserActionResultEventResult {
	return SharedSessionBrowserActionResultEventResult{
		TargetID:       strings.TrimSpace(preferredTargetID),
		ReviewDecision: strings.TrimSpace(reviewDecision),
		ReviewReady:    reviewReady,
		Note:           strings.TrimSpace(note),
	}
}

// ApplySharedSessionBrowserActionResultEvent applies a generic route-scoped
// runtime result through the shared target-tracking seam so primary and sibling
// providers can refresh from the same source-time writeback event.
func ApplySharedSessionBrowserActionResultEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserActionResultEventRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserActionResultEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.PreferredTargetID = strings.TrimSpace(req.PreferredTargetID)
	req.URL = strings.TrimSpace(req.URL)
	req.Title = strings.TrimSpace(req.Title)
	req.Source = strings.TrimSpace(req.Source)
	req.ReviewDecision = strings.TrimSpace(req.ReviewDecision)
	req.Note = strings.TrimSpace(req.Note)

	result := sharedSessionBrowserActionResult(
		req.PreferredTargetID,
		req.ReviewDecision,
		req.ReviewReady,
		req.Note,
	)
	if sessionRegistry == nil || req.SessionID == "" {
		return result
	}
	if req.TabIndex <= 0 && !req.TrackCurrent && !req.SetCurrent {
		return result
	}

	tracked := ApplySharedSessionBrowserTargetEvent(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		SharedSessionBrowserTargetEventRequest{
			SessionID:        req.SessionID,
			Route:            req.Route,
			ExplicitTargetID: req.PreferredTargetID,
			TabIndex:         req.TabIndex,
			URL:              req.URL,
			Title:            req.Title,
			Source:           req.Source,
			SetCurrent:       req.SetCurrent,
		},
		reconnectWindow,
	)
	result.TargetID = firstNonEmptyString(result.TargetID, strings.TrimSpace(tracked.TargetID))
	return result
}

// ApplySharedSessionBrowserActionResultWithContext applies a generic runtime
// result and routes the write through the top-level mutation seam when manager
// dependencies are available.
func ApplySharedSessionBrowserActionResultWithContext(
	ctx SharedSessionBrowserMutationContext,
	req SharedSessionBrowserActionResultEventRequest,
) SharedSessionBrowserActionResultEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.PreferredTargetID = strings.TrimSpace(req.PreferredTargetID)
	req.URL = strings.TrimSpace(req.URL)
	req.Title = strings.TrimSpace(req.Title)
	req.Source = strings.TrimSpace(req.Source)
	req.ReviewDecision = strings.TrimSpace(req.ReviewDecision)
	req.Note = strings.TrimSpace(req.Note)

	result := sharedSessionBrowserActionResult(
		req.PreferredTargetID,
		req.ReviewDecision,
		req.ReviewReady,
		req.Note,
	)
	if ctx.Registry == nil || req.SessionID == "" {
		return result
	}
	if req.TabIndex <= 0 && !req.TrackCurrent && !req.SetCurrent {
		return result
	}
	if ctx.usesWatchManagerEventSeam() {
		return ApplySharedSessionBrowserActionResultEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			req,
			ctx.ReconnectWindow,
		)
	}

	tracked := ApplySharedSessionBrowserTargetWithContext(
		ctx,
		SharedSessionBrowserTargetEventRequest{
			SessionID:        req.SessionID,
			Route:            req.Route,
			ExplicitTargetID: req.PreferredTargetID,
			TabIndex:         req.TabIndex,
			URL:              req.URL,
			Title:            req.Title,
			Source:           req.Source,
			SetCurrent:       req.SetCurrent,
		},
	)
	result.TargetID = firstNonEmptyString(result.TargetID, strings.TrimSpace(tracked.TargetID))
	return result
}

// ApplySharedSessionBrowserActionResult applies a generic runtime result
// through the shared browserruntime contract.
func ApplySharedSessionBrowserActionResult(
	registry *BrowserSessionRegistry,
	req SharedSessionBrowserActionResultEventRequest,
) SharedSessionBrowserActionResultEventResult {
	return ApplySharedSessionBrowserActionResultWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		req,
	)
}
