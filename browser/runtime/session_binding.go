package browserruntime

import (
	"context"
	"strings"
	"time"
)

// SharedSessionBrowserBindingSummary is the shared aggregate view used by
// browser runtime inspection payloads after route/run/profile snapshots are
// projected into stable session-scoped contracts.
type SharedSessionBrowserBindingSummary struct {
	CurrentTargetID             string
	RouteTargetCount            int
	PendingTargetReviewCount    int
	BlockedAutoFollowRouteCount int
	PopupStormRouteCount        int
	NodeRunCount                int
	ActiveNodeRunID             string
	NodeRunStatusCounts         map[string]int
	BrowserProfileCount         int
	ActiveBrowserProfile        string
	BrowserProfileStatusCounts  map[string]int
}

// SharedSessionBrowserBindingSnapshot is the shared session-scoped projection
// of route/run/profile state and current selections used by runtime/session
// inspection payloads.
type SharedSessionBrowserBindingSnapshot struct {
	CurrentTargetID          string
	SelectedProfileSelection *SharedSessionBrowserProfileSelection
	SelectedTargetSelection  *BrowserSessionTargetSelection
	Runs                     []SharedSessionRunInfo
	Profiles                 []SharedSessionBrowserProfileState
	Summary                  SharedSessionBrowserBindingSummary
}

// SharedSessionBrowserBindingEvaluation captures the shared session-scoped
// binding snapshot together with lifecycle-owned health and coordination
// evaluation for the selected route scope.
type SharedSessionBrowserBindingEvaluation struct {
	Routes        []SharedSessionBrowserRouteSnapshot
	Snapshot      SharedSessionBrowserBindingSnapshot
	ReferenceTime time.Time
	Health        SharedSessionBrowserHealthEvaluation
	Coordination  SharedSessionBrowserCoordinationEvaluation
	Handoff       *SharedSessionBrowserSessionHandoffSummary
}

// SharedSessionBrowserBindingObservation captures a single status/profiles
// watch cycle together with the binding/health/coordination evaluation that
// should consume that cycle as its shared source of truth.
type SharedSessionBrowserBindingObservation struct {
	Observation SharedSessionBrowserStatusAndProfilesObservation
	Evaluation  SharedSessionBrowserBindingEvaluation
}

