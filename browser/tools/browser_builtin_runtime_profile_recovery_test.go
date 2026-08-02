package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestBrowserRuntimeEnsurePreparedProfileDefersCooldownActiveHealthBlocker(t *testing.T) {
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStatusResult: BrowserProfileStatusResult{
					Backend:    "proxy",
					Profile:    "isolated",
					BrowserApp: "Chromium",
					Status:     "disconnected",
					Running:    false,
					Connected:  false,
					Note:       "cdp transport closed",
				},
			},
		},
	}

	result, err := browserRuntimeEnsurePreparedProfile(context.Background(), nil, control, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, &browserRuntimeSessionBinding{
		SessionHealthState:               "cooldown_active",
		SessionHealthReason:              "browser restart cooldown active for 900ms after 2 disconnects",
		SessionHealthRecoveryAction:      "browser action=wait",
		SessionHealthReconnectHint:       "retry_after_cooldown",
		SessionHealthCooldownRemainingMs: 900,
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "disconnected",
			Running:       false,
			Connected:     false,
			Note:          "cdp transport closed",
		}},
		BrowserProfileCount: 1,
	})
	if err != nil {
		t.Fatalf("browserRuntimeEnsurePreparedProfile returned error: %v", err)
	}
	if len(control.runtimeStatusReqs) != 1 || len(control.runtimeStopReqs) != 0 || len(control.runtimeStartReqs) != 0 {
		t.Fatalf("expected cooldown_active prepare to stop after status observation, got status=%#v stop=%#v start=%#v", control.runtimeStatusReqs, control.runtimeStopReqs, control.runtimeStartReqs)
	}
	if result.Decision != "cooldown_active" || result.Ready {
		t.Fatalf("expected cooldown_active prepare to defer lifecycle churn, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "disconnected" || result.ProfileStatus.Running || result.ProfileStatus.Connected {
		t.Fatalf("expected cooldown_active prepare to preserve disconnected status, got %#v", result.ProfileStatus)
	}
}

func TestBrowserRuntimeRestartProfileRequiresExplicitStartAfterPermanentRestartFailure(t *testing.T) {
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStatusResult: BrowserProfileStatusResult{
					Backend:    "proxy",
					Profile:    "isolated",
					BrowserApp: "Chromium",
					Status:     "disconnected",
					Running:    false,
					Connected:  false,
					Note:       "explicit browser.start required",
				},
			},
		},
	}

	result, err := browserRuntimeRestartProfile(context.Background(), nil, control, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, &browserRuntimeSessionBinding{
		SessionHealthState:          "restart_failed_permanent",
		SessionHealthReason:         "browser restart failed 2 times; explicit browser.start required",
		SessionHealthRecoveryAction: "browser action=start",
		SessionHealthReconnectHint:  "manual_restart_required",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "disconnected",
			Running:       false,
			Connected:     false,
			Note:          "explicit browser.start required",
		}},
		BrowserProfileCount: 1,
	}, false)
	if err != nil {
		t.Fatalf("browserRuntimeRestartProfile returned error: %v", err)
	}
	if len(control.runtimeStatusReqs) != 1 || len(control.runtimeStopReqs) != 0 || len(control.runtimeStartReqs) != 0 {
		t.Fatalf("expected restart_failed_permanent restart to stop after status observation, got status=%#v stop=%#v start=%#v", control.runtimeStatusReqs, control.runtimeStopReqs, control.runtimeStartReqs)
	}
	if result.Decision != "restart_failed_permanent" || result.Ready {
		t.Fatalf("expected restart_failed_permanent restart to require explicit start, got %#v", result)
	}
	if result.InvalidateSessionTargets {
		t.Fatalf("expected restart_failed_permanent blocker not to invalidate session targets, got %#v", result)
	}
	if result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "disconnected" || result.ProfileStatus.Running || result.ProfileStatus.Connected {
		t.Fatalf("expected restart_failed_permanent restart to preserve disconnected status, got %#v", result.ProfileStatus)
	}
}

