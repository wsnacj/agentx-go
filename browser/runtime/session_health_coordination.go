package browserruntime

import (
	"strings"
	"time"
)

func sharedSessionBrowserHealthCanonicalRecoveryAction(summary *SharedSessionBrowserHealthSummary) string {
	if summary == nil {
		return ""
	}
	if recovery := strings.TrimSpace(summary.RecoveryAction); recovery != "" {
		return recovery
	}
	switch strings.TrimSpace(summary.ReconnectHint) {
	case "retry_after_cooldown", "retry_after_backoff", "wait_for_restart":
		return "browser action=wait"
	case "manual_restart_required":
		return "browser action=start"
	}
	switch strings.TrimSpace(summary.State) {
	case "cooldown_active", "restart_pending":
		return "browser action=wait"
	case "restart_failed_permanent":
		return "browser action=start"
	default:
		return ""
	}
}

func RecommendSharedSessionBrowserHealthActions(blockedAutoFollow bool, evaluation SharedSessionBrowserHealthEvaluation) SharedSessionBrowserHealthActions {
	if blockedAutoFollow || evaluation.Summary == nil {
		return SharedSessionBrowserHealthActions{}
	}
	actions := SharedSessionBrowserHealthActions{}
	recovery := sharedSessionBrowserHealthCanonicalRecoveryAction(evaluation.Summary)
	switch strings.TrimSpace(evaluation.Summary.State) {
	case "profile_disconnected":
		if recovery != "" {
			actions.PrimaryAction = recovery
			actions.RestartAction = recovery
			actions.RecommendedActions = append(actions.RecommendedActions, recovery, "browser action=ensure")
		}
	case "profile_reconnecting":
		if recovery != "" {
			actions.PrimaryAction = recovery
			actions.RestartAction = recovery
			actions.RecommendedActions = append(actions.RecommendedActions, recovery, "browser action=ensure")
		}
		if !evaluation.ReconnectTimedOut {
			actions.SuppressRefresh = true
			actions.ClearRestartAction = true
		}
	case "profile_stopped":
		if recovery != "" {
			actions.PrimaryAction = recovery
			actions.RecommendedActions = append(actions.RecommendedActions, recovery)
		}
	case "stale_route_targets":
		if recovery != "" {
			actions.PrimaryAction = recovery
			actions.RecommendedActions = append(actions.RecommendedActions, recovery, "browser action=ensure")
		}
	case "cooldown_active", "restart_pending":
		if recovery != "" {
			actions.PrimaryAction = recovery
			actions.RecommendedActions = append(actions.RecommendedActions, recovery)
		}
		actions.ClearRestartAction = true
		actions.SuppressRefresh = true
	case "browser_disconnect_burst", "restart_failed":
		if recovery != "" {
			actions.PrimaryAction = recovery
			actions.RestartAction = recovery
			actions.RecommendedActions = append(actions.RecommendedActions, recovery, "browser action=ensure")
		}
	case "restart_failed_permanent":
		if recovery != "" {
			actions.PrimaryAction = recovery
			actions.RestartAction = recovery
			actions.RecommendedActions = append(actions.RecommendedActions, recovery, "browser action=ensure")
		}
	}
	actions.RecommendedActions = sharedSessionBrowserUniqueActions(actions.RecommendedActions)
	return actions
}

