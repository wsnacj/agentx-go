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

func TestRegisterBrowserTools_TabsListRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-tabs-list-managed-current-hidden-implicit-host")
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
					Action:      "list",
					Status:      "listed",
					ActiveIndex: 2,
					Tabs: []BrowserTab{
						{Index: 1, Title: "Home", URL: "https://93.184.216.34", Active: false},
						{Index: 2, Title: "Workbench", URL: "https://node.example/workbench", Active: true},
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

	sessionRegistry.TrackCurrentTarget("browser-tabs-list-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Name:      "browser_tabs",
		Arguments: `{"action":"list"}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs list managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.tabsReqs) != 0 {
		t.Fatalf("expected host backend tabs list to stay unused, got %#v", hostBackend.tabsReqs)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "list" || nodeBackend.tabsReqs[0].TabIndex != 2 {
		t.Fatalf("expected managed current tabs list to preserve current target binding before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	if nodeBackend.tabsReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected managed current tabs list to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	var payload struct {
		Backend       string       `json:"backend"`
		BrowserApp    string       `json:"browser_app"`
		Profile       string       `json:"profile"`
		RuntimeTarget string       `json:"runtime_target"`
		Action        string       `json:"action"`
		TabIndex      int          `json:"tab_index"`
		ActiveIndex   int          `json:"active_index"`
		Tabs          []BrowserTab `json:"tabs"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed current browser_tabs list output: %v", err)
	}
	if payload.Backend != "proxy-tabs" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "list" || payload.TabIndex != 2 || payload.ActiveIndex != 2 || len(payload.Tabs) != 2 {
		t.Fatalf("unexpected managed current browser_tabs list payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_TabsListRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				tabsResult: BrowserTabsResult{
					Backend:    "proxy-tabs",
					BrowserApp: "Chromium",
					Action:     "list",
					Status:     "listed",
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
		EnabledTools: []string{"browser_tabs"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"list"}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs list managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "list" || nodeBackend.tabsReqs[0].TabIndex != 0 {
		t.Fatalf("expected managed default tabs list to drive node backend before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	var payload struct {
		Backend       string                         `json:"backend"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Action        string                         `json:"action"`
		TabIndex      int                            `json:"tab_index"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_tabs list output: %v", err)
	}
	if payload.Backend != "proxy-tabs" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "list" || payload.TabIndex != 0 {
		t.Fatalf("unexpected managed default browser_tabs list payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "list_tabs_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "list_tabs_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "list_tabs_completed" ||
		payload.View == nil || payload.View.SummaryCode != "list_tabs_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_tabs list to expose stable tabs summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_TabsFocusRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				tabsResult: BrowserTabsResult{
					Backend:    "proxy-tabs",
					BrowserApp: "Chromium",
					Action:     "focus",
					Status:     "focused",
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
		EnabledTools: []string{"browser_tabs"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"focus","tab_index":2,"wait_ms":120}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs focus managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "focus" || nodeBackend.tabsReqs[0].TabIndex != 2 || nodeBackend.tabsReqs[0].WaitMs != 120 {
		t.Fatalf("expected managed default browser_tabs focus to drive node backend before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	var payload struct {
		Backend       string                         `json:"backend"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Action        string                         `json:"action"`
		Status        string                         `json:"status"`
		TabIndex      int                            `json:"tab_index"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_tabs focus output: %v", err)
	}
	if payload.Backend != "proxy-tabs" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "focus" || payload.Status != "focused" || payload.TabIndex != 2 {
		t.Fatalf("unexpected managed default browser_tabs focus payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "focus_tab_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "focus_tab_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "focus_tab_completed" ||
		payload.View == nil || payload.View.SummaryCode != "focus_tab_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_tabs focus to expose stable tabs summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_TabsCloseRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				tabsResult: BrowserTabsResult{
					Backend:    "proxy-tabs",
					BrowserApp: "Chromium",
					Action:     "close",
					Status:     "closed",
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
		EnabledTools: []string{"browser_tabs"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"close","tab_index":2}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs close managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "close" || nodeBackend.tabsReqs[0].TabIndex != 2 {
		t.Fatalf("expected managed default browser_tabs close to drive node backend before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	var payload struct {
		Backend       string                         `json:"backend"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Action        string                         `json:"action"`
		Status        string                         `json:"status"`
		TabIndex      int                            `json:"tab_index"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_tabs close output: %v", err)
	}
	if payload.Backend != "proxy-tabs" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "close" || payload.Status != "closed" || payload.TabIndex != 2 {
		t.Fatalf("unexpected managed default browser_tabs close payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "close_tab_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "close_tab_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "close_tab_completed" ||
		payload.View == nil || payload.View.SummaryCode != "close_tab_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_tabs close to expose stable tabs summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_ActListTabsRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-act-list-tabs-managed-current-hidden-implicit-host")
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
					Action:      "list",
					Status:      "listed",
					ActiveIndex: 2,
					Tabs: []BrowserTab{
						{Index: 1, Title: "Home", URL: "https://93.184.216.34", Active: false},
						{Index: 2, Title: "Workbench", URL: "https://node.example/workbench", Active: true},
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

	sessionRegistry.TrackCurrentTarget("browser-act-list-tabs-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Name:      "browser_act",
		Arguments: `{"kind":"list_tabs"}`,
	})
	if err != nil {
		t.Fatalf("browser_act list_tabs managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.tabsReqs) != 0 {
		t.Fatalf("expected host backend browser_act list_tabs to stay unused, got %#v", hostBackend.tabsReqs)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "list" || nodeBackend.tabsReqs[0].TabIndex != 2 {
		t.Fatalf("expected managed current browser_act list_tabs to preserve current target binding before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	if nodeBackend.tabsReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected managed current browser_act list_tabs to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	var payload struct {
		Kind          string       `json:"kind"`
		Action        string       `json:"action"`
		Backend       string       `json:"backend"`
		BrowserApp    string       `json:"browser_app"`
		Profile       string       `json:"profile"`
		RuntimeTarget string       `json:"runtime_target"`
		TabIndex      int          `json:"tab_index"`
		ActiveIndex   int          `json:"active_index"`
		Tabs          []BrowserTab `json:"tabs"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed current browser_act list_tabs output: %v", err)
	}
	if payload.Kind != "list_tabs" || payload.Action != "list" || payload.Backend != "proxy-tabs" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.TabIndex != 2 || payload.ActiveIndex != 2 || len(payload.Tabs) != 2 {
		t.Fatalf("unexpected managed current browser_act list_tabs payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActOpenRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				openResult: BrowserOpenResult{Backend: "proxy-open", Status: "opened"},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Open: true,
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
		Arguments: `{"kind":"open","url":"https://93.184.216.34"}`,
	})
	if err != nil {
		t.Fatalf("browser_act open managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.openReqs) != 1 || nodeBackend.openReqs[0].URL != "https://93.184.216.34" {
		t.Fatalf("expected managed default browser_act open to drive node backend before implicit host fallback, got %#v", nodeBackend.openReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_act open output: %v", err)
	}
	if payload.Kind != "open" || payload.Backend != "proxy-open" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "opened" {
		t.Fatalf("unexpected managed default browser_act open payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "open_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "open_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "open_completed" ||
		payload.View == nil || payload.View.SummaryCode != "open_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_act open to expose stable open summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_ActNavigateRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				navigateResult: BrowserNavigateResult{Backend: "proxy-navigate", FinalURL: "https://93.184.216.34", Status: "navigated"},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Navigate: true,
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
		Arguments: `{"kind":"navigate","url":"https://93.184.216.34"}`,
	})
	if err != nil {
		t.Fatalf("browser_act navigate managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.navigateReqs) != 1 || nodeBackend.navigateReqs[0].URL != "https://93.184.216.34" || nodeBackend.navigateReqs[0].TabIndex != 0 {
		t.Fatalf("expected managed default browser_act navigate to drive node backend before implicit host fallback, got %#v", nodeBackend.navigateReqs)
	}
	var payload struct {
		Kind                   string                                       `json:"kind"`
		Backend                string                                       `json:"backend"`
		Profile                string                                       `json:"profile"`
		RuntimeTarget          string                                       `json:"runtime_target"`
		Status                 string                                       `json:"status"`
		FinalURL               string                                       `json:"final_url"`
		PostNavigationSnapshot *browserPostNavigationSnapshotRecommendation `json:"post_navigation_snapshot"`
		Summary                *browserTopLevelSummary                      `json:"summary"`
		Display                *browserTopLevelDisplaySummary               `json:"display"`
		Surface                *browserTopLevelSurfaceSummary               `json:"surface"`
		View                   *browserTopLevelViewSummary                  `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_act navigate output: %v", err)
	}
	if payload.Kind != "navigate" || payload.Backend != "proxy-navigate" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "navigated" || payload.FinalURL != "https://93.184.216.34" {
		t.Fatalf("unexpected managed default browser_act navigate payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "navigate_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "navigate_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "navigate_completed" ||
		payload.View == nil || payload.View.SummaryCode != "navigate_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_act navigate to expose stable navigate summary surfaces, got %#v", payload)
	}
	if payload.PostNavigationSnapshot == nil ||
		payload.PostNavigationSnapshot.Recommendation != "take_compact_snapshot" ||
		payload.PostNavigationSnapshot.Tool != "browser_act" ||
		payload.PostNavigationSnapshot.Kind != "snapshot" ||
		!payload.PostNavigationSnapshot.Compact ||
		payload.PostNavigationSnapshot.MaxElements != browserPostNavigationSnapshotMaxElements {
		t.Fatalf("expected browser_act navigate to expose compact snapshot recommendation, got %#v", payload.PostNavigationSnapshot)
	}
}

func TestRegisterBrowserTools_ActListTabsRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				tabsResult: BrowserTabsResult{
					Backend:    "proxy-tabs",
					BrowserApp: "Chromium",
					Action:     "list",
					Status:     "listed",
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
		Arguments: `{"kind":"list_tabs"}`,
	})
	if err != nil {
		t.Fatalf("browser_act list_tabs managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "list" || nodeBackend.tabsReqs[0].TabIndex != 0 {
		t.Fatalf("expected managed default browser_act list_tabs to drive node backend before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Action        string                         `json:"action"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		TabIndex      int                            `json:"tab_index"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_act list_tabs output: %v", err)
	}
	if payload.Kind != "list_tabs" || payload.Action != "list" || payload.Backend != "proxy-tabs" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.TabIndex != 0 {
		t.Fatalf("unexpected managed default browser_act list_tabs payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "list_tabs_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "list_tabs_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "list_tabs_completed" ||
		payload.View == nil || payload.View.SummaryCode != "list_tabs_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_act list_tabs to expose stable tabs summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_ActFocusTabRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				tabsResult: BrowserTabsResult{
					Backend:    "proxy-tabs",
					BrowserApp: "Chromium",
					Action:     "focus",
					Status:     "focused",
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
		Arguments: `{"kind":"focus_tab","tab_index":2,"wait_ms":120}`,
	})
	if err != nil {
		t.Fatalf("browser_act focus_tab managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "focus" || nodeBackend.tabsReqs[0].TabIndex != 2 || nodeBackend.tabsReqs[0].WaitMs != 120 {
		t.Fatalf("expected managed default browser_act focus_tab to drive node backend before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Action        string                         `json:"action"`
		Status        string                         `json:"status"`
		TabIndex      int                            `json:"tab_index"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_act focus_tab output: %v", err)
	}
	if payload.Kind != "focus_tab" || payload.Backend != "proxy-tabs" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "focus" || payload.Status != "focused" || payload.TabIndex != 2 {
		t.Fatalf("unexpected managed default browser_act focus_tab payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "focus_tab_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "focus_tab_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "focus_tab_completed" ||
		payload.View == nil || payload.View.SummaryCode != "focus_tab_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_act focus_tab to expose stable tabs summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_ActCloseTabRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				tabsResult: BrowserTabsResult{
					Backend:    "proxy-tabs",
					BrowserApp: "Chromium",
					Action:     "close",
					Status:     "closed",
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
		Arguments: `{"kind":"close_tab","tab_index":2}`,
	})
	if err != nil {
		t.Fatalf("browser_act close_tab managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "close" || nodeBackend.tabsReqs[0].TabIndex != 2 {
		t.Fatalf("expected managed default browser_act close_tab to drive node backend before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Action        string                         `json:"action"`
		Status        string                         `json:"status"`
		TabIndex      int                            `json:"tab_index"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default browser_act close_tab output: %v", err)
	}
	if payload.Kind != "close_tab" || payload.Backend != "proxy-tabs" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "close" || payload.Status != "closed" || payload.TabIndex != 2 {
		t.Fatalf("unexpected managed default browser_act close_tab payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "close_tab_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "close_tab_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "close_tab_completed" ||
		payload.View == nil || payload.View.SummaryCode != "close_tab_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_act close_tab to expose stable tabs summary surfaces, got %#v", payload)
	}
}
