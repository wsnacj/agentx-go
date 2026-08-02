package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_RuntimeSelectProfilePropagatesToBrowserTabs(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			ActiveIndex: 1,
			Tabs: []BrowserTab{
				{Index: 1, URL: "https://host.example", Title: "Host"},
			},
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			tabsResult: BrowserTabsResult{
				ActiveIndex: 1,
				Tabs: []BrowserTab{
					{Index: 1, URL: "https://node.example/workbench", Title: "Workbench"},
				},
			},
			extractResult: BrowserExtractResult{
				Backend:    "proxy-extract",
				BrowserApp: "Chromium",
				Title:      "Workbench",
				Content:    "Remembered extract",
				FinalURL:   "https://node.example/workbench",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_tabs", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-tabs-session")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs after session profile selection: %v", err)
	}
	if len(hostBackend.tabsReqs) != 0 {
		t.Fatalf("expected host backend tabs to stay unused, got %#v", hostBackend.tabsReqs)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].TabIndex != 0 {
		t.Fatalf("expected node backend tabs to be used once, got %#v", nodeBackend.tabsReqs)
	}
	if !strings.Contains(out, `https://node.example/workbench`) {
		t.Fatalf("expected browser_tabs output to come from selected node profile route, got %s", out)
	}
}

func TestRegisterBrowserTools_RuntimePrepareRememberProfilePropagatesToBrowserTabs(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			ActiveIndex: 1,
			Tabs: []BrowserTab{
				{Index: 1, URL: "https://host.example", Title: "Host"},
			},
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
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
			tabsResult: BrowserTabsResult{
				ActiveIndex: 1,
				Tabs: []BrowserTab{
					{Index: 1, URL: "https://node.example/workbench", Title: "Workbench"},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_tabs", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-remember")
	sessionRegistry.TrackTab("browser-runtime-prepare-remember", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	prepareOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"prepare","runtime_target":"node","profile":"workbench","remember_profile":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare remember_profile: %v", err)
	}
	var preparePayload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		RememberDecision        string `json:"remember_decision"`
		RememberReady           bool   `json:"remember_ready"`
		SessionProfileSelection struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Source        string `json:"source"`
		} `json:"session_profile_selection"`
		SessionTargetSelection struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(prepareOut), &preparePayload); err != nil {
		t.Fatalf("decode prepare output: %v", err)
	}
	if preparePayload.Action != "prepare" || preparePayload.Status != "ok" || preparePayload.RememberDecision != "session_profile_remembered" || !preparePayload.RememberReady {
		t.Fatalf("unexpected prepare remember payload: %#v", preparePayload)
	}
	if preparePayload.SessionProfileSelection.Profile != "workbench" || preparePayload.SessionProfileSelection.RuntimeTarget != "node" {
		t.Fatalf("expected prepare to remember selected node profile, got %#v", preparePayload.SessionProfileSelection)
	}
	if preparePayload.SessionProfileSelection.Source != "remember_profile" {
		t.Fatalf("expected remember_profile source, got %#v", preparePayload.SessionProfileSelection)
	}
	if preparePayload.SessionTargetSelection.TabIndex != 1 || preparePayload.SessionTargetSelection.Source != "remember_profile" {
		t.Fatalf("expected remember_profile to sync current target, got %#v", preparePayload.SessionTargetSelection)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs after remember_profile prepare: %v", err)
	}
	if len(hostBackend.tabsReqs) != 0 {
		t.Fatalf("expected host backend tabs to stay unused, got %#v", hostBackend.tabsReqs)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].TabIndex != 1 {
		t.Fatalf("expected node backend tabs to reuse remembered target, got %#v", nodeBackend.tabsReqs)
	}
	if !strings.Contains(out, `https://node.example/workbench`) {
		t.Fatalf("expected browser_tabs output to come from remembered node profile route, got %s", out)
	}
	extractOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_act after remember_profile prepare: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused after remember_profile prepare, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 1 {
		t.Fatalf("expected node backend extract to reuse remembered target, got %#v", nodeBackend.extractReqs)
	}
	if !strings.Contains(extractOut, `"status":"extracted"`) || !strings.Contains(extractOut, `"tab_index":1`) {
		t.Fatalf("expected browser_act extract output after remember_profile prepare, got %s", extractOut)
	}
}

