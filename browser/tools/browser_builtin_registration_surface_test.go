package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"runtime"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_NodeManagedDefaultExpandsCompatSurface(t *testing.T) {
	reg := llmxtools.NewRegistry()
	node := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Navigate:   true,
				Tabs:       true,
				Click:      true,
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
		Root:        t.TempDir(),
		NodeBackend: node,
	})

	got := browserDefinitionNames(reg)
	for _, name := range []string{"browser_navigate", "browser_tabs", "browser_click", "browser_screenshot"} {
		if !browserStringSliceContains(got, name) {
			t.Fatalf("expected managed default route to expose %s in compat surface, got %#v", name, got)
		}
	}
}

func TestRegisterBrowserTools_DefaultSystemBackendHonorsCapabilities(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{Root: t.TempDir()})

	got := browserDefinitionNames(reg)
	want := []string{"browser", "browser_act", "browser_extract", "browser_open", "browser_runtime"}
	if runtime.GOOS == "darwin" {
		want = []string{"browser", "browser_act", "browser_click", "browser_eval", "browser_extract", "browser_navigate", "browser_open", "browser_runtime", "browser_screenshot", "browser_tabs", "browser_type"}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected default browser tool surface: got=%#v want=%#v", got, want)
	}
	if kinds := browserActDefinitionKinds(reg); !reflect.DeepEqual(kinds, browserActKindsForRegistration(BrowserToolOptions{})) {
		t.Fatalf("unexpected browser_act kinds in definition: got=%#v want=%#v", kinds, browserActKindsForRegistration(BrowserToolOptions{}))
	}
}

func TestRegisterBrowserTools_DefaultDarwinSystemBackendExposesInteractiveSurface(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-only browser surface guardrail")
	}
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{Root: t.TempDir()})

	got := browserDefinitionNames(reg)
	for _, name := range []string{"browser_navigate", "browser_tabs", "browser_click", "browser_type", "browser_eval", "browser_screenshot"} {
		if !browserStringSliceContains(got, name) {
			t.Fatalf("expected darwin default browser surface to include %s, got %#v", name, got)
		}
	}
	kinds := browserActDefinitionKinds(reg)
	for _, kind := range []string{"navigate", "click", "type", "evaluate", "list_tabs", "focus_tab", "close_tab", "screenshot"} {
		if !browserStringSliceContains(kinds, kind) {
			t.Fatalf("expected darwin browser_act surface to include %s, got %#v", kind, kinds)
		}
	}
}

func TestRegisterBrowserTools_CustomCapabilityBackendSkipsUnsupportedTools(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult:    BrowserOpenResult{Backend: "fake-open", Status: "opened"},
			extractResult: BrowserExtractResult{Backend: "fake-extract", Content: "ok"},
		},
		capabilities: BrowserCapabilities{
			Open:     true,
			Extract:  true,
			Snapshot: true,
			Wait:     true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:    t.TempDir(),
		Backend: backend,
	})

	got := browserDefinitionNames(reg)
	want := []string{"browser", "browser_act", "browser_extract", "browser_open", "browser_runtime"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected capability-constrained browser tool surface: got=%#v want=%#v", got, want)
	}
	wantKinds := []string{"open", "extract", "snapshot", "wait"}
	if kinds := browserActDefinitionKinds(reg); !reflect.DeepEqual(kinds, wantKinds) {
		t.Fatalf("unexpected capability-constrained browser_act kinds: got=%#v want=%#v", kinds, wantKinds)
	}
	if _, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","selector":"button.buy"}`,
	}); err == nil || !strings.Contains(err.Error(), `unsupported kind "click"`) {
		t.Fatalf("expected unsupported click kind error, got %v", err)
	}
}

