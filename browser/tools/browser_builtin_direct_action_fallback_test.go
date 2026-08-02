package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ClickRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser_click"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_click",
		Arguments: `{"selector":"button.buy","post_wait_ms":500}`,
	})
	if err != nil {
		t.Fatalf("browser_click managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.clickReqs) != 1 || nodeBackend.clickReqs[0].TabIndex != 0 || nodeBackend.clickReqs[0].URL != "" || nodeBackend.clickReqs[0].Selector != "button.buy" || nodeBackend.clickReqs[0].PostWaitMs != 500 || nodeBackend.clickReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs {
		t.Fatalf("expected managed default click to drive node backend before implicit host fallback, got %#v", nodeBackend.clickReqs)
	}
	var payload struct {
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
		t.Fatalf("decode managed default click output: %v", err)
	}
	if payload.Backend != "proxy-click" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "clicked" || payload.FinalURL != "https://node.example/after-click" || payload.TabIndex != 0 {
		t.Fatalf("unexpected managed default click payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "interaction" || payload.Summary.SummaryCode != "click_completed" {
		t.Fatalf("unexpected managed default click summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "interaction" || payload.Display.SummaryCode != "click_completed" {
		t.Fatalf("unexpected managed default click display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "interaction" || payload.Surface.SummaryCode != "click_completed" {
		t.Fatalf("unexpected managed default click surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "interaction" || payload.View.SummaryCode != "click_completed" {
		t.Fatalf("unexpected managed default click view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_ScreenshotRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools:    []string{"browser_screenshot"},
		PublishArtifact: testBrowserArtifactPublisher,
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_screenshot",
		Arguments: `{}`,
	})
	if err != nil {
		t.Fatalf("browser_screenshot managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.screenshotReqs) != 1 || nodeBackend.screenshotReqs[0].TabIndex != 0 || nodeBackend.screenshotReqs[0].URL != "" || nodeBackend.screenshotReqs[0].WaitMs != browserTabTargetWaitMs || strings.TrimSpace(nodeBackend.screenshotReqs[0].OutputPath) == "" {
		t.Fatalf("expected managed default screenshot to drive node backend before implicit host fallback, got %#v", nodeBackend.screenshotReqs)
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
		FinalURL      string `json:"final_url"`
		TabIndex      int    `json:"tab_index"`
		WaitMs        int    `json:"wait_ms"`
		Path          string `json:"path"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default screenshot output: %v", err)
	}
	if payload.Backend != "proxy-screenshot" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "captured" || payload.FinalURL != "https://node.example/default" || payload.TabIndex != 0 || payload.WaitMs != browserTabTargetWaitMs || strings.TrimSpace(payload.Path) == "" {
		t.Fatalf("unexpected managed default screenshot payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_TypeRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser_type"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_type",
		Arguments: `{"selector":"input[name=q]","text":"agentx","submit":true}`,
	})
	if err != nil {
		t.Fatalf("browser_type managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.typeReqs) != 1 || nodeBackend.typeReqs[0].TabIndex != 0 || nodeBackend.typeReqs[0].URL != "" || nodeBackend.typeReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs || nodeBackend.typeReqs[0].PostWaitMs != 250 || nodeBackend.typeReqs[0].Selector != "input[name=q]" || nodeBackend.typeReqs[0].Text != "agentx" || !nodeBackend.typeReqs[0].Submit {
		t.Fatalf("expected managed default type to drive node backend before implicit host fallback, got %#v", nodeBackend.typeReqs)
	}
	var payload struct {
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
		t.Fatalf("decode managed default type output: %v", err)
	}
	if payload.Backend != "proxy-type" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "typed" || payload.FinalURL != "https://node.example/form" || payload.TabIndex != 0 || payload.Value != "agentx" || !payload.Submitted {
		t.Fatalf("unexpected managed default type payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "form" || payload.Summary.SummaryCode != "type_completed" {
		t.Fatalf("unexpected managed default type summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "form" || payload.Display.SummaryCode != "type_completed" {
		t.Fatalf("unexpected managed default type display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "form" || payload.Surface.SummaryCode != "type_completed" {
		t.Fatalf("unexpected managed default type surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "form" || payload.View.SummaryCode != "type_completed" {
		t.Fatalf("unexpected managed default type view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_EvalRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser_eval"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_eval",
		Arguments: `{"script":"document.title","max_chars":24}`,
	})
	if err != nil {
		t.Fatalf("browser_eval managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.evalReqs) != 1 || nodeBackend.evalReqs[0].TabIndex != 0 || nodeBackend.evalReqs[0].URL != "" || nodeBackend.evalReqs[0].WaitMs != 0 || nodeBackend.evalReqs[0].Script != "document.title" || nodeBackend.evalReqs[0].MaxChars != 24 {
		t.Fatalf("expected managed default eval to drive node backend before implicit host fallback, got %#v", nodeBackend.evalReqs)
	}
	var payload struct {
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
		t.Fatalf("decode managed default eval output: %v", err)
	}
	if payload.Backend != "proxy-eval" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "evaluated" || payload.FinalURL != "https://node.example/default" || payload.TabIndex != 0 || payload.Result != "managed default eval" {
		t.Fatalf("unexpected managed default eval payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "script" || payload.Summary.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected managed default eval summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "script" || payload.Display.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected managed default eval display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "script" || payload.Surface.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected managed default eval surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "script" || payload.View.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected managed default eval view: %#v", payload.View)
	}
}
