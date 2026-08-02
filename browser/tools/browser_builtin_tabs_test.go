package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ActListTabs(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "list",
			Status:      "listed",
			ActiveIndex: 2,
			Tabs: []BrowserTab{
				{Index: 1, Title: "Home", URL: "https://93.184.216.34", Active: false},
				{Index: 2, Title: "Search", URL: "https://search.example", Active: true},
			},
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"list_tabs"}`,
	})
	if err != nil {
		t.Fatalf("browser_act list_tabs: %v", err)
	}
	if len(backend.tabsReqs) != 1 || backend.tabsReqs[0].Action != "list" {
		t.Fatalf("unexpected browser_act list_tabs dispatch: %#v", backend.tabsReqs)
	}
	var payload struct {
		Kind        string       `json:"kind"`
		Action      string       `json:"action"`
		ActiveIndex int          `json:"active_index"`
		Tabs        []BrowserTab `json:"tabs"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "list_tabs" || payload.Action != "list" || payload.ActiveIndex != 2 || len(payload.Tabs) != 2 {
		t.Fatalf("unexpected browser_act list_tabs output: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActFocusTab(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "focus",
			Status:      "focused",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 2, Title: "Docs", URL: "https://docs.example", Active: false},
				{Index: 3, Title: "AgentX", URL: "https://agentx.example", Active: true},
			},
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"focus_tab","tab_index":3,"wait_ms":120}`,
	})
	if err != nil {
		t.Fatalf("browser_act focus_tab: %v", err)
	}
	if len(backend.tabsReqs) != 1 || backend.tabsReqs[0].Action != "focus" || backend.tabsReqs[0].TabIndex != 3 || backend.tabsReqs[0].WaitMs != 120 {
		t.Fatalf("unexpected browser_act focus_tab dispatch: %#v", backend.tabsReqs)
	}
	var payload struct {
		Kind     string `json:"kind"`
		Action   string `json:"action"`
		Status   string `json:"status"`
		TabIndex int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "focus_tab" || payload.Action != "focus" || payload.Status != "focused" || payload.TabIndex != 3 {
		t.Fatalf("unexpected browser_act focus_tab output: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActFocusTabRequiresPendingPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-focus-popup-review")
	sessionRegistry.TrackTab("browser-act-focus-popup-review", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://93.184.216.34",
		Title:      "Home",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)
	popup := sessionRegistry.TrackTab("browser-act-focus-popup-review", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-act-focus-popup-review", BrowserSessionRoute{
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
		BrowserApp: "Safari",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         popup.ID,
		TabIndex:   popup.TabIndex,
		URL:        popup.URL,
		Title:      popup.Title,
		BrowserApp: popup.BrowserApp,
		Backend:    popup.Backend,
		Profile:    popup.Profile,
		Target:     popup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"focus_tab","tab_index":3,"wait_ms":120}`,
	})
	if err != nil {
		t.Fatalf("browser_act focus_tab popup review: %v", err)
	}
	if len(backend.tabsReqs) != 0 {
		t.Fatalf("expected focus_tab to be blocked before backend dispatch, got %#v", backend.tabsReqs)
	}
	var payload struct {
		Kind           string `json:"kind"`
		Action         string `json:"action"`
		Status         string `json:"status"`
		Force          bool   `json:"force"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TabIndex       int    `json:"tab_index"`
		TargetID       string `json:"target_id"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode focus_tab popup review output: %v", err)
	}
	if payload.Kind != "focus_tab" || payload.Action != "focus" || payload.Status != "review_required" || payload.Force || payload.ReviewDecision != "session_target_popup_review_required" || payload.ReviewReady || payload.TabIndex != 3 || strings.TrimSpace(payload.TargetID) != "" || !strings.Contains(payload.Note, "pending popup target") {
		t.Fatalf("unexpected browser_act focus_tab popup review payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActFocusTabConfirmsPendingPopupWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "focus",
			Status:      "focused",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 1, Title: "Home", URL: "https://93.184.216.34", Active: false},
				{Index: 3, Title: "Offer", URL: "https://popup.example/offer", Active: true},
			},
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-focus-popup-force")
	sessionRegistry.TrackTab("browser-act-focus-popup-force", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://93.184.216.34",
		Title:      "Home",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)
	popup := sessionRegistry.TrackTab("browser-act-focus-popup-force", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-act-focus-popup-force", BrowserSessionRoute{
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
		BrowserApp: "Safari",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         popup.ID,
		TabIndex:   popup.TabIndex,
		URL:        popup.URL,
		Title:      popup.Title,
		BrowserApp: popup.BrowserApp,
		Backend:    popup.Backend,
		Profile:    popup.Profile,
		Target:     popup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"focus_tab","tab_index":3,"wait_ms":120,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act focus_tab popup force: %v", err)
	}
	if len(backend.tabsReqs) != 1 || backend.tabsReqs[0].Action != "focus" || backend.tabsReqs[0].TabIndex != 3 || backend.tabsReqs[0].WaitMs != 120 {
		t.Fatalf("unexpected forced browser_act focus_tab dispatch: %#v", backend.tabsReqs)
	}
	var payload struct {
		Kind           string `json:"kind"`
		Action         string `json:"action"`
		Status         string `json:"status"`
		Force          bool   `json:"force"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TabIndex       int    `json:"tab_index"`
		TargetID       string `json:"target_id"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode focus_tab popup force output: %v", err)
	}
	if payload.Kind != "focus_tab" || payload.Action != "focus" || payload.Status != "focused" || !payload.Force || payload.ReviewDecision != "session_target_popup_review_confirmed" || !payload.ReviewReady || payload.TabIndex != 3 || strings.TrimSpace(payload.TargetID) == "" || !strings.Contains(payload.Note, "force=true") {
		t.Fatalf("unexpected browser_act focus_tab popup force payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActFocusTabRememberTargetPropagatesToLaterActions(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "focus",
			Status:      "focused",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 2, Title: "Docs", URL: "https://docs.example", Active: false},
				{Index: 3, Title: "AgentX", URL: "https://agentx.example", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "fake-extract",
			BrowserApp: "Safari",
			FinalURL:   "https://agentx.example",
			Title:      "AgentX",
			Content:    "AgentX remembered target",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-remember-target")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"focus_tab","tab_index":3,"remember_target":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act focus_tab remember_target: %v", err)
	}
	var payload struct {
		Kind                   string                                              `json:"kind"`
		Status                 string                                              `json:"status"`
		RememberTargetDecision string                                              `json:"remember_target_decision"`
		RememberTargetReady    bool                                                `json:"remember_target_ready"`
		SessionTargetSelection *agentxbrowserruntime.BrowserSessionTargetSelection `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode focus_tab remember_target output: %v", err)
	}
	if payload.Kind != "focus_tab" || payload.Status != "focused" || !payload.RememberTargetReady {
		t.Fatalf("unexpected focus_tab remember_target payload: %#v", payload)
	}
	if payload.RememberTargetDecision != "session_target_remembered" && payload.RememberTargetDecision != "session_target_already_remembered" {
		t.Fatalf("unexpected focus_tab remember_target decision: %#v", payload)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.TabIndex != 3 || payload.SessionTargetSelection.Source != "remember_target" || strings.TrimSpace(payload.SessionTargetSelection.ID) == "" {
		t.Fatalf("unexpected focus_tab remember_target selection: %#v", payload.SessionTargetSelection)
	}

	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after remember_target: %v", err)
	}
	if len(backend.extractReqs) != 1 || backend.extractReqs[0].TabIndex != 3 || backend.extractReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected extract to reuse remembered target, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_ActFocusTabRememberTargetPromotesSessionProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "proxy-tabs",
			BrowserApp:  "Chromium",
			Action:      "focus",
			Status:      "focused",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 2, Title: "Docs", URL: "https://node.example/docs", Active: false},
				{Index: 3, Title: "Workbench", URL: "https://node.example/workbench", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "proxy-extract",
			BrowserApp: "Chromium",
			FinalURL:   "https://node.example/workbench",
			Title:      "Workbench",
			Content:    "remembered node target",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_act", "browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-remember-target-promotes-profile")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"focus_tab","runtime_target":"node","profile":"workbench","tab_index":3,"remember_target":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act focus_tab remember_target node: %v", err)
	}
	var payload struct {
		Kind                    string                                              `json:"kind"`
		Status                  string                                              `json:"status"`
		RememberTargetDecision  string                                              `json:"remember_target_decision"`
		RememberTargetReady     bool                                                `json:"remember_target_ready"`
		SessionProfileSelection *browserRuntimeSessionProfileSelection              `json:"session_profile_selection"`
		SessionTargetSelection  *agentxbrowserruntime.BrowserSessionTargetSelection `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode focus_tab remember_target node output: %v", err)
	}
	if payload.Kind != "focus_tab" || payload.Status != "focused" || !payload.RememberTargetReady {
		t.Fatalf("unexpected focus_tab remember_target node payload: %#v", payload)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.TabIndex != 3 || payload.SessionTargetSelection.Source != "remember_target" {
		t.Fatalf("unexpected remembered target selection: %#v", payload.SessionTargetSelection)
	}
	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.RuntimeTarget != "node" || payload.SessionProfileSelection.Source != "remember_target" {
		t.Fatalf("expected remember_target to promote matching session profile, got %#v", payload.SessionProfileSelection)
	}

	statusOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"status"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime status after remember_target promote profile: %v", err)
	}
	var statusPayload struct {
		SelectedRoute struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
		} `json:"selected_route"`
		SessionProfileSelection struct {
			Profile string `json:"profile"`
			Source  string `json:"source"`
		} `json:"session_profile_selection"`
		RouteResolution struct {
			ProfileSource       string `json:"profile_source"`
			RuntimeTargetSource string `json:"runtime_target_source"`
			TargetSource        string `json:"target_source"`
		} `json:"route_resolution"`
	}
	if err := json.Unmarshal([]byte(statusOut), &statusPayload); err != nil {
		t.Fatalf("decode status output after remember_target promote profile: %v", err)
	}
	if statusPayload.SelectedRoute.Backend != "proxy" || statusPayload.SelectedRoute.Profile != "workbench" || statusPayload.SelectedRoute.RuntimeTarget != "node" {
		t.Fatalf("expected status to resolve node workbench route, got %#v", statusPayload.SelectedRoute)
	}
	if statusPayload.SessionProfileSelection.Profile != "workbench" || statusPayload.SessionProfileSelection.Source != "remember_target" {
		t.Fatalf("expected promoted session profile to persist after remember_target, got %#v", statusPayload.SessionProfileSelection)
	}
	if statusPayload.RouteResolution.ProfileSource != "remember_target" || statusPayload.RouteResolution.RuntimeTargetSource != "remember_target" {
		t.Fatalf("unexpected route resolution after remember_target promote profile: %#v", statusPayload.RouteResolution)
	}
}

