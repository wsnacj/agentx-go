package browserruntime

import (
	"context"
	"errors"
	"testing"
	"time"
)

type executionTestBackend struct {
	statusReqs   []BrowserProfileStatusRequest
	statusResp   BrowserProfileStatusResult
	statusErr    error
	startReqs    []BrowserProfileLifecycleRequest
	startResp    BrowserProfileStatusResult
	startErr     error
	rawStart     func(context.Context, string) SharedSessionBrowserRawLifecycleObservation
	stopReqs     []BrowserProfileLifecycleRequest
	stopResp     BrowserProfileStatusResult
	stopErr      error
	rawStop      func(context.Context, string) SharedSessionBrowserRawLifecycleObservation
	profilesReqs []BrowserProfilesRequest
	profilesResp BrowserProfilesResult
	profilesErr  error
	createReqs   []BrowserProfileCreateRequest
	createResp   BrowserProfileStatusResult
	createErr    error
	deleteReqs   []BrowserProfileDeleteRequest
	deleteResp   BrowserProfileStatusResult
	deleteErr    error
}

func (b *executionTestBackend) RuntimeStatus(_ context.Context, req BrowserProfileStatusRequest) (BrowserProfileStatusResult, error) {
	b.statusReqs = append(b.statusReqs, req)
	return b.statusResp, b.statusErr
}

func (b *executionTestBackend) RuntimeStart(_ context.Context, req BrowserProfileLifecycleRequest) (BrowserProfileStatusResult, error) {
	b.startReqs = append(b.startReqs, req)
	return b.startResp, b.startErr
}

func (b *executionTestBackend) RuntimeStop(_ context.Context, req BrowserProfileLifecycleRequest) (BrowserProfileStatusResult, error) {
	b.stopReqs = append(b.stopReqs, req)
	return b.stopResp, b.stopErr
}

func (b *executionTestBackend) ObserveRawBrowserRuntimeStart(ctx context.Context, profile string) SharedSessionBrowserRawLifecycleObservation {
	if b.rawStart == nil {
		return SharedSessionBrowserRawLifecycleObservation{}
	}
	return b.rawStart(ctx, profile)
}

func (b *executionTestBackend) ObserveRawBrowserRuntimeStop(ctx context.Context, profile string) SharedSessionBrowserRawLifecycleObservation {
	if b.rawStop == nil {
		return SharedSessionBrowserRawLifecycleObservation{}
	}
	return b.rawStop(ctx, profile)
}

func (b *executionTestBackend) RuntimeProfiles(_ context.Context, req BrowserProfilesRequest) (BrowserProfilesResult, error) {
	b.profilesReqs = append(b.profilesReqs, req)
	return b.profilesResp, b.profilesErr
}

func (b *executionTestBackend) RuntimeCreateProfile(_ context.Context, req BrowserProfileCreateRequest) (BrowserProfileStatusResult, error) {
	b.createReqs = append(b.createReqs, req)
	return b.createResp, b.createErr
}

func (b *executionTestBackend) RuntimeDeleteProfile(_ context.Context, req BrowserProfileDeleteRequest) (BrowserProfileStatusResult, error) {
	b.deleteReqs = append(b.deleteReqs, req)
	return b.deleteResp, b.deleteErr
}

func TestExecuteSharedSessionBrowserCreateProfileFallsBackToCreatedStatus(t *testing.T) {
	backend := &executionTestBackend{}
	result, err := ExecuteSharedSessionBrowserCreateProfile(context.Background(), backend, SharedSessionBrowserProfileCreateRequest{
		RequestedProfile: "workbench",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
		BrowserApp: "Chromium",
		Color:      "#ff5500",
		CopyFrom:   "isolated",
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserCreateProfile returned error: %v", err)
	}
	if len(backend.createReqs) != 1 || backend.createReqs[0].Profile != "workbench" || backend.createReqs[0].BrowserApp != "Chromium" || backend.createReqs[0].Color != "#ff5500" || backend.createReqs[0].CopyFrom != "isolated" {
		t.Fatalf("expected create request to reach backend, got %#v", backend.createReqs)
	}
	if result.Decision != "created" || !result.Ready {
		t.Fatalf("expected shared create execution to report created+ready, got %#v", result)
	}
	if result.ProfileStatus.Backend != "proxy" || result.ProfileStatus.Profile != "workbench" || result.ProfileStatus.BrowserApp != "Chromium" || result.ProfileStatus.Status != "created" {
		t.Fatalf("expected create fallback status to preserve backend/profile/browser app, got %#v", result.ProfileStatus)
	}
}

func TestBuildSharedSessionBrowserExecutionRequest(t *testing.T) {
	req := BuildSharedSessionBrowserExecutionRequest(
		nil,
		" sess-1 ",
		" work ",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "work", Target: "node"},
		true,
		" run-1 ",
		" work ",
		SharedSessionBrowserHealthInput{StoredState: "healthy"},
		time.Minute,
	)
	if req.SessionID != "sess-1" || req.RequestedProfile != "work" || req.ActiveNodeRunID != "run-1" || req.ActiveBrowserProfile != "work" || !req.Force || req.HealthInput.StoredState != "healthy" || req.ReconnectWindow != time.Minute {
		t.Fatalf("unexpected shared execution request: %#v", req)
	}
}

