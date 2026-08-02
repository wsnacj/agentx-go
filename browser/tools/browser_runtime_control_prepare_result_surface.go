package tools

import (
	"context"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func browserRuntimeApplyPrepareResultSurface(
	callCtx context.Context,
	payload *browserRuntimePayload,
	selectedInfo BrowserRuntimeInfo,
	projection agentxbrowserruntime.SharedSessionBrowserExecutionSurfaceProjection,
) {
	if payload == nil {
		return
	}
	browserRuntimeApplySharedActionSurface(
		callCtx,
		payload,
		agentxbrowserruntime.SharedSessionBrowserActionSurfaceProjection{
			Note:           projection.Surface.Note,
			ConfiguredInfo: selectedInfo,
			ProfileInventory: agentxbrowserruntime.ProjectSharedSessionBrowserTopLevelProfileInventoryFromExecutionInventory(
				selectedInfo,
				projection.InventoryProjection,
			),
		},
	)
	if projection.Surface.ClearedSessionTargets > 0 {
		payload.ClearedSessionTargets = projection.Surface.ClearedSessionTargets
	}
}

func browserRuntimeTopLevelProfileInventoryProjectionFromShared(
	configuredInfo BrowserRuntimeInfo,
	surface *agentxbrowserruntime.SharedSessionBrowserTopLevelProfileInventoryProjection,
) *browserRuntimeTopLevelProfileInventoryProjection {
	projection := browserRuntimeTopLevelProfileInventoryProjection{
		ConfiguredInfo: configuredInfo,
	}
	if surface == nil {
		return nil
	}
	hasProjection := false
	if surface.HasProfileStatus {
		projection.ProfileStatus = browserRuntimeProfileStatePtrFromSharedSessionState(surface.ProfileStatus)
		projection.ApplyProfileStatus = true
		hasProjection = true
	}
	if surface.ApplyProfileInventory {
		projection.Profiles = browserRuntimeProfileStatesFromProjected(surface.Profiles)
		projection.DiscoveredProfiles = append([]string(nil), surface.DiscoveredProfiles...)
		projection.DefaultProfile = surface.DefaultProfile
		projection.ApplyProfileInventory = true
		hasProjection = true
	}
	if !hasProjection {
		return nil
	}
	return &projection
}

func browserRuntimeSharedTopLevelSessionProjectionPtr(
	projection *browserRuntimeTopLevelSessionProjection,
) *agentxbrowserruntime.SharedSessionBrowserTopLevelSessionProjection {
	if projection == nil {
		return nil
	}
	shared := agentxbrowserruntime.SharedSessionBrowserTopLevelSessionProjection{
		TargetCount: projection.TargetCount,
		Handoff:     agentxbrowserruntime.CloneSharedSessionBrowserSessionHandoffSummary(projection.Handoff),
	}
	if len(projection.Routes) > 0 {
		shared.Routes = browserRuntimeSharedSessionRouteSnapshots(projection.Routes)
	}
	if len(projection.Runs) > 0 {
		shared.Runs = browserRuntimeSharedSessionRunsFromBinding(projection.Runs)
	}
	if len(projection.Profiles) > 0 {
		shared.Profiles = browserRuntimeSharedProjectedProfiles(projection.Profiles)
	}
	return &shared
}

func browserRuntimeSharedSessionProjectionPtr(
	ok bool,
	projection agentxbrowserruntime.SharedSessionBrowserTopLevelSessionProjection,
) *agentxbrowserruntime.SharedSessionBrowserTopLevelSessionProjection {
	if !ok {
		return nil
	}
	shared := projection
	return &shared
}

func browserRuntimeSharedProjectedProfiles(
	items []browserRuntimeProfileState,
) []agentxbrowserruntime.SharedSessionBrowserProjectedProfileState {
	if len(items) == 0 {
		return nil
	}
	out := make([]agentxbrowserruntime.SharedSessionBrowserProjectedProfileState, 0, len(items))
	for _, item := range items {
		out = append(out, agentxbrowserruntime.SharedSessionBrowserProjectedProfileState{
			State: agentxbrowserruntime.SharedSessionBrowserProfileState{
				Backend:       item.Backend,
				Profile:       item.Profile,
				RuntimeTarget: item.RuntimeTarget,
				BrowserApp:    item.BrowserApp,
				Status:        item.Status,
				Connected:     item.Connected,
				Running:       item.Running,
				Note:          item.Note,
			},
			Selected: item.Selected,
		})
	}
	return out
}
