package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserLocalPlannerSummaryForActResult(opts BrowserToolOptions, kind string, currentArgs map[string]any, result BrowserActResult) *agentxbrowserruntime.BrowserLocalPlannerResultSummary {
	if result.BrowserLocalPlanner != nil {
		return browserLocalPlannerCloneSummary(result.BrowserLocalPlanner)
	}
	if !opts.BrowserLocalPlannerDryRun {
		return nil
	}
	dryRun := BrowserLocalPlannerDryRunForActResult(
		kind,
		currentArgs,
		result,
		nil,
		browserLocalPlannerPolicyConstraintsForDryRun(currentArgs, result),
	)
	if !dryRun.Eligibility.Eligible &&
		strings.TrimSpace(dryRun.Eligibility.ReasonCode) == "" &&
		strings.TrimSpace(dryRun.Eligibility.FollowupKind) == "" &&
		strings.TrimSpace(dryRun.Eligibility.BlockedReason) == "" {
		return nil
	}
	return &agentxbrowserruntime.BrowserLocalPlannerResultSummary{
		Mode:            "dry_run",
		Eligible:        dryRun.Eligibility.Eligible,
		ReasonCode:      strings.TrimSpace(dryRun.Eligibility.ReasonCode),
		FollowupKind:    strings.TrimSpace(dryRun.Eligibility.FollowupKind),
		FailedCheck:     strings.TrimSpace(dryRun.Eligibility.FailedCheck),
		FailureReason:   strings.TrimSpace(dryRun.Eligibility.FailureReason),
		ManualRetryHint: strings.TrimSpace(dryRun.Eligibility.ManualRetryHint),
		RecoveryAction:  strings.TrimSpace(dryRun.Eligibility.RecoveryAction),
		RequiresForce:   dryRun.Eligibility.RequiresForce,
		BlockedReason:   strings.TrimSpace(dryRun.Eligibility.BlockedReason),
	}
}

func browserLocalPlannerPolicyConstraintsForDryRun(currentArgs map[string]any, result BrowserActResult) BrowserLocalPlannerPolicyConstraints {
	policy := BrowserLocalPlannerPolicyConstraints{
		RiskTier: "browser",
	}
	if firstBool(currentArgs, "force") || result.Force {
		policy.AllowForceRetry = true
	}
	return policy
}

func browserLocalPlannerCloneSummary(in *agentxbrowserruntime.BrowserLocalPlannerResultSummary) *agentxbrowserruntime.BrowserLocalPlannerResultSummary {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}
