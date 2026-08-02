package browserruntime

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SharedSessionBrowserExecutionResult is the lifecycle-owned execution outcome
// shared across prepare/restart/start/stop/teardown control paths.
type SharedSessionBrowserExecutionResult struct {
	Profile                  string
	Profiles                 *BrowserProfilesResult
	ProfilesObservedAt       time.Time
	ProfileStatus            BrowserProfileStatusResult
	ProfileStatusObservedAt  time.Time
	Decision                 string
	Ready                    bool
	InvalidateSessionTargets bool
	InvalidateSessionProfile bool
}

// SharedSessionBrowserExecutionRequest provides the route-scoped execution
// context needed to resolve effective lifecycle state around runtime actions.
type SharedSessionBrowserExecutionRequest struct {
	Registry             SharedSessionBrowserStateRegistry
	SessionID            string
	RequestedProfile     string
	SelectedInfo         BrowserRuntimeInfo
	Force                bool
	ActiveNodeRunID      string
	ActiveBrowserProfile string
	HealthInput          SharedSessionBrowserHealthInput
	ReconnectWindow      time.Duration
}

// BuildSharedSessionBrowserExecutionRequest assembles the shared execution
// request used by managed-browser lifecycle actions from already projected
// binding/session state.
func BuildSharedSessionBrowserExecutionRequest(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	requestedProfile string,
	selectedInfo BrowserRuntimeInfo,
	force bool,
	activeNodeRunID string,
	activeBrowserProfile string,
	healthInput SharedSessionBrowserHealthInput,
	reconnectWindow time.Duration,
) SharedSessionBrowserExecutionRequest {
	return SharedSessionBrowserExecutionRequest{
		Registry:             registry,
		SessionID:            strings.TrimSpace(sessionID),
		RequestedProfile:     strings.TrimSpace(requestedProfile),
		SelectedInfo:         selectedInfo,
		Force:                force,
		ActiveNodeRunID:      strings.TrimSpace(activeNodeRunID),
		ActiveBrowserProfile: strings.TrimSpace(activeBrowserProfile),
		HealthInput:          healthInput,
		ReconnectWindow:      reconnectWindow,
	}
}

// SharedSessionBrowserProfileCreateRequest carries the create-profile inputs
// needed by the shared execution contract.
type SharedSessionBrowserProfileCreateRequest struct {
	RequestedProfile string
	SelectedInfo     BrowserRuntimeInfo
	BrowserApp       string
	Color            string
	CopyFrom         string
}

// SharedSessionBrowserExecutionResolution captures the lifecycle-owned resolved
// status and optional scoped snapshot after applying an execution result to the
// shared session registry contract.
type SharedSessionBrowserExecutionResolution struct {
	ResolvedStatus BrowserProfileStatusResult
	SyncedState    SharedSessionBrowserProfileState
	HasSyncedState bool
	Snapshot       []SharedSessionBrowserProfileState
}

// SharedSessionBrowserExecutionCleanup captures the session-scoped selection
// and route cleanup produced by an execution result after lifecycle resolution.
type SharedSessionBrowserExecutionCleanup struct {
	ClearedSessionTargets int
	ClearedSessionProfile bool
}

// SharedSessionBrowserExecutionApplication captures the shared execution
// result after lifecycle resolution, scoped cleanup, and projected profile
// snapshot fallback have been applied.
type SharedSessionBrowserExecutionApplication struct {
	Resolution        SharedSessionBrowserExecutionResolution
	Cleanup           SharedSessionBrowserExecutionCleanup
	ProjectedProfiles []SharedSessionBrowserProjectedProfileState
}

func sharedSessionBrowserExecutionWatchManager(control BrowserRuntimeControlBackend, req SharedSessionBrowserExecutionRequest) SharedSessionBrowserWatchManager {
	return sharedSessionBrowserObserverManager(nil, nil, req.Registry, req.ReconnectWindow).Bind(control)
}

// ExecuteSharedSessionBrowserEnsurePrepared applies the shared prepare/ensure
// orchestration for a managed browser profile, including prepared-profile
// loading, recovery assessment, and lifecycle-owned ready/status resolution.
func ExecuteSharedSessionBrowserEnsurePrepared(ctx context.Context, control BrowserRuntimeControlBackend, req SharedSessionBrowserExecutionRequest) (SharedSessionBrowserExecutionResult, error) {
	profile, profilesResult, profilesObservedAt, err := LoadSharedSessionPreparedProfile(ctx, control, req.RequestedProfile, req.SelectedInfo)
	result := SharedSessionBrowserExecutionResult{
		Profile:            profile,
		Profiles:           profilesResult,
		ProfilesObservedAt: profilesObservedAt,
	}
	if err != nil {
		return result, err
	}
	return executeSharedSessionBrowserEnsurePrepared(ctx, sharedSessionBrowserExecutionWatchManager(control, req), req, result)
}

