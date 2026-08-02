package browserruntime

import (
	"fmt"
	"strings"
	"time"
)

// SharedSessionBrowserTabRememberReviewRequest carries the route-scoped tab
// lifecycle inputs used to gate remember_target adoption after list/focus/close
// style tab actions.
type SharedSessionBrowserTabRememberReviewRequest struct {
	Registry            *BrowserSessionRegistry
	RunRegistry         SharedSessionRunRegistry
	StateRegistry       SharedSessionBrowserStateRegistry
	ReconnectWindow     time.Duration
	SessionID           string
	Route               BrowserSessionRoute
	Action              string
	Force               bool
	RememberTarget      bool
	CandidateTargetID   string
	RequestedTabIndex   int
	ActiveIndex         int
	PriorActiveTargetID string
	PriorSelection      *BrowserSessionTargetSelection
	Tabs                []BrowserTab
}

// SharedSessionBrowserTabRememberReviewResult captures the shared remember
// review posture for a tab lifecycle action together with the target candidate
// that should be remembered when no popup-review gate blocks adoption.
type SharedSessionBrowserTabRememberReviewResult struct {
	RememberTargetID string
	RememberTabIndex int
	Decision         string
	Ready            bool
	Note             string
}

// ApplySharedSessionBrowserTabsResultWithContext applies a route-scoped tab
// action result and routes the write through the top-level mutation seam when
// manager dependencies are available.
func ApplySharedSessionBrowserTabsResultWithContext(
	ctx SharedSessionBrowserMutationContext,
	req SharedSessionBrowserTabsResultEventRequest,
) SharedSessionBrowserTabsResultEventResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.Action = strings.TrimSpace(req.Action)
	req.ExplicitTargetID = strings.TrimSpace(req.ExplicitTargetID)
	req.PriorActiveTargetID = strings.TrimSpace(req.PriorActiveTargetID)
	req.PriorRequestedTargetID = strings.TrimSpace(req.PriorRequestedTargetID)
	req.Actor = strings.TrimSpace(req.Actor)
	req.Note = strings.TrimSpace(req.Note)
	if req.PriorSelection != nil {
		priorSelectionCopy := *req.PriorSelection
		req.PriorSelection = &priorSelectionCopy
	}

	result := SharedSessionBrowserTabsResultEventResult{
		TargetID: req.ExplicitTargetID,
		Tabs:     append([]BrowserTab(nil), req.Tabs...),
	}
	if req.Review.Review != nil {
		result.ReviewDecision = SharedSessionBrowserPendingTargetReviewDecision(req.Review, req.Force)
		result.ReviewReady = req.Force
		result.Note = firstNonEmptyString(
			req.Note,
			SharedSessionBrowserPendingTargetReviewReason(
				firstNonEmptyString(req.Actor, "browser tabs action"),
				req.Review,
				req.Force,
			),
		)
	} else {
		result.Note = req.Note
	}
	if ctx.Registry == nil || req.SessionID == "" {
		return result
	}
	if ctx.usesWatchManagerEventSeam() {
		return ApplySharedSessionBrowserTabsResultEvent(
			ctx.Registry,
			ctx.RunRegistry,
			ctx.StateRegistry,
			req,
			ctx.ReconnectWindow,
		)
	}

	priorSelection := req.PriorSelection
	if priorSelection == nil {
		priorSelection = SnapshotSharedSessionBrowserCurrentTargetSelection(ctx.Registry, req.SessionID, req.Route)
	}
	priorActiveTargetID := req.PriorActiveTargetID
	if priorActiveTargetID == "" && req.ActiveIndex > 0 {
		priorActiveTargetID = ResolveSharedSessionBrowserTabTargetID(ctx.Registry, req.SessionID, req.Route, req.ActiveIndex)
	}
	priorRequestedTargetID := req.PriorRequestedTargetID
	if priorRequestedTargetID == "" && req.RequestedTabIndex > 0 {
		priorRequestedTargetID = ResolveSharedSessionBrowserTabTargetID(ctx.Registry, req.SessionID, req.Route, req.RequestedTabIndex)
	}

	result.Tabs = SyncSharedSessionBrowserTabsForRouteWithContext(
		ctx,
		req.SessionID,
		req.Route,
		req.ActiveIndex,
		req.Tabs,
	)

	result.TargetID = firstNonEmptyString(result.TargetID, priorRequestedTargetID)
	switch req.Action {
	case "list":
		if result.TargetID == "" && req.ActiveIndex > 0 {
			result.TargetID = ResolveSharedSessionBrowserTabTargetID(ctx.Registry, req.SessionID, req.Route, req.ActiveIndex)
		}
	case "focus":
		if result.TargetID == "" && req.RequestedTabIndex > 0 {
			result.TargetID = ResolveSharedSessionBrowserTabTargetID(ctx.Registry, req.SessionID, req.Route, req.RequestedTabIndex)
		}
		if result.TargetID == "" && req.ActiveIndex > 0 {
			result.TargetID = ResolveSharedSessionBrowserTabTargetID(ctx.Registry, req.SessionID, req.Route, req.ActiveIndex)
		}
	case "close":
		if result.TargetID == "" {
			result.TargetID = priorRequestedTargetID
		}
	}

	result.RememberReview = ApplySharedSessionBrowserTabRememberReview(SharedSessionBrowserTabRememberReviewRequest{
		Registry:            ctx.Registry,
		RunRegistry:         ctx.RunRegistry,
		StateRegistry:       ctx.StateRegistry,
		ReconnectWindow:     ctx.ReconnectWindow,
		SessionID:           req.SessionID,
		Route:               req.Route,
		Action:              req.Action,
		Force:               req.Force,
		RememberTarget:      req.RememberTarget,
		CandidateTargetID:   result.TargetID,
		RequestedTabIndex:   req.RequestedTabIndex,
		ActiveIndex:         req.ActiveIndex,
		PriorActiveTargetID: priorActiveTargetID,
		PriorSelection:      priorSelection,
		Tabs:                result.Tabs,
	})

	if req.Review.Review != nil {
		return result
	}
	if strings.TrimSpace(result.RememberReview.Decision) != "" {
		result.Note = firstNonEmptyString(req.Note, strings.TrimSpace(result.RememberReview.Note))
		return result
	}
	result.Note = req.Note

	return result
}

