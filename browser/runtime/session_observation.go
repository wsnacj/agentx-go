package browserruntime

import (
	"context"
	"time"
)

// SharedSessionBrowserStatusAndProfilesObservation captures route-scoped runtime
// status/profile discovery together with the lifecycle-owned resolved status and
// scoped synced snapshot when a registry session is available.
type SharedSessionBrowserStatusAndProfilesObservation struct {
	Status             *BrowserProfileStatusResult
	StatusErr          error
	StatusObservedAt   time.Time
	Profiles           *BrowserProfilesResult
	ProfilesErr        error
	ProfilesObservedAt time.Time
	ResolvedStatus     BrowserProfileStatusResult
	SyncedState        SharedSessionBrowserProfileState
	HasSyncedState     bool
	Snapshot           []SharedSessionBrowserProfileState
}

// SharedSessionBrowserStatusObservation captures a raw RuntimeStatus
// observation together with the lifecycle-owned resolved status when a session
// registry is available.
type SharedSessionBrowserStatusObservation struct {
	Status         *BrowserProfileStatusResult
	StatusErr      error
	ObservedAt     time.Time
	ResolvedStatus BrowserProfileStatusResult
	SyncedState    SharedSessionBrowserProfileState
	HasSyncedState bool
}

// ObserveSharedSessionBrowserStatus loads a RuntimeStatus observation and,
// when a session registry is available, resolves it through the shared scoped
// lifecycle snapshot before returning the lifecycle-owned status view.
func ObserveSharedSessionBrowserStatus(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	reconnectWindow time.Duration,
) SharedSessionBrowserStatusObservation {
	return sharedSessionBrowserObserverManager(
		nil,
		nil,
		registry,
		reconnectWindow,
	).ObserveStatus(ctx, control, sessionID, selectedInfo, requestedProfile)
}

// ObserveSharedSessionBrowserStatusAndProfiles loads optional status/profile
// observations from the backend and, when possible, syncs them through the
// shared session registry before returning the resolved lifecycle view.
func ObserveSharedSessionBrowserStatusAndProfiles(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	includeStatus bool,
	includeProfiles bool,
	reconnectWindow time.Duration,
) SharedSessionBrowserStatusAndProfilesObservation {
	return sharedSessionBrowserObserverManager(
		nil,
		nil,
		registry,
		reconnectWindow,
	).ObserveStatusAndProfiles(ctx, control, sessionID, selectedInfo, requestedProfile, includeStatus, includeProfiles)
}