func TestDispatchSharedSessionBrowserLifecycleActionUsesRefreshRestartContract(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
		startResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "started",
			Running:   true,
			Connected: false,
		},
	}
	dispatched := DispatchSharedSessionBrowserLifecycleAction(
		context.Background(),
		SharedSessionBrowserLifecycleActionDispatchRequest{
			Action:  "refresh",
			Control: backend,
			ExecutionRequest: SharedSessionBrowserExecutionRequest{
				RequestedProfile: "isolated",
				SelectedInfo: BrowserRuntimeInfo{
					Backend: "proxy",
					Target:  "node",
				},
				ReconnectWindow: time.Minute,
			},
		},
	)
	if !dispatched.Handled || dispatched.Err != nil {
		t.Fatalf("expected lifecycle dispatch helper to handle refresh restart contract, got %#v", dispatched)
	}
	if !dispatched.RememberProfile || dispatched.Result.Decision != "restart_started" || dispatched.Result.Profile != "isolated" {
		t.Fatalf("expected refresh dispatch to route through restart contract, got %#v", dispatched)
	}
	if len(backend.statusReqs) != 1 || len(backend.startReqs) != 1 {
		t.Fatalf("expected refresh dispatch to query status then start once, got status=%d start=%d", len(backend.statusReqs), len(backend.startReqs))
	}
}

func TestDispatchSharedSessionBrowserLifecycleActionUsesCreateProfileContract(t *testing.T) {
	backend := &executionTestBackend{
		createResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "workbench",
			Status:     "created",
			BrowserApp: "Chromium",
		},
	}
	dispatched := DispatchSharedSessionBrowserLifecycleAction(
		context.Background(),
		SharedSessionBrowserLifecycleActionDispatchRequest{
			Action:  "create_profile",
			Manager: backend,
			ProfileCreateRequest: SharedSessionBrowserProfileCreateRequest{
				RequestedProfile: "workbench",
				SelectedInfo:     BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
				BrowserApp:       "Chromium",
			},
		},
	)
	if !dispatched.Handled || dispatched.Err != nil {
		t.Fatalf("expected lifecycle dispatch helper to handle create_profile, got %#v", dispatched)
	}
	if !dispatched.RememberProfile || dispatched.Result.Decision != "created" || dispatched.Result.Profile != "workbench" {
		t.Fatalf("expected create_profile dispatch to preserve create contract, got %#v", dispatched)
	}
	if len(backend.createReqs) != 1 || backend.createReqs[0].Profile != "workbench" {
		t.Fatalf("expected create_profile dispatch to reach backend, got %#v", backend.createReqs)
	}
}

func TestDispatchSharedSessionBrowserLifecycleActionReturnsUnhandledForUnknownAction(t *testing.T) {
	dispatched := DispatchSharedSessionBrowserLifecycleAction(
		context.Background(),
		SharedSessionBrowserLifecycleActionDispatchRequest{Action: "unknown"},
	)
	if dispatched.Handled {
		t.Fatalf("expected unknown lifecycle action to remain unhandled, got %#v", dispatched)
	}
}

func TestObserveSharedSessionBrowserExecutionStatusUsesResolvedFallbackOnError(t *testing.T) {
	backend := &executionTestBackend{statusErr: errors.New("status unavailable")}
	observation := ObserveSharedSessionBrowserExecutionStatus(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		HealthInput: SharedSessionBrowserHealthInput{
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				Status:        "reconnecting",
				Running:       true,
				Connected:     false,
			}},
		},
		ReconnectWindow: time.Minute,
	}, "isolated", BrowserProfileStatusResult{})
	if observation.StatusErr == nil || observation.StatusErr.Error() != "status unavailable" {
		t.Fatalf("expected raw status error, got %#v", observation.StatusErr)
	}
	if observation.HasStatus {
		t.Fatalf("expected no raw status on error, got %#v", observation)
	}
	if observation.ResolvedStatus.Profile != "isolated" || observation.ResolvedStatus.Status != "reconnecting" || !observation.ResolvedStatus.Running || observation.ResolvedStatus.Connected {
		t.Fatalf("expected resolved fallback status, got %#v", observation.ResolvedStatus)
	}
}