// ExecuteSharedSessionBrowserStart applies the shared direct-start lifecycle
// orchestration for a managed browser profile.
func ExecuteSharedSessionBrowserStart(ctx context.Context, control BrowserRuntimeControlBackend, req SharedSessionBrowserExecutionRequest) (SharedSessionBrowserExecutionResult, error) {
	profile := firstNonEmptyString(strings.TrimSpace(req.RequestedProfile), strings.TrimSpace(req.SelectedInfo.Profile))
	result := SharedSessionBrowserExecutionResult{
		Profile: profile,
	}
	watchManager := sharedSessionBrowserExecutionWatchManager(control, req)
	if watchManager.Control == nil {
		result.Decision = "start_unsupported"
		return result, nil
	}
	startObservation := watchManager.ObserveExecutionStart(ctx, req, profile, "started", resolveSharedSessionBrowserExecutionStatus(req, profile, result.ProfileStatus))
	if startObservation.Err != nil {
		result.Decision = "start_failed"
		result.ProfileStatus = startObservation.Status
		result.ProfileStatusObservedAt = startObservation.ObservedAt
		return result, startObservation.Err
	}
	result.Profile = startObservation.Profile
	result.Decision = "started"
	result.ProfileStatus = startObservation.Status
	result.ProfileStatusObservedAt = startObservation.ObservedAt
	result.Ready = startObservation.Ready
	return result, nil
}

// ExecuteSharedSessionBrowserCreateProfile applies the shared create-profile
// orchestration for a managed browser profile.
func ExecuteSharedSessionBrowserCreateProfile(ctx context.Context, manager BrowserRuntimeProfileManagementBackend, req SharedSessionBrowserProfileCreateRequest) (SharedSessionBrowserExecutionResult, error) {
	profile := strings.TrimSpace(req.RequestedProfile)
	result := SharedSessionBrowserExecutionResult{
		Profile: profile,
		Ready:   false,
	}
	if profile == "" {
		result.Decision = "create_profile_missing"
		return result, fmt.Errorf("browser_runtime: profile is required for action create_profile")
	}
	if manager == nil {
		result.Decision = "create_profile_unsupported"
		return result, nil
	}
	createResult, err := manager.RuntimeCreateProfile(ctx, BrowserProfileCreateRequest{
		Profile:    profile,
		BrowserApp: strings.TrimSpace(req.BrowserApp),
		Color:      strings.TrimSpace(req.Color),
		CopyFrom:   strings.TrimSpace(req.CopyFrom),
	})
	if err != nil {
		result.Decision = "create_profile_failed"
		return result, err
	}
	result.Profile = firstNonEmptyString(strings.TrimSpace(createResult.Profile), profile)
	result.ProfileStatus = createResult
	result.Decision = firstNonEmptyString(strings.TrimSpace(createResult.Status), "created")
	result.Ready = result.Profile != ""
	if sharedSessionBrowserProfileStatusResultEmpty(result.ProfileStatus) {
		result.ProfileStatus = BrowserProfileStatusResult{
			Backend:    strings.TrimSpace(req.SelectedInfo.Backend),
			Profile:    result.Profile,
			BrowserApp: strings.TrimSpace(req.BrowserApp),
			Status:     "created",
		}
	}
	return result, nil
}

