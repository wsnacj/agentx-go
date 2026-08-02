package tools

import (
	"context"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_TabsFocusRequiresExplicitHostForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_tabs"},
	})

	expectBrowserCompatToolExplicitFallbackOrNotRegistered(t, reg, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"focus","tab_index":2}`,
	}, "browser_tabs")
}

func TestRegisterBrowserTools_TabsCloseRequiresExplicitHostForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_tabs"},
	})

	expectBrowserCompatToolExplicitFallbackOrNotRegistered(t, reg, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"close","tab_index":2}`,
	}, "browser_tabs")
}

func TestRegisterBrowserTools_ActFocusTabRequiresExplicitHostForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_act"},
	})

	expectBrowserActKindExplicitFallbackOrUnsupported(t, reg, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"focus_tab","tab_index":2}`,
	}, "focus_tab")
}

func TestRegisterBrowserTools_ActCloseTabRequiresExplicitHostForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_act"},
	})

	expectBrowserActKindExplicitFallbackOrUnsupported(t, reg, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"close_tab","tab_index":2}`,
	}, "close_tab")
}

func TestRegisterBrowserTools_OpenRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_open"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_open",
		Arguments: `{"url":"https://93.184.216.34"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target`) {
		t.Fatalf("expected implicit host browser_open to require explicit runtime_target, got %v", err)
	}
}

func TestRegisterBrowserTools_NavigateRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_navigate"},
	})

	expectBrowserCompatToolExplicitFallbackOrNotRegistered(t, reg, types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: `{"url":"https://93.184.216.34"}`,
	}, "browser_navigate")
}

func TestRegisterBrowserTools_ActOpenRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"open","url":"https://93.184.216.34"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target`) {
		t.Fatalf("expected implicit host browser_act open to require explicit runtime_target, got %v", err)
	}
}

func TestRegisterBrowserTools_ActNavigateRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_act"},
	})

	expectBrowserActKindExplicitFallbackOrUnsupported(t, reg, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"navigate","url":"https://93.184.216.34"}`,
	}, "navigate")
}

func TestBrowserRuntimeActionPrepareRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	if !browserRuntimeActionRequiresExplicitRuntimeTargetForImplicitFallback("prepare") {
		t.Fatalf("expected runtime prepare to require explicit runtime_target when default browser route falls back to the legacy system host path")
	}
}

func TestBrowserRuntimeActionStartRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	if !browserRuntimeActionRequiresExplicitRuntimeTargetForImplicitFallback("start") {
		t.Fatalf("expected runtime start to require explicit runtime_target when default browser route falls back to the legacy system host path")
	}
}

func TestBrowserRuntimeActionStatusUsesImplicitHostDiagnosticsPath(t *testing.T) {
	if browserRuntimeActionRequiresExplicitRuntimeTargetForImplicitFallback("") {
		t.Fatalf("expected runtime default action to stay on the implicit-host diagnostics path")
	}
	if !browserRuntimeUsesImplicitLegacyHostDiagnosticsDegradePath("", true, "", "", false) {
		t.Fatalf("expected runtime default action to use the implicit-host diagnostics degrade path")
	}
	if browserRuntimeActionRequiresExplicitRuntimeTargetForImplicitFallback("status") {
		t.Fatalf("expected runtime status to stay on the implicit-host diagnostics path")
	}
	if !browserRuntimeUsesImplicitLegacyHostDiagnosticsDegradePath("status", true, "", "", false) {
		t.Fatalf("expected runtime status to use the implicit-host diagnostics degrade path")
	}
	if browserRuntimeUsesImplicitLegacyHostDiagnosticsDegradePath("status", true, "", "", true) {
		t.Fatalf("expected pure targetless runtime status to prefer managed current route over implicit-host diagnostics degrade when available")
	}
	if !browserRuntimeUsesImplicitLegacyHostDiagnosticsDegradePath("status", true, "default", "", true) {
		t.Fatalf("expected runtime status profile=default to reuse the implicit-host diagnostics degrade path")
	}
}