func TestBrowserRuntimeRestartProfileDefersBackendCooldownActiveSessionHealth(t *testing.T) {
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStatusResult: BrowserProfileStatusResult{
					Backend:    "proxy",
					Profile:    "isolated",
					BrowserApp: "Chromium",
					Status:     "disconnected",
					Running:    false,
					Connected:  false,
					Note:       "cdp transport closed",
					SessionHealth: &agentxbrowserruntime.BrowserSessionHealthSummary{
						State:               "cooldown_active",
						Reason:              "browser restart cooldown active for 900ms after 2 disconnects",
						RecoveryAction:      "browser action=wait",
						ReconnectHint:       "retry_after_cooldown",
						CooldownRemainingMs: 900,
					},
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
	if len(control.runtimeStatusReqs) != 1 || len(control.runtimeStopReqs) != 0 || len(control.runtimeStartReqs) != 0 {
		t.Fatalf("expected backend cooldown_active session health to stop after status observation, got status=%#v stop=%#v start=%#v", control.runtimeStatusReqs, control.runtimeStopReqs, control.runtimeStartReqs)
	}
	if result.Decision != "cooldown_active" || result.Ready {
		t.Fatalf("expected backend cooldown_active session health to defer lifecycle churn, got %#v", result)
	}
	if result.InvalidateSessionTargets {
		t.Fatalf("expected backend cooldown_active session health not to invalidate session targets, got %#v", result)
	}
}

func TestBrowserRuntimeApplyPrepareResultBlockedRestartPreservesTrackedTargets(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-apply-prepare-result-blocked-restart")
	payload := &browserRuntimePayload{SessionID: "browser-runtime-apply-prepare-result-blocked-restart"}

	tracked := sessionRegistry.TrackTab("browser-runtime-apply-prepare-result-blocked-restart", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)

	browserRuntimeApplyPrepareResult(callCtx, payload, agentxbrowserruntime.SharedSessionBrowserObserverManager{}, sessionRegistry, sessionStateRegistry, BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, browserRuntimePrepareResult{
		Profile:  "isolated",
		Decision: "cooldown_active",
		ProfileStatus: BrowserProfileStatusResult{
			Backend:    "proxy",
			BrowserApp: "Chromium",
			Profile:    "isolated",
			Status:     "disconnected",
			Running:    false,
			Connected:  false,
			Note:       "cdp transport closed",
		},
	})

	if payload.ClearedSessionTargets != 0 {
		t.Fatalf("expected blocked restart apply surface not to report cleared session targets, got %#v", payload.ClearedSessionTargets)
	}
	if got, ok := sessionRegistry.CurrentTargetForRoute("browser-runtime-apply-prepare-result-blocked-restart", BrowserSessionRoute{Target: "node", Profile: "isolated"}); !ok || got.ID != tracked.ID {
		t.Fatalf("expected blocked restart apply surface to preserve tracked target, got %#v ok=%v", got, ok)
	}
}

func TestBrowserRuntimeEnsurePreparedProfileExplicitHealthyRawStatusUsesLifecycleIdentity(t *testing.T) {
	control := &runtimeControlBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				runtimeStatusResult: BrowserProfileStatusResult{
					Status:    "connected",
					Running:   true,
					Connected: true,
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
	if result.Decision != "already_ready" || !result.Ready {
		t.Fatalf("expected already_ready result, got %#v", result)
	}
	if result.ProfileStatus.Backend != "proxy" || result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "connected" || !result.ProfileStatus.Running || !result.ProfileStatus.Connected {
		t.Fatalf("expected explicit healthy raw status to inherit lifecycle route identity, got %#v", result.ProfileStatus)
	}
}

func TestBrowserRuntimeRestartProfileBlockedActiveNodeRunKeepsEffectiveLifecycleState(t *testing.T) {
	result, err := browserRuntimeRestartProfile(context.Background(), nil, &runtimeControlBrowserBackend{}, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, &browserRuntimeSessionBinding{
		ActiveNodeRunID: "run-91",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
		BrowserProfileCount: 1,
	}, false)
	if err != nil {
		t.Fatalf("browserRuntimeRestartProfile returned error: %v", err)
	}
	if result.Decision != "restart_blocked_active_node_run" || result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "running" || !result.ProfileStatus.Running || !result.ProfileStatus.Connected {
		t.Fatalf("expected blocked restart to preserve effective lifecycle state, got %#v", result)
	}
}

func TestBrowserRuntimeStopProfileBlockedActiveNodeRunKeepsEffectiveLifecycleState(t *testing.T) {
	result, err := browserRuntimeStopProfile(context.Background(), nil, &runtimeControlBrowserBackend{}, "isolated", BrowserRuntimeInfo{
		Backend: "proxy",
		Profile: "isolated",
		Target:  "node",
	}, &browserRuntimeSessionBinding{
		ActiveNodeRunID: "run-92",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
		BrowserProfileCount: 1,
	}, false)
	if err != nil {
		t.Fatalf("browserRuntimeStopProfile returned error: %v", err)
	}
	if result.Decision != "stop_blocked_active_node_run" || result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "running" || !result.ProfileStatus.Running || !result.ProfileStatus.Connected {
		t.Fatalf("expected blocked stop to preserve effective lifecycle state, got %#v", result)
	}
}

func TestRegisterBrowserTools_RuntimePrepareSuppressesReconnectChurnWithinWatchdog(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-reconnecting")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	reconnectingAt := time.Now().Add(-12 * time.Second)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-reconnecting", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "started",
				Running:    true,
				Connected:  false,
				Note:       "cdp reconnect in progress",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"prepare","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare reconnecting: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 0 || len(nodeBackend.runtimeStartReqs) != 0 {
		t.Fatalf("expected prepare during reconnect watchdog window to avoid lifecycle churn, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		PrepareDecision string `json:"prepare_decision"`
		PrepareReady    bool   `json:"prepare_ready"`
		ProfileStatus   struct {
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
		SessionBinding struct {
			SessionHealthState          string `json:"session_health_state"`
			SessionHealthRecoveryAction string `json:"session_health_recovery_action"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.PrepareDecision != "restart_reconnect_in_progress" || payload.PrepareReady {
		t.Fatalf("expected prepare during reconnect watchdog window to stay in-progress, got %#v", payload)
	}
	if payload.ProfileStatus.Status != "reconnecting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected prepare reconnecting status to be preserved, got %#v", payload.ProfileStatus)
	}
	if payload.SessionBinding.SessionHealthState != "profile_reconnecting" || payload.SessionBinding.SessionHealthRecoveryAction != "" {
		t.Fatalf("expected reconnecting session health to remain observed without refresh escalation, got %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimePrepareRecoversReconnectTimeoutProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-reconnect-timeout")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	staleAt := time.Now().Add(-2 * time.Minute)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-reconnect-timeout", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		ObservedAt:    staleAt,
		StatusSince:   staleAt,
		Note:          "cdp reconnect in progress",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "started",
				Running:    true,
				Connected:  false,
				Note:       "cdp reconnect in progress",
			},
			runtimeStopResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "stopped",
			},
			runtimeStartResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "started",
				Running:    true,
				Connected:  false,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"prepare","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare reconnect timeout profile: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("expected prepare recovery to stop/start reconnect-timeout profile, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		PrepareDecision string `json:"prepare_decision"`
		ProfileStatus   struct {
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
		SessionBinding struct {
			SessionHealthState          string `json:"session_health_state"`
			SessionHealthRecoveryAction string `json:"session_health_recovery_action"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.PrepareDecision != "restarted" && payload.PrepareDecision != "restart_started" {
		t.Fatalf("expected reconnect-timeout prepare recovery to reuse restart semantics, got %#v", payload)
	}
	if payload.ProfileStatus.Status != "reconnecting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected reconnect-timeout prepare recovery to advance profile status, got %#v", payload.ProfileStatus)
	}
	if payload.SessionBinding.SessionHealthState != "profile_reconnecting" {
		t.Fatalf("expected reconnect-timeout prepare recovery to transition session health to reconnecting, got %#v", payload.SessionBinding)
	}
	if payload.SessionBinding.SessionHealthRecoveryAction != "" {
		t.Fatalf("expected fresh reconnecting state to clear refresh recovery action, got %#v", payload.SessionBinding)
	}
}

func TestBrowserRuntimeTeardownProfileBlockedActiveNodeRunKeepsEffectiveLifecycleState(t *testing.T) {
	result, err := browserRuntimeTeardownProfile(context.Background(), nil, &runtimeControlBrowserBackend{}, "", BrowserRuntimeInfo{
		Backend: "proxy",
		Target:  "node",
	}, &browserRuntimeSessionBinding{
		ActiveNodeRunID:      "run-93",
		ActiveBrowserProfile: "isolated",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
		BrowserProfileCount: 1,
	})
	if err != nil {
		t.Fatalf("browserRuntimeTeardownProfile returned error: %v", err)
	}
	if result.Decision != "teardown_blocked_active_node_run" || result.Profile != "isolated" || result.ProfileStatus.Profile != "isolated" || result.ProfileStatus.Status != "running" || !result.ProfileStatus.Running || !result.ProfileStatus.Connected {
		t.Fatalf("expected blocked teardown to preserve effective lifecycle state, got %#v", result)
	}
}

func TestBrowserRuntimeRecordProfileStatusInvalidatesManagedCurrentTargetSelection(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-record-profile-status-invalidate-managed")
	sessionID := ToolSessionIDFromContext(callCtx)
	route := BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"}
	sessionRegistry.TrackTab("browser-runtime-record-profile-status-invalidate-managed", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/home",
		Title:      "Home",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	if _, ok := sessionRegistry.CurrentTargetForRoute("browser-runtime-record-profile-status-invalidate-managed", route); !ok {
		t.Fatalf("expected initial managed current target selection")
	}

	agentxbrowserruntime.SyncSharedSessionBrowserProfileStatusEvent(sessionStateRegistry, sessionRegistry, sessionID, agentxbrowserruntime.BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}, agentxbrowserruntime.BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "disconnected",
		Running:    true,
		Connected:  false,
	}, time.Time{}, browserRuntimeReconnectWatchdogWindow)

	if _, ok := sessionRegistry.CurrentTargetForRoute("browser-runtime-record-profile-status-invalidate-managed", route); ok {
		t.Fatalf("expected managed current target selection to be invalidated")
	}
	if _, ok := sessionRegistry.ResolveTabForRoute("browser-runtime-record-profile-status-invalidate-managed", route, 1); !ok {
		t.Fatalf("expected managed tracked tab to remain after selection invalidation")
	}
}

func TestBrowserRuntimeRecordProfileStatusPreservesManagedTransitionState(t *testing.T) {
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-record-profile-status-preserve-transition")
	sessionID := ToolSessionIDFromContext(callCtx)
	base := time.Now().Add(-2 * time.Minute)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-record-profile-status-preserve-transition", agentxbrowserruntime.SharedSessionBrowserProfileState{
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

	agentxbrowserruntime.SyncSharedSessionBrowserProfileStatusEvent(sessionStateRegistry, nil, sessionID, agentxbrowserruntime.BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}, agentxbrowserruntime.BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "running",
		Running:    true,
		Connected:  false,
	}, time.Time{}, browserRuntimeReconnectWatchdogWindow)

	snapshot := sessionStateRegistry.SnapshotSessionBrowserProfiles("browser-runtime-record-profile-status-preserve-transition")
	if len(snapshot) != 1 {
		t.Fatalf("expected one managed profile state, got %#v", snapshot)
	}
	if snapshot[0].Status != "reconnecting" || snapshot[0].Note != "cdp reconnect in progress" {
		t.Fatalf("expected generic status poll to preserve reconnecting transition state, got %#v", snapshot[0])
	}
	if !snapshot[0].StatusSince.Equal(base) {
		t.Fatalf("expected transition-preserving status sync to keep status_since, got %#v", snapshot[0])
	}
}

