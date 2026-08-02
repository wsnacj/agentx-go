package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_RuntimeProfiles(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
					{Profile: "relay", BrowserApp: "Chromium", Status: "stopped"},
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
		Arguments: `{"action":"profiles","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles: %v", err)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 1 {
		t.Fatalf("unexpected runtime profiles requests: %#v", nodeBackend.runtimeProfilesReqs)
	}
	if !strings.Contains(out, `"action":"profiles"`) || !strings.Contains(out, `"profiles":[`) || !strings.Contains(out, `"profile":"isolated"`) || !strings.Contains(out, `"profile":"relay"`) {
		t.Fatalf("unexpected runtime profiles output: %s", out)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesFilteredSyncPreservesOtherRouteProfiles(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profiles-filtered-sync")
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profiles-filtered-sync", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "relay",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       true,
		Connected:     false,
		Note:          "stale relay state",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profiles-filtered-sync", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})

	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
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

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"profiles","runtime_target":"node"}`,
	}); err != nil {
		t.Fatalf("browser_runtime profiles filtered sync: %v", err)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"profiles","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles filtered sync output: %v", err)
	}
	if !strings.Contains(out, `"action":"profiles"`) || !strings.Contains(out, `"profile":"isolated"`) || !strings.Contains(out, `"profile":"relay"`) {
		t.Fatalf("expected session-backed profiles output to retain full route inventory, got %s", out)
	}

	snapshot := sessionStateRegistry.SnapshotSessionBrowserProfiles("browser-runtime-profiles-filtered-sync")
	if len(snapshot) != 3 {
		t.Fatalf("expected filtered profiles sync to preserve other route-profile state, got %#v", snapshot)
	}
	if snapshot[0].Backend != "proxy" || snapshot[0].Profile != "isolated" || snapshot[1].Backend != "proxy" || snapshot[1].Profile != "relay" || snapshot[2].Backend != "system" || snapshot[2].Profile != "default" {
		t.Fatalf("unexpected filtered profiles sync snapshot: %#v", snapshot)
	}
}

func TestRegisterBrowserTools_RuntimeProfilesUsesSyncedLifecycleState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profiles-synced-state")
	base := time.Now().Add(-2 * time.Minute)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profiles-synced-state", agentxbrowserruntime.SharedSessionBrowserProfileState{
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

	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: false},
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
		Arguments: `{"action":"profiles","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime profiles synced state: %v", err)
	}
	var payload struct {
		Action   string `json:"action"`
		Status   string `json:"status"`
		Profiles []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Status        string `json:"status"`
			Running       bool   `json:"running"`
			Connected     bool   `json:"connected"`
			Note          string `json:"note"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode runtime profiles synced-state output: %v", err)
	}
	if payload.Action != "profiles" || payload.Status != "ok" || len(payload.Profiles) != 1 {
		t.Fatalf("unexpected runtime profiles synced-state payload: %#v", payload)
	}
	if payload.Profiles[0].Backend != "proxy" || payload.Profiles[0].Profile != "isolated" || payload.Profiles[0].RuntimeTarget != "node" || payload.Profiles[0].Status != "reconnecting" || !payload.Profiles[0].Running || payload.Profiles[0].Connected || payload.Profiles[0].Note != "cdp reconnect in progress" {
		t.Fatalf("expected profiles payload to use synced lifecycle state, got %#v", payload.Profiles[0])
	}
}
