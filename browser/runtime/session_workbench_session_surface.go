package browserruntime

// SharedSessionBrowserWorkbenchSessionProjectionApplication captures a shared
// session projection plus the configured-info/apply flags that tools bridge
// into workbench payload session shells.
type SharedSessionBrowserWorkbenchSessionProjectionApplication struct {
	Projection              *SharedSessionBrowserTopLevelSessionProjection
	ConfiguredInfo          BrowserRuntimeInfo
	ApplyConfiguredProfiles bool
	MissingSessionIDNote    string
}

// SharedSessionBrowserWorkbenchSessionSurfaceProjection captures the shared
// binding/session/profile-inventory pieces that tools bridge into workbench
// payload shells.
type SharedSessionBrowserWorkbenchSessionSurfaceProjection struct {
	BindingProjection         *SharedSessionBrowserTopLevelBindingProjection
	SessionProjection         *SharedSessionBrowserWorkbenchSessionProjectionApplication
	FallbackSessionProjection *SharedSessionBrowserWorkbenchSessionProjectionApplication
	ProfileInventory          *SharedSessionBrowserTopLevelProfileInventoryProjection
}

// SharedSessionBrowserWorkbenchSessionSurfaceRequest carries the shared inputs
// needed to build the lifecycle-owned workbench session surface contract.
type SharedSessionBrowserWorkbenchSessionSurfaceRequest struct {
	SelectedInfo              BrowserRuntimeInfo
	RequestedDefaultProfile   string
	NeedProfileStatus         bool
	NeedProfileInventory      bool
	ConfiguredInfo            BrowserRuntimeInfo
	ApplyConfiguredProfiles   bool
	ApplyMissingSessionIDNote bool
	BindingProjection         *SharedSessionBrowserTopLevelBindingProjection
	BindingEvaluation         *SharedSessionBrowserBindingEvaluation
	ApplyBindingEvaluation    bool
	SessionProjection         *SharedSessionBrowserTopLevelSessionProjection
}

// BuildSharedSessionBrowserWorkbenchSessionSurfaceProjection lowers explicit
// binding/session projections plus optional direct binding evaluation into the
// shared workbench-session surface contract so tools callers do not need to
// maintain a parallel direct-binding/session/profile-inventory control flow.
func BuildSharedSessionBrowserWorkbenchSessionSurfaceProjection(
	req SharedSessionBrowserWorkbenchSessionSurfaceRequest,
) SharedSessionBrowserWorkbenchSessionSurfaceProjection {
	configuredInfo := sharedSessionBrowserNormalizedRuntimeInfo(req.ConfiguredInfo)
	if configuredInfo == (BrowserRuntimeInfo{}) {
		configuredInfo = sharedSessionBrowserNormalizedRuntimeInfo(req.SelectedInfo)
	}
	projection := SharedSessionBrowserWorkbenchSessionSurfaceProjection{
		BindingProjection: sharedSessionBrowserWorkbenchBindingProjection(req),
	}
	evaluation := sharedSessionBrowserWorkbenchBindingEvaluation(req, projection.BindingProjection)
	primarySessionProjection := cloneSharedSessionBrowserTopLevelSessionProjection(req.SessionProjection)
	if primarySessionProjection == nil && evaluation != nil {
		sessionProjection := ProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluation(*evaluation)
		if !sharedSessionBrowserTopLevelSessionProjectionEmpty(sessionProjection) {
			primarySessionProjection = &sessionProjection
		}
	}
	projection.SessionProjection = sharedSessionBrowserWorkbenchSessionProjectionApplication(
		primarySessionProjection,
		configuredInfo,
		req.ApplyConfiguredProfiles,
		req.ApplyMissingSessionIDNote,
	)
	if evaluation != nil {
		bindingSessionProjection := ProjectSharedSessionBrowserTopLevelSessionFromBindingEvaluation(*evaluation)
		if !sharedSessionBrowserTopLevelSessionProjectionEmpty(bindingSessionProjection) &&
			projection.SessionProjection != nil &&
			sharedSessionBrowserWorkbenchNeedsFallbackSessionProjection(
				projection.SessionProjection.Projection,
				bindingSessionProjection,
				evaluation.Snapshot.Summary.ActiveNodeRunID,
			) {
			projection.FallbackSessionProjection = sharedSessionBrowserWorkbenchSessionProjectionApplication(
				&bindingSessionProjection,
				configuredInfo,
				req.ApplyConfiguredProfiles,
				req.ApplyMissingSessionIDNote,
			)
		}
	}
	if projection.SessionProjection != nil {
		projection.ProfileInventory = ProjectSharedSessionBrowserTopLevelProfileInventoryFromSessionProjection(
			SharedSessionBrowserSessionProjectionProfileInventoryRequest{
				SelectedInfo:            req.SelectedInfo,
				RequestedDefaultProfile: req.RequestedDefaultProfile,
				NeedProfileStatus:       req.NeedProfileStatus,
				NeedProfileInventory:    req.NeedProfileInventory,
				SessionProjection:       projection.SessionProjection.Projection,
			},
		)
	}
	if projection.ProfileInventory == nil && projection.FallbackSessionProjection != nil {
		projection.ProfileInventory = ProjectSharedSessionBrowserTopLevelProfileInventoryFromSessionProjection(
			SharedSessionBrowserSessionProjectionProfileInventoryRequest{
				SelectedInfo:            req.SelectedInfo,
				RequestedDefaultProfile: req.RequestedDefaultProfile,
				NeedProfileStatus:       req.NeedProfileStatus,
				NeedProfileInventory:    req.NeedProfileInventory,
				SessionProjection:       projection.FallbackSessionProjection.Projection,
			},
		)
	}
	return projection
}

