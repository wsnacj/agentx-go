package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeSessionHealthEvaluation struct {
	Summary           *browserRuntimeSessionHealthSummary
	Profile           browserRuntimeProfileState
	HasProfile        bool
	ReconnectTimedOut bool
}

type browserRuntimeRouteProfileRecoveryAssessment struct {
	EffectiveStatus          BrowserProfileStatusResult
	Health                   browserRuntimeSessionHealthEvaluation
	NeedsRefreshRecovery     bool
	ShouldStopBeforeRecovery bool
	ReconnectInProgress      bool
	HasSyntheticStatus       bool
	SyntheticStatus          BrowserProfileStatusResult
}

func browserRuntimeSessionHealthEvaluationFromBinding(binding browserRuntimeSessionBinding) browserRuntimeSessionHealthEvaluation {
	return browserRuntimeSessionHealthEvaluationFromShared(
		agentxbrowserruntime.EvaluateSharedSessionBrowserHealth(
			browserRuntimeSessionHealthInputFromBinding(binding),
			browserRuntimeReconnectWatchdogWindow,
		),
	)
}

func browserRuntimeApplySessionHealthEvaluation(binding *browserRuntimeSessionBinding, evaluation browserRuntimeSessionHealthEvaluation) {
	if binding == nil || evaluation.Summary == nil {
		if binding != nil {
			binding.SessionHealthState = ""
			binding.SessionHealthReason = ""
			binding.SessionHealthRecoveryAction = ""
			binding.SessionHealthReconnectHint = ""
			binding.SessionHealthDisconnectCount = 0
			binding.SessionHealthDisconnectBurstCount = 0
			binding.SessionHealthDisconnectBurstWindowMs = 0
			binding.SessionHealthCooldownRemainingMs = 0
			binding.SessionHealthRetryBackoffRemainingMs = 0
			binding.SessionHealthRestartAttemptCount = 0
			binding.SessionHealthRestartFailureCount = 0
			binding.SessionHealthLastDisconnectUnixMilli = 0
			binding.SessionHealthLastReconnectUnixMilli = 0
			binding.SessionHealthLastRestartAttemptUnixMilli = 0
			binding.SessionHealthLastRestartResult = ""
			binding.SessionHealthLastRestartError = ""
			binding.SessionHealthRecommendedBackoffMs = 0
			binding.SessionHealthResolverBlockedBy = ""
			binding.SessionHealthResolverAmbiguityClass = ""
			binding.SessionHealthResolverCandidateKind = ""
			binding.SessionHealthResolverStrength = ""
			binding.SessionHealthResolverRetryDisposition = ""
			binding.SessionHealthResolverManualRetryHint = ""
			binding.SessionHealthResolverNextStepAlias = ""
			binding.SessionHealthResolverSpecificityFields = nil
			if binding.HasSharedEvaluation {
				binding.SharedEvaluation.Health.Summary = nil
			}
		}
		return
	}
	binding.SessionHealthState = strings.TrimSpace(evaluation.Summary.State)
	binding.SessionHealthReason = strings.TrimSpace(evaluation.Summary.Reason)
	binding.SessionHealthRecoveryAction = strings.TrimSpace(evaluation.Summary.RecoveryAction)
	binding.SessionHealthReconnectHint = strings.TrimSpace(evaluation.Summary.ReconnectHint)
	binding.SessionHealthDisconnectCount = evaluation.Summary.DisconnectCount
	binding.SessionHealthDisconnectBurstCount = evaluation.Summary.DisconnectBurstCount
	binding.SessionHealthDisconnectBurstWindowMs = evaluation.Summary.DisconnectBurstWindowMs
	binding.SessionHealthCooldownRemainingMs = evaluation.Summary.CooldownRemainingMs
	binding.SessionHealthRetryBackoffRemainingMs = evaluation.Summary.RetryBackoffRemainingMs
	binding.SessionHealthRestartAttemptCount = evaluation.Summary.RestartAttemptCount
	binding.SessionHealthRestartFailureCount = evaluation.Summary.RestartFailureCount
	binding.SessionHealthLastDisconnectUnixMilli = evaluation.Summary.LastDisconnectUnixMilli
	binding.SessionHealthLastReconnectUnixMilli = evaluation.Summary.LastReconnectUnixMilli
	binding.SessionHealthLastRestartAttemptUnixMilli = evaluation.Summary.LastRestartAttemptUnixMilli
	binding.SessionHealthLastRestartResult = strings.TrimSpace(evaluation.Summary.LastRestartResult)
	binding.SessionHealthLastRestartError = strings.TrimSpace(evaluation.Summary.LastRestartError)
	binding.SessionHealthRecommendedBackoffMs = evaluation.Summary.RecommendedBackoffMs
	binding.SessionHealthResolverBlockedBy = strings.TrimSpace(evaluation.Summary.ResolverBlockedBy)
	binding.SessionHealthResolverAmbiguityClass = strings.TrimSpace(evaluation.Summary.AmbiguityClass)
	binding.SessionHealthResolverCandidateKind = strings.TrimSpace(evaluation.Summary.CandidateKind)
	binding.SessionHealthResolverStrength = strings.TrimSpace(evaluation.Summary.CandidateStrength)
	binding.SessionHealthResolverRetryDisposition = strings.TrimSpace(evaluation.Summary.RetryDisposition)
	binding.SessionHealthResolverManualRetryHint = strings.TrimSpace(evaluation.Summary.ManualRetryHint)
	binding.SessionHealthResolverNextStepAlias = strings.TrimSpace(evaluation.Summary.NextStepAlias)
	binding.SessionHealthResolverSpecificityFields = append([]string(nil), evaluation.Summary.SpecificityFields...)
	if binding.HasSharedEvaluation {
		binding.SharedEvaluation.Health.Summary = browserRuntimeSharedSessionHealthSummary(*binding)
	}
}

