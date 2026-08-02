package tools

import (
	"context"
	"encoding/json"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_RuntimePrepare(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-session")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
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
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"prepare","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare: %v", err)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 0 {
		t.Fatalf("unexpected runtime profiles requests: %#v", nodeBackend.runtimeProfilesReqs)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "isolated" {
		t.Fatalf("unexpected runtime status requests: %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeStartReqs) != 1 || nodeBackend.runtimeStartReqs[0].Profile != "isolated" {
		t.Fatalf("unexpected runtime start requests: %#v", nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		Action          string   `json:"action"`
		Status          string   `json:"status"`
		PreparedProfile string   `json:"prepared_profile"`
		PrepareDecision string   `json:"prepare_decision"`
		PrepareReady    bool     `json:"prepare_ready"`
		RuntimeActions  []string `json:"runtime_actions"`
		SelectedRoute   struct {
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
		SessionBinding struct {
			SessionKey           string `json:"session_key"`
			BrowserProfileCount  int    `json:"browser_profile_count"`
			ActiveBrowserProfile string `json:"active_browser_profile"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "prepare" || payload.Status != "ok" || payload.PreparedProfile != "isolated" || payload.PrepareDecision != "started" || !payload.PrepareReady {
		t.Fatalf("unexpected runtime prepare payload: %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "isolated" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected node prepare route to use the promoted managed default profile, got %#v", payload.SelectedRoute)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "prepare") {
		t.Fatalf("expected runtime actions to include prepare, got %#v", payload.RuntimeActions)
	}
	if payload.ProfileStatus.Profile != "isolated" || payload.ProfileStatus.Status != "started" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected runtime prepare profile status: %#v", payload.ProfileStatus)
	}
	if payload.LaunchDiagnostics.Source != "runtime_status" ||
		payload.LaunchDiagnostics.Profile != "isolated" ||
		payload.LaunchDiagnostics.RuntimeTarget != "node" ||
		payload.LaunchDiagnostics.Status != "started" ||
		payload.LaunchDiagnostics.NodeVersion != "24.2.0" ||
		payload.LaunchDiagnostics.SelectedLaunchSource != "runtime_observed" ||
		payload.LaunchDiagnostics.SelectedLaunchReady == nil || !*payload.LaunchDiagnostics.SelectedLaunchReady ||
		payload.LaunchDiagnostics.LaunchReady == nil || !*payload.LaunchDiagnostics.LaunchReady {
		t.Fatalf("unexpected runtime prepare diagnostics payload: %#v", payload.LaunchDiagnostics)
	}
	if payload.SessionBinding.SessionKey != "browser-runtime-prepare-session" {
		t.Fatalf("unexpected runtime prepare session binding: %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimePrepareStartedButNotYetConnectedIsNotReady(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-starting")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
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
		t.Fatalf("browser_runtime prepare starting: %v", err)
	}
	var payload struct {
		PrepareDecision string `json:"prepare_decision"`
		PrepareReady    bool   `json:"prepare_ready"`
		ProfileStatus   struct {
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.PrepareDecision != "started" || payload.PrepareReady {
		t.Fatalf("expected prepare started-with-weak-observation to remain not ready, got %#v", payload)
	}
	if payload.ProfileStatus.Status != "starting" || payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected prepare payload to expose lifecycle-owned starting state, got %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_RuntimePrepareRecoversDisconnectedProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-disconnected")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-disconnected", agentxbrowserruntime.SharedSessionBrowserProfileState{
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

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"prepare","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare disconnected profile: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 1 || len(nodeBackend.runtimeStartReqs) != 1 {
		t.Fatalf("expected prepare recovery to stop/start disconnected profile, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		PrepareDecision string `json:"prepare_decision"`
		ProfileStatus   struct {
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
		SessionBinding struct {
			SessionHealthState string `json:"session_health_state"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.PrepareDecision != "restarted" && payload.PrepareDecision != "restart_started" {
		t.Fatalf("expected prepare recovery decision to reuse restart semantics, got %#v", payload)
	}
	if payload.ProfileStatus.Status != "reconnecting" || !payload.ProfileStatus.Running || payload.ProfileStatus.Connected {
		t.Fatalf("expected prepare recovery to advance profile status, got %#v", payload.ProfileStatus)
	}
	if payload.SessionBinding.SessionHealthState != "profile_reconnecting" {
		t.Fatalf("expected prepare recovery to transition session health to reconnecting, got %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimePrepareIgnoresUnrelatedDisconnectedProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-ignore-unrelated-disconnected")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-ignore-unrelated-disconnected", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-ignore-unrelated-disconnected", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "relay",
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
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"prepare","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare unrelated disconnected profile: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 0 || len(nodeBackend.runtimeStartReqs) != 0 {
		t.Fatalf("expected prepare for healthy profile to ignore unrelated disconnected route peer, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		PrepareDecision string `json:"prepare_decision"`
		PrepareReady    bool   `json:"prepare_ready"`
		SessionBinding  struct {
			SessionHealthState          string `json:"session_health_state"`
			SessionHealthRecoveryAction string `json:"session_health_recovery_action"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.PrepareDecision != "already_ready" || !payload.PrepareReady {
		t.Fatalf("expected prepare to remain already_ready for healthy requested profile, got %#v", payload)
	}
	if payload.SessionBinding.SessionHealthState != "healthy" || payload.SessionBinding.SessionHealthRecoveryAction != "" {
		t.Fatalf("expected session health to follow requested profile scope, got %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimePrepareUsesStoredHealthyStateWhenRawStatusIsWeak(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-healthy-stored-state")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-healthy-stored-state", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		t.Fatalf("browser_runtime prepare with stored healthy state: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 0 || len(nodeBackend.runtimeStartReqs) != 0 {
		t.Fatalf("expected prepare to trust stored healthy state without lifecycle churn, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
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
			SessionHealthState string `json:"session_health_state"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.PrepareDecision != "already_ready" || !payload.PrepareReady {
		t.Fatalf("expected prepare to stay already_ready when registry-owned lifecycle state is healthy, got %#v", payload)
	}
	if payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("expected prepare payload to use stored healthy state, got %#v", payload.ProfileStatus)
	}
	if payload.SessionBinding.SessionHealthState != "healthy" {
		t.Fatalf("expected session health to remain healthy, got %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_RuntimePrepareExplicitHealthyRawStatusOverridesStoredDisconnectedState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-healthy-raw-overrides-disconnected")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-prepare-healthy-raw-overrides-disconnected", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       true,
		Connected:     false,
		Note:          "stale cdp disconnect",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "connected",
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
		Arguments: `{"action":"prepare","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare healthy raw overrides disconnected: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || len(nodeBackend.runtimeStopReqs) != 0 || len(nodeBackend.runtimeStartReqs) != 0 {
		t.Fatalf("expected explicit healthy raw status to avoid restart churn, got status=%#v stop=%#v start=%#v", nodeBackend.runtimeStatusReqs, nodeBackend.runtimeStopReqs, nodeBackend.runtimeStartReqs)
	}
	var payload struct {
		PrepareDecision string `json:"prepare_decision"`
		PrepareReady    bool   `json:"prepare_ready"`
		ProfileStatus   struct {
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.PrepareDecision != "already_ready" || !payload.PrepareReady {
		t.Fatalf("expected explicit healthy raw status to keep prepare already_ready, got %#v", payload)
	}
	if payload.ProfileStatus.Status != "connected" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("expected prepare to surface explicit healthy raw status, got %#v", payload.ProfileStatus)
	}
}
