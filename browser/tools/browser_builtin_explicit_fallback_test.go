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

func TestRegisterBrowserTools_ClickPageBoundRefRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
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
		MaxChars:        64,
		EnabledTools:    []string{"browser_click"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-click-ref-managed-hidden-implicit-host")
	sessionRegistry.TrackCurrentTarget("browser-click-ref-managed-hidden-implicit-host", BrowserSessionTarget{
		ID:         "node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")
	ref := browserElementRefForSnapshotElement(BrowserSnapshotElement{
		Selector: `button[data-action="run"]`,
		Label:    "Run",
		Role:     "button",
	}, "https://node.example/workbench", "Workbench")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_click",
		Arguments: `{"ref":"` + ref + `"}`,
	})
	if err != nil {
		t.Fatalf("browser_click page-bound ref managed hidden implicit host: %v", err)
	}
	if len(hostBackend.clickReqs) != 0 {
		t.Fatalf("expected host backend click to stay unused, got %#v", hostBackend.clickReqs)
	}
	if len(nodeBackend.clickReqs) != 1 {
		t.Fatalf("expected managed page-bound ref click to route through node backend before implicit host fallback, got %#v", nodeBackend.clickReqs)
	}
	if nodeBackend.clickReqs[0].Selector != `button[data-action="run"]` || nodeBackend.clickReqs[0].ElementRef != ref {
		t.Fatalf("unexpected managed page-bound ref click request: %#v", nodeBackend.clickReqs[0])
	}
	if nodeBackend.clickReqs[0].BrowserApp != "Chromium" || nodeBackend.clickReqs[0].TabIndex != 2 {
		t.Fatalf("expected managed page-bound ref click to inherit managed current target before host fallback, got %#v", nodeBackend.clickReqs[0])
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Ref           string `json:"ref"`
		Target        string `json:"target"`
		TargetID      string `json:"target_id"`
		TabIndex      int    `json:"tab_index"`
		FinalURL      string `json:"final_url"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode managed page-bound ref click output: %v", err)
	}
	if payload.Backend != "proxy-click" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Ref != ref || !strings.HasPrefix(payload.Target, "target:") || strings.TrimSpace(payload.TargetID) == "" || payload.TabIndex != 2 || payload.FinalURL != "https://node.example/workbench" {
		t.Fatalf("unexpected managed page-bound ref click payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ExplicitManagedCompatClickRegistersOutsideStaticVisibleSurface(t *testing.T) {
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
		EnabledTools: []string{"browser_click"},
	})

	got := browserDefinitionNames(reg)
	if !browserStringSliceContains(got, "browser_click") {
		t.Fatalf("expected explicit managed compat click to stay registered, got %#v", got)
	}
	if kinds := browserActDefinitionKinds(reg); browserStringSliceContains(kinds, "click") {
		t.Fatalf("expected static browser_act kinds to remain conservative, got %#v", kinds)
	}

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_click",
		Arguments: `{"runtime_target":"node","profile":"workbench","selector":"button.buy"}`,
	})
	if err != nil {
		t.Fatalf("browser_click explicit managed compat registration: %v", err)
	}
	if len(hostBackend.clickReqs) != 0 {
		t.Fatalf("expected explicit managed compat click to avoid host backend, got %#v", hostBackend.clickReqs)
	}
	if len(nodeBackend.clickReqs) != 1 ||
		nodeBackend.clickReqs[0].Selector != "button.buy" ||
		nodeBackend.clickReqs[0].TabIndex != 0 ||
		nodeBackend.clickReqs[0].PostWaitMs != 750 {
		t.Fatalf("unexpected explicit managed compat click dispatch: %#v", nodeBackend.clickReqs)
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode explicit managed compat click output: %v", err)
	}
	if payload.Backend != "proxy-click" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "clicked" {
		t.Fatalf("unexpected explicit managed compat click payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ExplicitCompatClickDoesNotReviveBrokenHostOnlySurface(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities:       BrowserCapabilitiesForActKinds([]string{"click"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		EnabledTools: []string{"browser_click"},
	})

	if got := browserDefinitionNames(reg); browserStringSliceContains(got, "browser_click") {
		t.Fatalf("expected broken host-only surface not to revive browser_click via explicit enable, got %#v", got)
	}
}

func TestRegisterBrowserTools_ExtractRequiresExplicitHostCurrentPageForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		MaxChars:     64,
		EnabledTools: []string{"browser_extract"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_extract",
		Arguments: `{"max_chars":32}`,
	})
	if err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target`) {
		t.Fatalf("expected targetless implicit host extract to require explicit host runtime_target or target/url, got %v", err)
	}
}

func TestRegisterBrowserTools_ExtractURLRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		MaxChars:     64,
		EnabledTools: []string{"browser_extract"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_extract",
		Arguments: `{"url":"https://93.184.216.34","max_chars":32}`,
	})
	if err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target`) {
		t.Fatalf("expected explicit-url implicit host extract to require explicit runtime_target, got %v", err)
	}
}

func TestRegisterBrowserTools_ActExtractRequiresExplicitHostCurrentPageForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		MaxChars:     64,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target`) {
		t.Fatalf("expected targetless implicit host browser_act extract to require explicit host runtime_target or target/url, got %v", err)
	}
}

func TestRegisterBrowserTools_ClickURLRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_click"},
	})

	expectBrowserCompatToolExplicitFallbackOrNotRegistered(t, reg, types.FunctionCall{
		Name:      "browser_click",
		Arguments: `{"url":"https://93.184.216.34","selector":"#submit"}`,
	}, "browser_click")
}

func TestRegisterBrowserTools_ActClickURLRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_act"},
	})

	expectBrowserActKindExplicitFallbackOrUnsupported(t, reg, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","url":"https://93.184.216.34","selector":"#submit"}`,
	}, "click")
}

func TestRegisterBrowserTools_TabsListRequiresExplicitHostForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_tabs"},
	})

	expectBrowserCompatToolExplicitFallbackOrNotRegistered(t, reg, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"list"}`,
	}, "browser_tabs")
}

func TestRegisterBrowserTools_ActListTabsRequiresExplicitHostForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_act"},
	})

	expectBrowserActKindExplicitFallbackOrUnsupported(t, reg, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"list_tabs"}`,
	}, "list_tabs")
}