func browserRuntimeEvaluateSessionHealth(binding browserRuntimeSessionBinding) browserRuntimeSessionHealthEvaluation {
	return browserRuntimeSessionHealthEvaluationFromBinding(binding)
}

func browserRuntimeSessionHealthEvaluationWithRegistry(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry, binding *browserRuntimeSessionBinding, selectedRoute *browserRuntimeRouteDescriptor) browserRuntimeSessionHealthEvaluation {
	var input agentxbrowserruntime.SharedSessionBrowserHealthInput
	if binding != nil {
		input = browserRuntimeSessionHealthInputFromBinding(*binding)
	}
	selectedInfo := BrowserRuntimeInfo{}
	requestedProfile := ""
	if selectedRoute != nil {
		selectedInfo.Backend = strings.TrimSpace(selectedRoute.Backend)
		selectedInfo.Target = strings.TrimSpace(selectedRoute.RuntimeTarget)
		requestedProfile = strings.TrimSpace(firstNonEmpty(selectedRoute.Profile, inputProfileFromBinding(binding)))
	}
	sessionID, _ := browserRuntimeSessionStateRegistrySessionID(ctx, registry)
	return browserRuntimeSessionHealthEvaluationFromShared(
		agentxbrowserruntime.EvaluateSharedSessionBrowserHealthForScope(
			registry,
			sessionID,
			selectedInfo,
			requestedProfile,
			input,
			browserRuntimeReconnectWatchdogWindow,
		),
	)
}

func inputProfileFromBinding(binding *browserRuntimeSessionBinding) string {
	if binding == nil {
		return ""
	}
	return strings.TrimSpace(binding.SelectedBrowserProfile)
}

func browserRuntimeAssessRouteProfileRecovery(binding *browserRuntimeSessionBinding, profile string, selectedInfo BrowserRuntimeInfo, status BrowserProfileStatusResult) browserRuntimeRouteProfileRecoveryAssessment {
	input := agentxbrowserruntime.SharedSessionBrowserHealthInput{}
	if binding != nil {
		input = browserRuntimeSessionHealthInputFromBinding(*binding)
	}
	health, assessment := agentxbrowserruntime.AssessSharedSessionBrowserProfileRecoveryForInputScope(
		input,
		selectedInfo,
		profile,
		status,
		browserRuntimeReconnectWatchdogWindow,
	)
	return browserRuntimeRouteProfileRecoveryAssessmentFromShared(
		browserRuntimeSessionHealthEvaluationFromShared(health),
		assessment,
	)
}

func browserRuntimeAssessRouteProfileRecoveryWithRegistry(ctx context.Context, registry agentxbrowserruntime.SharedSessionBrowserStateRegistry, binding *browserRuntimeSessionBinding, profile string, selectedInfo BrowserRuntimeInfo, status BrowserProfileStatusResult) browserRuntimeRouteProfileRecoveryAssessment {
	var input agentxbrowserruntime.SharedSessionBrowserHealthInput
	if binding != nil {
		input = browserRuntimeSessionHealthInputFromBinding(*binding)
	}
	sessionID, _ := browserRuntimeSessionStateRegistrySessionID(ctx, registry)
	health, assessment := agentxbrowserruntime.AssessSharedSessionBrowserProfileRecoveryForScope(
		registry,
		sessionID,
		selectedInfo,
		profile,
		input,
		status,
		browserRuntimeReconnectWatchdogWindow,
	)
	return browserRuntimeRouteProfileRecoveryAssessmentFromShared(
		browserRuntimeSessionHealthEvaluationFromShared(health),
		assessment,
	)
}

