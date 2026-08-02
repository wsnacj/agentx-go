package browserruntime

import (
	"strings"
	"time"
)

// SharedSessionBrowserCoordinationSummary captures the tool-facing
// coordination state/decision/ready posture derived from a finalized action.
type SharedSessionBrowserCoordinationSummary struct {
	State    string
	Decision string
	Ready    bool
}

// SharedSessionBrowserCoordinationSummaryInput carries the shared inputs used
// to derive the finalized coordination summary for refresh/coordinate actions.
type SharedSessionBrowserCoordinationSummaryInput struct {
	Action            string
	CoordinationGoal  string
	Evaluation        *SharedSessionBrowserBindingEvaluation
	CoordinationState string
	HealthState       string
	ProfileStatus     *BrowserProfileStatusResult
	PrepareDecision   string
	RestartDecision   string
	SyncSessionReady  bool
}

// SharedSessionBrowserFinalizedActionSurface captures the shared finalized
// binding/session surface that tools can map into payload-specific views.
type SharedSessionBrowserFinalizedActionSurface struct {
	BindingProjection       *SharedSessionBrowserTopLevelBindingProjection
	WorkbenchSurface        *SharedSessionBrowserWorkbenchSessionSurfaceProjection
	SessionProjection       *SharedSessionBrowserWorkbenchSessionProjectionApplication
	UseWorkbenchSurface     bool
	SyncCoordinationSurface bool
	CoordinationSummary     *SharedSessionBrowserCoordinationSummary
}

// SharedSessionBrowserFinalizedActionSurfaceRequest carries the shared inputs
// needed to build the finalized action surface after tool execution completes.
type SharedSessionBrowserFinalizedActionSurfaceRequest struct {
	Action              string
	CoordinationGoal    string
	SessionID           string
	Route               BrowserSessionRoute
	Routes              []SharedSessionBrowserRouteSnapshot
	Registry            *BrowserSessionRegistry
	RunRegistry         SharedSessionRunRegistry
	StateRegistry       SharedSessionBrowserStateRegistry
	BindingEvaluation   *SharedSessionBrowserBindingEvaluation
	ProfileStatus       *BrowserProfileStatusResult
	CurrentProfile      *SharedSessionBrowserProfileState
	HealthSummary       *SharedSessionBrowserHealthSummary
	ReconnectWindow     time.Duration
	CoordinationSummary SharedSessionBrowserCoordinationSummaryInput
}

// BuildSharedSessionBrowserCoordinationSummary derives the finalized
// coordination summary for refresh/coordinate actions, including typed
// health-state overrides when a binding evaluation or fallback health state is
// available.
func BuildSharedSessionBrowserCoordinationSummary(
	input SharedSessionBrowserCoordinationSummaryInput,
) (SharedSessionBrowserCoordinationSummary, bool) {
	action := strings.ToLower(strings.TrimSpace(input.Action))
	if action != "coordinate" && action != "refresh" {
		return SharedSessionBrowserCoordinationSummary{}, false
	}
	state := strings.TrimSpace(input.CoordinationState)
	healthState := strings.TrimSpace(input.HealthState)
	if input.Evaluation != nil {
		if candidate := strings.TrimSpace(input.Evaluation.Coordination.Plan.State); candidate != "" {
			state = candidate
		}
		if input.Evaluation.Health.Summary != nil {
			if candidate := strings.TrimSpace(input.Evaluation.Health.Summary.State); candidate != "" {
				healthState = candidate
			}
		}
	}
	decision, goal := sharedSessionBrowserCoordinationSummaryInput(
		action,
		input.CoordinationGoal,
		input.PrepareDecision,
		input.RestartDecision,
	)
	summary := SharedSessionBrowserCoordinationSummary{State: state}
	if override, ok := sharedSessionBrowserCoordinationSummaryHealthOverride(healthState); ok {
		summary.Decision = strings.TrimSpace(override.Decision)
		summary.Ready = override.Ready
		return summary, true
	}
	status := EvaluateSharedSessionBrowserCoordinationStatus(
		state,
		goal,
		input.ProfileStatus,
		input.SyncSessionReady,
		decision,
	)
	summary.Decision = strings.TrimSpace(status.Decision)
	summary.Ready = status.Ready
	return summary, true
}

