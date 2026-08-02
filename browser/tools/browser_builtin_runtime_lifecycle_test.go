package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_RuntimeClearSessionResetsManagedRouteState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "system-extract",
			BrowserApp: "Safari",
			Title:      "Host",
			Content:    "Host content",
			FinalURL:   "https://host.example/home",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			extractResult: BrowserExtractResult{
				Backend:    "proxy-extract",
				BrowserApp: "Chromium",
				Title:      "Workbench",
				Content:    "Node content",
				FinalURL:   "https://node.example/workbench",
			},
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-session")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	}); err != nil {
		t.Fatalf("browser_runtime status before clear_session: %v", err)
	}
	first := sessionRegistry.TrackTab("browser-runtime-clear-session", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	second := sessionRegistry.TrackTab("browser-runtime-clear-session", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","target":"target:%s"}`, second.ID),
	}); err != nil {
		t.Fatalf("browser_runtime select_target: %v", err)
	}
	if got, ok := sessionRegistry.ResolveTarget("browser-runtime-clear-session", first.ID); !ok || got.TabIndex != 1 {
		t.Fatalf("expected first tracked target to exist before clear_session, got %#v ok=%v", got, ok)
	}

	clearOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_session"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_session: %v", err)
	}
	var clearPayload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		ClearSessionDecision    string `json:"clear_session_decision"`
		ClearSessionReady       bool   `json:"clear_session_ready"`
		ClearedSessionProfiles  int    `json:"cleared_session_profiles"`
		ClearedSessionTargets   int    `json:"cleared_session_targets"`
		SessionProfileSelection any    `json:"session_profile_selection"`
		SessionTargetSelection  any    `json:"session_target_selection"`
		SessionBinding          struct {
			RouteTargetCount        int    `json:"route_target_count"`
			BrowserProfileCount     int    `json:"browser_profile_count"`
			SelectedBrowserProfile  string `json:"selected_browser_profile"`
			SelectedBrowserTargetID string `json:"selected_browser_target_id"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(clearOut), &clearPayload); err != nil {
		t.Fatalf("decode clear_session output: %v", err)
	}
	if clearPayload.Action != "clear_session" || clearPayload.Status != "ok" || clearPayload.ClearSessionDecision != "session_route_cleared" || !clearPayload.ClearSessionReady {
		t.Fatalf("unexpected clear_session payload: %#v", clearPayload)
	}
	if clearPayload.ClearedSessionProfiles != 1 || clearPayload.ClearedSessionTargets != 2 {
		t.Fatalf("expected clear_session to clear 1 profile state and 2 tracked targets, got %#v", clearPayload)
	}
	if clearPayload.SessionProfileSelection != nil || clearPayload.SessionTargetSelection != nil {
		t.Fatalf("expected clear_session to clear session selections, got profile=%#v target=%#v", clearPayload.SessionProfileSelection, clearPayload.SessionTargetSelection)
	}
	if clearPayload.SessionBinding.RouteTargetCount != 0 || clearPayload.SessionBinding.BrowserProfileCount != 0 || clearPayload.SessionBinding.SelectedBrowserProfile != "" || clearPayload.SessionBinding.SelectedBrowserTargetID != "" {
		t.Fatalf("expected clear_session to reset session binding, got %#v", clearPayload.SessionBinding)
	}
	if _, ok := sessionRegistry.ResolveTarget("browser-runtime-clear-session", second.ID); ok {
		t.Fatalf("expected tracked node target to be removed by clear_session")
	}

	statusOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status after clear_session: %v", err)
	}
	var statusPayload struct {
		SelectedRoute struct {
			Backend       string `json:"backend"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionProfileSelection any `json:"session_profile_selection"`
		SessionTargetSelection  any `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(statusOut), &statusPayload); err != nil {
		t.Fatalf("decode status output after clear_session: %v", err)
	}
	if statusPayload.SelectedRoute.Backend != "proxy" || statusPayload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected clear_session to fall back to promoted node route, got %#v", statusPayload.SelectedRoute)
	}
	if statusPayload.SessionProfileSelection != nil || statusPayload.SessionTargetSelection != nil {
		t.Fatalf("expected status after clear_session to have no session selections, got %#v %#v", statusPayload.SessionProfileSelection, statusPayload.SessionTargetSelection)
	}

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	}); err != nil {
		t.Fatalf("browser_act extract after clear_session: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused after clear_session, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 0 {
		t.Fatalf("expected promoted node extract after clear_session, got %#v", nodeBackend.extractReqs)
	}
}

func TestRegisterBrowserTools_RuntimeClearSessionBlockedByActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-clear-session-blocked", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-71",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionRunRegistry:   sessionRunRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-session-blocked")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-clear-session-blocked", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	sessionRegistry.TrackTab("browser-runtime-clear-session-blocked", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_session","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_session blocked: %v", err)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		ClearSessionDecision string `json:"clear_session_decision"`
		ClearSessionReady    bool   `json:"clear_session_ready"`
		Force                bool   `json:"force"`
		SessionBinding       struct {
			ActiveNodeRunID string `json:"active_node_run_id"`
		} `json:"session_binding"`
		ProfileStatus struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode clear_session blocked output: %v", err)
	}
	if payload.Action != "clear_session" || payload.Status != "ok" || payload.ClearSessionDecision != "clear_session_blocked_active_node_run" || payload.ClearSessionReady || payload.Force {
		t.Fatalf("unexpected blocked clear_session payload: %#v", payload)
	}
	if payload.SessionBinding.ActiveNodeRunID != "run-71" {
		t.Fatalf("expected active node run to remain visible in blocked clear_session payload, got %#v", payload)
	}
	if payload.ProfileStatus.Backend != "proxy" || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.RuntimeTarget != "node" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("expected blocked clear_session to preserve effective lifecycle state, got %#v", payload.ProfileStatus)
	}
	if got, ok := sessionRegistry.CurrentTargetForRoute("browser-runtime-clear-session-blocked", BrowserSessionRoute{Target: "node", Profile: "workbench"}); !ok || got.TabIndex != 1 {
		t.Fatalf("expected blocked clear_session not to clear tracked target, got %#v ok=%v", got, ok)
	}
}

func TestRegisterBrowserTools_RuntimeClearSessionForceBypassesActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-clear-session-force", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-72",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionRunRegistry:   sessionRunRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-session-force")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-clear-session-force", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	sessionRegistry.TrackTab("browser-runtime-clear-session-force", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_session","runtime_target":"node","profile":"workbench","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_session force: %v", err)
	}
	var payload struct {
		Action                string `json:"action"`
		Status                string `json:"status"`
		ClearSessionDecision  string `json:"clear_session_decision"`
		ClearSessionReady     bool   `json:"clear_session_ready"`
		Force                 bool   `json:"force"`
		ClearedSessionTargets int    `json:"cleared_session_targets"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode clear_session force output: %v", err)
	}
	if payload.Action != "clear_session" || payload.Status != "ok" || payload.ClearSessionDecision != "session_route_cleared" || !payload.ClearSessionReady || !payload.Force || payload.ClearedSessionTargets != 1 {
		t.Fatalf("unexpected forced clear_session payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeRestartClearsImplicitTargetReuse(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			extractResult: BrowserExtractResult{
				Backend:    "proxy-extract",
				BrowserApp: "Chromium",
				Title:      "Workbench",
				Content:    "Extracted content",
				FinalURL:   "https://node.example/workbench",
			},
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "stopped",
			},
			runtimeStartResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "started",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-restart-clears-target")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	tracked := sessionRegistry.TrackTab("browser-runtime-restart-clears-target", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","target":"target:%s"}`, tracked.ID),
	}); err != nil {
		t.Fatalf("browser_runtime select_target: %v", err)
	}

	restartOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"restart","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart: %v", err)
	}
	var restartPayload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		RestartDecision        string `json:"restart_decision"`
		RestartReady           bool   `json:"restart_ready"`
		ClearedSessionTargets  int    `json:"cleared_session_targets"`
		SessionTargetSelection any    `json:"session_target_selection"`
		SessionBinding         struct {
			SessionHealthState               string `json:"session_health_state"`
			SessionHealthReason              string `json:"session_health_reason"`
			SessionHealthRecoveryAction      string `json:"session_health_recovery_action"`
			SessionHealthReconnectHint       string `json:"session_health_reconnect_hint"`
			SessionHealthCooldownRemainingMs int    `json:"session_health_cooldown_remaining_ms"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(restartOut), &restartPayload); err != nil {
		t.Fatalf("decode restart output: %v", err)
	}
	if restartPayload.Action != "restart" || restartPayload.Status != "ok" || restartPayload.RestartDecision != "restart_started" || !restartPayload.RestartReady {
		t.Fatalf("unexpected restart payload: %#v", restartPayload)
	}
	if restartPayload.ClearedSessionTargets != 1 {
		t.Fatalf("expected restart to clear 1 tracked session target, got %#v", restartPayload)
	}
	if restartPayload.SessionTargetSelection != nil {
		t.Fatalf("expected restart to clear session target selection, got %#v", restartPayload.SessionTargetSelection)
	}

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	}); err != nil {
		t.Fatalf("browser_act extract after restart: %v", err)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 0 {
		t.Fatalf("expected node backend extract to stop reusing selected target after restart, got %#v", nodeBackend.extractReqs)
	}
}