func EvaluateSharedSessionBrowserCoordination(input SharedSessionBrowserCoordinationInput) SharedSessionBrowserCoordinationPlan {
	input.ActiveNodeRunID = strings.TrimSpace(input.ActiveNodeRunID)
	input.SelectedBrowserProfile = strings.TrimSpace(input.SelectedBrowserProfile)
	input.SelectedBrowserTargetID = strings.TrimSpace(input.SelectedBrowserTargetID)
	browserOnNode := false
	for _, profile := range input.Profiles {
		if strings.EqualFold(strings.TrimSpace(profile.RuntimeTarget), "node") || strings.EqualFold(strings.TrimSpace(profile.Backend), "proxy") {
			browserOnNode = true
			break
		}
	}
	hasActiveNodeRun := input.ActiveNodeRunID != ""
	hasRunningBrowserProfile := sharedSessionBrowserHasRunningProfile(input.Profiles)
	state := "idle"
	switch {
	case browserOnNode && hasActiveNodeRun && hasRunningBrowserProfile:
		state = "coordinated"
	case browserOnNode && hasRunningBrowserProfile:
		state = "browser_ready"
	case hasActiveNodeRun:
		state = "nodes_active"
	case browserOnNode || len(input.Profiles) > 0 || input.RouteTargetCount > 0:
		state = "browser_attached"
	}
	needsSessionSync := browserOnNode &&
		hasRunningBrowserProfile &&
		(input.SelectedBrowserProfile == "" || (input.RouteTargetCount > 0 && input.SelectedBrowserTargetID == ""))
	plan := SharedSessionBrowserCoordinationPlan{
		State:                    state,
		BrowserOnNode:            browserOnNode,
		HasActiveNodeRun:         hasActiveNodeRun,
		HasRunningBrowserProfile: hasRunningBrowserProfile,
		NeedsSessionSync:         needsSessionSync,
	}
	switch state {
	case "coordinated":
		if needsSessionSync {
			plan.RecommendedBrowserActions = append(plan.RecommendedBrowserActions, "browser action=sync")
		}
		plan.RecommendedBrowserActions = append(plan.RecommendedBrowserActions, "browser")
		plan.RecommendedNodeActions = append(plan.RecommendedNodeActions, "nodes action=run_status", "nodes action=run_wait", "nodes action=run_cancel")
	case "browser_ready":
		if needsSessionSync {
			plan.RecommendedBrowserActions = append(plan.RecommendedBrowserActions, "browser action=sync")
		}
		plan.RecommendedBrowserActions = append(plan.RecommendedBrowserActions, "browser")
		if browserOnNode && hasRunningBrowserProfile && !hasActiveNodeRun {
			plan.RecommendedBrowserActions = append(plan.RecommendedBrowserActions, "browser action=refresh")
		}
		plan.RecommendedNodeActions = append(plan.RecommendedNodeActions, "nodes action=run")
	case "nodes_active":
		plan.RecommendedBrowserActions = append(plan.RecommendedBrowserActions, "browser action=ensure", "browser")
		plan.RecommendedNodeActions = append(plan.RecommendedNodeActions, "nodes action=run_status", "nodes action=run_wait", "nodes action=run_cancel")
	case "browser_attached":
		plan.RecommendedBrowserActions = append(plan.RecommendedBrowserActions, "browser action=ensure", "browser")
		if browserOnNode {
			plan.RecommendedNodeActions = append(plan.RecommendedNodeActions, "nodes action=run")
		}
	}
	plan.RecommendedBrowserActions = sharedSessionBrowserUniqueActions(plan.RecommendedBrowserActions)
	plan.RecommendedNodeActions = sharedSessionBrowserUniqueActions(plan.RecommendedNodeActions)
	if needsSessionSync {
		plan.SyncAction = "browser action=sync"
	}
	if browserOnNode && !hasRunningBrowserProfile {
		plan.PrepareAction = "browser action=ensure"
	}
	if browserOnNode && hasRunningBrowserProfile && !hasActiveNodeRun {
		plan.RestartAction = "browser action=refresh"
		plan.TeardownAction = "browser action=teardown"
	}
	if len(plan.RecommendedBrowserActions) > 0 {
		plan.PrimaryBrowserAction = plan.RecommendedBrowserActions[0]
	}
	if len(plan.RecommendedNodeActions) > 0 {
		plan.PrimaryNodeAction = plan.RecommendedNodeActions[0]
	}
	switch state {
	case "coordinated":
		if needsSessionSync && plan.PrimaryBrowserAction != "" {
			plan.NextStep = plan.PrimaryBrowserAction
		} else {
			plan.NextStep = plan.PrimaryNodeAction
		}
	case "nodes_active":
		plan.NextStep = plan.PrimaryNodeAction
	default:
		plan.NextStep = firstNonEmptyString(plan.PrimaryBrowserAction, plan.PrimaryNodeAction)
	}
	return plan
}