// BuildSharedSessionBrowserFinalizedActionSurface projects the shared binding
// and coordination surface consumed by tools during final payload refresh.
func BuildSharedSessionBrowserFinalizedActionSurface(
	req SharedSessionBrowserFinalizedActionSurfaceRequest,
) SharedSessionBrowserFinalizedActionSurface {
	action := strings.ToLower(strings.TrimSpace(req.Action))
	route := normalizeBrowserSessionRoute(req.Route)
	currentProfile := sharedSessionBrowserFinalizedActionCurrentProfile(req, route)
	projection := SharedSessionBrowserTopLevelBindingProjection{}
	if !sharedSessionBrowserRouteEmpty(route) {
		projection = ProjectSharedSessionBrowserTopLevelBinding(
			strings.TrimSpace(req.SessionID),
			BrowserRuntimeInfo{
				Backend: strings.TrimSpace(route.Backend),
				Profile: strings.TrimSpace(route.Profile),
				Target:  strings.TrimSpace(route.Target),
			},
			route,
			req.Routes,
			req.Registry,
			req.RunRegistry,
			req.StateRegistry,
			req.BindingEvaluation,
			currentProfile,
			req.ReconnectWindow,
		)
	} else if req.BindingEvaluation != nil {
		projection = ProjectSharedSessionBrowserTopLevelBindingFromEvaluation(
			strings.TrimSpace(req.SessionID),
			req.Routes,
			req.Registry,
			req.RunRegistry,
			req.StateRegistry,
			*req.BindingEvaluation,
			currentProfile,
			req.ReconnectWindow,
		)
	}
	if req.HealthSummary != nil && (!sharedSessionBrowserRouteEmpty(route) || req.BindingEvaluation != nil || currentProfile != nil) {
		selectedInfo := sharedSessionBrowserTopLevelBindingSelectedInfo(projection.Evaluation, currentProfile)
		requestedProfile := firstNonEmptyBindingString(
			strings.TrimSpace(selectedInfo.Profile),
			strings.TrimSpace(route.Profile),
		)
		projection.Evaluation = MergeSharedSessionBrowserBindingEvaluationHealthSummary(
			req.StateRegistry,
			strings.TrimSpace(req.SessionID),
			selectedInfo,
			requestedProfile,
			projection.Evaluation,
			req.HealthSummary,
			req.ReconnectWindow,
		)
	}
	surface := SharedSessionBrowserFinalizedActionSurface{
		BindingProjection: &projection,
	}
	if action == "workbench" {
		sessionProjection := ProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluation(projection.Evaluation)
		if len(sessionProjection.Routes) > 0 || sessionProjection.TargetCount > 0 || len(sessionProjection.Runs) > 0 || len(sessionProjection.Profiles) > 0 {
			surface.SessionProjection = sharedSessionBrowserWorkbenchSessionProjectionApplication(
				&sessionProjection,
				BrowserRuntimeInfo{},
				false,
				false,
			)
		}
		surface.WorkbenchSurface = sharedSessionBrowserFinalizedWorkbenchSurface(
			projection,
			surface.SessionProjection,
			req.ProfileStatus == nil || sharedSessionBrowserProfileStatusResultEmpty(*req.ProfileStatus),
			currentProfile,
		)
		surface.UseWorkbenchSurface = true
		surface.SyncCoordinationSurface = true
	}
	coordinationInput := req.CoordinationSummary
	if strings.TrimSpace(coordinationInput.Action) == "" {
		coordinationInput.Action = action
	}
	if strings.TrimSpace(coordinationInput.CoordinationGoal) == "" {
		coordinationInput.CoordinationGoal = strings.TrimSpace(req.CoordinationGoal)
	}
	if coordinationInput.ProfileStatus == nil {
		coordinationInput.ProfileStatus = req.ProfileStatus
	}
	if coordinationInput.Evaluation == nil {
		coordinationInput.Evaluation = &projection.Evaluation
	}
	if summary, ok := BuildSharedSessionBrowserCoordinationSummary(coordinationInput); ok {
		surface.CoordinationSummary = &summary
	}
	return surface
}

