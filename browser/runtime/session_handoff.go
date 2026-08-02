package browserruntime

import "strings"

const (
	SharedSessionBrowserSessionHandoffStateReady                   = "ready"
	SharedSessionBrowserSessionHandoffStateHealthAttention         = "health_attention"
	SharedSessionBrowserSessionHandoffStateTargetReviewRequired    = "target_review_required"
	SharedSessionBrowserSessionHandoffStateTargetSelectionRequired = "target_selection_required"
	SharedSessionBrowserSessionHandoffStateSessionContextOnly      = "session_context_only"
	SharedSessionBrowserSessionHandoffStateEmpty                   = "empty"
)

// SharedSessionBrowserSessionHandoffTargetSummary is the compact current or
// pending target identity needed to resume a long browser session without
// replaying the full route target inventory.
type SharedSessionBrowserSessionHandoffTargetSummary struct {
	ID            string `json:"id,omitempty"`
	TabIndex      int    `json:"tab_index,omitempty"`
	URL           string `json:"url,omitempty"`
	Title         string `json:"title,omitempty"`
	BrowserApp    string `json:"browser_app,omitempty"`
	Backend       string `json:"backend,omitempty"`
	Profile       string `json:"profile,omitempty"`
	RuntimeTarget string `json:"runtime_target,omitempty"`
	Source        string `json:"source,omitempty"`
}

// SharedSessionBrowserSessionHandoffSummary is a compact, derived-only resume
// summary for long-running browser sessions. It is intentionally built from
// existing route, run, profile, and health projections rather than becoming a
// new session owner.
type SharedSessionBrowserSessionHandoffSummary struct {
	State                    string                                           `json:"state,omitempty"`
	NextStepAlias            string                                           `json:"next_step_alias,omitempty"`
	TargetCount              int                                              `json:"target_count,omitempty"`
	RouteCount               int                                              `json:"route_count,omitempty"`
	PendingTargetReviewCount int                                              `json:"pending_target_review_count,omitempty"`
	CurrentTarget            *SharedSessionBrowserSessionHandoffTargetSummary `json:"current_target,omitempty"`
	PendingTargetReview      *SharedSessionBrowserSessionHandoffTargetSummary `json:"pending_target_review,omitempty"`
	ActiveRunID              string                                           `json:"active_run_id,omitempty"`
	RunCount                 int                                              `json:"run_count,omitempty"`
	ActiveRunCount           int                                              `json:"active_run_count,omitempty"`
	SelectedProfile          string                                           `json:"selected_profile,omitempty"`
	SelectedProfileSource    string                                           `json:"selected_profile_source,omitempty"`
	SelectedProfileStatus    string                                           `json:"selected_profile_status,omitempty"`
	ActiveProfile            string                                           `json:"active_profile,omitempty"`
	BrowserApp               string                                           `json:"browser_app,omitempty"`
	FollowPolicyState        string                                           `json:"follow_policy_state,omitempty"`
	PopupPolicyState         string                                           `json:"popup_policy_state,omitempty"`
	HealthState              string                                           `json:"health_state,omitempty"`
	HealthReason             string                                           `json:"health_reason,omitempty"`
	RecoveryAction           string                                           `json:"recovery_action,omitempty"`
	ReconnectHint            string                                           `json:"reconnect_hint,omitempty"`
}

// SharedSessionBrowserSessionHandoffRequest contains the existing session
// projections used to derive a compact handoff summary.
type SharedSessionBrowserSessionHandoffRequest struct {
	Routes          []SharedSessionBrowserRouteSnapshot
	Runs            []SharedSessionRunInfo
	Profiles        []SharedSessionBrowserProjectedProfileState
	TargetCount     int
	SelectedProfile *SharedSessionBrowserProfileSelection
	SelectedTarget  *BrowserSessionTargetSelection
	Health          *SharedSessionBrowserHealthSummary
}

