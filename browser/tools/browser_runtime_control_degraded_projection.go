package tools

import (
	"context"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserRuntimeDegradedRouteProjection struct {
	Route                    BrowserRuntimeInfo
	DefaultProfile           string
	ProfileStatus            *browserRuntimeProfileState
	Profiles                 []browserRuntimeProfileState
	SessionRoutes            []browserRuntimeSessionRoute
	SessionTargetCount       int
	SessionRuns              []browserRuntimeSessionRun
	SessionProfiles          []browserRuntimeProfileState
	WorkbenchSessionProfiles []browserRuntimeProfileState
	SessionHandoff           *browserRuntimeSessionHandoffSummary
	WorkbenchSessionHandoff  *browserRuntimeSessionHandoffSummary
	SessionProfileSelection  *browserRuntimeSessionProfileSelection
	SessionTargetSelection   *browserRuntimeSessionTargetSelection
	SessionBinding           *browserRuntimeSessionBinding
}

func browserRuntimeDegradedRouteProjectionFromSnapshot(
	ctx context.Context,
	sessionRegistry *BrowserSessionRegistry,
	runRegistry agentxbrowserruntime.SharedSessionRunRegistry,
	stateRegistry agentxbrowserruntime.SharedSessionBrowserStateRegistry,
	route BrowserRuntimeInfo,
	requestedProfile string,
) browserRuntimeDegradedRouteProjection {
	route = normalizeBrowserRuntimeInfo(route)
	requestedProfile = strings.TrimSpace(requestedProfile)

	bindingInfo := route
	if requestedProfile != "" {
		bindingInfo.Profile = requestedProfile
	}
	bindingRoute := browserRuntimeRouteDescriptorPtr(bindingInfo)

	sessionView := browserRuntimeDegradedSessionViewFromRouteSnapshot(
		ctx,
		sessionRegistry,
		runRegistry,
		stateRegistry,
		route,
		requestedProfile,
	)
	fallbackProjectedProfiles := browserRuntimeDegradedProjectedProfilesFromSessionView(
		route,
		requestedProfile,
		sessionView,
		nil,
	)
	sessionProjection := agentxbrowserruntime.ProjectSharedSessionBrowserTopLevelSessionView(sessionView, fallbackProjectedProfiles)
	sessionRoutes := browserRuntimeSessionRoutesFromShared(sessionProjection.Routes)
	bindingProjection := browserRuntimeTopLevelBindingProjection(
		ctx,
		sessionRegistry,
		runRegistry,
		stateRegistry,
		bindingRoute,
		sessionRoutes,
		nil,
		nil,
	)
	if len(fallbackProjectedProfiles) == 0 {
		fallbackProjectedProfiles = browserRuntimeDegradedProjectedProfilesFromSessionView(
			route,
			requestedProfile,
			sessionView,
			browserRuntimeSelectionPtrValue(bindingProjection.ProfileSelection),
		)
	}
	workbenchSessionProjection := agentxbrowserruntime.ProjectSharedSessionBrowserTopLevelSessionView(sessionView, fallbackProjectedProfiles)

	profileStatus := browserRuntimeDegradedProfileStatusFromRouteSnapshot(ctx, stateRegistry, route, requestedProfile)
	profiles := browserRuntimeDegradedProfilesFromRouteSnapshot(ctx, stateRegistry, route, requestedProfile)
	if profileStatus == nil {
		profileStatus = browserRuntimeProfileStatusFromStates(
			browserRuntimeProfileStatesFromProjected(fallbackProjectedProfiles),
			strings.TrimSpace(firstNonEmpty(requestedProfile, route.Profile)),
		)
	}
	if len(profiles) == 0 {
		profiles = browserRuntimeProfileStatesFromProjected(fallbackProjectedProfiles)
	}

	sessionRuns := browserRuntimeSessionRunsFromShared(sessionProjection.Runs)
	sessionProfiles := browserRuntimeProfileStatesFromProjected(sessionProjection.Profiles)
	workbenchSessionProfiles := browserRuntimeProfileStatesFromProjected(workbenchSessionProjection.Profiles)

	return browserRuntimeDegradedRouteProjection{
		Route:                    route,
		DefaultProfile:           strings.TrimSpace(route.Profile),
		ProfileStatus:            profileStatus,
		Profiles:                 profiles,
		SessionRoutes:            sessionRoutes,
		SessionTargetCount:       sessionProjection.TargetCount,
		SessionRuns:              sessionRuns,
		SessionProfiles:          sessionProfiles,
		WorkbenchSessionProfiles: workbenchSessionProfiles,
		SessionHandoff:           agentxbrowserruntime.CloneSharedSessionBrowserSessionHandoffSummary(sessionProjection.Handoff),
		WorkbenchSessionHandoff:  agentxbrowserruntime.CloneSharedSessionBrowserSessionHandoffSummary(workbenchSessionProjection.Handoff),
		SessionProfileSelection:  browserRuntimeSelectionPtrValue(bindingProjection.ProfileSelection),
		SessionTargetSelection:   browserRuntimeSessionTargetSelectionPtrFromShared(bindingProjection.TargetSelection),
		SessionBinding:           browserRuntimeBuildSessionBinding(ctx, nil, nil, nil, bindingRoute, nil, &bindingProjection.Evaluation),
	}
}

func browserRuntimeDegradedActionSurface(
	action string,
	projection browserRuntimeDegradedRouteProjection,
) agentxbrowserruntime.SharedSessionBrowserActionSurfaceProjection {
	return agentxbrowserruntime.BuildSharedSessionBrowserDegradedActionSurfaceProjection(
		action,
		agentxbrowserruntime.SharedSessionBrowserDegradedActionSurfaceRequest{
			SelectedInfo:            projection.Route,
			RequestedDefaultProfile: projection.DefaultProfile,
			ProfileStatus: func() *agentxbrowserruntime.SharedSessionBrowserProfileState {
				if projection.ProfileStatus == nil {
					return nil
				}
				state := browserRuntimeSharedSessionProfileState(*projection.ProfileStatus)
				return &state
			}(),
			Profiles: browserRuntimeSharedProjectedProfiles(projection.Profiles),
			SessionProjection: &agentxbrowserruntime.SharedSessionBrowserTopLevelSessionProjection{
				Routes:      browserRuntimeSharedSessionRouteSnapshots(projection.SessionRoutes),
				TargetCount: projection.SessionTargetCount,
				Runs:        browserRuntimeSharedSessionRunsFromBinding(projection.SessionRuns),
				Profiles:    browserRuntimeSharedProjectedProfiles(projection.SessionProfiles),
				Handoff:     agentxbrowserruntime.CloneSharedSessionBrowserSessionHandoffSummary(projection.SessionHandoff),
			},
			WorkbenchSurface: browserRuntimeSharedWorkbenchSessionSurfaceFromDegradedProjection(projection),
		},
	)
}

func browserRuntimeSharedWorkbenchSessionSurfaceFromDegradedProjection(
	projection browserRuntimeDegradedRouteProjection,
) *agentxbrowserruntime.SharedSessionBrowserWorkbenchSessionSurfaceProjection {
	var bindingEvaluation *agentxbrowserruntime.SharedSessionBrowserBindingEvaluation
	if projection.SessionBinding != nil {
		evaluation := browserRuntimeSharedBindingEvaluation(*projection.SessionBinding, projection.SessionRoutes)
		bindingEvaluation = &evaluation
	}
	return agentxbrowserruntime.BuildSharedSessionBrowserDegradedWorkbenchSurface(
		agentxbrowserruntime.SharedSessionBrowserDegradedWorkbenchSurfaceRequest{
			SelectedInfo:            projection.Route,
			RequestedDefaultProfile: projection.DefaultProfile,
			ProfileStatus: func() *agentxbrowserruntime.SharedSessionBrowserProfileState {
				if projection.ProfileStatus == nil {
					return nil
				}
				state := browserRuntimeSharedSessionProfileState(*projection.ProfileStatus)
				return &state
			}(),
			Profiles: browserRuntimeSharedProjectedProfiles(projection.Profiles),
			SessionProjection: &agentxbrowserruntime.SharedSessionBrowserTopLevelSessionProjection{
				Routes:      browserRuntimeSharedSessionRouteSnapshots(projection.SessionRoutes),
				TargetCount: projection.SessionTargetCount,
				Runs:        browserRuntimeSharedSessionRunsFromBinding(projection.SessionRuns),
				Profiles:    browserRuntimeSharedProjectedProfiles(projection.WorkbenchSessionProfiles),
				Handoff:     agentxbrowserruntime.CloneSharedSessionBrowserSessionHandoffSummary(projection.WorkbenchSessionHandoff),
			},
			BindingEvaluation: bindingEvaluation,
			ProfileSelection:  browserRuntimeSharedProfileSelection(projection.SessionProfileSelection),
			TargetSelection:   browserRuntimeSharedTargetSelection(projection.SessionTargetSelection),
		},
	)
}
