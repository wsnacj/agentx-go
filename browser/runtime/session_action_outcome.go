package browserruntime

import (
	"fmt"
	"strings"
)

// SharedSessionBrowserSelectionProjection captures the shared profile/target
// selection writeback that tools can map into their payload-specific surface.
type SharedSessionBrowserSelectionProjection struct {
	ProfileSelection   *SharedSessionBrowserProfileSelection
	TargetSelection    *BrowserSessionTargetSelection
	ApplyTargetToRoute bool
}

// SharedSessionBrowserActionOutcome captures the shared tool-facing session
// action result contract after browserruntime dispatch/mutation completes.
type SharedSessionBrowserActionOutcome struct {
	ApplyDecision                    bool
	Action                           string
	CoordinationGoal                 string
	Decision                         string
	Ready                            bool
	SelectDecision                   string
	SelectReady                      bool
	SelectTargetDecision             string
	SelectTargetReady                bool
	SyncSessionDecision              string
	SyncSessionReady                 bool
	RememberDecision                 string
	RememberReady                    bool
	ClearDecision                    string
	ClearReady                       bool
	ClearSessionDecision             string
	ClearSessionReady                bool
	ClearTargetDecision              string
	ClearTargetReady                 bool
	Status                           string
	Note                             string
	Err                              error
	ApplyCoordinationDecisionOnError bool
	SelectionProjection              *SharedSessionBrowserSelectionProjection
	ProfileInventoryProjection       *SharedSessionBrowserTopLevelProfileInventoryProjection
	ClearProfileStatus               bool
	ClearedSessionProfiles           int
	ClearedSessionTargets            int
}

// SharedSessionBrowserProjectedDecision is the runtime-owned action decision
// projection that tools bridge into their payload-specific JSON fields.
type SharedSessionBrowserProjectedDecision struct {
	Action                 string
	Decision               string
	Ready                  bool
	ClearProfileStatus     bool
	ClearedSessionProfiles int
	ClearedSessionTargets  int
}

// SharedSessionBrowserSelectionActionOutcomeRequest carries the runtime-owned
// inputs needed to lower a selection dispatch result into the shared action
// outcome contract consumed by tools.
type SharedSessionBrowserSelectionActionOutcomeRequest struct {
	Action             string
	DispatchResult     SharedSessionBrowserSelectionActionDispatchResult
	MissingNote        string
	ApplyTargetToRoute bool
}

// SharedSessionBrowserBasicActionOutcomeRequest carries action-local decision,
// status, note, and error fields that browserruntime should lower into the
// shared action-outcome contract before tools bridge into payload fields.
type SharedSessionBrowserBasicActionOutcomeRequest struct {
	ApplyDecision                    bool
	Action                           string
	CoordinationGoal                 string
	Decision                         string
	Ready                            bool
	Status                           string
	Note                             string
	Err                              error
	ApplyCoordinationDecisionOnError bool
	ClearProfileStatus               bool
	ClearedSessionProfiles           int
	ClearedSessionTargets            int
}

// SharedSessionBrowserActionDecisionSurface returns the action-specific
// decision surface that should receive an outcome's decision/ready pair.
func SharedSessionBrowserActionDecisionSurface(action string, coordinationGoal string) string {
	action = strings.ToLower(strings.TrimSpace(action))
	if action == "coordinate" && strings.EqualFold(strings.TrimSpace(coordinationGoal), "sync") {
		return "sync_session"
	}
	return action
}

