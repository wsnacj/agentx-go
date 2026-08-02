package browserruntime

import (
	"context"
	"time"
)

// SharedSessionBrowserExecutionStatusObservation captures the raw RuntimeStatus
// load used by execution paths together with the lifecycle-owned resolved
// status they should consume on both success and fallback paths.
type SharedSessionBrowserExecutionStatusObservation struct {
	Status         BrowserProfileStatusResult
	HasStatus      bool
	StatusErr      error
	ObservedAt     time.Time
	ResolvedStatus BrowserProfileStatusResult
}

// SharedSessionBrowserExecutionLifecycleObservation captures the raw
// RuntimeStart/RuntimeStop load used by execution paths together with the
// lifecycle-owned mapped status and ready state they should consume.
type SharedSessionBrowserExecutionLifecycleObservation struct {
	Profile    string
	Status     BrowserProfileStatusResult
	Ready      bool
	Err        error
	ObservedAt time.Time
}

// ObserveSharedSessionBrowserExecutionStatus loads a RuntimeStatus observation
// for a managed profile and applies the shared execution fallback resolution
// when the backend cannot return a raw status.
func ObserveSharedSessionBrowserExecutionStatus(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserExecutionRequest,
	profile string,
	fallback BrowserProfileStatusResult,
) SharedSessionBrowserExecutionStatusObservation {
	return sharedSessionBrowserObserverManager(
		nil,
		nil,
		nil,
		req.ReconnectWindow,
	).ObserveExecutionStatus(ctx, control, req, profile, fallback)
}

// ObserveSharedSessionBrowserExecutionStart loads a RuntimeStart observation
// for a managed profile and applies the shared lifecycle decision mapping
// expected by execution paths.
func ObserveSharedSessionBrowserExecutionStart(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserExecutionRequest,
	profile string,
	decision string,
	fallback BrowserProfileStatusResult,
) SharedSessionBrowserExecutionLifecycleObservation {
	return sharedSessionBrowserObserverManager(
		nil,
		nil,
		nil,
		req.ReconnectWindow,
	).ObserveExecutionStart(ctx, control, req, profile, decision, fallback)
}

// ObserveSharedSessionBrowserExecutionStop loads a RuntimeStop observation for
// a managed profile and applies the shared lifecycle decision mapping expected
// by execution paths.
func ObserveSharedSessionBrowserExecutionStop(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	req SharedSessionBrowserExecutionRequest,
	profile string,
	decision string,
	fallback BrowserProfileStatusResult,
) SharedSessionBrowserExecutionLifecycleObservation {
	return sharedSessionBrowserObserverManager(
		nil,
		nil,
		nil,
		req.ReconnectWindow,
	).ObserveExecutionStop(ctx, control, req, profile, decision, fallback)
}
