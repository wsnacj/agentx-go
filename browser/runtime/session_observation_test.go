package browserruntime

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type statusProfilesObservationTestBackend struct {
	mu                        sync.Mutex
	statusReqs                []BrowserProfileStatusRequest
	statusResp                BrowserProfileStatusResult
	statusErr                 error
	statusStarted             chan struct{}
	statusBlock               <-chan struct{}
	profilesReqs              []BrowserProfilesRequest
	profilesResp              BrowserProfilesResult
	profilesErr               error
	profilesStarted           chan struct{}
	profilesBlock             <-chan struct{}
	rawStatusCalls            int
	rawStatus                 func(context.Context, string) SharedSessionBrowserRawStatusObservation
	rawProfilesCalls          int
	rawProfiles               func(context.Context, string) SharedSessionBrowserRawProfilesObservation
	rawStatusAndProfilesCalls int
	rawStatusAndProfiles      func(context.Context, string, bool, bool) SharedSessionBrowserRawStatusAndProfilesObservation
	rawTabsCalls              int
	rawTabs                   func(context.Context, string) SharedSessionBrowserRawTabsObservation
	rawNavigationCalls        int
	rawNavigation             func(context.Context, string) SharedSessionBrowserRawNavigationObservation
	rawOpenCalls              int
	rawOpen                   func(context.Context, string) SharedSessionBrowserRawOpenObservation
	rawTargetCalls            int
	rawTarget                 func(context.Context, string) SharedSessionBrowserRawTargetObservation
	rawRouteMutationCalls     int
	rawRouteMutation          func(context.Context, string) SharedSessionBrowserRawRouteMutationObservation
}

func (b *statusProfilesObservationTestBackend) RuntimeStatus(_ context.Context, req BrowserProfileStatusRequest) (BrowserProfileStatusResult, error) {
	b.mu.Lock()
	b.statusReqs = append(b.statusReqs, req)
	b.mu.Unlock()
	if b.statusStarted != nil {
		select {
		case b.statusStarted <- struct{}{}:
		default:
		}
	}
	if b.statusBlock != nil {
		<-b.statusBlock
	}
	return b.statusResp, b.statusErr
}

func (b *statusProfilesObservationTestBackend) RuntimeStart(context.Context, BrowserProfileLifecycleRequest) (BrowserProfileStatusResult, error) {
	return BrowserProfileStatusResult{}, nil
}

func (b *statusProfilesObservationTestBackend) RuntimeStop(context.Context, BrowserProfileLifecycleRequest) (BrowserProfileStatusResult, error) {
	return BrowserProfileStatusResult{}, nil
}

func (b *statusProfilesObservationTestBackend) RuntimeProfiles(_ context.Context, req BrowserProfilesRequest) (BrowserProfilesResult, error) {
	b.mu.Lock()
	b.profilesReqs = append(b.profilesReqs, req)
	b.mu.Unlock()
	if b.profilesStarted != nil {
		select {
		case b.profilesStarted <- struct{}{}:
		default:
		}
	}
	if b.profilesBlock != nil {
		<-b.profilesBlock
	}
	return b.profilesResp, b.profilesErr
}

func (b *statusProfilesObservationTestBackend) ObserveRawBrowserRuntimeStatus(ctx context.Context, requestedProfile string) SharedSessionBrowserRawStatusObservation {
	b.mu.Lock()
	b.rawStatusCalls++
	fn := b.rawStatus
	b.mu.Unlock()
	if fn == nil {
		return SharedSessionBrowserRawStatusObservation{}
	}
	return fn(ctx, requestedProfile)
}

func (b *statusProfilesObservationTestBackend) ObserveRawBrowserRuntimeProfiles(ctx context.Context, requestedProfile string) SharedSessionBrowserRawProfilesObservation {
	b.mu.Lock()
	b.rawProfilesCalls++
	fn := b.rawProfiles
	b.mu.Unlock()
	if fn == nil {
		return SharedSessionBrowserRawProfilesObservation{}
	}
	return fn(ctx, requestedProfile)
}

func (b *statusProfilesObservationTestBackend) ObserveRawBrowserRuntimeStatusAndProfiles(ctx context.Context, requestedProfile string, includeStatus bool, includeProfiles bool) SharedSessionBrowserRawStatusAndProfilesObservation {
	b.mu.Lock()
	b.rawStatusAndProfilesCalls++
	fn := b.rawStatusAndProfiles
	b.mu.Unlock()
	if fn == nil {
		return SharedSessionBrowserRawStatusAndProfilesObservation{}
	}
	return fn(ctx, requestedProfile, includeStatus, includeProfiles)
}

