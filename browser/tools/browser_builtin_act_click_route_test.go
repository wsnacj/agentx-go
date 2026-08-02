package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ActDispatchesClick(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		clickResult: BrowserClickResult{Backend: "fake-click", BrowserApp: "Safari", FinalURL: "https://93.184.216.34/after", Status: "clicked"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		NodeBackend:  backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","target":"tab:3","selector":"button.buy","post_wait_ms":500}`,
	})
	if err != nil {
		t.Fatalf("browser_act click: %v", err)
	}
	if len(backend.clickReqs) != 1 || backend.clickReqs[0].Selector != "button.buy" || backend.clickReqs[0].PostWaitMs != 500 || backend.clickReqs[0].TabIndex != 3 || backend.clickReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs {
		t.Fatalf("unexpected browser_act click dispatch: %#v", backend.clickReqs)
	}
	var payload struct {
		Kind     string                         `json:"kind"`
		Backend  string                         `json:"backend"`
		Status   string                         `json:"status"`
		Target   string                         `json:"target"`
		TabIndex int                            `json:"tab_index"`
		Summary  *browserTopLevelSummary        `json:"summary"`
		Display  *browserTopLevelDisplaySummary `json:"display"`
		Surface  *browserTopLevelSurfaceSummary `json:"surface"`
		View     *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "click" || payload.Backend != "fake-click" || payload.Status != "clicked" || payload.Target != "tab:3" || payload.TabIndex != 3 {
		t.Fatalf("unexpected browser_act click output: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "interaction" || payload.Summary.SummaryCode != "click_completed" {
		t.Fatalf("unexpected browser_act click summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "interaction" || payload.Display.SummaryCode != "click_completed" {
		t.Fatalf("unexpected browser_act click display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "interaction" || payload.Surface.SummaryCode != "click_completed" {
		t.Fatalf("unexpected browser_act click surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "interaction" || payload.View.SummaryCode != "click_completed" {
		t.Fatalf("unexpected browser_act click view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_ActClickAllowsExplicitManagedLaneOutsideStaticActKinds(t *testing.T) {
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
		EnabledTools: []string{"browser_act"},
	})

	rawKinds := browserActDefinitionKinds(reg)
	if len(rawKinds) == 0 {
		t.Fatalf("expected browser_act registration")
	}
	if browserStringSliceContains(rawKinds, "click") {
		t.Fatalf("expected static browser_act kind list to remain conservative, got %#v", rawKinds)
	}

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","runtime_target":"node","profile":"workbench","selector":"button.buy"}`,
	})
	if err != nil {
		t.Fatalf("browser_act explicit managed click outside static kinds: %v", err)
	}
	if len(hostBackend.clickReqs) != 0 {
		t.Fatalf("expected explicit managed click to avoid host backend, got %#v", hostBackend.clickReqs)
	}
	if len(nodeBackend.clickReqs) != 1 ||
		nodeBackend.clickReqs[0].Selector != "button.buy" ||
		nodeBackend.clickReqs[0].TabIndex != 0 ||
		nodeBackend.clickReqs[0].PostWaitMs != 750 {
		t.Fatalf("unexpected explicit managed browser_act click dispatch: %#v", nodeBackend.clickReqs)
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
		t.Fatalf("decode explicit managed browser_act click output: %v", err)
	}
	if payload.Kind != "click" || payload.Backend != "proxy-click" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "clicked" {
		t.Fatalf("unexpected explicit managed browser_act click payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActClickRegistersExplicitManagedOptInSurfaceWithoutVisibleActSurface(t *testing.T) {
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
		EnabledTools: []string{"browser_act"},
	})

	if got := browserDefinitionNames(reg); !browserStringSliceContains(got, "browser_act") {
		t.Fatalf("expected browser_act explicit managed opt-in registration, got %#v", got)
	}
	if kinds := browserActDefinitionKinds(reg); !browserStringSliceContains(kinds, "click") {
		t.Fatalf("expected explicit managed opt-in browser_act kinds to include click, got %#v", kinds)
	}

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","runtime_target":"node","profile":"workbench","selector":"button.buy"}`,
	})
	if err != nil {
		t.Fatalf("browser_act explicit managed opt-in click: %v", err)
	}
	if len(hostBackend.clickReqs) != 0 {
		t.Fatalf("expected explicit managed opt-in click to avoid host backend, got %#v", hostBackend.clickReqs)
	}
	if len(nodeBackend.clickReqs) != 1 ||
		nodeBackend.clickReqs[0].Selector != "button.buy" ||
		nodeBackend.clickReqs[0].TabIndex != 0 ||
		nodeBackend.clickReqs[0].PostWaitMs != 750 {
		t.Fatalf("unexpected explicit managed opt-in browser_act click dispatch: %#v", nodeBackend.clickReqs)
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
		t.Fatalf("decode explicit managed opt-in browser_act click output: %v", err)
	}
	if payload.Kind != "click" || payload.Backend != "proxy-click" || payload.BrowserApp != "Chromium" || payload.Profile != "workbench" || payload.RuntimeTarget != "node" || payload.Status != "clicked" {
		t.Fatalf("unexpected explicit managed opt-in browser_act click payload: %#v", payload)
	}
	if payload.BrowserSurface != "explicit_managed_opt_in" ||
		len(payload.BrowserOptInTargets) != 1 ||
		payload.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected explicit managed opt-in browser_act click payload to expose route surface, got %#v", payload)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_act") ||
		browserStringSliceContains(payload.ArtifactTools, "browser_screenshot") ||
		browserStringSliceContains(payload.ArtifactKinds, "screenshot") ||
		payload.ArtifactContract != "" ||
		!browserStringSliceContains(payload.BrowserActKinds, "click") {
		t.Fatalf("expected explicit managed opt-in browser_act click payload to expose selected-route capabilities, got %#v", payload)
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
		t.Fatalf("expected explicit managed opt-in browser_act click payload to expose success action hints, got %#v", payload)
	}
}