func TestRegisterBrowserTools_RuntimePrepareRememberProfileClearsMismatchedTargetSelection(t *testing.T) {
	reg := llmxtools.NewRegistry()
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
				Title:      "Alternate",
				Content:    "Alternate content",
				FinalURL:   "https://node.example/alternate",
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
					{Profile: "alternate", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "alternate",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-prepare-remember-clears-mismatched-target")
	sessionRegistry.TrackTab("browser-runtime-prepare-remember-clears-mismatched-target", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, true)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_target","runtime_target":"node","profile":"workbench","target":"tab:1"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_target before remember_profile switch: %v", err)
	}

	prepareOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"prepare","runtime_target":"node","profile":"alternate","remember_profile":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime prepare alternate remember_profile: %v", err)
	}
	var preparePayload struct {
		Action                  string                                 `json:"action"`
		Status                  string                                 `json:"status"`
		RememberDecision        string                                 `json:"remember_decision"`
		RememberReady           bool                                   `json:"remember_ready"`
		SessionProfileSelection *browserRuntimeSessionProfileSelection `json:"session_profile_selection"`
		SessionTargetSelection  *browserRuntimeSessionTargetSelection  `json:"session_target_selection"`
		SessionBinding          struct {
			SelectedBrowserProfile      string `json:"selected_browser_profile"`
			SelectedBrowserTargetID     string `json:"selected_browser_target_id"`
			SelectedBrowserTargetSource string `json:"selected_browser_target_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(prepareOut), &preparePayload); err != nil {
		t.Fatalf("decode prepare alternate output: %v", err)
	}
	if preparePayload.Action != "prepare" || preparePayload.Status != "ok" || preparePayload.RememberDecision != "session_profile_remembered" || !preparePayload.RememberReady {
		t.Fatalf("unexpected prepare alternate remember payload: %#v", preparePayload)
	}
	if preparePayload.SessionProfileSelection == nil || preparePayload.SessionProfileSelection.Profile != "alternate" || preparePayload.SessionProfileSelection.Source != "remember_profile" {
		t.Fatalf("expected alternate profile remembered, got %#v", preparePayload.SessionProfileSelection)
	}
	if preparePayload.SessionTargetSelection != nil {
		t.Fatalf("expected mismatched target selection to clear during remember_profile switch, got %#v", preparePayload.SessionTargetSelection)
	}
	if preparePayload.SessionBinding.SelectedBrowserProfile != "alternate" || preparePayload.SessionBinding.SelectedBrowserTargetID != "" || preparePayload.SessionBinding.SelectedBrowserTargetSource != "" {
		t.Fatalf("unexpected session binding after remember_profile switch: %#v", preparePayload.SessionBinding)
	}

	extractOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_act after alternate remember_profile prepare: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused after alternate remember_profile prepare, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 0 {
		t.Fatalf("expected node backend extract to avoid old remembered target after remember_profile switch, got %#v", nodeBackend.extractReqs)
	}
	if !strings.Contains(extractOut, `"status":"extracted"`) || strings.Contains(extractOut, `"tab_index":1`) {
		t.Fatalf("expected browser_act extract output after alternate remember_profile prepare, got %s", extractOut)
	}
}

func TestRegisterBrowserTools_RuntimeSyncSessionPromotesCurrentRouteDefaults(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "system-extract",
			BrowserApp: "Safari",
			Title:      "Host",
			Content:    "Host content",
			FinalURL:   "https://host.example",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
			extractResult: BrowserExtractResult{
				Backend:    "proxy-extract",
				BrowserApp: "Chromium",
				Title:      "Workbench",
				Content:    "Extracted content",
				FinalURL:   "https://node.example/workbench",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sync-session")
	tracked := sessionRegistry.TrackTabs("browser-runtime-sync-session", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://node.example/workbench",
			Title:      "Workbench",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
	}, 1)
	if len(tracked) != 1 || strings.TrimSpace(tracked[0].ID) == "" {
		t.Fatalf("expected one tracked target before sync_session, got %#v", tracked)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sync_session","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sync_session: %v", err)
	}
	var payload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		SyncSessionDecision     string `json:"sync_session_decision"`
		SyncSessionReady        bool   `json:"sync_session_ready"`
		SessionProfileSelection struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Source        string `json:"source"`
		} `json:"session_profile_selection"`
		SessionTargetSelection struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
		RouteResolution struct {
			ProfileSource       string `json:"profile_source"`
			RuntimeTargetSource string `json:"runtime_target_source"`
			TargetSource        string `json:"target_source"`
		} `json:"route_resolution"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode sync_session output: %v", err)
	}
	if payload.Action != "sync_session" || payload.Status != "ok" || payload.SyncSessionDecision != "session_route_synced" || !payload.SyncSessionReady {
		t.Fatalf("unexpected sync_session payload: %#v", payload)
	}
	if payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.RuntimeTarget != "node" || payload.SessionProfileSelection.Source != "sync_session" {
		t.Fatalf("unexpected sync_session profile selection: %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection.ID != tracked[0].ID || payload.SessionTargetSelection.TabIndex != 1 || payload.SessionTargetSelection.Source != "sync_session" {
		t.Fatalf("unexpected sync_session target selection: %#v", payload.SessionTargetSelection)
	}
	if payload.RouteResolution.ProfileSource != "explicit_request" || payload.RouteResolution.RuntimeTargetSource != "explicit_request" || payload.RouteResolution.TargetSource != "sync_session" {
		t.Fatalf("unexpected sync_session route resolution: %#v", payload.RouteResolution)
	}

	extractOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after sync_session: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused after sync_session, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 1 {
		t.Fatalf("expected node backend extract to inherit synced target, got %#v", nodeBackend.extractReqs)
	}
	if !strings.Contains(extractOut, `"status":"extracted"`) || !strings.Contains(extractOut, `"tab_index":1`) {
		t.Fatalf("expected browser_act extract output after sync_session, got %s", extractOut)
	}

	againOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sync_session","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sync_session again: %v", err)
	}
	var againPayload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		SyncSessionDecision string `json:"sync_session_decision"`
		SyncSessionReady    bool   `json:"sync_session_ready"`
	}
	if err := json.Unmarshal([]byte(againOut), &againPayload); err != nil {
		t.Fatalf("decode second sync_session output: %v", err)
	}
	if againPayload.Action != "sync_session" || againPayload.Status != "ok" || !againPayload.SyncSessionReady {
		t.Fatalf("unexpected second sync_session payload: %#v", againPayload)
	}
	if againPayload.SyncSessionDecision != "session_route_already_synced" && againPayload.SyncSessionDecision != "session_profile_already_synced" {
		t.Fatalf("unexpected second sync_session payload: %#v", againPayload)
	}
}