func browserRuntimeSessionHealthInputFromBinding(binding browserRuntimeSessionBinding) agentxbrowserruntime.SharedSessionBrowserHealthInput {
	if binding.HasSharedEvaluation {
		return agentxbrowserruntime.BuildSharedSessionBrowserHealthInputFromBindingEvaluation(binding.SharedEvaluation)
	}
	input := agentxbrowserruntime.BuildSharedSessionBrowserHealthInputAt(
		binding.ActiveNodeRunID,
		binding.RouteTargetCount,
		binding.SessionHealthState,
		binding.SessionHealthReason,
		binding.SessionHealthRecoveryAction,
		binding.SessionHealthResolverBlockedBy,
		binding.SessionHealthResolverAmbiguityClass,
		binding.SessionHealthResolverCandidateKind,
		binding.SessionHealthResolverStrength,
		binding.SessionHealthResolverRetryDisposition,
		binding.SessionHealthResolverManualRetryHint,
		binding.SessionHealthResolverNextStepAlias,
		binding.SessionHealthResolverSpecificityFields,
		browserRuntimeSharedSessionProfileStates(binding.BrowserProfiles),
		binding.ReferenceTime,
	)
	return agentxbrowserruntime.ApplySharedSessionBrowserHealthStoredSummary(
		input,
		browserRuntimeSharedSessionHealthSummary(binding),
	)
}

func browserRuntimeSessionHealthEvaluationFromShared(evaluation agentxbrowserruntime.SharedSessionBrowserHealthEvaluation) browserRuntimeSessionHealthEvaluation {
	out := browserRuntimeSessionHealthEvaluation{
		HasProfile:        evaluation.HasProfile,
		ReconnectTimedOut: evaluation.ReconnectTimedOut,
	}
	if evaluation.Summary != nil {
		out.Summary = &browserRuntimeSessionHealthSummary{
			State:                       strings.TrimSpace(evaluation.Summary.State),
			Reason:                      strings.TrimSpace(evaluation.Summary.Reason),
			RecoveryAction:              strings.TrimSpace(evaluation.Summary.RecoveryAction),
			ReconnectHint:               strings.TrimSpace(evaluation.Summary.ReconnectHint),
			DisconnectCount:             evaluation.Summary.DisconnectCount,
			DisconnectBurstCount:        evaluation.Summary.DisconnectBurstCount,
			DisconnectBurstWindowMs:     evaluation.Summary.DisconnectBurstWindowMs,
			CooldownRemainingMs:         evaluation.Summary.CooldownRemainingMs,
			RetryBackoffRemainingMs:     evaluation.Summary.RetryBackoffRemainingMs,
			RestartAttemptCount:         evaluation.Summary.RestartAttemptCount,
			RestartFailureCount:         evaluation.Summary.RestartFailureCount,
			LastDisconnectUnixMilli:     evaluation.Summary.LastDisconnectUnixMilli,
			LastReconnectUnixMilli:      evaluation.Summary.LastReconnectUnixMilli,
			LastRestartAttemptUnixMilli: evaluation.Summary.LastRestartAttemptUnixMilli,
			LastRestartResult:           strings.TrimSpace(evaluation.Summary.LastRestartResult),
			LastRestartError:            strings.TrimSpace(evaluation.Summary.LastRestartError),
			RecommendedBackoffMs:        evaluation.Summary.RecommendedBackoffMs,
			ResolverBlockedBy:           strings.TrimSpace(evaluation.Summary.ResolverBlockedBy),
			AmbiguityClass:              strings.TrimSpace(evaluation.Summary.AmbiguityClass),
			CandidateKind:               strings.TrimSpace(evaluation.Summary.CandidateKind),
			CandidateStrength:           strings.TrimSpace(evaluation.Summary.CandidateStrength),
			RetryDisposition:            strings.TrimSpace(evaluation.Summary.RetryDisposition),
			ManualRetryHint:             strings.TrimSpace(evaluation.Summary.ManualRetryHint),
			NextStepAlias:               strings.TrimSpace(evaluation.Summary.NextStepAlias),
			SpecificityFields:           append([]string(nil), evaluation.Summary.SpecificityFields...),
		}
	}
	if evaluation.HasProfile {
		out.Profile = browserRuntimeProfileStateFromSharedSessionState(evaluation.Profile)
	}
	return out
}

func browserRuntimeStatusExplicitlyHealthy(status BrowserProfileStatusResult) bool {
	return agentxbrowserruntime.SharedSessionBrowserStatusExplicitlyHealthy(status)
}

func browserRuntimeRouteProfileRecoveryAssessmentFromShared(health browserRuntimeSessionHealthEvaluation, assessment agentxbrowserruntime.SharedSessionBrowserProfileRecoveryAssessment) browserRuntimeRouteProfileRecoveryAssessment {
	return browserRuntimeRouteProfileRecoveryAssessment{
		EffectiveStatus:          assessment.EffectiveStatus,
		Health:                   health,
		NeedsRefreshRecovery:     assessment.NeedsRefreshRecovery,
		ShouldStopBeforeRecovery: assessment.ShouldStopBeforeRecovery,
		ReconnectInProgress:      assessment.ReconnectInProgress,
		HasSyntheticStatus:       assessment.HasSyntheticStatus,
		SyntheticStatus:          assessment.SyntheticStatus,
	}
}