func TestRegisterBrowserTools_ActUsesRouteScopedCapabilities(t *testing.T) {
	reg := llmxtools.NewRegistry()
	host := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"fill"}),
	}
	node := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"snapshot"}),
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      host,
		NodeBackend:  node,
		EnabledTools: []string{"browser_act", "browser_runtime"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"fill","runtime_target":"node"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `unsupported kind "fill" for runtime_target="node" backend="proxy"`) {
		t.Fatalf("expected route-scoped unsupported fill kind error, got %v", err)
	}
	if len(host.fillReqs) != 0 || len(node.fillReqs) != 0 {
		t.Fatalf("expected route-scoped capability gate to block fill before backend dispatch, host=%#v node=%#v", host.fillReqs, node.fillReqs)
	}
}

func TestRegisterBrowserTools_RuntimeUsesRouteScopedCapabilities(t *testing.T) {
	reg := llmxtools.NewRegistry()
	host := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"fill"}),
	}
	node := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"snapshot"}),
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      host,
		NodeBackend:  node,
		EnabledTools: []string{"browser_runtime", "browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"runtime_target":"node","include_routes":true}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime route-scoped capabilities: %v", err)
	}
	var payload struct {
		SelectedRoute struct {
			Backend       string `json:"backend"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		BrowserActKinds     []string        `json:"browser_act_kinds"`
		BrowserSurface      string          `json:"browser_surface"`
		BrowserOptInTargets []string        `json:"browser_opt_in_targets"`
		Capabilities        map[string]bool `json:"capabilities"`
		Routes              []struct {
			Backend             string          `json:"backend"`
			RuntimeTarget       string          `json:"runtime_target"`
			Status              string          `json:"status"`
			BrowserActKinds     []string        `json:"browser_act_kinds"`
			BrowserSurface      string          `json:"browser_surface"`
			BrowserOptInTargets []string        `json:"browser_opt_in_targets"`
			Capabilities        map[string]bool `json:"capabilities"`
		} `json:"routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime route-scoped capabilities output: %v", err)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected runtime payload to resolve explicit node route, got %#v", payload.SelectedRoute)
	}
	if !browserStringSliceContains(payload.BrowserActKinds, "snapshot") || browserStringSliceContains(payload.BrowserActKinds, "fill") {
		t.Fatalf("expected selected route browser_act kinds to reflect concrete node capabilities, got %#v", payload.BrowserActKinds)
	}
	if payload.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.BrowserOptInTargets) != 1 ||
		payload.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected selected route payload to expose explicit managed opt-in surface, got %#v", payload)
	}
	if payload.Capabilities["snapshot"] != true || payload.Capabilities["fill"] {
		t.Fatalf("expected selected route capabilities map to reflect concrete node route, got %#v", payload.Capabilities)
	}
	foundNodeRoute := false
	for _, route := range payload.Routes {
		if route.Backend != "proxy" || route.RuntimeTarget != "node" || (route.Status != "default" && route.Status != "available") {
			continue
		}
		foundNodeRoute = true
		if !browserStringSliceContains(route.BrowserActKinds, "snapshot") || browserStringSliceContains(route.BrowserActKinds, "fill") {
			t.Fatalf("expected node route matrix entry to reflect concrete node capabilities, got %#v", route)
		}
		if route.BrowserSurface != "explicit_managed_opt_in" ||
			len(route.BrowserOptInTargets) != 1 ||
			route.BrowserOptInTargets[0] != "node" {
			t.Fatalf("expected node route matrix entry to expose explicit managed opt-in surface, got %#v", route)
		}
		if route.Capabilities["snapshot"] != true || route.Capabilities["fill"] {
			t.Fatalf("expected node route capability map to stay route-scoped, got %#v", route.Capabilities)
		}
	}
	if !foundNodeRoute {
		t.Fatalf("expected runtime route matrix to include node route, got %#v", payload.Routes)
	}
}

