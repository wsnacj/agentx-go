package browserruntime

import (
	"strings"
	"time"
)

func ResolveSharedSessionBrowserProfileStatus(input SharedSessionBrowserHealthInput, selectedInfo BrowserRuntimeInfo, requestedProfile string, fallback BrowserProfileStatusResult, reconnectWindow time.Duration) BrowserProfileStatusResult {
	if SharedSessionBrowserHasExplicitProfileLifecycleObservation(fallback) {
		return fallback
	}
	input = ApplySharedSessionBrowserHealthSummary(input, SharedSessionBrowserHealthSummaryFromBrowserSessionHealth(fallback.SessionHealth))
	requestedProfile = firstNonEmptyString(
		strings.TrimSpace(requestedProfile),
		strings.TrimSpace(fallback.Profile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	return SharedSessionBrowserProfileRecoveryStatus(
		EvaluateSharedSessionBrowserHealthForInputScope(input, selectedInfo, requestedProfile, reconnectWindow),
		selectedInfo,
		requestedProfile,
		fallback,
	)
}

func ResolveSharedSessionBrowserProfileStatusForScope(registry SharedSessionBrowserStateRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, input SharedSessionBrowserHealthInput, fallback BrowserProfileStatusResult, reconnectWindow time.Duration) BrowserProfileStatusResult {
	if SharedSessionBrowserHasExplicitProfileLifecycleObservation(fallback) {
		return fallback
	}
	input = ApplySharedSessionBrowserHealthSummary(input, SharedSessionBrowserHealthSummaryFromBrowserSessionHealth(fallback.SessionHealth))
	requestedProfile = firstNonEmptyString(
		strings.TrimSpace(requestedProfile),
		strings.TrimSpace(fallback.Profile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	return SharedSessionBrowserProfileRecoveryStatus(
		EvaluateSharedSessionBrowserHealthForScope(registry, sessionID, selectedInfo, requestedProfile, input, reconnectWindow),
		selectedInfo,
		requestedProfile,
		fallback,
	)
}

func AssessSharedSessionBrowserProfileRecoveryForScope(registry SharedSessionBrowserStateRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, input SharedSessionBrowserHealthInput, fallback BrowserProfileStatusResult, reconnectWindow time.Duration) (SharedSessionBrowserHealthEvaluation, SharedSessionBrowserProfileRecoveryAssessment) {
	input = ApplySharedSessionBrowserHealthSummary(input, SharedSessionBrowserHealthSummaryFromBrowserSessionHealth(fallback.SessionHealth))
	requestedProfile = firstNonEmptyString(
		strings.TrimSpace(requestedProfile),
		strings.TrimSpace(fallback.Profile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	evaluation := EvaluateSharedSessionBrowserHealthForScope(
		registry,
		sessionID,
		selectedInfo,
		requestedProfile,
		input,
		reconnectWindow,
	)
	return evaluation, AssessSharedSessionBrowserProfileRecovery(evaluation, selectedInfo, requestedProfile, fallback)
}

func AssessSharedSessionBrowserProfileRecoveryForInputScope(input SharedSessionBrowserHealthInput, selectedInfo BrowserRuntimeInfo, requestedProfile string, fallback BrowserProfileStatusResult, reconnectWindow time.Duration) (SharedSessionBrowserHealthEvaluation, SharedSessionBrowserProfileRecoveryAssessment) {
	input = ApplySharedSessionBrowserHealthSummary(input, SharedSessionBrowserHealthSummaryFromBrowserSessionHealth(fallback.SessionHealth))
	requestedProfile = firstNonEmptyString(
		strings.TrimSpace(requestedProfile),
		strings.TrimSpace(fallback.Profile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	evaluation := EvaluateSharedSessionBrowserHealthForInputScope(input, selectedInfo, requestedProfile, reconnectWindow)
	return evaluation, AssessSharedSessionBrowserProfileRecovery(evaluation, selectedInfo, requestedProfile, fallback)
}

func SharedSessionBrowserProfileRecoveryStatus(evaluation SharedSessionBrowserHealthEvaluation, selectedInfo BrowserRuntimeInfo, profile string, fallback BrowserProfileStatusResult) BrowserProfileStatusResult {
	if evaluation.HasProfile {
		return SharedSessionBrowserProfileStatusResultFromState(evaluation.Profile, selectedInfo, profile)
	}
	return fallback
}

func AssessSharedSessionBrowserProfileRecovery(evaluation SharedSessionBrowserHealthEvaluation, selectedInfo BrowserRuntimeInfo, profile string, fallback BrowserProfileStatusResult) SharedSessionBrowserProfileRecoveryAssessment {
	effectiveStatus := SharedSessionBrowserProfileRecoveryStatus(evaluation, selectedInfo, profile, fallback)
	assessment := SharedSessionBrowserProfileRecoveryAssessment{
		EffectiveStatus:          effectiveStatus,
		ShouldStopBeforeRecovery: sharedSessionBrowserStatusRequiresStopBeforeRecovery(effectiveStatus),
	}
	if SharedSessionBrowserHasExplicitProfileLifecycleObservation(fallback) && !sharedSessionBrowserStatusRequiresStopBeforeRecovery(fallback) {
		assessment.ShouldStopBeforeRecovery = false
	}
	if evaluation.Summary != nil {
		if sharedSessionBrowserHealthBlocksLifecycleRecovery(evaluation.Summary) {
			assessment.NeedsRefreshRecovery = false
			assessment.ShouldStopBeforeRecovery = false
			return assessment
		}
		if strings.EqualFold(strings.TrimSpace(evaluation.Summary.RecoveryAction), "browser action=refresh") {
			assessment.NeedsRefreshRecovery = true
		}
		if strings.EqualFold(strings.TrimSpace(evaluation.Summary.State), "profile_reconnecting") &&
			!evaluation.ReconnectTimedOut && evaluation.HasProfile {
			assessment.ReconnectInProgress = true
			assessment.HasSyntheticStatus = true
			assessment.SyntheticStatus = SharedSessionBrowserProfileStatusResultFromState(evaluation.Profile, selectedInfo, profile)
		}
	}
	if sharedSessionBrowserStatusNeedsRefreshRecovery(effectiveStatus) {
		assessment.NeedsRefreshRecovery = true
	}
	return assessment
}

func SharedSessionBrowserExecutionBlockedDecision(evaluation SharedSessionBrowserHealthEvaluation) (string, bool) {
	if evaluation.Summary == nil || !sharedSessionBrowserHealthBlocksLifecycleRecovery(evaluation.Summary) {
		return "", false
	}
	return strings.TrimSpace(evaluation.Summary.State), true
}

// SharedSessionBrowserProfileStatusResultFromState converts a shared session
// lifecycle snapshot into the stable profile-status payload surface.
func SharedSessionBrowserProfileStatusResultFromState(state SharedSessionBrowserProfileState, selectedInfo BrowserRuntimeInfo, profile string) BrowserProfileStatusResult {
	return BrowserProfileStatusResult{
		Backend:    firstNonEmptyString(strings.TrimSpace(state.Backend), strings.TrimSpace(selectedInfo.Backend)),
		BrowserApp: strings.TrimSpace(state.BrowserApp),
		Profile:    firstNonEmptyString(strings.TrimSpace(state.Profile), strings.TrimSpace(profile), strings.TrimSpace(selectedInfo.Profile)),
		Status:     strings.TrimSpace(state.Status),
		Running:    state.Running,
		Connected:  state.Connected,
		Note:       strings.TrimSpace(state.Note),
	}
}

func sharedSessionBrowserStatusNeedsRefreshRecovery(status BrowserProfileStatusResult) bool {
	switch strings.ToLower(strings.TrimSpace(status.Status)) {
	case "disconnected", "crashed":
		return true
	default:
		return false
	}
}

func sharedSessionBrowserStatusRequiresStopBeforeRecovery(status BrowserProfileStatusResult) bool {
	state := strings.ToLower(strings.TrimSpace(status.Status))
	switch state {
	case "stopped", "starting", "deleted", "delete_requested", "teardown_stopped", "teardown_already_stopped":
		return false
	}
	return status.Running || status.Connected || state != ""
}

func sharedSessionBrowserHealthBlocksLifecycleRecovery(summary *SharedSessionBrowserHealthSummary) bool {
	if summary == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(summary.State)) {
	case "cooldown_active", "restart_pending", "restart_failed_permanent":
		return true
	default:
		return false
	}
}