func TestObserveSharedSessionBrowserExecutionStatusUsesExecutionResolutionOnSuccess(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Profile:   "work",
			Status:    "running",
			Connected: true,
		},
	}
	registry := NewBrowserSessionStateRegistry()
	observation := ObserveSharedSessionBrowserExecutionStatus(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		Registry:  registry,
		SessionID: "s1",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "work",
			Target:  "node",
		},
		ReconnectWindow: time.Minute,
	}, "work", BrowserProfileStatusResult{})
	if !observation.HasStatus || observation.Status.Profile != "work" {
		t.Fatalf("expected raw status observation, got %#v", observation)
	}
	if observation.ResolvedStatus.Profile != "work" || observation.ResolvedStatus.Status != "running" || !observation.ResolvedStatus.Connected {
		t.Fatalf("expected execution-resolved status to preserve raw success semantics, got %#v", observation.ResolvedStatus)
	}
}

func TestObserveSharedSessionBrowserExecutionStartUsesLifecycleDecisionStatus(t *testing.T) {
	backend := &executionTestBackend{
		startResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "started",
			Running:   true,
			Connected: false,
		},
	}
	observation := ObserveSharedSessionBrowserExecutionStart(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
	}, "isolated", "started", BrowserProfileStatusResult{})
	if observation.Err != nil {
		t.Fatalf("expected successful start observation, got %v", observation.Err)
	}
	if observation.Profile != "isolated" || observation.Status.Status != "starting" || !observation.Status.Running || observation.Status.Connected || observation.Status.Note != "start requested" || observation.Ready {
		t.Fatalf("expected lifecycle-owned starting observation, got %#v", observation)
	}
}

func TestObserveSharedSessionBrowserExecutionStartPreservesFallbackOnError(t *testing.T) {
	backend := &executionTestBackend{startErr: errors.New("start failed")}
	observation := ObserveSharedSessionBrowserExecutionStart(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
	}, "isolated", "started", BrowserProfileStatusResult{
		Backend:   "proxy",
		Profile:   "isolated",
		Status:    "stopped",
		Running:   false,
		Connected: false,
	})
	if observation.Err == nil || observation.Err.Error() != "start failed" {
		t.Fatalf("expected start error, got %#v", observation.Err)
	}
	if observation.Status.Profile != "isolated" || observation.Status.Status != "stopped" {
		t.Fatalf("expected fallback status on start error, got %#v", observation.Status)
	}
}

func TestObserveSharedSessionBrowserExecutionStopUsesLifecycleDecisionStatus(t *testing.T) {
	backend := &executionTestBackend{
		stopResp: BrowserProfileStatusResult{
			Backend: "proxy",
			Profile: "isolated",
		},
	}
	observation := ObserveSharedSessionBrowserExecutionStop(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
	}, "isolated", "stopped", BrowserProfileStatusResult{
		Backend:   "proxy",
		Profile:   "isolated",
		Status:    "connected",
		Running:   true,
		Connected: true,
	})
	if observation.Err != nil {
		t.Fatalf("expected successful stop observation, got %v", observation.Err)
	}
	if observation.Status.Profile != "isolated" || observation.Status.Status != "stopped" || observation.Status.Running || observation.Status.Connected || !observation.Ready {
		t.Fatalf("expected lifecycle-owned stopped observation, got %#v", observation)
	}
}

func TestExecuteSharedSessionBrowserDeleteProfileBlockedActiveNodeRunKeepsEffectiveLifecycleState(t *testing.T) {
	result, err := ExecuteSharedSessionBrowserDeleteProfile(context.Background(), &executionTestBackend{}, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		ActiveNodeRunID: "run-93",
		HealthInput: SharedSessionBrowserHealthInput{
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
			}},
		},
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserDeleteProfile returned error: %v", err)
	}
	if result.Decision != "delete_profile_blocked_active_node_run" || result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "running" || !result.ProfileStatus.Running || !result.ProfileStatus.Connected {
		t.Fatalf("expected blocked delete to preserve effective lifecycle state, got %#v", result)
	}
}

func TestExecuteSharedSessionBrowserDeleteProfileDeletedInvalidatesSessionState(t *testing.T) {
	backend := &executionTestBackend{
		deleteResp: BrowserProfileStatusResult{
			Backend: "proxy",
			Profile: "workbench",
			Status:  "deleted",
		},
	}
	result, err := ExecuteSharedSessionBrowserDeleteProfile(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "workbench",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
		Force: true,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserDeleteProfile returned error: %v", err)
	}
	if len(backend.deleteReqs) != 1 || backend.deleteReqs[0].Profile != "workbench" || !backend.deleteReqs[0].Force {
		t.Fatalf("expected delete request to reach backend, got %#v", backend.deleteReqs)
	}
	if result.Decision != "deleted" || !result.Ready || !result.InvalidateSessionTargets || !result.InvalidateSessionProfile {
		t.Fatalf("expected shared delete execution to invalidate session state, got %#v", result)
	}
}

