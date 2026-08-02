package browserruntime

import (
	"fmt"
	"strings"
	"time"
)

func EvaluateSharedSessionBrowserHealth(input SharedSessionBrowserHealthInput, reconnectWindow time.Duration) SharedSessionBrowserHealthEvaluation {
	input.ActiveNodeRunID = strings.TrimSpace(input.ActiveNodeRunID)
	input.StoredState = strings.TrimSpace(input.StoredState)
	input.StoredReason = strings.TrimSpace(input.StoredReason)
	input.StoredRecoveryAction = strings.TrimSpace(input.StoredRecoveryAction)
	input.StoredReconnectHint = strings.TrimSpace(input.StoredReconnectHint)
	input.StoredLastRestartResult = strings.TrimSpace(input.StoredLastRestartResult)
	input.StoredLastRestartError = strings.TrimSpace(input.StoredLastRestartError)
	input.StoredResolverBlockedBy = strings.TrimSpace(input.StoredResolverBlockedBy)
	input.StoredAmbiguityClass = strings.TrimSpace(input.StoredAmbiguityClass)
	input.StoredCandidateKind = strings.TrimSpace(input.StoredCandidateKind)
	input.StoredCandidateStrength = strings.TrimSpace(input.StoredCandidateStrength)
	input.StoredRetryDisposition = strings.TrimSpace(input.StoredRetryDisposition)
	input.StoredManualRetryHint = strings.TrimSpace(input.StoredManualRetryHint)
	input.StoredNextStepAlias = strings.TrimSpace(input.StoredNextStepAlias)
	if input.StoredState == "" &&
		input.StoredReason == "" &&
		input.StoredRecoveryAction == "" &&
		input.StoredReconnectHint == "" &&
		input.StoredDisconnectCount == 0 &&
		input.StoredDisconnectBurstCount == 0 &&
		input.StoredDisconnectBurstWindowMs == 0 &&
		input.StoredCooldownRemainingMs == 0 &&
		input.StoredRetryBackoffRemainingMs == 0 &&
		input.StoredRestartAttemptCount == 0 &&
		input.StoredRestartFailureCount == 0 &&
		input.StoredLastDisconnectUnixMilli == 0 &&
		input.StoredLastReconnectUnixMilli == 0 &&
		input.StoredLastRestartAttemptUnixMilli == 0 &&
		input.StoredLastRestartResult == "" &&
		input.StoredLastRestartError == "" &&
		input.StoredRecommendedBackoffMs == 0 &&
		input.StoredResolverBlockedBy == "" &&
		input.StoredAmbiguityClass == "" &&
		input.StoredCandidateKind == "" &&
		input.StoredCandidateStrength == "" &&
		input.StoredRetryDisposition == "" &&
		input.StoredManualRetryHint == "" &&
		input.StoredNextStepAlias == "" &&
		len(input.StoredSpecificityFields) == 0 {
		return evaluateObservedSharedSessionBrowserHealth(input, reconnectWindow)
	}
	evaluation := SharedSessionBrowserHealthEvaluation{
		Summary: &SharedSessionBrowserHealthSummary{
			State:                       input.StoredState,
			Reason:                      input.StoredReason,
			RecoveryAction:              input.StoredRecoveryAction,
			ReconnectHint:               input.StoredReconnectHint,
			DisconnectCount:             input.StoredDisconnectCount,
			DisconnectBurstCount:        input.StoredDisconnectBurstCount,
			DisconnectBurstWindowMs:     input.StoredDisconnectBurstWindowMs,
			CooldownRemainingMs:         input.StoredCooldownRemainingMs,
			RetryBackoffRemainingMs:     input.StoredRetryBackoffRemainingMs,
			RestartAttemptCount:         input.StoredRestartAttemptCount,
			RestartFailureCount:         input.StoredRestartFailureCount,
			LastDisconnectUnixMilli:     input.StoredLastDisconnectUnixMilli,
			LastReconnectUnixMilli:      input.StoredLastReconnectUnixMilli,
			LastRestartAttemptUnixMilli: input.StoredLastRestartAttemptUnixMilli,
			LastRestartResult:           input.StoredLastRestartResult,
			LastRestartError:            input.StoredLastRestartError,
			RecommendedBackoffMs:        input.StoredRecommendedBackoffMs,
			ResolverBlockedBy:           input.StoredResolverBlockedBy,
			AmbiguityClass:              input.StoredAmbiguityClass,
			CandidateKind:               input.StoredCandidateKind,
			CandidateStrength:           input.StoredCandidateStrength,
			RetryDisposition:            input.StoredRetryDisposition,
			ManualRetryHint:             input.StoredManualRetryHint,
			NextStepAlias:               input.StoredNextStepAlias,
			SpecificityFields:           append([]string(nil), input.StoredSpecificityFields...),
		},
	}
	switch input.StoredState {
	case "profile_reconnecting":
		if profile, ok := sharedSessionBrowserReconnectingProfile(input.Profiles); ok {
			evaluation.Profile = profile
			evaluation.HasProfile = true
			evaluation.ReconnectTimedOut = sharedSessionBrowserReconnectTimedOut(profile, reconnectWindow, input.ReferenceTime)
		}
	case "profile_disconnected":
		if profile, ok := sharedSessionBrowserDisconnectedProfile(input.Profiles); ok {
			evaluation.Profile = profile
			evaluation.HasProfile = true
		}
	case "profile_stopped":
		if profile, ok := sharedSessionBrowserPreferredProfile(input.Profiles); ok {
			evaluation.Profile = profile
			evaluation.HasProfile = true
		}
	case "healthy":
		if profile, ok := sharedSessionBrowserConnectedProfile(input.Profiles); ok {
			evaluation.Profile = profile
			evaluation.HasProfile = true
		}
	}
	switch input.StoredState {
	case "profile_reconnecting", "profile_disconnected", "profile_stopped", "healthy":
		if !evaluation.HasProfile {
			return evaluateObservedSharedSessionBrowserHealth(input, reconnectWindow)
		}
	}
	return evaluation
}