func TestRegisterBrowserTools_RuntimeRestartPreservesImplicitTargetReuseWhenBackendCooldownActive(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			extractResult: BrowserExtractResult{
				Backend:    "proxy-extract",
				BrowserApp: "Chromium",
				Title:      "Workbench",
				Content:    "Extracted content",
				FinalURL:   "https://node.example/workbench",
			},
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
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
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-restart-preserves-target-cooldown")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	tracked := sessionRegistry.TrackTab("browser-runtime-restart-preserves-target-cooldown", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","target":"target:%s"}`, tracked.ID),
	}); err != nil {
		t.Fatalf("browser_runtime select_target: %v", err)
	}

	restartOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"restart","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart: %v", err)
	}
	var restartPayload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		RestartDecision        string `json:"restart_decision"`
		RestartReady           bool   `json:"restart_ready"`
		ClearedSessionTargets  int    `json:"cleared_session_targets"`
		SessionTargetSelection any    `json:"session_target_selection"`
		SessionBinding         struct {
			SessionHealthState               string `json:"session_health_state"`
			SessionHealthReason              string `json:"session_health_reason"`
			SessionHealthRecoveryAction      string `json:"session_health_recovery_action"`
			SessionHealthReconnectHint       string `json:"session_health_reconnect_hint"`
			SessionHealthCooldownRemainingMs int    `json:"session_health_cooldown_remaining_ms"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(restartOut), &restartPayload); err != nil {
		t.Fatalf("decode restart output: %v", err)
	}
	if restartPayload.Action != "restart" || restartPayload.Status != "ok" || restartPayload.RestartDecision != "cooldown_active" || restartPayload.RestartReady {
		t.Fatalf("unexpected cooldown_active restart payload: %#v", restartPayload)
	}
	if restartPayload.ClearedSessionTargets != 0 {
		t.Fatalf("expected cooldown_active restart not to clear tracked session targets, got %#v", restartPayload)
	}
	if restartPayload.SessionTargetSelection == nil {
		t.Fatalf("expected cooldown_active restart to preserve session target selection, got %#v", restartPayload.SessionTargetSelection)
	}
	if restartPayload.SessionBinding.SessionHealthState != "cooldown_active" ||
		!strings.Contains(restartPayload.SessionBinding.SessionHealthReason, "cooldown active") ||
		restartPayload.SessionBinding.SessionHealthRecoveryAction != "browser action=wait" ||
		restartPayload.SessionBinding.SessionHealthReconnectHint != "retry_after_cooldown" ||
		restartPayload.SessionBinding.SessionHealthCooldownRemainingMs != 900 {
		t.Fatalf("expected cooldown_active restart to publish backend session health blocker in final payload, got %#v", restartPayload.SessionBinding)
	}

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	}); err != nil {
		t.Fatalf("browser_act extract after cooldown_active restart: %v", err)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 2 {
		t.Fatalf("expected node backend extract to keep reusing selected target after cooldown_active restart, got %#v", nodeBackend.extractReqs)
	}
}

