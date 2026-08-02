package tools

import (
	"context"
	"errors"
	"testing"
	"time"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func TestBrowserRuntimeRecordProfilesSyncsFullRouteScopeWhenProfileFilterBlank(t *testing.T) {
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-record-profiles-full-sync")
	sessionID := ToolSessionIDFromContext(callCtx)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-record-profiles-full-sync", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "relay",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       true,
		Connected:     false,
		Note:          "stale relay state",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-record-profiles-full-sync", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	scopedSnapshot := agentxbrowserruntime.SyncSharedSessionBrowserProfilesEvent(sessionStateRegistry, sessionID, agentxbrowserruntime.BrowserRuntimeInfo{Backend: "proxy", Target: "node"}, "", agentxbrowserruntime.BrowserProfilesResult{
		Backend: "proxy",
		Profiles: []agentxbrowserruntime.BrowserProfileInfo{
			{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
		},
	}, time.Time{}, browserRuntimeReconnectWatchdogWindow)
	if len(scopedSnapshot) != 1 || scopedSnapshot[0].Profile != "isolated" || scopedSnapshot[0].RuntimeTarget != "node" {
		t.Fatalf("expected recordProfiles to return scoped synced snapshot, got %#v", scopedSnapshot)
	}

	snapshot := sessionStateRegistry.SnapshotSessionBrowserProfiles("browser-runtime-record-profiles-full-sync")
	if len(snapshot) != 2 {
		t.Fatalf("expected full-scope profile sync to keep host state and replace node route state, got %#v", snapshot)
	}
	if snapshot[0].Backend != "proxy" || snapshot[0].Profile != "isolated" || snapshot[1].Backend != "system" || snapshot[1].Profile != "default" {
		t.Fatalf("unexpected full-scope profile sync snapshot: %#v", snapshot)
	}
}

func TestBrowserRuntimeRecordProfilesPreservesManagedTransitionState(t *testing.T) {
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-record-profiles-preserve-transition")
	sessionID := ToolSessionIDFromContext(callCtx)
	base := time.Now().Add(-2 * time.Minute)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-record-profiles-preserve-transition", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
		Note:          "cdp reconnect in progress",
	})

	scopedSnapshot := agentxbrowserruntime.SyncSharedSessionBrowserProfilesEvent(sessionStateRegistry, sessionID, agentxbrowserruntime.BrowserRuntimeInfo{Backend: "proxy", Target: "node"}, "isolated", agentxbrowserruntime.BrowserProfilesResult{
		Backend: "proxy",
		Profiles: []agentxbrowserruntime.BrowserProfileInfo{
			{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: false},
		},
	}, time.Time{}, browserRuntimeReconnectWatchdogWindow)
	if len(scopedSnapshot) != 1 || scopedSnapshot[0].Status != "reconnecting" || scopedSnapshot[0].Note != "cdp reconnect in progress" {
		t.Fatalf("expected recordProfiles to return lifecycle-owned scoped snapshot, got %#v", scopedSnapshot)
	}

	snapshot := sessionStateRegistry.SnapshotSessionBrowserProfiles("browser-runtime-record-profiles-preserve-transition")
	if len(snapshot) != 1 {
		t.Fatalf("expected one synced profile state, got %#v", snapshot)
	}
	if snapshot[0].Status != "reconnecting" || snapshot[0].Note != "cdp reconnect in progress" {
		t.Fatalf("expected profiles sync to preserve managed reconnect transition, got %#v", snapshot[0])
	}
	if !snapshot[0].StatusSince.Equal(base) {
		t.Fatalf("expected profiles sync to preserve status_since, got %#v", snapshot[0])
	}
}

func TestBrowserRuntimeRecordProfileStatusSyncsRouteProfileScope(t *testing.T) {
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-record-profile-status-sync")
	sessionID := ToolSessionIDFromContext(callCtx)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-record-profile-status-sync", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		Note:          "stale chromium state",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-record-profile-status-sync", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chrome",
		Status:        "stopped",
		Running:       false,
		Connected:     false,
		Note:          "stale chrome alias",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-record-profile-status-sync", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	agentxbrowserruntime.SyncSharedSessionBrowserProfileStatusEvent(sessionStateRegistry, nil, sessionID, agentxbrowserruntime.BrowserRuntimeInfo{Backend: "proxy", Target: "node"}, agentxbrowserruntime.BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "running",
		Running:    true,
		Connected:  true,
	}, time.Time{}, browserRuntimeReconnectWatchdogWindow)

	snapshot := sessionStateRegistry.SnapshotSessionBrowserProfiles("browser-runtime-record-profile-status-sync")
	if len(snapshot) != 2 {
		t.Fatalf("expected route-profile sync to replace stale duplicate states, got %#v", snapshot)
	}
	if snapshot[0].Backend != "proxy" || snapshot[0].Profile != "isolated" || snapshot[0].BrowserApp != "Chromium" || snapshot[0].Status != "running" {
		t.Fatalf("unexpected synced proxy profile state: %#v", snapshot[0])
	}
	if snapshot[1].Backend != "system" || snapshot[1].Profile != "default" {
		t.Fatalf("expected host state to be preserved, got %#v", snapshot)
	}
}

