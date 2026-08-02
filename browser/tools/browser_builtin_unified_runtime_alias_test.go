package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_UnifiedBrowserDelegatesRuntimeAction(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", Status: "running", Running: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(WithToolSessionID(context.Background(), "browser-unified-runtime"), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"workbench","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser workbench: %v", err)
	}
	var payload struct {
		Action          string                      `json:"action"`
		Status          string                      `json:"status"`
		BrowserTools    []string                    `json:"browser_tools"`
		BrowserActKinds []string                    `json:"browser_act_kinds"`
		View            *browserTopLevelViewSummary `json:"view"`
		SelectedRoute   struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "workbench" || payload.Status != "ok" {
		t.Fatalf("unexpected browser workbench payload: %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected unified runtime delegation to keep selected route, got %#v", payload.SelectedRoute)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_runtime") || !browserStringSliceContains(payload.BrowserTools, "browser_act") {
		t.Fatalf("expected unified runtime delegation payload to expose implicit specialist surface, got %#v", payload.BrowserTools)
	}
	if !browserStringSliceContains(payload.BrowserActKinds, "click") {
		t.Fatalf("expected unified runtime delegation payload to expose act kinds from selected route, got %#v", payload.BrowserActKinds)
	}
	if payload.View == nil ||
		payload.View.Kind != "workbench" ||
		payload.View.Category != "coordination" ||
		payload.View.State != "action_plan_available" ||
		payload.View.SummaryCode != "workbench_action_plan" {
		t.Fatalf("expected unified runtime delegation payload to expose workbench view, got %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserRefreshAliasDelegatesRestartCoordination(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-refresh")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-refresh", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  true,
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
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"refresh","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser refresh: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("unexpected refresh lifecycle requests: status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action               string `json:"action"`
		CoordinationGoal     string `json:"coordination_goal"`
		CoordinationDecision string `json:"coordination_decision"`
		RestartDecision      string `json:"restart_decision"`
		RestartReady         bool   `json:"restart_ready"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "refresh" || payload.CoordinationGoal != "restart" || payload.CoordinationDecision != "restart_ready" {
		t.Fatalf("unexpected browser refresh payload: %#v", payload)
	}
	if payload.RestartDecision != "restarted" || !payload.RestartReady {
		t.Fatalf("expected refresh alias to reuse restart coordination path, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeRefreshUsesRestartRecoveryPath(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  true,
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
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"refresh","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime refresh: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "isolated" {
		t.Fatalf("unexpected runtime status requests: %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "isolated" {
		t.Fatalf("unexpected runtime stop requests: %#v", nodeBackend.runtimeStopReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "isolated" {
		t.Fatalf("unexpected runtime start requests: %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action               string   `json:"action"`
		Status               string   `json:"status"`
		PreparedProfile      string   `json:"prepared_profile"`
		CoordinationGoal     string   `json:"coordination_goal"`
		CoordinationDecision string   `json:"coordination_decision"`
		RestartDecision      string   `json:"restart_decision"`
		RestartReady         bool     `json:"restart_ready"`
		RuntimeActions       []string `json:"runtime_actions"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "refresh" || payload.Status != "ok" || payload.PreparedProfile != "isolated" {
		t.Fatalf("unexpected runtime refresh payload: %#v", payload)
	}
	if payload.CoordinationGoal != "restart" || payload.CoordinationDecision != "restart_ready" {
		t.Fatalf("expected runtime refresh to surface restart coordination semantics, got %#v", payload)
	}
	if payload.RestartDecision != "restarted" || !payload.RestartReady {
		t.Fatalf("expected runtime refresh to reuse restart semantics, got %#v", payload)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "refresh") || !browserStringSliceContains(payload.RuntimeActions, "restart") {
		t.Fatalf("expected runtime actions to include refresh and restart, got %#v", payload.RuntimeActions)
	}
}

func TestRegisterBrowserTools_RuntimeRefreshSuppressesReconnectChurnWithinWatchdog(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-refresh-reconnecting")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	reconnectingAt := time.Now().Add(-12 * time.Second)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-refresh-reconnecting", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
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
		Arguments: `{"action":"refresh","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime refresh reconnecting: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 0 || len(nodeBackend.runtimeStopReqs) != 0 || len(nodeBackend.runtimeStartReqs) != 0 {
		t.Fatalf("expected refresh during reconnect watchdog window to avoid lifecycle churn, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		CoordinationGoal     string `json:"coordination_goal"`
		CoordinationDecision string `json:"coordination_decision"`
		CoordinationReady    bool   `json:"coordination_ready"`
		RestartDecision      string `json:"restart_decision"`
		RestartReady         bool   `json:"restart_ready"`
		ProfileStatus        struct {
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
	if payload.Action != "refresh" || payload.Status != "ok" || payload.CoordinationGoal != "restart" {
		t.Fatalf("unexpected runtime refresh reconnecting payload: %#v", payload)
	}
	if payload.RestartDecision != "restart_reconnect_in_progress" || payload.RestartReady || payload.CoordinationDecision != "restart_reconnect_in_progress" || payload.CoordinationReady {
		t.Fatalf("expected refresh during reconnect watchdog window to stay in-progress, got %#v", payload)
	}
	if payload.ProfileStatus.Status != "reconnecting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected refresh reconnecting status to be preserved, got %#v", payload.ProfileStatus)
	}
	if payload.SessionBinding.SessionHealthState != "profile_reconnecting" || payload.SessionBinding.SessionHealthRecoveryAction != "" {
		t.Fatalf("expected reconnecting session health to remain observed without refresh escalation, got %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeRefreshForceBypassesReconnectSuppression(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-refresh-force-reconnecting")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	reconnectingAt := time.Now().Add(-12 * time.Second)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-refresh-force-reconnecting", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		Arguments: `{"action":"refresh","profile":"isolated","runtime_target":"node","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime refresh reconnecting force: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("expected forced refresh to bypass reconnect suppression, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		RestartDecision string `json:"restart_decision"`
		RestartReady    bool   `json:"restart_ready"`
		ProfileStatus   struct {
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.RestartDecision != "restarted" || payload.RestartReady {
		t.Fatalf("expected forced refresh to reuse restart path, got %#v", payload)
	}
	if payload.ProfileStatus.Status != "reconnecting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected forced refresh payload to reflect reconnecting lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeRefreshRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-refresh-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
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
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-refresh-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"refresh"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime refresh managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime refresh to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime refresh to reuse managed current route stop profile before implicit host fallback, got %#v", nodeBackend.runtimeStopReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime refresh to reuse managed current route start profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		PreparedProfile      string `json:"prepared_profile"`
		CoordinationGoal     string `json:"coordination_goal"`
		CoordinationDecision string `json:"coordination_decision"`
		RestartDecision      string `json:"restart_decision"`
		RestartReady         bool   `json:"restart_ready"`
		RequestedBrowserApp  string `json:"requested_browser_app"`
		SelectedRoute        struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed current runtime refresh output: %v", err)
	}
	if payload.Action != "refresh" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.CoordinationGoal != "restart" || payload.CoordinationDecision != "restart_ready" {
		t.Fatalf("unexpected managed current runtime refresh payload: %#v", payload)
	}
	if payload.RestartDecision != "restarted" || !payload.RestartReady {
		t.Fatalf("unexpected managed current runtime refresh restart result: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime refresh to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime refresh to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected managed current runtime refresh profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserResetAliasDelegatesClearSession(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	target := sessionRegistry.TrackTab("browser-unified-reset", agentxbrowserruntime.BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/reset",
		Title:      "Reset",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-reset", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}},
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(WithToolSessionID(context.Background(), "browser-unified-reset"), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"reset","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser reset: %v", err)
	}
	var payload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		ClearSessionDecision   string `json:"clear_session_decision"`
		ClearSessionReady      bool   `json:"clear_session_ready"`
		ClearedSessionProfiles int    `json:"cleared_session_profiles"`
		ClearedSessionTargets  int    `json:"cleared_session_targets"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "clear_session" || payload.Status != "ok" || payload.ClearSessionDecision != "session_route_cleared" || !payload.ClearSessionReady {
		t.Fatalf("unexpected browser reset payload: %#v", payload)
	}
	if payload.ClearedSessionProfiles != 1 || payload.ClearedSessionTargets != 1 {
		t.Fatalf("expected reset alias to reuse clear_session path, got %#v", payload)
	}
	if _, ok := sessionRegistry.ResolveTarget("browser-unified-reset", target.ID); ok {
		t.Fatalf("expected reset alias to clear tracked target")
	}
}

func TestRegisterBrowserTools_UnifiedBrowserPinProfileAliasDelegatesSelectProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", Status: "running", BrowserApp: "Chromium", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      NewBrowserSessionRegistry(),
		SessionStateRegistry: NewBrowserSessionStateRegistry(),
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(WithToolSessionID(context.Background(), "browser-unified-pin-profile"), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"pin_profile","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser pin_profile: %v", err)
	}
	var payload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		SelectDecision          string `json:"select_decision"`
		SelectReady             bool   `json:"select_ready"`
		SessionProfileSelection struct {
			Profile string `json:"profile"`
			Source  string `json:"source"`
		} `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "select_profile" || payload.Status != "ok" || payload.SelectDecision != "session_profile_selected" || !payload.SelectReady {
		t.Fatalf("unexpected browser pin_profile payload: %#v", payload)
	}
	if payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected pin_profile alias to reuse select_profile path, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserAdoptAliasDelegatesSyncSession(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	tracked := sessionRegistry.TrackTabs("browser-unified-adopt", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://node.example/adopt",
			Title:      "Adopt",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
	}, 1)
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}},
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: NewBrowserSessionStateRegistry(),
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(WithToolSessionID(context.Background(), "browser-unified-adopt"), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"adopt","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser adopt: %v", err)
	}
	var payload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		SyncSessionDecision     string `json:"sync_session_decision"`
		SyncSessionReady        bool   `json:"sync_session_ready"`
		SessionProfileSelection struct {
			Profile string `json:"profile"`
			Source  string `json:"source"`
		} `json:"session_profile_selection"`
		SessionTargetSelection struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "sync_session" || payload.Status != "ok" || payload.SyncSessionDecision != "session_route_synced" || !payload.SyncSessionReady {
		t.Fatalf("unexpected browser adopt payload: %#v", payload)
	}
	if payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.Source != "sync_session" {
		t.Fatalf("expected adopt alias to reuse sync_session path for profile selection, got %#v", payload)
	}
	if payload.SessionTargetSelection.ID != tracked[0].ID || payload.SessionTargetSelection.TabIndex != 1 || payload.SessionTargetSelection.Source != "sync_session" {
		t.Fatalf("expected adopt alias to reuse sync_session path for target selection, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserNewProfileAliasDelegatesCreateProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeCreateResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "created",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"new_profile","profile":"workbench","browser_app":"Chromium","color":"#ff5500","copy_from":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser new_profile: %v", err)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		PreparedProfile     string `json:"prepared_profile"`
		CreateDecision      string `json:"create_decision"`
		CreateReady         bool   `json:"create_ready"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		RequestedColor      string `json:"requested_color"`
		RequestedCopyFrom   string `json:"requested_copy_from"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "create_profile" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.CreateDecision != "created" || !payload.CreateReady {
		t.Fatalf("unexpected browser new_profile payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" || payload.RequestedColor != "#ff5500" || payload.RequestedCopyFrom != "isolated" {
		t.Fatalf("expected new_profile alias to reuse create_profile path, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserRemoveProfileAliasDelegatesDeleteProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeDeleteResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "deleted",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"remove_profile","profile":"workbench","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser remove_profile: %v", err)
	}
	var payload struct {
		Action          string `json:"action"`
		Status          string `json:"status"`
		PreparedProfile string `json:"prepared_profile"`
		DeleteDecision  string `json:"delete_decision"`
		DeleteReady     bool   `json:"delete_ready"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "delete_profile" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.DeleteDecision != "deleted" || !payload.DeleteReady {
		t.Fatalf("unexpected browser remove_profile payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserNewProfileRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-new-profile-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeCreateResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "draft",
				Status:     "created",
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-new-profile-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"new_profile","profile":"draft","color":"#ff5500","copy_from":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser new_profile managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeCreateReqs) != 1 {
		t.Fatalf("expected browser new_profile to reuse managed current route before implicit host fallback, got %#v", nodeBackend.runtimeCreateReqs)
	}
	if nodeBackend.runtimeCreateReqs[0].Profile != "draft" || nodeBackend.runtimeCreateReqs[0].BrowserApp != "Chromium" || nodeBackend.runtimeCreateReqs[0].Color != "#ff5500" || nodeBackend.runtimeCreateReqs[0].CopyFrom != "isolated" {
		t.Fatalf("unexpected browser new_profile create request payload: %#v", nodeBackend.runtimeCreateReqs[0])
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		PreparedProfile     string `json:"prepared_profile"`
		CreateDecision      string `json:"create_decision"`
		CreateReady         bool   `json:"create_ready"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		RequestedColor      string `json:"requested_color"`
		RequestedCopyFrom   string `json:"requested_copy_from"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile    string `json:"profile"`
			BrowserApp string `json:"browser_app"`
			Status     string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser new_profile managed current output: %v", err)
	}
	if payload.Action != "create_profile" || payload.Status != "ok" || payload.PreparedProfile != "draft" || payload.CreateDecision != "created" || !payload.CreateReady {
		t.Fatalf("unexpected browser new_profile managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" || payload.RequestedColor != "#ff5500" || payload.RequestedCopyFrom != "isolated" {
		t.Fatalf("unexpected browser new_profile request echo: %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "draft" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser new_profile to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "draft" || payload.ProfileStatus.BrowserApp != "Chromium" || payload.ProfileStatus.Status != "created" {
		t.Fatalf("unexpected browser new_profile profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserRemoveProfileRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-remove-profile-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeDeleteResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "draft",
				Status:     "deleted",
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-remove-profile-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"remove_profile","profile":"draft","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser remove_profile managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeDeleteReqs) != 1 {
		t.Fatalf("expected browser remove_profile to reuse managed current route before implicit host fallback, got %#v", nodeBackend.runtimeDeleteReqs)
	}
	if nodeBackend.runtimeDeleteReqs[0].Profile != "draft" || !nodeBackend.runtimeDeleteReqs[0].Force {
		t.Fatalf("unexpected browser remove_profile delete request payload: %#v", nodeBackend.runtimeDeleteReqs[0])
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		PreparedProfile     string `json:"prepared_profile"`
		DeleteDecision      string `json:"delete_decision"`
		DeleteReady         bool   `json:"delete_ready"`
		Force               bool   `json:"force"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser remove_profile managed current output: %v", err)
	}
	if payload.Action != "delete_profile" || payload.Status != "ok" || payload.PreparedProfile != "draft" || payload.DeleteDecision != "deleted" || !payload.DeleteReady || !payload.Force {
		t.Fatalf("unexpected browser remove_profile managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser remove_profile to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "draft" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser remove_profile to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "draft" || payload.ProfileStatus.Status != "deleted" {
		t.Fatalf("unexpected browser remove_profile profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserResetRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-reset-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
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
		EnabledTools:         []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-reset-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"reset"}`,
	})
	if err != nil {
		t.Fatalf("browser reset managed current hidden implicit host: %v", err)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		ClearSessionDecision string `json:"clear_session_decision"`
		ClearSessionReady    bool   `json:"clear_session_ready"`
		RequestedBrowserApp  string `json:"requested_browser_app"`
		SelectedRoute        struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser reset managed current output: %v", err)
	}
	if payload.Action != "clear_session" || payload.Status != "ok" || payload.ClearSessionDecision == "" || !payload.ClearSessionReady {
		t.Fatalf("unexpected browser reset managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser reset to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser reset to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if _, ok := sessionRegistry.ResolveTarget("browser-unified-reset-managed-current-hidden-implicit-host", tracked.ID); ok {
		t.Fatalf("expected browser reset to clear tracked target after managed current route reset")
	}
}

func TestRegisterBrowserTools_UnifiedBrowserAdoptRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-adopt-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
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
		EnabledTools:         []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-adopt-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"adopt"}`,
	})
	if err != nil {
		t.Fatalf("browser adopt managed current hidden implicit host: %v", err)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		SyncSessionDecision string `json:"sync_session_decision"`
		SyncSessionReady    bool   `json:"sync_session_ready"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionProfileSelection struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			BrowserApp    string `json:"browser_app"`
			Source        string `json:"source"`
		} `json:"session_profile_selection"`
		SessionTargetSelection struct {
			ID            string `json:"id"`
			TabIndex      int    `json:"tab_index"`
			RuntimeTarget string `json:"runtime_target"`
			Profile       string `json:"profile"`
			BrowserApp    string `json:"browser_app"`
			Source        string `json:"source"`
		} `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser adopt managed current output: %v", err)
	}
	if payload.Action != "sync_session" || payload.Status != "ok" || payload.SyncSessionDecision != "session_route_synced" || !payload.SyncSessionReady {
		t.Fatalf("unexpected browser adopt managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser adopt to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser adopt to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.RuntimeTarget != "node" || payload.SessionProfileSelection.BrowserApp != "Chromium" || payload.SessionProfileSelection.Source != "sync_session" {
		t.Fatalf("unexpected browser adopt managed current profile selection: %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection.ID != tracked.ID || payload.SessionTargetSelection.TabIndex != 2 || payload.SessionTargetSelection.RuntimeTarget != "node" || payload.SessionTargetSelection.Profile != "workbench" || payload.SessionTargetSelection.BrowserApp != "Chromium" || payload.SessionTargetSelection.Source != "sync_session" {
		t.Fatalf("unexpected browser adopt managed current target selection: %#v", payload.SessionTargetSelection)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserPinProfileRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-pin-profile-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-pin-profile-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"pin_profile","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser pin_profile managed current hidden implicit host: %v", err)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		SelectDecision      string `json:"select_decision"`
		SelectReady         bool   `json:"select_ready"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionProfileSelection struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			BrowserApp    string `json:"browser_app"`
			Source        string `json:"source"`
		} `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser pin_profile managed current output: %v", err)
	}
	if payload.Action != "select_profile" || payload.Status != "ok" || payload.SelectDecision != "session_profile_selected" || !payload.SelectReady {
		t.Fatalf("unexpected browser pin_profile managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser pin_profile to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser pin_profile to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.RuntimeTarget != "node" || payload.SessionProfileSelection.BrowserApp != "Chromium" || payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("unexpected browser pin_profile managed current selection: %#v", payload.SessionProfileSelection)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserPinTargetRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-pin-target-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-pin-target-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	second := sessionRegistry.TrackTab("browser-unified-pin-target-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-second",
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"pin_target","target":"tab:2"}`,
	})
	if err != nil {
		t.Fatalf("browser pin_target managed current hidden implicit host: %v", err)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		SelectTargetDecision string `json:"select_target_decision"`
		SelectTargetReady    bool   `json:"select_target_ready"`
		RequestedBrowserApp  string `json:"requested_browser_app"`
		SelectedRoute        struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionTargetSelection struct {
			ID            string `json:"id"`
			TabIndex      int    `json:"tab_index"`
			RuntimeTarget string `json:"runtime_target"`
			Profile       string `json:"profile"`
			BrowserApp    string `json:"browser_app"`
			Source        string `json:"source"`
		} `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser pin_target managed current output: %v", err)
	}
	if payload.Action != "select_target" || payload.Status != "ok" || payload.SelectTargetDecision != "session_target_selected" || !payload.SelectTargetReady {
		t.Fatalf("unexpected browser pin_target managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser pin_target to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser pin_target to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SessionTargetSelection.ID != second.ID || payload.SessionTargetSelection.TabIndex != 2 || payload.SessionTargetSelection.RuntimeTarget != "node" || payload.SessionTargetSelection.Profile != "workbench" || payload.SessionTargetSelection.BrowserApp != "Chromium" || payload.SessionTargetSelection.Source != "select_target" {
		t.Fatalf("unexpected browser pin_target managed current selection: %#v", payload.SessionTargetSelection)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserSyncAliasDelegatesCoordinate(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	nodeBackend.runtimeProfilesResult = BrowserProfilesResult{
		Backend:        "proxy",
		DefaultProfile: "workbench",
		Profiles: []BrowserProfileInfo{
			{Profile: "workbench", Status: "running", BrowserApp: "Chromium"},
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser"},
	})

	tracked := sessionRegistry.TrackTabs("browser-unified-sync", []agentxbrowserruntime.BrowserSessionTarget{
		{
			ID:         "target-1",
			TabIndex:   1,
			URL:        "https://node.example/workbench",
			Title:      "Workbench",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
	}, 1)
	out, err := reg.Execute(WithToolSessionID(context.Background(), "browser-unified-sync"), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"sync","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser sync: %v", err)
	}
	var payload struct {
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		CoordinationGoal       string `json:"coordination_goal"`
		CoordinationDecision   string `json:"coordination_decision"`
		CoordinationReady      bool   `json:"coordination_ready"`
		SyncSessionDecision    string `json:"sync_session_decision"`
		SyncSessionReady       bool   `json:"sync_session_ready"`
		SessionTargetSelection struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser sync output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.CoordinationGoal != "sync" || !payload.CoordinationReady {
		t.Fatalf("unexpected browser sync payload: %#v", payload)
	}
	if payload.CoordinationDecision != "started" || payload.SyncSessionDecision != "session_target_synced" || !payload.SyncSessionReady {
		t.Fatalf("unexpected browser sync coordination decisions: %#v", payload)
	}
	if payload.SessionTargetSelection.ID != tracked[0].ID || payload.SessionTargetSelection.TabIndex != 1 || payload.SessionTargetSelection.Source != "sync_session" {
		t.Fatalf("expected sync alias to reuse coordinate sync path, got %#v", payload.SessionTargetSelection)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserTeardownAliasDelegatesCoordinate(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-teardown")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-teardown", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeStopResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "stopped",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"teardown","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser teardown: %v", err)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "isolated" {
		t.Fatalf("expected teardown alias to reuse coordinate teardown path, got %#v", nodeBackend.runtimeStopReqs)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		CoordinationGoal     string `json:"coordination_goal"`
		CoordinationDecision string `json:"coordination_decision"`
		CoordinationReady    bool   `json:"coordination_ready"`
		ProfileStatus        struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser teardown output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.CoordinationGoal != "teardown" || payload.CoordinationDecision != "teardown_stopped" || !payload.CoordinationReady {
		t.Fatalf("unexpected browser teardown payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "stopped" {
		t.Fatalf("expected teardown alias to reuse coordinate teardown profile status, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserUnpinProfileAliasDelegatesClearProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.SelectBrowserProfile("browser-unified-unpin-profile", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"}},
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(WithToolSessionID(context.Background(), "browser-unified-unpin-profile"), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"unpin_profile","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser unpin_profile: %v", err)
	}
	var payload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		ClearDecision           string `json:"clear_decision"`
		ClearReady              bool   `json:"clear_ready"`
		SessionProfileSelection any    `json:"session_profile_selection"`
		SessionBinding          struct {
			SelectedBrowserProfile string `json:"selected_browser_profile"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser unpin_profile output: %v", err)
	}
	if payload.Action != "clear_profile" || payload.Status != "ok" || payload.ClearDecision != "session_profile_cleared" || !payload.ClearReady {
		t.Fatalf("unexpected browser unpin_profile payload: %#v", payload)
	}
	if payload.SessionProfileSelection != nil || payload.SessionBinding.SelectedBrowserProfile != "" {
		t.Fatalf("expected unpin_profile alias to reuse clear_profile path, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserUnpinTargetAliasDelegatesClearTarget(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-unpin-target")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-unpin-target", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	second := sessionRegistry.TrackTab("browser-unified-unpin-target", BrowserSessionTarget{
		ID:         "node-second",
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"pin_profile","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser pin_profile setup: %v", err)
	}
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"pin_target","target":"tab:2"}`,
	}); err != nil {
		t.Fatalf("browser pin_target setup: %v", err)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: fmt.Sprintf(`{"action":"unpin_target","target":"target:%s"}`, second.ID),
	})
	if err != nil {
		t.Fatalf("browser unpin_target: %v", err)
	}
	var payload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		ClearTargetDecision     string `json:"clear_target_decision"`
		ClearTargetReady        bool   `json:"clear_target_ready"`
		SessionTargetSelection  any    `json:"session_target_selection"`
		SessionProfileSelection struct {
			Profile string `json:"profile"`
			Source  string `json:"source"`
		} `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser unpin_target output: %v", err)
	}
	if payload.Action != "clear_target" || payload.Status != "ok" || payload.ClearTargetDecision != "session_target_cleared" || !payload.ClearTargetReady {
		t.Fatalf("unexpected browser unpin_target payload: %#v", payload)
	}
	if payload.SessionTargetSelection != nil {
		t.Fatalf("expected unpin_target alias to clear session target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected unpin_target alias to preserve explicit session profile selection, got %#v", payload.SessionProfileSelection)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserSyncRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-sync-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	nodeBackend.runtimeProfilesResult = BrowserProfilesResult{
		Backend:        "proxy",
		DefaultProfile: "workbench",
		Profiles: []BrowserProfileInfo{
			{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-sync-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"sync"}`,
	})
	if err != nil {
		t.Fatalf("browser sync managed current hidden implicit host: %v", err)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		CoordinationGoal    string `json:"coordination_goal"`
		CoordinationReady   bool   `json:"coordination_ready"`
		SyncSessionDecision string `json:"sync_session_decision"`
		SyncSessionReady    bool   `json:"sync_session_ready"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionTargetSelection struct {
			ID            string `json:"id"`
			TabIndex      int    `json:"tab_index"`
			RuntimeTarget string `json:"runtime_target"`
			Profile       string `json:"profile"`
			BrowserApp    string `json:"browser_app"`
			Source        string `json:"source"`
		} `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser sync managed current output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.CoordinationGoal != "sync" || !payload.CoordinationReady || payload.SyncSessionDecision == "" || !payload.SyncSessionReady {
		t.Fatalf("unexpected browser sync managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser sync to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser sync to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SessionTargetSelection.ID != tracked.ID || payload.SessionTargetSelection.TabIndex != 2 || payload.SessionTargetSelection.RuntimeTarget != "node" || payload.SessionTargetSelection.Profile != "workbench" || payload.SessionTargetSelection.BrowserApp != "Chromium" || payload.SessionTargetSelection.Source != "sync_session" {
		t.Fatalf("unexpected browser sync managed current target selection: %#v", payload.SessionTargetSelection)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserTeardownRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-teardown-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-teardown-managed-current-hidden-implicit-host", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
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
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-teardown-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"teardown"}`,
	})
	if err != nil {
		t.Fatalf("browser teardown managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser teardown to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser teardown to reuse managed current route stop profile before implicit host fallback, got %#v", nodeBackend.runtimeStopReqs)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		CoordinationGoal     string `json:"coordination_goal"`
		CoordinationDecision string `json:"coordination_decision"`
		CoordinationReady    bool   `json:"coordination_ready"`
		RequestedBrowserApp  string `json:"requested_browser_app"`
		SelectedRoute        struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser teardown managed current output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.CoordinationGoal != "teardown" || payload.CoordinationDecision != "teardown_stopped" || !payload.CoordinationReady {
		t.Fatalf("unexpected browser teardown managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser teardown to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser teardown to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "stopped" {
		t.Fatalf("unexpected browser teardown managed current profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserUnpinProfileRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-unpin-profile-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.SelectBrowserProfile("browser-unified-unpin-profile-managed-current-hidden-implicit-host", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Source:        "select_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"}},
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-unpin-profile-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"unpin_profile"}`,
	})
	if err != nil {
		t.Fatalf("browser unpin_profile managed current hidden implicit host: %v", err)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		ClearDecision       string `json:"clear_decision"`
		ClearReady          bool   `json:"clear_ready"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionProfileSelection any `json:"session_profile_selection"`
		SessionBinding          struct {
			SelectedBrowserProfile string `json:"selected_browser_profile"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser unpin_profile managed current output: %v", err)
	}
	if payload.Action != "clear_profile" || payload.Status != "ok" || payload.ClearDecision != "session_profile_cleared" || !payload.ClearReady {
		t.Fatalf("unexpected browser unpin_profile managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser unpin_profile to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser unpin_profile to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SessionProfileSelection != nil || payload.SessionBinding.SelectedBrowserProfile != "" {
		t.Fatalf("expected browser unpin_profile to clear explicit session profile selection, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserUnpinTargetRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-unpin-target-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-unpin-target-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	second := sessionRegistry.TrackTab("browser-unified-unpin-target-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-second",
		TabIndex:   2,
		URL:        "https://node.example/two",
		Title:      "Two",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, false)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"pin_profile","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser pin_profile setup managed current hidden implicit host: %v", err)
	}
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"pin_target","target":"tab:2"}`,
	}); err != nil {
		t.Fatalf("browser pin_target setup managed current hidden implicit host: %v", err)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: fmt.Sprintf(`{"action":"unpin_target","target":"target:%s"}`, second.ID),
	})
	if err != nil {
		t.Fatalf("browser unpin_target managed current hidden implicit host: %v", err)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		ClearTargetDecision string `json:"clear_target_decision"`
		ClearTargetReady    bool   `json:"clear_target_ready"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionTargetSelection  any `json:"session_target_selection"`
		SessionProfileSelection struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			BrowserApp    string `json:"browser_app"`
			Source        string `json:"source"`
		} `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser unpin_target managed current output: %v", err)
	}
	if payload.Action != "clear_target" || payload.Status != "ok" || payload.ClearTargetDecision != "session_target_cleared" || !payload.ClearTargetReady {
		t.Fatalf("unexpected browser unpin_target managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser unpin_target to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser unpin_target to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SessionTargetSelection != nil {
		t.Fatalf("expected browser unpin_target to clear explicit session target selection, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.RuntimeTarget != "node" || payload.SessionProfileSelection.BrowserApp != "Chromium" || payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("expected browser unpin_target to preserve explicit session profile selection, got %#v", payload.SessionProfileSelection)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserRefreshRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-refresh-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-refresh-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"refresh"}`,
	})
	if err != nil {
		t.Fatalf("browser refresh managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser refresh to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser refresh to reuse managed current route stop profile before implicit host fallback, got %#v", nodeBackend.runtimeStopReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser refresh to reuse managed current route start profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		PreparedProfile      string `json:"prepared_profile"`
		CoordinationGoal     string `json:"coordination_goal"`
		CoordinationDecision string `json:"coordination_decision"`
		RestartDecision      string `json:"restart_decision"`
		RestartReady         bool   `json:"restart_ready"`
		RequestedBrowserApp  string `json:"requested_browser_app"`
		SelectedRoute        struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser refresh managed current output: %v", err)
	}
	if payload.Action != "refresh" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.CoordinationGoal != "restart" || payload.CoordinationDecision != "restart_ready" {
		t.Fatalf("unexpected browser refresh managed current payload: %#v", payload)
	}
	if payload.RestartDecision != "restarted" || !payload.RestartReady {
		t.Fatalf("unexpected browser refresh managed current restart result: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser refresh to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser refresh to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected browser refresh managed current profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserCoordinateRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-coordinate-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-coordinate-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"coordinate","coordination_goal":"restart"}`,
	})
	if err != nil {
		t.Fatalf("browser coordinate managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser coordinate to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser coordinate to reuse managed current route stop profile before implicit host fallback, got %#v", nodeBackend.runtimeStopReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser coordinate to reuse managed current route start profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		PreparedProfile      string `json:"prepared_profile"`
		CoordinationGoal     string `json:"coordination_goal"`
		CoordinationDecision string `json:"coordination_decision"`
		CoordinationReady    bool   `json:"coordination_ready"`
		RestartDecision      string `json:"restart_decision"`
		RestartReady         bool   `json:"restart_ready"`
		RequestedBrowserApp  string `json:"requested_browser_app"`
		SelectedRoute        struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser coordinate managed current output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.CoordinationGoal != "restart" || payload.CoordinationDecision != "restart_ready" || !payload.CoordinationReady {
		t.Fatalf("unexpected browser coordinate managed current payload: %#v", payload)
	}
	if payload.RestartDecision != "restarted" || !payload.RestartReady {
		t.Fatalf("unexpected browser coordinate managed current restart result: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser coordinate to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser coordinate to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected browser coordinate managed current profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserInspectAliasDelegatesStatus(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-inspect")
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"inspect","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser inspect: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 {
		t.Fatalf("expected inspect alias to reuse status path, got %#v", nodeBackend.runtimeStatusReqs)
	}
	var payload struct {
		Action string `json:"action"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "status" || payload.Status != "ok" {
		t.Fatalf("unexpected browser inspect payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserDoctorAliasDelegatesStatus(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-doctor")
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  true,
				PlaywrightCache: &agentxbrowserruntime.BrowserPlaywrightCacheSummary{
					SelectedLaunchSource: "runtime_observed",
					SelectedLaunchReady:  true,
					LaunchReady:          true,
					BootstrapState:       "ready",
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"doctor","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser doctor: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 {
		t.Fatalf("expected doctor alias to reuse status path, got %#v", nodeBackend.runtimeStatusReqs)
	}
	var payload struct {
		Action string `json:"action"`
		Status string `json:"status"`
		Doctor struct {
			Status string `json:"status"`
			Ready  bool   `json:"ready"`
			Route  struct {
				Status string `json:"status"`
				Code   string `json:"code"`
			} `json:"route"`
			Launch struct {
				Status string `json:"status"`
				Code   string `json:"code"`
			} `json:"launch"`
		} `json:"doctor"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "status" || payload.Status != "ok" || payload.Doctor.Status != "ok" || !payload.Doctor.Ready || payload.Doctor.Route.Code != "managed_default_route" || payload.Doctor.Launch.Code != "launch_ready" {
		t.Fatalf("unexpected browser doctor payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_BrowserRuntimeDoctorActionDelegatesStatus(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  true,
				PlaywrightCache: &agentxbrowserruntime.BrowserPlaywrightCacheSummary{
					SelectedLaunchSource: "runtime_observed",
					SelectedLaunchReady:  true,
					LaunchReady:          true,
					BootstrapState:       "ready",
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"doctor","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime doctor: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 {
		t.Fatalf("expected browser_runtime doctor to reuse status path, got %#v", nodeBackend.runtimeStatusReqs)
	}
	var payload struct {
		Action string `json:"action"`
		Status string `json:"status"`
		Doctor struct {
			Status string `json:"status"`
			Ready  bool   `json:"ready"`
		} `json:"doctor"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "status" || payload.Status != "ok" || payload.Doctor.Status != "ok" || !payload.Doctor.Ready {
		t.Fatalf("unexpected browser_runtime doctor payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserLaunchAliasDelegatesStart(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStartResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "started",
				Running:    true,
				Connected:  true,
				PlaywrightCache: &agentxbrowserruntime.BrowserPlaywrightCacheSummary{
					NodeVersion:          "24.2.0",
					SelectedLaunchSource: "runtime_observed",
					SelectedLaunchReady:  true,
					LaunchReady:          true,
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"launch","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser launch: %v", err)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "isolated" {
		t.Fatalf("expected launch alias to reuse start path, got %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action            string `json:"action"`
		Status            string `json:"status"`
		LaunchDiagnostics struct {
			Source               string `json:"source"`
			Profile              string `json:"profile"`
			RuntimeTarget        string `json:"runtime_target"`
			Status               string `json:"status"`
			NodeVersion          string `json:"node_version"`
			SelectedLaunchSource string `json:"selected_launch_source"`
			SelectedLaunchReady  *bool  `json:"selected_launch_ready"`
			LaunchReady          *bool  `json:"launch_ready"`
		} `json:"launch_diagnostics"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "start" || payload.Status != "ok" {
		t.Fatalf("unexpected browser launch payload: %#v", payload)
	}
	if payload.LaunchDiagnostics.Source != "runtime_status" ||
		payload.LaunchDiagnostics.Profile != "isolated" ||
		payload.LaunchDiagnostics.RuntimeTarget != "node" ||
		payload.LaunchDiagnostics.Status != "started" ||
		payload.LaunchDiagnostics.NodeVersion != "24.2.0" ||
		payload.LaunchDiagnostics.SelectedLaunchSource != "runtime_observed" ||
		payload.LaunchDiagnostics.SelectedLaunchReady == nil || !*payload.LaunchDiagnostics.SelectedLaunchReady ||
		payload.LaunchDiagnostics.LaunchReady == nil || !*payload.LaunchDiagnostics.LaunchReady {
		t.Fatalf("unexpected browser launch diagnostics payload: %#v", payload.LaunchDiagnostics)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserHaltAliasDelegatesStop(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeStopResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "stopped",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"halt","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser halt: %v", err)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "isolated" {
		t.Fatalf("expected halt alias to reuse stop path, got %#v", nodeBackend.runtimeStopReqs)
	}
	var payload struct {
		Action       string `json:"action"`
		Status       string `json:"status"`
		StopDecision string `json:"stop_decision"`
		StopReady    bool   `json:"stop_ready"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "stop" || payload.Status != "ok" || payload.StopDecision != "stopped" || !payload.StopReady {
		t.Fatalf("unexpected browser halt payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserHaltRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-halt-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
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
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-halt-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"halt"}`,
	})
	if err != nil {
		t.Fatalf("browser halt managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser halt to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser halt to reuse managed current route stop profile before implicit host fallback, got %#v", nodeBackend.runtimeStopReqs)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		StopDecision        string `json:"stop_decision"`
		StopReady           bool   `json:"stop_ready"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser halt managed current output: %v", err)
	}
	if payload.Action != "stop" || payload.Status != "ok" || payload.StopDecision != "stopped" || !payload.StopReady {
		t.Fatalf("unexpected browser halt managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser halt to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser halt to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "stopped" {
		t.Fatalf("unexpected browser halt managed current profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserRestartRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-restart-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-restart-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"restart"}`,
	})
	if err != nil {
		t.Fatalf("browser restart managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser restart to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser restart to reuse managed current route stop profile before implicit host fallback, got %#v", nodeBackend.runtimeStopReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser restart to reuse managed current route start profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		PreparedProfile     string `json:"prepared_profile"`
		RestartDecision     string `json:"restart_decision"`
		RestartReady        bool   `json:"restart_ready"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser restart managed current output: %v", err)
	}
	if payload.Action != "restart" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.RestartDecision != "restarted" || !payload.RestartReady {
		t.Fatalf("unexpected browser restart managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser restart to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser restart to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected browser restart managed current profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserReadyAliasDelegatesPrepare(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-ready")
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
				Connected:  true,
				PlaywrightCache: &agentxbrowserruntime.BrowserPlaywrightCacheSummary{
					NodeVersion:          "24.2.0",
					SelectedLaunchSource: "runtime_observed",
					SelectedLaunchReady:  true,
					LaunchReady:          true,
				},
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "stopped"},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"ready","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser ready: %v", err)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 0 || len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("expected ready alias to reuse prepare lifecycle path, got profiles=%#v status=%#v start=%#v", nodeBackend.runtimeProfilesReqs, nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action            string `json:"action"`
		Status            string `json:"status"`
		PreparedProfile   string `json:"prepared_profile"`
		PrepareDecision   string `json:"prepare_decision"`
		PrepareReady      bool   `json:"prepare_ready"`
		LaunchDiagnostics struct {
			Source               string `json:"source"`
			Profile              string `json:"profile"`
			RuntimeTarget        string `json:"runtime_target"`
			Status               string `json:"status"`
			NodeVersion          string `json:"node_version"`
			SelectedLaunchSource string `json:"selected_launch_source"`
			SelectedLaunchReady  *bool  `json:"selected_launch_ready"`
			LaunchReady          *bool  `json:"launch_ready"`
		} `json:"launch_diagnostics"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "prepare" || payload.Status != "ok" || payload.PreparedProfile != "isolated" || payload.PrepareDecision != "started" || !payload.PrepareReady {
		t.Fatalf("unexpected browser ready payload: %#v", payload)
	}
	if payload.LaunchDiagnostics.Source != "runtime_status" ||
		payload.LaunchDiagnostics.Profile != "isolated" ||
		payload.LaunchDiagnostics.RuntimeTarget != "node" ||
		payload.LaunchDiagnostics.Status != "started" ||
		payload.LaunchDiagnostics.NodeVersion != "24.2.0" ||
		payload.LaunchDiagnostics.SelectedLaunchSource != "runtime_observed" ||
		payload.LaunchDiagnostics.SelectedLaunchReady == nil || !*payload.LaunchDiagnostics.SelectedLaunchReady ||
		payload.LaunchDiagnostics.LaunchReady == nil || !*payload.LaunchDiagnostics.LaunchReady {
		t.Fatalf("unexpected browser ready diagnostics payload: %#v", payload.LaunchDiagnostics)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserEnsureAliasDelegatesCoordinate(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-ensure")
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
				Connected:  true,
				PlaywrightCache: &agentxbrowserruntime.BrowserPlaywrightCacheSummary{
					NodeVersion:          "24.2.0",
					SelectedLaunchSource: "runtime_observed",
					SelectedLaunchReady:  true,
					LaunchReady:          true,
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"ensure","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser ensure: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("expected ensure alias to reuse coordinate lifecycle path, got status=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		PreparedProfile      string `json:"prepared_profile"`
		CoordinationGoal     string `json:"coordination_goal"`
		CoordinationDecision string `json:"coordination_decision"`
		CoordinationReady    bool   `json:"coordination_ready"`
		LaunchDiagnostics    struct {
			Source               string `json:"source"`
			Profile              string `json:"profile"`
			RuntimeTarget        string `json:"runtime_target"`
			Status               string `json:"status"`
			NodeVersion          string `json:"node_version"`
			SelectedLaunchSource string `json:"selected_launch_source"`
			SelectedLaunchReady  *bool  `json:"selected_launch_ready"`
			LaunchReady          *bool  `json:"launch_ready"`
		} `json:"launch_diagnostics"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.PreparedProfile != "isolated" || payload.CoordinationGoal != "ensure" || payload.CoordinationReady == false {
		t.Fatalf("unexpected browser ensure payload: %#v", payload)
	}
	if payload.CoordinationDecision != "started" && payload.CoordinationDecision != "started_browser_profile" {
		t.Fatalf("unexpected browser ensure payload: %#v", payload)
	}
	if payload.LaunchDiagnostics.Source != "runtime_status" ||
		payload.LaunchDiagnostics.Profile != "isolated" ||
		payload.LaunchDiagnostics.RuntimeTarget != "node" ||
		payload.LaunchDiagnostics.Status != "started" ||
		payload.LaunchDiagnostics.NodeVersion != "24.2.0" ||
		payload.LaunchDiagnostics.SelectedLaunchSource != "runtime_observed" ||
		payload.LaunchDiagnostics.SelectedLaunchReady == nil || !*payload.LaunchDiagnostics.SelectedLaunchReady ||
		payload.LaunchDiagnostics.LaunchReady == nil || !*payload.LaunchDiagnostics.LaunchReady {
		t.Fatalf("unexpected browser ensure diagnostics payload: %#v", payload.LaunchDiagnostics)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserReadyRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-ready-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-ready-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"ready"}`,
	})
	if err != nil {
		t.Fatalf("browser ready managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser ready to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser ready to reuse managed current route start profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		PreparedProfile     string `json:"prepared_profile"`
		PrepareDecision     string `json:"prepare_decision"`
		PrepareReady        bool   `json:"prepare_ready"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser ready managed current output: %v", err)
	}
	if payload.Action != "prepare" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.PrepareDecision != "started" || !payload.PrepareReady {
		t.Fatalf("unexpected browser ready managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser ready to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser ready to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected browser ready managed current profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserPrepareRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-prepare-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-prepare-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"prepare"}`,
	})
	if err != nil {
		t.Fatalf("browser prepare managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser prepare to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser prepare to reuse managed current route start profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		PreparedProfile     string `json:"prepared_profile"`
		PrepareDecision     string `json:"prepare_decision"`
		PrepareReady        bool   `json:"prepare_ready"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser prepare managed current output: %v", err)
	}
	if payload.Action != "prepare" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.PrepareDecision != "started" || !payload.PrepareReady {
		t.Fatalf("unexpected browser prepare managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser prepare to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser prepare to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected browser prepare managed current profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserEnsureRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-ensure-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-ensure-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"ensure"}`,
	})
	if err != nil {
		t.Fatalf("browser ensure managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser ensure to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser ensure to reuse managed current route start profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		PreparedProfile      string `json:"prepared_profile"`
		CoordinationGoal     string `json:"coordination_goal"`
		CoordinationDecision string `json:"coordination_decision"`
		CoordinationReady    bool   `json:"coordination_ready"`
		RequestedBrowserApp  string `json:"requested_browser_app"`
		SelectedRoute        struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser ensure managed current output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.CoordinationGoal != "ensure" || !payload.CoordinationReady {
		t.Fatalf("unexpected browser ensure managed current payload: %#v", payload)
	}
	if payload.CoordinationDecision != "started" && payload.CoordinationDecision != "started_browser_profile" {
		t.Fatalf("unexpected browser ensure managed current coordination payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser ensure to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser ensure to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected browser ensure managed current profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserLaunchRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-launch-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-launch-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"launch"}`,
	})
	if err != nil {
		t.Fatalf("browser launch managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser launch to reuse managed current route start profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser launch managed current output: %v", err)
	}
	if payload.Action != "start" || payload.Status != "ok" {
		t.Fatalf("unexpected browser launch managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser launch to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser launch to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected browser launch managed current profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserStartRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-start-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
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
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-start-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"start"}`,
	})
	if err != nil {
		t.Fatalf("browser start managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser start to reuse managed current route start profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser start managed current output: %v", err)
	}
	if payload.Action != "start" || payload.Status != "ok" {
		t.Fatalf("unexpected browser start managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser start to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser start to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected browser start managed current profile status: %#v", payload.ProfileStatus)
	}
}