// ProjectSharedSessionBrowserActionDecision lowers the outcome into one
// decision/ready projection so tools do not need to re-own action aliases.
func ProjectSharedSessionBrowserActionDecision(outcome SharedSessionBrowserActionOutcome) (SharedSessionBrowserProjectedDecision, bool) {
	if !outcome.ApplyDecision {
		return SharedSessionBrowserProjectedDecision{}, false
	}
	projection := SharedSessionBrowserProjectedDecision{
		Action:                 SharedSessionBrowserActionDecisionSurface(outcome.Action, outcome.CoordinationGoal),
		ClearProfileStatus:     outcome.ClearProfileStatus,
		ClearedSessionProfiles: outcome.ClearedSessionProfiles,
		ClearedSessionTargets:  outcome.ClearedSessionTargets,
	}
	switch projection.Action {
	case "select_profile":
		projection.Decision = strings.TrimSpace(outcome.SelectDecision)
		projection.Ready = outcome.SelectReady
	case "select_target":
		projection.Decision = strings.TrimSpace(outcome.SelectTargetDecision)
		projection.Ready = outcome.SelectTargetReady
	case "sync_session":
		projection.Decision = strings.TrimSpace(outcome.SyncSessionDecision)
		projection.Ready = outcome.SyncSessionReady
	case "remember_profile":
		projection.Decision = strings.TrimSpace(outcome.RememberDecision)
		projection.Ready = outcome.RememberReady
	case "clear_profile":
		projection.Decision = strings.TrimSpace(outcome.ClearDecision)
		projection.Ready = outcome.ClearReady
	case "clear_session":
		projection.Decision = strings.TrimSpace(outcome.ClearSessionDecision)
		projection.Ready = outcome.ClearSessionReady
	case "clear_target":
		projection.Decision = strings.TrimSpace(outcome.ClearTargetDecision)
		projection.Ready = outcome.ClearTargetReady
	default:
		return SharedSessionBrowserProjectedDecision{}, false
	}
	return projection, true
}

// SharedSessionBrowserDecisionRequiresReview reports whether the decision
// leaves the caller in a manual review posture.
func SharedSessionBrowserDecisionRequiresReview(decision string) bool {
	return strings.HasSuffix(strings.TrimSpace(decision), "_review_required")
}

func projectSharedSessionBrowserActionDecisionFields(outcome *SharedSessionBrowserActionOutcome) {
	if outcome == nil {
		return
	}
	decision := strings.TrimSpace(outcome.Decision)
	ready := outcome.Ready
	action := SharedSessionBrowserActionDecisionSurface(outcome.Action, outcome.CoordinationGoal)
	switch action {
	case "select_profile":
		outcome.SelectDecision = decision
		outcome.SelectReady = ready
	case "select_target":
		outcome.SelectTargetDecision = decision
		outcome.SelectTargetReady = ready
	case "sync_session":
		outcome.SyncSessionDecision = decision
		outcome.SyncSessionReady = ready
	case "remember_profile":
		outcome.RememberDecision = decision
		outcome.RememberReady = ready
	case "clear_profile":
		outcome.ClearDecision = decision
		outcome.ClearReady = ready
	case "clear_session":
		outcome.ClearSessionDecision = decision
		outcome.ClearSessionReady = ready
	case "clear_target":
		outcome.ClearTargetDecision = decision
		outcome.ClearTargetReady = ready
	}
}

func sharedSessionBrowserMissingInputNote(action string) string {
	action = strings.ToLower(strings.TrimSpace(action))
	switch action {
	case "select_profile":
		return "browser_runtime: profile is required for action select_profile"
	case "select_target":
		return "browser_runtime: target or tab_index is required for action select_target"
	default:
		return ""
	}
}

func sharedSessionBrowserApplyErrorTerminal(outcome *SharedSessionBrowserActionOutcome) {
	if outcome == nil || outcome.Err == nil {
		return
	}
	if strings.TrimSpace(outcome.Status) == "" {
		outcome.Status = "error"
	}
	if strings.TrimSpace(outcome.Note) == "" {
		outcome.Note = outcome.Err.Error()
	}
}

// BuildSharedSessionBrowserBasicActionOutcome lowers a minimal action-local
// result into the shared action outcome contract so tools callers do not need
// to mirror action-specific decision-field projection logic.
func BuildSharedSessionBrowserBasicActionOutcome(
	req SharedSessionBrowserBasicActionOutcomeRequest,
) SharedSessionBrowserActionOutcome {
	outcome := SharedSessionBrowserActionOutcome{
		ApplyDecision:                    req.ApplyDecision,
		Action:                           strings.ToLower(strings.TrimSpace(req.Action)),
		CoordinationGoal:                 strings.TrimSpace(req.CoordinationGoal),
		Decision:                         strings.TrimSpace(req.Decision),
		Ready:                            req.Ready,
		Status:                           strings.TrimSpace(req.Status),
		Note:                             strings.TrimSpace(req.Note),
		Err:                              req.Err,
		ApplyCoordinationDecisionOnError: req.ApplyCoordinationDecisionOnError,
		ClearProfileStatus:               req.ClearProfileStatus,
		ClearedSessionProfiles:           req.ClearedSessionProfiles,
		ClearedSessionTargets:            req.ClearedSessionTargets,
	}
	sharedSessionBrowserApplyErrorTerminal(&outcome)
	projectSharedSessionBrowserActionDecisionFields(&outcome)
	return outcome
}

