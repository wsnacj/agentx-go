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

func TestRegisterBrowserTools_Open(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		openResult: BrowserOpenResult{Backend: "fake-open", BrowserApp: "Safari", Status: "opened"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_open"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_open",
		Arguments: `{"url":"https://93.184.216.34","browser":"Safari","wait_ms":1200}`,
	})
	if err != nil {
		t.Fatalf("browser_open: %v", err)
	}
	if len(backend.openReqs) != 1 || backend.openReqs[0].URL != "https://93.184.216.34" || backend.openReqs[0].WaitMs != 1200 {
		t.Fatalf("unexpected open request: %#v", backend.openReqs)
	}
	var payload struct {
		Backend     string                         `json:"backend"`
		BrowserApp  string                         `json:"browser_app"`
		Status      string                         `json:"status"`
		Explanation *browserTopLevelSummary        `json:"explanation"`
		Diagnostics *browserTopLevelSummary        `json:"diagnostics"`
		Summary     *browserTopLevelSummary        `json:"summary"`
		Display     *browserTopLevelDisplaySummary `json:"display"`
		Surface     *browserTopLevelSurfaceSummary `json:"surface"`
		View        *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Backend != "fake-open" || payload.BrowserApp != "Safari" || payload.Status != "opened" {
		t.Fatalf("unexpected open output: %#v", payload)
	}
	if payload.Explanation == nil || payload.Explanation.SummaryCode != "open_completed" ||
		payload.Diagnostics == nil || payload.Diagnostics.SummaryCode != "open_completed" ||
		payload.Summary == nil || payload.Summary.SummaryCode != "open_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "open_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "open_completed" ||
		payload.View == nil || payload.View.SummaryCode != "open_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_open to expose stable open summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_OpenContextNetworkGuardAllowsPrivateHost(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		openResult: BrowserOpenResult{Backend: "fake-open", BrowserApp: "Safari", Status: "opened"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_open"},
	})

	out, err := reg.Execute(WithToolRuntimeNetworkGuard(context.Background(), ToolRuntimeNetworkGuard{
		WebFetchAllowPrivateHosts: ToolRuntimeBool{Set: true, Value: true},
	}), types.FunctionCall{
		Name:      "browser_open",
		Arguments: `{"url":"http://127.0.0.1:8123","browser":"Safari","wait_ms":300}`,
	})
	if err != nil {
		t.Fatalf("expected runtime network guard to allow private host browser url, got %v", err)
	}
	if len(backend.openReqs) != 1 || backend.openReqs[0].URL != "http://127.0.0.1:8123" || backend.openReqs[0].WaitMs != 300 {
		t.Fatalf("unexpected open request: %#v", backend.openReqs)
	}
	if !strings.Contains(out, `"status":"opened"`) {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestRegisterBrowserTools_OpenTracksCurrentTargetForLaterExtract(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "system_open", BrowserApp: "Safari", Status: "opened"},
			extractResult: BrowserExtractResult{
				Backend:     "fake-extract",
				BrowserApp:  "Safari",
				Title:       "Opened Page",
				Content:     "tracked current content",
				FinalURL:    publicExampleIPURL,
				ContentType: "text/plain",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser_open", "browser_extract"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-open-track")
	openOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_open",
		Arguments: `{"url":"` + publicExampleIPURL + `"}`,
	})
	if err != nil {
		t.Fatalf("browser_open tracked current target: %v", err)
	}
	var openPayload struct {
		Target   string `json:"target"`
		TargetID string `json:"target_id"`
	}
	if err := json.Unmarshal([]byte(openOut), &openPayload); err != nil {
		t.Fatalf("decode open output: %v", err)
	}
	if openPayload.Target != "current" || strings.TrimSpace(openPayload.TargetID) == "" {
		t.Fatalf("expected tracked current target from browser_open, got %#v", openPayload)
	}

	extractOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_extract",
		Arguments: `{"target":"current","max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser_extract current after open: %v", err)
	}
	if len(backend.extractReqs) != 1 || backend.extractReqs[0].URL != publicExampleIPURL || backend.extractReqs[0].TabIndex != 0 {
		t.Fatalf("expected current extract to reuse tracked open url, got %#v", backend.extractReqs)
	}
	var extractPayload struct {
		FinalURL string `json:"final_url"`
		TargetID string `json:"target_id"`
	}
	if err := json.Unmarshal([]byte(extractOut), &extractPayload); err != nil {
		t.Fatalf("decode extract output: %v", err)
	}
	if extractPayload.FinalURL != publicExampleIPURL || strings.TrimSpace(extractPayload.TargetID) == "" {
		t.Fatalf("unexpected extract payload after open tracking: %#v", extractPayload)
	}
}

func TestRegisterBrowserTools_ExtractTruncatesContent(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:     "fake-extract",
			BrowserApp:  "Safari",
			Title:       "Example",
			Content:     strings.Repeat("x", 80),
			FinalURL:    "https://93.184.216.34/final",
			ContentType: "text/plain",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		MaxChars:     64,
		EnabledTools: []string{"browser_extract"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_extract",
		Arguments: `{"url":"` + publicExampleIPURL + `","max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser_extract: %v", err)
	}
	if len(backend.extractReqs) != 1 || backend.extractReqs[0].MaxChars != 32 {
		t.Fatalf("unexpected extract request: %#v", backend.extractReqs)
	}
	var payload struct {
		Backend   string                         `json:"backend"`
		FinalURL  string                         `json:"final_url"`
		Content   string                         `json:"content"`
		Truncated bool                           `json:"truncated"`
		Summary   *browserTopLevelSummary        `json:"summary"`
		Display   *browserTopLevelDisplaySummary `json:"display"`
		Surface   *browserTopLevelSurfaceSummary `json:"surface"`
		View      *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Backend != "fake-extract" || payload.FinalURL != "https://93.184.216.34/final" || len(payload.Content) != 32 || !payload.Truncated {
		t.Fatalf("unexpected extract output: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "extract_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "extract_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "extract_completed" ||
		payload.View == nil || payload.View.SummaryCode != "extract_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_extract to expose stable extract summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_ExtractTargetsCurrentWithoutURL(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:     "fake-extract",
			BrowserApp:  "Safari",
			Title:       "Current Tab",
			Content:     "tab content",
			FinalURL:    "https://93.184.216.34/current",
			ContentType: "text/plain",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		MaxChars:     64,
		EnabledTools: []string{"browser_extract"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_extract",
		Arguments: `{"target":"current","max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser_extract tab target: %v", err)
	}
	if len(backend.extractReqs) != 1 || backend.extractReqs[0].URL != "" || backend.extractReqs[0].TabIndex != 0 || backend.extractReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("unexpected extract tab request: %#v", backend.extractReqs)
	}
	var payload struct {
		FinalURL string `json:"final_url"`
		Target   string `json:"target"`
		TabIndex int    `json:"tab_index"`
		WaitMs   int    `json:"wait_ms"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.FinalURL != "https://93.184.216.34/current" || payload.Target != "current" || payload.TabIndex != 0 || payload.WaitMs != browserTabTargetWaitMs {
		t.Fatalf("unexpected extract tab output: %#v", payload)
	}
}

func TestRegisterBrowserTools_OpenRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools:    []string{"browser_open"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-open-managed-current-hidden-implicit-host")
	sessionRegistry.TrackCurrentTarget("browser-open-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Name:      "browser_open",
		Arguments: `{"url":"` + publicExampleIPURL + `"}`,
	})
	if err != nil {
		t.Fatalf("browser_open managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.openReqs) != 0 {
		t.Fatalf("expected host backend open to stay unused, got %#v", hostBackend.openReqs)
	}
	if len(nodeBackend.openReqs) != 1 || nodeBackend.openReqs[0].URL != publicExampleIPURL {
		t.Fatalf("expected managed current route to drive node open before implicit host fallback, got %#v", nodeBackend.openReqs)
	}
	if nodeBackend.openReqs[0].BrowserApp != "Chromium" {
		t.Fatalf("expected managed current open to inherit managed browser app before host fallback, got %#v", nodeBackend.openReqs)
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed current open output: %v", err)
	}
	if payload.Backend != "proxy-open" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "opened" {
		t.Fatalf("unexpected managed current open payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_OpenRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser_open"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_open",
		Arguments: `{"url":"https://93.184.216.34"}`,
	})
	if err != nil {
		t.Fatalf("browser_open managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.openReqs) != 1 || nodeBackend.openReqs[0].URL != "https://93.184.216.34" {
		t.Fatalf("expected managed default route to drive node open before implicit host fallback, got %#v", nodeBackend.openReqs)
	}
	var payload struct {
		Backend       string `json:"backend"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default open output: %v", err)
	}
	if payload.Backend != "proxy-open" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "opened" {
		t.Fatalf("unexpected managed default open payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_NavigateRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
				navigateResult: BrowserNavigateResult{Backend: "proxy-navigate", BrowserApp: "Chromium", FinalURL: publicExampleIPURL, Title: "Workbench", Status: "navigated"},
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
		EnabledTools:    []string{"browser_navigate"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-navigate-managed-current-hidden-implicit-host")
	sessionRegistry.TrackCurrentTarget("browser-navigate-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Name:      "browser_navigate",
		Arguments: `{"url":"` + publicExampleIPURL + `"}`,
	})
	if err != nil {
		t.Fatalf("browser_navigate managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.navigateReqs) != 0 {
		t.Fatalf("expected host backend navigate to stay unused, got %#v", hostBackend.navigateReqs)
	}
	if len(nodeBackend.navigateReqs) != 1 || nodeBackend.navigateReqs[0].URL != publicExampleIPURL {
		t.Fatalf("expected managed current route to drive node navigate before implicit host fallback, got %#v", nodeBackend.navigateReqs)
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
		FinalURL      string `json:"final_url"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed current navigate output: %v", err)
	}
	if payload.Backend != "proxy-navigate" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "navigated" || payload.FinalURL != publicExampleIPURL {
		t.Fatalf("unexpected managed current navigate payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_NavigateRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				navigateResult: BrowserNavigateResult{Backend: "proxy-navigate", FinalURL: publicExampleIPURL, Status: "navigated"},
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
		EnabledTools: []string{"browser_navigate"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: `{"url":"` + publicExampleIPURL + `"}`,
	})
	if err != nil {
		t.Fatalf("browser_navigate managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.navigateReqs) != 1 || nodeBackend.navigateReqs[0].URL != publicExampleIPURL || nodeBackend.navigateReqs[0].TabIndex != 0 {
		t.Fatalf("expected managed default route to drive node navigate before implicit host fallback, got %#v", nodeBackend.navigateReqs)
	}
	var payload struct {
		Backend       string `json:"backend"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
		FinalURL      string `json:"final_url"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default navigate output: %v", err)
	}
	if payload.Backend != "proxy-navigate" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "navigated" || payload.FinalURL != publicExampleIPURL {
		t.Fatalf("unexpected managed default navigate payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ExtractCurrentRequiresExplicitHostTargetForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionRegistry.TrackCurrentTarget("browser-extract-current-implicit-host", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://93.184.216.34/current",
		Title:      "Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser_extract"},
	})

	_, err := reg.Execute(WithToolSessionID(context.Background(), "browser-extract-current-implicit-host"), types.FunctionCall{
		Name:      "browser_extract",
		Arguments: `{"target":"current","max_chars":32}`,
	})
	if err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target`) {
		t.Fatalf("expected implicit host current target to require explicit host runtime_target, got %v", err)
	}
}

func TestRegisterBrowserTools_ExtractCurrentRoutesToManagedTargetBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools:    []string{"browser_extract"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-extract-current-managed-hidden-implicit-host")
	tracked := sessionRegistry.TrackCurrentTarget("browser-extract-current-managed-hidden-implicit-host", BrowserSessionTarget{
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
		Name:      "browser_extract",
		Arguments: `{"target":"current","max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser_extract current managed hidden implicit host: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 2 || nodeBackend.extractReqs[0].WaitMs != 4000 {
		t.Fatalf("expected managed current target to route through node backend before implicit host fallback, got %#v", nodeBackend.extractReqs)
	}
	if nodeBackend.extractReqs[0].BrowserApp != "Chromium" || nodeBackend.extractReqs[0].URL != publicExampleIPURL {
		t.Fatalf("expected managed current extract to inherit Chromium route and tracked url before backend dispatch, got %#v", nodeBackend.extractReqs[0])
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Target        string `json:"target"`
		TargetID      string `json:"target_id"`
		FinalURL      string `json:"final_url"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed current extract output: %v", err)
	}
	if payload.Backend != "proxy-extract" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.FinalURL != publicExampleIPURL {
		t.Fatalf("unexpected managed current extract payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ExtractRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools:    []string{"browser_extract"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-extract-managed-current-hidden-implicit-host")
	tracked := sessionRegistry.TrackCurrentTarget("browser-extract-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Name:      "browser_extract",
		Arguments: `{"max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser_extract managed current hidden implicit host: %v", err)
	}
	if len(hostBackend.extractReqs) != 0 {
		t.Fatalf("expected host backend extract to stay unused, got %#v", hostBackend.extractReqs)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 2 || nodeBackend.extractReqs[0].WaitMs != 4000 {
		t.Fatalf("expected targetless extract to reuse managed current route before implicit host fallback, got %#v", nodeBackend.extractReqs)
	}
	if nodeBackend.extractReqs[0].BrowserApp != "Chromium" || nodeBackend.extractReqs[0].URL != publicExampleIPURL {
		t.Fatalf("expected targetless extract to inherit managed browser app and tracked url before backend dispatch, got %#v", nodeBackend.extractReqs[0])
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Target        string `json:"target"`
		TargetID      string `json:"target_id"`
		FinalURL      string `json:"final_url"`
		TabIndex      int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed targetless extract output: %v", err)
	}
	if payload.Backend != "proxy-extract" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Target != "target:"+tracked.ID || payload.TargetID != tracked.ID || payload.FinalURL != publicExampleIPURL || payload.TabIndex != 2 {
		t.Fatalf("unexpected managed targetless extract payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ExtractRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser_extract"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_extract",
		Arguments: `{"max_chars":32}`,
	})
	if err != nil {
		t.Fatalf("browser_extract managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 0 || nodeBackend.extractReqs[0].URL != "" || nodeBackend.extractReqs[0].WaitMs != browserTabTargetWaitMs || nodeBackend.extractReqs[0].MaxChars != 32 {
		t.Fatalf("expected managed default route to drive extract before implicit host fallback, got %#v", nodeBackend.extractReqs)
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
		FinalURL      string `json:"final_url"`
		TabIndex      int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed default extract output: %v", err)
	}
	if payload.Backend != "proxy-extract" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "extracted" || payload.FinalURL != "https://node.example/default" || payload.TabIndex != 0 {
		t.Fatalf("unexpected managed default extract payload: %#v", payload)
	}
}