func TestExecuteSharedSessionBrowserStopBlockedActiveNodeRunKeepsEffectiveLifecycleState(t *testing.T) {
	result, err := ExecuteSharedSessionBrowserStop(context.Background(), &executionTestBackend{}, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		ActiveNodeRunID: "run-92",
		HealthInput: SharedSessionBrowserHealthInput{
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
			}},
		},
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserStop returned error: %v", err)
	}
	if result.Decision != "stop_blocked_active_node_run" || result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "running" || !result.ProfileStatus.Running || !result.ProfileStatus.Connected {
		t.Fatalf("expected blocked stop to preserve effective lifecycle state, got %#v", result)
	}
}

func TestExecuteSharedSessionBrowserEnsurePreparedUsesLifecycleStateForWeakStartResult(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "stopped",
			Running:   false,
			Connected: false,
		},
		startResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "started",
			Running:   true,
			Connected: false,
		},
	}
	result, err := ExecuteSharedSessionBrowserEnsurePrepared(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserEnsurePrepared returned error: %v", err)
	}
	if result.Decision != "started" || result.Ready {
		t.Fatalf("expected started but not-yet-ready prepare result, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "starting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "start requested" {
		t.Fatalf("expected weak start result to collapse into lifecycle-owned starting state, got %#v", result.ProfileStatus)
	}
}

func TestExecuteSharedSessionBrowserStartUsesLifecycleStateForWeakStartResult(t *testing.T) {
	backend := &executionTestBackend{
		startResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "started",
			Running:   true,
			Connected: false,
		},
	}
	result, err := ExecuteSharedSessionBrowserStart(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserStart returned error: %v", err)
	}
	if result.Decision != "started" || result.Ready {
		t.Fatalf("expected started but not-yet-ready runtime start result, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "starting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "start requested" {
		t.Fatalf("expected weak start result to collapse into lifecycle-owned starting state, got %#v", result.ProfileStatus)
	}
}

func TestExecuteSharedSessionBrowserRestartBlockedActiveNodeRunKeepsEffectiveLifecycleState(t *testing.T) {
	result, err := ExecuteSharedSessionBrowserRestart(context.Background(), &executionTestBackend{}, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		ActiveNodeRunID: "run-91",
		HealthInput: SharedSessionBrowserHealthInput{
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
			}},
		},
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserRestart returned error: %v", err)
	}
	if result.Decision != "restart_blocked_active_node_run" || result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "running" || !result.ProfileStatus.Running || !result.ProfileStatus.Connected {
		t.Fatalf("expected blocked restart to preserve effective lifecycle state, got %#v", result)
	}
}

func TestExecuteSharedSessionBrowserRestartUsesLifecycleStateForWeakStartResult(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "connected",
			Running:   true,
			Connected: true,
		},
		startResp: BrowserProfileStatusResult{
			Backend:   "proxy",
			Profile:   "isolated",
			Status:    "started",
			Running:   true,
			Connected: false,
		},
	}
	result, err := ExecuteSharedSessionBrowserRestart(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		Force:           false,
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserRestart returned error: %v", err)
	}
	if len(backend.stopReqs) != 1 || backend.stopReqs[0].Profile != "isolated" {
		t.Fatalf("expected restart recovery to issue one stop for isolated profile, got %#v", backend.stopReqs)
	}
	if len(backend.startReqs) != 1 || backend.startReqs[0].Profile != "isolated" {
		t.Fatalf("expected restart recovery to issue one start for isolated profile, got %#v", backend.startReqs)
	}
	if result.Decision != "restarted" || result.Ready {
		t.Fatalf("expected restarted but not-yet-ready restart result, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "reconnecting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "restart requested" {
		t.Fatalf("expected weak restart start result to collapse into lifecycle-owned reconnecting state, got %#v", result.ProfileStatus)
	}
}

func TestExecuteSharedSessionBrowserRestartDefersDuringCooldownActive(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "isolated",
			Status:     "disconnected",
			Running:    false,
			Connected:  false,
			Note:       "cdp transport closed",
		},
	}
	result, err := ExecuteSharedSessionBrowserRestart(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		HealthInput: SharedSessionBrowserHealthInput{
			StoredState:               "cooldown_active",
			StoredReason:              "browser restart cooldown active for 900ms after 2 disconnects",
			StoredRecoveryAction:      "browser action=wait",
			StoredReconnectHint:       "retry_after_cooldown",
			StoredCooldownRemainingMs: 900,
		},
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserRestart returned error: %v", err)
	}
	if len(backend.statusReqs) != 1 || len(backend.stopReqs) != 0 || len(backend.startReqs) != 0 {
		t.Fatalf("expected cooldown_active restart to stop after status observation, got status=%d stop=%d start=%d", len(backend.statusReqs), len(backend.stopReqs), len(backend.startReqs))
	}
	if result.Decision != "cooldown_active" || result.Ready {
		t.Fatalf("expected cooldown_active restart to defer lifecycle churn, got %#v", result)
	}
	if result.InvalidateSessionTargets {
		t.Fatalf("expected cooldown_active restart blocker not to invalidate session targets, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "disconnected" || result.ProfileStatus.Running || result.ProfileStatus.Connected {
		t.Fatalf("expected cooldown_active restart to preserve disconnected status, got %#v", result.ProfileStatus)
	}
}

