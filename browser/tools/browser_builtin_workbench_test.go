package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_RuntimeWorkbench(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-workbench", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-71",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{
						Profile:    "workbench",
						BrowserApp: "Chromium",
						Status:     "running",
						Running:    true,
						Connected:  true,
					},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		SessionRunRegistry:   sessionRunRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	sessionRegistry.TrackTabs("browser-runtime-workbench", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/dashboard",
			Title:      "Dashboard",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
	}, 1)

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"workbench","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench: %v", err)
	}
	var payload struct {
		Action                             string   `json:"action"`
		Status                             string   `json:"status"`
		BrowserSurface                     string   `json:"browser_surface"`
		BrowserOptInTargets                []string `json:"browser_opt_in_targets"`
		WorkbenchReady                     bool     `json:"workbench_ready"`
		WorkbenchSections                  []string `json:"workbench_sections"`
		WorkbenchPrimaryBrowserAction      string   `json:"workbench_primary_browser_action"`
		WorkbenchPrimaryNodeAction         string   `json:"workbench_primary_node_action"`
		WorkbenchNextStep                  string   `json:"workbench_next_step"`
		WorkbenchRecommendedBrowserActions []string `json:"workbench_recommended_browser_actions"`
		WorkbenchRecommendedNodeActions    []string `json:"workbench_recommended_node_actions"`
		WorkbenchDiagnostics               *struct {
			Category             string `json:"category"`
			State                string `json:"state"`
			SummaryCode          string `json:"summary_code"`
			PrimaryBrowserAction string `json:"primary_browser_action"`
			PrimaryNodeAction    string `json:"primary_node_action"`
			NextStep             string `json:"next_step"`
		} `json:"workbench_diagnostics"`
		Workbench        *browserRuntimeWorkbenchSurfaceSummary `json:"workbench"`
		WorkbenchDisplay *browserRuntimeWorkbenchDisplaySummary `json:"workbench_display"`
		Review           *browserReviewSurfaceSummary           `json:"review"`
		Explanation      *browserTopLevelSummary                `json:"explanation"`
		Diagnostics      *browserTopLevelSummary                `json:"diagnostics"`
		WorkbenchSummary *browserTopLevelSummary                `json:"workbench_summary"`
		Summary          *browserTopLevelSummary                `json:"summary"`
		Display          *browserTopLevelDisplaySummary         `json:"display"`
		RouteResolution  struct {
			ProfileSource       string `json:"profile_source"`
			RuntimeTargetSource string `json:"runtime_target_source"`
			TargetSource        string `json:"target_source"`
		} `json:"route_resolution"`
		RuntimeActions []string `json:"runtime_actions"`
		ProfileStatus  struct {
			Profile string `json:"profile"`
			Status  string `json:"status"`
		} `json:"profile_status"`
		Profiles []struct {
			Profile  string `json:"profile"`
			Selected bool   `json:"selected"`
		} `json:"profiles"`
		SessionTargetCount int `json:"session_target_count"`
		SessionBinding     struct {
			ActiveNodeRunID              string `json:"active_node_run_id"`
			ActiveBrowserProfile         string `json:"active_browser_profile"`
			SelectedBrowserProfileSource string `json:"selected_browser_profile_source"`
			SelectedBrowserTargetSource  string `json:"selected_browser_target_source"`
			Coordination                 struct {
				State string `json:"state"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode workbench output: %v", err)
	}
	if payload.Action != "workbench" || payload.Status != "ok" || !payload.WorkbenchReady {
		t.Fatalf("unexpected runtime workbench payload: %#v", payload)
	}
	for _, want := range []string{"workbench", "status", "profiles", "route", "sessions", "coordination"} {
		if want == "workbench" {
			if !browserStringSliceContains(payload.RuntimeActions, want) {
				t.Fatalf("expected runtime actions to include workbench, got %#v", payload.RuntimeActions)
			}
			continue
		}
		if !browserStringSliceContains(payload.WorkbenchSections, want) {
			t.Fatalf("expected workbench sections to include %q, got %#v", want, payload.WorkbenchSections)
		}
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "running" {
		t.Fatalf("unexpected workbench profile status: %#v", payload.ProfileStatus)
	}
	if payload.RouteResolution.ProfileSource != "explicit_request" || payload.RouteResolution.RuntimeTargetSource != "explicit_request" || payload.RouteResolution.TargetSource != "tracked_active_tab" {
		t.Fatalf("unexpected route resolution: %#v", payload.RouteResolution)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Profile != "workbench" || !payload.Profiles[0].Selected {
		t.Fatalf("unexpected workbench profiles: %#v", payload.Profiles)
	}
	if payload.SessionTargetCount != 1 {
		t.Fatalf("expected one tracked session target in workbench view, got %#v", payload.SessionTargetCount)
	}
	if payload.SessionBinding.ActiveNodeRunID != "run-71" || payload.SessionBinding.ActiveBrowserProfile != "workbench" {
		t.Fatalf("unexpected workbench session binding: %#v", payload.SessionBinding)
	}
	if payload.SessionBinding.SelectedBrowserProfileSource != "select_profile" || payload.SessionBinding.SelectedBrowserTargetSource != "tracked_active_tab" {
		t.Fatalf("unexpected workbench session binding selection sources: %#v", payload.SessionBinding)
	}
	if payload.SessionBinding.Coordination.State == "" {
		t.Fatalf("expected workbench coordination summary, got %#v", payload.SessionBinding)
	}
	if payload.WorkbenchPrimaryBrowserAction != "browser_runtime action=workbench" || payload.WorkbenchPrimaryNodeAction != "nodes action=run_status" || payload.WorkbenchNextStep != "nodes action=run_status" {
		t.Fatalf("unexpected top-level workbench action plan: %#v", payload)
	}
	if !browserStringSliceContains(payload.WorkbenchRecommendedBrowserActions, "browser_runtime action=workbench") || !browserStringSliceContains(payload.WorkbenchRecommendedNodeActions, "nodes action=run_wait") {
		t.Fatalf("unexpected top-level workbench recommended actions: %#v", payload)
	}
	if payload.WorkbenchDiagnostics == nil ||
		payload.WorkbenchDiagnostics.Category != "coordination" ||
		payload.WorkbenchDiagnostics.State != "action_plan_available" ||
		payload.WorkbenchDiagnostics.SummaryCode != "workbench_action_plan" ||
		payload.WorkbenchDiagnostics.PrimaryBrowserAction != "browser_runtime action=workbench" ||
		payload.WorkbenchDiagnostics.PrimaryNodeAction != "nodes action=run_status" ||
		payload.WorkbenchDiagnostics.NextStep != "nodes action=run_status" {
		t.Fatalf("unexpected workbench diagnostics summary: %#v", payload.WorkbenchDiagnostics)
	}
	if payload.Workbench == nil ||
		!payload.Workbench.Ready ||
		!browserStringSliceContains(payload.Workbench.Sections, "coordination") ||
		payload.Workbench.Explanation != nil ||
		payload.Workbench.Diagnostics == nil ||
		payload.Workbench.Diagnostics.Category != "coordination" ||
		payload.Workbench.Diagnostics.State != "action_plan_available" ||
		payload.Workbench.Diagnostics.SummaryCode != "workbench_action_plan" ||
		payload.Workbench.Diagnostics.PrimaryBrowserAction != "browser_runtime action=workbench" ||
		payload.Workbench.Diagnostics.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Workbench.Diagnostics.NextStep != "nodes action=run_status" ||
		payload.Workbench.Summary == nil ||
		payload.Workbench.Summary.Category != "coordination" ||
		payload.Workbench.Summary.State != "action_plan_available" ||
		payload.Workbench.Summary.SummaryCode != "workbench_action_plan" ||
		payload.Workbench.PrimaryBrowserAction != "browser_runtime action=workbench" ||
		payload.Workbench.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Workbench.NextStep != "nodes action=run_status" ||
		!browserStringSliceContains(payload.Workbench.RecommendedBrowserActions, "browser_runtime action=workbench") ||
		!browserStringSliceContains(payload.Workbench.RecommendedNodeActions, "nodes action=run_wait") {
		t.Fatalf("unexpected runtime workbench surface summary: %#v", payload.Workbench)
	}
	if payload.WorkbenchDisplay == nil ||
		!payload.WorkbenchDisplay.Ready ||
		!browserStringSliceContains(payload.WorkbenchDisplay.Sections, "coordination") ||
		payload.WorkbenchDisplay.Category != "coordination" ||
		payload.WorkbenchDisplay.State != "action_plan_available" ||
		payload.WorkbenchDisplay.SummaryCode != "workbench_action_plan" ||
		payload.WorkbenchDisplay.PrimaryBrowserAction != "browser_runtime action=workbench" ||
		payload.WorkbenchDisplay.PrimaryNodeAction != "nodes action=run_status" ||
		payload.WorkbenchDisplay.NextStep != "nodes action=run_status" ||
		payload.WorkbenchDisplay.ResolvedViaFallback {
		t.Fatalf("unexpected runtime workbench display summary: %#v", payload.WorkbenchDisplay)
	}
	if payload.Explanation != nil {
		t.Fatalf("expected runtime workbench coordination payload not to synthesize explanation, got %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "coordination" ||
		payload.Diagnostics.State != "action_plan_available" ||
		payload.Diagnostics.SummaryCode != "workbench_action_plan" ||
		payload.Diagnostics.PrimaryBrowserAction != "browser_runtime action=workbench" ||
		payload.Diagnostics.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Diagnostics.NextStep != "nodes action=run_status" ||
		payload.Diagnostics.ResolvedViaFallback {
		t.Fatalf("unexpected runtime top-level diagnostics alias: %#v", payload.Diagnostics)
	}
	if payload.WorkbenchSummary == nil ||
		payload.WorkbenchSummary.Category != "coordination" ||
		payload.WorkbenchSummary.State != "action_plan_available" ||
		payload.WorkbenchSummary.SummaryCode != "workbench_action_plan" ||
		payload.WorkbenchSummary.PrimaryBrowserAction != "browser_runtime action=workbench" ||
		payload.WorkbenchSummary.PrimaryNodeAction != "nodes action=run_status" ||
		payload.WorkbenchSummary.NextStep != "nodes action=run_status" ||
		payload.WorkbenchSummary.ResolvedViaFallback {
		t.Fatalf("unexpected workbench summary alias: %#v", payload.WorkbenchSummary)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "coordination" ||
		payload.Summary.State != "action_plan_available" ||
		payload.Summary.SummaryCode != "workbench_action_plan" ||
		payload.Summary.PrimaryBrowserAction != "browser_runtime action=workbench" ||
		payload.Summary.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Summary.NextStep != "nodes action=run_status" ||
		payload.Summary.ResolvedViaFallback {
		t.Fatalf("unexpected workbench top-level summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		!payload.Display.Ready ||
		!browserStringSliceContains(payload.Display.Sections, "coordination") ||
		payload.Display.Category != "coordination" ||
		payload.Display.State != "action_plan_available" ||
		payload.Display.SummaryCode != "workbench_action_plan" ||
		payload.Display.PrimaryBrowserAction != "browser_runtime action=workbench" ||
		payload.Display.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Display.NextStep != "nodes action=run_status" ||
		payload.Display.ResolvedViaFallback {
		t.Fatalf("unexpected runtime top-level display: %#v", payload.Display)
	}
}

func TestRegisterBrowserTools_RuntimeWorkbenchCarriesExplicitManagedRouteSurface(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-workbench-explicit-opt-in", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-71",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run",
	})
	hostBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
				fakeBrowserBackend: &fakeBrowserBackend{},
				runtimeInfo:        BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"},
			}},
			capabilities: BrowserCapabilitiesForActKinds([]string{"open"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			requested = normalizeBrowserRuntimeInfo(requested)
			if strings.TrimSpace(requested.Target) == "host" {
				return BrowserRuntimeInfo{Backend: "custom-playwright", Profile: "workbench", Target: "host"}, nil
			}
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	nodeBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
				fakeBrowserBackend: &fakeBrowserBackend{
					runtimeStatusResult: BrowserProfileStatusResult{
						Backend:    "proxy",
						BrowserApp: "Chromium",
						Profile:    "workbench",
						Status:     "running",
						Running:    true,
						Connected:  true,
					},
					runtimeProfilesResult: BrowserProfilesResult{
						Backend:        "proxy",
						DefaultProfile: "workbench",
						Profiles: []BrowserProfileInfo{
							{
								Profile:    "workbench",
								BrowserApp: "Chromium",
								Status:     "running",
								Running:    true,
								Connected:  true,
							},
						},
					},
				},
				runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
			}},
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
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		SessionRunRegistry:   sessionRunRegistry,
		EnabledTools:         []string{"browser_runtime", "browser_click", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench-explicit-opt-in")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile explicit opt-in: %v", err)
	}
	sessionRegistry.TrackTabs("browser-runtime-workbench-explicit-opt-in", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/dashboard",
			Title:      "Dashboard",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
	}, 1)

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"workbench","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench explicit opt-in: %v", err)
	}
	var payload struct {
		Status              string                                 `json:"status"`
		BrowserTools        []string                               `json:"browser_tools"`
		ArtifactTools       []string                               `json:"artifact_tools"`
		ArtifactKinds       []string                               `json:"artifact_kinds"`
		ArtifactContract    string                                 `json:"artifact_contract"`
		BrowserActKinds     []string                               `json:"browser_act_kinds"`
		BrowserSurface      string                                 `json:"browser_surface"`
		BrowserOptInTargets []string                               `json:"browser_opt_in_targets"`
		Workbench           *browserRuntimeWorkbenchSurfaceSummary `json:"workbench"`
		Surface             *browserTopLevelSurfaceSummary         `json:"surface"`
		View                *browserTopLevelViewSummary            `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode runtime workbench explicit opt-in output: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("unexpected runtime workbench explicit opt-in payload: %#v", payload)
	}
	if payload.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected runtime workbench payload to expose explicit managed route surface, got %#v", payload)
	}
	if !browserStringSliceContains(payload.BrowserTools, "browser_runtime") ||
		!browserStringSliceContains(payload.BrowserTools, "browser_act") ||
		!browserStringSliceContains(payload.BrowserActKinds, "click") {
		t.Fatalf("expected runtime workbench payload to expose selected-route capability metadata, got %#v", payload)
	}
	if payload.Workbench == nil ||
		payload.Workbench.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.Workbench.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected runtime workbench surface to inherit explicit managed route surface, got %#v", payload.Workbench)
	}
	if payload.Workbench == nil ||
		!browserStringSliceContains(payload.Workbench.BrowserTools, "browser_runtime") ||
		!reflect.DeepEqual(payload.Workbench.ArtifactTools, payload.ArtifactTools) ||
		!reflect.DeepEqual(payload.Workbench.ArtifactKinds, payload.ArtifactKinds) ||
		payload.Workbench.ArtifactContract != payload.ArtifactContract ||
		!browserStringSliceContains(payload.Workbench.BrowserActKinds, "click") {
		t.Fatalf("expected runtime workbench surface to inherit selected-route capability metadata, got %#v", payload.Workbench)
	}
	if payload.Surface == nil ||
		payload.Surface.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.Surface.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected top-level surface to keep explicit managed route surface, got %#v", payload.Surface)
	}
	if payload.Surface == nil ||
		!browserStringSliceContains(payload.Surface.BrowserTools, "browser_runtime") ||
		!reflect.DeepEqual(payload.Surface.ArtifactTools, payload.ArtifactTools) ||
		!reflect.DeepEqual(payload.Surface.ArtifactKinds, payload.ArtifactKinds) ||
		payload.Surface.ArtifactContract != payload.ArtifactContract ||
		!browserStringSliceContains(payload.Surface.BrowserActKinds, "click") {
		t.Fatalf("expected top-level surface to keep selected-route capability metadata, got %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.View.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected top-level view to keep explicit managed route surface, got %#v", payload.View)
	}
	if payload.View == nil ||
		!browserStringSliceContains(payload.View.BrowserTools, "browser_runtime") ||
		!reflect.DeepEqual(payload.View.ArtifactTools, payload.ArtifactTools) ||
		!reflect.DeepEqual(payload.View.ArtifactKinds, payload.ArtifactKinds) ||
		payload.View.ArtifactContract != payload.ArtifactContract ||
		!browserStringSliceContains(payload.View.BrowserActKinds, "click") {
		t.Fatalf("expected top-level view to keep selected-route capability metadata, got %#v", payload.View)
	}
}