// ApplySharedSessionBrowserTabsResult applies a route-scoped tab action result
// through the shared tab lifecycle contract.
func ApplySharedSessionBrowserTabsResult(
	registry *BrowserSessionRegistry,
	req SharedSessionBrowserTabsResultEventRequest,
) SharedSessionBrowserTabsResultEventResult {
	return ApplySharedSessionBrowserTabsResultWithContext(
		SharedSessionBrowserMutationContext{Registry: registry},
		req,
	)
}

// ApplySharedSessionBrowserTabRememberReview applies the shared popup-review
// guardrail for tab lifecycle actions and restores prior selection state when a
// newly surfaced active tab should not yet become the remembered target.
func ApplySharedSessionBrowserTabRememberReview(req SharedSessionBrowserTabRememberReviewRequest) SharedSessionBrowserTabRememberReviewResult {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.Route = normalizeBrowserSessionRoute(req.Route)
	req.Action = strings.TrimSpace(req.Action)
	req.CandidateTargetID = strings.TrimSpace(req.CandidateTargetID)
	req.PriorActiveTargetID = strings.TrimSpace(req.PriorActiveTargetID)

	result := SharedSessionBrowserTabRememberReviewResult{}
	if req.Registry == nil || req.SessionID == "" {
		return result
	}
	if req.ReconnectWindow > 0 && (req.RunRegistry != nil || req.StateRegistry != nil) {
		return ApplySharedSessionBrowserTabRememberReviewEvent(
			req.Registry,
			req.RunRegistry,
			req.StateRegistry,
			req,
			req.ReconnectWindow,
		)
	}

	if (req.Action == "list" || req.Action == "close") && req.ActiveIndex > 0 {
		activeReview := SharedSessionBrowserPendingTargetReviewStateForTarget(req.Registry, req.SessionID, req.Route, "", req.ActiveIndex)
		if activeReview.Review != nil && !(req.RememberTarget && req.Force) {
			RestoreSharedSessionBrowserCurrentTargetSelection(req.Registry, req.SessionID, req.Route, req.PriorSelection, "popup_review_restore")
		}
	}

	switch req.Action {
	case "list", "close":
		result.RememberTabIndex = req.ActiveIndex
	case "focus":
		if req.ActiveIndex > 0 {
			result.RememberTabIndex = req.ActiveIndex
		} else {
			result.RememberTabIndex = req.RequestedTabIndex
		}
	}
	if result.RememberTabIndex <= 0 {
		result.RememberTargetID = req.CandidateTargetID
	}

	popupReview := req.RememberTarget && sharedSessionBrowserRememberTargetNeedsPopupReview(req.Action, result.RememberTabIndex, req.PriorActiveTargetID)
	if !popupReview {
		return result
	}

	activeTab := sharedSessionBrowserTabByIndex(req.Tabs, result.RememberTabIndex)
	result.Decision = SharedSessionBrowserRememberTargetPopupReviewDecision(req.Force)
	result.Ready = req.Force
	result.Note = SharedSessionBrowserRememberTargetPopupReviewReason(activeTab, req.Force)
	if req.Force {
		return result
	}

	RestoreSharedSessionBrowserCurrentTargetSelection(req.Registry, req.SessionID, req.Route, req.PriorSelection, "popup_review_restore")
	RecordSharedSessionBrowserPendingTargetPopupReview(
		req.Registry,
		req.SessionID,
		req.Route,
		activeTab,
		result.Decision,
		result.Note,
	)
	return result
}

// SharedSessionBrowserRememberTargetPopupReviewDecision returns the decision
// token used when tab adoption requires popup review.
func SharedSessionBrowserRememberTargetPopupReviewDecision(force bool) string {
	if force {
		return "session_target_popup_review_confirmed"
	}
	return "session_target_popup_review_required"
}

// SharedSessionBrowserRememberTargetPopupReviewReason returns the user-facing
// reason for a remember_target popup review posture.
func SharedSessionBrowserRememberTargetPopupReviewReason(activeTab BrowserTab, force bool) string {
	label := strings.TrimSpace(activeTab.Title)
	if label == "" {
		label = strings.TrimSpace(activeTab.URL)
	}
	if label == "" && activeTab.Index > 0 {
		label = fmt.Sprintf("tab:%d", activeTab.Index)
	}
	if force {
		return fmt.Sprintf("browser tab adoption review acknowledged via force=true; confirm the newly opened active tab %q is intended before continuing", label)
	}
	return fmt.Sprintf("browser tab listing surfaced a newly opened active tab %q; rerun with force=true before adopting it as the remembered session target", label)
}

func sharedSessionBrowserRememberTargetNeedsPopupReview(action string, activeTabIndex int, priorActiveTargetID string) bool {
	if activeTabIndex <= 0 || strings.TrimSpace(priorActiveTargetID) != "" {
		return false
	}
	switch strings.TrimSpace(action) {
	case "list", "close":
		return true
	default:
		return false
	}
}

func sharedSessionBrowserTabByIndex(tabs []BrowserTab, index int) BrowserTab {
	for _, tab := range tabs {
		if tab.Index == index {
			return tab
		}
	}
	return BrowserTab{Index: index}
}