func TestRegisterBrowserTools_RuntimeSyncSessionClearsMismatchedTargetSelection(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:    "system-extract",
			BrowserApp: "Safari",
			Title:      "Host",
			Content:    "Host content",
			FinalURL:   "https://host.example",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend: "proxy",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
					{Profile: "alternate", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
			extractResult: BrowserExtractResult{
				Backend:    "proxy-extract",
				BrowserApp: "Chromium",
				Title:      "Alternate",
				Content:    "Alternate content",
				FinalURL:   "https://node.example/alternate",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-sync-session-clears-mismatched-target")
	sessionRegistry.TrackTabs("browser-runtime-sync-session-clears-mismatched-target", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://node.example/workbench",
			Title:      "Workbench",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
	}, 1)
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_target","runtime_target":"node","profile":"workbench","target":"tab:1"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_target before sync_session switch: %v", err)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sync_session","runtime_target":"node","profile":"alternate"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sync_session alternate: %v", err)
	}
	var payload struct {
		Action                  string                                 `json:"action"`
		Status                  string                                 `json:"status"`
		SyncSessionDecision     string                                 `json:"sync_session_decision"`
		SyncSessionReady        bool                                   `json:"sync_session_ready"`
		SessionProfileSelection *browserRuntimeSessionProfileSelection `json:"session_profile_selection"`
		SessionTargetSelection  *browserRuntimeSessionTargetSelection  `json:"session_target_selection"`
		SessionBinding          struct {
			SelectedBrowserProfile      string `json:"selected_browser_profile"`
			SelectedBrowserTargetID     string `json:"selected_browser_target_id"`
			SelectedBrowserTargetSource string `json:"selected_browser_target_source"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode sync_session alternate output: %v", err)
	}
	if payload.Action != "sync_session" || payload.Status != "ok" || payload.SyncSessionDecision != "session_profile_synced" || !payload.SyncSessionReady {
		t.Fatalf("unexpected sync_session alternate payload: %#v", payload)
	}
	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "alternate" || payload.SessionProfileSelection.Source != "sync_session" {
		t.Fatalf("expected alternate profile synced, got %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection != nil {
		t.Fatalf("expected mismatched target selection to clear during sync_session switch, got %#v", payload.SessionTargetSelection)
	}
	if payload.SessionBinding.SelectedBrowserProfile != "alternate" || payload.SessionBinding.SelectedBrowserTargetID != "" || payload.SessionBinding.SelectedBrowserTargetSource != "" {
		t.Fatalf("unexpected session binding after sync_session alternate: %#v", payload.SessionBinding)
	}

	extractOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after sync_session alternate: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused after sync_session alternate, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 0 {
		t.Fatalf("expected node backend extract to avoid old remembered target after sync_session switch, got %#v", nodeBackend.extractReqs)
	}
	if !strings.Contains(extractOut, `"status":"extracted"`) || strings.Contains(extractOut, `"tab_index":1`) {
		t.Fatalf("expected browser_act extract output after sync_session alternate, got %s", extractOut)
	}
}

func TestRegisterBrowserTools_RuntimeCoordinateSync(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &fakeBrowserBackend{
		extractResult: BrowserExtractResult{FinalURL: "https://host.example", Content: "host fallback"},
	}
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			extractResult: BrowserExtractResult{
				FinalURL: "https://node.example/workbench",
				Content:  "proxy extract",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	nodeBackend.runtimeProfilesResult = BrowserProfilesResult{
		Backend:        "proxy",
		DefaultProfile: "workbench",
		Profiles: []BrowserProfileInfo{
			{Profile: "workbench", Status: "running", BrowserApp: "Chromium"},
		},
	}
	sessionRegistry := NewBrowserSessionRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         hostBackend,
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime", "browser_tabs", "browser_act"},
	})
	callCtx := WithToolSessionID(context.Background(), "tool-session")
	tracked := sessionRegistry.TrackTabs("tool-session", []agentxbrowserruntime.BrowserSessionTarget{
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
	if len(tracked) != 1 || strings.TrimSpace(tracked[0].ID) == "" {
		t.Fatalf("expected one tracked target before coordinate sync, got %#v", tracked)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"coordinate","runtime_target":"node","profile":"workbench","coordination_goal":"sync"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime coordinate sync: %v", err)
	}
	var payload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		CoordinationGoal        string `json:"coordination_goal"`
		CoordinationDecision    string `json:"coordination_decision"`
		CoordinationReady       bool   `json:"coordination_ready"`
		SyncSessionDecision     string `json:"sync_session_decision"`
		SyncSessionReady        bool   `json:"sync_session_ready"`
		SessionProfileSelection struct {
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Source        string `json:"source"`
		} `json:"session_profile_selection"`
		SessionTargetSelection struct {
			ID       string `json:"id"`
			TabIndex int    `json:"tab_index"`
			Source   string `json:"source"`
		} `json:"session_target_selection"`
		RouteResolution struct {
			TargetSource string `json:"target_source"`
		} `json:"route_resolution"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode coordinate sync output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "ok" || payload.CoordinationGoal != "sync" {
		t.Fatalf("unexpected coordinate sync payload: %#v", payload)
	}
	if payload.CoordinationDecision != "started" || payload.SyncSessionDecision != "session_target_synced" || !payload.SyncSessionReady || !payload.CoordinationReady {
		t.Fatalf("unexpected coordinate sync decisions: %#v", payload)
	}
	if payload.SessionProfileSelection.Profile != "" || payload.SessionProfileSelection.RuntimeTarget != "" || payload.SessionProfileSelection.Source != "" {
		t.Fatalf("unexpected coordinate sync profile selection: %#v", payload.SessionProfileSelection)
	}
	if payload.SessionTargetSelection.ID != tracked[0].ID || payload.SessionTargetSelection.TabIndex != 1 || payload.SessionTargetSelection.Source != "sync_session" {
		t.Fatalf("unexpected coordinate sync target selection: %#v", payload.SessionTargetSelection)
	}
	if payload.RouteResolution.TargetSource != "sync_session" {
		t.Fatalf("unexpected coordinate sync route resolution: %#v", payload.RouteResolution)
	}

	extractOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after coordinate sync: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused after coordinate sync, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 1 {
		t.Fatalf("expected node backend extract to inherit synced target, got %#v", nodeBackend.extractReqs)
	}
	if !strings.Contains(extractOut, `"status":"extracted"`) || !strings.Contains(extractOut, `"tab_index":1`) {
		t.Fatalf("expected browser_act extract output after coordinate sync, got %s", extractOut)
	}
}