// ObserveSharedSessionBrowserBindingForScope runs a scoped status/profiles
// watch cycle and refreshes the shared binding evaluation from the same source
// snapshot so control-plane consumers do not need to stitch these phases
// together locally.
func ObserveSharedSessionBrowserBindingForScope(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	requestedProfile string,
	includeStatus bool,
	includeProfiles bool,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserBindingObservation {
	return sharedSessionBrowserObserverManager(
		registry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ObserveBinding(ctx, control, SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     selectedInfo,
		BindingRoute:     route,
		RequestedProfile: requestedProfile,
		IncludeStatus:    includeStatus,
		IncludeProfiles:  includeProfiles,
	})
}

func observeSharedSessionBrowserBindingForScopeFromCycle(
	req SharedSessionBrowserObserverRequest,
	cycle SharedSessionBrowserEventCycleObservation,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserBindingObservation {
	observation, _ := observeSharedSessionBrowserBindingForScopeFromCycleWithInvalidation(
		req,
		cycle,
		registry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	)
	return observation
}

func observeSharedSessionBrowserBindingForScopeFromCycleWithInvalidation(
	req SharedSessionBrowserObserverRequest,
	cycle SharedSessionBrowserEventCycleObservation,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) (SharedSessionBrowserBindingObservation, bool) {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.SelectedInfo.Backend = strings.TrimSpace(req.SelectedInfo.Backend)
	req.SelectedInfo.Profile = strings.TrimSpace(req.SelectedInfo.Profile)
	req.SelectedInfo.Target = strings.TrimSpace(req.SelectedInfo.Target)
	req.BindingRoute = normalizeBrowserSessionRoute(req.BindingRoute)
	req.RequestedProfile = firstNonEmptyBindingString(
		strings.TrimSpace(req.RequestedProfile),
		strings.TrimSpace(req.SelectedInfo.Profile),
		strings.TrimSpace(req.BindingRoute.Profile),
	)

	observation := cycle.Observation
	cleared := false
	if observation.HasSyncedState {
		cleared = invalidateSharedSessionBrowserCurrentTargetForProfileState(registry, req.SessionID, observation.SyncedState, false)
	}

	binding := EvaluateSharedSessionBrowserBindingForScope(
		req.SessionID,
		req.SelectedInfo,
		req.BindingRoute,
		nil,
		registry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	)
	if len(observation.Snapshot) > 0 {
		binding = MergeSharedSessionBrowserBindingEvaluationProfileSnapshot(
			stateRegistry,
			req.SessionID,
			req.SelectedInfo,
			req.RequestedProfile,
			binding,
			observation.Snapshot,
			reconnectWindow,
		)
	} else if observation.Status != nil {
		binding = MergeSharedSessionBrowserBindingEvaluationProfileState(
			stateRegistry,
			req.SessionID,
			req.SelectedInfo,
			req.RequestedProfile,
			binding,
			SharedSessionBrowserProfileStateFromObservedStatus(
				req.SelectedInfo,
				observation.ResolvedStatus,
				observation.StatusObservedAt,
			),
			reconnectWindow,
		)
	}

	return SharedSessionBrowserBindingObservation{
		Observation: observation,
		Evaluation:  binding,
	}, cleared
}

// MergeSharedSessionBrowserBindingEvaluationProfileState merges a newly
// observed current profile state into an existing binding evaluation and
// refreshes the derived summary/health/coordination contracts from the same
// shared owner.
func MergeSharedSessionBrowserBindingEvaluationProfileState(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	evaluation SharedSessionBrowserBindingEvaluation,
	current SharedSessionBrowserProfileState,
	reconnectWindow time.Duration,
) SharedSessionBrowserBindingEvaluation {
	selectedInfo.Backend = strings.TrimSpace(selectedInfo.Backend)
	selectedInfo.Profile = strings.TrimSpace(selectedInfo.Profile)
	selectedInfo.Target = strings.TrimSpace(selectedInfo.Target)
	requestedProfile = firstNonEmptyBindingString(
		strings.TrimSpace(requestedProfile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	current.Backend = firstNonEmptyBindingString(strings.TrimSpace(current.Backend), strings.TrimSpace(selectedInfo.Backend))
	current.Profile = firstNonEmptyBindingString(strings.TrimSpace(current.Profile), requestedProfile)
	current.RuntimeTarget = firstNonEmptyBindingString(strings.TrimSpace(current.RuntimeTarget), strings.TrimSpace(selectedInfo.Target))
	current.BrowserApp = strings.TrimSpace(current.BrowserApp)
	current.Status = strings.TrimSpace(current.Status)
	current.Note = strings.TrimSpace(current.Note)
	if current.Profile == "" && current.Status == "" && !current.Running && !current.Connected {
		return evaluation
	}

	return refreshSharedSessionBrowserBindingEvaluationProfiles(
		registry,
		sessionID,
		selectedInfo,
		requestedProfile,
		evaluation,
		mergeSharedSessionBrowserBindingProfile(evaluation.Snapshot.Profiles, current),
		reconnectWindow,
	)
}

// MergeSharedSessionBrowserBindingEvaluationProfileSnapshot replaces the
// scoped profile snapshot consumed by a binding evaluation and recomputes the
// derived health/coordination posture from that shared profile source.
func MergeSharedSessionBrowserBindingEvaluationProfileSnapshot(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	evaluation SharedSessionBrowserBindingEvaluation,
	profiles []SharedSessionBrowserProfileState,
	reconnectWindow time.Duration,
) SharedSessionBrowserBindingEvaluation {
	return refreshSharedSessionBrowserBindingEvaluationProfiles(
		registry,
		sessionID,
		selectedInfo,
		requestedProfile,
		evaluation,
		append([]SharedSessionBrowserProfileState(nil), profiles...),
		reconnectWindow,
	)
}

// MergeSharedSessionBrowserBindingEvaluationHealthSummary overlays a newly
// observed machine-readable health summary onto an existing binding evaluation
// and refreshes the derived health/coordination posture from the same shared
// owner. This is used when backend runtime status exposes typed blocker state
// that does not naturally round-trip through the scoped profile snapshot.
func MergeSharedSessionBrowserBindingEvaluationHealthSummary(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	evaluation SharedSessionBrowserBindingEvaluation,
	summary *SharedSessionBrowserHealthSummary,
	reconnectWindow time.Duration,
) SharedSessionBrowserBindingEvaluation {
	if summary == nil {
		return evaluation
	}
	selectedInfo.Backend = strings.TrimSpace(selectedInfo.Backend)
	selectedInfo.Profile = strings.TrimSpace(selectedInfo.Profile)
	selectedInfo.Target = strings.TrimSpace(selectedInfo.Target)
	requestedProfile = firstNonEmptyBindingString(
		strings.TrimSpace(requestedProfile),
		strings.TrimSpace(selectedInfo.Profile),
	)

	snapshot := evaluation.Snapshot
	referenceTime := evaluation.ReferenceTime
	if referenceTime.IsZero() {
		referenceTime = SharedSessionBrowserLatestObservedAt(snapshot.Profiles)
	}
	healthInput := BuildSharedSessionBrowserHealthInputAt(
		snapshot.Summary.ActiveNodeRunID,
		snapshot.Summary.RouteTargetCount,
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
		snapshot.Profiles,
		referenceTime,
	)
	healthInput = ApplySharedSessionBrowserHealthSummary(healthInput, summary)

	evaluation.ReferenceTime = referenceTime
	sessionID = strings.TrimSpace(sessionID)
	if registry != nil && sessionID != "" {
		evaluation.Health = EvaluateSharedSessionBrowserHealthForScope(
			registry,
			sessionID,
			selectedInfo,
			requestedProfile,
			healthInput,
			reconnectWindow,
		)
		evaluation.Coordination = EvaluateSharedSessionBrowserCoordinationEvaluationForScope(
			registry,
			sessionID,
			selectedInfo,
			requestedProfile,
			healthInput,
			BuildSharedSessionBrowserCoordinationInput(
				snapshot.Summary.ActiveNodeRunID,
				snapshot.Summary.RouteTargetCount,
				snapshot.SelectedProfileSelection,
				snapshot.SelectedTargetSelection,
				snapshot.Profiles,
			),
			SharedSessionBrowserRouteCoordinationInputs(evaluation.Routes),
			reconnectWindow,
			snapshot.Summary.BlockedAutoFollowRouteCount > 0,
		)
		return sharedSessionBrowserFinalizeBindingEvaluationHandoff(evaluation)
	}

	evaluation.Health = EvaluateSharedSessionBrowserHealthForInputScope(
		healthInput,
		selectedInfo,
		requestedProfile,
		reconnectWindow,
	)
	evaluation.Coordination = EvaluateSharedSessionBrowserCoordinationEvaluation(SharedSessionBrowserCoordinationEvaluationInput{
		Coordination: BuildSharedSessionBrowserCoordinationInput(
			snapshot.Summary.ActiveNodeRunID,
			snapshot.Summary.RouteTargetCount,
			snapshot.SelectedProfileSelection,
			snapshot.SelectedTargetSelection,
			snapshot.Profiles,
		),
		Routes:            SharedSessionBrowserRouteCoordinationInputs(evaluation.Routes),
		HealthEvaluation:  evaluation.Health,
		BlockedAutoFollow: snapshot.Summary.BlockedAutoFollowRouteCount > 0,
	})
	return sharedSessionBrowserFinalizeBindingEvaluationHandoff(evaluation)
}

// SnapshotSharedSessionBrowserBinding projects the shared session-scoped route,
// run, profile, and selection state needed to build a runtime binding payload.
func SnapshotSharedSessionBrowserBinding(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	routes []SharedSessionBrowserRouteSnapshot,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
) SharedSessionBrowserBindingSnapshot {
	sessionID = strings.TrimSpace(sessionID)
	selectedInfo.Backend = strings.TrimSpace(selectedInfo.Backend)
	selectedInfo.Profile = strings.TrimSpace(selectedInfo.Profile)
	selectedInfo.Target = strings.TrimSpace(selectedInfo.Target)
	route = normalizeBrowserSessionRoute(route)
	snapshot := SharedSessionBrowserBindingSnapshot{}
	if sessionID == "" {
		return snapshot
	}
	if stateRegistry != nil {
		if selection, ok := stateRegistry.SelectedBrowserProfile(sessionID, strings.TrimSpace(selectedInfo.Target)); ok {
			selection.Backend = strings.TrimSpace(selection.Backend)
			selection.Profile = strings.TrimSpace(selection.Profile)
			selection.RuntimeTarget = strings.TrimSpace(selection.RuntimeTarget)
			selection.BrowserApp = strings.TrimSpace(selection.BrowserApp)
			selection.Source = strings.TrimSpace(selection.Source)
			snapshot.SelectedProfileSelection = &selection
		}
		snapshot.Profiles = stateRegistry.SnapshotSessionBrowserProfilesForScope(sessionID, selectedInfo, firstNonEmptyBindingString(strings.TrimSpace(selectedInfo.Profile), strings.TrimSpace(route.Profile)))
	}
	if runRegistry != nil {
		snapshot.Runs = runRegistry.SnapshotSessionRuns(sessionID)
	}
	if registry != nil {
		snapshot.SelectedTargetSelection = CurrentSharedSessionBrowserTargetSelection(registry, sessionID, route)
		if routes == nil {
			routes = SnapshotSharedSessionBrowserRoutes(registry, sessionID, route)
		}
	}
	snapshot.Summary = SummarizeSharedSessionBrowserBinding(routes, snapshot.Runs, snapshot.Profiles)
	if snapshot.SelectedTargetSelection != nil && strings.TrimSpace(snapshot.SelectedTargetSelection.ID) != "" {
		snapshot.CurrentTargetID = strings.TrimSpace(snapshot.SelectedTargetSelection.ID)
	} else {
		snapshot.CurrentTargetID = strings.TrimSpace(snapshot.Summary.CurrentTargetID)
	}
	return snapshot
}

// EvaluateSharedSessionBrowserBindingForScope projects the shared binding
// snapshot for a scoped route and immediately evaluates the corresponding
// health and coordination posture from the same shared owner.
func EvaluateSharedSessionBrowserBindingForScope(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	route BrowserSessionRoute,
	routes []SharedSessionBrowserRouteSnapshot,
	registry *BrowserSessionRegistry,
	runRegistry SharedSessionRunRegistry,
	stateRegistry SharedSessionBrowserStateRegistry,
	reconnectWindow time.Duration,
) SharedSessionBrowserBindingEvaluation {
	sessionID = strings.TrimSpace(sessionID)
	selectedInfo.Backend = strings.TrimSpace(selectedInfo.Backend)
	selectedInfo.Profile = strings.TrimSpace(selectedInfo.Profile)
	selectedInfo.Target = strings.TrimSpace(selectedInfo.Target)
	route = normalizeBrowserSessionRoute(route)

	evaluation := SharedSessionBrowserBindingEvaluation{}
	if routes == nil && registry != nil {
		routes = SnapshotSharedSessionBrowserRoutes(registry, sessionID, route)
	}
	if len(routes) > 0 {
		evaluation.Routes = append([]SharedSessionBrowserRouteSnapshot(nil), routes...)
	}
	evaluation.Snapshot = SnapshotSharedSessionBrowserBinding(
		sessionID,
		selectedInfo,
		route,
		evaluation.Routes,
		registry,
		runRegistry,
		stateRegistry,
	)

	requestedProfile := firstNonEmptyBindingString(strings.TrimSpace(selectedInfo.Profile), strings.TrimSpace(route.Profile))
	healthInput := BuildSharedSessionBrowserHealthInput(
		evaluation.Snapshot.Summary.ActiveNodeRunID,
		evaluation.Snapshot.Summary.RouteTargetCount,
		"",
		"",
		"",
		evaluation.Snapshot.Profiles,
	)
	evaluation.Health = EvaluateSharedSessionBrowserHealthForScope(
		stateRegistry,
		sessionID,
		selectedInfo,
		requestedProfile,
		healthInput,
		reconnectWindow,
	)
	evaluation.Coordination = EvaluateSharedSessionBrowserCoordinationEvaluationForScope(
		stateRegistry,
		sessionID,
		selectedInfo,
		requestedProfile,
		healthInput,
		BuildSharedSessionBrowserCoordinationInput(
			evaluation.Snapshot.Summary.ActiveNodeRunID,
			evaluation.Snapshot.Summary.RouteTargetCount,
			evaluation.Snapshot.SelectedProfileSelection,
			evaluation.Snapshot.SelectedTargetSelection,
			evaluation.Snapshot.Profiles,
		),
		SharedSessionBrowserRouteCoordinationInputs(evaluation.Routes),
		reconnectWindow,
		evaluation.Snapshot.Summary.BlockedAutoFollowRouteCount > 0,
	)
	return sharedSessionBrowserFinalizeBindingEvaluationHandoff(evaluation)
}

func refreshSharedSessionBrowserBindingEvaluationProfiles(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	evaluation SharedSessionBrowserBindingEvaluation,
	profiles []SharedSessionBrowserProfileState,
	reconnectWindow time.Duration,
) SharedSessionBrowserBindingEvaluation {
	snapshot := evaluation.Snapshot
	snapshot.Profiles = profiles
	summary := snapshot.Summary
	summary.BrowserProfileCount = len(snapshot.Profiles)
	summary.ActiveBrowserProfile, summary.BrowserProfileStatusCounts = SummarizeSharedSessionBrowserProfiles(snapshot.Profiles)
	snapshot.Summary = summary
	if strings.TrimSpace(snapshot.CurrentTargetID) == "" {
		snapshot.CurrentTargetID = strings.TrimSpace(snapshot.Summary.CurrentTargetID)
	}
	referenceTime := SharedSessionBrowserLatestObservedAt(snapshot.Profiles)

	healthInput := BuildSharedSessionBrowserHealthInputAt(
		snapshot.Summary.ActiveNodeRunID,
		snapshot.Summary.RouteTargetCount,
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
		snapshot.Profiles,
		referenceTime,
	)
	healthInput = ApplySharedSessionBrowserHealthStoredSummary(healthInput, evaluation.Health.Summary)

	evaluation.Snapshot = snapshot
	evaluation.ReferenceTime = referenceTime
	sessionID = strings.TrimSpace(sessionID)
	if registry != nil && sessionID != "" {
		evaluation.Health = EvaluateSharedSessionBrowserHealthForScope(
			registry,
			sessionID,
			selectedInfo,
			requestedProfile,
			healthInput,
			reconnectWindow,
		)
		evaluation.Coordination = EvaluateSharedSessionBrowserCoordinationEvaluationForScope(
			registry,
			sessionID,
			selectedInfo,
			requestedProfile,
			healthInput,
			BuildSharedSessionBrowserCoordinationInput(
				snapshot.Summary.ActiveNodeRunID,
				snapshot.Summary.RouteTargetCount,
				snapshot.SelectedProfileSelection,
				snapshot.SelectedTargetSelection,
				snapshot.Profiles,
			),
			SharedSessionBrowserRouteCoordinationInputs(evaluation.Routes),
			reconnectWindow,
			snapshot.Summary.BlockedAutoFollowRouteCount > 0,
		)
		return sharedSessionBrowserFinalizeBindingEvaluationHandoff(evaluation)
	}

	evaluation.Health = EvaluateSharedSessionBrowserHealthForInputScope(
		healthInput,
		selectedInfo,
		requestedProfile,
		reconnectWindow,
	)
	evaluation.Coordination = EvaluateSharedSessionBrowserCoordinationEvaluation(SharedSessionBrowserCoordinationEvaluationInput{
		Coordination: BuildSharedSessionBrowserCoordinationInput(
			snapshot.Summary.ActiveNodeRunID,
			snapshot.Summary.RouteTargetCount,
			snapshot.SelectedProfileSelection,
			snapshot.SelectedTargetSelection,
			snapshot.Profiles,
		),
		Routes:            SharedSessionBrowserRouteCoordinationInputs(evaluation.Routes),
		HealthEvaluation:  evaluation.Health,
		BlockedAutoFollow: snapshot.Summary.BlockedAutoFollowRouteCount > 0,
	})
	return sharedSessionBrowserFinalizeBindingEvaluationHandoff(evaluation)
}

func firstNonEmptyBindingString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func mergeSharedSessionBrowserBindingProfile(items []SharedSessionBrowserProfileState, current SharedSessionBrowserProfileState) []SharedSessionBrowserProfileState {
	if len(items) == 0 {
		return []SharedSessionBrowserProfileState{current}
	}
	out := append([]SharedSessionBrowserProfileState(nil), items...)
	for i, item := range out {
		if !sameSharedSessionBrowserBindingProfileState(item, current) {
			continue
		}
		if item.Backend == "" && current.Backend != "" {
			item.Backend = current.Backend
		}
		if item.Profile == "" && current.Profile != "" {
			item.Profile = current.Profile
		}
		if item.RuntimeTarget == "" && current.RuntimeTarget != "" {
			item.RuntimeTarget = current.RuntimeTarget
		}
		if item.BrowserApp == "" && current.BrowserApp != "" {
			item.BrowserApp = current.BrowserApp
		}
		if item.Note == "" && current.Note != "" {
			item.Note = current.Note
		}
		out[i] = item
		return out
	}
	return append(out, current)
}

func sameSharedSessionBrowserBindingProfileState(left SharedSessionBrowserProfileState, right SharedSessionBrowserProfileState) bool {
	return strings.EqualFold(strings.TrimSpace(left.Backend), strings.TrimSpace(right.Backend)) &&
		strings.EqualFold(strings.TrimSpace(left.RuntimeTarget), strings.TrimSpace(right.RuntimeTarget)) &&
		strings.EqualFold(strings.TrimSpace(left.Profile), strings.TrimSpace(right.Profile)) &&
		strings.EqualFold(strings.TrimSpace(left.BrowserApp), strings.TrimSpace(right.BrowserApp))
}

func sharedSessionBrowserHealthSummaryState(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.State)
}

func sharedSessionBrowserHealthSummaryReason(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.Reason)
}

func sharedSessionBrowserHealthSummaryRecoveryAction(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.RecoveryAction)
}

func sharedSessionBrowserHealthSummaryReconnectHint(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.ReconnectHint)
}

