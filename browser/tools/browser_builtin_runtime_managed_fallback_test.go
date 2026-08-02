package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_RuntimeCoordinateRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-coordinate-managed-current-hidden-implicit-host")
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

	sessionRegistry.TrackCurrentTarget("browser-runtime-coordinate-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"coordinate","coordination_goal":"restart"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime coordinate managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime coordinate to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime coordinate to reuse managed current route stop profile before implicit host fallback, got %#v", nodeBackend.runtimeStopReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime coordinate to reuse managed current route start profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
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
		t.Fatalf("decode managed current runtime coordinate output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.CoordinationGoal != "restart" || payload.CoordinationDecision != "restart_ready" || !payload.CoordinationReady {
		t.Fatalf("unexpected managed current runtime coordinate payload: %#v", payload)
	}
	if payload.RestartDecision != "restarted" || !payload.RestartReady {
		t.Fatalf("unexpected managed current runtime coordinate restart result: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime coordinate to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime coordinate to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected managed current runtime coordinate profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimePrepareRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-managed-current-hidden-implicit-host")
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
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-prepare-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"prepare"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime prepare to reuse managed current route profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime prepare to start managed current route profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		PreparedProfile     string `json:"prepared_profile"`
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
		t.Fatalf("decode managed current runtime prepare output: %v", err)
	}
	if payload.Action != "prepare" || payload.Status != "ok" || payload.PreparedProfile != "workbench" {
		t.Fatalf("unexpected managed current runtime prepare payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime prepare to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime prepare to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected managed current runtime prepare profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeSelectProfileRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-profile-managed-current-hidden-implicit-host")
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
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-select-profile-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"select_profile","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime select_profile managed current hidden implicit host: %v", err)
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
		t.Fatalf("decode managed current runtime select_profile output: %v", err)
	}
	if payload.Action != "select_profile" || payload.Status != "ok" || payload.SelectDecision != "session_profile_selected" || !payload.SelectReady {
		t.Fatalf("unexpected managed current runtime select_profile payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime select_profile to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime select_profile to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.RuntimeTarget != "node" || payload.SessionProfileSelection.Source != "select_profile" {
		t.Fatalf("unexpected session profile selection payload: %#v", payload.SessionProfileSelection)
	}
	if payload.SessionProfileSelection.BrowserApp != "Chromium" {
		t.Fatalf("expected session profile selection to inherit managed current browser_app before implicit host fallback, got %#v", payload.SessionProfileSelection)
	}
}