func TestBrowserRuntimeMergeCurrentProfileStatePrefersTrackedRegistryState(t *testing.T) {
	binding := &browserRuntimeSessionBinding{
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			Note:          "cdp reconnect in progress",
		}},
		BrowserProfileCount:  1,
		ActiveBrowserProfile: "isolated",
	}
	sharedEvaluation := browserRuntimeSharedBindingEvaluation(*binding, nil)

	projection := agentxbrowserruntime.ProjectSharedSessionBrowserTopLevelBinding(
		"",
		agentxbrowserruntime.BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
		nil,
		nil,
		nil,
		nil,
		&sharedEvaluation,
		&agentxbrowserruntime.SharedSessionBrowserProfileState{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "started",
			Running:       true,
			Connected:     true,
			Note:          "raw backend started",
		},
		browserRuntimeReconnectWatchdogWindow,
	)
	browserRuntimeApplySharedBindingEvaluation(binding, projection.Evaluation)

	if len(binding.BrowserProfiles) != 1 {
		t.Fatalf("expected merged binding to keep one tracked profile, got %#v", binding.BrowserProfiles)
	}
	if binding.BrowserProfiles[0].Status != "reconnecting" || !binding.BrowserProfiles[0].Running || binding.BrowserProfiles[0].Connected {
		t.Fatalf("expected binding merge to preserve tracked lifecycle state, got %#v", binding.BrowserProfiles[0])
	}
	if binding.BrowserProfiles[0].Note != "cdp reconnect in progress" {
		t.Fatalf("expected binding merge to preserve tracked transition note, got %#v", binding.BrowserProfiles[0])
	}
	if binding.SessionHealthState != "profile_reconnecting" || binding.SessionHealthRecoveryAction != "" {
		t.Fatalf("expected binding health to continue using tracked reconnecting state, got %#v", binding)
	}
}

