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

func TestRegisterBrowserTools_RuntimeRestartStartFailureUsesStoppedLifecycleStateAfterWeakStop(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-restart-start-failed-after-stop")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-restart-start-failed-after-stop", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
			},
			runtimeStopResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
			},
			runtimeStartErr: fmt.Errorf("start failed"),
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
		Arguments: `{"action":"restart","runtime_target":"node","profile":"isolated","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart start failure after weak stop: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("expected restart recovery to stop then fail start, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Status          string `json:"status"`
		RestartDecision string `json:"restart_decision"`
		RestartReady    bool   `json:"restart_ready"`
		ProfileStatus   struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "error" || payload.RestartDecision != "restart_start_failed" || payload.RestartReady {
		t.Fatalf("unexpected restart failure payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "stopped" {
		t.Fatalf("expected restart failure to keep lifecycle-owned stopped status, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeRestartStopFailureUsesEffectiveLifecycleState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-restart-stop-failed")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-restart-stop-failed", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		Arguments: `{"action":"restart","runtime_target":"node","profile":"isolated","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart stop failure: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 0 {
		t.Fatalf("expected restart recovery to fail at stop before start, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Status          string `json:"status"`
		RestartDecision string `json:"restart_decision"`
		RestartReady    bool   `json:"restart_ready"`
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
	if payload.Status != "error" || payload.RestartDecision != "restart_stop_failed" || payload.RestartReady {
		t.Fatalf("unexpected restart stop failure payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "disconnected" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected restart stop failure payload to keep effective disconnected lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeRestartStatusFailureUsesEffectiveLifecycleState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-restart-status-failed")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-restart-status-failed", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		Arguments: `{"action":"restart","runtime_target":"node","profile":"isolated","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime restart status failure: %v", err)
	}
	var payload struct {
		Status          string `json:"status"`
		RestartDecision string `json:"restart_decision"`
		RestartReady    bool   `json:"restart_ready"`
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
	if payload.Status != "error" || payload.RestartDecision != "restart_status_failed" || payload.RestartReady {
		t.Fatalf("unexpected restart status failure payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "reconnecting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected restart status failure payload to keep effective lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeStop(t *testing.T) {
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
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"stop","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime stop: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "isolated" {
		t.Fatalf("unexpected runtime stop status requests: %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStopReqs) != 1 || nodeBackend.runtimeStopReqs[0].Profile != "isolated" {
		t.Fatalf("unexpected runtime stop requests: %#v", nodeBackend.runtimeStopReqs)
	}
	if !strings.Contains(out, `"action":"stop"`) || !strings.Contains(out, `"status":"ok"`) || !strings.Contains(out, `"stop_decision":"stopped"`) || !strings.Contains(out, `"profile":"isolated"`) || !strings.Contains(out, `"status":"stopped"`) {
		t.Fatalf("unexpected runtime stop output: %s", out)
	}
}

func TestRegisterBrowserTools_RuntimeStopUsesStoredStoppedStateWhenRawStatusIsWeak(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-stop-stored-stopped")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-stop-stored-stopped", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		Arguments: `{"action":"stop","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime stop stored stopped state: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 0 {
		t.Fatalf("expected stop to trust stored stopped state without issuing RuntimeStop, got status=%#v stop=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs)
	}
	if !strings.Contains(out, `"stop_decision":"stop_already_stopped"`) || !strings.Contains(out, `"status":"stopped"`) {
		t.Fatalf("unexpected runtime stop stored-stopped output: %s", out)
	}
}

func TestRegisterBrowserTools_RuntimeStopStatusFailureUsesEffectiveLifecycleState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-stop-status-failed")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-stop-status-failed", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		Arguments: `{"action":"stop","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime stop status failure: %v", err)
	}
	var payload struct {
		Status        string `json:"status"`
		StopDecision  string `json:"stop_decision"`
		StopReady     bool   `json:"stop_ready"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "error" || payload.StopDecision != "stop_status_failed" || payload.StopReady {
		t.Fatalf("unexpected stop status failure payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "reconnecting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected stop status failure payload to keep effective lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeStopFailureUsesEffectiveLifecycleState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-stop-failed")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-stop-failed", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		Arguments: `{"action":"stop","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime stop failure: %v", err)
	}
	var payload struct {
		Status        string `json:"status"`
		StopDecision  string `json:"stop_decision"`
		StopReady     bool   `json:"stop_ready"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "error" || payload.StopDecision != "stop_failed" || payload.StopReady {
		t.Fatalf("unexpected stop failure payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("expected stop failure payload to keep effective lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeStopWeakStopResultUsesLifecycleStoppedState(t *testing.T) {
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
		Arguments: `{"action":"stop","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime stop weak stop result: %v", err)
	}
	var payload struct {
		StopDecision  string `json:"stop_decision"`
		StopReady     bool   `json:"stop_ready"`
		ProfileStatus struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.StopDecision != "stopped" || !payload.StopReady {
		t.Fatalf("unexpected stop result when RuntimeStop is weak: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "stopped" {
		t.Fatalf("expected lifecycle-owned stopped profile status, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimeStopBlockedByActiveNodeRun(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-stop-blocked")
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-stop-blocked", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-93",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:               t.TempDir(),
		Backend:            &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:        nodeBackend,
		SessionRunRegistry: sessionRunRegistry,
		EnabledTools:       []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"stop","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime stop blocked: %v", err)
	}
	if len(nodeBackend.runtimeStopReqs) != 0 || len(nodeBackend.runtimeStatusReqs) != 0 {
		t.Fatalf("expected blocked stop to avoid runtime lifecycle calls, got status=%#v stop=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs)
	}
	if !strings.Contains(out, `"stop_decision":"stop_blocked_active_node_run"`) {
		t.Fatalf("unexpected runtime stop-blocked output: %s", out)
	}
}

func TestRegisterBrowserTools_RuntimePrepareStartFailureUsesStoppedLifecycleStateAfterWeakStop(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-start-failed-after-stop")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-start-failed-after-stop", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
			},
			runtimeStopResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
			},
			runtimeStartErr: fmt.Errorf("start failed"),
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
		t.Fatalf("browser_runtime prepare start failure after weak stop: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("expected prepare recovery to stop then fail start, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Status          string `json:"status"`
		PrepareDecision string `json:"prepare_decision"`
		PrepareReady    bool   `json:"prepare_ready"`
		ProfileStatus   struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Status != "error" || payload.PrepareDecision != "restart_start_failed" || payload.PrepareReady {
		t.Fatalf("unexpected prepare failure payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "stopped" {
		t.Fatalf("expected prepare failure to keep lifecycle-owned stopped status, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimePrepareStopFailureUsesEffectiveLifecycleState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-stop-failed")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-stop-failed", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		Arguments: `{"action":"prepare","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare stop failure: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 0 {
		t.Fatalf("expected prepare recovery to fail at stop before start, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Status          string `json:"status"`
		PrepareDecision string `json:"prepare_decision"`
		PrepareReady    bool   `json:"prepare_ready"`
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
	if payload.Status != "error" || payload.PrepareDecision != "restart_stop_failed" || payload.PrepareReady {
		t.Fatalf("unexpected prepare stop failure payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "disconnected" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected prepare stop failure payload to keep effective disconnected lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimePrepareStatusFailureUsesEffectiveLifecycleState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-status-failed")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-status-failed", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		Arguments: `{"action":"prepare","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare status failure: %v", err)
	}
	var payload struct {
		Status          string `json:"status"`
		PrepareDecision string `json:"prepare_decision"`
		PrepareReady    bool   `json:"prepare_ready"`
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
	if payload.Status != "error" || payload.PrepareDecision != "status_failed" || payload.PrepareReady {
		t.Fatalf("unexpected prepare status failure payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "reconnecting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected prepare status failure payload to keep effective lifecycle state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimePrepareStartFailureUsesEffectiveLifecycleState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-start-failed")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-start-failed", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "stopped",
		Running:       false,
		Connected:     false,
		Note:          "runtime stopped",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
			},
			runtimeStartErr: fmt.Errorf("start failed"),
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
		t.Fatalf("browser_runtime prepare start failure: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 0 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("expected prepare to attempt start once without stop, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Status          string `json:"status"`
		PrepareDecision string `json:"prepare_decision"`
		PrepareReady    bool   `json:"prepare_ready"`
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
	if payload.Status != "error" || payload.PrepareDecision != "start_failed" || payload.PrepareReady {
		t.Fatalf("unexpected prepare start failure payload: %#v", payload)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "stopped" || payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected prepare start failure payload to keep effective lifecycle state, got %#v", payload.ProfileStatus)
	}
}