func EvaluateSharedSessionBrowserFollowPolicyActions(routes []SharedSessionBrowserRouteCoordinationInput) SharedSessionBrowserCoordinationActions {
	selectedState := ""
	for _, route := range routes {
		state := strings.TrimSpace(route.FollowPolicyState)
		if sharedSessionBrowserFollowPolicyPriority(state) > sharedSessionBrowserFollowPolicyPriority(selectedState) {
			selectedState = state
		}
	}
	switch selectedState {
	case "popup_storm_review_required":
		return SharedSessionBrowserCoordinationActions{
			PrimaryAction:      "browser action=close",
			RecommendedActions: []string{"browser action=close", "browser action=tabs", "browser action=pin_target"},
		}
	case "redirect_review_required":
		return SharedSessionBrowserCoordinationActions{
			PrimaryAction:      "browser action=pin_target",
			RecommendedActions: []string{"browser action=pin_target", "browser action=tabs"},
		}
	case "popup_review_required":
		return SharedSessionBrowserCoordinationActions{
			PrimaryAction:      "browser action=tabs",
			RecommendedActions: []string{"browser action=tabs", "browser action=focus", "browser action=close", "browser action=pin_target"},
		}
	default:
		return SharedSessionBrowserCoordinationActions{}
	}
}

func EvaluateSharedSessionBrowserManagedRouteRecoveryActions(input SharedSessionBrowserManagedRouteRecoveryInput) SharedSessionBrowserCoordinationActions {
	if len(input.Profiles) == 0 || strings.TrimSpace(input.ActiveNodeRunID) != "" || sharedSessionBrowserHasRunningProfile(input.Profiles) {
		return SharedSessionBrowserCoordinationActions{}
	}
	blockedManagedRoute := false
	for _, route := range input.Routes {
		if sharedSessionBrowserFollowPolicyPriority(route.FollowPolicyState) == 0 {
			continue
		}
		if route.ManagedRuntime {
			blockedManagedRoute = true
			continue
		}
		return SharedSessionBrowserCoordinationActions{}
	}
	if !blockedManagedRoute {
		return SharedSessionBrowserCoordinationActions{}
	}
	return SharedSessionBrowserCoordinationActions{
		PrimaryAction:      "browser action=reset",
		RecommendedActions: []string{"browser action=reset", "browser action=ensure"},
	}
}

func ApplySharedSessionBrowserCoordinationActions(guidance SharedSessionBrowserCoordinationGuidance, actions SharedSessionBrowserCoordinationActions) SharedSessionBrowserCoordinationGuidance {
	if len(actions.RecommendedActions) > 0 {
		merged := make([]string, 0, len(actions.RecommendedActions)+len(guidance.RecommendedActions))
		merged = append(merged, actions.RecommendedActions...)
		merged = append(merged, guidance.RecommendedActions...)
		guidance.RecommendedActions = sharedSessionBrowserUniqueActions(merged)
	}
	if primary := strings.TrimSpace(actions.PrimaryAction); primary != "" {
		guidance.PrimaryAction = primary
		guidance.NextStep = primary
	}
	return guidance
}