func TestBrowserRuntimeRecordProfileStatusPreservesHostCurrentTargetSelection(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-record-profile-status-preserve-host")
	sessionID := ToolSessionIDFromContext(callCtx)
	route := BrowserSessionRoute{Backend: "system", Profile: "default", Target: "host"}
	sessionRegistry.TrackTab("browser-runtime-record-profile-status-preserve-host", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://host.example/home",
		Title:      "Home",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)

	agentxbrowserruntime.SyncSharedSessionBrowserProfileStatusEvent(sessionStateRegistry, sessionRegistry, sessionID, agentxbrowserruntime.BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}, agentxbrowserruntime.BrowserProfileStatusResult{
		Backend:    "system",
		Profile:    "default",
		BrowserApp: "Safari",
		Status:     "disconnected",
		Running:    true,
		Connected:  false,
	}, time.Time{}, browserRuntimeReconnectWatchdogWindow)

	if _, ok := sessionRegistry.CurrentTargetForRoute("browser-runtime-record-profile-status-preserve-host", route); !ok {
		t.Fatalf("expected host current target selection to remain intact")
	}
}

func TestBrowserRuntimeEvaluateSessionHealthReconnectWindow(t *testing.T) {
	reconnectingAt := time.Now().Add(-12 * time.Second)
	evaluation := browserRuntimeEvaluateSessionHealth(browserRuntimeSessionBinding{
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			StatusSince:   reconnectingAt,
			ObservedAt:    reconnectingAt,
			Note:          "cdp reconnect in progress",
		}},
		BrowserProfileCount: 1,
	})
	if evaluation.Summary == nil || evaluation.Summary.State != "profile_reconnecting" {
		t.Fatalf("expected reconnecting health evaluation, got %#v", evaluation)
	}
	if evaluation.ReconnectTimedOut {
		t.Fatalf("expected reconnecting profile inside watchdog window to avoid timeout, got %#v", evaluation)
	}

	staleAt := time.Now().Add(-2 * time.Minute)
	evaluation = browserRuntimeEvaluateSessionHealth(browserRuntimeSessionBinding{
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			StatusSince:   staleAt,
			ObservedAt:    staleAt,
			Note:          "cdp reconnect in progress",
		}},
		BrowserProfileCount: 1,
	})
	if evaluation.Summary == nil || evaluation.Summary.RecoveryAction != "browser action=refresh" || !evaluation.ReconnectTimedOut {
		t.Fatalf("expected reconnect timeout to escalate to refresh in shared evaluation, got %#v", evaluation)
	}
}

func TestBrowserRuntimeSessionHealthInputFromBindingPreservesReferenceTime(t *testing.T) {
	referenceTime := time.Date(2026, time.March, 29, 12, 0, 0, 0, time.UTC)
	input := browserRuntimeSessionHealthInputFromBinding(browserRuntimeSessionBinding{
		ActiveNodeRunID:    "run-1",
		RouteTargetCount:   2,
		SessionHealthState: "profile_reconnecting",
		ReferenceTime:      referenceTime,
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			StatusSince:   referenceTime,
			ObservedAt:    referenceTime,
		}},
		BrowserProfileCount: 1,
	})

	if !input.ReferenceTime.Equal(referenceTime) {
		t.Fatalf("expected binding health input to preserve reference time, got %v want %v", input.ReferenceTime, referenceTime)
	}
	if input.ActiveNodeRunID != "run-1" || input.RouteTargetCount != 2 || len(input.Profiles) != 1 {
		t.Fatalf("unexpected binding health input projection: %#v", input)
	}
}

func TestBrowserRuntimeSessionHealthInputFromBindingPrefersSharedEvaluation(t *testing.T) {
	referenceTime := time.Date(2026, time.March, 29, 12, 5, 0, 0, time.UTC)
	input := browserRuntimeSessionHealthInputFromBinding(browserRuntimeSessionBinding{
		ActiveNodeRunID:  "stale-run",
		RouteTargetCount: 1,
		ReferenceTime:    time.Date(2026, time.March, 29, 11, 0, 0, 0, time.UTC),
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "stale",
			RuntimeTarget: "node",
			Status:        "stopped",
		}},
		SharedEvaluation: agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
			Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
				Profiles: []agentxbrowserruntime.SharedSessionBrowserProfileState{{
					Backend:       "proxy",
					Profile:       "shared",
					RuntimeTarget: "node",
					Status:        "running",
				}},
				Summary: agentxbrowserruntime.SharedSessionBrowserBindingSummary{
					ActiveNodeRunID:  "shared-run",
					RouteTargetCount: 3,
				},
			},
			ReferenceTime: referenceTime,
			Health: agentxbrowserruntime.SharedSessionBrowserHealthEvaluation{
				Summary: &agentxbrowserruntime.SharedSessionBrowserHealthSummary{
					State:             "healthy",
					Reason:            "shared summary",
					RecoveryAction:    "none",
					ResolverBlockedBy: "multiple_candidates_filtered",
					AmbiguityClass:    "filtered_residual",
					CandidateKind:     "label",
					CandidateStrength: "medium",
					RetryDisposition:  "manual_only",
					ManualRetryHint:   "add_ordinal",
					NextStepAlias:     "snapshot",
					SpecificityFields: []string{"tag", "type"},
				},
			},
		},
		HasSharedEvaluation: true,
	})

	if input.ActiveNodeRunID != "shared-run" || input.RouteTargetCount != 3 || input.StoredState != "healthy" {
		t.Fatalf("expected shared binding evaluation to drive health input, got %#v", input)
	}
	if input.StoredResolverBlockedBy != "multiple_candidates_filtered" ||
		input.StoredAmbiguityClass != "filtered_residual" ||
		input.StoredCandidateKind != "label" ||
		input.StoredCandidateStrength != "medium" ||
		input.StoredRetryDisposition != "manual_only" ||
		input.StoredManualRetryHint != "add_ordinal" ||
		input.StoredNextStepAlias != "snapshot" ||
		!reflect.DeepEqual(input.StoredSpecificityFields, []string{"tag", "type"}) {
		t.Fatalf("expected shared binding evaluation to preserve resolver guidance, got %#v", input)
	}
	if !input.ReferenceTime.Equal(referenceTime) || len(input.Profiles) != 1 || input.Profiles[0].Profile != "shared" {
		t.Fatalf("expected shared binding evaluation to preserve source-time/profile snapshot, got %#v", input)
	}
}