func TestExecuteSharedSessionBrowserEnsurePreparedDefersPermanentRestartFailure(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "isolated",
			Status:     "disconnected",
			Running:    false,
			Connected:  false,
			Note:       "explicit browser.start required",
		},
	}
	result, err := ExecuteSharedSessionBrowserEnsurePrepared(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		HealthInput: SharedSessionBrowserHealthInput{
			StoredState:          "restart_failed_permanent",
			StoredReason:         "browser restart failed 2 times; explicit browser.start required",
			StoredRecoveryAction: "browser action=start",
			StoredReconnectHint:  "manual_restart_required",
		},
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserEnsurePrepared returned error: %v", err)
	}
	if len(backend.statusReqs) != 1 || len(backend.stopReqs) != 0 || len(backend.startReqs) != 0 {
		t.Fatalf("expected restart_failed_permanent prepare to stop after status observation, got status=%d stop=%d start=%d", len(backend.statusReqs), len(backend.stopReqs), len(backend.startReqs))
	}
	if result.Decision != "restart_failed_permanent" || result.Ready {
		t.Fatalf("expected restart_failed_permanent prepare to require explicit start, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "disconnected" || result.ProfileStatus.Running || result.ProfileStatus.Connected {
		t.Fatalf("expected restart_failed_permanent prepare to preserve disconnected status, got %#v", result.ProfileStatus)
	}
}

func TestExecuteSharedSessionBrowserEnsurePreparedKeepsStartingStatusInProgress(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "isolated",
			Status:     "starting",
			Running:    true,
			Connected:  false,
			Note:       "start requested",
		},
	}
	result, err := ExecuteSharedSessionBrowserEnsurePrepared(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserEnsurePrepared returned error: %v", err)
	}
	if len(backend.statusReqs) != 1 || len(backend.stopReqs) != 0 || len(backend.startReqs) != 0 {
		t.Fatalf("expected starting prepare to stop after status observation, got status=%d stop=%d start=%d", len(backend.statusReqs), len(backend.stopReqs), len(backend.startReqs))
	}
	if result.Decision != "started" || result.Ready {
		t.Fatalf("expected starting prepare to remain in-progress, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "starting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "start requested" {
		t.Fatalf("expected starting prepare to preserve lifecycle-owned starting state, got %#v", result.ProfileStatus)
	}
}

func TestExecuteSharedSessionBrowserEnsurePreparedKeepsStartedButNotConnectedStatusInProgress(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "isolated",
			Status:     "started",
			Running:    true,
			Connected:  false,
		},
	}
	result, err := ExecuteSharedSessionBrowserEnsurePrepared(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserEnsurePrepared returned error: %v", err)
	}
	if len(backend.statusReqs) != 1 || len(backend.stopReqs) != 0 || len(backend.startReqs) != 0 {
		t.Fatalf("expected started-but-not-connected prepare to stop after status observation, got status=%d stop=%d start=%d", len(backend.statusReqs), len(backend.stopReqs), len(backend.startReqs))
	}
	if result.Decision != "started" || result.Ready {
		t.Fatalf("expected started-but-not-connected prepare to remain in-progress, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "starting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "start requested" {
		t.Fatalf("expected started-but-not-connected prepare to collapse into lifecycle-owned starting state, got %#v", result.ProfileStatus)
	}
}

func TestExecuteSharedSessionBrowserEnsurePreparedKeepsReconnectingStatusInProgress(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "isolated",
			Status:     "reconnecting",
			Running:    true,
			Connected:  false,
			Note:       "cdp reconnect in progress",
		},
	}
	result, err := ExecuteSharedSessionBrowserEnsurePrepared(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Profile: "isolated",
			Target:  "node",
		},
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserEnsurePrepared returned error: %v", err)
	}
	if len(backend.statusReqs) != 1 || len(backend.stopReqs) != 0 || len(backend.startReqs) != 0 {
		t.Fatalf("expected reconnecting prepare to stop after status observation, got status=%d stop=%d start=%d", len(backend.statusReqs), len(backend.stopReqs), len(backend.startReqs))
	}
	if result.Decision != "restart_reconnect_in_progress" || result.Ready {
		t.Fatalf("expected reconnecting prepare to remain in-progress, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "reconnecting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "cdp reconnect in progress" {
		t.Fatalf("expected reconnecting prepare to preserve lifecycle-owned reconnecting state, got %#v", result.ProfileStatus)
	}
}

func TestResolveSharedSessionBrowserExecutionResultUsesRegistryResolution(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-execution-resolution-registry"
	resolution := ResolveSharedSessionBrowserExecutionResult(
		registry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"",
		SharedSessionBrowserExecutionResult{
			Profile: "isolated",
			Profiles: &BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: false},
					{Profile: "relay", BrowserApp: "Chromium", Status: "stopped", Running: false, Connected: false},
				},
			},
			Decision: "started",
			ProfileStatus: BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "isolated",
				Status:    "started",
				Running:   true,
				Connected: false,
			},
		},
		time.Minute,
	)
	if !resolution.HasSyncedState {
		t.Fatalf("expected registry-backed execution resolution to return synced state")
	}
	if resolution.ResolvedStatus.Profile != "isolated" || resolution.ResolvedStatus.Status != "starting" || !resolution.ResolvedStatus.Running || resolution.ResolvedStatus.Connected {
		t.Fatalf("expected resolved status to use lifecycle-owned starting state, got %#v", resolution.ResolvedStatus)
	}
	if len(resolution.Snapshot) != 2 || resolution.Snapshot[0].Profile != "isolated" || resolution.Snapshot[0].Status != "starting" {
		t.Fatalf("expected registry-backed execution resolution to return scoped snapshot, got %#v", resolution.Snapshot)
	}
}