func TestBrowserRuntimeAdvanceSessionProfileStateSyncsRouteProfileScope(t *testing.T) {
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-advance-profile-state-sync")
	sessionID := ToolSessionIDFromContext(callCtx)
	base := time.Now().Add(-12 * time.Second)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-advance-profile-state-sync", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    base,
		StatusSince:   base,
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-advance-profile-state-sync", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chrome",
		Status:        "stopped",
	})

	agentxbrowserruntime.SyncSharedSessionBrowserProfileLifecycleEvent(sessionStateRegistry, nil, nil, sessionID, agentxbrowserruntime.BrowserRuntimeInfo{Backend: "proxy", Target: "node"}, "isolated", agentxbrowserruntime.BrowserProfileStatusResult{
		Backend:   "proxy",
		Profile:   "isolated",
		Status:    "starting",
		Running:   true,
		Connected: false,
	}, "restart_started", time.Time{}, browserRuntimeReconnectWatchdogWindow)

	snapshot := sessionStateRegistry.SnapshotSessionBrowserProfiles("browser-runtime-advance-profile-state-sync")
	if len(snapshot) != 1 {
		t.Fatalf("expected prepare lifecycle sync to collapse stale duplicate profile states, got %#v", snapshot)
	}
	if snapshot[0].BrowserApp != "Chromium" || snapshot[0].Status != "reconnecting" || snapshot[0].Note != "restart requested" {
		t.Fatalf("unexpected lifecycle-synced profile state: %#v", snapshot[0])
	}
	if !snapshot[0].StatusSince.Equal(base) {
		t.Fatalf("expected lifecycle sync to preserve status_since for unchanged reconnecting health, got %#v", snapshot[0])
	}
}

func TestBrowserRuntimeApplyPrepareResultPrefersAdvancedLifecycleState(t *testing.T) {
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-apply-prepare-result-advance")
	payload := &browserRuntimePayload{}

	browserRuntimeApplyPrepareResult(callCtx, payload, agentxbrowserruntime.SharedSessionBrowserObserverManager{}, nil, sessionStateRegistry, BrowserRuntimeInfo{Backend: "proxy", Target: "node"}, browserRuntimePrepareResult{
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
	})

	if payload.ProfileStatus == nil {
		t.Fatalf("expected payload profile status to be populated")
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "starting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected applyPrepareResult to surface advanced lifecycle state, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 2 {
		t.Fatalf("expected applyPrepareResult to surface synced route profiles, got %#v", payload.Profiles)
	}
	if payload.Profiles[0].Profile != "isolated" || payload.Profiles[0].Status != "starting" || !payload.Profiles[0].Running || payload.Profiles[0].Connected {
		t.Fatalf("expected applyPrepareResult profiles to use final synced lifecycle state, got %#v", payload.Profiles)
	}
	if payload.Profiles[1].Profile != "relay" || payload.Profiles[1].Status != "stopped" {
		t.Fatalf("expected applyPrepareResult to keep other synced route profiles, got %#v", payload.Profiles)
	}

	snapshot := sessionStateRegistry.SnapshotSessionBrowserProfiles("browser-runtime-apply-prepare-result-advance")
	if len(snapshot) != 2 {
		t.Fatalf("expected synced route profiles in snapshot, got %#v", snapshot)
	}
	if snapshot[0].Profile != "isolated" || snapshot[0].Status != "starting" || !snapshot[0].Running || snapshot[0].Connected || snapshot[0].Note != "start requested" {
		t.Fatalf("unexpected advanced lifecycle snapshot: %#v", snapshot[0])
	}
	if snapshot[1].Profile != "relay" || snapshot[1].Status != "stopped" {
		t.Fatalf("unexpected secondary synced route profile snapshot: %#v", snapshot[1])
	}
}