func TestBrowserRuntimeExecutionAndClearRequestPreferSharedBindingEvaluation(t *testing.T) {
	referenceTime := time.Date(2026, time.March, 29, 12, 10, 0, 0, time.UTC)
	binding := &browserRuntimeSessionBinding{
		ActiveNodeRunID:      "stale-run",
		ActiveBrowserProfile: "stale-profile",
		SharedEvaluation: agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
			Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
				Profiles: []agentxbrowserruntime.SharedSessionBrowserProfileState{{
					Backend:       "proxy",
					Profile:       "shared-profile",
					RuntimeTarget: "node",
					Status:        "running",
				}},
				Summary: agentxbrowserruntime.SharedSessionBrowserBindingSummary{
					ActiveNodeRunID:      "shared-run",
					ActiveBrowserProfile: "shared-profile",
					RouteTargetCount:     2,
				},
			},
			ReferenceTime: referenceTime,
			Health: agentxbrowserruntime.SharedSessionBrowserHealthEvaluation{
				Summary: &agentxbrowserruntime.SharedSessionBrowserHealthSummary{
					State:          "healthy",
					Reason:         "shared summary",
					RecoveryAction: "none",
				},
			},
		},
		HasSharedEvaluation: true,
	}

	execution := browserRuntimeExecutionRequest(
		context.Background(),
		nil,
		"requested",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "requested", Target: "node"},
		binding,
		true,
	)
	if execution.ActiveNodeRunID != "shared-run" || execution.ActiveBrowserProfile != "shared-profile" || !execution.Force {
		t.Fatalf("expected execution request to prefer shared evaluation summary, got %#v", execution)
	}
	if !execution.HealthInput.ReferenceTime.Equal(referenceTime) || execution.HealthInput.StoredState != "healthy" {
		t.Fatalf("expected execution request to preserve shared health input, got %#v", execution.HealthInput)
	}

	clearReq := browserRuntimeClearRequest(
		context.Background(),
		nil,
		nil,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "requested", Target: "node"},
		&browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "requested", RuntimeTarget: "node"},
		binding,
		true,
	)
	if clearReq.ActiveNodeRunID != "shared-run" || !clearReq.Force {
		t.Fatalf("expected clear request to prefer shared evaluation summary, got %#v", clearReq)
	}
	if !clearReq.HealthInput.ReferenceTime.Equal(referenceTime) || clearReq.HealthInput.StoredState != "healthy" {
		t.Fatalf("expected clear request to preserve shared health input, got %#v", clearReq.HealthInput)
	}
}

func TestBrowserRuntimeExecutionAndClearRequestFromPayloadUseOverlayProfiles(t *testing.T) {
	payload := browserRuntimePayload{
		SessionBinding: &browserRuntimeSessionBinding{
			SelectedBrowserBackend:       "proxy",
			SelectedBrowserProfile:       "default",
			SelectedBrowserTarget:        "node",
			SelectedBrowserProfileSource: "select_profile",
			SharedEvaluation: agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
				Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
					Profiles: []agentxbrowserruntime.SharedSessionBrowserProfileState{{
						Backend:       "proxy",
						Profile:       "default",
						RuntimeTarget: "node",
						BrowserApp:    "Chromium",
						Status:        "stopped",
					}},
					Summary: agentxbrowserruntime.SharedSessionBrowserBindingSummary{
						ActiveNodeRunID:      "shared-run",
						ActiveBrowserProfile: "default",
						RouteTargetCount:     1,
					},
				},
				Health: agentxbrowserruntime.SharedSessionBrowserHealthEvaluation{
					Summary: &agentxbrowserruntime.SharedSessionBrowserHealthSummary{
						State:          "healthy",
						Reason:         "shared summary",
						RecoveryAction: "none",
					},
				},
			},
			HasSharedEvaluation: true,
		},
		SessionProfileSelection: &browserRuntimeSessionProfileSelection{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "select_profile",
		},
		Profiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
	}

	execution := browserRuntimeExecutionRequestFromPayload(
		context.Background(),
		nil,
		"requested",
		BrowserRuntimeInfo{Backend: "proxy", Profile: "requested", Target: "node"},
		payload,
		true,
	)
	if execution.ActiveNodeRunID != "shared-run" || execution.ActiveBrowserProfile != "default" || !execution.Force {
		t.Fatalf("expected payload-aware execution request to preserve shared summary, got %#v", execution)
	}
	if len(execution.HealthInput.Profiles) != 2 ||
		execution.HealthInput.Profiles[0].Profile != "isolated" ||
		execution.HealthInput.Profiles[1].Profile != "default" {
		t.Fatalf("expected payload-aware execution request to prepend payload overlay profiles, got %#v", execution.HealthInput)
	}

	clearReq := browserRuntimeClearRequestFromPayload(
		context.Background(),
		nil,
		nil,
		BrowserRuntimeInfo{Backend: "proxy", Profile: "requested", Target: "node"},
		&browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "requested", RuntimeTarget: "node"},
		payload,
		true,
	)
	if clearReq.ActiveNodeRunID != "shared-run" || !clearReq.Force {
		t.Fatalf("expected payload-aware clear request to preserve shared summary, got %#v", clearReq)
	}
	if len(clearReq.HealthInput.Profiles) != 2 ||
		clearReq.HealthInput.Profiles[0].Profile != "isolated" ||
		clearReq.HealthInput.Profiles[1].Profile != "default" {
		t.Fatalf("expected payload-aware clear request to prepend payload overlay profiles, got %#v", clearReq.HealthInput)
	}
}