func TestResolveSharedSessionBrowserExecutionResultFallsBackWithoutRegistry(t *testing.T) {
	resolution := ResolveSharedSessionBrowserExecutionResult(
		nil,
		"",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		"",
		SharedSessionBrowserExecutionResult{
			Profile: "isolated",
			Profiles: &BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: false},
				},
			},
			Decision: "started",
			ProfileStatus: BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "isolated",
				Status:    "started",
				Running:   true,
				Connected: false,
			},
		},
		time.Minute,
	)
	if !resolution.HasSyncedState {
		t.Fatalf("expected lifecycle fallback resolution to produce synced state")
	}
	if resolution.ResolvedStatus.Profile != "isolated" || resolution.ResolvedStatus.Status != "starting" || !resolution.ResolvedStatus.Running || resolution.ResolvedStatus.Connected {
		t.Fatalf("expected fallback resolution to use lifecycle-owned starting state, got %#v", resolution.ResolvedStatus)
	}
	if len(resolution.Snapshot) != 1 || resolution.Snapshot[0].Profile != "isolated" || resolution.Snapshot[0].Status != "running" {
		t.Fatalf("expected fallback resolution to preserve raw discovered snapshot, got %#v", resolution.Snapshot)
	}
}

func TestApplySharedSessionBrowserExecutionCleanupClearsMatchingRouteStateAndProfileSelection(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-execution-cleanup"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})

	cleanup := ApplySharedSessionBrowserExecutionCleanup(
		sessionRegistry,
		stateRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		SharedSessionBrowserExecutionResult{
			Profile:                  "workbench",
			InvalidateSessionTargets: true,
			InvalidateSessionProfile: true,
		},
		SharedSessionBrowserExecutionResolution{
			ResolvedStatus: BrowserProfileStatusResult{
				Backend: "proxy",
				Profile: "workbench",
				Status:  "deleted",
			},
		},
	)
	if cleanup.ClearedSessionTargets != 1 || !cleanup.ClearedSessionProfile {
		t.Fatalf("expected cleanup to clear matching route state and selection, got %#v", cleanup)
	}
	if selection := CurrentSharedSessionBrowserTargetSelection(sessionRegistry, sessionID, route); selection != nil {
		t.Fatalf("expected route target selection to be cleared, got %#v", selection)
	}
	if _, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); ok {
		t.Fatalf("expected matching remembered profile selection to be cleared")
	}
}

func TestApplySharedSessionBrowserExecutionCleanupPreservesDifferentRememberedProfile(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-execution-cleanup-preserve"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})

	cleanup := ApplySharedSessionBrowserExecutionCleanup(
		sessionRegistry,
		stateRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		SharedSessionBrowserExecutionResult{
			Profile:                  "other",
			InvalidateSessionTargets: true,
			InvalidateSessionProfile: true,
		},
		SharedSessionBrowserExecutionResolution{
			ResolvedStatus: BrowserProfileStatusResult{
				Backend: "proxy",
				Profile: "other",
				Status:  "deleted",
			},
		},
	)
	if cleanup.ClearedSessionTargets != 0 || cleanup.ClearedSessionProfile {
		t.Fatalf("expected cleanup to preserve different route/profile state, got %#v", cleanup)
	}
	if selection := CurrentSharedSessionBrowserTargetSelection(sessionRegistry, sessionID, route); selection == nil || selection.Profile != "workbench" {
		t.Fatalf("expected workbench route target selection to be preserved, got %#v", selection)
	}
	if selection, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); !ok || selection.Profile != "workbench" {
		t.Fatalf("expected remembered workbench selection to remain, got %#v ok=%v", selection, ok)
	}
}