func sharedSessionBrowserHealthSummaryDisconnectCount(summary *SharedSessionBrowserHealthSummary) int {
	if summary == nil {
		return 0
	}
	return summary.DisconnectCount
}

func sharedSessionBrowserHealthSummaryDisconnectBurstCount(summary *SharedSessionBrowserHealthSummary) int {
	if summary == nil {
		return 0
	}
	return summary.DisconnectBurstCount
}

func sharedSessionBrowserHealthSummaryDisconnectBurstWindowMs(summary *SharedSessionBrowserHealthSummary) int {
	if summary == nil {
		return 0
	}
	return summary.DisconnectBurstWindowMs
}

func sharedSessionBrowserHealthSummaryCooldownRemainingMs(summary *SharedSessionBrowserHealthSummary) int {
	if summary == nil {
		return 0
	}
	return summary.CooldownRemainingMs
}

func sharedSessionBrowserHealthSummaryRetryBackoffRemainingMs(summary *SharedSessionBrowserHealthSummary) int {
	if summary == nil {
		return 0
	}
	return summary.RetryBackoffRemainingMs
}

func sharedSessionBrowserHealthSummaryRestartAttemptCount(summary *SharedSessionBrowserHealthSummary) int {
	if summary == nil {
		return 0
	}
	return summary.RestartAttemptCount
}