func TestBrowserRuntimeSharedBindingEvaluationPtrFromPayloadAppliesPayloadOverlays(t *testing.T) {
	evaluation := browserRuntimeSharedBindingEvaluationPtrFromPayload(browserRuntimePayload{
		SessionBinding: &browserRuntimeSessionBinding{
			SharedEvaluation: agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
				Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
					Profiles: []agentxbrowserruntime.SharedSessionBrowserProfileState{{
						Backend:       "proxy",
						Profile:       "default",
						RuntimeTarget: "node",
						Status:        "stopped",
					}},
				},
			},
			HasSharedEvaluation: true,
		},
		SessionProfileSelection: &browserRuntimeSessionProfileSelection{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "select_profile",
		},
		Profiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
	})
	if evaluation == nil {
		t.Fatalf("expected payload-aware binding evaluation")
	}
	if evaluation.Snapshot.SelectedProfileSelection == nil || evaluation.Snapshot.SelectedProfileSelection.Profile != "isolated" {
		t.Fatalf("expected payload-aware binding evaluation to project payload selection, got %#v", evaluation.Snapshot.SelectedProfileSelection)
	}
	if len(evaluation.Snapshot.Profiles) != 2 ||
		evaluation.Snapshot.Profiles[0].Profile != "isolated" ||
		evaluation.Snapshot.Profiles[1].Profile != "default" {
		t.Fatalf("expected payload-aware binding evaluation to prepend overlay profiles, got %#v", evaluation.Snapshot.Profiles)
	}
}

func TestBrowserRuntimeSharedWorkbenchProjectionRequestUsesPayloadAwareEvaluation(t *testing.T) {
	payload := browserRuntimePayload{
		SessionBinding: &browserRuntimeSessionBinding{
			SharedEvaluation: agentxbrowserruntime.SharedSessionBrowserBindingEvaluation{
				Snapshot: agentxbrowserruntime.SharedSessionBrowserBindingSnapshot{
					Profiles: []agentxbrowserruntime.SharedSessionBrowserProfileState{{
						Backend:       "proxy",
						Profile:       "default",
						RuntimeTarget: "node",
						Status:        "stopped",
					}},
				},
				Coordination: agentxbrowserruntime.SharedSessionBrowserCoordinationEvaluation{
					Plan: agentxbrowserruntime.SharedSessionBrowserCoordinationPlan{
						State:                "browser_ready",
						NeedsSessionSync:     true,
						PrimaryBrowserAction: "browser action=tabs",
						NextStep:             "browser action=tabs",
					},
				},
			},
			HasSharedEvaluation: true,
		},
		SessionProfileSelection: &browserRuntimeSessionProfileSelection{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Source:        "select_profile",
		},
		Profiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "running",
			Running:       true,
			Connected:     true,
		}},
	}

	request := browserRuntimeSharedWorkbenchProjectionRequest(&payload, browserRuntimeWorkbenchProjectionSync{
		SyncCoordinationSurface: true,
	})
	if request.Evaluation == nil {
		t.Fatalf("expected workbench projection request to keep payload-aware evaluation")
	}
	if request.Evaluation.Snapshot.SelectedProfileSelection == nil || request.Evaluation.Snapshot.SelectedProfileSelection.Profile != "isolated" {
		t.Fatalf("expected workbench projection request to preserve payload-aware selection, got %#v", request.Evaluation.Snapshot.SelectedProfileSelection)
	}
	if len(request.Evaluation.Snapshot.Profiles) != 2 ||
		request.Evaluation.Snapshot.Profiles[0].Profile != "isolated" ||
		request.Evaluation.Snapshot.Profiles[1].Profile != "default" {
		t.Fatalf("expected workbench projection request to preserve payload-aware profile inventory, got %#v", request.Evaluation.Snapshot.Profiles)
	}
	if request.CoordinationPlan == nil ||
		request.CoordinationPlan.State != "browser_ready" ||
		!request.CoordinationPlan.NeedsSessionSync ||
		request.CoordinationPlan.PrimaryBrowserAction != "browser action=tabs" ||
		request.CoordinationPlan.NextStep != "browser action=tabs" {
		t.Fatalf("expected workbench projection request to prefer payload-aware coordination plan, got %#v", request.CoordinationPlan)
	}
}