// BuildSharedSessionBrowserMissingInputActionOutcome lowers a missing-input
// action failure into the shared action outcome contract so tools callers do
// not keep owning canonical browser_runtime missing-input notes.
func BuildSharedSessionBrowserMissingInputActionOutcome(
	action string,
	decision string,
) SharedSessionBrowserActionOutcome {
	return BuildSharedSessionBrowserBasicActionOutcome(
		SharedSessionBrowserBasicActionOutcomeRequest{
			ApplyDecision: true,
			Action:        action,
			Decision:      decision,
			Status:        "error",
			Note:          sharedSessionBrowserMissingInputNote(action),
		},
	)
}

// BuildSharedSessionBrowserInvalidSelectTargetActionOutcome lowers an invalid
// select_target request into the shared action outcome contract so tools do not
// keep owning the canonical invalid-target decision.
func BuildSharedSessionBrowserInvalidSelectTargetActionOutcome(err error) SharedSessionBrowserActionOutcome {
	return BuildSharedSessionBrowserBasicActionOutcome(
		SharedSessionBrowserBasicActionOutcomeRequest{
			ApplyDecision: true,
			Action:        "select_target",
			Decision:      "session_target_invalid",
			Err:           err,
		},
	)
}

// BuildSharedSessionBrowserUnsupportedActionOutcome lowers an unsupported
// action into the shared terminal action-outcome contract so tools can stop
// duplicating the same unsupported status/note owner.
func BuildSharedSessionBrowserUnsupportedActionOutcome(action string) SharedSessionBrowserActionOutcome {
	return BuildSharedSessionBrowserUnsupportedRouteActionOutcome(action, nil)
}

// BuildSharedSessionBrowserUnsupportedRouteActionOutcome lowers an unsupported
// route/action terminal into the shared action-outcome contract so tools do not
// keep owning route-error unsupported notes.
func BuildSharedSessionBrowserUnsupportedRouteActionOutcome(action string, routeErr error) SharedSessionBrowserActionOutcome {
	action = strings.ToLower(strings.TrimSpace(action))
	note := ""
	if routeErr != nil {
		note = strings.TrimSpace(routeErr.Error())
	}
	if note == "" {
		note = fmt.Sprintf("browser_runtime: selected route does not support action %s", action)
	}
	return BuildSharedSessionBrowserBasicActionOutcome(
		SharedSessionBrowserBasicActionOutcomeRequest{
			Action: action,
			Status: "unsupported",
			Note:   note,
		},
	)
}

// BuildSharedSessionBrowserSelectionActionOutcome lowers a selection dispatch
// result into the shared session action outcome contract so tool callers only
// keep payload mapping.
func BuildSharedSessionBrowserSelectionActionOutcome(
	req SharedSessionBrowserSelectionActionOutcomeRequest,
) SharedSessionBrowserActionOutcome {
	action := strings.ToLower(strings.TrimSpace(req.Action))
	dispatched := req.DispatchResult
	missingNote := firstNonEmptyString(strings.TrimSpace(req.MissingNote), sharedSessionBrowserMissingInputNote(action))
	outcome := SharedSessionBrowserActionOutcome{
		ApplyDecision: true,
		Action:        action,
		Decision:      strings.TrimSpace(dispatched.Decision),
		Ready:         dispatched.Ready,
		Note:          strings.TrimSpace(dispatched.Note),
		Err:           dispatched.Err,
	}
	if dispatched.Err != nil {
		sharedSessionBrowserApplyErrorTerminal(&outcome)
		projectSharedSessionBrowserActionDecisionFields(&outcome)
		return outcome
	}
	switch action {
	case "select_profile":
		if dispatched.ProfileSelection == nil {
			outcome.Status = "error"
			outcome.Note = missingNote
			projectSharedSessionBrowserActionDecisionFields(&outcome)
			return outcome
		}
		outcome.SelectionProjection = &SharedSessionBrowserSelectionProjection{
			ProfileSelection: dispatched.ProfileSelection,
			TargetSelection:  dispatched.TargetSelection,
		}
	case "select_target":
		if dispatched.TargetSelection == nil {
			if SharedSessionBrowserDecisionRequiresReview(outcome.Decision) {
				outcome.Status = "review_required"
				projectSharedSessionBrowserActionDecisionFields(&outcome)
				return outcome
			}
			outcome.Status = "error"
			outcome.Note = missingNote
			projectSharedSessionBrowserActionDecisionFields(&outcome)
			return outcome
		}
		outcome.SelectionProjection = &SharedSessionBrowserSelectionProjection{
			ProfileSelection:   dispatched.ProfileSelection,
			TargetSelection:    dispatched.TargetSelection,
			ApplyTargetToRoute: req.ApplyTargetToRoute,
		}
	}
	projectSharedSessionBrowserActionDecisionFields(&outcome)
	return outcome
}

