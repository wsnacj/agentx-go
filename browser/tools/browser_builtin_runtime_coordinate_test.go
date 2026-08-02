package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_RuntimeCoordinate(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-session")
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry.Record("browser-runtime-coordinate-session", BrowserSessionRunInfo{
		RunID:    "run-55",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
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
					{Profile: "relay", BrowserApp: "Chromium", Status: "stopped"},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRunRegistry:   sessionRunRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"coordinate","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime coordinate: %v", err)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 0 || len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("unexpected runtime coordination requests: profiles=%#v status=%#v start=%#v", nodeBackend.runtimeProfilesReqs, nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action               string   `json:"action"`
		Status               string   `json:"status"`
		PreparedProfile      string   `json:"prepared_profile"`
		CoordinationDecision string   `json:"coordination_decision"`
		CoordinationState    string   `json:"coordination_state"`
		CoordinationReady    bool     `json:"coordination_ready"`
		RuntimeActions       []string `json:"runtime_actions"`
		SessionBinding       struct {
			ActiveNodeRunID      string `json:"active_node_run_id"`
			ActiveBrowserProfile string `json:"active_browser_profile"`
			Coordination         struct {
				State                string `json:"state"`
				SyncBrowserAction    string `json:"sync_browser_action"`
				PrimaryBrowserAction string `json:"primary_browser_action"`
				PrimaryNodeAction    string `json:"primary_node_action"`
				NextStep             string `json:"next_step"`
			} `json:"coordination"`
		} `json:"session_binding"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
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
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.PreparedProfile != "isolated" {
		t.Fatalf("unexpected runtime coordinate payload: %#v", payload)
	}
	if payload.CoordinationDecision != "started_for_active_node_run" || payload.CoordinationState != "coordinated" || !payload.CoordinationReady {
		t.Fatalf("unexpected coordination result: %#v", payload)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "coordinate") {
		t.Fatalf("expected runtime actions to include coordinate, got %#v", payload.RuntimeActions)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected runtime coordinate profile status: %#v", payload.ProfileStatus)
	}
	if payload.LaunchDiagnostics.Source != "runtime_status" ||
		payload.LaunchDiagnostics.Profile != "isolated" ||
		payload.LaunchDiagnostics.RuntimeTarget != "node" ||
		payload.LaunchDiagnostics.Status != "started" ||
		payload.LaunchDiagnostics.NodeVersion != "24.2.0" ||
		payload.LaunchDiagnostics.SelectedLaunchSource != "runtime_observed" ||
		payload.LaunchDiagnostics.SelectedLaunchReady == nil || !*payload.LaunchDiagnostics.SelectedLaunchReady ||
		payload.LaunchDiagnostics.LaunchReady == nil || !*payload.LaunchDiagnostics.LaunchReady {
		t.Fatalf("unexpected runtime coordinate diagnostics payload: %#v", payload.LaunchDiagnostics)
	}
	if payload.SessionBinding.ActiveNodeRunID != "run-55" || payload.SessionBinding.ActiveBrowserProfile != "isolated" || payload.SessionBinding.Coordination.State != "coordinated" {
		t.Fatalf("unexpected runtime coordinate session binding: %#v", payload.SessionBinding)
	}
	if payload.SessionBinding.Coordination.SyncBrowserAction != "" || payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser_runtime action=workbench" || payload.SessionBinding.Coordination.PrimaryNodeAction != "nodes action=run_status" || payload.SessionBinding.Coordination.NextStep != "nodes action=run_status" {
		t.Fatalf("unexpected runtime coordinate action plan: %#v", payload.SessionBinding.Coordination)
	}
}

func TestRegisterBrowserTools_RuntimeCoordinateTeardown(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-teardown")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-coordinate-teardown", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"coordinate","runtime_target":"node","coordination_goal":"teardown"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime coordinate teardown: %v", err)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "isolated" {
		t.Fatalf("unexpected runtime stop requests: %#v", nodeBackend.runtimeStopReqs)
	}
	var payload struct {
		Action               string `json:"action"`
		Status               string `json:"status"`
		CoordinationGoal     string `json:"coordination_goal"`
		CoordinationDecision string `json:"coordination_decision"`
		CoordinationState    string `json:"coordination_state"`
		CoordinationReady    bool   `json:"coordination_ready"`
		ProfileStatus        struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.CoordinationGoal != "teardown" {
		t.Fatalf("unexpected runtime coordinate teardown payload: %#v", payload)
	}
	if payload.CoordinationDecision != "teardown_stopped" || !payload.CoordinationReady {
		t.Fatalf("unexpected teardown coordination result: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "stopped" {
		t.Fatalf("unexpected teardown profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeCoordinateTeardownUsesStoredStoppedStateWhenRawStatusIsWeak(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-teardown-stored-stopped")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-coordinate-teardown-stored-stopped", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "stopped",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
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
		Arguments: `{"action":"coordinate","runtime_target":"node","profile":"isolated","coordination_goal":"teardown"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime coordinate teardown stored stopped state: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 0 {
		t.Fatalf("expected teardown to trust stored stopped state without issuing RuntimeStop, got status=%#v stop=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs)
	}
	if !strings.Contains(out, `"coordination_decision":"teardown_already_stopped"`) || !strings.Contains(out, `"status":"stopped"`) {
		t.Fatalf("unexpected runtime teardown stored-stopped output: %s", out)
	}
}

func TestRegisterBrowserTools_RuntimeCoordinateTeardownStatusFailureUsesEffectiveLifecycleState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-teardown-status-failed")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-coordinate-teardown-status-failed", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "reconnecting",
		Running:       true,
		Connected:     false,
		Note:          "cdp reconnect in progress",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusErr: fmt.Errorf("status failed"),
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
		Arguments: `{"action":"coordinate","runtime_target":"node","profile":"isolated","coordination_goal":"teardown"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime coordinate teardown status failure: %v", err)
	}
	var payload struct {
		Status               string `json:"status"`
		CoordinationDecision string `json:"coordination_decision"`
		CoordinationReady    bool   `json:"coordination_ready"`
		ProfileStatus        struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "error" || payload.CoordinationDecision != "teardown_status_failed" || payload.CoordinationReady {
		t.Fatalf("unexpected teardown status failure payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "reconnecting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected teardown status failure payload to keep effective lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeCoordinateTeardownStopFailureUsesEffectiveLifecycleState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-teardown-stop-failed")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-coordinate-teardown-stop-failed", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "browser active",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
			},
			runtimeStopErr: fmt.Errorf("stop failed"),
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
		Arguments: `{"action":"coordinate","runtime_target":"node","profile":"isolated","coordination_goal":"teardown"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime coordinate teardown stop failure: %v", err)
	}
	var payload struct {
		Status               string `json:"status"`
		CoordinationDecision string `json:"coordination_decision"`
		CoordinationReady    bool   `json:"coordination_ready"`
		ProfileStatus        struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "error" || payload.CoordinationDecision != "teardown_stop_failed" || payload.CoordinationReady {
		t.Fatalf("unexpected teardown stop failure payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("expected teardown stop failure payload to keep effective lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeCoordinateTeardownWeakStopResultUsesLifecycleStoppedState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-teardown-weak-stop")
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

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"coordinate","runtime_target":"node","profile":"isolated","coordination_goal":"teardown"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime coordinate teardown weak stop result: %v", err)
	}
	var payload struct {
		CoordinationDecision string `json:"coordination_decision"`
		CoordinationReady    bool   `json:"coordination_ready"`
		ProfileStatus        struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.CoordinationDecision != "teardown_stopped" || !payload.CoordinationReady {
		t.Fatalf("unexpected teardown result when RuntimeStop is weak: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "stopped" {
		t.Fatalf("expected lifecycle-owned stopped teardown profile status, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeCoordinateTeardownBlockedByActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-blocked")
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry.Record("browser-runtime-coordinate-blocked", BrowserSessionRunInfo{
		RunID:    "run-88",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-coordinate-blocked", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRunRegistry:   sessionRunRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"coordinate","runtime_target":"node","coordination_goal":"teardown"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime coordinate teardown blocked: %v", err)
	}
	if len(nodeBackend.runtimeStopReqs) != 0 {
		t.Fatalf("expected no runtime stop when active node run exists, got %#v", nodeBackend.runtimeStopReqs)
	}
	var payload struct {
		CoordinationDecision string `json:"coordination_decision"`
		CoordinationReady    bool   `json:"coordination_ready"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.CoordinationDecision != "teardown_blocked_active_node_run" || payload.CoordinationReady {
		t.Fatalf("unexpected blocked teardown coordination result: %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeCoordinateRestart(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-restart")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-coordinate-restart", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		Arguments: `{"action":"coordinate","runtime_target":"node","profile":"isolated","coordination_goal":"restart"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime coordinate restart: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("unexpected runtime coordinate restart lifecycle requests: status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
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
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.CoordinationGoal != "restart" {
		t.Fatalf("unexpected runtime coordinate restart payload: %#v", payload)
	}
	if payload.RestartDecision != "restarted" || !payload.RestartReady || payload.CoordinationDecision != "restart_ready" || !payload.CoordinationReady {
		t.Fatalf("unexpected coordinate restart result: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "started" {
		t.Fatalf("unexpected coordinate restart profile status: %#v", payload.ProfileStatus)
	}
}