// BuildSharedSessionBrowserSessionHandoffSummary builds a compact handoff from
// the existing browser-session projections.
func BuildSharedSessionBrowserSessionHandoffSummary(
	req SharedSessionBrowserSessionHandoffRequest,
) *SharedSessionBrowserSessionHandoffSummary {
	targetCount := req.TargetCount
	if targetCount <= 0 {
		targetCount = SharedSessionBrowserRouteTargetCount(req.Routes)
	}
	pendingReviewCount, pendingReview := sharedSessionBrowserHandoffPendingTargetReview(req.Routes)
	currentTarget := sharedSessionBrowserHandoffCurrentTarget(req.Routes, req.SelectedTarget)
	activeRunID, activeRunCount := sharedSessionBrowserHandoffRuns(req.Runs)
	selectedProfile, selectedProfileSource, selectedProfileStatus, browserApp := sharedSessionBrowserHandoffSelectedProfile(req.Profiles, req.SelectedProfile)
	activeProfile := sharedSessionBrowserHandoffActiveProfile(req.Profiles)
	followPolicyState, popupPolicyState := sharedSessionBrowserHandoffPolicyStates(req.Routes)

	summary := &SharedSessionBrowserSessionHandoffSummary{
		TargetCount:              targetCount,
		RouteCount:               len(req.Routes),
		PendingTargetReviewCount: pendingReviewCount,
		CurrentTarget:            currentTarget,
		PendingTargetReview:      pendingReview,
		ActiveRunID:              activeRunID,
		RunCount:                 len(req.Runs),
		ActiveRunCount:           activeRunCount,
		SelectedProfile:          selectedProfile,
		SelectedProfileSource:    selectedProfileSource,
		SelectedProfileStatus:    selectedProfileStatus,
		ActiveProfile:            activeProfile,
		BrowserApp:               firstNonEmptyBindingString(browserApp, sharedSessionBrowserHandoffTargetBrowserApp(currentTarget), sharedSessionBrowserHandoffTargetBrowserApp(pendingReview)),
		FollowPolicyState:        followPolicyState,
		PopupPolicyState:         popupPolicyState,
	}
	if req.Health != nil {
		summary.HealthState = strings.TrimSpace(req.Health.State)
		summary.HealthReason = strings.TrimSpace(req.Health.Reason)
		summary.RecoveryAction = strings.TrimSpace(req.Health.RecoveryAction)
		summary.ReconnectHint = strings.TrimSpace(req.Health.ReconnectHint)
	}
	if sharedSessionBrowserSessionHandoffEmpty(summary) {
		return nil
	}
	summary.State, summary.NextStepAlias = sharedSessionBrowserSessionHandoffState(summary, req.Health)
	return summary
}

// CloneSharedSessionBrowserSessionHandoffSummary returns a deep-enough clone of
// a handoff summary for projection bridges.
func CloneSharedSessionBrowserSessionHandoffSummary(
	summary *SharedSessionBrowserSessionHandoffSummary,
) *SharedSessionBrowserSessionHandoffSummary {
	if summary == nil {
		return nil
	}
	cloned := *summary
	if summary.CurrentTarget != nil {
		target := *summary.CurrentTarget
		cloned.CurrentTarget = &target
	}
	if summary.PendingTargetReview != nil {
		target := *summary.PendingTargetReview
		cloned.PendingTargetReview = &target
	}
	return &cloned
}

func sharedSessionBrowserBindingEvaluationHandoff(
	evaluation SharedSessionBrowserBindingEvaluation,
	routes []SharedSessionBrowserRouteSnapshot,
	runs []SharedSessionRunInfo,
	profiles []SharedSessionBrowserProjectedProfileState,
	targetCount int,
) *SharedSessionBrowserSessionHandoffSummary {
	if targetCount <= 0 {
		targetCount = evaluation.Snapshot.Summary.RouteTargetCount
	}
	return BuildSharedSessionBrowserSessionHandoffSummary(SharedSessionBrowserSessionHandoffRequest{
		Routes:          routes,
		Runs:            runs,
		Profiles:        profiles,
		TargetCount:     targetCount,
		SelectedProfile: evaluation.Snapshot.SelectedProfileSelection,
		SelectedTarget:  evaluation.Snapshot.SelectedTargetSelection,
		Health:          evaluation.Health.Summary,
	})
}