// ExecuteSharedSessionBrowserRestart applies the shared refresh/restart
// orchestration for a managed browser profile, including recovery assessment
// and lifecycle-owned ready/status resolution.
func ExecuteSharedSessionBrowserRestart(ctx context.Context, control BrowserRuntimeControlBackend, req SharedSessionBrowserExecutionRequest) (SharedSessionBrowserExecutionResult, error) {
	profile, profilesResult, profilesObservedAt, err := LoadSharedSessionPreparedProfile(ctx, control, req.RequestedProfile, req.SelectedInfo)
	result := SharedSessionBrowserExecutionResult{
		Profile:            profile,
		Profiles:           profilesResult,
		ProfilesObservedAt: profilesObservedAt,
	}
	if err != nil {
		return result, err
	}
	watchManager := sharedSessionBrowserExecutionWatchManager(control, req)
	if watchManager.Control == nil {
		result.Decision = "restart_unsupported"
		return result, nil
	}
	if !req.Force && strings.TrimSpace(req.ActiveNodeRunID) != "" {
		result.ProfileStatus = resolveSharedSessionBrowserExecutionStatus(req, profile, BrowserProfileStatusResult{
			Backend: strings.TrimSpace(req.SelectedInfo.Backend),
			Profile: strings.TrimSpace(profile),
		})
		result.Decision = "restart_blocked_active_node_run"
		return result, nil
	}
	if !req.Force {
		_, assessment := assessSharedSessionBrowserExecutionRecovery(req, profile, BrowserProfileStatusResult{})
		if assessment.ReconnectInProgress && assessment.HasSyntheticStatus {
			result.ProfileStatus = assessment.SyntheticStatus
			result.Decision = "restart_reconnect_in_progress"
			return result, nil
		}
	}
	statusObservation := watchManager.ObserveExecutionStatus(ctx, req, profile, result.ProfileStatus)
	if statusObservation.StatusErr != nil {
		result.ProfileStatus = statusObservation.ResolvedStatus
		result.ProfileStatusObservedAt = statusObservation.ObservedAt
		result.Decision = "restart_status_failed"
		return result, statusObservation.StatusErr
	}
	result.ProfileStatus = statusObservation.Status
	result.ProfileStatusObservedAt = statusObservation.ObservedAt
	evaluation, assessment := assessSharedSessionBrowserExecutionRecovery(req, profile, statusObservation.Status)
	effectiveStatus := assessment.EffectiveStatus
	result.ProfileStatus = effectiveStatus
	if !req.Force {
		if decision, blocked := SharedSessionBrowserExecutionBlockedDecision(evaluation); blocked {
			result.Decision = decision
			return result, nil
		}
	}
	result.InvalidateSessionTargets = true
	restarted := false
	if assessment.ShouldStopBeforeRecovery {
		stopObservation := watchManager.ObserveExecutionStop(ctx, req, profile, "stopped", effectiveStatus)
		if stopObservation.Err != nil {
			result.Decision = "restart_stop_failed"
			return result, stopObservation.Err
		}
		result.ProfileStatus = stopObservation.Status
		result.ProfileStatusObservedAt = stopObservation.ObservedAt
		restarted = true
	}
	decision := "restart_started"
	if restarted {
		decision = "restarted"
	}
	startObservation := watchManager.ObserveExecutionStart(ctx, req, profile, decision, result.ProfileStatus)
	if startObservation.Err != nil {
		result.Decision = "restart_start_failed"
		return result, startObservation.Err
	}
	result.Profile = startObservation.Profile
	result.Decision = decision
	result.ProfileStatus = startObservation.Status
	result.ProfileStatusObservedAt = startObservation.ObservedAt
	result.Ready = startObservation.Ready
	return result, nil
}

// ExecuteSharedSessionBrowserStop applies the shared stop orchestration for a
// managed browser profile, including effective-status resolution and fallback
// prepared-profile loading.
func ExecuteSharedSessionBrowserStop(ctx context.Context, control BrowserRuntimeControlBackend, req SharedSessionBrowserExecutionRequest) (SharedSessionBrowserExecutionResult, error) {
	return executeSharedSessionBrowserStop(ctx, sharedSessionBrowserExecutionWatchManager(control, req), req, false)
}

// ExecuteSharedSessionBrowserTeardown applies the shared teardown orchestration
// for a managed browser profile, including effective-status resolution and
// fallback prepared-profile loading.
func ExecuteSharedSessionBrowserTeardown(ctx context.Context, control BrowserRuntimeControlBackend, req SharedSessionBrowserExecutionRequest) (SharedSessionBrowserExecutionResult, error) {
	return executeSharedSessionBrowserStop(ctx, sharedSessionBrowserExecutionWatchManager(control, req), req, true)
}

// ExecuteSharedSessionBrowserDeleteProfile applies the shared delete-profile
// orchestration for a managed browser profile, including active-node-run
// blocking and lifecycle-owned effective-status fallback.
func ExecuteSharedSessionBrowserDeleteProfile(ctx context.Context, manager BrowserRuntimeProfileManagementBackend, req SharedSessionBrowserExecutionRequest) (SharedSessionBrowserExecutionResult, error) {
	profile := strings.TrimSpace(req.RequestedProfile)
	result := SharedSessionBrowserExecutionResult{
		Profile: profile,
		Ready:   false,
	}
	if profile == "" {
		result.Decision = "delete_profile_missing"
		return result, fmt.Errorf("browser_runtime: profile is required for action delete_profile")
	}
	if !req.Force && strings.TrimSpace(req.ActiveNodeRunID) != "" {
		result.ProfileStatus = resolveSharedSessionBrowserExecutionStatus(req, profile, BrowserProfileStatusResult{
			Backend: strings.TrimSpace(req.SelectedInfo.Backend),
			Profile: profile,
		})
		result.Decision = "delete_profile_blocked_active_node_run"
		return result, nil
	}
	if manager == nil {
		result.Decision = "delete_profile_unsupported"
		return result, nil
	}
	deleteResult, err := manager.RuntimeDeleteProfile(ctx, BrowserProfileDeleteRequest{
		Profile: profile,
		Force:   req.Force,
	})
	if err != nil {
		result.Decision = "delete_profile_failed"
		return result, err
	}
	result.Profile = firstNonEmptyString(strings.TrimSpace(deleteResult.Profile), profile)
	result.ProfileStatus = deleteResult
	result.Decision = firstNonEmptyString(strings.TrimSpace(deleteResult.Status), "deleted")
	result.Ready = strings.EqualFold(result.Decision, "deleted") || strings.EqualFold(result.Decision, "delete_requested")
	result.InvalidateSessionTargets = result.Ready
	result.InvalidateSessionProfile = result.Ready
	if sharedSessionBrowserProfileStatusResultEmpty(result.ProfileStatus) {
		result.ProfileStatus = BrowserProfileStatusResult{
			Backend: strings.TrimSpace(req.SelectedInfo.Backend),
			Profile: result.Profile,
			Status:  "deleted",
		}
	}
	return result, nil
}