func TestApplySharedSessionBrowserExecutionCleanupUsesResolvedStatusFallback(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-execution-cleanup-resolved"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://node.example/isolated",
		Title:      "Isolated",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	cleanup := ApplySharedSessionBrowserExecutionCleanup(
		sessionRegistry,
		stateRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		SharedSessionBrowserExecutionResult{
			InvalidateSessionTargets: true,
			InvalidateSessionProfile: true,
		},
		SharedSessionBrowserExecutionResolution{
			ResolvedStatus: BrowserProfileStatusResult{
				Backend: "proxy",
				Profile: "isolated",
				Status:  "stopped",
			},
		},
	)
	if cleanup.ClearedSessionTargets != 1 || !cleanup.ClearedSessionProfile {
		t.Fatalf("expected cleanup to use resolved status fallback, got %#v", cleanup)
	}
	if selection := CurrentSharedSessionBrowserTargetSelection(sessionRegistry, sessionID, route); selection != nil {
		t.Fatalf("expected resolved-status cleanup to clear route target selection, got %#v", selection)
	}
	if _, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); ok {
		t.Fatalf("expected resolved-status cleanup to clear remembered profile selection")
	}
}

func TestApplySharedSessionBrowserExecutionResultProjectsScopedSnapshotBeforeCleanup(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "browser-runtime-execution-application"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}

	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Arc",
		Source:        "remember_profile",
	})

	application := ApplySharedSessionBrowserExecutionResult(
		sessionRegistry,
		stateRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"",
		SharedSessionBrowserExecutionResult{
			Profile: "workbench",
			Profiles: &BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", Status: "running", Running: true, Connected: false},
					{Profile: "relay", Status: "stopped"},
				},
			},
			Decision: "started",
			ProfileStatus: BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "workbench",
				Status:    "started",
				Running:   true,
				Connected: false,
			},
			InvalidateSessionTargets: true,
			InvalidateSessionProfile: true,
		},
		time.Minute,
	)
	if application.Resolution.ResolvedStatus.Status != "starting" || !application.Resolution.HasSyncedState {
		t.Fatalf("expected application to preserve lifecycle-owned resolution, got %#v", application.Resolution)
	}
	if len(application.ProjectedProfiles) != 2 {
		t.Fatalf("expected scoped projected profiles to use pre-cleanup snapshot selection, got %#v", application.ProjectedProfiles)
	}
	var workbench *SharedSessionBrowserProjectedProfileState
	for i := range application.ProjectedProfiles {
		if application.ProjectedProfiles[i].State.Profile == "workbench" {
			workbench = &application.ProjectedProfiles[i]
			break
		}
	}
	if workbench == nil || !workbench.Selected || workbench.State.BrowserApp != "Arc" || workbench.State.Status != "starting" {
		t.Fatalf("expected scoped projected workbench profile to retain pre-cleanup selection, got %#v", application.ProjectedProfiles)
	}
	if application.Cleanup.ClearedSessionTargets != 1 || !application.Cleanup.ClearedSessionProfile {
		t.Fatalf("expected application cleanup to clear matching session state, got %#v", application.Cleanup)
	}
	if selection := CurrentSharedSessionBrowserTargetSelection(sessionRegistry, sessionID, route); selection != nil {
		t.Fatalf("expected cleanup to clear tracked route target selection, got %#v", selection)
	}
	if _, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); ok {
		t.Fatalf("expected cleanup to clear remembered profile selection")
	}
}