func sharedSessionBrowserHealthSummaryRestartFailureCount(summary *SharedSessionBrowserHealthSummary) int {
	if summary == nil {
		return 0
	}
	return summary.RestartFailureCount
}

func sharedSessionBrowserHealthSummaryLastDisconnectUnixMilli(summary *SharedSessionBrowserHealthSummary) int64 {
	if summary == nil {
		return 0
	}
	return summary.LastDisconnectUnixMilli
}

func sharedSessionBrowserHealthSummaryLastReconnectUnixMilli(summary *SharedSessionBrowserHealthSummary) int64 {
	if summary == nil {
		return 0
	}
	return summary.LastReconnectUnixMilli
}

func sharedSessionBrowserHealthSummaryLastRestartAttemptUnixMilli(summary *SharedSessionBrowserHealthSummary) int64 {
	if summary == nil {
		return 0
	}
	return summary.LastRestartAttemptUnixMilli
}

func sharedSessionBrowserHealthSummaryLastRestartResult(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.LastRestartResult)
}

func sharedSessionBrowserHealthSummaryLastRestartError(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.LastRestartError)
}

func sharedSessionBrowserHealthSummaryRecommendedBackoffMs(summary *SharedSessionBrowserHealthSummary) int {
	if summary == nil {
		return 0
	}
	return summary.RecommendedBackoffMs
}

