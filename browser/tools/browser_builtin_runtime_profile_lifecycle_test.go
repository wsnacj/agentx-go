package tools

import (
	"context"
	"encoding/json"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_RuntimeStart(t *testing.T) {
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
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"start","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime start: %v", err)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "isolated" {
		t.Fatalf("unexpected runtime start requests: %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action        string `json:"action"`
		Status        string `json:"status"`
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
	if payload.Action != "start" || payload.Status != "ok" {
		t.Fatalf("unexpected runtime start payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected runtime start profile status: %#v", payload.ProfileStatus)
	}
	if payload.LaunchDiagnostics.Source != "runtime_status" ||
		payload.LaunchDiagnostics.Profile != "isolated" ||
		payload.LaunchDiagnostics.RuntimeTarget != "node" ||
		payload.LaunchDiagnostics.Status != "started" ||
		payload.LaunchDiagnostics.NodeVersion != "24.2.0" ||
		payload.LaunchDiagnostics.SelectedLaunchSource != "runtime_observed" ||
		payload.LaunchDiagnostics.SelectedLaunchReady == nil || !*payload.LaunchDiagnostics.SelectedLaunchReady ||
		payload.LaunchDiagnostics.LaunchReady == nil || !*payload.LaunchDiagnostics.LaunchReady {
		t.Fatalf("unexpected runtime start diagnostics payload: %#v", payload.LaunchDiagnostics)
	}
}

func TestRegisterBrowserTools_RuntimeStartWeakStartResultUsesLifecycleState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStartResult: BrowserProfileStatusResult{
				Backend:   "proxy",
				Profile:   "isolated",
				Status:    "started",
				Running:   true,
				Connected: false,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionStateRegistry: NewBrowserSessionStateRegistry(),
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(WithToolSessionID(context.Background(), "browser-runtime-start-lifecycle"), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"start","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime start weak observation: %v", err)
	}
	var payload struct {
		Action        string `json:"action"`
		Status        string `json:"status"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
			Note      string `json:"note"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode weak start output: %v", err)
	}
	if payload.Action != "start" || payload.Status != "ok" {
		t.Fatalf("unexpected weak start payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "starting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected || payload.ProfileStatus.Note != "start requested" {
		t.Fatalf("expected runtime start to surface lifecycle-owned starting state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeCreateProfile(t *testing.T) {
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
		Arguments: `{"action":"create_profile","profile":"workbench","browser_app":"Chromium","color":"#ff5500","copy_from":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime create_profile: %v", err)
	}
	if len(nodeBackend.runtimeCreateReqs) != 1 {
		t.Fatalf("unexpected runtime create requests: %#v", nodeBackend.runtimeCreateReqs)
	}
	if nodeBackend.runtimeCreateReqs[0].Profile != "workbench" || nodeBackend.runtimeCreateReqs[0].BrowserApp != "Chromium" || nodeBackend.runtimeCreateReqs[0].Color != "#ff5500" || nodeBackend.runtimeCreateReqs[0].CopyFrom != "isolated" {
		t.Fatalf("unexpected runtime create request payload: %#v", nodeBackend.runtimeCreateReqs[0])
	}
	var payload struct {
		Action              string   `json:"action"`
		Status              string   `json:"status"`
		PreparedProfile     string   `json:"prepared_profile"`
		CreateDecision      string   `json:"create_decision"`
		CreateReady         bool     `json:"create_ready"`
		RequestedBrowserApp string   `json:"requested_browser_app"`
		RequestedColor      string   `json:"requested_color"`
		RequestedCopyFrom   string   `json:"requested_copy_from"`
		RuntimeActions      []string `json:"runtime_actions"`
		ProfileStatus       struct {
			Profile    string `json:"profile"`
			BrowserApp string `json:"browser_app"`
			Status     string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "create_profile" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.CreateDecision != "created" || !payload.CreateReady {
		t.Fatalf("unexpected runtime create_profile payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" || payload.RequestedColor != "#ff5500" || payload.RequestedCopyFrom != "isolated" {
		t.Fatalf("unexpected runtime create request echo: %#v", payload)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "create_profile") {
		t.Fatalf("expected runtime actions to include create_profile, got %#v", payload.RuntimeActions)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.BrowserApp != "Chromium" || payload.ProfileStatus.Status != "created" {
		t.Fatalf("unexpected runtime create profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeRestart(t *testing.T) {
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
		Arguments: `{"action":"restart","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart: %v", err)
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
		Action          string   `json:"action"`
		Status          string   `json:"status"`
		PreparedProfile string   `json:"prepared_profile"`
		RestartDecision string   `json:"restart_decision"`
		RestartReady    bool     `json:"restart_ready"`
		RuntimeActions  []string `json:"runtime_actions"`
		ProfileStatus   struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "restart" || payload.Status != "ok" || payload.PreparedProfile != "isolated" || payload.RestartDecision != "restarted" || !payload.RestartReady {
		t.Fatalf("unexpected runtime restart payload: %#v", payload)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "restart") {
		t.Fatalf("expected runtime actions to include restart, got %#v", payload.RuntimeActions)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected runtime restart profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeDeleteProfile(t *testing.T) {
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
		Arguments: `{"action":"delete_profile","profile":"workbench","runtime_target":"node","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime delete_profile: %v", err)
	}
	if len(nodeBackend.runtimeDeleteReqs) != 1 {
		t.Fatalf("unexpected runtime delete requests: %#v", nodeBackend.runtimeDeleteReqs)
	}
	if nodeBackend.runtimeDeleteReqs[0].Profile != "workbench" || !nodeBackend.runtimeDeleteReqs[0].Force {
		t.Fatalf("unexpected runtime delete request payload: %#v", nodeBackend.runtimeDeleteReqs[0])
	}
	var payload struct {
		Action          string   `json:"action"`
		Status          string   `json:"status"`
		PreparedProfile string   `json:"prepared_profile"`
		DeleteDecision  string   `json:"delete_decision"`
		DeleteReady     bool     `json:"delete_ready"`
		Force           bool     `json:"force"`
		RuntimeActions  []string `json:"runtime_actions"`
		ProfileStatus   struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "delete_profile" || payload.Status != "ok" || payload.PreparedProfile != "workbench" || payload.DeleteDecision != "deleted" || !payload.DeleteReady || !payload.Force {
		t.Fatalf("unexpected runtime delete_profile payload: %#v", payload)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "delete_profile") {
		t.Fatalf("expected runtime actions to include delete_profile, got %#v", payload.RuntimeActions)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "deleted" {
		t.Fatalf("unexpected runtime delete profile status: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeDeleteProfileBlockedByActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-delete-blocked", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-61",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeDeleteResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "deleted",
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

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-delete-blocked")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-delete-blocked", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"delete_profile","profile":"workbench","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime delete_profile blocked: %v", err)
	}
	if len(nodeBackend.runtimeDeleteReqs) != 0 {
		t.Fatalf("expected delete_profile to be blocked before backend delete, got %#v", nodeBackend.runtimeDeleteReqs)
	}
	var payload struct {
		Action         string `json:"action"`
		Status         string `json:"status"`
		DeleteDecision string `json:"delete_decision"`
		DeleteReady    bool   `json:"delete_ready"`
		Force          bool   `json:"force"`
		ProfileStatus  struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode delete blocked output: %v", err)
	}
	if payload.Action != "delete_profile" || payload.Status != "ok" || payload.DeleteDecision != "delete_profile_blocked_active_node_run" || payload.DeleteReady || payload.Force {
		t.Fatalf("unexpected blocked delete payload: %#v", payload)
	}
	if payload.ProfileStatus.Backend != "proxy" || payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.RuntimeTarget != "node" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("expected blocked delete_profile to preserve effective lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeDeleteProfileForceBypassesActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-delete-force", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-62",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeDeleteResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "deleted",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:               t.TempDir(),
		Backend:            &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:        nodeBackend,
		SessionRunRegistry: sessionRunRegistry,
		EnabledTools:       []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-delete-force")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"delete_profile","profile":"workbench","runtime_target":"node","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime delete_profile force: %v", err)
	}
	if len(nodeBackend.runtimeDeleteReqs) != 1 || nodeBackend.runtimeDeleteReqs[0].Profile != "workbench" || !nodeBackend.runtimeDeleteReqs[0].Force {
		t.Fatalf("expected forced delete_profile to reach backend delete, got %#v", nodeBackend.runtimeDeleteReqs)
	}
	var payload struct {
		Action         string `json:"action"`
		Status         string `json:"status"`
		DeleteDecision string `json:"delete_decision"`
		DeleteReady    bool   `json:"delete_ready"`
		Force          bool   `json:"force"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode delete force output: %v", err)
	}
	if payload.Action != "delete_profile" || payload.Status != "ok" || payload.DeleteDecision != "deleted" || !payload.DeleteReady || !payload.Force {
		t.Fatalf("unexpected forced delete payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeDeleteProfileClearsMatchingSelectedProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "system",
				BrowserApp: "Safari",
				Profile:    "default",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeDeleteResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "deleted",
			},
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "system",
				BrowserApp: "Safari",
				Profile:    "default",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-delete-clears-profile")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}

	deleteOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"delete_profile","runtime_target":"node","profile":"workbench","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime delete_profile: %v", err)
	}
	var deletePayload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		DeleteDecision          string `json:"delete_decision"`
		DeleteReady             bool   `json:"delete_ready"`
		SessionProfileSelection any    `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(deleteOut), &deletePayload); err != nil {
		t.Fatalf("decode delete output: %v", err)
	}
	if deletePayload.Action != "delete_profile" || deletePayload.Status != "ok" || deletePayload.DeleteDecision != "deleted" || !deletePayload.DeleteReady {
		t.Fatalf("unexpected delete payload: %#v", deletePayload)
	}
	if deletePayload.SessionProfileSelection != nil {
		t.Fatalf("expected matching delete_profile to clear session profile selection, got %#v", deletePayload.SessionProfileSelection)
	}

	statusOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status after delete_profile: %v", err)
	}
	if len(hostBackend.runtimeStatusReqs) != 0 {
		t.Fatalf("expected host backend status to stay unused after deleting selected profile, got %#v", hostBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "isolated" {
		t.Fatalf("expected default node route status after deleting selected profile, got %#v", nodeBackend.runtimeStatusReqs)
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
		t.Fatalf("expected status to fall back to promoted node route after deleting selected profile, got %#v", statusPayload.SelectedRoute)
	}
	if statusPayload.SessionProfileSelection != nil {
		t.Fatalf("expected session profile selection to stay cleared after deleting selected profile, got %#v", statusPayload.SessionProfileSelection)
	}
}

func TestRegisterBrowserTools_RuntimeDeleteProfilePreservesDifferentSelectedProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "system",
				BrowserApp: "Safari",
				Profile:    "default",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeDeleteResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "other",
				Status:     "deleted",
			},
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-delete-preserves-profile")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}

	deleteOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"delete_profile","runtime_target":"node","profile":"other","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime delete_profile other: %v", err)
	}
	var deletePayload struct {
		SessionProfileSelection struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(deleteOut), &deletePayload); err != nil {
		t.Fatalf("decode delete output: %v", err)
	}
	if deletePayload.SessionProfileSelection.Profile != "workbench" || deletePayload.SessionProfileSelection.RuntimeTarget != "node" {
		t.Fatalf("expected deleting another profile to preserve current selection, got %#v", deletePayload.SessionProfileSelection)
	}

	statusOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status after deleting other profile: %v", err)
	}
	if len(hostBackend.runtimeStatusReqs) != 0 {
		t.Fatalf("expected host backend status to stay unused, got %#v", hostBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected node backend status to keep using selected profile after deleting other profile, got %#v", nodeBackend.runtimeStatusReqs)
	}
	var statusPayload struct {
		SelectedRoute struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionProfileSelection struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"session_profile_selection"`
	}
	if err := json.Unmarshal([]byte(statusOut), &statusPayload); err != nil {
		t.Fatalf("decode status output: %v", err)
	}
	if statusPayload.SelectedRoute.Profile != "workbench" || statusPayload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected selected route to keep using remembered profile after deleting other profile, got %#v", statusPayload.SelectedRoute)
	}
	if statusPayload.SessionProfileSelection.Profile != "workbench" || statusPayload.SessionProfileSelection.RuntimeTarget != "node" {
		t.Fatalf("expected session profile selection to survive deleting other profile, got %#v", statusPayload.SessionProfileSelection)
	}
}