func TestApplySharedSessionBrowserExecutionResultWithManagerUsesFallbackDependencies(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-apply-execution-manager"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	application := ApplySharedSessionBrowserExecutionResultWithManager(
		SharedSessionBrowserObserverManager{},
		sessionRegistry,
		stateRegistry,
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"",
		SharedSessionBrowserExecutionResult{
			Profile: "workbench",
			Profiles: &BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", Status: "running", Running: true, Connected: false},
				},
			},
			Decision: "started",
			ProfileStatus: BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "workbench",
				Status:    "started",
				Running:   true,
				Connected: false,
			},
			InvalidateSessionTargets: true,
			InvalidateSessionProfile: true,
		},
		time.Minute,
	)
	if application.Resolution.ResolvedStatus.Status != "starting" || !application.Resolution.HasSyncedState {
		t.Fatalf("expected fallback execution manager helper to preserve lifecycle-owned resolution, got %#v", application.Resolution)
	}
	if application.Cleanup.ClearedSessionTargets != 1 || !application.Cleanup.ClearedSessionProfile {
		t.Fatalf("expected fallback execution manager helper to use provided registries for cleanup, got %#v", application.Cleanup)
	}
	if selection := CurrentSharedSessionBrowserTargetSelection(sessionRegistry, sessionID, route); selection != nil {
		t.Fatalf("expected fallback execution manager helper to clear tracked route target selection, got %#v", selection)
	}
	if _, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); ok {
		t.Fatalf("expected fallback execution manager helper to clear remembered profile selection")
	}
}

func TestApplySharedSessionBrowserExecutionResultWithMutationContextUsesFallbackDependencies(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	stateRegistry := NewBrowserSessionStateRegistry()
	sessionID := "sess-apply-execution-mutation-context"
	route := BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"}
	sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		ID:         "tab-1",
		TabIndex:   1,
		URL:        "https://example.com/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	stateRegistry.SelectBrowserProfile(sessionID, SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "remember_profile",
	})

	application := ApplySharedSessionBrowserExecutionResultWithMutationContext(
		SharedSessionBrowserMutationContextFor(
			SharedSessionBrowserObserverManager{},
			sessionRegistry,
			stateRegistry,
			time.Minute,
		),
		sessionID,
		BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
		"",
		SharedSessionBrowserExecutionResult{
			Profile: "workbench",
			Profiles: &BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", Status: "running", Running: true, Connected: false},
				},
			},
			Decision: "started",
			ProfileStatus: BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "workbench",
				Status:    "started",
				Running:   true,
				Connected: false,
			},
			InvalidateSessionTargets: true,
			InvalidateSessionProfile: true,
		},
	)
	if application.Resolution.ResolvedStatus.Status != "starting" || !application.Resolution.HasSyncedState {
		t.Fatalf("expected mutation-context execution helper to preserve lifecycle-owned resolution, got %#v", application.Resolution)
	}
	if application.Cleanup.ClearedSessionTargets != 1 || !application.Cleanup.ClearedSessionProfile {
		t.Fatalf("expected mutation-context execution helper to use provided registries for cleanup, got %#v", application.Cleanup)
	}
	if selection := CurrentSharedSessionBrowserTargetSelection(sessionRegistry, sessionID, route); selection != nil {
		t.Fatalf("expected mutation-context execution helper to clear tracked route target selection, got %#v", selection)
	}
	if _, ok := stateRegistry.SelectedBrowserProfile(sessionID, "node"); ok {
		t.Fatalf("expected mutation-context execution helper to clear remembered profile selection")
	}
}

func TestExecuteSharedSessionBrowserTeardownWeakStopResultUsesLifecycleStoppedState(t *testing.T) {
	backend := &executionTestBackend{
		statusResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "isolated",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
		stopResp: BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "isolated",
		},
	}
	result, err := ExecuteSharedSessionBrowserTeardown(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
		ReconnectWindow: time.Minute,
	})
	if err != nil {
		t.Fatalf("ExecuteSharedSessionBrowserTeardown returned error: %v", err)
	}
	if result.Decision != "teardown_stopped" || !result.Ready {
		t.Fatalf("expected teardown stopped decision with ready result, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "stopped" {
		t.Fatalf("expected lifecycle-owned stopped teardown profile status, got %#v", result.ProfileStatus)
	}
}

func TestExecuteSharedSessionBrowserStopStatusFailureUsesEffectiveLifecycleState(t *testing.T) {
	backend := &executionTestBackend{
		statusErr: errors.New("status failed"),
	}
	result, err := ExecuteSharedSessionBrowserStop(context.Background(), backend, SharedSessionBrowserExecutionRequest{
		RequestedProfile: "isolated",
		SelectedInfo: BrowserRuntimeInfo{
			Backend: "proxy",
			Target:  "node",
		},
		Registry:  NewBrowserSessionStateRegistry(),
		SessionID: "s1",
		HealthInput: SharedSessionBrowserHealthInput{
			Profiles: []SharedSessionBrowserProfileState{{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "reconnecting",
				Running:       true,
				Connected:     false,
				Note:          "cdp reconnect in progress",
			}},
		},
		ReconnectWindow: time.Minute,
	})
	if err == nil || err.Error() != "status failed" {
		t.Fatalf("expected status failure, got result=%#v err=%v", result, err)
	}
	if result.Decision != "stop_status_failed" || result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "reconnecting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected {
		t.Fatalf("expected stop status failure to keep effective lifecycle state, got %#v", result)
	}
}