func sharedSessionBrowserHealthSummaryResolverBlockedBy(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.ResolverBlockedBy)
}

func sharedSessionBrowserHealthSummaryAmbiguityClass(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.AmbiguityClass)
}

func sharedSessionBrowserHealthSummaryCandidateKind(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.CandidateKind)
}

func sharedSessionBrowserHealthSummaryCandidateStrength(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.CandidateStrength)
}

func sharedSessionBrowserHealthSummaryRetryDisposition(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.RetryDisposition)
}

func sharedSessionBrowserHealthSummaryManualRetryHint(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.ManualRetryHint)
}

func sharedSessionBrowserHealthSummaryNextStepAlias(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	return strings.TrimSpace(summary.NextStepAlias)
}

func sharedSessionBrowserHealthSummarySpecificityFields(summary *SharedSessionBrowserHealthSummary) []string {
	if summary == nil || len(summary.SpecificityFields) == 0 {
		return nil
	}
	return append([]string(nil), summary.SpecificityFields...)
}

// ApplySharedSessionBrowserHealthStoredSummary overlays the machine-readable
// reconnect/watchdog fields from a stored shared summary onto a stable health
// input without forcing every caller to rebuild that input field-by-field.
func ApplySharedSessionBrowserHealthStoredSummary(input SharedSessionBrowserHealthInput, summary *SharedSessionBrowserHealthSummary) SharedSessionBrowserHealthInput {
	input.StoredReconnectHint = sharedSessionBrowserHealthSummaryReconnectHint(summary)
	input.StoredDisconnectCount = sharedSessionBrowserHealthSummaryDisconnectCount(summary)
	input.StoredDisconnectBurstCount = sharedSessionBrowserHealthSummaryDisconnectBurstCount(summary)
	input.StoredDisconnectBurstWindowMs = sharedSessionBrowserHealthSummaryDisconnectBurstWindowMs(summary)
	input.StoredCooldownRemainingMs = sharedSessionBrowserHealthSummaryCooldownRemainingMs(summary)
	input.StoredRetryBackoffRemainingMs = sharedSessionBrowserHealthSummaryRetryBackoffRemainingMs(summary)
	input.StoredRestartAttemptCount = sharedSessionBrowserHealthSummaryRestartAttemptCount(summary)
	input.StoredRestartFailureCount = sharedSessionBrowserHealthSummaryRestartFailureCount(summary)
	input.StoredLastDisconnectUnixMilli = sharedSessionBrowserHealthSummaryLastDisconnectUnixMilli(summary)
	input.StoredLastReconnectUnixMilli = sharedSessionBrowserHealthSummaryLastReconnectUnixMilli(summary)
	input.StoredLastRestartAttemptUnixMilli = sharedSessionBrowserHealthSummaryLastRestartAttemptUnixMilli(summary)
	input.StoredLastRestartResult = sharedSessionBrowserHealthSummaryLastRestartResult(summary)
	input.StoredLastRestartError = sharedSessionBrowserHealthSummaryLastRestartError(summary)
	input.StoredRecommendedBackoffMs = sharedSessionBrowserHealthSummaryRecommendedBackoffMs(summary)
	return input
}