func (b *statusProfilesObservationTestBackend) ObserveRawBrowserTabs(ctx context.Context, requestedProfile string) SharedSessionBrowserRawTabsObservation {
	b.mu.Lock()
	b.rawTabsCalls++
	fn := b.rawTabs
	b.mu.Unlock()
	if fn == nil {
		return SharedSessionBrowserRawTabsObservation{}
	}
	return fn(ctx, requestedProfile)
}

func (b *statusProfilesObservationTestBackend) ObserveRawBrowserNavigation(ctx context.Context, requestedProfile string) SharedSessionBrowserRawNavigationObservation {
	b.mu.Lock()
	b.rawNavigationCalls++
	fn := b.rawNavigation
	b.mu.Unlock()
	if fn == nil {
		return SharedSessionBrowserRawNavigationObservation{}
	}
	return fn(ctx, requestedProfile)
}

func (b *statusProfilesObservationTestBackend) ObserveRawBrowserOpen(ctx context.Context, requestedProfile string) SharedSessionBrowserRawOpenObservation {
	b.mu.Lock()
	b.rawOpenCalls++
	fn := b.rawOpen
	b.mu.Unlock()
	if fn == nil {
		return SharedSessionBrowserRawOpenObservation{}
	}
	return fn(ctx, requestedProfile)
}

func (b *statusProfilesObservationTestBackend) ObserveRawBrowserTarget(ctx context.Context, requestedProfile string) SharedSessionBrowserRawTargetObservation {
	b.mu.Lock()
	b.rawTargetCalls++
	fn := b.rawTarget
	b.mu.Unlock()
	if fn == nil {
		return SharedSessionBrowserRawTargetObservation{}
	}
	return fn(ctx, requestedProfile)
}

func (b *statusProfilesObservationTestBackend) ObserveRawBrowserRouteMutation(ctx context.Context, requestedProfile string) SharedSessionBrowserRawRouteMutationObservation {
	b.mu.Lock()
	b.rawRouteMutationCalls++
	fn := b.rawRouteMutation
	b.mu.Unlock()
	if fn == nil {
		return SharedSessionBrowserRawRouteMutationObservation{}
	}
	return fn(ctx, requestedProfile)
}

func (b *statusProfilesObservationTestBackend) statusRequestCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.statusReqs)
}

func (b *statusProfilesObservationTestBackend) profilesRequestCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.profilesReqs)
}

func TestObserveSharedSessionBrowserStatusAndProfilesUsesRegistryResolution(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "work",
			Status:    "running",
			Connected: true,
			Note:      "status ok",
		},
		profilesResp: BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "work",
			Profiles: []BrowserProfileInfo{
				{Profile: "work", Status: "connected", Connected: true},
			},
		},
	}
	registry := NewBrowserSessionStateRegistry()
	selectedInfo := BrowserRuntimeInfo{Backend: "proxy", Target: "node", Profile: "default"}

	observation := ObserveSharedSessionBrowserStatusAndProfiles(
		context.Background(),
		backend,
		registry,
		"s1",
		selectedInfo,
		"work",
		true,
		true,
		time.Minute,
	)

	if len(backend.statusReqs) != 1 || backend.statusReqs[0].Profile != "work" {
		t.Fatalf("expected one status request for work profile, got %#v", backend.statusReqs)
	}
	if len(backend.profilesReqs) != 1 || backend.profilesReqs[0].Profile != "work" {
		t.Fatalf("expected one profiles request for work profile, got %#v", backend.profilesReqs)
	}
	if observation.Status == nil || observation.Profiles == nil {
		t.Fatalf("expected successful status and profiles observations, got %#v", observation)
	}
	if !observation.HasSyncedState {
		t.Fatalf("expected synced state from registry, got %#v", observation)
	}
	if observation.ResolvedStatus.Backend != "proxy" || observation.ResolvedStatus.Profile != "work" || observation.ResolvedStatus.Status != "running" {
		t.Fatalf("expected resolved status to inherit route identity, got %#v", observation.ResolvedStatus)
	}
	if len(observation.Snapshot) != 1 || observation.Snapshot[0].Profile != "work" || observation.Snapshot[0].RuntimeTarget != "node" {
		t.Fatalf("expected scoped snapshot for work@node, got %#v", observation.Snapshot)
	}
}