func TestRegisterBrowserTools_BrowserRuntimeWorkbenchReturnsReviewDiagnosticsSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-runtime-workbench-review", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-71",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		SessionRunRegistry:   sessionRunRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench-review")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser_runtime select_profile: %v", err)
	}
	tracked := sessionRegistry.TrackTabs("browser-runtime-workbench-review", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/dashboard",
			Title:      "Dashboard",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
		{
			TabIndex:   3,
			URL:        "https://popup.example/offer",
			Title:      "Offer",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
	}, 1)
	if len(tracked) != 2 {
		t.Fatalf("expected tracked tabs, got %#v", tracked)
	}
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-workbench-review", agentxbrowserruntime.BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "workbench",
		Target:     "node",
		BrowserApp: "Chromium",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         tracked[1].ID,
		TabIndex:   tracked[1].TabIndex,
		URL:        tracked[1].URL,
		Title:      tracked[1].Title,
		BrowserApp: tracked[1].BrowserApp,
		Backend:    tracked[1].Backend,
		Profile:    tracked[1].Profile,
		Target:     tracked[1].Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"workbench","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench review: %v", err)
	}
	var payload struct {
		WorkbenchPrimaryBrowserAction string `json:"workbench_primary_browser_action"`
		WorkbenchPrimaryNodeAction    string `json:"workbench_primary_node_action"`
		WorkbenchNextStep             string `json:"workbench_next_step"`
		WorkbenchDiagnostics          *struct {
			Category             string `json:"category"`
			State                string `json:"state"`
			SummaryCode          string `json:"summary_code"`
			NextStepAlias        string `json:"next_step_alias"`
			ManualRetryHint      string `json:"manual_retry_hint"`
			PrimaryBrowserAction string `json:"primary_browser_action"`
			PrimaryNodeAction    string `json:"primary_node_action"`
			NextStep             string `json:"next_step"`
		} `json:"workbench_diagnostics"`
		Workbench        *browserRuntimeWorkbenchSurfaceSummary `json:"workbench"`
		WorkbenchDisplay *browserRuntimeWorkbenchDisplaySummary `json:"workbench_display"`
		Review           *browserReviewSurfaceSummary           `json:"review"`
		View             *browserTopLevelViewSummary            `json:"view"`
		Explanation      *browserTopLevelSummary                `json:"explanation"`
		Diagnostics      *browserTopLevelSummary                `json:"diagnostics"`
		Summary          *browserTopLevelSummary                `json:"summary"`
		Display          *browserTopLevelDisplaySummary         `json:"display"`
		Surface          *browserTopLevelSurfaceSummary         `json:"surface"`
		SessionBinding   struct {
			PendingTargetReviewCount    int `json:"pending_target_review_count"`
			BlockedAutoFollowRouteCount int `json:"blocked_auto_follow_route_count"`
		} `json:"session_binding"`
		SessionRoutes []struct {
			FollowPolicyState string `json:"follow_policy_state"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode runtime workbench review output: %v", err)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].FollowPolicyState != "popup_review_required" {
		t.Fatalf("expected popup review route summary, got %#v", payload.SessionRoutes)
	}
	if payload.SessionBinding.PendingTargetReviewCount != 1 || payload.SessionBinding.BlockedAutoFollowRouteCount != 1 {
		t.Fatalf("expected workbench review counters, got %#v", payload.SessionBinding)
	}
	if payload.WorkbenchPrimaryBrowserAction != "browser_runtime action=sessions" || payload.WorkbenchPrimaryNodeAction != "nodes action=run_status" || payload.WorkbenchNextStep != "browser_runtime action=sessions" {
		t.Fatalf("expected workbench action plan to prefer tabs review, got %#v", payload)
	}
	if payload.WorkbenchDiagnostics == nil ||
		payload.WorkbenchDiagnostics.Category != "review" ||
		payload.WorkbenchDiagnostics.State != "manual_confirmation_required" ||
		payload.WorkbenchDiagnostics.SummaryCode != "popup_review_required" ||
		payload.WorkbenchDiagnostics.NextStepAlias != "sessions" ||
		payload.WorkbenchDiagnostics.ManualRetryHint != "rerun_with_force" ||
		payload.WorkbenchDiagnostics.PrimaryBrowserAction != "browser_runtime action=sessions" ||
		payload.WorkbenchDiagnostics.PrimaryNodeAction != "nodes action=run_status" ||
		payload.WorkbenchDiagnostics.NextStep != "browser_runtime action=sessions" {
		t.Fatalf("unexpected runtime workbench review diagnostics: %#v", payload.WorkbenchDiagnostics)
	}
	if payload.Workbench == nil ||
		payload.Workbench.Review == nil ||
		payload.Workbench.Review.PolicyState != "popup_review_required" ||
		payload.Workbench.Review.Decision != "session_target_popup_review_required" ||
		payload.Workbench.Review.Summary == nil ||
		payload.Workbench.Review.Summary.SummaryCode != "popup_review_required" ||
		payload.Workbench.Review.Summary.ManualRetryHint != "rerun_with_force" ||
		payload.Workbench.Review.Display == nil ||
		payload.Workbench.Review.Display.SummaryCode != "popup_review_required" ||
		payload.Workbench.Review.Display.ManualRetryHint != "rerun_with_force" ||
		payload.Workbench.Explanation == nil ||
		payload.Workbench.Explanation.Category != "review" ||
		payload.Workbench.Explanation.State != "manual_confirmation_required" ||
		payload.Workbench.Explanation.SummaryCode != "popup_review_required" ||
		payload.Workbench.Explanation.NextStepAlias != "sessions" ||
		payload.Workbench.Explanation.ManualRetryHint != "rerun_with_force" ||
		payload.Workbench.Diagnostics == nil ||
		payload.Workbench.Diagnostics.Category != "review" ||
		payload.Workbench.Diagnostics.State != "manual_confirmation_required" ||
		payload.Workbench.Diagnostics.SummaryCode != "popup_review_required" ||
		payload.Workbench.Diagnostics.ManualRetryHint != "rerun_with_force" ||
		payload.Workbench.Summary == nil ||
		payload.Workbench.Summary.Category != "review" ||
		payload.Workbench.Summary.State != "manual_confirmation_required" ||
		payload.Workbench.Summary.SummaryCode != "popup_review_required" ||
		payload.Workbench.Summary.ManualRetryHint != "rerun_with_force" ||
		payload.Workbench.PrimaryBrowserAction != "browser_runtime action=sessions" ||
		payload.Workbench.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Workbench.NextStep != "browser_runtime action=sessions" {
		t.Fatalf("unexpected runtime workbench review surface: %#v", payload.Workbench)
	}
	if payload.WorkbenchDisplay == nil ||
		payload.WorkbenchDisplay.Category != "review" ||
		payload.WorkbenchDisplay.State != "manual_confirmation_required" ||
		payload.WorkbenchDisplay.SummaryCode != "popup_review_required" ||
		payload.WorkbenchDisplay.NextStepAlias != "sessions" ||
		payload.WorkbenchDisplay.ManualRetryHint != "rerun_with_force" ||
		payload.WorkbenchDisplay.PrimaryBrowserAction != "browser_runtime action=sessions" ||
		payload.WorkbenchDisplay.NextStep != "browser_runtime action=sessions" {
		t.Fatalf("unexpected runtime workbench review display: %#v", payload.WorkbenchDisplay)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "review" ||
		payload.Explanation.State != "manual_confirmation_required" ||
		payload.Explanation.SummaryCode != "popup_review_required" ||
		payload.Explanation.NextStepAlias != "sessions" ||
		payload.Explanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected top-level review explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "review" ||
		payload.Diagnostics.State != "manual_confirmation_required" ||
		payload.Diagnostics.SummaryCode != "popup_review_required" ||
		payload.Diagnostics.NextStepAlias != "sessions" ||
		payload.Diagnostics.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected top-level review diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "review" ||
		payload.Summary.State != "manual_confirmation_required" ||
		payload.Summary.SummaryCode != "popup_review_required" ||
		payload.Summary.NextStepAlias != "sessions" ||
		payload.Summary.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected top-level review summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "review" ||
		payload.Display.State != "manual_confirmation_required" ||
		payload.Display.SummaryCode != "popup_review_required" ||
		payload.Display.NextStepAlias != "sessions" ||
		payload.Display.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected top-level review display: %#v", payload.Display)
	}
	if payload.Review == nil ||
		payload.Review.PolicyState != "popup_review_required" ||
		payload.Review.Decision != "session_target_popup_review_required" ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.SummaryCode != "popup_review_required" ||
		payload.Review.Summary.ManualRetryHint != "rerun_with_force" ||
		payload.Review.Display == nil ||
		payload.Review.Display.SummaryCode != "popup_review_required" ||
		payload.Review.Display.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected top-level review surface: %#v", payload.Review)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "review" ||
		payload.Surface.State != "manual_confirmation_required" ||
		payload.Surface.SummaryCode != "popup_review_required" ||
		payload.Surface.NextStepAlias != "sessions" ||
		payload.Surface.ManualRetryHint != "rerun_with_force" ||
		payload.Surface.PrimaryBrowserAction != "browser_runtime action=sessions" ||
		payload.Surface.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Surface.NextStep != "browser_runtime action=sessions" ||
		payload.Surface.ReviewPolicyState != "popup_review_required" ||
		payload.Surface.ReviewDecision != "session_target_popup_review_required" ||
		payload.Surface.ReviewReady {
		t.Fatalf("unexpected top-level review surface alias: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "workbench" ||
		!payload.View.Ready ||
		!browserStringSliceContains(payload.View.Sections, "coordination") ||
		payload.View.Category != "review" ||
		payload.View.State != "manual_confirmation_required" ||
		payload.View.SummaryCode != "popup_review_required" ||
		payload.View.NextStepAlias != "sessions" ||
		payload.View.ManualRetryHint != "rerun_with_force" ||
		payload.View.PrimaryBrowserAction != "browser_runtime action=sessions" ||
		payload.View.NextStep != "browser_runtime action=sessions" ||
		payload.View.Review == nil ||
		payload.View.Review.PolicyState != "popup_review_required" ||
		payload.View.Review.Decision != "session_target_popup_review_required" ||
		payload.View.Review.Display == nil ||
		payload.View.Review.Display.ManualRetryHint != "rerun_with_force" ||
		payload.View.Review.Ready {
		t.Fatalf("unexpected top-level review view alias: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserWorkbenchReturnsUnifiedDiagnosticsAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionRunRegistry.Record("browser-unified-workbench", agentxbrowserruntime.SharedSessionRunInfo{
		RunID:    "run-71",
		NodeID:   "node-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "workbench",
				Status:     "running",
				Running:    true,
				Connected:  true,
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{
						Profile:    "workbench",
						BrowserApp: "Chromium",
						Status:     "running",
						Running:    true,
						Connected:  true,
					},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "workbench", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		SessionRunRegistry:   sessionRunRegistry,
		EnabledTools:         []string{"browser"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-unified-workbench")
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"select_profile","runtime_target":"node","profile":"workbench"}`,
	}); err != nil {
		t.Fatalf("browser unified select_profile: %v", err)
	}
	sessionRegistry.TrackTabs("browser-unified-workbench", []agentxbrowserruntime.BrowserSessionTarget{
		{
			TabIndex:   1,
			URL:        "https://93.184.216.34/dashboard",
			Title:      "Dashboard",
			BrowserApp: "Chromium",
			Backend:    "proxy",
			Profile:    "workbench",
			Target:     "node",
		},
	}, 1)

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"workbench","runtime_target":"node","profile":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser unified workbench diagnostics alias: %v", err)
	}
	var payload struct {
		Action               string                                     `json:"action"`
		Status               string                                     `json:"status"`
		BrowserSurface       string                                     `json:"browser_surface"`
		BrowserOptInTargets  []string                                   `json:"browser_opt_in_targets"`
		WorkbenchDiagnostics *browserRuntimeWorkbenchDiagnosticsSummary `json:"workbench_diagnostics"`
		Workbench            *browserRuntimeWorkbenchSurfaceSummary     `json:"workbench"`
		WorkbenchDisplay     *browserRuntimeWorkbenchDisplaySummary     `json:"workbench_display"`
		Explanation          *browserTopLevelSummary                    `json:"explanation"`
		WorkbenchSummary     *browserTopLevelSummary                    `json:"workbench_summary"`
		Diagnostics          *browserTopLevelSummary                    `json:"diagnostics"`
		Summary              *browserTopLevelSummary                    `json:"summary"`
		Display              *browserTopLevelDisplaySummary             `json:"display"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode unified workbench diagnostics output: %v", err)
	}
	if payload.Action != "workbench" || payload.Status != "ok" {
		t.Fatalf("unexpected unified workbench payload: %#v", payload)
	}
	if payload.WorkbenchDiagnostics == nil ||
		payload.WorkbenchDiagnostics.Category != "coordination" ||
		payload.WorkbenchDiagnostics.State != "action_plan_available" ||
		payload.WorkbenchDiagnostics.SummaryCode != "workbench_action_plan" {
		t.Fatalf("unexpected unified workbench diagnostics summary: %#v", payload.WorkbenchDiagnostics)
	}
	if payload.Workbench == nil ||
		!payload.Workbench.Ready ||
		!browserStringSliceContains(payload.Workbench.Sections, "coordination") ||
		payload.Workbench.Explanation != nil ||
		payload.Workbench.Diagnostics == nil ||
		payload.Workbench.Diagnostics.Category != "coordination" ||
		payload.Workbench.Diagnostics.State != "action_plan_available" ||
		payload.Workbench.Diagnostics.SummaryCode != "workbench_action_plan" ||
		payload.Workbench.Diagnostics.PrimaryBrowserAction != "browser" ||
		payload.Workbench.Diagnostics.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Workbench.Diagnostics.NextStep != "nodes action=run_status" ||
		payload.Workbench.Summary == nil ||
		payload.Workbench.Summary.Category != "coordination" ||
		payload.Workbench.Summary.State != "action_plan_available" ||
		payload.Workbench.Summary.SummaryCode != "workbench_action_plan" ||
		payload.Workbench.PrimaryBrowserAction != "browser" ||
		payload.Workbench.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Workbench.NextStep != "nodes action=run_status" {
		t.Fatalf("unexpected unified workbench surface summary: %#v", payload.Workbench)
	}
	if payload.WorkbenchDisplay == nil ||
		!payload.WorkbenchDisplay.Ready ||
		!browserStringSliceContains(payload.WorkbenchDisplay.Sections, "coordination") ||
		payload.WorkbenchDisplay.Category != "coordination" ||
		payload.WorkbenchDisplay.State != "action_plan_available" ||
		payload.WorkbenchDisplay.SummaryCode != "workbench_action_plan" ||
		payload.WorkbenchDisplay.PrimaryBrowserAction != "browser" ||
		payload.WorkbenchDisplay.PrimaryNodeAction != "nodes action=run_status" ||
		payload.WorkbenchDisplay.NextStep != "nodes action=run_status" ||
		payload.WorkbenchDisplay.ResolvedViaFallback {
		t.Fatalf("unexpected unified workbench display summary: %#v", payload.WorkbenchDisplay)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "coordination" ||
		payload.Diagnostics.State != "action_plan_available" ||
		payload.Diagnostics.SummaryCode != "workbench_action_plan" ||
		payload.Diagnostics.PrimaryBrowserAction != "browser" ||
		payload.Diagnostics.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Diagnostics.NextStep != "nodes action=run_status" ||
		payload.Diagnostics.ResolvedViaFallback {
		t.Fatalf("unexpected unified top-level diagnostics alias: %#v", payload.Diagnostics)
	}
	if payload.Explanation != nil {
		t.Fatalf("expected unified workbench coordination payload not to synthesize explanation, got %#v", payload.Explanation)
	}
	if payload.WorkbenchSummary == nil ||
		payload.WorkbenchSummary.Category != "coordination" ||
		payload.WorkbenchSummary.State != "action_plan_available" ||
		payload.WorkbenchSummary.SummaryCode != "workbench_action_plan" ||
		payload.WorkbenchSummary.PrimaryBrowserAction != "browser" ||
		payload.WorkbenchSummary.PrimaryNodeAction != "nodes action=run_status" ||
		payload.WorkbenchSummary.NextStep != "nodes action=run_status" ||
		payload.WorkbenchSummary.ResolvedViaFallback {
		t.Fatalf("unexpected unified workbench summary alias: %#v", payload.WorkbenchSummary)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "coordination" ||
		payload.Summary.State != "action_plan_available" ||
		payload.Summary.SummaryCode != "workbench_action_plan" ||
		payload.Summary.PrimaryBrowserAction != "browser" ||
		payload.Summary.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Summary.NextStep != "nodes action=run_status" ||
		payload.Summary.ResolvedViaFallback {
		t.Fatalf("unexpected unified top-level summary alias: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		!payload.Display.Ready ||
		!browserStringSliceContains(payload.Display.Sections, "coordination") ||
		payload.Display.Category != "coordination" ||
		payload.Display.State != "action_plan_available" ||
		payload.Display.SummaryCode != "workbench_action_plan" ||
		payload.Display.PrimaryBrowserAction != "browser" ||
		payload.Display.PrimaryNodeAction != "nodes action=run_status" ||
		payload.Display.NextStep != "nodes action=run_status" ||
		payload.Display.ResolvedViaFallback {
		t.Fatalf("unexpected unified top-level display alias: %#v", payload.Display)
	}
}
