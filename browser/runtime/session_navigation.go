package browserruntime

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// SharedSessionBrowserNavigationResultEventRequest carries the route-scoped
// runtime result from a navigation action so redirect review and resolved-target
// tracking can share one browserruntime-owned contract.
type SharedSessionBrowserNavigationResultEventRequest struct {
	SessionID        string
	Route            BrowserSessionRoute
	ExplicitTargetID string
	TabIndex         int
	RequestedURL     string
	FinalURL         string
	Title            string
	Source           string
	Force            bool
	PriorSelection   *BrowserSessionTargetSelection
	Note             string
}

// SharedSessionBrowserNavigationResultEventResult captures the tracked target
// together with any redirect-review posture produced by a navigation result.
type SharedSessionBrowserNavigationResultEventResult struct {
	TargetID          string
	Review            *BrowserSessionTargetReview
	RestoredSelection *BrowserSessionTargetSelection
	ReviewRequired    bool
	ReviewDecision    string
	ReviewReady       bool
	Note              string
}

func sharedSessionBrowserNormalizedURLOrigin(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == nil {
		return ""
	}
	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if scheme == "" || host == "" {
		return ""
	}
	port := strings.TrimSpace(parsed.Port())
	if port == "" {
		switch scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		}
	}
	return scheme + "://" + host + ":" + port
}

// SharedSessionBrowserNavigationNeedsRedirectReview reports whether a
// navigation crossed origin and therefore requires explicit redirect review.
func SharedSessionBrowserNavigationNeedsRedirectReview(requestedURL string, finalURL string) bool {
	requestedOrigin := sharedSessionBrowserNormalizedURLOrigin(requestedURL)
	finalOrigin := sharedSessionBrowserNormalizedURLOrigin(finalURL)
	if requestedOrigin == "" || finalOrigin == "" {
		return false
	}
	return requestedOrigin != finalOrigin
}

// SharedSessionBrowserNavigationReviewDecision returns the tool-facing review
// token used when a navigation crosses origin.
func SharedSessionBrowserNavigationReviewDecision(force bool) string {
	if force {
		return "navigate_redirect_review_confirmed"
	}
	return "navigate_redirect_review_required"
}

// SharedSessionBrowserTargetRedirectReviewDecision returns the route-scoped
// pending-review token used when a redirected target must be confirmed before
// auto-follow/adoption.
func SharedSessionBrowserTargetRedirectReviewDecision(force bool) string {
	if force {
		return "session_target_redirect_review_confirmed"
	}
	return "session_target_redirect_review_required"
}

// SharedSessionBrowserNavigationReviewReason returns the user-facing reason for
// a redirect-review posture.
func SharedSessionBrowserNavigationReviewReason(requestedURL string, finalURL string, force bool) string {
	if force {
		return fmt.Sprintf("browser navigation redirect review acknowledged via force=true; confirm the final origin %q is intended before continuing", strings.TrimSpace(finalURL))
	}
	return fmt.Sprintf("browser navigation redirected across origin from %q to %q; rerun with force=true after review", strings.TrimSpace(requestedURL), strings.TrimSpace(finalURL))
}

// ApplySharedSessionBrowserNavigationResultEvent applies a navigation runtime
// result through the shared resolved-target seam and projects any redirect
// review back onto the returned runtime-result contract.
func (m SharedSessionBrowserObserverManager) ApplyNavigationResultEvent(
	req SharedSessionBrowserNavigationResultEventRequest,
) SharedSessionBrowserNavigationResultEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.ExplicitTargetID = strings.TrimSpace(req.ExplicitTargetID)
	req.RequestedURL = strings.TrimSpace(req.RequestedURL)
	req.FinalURL = strings.TrimSpace(req.FinalURL)
	req.Title = strings.TrimSpace(req.Title)
	req.Source = strings.TrimSpace(req.Source)
	req.Note = strings.TrimSpace(req.Note)

	redirectReview := SharedSessionBrowserNavigationNeedsRedirectReview(req.RequestedURL, req.FinalURL)
	reviewReason := SharedSessionBrowserNavigationReviewReason(req.RequestedURL, req.FinalURL, req.Force)
	result := SharedSessionBrowserNavigationResultEventResult{
		ReviewRequired: redirectReview && !req.Force,
		ReviewReady:    redirectReview && req.Force,
	}
	if redirectReview {
		result.ReviewDecision = SharedSessionBrowserNavigationReviewDecision(req.Force)
		result.Note = firstNonEmptyString(req.Note, reviewReason)
	} else {
		result.Note = req.Note
	}
	if m.SessionRegistry == nil || req.SessionID == "" {
		return result
	}

	resolved := m.ApplyResolvedTargetEvent(
		SharedSessionBrowserResolvedTargetEventRequest{
			SessionID:             req.SessionID,
			Route:                 req.Route,
			ExplicitTargetID:      req.ExplicitTargetID,
			TabIndex:              req.TabIndex,
			URL:                   req.FinalURL,
			Title:                 req.Title,
			Source:                req.Source,
			PendingReview:         redirectReview && !req.Force,
			PendingReviewDecision: SharedSessionBrowserTargetRedirectReviewDecision(req.Force),
			PendingReviewReason:   reviewReason,
			PriorSelection:        req.PriorSelection,
		},
	)
	result.TargetID = strings.TrimSpace(resolved.TargetID)
	result.Review = resolved.Review
	result.RestoredSelection = resolved.RestoredSelection
	return result
}