func TestBrowserRuntimeApplyPrepareResultUsesSyncedStatusSnapshotWhenDecisionDoesNotAdvance(t *testing.T) {
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-apply-prepare-result-status-sync")
	payload := &browserRuntimePayload{}

	browserRuntimeApplyPrepareResult(callCtx, payload, agentxbrowserruntime.SharedSessionBrowserObserverManager{}, nil, sessionStateRegistry, BrowserRuntimeInfo{Backend: "proxy", Target: "node"}, browserRuntimePrepareResult{
		Profile: "isolated",
		Profiles: &BrowserProfilesResult{
			Backend:        "proxy",
			DefaultProfile: "isolated",
			Profiles: []BrowserProfileInfo{
				{Profile: "isolated", BrowserApp: "Chromium", Status: "connected", Running: true, Connected: true},
				{Profile: "relay", BrowserApp: "Chromium", Status: "stopped", Running: false, Connected: false},
			},
		},
		Decision: "status_failed",
		ProfileStatus: BrowserProfileStatusResult{
			Backend:    "proxy",
			Profile:    "isolated",
			BrowserApp: "Chromium",
			Status:     "connected",
			Running:    true,
			Connected:  true,
		},
	})

	if payload.ProfileStatus == nil {
		t.Fatalf("expected payload profile status to be populated")
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "connected" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("expected applyPrepareResult to surface synced status state, got %#v", payload.ProfileStatus)
	}
	if len(payload.Profiles) != 2 {
		t.Fatalf("expected applyPrepareResult to surface synced scoped snapshot, got %#v", payload.Profiles)
	}
	if payload.Profiles[0].Profile != "isolated" || payload.Profiles[0].Status != "connected" || !payload.Profiles[0].Running || !payload.Profiles[0].Connected {
		t.Fatalf("expected scoped snapshot to keep synced status state, got %#v", payload.Profiles)
	}
	if payload.Profiles[1].Profile != "relay" || payload.Profiles[1].Status != "stopped" {
		t.Fatalf("expected other scoped profiles to remain available, got %#v", payload.Profiles)
	}
}