func TestRegisterBrowserTools_ActListTabsRememberTargetRequiresPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "list",
			Status:      "ok",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 2, Title: "Docs", URL: "https://93.184.216.34/docs", Active: false},
				{Index: 3, Title: "Popup", URL: "https://93.184.216.35/popup", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "fake-extract",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.34/docs",
			Title:      "Docs",
			Content:    "still default target",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-list-tabs-remember-popup-review")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"list_tabs","remember_target":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act list_tabs remember_target popup review: %v", err)
	}
	var payload struct {
		Kind                   string                                              `json:"kind"`
		Action                 string                                              `json:"action"`
		Status                 string                                              `json:"status"`
		RememberTargetDecision string                                              `json:"remember_target_decision"`
		RememberTargetReady    bool                                                `json:"remember_target_ready"`
		SessionTargetSelection *agentxbrowserruntime.BrowserSessionTargetSelection `json:"session_target_selection"`
		Summary                *browserTopLevelSummary                             `json:"summary"`
		Display                *browserTopLevelDisplaySummary                      `json:"display"`
		Review                 *browserReviewSurfaceSummary                        `json:"review"`
		Surface                *browserTopLevelSurfaceSummary                      `json:"surface"`
		View                   *browserTopLevelViewSummary                         `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act list_tabs popup review output: %v", err)
	}
	if payload.Kind != "list_tabs" || payload.Action != "list" || payload.Status != "ok" || payload.RememberTargetDecision != "session_target_popup_review_required" || payload.RememberTargetReady || payload.SessionTargetSelection != nil {
		t.Fatalf("unexpected browser_act list_tabs popup review payload: %#v", payload)
	}

	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after list_tabs popup review: %v", err)
	}
	var extractPayload struct {
		Kind           string `json:"kind"`
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TargetID       string `json:"target_id"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &extractPayload); err != nil {
		t.Fatalf("decode browser_act extract after list_tabs popup review output: %v", err)
	}
	if extractPayload.Kind != "extract" || extractPayload.Status != "review_required" || extractPayload.ReviewDecision != "session_target_popup_review_required" || extractPayload.ReviewReady || strings.TrimSpace(extractPayload.TargetID) == "" || !strings.Contains(extractPayload.Note, "pending popup target") {
		t.Fatalf("unexpected browser_act extract payload after list_tabs popup review: %#v", extractPayload)
	}
	if len(backend.extractReqs) != 0 {
		t.Fatalf("expected extract not to adopt popup target before review, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_ActListTabsRememberTargetPopupConfirmedWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "list",
			Status:      "ok",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 2, Title: "Docs", URL: "https://93.184.216.34/docs", Active: false},
				{Index: 3, Title: "Popup", URL: "https://93.184.216.35/popup", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "fake-extract",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.35/popup",
			Title:      "Popup",
			Content:    "popup adopted",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-list-tabs-remember-popup-confirmed")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"list_tabs","remember_target":true,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act list_tabs remember_target popup confirm: %v", err)
	}
	var payload struct {
		Kind                   string                                              `json:"kind"`
		Action                 string                                              `json:"action"`
		Status                 string                                              `json:"status"`
		RememberTargetDecision string                                              `json:"remember_target_decision"`
		RememberTargetReady    bool                                                `json:"remember_target_ready"`
		SessionTargetSelection *agentxbrowserruntime.BrowserSessionTargetSelection `json:"session_target_selection"`
		Summary                *browserTopLevelSummary                             `json:"summary"`
		Display                *browserTopLevelDisplaySummary                      `json:"display"`
		Review                 *browserReviewSurfaceSummary                        `json:"review"`
		Surface                *browserTopLevelSurfaceSummary                      `json:"surface"`
		View                   *browserTopLevelViewSummary                         `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act list_tabs popup confirm output: %v", err)
	}
	if payload.Kind != "list_tabs" || payload.Action != "list" || payload.Status != "ok" || payload.RememberTargetDecision != "session_target_popup_review_confirmed" || !payload.RememberTargetReady || payload.SessionTargetSelection == nil || payload.SessionTargetSelection.TabIndex != 3 {
		t.Fatalf("unexpected browser_act list_tabs popup confirm payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "list_tabs_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "list_tabs_completed" ||
		payload.Review == nil || payload.Review.PolicyState != "session_target_popup_review_confirmed" || !payload.Review.Ready || payload.Review.Summary == nil || payload.Review.Summary.SummaryCode != "list_tabs_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "list_tabs_completed" || payload.Surface.ReviewPolicyState != "session_target_popup_review_confirmed" || !payload.Surface.ReviewReady ||
		payload.View == nil || payload.View.Kind != "review" || payload.View.SummaryCode != "list_tabs_completed" || payload.View.Review == nil || payload.View.Review.PolicyState != "session_target_popup_review_confirmed" {
		t.Fatalf("expected browser_act list_tabs popup confirm to expose success summary plus confirmed review surface, got %#v", payload)
	}

	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after list_tabs popup confirm: %v", err)
	}
	if len(backend.extractReqs) != 1 || backend.extractReqs[0].TabIndex != 3 {
		t.Fatalf("expected extract to adopt popup target after confirmation, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_ActCloseTabRememberTargetPopupConfirmedWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "close",
			Status:      "closed",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 2, Title: "Docs", URL: "https://93.184.216.34/docs", Active: false},
				{Index: 3, Title: "Popup", URL: "https://93.184.216.35/popup", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "fake-extract",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.35/popup",
			Title:      "Popup",
			Content:    "popup adopted after close",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-close-tab-remember-popup-confirmed")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"close_tab","tab_index":2,"remember_target":true,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act close_tab remember_target popup confirm: %v", err)
	}
	var payload struct {
		Kind                   string                                              `json:"kind"`
		Action                 string                                              `json:"action"`
		Status                 string                                              `json:"status"`
		RememberTargetDecision string                                              `json:"remember_target_decision"`
		RememberTargetReady    bool                                                `json:"remember_target_ready"`
		SessionTargetSelection *agentxbrowserruntime.BrowserSessionTargetSelection `json:"session_target_selection"`
		Summary                *browserTopLevelSummary                             `json:"summary"`
		Display                *browserTopLevelDisplaySummary                      `json:"display"`
		Review                 *browserReviewSurfaceSummary                        `json:"review"`
		Surface                *browserTopLevelSurfaceSummary                      `json:"surface"`
		View                   *browserTopLevelViewSummary                         `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act close_tab popup confirm output: %v", err)
	}
	if payload.Kind != "close_tab" || payload.Action != "close" || payload.Status != "closed" || payload.RememberTargetDecision != "session_target_popup_review_confirmed" || !payload.RememberTargetReady || payload.SessionTargetSelection == nil || payload.SessionTargetSelection.TabIndex != 3 {
		t.Fatalf("unexpected browser_act close_tab popup confirm payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "close_tab_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "close_tab_completed" ||
		payload.Review == nil || payload.Review.PolicyState != "session_target_popup_review_confirmed" || !payload.Review.Ready || payload.Review.Summary == nil || payload.Review.Summary.SummaryCode != "close_tab_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "close_tab_completed" || payload.Surface.ReviewPolicyState != "session_target_popup_review_confirmed" || !payload.Surface.ReviewReady ||
		payload.View == nil || payload.View.Kind != "review" || payload.View.SummaryCode != "close_tab_completed" || payload.View.Review == nil || payload.View.Review.PolicyState != "session_target_popup_review_confirmed" {
		t.Fatalf("expected browser_act close_tab popup confirm to expose success summary plus confirmed review surface, got %#v", payload)
	}

	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after close_tab popup confirm: %v", err)
	}
	if len(backend.extractReqs) != 1 || backend.extractReqs[0].TabIndex != 3 {
		t.Fatalf("expected extract to adopt popup target after close confirmation, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_TabsFocusRememberTargetPropagatesToBrowserAct(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "focus",
			Status:      "focused",
			ActiveIndex: 2,
			Tabs: []BrowserTab{
				{Index: 1, Title: "Home", URL: "https://93.184.216.34", Active: false},
				{Index: 2, Title: "Search", URL: "https://search.example", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "fake-extract",
			BrowserApp: "Safari",
			FinalURL:   "https://search.example",
			Title:      "Search",
			Content:    "remembered via browser_tabs",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_tabs", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-tabs-remember-target")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"focus","tab_index":2,"remember_target":true}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs focus remember_target: %v", err)
	}
	var payload struct {
		Action                 string                                              `json:"action"`
		Status                 string                                              `json:"status"`
		RememberTargetDecision string                                              `json:"remember_target_decision"`
		RememberTargetReady    bool                                                `json:"remember_target_ready"`
		SessionTargetSelection *agentxbrowserruntime.BrowserSessionTargetSelection `json:"session_target_selection"`
		Explanation            *browserTopLevelSummary                             `json:"explanation"`
		Diagnostics            *browserTopLevelSummary                             `json:"diagnostics"`
		Summary                *browserTopLevelSummary                             `json:"summary"`
		Display                *browserTopLevelDisplaySummary                      `json:"display"`
		Review                 *browserReviewSurfaceSummary                        `json:"review"`
		Surface                *browserTopLevelSurfaceSummary                      `json:"surface"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_tabs remember_target output: %v", err)
	}
	if payload.Action != "focus" || payload.Status != "focused" || !payload.RememberTargetReady {
		t.Fatalf("unexpected browser_tabs remember_target payload: %#v", payload)
	}
	if payload.RememberTargetDecision != "session_target_remembered" && payload.RememberTargetDecision != "session_target_already_remembered" {
		t.Fatalf("unexpected browser_tabs remember_target decision: %#v", payload)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.TabIndex != 2 || payload.SessionTargetSelection.Source != "remember_target" || strings.TrimSpace(payload.SessionTargetSelection.ID) == "" {
		t.Fatalf("unexpected browser_tabs remember_target selection: %#v", payload.SessionTargetSelection)
	}

	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after browser_tabs remember_target: %v", err)
	}
	if len(backend.extractReqs) != 1 || backend.extractReqs[0].TabIndex != 2 || backend.extractReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected extract to reuse browser_tabs remembered target, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_TabsFocusRequiresPendingPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_tabs", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-tabs-focus-popup-review")
	sessionRegistry.TrackTab("browser-tabs-focus-popup-review", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://93.184.216.34",
		Title:      "Home",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)
	popup := sessionRegistry.TrackTab("browser-tabs-focus-popup-review", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-tabs-focus-popup-review", BrowserSessionRoute{
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
		BrowserApp: "Safari",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         popup.ID,
		TabIndex:   popup.TabIndex,
		URL:        popup.URL,
		Title:      popup.Title,
		BrowserApp: popup.BrowserApp,
		Backend:    popup.Backend,
		Profile:    popup.Profile,
		Target:     popup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"focus","tab_index":3,"wait_ms":120}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs focus popup review: %v", err)
	}
	if len(backend.tabsReqs) != 0 {
		t.Fatalf("expected browser_tabs focus to be blocked before backend dispatch, got %#v", backend.tabsReqs)
	}
	var payload struct {
		Action         string                         `json:"action"`
		Status         string                         `json:"status"`
		Force          bool                           `json:"force"`
		ReviewDecision string                         `json:"review_decision"`
		ReviewReady    bool                           `json:"review_ready"`
		TabIndex       int                            `json:"tab_index"`
		TargetID       string                         `json:"target_id"`
		Note           string                         `json:"note"`
		Explanation    *browserTopLevelSummary        `json:"explanation"`
		Diagnostics    *browserTopLevelSummary        `json:"diagnostics"`
		Summary        *browserTopLevelSummary        `json:"summary"`
		Display        *browserTopLevelDisplaySummary `json:"display"`
		Review         *browserReviewSurfaceSummary   `json:"review"`
		Surface        *browserTopLevelSurfaceSummary `json:"surface"`
		View           *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_tabs focus popup review output: %v", err)
	}
	if payload.Action != "focus" || payload.Status != "review_required" || payload.Force || payload.ReviewDecision != "session_target_popup_review_required" || payload.ReviewReady || payload.TabIndex != 3 || strings.TrimSpace(payload.TargetID) != "" || !strings.Contains(payload.Note, "pending popup target") {
		t.Fatalf("unexpected browser_tabs focus popup review payload: %#v", payload)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "review" ||
		payload.Explanation.State != "manual_confirmation_required" ||
		payload.Explanation.SummaryCode != "popup_review_required" ||
		payload.Explanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected browser_tabs focus review explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil || payload.Diagnostics.SummaryCode != "popup_review_required" {
		t.Fatalf("unexpected browser_tabs focus review diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "popup_review_required" {
		t.Fatalf("unexpected browser_tabs focus review summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "review" ||
		payload.Display.State != "manual_confirmation_required" ||
		payload.Display.SummaryCode != "popup_review_required" ||
		payload.Display.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected browser_tabs focus review display: %#v", payload.Display)
	}
	if payload.Review == nil ||
		payload.Review.PolicyState != "popup_review_required" ||
		payload.Review.Decision != "session_target_popup_review_required" ||
		payload.Review.Ready ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.SummaryCode != "popup_review_required" {
		t.Fatalf("unexpected browser_tabs focus review surface: %#v", payload.Review)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "review" ||
		payload.Surface.State != "manual_confirmation_required" ||
		payload.Surface.SummaryCode != "popup_review_required" ||
		payload.Surface.ManualRetryHint != "rerun_with_force" ||
		payload.Surface.ReviewPolicyState != "popup_review_required" ||
		payload.Surface.ReviewDecision != "session_target_popup_review_required" ||
		payload.Surface.ReviewReady {
		t.Fatalf("unexpected browser_tabs focus review surface alias: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "review" ||
		payload.View.Category != "review" ||
		payload.View.State != "manual_confirmation_required" ||
		payload.View.SummaryCode != "popup_review_required" ||
		payload.View.ManualRetryHint != "rerun_with_force" ||
		payload.View.Review == nil ||
		payload.View.Review.PolicyState != "popup_review_required" ||
		payload.View.Review.Decision != "session_target_popup_review_required" ||
		payload.View.Review.Ready {
		t.Fatalf("unexpected browser_tabs focus review view alias: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_TabsFocusConfirmsPendingPopupWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "focus",
			Status:      "focused",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 1, Title: "Home", URL: "https://93.184.216.34", Active: false},
				{Index: 3, Title: "Offer", URL: "https://popup.example/offer", Active: true},
			},
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_tabs", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-tabs-focus-popup-force")
	sessionRegistry.TrackTab("browser-tabs-focus-popup-force", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://93.184.216.34",
		Title:      "Home",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)
	popup := sessionRegistry.TrackTab("browser-tabs-focus-popup-force", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-tabs-focus-popup-force", BrowserSessionRoute{
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
		BrowserApp: "Safari",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         popup.ID,
		TabIndex:   popup.TabIndex,
		URL:        popup.URL,
		Title:      popup.Title,
		BrowserApp: popup.BrowserApp,
		Backend:    popup.Backend,
		Profile:    popup.Profile,
		Target:     popup.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"focus","tab_index":3,"wait_ms":120,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs focus popup force: %v", err)
	}
	if len(backend.tabsReqs) != 1 || backend.tabsReqs[0].Action != "focus" || backend.tabsReqs[0].TabIndex != 3 || backend.tabsReqs[0].WaitMs != 120 {
		t.Fatalf("unexpected forced browser_tabs focus dispatch: %#v", backend.tabsReqs)
	}
	var payload struct {
		Action         string `json:"action"`
		Status         string `json:"status"`
		Force          bool   `json:"force"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TabIndex       int    `json:"tab_index"`
		TargetID       string `json:"target_id"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_tabs focus popup force output: %v", err)
	}
	if payload.Action != "focus" || payload.Status != "focused" || !payload.Force || payload.ReviewDecision != "session_target_popup_review_confirmed" || !payload.ReviewReady || payload.TabIndex != 3 || strings.TrimSpace(payload.TargetID) == "" || !strings.Contains(payload.Note, "force=true") {
		t.Fatalf("unexpected browser_tabs focus popup force payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_TabsFocusRememberTargetPromotesSessionProfile(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	hostBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	nodeBackend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "proxy-tabs",
			BrowserApp:  "Chromium",
			Action:      "focus",
			Status:      "focused",
			ActiveIndex: 2,
			Tabs: []BrowserTab{
				{Index: 1, Title: "Home", URL: "https://node.example/home", Active: false},
				{Index: 2, Title: "Search", URL: "https://node.example/search", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "proxy-extract",
			BrowserApp: "Chromium",
			FinalURL:   "https://node.example/search",
			Title:      "Search",
			Content:    "remembered via browser_tabs node",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              hostBackend,
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_tabs", "browser_act", "browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-tabs-remember-target-promotes-profile")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"focus","runtime_target":"node","profile":"workbench","tab_index":2,"remember_target":true}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs focus remember_target node: %v", err)
	}
	var payload struct {
		Action                  string                                              `json:"action"`
		Status                  string                                              `json:"status"`
		RememberTargetDecision  string                                              `json:"remember_target_decision"`
		RememberTargetReady     bool                                                `json:"remember_target_ready"`
		SessionProfileSelection *browserRuntimeSessionProfileSelection              `json:"session_profile_selection"`
		SessionTargetSelection  *agentxbrowserruntime.BrowserSessionTargetSelection `json:"session_target_selection"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_tabs remember_target node output: %v", err)
	}
	if payload.Action != "focus" || payload.Status != "focused" || !payload.RememberTargetReady {
		t.Fatalf("unexpected browser_tabs remember_target node payload: %#v", payload)
	}
	if payload.SessionTargetSelection == nil || payload.SessionTargetSelection.TabIndex != 2 || payload.SessionTargetSelection.Source != "remember_target" {
		t.Fatalf("unexpected remembered browser_tabs target selection: %#v", payload.SessionTargetSelection)
	}
	if payload.SessionProfileSelection == nil || payload.SessionProfileSelection.Profile != "workbench" || payload.SessionProfileSelection.RuntimeTarget != "node" || payload.SessionProfileSelection.Source != "remember_target" {
		t.Fatalf("expected browser_tabs remember_target to promote session profile, got %#v", payload.SessionProfileSelection)
	}

	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after browser_tabs remember_target promote profile: %v", err)
	}
	if len(nodeBackend.extractReqs) != 1 || nodeBackend.extractReqs[0].TabIndex != 2 || nodeBackend.extractReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("expected extract to reuse browser_tabs remembered node target, got %#v", nodeBackend.extractReqs)
	}
}