// ApplySharedSessionBrowserHealthSummary overlays the full stored health
// summary, including typed blocker state/reason/recovery guidance, onto a
// stable health input.
func ApplySharedSessionBrowserHealthSummary(input SharedSessionBrowserHealthInput, summary *SharedSessionBrowserHealthSummary) SharedSessionBrowserHealthInput {
	if summary == nil {
		return input
	}
	input.StoredState = sharedSessionBrowserHealthSummaryState(summary)
	input.StoredReason = sharedSessionBrowserHealthSummaryReason(summary)
	input.StoredRecoveryAction = sharedSessionBrowserHealthSummaryRecoveryAction(summary)
	input.StoredResolverBlockedBy = sharedSessionBrowserHealthSummaryResolverBlockedBy(summary)
	input.StoredAmbiguityClass = sharedSessionBrowserHealthSummaryAmbiguityClass(summary)
	input.StoredCandidateKind = sharedSessionBrowserHealthSummaryCandidateKind(summary)
	input.StoredCandidateStrength = sharedSessionBrowserHealthSummaryCandidateStrength(summary)
	input.StoredRetryDisposition = sharedSessionBrowserHealthSummaryRetryDisposition(summary)
	input.StoredManualRetryHint = sharedSessionBrowserHealthSummaryManualRetryHint(summary)
	input.StoredNextStepAlias = sharedSessionBrowserHealthSummaryNextStepAlias(summary)
	input.StoredSpecificityFields = sharedSessionBrowserHealthSummarySpecificityFields(summary)
	return ApplySharedSessionBrowserHealthStoredSummary(input, summary)
}

