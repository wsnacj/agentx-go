package tools

import (
	"context"
	"strconv"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserActTabsContext struct {
	CallCtx                       context.Context
	RoutedBackend                 BrowserBackend
	SessionRegistry               *BrowserSessionRegistry
	SharedMutationCtx             agentxbrowserruntime.SharedSessionBrowserMutationContext
	SessionID                     string
	RuntimeInfo                   BrowserRuntimeInfo
	HiddenImplicitHostDefaultBase bool
	Target                        browserToolTarget
	BrowserApp                    string
	Force                         bool
	TabIndex                      int
}

type browserActTabsReviewState struct {
	Review browserPendingTargetReviewState
}

type browserActTabsEventApplyOptions struct {
	ResultBackend          string
	ResultBrowserApp       string
	Action                 string
	RequestedTabIndex      int
	ActiveIndex            int
	Tabs                   []BrowserTab
	ExplicitTargetID       string
	PriorSelection         *agentxbrowserruntime.BrowserSessionTargetSelection
	PriorRequestedTargetID string
	Force                  bool
	RememberTarget         bool
	Review                 browserPendingTargetReviewState
	Actor                  string
	Note                   string
}

type browserActTabsResultOptions struct {
	Kind             string
	Action           string
	ResultBackend    string
	BrowserApp       string
	Target           browserToolTarget
	TargetID         string
	RuntimeInfo      BrowserRuntimeInfo
	Status           string
	Force            bool
	ReviewDecision   string
	ReviewReady      bool
	TabIndex         int
	Tabs             []BrowserTab
	ActiveIndex      int
	RememberDecision string
	RememberReady    bool
	Note             string
}

func browserActTabsAction(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "list_tabs":
		return "list"
	case "focus_tab":
		return "focus"
	case "close_tab":
		return "close"
	default:
		return ""
	}
}

func browserActTabsWaitMs(requestedWaitMs int) int {
	if requestedWaitMs > 0 {
		return requestedWaitMs
	}
	return 200
}

func resolveBrowserActTabsReviewState(
	ctx context.Context,
	sessionRegistry *BrowserSessionRegistry,
	runtimeInfo BrowserRuntimeInfo,
	browserApp string,
	target browserToolTarget,
	action string,
) browserActTabsReviewState {
	review := browserPendingTargetReviewState{}
	if strings.EqualFold(strings.TrimSpace(action), "focus") {
		review = browserPendingTargetReviewStateForToolTarget(
			ctx,
			sessionRegistry,
			browserSessionRoute(runtimeInfo, browserApp, ""),
			target,
		)
	}
	return browserActTabsReviewState{Review: review}
}

func browserActTabsReviewBlockedResult(kind string, action string, runtimeInfo BrowserRuntimeInfo, browserApp string, target browserToolTarget, review browserPendingTargetReviewState, force bool, note string) BrowserActResult {
	return BrowserActResult{
		Kind:           kind,
		Action:         action,
		Backend:        strings.TrimSpace(runtimeInfo.Backend),
		BrowserApp:     strings.TrimSpace(browserApp),
		Target:         target.Value,
		TargetID:       strings.TrimSpace(target.TargetID),
		Profile:        runtimeInfo.Profile,
		RuntimeTarget:  runtimeInfo.Target,
		Status:         "review_required",
		Force:          force,
		ReviewDecision: browserPendingTargetReviewDecisionWithState(review, force),
		ReviewReady:    false,
		TabIndex:       target.TabIndex,
		Note:           strings.TrimSpace(note),
	}
}

func applyBrowserActTabsEventResult(
	sharedMutationCtx agentxbrowserruntime.SharedSessionBrowserMutationContext,
	sessionID string,
	runtimeInfo BrowserRuntimeInfo,
	defaultBrowserApp string,
	options browserActTabsEventApplyOptions,
) agentxbrowserruntime.SharedSessionBrowserTabsResultEventResult {
	return agentxbrowserruntime.ApplySharedSessionBrowserTabsResultWithContext(
		sharedMutationCtx,
		agentxbrowserruntime.SharedSessionBrowserTabsResultEventRequest{
			SessionID:              sessionID,
			Route:                  browserSessionRoute(runtimeInfo, firstNonEmpty(strings.TrimSpace(options.ResultBrowserApp), defaultBrowserApp), strings.TrimSpace(options.ResultBackend)),
			Action:                 options.Action,
			RequestedTabIndex:      options.RequestedTabIndex,
			ActiveIndex:            options.ActiveIndex,
			Tabs:                   options.Tabs,
			ExplicitTargetID:       options.ExplicitTargetID,
			PriorSelection:         options.PriorSelection,
			PriorRequestedTargetID: options.PriorRequestedTargetID,
			Force:                  options.Force,
			RememberTarget:         options.RememberTarget,
			Review:                 options.Review,
			Actor:                  options.Actor,
			Note:                   strings.TrimSpace(options.Note),
		},
	)
}

func browserActTabsResult(options browserActTabsResultOptions) BrowserActResult {
	return BrowserActResult{
		Kind:             options.Kind,
		Action:           options.Action,
		Backend:          options.ResultBackend,
		BrowserApp:       options.BrowserApp,
		Target:           options.Target.Value,
		TargetID:         options.TargetID,
		Profile:          options.RuntimeInfo.Profile,
		RuntimeTarget:    options.RuntimeInfo.Target,
		Status:           options.Status,
		Force:            options.Force,
		ReviewDecision:   options.ReviewDecision,
		ReviewReady:      options.ReviewReady,
		TabIndex:         options.TabIndex,
		Tabs:             options.Tabs,
		ActiveIndex:      options.ActiveIndex,
		RememberDecision: options.RememberDecision,
		RememberReady:    options.RememberReady,
		Note:             options.Note,
	}
}