// ResolveSharedSessionBrowserExecutionResult applies an execution result through
// the shared registry contract when a session scope exists and otherwise
// computes the lifecycle-owned resolved status directly from the shared
// lifecycle contract.
func ResolveSharedSessionBrowserExecutionResult(registry SharedSessionBrowserStateRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, result SharedSessionBrowserExecutionResult, reconnectWindow time.Duration) SharedSessionBrowserExecutionResolution {
	return ResolveSharedSessionBrowserExecutionEvent(
		registry,
		sessionID,
		selectedInfo,
		requestedProfile,
		result,
		reconnectWindow,
	)
}

// ApplySharedSessionBrowserExecutionCleanup clears any route-scoped targets or
// remembered profile selections invalidated by an execution result using the
// lifecycle-owned resolved status as the cleanup source of truth.
func ApplySharedSessionBrowserExecutionCleanup(sessionRegistry *BrowserSessionRegistry, stateRegistry SharedSessionBrowserStateRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, result SharedSessionBrowserExecutionResult, resolution SharedSessionBrowserExecutionResolution) SharedSessionBrowserExecutionCleanup {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return SharedSessionBrowserExecutionCleanup{}
	}
	cleanup := SharedSessionBrowserExecutionCleanup{}
	if result.InvalidateSessionTargets {
		cleanup.ClearedSessionTargets = clearSharedSessionBrowserTargetsForExecution(sessionRegistry, sessionID, selectedInfo, result, resolution)
	}
	if result.InvalidateSessionProfile {
		cleanup.ClearedSessionProfile = clearSharedSessionBrowserProfileSelectionForExecution(stateRegistry, sessionID, selectedInfo, result, resolution)
	}
	return cleanup
}

// ApplySharedSessionBrowserExecutionResultWithContext applies shared lifecycle
// resolution, projected scoped profile snapshot fallback, and any route/profile
// cleanup invalidated by the execution result. When a shared run registry is
// available, execution writeback reuses the shared observer-manager seam so
// sibling providers can refresh from the same source-time event.
func ApplySharedSessionBrowserExecutionResultWithContext(sessionRegistry *BrowserSessionRegistry, runRegistry SharedSessionRunRegistry, stateRegistry SharedSessionBrowserStateRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, result SharedSessionBrowserExecutionResult, reconnectWindow time.Duration) SharedSessionBrowserExecutionApplication {
	return sharedSessionBrowserObserverManager(
		sessionRegistry,
		runRegistry,
		stateRegistry,
		reconnectWindow,
	).ApplyExecutionResult(
		sessionID,
		selectedInfo,
		requestedProfile,
		result,
	)
}

// ApplySharedSessionBrowserExecutionResult applies shared lifecycle resolution,
// projected scoped profile snapshot fallback, and any route/profile cleanup
// invalidated by the execution result.
func ApplySharedSessionBrowserExecutionResult(sessionRegistry *BrowserSessionRegistry, stateRegistry SharedSessionBrowserStateRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, result SharedSessionBrowserExecutionResult, reconnectWindow time.Duration) SharedSessionBrowserExecutionApplication {
	return ApplySharedSessionBrowserExecutionResultWithContext(
		sessionRegistry,
		nil,
		stateRegistry,
		sessionID,
		selectedInfo,
		requestedProfile,
		result,
		reconnectWindow,
	)
}