func TestBrowserRuntimeSessionHealthEvaluationFromBindingPrefersStoredSummary(t *testing.T) {
	reconnectingAt := time.Now().Add(-12 * time.Second)
	evaluation := browserRuntimeSessionHealthEvaluationFromBinding(browserRuntimeSessionBinding{
		SessionHealthState:                     "profile_reconnecting",
		SessionHealthReason:                    "stored reconnecting posture",
		SessionHealthRecoveryAction:            "",
		SessionHealthResolverBlockedBy:         "multiple_candidates_filtered",
		SessionHealthResolverAmbiguityClass:    "filtered_residual",
		SessionHealthResolverCandidateKind:     "label",
		SessionHealthResolverStrength:          "medium",
		SessionHealthResolverRetryDisposition:  "manual_only",
		SessionHealthResolverManualRetryHint:   "add_ordinal",
		SessionHealthResolverNextStepAlias:     "snapshot",
		SessionHealthResolverSpecificityFields: []string{"tag", "type"},
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			StatusSince:   reconnectingAt,
			ObservedAt:    reconnectingAt,
			Note:          "cdp reconnect in progress",
		}},
		BrowserProfileCount: 1,
	})
	if evaluation.Summary == nil || evaluation.Summary.State != "profile_reconnecting" || evaluation.Summary.Reason != "stored reconnecting posture" {
		t.Fatalf("expected binding evaluation to reuse stored summary, got %#v", evaluation)
	}
	if evaluation.Summary.ResolverBlockedBy != "multiple_candidates_filtered" ||
		evaluation.Summary.AmbiguityClass != "filtered_residual" ||
		evaluation.Summary.CandidateKind != "label" ||
		evaluation.Summary.CandidateStrength != "medium" ||
		evaluation.Summary.RetryDisposition != "manual_only" ||
		evaluation.Summary.ManualRetryHint != "add_ordinal" ||
		evaluation.Summary.NextStepAlias != "snapshot" ||
		!reflect.DeepEqual(evaluation.Summary.SpecificityFields, []string{"tag", "type"}) {
		t.Fatalf("expected binding evaluation to preserve stored resolver guidance, got %#v", evaluation)
	}
	if !evaluation.HasProfile || evaluation.Profile.Status != "reconnecting" || evaluation.ReconnectTimedOut {
		t.Fatalf("expected binding evaluation to still surface tracked reconnecting profile, got %#v", evaluation)
	}
}

func TestBrowserRuntimeAssessRouteProfileRecoveryReconnectInProgress(t *testing.T) {
	reconnectingAt := time.Now().Add(-12 * time.Second)
	assessment := browserRuntimeAssessRouteProfileRecovery(&browserRuntimeSessionBinding{
		BrowserProfiles: []browserRuntimeProfileState{
			{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "reconnecting",
				Running:       true,
				Connected:     false,
				StatusSince:   reconnectingAt,
				ObservedAt:    reconnectingAt,
				Note:          "cdp reconnect in progress",
			},
			{
				Backend:       "proxy",
				Profile:       "relay",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "disconnected",
				Running:       true,
				Connected:     false,
			},
		},
		BrowserProfileCount: 2,
	}, "isolated", BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}, BrowserProfileStatusResult{})
	if !assessment.ReconnectInProgress || !assessment.HasSyntheticStatus {
		t.Fatalf("expected route-scoped recovery assessment to preserve reconnect_in_progress, got %#v", assessment)
	}
	if assessment.NeedsRefreshRecovery {
		t.Fatalf("expected reconnecting profile inside watchdog window to avoid refresh escalation, got %#v", assessment)
	}
	if assessment.SyntheticStatus.Status != "reconnecting" || !assessment.SyntheticStatus.Running || assessment.SyntheticStatus.Connected {
		t.Fatalf("expected synthetic status to mirror reconnecting health profile, got %#v", assessment.SyntheticStatus)
	}
}

func TestBrowserRuntimeAssessRouteProfileRecoveryPrefersScopedStoredSummary(t *testing.T) {
	reconnectingAt := time.Now().Add(-12 * time.Second)
	assessment := browserRuntimeAssessRouteProfileRecovery(&browserRuntimeSessionBinding{
		SessionHealthState:          "profile_reconnecting",
		SessionHealthReason:         "stored scoped reconnecting posture",
		SessionHealthRecoveryAction: "",
		BrowserProfiles: []browserRuntimeProfileState{
			{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "reconnecting",
				Running:       true,
				Connected:     false,
				StatusSince:   reconnectingAt,
				ObservedAt:    reconnectingAt,
			},
		},
		BrowserProfileCount: 1,
	}, "isolated", BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}, BrowserProfileStatusResult{})
	if assessment.Health.Summary == nil || assessment.Health.Summary.Reason != "stored scoped reconnecting posture" {
		t.Fatalf("expected route-scoped recovery to reuse stored summary, got %#v", assessment)
	}
	if !assessment.ReconnectInProgress || !assessment.HasSyntheticStatus {
		t.Fatalf("expected route-scoped recovery to preserve reconnecting transition, got %#v", assessment)
	}
}

func TestBrowserRuntimeAssessRouteProfileRecoveryIgnoresUnrelatedStoredSummary(t *testing.T) {
	assessment := browserRuntimeAssessRouteProfileRecovery(&browserRuntimeSessionBinding{
		SessionHealthState:          "profile_disconnected",
		SessionHealthReason:         "stored unrelated disconnected posture",
		SessionHealthRecoveryAction: "browser action=refresh",
		BrowserProfiles: []browserRuntimeProfileState{
			{
				Backend:       "proxy",
				Profile:       "isolated",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "running",
				Running:       true,
				Connected:     true,
			},
			{
				Backend:       "proxy",
				Profile:       "relay",
				RuntimeTarget: "node",
				BrowserApp:    "Chromium",
				Status:        "disconnected",
				Running:       true,
				Connected:     false,
			},
		},
		BrowserProfileCount: 2,
	}, "isolated", BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}, BrowserProfileStatusResult{})
	if assessment.Health.Summary == nil || assessment.Health.Summary.State != "healthy" {
		t.Fatalf("expected route-scoped recovery to ignore unrelated stored summary, got %#v", assessment)
	}
	if assessment.NeedsRefreshRecovery || assessment.ReconnectInProgress {
		t.Fatalf("expected healthy scoped route not to inherit unrelated recovery posture, got %#v", assessment)
	}
}

