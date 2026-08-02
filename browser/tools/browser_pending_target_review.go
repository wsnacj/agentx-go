package tools

import (
	"context"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserPendingTargetReviewState = agentxbrowserruntime.SharedSessionBrowserPendingTargetReviewState

func browserTargetlessImplicitLegacyHostFallback(hiddenImplicitHostDefaultBase bool, target browserToolTarget) bool {
	return browserTargetlessImplicitLegacyHostCurrentPageFallback(hiddenImplicitHostDefaultBase, target, "")
}

func browserPendingTargetReviewStateForToolTarget(ctx context.Context, registry *BrowserSessionRegistry, route BrowserSessionRoute, target browserToolTarget) browserPendingTargetReviewState {
	if registry == nil {
		return browserPendingTargetReviewState{}
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return browserPendingTargetReviewState{}
	}
	return agentxbrowserruntime.SharedSessionBrowserPendingTargetReviewStateForTarget(registry, sessionID, route, target.TargetID, target.TabIndex)
}

func browserPendingTargetReviewForToolTarget(ctx context.Context, registry *BrowserSessionRegistry, route BrowserSessionRoute, target browserToolTarget) *agentxbrowserruntime.BrowserSessionTargetReview {
	return browserPendingTargetReviewStateForToolTarget(ctx, registry, route, target).Review
}

func browserPendingTargetReviewForRuntimeTarget(ctx context.Context, registry *BrowserSessionRegistry, runtimeInfo BrowserRuntimeInfo, browserApp string, target browserToolTarget) *agentxbrowserruntime.BrowserSessionTargetReview {
	return browserPendingTargetReviewForToolTarget(ctx, registry, browserSessionRoute(runtimeInfo, browserApp, ""), target)
}

func browserPendingTargetReviewStateForRuntimeTarget(ctx context.Context, registry *BrowserSessionRegistry, runtimeInfo BrowserRuntimeInfo, browserApp string, target browserToolTarget) browserPendingTargetReviewState {
	return browserPendingTargetReviewStateForToolTarget(ctx, registry, browserSessionRoute(runtimeInfo, browserApp, ""), target)
}

func browserPendingTargetReviewStateForRoute(ctx context.Context, registry *BrowserSessionRegistry, route BrowserSessionRoute) browserPendingTargetReviewState {
	if registry == nil {
		return browserPendingTargetReviewState{}
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return browserPendingTargetReviewState{}
	}
	return agentxbrowserruntime.SharedSessionBrowserPendingTargetReviewStateForRoute(registry, sessionID, route)
}

func browserAutoFollowPendingTargetReviewStateForRuntimeTarget(ctx context.Context, registry *BrowserSessionRegistry, runtimeInfo BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, browserApp string, target browserToolTarget) browserPendingTargetReviewState {
	if browserTargetlessImplicitLegacyHostFallback(hiddenImplicitHostDefaultBase, target) {
		return browserPendingTargetReviewState{}
	}
	if registry == nil {
		return browserPendingTargetReviewState{}
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return browserPendingTargetReviewState{}
	}
	return agentxbrowserruntime.SharedSessionBrowserAutoFollowPendingTargetReviewState(
		registry,
		sessionID,
		browserSessionRoute(runtimeInfo, browserApp, ""),
		target.TargetID,
		target.TabIndex,
	)
}

func browserPendingTargetReviewLabel(review *agentxbrowserruntime.BrowserSessionTargetReview) string {
	return agentxbrowserruntime.SharedSessionBrowserTargetReviewLabel(review)
}

func browserPendingTargetReviewCategory(review *agentxbrowserruntime.BrowserSessionTargetReview) string {
	return agentxbrowserruntime.SharedSessionBrowserPendingTargetReviewCategory(review)
}

func browserPendingTargetReviewDecisionWithState(state browserPendingTargetReviewState, force bool) string {
	return agentxbrowserruntime.SharedSessionBrowserPendingTargetReviewDecision(state, force)
}

func browserPendingTargetReviewPolicy(review *agentxbrowserruntime.BrowserSessionTargetReview, count int) (string, string) {
	return agentxbrowserruntime.SharedSessionBrowserPendingTargetReviewPolicy(review, count)
}

func browserPendingTargetFollowPolicy(state browserPendingTargetReviewState) (string, string) {
	return agentxbrowserruntime.SharedSessionBrowserPendingTargetFollowPolicy(state.Review, state.Count)
}

func browserPendingTargetReviewReason(actor string, review *agentxbrowserruntime.BrowserSessionTargetReview, force bool) string {
	return browserPendingTargetReviewReasonWithState(actor, browserPendingTargetReviewState{Review: review}, force)
}

func browserPendingTargetReviewReasonWithState(actor string, state browserPendingTargetReviewState, force bool) string {
	return agentxbrowserruntime.SharedSessionBrowserPendingTargetReviewReason(actor, state, force)
}