// ApplyExecutionResult applies shared lifecycle resolution, projected scoped
// profile snapshot fallback, and any route/profile cleanup invalidated by the
// execution result. When the provider has bound watch managers, execution
// writeback also invalidates their short-lived source-time caches so follow-up
// watch/inspection reads do not reuse stale session snapshots.
func (m SharedSessionBrowserObserverManager) ApplyExecutionResult(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, result SharedSessionBrowserExecutionResult) SharedSessionBrowserExecutionApplication {
	m.touchProvider()
	sessionID = strings.TrimSpace(sessionID)
	application := SharedSessionBrowserExecutionApplication{
		Resolution: ResolveSharedSessionBrowserExecutionResult(
			m.StateRegistry,
			sessionID,
			selectedInfo,
			requestedProfile,
			result,
			m.ReconnectWindow,
		),
	}
	if result.Profiles != nil && len(application.Resolution.Snapshot) > 0 {
		application.ProjectedProfiles = ProjectSharedSessionBrowserProfileSnapshot(
			application.Resolution.Snapshot,
			sharedSessionBrowserSelectedProfileForTarget(m.StateRegistry, sessionID, selectedInfo.Target),
		)
	}
	application.Cleanup = ApplySharedSessionBrowserExecutionCleanup(
		m.SessionRegistry,
		m.StateRegistry,
		sessionID,
		selectedInfo,
		result,
		application.Resolution,
	)
	if result.Profiles != nil && len(application.ProjectedProfiles) == 0 {
		application.ProjectedProfiles = SnapshotSharedSessionBrowserProjectedProfilesForScope(
			m.StateRegistry,
			sessionID,
			selectedInfo,
			requestedProfile,
		)
	}
	if result.Profiles != nil && len(application.ProjectedProfiles) == 0 {
		application.ProjectedProfiles = ProjectSharedSessionBrowserObservedProfiles(
			result.Profiles.Backend,
			selectedInfo.Target,
			result.Profiles.Profiles,
			sharedSessionBrowserSelectedProfileForTarget(m.StateRegistry, sessionID, selectedInfo.Target),
		)
	}
	if sessionID != "" {
		if !m.seedBoundManagersForExecutionResult(sessionID, selectedInfo, requestedProfile, result, application.Resolution) {
			m.refreshBoundManagersProjectionCaches()
		}
	}
	return application
}

func (m SharedSessionBrowserObserverManager) seedBoundManagersForExecutionResult(
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	result SharedSessionBrowserExecutionResult,
	resolution SharedSessionBrowserExecutionResolution,
) bool {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return false
	}
	selectedInfo = BrowserRuntimeInfo{
		Backend: firstNonEmptyString(
			strings.TrimSpace(resolution.ResolvedStatus.Backend),
			strings.TrimSpace(resolution.SyncedState.Backend),
			strings.TrimSpace(result.ProfileStatus.Backend),
			strings.TrimSpace(selectedInfo.Backend),
		),
		Profile: firstNonEmptyString(
			strings.TrimSpace(resolution.ResolvedStatus.Profile),
			strings.TrimSpace(resolution.SyncedState.Profile),
			strings.TrimSpace(result.Profile),
			strings.TrimSpace(result.ProfileStatus.Profile),
			strings.TrimSpace(requestedProfile),
			strings.TrimSpace(selectedInfo.Profile),
		),
		Target: strings.TrimSpace(selectedInfo.Target),
	}

	var rawStatus *SharedSessionBrowserRawStatusObservation
	if !sharedSessionBrowserProfileStatusResultEmpty(result.ProfileStatus) {
		observedAt := result.ProfileStatusObservedAt
		if observedAt.IsZero() {
			observedAt = time.Now()
		}
		normalized := normalizeSharedSessionBrowserRawStatusObservation(SharedSessionBrowserRawStatusObservation{
			RequestedProfile: strings.TrimSpace(requestedProfile),
			Status:           &result.ProfileStatus,
			ObservedAt:       observedAt,
		}, strings.TrimSpace(requestedProfile))
		rawStatus = &normalized
	}

	var rawProfiles *SharedSessionBrowserRawProfilesObservation
	if result.Profiles != nil {
		observedAt := result.ProfilesObservedAt
		if observedAt.IsZero() {
			observedAt = time.Now()
		}
		normalized := normalizeSharedSessionBrowserRawProfilesObservation(SharedSessionBrowserRawProfilesObservation{
			RequestedProfile: strings.TrimSpace(requestedProfile),
			Profiles:         result.Profiles,
			ObservedAt:       observedAt,
		}, strings.TrimSpace(requestedProfile))
		rawProfiles = &normalized
	}

	if rawStatus == nil || rawProfiles == nil {
		cachedStatus, cachedProfiles := m.cachedRawObservationsForRouteMutation(sessionID, selectedInfo, requestedProfile)
		if rawStatus == nil && cachedStatus != nil {
			rawStatus = cachedStatus
		}
		if rawProfiles == nil && cachedProfiles != nil {
			rawProfiles = cachedProfiles
		}
	}

	if rawStatus == nil && rawProfiles == nil {
		return false
	}

	requestedProfiles := sharedSessionBrowserRawObservationCacheKeys(
		requestedProfile,
		selectedInfo.Profile,
		strings.TrimSpace(result.Profile),
		strings.TrimSpace(result.ProfileStatus.Profile),
		func() string {
			if result.Profiles == nil {
				return ""
			}
			return strings.TrimSpace(result.Profiles.DefaultProfile)
		}(),
	)
	m.seedBoundManagersRawObservations(
		sessionID,
		selectedInfo,
		requestedProfiles,
		rawStatus,
		rawProfiles,
	)
	seedRelatedSharedSessionBrowserObserverManagersRawObservations(
		m,
		m.SessionRegistry,
		m.StateRegistry,
		sessionID,
		selectedInfo,
		requestedProfiles,
		rawStatus,
		rawProfiles,
	)
	return true
}

