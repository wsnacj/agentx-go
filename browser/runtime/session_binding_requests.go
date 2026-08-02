package browserruntime

import "time"

// BuildSharedSessionBrowserHealthInputFromBindingEvaluation projects the
// lifecycle-owned binding evaluation back into the stable health input used by
// recovery and coordination helpers.
func BuildSharedSessionBrowserHealthInputFromBindingEvaluation(evaluation SharedSessionBrowserBindingEvaluation) SharedSessionBrowserHealthInput {
	return ApplySharedSessionBrowserHealthStoredSummary(BuildSharedSessionBrowserHealthInputAt(
		evaluation.Snapshot.Summary.ActiveNodeRunID,
		evaluation.Snapshot.Summary.RouteTargetCount,
		sharedSessionBrowserHealthSummaryState(evaluation.Health.Summary),
		sharedSessionBrowserHealthSummaryReason(evaluation.Health.Summary),
		sharedSessionBrowserHealthSummaryRecoveryAction(evaluation.Health.Summary),
		sharedSessionBrowserHealthSummaryResolverBlockedBy(evaluation.Health.Summary),
		sharedSessionBrowserHealthSummaryAmbiguityClass(evaluation.Health.Summary),
		sharedSessionBrowserHealthSummaryCandidateKind(evaluation.Health.Summary),
		sharedSessionBrowserHealthSummaryCandidateStrength(evaluation.Health.Summary),
		sharedSessionBrowserHealthSummaryRetryDisposition(evaluation.Health.Summary),
		sharedSessionBrowserHealthSummaryManualRetryHint(evaluation.Health.Summary),
		sharedSessionBrowserHealthSummaryNextStepAlias(evaluation.Health.Summary),
		sharedSessionBrowserHealthSummarySpecificityFields(evaluation.Health.Summary),
		evaluation.Snapshot.Profiles,
		evaluation.ReferenceTime,
	), evaluation.Health.Summary)
}

// BuildSharedSessionBrowserCoordinationInputFromBindingEvaluation projects the
// lifecycle-owned binding evaluation into the compact coordination input used
// by the shared route/session action planner.
func BuildSharedSessionBrowserCoordinationInputFromBindingEvaluation(evaluation SharedSessionBrowserBindingEvaluation) SharedSessionBrowserCoordinationInput {
	return BuildSharedSessionBrowserCoordinationInput(
		evaluation.Snapshot.Summary.ActiveNodeRunID,
		evaluation.Snapshot.Summary.RouteTargetCount,
		evaluation.Snapshot.SelectedProfileSelection,
		evaluation.Snapshot.SelectedTargetSelection,
		evaluation.Snapshot.Profiles,
	)
}

// BuildSharedSessionBrowserExecutionRequestFromBindingEvaluation assembles the
// shared execution request directly from a lifecycle-owned binding evaluation
// so callers do not need to reassemble route/run/profile state field-by-field.
func BuildSharedSessionBrowserExecutionRequestFromBindingEvaluation(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	requestedProfile string,
	selectedInfo BrowserRuntimeInfo,
	force bool,
	evaluation SharedSessionBrowserBindingEvaluation,
	reconnectWindow time.Duration,
) SharedSessionBrowserExecutionRequest {
	return BuildSharedSessionBrowserExecutionRequest(
		registry,
		sessionID,
		requestedProfile,
		selectedInfo,
		force,
		evaluation.Snapshot.Summary.ActiveNodeRunID,
		evaluation.Snapshot.Summary.ActiveBrowserProfile,
		BuildSharedSessionBrowserHealthInputFromBindingEvaluation(evaluation),
		reconnectWindow,
	)
}

// BuildSharedSessionBrowserExecutionRequestForBindingEvaluation assembles the
// shared execution request from an optional binding evaluation so callers can
// delegate nil/empty evaluation handling to browserruntime.
func BuildSharedSessionBrowserExecutionRequestForBindingEvaluation(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	requestedProfile string,
	selectedInfo BrowserRuntimeInfo,
	force bool,
	evaluation *SharedSessionBrowserBindingEvaluation,
	reconnectWindow time.Duration,
) SharedSessionBrowserExecutionRequest {
	sharedEvaluation := SharedSessionBrowserBindingEvaluation{}
	if evaluation != nil {
		sharedEvaluation = *evaluation
	}
	return BuildSharedSessionBrowserExecutionRequestFromBindingEvaluation(
		registry,
		sessionID,
		requestedProfile,
		selectedInfo,
		force,
		sharedEvaluation,
		reconnectWindow,
	)
}

// BuildSharedSessionBrowserClearRequestFromBindingEvaluation assembles the
// shared clear-action request directly from a lifecycle-owned binding
// evaluation so clear paths reuse the same health/source-time contract as the
// broader session watch loop.
func BuildSharedSessionBrowserClearRequestFromBindingEvaluation(
	sessionRegistry *BrowserSessionRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	force bool,
	evaluation SharedSessionBrowserBindingEvaluation,
	reconnectWindow time.Duration,
) SharedSessionBrowserClearRequest {
	return BuildSharedSessionBrowserClearRequest(
		sessionRegistry,
		stateRegistry,
		sessionID,
		selectedInfo,
		route,
		force,
		evaluation.Snapshot.Summary.ActiveNodeRunID,
		BuildSharedSessionBrowserHealthInputFromBindingEvaluation(evaluation),
		reconnectWindow,
	)
}

// BuildSharedSessionBrowserClearRequestForBindingEvaluation assembles the
// shared clear request from an optional binding evaluation so callers can
// delegate nil/empty evaluation handling to browserruntime.
func BuildSharedSessionBrowserClearRequestForBindingEvaluation(
	sessionRegistry *BrowserSessionRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	force bool,
	evaluation *SharedSessionBrowserBindingEvaluation,
	reconnectWindow time.Duration,
) SharedSessionBrowserClearRequest {
	sharedEvaluation := SharedSessionBrowserBindingEvaluation{}
	if evaluation != nil {
		sharedEvaluation = *evaluation
	}
	return BuildSharedSessionBrowserClearRequestFromBindingEvaluation(
		sessionRegistry,
		stateRegistry,
		sessionID,
		selectedInfo,
		route,
		force,
		sharedEvaluation,
		reconnectWindow,
	)
}