func EvaluateSharedSessionBrowserHealthForInputScope(input SharedSessionBrowserHealthInput, selectedInfo BrowserRuntimeInfo, requestedProfile string, reconnectWindow time.Duration) SharedSessionBrowserHealthEvaluation {
	requestedProfile = firstNonEmptyString(
		strings.TrimSpace(requestedProfile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	return EvaluateSharedSessionBrowserHealth(
		sharedSessionBrowserHealthInputForScope(input, selectedInfo, requestedProfile),
		reconnectWindow,
	)
}

func EvaluateSharedSessionBrowserHealthForScope(registry SharedSessionBrowserStateRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, input SharedSessionBrowserHealthInput, reconnectWindow time.Duration) SharedSessionBrowserHealthEvaluation {
	sessionID = strings.TrimSpace(sessionID)
	requestedProfile = strings.TrimSpace(requestedProfile)
	snapshot := sharedSessionBrowserProfilesForScope(registry, sessionID, selectedInfo, requestedProfile)
	if len(snapshot) == 0 {
		return EvaluateSharedSessionBrowserHealth(input, reconnectWindow)
	}
	if !sharedSessionBrowserStoredSummaryBlocksLifecycle(input) {
		input = clearSharedSessionBrowserHealthStoredSummary(input)
	}
	input.Profiles = append([]SharedSessionBrowserProfileState(nil), snapshot...)
	return EvaluateSharedSessionBrowserHealth(input, reconnectWindow)
}

func clearSharedSessionBrowserHealthStoredSummary(input SharedSessionBrowserHealthInput) SharedSessionBrowserHealthInput {
	input.StoredState = ""
	input.StoredReason = ""
	input.StoredRecoveryAction = ""
	input.StoredReconnectHint = ""
	input.StoredDisconnectCount = 0
	input.StoredDisconnectBurstCount = 0
	input.StoredDisconnectBurstWindowMs = 0
	input.StoredCooldownRemainingMs = 0
	input.StoredRetryBackoffRemainingMs = 0
	input.StoredRestartAttemptCount = 0
	input.StoredRestartFailureCount = 0
	input.StoredLastDisconnectUnixMilli = 0
	input.StoredLastReconnectUnixMilli = 0
	input.StoredLastRestartAttemptUnixMilli = 0
	input.StoredLastRestartResult = ""
	input.StoredLastRestartError = ""
	input.StoredRecommendedBackoffMs = 0
	input.StoredResolverBlockedBy = ""
	input.StoredAmbiguityClass = ""
	input.StoredCandidateKind = ""
	input.StoredCandidateStrength = ""
	input.StoredRetryDisposition = ""
	input.StoredManualRetryHint = ""
	input.StoredNextStepAlias = ""
	input.StoredSpecificityFields = nil
	return input
}

func sharedSessionBrowserStoredSummaryBlocksLifecycle(input SharedSessionBrowserHealthInput) bool {
	return sharedSessionBrowserHealthBlocksLifecycleRecovery(&SharedSessionBrowserHealthSummary{
		State: strings.TrimSpace(input.StoredState),
	})
}

func evaluateObservedSharedSessionBrowserHealth(input SharedSessionBrowserHealthInput, reconnectWindow time.Duration) SharedSessionBrowserHealthEvaluation {
	if profile, ok := sharedSessionBrowserReconnectingProfile(input.Profiles); ok {
		reconnectTimedOut := sharedSessionBrowserReconnectTimedOut(profile, reconnectWindow, input.ReferenceTime)
		recoveryAction := ""
		if reconnectTimedOut {
			recoveryAction = "browser action=refresh"
		}
		return SharedSessionBrowserHealthEvaluation{
			Summary: &SharedSessionBrowserHealthSummary{
				State:          "profile_reconnecting",
				Reason:         sharedSessionBrowserProfileReconnectingReason(profile, reconnectWindow, input.ReferenceTime),
				RecoveryAction: recoveryAction,
			},
			Profile:           profile,
			HasProfile:        true,
			ReconnectTimedOut: reconnectTimedOut,
		}
	}
	if profile, ok := sharedSessionBrowserDisconnectedProfile(input.Profiles); ok {
		recoveryAction := ""
		if input.ActiveNodeRunID == "" {
			recoveryAction = "browser action=refresh"
		}
		return SharedSessionBrowserHealthEvaluation{
			Summary: &SharedSessionBrowserHealthSummary{
				State:          "profile_disconnected",
				Reason:         sharedSessionBrowserProfileDisconnectedReason(profile),
				RecoveryAction: recoveryAction,
			},
			Profile:    profile,
			HasProfile: true,
		}
	}
	if input.ActiveNodeRunID == "" && !sharedSessionBrowserHasRunningProfile(input.Profiles) && input.RouteTargetCount > 0 {
		return SharedSessionBrowserHealthEvaluation{
			Summary: &SharedSessionBrowserHealthSummary{
				State:          "stale_route_targets",
				Reason:         "session still tracks browser targets after the managed browser profile stopped",
				RecoveryAction: "browser action=reset",
			},
		}
	}
	if input.ActiveNodeRunID == "" && !sharedSessionBrowserHasRunningProfile(input.Profiles) && len(input.Profiles) > 0 {
		if profile, ok := sharedSessionBrowserPreferredProfile(input.Profiles); ok {
			return SharedSessionBrowserHealthEvaluation{
				Summary: &SharedSessionBrowserHealthSummary{
					State:          "profile_stopped",
					Reason:         sharedSessionBrowserProfileStoppedReason(profile),
					RecoveryAction: "browser action=ensure",
				},
				Profile:    profile,
				HasProfile: true,
			}
		}
		return SharedSessionBrowserHealthEvaluation{
			Summary: &SharedSessionBrowserHealthSummary{
				State:          "profile_stopped",
				Reason:         "managed browser profile is stopped",
				RecoveryAction: "browser action=ensure",
			},
		}
	}
	if sharedSessionBrowserHasRunningProfile(input.Profiles) {
		if profile, ok := sharedSessionBrowserConnectedProfile(input.Profiles); ok {
			return SharedSessionBrowserHealthEvaluation{
				Summary: &SharedSessionBrowserHealthSummary{
					State:  "healthy",
					Reason: sharedSessionBrowserProfileHealthyReason(profile),
				},
				Profile:    profile,
				HasProfile: true,
			}
		}
		return SharedSessionBrowserHealthEvaluation{
			Summary: &SharedSessionBrowserHealthSummary{
				State:  "healthy",
				Reason: "managed browser profile is running",
			},
		}
	}
	return SharedSessionBrowserHealthEvaluation{}
}

func sharedSessionBrowserProfilesForScope(registry SharedSessionBrowserStateRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string) []SharedSessionBrowserProfileState {
	sessionID = strings.TrimSpace(sessionID)
	requestedProfile = strings.TrimSpace(requestedProfile)
	if registry == nil || sessionID == "" {
		return nil
	}
	snapshot := registry.SnapshotSessionBrowserProfilesForScope(sessionID, selectedInfo, requestedProfile)
	if len(snapshot) == 0 {
		return nil
	}
	return snapshot
}

func sharedSessionBrowserHealthInputForScope(input SharedSessionBrowserHealthInput, selectedInfo BrowserRuntimeInfo, requestedProfile string) SharedSessionBrowserHealthInput {
	if len(input.Profiles) == 0 {
		return input
	}
	filtered := make([]SharedSessionBrowserProfileState, 0, len(input.Profiles))
	for _, item := range input.Profiles {
		if !sharedSessionBrowserProfileMatchesSelectedInfo(item, requestedProfile, selectedInfo) {
			continue
		}
		filtered = append(filtered, item)
	}
	if len(filtered) == 0 {
		input.Profiles = nil
		return input
	}
	input.Profiles = filtered
	return input
}

func SharedSessionBrowserHasExplicitProfileLifecycleObservation(status BrowserProfileStatusResult) bool {
	return strings.TrimSpace(status.Status) != "" || status.Running || status.Connected
}

func SharedSessionBrowserStatusExplicitlyHealthy(status BrowserProfileStatusResult) bool {
	if !SharedSessionBrowserHasExplicitProfileLifecycleObservation(status) {
		return false
	}
	if status.Connected {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(status.Status)) {
	case "connected", "ready":
		return true
	default:
		return false
	}
}

func SharedSessionBrowserProfileReconnectTimedOut(profile SharedSessionBrowserProfileState, reconnectWindow time.Duration) bool {
	return sharedSessionBrowserReconnectTimedOut(profile, reconnectWindow, time.Time{})
}

func sharedSessionBrowserHasRunningProfile(items []SharedSessionBrowserProfileState) bool {
	for _, profile := range items {
		status := strings.ToLower(strings.TrimSpace(profile.Status))
		if profile.Running || profile.Connected || status == "running" || status == "started" {
			return true
		}
	}
	return false
}

func sharedSessionBrowserConnectedProfile(items []SharedSessionBrowserProfileState) (SharedSessionBrowserProfileState, bool) {
	for _, profile := range items {
		if profile.Connected {
			return profile, true
		}
	}
	return sharedSessionBrowserPreferredProfile(items)
}

func sharedSessionBrowserReconnectingProfile(items []SharedSessionBrowserProfileState) (SharedSessionBrowserProfileState, bool) {
	for _, profile := range items {
		if strings.EqualFold(strings.TrimSpace(profile.Status), "reconnecting") {
			return profile, true
		}
	}
	return SharedSessionBrowserProfileState{}, false
}

func sharedSessionBrowserDisconnectedProfile(items []SharedSessionBrowserProfileState) (SharedSessionBrowserProfileState, bool) {
	for _, profile := range items {
		status := strings.ToLower(strings.TrimSpace(profile.Status))
		switch {
		case status == "disconnected" || status == "crashed":
			return profile, true
		case profile.Running && !profile.Connected:
			return profile, true
		}
	}
	return SharedSessionBrowserProfileState{}, false
}

func sharedSessionBrowserPreferredProfile(items []SharedSessionBrowserProfileState) (SharedSessionBrowserProfileState, bool) {
	if len(items) == 0 {
		return SharedSessionBrowserProfileState{}, false
	}
	return items[0], true
}

func sharedSessionBrowserReconnectTimedOut(profile SharedSessionBrowserProfileState, reconnectWindow time.Duration, referenceTime time.Time) bool {
	if reconnectWindow <= 0 {
		return false
	}
	return sharedSessionBrowserProfileHealthAge(profile, referenceTime) >= reconnectWindow
}

func sharedSessionBrowserProfileHealthAge(profile SharedSessionBrowserProfileState, referenceTime time.Time) time.Duration {
	if profile.StatusSince.IsZero() {
		return 0
	}
	if referenceTime.IsZero() {
		referenceTime = time.Now()
	}
	elapsed := referenceTime.Sub(profile.StatusSince)
	if elapsed < 0 {
		return 0
	}
	return elapsed
}

func sharedSessionBrowserProfileDisconnectedReason(profile SharedSessionBrowserProfileState) string {
	name := firstNonEmptyString(strings.TrimSpace(profile.Profile), "managed browser profile")
	status := strings.ToLower(strings.TrimSpace(profile.Status))
	switch status {
	case "reconnecting":
		return fmt.Sprintf("browser profile %q is reconnecting", name)
	case "crashed":
		return fmt.Sprintf("browser profile %q crashed and needs recovery", name)
	default:
		reason := fmt.Sprintf("browser profile %q is disconnected", name)
		if note := strings.TrimSpace(profile.Note); note != "" {
			reason = reason + ": " + note
		}
		return reason
	}
}

func sharedSessionBrowserProfileReconnectingReason(profile SharedSessionBrowserProfileState, reconnectWindow time.Duration, referenceTime time.Time) string {
	name := firstNonEmptyString(strings.TrimSpace(profile.Profile), "managed browser profile")
	reason := fmt.Sprintf("browser profile %q is reconnecting", name)
	if elapsed := sharedSessionBrowserProfileHealthAge(profile, referenceTime); elapsed > 0 {
		if sharedSessionBrowserReconnectTimedOut(profile, reconnectWindow, referenceTime) {
			reason = fmt.Sprintf("%s beyond watchdog window (%s)", reason, sharedSessionBrowserFormatDuration(elapsed))
		} else {
			reason = fmt.Sprintf("%s (%s elapsed)", reason, sharedSessionBrowserFormatDuration(elapsed))
		}
	}
	if note := strings.TrimSpace(profile.Note); note != "" {
		reason = reason + ": " + note
	}
	return reason
}

func sharedSessionBrowserProfileStoppedReason(profile SharedSessionBrowserProfileState) string {
	name := firstNonEmptyString(strings.TrimSpace(profile.Profile), "managed browser profile")
	reason := fmt.Sprintf("browser profile %q is stopped", name)
	if note := strings.TrimSpace(profile.Note); note != "" {
		reason = reason + ": " + note
	}
	return reason
}

func sharedSessionBrowserProfileHealthyReason(profile SharedSessionBrowserProfileState) string {
	name := firstNonEmptyString(strings.TrimSpace(profile.Profile), "managed browser profile")
	if profile.Connected {
		return fmt.Sprintf("browser profile %q is running and connected", name)
	}
	return fmt.Sprintf("browser profile %q is running", name)
}

func sharedSessionBrowserFormatDuration(d time.Duration) string {
	if d <= 0 {
		return "0s"
	}
	seconds := int(d.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return fmt.Sprintf("%ds", seconds)
}