// SharedSessionBrowserHealthSummaryFromBrowserSessionHealth lowers the
// backend/runtime-facing session_health payload into the shared health summary
// contract consumed by recovery and coordination helpers.
func SharedSessionBrowserHealthSummaryFromBrowserSessionHealth(summary *BrowserSessionHealthSummary) *SharedSessionBrowserHealthSummary {
	if summary == nil {
		return nil
	}
	return &SharedSessionBrowserHealthSummary{
		State:                       strings.TrimSpace(summary.State),
		Reason:                      strings.TrimSpace(summary.Reason),
		RecoveryAction:              strings.TrimSpace(summary.RecoveryAction),
		ReconnectHint:               strings.TrimSpace(summary.ReconnectHint),
		DisconnectCount:             summary.DisconnectCount,
		DisconnectBurstCount:        summary.DisconnectBurstCount,
		DisconnectBurstWindowMs:     summary.DisconnectBurstWindowMs,
		CooldownRemainingMs:         summary.CooldownRemainingMs,
		RetryBackoffRemainingMs:     summary.RetryBackoffRemainingMs,
		RestartAttemptCount:         summary.RestartAttemptCount,
		RestartFailureCount:         summary.RestartFailureCount,
		LastDisconnectUnixMilli:     summary.LastDisconnectUnixMilli,
		LastReconnectUnixMilli:      summary.LastReconnectUnixMilli,
		LastRestartAttemptUnixMilli: summary.LastRestartAttemptUnixMilli,
		LastRestartResult:           strings.TrimSpace(summary.LastRestartResult),
		LastRestartError:            strings.TrimSpace(summary.LastRestartError),
		RecommendedBackoffMs:        summary.RecommendedBackoffMs,
	}
}

// BuildSharedSessionBrowserHealthInput projects shared binding lifecycle state
// into the stable health input consumed by lifecycle evaluation/recovery rules.
func BuildSharedSessionBrowserHealthInput(activeNodeRunID string, routeTargetCount int, storedState string, storedReason string, storedRecoveryAction string, profiles []SharedSessionBrowserProfileState) SharedSessionBrowserHealthInput {
	return BuildSharedSessionBrowserHealthInputAt(activeNodeRunID, routeTargetCount, storedState, storedReason, storedRecoveryAction, "", "", "", "", "", "", "", nil, profiles, time.Time{})
}

// BuildSharedSessionBrowserHealthInputAt projects shared binding lifecycle
// state into the stable health input consumed by lifecycle evaluation/recovery
// rules, using an explicit source-time reference when a watch cycle already
// owns the observation clock.
func BuildSharedSessionBrowserHealthInputAt(
	activeNodeRunID string,
	routeTargetCount int,
	storedState string,
	storedReason string,
	storedRecoveryAction string,
	storedResolverBlockedBy string,
	storedAmbiguityClass string,
	storedCandidateKind string,
	storedCandidateStrength string,
	storedRetryDisposition string,
	storedManualRetryHint string,
	storedNextStepAlias string,
	storedSpecificityFields []string,
	profiles []SharedSessionBrowserProfileState,
	referenceTime time.Time,
) SharedSessionBrowserHealthInput {
	input := SharedSessionBrowserHealthInput{
		ActiveNodeRunID:         strings.TrimSpace(activeNodeRunID),
		RouteTargetCount:        routeTargetCount,
		StoredState:             strings.TrimSpace(storedState),
		StoredReason:            strings.TrimSpace(storedReason),
		StoredRecoveryAction:    strings.TrimSpace(storedRecoveryAction),
		StoredResolverBlockedBy: strings.TrimSpace(storedResolverBlockedBy),
		StoredAmbiguityClass:    strings.TrimSpace(storedAmbiguityClass),
		StoredCandidateKind:     strings.TrimSpace(storedCandidateKind),
		StoredCandidateStrength: strings.TrimSpace(storedCandidateStrength),
		StoredRetryDisposition:  strings.TrimSpace(storedRetryDisposition),
		StoredManualRetryHint:   strings.TrimSpace(storedManualRetryHint),
		StoredNextStepAlias:     strings.TrimSpace(storedNextStepAlias),
		ReferenceTime:           referenceTime,
	}
	if len(storedSpecificityFields) > 0 {
		input.StoredSpecificityFields = append([]string(nil), storedSpecificityFields...)
	}
	if len(profiles) > 0 {
		input.Profiles = append([]SharedSessionBrowserProfileState(nil), profiles...)
	}
	return input
}

func SharedSessionBrowserLatestObservedAt(profiles []SharedSessionBrowserProfileState) time.Time {
	var latest time.Time
	for _, profile := range profiles {
		if profile.ObservedAt.After(latest) {
			latest = profile.ObservedAt
		}
	}
	return latest
}

