package browserruntime

import (
	"fmt"
	"strings"
)

// SharedSessionBrowserRouteTarget is a route-scoped tracked target projection
// that can be surfaced by browser runtime inspection tools.
type SharedSessionBrowserRouteTarget struct {
	ID            string
	TabIndex      int
	URL           string
	Title         string
	BrowserApp    string
	Backend       string
	Profile       string
	RuntimeTarget string
	Current       bool
}

// SharedSessionBrowserRouteSnapshot is a route-scoped session snapshot with
// pending-review posture and follow-policy projection.
type SharedSessionBrowserRouteSnapshot struct {
	Backend                  string
	Profile                  string
	RuntimeTarget            string
	BrowserApp               string
	CurrentTargetID          string
	CurrentTargetSource      string
	PendingTargetReview      *BrowserSessionTargetReview
	PendingTargetReviewCount int
	FollowPolicyState        string
	FollowPolicyReason       string
	PopupPolicyState         string
	PopupPolicyReason        string
	Targets                  []SharedSessionBrowserRouteTarget
}

// SnapshotSharedSessionBrowserRoutes returns the filtered route snapshot for a
// session after pruning stale route state.
func SnapshotSharedSessionBrowserRoutes(registry *BrowserSessionRegistry, sessionID string, filter BrowserSessionRoute) []SharedSessionBrowserRouteSnapshot {
	if registry == nil {
		return nil
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}
	registry.PruneStaleRouteState(sessionID, filter)
	snapshot := registry.Snapshot(sessionID)
	if len(snapshot) == 0 {
		return nil
	}
	out := make([]SharedSessionBrowserRouteSnapshot, 0, len(snapshot))
	for _, routeState := range snapshot {
		if !browserSessionRouteMatchesFilter(routeState.Route, filter) {
			continue
		}
		followPolicyState, followPolicyReason := SharedSessionBrowserPendingTargetFollowPolicy(routeState.PendingTargetReview, routeState.PendingTargetReviewCount)
		popupPolicyState, popupPolicyReason := "", ""
		if SharedSessionBrowserPendingTargetReviewCategory(routeState.PendingTargetReview) != "redirect" {
			popupPolicyState, popupPolicyReason = SharedSessionBrowserPendingTargetReviewPolicy(routeState.PendingTargetReview, routeState.PendingTargetReviewCount)
		}
		targets := make([]SharedSessionBrowserRouteTarget, 0, len(routeState.Targets))
		for _, target := range routeState.Targets {
			targets = append(targets, SharedSessionBrowserRouteTarget{
				ID:            strings.TrimSpace(target.ID),
				TabIndex:      target.TabIndex,
				URL:           strings.TrimSpace(target.URL),
				Title:         strings.TrimSpace(target.Title),
				BrowserApp:    strings.TrimSpace(target.BrowserApp),
				Backend:       strings.TrimSpace(target.Backend),
				Profile:       strings.TrimSpace(target.Profile),
				RuntimeTarget: strings.TrimSpace(target.Target),
				Current:       strings.EqualFold(strings.TrimSpace(routeState.CurrentTargetID), strings.TrimSpace(target.ID)),
			})
		}
		out = append(out, SharedSessionBrowserRouteSnapshot{
			Backend:                  strings.TrimSpace(routeState.Route.Backend),
			Profile:                  strings.TrimSpace(routeState.Route.Profile),
			RuntimeTarget:            strings.TrimSpace(routeState.Route.Target),
			BrowserApp:               strings.TrimSpace(routeState.Route.BrowserApp),
			CurrentTargetID:          strings.TrimSpace(routeState.CurrentTargetID),
			CurrentTargetSource:      strings.TrimSpace(routeState.CurrentTargetSource),
			PendingTargetReview:      routeState.PendingTargetReview,
			PendingTargetReviewCount: routeState.PendingTargetReviewCount,
			FollowPolicyState:        followPolicyState,
			FollowPolicyReason:       followPolicyReason,
			PopupPolicyState:         popupPolicyState,
			PopupPolicyReason:        popupPolicyReason,
			Targets:                  targets,
		})
	}
	return out
}

// SharedSessionBrowserTargetReviewLabel returns a stable user-facing label for
// a pending target review.
func SharedSessionBrowserTargetReviewLabel(review *BrowserSessionTargetReview) string {
	if review == nil {
		return ""
	}
	label := strings.TrimSpace(review.Title)
	if label == "" {
		label = strings.TrimSpace(review.URL)
	}
	if label == "" && review.TabIndex > 0 {
		label = fmt.Sprintf("tab:%d", review.TabIndex)
	}
	return label
}

// SharedSessionBrowserPendingTargetReviewCategory classifies the pending review
// posture so follow-policy and confirmation rules can share the same owner.
func SharedSessionBrowserPendingTargetReviewCategory(review *BrowserSessionTargetReview) string {
	if review == nil {
		return ""
	}
	switch strings.TrimSpace(review.Decision) {
	case "session_target_redirect_review_required", "session_target_redirect_review_confirmed":
		return "redirect"
	default:
		return "popup"
	}
}

// SharedSessionBrowserPendingTargetReviewPolicy projects the pending popup
// review posture for sessions/workbench payloads.
func SharedSessionBrowserPendingTargetReviewPolicy(review *BrowserSessionTargetReview, count int) (string, string) {
	if review == nil {
		return "", ""
	}
	if count <= 0 {
		count = 1
	}
	label := SharedSessionBrowserTargetReviewLabel(review)
	if count >= 2 {
		return "popup_storm_review_required", fmt.Sprintf("browser session accumulated %d pending popup targets on this route; close or explicitly force a target before following or adopting popup target %q", count, label)
	}
	return "popup_review_required", fmt.Sprintf("browser session has a pending popup target %q; rerun with force=true before following or adopting it", label)
}

// SharedSessionBrowserPendingTargetFollowPolicy projects the route-level
// follow-policy posture for the current pending review state.
func SharedSessionBrowserPendingTargetFollowPolicy(review *BrowserSessionTargetReview, count int) (string, string) {
	if review == nil {
		return "auto_follow_allowed", ""
	}
	label := SharedSessionBrowserTargetReviewLabel(review)
	if SharedSessionBrowserPendingTargetReviewCategory(review) == "redirect" {
		return "redirect_review_required", fmt.Sprintf("browser session has a redirected target %q pending confirmation; rerun with force=true before auto-following or adopting it", label)
	}
	return SharedSessionBrowserPendingTargetReviewPolicy(review, count)
}