// BuildSharedSessionBrowserSyncActionOutcome lowers a sync/coordinated session
// dispatch result into the shared action outcome contract.
func BuildSharedSessionBrowserSyncActionOutcome(
	action string,
	coordinationGoal string,
	dispatched SharedSessionBrowserSyncActionDispatchResult,
	applyTargetToRoute bool,
	applyCoordinationDecisionOnError bool,
) SharedSessionBrowserActionOutcome {
	result := dispatched.Result
	var projection *SharedSessionBrowserSelectionProjection
	if result.ProfileSelection != nil || result.TargetSelection != nil {
		projection = &SharedSessionBrowserSelectionProjection{
			ProfileSelection:   result.ProfileSelection,
			TargetSelection:    result.TargetSelection,
			ApplyTargetToRoute: applyTargetToRoute,
		}
	}
	outcome := SharedSessionBrowserActionOutcome{
		ApplyDecision:                    true,
		Action:                           strings.ToLower(strings.TrimSpace(action)),
		CoordinationGoal:                 strings.TrimSpace(coordinationGoal),
		Decision:                         strings.TrimSpace(result.Decision),
		Ready:                            result.Ready,
		Err:                              dispatched.Err,
		ApplyCoordinationDecisionOnError: applyCoordinationDecisionOnError,
		SelectionProjection:              projection,
	}
	sharedSessionBrowserApplyErrorTerminal(&outcome)
	projectSharedSessionBrowserActionDecisionFields(&outcome)
	return outcome
}

// BuildSharedSessionBrowserClearActionOutcome lowers a clear-action result into
// the shared action outcome contract.
func BuildSharedSessionBrowserClearActionOutcome(
	action string,
	selectedInfo BrowserRuntimeInfo,
	result SharedSessionBrowserClearResult,
) SharedSessionBrowserActionOutcome {
	outcome := SharedSessionBrowserActionOutcome{
		ApplyDecision:              true,
		Action:                     strings.ToLower(strings.TrimSpace(action)),
		ProfileInventoryProjection: ProjectSharedSessionBrowserTopLevelProfileInventoryFromClearResult(selectedInfo, result),
		ClearProfileStatus:         true,
		Decision:                   strings.TrimSpace(result.Decision),
		Ready:                      result.Ready,
		ClearedSessionProfiles:     result.ClearedSessionProfiles,
		ClearedSessionTargets:      result.ClearedSessionTargets,
	}
	projectSharedSessionBrowserActionDecisionFields(&outcome)
	return outcome
}

// BuildSharedSessionBrowserRememberActionOutcome lowers a remember-profile
// dispatch result into the shared session action outcome contract.
func BuildSharedSessionBrowserRememberActionOutcome(
	result SharedSessionBrowserRememberProfileResult,
) SharedSessionBrowserActionOutcome {
	outcome := SharedSessionBrowserActionOutcome{
		ApplyDecision: true,
		Action:        "remember_profile",
		Decision:      strings.TrimSpace(result.Decision),
		Ready:         result.Ready,
		SelectionProjection: func() *SharedSessionBrowserSelectionProjection {
			if result.SelectionProjection == nil {
				return nil
			}
			cloned := *result.SelectionProjection
			return &cloned
		}(),
	}
	projectSharedSessionBrowserActionDecisionFields(&outcome)
	return outcome
}
