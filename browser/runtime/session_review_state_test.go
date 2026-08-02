package browserruntime

import "testing"

func TestSelectSharedSessionBrowserReviewStatePrefersExplicitDecision(t *testing.T) {
	state := SelectSharedSessionBrowserReviewState(
		SharedSessionBrowserReviewStateRequest{
			Candidates: []SharedSessionBrowserReviewDecisionCandidate{
				{Decision: " session_target_popup_review_required ", Ready: false},
				{Decision: "session_target_redirect_review_required", Ready: true},
			},
			Routes: []SharedSessionBrowserRouteCoordinationInput{
				{FollowPolicyState: "redirect_review_required"},
			},
		},
	)

	if state.Decision != "session_target_popup_review_required" || state.Ready {
		t.Fatalf("expected explicit review state to win, got %#v", state)
	}
}

func TestSelectSharedSessionBrowserReviewStateFallsBackToRoutes(t *testing.T) {
	redirect := SelectSharedSessionBrowserReviewState(
		SharedSessionBrowserReviewStateRequest{
			Routes: []SharedSessionBrowserRouteCoordinationInput{
				{FollowPolicyState: "redirect_review_required"},
			},
		},
	)
	if redirect.Decision != "session_target_redirect_review_required" || redirect.Ready {
		t.Fatalf("expected redirect follow-policy fallback, got %#v", redirect)
	}

	popup := SelectSharedSessionBrowserReviewState(
		SharedSessionBrowserReviewStateRequest{
			Routes: []SharedSessionBrowserRouteCoordinationInput{
				{FollowPolicyState: "popup_review_required"},
			},
		},
	)
	if popup.Decision != "session_target_popup_review_required" || popup.Ready {
		t.Fatalf("expected popup follow-policy fallback, got %#v", popup)
	}
}

func TestSelectSharedSessionBrowserReviewStateIgnoresPopupStormFallback(t *testing.T) {
	state := SelectSharedSessionBrowserReviewState(
		SharedSessionBrowserReviewStateRequest{
			Routes: []SharedSessionBrowserRouteCoordinationInput{
				{FollowPolicyState: "popup_storm_review_required"},
			},
		},
	)

	if state.Decision != "" || state.Ready {
		t.Fatalf("expected popup-storm posture not to synthesize top-level review decision, got %#v", state)
	}
}
