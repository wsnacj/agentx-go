package tools

import (
	"fmt"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserActDialogWaitMs(requestedWaitMs int, defaultWaitMs int) int {
	if requestedWaitMs > 0 {
		return requestedWaitMs
	}
	return maxInt(defaultWaitMs, defaultBrowserWaitDownloadMs)
}

func browserActDialogReviewState(action string, force bool) (string, bool) {
	if strings.EqualFold(strings.TrimSpace(action), "accept") {
		return agentxbrowserruntime.SharedSessionBrowserActionReviewDecision("dialog", force), force
	}
	return "", false
}

func browserActActionRuntimeResultEventResultForOutcome(
	pageCtx browserActPageActionContext,
	resolverOutcome *agentxbrowserruntime.BrowserElementResolverOutcome,
	options browserActActionRuntimeEventOptions,
) agentxbrowserruntime.SharedSessionBrowserActionResultEventResult {
	if browserResolverOutcomeAllowsTargetTracking(resolverOutcome) {
		return browserActApplyActionRuntimeResultEvent(pageCtx, options)
	}
	return agentxbrowserruntime.SharedSessionBrowserActionResultEventResult{
		TargetID:       strings.TrimSpace(pageCtx.Target.TargetID),
		ReviewDecision: strings.TrimSpace(options.ReviewDecision),
		ReviewReady:    options.ReviewReady,
		Note:           strings.TrimSpace(options.Note),
	}
}

func browserActExecuteTraceStart(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	traceBackend, ok := pageCtx.RoutedBackend.(BrowserTraceActionBackend)
	if !ok {
		return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(pageCtx.RuntimeInfo.Backend), "trace_start")
	}
	waitMs := browserActTraceWaitMs(firstInt(params, "wait_ms"), pageCtx.Target)
	result, err := traceBackend.Trace(pageCtx.CallCtx, BrowserTraceRequest{
		BrowserApp:        pageCtx.BrowserApp,
		WaitMs:            waitMs,
		TabIndex:          pageCtx.TabIndex,
		PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
		Action:            "start",
	})
	if err != nil {
		return BrowserActResult{}, err
	}
	actionResult := browserActApplyActionRuntimeResultEvent(
		pageCtx,
		browserActActionRuntimeEventOptions{
			ResultBackend:     result.Backend,
			ResultBrowserApp:  result.BrowserApp,
			PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
			FinalURL:          strings.TrimSpace(result.FinalURL),
			Title:             strings.TrimSpace(result.Title),
			Source:            "browser_act_trace_start",
			SetCurrent:        pageCtx.TabIndex > 0,
			Note:              strings.TrimSpace(result.Note),
		},
	)
	return BrowserActResult{
		Kind:          "trace_start",
		Backend:       result.Backend,
		BrowserApp:    firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:        pageCtx.Target.Value,
		TargetID:      actionResult.TargetID,
		Profile:       pageCtx.RuntimeInfo.Profile,
		RuntimeTarget: pageCtx.RuntimeInfo.Target,
		FinalURL:      strings.TrimSpace(result.FinalURL),
		Title:         strings.TrimSpace(result.Title),
		Status:        firstNonEmpty(strings.TrimSpace(result.Status), "started"),
		TabIndex:      pageCtx.TabIndex,
		Note:          actionResult.Note,
	}, nil
}

func browserActExecuteDialog(pageCtx browserActPageActionContext, params map[string]any) (BrowserActResult, error) {
	dialogBackend, ok := pageCtx.RoutedBackend.(BrowserDialogActionBackend)
	if !ok {
		return BrowserActResult{}, fmt.Errorf("browser_act: backend %q does not support kind %q", strings.TrimSpace(pageCtx.RuntimeInfo.Backend), "dialog")
	}
	waitMs := browserActDialogWaitMs(firstInt(params, "wait_ms"), pageCtx.DefaultWaitMs)
	action, err := resolveBrowserDialogAction(params)
	if err != nil {
		return BrowserActResult{}, err
	}
	if action == "accept" && !pageCtx.Force {
		blocked := browserActReviewBlockedResult("dialog", pageCtx.RuntimeInfo, pageCtx.BrowserApp, pageCtx.Target, pageCtx.Force, agentxbrowserruntime.SharedSessionBrowserActionReviewReason("dialog", pageCtx.Force))
		blocked.Action = action
		return blocked, nil
	}
	promptText := strings.TrimSpace(firstString(params, "prompt_text", "prompt"))
	if action == "dismiss" {
		promptText = ""
	}
	reviewDecision, reviewReady := browserActDialogReviewState(action, pageCtx.Force)
	result, err := dialogBackend.Dialog(pageCtx.CallCtx, BrowserDialogRequest{
		BrowserApp:        pageCtx.BrowserApp,
		WaitMs:            waitMs,
		TabIndex:          pageCtx.TabIndex,
		PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
		ReviewDecision:    reviewDecision,
		ReviewReady:       reviewReady,
		Note:              agentxbrowserruntime.SharedSessionBrowserActionReviewReason("dialog", pageCtx.Force),
		Action:            action,
		PromptText:        promptText,
	})
	if err != nil {
		return BrowserActResult{}, err
	}
	actionResult := browserActApplyActionRuntimeResultEvent(
		pageCtx,
		browserActActionRuntimeEventOptions{
			ResultBackend:     result.Backend,
			ResultBrowserApp:  result.BrowserApp,
			PreferredTargetID: strings.TrimSpace(pageCtx.Target.TargetID),
			FinalURL:          strings.TrimSpace(result.FinalURL),
			Title:             strings.TrimSpace(result.Title),
			Source:            "browser_act_dialog",
			SetCurrent:        pageCtx.TabIndex > 0,
			ReviewDecision:    reviewDecision,
			ReviewReady:       reviewReady,
			Note:              firstNonEmpty(strings.TrimSpace(result.Note), agentxbrowserruntime.SharedSessionBrowserActionReviewReason("dialog", pageCtx.Force)),
		},
	)
	return BrowserActResult{
		Kind:           "dialog",
		Action:         action,
		Backend:        result.Backend,
		BrowserApp:     firstNonEmpty(result.BrowserApp, pageCtx.BrowserApp),
		Target:         pageCtx.Target.Value,
		TargetID:       actionResult.TargetID,
		Profile:        pageCtx.RuntimeInfo.Profile,
		RuntimeTarget:  pageCtx.RuntimeInfo.Target,
		FinalURL:       strings.TrimSpace(result.FinalURL),
		Title:          strings.TrimSpace(result.Title),
		Status:         firstNonEmpty(strings.TrimSpace(result.Status), "armed"),
		Force:          pageCtx.Force,
		ReviewDecision: actionResult.ReviewDecision,
		ReviewReady:    actionResult.ReviewReady,
		TabIndex:       pageCtx.TabIndex,
		Note:           actionResult.Note,
	}, nil
}
