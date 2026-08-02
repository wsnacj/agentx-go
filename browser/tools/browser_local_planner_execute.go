package tools

import (
	"strings"
	"time"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

const browserLocalPlannerExecuteWaitDownloadRetryMs = 30_000

func browserLocalPlannerExecuteForActResult(pageCtx browserActPageActionContext, currentArgs map[string]any, result BrowserActResult) BrowserActResult {
	if !pageCtx.Options.BrowserLocalPlannerExecute || result.BrowserLocalPlanner != nil {
		return result
	}
	eligibility := EvaluateBrowserLocalPlannerEligibilityForActResult(
		result.Kind,
		currentArgs,
		result,
		browserLocalPlannerPolicyConstraintsForExecution(pageCtx, currentArgs, result),
	)
	if !eligibility.Eligible || normalizeBrowserLocalPlannerToken(eligibility.FollowupKind) != "wait_download" {
		return result
	}

	if strings.TrimSpace(pageCtx.Options.BrowserLocalPlannerModel) != "" {
		if planned := browserLocalPlannerExecuteModelFollowup(pageCtx, currentArgs, result, eligibility); planned != nil {
			return *planned
		}
	}

	return browserLocalPlannerExecuteBuiltinWaitDownload(pageCtx, currentArgs, result, eligibility)
}

func browserLocalPlannerExecuteModelFollowup(pageCtx browserActPageActionContext, currentArgs map[string]any, result BrowserActResult, eligibility BrowserLocalPlannerEligibility) *BrowserActResult {
	plannerContext := BuildBrowserLocalPlannerContextForActResult(
		result.Kind,
		currentArgs,
		result,
		nil,
		browserLocalPlannerPolicyConstraintsForExecution(pageCtx, currentArgs, result),
	)
	decision, telemetry, err := invokeBrowserLocalPlanner(pageCtx.CallCtx, pageCtx.Options, eligibility, plannerContext)

	summary := &agentxbrowserruntime.BrowserLocalPlannerResultSummary{
		Mode:                   "execute",
		Eligible:               true,
		ReasonCode:             strings.TrimSpace(eligibility.ReasonCode),
		FollowupKind:           strings.TrimSpace(eligibility.FollowupKind),
		FailedCheck:            strings.TrimSpace(eligibility.FailedCheck),
		FailureReason:          strings.TrimSpace(eligibility.FailureReason),
		ManualRetryHint:        strings.TrimSpace(eligibility.ManualRetryHint),
		RecoveryAction:         strings.TrimSpace(eligibility.RecoveryAction),
		RequiresForce:          eligibility.RequiresForce,
		Decision:               strings.TrimSpace(telemetry.Decision),
		Model:                  firstNonEmpty(strings.TrimSpace(telemetry.Model), strings.TrimSpace(pageCtx.Options.BrowserLocalPlannerModel)),
		LatencyMs:              telemetry.LatencyMs,
		DiscardedInvalidOutput: telemetry.DiscardedInvalidOutput,
	}
	if err != nil {
		result.BrowserLocalPlanner = summary
		return &result
	}
	if decision.Action == nil {
		if strings.TrimSpace(summary.Decision) == "" {
			summary.Decision = decision.Decision
		}
		result.BrowserLocalPlanner = summary
		return &result
	}
	if decision.Decision != "retry_one_step" || normalizeBrowserLocalPlannerToken(decision.Action.Kind) != "wait_download" {
		if strings.TrimSpace(summary.Decision) == "" {
			summary.Decision = decision.Decision
		}
		summary.FollowupKind = decision.Action.Kind
		result.BrowserLocalPlanner = summary
		return &result
	}
	followupResult, ok := browserLocalPlannerExecuteWaitDownloadFollowup(pageCtx, currentArgs, decision.Action.Params, summary)
	if ok {
		return &followupResult
	}
	result.BrowserLocalPlanner = summary
	return &result
}

func browserLocalPlannerExecuteBuiltinWaitDownload(pageCtx browserActPageActionContext, currentArgs map[string]any, result BrowserActResult, eligibility BrowserLocalPlannerEligibility) BrowserActResult {
	summary := &agentxbrowserruntime.BrowserLocalPlannerResultSummary{
		Mode:            "execute",
		Eligible:        true,
		ReasonCode:      strings.TrimSpace(eligibility.ReasonCode),
		FollowupKind:    strings.TrimSpace(eligibility.FollowupKind),
		FailedCheck:     strings.TrimSpace(eligibility.FailedCheck),
		FailureReason:   strings.TrimSpace(eligibility.FailureReason),
		ManualRetryHint: strings.TrimSpace(eligibility.ManualRetryHint),
		RecoveryAction:  strings.TrimSpace(eligibility.RecoveryAction),
		RequiresForce:   eligibility.RequiresForce,
		Decision:        "retry_one_step",
		Model:           "builtin",
	}
	if normalizeBrowserLocalPlannerToken(eligibility.ReasonCode) != "wait_download_review_required" {
		result.BrowserLocalPlanner = summary
		return result
	}

	followupResult, ok := browserLocalPlannerExecuteWaitDownloadFollowup(pageCtx, currentArgs, nil, summary)
	if ok {
		return followupResult
	}
	result.BrowserLocalPlanner = summary
	return result
}

func browserLocalPlannerExecuteWaitDownloadFollowup(pageCtx browserActPageActionContext, currentArgs map[string]any, followupOverrides map[string]any, summary *agentxbrowserruntime.BrowserLocalPlannerResultSummary) (BrowserActResult, bool) {
	followupParams := browserUnifiedCloneParams(currentArgs)
	followupParams["kind"] = "wait_download"
	followupParams["force"] = true
	followupParams["wait_ms"] = browserLocalPlannerExecuteWaitDownloadRetryMsForArgs(currentArgs)
	followupParams["allow_recent_download_reuse"] = true
	for key, value := range followupOverrides {
		followupParams[key] = value
	}
	followupParams["kind"] = "wait_download"
	followupParams["force"] = true
	followupParams["allow_recent_download_reuse"] = true
	followupParams["wait_ms"] = browserLocalPlannerClampWaitDownloadRetryMs(firstInt(followupParams, "wait_ms", "timeout_ms"))
	followupCtx := pageCtx
	followupCtx.Force = true
	started := time.Now()
	followupResult, err := browserActExecuteWaitDownload(followupCtx, followupParams)
	if summary.LatencyMs <= 0 {
		summary.LatencyMs = time.Since(started).Milliseconds()
	}
	if err != nil {
		return BrowserActResult{}, false
	}

	summary.FollowupOK = strings.TrimSpace(followupResult.Status) != ""
	summary.FollowupRecovered = normalizeBrowserLocalPlannerToken(followupResult.Status) == "downloaded"
	followupResult.BrowserLocalPlanner = summary
	if summary.FollowupRecovered {
		return followupResult, true
	}
	return BrowserActResult{}, false
}

func browserLocalPlannerPolicyConstraintsForExecution(pageCtx browserActPageActionContext, currentArgs map[string]any, result BrowserActResult) BrowserLocalPlannerPolicyConstraints {
	policy := browserLocalPlannerPolicyConstraintsForDryRun(currentArgs, result)
	policy.AllowForceRetry = pageCtx.Options.BrowserLocalPlannerExecute || policy.AllowForceRetry
	return policy
}

func browserLocalPlannerExecuteWaitDownloadRetryMsForArgs(currentArgs map[string]any) int {
	return browserLocalPlannerClampWaitDownloadRetryMs(firstInt(currentArgs, "wait_ms", "timeout_ms"))
}

func browserLocalPlannerClampWaitDownloadRetryMs(waitMs int) int {
	if waitMs <= 0 {
		return browserLocalPlannerExecuteWaitDownloadRetryMs
	}
	if waitMs > browserLocalPlannerExecuteWaitDownloadRetryMs {
		return browserLocalPlannerExecuteWaitDownloadRetryMs
	}
	return waitMs
}