func TestRegisterBrowserTools_TabsListRememberTargetRequiresPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "list",
			Status:      "ok",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 2, Title: "Docs", URL: "https://93.184.216.34/docs", Active: false},
				{Index: 3, Title: "Popup", URL: "https://93.184.216.35/popup", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "fake-extract",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.34/docs",
			Title:      "Docs",
			Content:    "still default target",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_tabs", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-tabs-list-remember-popup-review")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"list","remember_target":true}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs list remember_target popup review: %v", err)
	}
	var payload struct {
		Action                 string                                              `json:"action"`
		Status                 string                                              `json:"status"`
		RememberTargetDecision string                                              `json:"remember_target_decision"`
		RememberTargetReady    bool                                                `json:"remember_target_ready"`
		SessionTargetSelection *agentxbrowserruntime.BrowserSessionTargetSelection `json:"session_target_selection"`
		Explanation            *browserTopLevelSummary                             `json:"explanation"`
		Diagnostics            *browserTopLevelSummary                             `json:"diagnostics"`
		Summary                *browserTopLevelSummary                             `json:"summary"`
		Display                *browserTopLevelDisplaySummary                      `json:"display"`
		Review                 *browserReviewSurfaceSummary                        `json:"review"`
		Surface                *browserTopLevelSurfaceSummary                      `json:"surface"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_tabs popup review output: %v", err)
	}
	if payload.Action != "list" || payload.Status != "ok" || payload.RememberTargetDecision != "session_target_popup_review_required" || payload.RememberTargetReady || payload.SessionTargetSelection != nil {
		t.Fatalf("unexpected browser_tabs popup review payload: %#v", payload)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "review" ||
		payload.Explanation.State != "manual_confirmation_required" ||
		payload.Explanation.SummaryCode != "popup_review_required" ||
		payload.Explanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected browser_tabs remember review explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil || payload.Diagnostics.SummaryCode != "popup_review_required" {
		t.Fatalf("unexpected browser_tabs remember review diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "popup_review_required" {
		t.Fatalf("unexpected browser_tabs remember review summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "review" ||
		payload.Display.State != "manual_confirmation_required" ||
		payload.Display.SummaryCode != "popup_review_required" ||
		payload.Display.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected browser_tabs remember review display: %#v", payload.Display)
	}
	if payload.Review == nil ||
		payload.Review.PolicyState != "popup_review_required" ||
		payload.Review.Decision != "session_target_popup_review_required" ||
		payload.Review.Ready ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.SummaryCode != "popup_review_required" {
		t.Fatalf("unexpected browser_tabs remember review surface: %#v", payload.Review)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "review" ||
		payload.Surface.State != "manual_confirmation_required" ||
		payload.Surface.SummaryCode != "popup_review_required" ||
		payload.Surface.ManualRetryHint != "rerun_with_force" ||
		payload.Surface.ReviewPolicyState != "popup_review_required" ||
		payload.Surface.ReviewDecision != "session_target_popup_review_required" ||
		payload.Surface.ReviewReady {
		t.Fatalf("unexpected browser_tabs remember review surface alias: %#v", payload.Surface)
	}

	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after browser_tabs popup review: %v", err)
	}
	var extractPayload struct {
		Kind           string `json:"kind"`
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TargetID       string `json:"target_id"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &extractPayload); err != nil {
		t.Fatalf("decode browser_act extract after browser_tabs popup review output: %v", err)
	}
	if extractPayload.Kind != "extract" || extractPayload.Status != "review_required" || extractPayload.ReviewDecision != "session_target_popup_review_required" || extractPayload.ReviewReady || strings.TrimSpace(extractPayload.TargetID) == "" || !strings.Contains(extractPayload.Note, "pending popup target") {
		t.Fatalf("unexpected browser_act extract payload after browser_tabs popup review: %#v", extractPayload)
	}
	if len(backend.extractReqs) != 0 {
		t.Fatalf("expected extract not to adopt popup target before review, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_TabsListRememberTargetPopupConfirmedWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "list",
			Status:      "ok",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 2, Title: "Docs", URL: "https://93.184.216.34/docs", Active: false},
				{Index: 3, Title: "Popup", URL: "https://93.184.216.35/popup", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "fake-extract",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.35/popup",
			Title:      "Popup",
			Content:    "popup adopted",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_tabs", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-tabs-list-remember-popup-confirmed")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"list","remember_target":true,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs list remember_target popup confirm: %v", err)
	}
	var payload struct {
		Action                 string                                              `json:"action"`
		Status                 string                                              `json:"status"`
		RememberTargetDecision string                                              `json:"remember_target_decision"`
		RememberTargetReady    bool                                                `json:"remember_target_ready"`
		SessionTargetSelection *agentxbrowserruntime.BrowserSessionTargetSelection `json:"session_target_selection"`
		Summary                *browserTopLevelSummary                             `json:"summary"`
		Display                *browserTopLevelDisplaySummary                      `json:"display"`
		Review                 *browserReviewSurfaceSummary                        `json:"review"`
		Surface                *browserTopLevelSurfaceSummary                      `json:"surface"`
		View                   *browserTopLevelViewSummary                         `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_tabs popup confirm output: %v", err)
	}
	if payload.Action != "list" || payload.Status != "ok" || payload.RememberTargetDecision != "session_target_popup_review_confirmed" || !payload.RememberTargetReady || payload.SessionTargetSelection == nil || payload.SessionTargetSelection.TabIndex != 3 {
		t.Fatalf("unexpected browser_tabs popup confirm payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "list_tabs_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "list_tabs_completed" ||
		payload.Review == nil || payload.Review.PolicyState != "session_target_popup_review_confirmed" || !payload.Review.Ready || payload.Review.Summary == nil || payload.Review.Summary.SummaryCode != "list_tabs_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "list_tabs_completed" || payload.Surface.ReviewPolicyState != "session_target_popup_review_confirmed" || !payload.Surface.ReviewReady ||
		payload.View == nil || payload.View.Kind != "review" || payload.View.SummaryCode != "list_tabs_completed" || payload.View.Review == nil || payload.View.Review.PolicyState != "session_target_popup_review_confirmed" {
		t.Fatalf("expected browser_tabs popup confirm to expose success summary plus confirmed review surface, got %#v", payload)
	}

	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after browser_tabs popup confirm: %v", err)
	}
	if len(backend.extractReqs) != 1 || backend.extractReqs[0].TabIndex != 3 {
		t.Fatalf("expected extract to adopt popup target after confirmation, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_ActCloseTabRememberTargetRequiresPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "close",
			Status:      "closed",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 1, Title: "Home", URL: "https://93.184.216.34", Active: false},
				{Index: 3, Title: "Popup", URL: "https://93.184.216.35/popup", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "fake-extract",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.34",
			Title:      "Home",
			Content:    "still default target",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-close-tab-remember-popup-review")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"close_tab","tab_index":2,"remember_target":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act close_tab remember_target popup review: %v", err)
	}
	var payload struct {
		Kind                   string `json:"kind"`
		Action                 string `json:"action"`
		Status                 string `json:"status"`
		RememberTargetDecision string `json:"remember_target_decision"`
		RememberTargetReady    bool   `json:"remember_target_ready"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act close_tab popup review output: %v", err)
	}
	if payload.Kind != "close_tab" || payload.Action != "close" || payload.Status != "closed" || payload.RememberTargetDecision != "session_target_popup_review_required" || payload.RememberTargetReady {
		t.Fatalf("unexpected browser_act close_tab popup review payload: %#v", payload)
	}

	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after close_tab popup review: %v", err)
	}
	var extractPayload struct {
		Kind           string `json:"kind"`
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TargetID       string `json:"target_id"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &extractPayload); err != nil {
		t.Fatalf("decode browser_act extract after close_tab popup review output: %v", err)
	}
	if extractPayload.Kind != "extract" || extractPayload.Status != "review_required" || extractPayload.ReviewDecision != "session_target_popup_review_required" || extractPayload.ReviewReady || strings.TrimSpace(extractPayload.TargetID) == "" || !strings.Contains(extractPayload.Note, "pending popup target") {
		t.Fatalf("unexpected browser_act extract payload after close_tab popup review: %#v", extractPayload)
	}
	if len(backend.extractReqs) != 0 {
		t.Fatalf("expected extract not to adopt popup target after close_tab popup review, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_TabsCloseRememberTargetRequiresPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "close",
			Status:      "closed",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 1, Title: "Home", URL: "https://93.184.216.34", Active: false},
				{Index: 3, Title: "Popup", URL: "https://93.184.216.35/popup", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "fake-extract",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.34",
			Title:      "Home",
			Content:    "still default target",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_tabs", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-tabs-close-remember-popup-review")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"close","tab_index":2,"remember_target":true}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs close remember_target popup review: %v", err)
	}
	var payload struct {
		Action                 string                         `json:"action"`
		Status                 string                         `json:"status"`
		RememberTargetDecision string                         `json:"remember_target_decision"`
		RememberTargetReady    bool                           `json:"remember_target_ready"`
		Explanation            *browserTopLevelSummary        `json:"explanation"`
		Diagnostics            *browserTopLevelSummary        `json:"diagnostics"`
		Summary                *browserTopLevelSummary        `json:"summary"`
		Display                *browserTopLevelDisplaySummary `json:"display"`
		Review                 *browserReviewSurfaceSummary   `json:"review"`
		Surface                *browserTopLevelSurfaceSummary `json:"surface"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_tabs close popup review output: %v", err)
	}
	if payload.Action != "close" || payload.Status != "closed" || payload.RememberTargetDecision != "session_target_popup_review_required" || payload.RememberTargetReady {
		t.Fatalf("unexpected browser_tabs close popup review payload: %#v", payload)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "review" ||
		payload.Explanation.State != "manual_confirmation_required" ||
		payload.Explanation.SummaryCode != "popup_review_required" ||
		payload.Explanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected browser_tabs close review explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil || payload.Diagnostics.SummaryCode != "popup_review_required" {
		t.Fatalf("unexpected browser_tabs close review diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "popup_review_required" {
		t.Fatalf("unexpected browser_tabs close review summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "review" ||
		payload.Display.State != "manual_confirmation_required" ||
		payload.Display.SummaryCode != "popup_review_required" ||
		payload.Display.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected browser_tabs close review display: %#v", payload.Display)
	}
	if payload.Review == nil ||
		payload.Review.PolicyState != "popup_review_required" ||
		payload.Review.Decision != "session_target_popup_review_required" ||
		payload.Review.Ready ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.SummaryCode != "popup_review_required" {
		t.Fatalf("unexpected browser_tabs close review surface: %#v", payload.Review)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "review" ||
		payload.Surface.State != "manual_confirmation_required" ||
		payload.Surface.SummaryCode != "popup_review_required" ||
		payload.Surface.ManualRetryHint != "rerun_with_force" ||
		payload.Surface.ReviewPolicyState != "popup_review_required" ||
		payload.Surface.ReviewDecision != "session_target_popup_review_required" ||
		payload.Surface.ReviewReady {
		t.Fatalf("unexpected browser_tabs close review surface alias: %#v", payload.Surface)
	}

	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after browser_tabs close popup review: %v", err)
	}
	var extractPayload struct {
		Kind           string `json:"kind"`
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TargetID       string `json:"target_id"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &extractPayload); err != nil {
		t.Fatalf("decode browser_act extract after browser_tabs close popup review output: %v", err)
	}
	if extractPayload.Kind != "extract" || extractPayload.Status != "review_required" || extractPayload.ReviewDecision != "session_target_popup_review_required" || extractPayload.ReviewReady || strings.TrimSpace(extractPayload.TargetID) == "" || !strings.Contains(extractPayload.Note, "pending popup target") {
		t.Fatalf("unexpected browser_act extract payload after browser_tabs close popup review: %#v", extractPayload)
	}
	if len(backend.extractReqs) != 0 {
		t.Fatalf("expected extract not to adopt popup target after browser_tabs close popup review, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_TabsCloseRememberTargetPopupConfirmedWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "close",
			Status:      "closed",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 1, Title: "Home", URL: "https://93.184.216.34", Active: false},
				{Index: 3, Title: "Popup", URL: "https://93.184.216.35/popup", Active: true},
			},
		},
		extractResult: BrowserExtractResult{
			Backend:    "fake-extract",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.35/popup",
			Title:      "Popup",
			Content:    "popup adopted after close",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_tabs", "browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-tabs-close-remember-popup-confirmed")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"close","tab_index":2,"remember_target":true,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs close remember_target popup confirm: %v", err)
	}
	var payload struct {
		Action                 string                         `json:"action"`
		Status                 string                         `json:"status"`
		RememberTargetDecision string                         `json:"remember_target_decision"`
		RememberTargetReady    bool                           `json:"remember_target_ready"`
		Explanation            *browserTopLevelSummary        `json:"explanation"`
		Diagnostics            *browserTopLevelSummary        `json:"diagnostics"`
		Summary                *browserTopLevelSummary        `json:"summary"`
		Display                *browserTopLevelDisplaySummary `json:"display"`
		Review                 *browserReviewSurfaceSummary   `json:"review"`
		Surface                *browserTopLevelSurfaceSummary `json:"surface"`
		View                   *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_tabs close popup confirm output: %v", err)
	}
	if payload.Action != "close" || payload.Status != "closed" || payload.RememberTargetDecision != "session_target_popup_review_confirmed" || !payload.RememberTargetReady {
		t.Fatalf("unexpected browser_tabs close popup confirm payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "close_tab_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "close_tab_completed" ||
		payload.Review == nil || payload.Review.PolicyState != "session_target_popup_review_confirmed" || !payload.Review.Ready || payload.Review.Summary == nil || payload.Review.Summary.SummaryCode != "close_tab_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "close_tab_completed" || payload.Surface.ReviewPolicyState != "session_target_popup_review_confirmed" || !payload.Surface.ReviewReady ||
		payload.View == nil || payload.View.Kind != "review" || payload.View.SummaryCode != "close_tab_completed" || payload.View.Review == nil || payload.View.Review.PolicyState != "session_target_popup_review_confirmed" {
		t.Fatalf("expected browser_tabs close popup confirm to expose success summary plus confirmed review surface, got %#v", payload)
	}

	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after browser_tabs close popup confirm: %v", err)
	}
	if len(backend.extractReqs) != 1 || backend.extractReqs[0].TabIndex != 3 {
		t.Fatalf("expected extract to adopt popup target after browser_tabs close confirmation, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_TabsCloseCleansUpLatestPopupStormReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "close",
			Status:      "closed",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 1, Title: "Home", URL: "https://93.184.216.34", Active: false},
				{Index: 3, Title: "Popup", URL: "https://93.184.216.35/popup", Active: true},
			},
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_tabs", "browser_runtime"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-tabs-close-popup-storm-cleanup")
	home := sessionRegistry.TrackTab("browser-tabs-close-popup-storm-cleanup", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://93.184.216.34",
		Title:      "Home",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)
	first := sessionRegistry.TrackTab("browser-tabs-close-popup-storm-cleanup", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://93.184.216.35/popup",
		Title:      "Popup",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	second := sessionRegistry.TrackTab("browser-tabs-close-popup-storm-cleanup", BrowserSessionTarget{
		TabIndex:   4,
		URL:        "https://93.184.216.35/bonus",
		Title:      "Bonus",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	route := agentxbrowserruntime.BrowserSessionRoute{
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
		BrowserApp: "Safari",
	}
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-tabs-close-popup-storm-cleanup", route, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         first.ID,
		TabIndex:   first.TabIndex,
		URL:        first.URL,
		Title:      first.Title,
		BrowserApp: first.BrowserApp,
		Backend:    first.Backend,
		Profile:    first.Profile,
		Target:     first.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-tabs-close-popup-storm-cleanup", route, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         second.ID,
		TabIndex:   second.TabIndex,
		URL:        second.URL,
		Title:      second.Title,
		BrowserApp: second.BrowserApp,
		Backend:    second.Backend,
		Profile:    second.Profile,
		Target:     second.Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"close","tab_index":4}`,
	}); err != nil {
		t.Fatalf("browser_tabs close popup storm cleanup: %v", err)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"host","profile":"default"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions after popup storm cleanup: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			CurrentTargetID             string `json:"current_target_id"`
			PendingTargetReviewCount    int    `json:"pending_target_review_count"`
			BlockedAutoFollowRouteCount int    `json:"blocked_auto_follow_route_count"`
			PopupStormRouteCount        int    `json:"popup_storm_route_count"`
		} `json:"session_binding"`
		SessionRoutes []struct {
			CurrentTargetID          string `json:"current_target_id"`
			PendingTargetReviewCount int    `json:"pending_target_review_count"`
			FollowPolicyState        string `json:"follow_policy_state"`
			PopupPolicyState         string `json:"popup_policy_state"`
			PendingTargetReview      *struct {
				ID       string `json:"id"`
				TabIndex int    `json:"tab_index"`
			} `json:"pending_target_review"`
			Targets []struct {
				ID       string `json:"id"`
				TabIndex int    `json:"tab_index"`
				Current  bool   `json:"current"`
			} `json:"targets"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime sessions popup storm cleanup output: %v", err)
	}
	if payload.SessionBinding.CurrentTargetID != home.ID || payload.SessionBinding.PendingTargetReviewCount != 1 || payload.SessionBinding.BlockedAutoFollowRouteCount != 1 || payload.SessionBinding.PopupStormRouteCount != 0 {
		t.Fatalf("expected session binding cleanup after closing latest popup, got binding=%#v routes=%#v", payload.SessionBinding, payload.SessionRoutes)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].CurrentTargetID != home.ID {
		t.Fatalf("expected single route with preserved current target, got %#v", payload.SessionRoutes)
	}
	if payload.SessionRoutes[0].PendingTargetReview == nil || payload.SessionRoutes[0].PendingTargetReview.ID != first.ID || payload.SessionRoutes[0].PendingTargetReview.TabIndex != 3 || payload.SessionRoutes[0].PendingTargetReviewCount != 1 || payload.SessionRoutes[0].FollowPolicyState != "popup_review_required" || payload.SessionRoutes[0].PopupPolicyState != "popup_review_required" {
		t.Fatalf("expected fallback to earlier popup review after close cleanup, got %#v", payload.SessionRoutes[0])
	}
	if len(payload.SessionRoutes[0].Targets) != 2 || payload.SessionRoutes[0].Targets[0].ID != home.ID || !payload.SessionRoutes[0].Targets[0].Current {
		t.Fatalf("expected home and first popup targets after cleanup, got %#v", payload.SessionRoutes[0].Targets)
	}
	for _, target := range payload.SessionRoutes[0].Targets {
		if target.ID == second.ID || target.TabIndex == 4 {
			t.Fatalf("expected closed latest popup to be removed from targets, got %#v", payload.SessionRoutes[0].Targets)
		}
	}
}

func TestRegisterBrowserTools_ActCloseTab(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "close",
			Status:      "closed",
			ActiveIndex: 1,
			Tabs: []BrowserTab{
				{Index: 1, Title: "Remaining", URL: "https://remaining.example", Active: true},
			},
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"close_tab","tab_index":2}`,
	})
	if err != nil {
		t.Fatalf("browser_act close_tab: %v", err)
	}
	if len(backend.tabsReqs) != 1 || backend.tabsReqs[0].Action != "close" || backend.tabsReqs[0].TabIndex != 2 {
		t.Fatalf("unexpected browser_act close_tab dispatch: %#v", backend.tabsReqs)
	}
	var payload struct {
		Kind     string `json:"kind"`
		Action   string `json:"action"`
		Status   string `json:"status"`
		TabIndex int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "close_tab" || payload.Action != "close" || payload.Status != "closed" || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act close_tab output: %#v", payload)
	}
}