func TestRegisterBrowserTools_ActClickPayloadUsesConcreteSelectedRouteCapabilities(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoCapabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
		capabilities:       BrowserCapabilitiesForActKinds([]string{"open"}),
	}
	var nodeBackend *countingExecutionRouteResolverBrowserBackend
	nodeBackend = &countingExecutionRouteResolverBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				clickResult: BrowserClickResult{
					Backend:    "proxy-click",
					BrowserApp: "Chromium",
					FinalURL:   "https://node.example/workbench",
					Status:     "clicked",
				},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		},
		resolveExecution: func(requested BrowserRuntimeInfo) (browserResolvedExecutionRoute, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) != "node" {
				return browserResolvedExecutionRoute{}, context.DeadlineExceeded
			}
			if !strings.EqualFold(strings.TrimSpace(requested.Profile), "workbench") {
				return browserResolvedExecutionRoute{}, context.DeadlineExceeded
			}
			return browserResolvedExecutionRoute{
				Backend:      nodeBackend,
				RuntimeInfo:  BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
				Capabilities: BrowserCapabilitiesForActKinds([]string{"click"}),
			}, nil
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_act"},
	})
	resolveCallsAfterRegister := nodeBackend.executionCalls

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","runtime_target":"node","profile":"workbench","selector":"button.buy"}`,
	})
	if err != nil {
		t.Fatalf("browser_act concrete selected-route click: %v", err)
	}
	if len(hostBackend.clickReqs) != 0 {
		t.Fatalf("expected browser_act concrete selected-route click to avoid host backend, got %#v", hostBackend.clickReqs)
	}
	if nodeBackend.executionCalls != resolveCallsAfterRegister+1 {
		t.Fatalf("expected browser_act concrete selected-route click to resolve node lane once during execution, registration=%d total=%d", resolveCallsAfterRegister, nodeBackend.executionCalls)
	}
	if nodeBackend.runtimeCalls != 0 || nodeBackend.backendCalls != 0 {
		t.Fatalf("expected browser_act concrete selected-route click to avoid runtime/backend resolver fallbacks, runtime=%d backend=%d", nodeBackend.runtimeCalls, nodeBackend.backendCalls)
	}
	if len(nodeBackend.clickReqs) != 1 {
		t.Fatalf("expected browser_act concrete selected-route click to dispatch exactly once on node backend, got %#v", nodeBackend.clickReqs)
	}
	var payload struct {
		BrowserTools     []string `json:"browser_tools"`
		ArtifactTools    []string `json:"artifact_tools"`
		ArtifactKinds    []string `json:"artifact_kinds"`
		ArtifactContract string   `json:"artifact_contract"`
		BrowserActKinds  []string `json:"browser_act_kinds"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act concrete selected-route click output: %v", err)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_act") ||
		browserStringSliceContains(payload.BrowserTools, "browser_screenshot") ||
		browserStringSliceContains(payload.ArtifactTools, "browser_act") ||
		browserStringSliceContains(payload.ArtifactKinds, "screenshot") ||
		payload.ArtifactContract != "" ||
		!browserStringSliceContains(payload.BrowserActKinds, "click") ||
		browserStringSliceContains(payload.BrowserActKinds, "screenshot") {
		t.Fatalf("expected browser_act payload to use concrete selected-route capabilities, got %#v", payload)
	}
}
