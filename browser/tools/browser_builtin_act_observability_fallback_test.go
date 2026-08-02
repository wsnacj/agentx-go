package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ActConsoleStaysHostUnsupportedWithoutManagedCurrentRoute(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
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
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"console"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `unsupported kind "console" for runtime_target="host" backend="system"`) {
		t.Fatalf("expected implicit host browser_act console without managed current route to stay on host unsupported path, got %v", err)
	}
}

func TestRegisterBrowserTools_ActRequestsStaysHostUnsupportedWithoutManagedCurrentRoute(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoCapabilityRouteResolverBrowserBackend{
		runtimeInfoCapabilityBrowserBackend: &runtimeInfoCapabilityBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{},
			runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
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
			return requested, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"requests"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `unsupported kind "requests" for runtime_target="host" backend="system"`) {
		t.Fatalf("expected implicit host browser_act requests without managed current route to stay on host unsupported path, got %v", err)
	}
}

func TestRegisterBrowserTools_ActConsoleRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"console","level":"error"}`,
	})
	if err != nil {
		t.Fatalf("browser_act console managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.consoleReqs) != 1 || nodeBackend.consoleReqs[0].Level != "error" || nodeBackend.consoleReqs[0].TabIndex != 0 {
		t.Fatalf("expected managed default browser_act console to drive node backend before implicit host fallback, got %#v", nodeBackend.consoleReqs)
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
		t.Fatalf("decode managed default browser_act console output: %v", err)
	}
	if payload.Kind != "console" || payload.Backend != "proxy-console" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.FinalURL != "https://node.example/workbench" || payload.Status != "ok" || len(payload.Messages) != 1 || payload.Messages[0].Level != "error" {
		t.Fatalf("unexpected managed default browser_act console payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "console_collected" {
		t.Fatalf("unexpected managed default browser_act console summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "console_collected" {
		t.Fatalf("unexpected managed default browser_act console display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "console_collected" {
		t.Fatalf("unexpected managed default browser_act console surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "console_collected" {
		t.Fatalf("unexpected managed default browser_act console view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_ActRequestsRoutesManagedDefaultRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
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
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"requests","filter":"api"}`,
	})
	if err != nil {
		t.Fatalf("browser_act requests managed default hidden implicit host: %v", err)
	}
	if len(nodeBackend.requestsReqs) != 1 || nodeBackend.requestsReqs[0].Filter != "api" || nodeBackend.requestsReqs[0].TabIndex != 0 {
		t.Fatalf("expected managed default browser_act requests to drive node backend before implicit host fallback, got %#v", nodeBackend.requestsReqs)
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
		t.Fatalf("decode managed default browser_act requests output: %v", err)
	}
	if payload.Kind != "requests" || payload.Backend != "proxy-requests" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.FinalURL != "https://node.example/workbench" || payload.Status != "ok" || len(payload.Requests) != 1 || payload.Requests[0].URL != "https://node.example/api/items" {
		t.Fatalf("unexpected managed default browser_act requests payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected managed default browser_act requests summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected managed default browser_act requests display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected managed default browser_act requests surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected managed default browser_act requests view: %#v", payload.View)
	}
}