func executeSharedSessionBrowserEnsurePrepared(ctx context.Context, watchManager SharedSessionBrowserWatchManager, req SharedSessionBrowserExecutionRequest, result SharedSessionBrowserExecutionResult) (SharedSessionBrowserExecutionResult, error) {
	profile := strings.TrimSpace(result.Profile)
	if watchManager.Control == nil {
		result.Decision = "prepare_unsupported"
		return result, nil
	}

	statusObservation := watchManager.ObserveExecutionStatus(ctx, req, profile, result.ProfileStatus)
	if statusObservation.StatusErr != nil {
		result.ProfileStatus = statusObservation.ResolvedStatus
		result.ProfileStatusObservedAt = statusObservation.ObservedAt
		result.Decision = "status_failed"
		return result, statusObservation.StatusErr
	}
	result.ProfileStatus = statusObservation.Status
	result.ProfileStatusObservedAt = statusObservation.ObservedAt

	evaluation, assessment := assessSharedSessionBrowserExecutionRecovery(req, profile, statusObservation.Status)
	if SharedSessionBrowserStatusExplicitlyHealthy(statusObservation.Status) && SharedSessionBrowserProfileReady(statusObservation.Status) {
		result.ProfileStatus = SharedSessionBrowserLifecycleDecisionStatus(req.SelectedInfo, profile, statusObservation.Status, "already_ready")
		result.Decision = "already_ready"
		result.Ready = true
		return result, nil
	}

	effectiveStatus := assessment.EffectiveStatus
	if assessment.ReconnectInProgress && assessment.HasSyntheticStatus {
		result.ProfileStatus = effectiveStatus
		result.Decision = "restart_reconnect_in_progress"
		return result, nil
	}
	if !req.Force {
		if decision, blocked := SharedSessionBrowserExecutionBlockedDecision(evaluation); blocked {
			result.ProfileStatus = effectiveStatus
			result.Decision = decision
			return result, nil
		}
	}
	if assessment.NeedsRefreshRecovery {
		result.InvalidateSessionTargets = true
		restarted := false
		result.ProfileStatus = effectiveStatus
		if assessment.ShouldStopBeforeRecovery {
			stopObservation := watchManager.ObserveExecutionStop(ctx, req, profile, "stopped", effectiveStatus)
			if stopObservation.Err != nil {
				result.Decision = "restart_stop_failed"
				return result, stopObservation.Err
			}
			result.ProfileStatus = stopObservation.Status
			result.ProfileStatusObservedAt = stopObservation.ObservedAt
			restarted = true
		}
		decision := "restart_started"
		if restarted {
			decision = "restarted"
		}
		startObservation := watchManager.ObserveExecutionStart(ctx, req, profile, decision, result.ProfileStatus)
		if startObservation.Err != nil {
			result.Decision = "restart_start_failed"
			return result, startObservation.Err
		}
		result.Profile = startObservation.Profile
		result.Decision = decision
		result.ProfileStatus = startObservation.Status
		result.ProfileStatusObservedAt = startObservation.ObservedAt
		result.Ready = startObservation.Ready
		return result, nil
	}
	if inProgressStatus, decision, ok := SharedSessionBrowserEnsurePreparedInProgressStatus(req.SelectedInfo, profile, effectiveStatus); ok {
		result.ProfileStatus = inProgressStatus
		result.Decision = decision
		return result, nil
	}
	if SharedSessionBrowserProfileReady(effectiveStatus) {
		result.ProfileStatus = effectiveStatus
		result.Decision = "already_ready"
		result.Ready = true
		return result, nil
	}
	startObservation := watchManager.ObserveExecutionStart(ctx, req, profile, "started", effectiveStatus)
	if startObservation.Err != nil {
		result.ProfileStatus = effectiveStatus
		result.Decision = "start_failed"
		return result, startObservation.Err
	}
	result.Profile = startObservation.Profile
	result.Decision = "started"
	result.ProfileStatus = startObservation.Status
	result.ProfileStatusObservedAt = startObservation.ObservedAt
	result.Ready = startObservation.Ready
	return result, nil
}