func sharedSessionBrowserFinalizeBindingEvaluationHandoff(
	evaluation SharedSessionBrowserBindingEvaluation,
) SharedSessionBrowserBindingEvaluation {
	profiles := ProjectSharedSessionBrowserProfileSnapshot(
		evaluation.Snapshot.Profiles,
		evaluation.Snapshot.SelectedProfileSelection,
	)
	evaluation.Handoff = sharedSessionBrowserBindingEvaluationHandoff(
		evaluation,
		evaluation.Routes,
		evaluation.Snapshot.Runs,
		profiles,
		evaluation.Snapshot.Summary.RouteTargetCount,
	)
	return evaluation
}

func sharedSessionBrowserHandoffPendingTargetReview(
	routes []SharedSessionBrowserRouteSnapshot,
) (int, *SharedSessionBrowserSessionHandoffTargetSummary) {
	count := 0
	var pending *SharedSessionBrowserSessionHandoffTargetSummary
	for _, route := range routes {
		if route.PendingTargetReviewCount > 0 {
			count += route.PendingTargetReviewCount
		} else if route.PendingTargetReview != nil {
			count++
		}
		if pending != nil || route.PendingTargetReview == nil {
			continue
		}
		review := route.PendingTargetReview
		pending = sharedSessionBrowserNormalizeHandoffTarget(SharedSessionBrowserSessionHandoffTargetSummary{
			ID:            strings.TrimSpace(review.ID),
			TabIndex:      review.TabIndex,
			URL:           strings.TrimSpace(review.URL),
			Title:         strings.TrimSpace(review.Title),
			BrowserApp:    firstNonEmptyBindingString(review.BrowserApp, route.BrowserApp),
			Backend:       firstNonEmptyBindingString(review.Backend, route.Backend),
			Profile:       firstNonEmptyBindingString(review.Profile, route.Profile),
			RuntimeTarget: firstNonEmptyBindingString(review.Target, route.RuntimeTarget),
			Source:        "pending_target_review",
		})
	}
	return count, pending
}

func sharedSessionBrowserHandoffCurrentTarget(
	routes []SharedSessionBrowserRouteSnapshot,
	selected *BrowserSessionTargetSelection,
) *SharedSessionBrowserSessionHandoffTargetSummary {
	if selected != nil {
		if target := sharedSessionBrowserNormalizeHandoffTarget(SharedSessionBrowserSessionHandoffTargetSummary{
			ID:            strings.TrimSpace(selected.ID),
			TabIndex:      selected.TabIndex,
			URL:           strings.TrimSpace(selected.URL),
			Title:         strings.TrimSpace(selected.Title),
			BrowserApp:    strings.TrimSpace(selected.BrowserApp),
			Backend:       strings.TrimSpace(selected.Backend),
			Profile:       strings.TrimSpace(selected.Profile),
			RuntimeTarget: strings.TrimSpace(selected.RuntimeTarget),
			Source:        strings.TrimSpace(selected.Source),
		}); target != nil {
			return target
		}
	}
	for _, route := range routes {
		currentID := strings.TrimSpace(route.CurrentTargetID)
		for _, target := range route.Targets {
			if !target.Current && (currentID == "" || !strings.EqualFold(strings.TrimSpace(target.ID), currentID)) {
				continue
			}
			if out := sharedSessionBrowserHandoffTargetFromRouteTarget(route, target, route.CurrentTargetSource); out != nil {
				return out
			}
		}
		if currentID != "" {
			if out := sharedSessionBrowserNormalizeHandoffTarget(SharedSessionBrowserSessionHandoffTargetSummary{
				ID:            currentID,
				BrowserApp:    strings.TrimSpace(route.BrowserApp),
				Backend:       strings.TrimSpace(route.Backend),
				Profile:       strings.TrimSpace(route.Profile),
				RuntimeTarget: strings.TrimSpace(route.RuntimeTarget),
				Source:        strings.TrimSpace(route.CurrentTargetSource),
			}); out != nil {
				return out
			}
		}
	}
	return nil
}