func TestBrowserRuntimeEnsurePreparedProfileUsesLifecycleStateForWeakStartResult(t *testing.T) {
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStatusResult: BrowserProfileStatusResult{
					Backend:   "proxy",
					Profile:   "isolated",
					Status:    "stopped",
					Running:   false,
					Connected: false,
				},
				runtimeStartResult: BrowserProfileStatusResult{
					Backend:   "proxy",
					Profile:   "isolated",
					Status:    "started",
					Running:   true,
					Connected: false,
				},
			},
		},
	}

	result, err := browserRuntimeEnsurePreparedProfile(context.Background(), nil, control, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, nil)
	if err != nil {
		t.Fatalf("browserRuntimeEnsurePreparedProfile returned error: %v", err)
	}
	if result.Decision != "started" || result.Ready {
		t.Fatalf("expected started but not-yet-ready prepare result, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "starting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "start requested" {
		t.Fatalf("expected weak start result to collapse into lifecycle-owned starting state, got %#v", result.ProfileStatus)
	}
}

func TestBrowserRuntimeEnsurePreparedProfileKeepsStartingStatusInProgress(t *testing.T) {
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStatusResult: BrowserProfileStatusResult{
					Backend:    "proxy",
					Profile:    "isolated",
					BrowserApp: "Chromium",
					Status:     "starting",
					Running:    true,
					Connected:  false,
					Note:       "start requested",
				},
			},
		},
	}

	result, err := browserRuntimeEnsurePreparedProfile(context.Background(), nil, control, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, nil)
	if err != nil {
		t.Fatalf("browserRuntimeEnsurePreparedProfile returned error: %v", err)
	}
	if len(control.runtimeStatusReqs) != 1 || len(control.runtimeStopReqs) != 0 || len(control.runtimeStartReqs) != 0 {
		t.Fatalf("expected starting prepare to stop after status observation, got status=%#v stop=%#v start=%#v", control.runtimeStatusReqs, control.runtimeStopReqs, control.runtimeStartReqs)
	}
	if result.Decision != "started" || result.Ready {
		t.Fatalf("expected starting prepare result to remain in-progress, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "starting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "start requested" {
		t.Fatalf("expected starting prepare result to preserve lifecycle-owned starting state, got %#v", result.ProfileStatus)
	}
}

func TestBrowserRuntimeEnsurePreparedProfileKeepsStartedButNotConnectedStatusInProgress(t *testing.T) {
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStatusResult: BrowserProfileStatusResult{
					Backend:    "proxy",
					Profile:    "isolated",
					BrowserApp: "Chromium",
					Status:     "started",
					Running:    true,
					Connected:  false,
				},
			},
		},
	}

	result, err := browserRuntimeEnsurePreparedProfile(context.Background(), nil, control, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, nil)
	if err != nil {
		t.Fatalf("browserRuntimeEnsurePreparedProfile returned error: %v", err)
	}
	if len(control.runtimeStatusReqs) != 1 || len(control.runtimeStopReqs) != 0 || len(control.runtimeStartReqs) != 0 {
		t.Fatalf("expected started-but-not-connected prepare to stop after status observation, got status=%#v stop=%#v start=%#v", control.runtimeStatusReqs, control.runtimeStopReqs, control.runtimeStartReqs)
	}
	if result.Decision != "started" || result.Ready {
		t.Fatalf("expected started-but-not-connected prepare result to remain in-progress, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "starting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "start requested" {
		t.Fatalf("expected started-but-not-connected prepare result to collapse into lifecycle-owned starting state, got %#v", result.ProfileStatus)
	}
}

func TestBrowserRuntimeEnsurePreparedProfileKeepsReconnectingStatusInProgress(t *testing.T) {
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStatusResult: BrowserProfileStatusResult{
					Backend:    "proxy",
					Profile:    "isolated",
					BrowserApp: "Chromium",
					Status:     "reconnecting",
					Running:    true,
					Connected:  false,
					Note:       "cdp reconnect in progress",
				},
			},
		},
	}

	result, err := browserRuntimeEnsurePreparedProfile(context.Background(), nil, control, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, nil)
	if err != nil {
		t.Fatalf("browserRuntimeEnsurePreparedProfile returned error: %v", err)
	}
	if len(control.runtimeStatusReqs) != 1 || len(control.runtimeStopReqs) != 0 || len(control.runtimeStartReqs) != 0 {
		t.Fatalf("expected reconnecting prepare to stop after status observation, got status=%#v stop=%#v start=%#v", control.runtimeStatusReqs, control.runtimeStopReqs, control.runtimeStartReqs)
	}
	if result.Decision != "restart_reconnect_in_progress" || result.Ready {
		t.Fatalf("expected reconnecting prepare result to remain in-progress, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "reconnecting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "cdp reconnect in progress" {
		t.Fatalf("expected reconnecting prepare result to preserve lifecycle-owned reconnecting state, got %#v", result.ProfileStatus)
	}
}

func TestBrowserRuntimeStartProfileUsesLifecycleStateForWeakStartResult(t *testing.T) {
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStartResult: BrowserProfileStatusResult{
					Backend:   "proxy",
					Profile:   "isolated",
					Status:    "started",
					Running:   true,
					Connected: false,
				},
			},
		},
	}

	result, err := browserRuntimeStartProfile(context.Background(), nil, control, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, nil)
	if err != nil {
		t.Fatalf("browserRuntimeStartProfile returned error: %v", err)
	}
	if result.Decision != "started" || result.Ready {
		t.Fatalf("expected started but not-yet-ready runtime start result, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "starting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "start requested" {
		t.Fatalf("expected weak start result to collapse into lifecycle-owned starting state, got %#v", result.ProfileStatus)
	}
}

func TestBrowserRuntimeEnsurePreparedProfileStatusFailurePrefersRegistryResolvedState(t *testing.T) {
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-status-failure-registry-state")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-status-failure-registry-state", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		Note:          "cdp reconnect in progress",
	})
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStatusErr: errors.New("status unavailable"),
			},
		},
	}
	binding := &browserRuntimeSessionBinding{
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "stopped",
			Running:       false,
			Connected:     false,
			Note:          "stale binding state",
		}},
	}

	result, err := browserRuntimeEnsurePreparedProfile(callCtx, sessionStateRegistry, control, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, binding)
	if err == nil {
		t.Fatalf("expected browserRuntimeEnsurePreparedProfile to return status error")
	}
	if result.Decision != "status_failed" {
		t.Fatalf("expected status_failed prepare result, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "reconnecting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "cdp reconnect in progress" {
		t.Fatalf("expected status failure to prefer registry-resolved lifecycle state, got %#v", result.ProfileStatus)
	}
}

func TestBrowserRuntimeAssessRouteProfileRecoveryPrefersRegistryScopedHealth(t *testing.T) {
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-assess-recovery-registry-health")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-assess-recovery-registry-health", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		Note:          "cdp reconnect in progress",
	})
	assessment := browserRuntimeAssessRouteProfileRecoveryWithRegistry(callCtx, sessionStateRegistry, &browserRuntimeSessionBinding{
		SessionHealthState:          "profile_stopped",
		SessionHealthReason:         "stale stopped posture",
		SessionHealthRecoveryAction: "browser action=ensure",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "stopped",
			Running:       false,
			Connected:     false,
		}},
	}, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, BrowserProfileStatusResult{})
	if assessment.Health.Summary == nil || assessment.Health.Summary.State != "profile_reconnecting" {
		t.Fatalf("expected registry-scoped health to override stale binding summary, got %#v", assessment)
	}
	if !assessment.ReconnectInProgress || !assessment.HasSyntheticStatus || assessment.NeedsRefreshRecovery {
		t.Fatalf("expected reconnect-in-progress assessment from registry-scoped health, got %#v", assessment)
	}
	if assessment.SyntheticStatus.Profile != "isolated" || assessment.SyntheticStatus.Status != "reconnecting" {
		t.Fatalf("expected reconnecting synthetic status from registry-scoped health, got %#v", assessment.SyntheticStatus)
	}
}