func executeSharedSessionBrowserStop(ctx context.Context, watchManager SharedSessionBrowserWatchManager, req SharedSessionBrowserExecutionRequest, teardown bool) (SharedSessionBrowserExecutionResult, error) {
	result := SharedSessionBrowserExecutionResult{
		Profile: strings.TrimSpace(req.RequestedProfile),
	}
	if watchManager.Control == nil {
		result.Decision = sharedSessionBrowserStopDecisionPrefix(teardown) + "_unsupported"
		return result, nil
	}
	if !req.Force && strings.TrimSpace(req.ActiveNodeRunID) != "" {
		profile := strings.TrimSpace(req.RequestedProfile)
		if teardown && profile == "" {
			profile = strings.TrimSpace(req.ActiveBrowserProfile)
		}
		result.Profile = profile
		result.ProfileStatus = resolveSharedSessionBrowserExecutionStatus(req, profile, BrowserProfileStatusResult{
			Backend: strings.TrimSpace(req.SelectedInfo.Backend),
			Profile: strings.TrimSpace(profile),
		})
		result.Decision = sharedSessionBrowserBlockedDecision(teardown)
		return result, nil
	}

	profile := strings.TrimSpace(req.RequestedProfile)
	if profile == "" && strings.TrimSpace(req.ActiveBrowserProfile) != "" {
		profile = strings.TrimSpace(req.ActiveBrowserProfile)
	}
	if profile == "" {
		resolved, profilesResult, profilesObservedAt, err := LoadSharedSessionPreparedProfile(ctx, watchManager.Control, req.RequestedProfile, req.SelectedInfo)
		if err == nil {
			profile = resolved
			result.Profiles = profilesResult
			result.ProfilesObservedAt = profilesObservedAt
		}
	}
	result.Profile = profile
	if profile == "" {
		result.Decision = sharedSessionBrowserNoProfileDecision(teardown)
		result.Ready = true
		return result, nil
	}

	statusObservation := watchManager.ObserveExecutionStatus(ctx, req, profile, result.ProfileStatus)
	if statusObservation.StatusErr != nil {
		result.ProfileStatus = statusObservation.ResolvedStatus
		result.ProfileStatusObservedAt = statusObservation.ObservedAt
		result.Decision = sharedSessionBrowserStatusFailedDecision(teardown)
		return result, statusObservation.StatusErr
	}

	effectiveStatus := statusObservation.ResolvedStatus
	result.ProfileStatus = effectiveStatus
	result.ProfileStatusObservedAt = statusObservation.ObservedAt
	if !SharedSessionBrowserProfileReady(effectiveStatus) {
		result.Decision = sharedSessionBrowserAlreadyStoppedDecision(teardown)
		result.InvalidateSessionTargets = true
		result.InvalidateSessionProfile = true
		result.Ready = true
		return result, nil
	}

	stopObservation := watchManager.ObserveExecutionStop(ctx, req, profile, sharedSessionBrowserStoppedDecision(teardown), effectiveStatus)
	if stopObservation.Err != nil {
		result.ProfileStatus = effectiveStatus
		result.Decision = sharedSessionBrowserStopFailedDecision(teardown)
		return result, stopObservation.Err
	}
	result.Decision = sharedSessionBrowserStoppedDecision(teardown)
	result.ProfileStatus = stopObservation.Status
	result.ProfileStatusObservedAt = stopObservation.ObservedAt
	result.InvalidateSessionTargets = true
	result.InvalidateSessionProfile = true
	result.Ready = stopObservation.Ready
	return result, nil
}

func assessSharedSessionBrowserExecutionRecovery(req SharedSessionBrowserExecutionRequest, profile string, fallback BrowserProfileStatusResult) (SharedSessionBrowserHealthEvaluation, SharedSessionBrowserProfileRecoveryAssessment) {
	sessionID := strings.TrimSpace(req.SessionID)
	if req.Registry != nil && sessionID != "" {
		return AssessSharedSessionBrowserProfileRecoveryForScope(
			req.Registry,
			sessionID,
			req.SelectedInfo,
			profile,
			req.HealthInput,
			fallback,
			req.ReconnectWindow,
		)
	}
	return AssessSharedSessionBrowserProfileRecoveryForInputScope(
		req.HealthInput,
		req.SelectedInfo,
		profile,
		fallback,
		req.ReconnectWindow,
	)
}