func sharedSessionBrowserHandoffTargetFromRouteTarget(
	route SharedSessionBrowserRouteSnapshot,
	target SharedSessionBrowserRouteTarget,
	source string,
) *SharedSessionBrowserSessionHandoffTargetSummary {
	return sharedSessionBrowserNormalizeHandoffTarget(SharedSessionBrowserSessionHandoffTargetSummary{
		ID:            strings.TrimSpace(target.ID),
		TabIndex:      target.TabIndex,
		URL:           strings.TrimSpace(target.URL),
		Title:         strings.TrimSpace(target.Title),
		BrowserApp:    firstNonEmptyBindingString(target.BrowserApp, route.BrowserApp),
		Backend:       firstNonEmptyBindingString(target.Backend, route.Backend),
		Profile:       firstNonEmptyBindingString(target.Profile, route.Profile),
		RuntimeTarget: firstNonEmptyBindingString(target.RuntimeTarget, route.RuntimeTarget),
		Source:        strings.TrimSpace(source),
	})
}

func sharedSessionBrowserNormalizeHandoffTarget(
	target SharedSessionBrowserSessionHandoffTargetSummary,
) *SharedSessionBrowserSessionHandoffTargetSummary {
	target.ID = strings.TrimSpace(target.ID)
	target.URL = strings.TrimSpace(target.URL)
	target.Title = strings.TrimSpace(target.Title)
	target.BrowserApp = strings.TrimSpace(target.BrowserApp)
	target.Backend = strings.TrimSpace(target.Backend)
	target.Profile = strings.TrimSpace(target.Profile)
	target.RuntimeTarget = strings.TrimSpace(target.RuntimeTarget)
	target.Source = strings.TrimSpace(target.Source)
	if target.ID == "" &&
		target.TabIndex <= 0 &&
		target.URL == "" &&
		target.Title == "" &&
		target.BrowserApp == "" &&
		target.Backend == "" &&
		target.Profile == "" &&
		target.RuntimeTarget == "" &&
		target.Source == "" {
		return nil
	}
	return &target
}

func sharedSessionBrowserHandoffRuns(runs []SharedSessionRunInfo) (string, int) {
	activeRunID := ""
	activeRunCount := 0
	for _, run := range runs {
		status := strings.ToLower(strings.TrimSpace(run.Status))
		active := status == "running" || status == "pending" || status == "starting"
		if active {
			activeRunCount++
		}
		if activeRunID == "" && active {
			activeRunID = strings.TrimSpace(run.RunID)
		}
	}
	if activeRunID == "" && len(runs) > 0 {
		activeRunID = strings.TrimSpace(runs[0].RunID)
	}
	return activeRunID, activeRunCount
}

func sharedSessionBrowserHandoffSelectedProfile(
	profiles []SharedSessionBrowserProjectedProfileState,
	selection *SharedSessionBrowserProfileSelection,
) (string, string, string, string) {
	profile := ""
	source := ""
	if selection != nil {
		profile = strings.TrimSpace(selection.Profile)
		source = strings.TrimSpace(selection.Source)
	}
	var fallback *SharedSessionBrowserProfileState
	for i := range profiles {
		state := profiles[i].State
		if profiles[i].Selected {
			if profile == "" {
				profile = strings.TrimSpace(state.Profile)
			}
			return profile, source, strings.TrimSpace(state.Status), strings.TrimSpace(firstNonEmptyBindingString(state.BrowserApp, firstStringValue(selection, func(v *SharedSessionBrowserProfileSelection) string { return v.BrowserApp })))
		}
		if fallback == nil && (state.Running || state.Connected || strings.EqualFold(strings.TrimSpace(state.Status), "running")) {
			candidate := state
			fallback = &candidate
		}
	}
	if fallback == nil && len(profiles) > 0 {
		candidate := profiles[0].State
		fallback = &candidate
	}
	if fallback != nil {
		if profile == "" {
			profile = strings.TrimSpace(fallback.Profile)
		}
		return profile, source, strings.TrimSpace(fallback.Status), strings.TrimSpace(firstNonEmptyBindingString(fallback.BrowserApp, firstStringValue(selection, func(v *SharedSessionBrowserProfileSelection) string { return v.BrowserApp })))
	}
	return profile, source, "", strings.TrimSpace(firstStringValue(selection, func(v *SharedSessionBrowserProfileSelection) string { return v.BrowserApp }))
}

