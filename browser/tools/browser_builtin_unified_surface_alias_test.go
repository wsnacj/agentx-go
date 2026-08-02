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

func TestRegisterBrowserTools_UnifiedBrowserInventoryAliasDelegatesProfiles(t *testing.T) {
	reg := llmxtools.NewRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
					{Profile: "relay", BrowserApp: "Chromium", Status: "stopped"},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"inventory","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser inventory: %v", err)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 1 {
		t.Fatalf("expected inventory alias to reuse profiles path, got %#v", nodeBackend.runtimeProfilesReqs)
	}
	if !strings.Contains(out, `"action":"profiles"`) || !strings.Contains(out, `"profiles":[`) || !strings.Contains(out, `"profile":"isolated"`) || !strings.Contains(out, `"profile":"relay"`) {
		t.Fatalf("unexpected browser inventory output: %s", out)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserInventoryAliasHidesImplicitHostProfileSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-inventory-implicit-host-profile-summary")
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-inventory-implicit-host-profile-summary", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-inventory-implicit-host-profile-summary", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"inventory"}`,
	})
	if err != nil {
		t.Fatalf("browser inventory implicit host profile summary: %v", err)
	}
	var payload struct {
		Action       string `json:"action"`
		Status       string `json:"status"`
		Note         string `json:"note"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute *struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		RouteResolution any    `json:"route_resolution"`
		DefaultProfile  string `json:"default_profile"`
		Profiles        []any  `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "profiles" || payload.Status != "unsupported" {
		t.Fatalf("unexpected inventory alias payload: %#v", payload)
	}
	if payload.Note != "" {
		t.Fatalf("expected inventory alias to short-circuit hidden implicit-host diagnostics without route note, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "" || payload.DefaultRoute.Profile != "" || payload.DefaultRoute.RuntimeTarget != "" {
		t.Fatalf("expected inventory alias to keep implicit host out of default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected inventory alias to hide implicit host top-level route state, got %#v", payload)
	}
	if payload.DefaultProfile != "" || len(payload.Profiles) != 0 {
		t.Fatalf("expected inventory alias to hide implicit host profile summary, got %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserHandlesAliasDelegatesSessions(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-handles")
	sessionRegistry.TrackTab("browser-unified-handles", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/handles",
		Title:      "Handles",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-handles", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}},
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"handles","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser handles: %v", err)
	}
	var payload struct {
		Action             string `json:"action"`
		Status             string `json:"status"`
		SessionID          string `json:"session_id"`
		SessionTargetCount int    `json:"session_target_count"`
		SessionBinding     struct {
			ActiveBrowserProfile string `json:"active_browser_profile"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "sessions" || payload.Status != "ok" || payload.SessionID != "browser-unified-handles" || payload.SessionTargetCount != 1 || payload.SessionBinding.ActiveBrowserProfile != "isolated" {
		t.Fatalf("unexpected browser handles payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserHandlesAliasHidesImplicitHostSelections(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-handles-implicit-host-selections")
	sessionRegistry.TrackTab("browser-unified-handles-implicit-host-selections", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://93.184.216.34/hidden-host",
		Title:      "Hidden Host",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-handles-implicit-host-selections", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-handles-implicit-host-selections", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"handles"}`,
	})
	if err != nil {
		t.Fatalf("browser handles implicit host selections: %v", err)
	}
	var payload struct {
		Action                  string `json:"action"`
		Status                  string `json:"status"`
		SelectedRoute           any    `json:"selected_route"`
		RouteResolution         any    `json:"route_resolution"`
		SessionID               string `json:"session_id"`
		SessionTargetCount      int    `json:"session_target_count"`
		SessionProfileSelection string `json:"session_profile_selection"`
		SessionTargetSelection  any    `json:"session_target_selection"`
		SessionBinding          struct {
			SelectedBrowserProfile string `json:"selected_browser_profile"`
			CurrentTargetID        string `json:"current_target_id"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "sessions" || payload.Status != "ok" || payload.SessionID != "browser-unified-handles-implicit-host-selections" || payload.SessionTargetCount != 1 {
		t.Fatalf("unexpected handles alias payload: %#v", payload)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected handles alias to hide implicit host top-level route state, got %#v", payload)
	}
	if payload.SessionProfileSelection != "" || payload.SessionTargetSelection != nil {
		t.Fatalf("expected handles alias to hide implicit host top-level selections, got %#v", payload)
	}
	if payload.SessionBinding.SelectedBrowserProfile != "" || payload.SessionBinding.CurrentTargetID != "" {
		t.Fatalf("expected handles alias to hide implicit host session binding selections, got %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserInspectAliasHidesImplicitHostProfileSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-inspect-implicit-host-profile-summary")
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-inspect-implicit-host-profile-summary", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-inspect-implicit-host-profile-summary", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"inspect"}`,
	})
	if err != nil {
		t.Fatalf("browser inspect implicit host profile summary: %v", err)
	}
	var payload struct {
		Action       string `json:"action"`
		Status       string `json:"status"`
		DefaultRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"default_route"`
		SelectedRoute   any    `json:"selected_route"`
		RouteResolution any    `json:"route_resolution"`
		DefaultProfile  string `json:"default_profile"`
		ProfileStatus   any    `json:"profile_status"`
		SessionBinding  struct {
			BrowserProfileCount    int    `json:"browser_profile_count"`
			ActiveBrowserProfile   string `json:"active_browser_profile"`
			SelectedBrowserProfile string `json:"selected_browser_profile"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "status" || payload.Status != "ok" {
		t.Fatalf("unexpected inspect alias payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "" || payload.DefaultRoute.Profile != "" || payload.DefaultRoute.RuntimeTarget != "" {
		t.Fatalf("expected inspect alias to keep implicit host out of default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected inspect alias to hide implicit host top-level route state, got %#v", payload)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil {
		t.Fatalf("expected inspect alias to hide implicit host top-level profile summary, got %#v", payload)
	}
	if payload.SessionBinding.BrowserProfileCount != 0 || payload.SessionBinding.ActiveBrowserProfile != "" || payload.SessionBinding.SelectedBrowserProfile != "" {
		t.Fatalf("expected inspect alias to hide implicit host session binding profile summary, got %#v", payload.SessionBinding)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserInspectAliasSurfacesHiddenManagedRouteSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-inspect-hidden-managed-route")
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-inspect-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-inspect-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: browserHiddenManagedRouteDoctorGateRuntimeControlNodeBackend(BrowserCapabilities{
			Open:           true,
			RuntimeStatus:  true,
			RuntimePrepare: true,
		}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"inspect"}`,
	})
	if err != nil {
		t.Fatalf("browser inspect hidden managed route: %v", err)
	}
	var payload struct {
		Action            string                         `json:"action"`
		Status            string                         `json:"status"`
		Note              string                         `json:"note"`
		DefaultRoute      browserRuntimeRouteDescriptor  `json:"default_route"`
		DefaultCandidate  browserRuntimeRouteDescriptor  `json:"default_candidate_route"`
		SelectedRoute     any                            `json:"selected_route"`
		RouteResolution   any                            `json:"route_resolution"`
		DefaultProfile    string                         `json:"default_profile"`
		ProfileStatus     any                            `json:"profile_status"`
		BrowserSurface    string                         `json:"browser_surface"`
		BrowserOptInHints []string                       `json:"browser_opt_in_targets"`
		Diagnostics       *browserTopLevelSummary        `json:"diagnostics"`
		Summary           *browserTopLevelSummary        `json:"summary"`
		Display           *browserTopLevelDisplaySummary `json:"display"`
		Surface           *browserTopLevelSurfaceSummary `json:"surface"`
		View              *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode inspect hidden managed route output: %v", err)
	}
	if payload.Action != "status" || payload.Status != "ok" {
		t.Fatalf("unexpected inspect hidden managed route payload: %#v", payload)
	}
	if !strings.Contains(payload.Note, "not the default") {
		t.Fatalf("expected inspect hidden managed route note to preserve doctor guidance, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected inspect alias to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected inspect alias to keep hidden implicit-host route state scrubbed, got %#v", payload)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil {
		t.Fatalf("expected inspect alias to keep hidden implicit-host profile summary scrubbed, got %#v", payload)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Summary)
	if payload.Summary == nil || payload.Summary.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected unified status summary to expose managed default_candidate_route, got %#v", payload.Summary)
	}
	assertUnifiedManagedRouteDoctorAliasDisplaySummary(t, payload.Display)
	assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t, payload.Surface)
	assertUnifiedManagedRouteDoctorAliasViewSummary(t, payload.View)
	if payload.BrowserSurface != "explicit_managed_opt_in" || len(payload.BrowserOptInHints) != 1 || payload.BrowserOptInHints[0] != "node" {
		t.Fatalf("expected inspect alias to surface managed opt-in root metadata, got surface=%q targets=%#v", payload.BrowserSurface, payload.BrowserOptInHints)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserStatusActionSurfacesHiddenManagedRouteSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-status-hidden-managed-route")
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-status-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-status-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: browserHiddenManagedRouteDoctorGateRuntimeControlNodeBackend(BrowserCapabilities{
			RuntimeStatus:  true,
			RuntimePrepare: true,
			Open:           true,
		}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser status hidden managed route: %v", err)
	}
	var payload struct {
		Action            string                         `json:"action"`
		Status            string                         `json:"status"`
		Note              string                         `json:"note"`
		DefaultRoute      browserRuntimeRouteDescriptor  `json:"default_route"`
		DefaultCandidate  browserRuntimeRouteDescriptor  `json:"default_candidate_route"`
		SelectedRoute     any                            `json:"selected_route"`
		RouteResolution   any                            `json:"route_resolution"`
		DefaultProfile    string                         `json:"default_profile"`
		ProfileStatus     any                            `json:"profile_status"`
		BrowserSurface    string                         `json:"browser_surface"`
		BrowserOptInHints []string                       `json:"browser_opt_in_targets"`
		Diagnostics       *browserTopLevelSummary        `json:"diagnostics"`
		Summary           *browserTopLevelSummary        `json:"summary"`
		Display           *browserTopLevelDisplaySummary `json:"display"`
		Surface           *browserTopLevelSurfaceSummary `json:"surface"`
		View              *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode status hidden managed route output: %v", err)
	}
	if payload.Action != "status" || payload.Status != "ok" {
		t.Fatalf("unexpected status hidden managed route payload: %#v", payload)
	}
	if !strings.Contains(payload.Note, "not the default") {
		t.Fatalf("expected status hidden managed route note to preserve doctor guidance, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected status action to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.DefaultCandidate.Backend != "proxy" || payload.DefaultCandidate.Profile != "isolated" || payload.DefaultCandidate.RuntimeTarget != "node" {
		t.Fatalf("expected status action to expose managed default_candidate_route, got %#v", payload.DefaultCandidate)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected status action to keep hidden implicit-host route state scrubbed, got %#v", payload)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil {
		t.Fatalf("expected status action to keep hidden implicit-host profile summary scrubbed, got %#v", payload)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Summary)
	if payload.Summary == nil || payload.Summary.DefaultCandidateRoute != payload.DefaultCandidate {
		t.Fatalf("expected status summary to expose managed default_candidate_route, got %#v", payload.Summary)
	}
	assertUnifiedManagedRouteDoctorAliasDisplaySummary(t, payload.Display)
	assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t, payload.Surface)
	assertUnifiedManagedRouteDoctorAliasViewSummary(t, payload.View)
	if payload.BrowserSurface != "explicit_managed_opt_in" || len(payload.BrowserOptInHints) != 1 || payload.BrowserOptInHints[0] != "node" {
		t.Fatalf("expected status action to surface managed opt-in root metadata, got surface=%q targets=%#v", payload.BrowserSurface, payload.BrowserOptInHints)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserStatusActionRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-status-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
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
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-status-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser status managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser status to route through managed current route before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser status managed current output: %v", err)
	}
	if payload.Action != "status" || payload.Status != "ok" {
		t.Fatalf("unexpected browser status managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser status to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser status to select managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected browser status managed current profile payload: %#v", payload.ProfileStatus)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserDoctorAliasSurfacesHiddenManagedRouteSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-doctor-hidden-managed-route")
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-doctor-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-doctor-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: browserHiddenManagedRouteDoctorGateResolvableRuntimeControlNodeBackend(BrowserCapabilities{
			RuntimeStatus:  true,
			RuntimePrepare: true,
			Open:           true,
		}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"doctor"}`,
	})
	if err != nil {
		t.Fatalf("browser doctor hidden managed route: %v", err)
	}
	var payload struct {
		Action            string                         `json:"action"`
		Status            string                         `json:"status"`
		Note              string                         `json:"note"`
		DefaultRoute      browserRuntimeRouteDescriptor  `json:"default_route"`
		SelectedRoute     any                            `json:"selected_route"`
		RouteResolution   any                            `json:"route_resolution"`
		DefaultProfile    string                         `json:"default_profile"`
		ProfileStatus     any                            `json:"profile_status"`
		BrowserSurface    string                         `json:"browser_surface"`
		BrowserOptInHints []string                       `json:"browser_opt_in_targets"`
		Doctor            *BrowserDoctorSummary          `json:"doctor"`
		Diagnostics       *browserTopLevelSummary        `json:"diagnostics"`
		Summary           *browserTopLevelSummary        `json:"summary"`
		Display           *browserTopLevelDisplaySummary `json:"display"`
		Surface           *browserTopLevelSurfaceSummary `json:"surface"`
		View              *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode doctor hidden managed route output: %v", err)
	}
	if payload.Action != "status" || payload.Status != "ok" {
		t.Fatalf("unexpected doctor hidden managed route payload: %#v", payload)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil || payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" {
		t.Fatalf("expected doctor alias to keep doctor route summary, got %#v", payload.Doctor)
	}
	if !strings.Contains(payload.Note, "not the default") {
		t.Fatalf("expected doctor hidden managed route note to preserve doctor guidance, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected doctor alias to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected doctor alias to keep hidden implicit-host route state scrubbed, got %#v", payload)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil {
		t.Fatalf("expected doctor alias to keep hidden implicit-host profile summary scrubbed, got %#v", payload)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Summary)
	assertUnifiedManagedRouteDoctorAliasDisplaySummary(t, payload.Display)
	assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t, payload.Surface)
	assertUnifiedManagedRouteDoctorAliasViewSummary(t, payload.View)
	if payload.BrowserSurface != "explicit_managed_opt_in" || len(payload.BrowserOptInHints) != 1 || payload.BrowserOptInHints[0] != "node" {
		t.Fatalf("expected doctor alias to surface managed opt-in root metadata, got surface=%q targets=%#v", payload.BrowserSurface, payload.BrowserOptInHints)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserReadySurfacesHiddenManagedRouteSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-ready-hidden-managed-route")
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-ready-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-ready-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: browserHiddenManagedRouteDoctorGateResolvableRuntimeControlNodeBackend(BrowserCapabilities{
			RuntimeStatus:  true,
			RuntimePrepare: true,
			Open:           true,
		}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"ready"}`,
	})
	if err != nil {
		t.Fatalf("browser ready hidden managed route: %v", err)
	}
	var payload struct {
		Action            string                         `json:"action"`
		Status            string                         `json:"status"`
		Note              string                         `json:"note"`
		DefaultRoute      browserRuntimeRouteDescriptor  `json:"default_route"`
		SelectedRoute     any                            `json:"selected_route"`
		RouteResolution   any                            `json:"route_resolution"`
		DefaultProfile    string                         `json:"default_profile"`
		ProfileStatus     any                            `json:"profile_status"`
		BrowserSurface    string                         `json:"browser_surface"`
		BrowserOptInHints []string                       `json:"browser_opt_in_targets"`
		Doctor            *BrowserDoctorSummary          `json:"doctor"`
		Diagnostics       *browserTopLevelSummary        `json:"diagnostics"`
		Summary           *browserTopLevelSummary        `json:"summary"`
		Display           *browserTopLevelDisplaySummary `json:"display"`
		Surface           *browserTopLevelSurfaceSummary `json:"surface"`
		View              *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode ready hidden managed route output: %v", err)
	}
	if payload.Action != "prepare" || payload.Status != "unsupported" {
		t.Fatalf("unexpected ready hidden managed route payload: %#v", payload)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil || payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" {
		t.Fatalf("expected ready alias to keep doctor route summary, got %#v", payload.Doctor)
	}
	if !strings.Contains(payload.Note, "not the default") {
		t.Fatalf("expected ready hidden managed route note to preserve doctor guidance, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected ready alias to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected ready alias to keep hidden implicit-host route state scrubbed, got %#v", payload)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil {
		t.Fatalf("expected ready alias to keep hidden implicit-host profile summary scrubbed, got %#v", payload)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Summary)
	assertUnifiedManagedRouteDoctorAliasDisplaySummary(t, payload.Display)
	assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t, payload.Surface)
	assertUnifiedManagedRouteDoctorAliasViewSummary(t, payload.View)
	if payload.BrowserSurface != "explicit_managed_opt_in" || len(payload.BrowserOptInHints) != 1 || payload.BrowserOptInHints[0] != "node" {
		t.Fatalf("expected ready alias to surface managed opt-in root metadata, got surface=%q targets=%#v", payload.BrowserSurface, payload.BrowserOptInHints)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserEnsureSurfacesHiddenManagedRouteSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-ensure-hidden-managed-route")
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-ensure-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-ensure-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: browserHiddenManagedRouteDoctorGateResolvableRuntimeControlNodeBackend(BrowserCapabilities{
			RuntimeCoordinate: true,
			Open:              true,
		}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"ensure"}`,
	})
	if err != nil {
		t.Fatalf("browser ensure hidden managed route: %v", err)
	}
	var payload struct {
		Action            string                         `json:"action"`
		Status            string                         `json:"status"`
		Note              string                         `json:"note"`
		DefaultRoute      browserRuntimeRouteDescriptor  `json:"default_route"`
		SelectedRoute     any                            `json:"selected_route"`
		RouteResolution   any                            `json:"route_resolution"`
		DefaultProfile    string                         `json:"default_profile"`
		ProfileStatus     any                            `json:"profile_status"`
		BrowserSurface    string                         `json:"browser_surface"`
		BrowserOptInHints []string                       `json:"browser_opt_in_targets"`
		Doctor            *BrowserDoctorSummary          `json:"doctor"`
		Diagnostics       *browserTopLevelSummary        `json:"diagnostics"`
		Summary           *browserTopLevelSummary        `json:"summary"`
		Display           *browserTopLevelDisplaySummary `json:"display"`
		Surface           *browserTopLevelSurfaceSummary `json:"surface"`
		View              *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode ensure hidden managed route output: %v", err)
	}
	if payload.Action != "coordinate" || payload.Status != "unsupported" {
		t.Fatalf("unexpected ensure hidden managed route payload: %#v", payload)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil || payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" {
		t.Fatalf("expected ensure alias to keep doctor route summary, got %#v", payload.Doctor)
	}
	if !strings.Contains(payload.Note, "not the default") {
		t.Fatalf("expected ensure hidden managed route note to preserve doctor guidance, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected ensure alias to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected ensure alias to keep hidden implicit-host route state scrubbed, got %#v", payload)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil {
		t.Fatalf("expected ensure alias to keep hidden implicit-host profile summary scrubbed, got %#v", payload)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Summary)
	assertUnifiedManagedRouteDoctorAliasDisplaySummary(t, payload.Display)
	assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t, payload.Surface)
	assertUnifiedManagedRouteDoctorAliasViewSummary(t, payload.View)
	if payload.BrowserSurface != "explicit_managed_opt_in" || len(payload.BrowserOptInHints) != 1 || payload.BrowserOptInHints[0] != "node" {
		t.Fatalf("expected ensure alias to surface managed opt-in root metadata, got surface=%q targets=%#v", payload.BrowserSurface, payload.BrowserOptInHints)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserLaunchSurfacesHiddenManagedRouteSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-launch-hidden-managed-route")
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-launch-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-launch-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root: t.TempDir(),
		NodeBackend: browserHiddenManagedRouteDoctorGateResolvableRuntimeControlNodeBackend(BrowserCapabilities{
			RuntimeStart: true,
			Open:         true,
		}),
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"launch"}`,
	})
	if err != nil {
		t.Fatalf("browser launch hidden managed route: %v", err)
	}
	var payload struct {
		Action            string                         `json:"action"`
		Status            string                         `json:"status"`
		Note              string                         `json:"note"`
		DefaultRoute      browserRuntimeRouteDescriptor  `json:"default_route"`
		SelectedRoute     any                            `json:"selected_route"`
		RouteResolution   any                            `json:"route_resolution"`
		DefaultProfile    string                         `json:"default_profile"`
		ProfileStatus     any                            `json:"profile_status"`
		BrowserSurface    string                         `json:"browser_surface"`
		BrowserOptInHints []string                       `json:"browser_opt_in_targets"`
		Doctor            *BrowserDoctorSummary          `json:"doctor"`
		Diagnostics       *browserTopLevelSummary        `json:"diagnostics"`
		Summary           *browserTopLevelSummary        `json:"summary"`
		Display           *browserTopLevelDisplaySummary `json:"display"`
		Surface           *browserTopLevelSurfaceSummary `json:"surface"`
		View              *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode launch hidden managed route output: %v", err)
	}
	if payload.Action != "start" || payload.Status != "unsupported" {
		t.Fatalf("unexpected launch hidden managed route payload: %#v", payload)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil || payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" {
		t.Fatalf("expected launch alias to keep doctor route summary, got %#v", payload.Doctor)
	}
	if !strings.Contains(payload.Note, "not the default") {
		t.Fatalf("expected launch hidden managed route note to preserve doctor guidance, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected launch alias to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected launch alias to keep hidden implicit-host route state scrubbed, got %#v", payload)
	}
	if payload.DefaultProfile != "" || payload.ProfileStatus != nil {
		t.Fatalf("expected launch alias to keep hidden implicit-host profile summary scrubbed, got %#v", payload)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Summary)
	assertUnifiedManagedRouteDoctorAliasDisplaySummary(t, payload.Display)
	assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t, payload.Surface)
	assertUnifiedManagedRouteDoctorAliasViewSummary(t, payload.View)
	if payload.BrowserSurface != "explicit_managed_opt_in" || len(payload.BrowserOptInHints) != 1 || payload.BrowserOptInHints[0] != "node" {
		t.Fatalf("expected launch alias to surface managed opt-in root metadata, got surface=%q targets=%#v", payload.BrowserSurface, payload.BrowserOptInHints)
	}
}

func assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t *testing.T, summary *browserTopLevelSummary) {
	t.Helper()
	if summary == nil ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.NextStepAlias != "ready" ||
		summary.ManualRetryHint != "promote_managed_default" ||
		!summary.ResolvedViaFallback ||
		summary.PrimaryBrowserAction != "browser action=ready" ||
		summary.PrimaryNodeAction != "" ||
		summary.NextStep != "browser action=ready" {
		t.Fatalf("unexpected unified doctor-route top-level summary: %#v", summary)
	}
}

func assertUnifiedManagedRouteDoctorAliasDisplaySummary(t *testing.T, summary *browserTopLevelDisplaySummary) {
	t.Helper()
	if summary == nil ||
		summary.Ready ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		summary.NextStepAlias != "ready" ||
		summary.ManualRetryHint != "promote_managed_default" ||
		!summary.ResolvedViaFallback ||
		summary.PrimaryBrowserAction != "browser action=ready" ||
		summary.PrimaryNodeAction != "" ||
		summary.NextStep != "browser action=ready" {
		t.Fatalf("unexpected unified doctor-route display summary: %#v", summary)
	}
}

func assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t *testing.T, summary *browserTopLevelSurfaceSummary) {
	t.Helper()
	if summary == nil ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		summary.NextStepAlias != "ready" ||
		summary.ManualRetryHint != "promote_managed_default" ||
		!summary.ResolvedViaFallback ||
		summary.PrimaryBrowserAction != "browser action=ready" ||
		summary.PrimaryNodeAction != "" ||
		summary.NextStep != "browser action=ready" ||
		summary.BrowserSurface != "explicit_managed_opt_in" ||
		len(summary.BrowserOptInTargets) != 1 ||
		summary.BrowserOptInTargets[0] != "node" {
		t.Fatalf("unexpected unified doctor-route surface summary: %#v", summary)
	}
}