func sharedSessionBrowserFinalizedWorkbenchSurface(
	projection SharedSessionBrowserTopLevelBindingProjection,
	sessionProjection *SharedSessionBrowserWorkbenchSessionProjectionApplication,
	needProfileStatus bool,
	currentProfile *SharedSessionBrowserProfileState,
) *SharedSessionBrowserWorkbenchSessionSurfaceProjection {
	selectedInfo := sharedSessionBrowserTopLevelBindingSelectedInfo(projection.Evaluation, currentProfile)
	var sessionView *SharedSessionBrowserTopLevelSessionProjection
	if sessionProjection != nil {
		sessionView = sessionProjection.Projection
	}
	surface := BuildSharedSessionBrowserWorkbenchSessionSurfaceProjection(
		SharedSessionBrowserWorkbenchSessionSurfaceRequest{
			SelectedInfo:         selectedInfo,
			NeedProfileStatus:    needProfileStatus,
			NeedProfileInventory: true,
			BindingProjection:    &projection,
			SessionProjection:    sessionView,
		},
	)
	if surface.BindingProjection == nil &&
		surface.SessionProjection == nil &&
		surface.FallbackSessionProjection == nil &&
		surface.ProfileInventory == nil {
		return nil
	}
	return &surface
}

func sharedSessionBrowserFinalizedActionCurrentProfile(
	req SharedSessionBrowserFinalizedActionSurfaceRequest,
	route BrowserSessionRoute,
) *SharedSessionBrowserProfileState {
	if req.CurrentProfile != nil {
		current := *req.CurrentProfile
		return &current
	}
	if req.BindingEvaluation != nil {
		return nil
	}
	if req.ProfileStatus == nil || sharedSessionBrowserProfileStatusResultEmpty(*req.ProfileStatus) {
		return nil
	}
	selectedInfo := BrowserRuntimeInfo{
		Backend: firstNonEmptyBindingString(
			strings.TrimSpace(req.ProfileStatus.Backend),
			strings.TrimSpace(route.Backend),
		),
		Profile: firstNonEmptyBindingString(
			strings.TrimSpace(req.ProfileStatus.Profile),
			strings.TrimSpace(route.Profile),
		),
		Target: strings.TrimSpace(route.Target),
	}
	current := SharedSessionBrowserProfileStateFromStatus(selectedInfo, *req.ProfileStatus)
	return &current
}

func sharedSessionBrowserCoordinationSummaryHealthOverride(
	healthState string,
) (SharedSessionBrowserCoordinationStatus, bool) {
	switch strings.TrimSpace(healthState) {
	case "cooldown_active", "restart_pending", "browser_disconnect_burst", "restart_failed", "restart_failed_permanent":
		return SharedSessionBrowserCoordinationStatus{
			Decision: strings.TrimSpace(healthState),
			Ready:    false,
		}, true
	default:
		return SharedSessionBrowserCoordinationStatus{}, false
	}
}

func sharedSessionBrowserCoordinationSummaryInput(
	action string,
	coordinationGoal string,
	prepareDecision string,
	restartDecision string,
) (string, string) {
	action = strings.ToLower(strings.TrimSpace(action))
	switch action {
	case "refresh":
		return strings.TrimSpace(restartDecision), "restart"
	case "coordinate":
		return strings.TrimSpace(prepareDecision), strings.TrimSpace(coordinationGoal)
	default:
		return "", strings.TrimSpace(coordinationGoal)
	}
}