func TestRegisterBrowserTools_RuntimeDefaultActionTreatsImplicitHostAsDiagnosticsRequest(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_runtime"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{}`,
	})
	if err != nil {
		t.Fatalf("expected implicit host default browser_runtime action to stay on diagnostics path, got %v", err)
	}
	if strings.Contains(out, `"selected_route"`) {
		t.Fatalf("expected implicit host default browser_runtime payload to hide selected_route, got %s", out)
	}
}

func TestBrowserImplicitLegacyHostManagedActKindFallbackError(t *testing.T) {
	if err := browserImplicitLegacyHostManagedActKindFallbackError(true, false, BrowserRuntimeInfo{Target: "host"}, "console"); err == nil {
		t.Fatalf("expected managed-only act kind to require explicit runtime_target on implicit host fallback")
	}
	if err := browserImplicitLegacyHostManagedActKindFallbackError(true, false, BrowserRuntimeInfo{Target: "host"}, "click"); err != nil {
		t.Fatalf("expected legacy-host-supported act kind to stay allowed on implicit host fallback, got %v", err)
	}
	if err := browserImplicitLegacyHostManagedActKindFallbackError(true, true, BrowserRuntimeInfo{Target: "host"}, "console"); err != nil {
		t.Fatalf("expected explicit runtime_target to bypass implicit-host managed act fallback, got %v", err)
	}
	if err := browserImplicitLegacyHostManagedActKindFallbackError(true, false, BrowserRuntimeInfo{Target: "node"}, "console"); err != nil {
		t.Fatalf("expected managed runtime route to bypass implicit-host managed act fallback, got %v", err)
	}
}

func TestBrowserImplicitLegacyHostSupportsActKind(t *testing.T) {
	for _, kind := range []string{"open", "navigate", "extract", "snapshot", "screenshot", "click", "type", "evaluate", "wait", "list_tabs", "focus_tab", "close_tab"} {
		if !browserImplicitLegacyHostSupportsActKind(kind) {
			t.Fatalf("expected legacy host lane to allow kind %q", kind)
		}
	}
	for _, kind := range []string{"console", "requests", "cookies", "storage", "highlight", "upload", "fill"} {
		if browserImplicitLegacyHostSupportsActKind(kind) {
			t.Fatalf("expected legacy host lane to reject managed-only kind %q", kind)
		}
	}
}

func TestBrowserImplicitLegacyHostPageExecutionFallbackError(t *testing.T) {
	if err := browserImplicitLegacyHostPageExecutionFallbackError("browser_extract", true, false, BrowserRuntimeInfo{Target: "host"}, browserToolTarget{}, "https://93.184.216.34"); err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target because`) {
		t.Fatalf("expected implicit host URL route to require explicit runtime_target, got %v", err)
	}
	if err := browserImplicitLegacyHostPageExecutionFallbackError("browser_extract", true, false, BrowserRuntimeInfo{Target: "host"}, browserToolTarget{}, ""); err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target=host or an explicit target/url`) {
		t.Fatalf("expected implicit host current-page fallback to require explicit host targeting, got %v", err)
	}
	if err := browserImplicitLegacyHostPageExecutionFallbackError("browser_extract", true, true, BrowserRuntimeInfo{Target: "host"}, browserToolTarget{}, "https://93.184.216.34"); err != nil {
		t.Fatalf("expected explicit runtime_target to bypass implicit-host page execution fallback, got %v", err)
	}
}