func assertUnifiedManagedRouteDoctorAliasViewSummary(t *testing.T, summary *browserTopLevelViewSummary) {
	t.Helper()
	if summary == nil ||
		summary.Category != "coordination" ||
		summary.State != "managed_route_pending_default" ||
		summary.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		summary.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		summary.NextStepAlias != "ready" ||
		summary.ManualRetryHint != "promote_managed_default" ||
		!summary.ResolvedViaFallback ||
		summary.PrimaryBrowserAction != "browser action=ready" ||
		summary.PrimaryNodeAction != "" ||
		summary.NextStep != "browser action=ready" ||
		summary.BrowserSurface != "explicit_managed_opt_in" ||
		len(summary.BrowserOptInTargets) != 1 ||
		summary.BrowserOptInTargets[0] != "node" {
		t.Fatalf("unexpected unified doctor-route view summary: %#v", summary)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserInventoryAliasSurfacesHiddenManagedRouteSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-inventory-hidden-managed-route")
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-inventory-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-inventory-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	nodeBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
				runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
					fakeBrowserBackend: &fakeBrowserBackend{},
					runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
				},
			},
			capabilities: BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"inventory"}`,
	})
	if err != nil {
		t.Fatalf("browser inventory hidden managed route: %v", err)
	}
	var payload struct {
		Action            string                         `json:"action"`
		Status            string                         `json:"status"`
		Note              string                         `json:"note"`
		DefaultRoute      browserRuntimeRouteDescriptor  `json:"default_route"`
		SelectedRoute     any                            `json:"selected_route"`
		RouteResolution   any                            `json:"route_resolution"`
		DefaultProfile    string                         `json:"default_profile"`
		Profiles          []any                          `json:"profiles"`
		BrowserSurface    string                         `json:"browser_surface"`
		BrowserOptInHints []string                       `json:"browser_opt_in_targets"`
		Diagnostics       *browserTopLevelSummary        `json:"diagnostics"`
		Summary           *browserTopLevelSummary        `json:"summary"`
		Display           *browserTopLevelDisplaySummary `json:"display"`
		Surface           *browserTopLevelSurfaceSummary `json:"surface"`
		View              *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode inventory hidden managed route output: %v", err)
	}
	if payload.Action != "profiles" || payload.Status != "unsupported" {
		t.Fatalf("unexpected inventory hidden managed route payload: %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected inventory hidden managed route note to preserve specific route failure, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected inventory alias to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected inventory alias to keep hidden implicit-host route state scrubbed, got %#v", payload)
	}
	if payload.DefaultProfile != "" || len(payload.Profiles) != 0 {
		t.Fatalf("expected inventory alias to keep hidden implicit-host profile summary scrubbed, got %#v", payload)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Summary)
	assertUnifiedManagedRouteDoctorAliasDisplaySummary(t, payload.Display)
	assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t, payload.Surface)
	assertUnifiedManagedRouteDoctorAliasViewSummary(t, payload.View)
	if payload.BrowserSurface != "explicit_managed_opt_in" || len(payload.BrowserOptInHints) != 1 || payload.BrowserOptInHints[0] != "node" {
		t.Fatalf("expected inventory alias to surface managed opt-in root metadata, got surface=%q targets=%#v", payload.BrowserSurface, payload.BrowserOptInHints)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserHandlesAliasSurfacesHiddenManagedRouteSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-handles-hidden-managed-route")
	tracked := sessionRegistry.TrackTab("browser-unified-handles-hidden-managed-route", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://93.184.216.34/hidden-host",
		Title:      "Hidden Host",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-handles-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-handles-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	nodeBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
				runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
					fakeBrowserBackend: &fakeBrowserBackend{},
					runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
				},
			},
			capabilities: BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
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
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"handles"}`,
	})
	if err != nil {
		t.Fatalf("browser handles hidden managed route: %v", err)
	}
	var payload struct {
		Action               string                         `json:"action"`
		Status               string                         `json:"status"`
		Note                 string                         `json:"note"`
		DefaultRoute         browserRuntimeRouteDescriptor  `json:"default_route"`
		SelectedRoute        any                            `json:"selected_route"`
		RouteResolution      any                            `json:"route_resolution"`
		SessionID            string                         `json:"session_id"`
		SessionTargetCount   int                            `json:"session_target_count"`
		BrowserSurface       string                         `json:"browser_surface"`
		BrowserOptInTargets  []string                       `json:"browser_opt_in_targets"`
		Diagnostics          *browserTopLevelSummary        `json:"diagnostics"`
		Summary              *browserTopLevelSummary        `json:"summary"`
		Display              *browserTopLevelDisplaySummary `json:"display"`
		Surface              *browserTopLevelSurfaceSummary `json:"surface"`
		View                 *browserTopLevelViewSummary    `json:"view"`
		SessionProfileSelect string                         `json:"session_profile_selection"`
		SessionTargetSelect  any                            `json:"session_target_selection"`
		SessionBinding       struct {
			SelectedBrowserProfile string `json:"selected_browser_profile"`
			CurrentTargetID        string `json:"current_target_id"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode handles hidden managed route output: %v", err)
	}
	if payload.Action != "sessions" || payload.Status != "ok" {
		t.Fatalf("unexpected handles hidden managed route payload: %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected handles hidden managed route note to preserve specific route failure, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected handles alias to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected handles alias to keep hidden implicit-host route state scrubbed, got %#v", payload)
	}
	if payload.SessionID != "browser-unified-handles-hidden-managed-route" || payload.SessionTargetCount != 1 {
		t.Fatalf("expected handles alias to preserve session routing payload, got %#v", payload)
	}
	if payload.SessionProfileSelect != "" || payload.SessionTargetSelect != nil {
		t.Fatalf("expected handles alias to keep hidden implicit-host top-level selections scrubbed, got %#v", payload)
	}
	if payload.SessionBinding.SelectedBrowserProfile != "" || payload.SessionBinding.CurrentTargetID != "" {
		t.Fatalf("expected handles alias to keep hidden implicit-host binding selections scrubbed, got %#v", payload.SessionBinding)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Summary)
	assertUnifiedManagedRouteDoctorAliasDisplaySummary(t, payload.Display)
	assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t, payload.Surface)
	assertUnifiedManagedRouteDoctorAliasViewSummary(t, payload.View)
	if payload.BrowserSurface != "explicit_managed_opt_in" || len(payload.BrowserOptInTargets) != 1 || payload.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected handles alias to surface managed opt-in root metadata, got surface=%q targets=%#v", payload.BrowserSurface, payload.BrowserOptInTargets)
	}
	if tracked.ID == "" {
		t.Fatalf("expected tracked host target to remain available for hidden managed handles test")
	}
}

func TestRegisterBrowserTools_UnifiedBrowserProfilesActionSurfacesHiddenManagedRouteSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-profiles-hidden-managed-route")
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-profiles-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-profiles-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	nodeBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
				runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
					fakeBrowserBackend: &fakeBrowserBackend{},
					runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
				},
			},
			capabilities: BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
		},
		resolve: func(requested BrowserRuntimeInfo) (BrowserRuntimeInfo, error) {
			return BrowserRuntimeInfo{}, context.DeadlineExceeded
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"profiles"}`,
	})
	if err != nil {
		t.Fatalf("browser profiles hidden managed route: %v", err)
	}
	var payload struct {
		Action            string                         `json:"action"`
		Status            string                         `json:"status"`
		Note              string                         `json:"note"`
		DefaultRoute      browserRuntimeRouteDescriptor  `json:"default_route"`
		SelectedRoute     any                            `json:"selected_route"`
		RouteResolution   any                            `json:"route_resolution"`
		DefaultProfile    string                         `json:"default_profile"`
		Profiles          []any                          `json:"profiles"`
		BrowserSurface    string                         `json:"browser_surface"`
		BrowserOptInHints []string                       `json:"browser_opt_in_targets"`
		Diagnostics       *browserTopLevelSummary        `json:"diagnostics"`
		Summary           *browserTopLevelSummary        `json:"summary"`
		Display           *browserTopLevelDisplaySummary `json:"display"`
		Surface           *browserTopLevelSurfaceSummary `json:"surface"`
		View              *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode profiles hidden managed route output: %v", err)
	}
	if payload.Action != "profiles" || payload.Status != "unsupported" {
		t.Fatalf("unexpected profiles hidden managed route payload: %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected profiles hidden managed route note to preserve specific route failure, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected profiles action to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected profiles action to keep hidden implicit-host route state scrubbed, got %#v", payload)
	}
	if payload.DefaultProfile != "" || len(payload.Profiles) != 0 {
		t.Fatalf("expected profiles action to keep hidden implicit-host profile summary scrubbed, got %#v", payload)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Summary)
	assertUnifiedManagedRouteDoctorAliasDisplaySummary(t, payload.Display)
	assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t, payload.Surface)
	assertUnifiedManagedRouteDoctorAliasViewSummary(t, payload.View)
	if payload.BrowserSurface != "explicit_managed_opt_in" || len(payload.BrowserOptInHints) != 1 || payload.BrowserOptInHints[0] != "node" {
		t.Fatalf("expected profiles action to surface managed opt-in root metadata, got surface=%q targets=%#v", payload.BrowserSurface, payload.BrowserOptInHints)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserProfilesActionRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-profiles-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "workbench",
				Profiles: []BrowserProfileInfo{
					{Profile: "workbench", BrowserApp: "Chromium", Status: "running", Running: true, Connected: true},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	sessionRegistry.TrackCurrentTarget("browser-unified-profiles-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"profiles"}`,
	})
	if err != nil {
		t.Fatalf("browser profiles managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 1 || nodeBackend.runtimeProfilesReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser profiles to route through managed current route before implicit host fallback, got %#v", nodeBackend.runtimeProfilesReqs)
	}
	var payload struct {
		Action              string   `json:"action"`
		Status              string   `json:"status"`
		RequestedBrowserApp string   `json:"requested_browser_app"`
		DefaultProfile      string   `json:"default_profile"`
		ConfiguredProfiles  []string `json:"configured_profiles"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		Profiles []struct {
			Profile    string `json:"profile"`
			BrowserApp string `json:"browser_app"`
			Status     string `json:"status"`
			Running    bool   `json:"running"`
			Connected  bool   `json:"connected"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser profiles managed current output: %v", err)
	}
	if payload.Action != "profiles" || payload.Status != "ok" {
		t.Fatalf("unexpected browser profiles managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser profiles to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.DefaultProfile != "workbench" {
		t.Fatalf("expected browser profiles to expose managed current default profile, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser profiles to select managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if !browserStringSliceContains(payload.ConfiguredProfiles, "workbench") {
		t.Fatalf("expected browser profiles configured profiles to retain managed current route profile, got %#v", payload.ConfiguredProfiles)
	}
	if len(payload.Profiles) != 1 || payload.Profiles[0].Profile != "workbench" || payload.Profiles[0].BrowserApp != "Chromium" || payload.Profiles[0].Status != "running" || !payload.Profiles[0].Running || !payload.Profiles[0].Connected {
		t.Fatalf("unexpected browser profiles managed current payload: %#v", payload.Profiles)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserSessionsActionSurfacesHiddenManagedRouteSummary(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-sessions-hidden-managed-route")
	tracked := sessionRegistry.TrackTab("browser-unified-sessions-hidden-managed-route", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://93.184.216.34/hidden-host",
		Title:      "Hidden Host",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-sessions-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-sessions-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	nodeBackend := &routeResolverCapabilityRuntimeControlBrowserBackend{
		capabilityRuntimeControlBrowserBackend: &capabilityRuntimeControlBrowserBackend{
			runtimeControlBrowserBackend: &runtimeControlBrowserBackend{
				runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
					fakeBrowserBackend: &fakeBrowserBackend{},
					runtimeInfo:        BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
				},
			},
			capabilities: BrowserCapabilitiesForActKinds([]string{"open", "navigate", "list_tabs", "extract", "snapshot", "screenshot", "click", "type", "evaluate"}),
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
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser sessions hidden managed route: %v", err)
	}
	var payload struct {
		Action               string                         `json:"action"`
		Status               string                         `json:"status"`
		Note                 string                         `json:"note"`
		DefaultRoute         browserRuntimeRouteDescriptor  `json:"default_route"`
		SelectedRoute        any                            `json:"selected_route"`
		RouteResolution      any                            `json:"route_resolution"`
		SessionID            string                         `json:"session_id"`
		SessionTargetCount   int                            `json:"session_target_count"`
		BrowserSurface       string                         `json:"browser_surface"`
		BrowserOptInTargets  []string                       `json:"browser_opt_in_targets"`
		Diagnostics          *browserTopLevelSummary        `json:"diagnostics"`
		Summary              *browserTopLevelSummary        `json:"summary"`
		Display              *browserTopLevelDisplaySummary `json:"display"`
		Surface              *browserTopLevelSurfaceSummary `json:"surface"`
		View                 *browserTopLevelViewSummary    `json:"view"`
		SessionProfileSelect string                         `json:"session_profile_selection"`
		SessionTargetSelect  any                            `json:"session_target_selection"`
		SessionBinding       struct {
			SelectedBrowserProfile string `json:"selected_browser_profile"`
			CurrentTargetID        string `json:"current_target_id"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode sessions hidden managed route output: %v", err)
	}
	if payload.Action != "sessions" || payload.Status != "ok" {
		t.Fatalf("unexpected sessions hidden managed route payload: %#v", payload)
	}
	if !strings.Contains(payload.Note, "context deadline exceeded") {
		t.Fatalf("expected sessions hidden managed route note to preserve specific route failure, got %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected sessions action to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected sessions action to keep hidden implicit-host route state scrubbed, got %#v", payload)
	}
	if payload.SessionID != "browser-unified-sessions-hidden-managed-route" || payload.SessionTargetCount != 1 {
		t.Fatalf("expected sessions action to preserve session routing payload, got %#v", payload)
	}
	if payload.SessionProfileSelect != "" || payload.SessionTargetSelect != nil {
		t.Fatalf("expected sessions action to keep hidden implicit-host top-level selections scrubbed, got %#v", payload)
	}
	if payload.SessionBinding.SelectedBrowserProfile != "" || payload.SessionBinding.CurrentTargetID != "" {
		t.Fatalf("expected sessions action to keep hidden implicit-host binding selections scrubbed, got %#v", payload.SessionBinding)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Summary)
	assertUnifiedManagedRouteDoctorAliasDisplaySummary(t, payload.Display)
	assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t, payload.Surface)
	assertUnifiedManagedRouteDoctorAliasViewSummary(t, payload.View)
	if payload.BrowserSurface != "explicit_managed_opt_in" || len(payload.BrowserOptInTargets) != 1 || payload.BrowserOptInTargets[0] != "node" {
		t.Fatalf("expected sessions action to surface managed opt-in root metadata, got surface=%q targets=%#v", payload.BrowserSurface, payload.BrowserOptInTargets)
	}
	if tracked.ID == "" {
		t.Fatalf("expected tracked host target to remain available for hidden managed sessions test")
	}
}

func TestRegisterBrowserTools_UnifiedBrowserSessionsActionRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-sessions-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
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
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-sessions-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser sessions managed current hidden implicit host: %v", err)
	}
	var payload struct {
		Action              string `json:"action"`
		Status              string `json:"status"`
		RequestedBrowserApp string `json:"requested_browser_app"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionTargetCount int `json:"session_target_count"`
		SessionBinding     struct {
			CurrentTargetID string `json:"current_target_id"`
		} `json:"session_binding"`
		SessionRoutes []struct {
			Backend         string `json:"backend"`
			Profile         string `json:"profile"`
			RuntimeTarget   string `json:"runtime_target"`
			CurrentTargetID string `json:"current_target_id"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser sessions managed current output: %v", err)
	}
	if payload.Action != "sessions" || payload.Status != "ok" {
		t.Fatalf("unexpected browser sessions managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser sessions to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser sessions to select managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	if payload.SessionTargetCount != 1 || payload.SessionBinding.CurrentTargetID != tracked.ID {
		t.Fatalf("unexpected browser sessions managed current binding payload: %#v", payload)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].Backend != "proxy" || payload.SessionRoutes[0].Profile != "workbench" || payload.SessionRoutes[0].RuntimeTarget != "node" || payload.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("unexpected browser sessions managed current routes payload: %#v", payload.SessionRoutes)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserWorkbenchUsesDoctorRouteSummaryWhenManagedRouteStaysHidden(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-workbench-hidden-managed-route")
	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-workbench-hidden-managed-route", BrowserSessionTarget{
		ID:         "host-current",
		TabIndex:   1,
		URL:        "https://host.example/current",
		Title:      "Host Current",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, "tracked_active_tab")
	sessionStateRegistry.RecordBrowserProfileState("browser-unified-workbench-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Status:        "running",
		Running:       true,
		Connected:     true,
		Note:          "cached implicit host default profile",
	})
	sessionStateRegistry.SelectBrowserProfile("browser-unified-workbench-hidden-managed-route", agentxbrowserruntime.SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
		Source:        "remember_profile",
	})
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          browserHiddenManagedRouteDoctorGateNodeBackend(),
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser",
		Arguments: `{"action":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser workbench hidden managed route: %v", err)
	}
	var payload struct {
		Action                             string                                       `json:"action"`
		Status                             string                                       `json:"status"`
		Note                               string                                       `json:"note"`
		DefaultRoute                       browserRuntimeRouteDescriptor                `json:"default_route"`
		SelectedRoute                      any                                          `json:"selected_route"`
		RouteResolution                    any                                          `json:"route_resolution"`
		Doctor                             *BrowserDoctorSummary                        `json:"doctor"`
		WorkbenchReady                     bool                                         `json:"workbench_ready"`
		WorkbenchSections                  []string                                     `json:"workbench_sections"`
		WorkbenchPrimaryBrowserAction      string                                       `json:"workbench_primary_browser_action"`
		WorkbenchPrimaryNodeAction         string                                       `json:"workbench_primary_node_action"`
		WorkbenchNextStep                  string                                       `json:"workbench_next_step"`
		WorkbenchRecommendedBrowserActions []string                                     `json:"workbench_recommended_browser_actions"`
		WorkbenchExplanation               *browserRuntimeDiagnosticsExplanationSummary `json:"workbench_explanation"`
		WorkbenchDiagnostics               *browserRuntimeWorkbenchDiagnosticsSummary   `json:"workbench_diagnostics"`
		WorkbenchSummary                   *browserTopLevelSummary                      `json:"workbench_summary"`
		WorkbenchDisplay                   *browserRuntimeWorkbenchDisplaySummary       `json:"workbench_display"`
		Workbench                          *browserRuntimeWorkbenchSurfaceSummary       `json:"workbench"`
		Diagnostics                        *browserTopLevelSummary                      `json:"diagnostics"`
		Summary                            *browserTopLevelSummary                      `json:"summary"`
		Display                            *browserTopLevelDisplaySummary               `json:"display"`
		Surface                            *browserTopLevelSurfaceSummary               `json:"surface"`
		View                               *browserTopLevelViewSummary                  `json:"view"`
		BrowserSurface                     string                                       `json:"browser_surface"`
		BrowserOptInTargets                []string                                     `json:"browser_opt_in_targets"`
		SessionRoutes                      []browserRuntimeSessionRoute                 `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode workbench hidden managed route output: %v", err)
	}
	if payload.Action != "workbench" || payload.Status != "ok" {
		t.Fatalf("unexpected workbench hidden managed route payload: %#v", payload)
	}
	if payload.DefaultRoute.Backend != "proxy" || payload.DefaultRoute.Profile != "isolated" || payload.DefaultRoute.RuntimeTarget != "node" {
		t.Fatalf("expected workbench action to promote managed candidate default_route, got %#v", payload.DefaultRoute)
	}
	if payload.SelectedRoute != nil || payload.RouteResolution != nil {
		t.Fatalf("expected workbench action to keep hidden implicit-host route state scrubbed, got %#v", payload)
	}
	if payload.Doctor == nil || payload.Doctor.Route == nil || payload.Doctor.Route.Code != "managed_route_hidden_by_legacy_host_default" {
		t.Fatalf("expected workbench action to keep doctor route summary, got %#v", payload.Doctor)
	}
	if !strings.Contains(payload.Note, strings.TrimSpace(payload.Doctor.Route.Summary)) {
		t.Fatalf("expected workbench hidden managed route note to preserve doctor guidance, got %#v", payload)
	}
	if !payload.WorkbenchReady || !browserStringSliceContains(payload.WorkbenchSections, "route") {
		t.Fatalf("expected workbench action to keep route section visible, got %#v", payload)
	}
	if payload.WorkbenchPrimaryBrowserAction != "browser action=ready" ||
		payload.WorkbenchPrimaryNodeAction != "" ||
		payload.WorkbenchNextStep != "browser action=ready" {
		t.Fatalf("expected workbench action to promote unified ready guidance, got %#v", payload)
	}
	if !browserStringSliceContains(payload.WorkbenchRecommendedBrowserActions, "browser action=ready") {
		t.Fatalf("expected workbench action to recommend unified ready guidance, got %#v", payload.WorkbenchRecommendedBrowserActions)
	}
	if payload.WorkbenchExplanation == nil ||
		payload.WorkbenchExplanation.Category != "coordination" ||
		payload.WorkbenchExplanation.State != "managed_route_pending_default" ||
		payload.WorkbenchExplanation.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		payload.WorkbenchExplanation.NextStepAlias != "ready" ||
		payload.WorkbenchExplanation.ManualRetryHint != "promote_managed_default" {
		t.Fatalf("unexpected unified workbench explanation summary: %#v", payload.WorkbenchExplanation)
	}
	if payload.WorkbenchDiagnostics == nil ||
		payload.WorkbenchDiagnostics.Category != "coordination" ||
		payload.WorkbenchDiagnostics.State != "managed_route_pending_default" ||
		payload.WorkbenchDiagnostics.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		payload.WorkbenchDiagnostics.NextStepAlias != "ready" ||
		payload.WorkbenchDiagnostics.ManualRetryHint != "promote_managed_default" ||
		payload.WorkbenchDiagnostics.PrimaryBrowserAction != "browser action=ready" ||
		payload.WorkbenchDiagnostics.PrimaryNodeAction != "" ||
		payload.WorkbenchDiagnostics.NextStep != "browser action=ready" {
		t.Fatalf("unexpected unified workbench diagnostics summary: %#v", payload.WorkbenchDiagnostics)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.WorkbenchSummary)
	if payload.WorkbenchSummary == nil || payload.WorkbenchSummary.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) {
		t.Fatalf("expected unified workbench summary to expose managed default_candidate_route, got %#v", payload.WorkbenchSummary)
	}
	if payload.WorkbenchDisplay == nil ||
		!payload.WorkbenchDisplay.Ready ||
		!browserStringSliceContains(payload.WorkbenchDisplay.Sections, "route") ||
		payload.WorkbenchDisplay.Category != "coordination" ||
		payload.WorkbenchDisplay.State != "managed_route_pending_default" ||
		payload.WorkbenchDisplay.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		payload.WorkbenchDisplay.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		payload.WorkbenchDisplay.NextStepAlias != "ready" ||
		payload.WorkbenchDisplay.ManualRetryHint != "promote_managed_default" ||
		!payload.WorkbenchDisplay.ResolvedViaFallback ||
		payload.WorkbenchDisplay.PrimaryBrowserAction != "browser action=ready" ||
		payload.WorkbenchDisplay.PrimaryNodeAction != "" ||
		payload.WorkbenchDisplay.NextStep != "browser action=ready" {
		t.Fatalf("unexpected unified workbench display summary: %#v", payload.WorkbenchDisplay)
	}
	if payload.Workbench == nil ||
		!payload.Workbench.Ready ||
		!browserStringSliceContains(payload.Workbench.Sections, "route") ||
		payload.Workbench.Diagnostics == nil ||
		payload.Workbench.Summary == nil ||
		payload.Workbench.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		payload.Workbench.BrowserSurface != "explicit_managed_opt_in" ||
		!reflect.DeepEqual(payload.Workbench.BrowserOptInTargets, []string{"node"}) ||
		payload.Workbench.PrimaryBrowserAction != "browser action=ready" ||
		payload.Workbench.PrimaryNodeAction != "" ||
		payload.Workbench.NextStep != "browser action=ready" {
		t.Fatalf("unexpected unified workbench surface summary: %#v", payload.Workbench)
	}
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Workbench.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Workbench.Summary)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Diagnostics)
	assertUnifiedManagedRouteDoctorAliasTopLevelSummary(t, payload.Summary)
	if payload.Display == nil ||
		!payload.Display.Ready ||
		!browserStringSliceContains(payload.Display.Sections, "route") ||
		payload.Display.Category != "coordination" ||
		payload.Display.State != "managed_route_pending_default" ||
		payload.Display.SummaryCode != "managed_route_hidden_by_legacy_host_default" ||
		payload.Display.DefaultCandidateRoute != (browserRuntimeRouteDescriptor{Backend: "proxy", Profile: "isolated", RuntimeTarget: "node"}) ||
		payload.Display.NextStepAlias != "ready" ||
		payload.Display.ManualRetryHint != "promote_managed_default" ||
		!payload.Display.ResolvedViaFallback ||
		payload.Display.PrimaryBrowserAction != "browser action=ready" ||
		payload.Display.PrimaryNodeAction != "" ||
		payload.Display.NextStep != "browser action=ready" {
		t.Fatalf("unexpected unified doctor-route display summary: %#v", payload.Display)
	}
	assertUnifiedManagedRouteDoctorAliasSurfaceSummary(t, payload.Surface)
	assertUnifiedManagedRouteDoctorAliasViewSummary(t, payload.View)
	if payload.BrowserSurface != "explicit_managed_opt_in" || !reflect.DeepEqual(payload.BrowserOptInTargets, []string{"node"}) {
		t.Fatalf("expected workbench action to surface managed opt-in root metadata, got surface=%q targets=%#v", payload.BrowserSurface, payload.BrowserOptInTargets)
	}
	if len(payload.SessionRoutes) != 1 ||
		payload.SessionRoutes[0].Backend != "system" ||
		payload.SessionRoutes[0].Profile != "default" ||
		payload.SessionRoutes[0].RuntimeTarget != "host" ||
		payload.SessionRoutes[0].CurrentTargetID != tracked.ID {
		t.Fatalf("expected explicit host session route snapshot to remain visible, got %#v", payload.SessionRoutes)
	}
}

func TestRegisterBrowserTools_UnifiedBrowserWorkbenchActionRoutesToManagedCurrentRouteBeforeHiddenImplicitHostFallback(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-unified-workbench-managed-current-hidden-implicit-host")
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
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
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser"},
	})

	tracked := sessionRegistry.TrackCurrentTarget("browser-unified-workbench-managed-current-hidden-implicit-host", BrowserSessionTarget{
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
		Arguments: `{"action":"workbench"}`,
	})
	if err != nil {
		t.Fatalf("browser workbench managed current hidden implicit host: %v", err)
	}
	if len(nodeBackend.runtimeStatusReqs) != 1 || nodeBackend.runtimeStatusReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser workbench to route through managed current route status before implicit host fallback, got %#v", nodeBackend.runtimeStatusReqs)
	}
	if len(nodeBackend.runtimeProfilesReqs) != 1 || nodeBackend.runtimeProfilesReqs[0].Profile != "workbench" {
		t.Fatalf("expected browser workbench to route through managed current route profiles before implicit host fallback, got %#v", nodeBackend.runtimeProfilesReqs)
	}
	var payload struct {
		Action              string   `json:"action"`
		Status              string   `json:"status"`
		RequestedBrowserApp string   `json:"requested_browser_app"`
		WorkbenchReady      bool     `json:"workbench_ready"`
		WorkbenchSections   []string `json:"workbench_sections"`
		SelectedRoute       struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		ProfileStatus struct {
			Profile   string `json:"profile"`
			Status    string `json:"status"`
			Running   bool   `json:"running"`
			Connected bool   `json:"connected"`
		} `json:"profile_status"`
		SessionTargetCount int `json:"session_target_count"`
		SessionBinding     struct {
			CurrentTargetID string `json:"current_target_id"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser workbench managed current output: %v", err)
	}
	if payload.Action != "workbench" || payload.Status != "ok" || !payload.WorkbenchReady {
		t.Fatalf("unexpected browser workbench managed current payload: %#v", payload)
	}
	if payload.RequestedBrowserApp != "Chromium" {
		t.Fatalf("expected browser workbench to inherit managed current browser_app before implicit host fallback, got %#v", payload)
	}
	if payload.SelectedRoute.Backend != "proxy" || payload.SelectedRoute.Profile != "workbench" || payload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected browser workbench to select managed current route before implicit host fallback, got %#v", payload.SelectedRoute)
	}
	for _, want := range []string{"route", "status", "profiles", "sessions"} {
		if !browserStringSliceContains(payload.WorkbenchSections, want) {
			t.Fatalf("expected browser workbench to retain %q section on managed route, got %#v", want, payload.WorkbenchSections)
		}
	}
	if payload.ProfileStatus.Profile != "workbench" || payload.ProfileStatus.Status != "running" || !payload.ProfileStatus.Running || !payload.ProfileStatus.Connected {
		t.Fatalf("unexpected browser workbench managed current profile payload: %#v", payload.ProfileStatus)
	}
	if payload.SessionTargetCount != 1 || payload.SessionBinding.CurrentTargetID != tracked.ID {
		t.Fatalf("unexpected browser workbench managed current session binding payload: %#v", payload.SessionBinding)
	}
}