func sharedSessionBrowserWorkbenchBindingEvaluation(
	req SharedSessionBrowserWorkbenchSessionSurfaceRequest,
	projection *SharedSessionBrowserTopLevelBindingProjection,
) *SharedSessionBrowserBindingEvaluation {
	if projection != nil {
		return &projection.Evaluation
	}
	if req.BindingEvaluation == nil {
		return nil
	}
	return req.BindingEvaluation
}

func sharedSessionBrowserWorkbenchSessionProjectionApplication(
	projection *SharedSessionBrowserTopLevelSessionProjection,
	configuredInfo BrowserRuntimeInfo,
	applyConfiguredProfiles bool,
	applyMissingSessionIDNote bool,
) *SharedSessionBrowserWorkbenchSessionProjectionApplication {
	projection = cloneSharedSessionBrowserTopLevelSessionProjection(projection)
	if projection == nil {
		return nil
	}
	return &SharedSessionBrowserWorkbenchSessionProjectionApplication{
		Projection:              projection,
		ConfiguredInfo:          configuredInfo,
		ApplyConfiguredProfiles: applyConfiguredProfiles,
		MissingSessionIDNote:    sharedSessionBrowserMissingSessionIDNote(applyMissingSessionIDNote),
	}
}

func sharedSessionBrowserMissingSessionIDNote(enabled bool) string {
	if !enabled {
		return ""
	}
	return "browser_runtime: no tool session context is available"
}

func sharedSessionBrowserWorkbenchNeedsFallbackSessionProjection(
	current *SharedSessionBrowserTopLevelSessionProjection,
	binding SharedSessionBrowserTopLevelSessionProjection,
	activeNodeRunID string,
) bool {
	if current == nil {
		return true
	}
	return len(current.Routes) == 0 ||
		current.TargetCount == 0 ||
		(len(current.Runs) == 0 && activeNodeRunID != "" && len(binding.Runs) > 0) ||
		len(current.Profiles) == 0
}

func sharedSessionBrowserWorkbenchBindingProjection(
	req SharedSessionBrowserWorkbenchSessionSurfaceRequest,
) *SharedSessionBrowserTopLevelBindingProjection {
	if req.BindingProjection != nil {
		cloned := *req.BindingProjection
		return &cloned
	}
	if req.BindingEvaluation == nil || !req.ApplyBindingEvaluation {
		return nil
	}
	projection := SharedSessionBrowserTopLevelBindingProjection{
		Evaluation: *req.BindingEvaluation,
	}
	if selection := req.BindingEvaluation.Snapshot.SelectedProfileSelection; selection != nil {
		cloned := *selection
		projection.ProfileSelection = &cloned
	}
	if selection := req.BindingEvaluation.Snapshot.SelectedTargetSelection; selection != nil {
		cloned := *selection
		projection.TargetSelection = &cloned
	}
	if sharedSessionBrowserTopLevelBindingProjectionEmpty(projection) {
		return nil
	}
	return &projection
}

func cloneSharedSessionBrowserTopLevelSessionProjection(
	projection *SharedSessionBrowserTopLevelSessionProjection,
) *SharedSessionBrowserTopLevelSessionProjection {
	if projection == nil {
		return nil
	}
	cloned := *projection
	if len(projection.Routes) > 0 {
		cloned.Routes = append([]SharedSessionBrowserRouteSnapshot(nil), projection.Routes...)
	}
	if len(projection.Runs) > 0 {
		cloned.Runs = append([]SharedSessionRunInfo(nil), projection.Runs...)
	}
	if len(projection.Profiles) > 0 {
		cloned.Profiles = cloneSharedSessionBrowserProjectedProfiles(projection.Profiles)
	}
	cloned.Handoff = CloneSharedSessionBrowserSessionHandoffSummary(projection.Handoff)
	if cloned.Handoff == nil {
		cloned.Handoff = BuildSharedSessionBrowserSessionHandoffSummary(SharedSessionBrowserSessionHandoffRequest{
			Routes:      cloned.Routes,
			Runs:        cloned.Runs,
			Profiles:    cloned.Profiles,
			TargetCount: cloned.TargetCount,
		})
	}
	return &cloned
}

func sharedSessionBrowserTopLevelSessionProjectionEmpty(
	projection SharedSessionBrowserTopLevelSessionProjection,
) bool {
	return len(projection.Routes) == 0 &&
		projection.TargetCount == 0 &&
		len(projection.Runs) == 0 &&
		len(projection.Profiles) == 0 &&
		projection.Handoff == nil
}

func sharedSessionBrowserTopLevelBindingProjectionEmpty(
	projection SharedSessionBrowserTopLevelBindingProjection,
) bool {
	return projection.ProfileSelection == nil &&
		projection.TargetSelection == nil &&
		len(projection.Evaluation.Routes) == 0 &&
		len(projection.Evaluation.Snapshot.Runs) == 0 &&
		len(projection.Evaluation.Snapshot.Profiles) == 0 &&
		projection.Evaluation.Snapshot.SelectedProfileSelection == nil &&
		projection.Evaluation.Snapshot.SelectedTargetSelection == nil &&
		projection.Evaluation.Snapshot.CurrentTargetID == ""
}