// ApplySharedSessionBrowserNavigationResultEvent applies a navigation runtime
// result through the shared resolved-target seam and projects any redirect
// review back onto the returned runtime-result contract.
func ApplySharedSessionBrowserNavigationResultEvent(
	sessionRegistry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	req SharedSessionBrowserNavigationResultEventRequest,
	reconnectWindow time.Duration,
) SharedSessionBrowserNavigationResultEventResult {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ApplyNavigationResultEvent(req)
}

// ApplySharedSessionBrowserNavigationResultWithContext applies a navigation
// runtime result and routes the write through the top-level mutation seam when
// manager dependencies are available.
func ApplySharedSessionBrowserNavigationResultWithContext(
	ctx SharedSessionBrowserMutationContext,
	req SharedSessionBrowserNavigationResultEventRequest,
) SharedSessionBrowserNavigationResultEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.ExplicitTargetID = strings.TrimSpace(req.ExplicitTargetID)
	req.RequestedURL = strings.TrimSpace(req.RequestedURL)
	req.FinalURL = strings.TrimSpace(req.FinalURL)
	req.Title = strings.TrimSpace(req.Title)
	req.Source = strings.TrimSpace(req.Source)
	req.Note = strings.TrimSpace(req.Note)

	redirectReview := SharedSessionBrowserNavigationNeedsRedirectReview(req.RequestedURL, req.FinalURL)
	reviewReason := SharedSessionBrowserNavigationReviewReason(req.RequestedURL, req.FinalURL, req.Force)
	result := SharedSessionBrowserNavigationResultEventResult{
		ReviewRequired: redirectReview && !req.Force,
		ReviewReady:    redirectReview && req.Force,
	}
	if redirectReview {
		result.ReviewDecision = SharedSessionBrowserNavigationReviewDecision(req.Force)
		result.Note = firstNonEmptyString(req.Note, reviewReason)
	} else {
		result.Note = req.Note
	}
	if ctx.Registry == nil || req.SessionID == "" {
		return result
	}
	if ctx.usesWatchManagerEventSeam() {
		return ApplySharedSessionBrowserNavigationResultEvent(
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
			SessionID:             req.SessionID,
			Route:                 req.Route,
			ExplicitTargetID:      req.ExplicitTargetID,
			TabIndex:              req.TabIndex,
			URL:                   req.FinalURL,
			Title:                 req.Title,
			Source:                req.Source,
			PendingReview:         redirectReview && !req.Force,
			PendingReviewDecision: SharedSessionBrowserTargetRedirectReviewDecision(req.Force),
			PendingReviewReason:   reviewReason,
			PriorSelection:        req.PriorSelection,
		},
	)
	result.TargetID = strings.TrimSpace(resolved.TargetID)
	result.Review = resolved.Review
	result.RestoredSelection = resolved.RestoredSelection
	return result
}

// ApplySharedSessionBrowserNavigationResult applies a navigation runtime result
// through the shared browserruntime contract.
func ApplySharedSessionBrowserNavigationResult(
	registry *BrowserSessionRegistry,
	req SharedSessionBrowserNavigationResultEventRequest,
) SharedSessionBrowserNavigationResultEventResult {
	return ApplySharedSessionBrowserNavigationResultWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		req,
	)
}
