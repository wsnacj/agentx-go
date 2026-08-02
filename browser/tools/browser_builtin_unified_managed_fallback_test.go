package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_UnifiedBrowserOpenRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		Root:            t.TempDir(),
		NodeBackend:     nodeBackend,
		EnabledTools:    []string{"browser"},
		PublishArtifact: testBrowserArtifactPublisher,
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"open","url":"https://93.184.216.34"}`,
	})
	if err != nil {
		t.Fatalf("browser unified open managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.openReqs) != 1 || nodeBackend.openReqs[0].URL != "https://93.184.216.34" {
		t.Fatalf("expected unified browser open to drive managed default route before implicit host fallback, got %#v", nodeBackend.openReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		Explanation   *browserTopLevelSummary        `json:"explanation"`
		Diagnostics   *browserTopLevelSummary        `json:"diagnostics"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser open output: %v", err)
	}
	if payload.Kind != "open" || payload.Backend != "proxy-open" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "opened" {
		t.Fatalf("unexpected unified browser open payload: %#v", payload)
	}
	if payload.Explanation == nil || payload.Explanation.SummaryCode != "open_completed" ||
		payload.Diagnostics == nil || payload.Diagnostics.SummaryCode != "open_completed" ||
		payload.Summary == nil || payload.Summary.SummaryCode != "open_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "open_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "open_completed" ||
		payload.View == nil || payload.View.SummaryCode != "open_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected unified browser open to expose stable open summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserOpenRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-open-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				openResult: BrowserOpenResult{Backend: "proxy-open", BrowserApp: "Chromium", Status: "opened"},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Open: true,
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
		EnabledTools:    []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-open-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"open","url":"` + publicExampleIPURL + `"}`,
	})
	if err != nil {
		t.Fatalf("browser unified open managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.openReqs) != 0 {
		t.Fatalf("expected host backend open to stay unused, got %#v", hostBackend.openReqs)
	}
	if len(nodeBackend.openReqs) != 1 || nodeBackend.openReqs[0].URL != publicExampleIPURL {
		t.Fatalf("expected unified browser open to reuse managed current route before implicit host fallback, got %#v", nodeBackend.openReqs)
	}
	if nodeBackend.openReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser open to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.openReqs[0])
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		Target        string                         `json:"target"`
		TargetID      string                         `json:"target_id"`
		Explanation   *browserTopLevelSummary        `json:"explanation"`
		Diagnostics   *browserTopLevelSummary        `json:"diagnostics"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser open output: %v", err)
	}
	if payload.Kind != "open" || payload.Backend != "proxy-open" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "opened" || payload.Target != "current" || strings.TrimSpace(payload.TargetID) == "" {
		t.Fatalf("unexpected unified browser open payload: %#v", payload)
	}
	if payload.Explanation == nil || payload.Explanation.SummaryCode != "open_completed" ||
		payload.Diagnostics == nil || payload.Diagnostics.SummaryCode != "open_completed" ||
		payload.Summary == nil || payload.Summary.SummaryCode != "open_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "open_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "open_completed" ||
		payload.View == nil || payload.View.SummaryCode != "open_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected unified browser open to expose stable open summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserNavigateRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		Root:            t.TempDir(),
		NodeBackend:     nodeBackend,
		EnabledTools:    []string{"browser"},
		PublishArtifact: testBrowserArtifactPublisher,
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"navigate","url":"https://93.184.216.34"}`,
	})
	if err != nil {
		t.Fatalf("browser unified navigate managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.navigateReqs) != 1 || nodeBackend.navigateReqs[0].URL != "https://93.184.216.34" || nodeBackend.navigateReqs[0].TabIndex != 0 {
		t.Fatalf("expected unified browser navigate to drive managed default route before implicit host fallback, got %#v", nodeBackend.navigateReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		FinalURL      string                         `json:"final_url"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser navigate output: %v", err)
	}
	if payload.Kind != "navigate" || payload.Backend != "proxy-navigate" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "navigated" || payload.FinalURL != "https://93.184.216.34" {
		t.Fatalf("unexpected unified browser navigate payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "navigate_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "navigate_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "navigate_completed" ||
		payload.View == nil || payload.View.SummaryCode != "navigate_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected unified browser navigate to expose stable navigate summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserNavigateRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-navigate-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				navigateResult: BrowserNavigateResult{
					Backend:    "proxy-navigate",
					BrowserApp: "Chromium",
					FinalURL:   publicExampleIPURL,
					Title:      "Workbench",
					Status:     "navigated",
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Navigate: true,
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
		EnabledTools:    []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-navigate-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"navigate","url":"` + publicExampleIPURL + `"}`,
	})
	if err != nil {
		t.Fatalf("browser unified navigate managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.navigateReqs) != 0 {
		t.Fatalf("expected host backend navigate to stay unused, got %#v", hostBackend.navigateReqs)
	}
	if len(nodeBackend.navigateReqs) != 1 || nodeBackend.navigateReqs[0].URL != publicExampleIPURL {
		t.Fatalf("expected unified browser navigate to reuse managed current route before implicit host fallback, got %#v", nodeBackend.navigateReqs)
	}
	if nodeBackend.navigateReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser navigate to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.navigateReqs[0])
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		FinalURL      string                         `json:"final_url"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser navigate output: %v", err)
	}
	if payload.Kind != "navigate" || payload.Backend != "proxy-navigate" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "navigated" || payload.FinalURL != publicExampleIPURL {
		t.Fatalf("unexpected unified browser navigate payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "navigate_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "navigate_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "navigate_completed" ||
		payload.View == nil || payload.View.SummaryCode != "navigate_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected unified browser navigate to expose stable navigate summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserClickRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				clickResult: BrowserClickResult{
					Backend:    "proxy-click",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/after-click",
					Title:      "Clicked",
					Status:     "clicked",
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Click: true,
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
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"click","selector":"button.buy","post_wait_ms":500}`,
	})
	if err != nil {
		t.Fatalf("browser unified click managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.clickReqs) != 1 || nodeBackend.clickReqs[0].TabIndex != 0 || nodeBackend.clickReqs[0].URL != "" || nodeBackend.clickReqs[0].Selector != "button.buy" || nodeBackend.clickReqs[0].PostWaitMs != 500 || nodeBackend.clickReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs {
		t.Fatalf("expected unified browser click to drive managed default route before implicit host fallback, got %#v", nodeBackend.clickReqs)
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
		t.Fatalf("decode unified browser click output: %v", err)
	}
	if payload.Kind != "click" || payload.Backend != "proxy-click" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "clicked" || payload.FinalURL != "https://node.example/after-click" || payload.TabIndex != 0 {
		t.Fatalf("unexpected unified browser click payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "interaction" || payload.Summary.SummaryCode != "click_completed" {
		t.Fatalf("unexpected unified browser click summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "interaction" || payload.Display.SummaryCode != "click_completed" {
		t.Fatalf("unexpected unified browser click display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "interaction" || payload.Surface.SummaryCode != "click_completed" {
		t.Fatalf("unexpected unified browser click surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "interaction" || payload.View.SummaryCode != "click_completed" {
		t.Fatalf("unexpected unified browser click view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserClickRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-click-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				clickResult: BrowserClickResult{
					Backend:    "proxy-click",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/workbench",
					Title:      "Workbench",
					Status:     "clicked",
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Click: true,
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
		EnabledTools:    []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-click-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"click","selector":"button.buy","post_wait_ms":500}`,
	})
	if err != nil {
		t.Fatalf("browser unified click managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.clickReqs) != 0 {
		t.Fatalf("expected host backend click to stay unused, got %#v", hostBackend.clickReqs)
	}
	if len(nodeBackend.clickReqs) != 1 || nodeBackend.clickReqs[0].TabIndex != 2 || nodeBackend.clickReqs[0].URL != "" || nodeBackend.clickReqs[0].Selector != "button.buy" || nodeBackend.clickReqs[0].PostWaitMs != 500 || nodeBackend.clickReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs || nodeBackend.clickReqs[0].PreferredTargetID != tracked.ID {
		t.Fatalf("expected unified browser click to reuse managed current route before implicit host fallback, got %#v", nodeBackend.clickReqs)
	}
	if nodeBackend.clickReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser click to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.clickReqs[0])
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		Target        string                         `json:"target"`
		TargetID      string                         `json:"target_id"`
		FinalURL      string                         `json:"final_url"`
		TabIndex      int                            `json:"tab_index"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser click output: %v", err)
	}
	if payload.Kind != "click" || payload.Backend != "proxy-click" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "clicked" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.FinalURL != "https://node.example/workbench" || payload.TabIndex != 2 {
		t.Fatalf("unexpected unified browser click payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "interaction" || payload.Summary.SummaryCode != "click_completed" {
		t.Fatalf("unexpected unified browser click summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "interaction" || payload.Display.SummaryCode != "click_completed" {
		t.Fatalf("unexpected unified browser click display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "interaction" || payload.Surface.SummaryCode != "click_completed" {
		t.Fatalf("unexpected unified browser click surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "interaction" || payload.View.SummaryCode != "click_completed" {
		t.Fatalf("unexpected unified browser click view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserListTabsRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"list_tabs"}`,
	})
	if err != nil {
		t.Fatalf("browser unified list_tabs managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "list" || nodeBackend.tabsReqs[0].TabIndex != 0 {
		t.Fatalf("expected unified browser list_tabs to drive managed default route before implicit host fallback, got %#v", nodeBackend.tabsReqs)
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
		t.Fatalf("decode unified browser list_tabs output: %v", err)
	}
	if payload.Kind != "list_tabs" || payload.Action != "list" || payload.Backend != "proxy-tabs" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.TabIndex != 0 {
		t.Fatalf("unexpected unified browser list_tabs payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "list_tabs_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "list_tabs_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "list_tabs_completed" ||
		payload.View == nil || payload.View.SummaryCode != "list_tabs_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected unified browser list_tabs to expose stable tabs summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserFocusRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"focus","tab_index":2,"wait_ms":120}`,
	})
	if err != nil {
		t.Fatalf("browser unified focus managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "focus" || nodeBackend.tabsReqs[0].TabIndex != 2 || nodeBackend.tabsReqs[0].WaitMs != 120 {
		t.Fatalf("expected unified browser focus to drive managed default route before implicit host fallback, got %#v", nodeBackend.tabsReqs)
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
		t.Fatalf("decode unified browser focus output: %v", err)
	}
	if payload.Kind != "focus_tab" || payload.Backend != "proxy-tabs" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "focus" || payload.Status != "focused" || payload.TabIndex != 2 {
		t.Fatalf("unexpected unified browser focus payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "focus_tab_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "focus_tab_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "focus_tab_completed" ||
		payload.View == nil || payload.View.SummaryCode != "focus_tab_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected unified browser focus to expose stable tabs summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserCloseRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"close","tab_index":2}`,
	})
	if err != nil {
		t.Fatalf("browser unified close managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "close" || nodeBackend.tabsReqs[0].TabIndex != 2 {
		t.Fatalf("expected unified browser close to drive managed default route before implicit host fallback, got %#v", nodeBackend.tabsReqs)
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
		t.Fatalf("decode unified browser close output: %v", err)
	}
	if payload.Kind != "close_tab" || payload.Backend != "proxy-tabs" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "close" || payload.Status != "closed" || payload.TabIndex != 2 {
		t.Fatalf("unexpected unified browser close payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "close_tab_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "close_tab_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "close_tab_completed" ||
		payload.View == nil || payload.View.SummaryCode != "close_tab_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected unified browser close to expose stable tabs summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserExtractRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"extract","max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser unified extract managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 0 || nodeBackend.extractReqs[0].URL != "" || nodeBackend.extractReqs[0].WaitMs != browserTabTargetWaitMs || nodeBackend.extractReqs[0].MaxChars != 32 {
		t.Fatalf("expected unified browser extract to drive managed default route before implicit host fallback, got %#v", nodeBackend.extractReqs)
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
		t.Fatalf("decode unified browser extract output: %v", err)
	}
	if payload.Kind != "extract" || payload.Backend != "proxy-extract" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "extracted" || payload.FinalURL != "https://node.example/default" || payload.TabIndex != 0 {
		t.Fatalf("unexpected unified browser extract payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "extract_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "extract_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "extract_completed" ||
		payload.View == nil || payload.View.SummaryCode != "extract_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected unified browser extract to expose stable extract summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserSnapshotRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"snapshot","max_chars":24,"max_elements":6,"format":"ai","mode":"efficient"}`,
	})
	if err != nil {
		t.Fatalf("browser unified snapshot managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.snapshotReqs) != 1 || nodeBackend.snapshotReqs[0].TabIndex != 0 || nodeBackend.snapshotReqs[0].URL != "" || nodeBackend.snapshotReqs[0].WaitMs != 250 || nodeBackend.snapshotReqs[0].MaxChars != 24 || nodeBackend.snapshotReqs[0].MaxElements != 6 || nodeBackend.snapshotReqs[0].Format != "ai" {
		t.Fatalf("expected unified browser snapshot to drive managed default route before implicit host fallback, got %#v", nodeBackend.snapshotReqs)
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
		t.Fatalf("decode unified browser snapshot output: %v", err)
	}
	if payload.Kind != "snapshot" || payload.Backend != "proxy-snapshot" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.FinalURL != "https://node.example/default" || payload.Snapshot != "managed default snapshot" || payload.TabIndex != 0 {
		t.Fatalf("unexpected unified browser snapshot payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "content" || payload.Summary.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected unified browser snapshot summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "content" || payload.Display.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected unified browser snapshot display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "content" || payload.Surface.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected unified browser snapshot surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "content" || payload.View.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected unified browser snapshot view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserScreenshotRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools:    []string{"browser"},
		PublishArtifact: testBrowserArtifactPublisher,
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"screenshot"}`,
	})
	if err != nil {
		t.Fatalf("browser unified screenshot managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.screenshotReqs) != 1 || nodeBackend.screenshotReqs[0].TabIndex != 0 || nodeBackend.screenshotReqs[0].URL != "" || nodeBackend.screenshotReqs[0].WaitMs != browserTabTargetWaitMs || strings.TrimSpace(nodeBackend.screenshotReqs[0].OutputPath) == "" {
		t.Fatalf("expected unified browser screenshot to drive managed default route before implicit host fallback, got %#v", nodeBackend.screenshotReqs)
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
		t.Fatalf("decode unified browser screenshot output: %v", err)
	}
	if payload.Kind != "screenshot" || payload.Backend != "proxy-screenshot" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "captured" || payload.FinalURL != "https://node.example/default" || payload.TabIndex != 0 || strings.TrimSpace(payload.Path) == "" {
		t.Fatalf("unexpected unified browser screenshot payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "screenshot_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "screenshot_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "screenshot_completed" ||
		payload.View == nil || payload.View.SummaryCode != "screenshot_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected unified browser screenshot to expose stable screenshot summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserTypeRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"type","selector":"input[name=q]","text":"agentx","submit":true}`,
	})
	if err != nil {
		t.Fatalf("browser unified type managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.typeReqs) != 1 || nodeBackend.typeReqs[0].TabIndex != 0 || nodeBackend.typeReqs[0].URL != "" || nodeBackend.typeReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs || nodeBackend.typeReqs[0].PostWaitMs != 250 || nodeBackend.typeReqs[0].Selector != "input[name=q]" || nodeBackend.typeReqs[0].Text != "agentx" || !nodeBackend.typeReqs[0].Submit {
		t.Fatalf("expected unified browser type to drive managed default route before implicit host fallback, got %#v", nodeBackend.typeReqs)
	}
	var payload struct {
		Kind          string `json:"kind"`
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
		FinalURL      string `json:"final_url"`
		TabIndex      int    `json:"tab_index"`
		Value         string `json:"value"`
		Submitted     bool   `json:"submitted"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser type output: %v", err)
	}
	if payload.Kind != "type" || payload.Backend != "proxy-type" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "typed" || payload.FinalURL != "https://node.example/form" || payload.TabIndex != 0 || payload.Value != "agentx" || !payload.Submitted {
		t.Fatalf("unexpected unified browser type payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserEvaluateRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"evaluate","script":"document.title","max_chars":24}`,
	})
	if err != nil {
		t.Fatalf("browser unified evaluate managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.evalReqs) != 1 || nodeBackend.evalReqs[0].TabIndex != 0 || nodeBackend.evalReqs[0].URL != "" || nodeBackend.evalReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs || nodeBackend.evalReqs[0].Script != "document.title" || nodeBackend.evalReqs[0].MaxChars != 24 {
		t.Fatalf("expected unified browser evaluate to drive managed default route before implicit host fallback, got %#v", nodeBackend.evalReqs)
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
		t.Fatalf("decode unified browser evaluate output: %v", err)
	}
	if payload.Kind != "evaluate" || payload.Backend != "proxy-eval" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "evaluated" || payload.FinalURL != "https://node.example/default" || payload.TabIndex != 0 || payload.Result != "managed default eval" {
		t.Fatalf("unexpected unified browser evaluate payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "script" || payload.Summary.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected unified browser evaluate summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "script" || payload.Display.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected unified browser evaluate display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "script" || payload.Surface.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected unified browser evaluate surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "script" || payload.View.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected unified browser evaluate view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserTabsAliasRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-tabs-managed-current-hidden-implicit-host")
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
						{Index: 1, Title: "Docs", URL: "https://docs.example", Active: false},
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
		EnabledTools:    []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-tabs-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"tabs"}`,
	})
	if err != nil {
		t.Fatalf("browser tabs managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.tabsReqs) != 0 {
		t.Fatalf("expected host backend tabs list to stay unused, got %#v", hostBackend.tabsReqs)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "list" || nodeBackend.tabsReqs[0].TabIndex != 2 {
		t.Fatalf("expected unified browser tabs alias to reuse managed current route before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	if nodeBackend.tabsReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser tabs alias to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.tabsReqs[0])
	}
	var payload struct {
		Kind          string       `json:"kind"`
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
		t.Fatalf("decode unified browser tabs alias output: %v", err)
	}
	if payload.Kind != "list_tabs" || payload.Backend != "proxy-tabs" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "list" || payload.TabIndex != 2 || payload.ActiveIndex != 2 || len(payload.Tabs) != 2 {
		t.Fatalf("unexpected unified browser tabs alias payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserFocusRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-focus-managed-current-hidden-implicit-host")
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
		EnabledTools:    []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-focus-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   4,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"focus","tab_index":3,"wait_ms":120}`,
	})
	if err != nil {
		t.Fatalf("browser focus managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.tabsReqs) != 0 {
		t.Fatalf("expected host backend tabs focus to stay unused, got %#v", hostBackend.tabsReqs)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "focus" || nodeBackend.tabsReqs[0].TabIndex != 3 || nodeBackend.tabsReqs[0].WaitMs != 120 {
		t.Fatalf("expected unified browser focus to reuse managed current route before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	if nodeBackend.tabsReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser focus to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.tabsReqs[0])
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
		t.Fatalf("decode unified browser focus output: %v", err)
	}
	if payload.Kind != "focus_tab" || payload.Backend != "proxy-tabs" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "focus" || payload.Status != "focused" || payload.Target != "tab:3" || payload.TabIndex != 3 {
		t.Fatalf("unexpected unified browser focus payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserCloseRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-close-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
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
		EnabledTools:    []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-close-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"close","tab_index":2}`,
	})
	if err != nil {
		t.Fatalf("browser close managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.tabsReqs) != 0 {
		t.Fatalf("expected host backend tabs close to stay unused, got %#v", hostBackend.tabsReqs)
	}
	if len(nodeBackend.tabsReqs) != 1 || nodeBackend.tabsReqs[0].Action != "close" || nodeBackend.tabsReqs[0].TabIndex != 2 {
		t.Fatalf("expected unified browser close to reuse managed current route before implicit host fallback, got %#v", nodeBackend.tabsReqs)
	}
	if nodeBackend.tabsReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser close to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.tabsReqs[0])
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
		t.Fatalf("decode unified browser close output: %v", err)
	}
	if payload.Kind != "close_tab" || payload.Backend != "proxy-tabs" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Action != "close" || payload.Status != "closed" || payload.Target != "tab:2" || payload.TabIndex != 2 {
		t.Fatalf("unexpected unified browser close payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserExtractRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-extract-managed-current-hidden-implicit-host")
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
		EnabledTools:    []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-extract-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Name:      "browser",
		Arguments: `{"action":"extract","max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser extract managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 2 || nodeBackend.extractReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected unified browser extract to reuse managed current route before implicit host fallback, got %#v", nodeBackend.extractReqs)
	}
	if nodeBackend.extractReqs[0].BrowserApp != "Chromium" || nodeBackend.extractReqs[0].URL != "" {
		t.Fatalf("expected unified browser extract to preserve managed browser app and current-target tab routing before backend dispatch, got %#v", nodeBackend.extractReqs[0])
	}
	var payload struct {
		Kind          string `json:"kind"`
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
		Target        string `json:"target"`
		TargetID      string `json:"target_id"`
		FinalURL      string `json:"final_url"`
		TabIndex      int    `json:"tab_index"`
		Content       string `json:"content"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser extract output: %v", err)
	}
	if payload.Kind != "extract" || payload.Backend != "proxy-extract" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "extracted" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.FinalURL != publicExampleIPURL || payload.TabIndex != 2 || payload.Content != "managed current content" {
		t.Fatalf("unexpected unified browser extract payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserSnapshotRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-snapshot-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				snapshotResult: BrowserSnapshotResult{
					Backend:    "proxy-snapshot",
					BrowserApp: "Chromium",
					FinalURL:   publicExampleIPURL,
					Title:      "Workbench",
					Snapshot:   "managed current snapshot",
					Format:     "ai",
					Mode:       "efficient",
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
		EnabledTools:    []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-snapshot-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Name:      "browser",
		Arguments: `{"action":"snapshot","max_chars":24,"max_elements":6,"format":"ai","mode":"efficient"}`,
	})
	if err != nil {
		t.Fatalf("browser snapshot managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.snapshotReqs) != 0 {
		t.Fatalf("expected host backend snapshot to stay unused, got %#v", hostBackend.snapshotReqs)
	}
	if len(nodeBackend.snapshotReqs) != 1 || nodeBackend.snapshotReqs[0].TabIndex != 2 || nodeBackend.snapshotReqs[0].URL != "" || nodeBackend.snapshotReqs[0].WaitMs != browserTabTargetWaitMs || nodeBackend.snapshotReqs[0].MaxChars != 24 || nodeBackend.snapshotReqs[0].MaxElements != 6 || nodeBackend.snapshotReqs[0].Format != "ai" {
		t.Fatalf("expected unified browser snapshot to reuse managed current route before implicit host fallback, got %#v", nodeBackend.snapshotReqs)
	}
	if nodeBackend.snapshotReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser snapshot to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.snapshotReqs[0])
	}
	var payload struct {
		Kind          string `json:"kind"`
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
		Target        string `json:"target"`
		TargetID      string `json:"target_id"`
		FinalURL      string `json:"final_url"`
		TabIndex      int    `json:"tab_index"`
		Snapshot      string `json:"snapshot"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser snapshot output: %v", err)
	}
	if payload.Kind != "snapshot" || payload.Backend != "proxy-snapshot" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "snapshotted" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.FinalURL != publicExampleIPURL || payload.TabIndex != 2 || payload.Snapshot != "managed current snapshot" {
		t.Fatalf("unexpected unified browser snapshot payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserScreenshotRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	root := t.TempDir()
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-screenshot-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				screenshotResult: BrowserScreenshotResult{
					Backend:       "proxy-screenshot",
					BrowserApp:    "Chromium",
					FinalURL:      publicExampleIPURL,
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
		Root:            root,
		Backend:         hostBackend,
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser"},
		PublishArtifact: testBrowserArtifactPublisher,
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-screenshot-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Name:      "browser",
		Arguments: `{"action":"screenshot"}`,
	})
	if err != nil {
		t.Fatalf("browser screenshot managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.screenshotReqs) != 0 {
		t.Fatalf("expected host backend screenshot to stay unused, got %#v", hostBackend.screenshotReqs)
	}
	if len(nodeBackend.screenshotReqs) != 1 || nodeBackend.screenshotReqs[0].TabIndex != 2 || nodeBackend.screenshotReqs[0].URL != "" || nodeBackend.screenshotReqs[0].WaitMs != browserTabTargetWaitMs || strings.TrimSpace(nodeBackend.screenshotReqs[0].OutputPath) == "" {
		t.Fatalf("expected unified browser screenshot to reuse managed current route before implicit host fallback, got %#v", nodeBackend.screenshotReqs)
	}
	if nodeBackend.screenshotReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser screenshot to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.screenshotReqs[0])
	}
	if !strings.HasPrefix(filepath.ToSlash(nodeBackend.screenshotReqs[0].OutputPath), filepath.ToSlash(root)+"/.agentx/browser/") {
		t.Fatalf("expected unified browser screenshot path within workspace, got %s", nodeBackend.screenshotReqs[0].OutputPath)
	}
	var payload struct {
		Kind          string `json:"kind"`
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
		Target        string `json:"target"`
		TargetID      string `json:"target_id"`
		FinalURL      string `json:"final_url"`
		TabIndex      int    `json:"tab_index"`
		Path          string `json:"path"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser screenshot output: %v", err)
	}
	if payload.Kind != "screenshot" || payload.Backend != "proxy-screenshot" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "captured" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.FinalURL != publicExampleIPURL || payload.TabIndex != 2 || strings.TrimSpace(payload.Path) == "" {
		t.Fatalf("unexpected unified browser screenshot payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserTypeRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34/form"
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-type-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				typeResult: BrowserTypeResult{
					Backend:    "proxy-type",
					BrowserApp: "Chromium",
					FinalURL:   publicExampleIPURL,
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
		EnabledTools:    []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-type-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        publicExampleIPURL,
		Title:      "Form",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"type","selector":"input[name=q]","text":"agentx","submit":true}`,
	})
	if err != nil {
		t.Fatalf("browser type managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.typeReqs) != 0 {
		t.Fatalf("expected host backend type to stay unused, got %#v", hostBackend.typeReqs)
	}
	if len(nodeBackend.typeReqs) != 1 || nodeBackend.typeReqs[0].TabIndex != 2 || nodeBackend.typeReqs[0].URL != "" || nodeBackend.typeReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs || nodeBackend.typeReqs[0].PostWaitMs != 250 || nodeBackend.typeReqs[0].Selector != "input[name=q]" || nodeBackend.typeReqs[0].Text != "agentx" || !nodeBackend.typeReqs[0].Submit || nodeBackend.typeReqs[0].PreferredTargetID != tracked.ID {
		t.Fatalf("expected unified browser type to reuse managed current route before implicit host fallback, got %#v", nodeBackend.typeReqs)
	}
	if nodeBackend.typeReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser type to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.typeReqs[0])
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		Target        string                         `json:"target"`
		TargetID      string                         `json:"target_id"`
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
		t.Fatalf("decode unified browser type output: %v", err)
	}
	if payload.Kind != "type" || payload.Backend != "proxy-type" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "typed" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.FinalURL != publicExampleIPURL || payload.TabIndex != 2 || payload.Value != "agentx" || !payload.Submitted {
		t.Fatalf("unexpected unified browser type payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "form" || payload.Summary.SummaryCode != "type_completed" {
		t.Fatalf("unexpected unified browser type summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "form" || payload.Display.SummaryCode != "type_completed" {
		t.Fatalf("unexpected unified browser type display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "form" || payload.Surface.SummaryCode != "type_completed" {
		t.Fatalf("unexpected unified browser type surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "form" || payload.View.SummaryCode != "type_completed" {
		t.Fatalf("unexpected unified browser type view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserEvaluateRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34/workbench"
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-evaluate-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				evalResult: BrowserEvalResult{
					Backend:    "proxy-eval",
					BrowserApp: "Chromium",
					FinalURL:   publicExampleIPURL,
					Title:      "Workbench",
					Result:     "managed current eval",
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
		EnabledTools:    []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-evaluate-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Name:      "browser",
		Arguments: `{"action":"evaluate","script":"document.title","max_chars":24}`,
	})
	if err != nil {
		t.Fatalf("browser evaluate managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.evalReqs) != 0 {
		t.Fatalf("expected host backend evaluate to stay unused, got %#v", hostBackend.evalReqs)
	}
	if len(nodeBackend.evalReqs) != 1 || nodeBackend.evalReqs[0].TabIndex != 2 || nodeBackend.evalReqs[0].URL != "" || nodeBackend.evalReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs || nodeBackend.evalReqs[0].Script != "document.title" || nodeBackend.evalReqs[0].MaxChars != 24 || nodeBackend.evalReqs[0].PreferredTargetID != tracked.ID {
		t.Fatalf("expected unified browser evaluate to reuse managed current route before implicit host fallback, got %#v", nodeBackend.evalReqs)
	}
	if nodeBackend.evalReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser evaluate to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.evalReqs[0])
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		Target        string                         `json:"target"`
		TargetID      string                         `json:"target_id"`
		FinalURL      string                         `json:"final_url"`
		TabIndex      int                            `json:"tab_index"`
		Result        string                         `json:"result"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser evaluate output: %v", err)
	}
	if payload.Kind != "evaluate" || payload.Backend != "proxy-eval" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "evaluated" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.FinalURL != publicExampleIPURL || payload.TabIndex != 2 || payload.Result != "managed current eval" {
		t.Fatalf("unexpected unified browser evaluate payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "script" || payload.Summary.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected unified browser evaluate summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "script" || payload.Display.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected unified browser evaluate display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "script" || payload.Surface.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected unified browser evaluate surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "script" || payload.View.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected unified browser evaluate view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserActConsoleRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-act-console-managed-current-hidden-implicit-host")
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
		EnabledTools:    []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-act-console-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"act","kind":"console","level":"error"}`,
	})
	if err != nil {
		t.Fatalf("browser act console managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.consoleReqs) != 0 {
		t.Fatalf("expected host backend console to stay unused, got %#v", hostBackend.consoleReqs)
	}
	if len(nodeBackend.consoleReqs) != 1 || nodeBackend.consoleReqs[0].Level != "error" || nodeBackend.consoleReqs[0].TabIndex != 2 || nodeBackend.consoleReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected unified browser act console to reuse managed current route before implicit host fallback, got %#v", nodeBackend.consoleReqs)
	}
	if nodeBackend.consoleReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser act console to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.consoleReqs[0])
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		Target        string                         `json:"target"`
		TargetID      string                         `json:"target_id"`
		FinalURL      string                         `json:"final_url"`
		TabIndex      int                            `json:"tab_index"`
		Messages      []BrowserConsoleMessage        `json:"messages"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser act console output: %v", err)
	}
	if payload.Kind != "console" || payload.Backend != "proxy-console" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "ok" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.FinalURL != "https://node.example/workbench" || payload.TabIndex != 2 || len(payload.Messages) != 1 || payload.Messages[0].Level != "error" {
		t.Fatalf("unexpected unified browser act console payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "console_collected" {
		t.Fatalf("unexpected unified browser act console summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "console_collected" {
		t.Fatalf("unexpected unified browser act console display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "console_collected" {
		t.Fatalf("unexpected unified browser act console surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "console_collected" {
		t.Fatalf("unexpected unified browser act console view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserActRequestsRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-act-requests-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				requestsResult: BrowserRequestsResult{
					Backend:    "proxy-requests",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/workbench",
					Title:      "Workbench",
					Requests: []BrowserRequestEntry{
						{Method: "GET", URL: "https://node.example/api/items", Status: 200, ResourceType: "xhr"},
					},
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Requests: true,
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
		EnabledTools:    []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-act-requests-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"act","kind":"requests","filter":"api"}`,
	})
	if err != nil {
		t.Fatalf("browser act requests managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.requestsReqs) != 0 {
		t.Fatalf("expected host backend requests to stay unused, got %#v", hostBackend.requestsReqs)
	}
	if len(nodeBackend.requestsReqs) != 1 || nodeBackend.requestsReqs[0].Filter != "api" || nodeBackend.requestsReqs[0].TabIndex != 2 || nodeBackend.requestsReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected unified browser act requests to reuse managed current route before implicit host fallback, got %#v", nodeBackend.requestsReqs)
	}
	if nodeBackend.requestsReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected unified browser act requests to inherit managed browser app before implicit host fallback, got %#v", nodeBackend.requestsReqs[0])
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		Status        string                         `json:"status"`
		Target        string                         `json:"target"`
		TargetID      string                         `json:"target_id"`
		FinalURL      string                         `json:"final_url"`
		TabIndex      int                            `json:"tab_index"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
		Requests      []BrowserRequestEntry          `json:"requests"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser act requests output: %v", err)
	}
	if payload.Kind != "requests" || payload.Backend != "proxy-requests" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "ok" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.FinalURL != "https://node.example/workbench" || payload.TabIndex != 2 || len(payload.Requests) != 1 || payload.Requests[0].URL != "https://node.example/api/items" {
		t.Fatalf("unexpected unified browser act requests payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected unified browser act requests summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected unified browser act requests display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected unified browser act requests surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected unified browser act requests view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserActConsoleRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
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
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"act","kind":"console","level":"error"}`,
	})
	if err != nil {
		t.Fatalf("browser unified act console managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.consoleReqs) != 1 || nodeBackend.consoleReqs[0].Level != "error" || nodeBackend.consoleReqs[0].TabIndex != 0 {
		t.Fatalf("expected unified browser act console to drive managed default route before implicit host fallback, got %#v", nodeBackend.consoleReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		FinalURL      string                         `json:"final_url"`
		Status        string                         `json:"status"`
		Messages      []BrowserConsoleMessage        `json:"messages"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser act console output: %v", err)
	}
	if payload.Kind != "console" || payload.Backend != "proxy-console" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.FinalURL != "https://node.example/workbench" || payload.Status != "ok" || len(payload.Messages) != 1 || payload.Messages[0].Level != "error" {
		t.Fatalf("unexpected unified browser act console payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "console_collected" {
		t.Fatalf("unexpected unified browser act console summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "console_collected" {
		t.Fatalf("unexpected unified browser act console display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "console_collected" {
		t.Fatalf("unexpected unified browser act console surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "console_collected" {
		t.Fatalf("unexpected unified browser act console view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserActRequestsRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				requestsResult: BrowserRequestsResult{
					Backend:    "proxy-requests",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/workbench",
					Title:      "Workbench",
					Requests: []BrowserRequestEntry{
						{Method: "GET", URL: "https://node.example/api/items", Status: 200, ResourceType: "xhr"},
					},
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Requests: true,
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
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"act","kind":"requests","filter":"api"}`,
	})
	if err != nil {
		t.Fatalf("browser unified act requests managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.requestsReqs) != 1 || nodeBackend.requestsReqs[0].Filter != "api" || nodeBackend.requestsReqs[0].TabIndex != 0 {
		t.Fatalf("expected unified browser act requests to drive managed default route before implicit host fallback, got %#v", nodeBackend.requestsReqs)
	}
	var payload struct {
		Kind          string                         `json:"kind"`
		Backend       string                         `json:"backend"`
		BrowserApp    string                         `json:"browser_app"`
		Profile       string                         `json:"profile"`
		RuntimeTarget string                         `json:"runtime_target"`
		FinalURL      string                         `json:"final_url"`
		Status        string                         `json:"status"`
		Summary       *browserTopLevelSummary        `json:"summary"`
		Display       *browserTopLevelDisplaySummary `json:"display"`
		Surface       *browserTopLevelSurfaceSummary `json:"surface"`
		View          *browserTopLevelViewSummary    `json:"view"`
		Requests      []BrowserRequestEntry          `json:"requests"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified browser act requests output: %v", err)
	}
	if payload.Kind != "requests" || payload.Backend != "proxy-requests" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.FinalURL != "https://node.example/workbench" || payload.Status != "ok" || len(payload.Requests) != 1 || payload.Requests[0].URL != "https://node.example/api/items" {
		t.Fatalf("unexpected unified browser act requests payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected unified browser act requests summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected unified browser act requests display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected unified browser act requests surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected unified browser act requests view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserClickReturnsUnifiedExplanationAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				clickResult: BrowserClickResult{
					Backend:         "proxy-click",
					BrowserApp:      "Chromium",
					FinalURL:        "https://node.example/workbench",
					Status:          "clicked",
					ResolverOutcome: browserBuiltinSharedFallbackOutcomeForTest(),
				},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilitiesForActKinds([]string{"click"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			if !strings.EqualFold(strings.TrimSpace(requested.Profile), "workbench") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"click","runtime_target":"node","profile":"workbench","selector":"button.buy"}`,
	})
	if err != nil {
		t.Fatalf("browser unified click explanation alias: %v", err)
	}
	var payload struct {
		ResolverExplanation         *browserRuntimeResolverExplanationSummary  `json:"resolver_explanation"`
		DiagnosticsExplanation      *browserDiagnosticsExplanationSummary      `json:"diagnostics_explanation"`
		Explanation                 *browserTopLevelSummary                    `json:"explanation"`
		Diagnostics                 *browserTopLevelSummary                    `json:"diagnostics"`
		ResolverFallbackExplanation *browserResolverFallbackExplanationSummary `json:"resolver_fallback_explanation"`
		Summary                     *browserTopLevelSummary                    `json:"summary"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified click explanation alias output: %v", err)
	}
	if payload.ResolverFallbackExplanation == nil ||
		payload.ResolverFallbackExplanation.State != "resolved_via_fallback" ||
		payload.ResolverFallbackExplanation.SummaryCode != "label_filtered_residual" {
		t.Fatalf("unexpected unified click fallback explanation: %#v", payload.ResolverFallbackExplanation)
	}
	if payload.ResolverExplanation == nil ||
		payload.ResolverExplanation.State != "resolved_via_fallback" ||
		payload.ResolverExplanation.SummaryCode != "label_filtered_residual" ||
		payload.ResolverExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected unified click resolver explanation: %#v", payload.ResolverExplanation)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "resolver_fallback" ||
		payload.DiagnosticsExplanation.State != "resolved_via_fallback" ||
		payload.DiagnosticsExplanation.SummaryCode != "label_filtered_residual" ||
		payload.DiagnosticsExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected unified click diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "resolver_fallback" ||
		payload.Explanation.State != "resolved_via_fallback" ||
		payload.Explanation.SummaryCode != "label_filtered_residual" ||
		payload.Explanation.ManualRetryHint != "add_ordinal" ||
		!payload.Explanation.ResolvedViaFallback {
		t.Fatalf("unexpected unified click top-level explanation alias: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "resolver_fallback" ||
		payload.Diagnostics.State != "resolved_via_fallback" ||
		payload.Diagnostics.SummaryCode != "label_filtered_residual" ||
		payload.Diagnostics.ManualRetryHint != "add_ordinal" ||
		!payload.Diagnostics.ResolvedViaFallback {
		t.Fatalf("unexpected unified click top-level diagnostics alias: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "resolver_fallback" ||
		payload.Summary.State != "resolved_via_fallback" ||
		payload.Summary.SummaryCode != "label_filtered_residual" ||
		payload.Summary.ManualRetryHint != "add_ordinal" ||
		!payload.Summary.ResolvedViaFallback {
		t.Fatalf("unexpected unified click top-level summary alias: %#v", payload.Summary)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserRejectsUnsupportedManagedActOutsideStaticActionEnum(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"click"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			if !strings.EqualFold(strings.TrimSpace(requested.Profile), "workbench") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"hover","runtime_target":"node","profile":"workbench","selector":"button.buy"}`,
	})
	if err == nil || !strings.Contains(err.Error(), "action must be one of") {
		t.Fatalf("expected unified browser to reject unsupported managed hover action early, got %v", err)
	}
	if len(hostBackend.hoverReqs) != 0 {
		t.Fatalf("expected unsupported unified hover action to avoid host backend, got %#v", hostBackend.hoverReqs)
	}
	if len(nodeBackend.hoverReqs) != 0 {
		t.Fatalf("expected unsupported unified hover action to avoid node backend, got %#v", nodeBackend.hoverReqs)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserClickUsesExplicitManagedActOptInSurfaceWithoutVisibleActSurface(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				clickResult: BrowserClickResult{Backend: "proxy-click", BrowserApp: "Chromium", FinalURL: "https://node.example/workbench", Status: "clicked"},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilitiesForActKinds([]string{"click"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			if !strings.EqualFold(strings.TrimSpace(requested.Profile), "workbench") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"click","runtime_target":"node","profile":"workbench","selector":"button.buy"}`,
	})
	if err != nil {
		t.Fatalf("browser explicit managed opt-in unified click: %v", err)
	}
	if len(hostBackend.clickReqs) != 0 {
		t.Fatalf("expected unified explicit managed opt-in click to avoid host backend, got %#v", hostBackend.clickReqs)
	}
	if len(nodeBackend.clickReqs) != 1 ||
		nodeBackend.clickReqs[0].Selector != "button.buy" ||
		nodeBackend.clickReqs[0].TabIndex != 0 ||
		nodeBackend.clickReqs[0].PostWaitMs != 750 {
		t.Fatalf("unexpected unified explicit managed opt-in click dispatch: %#v", nodeBackend.clickReqs)
	}
	var payload struct {
		Kind                string                         `json:"kind"`
		Backend             string                         `json:"backend"`
		BrowserApp          string                         `json:"browser_app"`
		Profile             string                         `json:"profile"`
		RuntimeTarget       string                         `json:"runtime_target"`
		Status              string                         `json:"status"`
		BrowserTools        []string                       `json:"browser_tools"`
		ArtifactTools       []string                       `json:"artifact_tools"`
		ArtifactKinds       []string                       `json:"artifact_kinds"`
		ArtifactContract    string                         `json:"artifact_contract"`
		BrowserActKinds     []string                       `json:"browser_act_kinds"`
		BrowserSurface      string                         `json:"browser_surface"`
		BrowserOptInTargets []string                       `json:"browser_opt_in_targets"`
		Summary             *browserTopLevelSummary        `json:"summary"`
		Display             *browserTopLevelDisplaySummary `json:"display"`
		Surface             *browserTopLevelSurfaceSummary `json:"surface"`
		View                *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified explicit managed opt-in click output: %v", err)
	}
	if payload.Kind != "click" || payload.Backend != "proxy-click" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "clicked" {
		t.Fatalf("unexpected unified explicit managed opt-in click payload: %#v", payload)
	}
	if payload.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.BrowserOptInTargets) != 1 ||
		payload.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected unified explicit managed opt-in click payload to expose route surface, got %#v", payload)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_runtime") ||
		!browserStringSliceContains(payload.BrowserTools, "browser_act") ||
		browserStringSliceContains(payload.ArtifactTools, "browser_screenshot") ||
		browserStringSliceContains(payload.ArtifactKinds, "screenshot") ||
		payload.ArtifactContract != "" ||
		!browserStringSliceContains(payload.BrowserActKinds, "click") {
		t.Fatalf("expected unified explicit managed opt-in click payload to expose selected-route capabilities, got %#v", payload)
	}
	if payload.Summary == nil ||
		payload.Summary.PrimaryBrowserAction != "browser action=click" ||
		payload.Summary.NextStep != "browser action=click" ||
		payload.Display == nil ||
		payload.Display.PrimaryBrowserAction != "browser action=click" ||
		payload.Display.NextStep != "browser action=click" ||
		payload.Surface == nil ||
		payload.Surface.PrimaryBrowserAction != "browser action=click" ||
		payload.Surface.NextStep != "browser action=click" ||
		payload.View == nil ||
		payload.View.PrimaryBrowserAction != "browser action=click" ||
		payload.View.NextStep != "browser action=click" {
		t.Fatalf("expected unified explicit managed opt-in click payload to expose success action hints, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserClickAllowsExplicitManagedLaneOutsideStaticActionEnum(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				clickResult: BrowserClickResult{Backend: "proxy-click", BrowserApp: "Chromium", FinalURL: "https://node.example/workbench", Status: "clicked"},
			},
			runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilitiesForActKinds([]string{"click"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			if !strings.EqualFold(strings.TrimSpace(requested.Profile), "workbench") {
				return BrowserRuntimeInfo{}, context.DeadlineExceeded
			}
			return BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"}, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	rawActions := browserUnifiedDefinitionActions(reg)
	if len(rawActions) == 0 {
		t.Fatalf("expected browser registration")
	}
	if browserStringSliceContains(rawActions, "click") {
		t.Fatalf("expected static unified action enum to remain conservative, got %#v", rawActions)
	}

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"click","runtime_target":"node","profile":"workbench","selector":"button.buy"}`,
	})
	if err != nil {
		t.Fatalf("browser unified explicit managed click outside static action enum: %v", err)
	}
	if len(hostBackend.clickReqs) != 0 {
		t.Fatalf("expected unified explicit managed click to avoid host backend, got %#v", hostBackend.clickReqs)
	}
	if len(nodeBackend.clickReqs) != 1 ||
		nodeBackend.clickReqs[0].Selector != "button.buy" ||
		nodeBackend.clickReqs[0].TabIndex != 0 ||
		nodeBackend.clickReqs[0].PostWaitMs != 750 {
		t.Fatalf("unexpected unified explicit managed click dispatch: %#v", nodeBackend.clickReqs)
	}
	var payload struct {
		Kind          string `json:"kind"`
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified explicit managed click output: %v", err)
	}
	if payload.Kind != "click" || payload.Backend != "proxy-click" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "clicked" {
		t.Fatalf("unexpected unified explicit managed click payload: %#v", payload)
	}
}