// BuildSharedSessionBrowserCoordinationInput projects shared binding state
// into the stable coordination input consumed by shared coordination rules.
func BuildSharedSessionBrowserCoordinationInput(activeNodeRunID string, routeTargetCount int, profileSelection *SharedSessionBrowserProfileSelection, targetSelection *BrowserSessionTargetSelection, profiles []SharedSessionBrowserProfileState) SharedSessionBrowserCoordinationInput {
	input := SharedSessionBrowserCoordinationInput{
		ActiveNodeRunID:  strings.TrimSpace(activeNodeRunID),
		RouteTargetCount: routeTargetCount,
	}
	if profileSelection != nil {
		input.SelectedBrowserProfile = strings.TrimSpace(profileSelection.Profile)
	}
	if targetSelection != nil {
		input.SelectedBrowserTargetID = strings.TrimSpace(targetSelection.ID)
	}
	if len(profiles) > 0 {
		input.Profiles = append([]SharedSessionBrowserProfileState(nil), profiles...)
	}
	return input
}

// SharedSessionBrowserRouteCoordinationInputs projects shared route snapshots
// into the compact posture flags consumed by coordination overlays.
func SharedSessionBrowserRouteCoordinationInputs(routes []SharedSessionBrowserRouteSnapshot) []SharedSessionBrowserRouteCoordinationInput {
	if len(routes) == 0 {
		return nil
	}
	out := make([]SharedSessionBrowserRouteCoordinationInput, 0, len(routes))
	for _, route := range routes {
		out = append(out, SharedSessionBrowserRouteCoordinationInput{
			FollowPolicyState: strings.TrimSpace(route.FollowPolicyState),
			ManagedRuntime:    sharedSessionBrowserManagedRuntimeTarget(route.RuntimeTarget),
		})
	}
	return out
}

func sharedSessionBrowserManagedRuntimeTarget(runtimeTarget string) bool {
	switch strings.ToLower(strings.TrimSpace(runtimeTarget)) {
	case "node", "sandbox":
		return true
	default:
		return false
	}
}

// SummarizeSharedSessionBrowserBinding aggregates route/run/profile snapshots
// into the stable counters and active identifiers used by runtime payloads.
func SummarizeSharedSessionBrowserBinding(routes []SharedSessionBrowserRouteSnapshot, runs []SharedSessionRunInfo, profiles []SharedSessionBrowserProfileState) SharedSessionBrowserBindingSummary {
	activeNodeRunID, nodeRunStatusCounts := SummarizeSharedSessionRuns(runs)
	activeBrowserProfile, browserProfileStatusCounts := SummarizeSharedSessionBrowserProfiles(profiles)
	summary := SharedSessionBrowserBindingSummary{
		NodeRunCount:               len(runs),
		ActiveNodeRunID:            strings.TrimSpace(activeNodeRunID),
		NodeRunStatusCounts:        nodeRunStatusCounts,
		BrowserProfileCount:        len(profiles),
		ActiveBrowserProfile:       strings.TrimSpace(activeBrowserProfile),
		BrowserProfileStatusCounts: browserProfileStatusCounts,
	}
	for _, route := range routes {
		summary.RouteTargetCount += len(route.Targets)
		if route.PendingTargetReview != nil {
			summary.PendingTargetReviewCount++
		}
		if strings.TrimSpace(route.FollowPolicyState) != "" && !strings.EqualFold(strings.TrimSpace(route.FollowPolicyState), "auto_follow_allowed") {
			summary.BlockedAutoFollowRouteCount++
		}
		if strings.EqualFold(strings.TrimSpace(route.PopupPolicyState), "popup_storm_review_required") {
			summary.PopupStormRouteCount++
		}
		if summary.CurrentTargetID == "" {
			summary.CurrentTargetID = strings.TrimSpace(route.CurrentTargetID)
		}
	}
	return summary
}

// SummarizeSharedSessionRuns returns the preferred active run and per-status
// counts for a session-scoped node-run snapshot.
func SummarizeSharedSessionRuns(items []SharedSessionRunInfo) (string, map[string]int) {
	if len(items) == 0 {
		return "", nil
	}
	counts := map[string]int{}
	activeID := ""
	for _, item := range items {
		status := strings.ToLower(strings.TrimSpace(item.Status))
		if status == "" {
			status = "unknown"
		}
		counts[status]++
		if activeID == "" && (status == "running" || status == "pending" || status == "starting") {
			activeID = strings.TrimSpace(item.RunID)
		}
	}
	if activeID == "" {
		activeID = strings.TrimSpace(items[0].RunID)
	}
	return activeID, counts
}

// SummarizeSharedSessionBrowserProfiles returns the preferred active managed
// browser profile and per-status counts for a scoped lifecycle snapshot.
func SummarizeSharedSessionBrowserProfiles(items []SharedSessionBrowserProfileState) (string, map[string]int) {
	if len(items) == 0 {
		return "", nil
	}
	counts := map[string]int{}
	activeProfile := ""
	for _, item := range items {
		status := strings.ToLower(strings.TrimSpace(item.Status))
		if status == "" {
			status = "unknown"
		}
		counts[status]++
		if activeProfile == "" && (item.Running || item.Connected || status == "running" || status == "started") {
			activeProfile = strings.TrimSpace(item.Profile)
		}
	}
	if activeProfile == "" {
		activeProfile = strings.TrimSpace(items[0].Profile)
	}
	return activeProfile, counts
}