func EvaluateSharedSessionBrowserCoordinationEvaluation(input SharedSessionBrowserCoordinationEvaluationInput) SharedSessionBrowserCoordinationEvaluation {
	plan := EvaluateSharedSessionBrowserCoordination(input.Coordination)
	evaluation := SharedSessionBrowserCoordinationEvaluation{
		Plan:          plan,
		RestartAction: strings.TrimSpace(plan.RestartAction),
		Guidance: SharedSessionBrowserCoordinationGuidance{
			PrimaryAction:      strings.TrimSpace(plan.PrimaryBrowserAction),
			NextStep:           strings.TrimSpace(plan.NextStep),
			RecommendedActions: sharedSessionBrowserUniqueActions(plan.RecommendedBrowserActions),
		},
	}
	baseBrowserActions := sharedSessionBrowserUniqueActions(plan.RecommendedBrowserActions)
	if actions := EvaluateSharedSessionBrowserFollowPolicyActions(input.Routes); strings.TrimSpace(actions.PrimaryAction) != "" || len(actions.RecommendedActions) > 0 {
		evaluation.Guidance = ApplySharedSessionBrowserCoordinationActions(evaluation.Guidance, actions)
	}
	if actions := EvaluateSharedSessionBrowserManagedRouteRecoveryActions(SharedSessionBrowserManagedRouteRecoveryInput{
		ActiveNodeRunID: strings.TrimSpace(input.Coordination.ActiveNodeRunID),
		Profiles:        input.Coordination.Profiles,
		Routes:          input.Routes,
	}); strings.TrimSpace(actions.PrimaryAction) != "" || len(actions.RecommendedActions) > 0 {
		evaluation.Guidance = ApplySharedSessionBrowserCoordinationActions(SharedSessionBrowserCoordinationGuidance{
			PrimaryAction:      strings.TrimSpace(evaluation.Guidance.PrimaryAction),
			NextStep:           strings.TrimSpace(evaluation.Guidance.NextStep),
			RecommendedActions: baseBrowserActions,
		}, actions)
	}
	guidance := ApplySharedSessionBrowserHealthActions(SharedSessionBrowserHealthGuidance{
		RestartAction:      evaluation.RestartAction,
		PrimaryAction:      strings.TrimSpace(evaluation.Guidance.PrimaryAction),
		PrimaryNodeAction:  strings.TrimSpace(plan.PrimaryNodeAction),
		NextStep:           strings.TrimSpace(evaluation.Guidance.NextStep),
		RecommendedActions: evaluation.Guidance.RecommendedActions,
	}, RecommendSharedSessionBrowserHealthActions(input.BlockedAutoFollow, input.HealthEvaluation))
	evaluation.RestartAction = strings.TrimSpace(guidance.RestartAction)
	evaluation.Guidance.PrimaryAction = strings.TrimSpace(guidance.PrimaryAction)
	evaluation.Guidance.NextStep = strings.TrimSpace(guidance.NextStep)
	evaluation.Guidance.RecommendedActions = guidance.RecommendedActions
	return evaluation
}

func EvaluateSharedSessionBrowserCoordinationEvaluationForScope(registry SharedSessionBrowserStateRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, healthInput SharedSessionBrowserHealthInput, coordination SharedSessionBrowserCoordinationInput, routes []SharedSessionBrowserRouteCoordinationInput, reconnectWindow time.Duration, blockedAutoFollow bool) SharedSessionBrowserCoordinationEvaluation {
	if snapshot := sharedSessionBrowserProfilesForScope(registry, sessionID, selectedInfo, requestedProfile); len(snapshot) > 0 {
		healthInput.StoredState = ""
		healthInput.StoredReason = ""
		healthInput.StoredRecoveryAction = ""
		healthInput.Profiles = append([]SharedSessionBrowserProfileState(nil), snapshot...)
		coordination.Profiles = append([]SharedSessionBrowserProfileState(nil), snapshot...)
	}
	return EvaluateSharedSessionBrowserCoordinationEvaluation(SharedSessionBrowserCoordinationEvaluationInput{
		Coordination:      coordination,
		Routes:            routes,
		HealthEvaluation:  EvaluateSharedSessionBrowserHealth(healthInput, reconnectWindow),
		BlockedAutoFollow: blockedAutoFollow,
	})
}

func EvaluateSharedSessionBrowserCoordinationStatus(coordinationState string, goal string, profile *BrowserProfileStatusResult, syncReady bool, prepareDecision string) SharedSessionBrowserCoordinationStatus {
	status := SharedSessionBrowserCoordinationStatus{
		Decision: DecideSharedSessionBrowserCoordination(coordinationState, prepareDecision, goal),
		Ready:    SharedSessionBrowserCoordinationReady(coordinationState, goal, profile, syncReady),
	}
	if strings.EqualFold(strings.TrimSpace(goal), "restart") && status.Ready {
		switch strings.TrimSpace(status.Decision) {
		case "", "restarted", "restart_started", "already_ready", "browser_ready":
			status.Decision = "restart_ready"
		}
	}
	return status
}