func TestRegisterBrowserTools_RuntimeRestartTransitionsSessionHealthToReconnecting(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-restart-reconnecting")
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
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

	restartOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"restart","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart reconnecting: %v", err)
	}
	var restartPayload struct {
		RestartDecision string `json:"restart_decision"`
		RestartReady    bool   `json:"restart_ready"`
		ProfileStatus   struct {
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(restartOut), &restartPayload); err != nil {
		t.Fatalf("decode restart output: %v", err)
	}
	if restartPayload.RestartDecision != "restart_started" || restartPayload.RestartReady {
		t.Fatalf("expected restart_started but not yet ready, got %#v", restartPayload)
	}
	if restartPayload.ProfileStatus.Status != "reconnecting" || !restartPayload.ProfileStatus.Running || restartPayload.ProfileStatus.Connected {
		t.Fatalf("expected restart_started payload to expose lifecycle-owned reconnecting state, got %#v", restartPayload.ProfileStatus)
	}

	sessionsOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions after reconnecting restart: %v", err)
	}
	var sessionsPayload struct {
		SessionBinding struct {
			SessionHealthState          string `json:"session_health_state"`
			SessionHealthRecoveryAction string `json:"session_health_recovery_action"`
			Coordination                struct {
				RestartBrowserAction      string   `json:"restart_browser_action"`
				RecommendedBrowserActions []string `json:"recommended_browser_actions"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(sessionsOut), &sessionsPayload); err != nil {
		t.Fatalf("decode sessions output: %v", err)
	}
	if sessionsPayload.SessionBinding.SessionHealthState != "profile_reconnecting" || sessionsPayload.SessionBinding.SessionHealthRecoveryAction != "" {
		t.Fatalf("expected reconnecting health posture, got %#v", sessionsPayload.SessionBinding)
	}
	if sessionsPayload.SessionBinding.Coordination.RestartBrowserAction != "" {
		t.Fatalf("expected reconnecting state to suppress extra refresh hint, got %#v", sessionsPayload.SessionBinding.Coordination)
	}
	if browserStringSliceContains(sessionsPayload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=refresh") {
		t.Fatalf("expected reconnecting state to suppress refresh from recommended actions, got %#v", sessionsPayload.SessionBinding.Coordination)
	}
}

func TestRegisterBrowserTools_RuntimeRestartRecoversDisconnectedProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-restart-disconnected")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-restart-disconnected", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       true,
		Connected:     false,
		Note:          "cdp transport closed",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "disconnected",
				Running:    true,
				Connected:  false,
				Note:       "cdp transport closed",
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

	restartOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"restart","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart disconnected: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("expected disconnected restart to stop/start profile, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var restartPayload struct {
		RestartDecision string `json:"restart_decision"`
		ProfileStatus   struct {
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(restartOut), &restartPayload); err != nil {
		t.Fatalf("decode restart output: %v", err)
	}
	if restartPayload.RestartDecision != "restarted" {
		t.Fatalf("expected disconnected restart to go through stop+start, got %#v", restartPayload)
	}
	if restartPayload.ProfileStatus.Status != "reconnecting" || !restartPayload.ProfileStatus.Running || restartPayload.ProfileStatus.Connected {
		t.Fatalf("expected disconnected restart payload to reflect synced reconnecting state, got %#v", restartPayload.ProfileStatus)
	}
}

func TestBrowserRuntimeAssessRouteProfileRecoveryRawStoppedVetoesStoredDisconnectStop(t *testing.T) {
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
		Status:     "stopped",
	})
	if !assessment.NeedsRefreshRecovery {
		t.Fatalf("expected stored disconnected summary to keep refresh recovery, got %#v", assessment)
	}
	if assessment.ShouldStopBeforeRecovery {
		t.Fatalf("expected explicit raw stopped status to veto extra stop-before-recovery, got %#v", assessment)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsReconnectTimeoutEscalatesToRefresh(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-reconnect-timeout")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	staleAt := time.Now().Add(-2 * time.Minute)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-reconnect-timeout", agentxbrowserruntime.SharedSessionBrowserProfileState{
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

	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}},
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions reconnect timeout: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			SessionHealthState          string `json:"session_health_state"`
			SessionHealthReason         string `json:"session_health_reason"`
			SessionHealthRecoveryAction string `json:"session_health_recovery_action"`
			Coordination                struct {
				PrimaryBrowserAction      string   `json:"primary_browser_action"`
				NextStep                  string   `json:"next_step"`
				RecommendedBrowserActions []string `json:"recommended_browser_actions"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.SessionBinding.SessionHealthState != "profile_reconnecting" || !strings.Contains(payload.SessionBinding.SessionHealthReason, "watchdog window") {
		t.Fatalf("expected reconnect timeout health reason, got %#v", payload.SessionBinding)
	}
	if payload.SessionBinding.SessionHealthRecoveryAction != "browser_runtime action=refresh" {
		t.Fatalf("expected reconnect timeout to expose refresh recovery action, got %#v", payload.SessionBinding)
	}
	if payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser_runtime action=refresh" || payload.SessionBinding.Coordination.NextStep != "browser_runtime action=refresh" {
		t.Fatalf("expected reconnect timeout to escalate to refresh, got %#v", payload.SessionBinding.Coordination)
	}
	if !browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=refresh") {
		t.Fatalf("expected refresh in reconnect-timeout recommended actions, got %#v", payload.SessionBinding.Coordination)
	}
}

