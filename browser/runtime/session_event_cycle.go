package browserruntime

import (
	"context"
	"time"
)

// SharedSessionBrowserEventCycleObservation captures a single source-time
// status/profiles polling cycle after shared session-event sync has resolved
// the lifecycle-owned effective view.
type SharedSessionBrowserEventCycleObservation struct {
	Observation   SharedSessionBrowserStatusAndProfilesObservation
	ReferenceTime time.Time
}

// ObserveSharedSessionBrowserEventCycle runs the shared source-time status/
// profiles cycle used by watch, inspection, and execution observers before any
// higher-level binding/view projection is applied.
func ObserveSharedSessionBrowserEventCycle(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	includeStatus bool,
	includeProfiles bool,
	reconnectWindow time.Duration,
) SharedSessionBrowserEventCycleObservation {
	return sharedSessionBrowserObserverManager(
		nil,
		nil,
		registry,
		reconnectWindow,
	).ObserveEventCycle(ctx, control, SharedSessionBrowserObserverRequest{
		SessionID:        sessionID,
		SelectedInfo:     selectedInfo,
		BindingRoute:     BrowserSessionRoute{Backend: selectedInfo.Backend, Profile: selectedInfo.Profile, Target: selectedInfo.Target},
		RequestedProfile: requestedProfile,
		IncludeStatus:    includeStatus,
		IncludeProfiles:  includeProfiles,
	})
}

func sharedSessionBrowserLatestEventCycleObservedAt(statusObservedAt time.Time, profilesObservedAt time.Time) time.Time {
	if profilesObservedAt.After(statusObservedAt) {
		return profilesObservedAt
	}
	return statusObservedAt
}