func resolveSharedSessionBrowserExecutionStatus(req SharedSessionBrowserExecutionRequest, profile string, fallback BrowserProfileStatusResult) BrowserProfileStatusResult {
	sessionID := strings.TrimSpace(req.SessionID)
	if req.Registry != nil && sessionID != "" {
		return ResolveSharedSessionBrowserProfileStatusForScope(
			req.Registry,
			sessionID,
			req.SelectedInfo,
			profile,
			req.HealthInput,
			fallback,
			req.ReconnectWindow,
		)
	}
	return ResolveSharedSessionBrowserProfileStatus(
		req.HealthInput,
		req.SelectedInfo,
		profile,
		fallback,
		req.ReconnectWindow,
	)
}

func sharedSessionBrowserStopDecisionPrefix(teardown bool) string {
	if teardown {
		return "teardown"
	}
	return "stop"
}

func sharedSessionBrowserBlockedDecision(teardown bool) string {
	return sharedSessionBrowserStopDecisionPrefix(teardown) + "_blocked_active_node_run"
}

func sharedSessionBrowserNoProfileDecision(teardown bool) string {
	return sharedSessionBrowserStopDecisionPrefix(teardown) + "_no_profile"
}

func sharedSessionBrowserStatusFailedDecision(teardown bool) string {
	return sharedSessionBrowserStopDecisionPrefix(teardown) + "_status_failed"
}

func sharedSessionBrowserAlreadyStoppedDecision(teardown bool) string {
	return sharedSessionBrowserStopDecisionPrefix(teardown) + "_already_stopped"
}

func sharedSessionBrowserStopFailedDecision(teardown bool) string {
	if teardown {
		return "teardown_stop_failed"
	}
	return "stop_failed"
}

func clearSharedSessionBrowserTargetsForExecution(sessionRegistry *BrowserSessionRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, result SharedSessionBrowserExecutionResult, resolution SharedSessionBrowserExecutionResolution) int {
	if sessionRegistry == nil {
		return 0
	}
	profile := firstNonEmptyString(
		strings.TrimSpace(resolution.ResolvedStatus.Profile),
		strings.TrimSpace(resolution.SyncedState.Profile),
		strings.TrimSpace(result.Profile),
		strings.TrimSpace(result.ProfileStatus.Profile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	if profile == "" {
		return 0
	}
	return sessionRegistry.ClearRoute(sessionID, BrowserSessionRoute{
		Backend: strings.TrimSpace(firstNonEmptyString(
			strings.TrimSpace(resolution.ResolvedStatus.Backend),
			strings.TrimSpace(resolution.SyncedState.Backend),
			strings.TrimSpace(result.ProfileStatus.Backend),
			strings.TrimSpace(selectedInfo.Backend),
		)),
		Profile: profile,
		Target:  strings.TrimSpace(selectedInfo.Target),
	})
}

func clearSharedSessionBrowserProfileSelectionForExecution(stateRegistry SharedSessionBrowserStateRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, result SharedSessionBrowserExecutionResult, resolution SharedSessionBrowserExecutionResolution) bool {
	if stateRegistry == nil {
		return false
	}
	runtimeTarget := strings.TrimSpace(selectedInfo.Target)
	if runtimeTarget == "" {
		return false
	}
	selected, ok := stateRegistry.SelectedBrowserProfile(sessionID, runtimeTarget)
	if !ok {
		return false
	}
	profile := firstNonEmptyString(
		strings.TrimSpace(resolution.ResolvedStatus.Profile),
		strings.TrimSpace(resolution.SyncedState.Profile),
		strings.TrimSpace(result.Profile),
		strings.TrimSpace(result.ProfileStatus.Profile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	if profile == "" || !strings.EqualFold(strings.TrimSpace(selected.Profile), profile) {
		return false
	}
	selectedBackend := strings.TrimSpace(selected.Backend)
	resolvedBackend := strings.TrimSpace(firstNonEmptyString(
		strings.TrimSpace(resolution.ResolvedStatus.Backend),
		strings.TrimSpace(resolution.SyncedState.Backend),
		strings.TrimSpace(result.ProfileStatus.Backend),
		strings.TrimSpace(selectedInfo.Backend),
	))
	if selectedBackend != "" && resolvedBackend != "" && !strings.EqualFold(selectedBackend, resolvedBackend) {
		return false
	}
	stateRegistry.ClearSelectedBrowserProfile(sessionID, runtimeTarget)
	return true
}

func sharedSessionBrowserStoppedDecision(teardown bool) string {
	if teardown {
		return "teardown_stopped"
	}
	return "stopped"
}
