package browserruntime

import (
	"context"
	"strings"
)

// SharedSessionBrowserInspectionProjection captures the shared lifecycle-owned
// inspection projection that runtime status/workbench/profiles/sessions
// surfaces can marshal into tool-specific payloads.
type SharedSessionBrowserInspectionProjection struct {
	ProfileStatus      SharedSessionBrowserProfileState
	HasProfileStatus   bool
	RuntimeStatus      *BrowserProfileStatusResult
	Profiles           []SharedSessionBrowserProjectedProfileState
	HasProfiles        bool
	ProfilesErr        error
	DiscoveredProfiles []string
	DefaultProfile     string
	Note               string
	SessionProjection  SharedSessionBrowserTopLevelSessionProjection
	HasSessionView     bool
}

// SharedSessionBrowserInspectionActionObservation captures the shared
// request/watch/projection bundle produced by the browserruntime-owned
// inspection observation seam.
type SharedSessionBrowserInspectionActionObservation struct {
	Request    SharedSessionBrowserInspectionActionRequest
	Watch      SharedSessionBrowserInspectionObservation
	Projection SharedSessionBrowserInspectionProjection
}

// ObserveProjectedSharedSessionBrowserInspectionAction runs the shared
// action-aware inspection observation and projection owner so tools callers do
// not need to manually stitch watch/projection phases together.
func ObserveProjectedSharedSessionBrowserInspectionAction(
	ctx context.Context,
	watchManager SharedSessionBrowserWatchManager,
	req SharedSessionBrowserInspectionActionRequest,
	stateRegistry SharedSessionBrowserStateRegistry,
) SharedSessionBrowserInspectionActionObservation {
	req = normalizeSharedSessionBrowserInspectionActionRequest(req)
	watch := watchManager.ObserveInspectionAction(ctx, req)
	return SharedSessionBrowserInspectionActionObservation{
		Request:    req,
		Watch:      watch,
		Projection: ProjectSharedSessionBrowserInspectionAction(req, watch, stateRegistry),
	}
}

// ProjectSharedSessionBrowserInspectionAction lowers a shared watch
// observation onto the narrower inspection projection contract used by
// browser_runtime inspection-style payloads.
func ProjectSharedSessionBrowserInspectionAction(
	req SharedSessionBrowserInspectionActionRequest,
	watch SharedSessionBrowserWatchObservation,
	stateRegistry SharedSessionBrowserStateRegistry,
) SharedSessionBrowserInspectionProjection {
	req = normalizeSharedSessionBrowserInspectionActionRequest(req)

	observation := watch.View.Observation
	projection := SharedSessionBrowserInspectionProjection{
		ProfilesErr:        observation.ProfilesErr,
		DiscoveredProfiles: append([]string(nil), watch.DiscoveredProfiles...),
		DefaultProfile:     strings.TrimSpace(watch.DefaultProfile),
	}

	switch req.Action {
	case "profiles":
		if observation.ProfilesErr == nil {
			projection.Profiles = sharedSessionBrowserInspectionProjectedProfiles(req, watch, stateRegistry)
			projection.HasProfiles = observation.Profiles != nil || len(projection.Profiles) > 0
			projection.Note = strings.TrimSpace(watch.Note)
		}
	case "sessions":
		projection.SessionProjection = ProjectSharedSessionBrowserTopLevelSessionView(
			watch.View.Session,
			sharedSessionBrowserInspectionSessionFallbackProfiles(req, watch),
		)
		projection.HasSessionView = true
	default:
		if observation.HasSyncedState {
			projection.ProfileStatus = observation.SyncedState
			projection.HasProfileStatus = true
		} else if observation.Status != nil {
			projection.ProfileStatus = SharedSessionBrowserProfileStateFromStatus(req.SelectedInfo, observation.ResolvedStatus)
			projection.HasProfileStatus = true
		}
		if observation.Status != nil {
			status := *observation.Status
			projection.RuntimeStatus = &status
		}
		if observation.Profiles != nil {
			projection.Profiles = append([]SharedSessionBrowserProjectedProfileState(nil), watch.Profiles...)
			projection.HasProfiles = true
			projection.Note = strings.TrimSpace(watch.Note)
		}
		if req.Action == "workbench" && req.IncludeSessionView {
			projection.SessionProjection = ProjectSharedSessionBrowserTopLevelSessionView(
				watch.View.Session,
				sharedSessionBrowserInspectionSessionFallbackProfiles(req, watch),
			)
			projection.HasSessionView = true
		}
	}

	projection.Note = firstNonEmptyBindingString(
		strings.TrimSpace(sharedSessionBrowserInspectionErrString(observation.StatusErr)),
		strings.TrimSpace(sharedSessionBrowserInspectionErrString(observation.ProfilesErr)),
		strings.TrimSpace(projection.Note),
	)
	return projection
}

func sharedSessionBrowserInspectionProjectedProfiles(
	req SharedSessionBrowserInspectionActionRequest,
	watch SharedSessionBrowserWatchObservation,
	stateRegistry SharedSessionBrowserStateRegistry,
) []SharedSessionBrowserProjectedProfileState {
	projected := append([]SharedSessionBrowserProjectedProfileState(nil), watch.Profiles...)
	if req.ExplicitRequestedProfile {
		return projected
	}
	scoped := SnapshotSharedSessionBrowserProjectedProfilesForScope(
		stateRegistry,
		req.SessionID,
		BrowserRuntimeInfo{
			Backend: req.SelectedInfo.Backend,
			Target:  req.SelectedInfo.Target,
		},
		"",
	)
	if len(scoped) > 0 {
		return scoped
	}
	return projected
}

func sharedSessionBrowserInspectionSessionFallbackProfiles(
	req SharedSessionBrowserInspectionActionRequest,
	watch SharedSessionBrowserWatchObservation,
) []SharedSessionBrowserProjectedProfileState {
	if len(watch.View.Session.Profiles) > 0 || len(watch.View.Session.Routes) == 0 {
		return nil
	}
	if len(watch.Profiles) > 0 {
		return append([]SharedSessionBrowserProjectedProfileState(nil), watch.Profiles...)
	}
	var selection *SharedSessionBrowserProfileSelection
	if watch.View.Binding.Snapshot.SelectedProfileSelection != nil {
		cloned := *watch.View.Binding.Snapshot.SelectedProfileSelection
		selection = &cloned
	}
	return ProjectSharedSessionBrowserFallbackProfilesFromRouteSnapshots(
		req.SelectedInfo,
		req.RequestedProfile,
		watch.View.Session.Routes,
		selection,
	)
}

func sharedSessionBrowserInspectionErrString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