func TestObserveSharedSessionBrowserStatusUsesRegistryResolution(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "work",
			Status:    "running",
			Connected: true,
			Note:      "status ok",
		},
	}
	registry := NewBrowserSessionStateRegistry()

	observation := ObserveSharedSessionBrowserStatus(
		context.Background(),
		backend,
		registry,
		"s1",
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"work",
		time.Minute,
	)

	if len(backend.statusReqs) != 1 || backend.statusReqs[0].Profile != "work" {
		t.Fatalf("expected one status request for work profile, got %#v", backend.statusReqs)
	}
	if observation.Status == nil || observation.Status.Profile != "work" {
		t.Fatalf("expected raw status observation, got %#v", observation.Status)
	}
	if !observation.HasSyncedState || observation.ResolvedStatus.Backend != "proxy" || observation.ResolvedStatus.Profile != "work" || observation.ResolvedStatus.Status != "running" {
		t.Fatalf("expected registry-resolved status for work@proxy, got %#v", observation)
	}
}

func TestObserveSharedSessionBrowserStatusReusesSharedRegistryWatchManager(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "work",
			Status:    "running",
			Connected: true,
		},
	}
	registry := NewBrowserSessionStateRegistry()
	selectedInfo := BrowserRuntimeInfo{Backend: "proxy", Target: "node"}

	first := ObserveSharedSessionBrowserStatus(
		context.Background(),
		backend,
		registry,
		"s1",
		selectedInfo,
		"work",
		time.Minute,
	)
	second := ObserveSharedSessionBrowserStatus(
		context.Background(),
		backend,
		registry,
		"s1",
		selectedInfo,
		"work",
		time.Minute,
	)

	if first.Status == nil || second.Status == nil {
		t.Fatalf("expected successful registry-backed status observations, got first=%#v second=%#v", first, second)
	}
	if len(backend.statusReqs) != 1 {
		t.Fatalf("expected shared registry watch manager to reuse cached status, got %#v", backend.statusReqs)
	}
}

func TestObserveSharedSessionBrowserStatusPreservesIndependentError(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusErr: errors.New("status failed"),
	}

	observation := ObserveSharedSessionBrowserStatus(
		context.Background(),
		backend,
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"work",
		time.Minute,
	)

	if observation.StatusErr == nil || observation.StatusErr.Error() != "status failed" {
		t.Fatalf("expected independent status error, got %#v", observation.StatusErr)
	}
	if observation.Status != nil || observation.ResolvedStatus != (BrowserProfileStatusResult{}) || observation.HasSyncedState {
		t.Fatalf("expected empty fallback observation on status error, got %#v", observation)
	}
}

func TestObserveSharedSessionBrowserStatusAndProfilesFallsBackWithoutRegistry(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend: "proxy",
			Profile: "work",
			Status:  "running",
		},
		profilesResp: BrowserProfilesResult{
			Backend: "proxy",
			Profiles: []BrowserProfileInfo{
				{Profile: "work", Status: "running", Running: true},
			},
		},
	}

	observation := ObserveSharedSessionBrowserStatusAndProfiles(
		context.Background(),
		backend,
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Target: "node", Profile: "default"},
		"work",
		true,
		true,
		time.Minute,
	)

	if observation.HasSyncedState {
		t.Fatalf("expected no synced state without registry, got %#v", observation)
	}
	if observation.ResolvedStatus.Profile != "work" || observation.ResolvedStatus.Status != "running" {
		t.Fatalf("expected raw resolved status fallback, got %#v", observation.ResolvedStatus)
	}
	if len(observation.Snapshot) != 1 || observation.Snapshot[0].Profile != "work" {
		t.Fatalf("expected raw mapped snapshot fallback, got %#v", observation.Snapshot)
	}
}

func TestObserveSharedSessionBrowserStatusAndProfilesPreservesIndependentErrors(t *testing.T) {
	backend := &statusProfilesObservationTestBackend{
		statusErr: errors.New("status failed"),
		profilesResp: BrowserProfilesResult{
			Backend: "proxy",
			Profiles: []BrowserProfileInfo{
				{Profile: "work"},
			},
		},
	}

	observation := ObserveSharedSessionBrowserStatusAndProfiles(
		context.Background(),
		backend,
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"work",
		true,
		true,
		time.Minute,
	)

	if observation.StatusErr == nil || observation.StatusErr.Error() != "status failed" {
		t.Fatalf("expected independent status error, got %#v", observation.StatusErr)
	}
	if observation.Profiles == nil || len(observation.Snapshot) != 1 {
		t.Fatalf("expected successful profiles fallback snapshot, got %#v", observation)
	}
}
