package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ExtractTabTargetRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				extractResult: BrowserExtractResult{
					Backend:     "proxy-extract",
					BrowserApp:  "Chromium",
					Title:       "Workbench Tab",
					Content:     "managed tab content",
					FinalURL:    publicExampleIPURL,
					ContentType: "text/plain",
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Extract: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, fmt.Errorf("unexpected target %q", requested.Target)
			}
			if strings.TrimSpace(requested.Profile) == "" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			requested.Backend = "proxy"
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         hostBackend,
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser_extract"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-extract-tab-target-managed-hidden-implicit-host")
	tracked := sessionRegistry.TrackCurrentTarget("browser-extract-tab-target-managed-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   4,
		URL:        publicExampleIPURL,
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	if strings.TrimSpace(tracked.ID) == "" {
		t.Fatalf("expected tracked managed current target id")
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_extract",
		Arguments: `{"tab_index":2,"max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser_extract tab target managed hidden implicit host: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 2 || nodeBackend.extractReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected explicit tab target to reuse managed current route before implicit host fallback, got %#v", nodeBackend.extractReqs)
	}
	if nodeBackend.extractReqs[0].BrowserApp != "Chromium" || nodeBackend.extractReqs[0].URL != "" {
		t.Fatalf("expected managed tab extract to route by tab target with managed browser app and without host url fallback, got %#v", nodeBackend.extractReqs[0])
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Target        string `json:"target"`
		TargetID      string `json:"target_id"`
		TabIndex      int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed tab extract output: %v", err)
	}
	if payload.Backend != "proxy-extract" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Target != "tab:2" || payload.TabIndex != 2 || strings.TrimSpace(payload.TargetID) == "" {
		t.Fatalf("unexpected managed tab extract payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_TabsFocusRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				tabsResult: BrowserTabsResult{
					Backend:     "proxy-tabs",
					BrowserApp:  "Chromium",
					Action:      "focus",
					Status:      "focused",
					ActiveIndex: 3,
					Tabs: []BrowserTab{
						{Index: 2, Title: "Docs", URL: "https://docs.example", Active: false},
						{Index: 3, Title: "Workbench", URL: "https://node.example/workbench", Active: true},
					},
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Tabs: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, fmt.Errorf("unexpected target %q", requested.Target)
			}
			if strings.TrimSpace(requested.Profile) == "" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			requested.Backend = "proxy"
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         hostBackend,
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_tabs"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-tabs-focus-managed-hidden-implicit-host")
	tracked := sessionRegistry.TrackCurrentTarget("browser-tabs-focus-managed-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   4,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	if strings.TrimSpace(tracked.ID) == "" {
		t.Fatalf("expected tracked managed current target id")
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"focus","tab_index":3,"wait_ms":120}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs focus managed hidden implicit host: %v", err)
	}
	if len(hostBackend.tabsReqs) != 0 {
		t.Fatalf("expected host backend tabs to stay unused, got %#v", hostBackend.tabsReqs)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "focus" || nodeBackend.tabsReqs[0].TabIndex != 3 || nodeBackend.tabsReqs[0].WaitMs != 120 {
		t.Fatalf("expected explicit tab focus to reuse managed current route before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	if nodeBackend.tabsReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected managed tab focus to inherit managed browser app before host fallback, got %#v", nodeBackend.tabsReqs[0])
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Action        string `json:"action"`
		Status        string `json:"status"`
		Target        string `json:"target"`
		TabIndex      int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed tab focus output: %v", err)
	}
	if payload.Backend != "proxy-tabs" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "focus" || payload.Status != "focused" || payload.Target != "tab:3" || payload.TabIndex != 3 {
		t.Fatalf("unexpected managed tab focus payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActFocusTabRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				tabsResult: BrowserTabsResult{
					Backend:     "proxy-tabs",
					BrowserApp:  "Chromium",
					Action:      "focus",
					Status:      "focused",
					ActiveIndex: 3,
					Tabs: []BrowserTab{
						{Index: 2, Title: "Docs", URL: "https://docs.example", Active: false},
						{Index: 3, Title: "Workbench", URL: "https://node.example/workbench", Active: true},
					},
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Tabs: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, fmt.Errorf("unexpected target %q", requested.Target)
			}
			if strings.TrimSpace(requested.Profile) == "" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			requested.Backend = "proxy"
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         hostBackend,
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-focus-tab-managed-hidden-implicit-host")
	tracked := sessionRegistry.TrackCurrentTarget("browser-act-focus-tab-managed-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   4,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	if strings.TrimSpace(tracked.ID) == "" {
		t.Fatalf("expected tracked managed current target id")
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"focus_tab","tab_index":3,"wait_ms":120}`,
	})
	if err != nil {
		t.Fatalf("browser_act focus_tab managed hidden implicit host: %v", err)
	}
	if len(hostBackend.tabsReqs) != 0 {
		t.Fatalf("expected host backend tabs to stay unused, got %#v", hostBackend.tabsReqs)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "focus" || nodeBackend.tabsReqs[0].TabIndex != 3 || nodeBackend.tabsReqs[0].WaitMs != 120 {
		t.Fatalf("expected act focus_tab to reuse managed current route before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	var payload struct {
		Kind          string `json:"kind"`
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Action        string `json:"action"`
		Status        string `json:"status"`
		Target        string `json:"target"`
		TabIndex      int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed act focus_tab output: %v", err)
	}
	if payload.Kind != "focus_tab" || payload.Backend != "proxy-tabs" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "focus" || payload.Status != "focused" || payload.Target != "tab:3" || payload.TabIndex != 3 {
		t.Fatalf("unexpected managed act focus_tab payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActConsoleRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				consoleResult: BrowserConsoleResult{
					Backend:    "proxy-console",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/workbench",
					Title:      "Workbench",
					Messages: []BrowserConsoleMessage{
						{Level: "error", Text: "boom"},
					},
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Console: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, fmt.Errorf("unexpected target %q", requested.Target)
			}
			if strings.TrimSpace(requested.Profile) == "" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			requested.Backend = "proxy"
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         hostBackend,
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-console-managed-hidden-implicit-host")
	tracked := sessionRegistry.TrackCurrentTarget("browser-act-console-managed-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	if strings.TrimSpace(tracked.ID) == "" {
		t.Fatalf("expected tracked managed current target id")
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"console","level":"error"}`,
	})
	if err != nil {
		t.Fatalf("browser_act console managed hidden implicit host: %v", err)
	}
	if len(hostBackend.consoleReqs) != 0 {
		t.Fatalf("expected host backend console to stay unused, got %#v", hostBackend.consoleReqs)
	}
	if len(nodeBackend.consoleReqs) != 1 || nodeBackend.consoleReqs[0].Level != "error" || nodeBackend.consoleReqs[0].TabIndex != 2 || nodeBackend.consoleReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected act console to reuse managed current route before implicit host fallback, got %#v", nodeBackend.consoleReqs)
	}
	if nodeBackend.consoleReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected managed act console to inherit managed browser app before host fallback, got %#v", nodeBackend.consoleReqs[0])
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Target        string                         `json:"target"`
		TargetID      string                         `json:"target_id"`
		TabIndex      int                            `json:"tab_index"`
		FinalURL      string                         `json:"final_url"`
		Messages      []BrowserConsoleMessage        `json:"messages"`
		Status        string                         `json:"status"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed act console output: %v", err)
	}
	if payload.Kind != "console" || payload.Backend != "proxy-console" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.TabIndex != 2 || payload.FinalURL != "https://node.example/workbench" || payload.Status != "ok" || len(payload.Messages) != 1 || payload.Messages[0].Level != "error" {
		t.Fatalf("unexpected managed act console payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "console_collected" {
		t.Fatalf("unexpected managed act console summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "console_collected" {
		t.Fatalf("unexpected managed act console display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "console_collected" {
		t.Fatalf("unexpected managed act console surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "console_collected" {
		t.Fatalf("unexpected managed act console view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_ActExtractRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				extractResult: BrowserExtractResult{
					Backend:     "proxy-extract",
					BrowserApp:  "Chromium",
					Title:       "Workbench",
					Content:     "managed current content",
					FinalURL:    publicExampleIPURL,
					ContentType: "text/plain",
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Extract: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, fmt.Errorf("unexpected target %q", requested.Target)
			}
			if strings.TrimSpace(requested.Profile) == "" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			requested.Backend = "proxy"
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         hostBackend,
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-extract-managed-current-hidden-implicit-host")
	tracked := sessionRegistry.TrackCurrentTarget("browser-act-extract-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        publicExampleIPURL,
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend browser_act extract to stay unused, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 2 || nodeBackend.extractReqs[0].WaitMs != 250 {
		t.Fatalf("expected targetless browser_act extract to reuse managed current route before implicit host fallback, got %#v", nodeBackend.extractReqs)
	}
	if nodeBackend.extractReqs[0].BrowserApp != "Chromium" || nodeBackend.extractReqs[0].URL != "" {
		t.Fatalf("expected targetless browser_act extract to stay tab-targeted while inheriting managed browser app before backend dispatch, got %#v", nodeBackend.extractReqs[0])
	}
	var payload struct {
		Kind          string `json:"kind"`
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Target        string `json:"target"`
		TargetID      string `json:"target_id"`
		FinalURL      string `json:"final_url"`
		TabIndex      int    `json:"tab_index"`
		Status        string `json:"status"`
		Content       string `json:"content"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed targetless browser_act extract output: %v", err)
	}
	if payload.Kind != "extract" || payload.Backend != "proxy-extract" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.FinalURL != publicExampleIPURL || payload.TabIndex != 2 || payload.Status != "extracted" || payload.Content != "managed current content" {
		t.Fatalf("unexpected managed targetless browser_act extract payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActExtractRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				extractResult: BrowserExtractResult{
					Backend:     "proxy-extract",
					BrowserApp:  "Chromium",
					Title:       "Workbench",
					Content:     "managed default content",
					FinalURL:    "https://node.example/default",
					ContentType: "text/plain",
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Extract: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			requested.Backend = "proxy"
			requested.Profile = firstNonEmpty(strings.TrimSpace(requested.Profile), "workbench")
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		NodeBackend:  nodeBackend,
		MaxChars:     64,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 0 || nodeBackend.extractReqs[0].URL != "" || nodeBackend.extractReqs[0].WaitMs != browserTabTargetWaitMs || nodeBackend.extractReqs[0].MaxChars != 32 {
		t.Fatalf("expected managed default browser_act extract to drive node backend before implicit host fallback, got %#v", nodeBackend.extractReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		FinalURL      string                         `json:"final_url"`
		TabIndex      int                            `json:"tab_index"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_act extract output: %v", err)
	}
	if payload.Kind != "extract" || payload.Backend != "proxy-extract" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "extracted" || payload.FinalURL != "https://node.example/default" || payload.TabIndex != 0 {
		t.Fatalf("unexpected managed default browser_act extract payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "extract_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "extract_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "extract_completed" ||
		payload.View == nil || payload.View.SummaryCode != "extract_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_act extract to expose stable extract summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_ActSnapshotRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				snapshotResult: BrowserSnapshotResult{
					Backend:    "proxy-snapshot",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/default",
					Title:      "Workbench",
					Snapshot:   "managed default snapshot",
					Format:     "ai",
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Snapshot: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			requested.Backend = "proxy"
			requested.Profile = firstNonEmpty(strings.TrimSpace(requested.Profile), "workbench")
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		NodeBackend:  nodeBackend,
		MaxChars:     64,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"snapshot","max_chars":24,"max_elements":6,"format":"ai","mode":"efficient"}`,
	})
	if err != nil {
		t.Fatalf("browser_act snapshot managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.snapshotReqs) != 1 || nodeBackend.snapshotReqs[0].TabIndex != 0 || nodeBackend.snapshotReqs[0].URL != "" || nodeBackend.snapshotReqs[0].WaitMs != 250 || nodeBackend.snapshotReqs[0].MaxChars != 24 || nodeBackend.snapshotReqs[0].MaxElements != 6 || nodeBackend.snapshotReqs[0].Format != "ai" || nodeBackend.snapshotReqs[0].Mode != "efficient" {
		t.Fatalf("expected managed default browser_act snapshot to drive node backend before implicit host fallback, got %#v", nodeBackend.snapshotReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		FinalURL      string                         `json:"final_url"`
		Snapshot      string                         `json:"snapshot"`
		TabIndex      int                            `json:"tab_index"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_act snapshot output: %v", err)
	}
	if payload.Kind != "snapshot" || payload.Backend != "proxy-snapshot" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.FinalURL != "https://node.example/default" || payload.Snapshot != "managed default snapshot" || payload.TabIndex != 0 {
		t.Fatalf("unexpected managed default browser_act snapshot payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "content" || payload.Summary.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected managed default browser_act snapshot summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "content" || payload.Display.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected managed default browser_act snapshot display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "content" || payload.Surface.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected managed default browser_act snapshot surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "content" || payload.View.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected managed default browser_act snapshot view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_ActScreenshotRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				screenshotResult: BrowserScreenshotResult{
					Backend:       "proxy-screenshot",
					BrowserApp:    "Chromium",
					FinalURL:      "https://node.example/default",
					Title:         "Workbench",
					CaptureScope:  "viewport",
					CaptureWidth:  1280,
					CaptureHeight: 720,
					Status:        "captured",
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Screenshot: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			requested.Backend = "proxy"
			requested.Profile = firstNonEmpty(strings.TrimSpace(requested.Profile), "workbench")
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		NodeBackend:     nodeBackend,
		EnabledTools:    []string{"browser_act"},
		PublishArtifact: testBrowserArtifactPublisher,
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"screenshot"}`,
	})
	if err != nil {
		t.Fatalf("browser_act screenshot managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.screenshotReqs) != 1 || nodeBackend.screenshotReqs[0].TabIndex != 0 || nodeBackend.screenshotReqs[0].URL != "" || nodeBackend.screenshotReqs[0].WaitMs != browserTabTargetWaitMs || strings.TrimSpace(nodeBackend.screenshotReqs[0].OutputPath) == "" {
		t.Fatalf("expected managed default browser_act screenshot to drive node backend before implicit host fallback, got %#v", nodeBackend.screenshotReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		FinalURL      string                         `json:"final_url"`
		TabIndex      int                            `json:"tab_index"`
		Path          string                         `json:"path"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_act screenshot output: %v", err)
	}
	if payload.Kind != "screenshot" || payload.Backend != "proxy-screenshot" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "captured" || payload.FinalURL != "https://node.example/default" || payload.TabIndex != 0 || strings.TrimSpace(payload.Path) == "" {
		t.Fatalf("unexpected managed default browser_act screenshot payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "screenshot_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "screenshot_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "screenshot_completed" ||
		payload.View == nil || payload.View.SummaryCode != "screenshot_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_act screenshot to expose stable screenshot summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_ActTypeRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				typeResult: BrowserTypeResult{
					Backend:    "proxy-type",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/form",
					Title:      "Form",
					Value:      "agentx",
					Status:     "typed",
					Submitted:  true,
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				TypeText: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			requested.Backend = "proxy"
			requested.Profile = firstNonEmpty(strings.TrimSpace(requested.Profile), "workbench")
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"type","selector":"input[name=q]","text":"agentx","submit":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act type managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.typeReqs) != 1 || nodeBackend.typeReqs[0].TabIndex != 0 || nodeBackend.typeReqs[0].URL != "" || nodeBackend.typeReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs || nodeBackend.typeReqs[0].PostWaitMs != 250 || nodeBackend.typeReqs[0].Selector != "input[name=q]" || nodeBackend.typeReqs[0].Text != "agentx" || !nodeBackend.typeReqs[0].Submit {
		t.Fatalf("expected managed default browser_act type to drive node backend before implicit host fallback, got %#v", nodeBackend.typeReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		FinalURL      string                         `json:"final_url"`
		TabIndex      int                            `json:"tab_index"`
		Value         string                         `json:"value"`
		Submitted     bool                           `json:"submitted"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_act type output: %v", err)
	}
	if payload.Kind != "type" || payload.Backend != "proxy-type" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "typed" || payload.FinalURL != "https://node.example/form" || payload.TabIndex != 0 || payload.Value != "agentx" || !payload.Submitted {
		t.Fatalf("unexpected managed default browser_act type payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "form" || payload.Summary.SummaryCode != "type_completed" {
		t.Fatalf("unexpected managed default browser_act type summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "form" || payload.Display.SummaryCode != "type_completed" {
		t.Fatalf("unexpected managed default browser_act type display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "form" || payload.Surface.SummaryCode != "type_completed" {
		t.Fatalf("unexpected managed default browser_act type surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "form" || payload.View.SummaryCode != "type_completed" {
		t.Fatalf("unexpected managed default browser_act type view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_ActEvaluateRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				evalResult: BrowserEvalResult{
					Backend:    "proxy-eval",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/default",
					Title:      "Workbench",
					Result:     "managed default eval",
					Status:     "evaluated",
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Evaluate: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			requested.Backend = "proxy"
			requested.Profile = firstNonEmpty(strings.TrimSpace(requested.Profile), "workbench")
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		NodeBackend:  nodeBackend,
		MaxChars:     64,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"evaluate","script":"document.title","max_chars":24}`,
	})
	if err != nil {
		t.Fatalf("browser_act evaluate managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.evalReqs) != 1 || nodeBackend.evalReqs[0].TabIndex != 0 || nodeBackend.evalReqs[0].URL != "" || nodeBackend.evalReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs || nodeBackend.evalReqs[0].Script != "document.title" || nodeBackend.evalReqs[0].MaxChars != 24 {
		t.Fatalf("expected managed default browser_act evaluate to drive node backend before implicit host fallback, got %#v", nodeBackend.evalReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		FinalURL      string                         `json:"final_url"`
		TabIndex      int                            `json:"tab_index"`
		Result        string                         `json:"result"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_act evaluate output: %v", err)
	}
	if payload.Kind != "evaluate" || payload.Backend != "proxy-eval" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "evaluated" || payload.FinalURL != "https://node.example/default" || payload.TabIndex != 0 || payload.Result != "managed default eval" {
		t.Fatalf("unexpected managed default browser_act evaluate payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "script" || payload.Summary.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected managed default browser_act evaluate summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "script" || payload.Display.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected managed default browser_act evaluate display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "script" || payload.Surface.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected managed default browser_act evaluate surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "script" || payload.View.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected managed default browser_act evaluate view: %#v", payload.View)
	}
}