func DecideSharedSessionBrowserCoordination(coordinationState string, prepareDecision string, goal string) string {
	prepareDecision = strings.TrimSpace(prepareDecision)
	coordinationState = strings.TrimSpace(coordinationState)
	switch goal {
	case "sync":
		switch prepareDecision {
		case "session_route_synced", "session_route_already_synced", "session_profile_synced", "session_profile_already_synced", "session_target_synced", "session_target_already_synced", "session_route_sync_unavailable", "session_profile_validation_failed", "session_target_sync_failed":
			return prepareDecision
		}
		if coordinationState == "coordinated" || coordinationState == "browser_ready" {
			return "sync_ready"
		}
		return prepareDecision
	case "restart":
		switch prepareDecision {
		case "restart_blocked_active_node_run", "restart_status_failed", "restart_stop_failed", "restart_start_failed", "restart_reconnect_in_progress", "restarted", "restart_started":
			return prepareDecision
		}
		if coordinationState == "coordinated" || coordinationState == "browser_ready" {
			return "restart_ready"
		}
		return prepareDecision
	case "teardown":
		switch prepareDecision {
		case "teardown_blocked_active_node_run", "teardown_stopped", "teardown_already_stopped", "teardown_no_profile", "teardown_status_failed", "teardown_stop_failed":
			return prepareDecision
		}
	}
	switch coordinationState {
	case "coordinated":
		if prepareDecision == "started" {
			return "started_for_active_node_run"
		}
		return "already_coordinated"
	case "browser_ready":
		if prepareDecision == "started" {
			return "started_browser_profile"
		}
		return "browser_ready"
	default:
		if prepareDecision != "" {
			return prepareDecision
		}
		return coordinationState
	}
}

func SharedSessionBrowserCoordinationReady(coordinationState string, goal string, profile *BrowserProfileStatusResult, syncReady bool) bool {
	coordinationState = strings.TrimSpace(coordinationState)
	switch goal {
	case "sync":
		return syncReady
	case "restart":
		if profile != nil {
			status := strings.ToLower(strings.TrimSpace(profile.Status))
			if status == "reconnecting" {
				return false
			}
			return SharedSessionBrowserProfileReady(*profile)
		}
		return coordinationState == "coordinated" || coordinationState == "browser_ready"
	case "teardown":
		if profile == nil {
			return coordinationState == "idle"
		}
		return !SharedSessionBrowserProfileReady(*profile)
	default:
		return coordinationState == "coordinated" || coordinationState == "browser_ready"
	}
}

func ApplySharedSessionBrowserHealthActions(guidance SharedSessionBrowserHealthGuidance, actions SharedSessionBrowserHealthActions) SharedSessionBrowserHealthGuidance {
	if len(actions.RecommendedActions) > 0 {
		merged := make([]string, 0, len(actions.RecommendedActions)+len(guidance.RecommendedActions))
		merged = append(merged, actions.RecommendedActions...)
		merged = append(merged, guidance.RecommendedActions...)
		guidance.RecommendedActions = sharedSessionBrowserUniqueActions(merged)
	}
	if strings.TrimSpace(actions.PrimaryAction) != "" {
		guidance.PrimaryAction = strings.TrimSpace(actions.PrimaryAction)
		guidance.NextStep = guidance.PrimaryAction
	}
	if actions.ClearRestartAction {
		guidance.RestartAction = ""
	}
	if restartAction := strings.TrimSpace(actions.RestartAction); restartAction != "" {
		guidance.RestartAction = restartAction
	}
	if actions.SuppressRefresh {
		guidance.RestartAction = ""
		guidance.RecommendedActions = sharedSessionBrowserWithoutAction(guidance.RecommendedActions, "browser action=refresh")
		if strings.EqualFold(strings.TrimSpace(guidance.PrimaryAction), "browser action=refresh") {
			guidance.PrimaryAction = ""
		}
		if strings.EqualFold(strings.TrimSpace(guidance.NextStep), "browser action=refresh") {
			guidance.NextStep = firstNonEmptyString(strings.TrimSpace(guidance.PrimaryAction), strings.TrimSpace(guidance.PrimaryNodeAction))
		}
	}
	return guidance
}