func TestBrowserRuntimeSessionHealthEvaluationWithRegistryPrefersScopedSnapshot(t *testing.T) {
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-session-health-registry-snapshot")
	reconnectingAt := time.Now().Add(-12 * time.Second)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-session-health-registry-snapshot", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    reconnectingAt,
		StatusSince:   reconnectingAt,
		Note:          "cdp reconnect in progress",
	})
	evaluation := browserRuntimeSessionHealthEvaluationWithRegistry(callCtx, sessionStateRegistry, &browserRuntimeSessionBinding{
		SessionHealthState:          "profile_stopped",
		SessionHealthReason:         "stale stopped posture",
		SessionHealthRecoveryAction: "browser action=ensure",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "stopped",
			Running:       false,
			Connected:     false,
		}},
		BrowserProfileCount: 1,
	}, &browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"})
	if evaluation.Summary == nil || evaluation.Summary.State != "profile_reconnecting" {
		t.Fatalf("expected registry-scoped session health to override stale binding summary, got %#v", evaluation)
	}
	if !evaluation.HasProfile || evaluation.Profile.Profile != "isolated" || evaluation.Profile.Status != "reconnecting" || evaluation.ReconnectTimedOut {
		t.Fatalf("expected registry-scoped session health to keep reconnecting profile state, got %#v", evaluation)
	}
}

func TestBrowserRuntimeSessionCoordinationPrefersRegistryScopedHealth(t *testing.T) {
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-session-coordination-registry-health")
	reconnectingAt := time.Now().Add(-12 * time.Second)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-session-coordination-registry-health", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    reconnectingAt,
		StatusSince:   reconnectingAt,
		Note:          "cdp reconnect in progress",
	})
	coordination := browserRuntimeSessionCoordination(callCtx, sessionStateRegistry, browserRuntimeSessionBinding{
		SessionHealthState:          "profile_stopped",
		SessionHealthReason:         "stale stopped posture",
		SessionHealthRecoveryAction: "browser action=ensure",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "stopped",
			Running:       false,
			Connected:     false,
		}},
		BrowserProfileCount: 1,
	}, &browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}, nil)
	if coordination == nil || coordination.State != "browser_ready" {
		t.Fatalf("expected registry-scoped coordination to see running reconnecting profile, got %#v", coordination)
	}
	if coordination.RestartBrowserAction != "" {
		t.Fatalf("expected reconnect-in-progress guardrail to suppress refresh hint, got %#v", coordination)
	}
	if browserStringSliceContains(coordination.RecommendedBrowserActions, "browser action=refresh") {
		t.Fatalf("expected coordination to suppress refresh recommendation while reconnect is in progress, got %#v", coordination)
	}
}

func TestBrowserRuntimeRestartProfileUsesLifecycleStateForWeakStartResult(t *testing.T) {
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStatusResult: BrowserProfileStatusResult{
					Backend:   "proxy",
					Profile:   "isolated",
					Status:    "connected",
					Running:   true,
					Connected: true,
				},
				runtimeStartResult: BrowserProfileStatusResult{
					Backend:   "proxy",
					Profile:   "isolated",
					Status:    "started",
					Running:   true,
					Connected: false,
				},
			},
		},
	}

	result, err := browserRuntimeRestartProfile(context.Background(), nil, control, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, nil, false)
	if err != nil {
		t.Fatalf("browserRuntimeRestartProfile returned error: %v", err)
	}
	if result.Decision != "restarted" || result.Ready {
		t.Fatalf("expected restarted but not-yet-ready restart result, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "reconnecting" || !result.ProfileStatus.Running || result.ProfileStatus.Connected || result.ProfileStatus.Note != "restart requested" {
		t.Fatalf("expected weak restart start result to collapse into lifecycle-owned reconnecting state, got %#v", result.ProfileStatus)
	}
}