func browserActExecuteTabs(tabsCtx browserActTabsContext, params map[string]any, kind string) (BrowserActResult, error) {
	action := browserActTabsAction(kind)
	if action == "" {
		return BrowserActResult{}, browserInvalidArgumentError("browser_act", []string{"kind"}, "browser_act: unsupported tabs kind "+strconv.Quote(kind))
	}
	if action != "list" && tabsCtx.TabIndex <= 0 {
		return BrowserActResult{}, browserMissingRequiredArgumentError("browser_act", []string{"target", "tab_index"}, "browser_act: target or tab_index is required for kind "+strings.TrimSpace(kind))
	}
	if err := browserImplicitLegacyHostTabsActionFallbackError("browser_act kind list_tabs", tabsCtx.HiddenImplicitHostDefaultBase, tabsCtx.RuntimeInfo, action, tabsCtx.Target); err != nil {
		return BrowserActResult{}, err
	}
	waitMs := browserActTabsWaitMs(firstInt(params, "wait_ms"))
	tabsReviewState := resolveBrowserActTabsReviewState(tabsCtx.CallCtx, tabsCtx.SessionRegistry, tabsCtx.RuntimeInfo, tabsCtx.BrowserApp, tabsCtx.Target, action)
	focusReview := tabsReviewState.Review
	if focusReview.Review != nil && !tabsCtx.Force {
		return browserActTabsReviewBlockedResult(
			kind,
			action,
			tabsCtx.RuntimeInfo,
			tabsCtx.BrowserApp,
			tabsCtx.Target,
			focusReview,
			tabsCtx.Force,
			browserPendingTargetReviewReasonWithState("browser_act focus_tab", focusReview, tabsCtx.Force),
		), nil
	}
	rememberTarget := firstBool(params, "remember_target", "remember")
	tabsActor := "browser_act " + kind
	explicitTargetID := strings.TrimSpace(tabsCtx.Target.TargetID)
	priorSelection := browserCurrentTargetSelectionSnapshotForRoute(tabsCtx.CallCtx, tabsCtx.SessionRegistry, tabsCtx.RuntimeInfo, tabsCtx.BrowserApp, "")
	priorRequestedTargetID := browserTargetIDForTab(tabsCtx.CallCtx, tabsCtx.SessionRegistry, tabsCtx.RuntimeInfo, tabsCtx.BrowserApp, tabsCtx.TabIndex)
	result, err := tabsCtx.RoutedBackend.Tabs(tabsCtx.CallCtx, BrowserTabsRequest{
		BrowserApp:             tabsCtx.BrowserApp,
		Action:                 action,
		TabIndex:               tabsCtx.TabIndex,
		WaitMs:                 waitMs,
		Force:                  tabsCtx.Force,
		RememberTarget:         rememberTarget,
		Review:                 focusReview,
		Actor:                  tabsActor,
		ExplicitTargetID:       explicitTargetID,
		PriorSelection:         priorSelection,
		PriorRequestedTargetID: priorRequestedTargetID,
	})
	if err != nil {
		return BrowserActResult{}, err
	}
	resolvedBrowserApp := firstNonEmpty(result.BrowserApp, tabsCtx.BrowserApp)
	tabsResult := applyBrowserActTabsEventResult(tabsCtx.SharedMutationCtx, tabsCtx.SessionID, tabsCtx.RuntimeInfo, tabsCtx.BrowserApp, browserActTabsEventApplyOptions{
		ResultBackend:          result.Backend,
		ResultBrowserApp:       resolvedBrowserApp,
		Action:                 action,
		RequestedTabIndex:      tabsCtx.TabIndex,
		ActiveIndex:            result.ActiveIndex,
		Tabs:                   result.Tabs,
		ExplicitTargetID:       explicitTargetID,
		PriorSelection:         priorSelection,
		PriorRequestedTargetID: priorRequestedTargetID,
		Force:                  tabsCtx.Force,
		RememberTarget:         rememberTarget,
		Review:                 focusReview,
		Actor:                  tabsActor,
		Note:                   result.Note,
	})
	rememberReview := tabsResult.RememberReview
	return browserActTabsResult(browserActTabsResultOptions{
		Kind:             kind,
		Action:           firstNonEmpty(result.Action, action),
		ResultBackend:    result.Backend,
		BrowserApp:       resolvedBrowserApp,
		Target:           tabsCtx.Target,
		TargetID:         strings.TrimSpace(tabsResult.TargetID),
		RuntimeInfo:      tabsCtx.RuntimeInfo,
		Status:           result.Status,
		Force:            tabsCtx.Force,
		ReviewDecision:   tabsResult.ReviewDecision,
		ReviewReady:      tabsResult.ReviewReady,
		TabIndex:         tabsCtx.TabIndex,
		Tabs:             tabsResult.Tabs,
		ActiveIndex:      result.ActiveIndex,
		RememberDecision: strings.TrimSpace(rememberReview.Decision),
		RememberReady:    rememberReview.Ready,
		Note:             tabsResult.Note,
	}), nil
}