func sharedSessionBrowserHandoffActiveProfile(profiles []SharedSessionBrowserProjectedProfileState) string {
	for _, item := range profiles {
		state := item.State
		status := strings.ToLower(strings.TrimSpace(state.Status))
		if state.Running || state.Connected || status == "running" || status == "started" {
			return strings.TrimSpace(state.Profile)
		}
	}
	if len(profiles) > 0 {
		return strings.TrimSpace(profiles[0].State.Profile)
	}
	return ""
}

func sharedSessionBrowserHandoffPolicyStates(routes []SharedSessionBrowserRouteSnapshot) (string, string) {
	followPolicy := ""
	popupPolicy := ""
	for _, route := range routes {
		if followPolicy == "" && strings.TrimSpace(route.FollowPolicyState) != "" && !strings.EqualFold(strings.TrimSpace(route.FollowPolicyState), "auto_follow_allowed") {
			followPolicy = strings.TrimSpace(route.FollowPolicyState)
		}
		if popupPolicy == "" && strings.TrimSpace(route.PopupPolicyState) != "" {
			popupPolicy = strings.TrimSpace(route.PopupPolicyState)
		}
	}
	return followPolicy, popupPolicy
}

func sharedSessionBrowserHandoffTargetBrowserApp(target *SharedSessionBrowserSessionHandoffTargetSummary) string {
	if target == nil {
		return ""
	}
	return strings.TrimSpace(target.BrowserApp)
}

func sharedSessionBrowserSessionHandoffState(
	summary *SharedSessionBrowserSessionHandoffSummary,
	health *SharedSessionBrowserHealthSummary,
) (string, string) {
	if summary.PendingTargetReviewCount > 0 {
		return SharedSessionBrowserSessionHandoffStateTargetReviewRequired, "review_target"
	}
	healthState := strings.TrimSpace(summary.HealthState)
	if healthState != "" && !strings.EqualFold(healthState, "healthy") {
		alias := ""
		if health != nil {
			alias = strings.TrimSpace(health.NextStepAlias)
		}
		if alias == "" {
			alias = "recover_browser"
		}
		return SharedSessionBrowserSessionHandoffStateHealthAttention, alias
	}
	if summary.CurrentTarget != nil {
		return SharedSessionBrowserSessionHandoffStateReady, "continue_current_target"
	}
	if summary.TargetCount > 0 {
		return SharedSessionBrowserSessionHandoffStateTargetSelectionRequired, "select_target"
	}
	if summary.ActiveRunID != "" || summary.SelectedProfile != "" || summary.ActiveProfile != "" {
		return SharedSessionBrowserSessionHandoffStateSessionContextOnly, "snapshot"
	}
	return SharedSessionBrowserSessionHandoffStateEmpty, "prepare_browser"
}

func sharedSessionBrowserSessionHandoffEmpty(summary *SharedSessionBrowserSessionHandoffSummary) bool {
	return summary.TargetCount == 0 &&
		summary.RouteCount == 0 &&
		summary.PendingTargetReviewCount == 0 &&
		summary.CurrentTarget == nil &&
		summary.PendingTargetReview == nil &&
		summary.ActiveRunID == "" &&
		summary.RunCount == 0 &&
		summary.ActiveRunCount == 0 &&
		summary.SelectedProfile == "" &&
		summary.SelectedProfileSource == "" &&
		summary.SelectedProfileStatus == "" &&
		summary.ActiveProfile == "" &&
		summary.BrowserApp == "" &&
		summary.FollowPolicyState == "" &&
		summary.PopupPolicyState == "" &&
		summary.HealthState == "" &&
		summary.HealthReason == "" &&
		summary.RecoveryAction == "" &&
		summary.ReconnectHint == ""
}
