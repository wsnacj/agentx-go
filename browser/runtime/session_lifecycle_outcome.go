package browserruntime

import "strings"

// SharedSessionBrowserLifecycleActionOutcome captures the shared tool-facing
// lifecycle result contract after execution apply/surface projection finishes.
type SharedSessionBrowserLifecycleActionOutcome struct {
	Action                           string
	CoordinationGoal                 string
	Result                           SharedSessionBrowserExecutionResult
	PreparedProfile                  string
	PrepareDecision                  string
	PrepareReady                     bool
	RestartDecision                  string
	RestartReady                     bool
	StopDecision                     string
	StopReady                        bool
	CreateDecision                   string
	CreateReady                      bool
	DeleteDecision                   string
	DeleteReady                      bool
	ExecutionProjection              SharedSessionBrowserExecutionSurfaceProjection
	RememberOutcome                  *SharedSessionBrowserActionOutcome
	Status                           string
	Note                             string
	Err                              error
	ApplyCoordinationDecisionOnError bool
}

// SharedSessionBrowserLifecycleActionOutcomeRequest carries the shared inputs
// needed to lower a lifecycle dispatch result into the stable action outcome
// contract consumed by tools.
type SharedSessionBrowserLifecycleActionOutcomeRequest struct {
	Action                           string
	CoordinationGoal                 string
	MutationContext                  SharedSessionBrowserMutationContext
	SessionID                        string
	SelectedInfo                     BrowserRuntimeInfo
	Route                            BrowserSessionRoute
	RequestedProfile                 string
	RequestedBrowserApp              string
	DispatchResult                   SharedSessionBrowserLifecycleActionDispatchResult
	RememberProfile                  bool
	ApplyCoordinationDecisionOnError bool
}

// BuildSharedSessionBrowserLifecycleActionOutcome lowers a lifecycle dispatch
// result into the shared outcome contract, including execution-application
// surface projection through the mutation seam.
func BuildSharedSessionBrowserLifecycleActionOutcome(
	req SharedSessionBrowserLifecycleActionOutcomeRequest,
) SharedSessionBrowserLifecycleActionOutcome {
	result := req.DispatchResult.Result
	application := ApplySharedSessionBrowserExecutionResultWithMutationContext(
		req.MutationContext,
		strings.TrimSpace(req.SessionID),
		req.SelectedInfo,
		"",
		result,
	)
	surface := ProjectSharedSessionBrowserExecutionSurface(req.SelectedInfo, result, application)
	outcome := SharedSessionBrowserLifecycleActionOutcome{
		Action:                           strings.ToLower(strings.TrimSpace(req.Action)),
		CoordinationGoal:                 strings.TrimSpace(req.CoordinationGoal),
		Result:                           result,
		ExecutionProjection:              BuildSharedSessionBrowserExecutionSurfaceProjectionFromSurface(surface),
		RememberOutcome:                  buildSharedSessionBrowserLifecycleRememberActionOutcome(req, result, surface),
		Status:                           "",
		Note:                             "",
		Err:                              req.DispatchResult.Err,
		ApplyCoordinationDecisionOnError: req.ApplyCoordinationDecisionOnError,
	}
	if outcome.Err != nil {
		outcome.Status = "error"
		outcome.Note = outcome.Err.Error()
	}
	projectSharedSessionBrowserLifecycleDecisionFields(&outcome, result)
	return outcome
}

func projectSharedSessionBrowserLifecycleDecisionFields(
	outcome *SharedSessionBrowserLifecycleActionOutcome,
	result SharedSessionBrowserExecutionResult,
) {
	if outcome == nil {
		return
	}
	outcome.PreparedProfile = strings.TrimSpace(result.Profile)
	decision := strings.TrimSpace(result.Decision)
	switch strings.ToLower(strings.TrimSpace(outcome.Action)) {
	case "prepare":
		outcome.PrepareDecision = decision
		outcome.PrepareReady = result.Ready
	case "coordinate":
		switch strings.ToLower(strings.TrimSpace(outcome.CoordinationGoal)) {
		case "restart":
			outcome.RestartDecision = decision
			outcome.RestartReady = result.Ready
		default:
			outcome.PrepareDecision = decision
			outcome.PrepareReady = result.Ready
		}
	case "restart", "refresh":
		outcome.RestartDecision = decision
		outcome.RestartReady = result.Ready
	case "stop":
		outcome.StopDecision = decision
		outcome.StopReady = result.Ready
	case "create_profile":
		outcome.CreateDecision = decision
		outcome.CreateReady = result.Ready
	case "delete_profile":
		outcome.DeleteDecision = decision
		outcome.DeleteReady = result.Ready
	}
}

func buildSharedSessionBrowserLifecycleRememberActionOutcome(
	req SharedSessionBrowserLifecycleActionOutcomeRequest,
	result SharedSessionBrowserExecutionResult,
	surface SharedSessionBrowserExecutionSurface,
) *SharedSessionBrowserActionOutcome {
	if !req.RememberProfile || req.DispatchResult.Err != nil {
		return nil
	}
	remember := DispatchSharedSessionBrowserRememberProfile(
		SharedSessionBrowserRememberProfileDispatchRequest{
			MutationContext:     req.MutationContext,
			SessionID:           strings.TrimSpace(req.SessionID),
			SelectedInfo:        req.SelectedInfo,
			Route:               req.Route,
			ProfileStatus:       sharedSessionBrowserLifecycleRememberProfileStatus(req.SelectedInfo, result.Profile, surface),
			PreparedProfile:     strings.TrimSpace(result.Profile),
			RequestedProfile:    strings.TrimSpace(req.RequestedProfile),
			RequestedBrowserApp: strings.TrimSpace(req.RequestedBrowserApp),
		},
	)
	outcome := BuildSharedSessionBrowserRememberActionOutcome(remember)
	return &outcome
}

func sharedSessionBrowserLifecycleRememberProfileStatus(
	selectedInfo BrowserRuntimeInfo,
	preparedProfile string,
	surface SharedSessionBrowserExecutionSurface,
) *BrowserProfileStatusResult {
	switch {
	case surface.HasProfileState:
		status := SharedSessionBrowserProfileStatusResultFromState(
			surface.ProfileState,
			selectedInfo,
			preparedProfile,
		)
		return &status
	case surface.HasProfileStatus:
		status := surface.ProfileStatus
		return &status
	default:
		return nil
	}
}
