package tools

import (
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

// BrowserLocalPlannerEligibility describes whether a browser-local planner is
// allowed to consider a follow-up for the current low-risk browser seam.
type BrowserLocalPlannerEligibility struct {
	Eligible        bool   `json:"eligible,omitempty"`
	ReasonCode      string `json:"reason_code,omitempty"`
	FollowupKind    string `json:"followup_kind,omitempty"`
	FailedCheck     string `json:"failed_check,omitempty"`
	FailureReason   string `json:"failure_reason,omitempty"`
	ManualRetryHint string `json:"manual_retry_hint,omitempty"`
	RecoveryAction  string `json:"recovery_action,omitempty"`
	RequiresForce   bool   `json:"requires_force,omitempty"`
	BlockedReason   string `json:"blocked_reason,omitempty"`
}

// BrowserLocalPlannerDryRun combines seam eligibility with the compressed
// browser-local context that a future constrained planner invocation would use.
type BrowserLocalPlannerDryRun struct {
	Eligibility BrowserLocalPlannerEligibility `json:"eligibility,omitempty"`
	Context     *BrowserLocalPlannerContext    `json:"context,omitempty"`
}

// EvaluateBrowserLocalPlannerEligibilityForActResult determines whether a
// browser action result falls into an allowlisted browser-local planning seam.
func EvaluateBrowserLocalPlannerEligibilityForActResult(kind string, currentArgs map[string]any, result BrowserActResult, policy BrowserLocalPlannerPolicyConstraints) BrowserLocalPlannerEligibility {
	action := browserLocalPlannerResolveActionKind(kind, result.Kind)
	status := normalizeBrowserLocalPlannerToken(result.Status)
	actionability := browserLocalPlannerResultActionabilitySignal(result)
	manualRetryHint := browserLocalPlannerResultManualRetryHint(result)
	recoveryAction := browserLocalPlannerResultRecoveryAction(result)

	if !browserLocalPlannerResultHasReview(result) {
		if eligibility := browserLocalPlannerEligibilityForActionabilityFailure(action, actionability, manualRetryHint, recoveryAction); eligibility.ReasonCode != "" {
			return eligibility
		}
	}

	switch action {
	case "click":
		if browserLocalPlannerResultHasReview(result) {
			return BrowserLocalPlannerEligibility{}
		}
		if browserLocalPlannerResultResolverStatus(result) != "unresolved" && status != "unresolved" {
			return BrowserLocalPlannerEligibility{}
		}
		switch manualRetryHint {
		case "refresh_snapshot":
			return BrowserLocalPlannerEligibility{
				Eligible:        true,
				ReasonCode:      "click_unresolved_refresh_snapshot",
				FollowupKind:    "snapshot",
				ManualRetryHint: manualRetryHint,
				RecoveryAction:  recoveryAction,
			}
		case "add_specificity", "add_ordinal":
			return BrowserLocalPlannerEligibility{
				Eligible:        true,
				ReasonCode:      "click_unresolved_retry_click",
				FollowupKind:    "click",
				ManualRetryHint: manualRetryHint,
				RecoveryAction:  recoveryAction,
			}
		}
		if recoveryAction == "browser action=snapshot" {
			return BrowserLocalPlannerEligibility{
				Eligible:       true,
				ReasonCode:     "click_unresolved_refresh_snapshot",
				FollowupKind:   "snapshot",
				RecoveryAction: recoveryAction,
			}
		}
	case "wait_download":
		switch {
		case status == "review_required" && strings.EqualFold(strings.TrimSpace(result.ReviewDecision), "wait_download_review_required"):
			eligibility := BrowserLocalPlannerEligibility{
				ReasonCode:      "wait_download_review_required",
				FollowupKind:    "wait_download",
				ManualRetryHint: firstNonEmpty(manualRetryHint, "rerun_with_force"),
				RecoveryAction:  recoveryAction,
				RequiresForce:   true,
			}
			if policy.AllowForceRetry {
				eligibility.Eligible = true
			} else {
				eligibility.BlockedReason = "force_retry_not_allowed"
			}
			return eligibility
		case status == "timeout" || browserLocalPlannerResultTimedOut(result):
			return BrowserLocalPlannerEligibility{
				Eligible:        true,
				ReasonCode:      "wait_download_timeout",
				FollowupKind:    "wait_download",
				ManualRetryHint: manualRetryHint,
				RecoveryAction:  recoveryAction,
			}
		}
	}

	_ = currentArgs
	return BrowserLocalPlannerEligibility{}
}

// BuildBrowserLocalPlannerContextForActResult compresses a browser action
// result and surrounding browser-local facts into planner input context.
func BuildBrowserLocalPlannerContextForActResult(kind string, currentArgs map[string]any, result BrowserActResult, recentRuns []BrowserLocalPlannerRecentToolRun, policy BrowserLocalPlannerPolicyConstraints) BrowserLocalPlannerContext {
	action := browserLocalPlannerResolveActionKind(kind, result.Kind)
	context := BrowserLocalPlannerContext{
		Tool:           "browser",
		Action:         action,
		CurrentArgs:    browserLocalPlannerCloneParams(currentArgs),
		CurrentResult:  browserLocalPlannerCurrentResultMap(result),
		Page:           BrowserLocalPlannerPage{URL: strings.TrimSpace(result.FinalURL), Title: strings.TrimSpace(result.Title)},
		RecentToolRuns: browserLocalPlannerRecentRuns(recentRuns),
		LatestSnapshot: browserLocalPlannerSnapshotSummaryFromActResult(result),
		Policy:         policy,
	}
	return context
}

// BrowserLocalPlannerDryRunForActResult calculates whether a low-risk
// browser-local planner seam exists and returns the compressed context that a
// future behind-flag planner invocation would consume.
func BrowserLocalPlannerDryRunForActResult(kind string, currentArgs map[string]any, result BrowserActResult, recentRuns []BrowserLocalPlannerRecentToolRun, policy BrowserLocalPlannerPolicyConstraints) BrowserLocalPlannerDryRun {
	eligibility := EvaluateBrowserLocalPlannerEligibilityForActResult(kind, currentArgs, result, policy)
	dryRun := BrowserLocalPlannerDryRun{Eligibility: eligibility}
	if eligibility.Eligible {
		context := BuildBrowserLocalPlannerContextForActResult(kind, currentArgs, result, recentRuns, policy)
		dryRun.Context = &context
	}
	return dryRun
}

func browserLocalPlannerResolveActionKind(kind string, fallback string) string {
	action := normalizeBrowserLocalPlannerToken(kind)
	if action == "" {
		action = normalizeBrowserLocalPlannerToken(fallback)
	}
	return action
}

func browserLocalPlannerResultHasReview(result BrowserActResult) bool {
	return strings.EqualFold(strings.TrimSpace(result.Status), "review_required") ||
		strings.TrimSpace(result.ReviewDecision) != ""
}

func browserLocalPlannerResultResolverStatus(result BrowserActResult) string {
	if result.ResolverOutcome != nil {
		return normalizeBrowserLocalPlannerToken(result.ResolverOutcome.Status)
	}
	return ""
}

func browserLocalPlannerResultManualRetryHint(result BrowserActResult) string {
	if value := normalizeBrowserLocalPlannerToken(result.ResolverManualRetryHint); value != "" {
		return value
	}
	if result.ResolverOutcome != nil {
		if value := normalizeBrowserLocalPlannerToken(result.ResolverOutcome.ManualRetryHint); value != "" {
			return value
		}
	}
	if value := browserLocalPlannerResultActionabilitySignal(result).ManualRetryHint; value != "" {
		return value
	}
	if strings.EqualFold(strings.TrimSpace(result.Status), "review_required") &&
		strings.EqualFold(strings.TrimSpace(result.ReviewDecision), "wait_download_review_required") {
		return "rerun_with_force"
	}
	return ""
}

func browserLocalPlannerResultRecoveryAction(result BrowserActResult) string {
	if value := strings.TrimSpace(result.RecoveryAction); value != "" {
		return value
	}
	if result.FailureEvidence != nil {
		if value := strings.TrimSpace(result.FailureEvidence.RecoveryAction); value != "" {
			return value
		}
	}
	if value := browserLocalPlannerResultActionabilitySignal(result).RecoveryAction; value != "" {
		return value
	}
	if result.ResolverOutcome != nil {
		return strings.TrimSpace(result.ResolverOutcome.RecoveryAction)
	}
	return ""
}

type browserLocalPlannerActionabilitySignal struct {
	Status          string
	FailedCheck     string
	FailureReason   string
	ManualRetryHint string
	RecoveryAction  string
}

func browserLocalPlannerResultActionabilitySignal(result BrowserActResult) browserLocalPlannerActionabilitySignal {
	report := result.Actionability
	if report == nil && result.FailureEvidence != nil {
		report = result.FailureEvidence.Actionability
	}
	if report == nil {
		return browserLocalPlannerActionabilitySignal{}
	}
	return browserLocalPlannerActionabilitySignal{
		Status:          normalizeBrowserLocalPlannerToken(report.Status),
		FailedCheck:     normalizeBrowserLocalPlannerToken(report.FailedCheck),
		FailureReason:   normalizeBrowserLocalPlannerToken(report.FailureReason),
		ManualRetryHint: normalizeBrowserLocalPlannerToken(report.ManualRetryHint),
		RecoveryAction:  strings.TrimSpace(report.RecoveryAction),
	}
}

func browserLocalPlannerEligibilityForActionabilityFailure(action string, signal browserLocalPlannerActionabilitySignal, manualRetryHint string, recoveryAction string) BrowserLocalPlannerEligibility {
	if signal.Status != agentxbrowserruntime.BrowserActionabilityStatusFailed || signal.FailedCheck == "" || signal.FailedCheck == "resolve_target" {
		return BrowserLocalPlannerEligibility{}
	}
	followupKind := browserLocalPlannerActionabilityFollowupKind(signal, recoveryAction)
	if followupKind == "" {
		return BrowserLocalPlannerEligibility{}
	}
	action = normalizeBrowserLocalPlannerToken(action)
	if action == "" {
		action = "browser"
	}
	return BrowserLocalPlannerEligibility{
		Eligible:        true,
		ReasonCode:      action + "_actionability_" + signal.FailedCheck + "_" + followupKind,
		FollowupKind:    followupKind,
		FailedCheck:     signal.FailedCheck,
		FailureReason:   signal.FailureReason,
		ManualRetryHint: firstNonEmpty(signal.ManualRetryHint, manualRetryHint),
		RecoveryAction:  firstNonEmpty(signal.RecoveryAction, recoveryAction),
	}
}

func browserLocalPlannerActionabilityFollowupKind(signal browserLocalPlannerActionabilitySignal, recoveryAction string) string {
	switch signal.FailedCheck {
	case "stable", "navigation_wait":
		return "wait"
	case "attached", "visible", "receives_events", "frame_hit_target", "enabled", "editable":
		return "snapshot"
	}
	return browserLocalPlannerFollowupKindFromRecoveryAction(firstNonEmpty(signal.RecoveryAction, recoveryAction))
}

func browserLocalPlannerFollowupKindFromRecoveryAction(recoveryAction string) string {
	value := strings.ToLower(strings.TrimSpace(recoveryAction))
	value = strings.TrimPrefix(value, "browser action=")
	value = strings.TrimPrefix(value, "browser ")
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.Contains(value, " ") {
		value = strings.Fields(value)[0]
	}
	if browserLocalPlannerAllowedActionKinds[value] {
		return value
	}
	return ""
}

func browserLocalPlannerResultTimedOut(result BrowserActResult) bool {
	note := strings.ToLower(strings.TrimSpace(result.Note))
	return strings.Contains(note, "timed out") || strings.Contains(note, "timeout")
}

func browserLocalPlannerCloneParams(params map[string]any) map[string]any {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]any, len(params))
	for key, value := range params {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" || value == nil {
			continue
		}
		out[trimmed] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func browserLocalPlannerCurrentResultMap(result BrowserActResult) map[string]any {
	out := map[string]any{}
	if value := normalizeBrowserLocalPlannerToken(result.Status); value != "" {
		out["status"] = value
	}
	if value := strings.TrimSpace(result.Note); value != "" {
		out["note"] = value
	}
	if value := strings.TrimSpace(result.ReviewDecision); value != "" {
		out["review_decision"] = value
		out["review_ready"] = result.ReviewReady
	}
	if value := browserLocalPlannerResultManualRetryHint(result); value != "" {
		out["manual_retry_hint"] = value
	}
	if value := browserLocalPlannerResultRecoveryAction(result); value != "" {
		out["recovery_action"] = value
	}
	actionability := browserLocalPlannerResultActionabilitySignal(result)
	if value := actionability.Status; value != "" {
		out["actionability_status"] = value
	}
	if value := actionability.FailedCheck; value != "" {
		out["failed_check"] = value
	}
	if value := actionability.FailureReason; value != "" {
		out["failure_reason"] = value
	}
	if result.FailureEvidence != nil {
		if value := normalizeBrowserLocalPlannerToken(result.FailureEvidence.ReasonCode); value != "" {
			out["failure_reason_code"] = value
		}
	}
	if value := browserLocalPlannerResultResolverStatus(result); value != "" {
		out["resolver_status"] = value
	}
	if result.Force {
		out["force"] = true
	}
	if result.TabIndex > 0 {
		out["tab_index"] = result.TabIndex
	}
	if value := strings.TrimSpace(result.TargetID); value != "" {
		out["target_id"] = value
	}
	if result.Snapshot != "" || len(result.Elements) > 0 {
		out["snapshot_available"] = true
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func browserLocalPlannerRecentRuns(items []BrowserLocalPlannerRecentToolRun) []BrowserLocalPlannerRecentToolRun {
	if len(items) == 0 {
		return nil
	}
	out := make([]BrowserLocalPlannerRecentToolRun, 0, len(items))
	for _, item := range items {
		normalized := BrowserLocalPlannerRecentToolRun{
			Action:          normalizeBrowserLocalPlannerToken(item.Action),
			Status:          normalizeBrowserLocalPlannerToken(item.Status),
			SummaryCode:     normalizeBrowserLocalPlannerToken(item.SummaryCode),
			ManualRetryHint: normalizeBrowserLocalPlannerToken(item.ManualRetryHint),
		}
		if normalized.Action == "" && normalized.Status == "" && normalized.SummaryCode == "" && normalized.ManualRetryHint == "" {
			continue
		}
		out = append(out, normalized)
		if len(out) >= 5 {
			break
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func browserLocalPlannerSnapshotSummaryFromActResult(result BrowserActResult) *BrowserLocalPlannerSnapshotSummary {
	if strings.TrimSpace(result.Snapshot) == "" && len(result.Elements) == 0 {
		return nil
	}
	summary := &BrowserLocalPlannerSnapshotSummary{
		Format: normalizeBrowserLocalPlannerToken(result.SnapshotFormat),
	}
	summary.TopClickableLabels = browserLocalPlannerSnapshotTopLabels(result.Elements)
	summary.TopRefs = browserLocalPlannerSnapshotTopRefs(result.Elements)
	if summary.Format == "" && strings.TrimSpace(result.Snapshot) != "" {
		summary.Format = "ai"
	}
	if summary.Format == "" && len(summary.TopClickableLabels) == 0 && len(summary.TopRefs) == 0 {
		return nil
	}
	return summary
}

func browserLocalPlannerSnapshotTopLabels(elements []BrowserSnapshotElement) []string {
	if len(elements) == 0 {
		return nil
	}
	clickable := make([]string, 0, 6)
	fallback := make([]string, 0, 6)
	seen := map[string]bool{}
	for _, element := range elements {
		label := strings.TrimSpace(element.Label)
		if label == "" || seen[label] {
			continue
		}
		seen[label] = true
		if browserLocalPlannerElementLooksClickable(element) {
			clickable = append(clickable, label)
		} else {
			fallback = append(fallback, label)
		}
		if len(clickable) >= 6 {
			break
		}
	}
	labels := clickable
	if len(labels) == 0 {
		labels = fallback
	}
	if len(labels) > 6 {
		labels = labels[:6]
	}
	if len(labels) == 0 {
		return nil
	}
	return labels
}

func browserLocalPlannerSnapshotTopRefs(elements []BrowserSnapshotElement) []string {
	if len(elements) == 0 {
		return nil
	}
	refs := make([]string, 0, 6)
	seen := map[string]bool{}
	for _, element := range elements {
		ref := strings.TrimSpace(element.Ref)
		if ref == "" {
			ref = strings.TrimSpace(element.Selector)
		}
		if ref == "" || seen[ref] {
			continue
		}
		seen[ref] = true
		refs = append(refs, ref)
		if len(refs) >= 6 {
			break
		}
	}
	if len(refs) == 0 {
		return nil
	}
	return refs
}

func browserLocalPlannerElementLooksClickable(element BrowserSnapshotElement) bool {
	role := normalizeBrowserLocalPlannerToken(element.Role)
	tag := normalizeBrowserLocalPlannerToken(element.Tag)
	typ := normalizeBrowserLocalPlannerToken(element.Type)
	switch {
	case role == "link" || role == "button":
		return true
	case tag == "a" || tag == "button":
		return true
	case typ == "button" || typ == "submit":
		return true
	default:
		return false
	}
}
