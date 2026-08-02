package browserruntime

import "strings"

// SharedSessionBrowserReviewDecisionCandidate captures one candidate review
// decision/ready posture in priority order.
type SharedSessionBrowserReviewDecisionCandidate struct {
	Decision string
	Ready    bool
}

// SharedSessionBrowserReviewState captures the selected top-level review
// posture after explicit decisions and route-scoped policy blockers are
// considered.
type SharedSessionBrowserReviewState struct {
	Decision string
	Ready    bool
}

// SharedSessionBrowserReviewStateRequest carries the prioritized explicit
// review decisions plus optional route-scoped follow-policy posture used to
// derive the top-level review state.
type SharedSessionBrowserReviewStateRequest struct {
	Candidates []SharedSessionBrowserReviewDecisionCandidate
	Routes     []SharedSessionBrowserRouteCoordinationInput
}

// SelectSharedSessionBrowserReviewState chooses the highest-priority explicit
// review posture, or falls back to route-scoped popup/redirect blockers when
// no explicit decision is present.
func SelectSharedSessionBrowserReviewState(
	req SharedSessionBrowserReviewStateRequest,
) SharedSessionBrowserReviewState {
	for _, candidate := range req.Candidates {
		decision := strings.TrimSpace(candidate.Decision)
		if decision == "" {
			continue
		}
		return SharedSessionBrowserReviewState{
			Decision: decision,
			Ready:    candidate.Ready,
		}
	}
	switch SharedSessionBrowserSelectedFollowPolicyState(req.Routes) {
	case "redirect_review_required":
		return SharedSessionBrowserReviewState{
			Decision: SharedSessionBrowserTargetRedirectReviewDecision(false),
			Ready:    false,
		}
	case "popup_review_required":
		return SharedSessionBrowserReviewState{
			Decision: SharedSessionBrowserRememberTargetPopupReviewDecision(false),
			Ready:    false,
		}
	default:
		return SharedSessionBrowserReviewState{}
	}
}