func TestRegisterBrowserTools_RuntimeClearSessionRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-session-managed-current-hidden-implicit-host")
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
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-clear-session-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"clear_session"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime clear_session managed current hidden implicit host: %v", err)
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
		t.Fatalf("decode managed current runtime clear_session output: %v", err)
	}
	if payload.Action != "clear_session" || payload.Status != "ok" || payload.ClearSessionDecision == "" || !payload.ClearSessionReady {
		t.Fatalf("unexpected managed current runtime clear_session payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime clear_session to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime clear_session to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
}

func TestRegisterBrowserTools_RuntimeStartRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-start-managed-current-hidden-implicit-host")
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
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-start-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"start"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime start managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime start to reuse managed current route profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
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
		t.Fatalf("decode managed current runtime start output: %v", err)
	}
	if payload.Action != "start" || payload.Status != "ok" {
		t.Fatalf("unexpected managed current runtime start payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime start to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime start to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected managed current runtime start profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeStopRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-stop-managed-current-hidden-implicit-host")
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
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-stop-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"stop"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime stop managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime stop to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime stop to reuse managed current route profile before implicit host fallback, got %#v", nodeBackend.runtimeStopReqs)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		StopDecision        string `json:"stop_decision"`
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
		t.Fatalf("decode managed current runtime stop output: %v", err)
	}
	if payload.Action != "stop" || payload.Status != "ok" || payload.StopDecision == "" {
		t.Fatalf("unexpected managed current runtime stop payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime stop to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime stop to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "stopped" {
		t.Fatalf("unexpected managed current runtime stop profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeRestartRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-restart-managed-current-hidden-implicit-host")
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

	sessionRegistry.TrackCurrentTarget("browser-runtime-restart-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"restart"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime restart to reuse managed current route status profile before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime restart to reuse managed current route stop profile before implicit host fallback, got %#v", nodeBackend.runtimeStopReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime restart to reuse managed current route start profile before implicit host fallback, got %#v", nodeBackend.runtimeStartReqs)
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
		t.Fatalf("decode managed current runtime restart output: %v", err)
	}
	if payload.Action != "restart" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.RestartDecision != "restarted" || !payload.RestartReady {
		t.Fatalf("unexpected managed current runtime restart payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime restart to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime restart to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected managed current runtime restart profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeCreateProfileRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-create-profile-managed-current-hidden-implicit-host")
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
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-create-profile-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"create_profile","profile":"draft","color":"#ff5500","copy_from":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime create_profile managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeCreateReqs) != 1 {
		t.Fatalf("expected runtime create_profile to reuse managed current route before implicit host fallback, got %#v", nodeBackend.runtimeCreateReqs)
	}
	if nodeBackend.runtimeCreateReqs[0].Profile != "draft" || nodeBackend.runtimeCreateReqs[0].BrowserApp != "Chromium" || nodeBackend.runtimeCreateReqs[0].Color != "#ff5500" || nodeBackend.runtimeCreateReqs[0].CopyFrom != "isolated" {
		t.Fatalf("unexpected managed current runtime create request payload: %#v", nodeBackend.runtimeCreateReqs[0])
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
		t.Fatalf("decode managed current runtime create_profile output: %v", err)
	}
	if payload.Action != "create_profile" || payload.Status != "ok" || payload.PreparedProfile != "draft" || payload.CreateDecision != "created" || !payload.CreateReady {
		t.Fatalf("unexpected managed current runtime create_profile payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" || payload.RequestedColor != "#ff5500" || payload.RequestedCopyFrom != "isolated" {
		t.Fatalf("unexpected managed current runtime create_profile request echo: %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "draft" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime create_profile to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "draft" || payload.ProfileStatus.BrowserApp != "Chromium" || payload.ProfileStatus.Status != "created" {
		t.Fatalf("unexpected managed current runtime create_profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeDeleteProfileRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-delete-profile-managed-current-hidden-implicit-host")
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
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-delete-profile-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"delete_profile","profile":"draft","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime delete_profile managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeDeleteReqs) != 1 {
		t.Fatalf("expected runtime delete_profile to reuse managed current route before implicit host fallback, got %#v", nodeBackend.runtimeDeleteReqs)
	}
	if nodeBackend.runtimeDeleteReqs[0].Profile != "draft" || !nodeBackend.runtimeDeleteReqs[0].Force {
		t.Fatalf("unexpected managed current runtime delete request payload: %#v", nodeBackend.runtimeDeleteReqs[0])
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
		t.Fatalf("decode managed current runtime delete_profile output: %v", err)
	}
	if payload.Action != "delete_profile" || payload.Status != "ok" || payload.PreparedProfile != "draft" || payload.DeleteDecision != "deleted" || !payload.DeleteReady || !payload.Force {
		t.Fatalf("unexpected managed current runtime delete_profile payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime delete_profile to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "draft" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime delete_profile to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "draft" || payload.ProfileStatus.Status != "deleted" {
		t.Fatalf("unexpected managed current runtime delete_profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeStatusRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-managed-current-hidden-implicit-host")
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
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-status-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime status to route through managed current route before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
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
		t.Fatalf("decode managed current runtime status output: %v", err)
	}
	if payload.Action != "status" || payload.Status != "ok" {
		t.Fatalf("unexpected managed current runtime status payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime status to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime status to select managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected managed current runtime status profile payload: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeStatusRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-managed-default-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
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
				runtimeInfo:   BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
				routeSource:   "managed_browserd",
				routeEndpoint: "http://127.0.0.1:43123",
			}},
			capabilities: fullBrowserCapabilities(),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			if strings.EqualFold(strings.TrimSpace(requested.Profile), "isolated") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status managed default hidden implicit host: %v", err)
	}
	var payload struct {
		Action                     string   `json:"action"`
		Status                     string   `json:"status"`
		Note                       string   `json:"note"`
		SubstrateSelectionStrategy string   `json:"substrate_selection_strategy"`
		SubstrateSelectionReason   string   `json:"substrate_selection_reason"`
		ConfiguredProfiles         []string `json:"configured_profiles"`
		DefaultRoute               struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Source        string `json:"source"`
			Endpoint      string `json:"endpoint"`
		} `json:"default_route"`
		SelectedRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Source        string `json:"source"`
			Endpoint      string `json:"endpoint"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default runtime status output: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime status to route through managed default route before implicit host fallback, got reqs=%#v payload=%#v", nodeBackend.runtimeStatusReqs, payload)
	}
	if payload.Action != "status" || payload.Status != "ok" {
		t.Fatalf("unexpected managed default runtime status payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime status payload to refresh default_route from dynamic managed-default route, got %#v", payload.DefaultRoute)
	}
	if payload.DefaultRoute.Source != "managed_browserd" || payload.DefaultRoute.Endpoint != "http://127.0.0.1:43123" {
		t.Fatalf("expected runtime status payload to preserve default_route provenance, got %#v", payload.DefaultRoute)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy || payload.SubstrateSelectionReason == "" {
		t.Fatalf("expected runtime status payload to refresh substrate summary from dynamic managed-default route, got strategy=%q reason=%q", payload.SubstrateSelectionStrategy, payload.SubstrateSelectionReason)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime status to select managed default route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SelectedRoute.Source != "managed_browserd" || payload.SelectedRoute.Endpoint != "http://127.0.0.1:43123" {
		t.Fatalf("expected runtime status to preserve selected_route provenance, got %#v", payload.SelectedRoute)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected runtime status configured profiles to retain managed default route profile, got %#v", payload.ConfiguredProfiles)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected managed default runtime status profile payload: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeStatusRoutesManagedRuntimeDefaultRouteWhenOnlyRuntimeToolEnabled(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-status-managed-runtime-default-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
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
			}},
			capabilities: BrowserCapabilities{RuntimeStatus: true},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			if strings.EqualFold(strings.TrimSpace(requested.Profile), "isolated") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status managed runtime default hidden implicit host: %v", err)
	}
	var payload struct {
		Action                     string `json:"action"`
		Status                     string `json:"status"`
		SubstrateSelectionStrategy string `json:"substrate_selection_strategy"`
		DefaultRoute               struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute struct {
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
		t.Fatalf("decode managed runtime default runtime status output: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime-only status to route through managed runtime default before implicit host fallback, got reqs=%#v payload=%#v", nodeBackend.runtimeStatusReqs, payload)
	}
	if payload.Action != "status" || payload.Status != "ok" {
		t.Fatalf("unexpected managed runtime default status payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "workbench" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime-only status payload to refresh default_route from managed runtime default, got %#v", payload.DefaultRoute)
	}
	if payload.SubstrateSelectionStrategy != BrowserSubstrateSelectionPreferNodeOverLegacy {
		t.Fatalf("expected runtime-only status payload to advertise managed runtime default substrate, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime-only status to select managed runtime default before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected managed runtime default profile payload: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profiles-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
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

	sessionRegistry.TrackCurrentTarget("browser-runtime-profiles-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"profiles"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 1 || nodeBackend.runtimeProfilesReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime profiles to route through managed current route before implicit host fallback, got %#v", nodeBackend.runtimeProfilesReqs)
	}
	var payload struct {
		Action              string   `json:"action"`
		Status              string   `json:"status"`
		RequestedBrowserApp string   `json:"requested_browser_app"`
		DefaultProfile      string   `json:"default_profile"`
		ConfiguredProfiles  []string `json:"configured_profiles"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		Profiles []struct {
			Profile    string `json:"profile"`
			BrowserApp string `json:"browser_app"`
			Status     string `json:"status"`
			Running    bool   `json:"running"`
			Connected  bool   `json:"connected"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed current runtime profiles output: %v", err)
	}
	if payload.Action != "profiles" || payload.Status != "ok" {
		t.Fatalf("unexpected managed current runtime profiles payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime profiles to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected runtime profiles to expose managed current default profile, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime profiles to select managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected runtime profiles configured profiles to retain managed current route profile, got %#v", payload.ConfiguredProfiles)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Profile != "workbench" || payload.Profiles[0].BrowserApp != "Chromium" || payload.Profiles[0].Status != "running" || !payload.Profiles[0].Running || !payload.Profiles[0].Connected {
		t.Fatalf("unexpected managed current runtime profiles payload: %#v", payload.Profiles)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profiles-managed-default-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
				fakeBrowserBackend: &fakeBrowserBackend{
					runtimeProfilesResult: BrowserProfilesResult{
						Backend:        "proxy",
						DefaultProfile: "workbench",
						Profiles: []BrowserProfileInfo{
							{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
						},
					},
				},
				runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			}},
			capabilities: fullBrowserCapabilities(),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			if strings.EqualFold(strings.TrimSpace(requested.Profile), "isolated") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"profiles"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 1 || nodeBackend.runtimeProfilesReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime profiles to route through managed default route before implicit host fallback, got %#v", nodeBackend.runtimeProfilesReqs)
	}
	var payload struct {
		Action             string   `json:"action"`
		Status             string   `json:"status"`
		DefaultProfile     string   `json:"default_profile"`
		ConfiguredProfiles []string `json:"configured_profiles"`
		SelectedRoute      struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		Profiles []struct {
			Profile    string `json:"profile"`
			BrowserApp string `json:"browser_app"`
			Status     string `json:"status"`
			Running    bool   `json:"running"`
			Connected  bool   `json:"connected"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default runtime profiles output: %v", err)
	}
	if payload.Action != "profiles" || payload.Status != "ok" {
		t.Fatalf("unexpected managed default runtime profiles payload: %#v", payload)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected runtime profiles to expose managed default profile, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime profiles to select managed default route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected runtime profiles configured profiles to retain managed default route profile, got %#v", payload.ConfiguredProfiles)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Profile != "workbench" || payload.Profiles[0].BrowserApp != "Chromium" || payload.Profiles[0].Status != "running" || !payload.Profiles[0].Running || !payload.Profiles[0].Connected {
		t.Fatalf("unexpected managed default runtime profiles payload: %#v", payload.Profiles)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sessions-managed-current-hidden-implicit-host")
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
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
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

	tracked := sessionRegistry.TrackCurrentTarget("browser-runtime-sessions-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions managed current hidden implicit host: %v", err)
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
		SessionTargetCount int `json:"session_target_count"`
		SessionBinding     struct {
			CurrentTargetID string `json:"current_target_id"`
		} `json:"session_binding"`
		SessionRoutes []struct {
			Backend         string `json:"backend"`
			Profile         string `json:"profile"`
			RuntimeTarget   string `json:"runtime_target"`
			CurrentTargetID string `json:"current_target_id"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed current runtime sessions output: %v", err)
	}
	if payload.Action != "sessions" || payload.Status != "ok" {
		t.Fatalf("unexpected managed current runtime sessions payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime sessions to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime sessions to select managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SessionTargetCount != 1 || payload.SessionBinding.CurrentTargetID != tracked.ID {
		t.Fatalf("unexpected managed current runtime sessions binding payload: %#v", payload)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].Backend != "proxy" || payload.SessionRoutes[0].Profile != "workbench" || payload.SessionRoutes[0].RuntimeTarget != "node" || payload.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("unexpected managed current runtime sessions routes payload: %#v", payload.SessionRoutes)
	}
}

func TestRegisterBrowserTools_RuntimeWorkbenchRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench-managed-current-hidden-implicit-host")
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
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
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

	tracked := sessionRegistry.TrackCurrentTarget("browser-runtime-workbench-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime workbench to route through managed current route status before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 1 || nodeBackend.runtimeProfilesReqs[0].Profile != "workbench" {
		t.Fatalf("expected runtime workbench to route through managed current route profiles before implicit host fallback, got %#v", nodeBackend.runtimeProfilesReqs)
	}
	var payload struct {
		Action              string   `json:"action"`
		Status              string   `json:"status"`
		RequestedBrowserApp string   `json:"requested_browser_app"`
		WorkbenchReady      bool     `json:"workbench_ready"`
		WorkbenchSections   []string `json:"workbench_sections"`
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
		SessionTargetCount int `json:"session_target_count"`
		SessionBinding     struct {
			CurrentTargetID string `json:"current_target_id"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed current runtime workbench output: %v", err)
	}
	if payload.Action != "workbench" || payload.Status != "ok" || !payload.WorkbenchReady {
		t.Fatalf("unexpected managed current runtime workbench payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime workbench to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime workbench to select managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	for _, want := range []string{"route", "status", "profiles", "sessions"} {
		if !browserStringSliceContains(payload.WorkbenchSections, want) {
			t.Fatalf("expected runtime workbench to retain %q section on managed route, got %#v", want, payload.WorkbenchSections)
		}
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected managed current runtime workbench profile payload: %#v", payload.ProfileStatus)
	}
	if payload.SessionTargetCount != 1 || payload.SessionBinding.CurrentTargetID != tracked.ID {
		t.Fatalf("unexpected managed current runtime workbench session binding payload: %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimeSyncSessionRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sync-session-managed-current-hidden-implicit-host")
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
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-runtime-sync-session-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"sync_session"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sync_session managed current hidden implicit host: %v", err)
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
		t.Fatalf("decode managed current runtime sync_session output: %v", err)
	}
	if payload.Action != "sync_session" || payload.Status != "ok" || payload.SyncSessionDecision != "session_route_synced" || !payload.SyncSessionReady {
		t.Fatalf("unexpected managed current runtime sync_session payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime sync_session to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime sync_session to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.RuntimeTarget != "node" || payload.SessionProfileSelection.BrowserApp != "Chromium" || payload.SessionProfileSelection.Source != "sync_session" {
		t.Fatalf("unexpected managed current runtime sync_session profile selection: %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection.ID != tracked.ID || payload.SessionTargetSelection.TabIndex != 2 || payload.SessionTargetSelection.RuntimeTarget != "node" || payload.SessionTargetSelection.Profile != "workbench" || payload.SessionTargetSelection.BrowserApp != "Chromium" || payload.SessionTargetSelection.Source != "sync_session" {
		t.Fatalf("unexpected managed current runtime sync_session target selection: %#v", payload.SessionTargetSelection)
	}
}

func TestRegisterBrowserTools_RuntimeSelectTargetRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-select-target-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-select-target-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   1,
		URL:        "https://node.example/one",
		Title:      "One",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	second := sessionRegistry.TrackTab("browser-runtime-select-target-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Name:      "browser_runtime",
		Arguments: `{"action":"select_target","target":"tab:2"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime select_target managed current hidden implicit host: %v", err)
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
		t.Fatalf("decode managed current runtime select_target output: %v", err)
	}
	if payload.Action != "select_target" || payload.Status != "ok" || payload.SelectTargetDecision != "session_target_selected" || !payload.SelectTargetReady {
		t.Fatalf("unexpected managed current runtime select_target payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected runtime select_target to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime select_target to route through managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SessionTargetSelection.ID != second.ID || payload.SessionTargetSelection.TabIndex != 2 || payload.SessionTargetSelection.RuntimeTarget != "node" || payload.SessionTargetSelection.Profile != "workbench" || payload.SessionTargetSelection.BrowserApp != "Chromium" || payload.SessionTargetSelection.Source != "select_target" {
		t.Fatalf("unexpected managed current runtime select_target selection: %#v", payload.SessionTargetSelection)
	}
}
