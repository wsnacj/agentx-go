package browserruntime

import (
	"strings"
	"time"
)

// SharedSessionBrowserPageActionResultEventRequest carries the route-scoped
// runtime result from page actions like extract/snapshot/click/type/eval so
// resolved-target tracking and pending-review posture can share one
// browserruntime-owned contract.
type SharedSessionBrowserPageActionResultEventRequest struct {
	SessionID         string
	Route             BrowserSessionRoute
	PreferredTargetID string
	TabIndex          int
	URL               string
	Title             string
	Source            string
	Actor             string
	Force             bool
	Review            SharedSessionBrowserPendingTargetReviewState
}

// SharedSessionBrowserPageActionResultEventResult captures the tracked target
// together with any pending-review posture projected from a page-action
// runtime result.
type SharedSessionBrowserPageActionResultEventResult struct {
	TargetID       string
	ReviewDecision string
	ReviewReady    bool
	Note           string
}

func sharedSessionBrowserPageActionResult(
	preferredTargetID string,
	actor string,
	force bool,
	review SharedSessionBrowserPendingTargetReviewState,
) SharedSessionBrowserPageActionResultEventResult {
	result := SharedSessionBrowserPageActionResultEventResult{
		TargetID: strings.TrimSpace(preferredTargetID),
	}
	if review.Review == nil {
		return result
	}
	result.ReviewDecision = SharedSessionBrowserPendingTargetReviewDecision(review, force)
	result.ReviewReady = force
	result.Note = SharedSessionBrowserPendingTargetReviewReason(firstNonEmptyString(strings.TrimSpace(actor), "browser page action"), review, force)
	return result
}

// ApplySharedSessionBrowserPageActionResultEvent applies a page-action runtime
// result through the shared resolved-target seam so tracking and review posture
// can reuse the same source-time writeback contract.
func ApplySharedSessionBrowserPageActionResultEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserPageActionResultEventRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserPageActionResultEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.PreferredTargetID = strings.TrimSpace(req.PreferredTargetID)
	req.URL = strings.TrimSpace(req.URL)
	req.Title = strings.TrimSpace(req.Title)
	req.Source = strings.TrimSpace(req.Source)
	req.Actor = strings.TrimSpace(req.Actor)

	result := sharedSessionBrowserPageActionResult(req.PreferredTargetID, req.Actor, req.Force, req.Review)
	if sessionRegistry == nil || req.SessionID == "" {
		return result
	}

	resolved := ApplySharedSessionBrowserResolvedTargetEvent(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		SharedSessionBrowserResolvedTargetEventRequest{
			SessionID: req.SessionID,
			Route:     req.Route,
			TabIndex:  req.TabIndex,
			URL:       req.URL,
			Title:     req.Title,
			Source:    req.Source,
		},
		reconnectWindow,
	)
	result.TargetID = firstNonEmptyString(result.TargetID, strings.TrimSpace(resolved.TargetID))
	return result
}

// ApplySharedSessionBrowserPageActionResultWithContext applies a page-action
// runtime result and routes the write through the top-level mutation seam when
// manager dependencies are available.
func ApplySharedSessionBrowserPageActionResultWithContext(
	ctx SharedSessionBrowserMutationContext,
	req SharedSessionBrowserPageActionResultEventRequest,
) SharedSessionBrowserPageActionResultEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.PreferredTargetID = strings.TrimSpace(req.PreferredTargetID)
	req.URL = strings.TrimSpace(req.URL)
	req.Title = strings.TrimSpace(req.Title)
	req.Source = strings.TrimSpace(req.Source)
	req.Actor = strings.TrimSpace(req.Actor)

	result := sharedSessionBrowserPageActionResult(req.PreferredTargetID, req.Actor, req.Force, req.Review)
	if ctx.Registry == nil || req.SessionID == "" {
		return result
	}
	if ctx.usesWatchManagerEventSeam() {
		return ApplySharedSessionBrowserPageActionResultEvent(
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
			TabIndex:  req.TabIndex,
			URL:       req.URL,
			Title:     req.Title,
			Source:    req.Source,
		},
	)
	result.TargetID = firstNonEmptyString(result.TargetID, strings.TrimSpace(resolved.TargetID))
	return result
}

// ApplySharedSessionBrowserPageActionResult applies a page-action runtime
// result through the shared page-action contract.
func ApplySharedSessionBrowserPageActionResult(
	registry *BrowserSessionRegistry,
	req SharedSessionBrowserPageActionResultEventRequest,
) SharedSessionBrowserPageActionResultEventResult {
	return ApplySharedSessionBrowserPageActionResultWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		req,
	)
}