func TestRegisterBrowserTools_RuntimeStopClearsImplicitTargetReuse(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "system-extract",
			BrowserApp: "Safari",
			Title:      "Host",
			Content:    "Host content",
			FinalURL:   "https://host.example",
		},
		runtimeStatusResult: BrowserProfileStatusResult{
			Backend:    "system",
			BrowserApp: "Safari",
			Profile:    "default",
			Status:     "running",
			Running:    true,
			Connected:  true,
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			extractResult: BrowserExtractResult{
				Backend:    "proxy-extract",
				BrowserApp: "Chromium",
				Title:      "Workbench",
				Content:    "Extracted content",
				FinalURL:   "https://node.example/workbench",
			},
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeStopResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "stopped",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-stop-clears-target")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	tracked := sessionRegistry.TrackTab("browser-runtime-stop-clears-target", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: fmt.Sprintf(`{"action":"select_target","target":"target:%s"}`, tracked.ID),
	}); err != nil {
		t.Fatalf("browser_runtime select_target: %v", err)
	}

	stopOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"stop","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime stop: %v", err)
	}
	var stopPayload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		StopDecision            string `json:"stop_decision"`
		StopReady               bool   `json:"stop_ready"`
		ClearedSessionTargets   int    `json:"cleared_session_targets"`
		SessionProfileSelection any    `json:"session_profile_selection"`
		SessionTargetSelection  any    `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(stopOut), &stopPayload); err != nil {
		t.Fatalf("decode stop output: %v", err)
	}
	if stopPayload.Action != "stop" || stopPayload.Status != "ok" || stopPayload.StopDecision != "stopped" || !stopPayload.StopReady {
		t.Fatalf("unexpected stop payload: %#v", stopPayload)
	}
	if stopPayload.ClearedSessionTargets != 1 {
		t.Fatalf("expected stop to clear 1 tracked session target, got %#v", stopPayload)
	}
	if stopPayload.SessionProfileSelection != nil {
		t.Fatalf("expected stop to clear session profile selection, got %#v", stopPayload.SessionProfileSelection)
	}
	if stopPayload.SessionTargetSelection != nil {
		t.Fatalf("expected stop to clear session target selection, got %#v", stopPayload.SessionTargetSelection)
	}

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	}); err != nil {
		t.Fatalf("browser_act extract after stop: %v", err)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 0 {
		t.Fatalf("expected promoted node extract after stop clears selected profile, got %#v", nodeBackend.extractReqs)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused after stop clears selected profile, got %#v", hostBackend.extractReqs)
	}

	statusOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status after stop: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 2 || nodeBackend.runtimeStatusReqs[1].Profile != "isolated" {
		t.Fatalf("expected status after stop to reuse the promoted node default route, got %#v", nodeBackend.runtimeStatusReqs)
	}
	var statusPayload struct {
		SelectedRoute struct {
			Backend       string `json:"backend"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionProfileSelection any `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(statusOut), &statusPayload); err != nil {
		t.Fatalf("decode status output: %v", err)
	}
	if statusPayload.SelectedRoute.Backend != "proxy" || statusPayload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected status to fall back to promoted node route after stop clears selected profile, got %#v", statusPayload.SelectedRoute)
	}
	if statusPayload.SessionProfileSelection != nil {
		t.Fatalf("expected status to keep session profile selection cleared after stop, got %#v", statusPayload.SessionProfileSelection)
	}
}