func TestBrowserImplicitLegacyHostDirectPageActionFallbackError(t *testing.T) {
	if err := browserImplicitLegacyHostDirectPageActionFallbackError("browser backend extract", true, browserToolTarget{TabIndex: 4}, ""); err == nil || !strings.Contains(err.Error(), `target "tab:4" requires explicit runtime_target=host`) {
		t.Fatalf("expected implicit host tab-target page action to require explicit host route, got %v", err)
	}
	if err := browserImplicitLegacyHostDirectPageActionFallbackError("browser backend extract", true, browserToolTarget{}, "https://93.184.216.34"); err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target because`) {
		t.Fatalf("expected implicit host URL page action to require explicit runtime_target, got %v", err)
	}
	if err := browserImplicitLegacyHostDirectPageActionFallbackError("browser backend extract", true, browserToolTarget{}, ""); err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target=host or an explicit target/url`) {
		t.Fatalf("expected implicit host current-page action to require explicit host targeting, got %v", err)
	}
	if err := browserImplicitLegacyHostDirectPageActionFallbackError("browser backend extract", false, browserToolTarget{}, ""); err != nil {
		t.Fatalf("expected non-implicit-host direct page action to bypass fallback error, got %v", err)
	}
}

func TestBrowserImplicitLegacyHostURLNavigationFallbackError(t *testing.T) {
	if err := browserImplicitLegacyHostURLNavigationFallbackError("browser_open", true, false, BrowserRuntimeInfo{Target: "host"}, browserToolTarget{}, "https://93.184.216.34"); err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target because`) {
		t.Fatalf("expected implicit host open fallback to require explicit runtime_target, got %v", err)
	}
	if err := browserImplicitLegacyHostURLNavigationFallbackError("browser_navigate", true, false, BrowserRuntimeInfo{Target: "host"}, browserToolTarget{TabIndex: 3}, "https://93.184.216.34"); err == nil || !strings.Contains(err.Error(), `target "tab:3" requires explicit runtime_target=host`) {
		t.Fatalf("expected implicit host navigate tab target to require explicit host route, got %v", err)
	}
	if err := browserImplicitLegacyHostURLNavigationFallbackError("browser_navigate", true, true, BrowserRuntimeInfo{Target: "host"}, browserToolTarget{}, "https://93.184.216.34"); err != nil {
		t.Fatalf("expected explicit runtime_target to bypass implicit-host URL navigation fallback, got %v", err)
	}
	if err := browserImplicitLegacyHostURLNavigationFallbackError("browser_navigate", true, false, BrowserRuntimeInfo{Target: "node"}, browserToolTarget{}, "https://93.184.216.34"); err != nil {
		t.Fatalf("expected managed runtime_target to bypass implicit-host URL navigation fallback, got %v", err)
	}
}

func TestBrowserImplicitLegacyHostTabsActionFallbackError(t *testing.T) {
	if err := browserImplicitLegacyHostTabsActionFallbackError("browser_tabs", true, BrowserRuntimeInfo{Target: "host"}, "list", browserToolTarget{}); err == nil || !strings.Contains(err.Error(), `requires explicit runtime_target=host`) {
		t.Fatalf("expected implicit host list-tabs fallback to require explicit host route, got %v", err)
	}
	if err := browserImplicitLegacyHostTabsActionFallbackError("browser_tabs", true, BrowserRuntimeInfo{Target: "host"}, "focus", browserToolTarget{TabIndex: 2}); err == nil || !strings.Contains(err.Error(), `target "tab:2" requires explicit runtime_target=host`) {
		t.Fatalf("expected implicit host focus-tab fallback to require explicit host route, got %v", err)
	}
	if err := browserImplicitLegacyHostTabsActionFallbackError("browser_tabs", true, BrowserRuntimeInfo{Target: "host"}, "focus", browserToolTarget{}); err == nil || !strings.Contains(err.Error(), `target "current" requires explicit runtime_target=host`) {
		t.Fatalf("expected implicit host non-list tabs fallback to use target-scoped explicit host error, got %v", err)
	}
	if err := browserImplicitLegacyHostTabsActionFallbackError("browser_tabs", false, BrowserRuntimeInfo{Target: "host"}, "list", browserToolTarget{}); err != nil {
		t.Fatalf("expected non-implicit-host tabs action to bypass fallback error, got %v", err)
	}
	if err := browserImplicitLegacyHostTabsActionFallbackError("browser_tabs", true, BrowserRuntimeInfo{Target: "node"}, "focus", browserToolTarget{TabIndex: 2}); err != nil {
		t.Fatalf("expected managed tabs route to bypass implicit-host fallback error, got %v", err)
	}
}

func TestBrowserImplicitLegacyHostRuntimeDiagnosticsRequestedProfile(t *testing.T) {
	if normalized, ok := browserImplicitLegacyHostRuntimeDiagnosticsRequestedProfile("status", "", "default"); !ok || normalized != "" {
		t.Fatalf("expected empty status profile to stay on diagnostics path, got %q %v", normalized, ok)
	}
	if normalized, ok := browserImplicitLegacyHostRuntimeDiagnosticsRequestedProfile("status", "default", "default"); !ok || normalized != "" {
		t.Fatalf("expected default status profile to normalize onto diagnostics path, got %q %v", normalized, ok)
	}
	if normalized, ok := browserImplicitLegacyHostRuntimeDiagnosticsRequestedProfile("status", "relay", "default"); ok || normalized != "relay" {
		t.Fatalf("expected alternate status profile to stay off implicit-host diagnostics path, got %q %v", normalized, ok)
	}
	if normalized, ok := browserImplicitLegacyHostRuntimeCanUseCachedDiagnosticsSnapshot("profiles", "default", "default"); ok || normalized != "" {
		t.Fatalf("expected default profiles diagnostics to skip cached snapshot degrade, got %q %v", normalized, ok)
	}
	if normalized, ok := browserImplicitLegacyHostRuntimeCanUseCachedDiagnosticsSnapshot("sessions", "default", "default"); !ok || normalized != "" {
		t.Fatalf("expected default sessions diagnostics to allow cached snapshot degrade, got %q %v", normalized, ok)
	}
}

func TestRegisterBrowserTools_RuntimeSelectProfileRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_runtime"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","profile":"default"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `browser_runtime action select_profile requires explicit runtime_target because the default browser route falls back to the legacy system host path`) {
		t.Fatalf("expected implicit host browser_runtime select_profile to require explicit runtime_target, got %v", err)
	}
}

func TestRegisterBrowserTools_RuntimeClearSessionRequiresExplicitRuntimeTargetForImplicitFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		EnabledTools: []string{"browser_runtime"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_session"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `browser_runtime action clear_session requires explicit runtime_target because the default browser route falls back to the legacy system host path`) {
		t.Fatalf("expected implicit host browser_runtime clear_session to require explicit runtime_target, got %v", err)
	}
}

func TestRegisterBrowserTools_RuntimeClearSessionRequiresExplicitRuntimeTargetForUnresolvableManagedCurrentRoute(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-session-stale-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-clear-session-stale-managed-current-hidden-implicit-host", BrowserSessionTarget{
		ID:         "stale-node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	_, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_session"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `browser_runtime action clear_session requires explicit runtime_target because the default browser route falls back to the legacy system host path`) {
		t.Fatalf("expected stale managed current route to stay behind explicit runtime_target gate, got %v", err)
	}
}

func TestRegisterBrowserTools_RuntimeClearSessionProbesUnresolvableManagedCurrentRouteOnce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-clear-session-stale-managed-current-single-probe")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
				runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
					fakeBrowserBackend: &fakeBrowserBackend{},
					runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
				},
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	sessionRegistry.TrackCurrentTarget("browser-runtime-clear-session-stale-managed-current-single-probe", BrowserSessionTarget{
		ID:         "stale-node-current",
		TabIndex:   2,
		URL:        "https://node.example/workbench",
		Title:      "Workbench",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
	}, "tracked_active_tab")

	resolveCallsBefore := nodeBackend.resolveCalls
	_, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"clear_session"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `browser_runtime action clear_session requires explicit runtime_target because the default browser route falls back to the legacy system host path`) {
		t.Fatalf("expected stale managed current route to stay behind explicit runtime_target gate, got %v", err)
	}
	if got := nodeBackend.resolveCalls - resolveCallsBefore; got > 1 {
		t.Fatalf("expected stale managed current route probe to resolve at most once, got %d", got)
	}
}