func TestBrowserRuntimeAssessRouteProfileRecoveryTimedOutStoredSummaryWithoutRawStatus(t *testing.T) {
	staleAt := time.Now().Add(-2 * time.Minute)
	assessment := browserRuntimeAssessRouteProfileRecovery(&browserRuntimeSessionBinding{
		SessionHealthState:          "profile_reconnecting",
		SessionHealthReason:         "stored timed-out reconnecting posture",
		SessionHealthRecoveryAction: "browser action=refresh",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			StatusSince:   staleAt,
			ObservedAt:    staleAt,
			Note:          "cdp reconnect in progress",
		}},
		BrowserProfileCount: 1,
	}, "isolated", BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}, BrowserProfileStatusResult{})
	if !assessment.NeedsRefreshRecovery || !assessment.ShouldStopBeforeRecovery {
		t.Fatalf("expected timed-out stored summary to drive refresh+stop without raw status, got %#v", assessment)
	}
	if assessment.ReconnectInProgress {
		t.Fatalf("expected timed-out stored summary not to suppress restart, got %#v", assessment)
	}
}

func TestBrowserRuntimeAssessRouteProfileRecoveryDisconnectedStoredSummaryWithoutRawStatus(t *testing.T) {
	assessment := browserRuntimeAssessRouteProfileRecovery(&browserRuntimeSessionBinding{
		SessionHealthState:          "profile_disconnected",
		SessionHealthReason:         "stored disconnected posture",
		SessionHealthRecoveryAction: "browser action=refresh",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "disconnected",
			Running:       true,
			Connected:     false,
			Note:          "cdp transport closed",
		}},
		BrowserProfileCount: 1,
	}, "isolated", BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}, BrowserProfileStatusResult{})
	if !assessment.NeedsRefreshRecovery || !assessment.ShouldStopBeforeRecovery {
		t.Fatalf("expected disconnected stored summary to drive refresh+stop without raw status, got %#v", assessment)
	}
	if assessment.ReconnectInProgress {
		t.Fatalf("expected disconnected stored summary not to suppress restart, got %#v", assessment)
	}
	if assessment.EffectiveStatus.Status != "disconnected" || assessment.EffectiveStatus.Profile != "isolated" {
		t.Fatalf("expected effective status to come from shared recovery assessment, got %#v", assessment.EffectiveStatus)
	}
}

func TestBrowserRuntimeAssessRouteProfileRecoveryExplicitHealthyRawStatusDoesNotSuppressStoredRecovery(t *testing.T) {
	assessment := browserRuntimeAssessRouteProfileRecovery(&browserRuntimeSessionBinding{
		SessionHealthState:          "profile_disconnected",
		SessionHealthReason:         "stored disconnected posture",
		SessionHealthRecoveryAction: "browser action=refresh",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "disconnected",
			Running:       true,
			Connected:     false,
			Note:          "cdp transport closed",
		}},
		BrowserProfileCount: 1,
	}, "isolated", BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}, BrowserProfileStatusResult{
		Backend:    "proxy",
		BrowserApp: "Chromium",
		Profile:    "isolated",
		Status:     "connected",
		Running:    true,
		Connected:  true,
	})
	if !assessment.NeedsRefreshRecovery || !assessment.ShouldStopBeforeRecovery {
		t.Fatalf("expected shared assessment to preserve stored recovery posture, got %#v", assessment)
	}
	if assessment.ReconnectInProgress {
		t.Fatalf("expected disconnected stored summary not to become reconnect-in-progress, got %#v", assessment)
	}
}

func TestBrowserRuntimeAssessRouteProfileRecoveryDisconnectedStoredSummaryIgnoresWeakRawStatus(t *testing.T) {
	assessment := browserRuntimeAssessRouteProfileRecovery(&browserRuntimeSessionBinding{
		SessionHealthState:          "profile_disconnected",
		SessionHealthReason:         "stored disconnected posture",
		SessionHealthRecoveryAction: "browser action=refresh",
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "disconnected",
			Running:       true,
			Connected:     false,
			Note:          "cdp transport closed",
		}},
		BrowserProfileCount: 1,
	}, "isolated", BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}, BrowserProfileStatusResult{
		Backend:    "proxy",
		BrowserApp: "Chromium",
		Profile:    "isolated",
	})
	if !assessment.NeedsRefreshRecovery || !assessment.ShouldStopBeforeRecovery {
		t.Fatalf("expected weak raw status not to veto disconnected stored summary recovery, got %#v", assessment)
	}
	if assessment.ReconnectInProgress {
		t.Fatalf("expected disconnected stored summary not to suppress restart, got %#v", assessment)
	}
}

func TestBrowserRuntimeAssessRouteProfileRecoveryEscalatesRefreshAndStop(t *testing.T) {
	staleAt := time.Now().Add(-2 * time.Minute)
	assessment := browserRuntimeAssessRouteProfileRecovery(&browserRuntimeSessionBinding{
		BrowserProfiles: []browserRuntimeProfileState{{
			Backend:       "proxy",
			Profile:       "isolated",
			RuntimeTarget: "node",
			BrowserApp:    "Chromium",
			Status:        "reconnecting",
			Running:       true,
			Connected:     false,
			StatusSince:   staleAt,
			ObservedAt:    staleAt,
			Note:          "cdp reconnect in progress",
		}},
		BrowserProfileCount: 1,
	}, "isolated", BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}, BrowserProfileStatusResult{
		Backend:    "proxy",
		Profile:    "isolated",
		BrowserApp: "Chromium",
		Status:     "started",
		Running:    true,
		Connected:  false,
	})
	if !assessment.NeedsRefreshRecovery || !assessment.ShouldStopBeforeRecovery {
		t.Fatalf("expected reconnect-timeout recovery assessment to escalate refresh+stop, got %#v", assessment)
	}
	if assessment.ReconnectInProgress {
		t.Fatalf("expected timed-out reconnecting profile to stop suppressing restart, got %#v", assessment)
	}
}