func TestRegisterBrowserTools_NodeBackendCapabilitiesExpandVisibleSurface(t *testing.T) {
	reg := llmxtools.NewRegistry()
	host := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "fake-open", Status: "opened"},
		},
		capabilities: BrowserCapabilities{
			Open:     true,
			Extract:  true,
			Snapshot: true,
			Wait:     true,
		},
	}
	node := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			clickResult: BrowserClickResult{Backend: "fake-click", Status: "clicked"},
		},
		capabilities: BrowserCapabilities{
			Click: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:        t.TempDir(),
		Backend:     host,
		NodeBackend: node,
	})

	got := browserDefinitionNames(reg)
	if !browserStringSliceContains(got, "browser_click") || !browserStringSliceContains(got, "browser_act") {
		t.Fatalf("expected node backend capabilities to expand visible surface, got %#v", got)
	}
	if kinds := browserActDefinitionKinds(reg); !browserStringSliceContains(kinds, "click") {
		t.Fatalf("expected browser_act kinds to include click from node backend, got %#v", kinds)
	}
}

func TestRegisterBrowserTools_NodeBackendRouteResolutionFailureDoesNotExpandVisibleSurface(t *testing.T) {
	reg := llmxtools.NewRegistry()
	host := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "fake-open", Status: "opened"},
		},
		capabilities: BrowserCapabilities{
			Open:     true,
			Extract:  true,
			Snapshot: true,
			Wait:     true,
		},
	}
	node := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				clickResult: BrowserClickResult{Backend: "fake-click", Status: "clicked"},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
			capabilities: BrowserCapabilities{
				Click: true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:        t.TempDir(),
		Backend:     host,
		NodeBackend: node,
	})

	got := browserDefinitionNames(reg)
	if browserStringSliceContains(got, "browser_click") {
		t.Fatalf("expected unresolved node route to stay out of visible specialist surface, got %#v", got)
	}
	if kinds := browserActDefinitionKinds(reg); browserStringSliceContains(kinds, "click") {
		t.Fatalf("expected unresolved node route to stay out of browser_act kinds, got %#v", kinds)
	}
}

func TestRegisterBrowserTools_HostRouteResolutionFailureKeepsOnlyUnifiedRuntimeSurface(t *testing.T) {
	reg := llmxtools.NewRegistry()
	host := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				openResult: BrowserOpenResult{Backend: "custom-open", Status: "opened"},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities: BrowserCapabilities{
				Open:     true,
				Extract:  true,
				Snapshot: true,
				Wait:     true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:    t.TempDir(),
		Backend: host,
	})

	got := browserDefinitionNames(reg)
	want := []string{"browser", "browser_runtime"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected unresolved explicit host route to keep only unified/runtime surface, got=%#v want=%#v", got, want)
	}
	if kinds := browserActDefinitionKinds(reg); len(kinds) != 0 {
		t.Fatalf("expected unresolved explicit host route to suppress browser_act kinds, got %#v", kinds)
	}
}

func TestRegisterBrowserTools_HostRouteResolutionFailureKeepsExplicitNodeSurfaceAvailable(t *testing.T) {
	reg := llmxtools.NewRegistry()
	host := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				openResult: BrowserOpenResult{Backend: "custom-open", Status: "opened"},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			capabilities: BrowserCapabilities{
				Open:     true,
				Extract:  true,
				Snapshot: true,
				Wait:     true,
			},
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	node := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			clickResult: BrowserClickResult{Backend: "node-click", Status: "clicked"},
		},
		runtimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		capabilities: BrowserCapabilitiesForActKinds([]string{"click", "snapshot"}),
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:        t.TempDir(),
		Backend:     host,
		NodeBackend: node,
	})

	got := browserDefinitionNames(reg)
	want := []string{"browser", "browser_act", "browser_click", "browser_runtime"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected unresolved explicit host route to keep explicit node surface visible, got=%#v want=%#v", got, want)
	}
	if kinds := browserActDefinitionKinds(reg); !reflect.DeepEqual(kinds, []string{"snapshot", "click"}) {
		t.Fatalf("expected unresolved explicit host route to keep node browser_act kinds, got %#v", kinds)
	}
	if _, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","runtime_target":"node","selector":"button.buy"}`,
	}); err != nil {
		t.Fatalf("expected explicit node lane to stay executable after host route failure, got %v", err)
	}
	if len(node.clickReqs) != 1 {
		t.Fatalf("expected explicit node lane dispatch after host route failure, got %#v", node.clickReqs)
	}
}
