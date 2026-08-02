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

func TestBrowserEffectiveBrowserAppUsesTrackedTarget(t *testing.T) {
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "browser-effective-browser-app"
	tracked := sessionRegistry.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		URL:        "https://93.184.216.34",
		Title:      "Example",
		BrowserApp: "Safari",
		Backend:    "safari_javascript",
		Profile:    "default",
		Target:     "host",
	})
	callCtx := WithToolSessionID(context.Background(), sessionID)
	target := browserToolTarget{
		Value:    "current",
		TargetID: tracked.ID,
		Explicit: true,
	}
	got := browserEffectiveBrowserApp(callCtx, sessionRegistry, BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}, false, "", target)
	if got != "Safari" {
		t.Fatalf("expected tracked browser app Safari, got %q", got)
	}
}

func TestRegisterBrowserTools_Navigate(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		navigateResult: BrowserNavigateResult{Backend: "fake-navigate", BrowserApp: "Safari", FinalURL: "https://1.1.1.1/final", Title: "Final", Status: "navigated"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_navigate"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: `{"url":"https://1.1.1.1","wait_ms":900,"target":"tab:2"}`,
	})
	if err != nil {
		t.Fatalf("browser_navigate: %v", err)
	}
	if len(backend.navigateReqs) != 1 || backend.navigateReqs[0].URL != "https://1.1.1.1" || backend.navigateReqs[0].WaitMs != 900 || backend.navigateReqs[0].TabIndex != 2 {
		t.Fatalf("unexpected navigate request: %#v", backend.navigateReqs)
	}
	var payload struct {
		Backend                string                                       `json:"backend"`
		FinalURL               string                                       `json:"final_url"`
		Status                 string                                       `json:"status"`
		Target                 string                                       `json:"target"`
		TabIndex               int                                          `json:"tab_index"`
		PostNavigationSnapshot *browserPostNavigationSnapshotRecommendation `json:"post_navigation_snapshot"`
		Summary                *browserTopLevelSummary                      `json:"summary"`
		Display                *browserTopLevelDisplaySummary               `json:"display"`
		Surface                *browserTopLevelSurfaceSummary               `json:"surface"`
		View                   *browserTopLevelViewSummary                  `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Backend != "fake-navigate" || payload.FinalURL != "https://1.1.1.1/final" || payload.Status != "navigated" || payload.Target != "tab:2" || payload.TabIndex != 2 {
		t.Fatalf("unexpected navigate output: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "navigate_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "navigate_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "navigate_completed" ||
		payload.View == nil || payload.View.SummaryCode != "navigate_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_navigate to expose stable navigate summary surfaces, got %#v", payload)
	}
	if payload.PostNavigationSnapshot == nil ||
		payload.PostNavigationSnapshot.Recommendation != "take_compact_snapshot" ||
		payload.PostNavigationSnapshot.Tool != "browser_act" ||
		payload.PostNavigationSnapshot.Kind != "snapshot" ||
		!payload.PostNavigationSnapshot.Compact ||
		payload.PostNavigationSnapshot.MaxElements != browserPostNavigationSnapshotMaxElements {
		t.Fatalf("expected browser_navigate to expose compact snapshot recommendation, got %#v", payload.PostNavigationSnapshot)
	}
}

func TestRegisterBrowserTools_NavigateReportsBotDetectionWarning(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		navigateResult: BrowserNavigateResult{
			Backend:    "fake-navigate",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.34/challenge",
			Title:      "Verify you are human",
			Status:     "navigated",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_navigate"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: `{"url":"https://93.184.216.34"}`,
	})
	if err != nil {
		t.Fatalf("browser_navigate bot warning: %v", err)
	}
	var payload struct {
		BotDetectionWarning *browserBotDetectionWarning `json:"bot_detection_warning"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode bot warning output: %v", err)
	}
	if payload.BotDetectionWarning == nil ||
		payload.BotDetectionWarning.WarningCode != "browser_bot_detection_challenge" ||
		payload.BotDetectionWarning.Severity != "warning" ||
		payload.BotDetectionWarning.Source != "title" ||
		payload.BotDetectionWarning.Signal != "human_verification" {
		t.Fatalf("expected structured bot detection warning, got %#v", payload.BotDetectionWarning)
	}
}

func TestRegisterBrowserTools_NavigateRedirectRequiresReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		navigateResult: BrowserNavigateResult{
			Backend:    "fake-navigate",
			BrowserApp: "Safari",
			FinalURL:   "https://1.0.0.1/landing",
			Title:      "Redirected",
			Status:     "navigated",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_navigate"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: `{"url":"https://1.1.1.1","wait_ms":900,"target":"tab:2"}`,
	})
	if err != nil {
		t.Fatalf("browser_navigate redirect review: %v", err)
	}
	if len(backend.navigateReqs) != 1 || backend.navigateReqs[0].URL != "https://1.1.1.1" || backend.navigateReqs[0].WaitMs != 900 || backend.navigateReqs[0].TabIndex != 2 {
		t.Fatalf("unexpected navigate request: %#v", backend.navigateReqs)
	}
	var payload struct {
		FinalURL       string `json:"final_url"`
		Status         string `json:"status"`
		Force          bool   `json:"force"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		Target         string `json:"target"`
		TabIndex       int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.FinalURL != "https://1.0.0.1/landing" || payload.Status != "review_required" || payload.Force || payload.ReviewDecision != "navigate_redirect_review_required" || payload.ReviewReady || payload.Target != "tab:2" || payload.TabIndex != 2 {
		t.Fatalf("unexpected redirect review navigate output: %#v", payload)
	}
}

func TestRegisterBrowserTools_NavigateRedirectReviewBlocksLaterExtract(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-navigate-redirect-pending")
	sessionRegistry.TrackTab("browser-navigate-redirect-pending", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://1.1.1.1/start",
		Title:      "Start",
		BrowserApp: "Safari",
		Backend:    "system",
		Target:     "host",
	}, true)
	backend := &fakeBrowserBackend{
		navigateResult: BrowserNavigateResult{
			Backend:    "fake-navigate",
			BrowserApp: "Safari",
			FinalURL:   "https://1.0.0.1/landing",
			Title:      "Redirected",
			Status:     "navigated",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_navigate", "browser_extract", "browser_runtime"},
	})

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: `{"url":"https://1.1.1.1","target":"tab:2"}`,
	}); err != nil {
		t.Fatalf("browser_navigate redirect review: %v", err)
	}
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_extract",
		Arguments: `{"target":"tab:2"}`,
	})
	if err != nil {
		t.Fatalf("browser_extract after redirect review: %v", err)
	}
	var payload struct {
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TabIndex       int    `json:"tab_index"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode extract after redirect review output: %v", err)
	}
	if payload.Status != "review_required" || payload.ReviewDecision != "session_target_redirect_review_required" || payload.ReviewReady || payload.TabIndex != 2 || !strings.Contains(payload.Note, "redirected target") {
		t.Fatalf("unexpected extract payload after redirect review: %#v", payload)
	}
	if len(backend.extractReqs) != 0 {
		t.Fatalf("expected extract not to follow redirected target before review, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_NavigateRedirectReviewWithoutPriorSelectionBlocksImplicitFollow(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-navigate-redirect-implicit-follow")
	backend := &fakeBrowserBackend{
		navigateResult: BrowserNavigateResult{
			Backend:    "fake-navigate",
			BrowserApp: "Safari",
			FinalURL:   "https://1.0.0.1/landing",
			Title:      "Redirected",
			Status:     "navigated",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_navigate", "browser_extract", "browser_runtime"},
	})

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: `{"url":"https://1.1.1.1","target":"tab:2"}`,
	}); err != nil {
		t.Fatalf("browser_navigate redirect review without prior selection: %v", err)
	}

	sessionsOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions after redirect review: %v", err)
	}
	var sessionsPayload struct {
		SessionBinding struct {
			PendingTargetReviewCount    int `json:"pending_target_review_count"`
			BlockedAutoFollowRouteCount int `json:"blocked_auto_follow_route_count"`
		} `json:"session_binding"`
		SessionRoutes []struct {
			CurrentTargetID     string `json:"current_target_id"`
			CurrentTargetSource string `json:"current_target_source"`
			FollowPolicyState   string `json:"follow_policy_state"`
			FollowPolicyReason  string `json:"follow_policy_reason"`
			PopupPolicyState    string `json:"popup_policy_state"`
			PendingTargetReview *struct {
				ID       string `json:"id"`
				Decision string `json:"decision"`
			} `json:"pending_target_review"`
			Targets []struct {
				ID      string `json:"id"`
				Current bool   `json:"current"`
			} `json:"targets"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(sessionsOut), &sessionsPayload); err != nil {
		t.Fatalf("decode sessions after redirect review output: %v", err)
	}
	if sessionsPayload.SessionBinding.PendingTargetReviewCount != 1 || sessionsPayload.SessionBinding.BlockedAutoFollowRouteCount != 1 || len(sessionsPayload.SessionRoutes) != 1 {
		t.Fatalf("expected one pending redirect review route, got %#v", sessionsPayload)
	}
	if sessionsPayload.SessionRoutes[0].CurrentTargetID != "" || sessionsPayload.SessionRoutes[0].CurrentTargetSource != "" {
		t.Fatalf("expected no current target selection after redirect review on previously unselected route, got %#v", sessionsPayload.SessionRoutes[0])
	}
	if sessionsPayload.SessionRoutes[0].FollowPolicyState != "redirect_review_required" || !strings.Contains(sessionsPayload.SessionRoutes[0].FollowPolicyReason, "redirected target") || sessionsPayload.SessionRoutes[0].PopupPolicyState != "" {
		t.Fatalf("expected redirect follow policy without popup posture, got %#v", sessionsPayload.SessionRoutes[0])
	}
	if sessionsPayload.SessionRoutes[0].PendingTargetReview == nil || sessionsPayload.SessionRoutes[0].PendingTargetReview.Decision != "session_target_redirect_review_required" {
		t.Fatalf("expected pending redirect review in sessions payload, got %#v", sessionsPayload.SessionRoutes[0])
	}
	if len(sessionsPayload.SessionRoutes[0].Targets) != 1 || sessionsPayload.SessionRoutes[0].Targets[0].Current {
		t.Fatalf("expected redirected target to stay unselected until confirmed, got %#v", sessionsPayload.SessionRoutes[0].Targets)
	}

	extractOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_extract",
		Arguments: `{}`,
	})
	if err != nil {
		t.Fatalf("browser_extract implicit follow after redirect review: %v", err)
	}
	var extractPayload struct {
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TargetID       string `json:"target_id"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(extractOut), &extractPayload); err != nil {
		t.Fatalf("decode implicit extract after redirect review output: %v", err)
	}
	if extractPayload.Status != "review_required" || extractPayload.ReviewDecision != "session_target_redirect_review_required" || extractPayload.ReviewReady || strings.TrimSpace(extractPayload.TargetID) == "" || !strings.Contains(extractPayload.Note, "redirected target") {
		t.Fatalf("unexpected implicit extract payload after redirect review: %#v", extractPayload)
	}
	if len(backend.extractReqs) != 0 {
		t.Fatalf("expected implicit extract not to follow redirected target before review, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_NavigateRedirectConfirmedWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		navigateResult: BrowserNavigateResult{
			Backend:    "fake-navigate",
			BrowserApp: "Safari",
			FinalURL:   "https://1.0.0.1/landing",
			Title:      "Redirected",
			Status:     "navigated",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_navigate"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: `{"url":"https://1.1.1.1","wait_ms":900,"target":"tab:2","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_navigate redirect force: %v", err)
	}
	var payload struct {
		FinalURL       string `json:"final_url"`
		Status         string `json:"status"`
		Force          bool   `json:"force"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.FinalURL != "https://1.0.0.1/landing" || payload.Status != "navigated" || !payload.Force || payload.ReviewDecision != "navigate_redirect_review_confirmed" || !payload.ReviewReady {
		t.Fatalf("unexpected redirect force navigate output: %#v", payload)
	}
}

func TestRegisterBrowserTools_NavigateRejectsFinalURLPolicyViolation(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		navigateResult: BrowserNavigateResult{
			Backend:    "fake-navigate",
			BrowserApp: "Safari",
			FinalURL:   "https://127.0.0.1/admin",
			Title:      "Private Redirect",
			Status:     "navigated",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_navigate"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: `{"url":"https://1.1.1.1","wait_ms":900,"target":"tab:2"}`,
	})
	if err == nil || !strings.Contains(err.Error(), "final_url") {
		t.Fatalf("expected final_url policy error, got %v", err)
	}
}

func TestRegisterBrowserTools_Tabs(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "focus",
			Status:      "focused",
			ActiveIndex: 2,
			Tabs: []BrowserTab{
				{Index: 1, Title: "A", URL: "https://a.example", Active: false},
				{Index: 2, Title: "B", URL: "https://b.example", Active: true},
			},
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_tabs"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"focus","target":"tab:2","wait_ms":150}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs: %v", err)
	}
	if len(backend.tabsReqs) != 1 || backend.tabsReqs[0].Action != "focus" || backend.tabsReqs[0].TabIndex != 2 {
		t.Fatalf("unexpected tabs request: %#v", backend.tabsReqs)
	}
	var payload struct {
		Backend     string                         `json:"backend"`
		Action      string                         `json:"action"`
		Status      string                         `json:"status"`
		Target      string                         `json:"target"`
		TabIndex    int                            `json:"tab_index"`
		ActiveIndex int                            `json:"active_index"`
		Tabs        []BrowserTab                   `json:"tabs"`
		Summary     *browserTopLevelSummary        `json:"summary"`
		Display     *browserTopLevelDisplaySummary `json:"display"`
		Surface     *browserTopLevelSurfaceSummary `json:"surface"`
		View        *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Backend != "fake-tabs" || payload.Action != "focus" || payload.Status != "focused" || payload.Target != "tab:2" || payload.TabIndex != 2 || payload.ActiveIndex != 2 || len(payload.Tabs) != 2 {
		t.Fatalf("unexpected tabs output: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "focus_tab_completed" ||
		payload.Display == nil || payload.Display.SummaryCode != "focus_tab_completed" ||
		payload.Surface == nil || payload.Surface.SummaryCode != "focus_tab_completed" ||
		payload.View == nil || payload.View.SummaryCode != "focus_tab_completed" || payload.View.Kind != "result" {
		t.Fatalf("expected browser_tabs to expose stable focus summary surfaces, got %#v", payload)
	}
}

func TestRegisterBrowserTools_TargetHandleResolvesAcrossCalls(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "list",
			Status:      "ok",
			ActiveIndex: 2,
			Tabs: []BrowserTab{
				{Index: 1, Title: "A", URL: "https://a.example", Active: false},
				{Index: 2, Title: "B", URL: "https://b.example", Active: true},
			},
		},
		navigateResult: BrowserNavigateResult{
			Backend:    "fake-navigate",
			FinalURL:   "https://1.1.1.1/final",
			Title:      "Final",
			Status:     "navigated",
			BrowserApp: "Safari",
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: NewBrowserSessionRegistry(),
		EnabledTools:    []string{"browser_tabs", "browser_navigate"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-session-1")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"list"}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs: %v", err)
	}
	var tabsPayload struct {
		Tabs []BrowserTab `json:"tabs"`
	}
	if err := json.Unmarshal([]byte(out), &tabsPayload); err != nil {
		t.Fatalf("decode tabs output: %v", err)
	}
	if len(tabsPayload.Tabs) != 2 || tabsPayload.Tabs[1].TargetID == "" {
		t.Fatalf("expected tracked target ids in tabs payload, got %#v", tabsPayload)
	}

	targetID := tabsPayload.Tabs[1].TargetID
	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: fmt.Sprintf(`{"url":"https://1.1.1.1","target":"target:%s"}`, targetID),
	})
	if err != nil {
		t.Fatalf("browser_navigate: %v", err)
	}
	if len(backend.navigateReqs) != 1 || backend.navigateReqs[0].TabIndex != 2 {
		t.Fatalf("expected target handle to resolve to tab 2, got %#v", backend.navigateReqs)
	}
	var navigatePayload struct {
		Target   string `json:"target"`
		TargetID string `json:"target_id"`
	}
	if err := json.Unmarshal([]byte(out), &navigatePayload); err != nil {
		t.Fatalf("decode navigate output: %v", err)
	}
	if navigatePayload.Target != "target:"+targetID || navigatePayload.TargetID != targetID {
		t.Fatalf("unexpected navigate target payload: %#v", navigatePayload)
	}
}

func TestRegisterBrowserTools_TargetHandlesAreScopedByRuntimeRoute(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			tabsResult: BrowserTabsResult{
				Backend:     "host-tabs",
				BrowserApp:  "Safari",
				Action:      "list",
				Status:      "ok",
				ActiveIndex: 1,
				Tabs: []BrowserTab{
					{Index: 1, Title: "Host", URL: "https://93.184.216.34", Active: true},
				},
			},
			navigateResult: BrowserNavigateResult{
				Backend:    "host-navigate",
				BrowserApp: "Safari",
				FinalURL:   "https://93.184.216.34/final",
				Title:      "Host Final",
				Status:     "navigated",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			tabsResult: BrowserTabsResult{
				Backend:     "node-tabs",
				BrowserApp:  "Chromium",
				Action:      "list",
				Status:      "ok",
				ActiveIndex: 1,
				Tabs: []BrowserTab{
					{Index: 1, Title: "Node", URL: "https://93.184.216.35", Active: true},
				},
			},
			navigateResult: BrowserNavigateResult{
				Backend:    "node-navigate",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.35/final",
				Title:      "Node Final",
				Status:     "navigated",
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         hostBackend,
		NodeBackend:     nodeBackend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_tabs", "browser_navigate"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-route-session")

	hostOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"list","runtime_target":"host"}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs host: %v", err)
	}
	nodeOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"list","profile":"isolated","runtime_target":"node","browser":"Chromium"}`,
	})
	if err != nil {
		t.Fatalf("browser_tabs node: %v", err)
	}

	var hostPayload, nodePayload struct {
		Tabs []BrowserTab `json:"tabs"`
	}
	if err := json.Unmarshal([]byte(hostOut), &hostPayload); err != nil {
		t.Fatalf("decode host tabs: %v", err)
	}
	if err := json.Unmarshal([]byte(nodeOut), &nodePayload); err != nil {
		t.Fatalf("decode node tabs: %v", err)
	}
	if len(hostPayload.Tabs) != 1 || len(nodePayload.Tabs) != 1 {
		t.Fatalf("expected single tab per route, got host=%#v node=%#v", hostPayload, nodePayload)
	}
	hostTargetID := hostPayload.Tabs[0].TargetID
	nodeTargetID := nodePayload.Tabs[0].TargetID
	if hostTargetID == "" || nodeTargetID == "" || hostTargetID == nodeTargetID {
		t.Fatalf("expected distinct route-scoped target ids, got host=%q node=%q", hostTargetID, nodeTargetID)
	}

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: fmt.Sprintf(`{"url":"https://93.184.216.34/next","target":"target:%s"}`, hostTargetID),
	}); err != nil {
		t.Fatalf("browser_navigate host target: %v", err)
	}
	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_navigate",
		Arguments: fmt.Sprintf(`{"url":"https://93.184.216.35/next","target":"target:%s","profile":"isolated","runtime_target":"node","browser":"Chromium"}`, nodeTargetID),
	}); err != nil {
		t.Fatalf("browser_navigate node target: %v", err)
	}
	if len(hostBackend.navigateReqs) != 1 || hostBackend.navigateReqs[0].TabIndex != 1 {
		t.Fatalf("expected host target handle to stay on host tab 1, got %#v", hostBackend.navigateReqs)
	}
	if len(nodeBackend.navigateReqs) != 1 || nodeBackend.navigateReqs[0].TabIndex != 1 {
		t.Fatalf("expected node target handle to stay on node tab 1, got %#v", nodeBackend.navigateReqs)
	}
}

func TestRegisterBrowserTools_BrowserActSupportsTargetHandle(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "fake-tabs",
			BrowserApp:  "Safari",
			Action:      "list",
			Status:      "ok",
			ActiveIndex: 2,
			Tabs: []BrowserTab{
				{Index: 1, Title: "A", URL: "https://a.example", Active: false},
				{Index: 2, Title: "B", URL: "https://b.example", Active: true},
			},
		},
		clickResult: BrowserClickResult{
			Backend:    "fake-click",
			BrowserApp: "Safari",
			FinalURL:   "https://b.example/after",
			Title:      "B After",
			Status:     "clicked",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: NewBrowserSessionRegistry(),
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-session-act")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"list_tabs"}`,
	})
	if err != nil {
		t.Fatalf("browser_act list_tabs: %v", err)
	}
	var listPayload struct {
		Tabs []BrowserTab `json:"tabs"`
	}
	if err := json.Unmarshal([]byte(out), &listPayload); err != nil {
		t.Fatalf("decode act list output: %v", err)
	}
	if len(listPayload.Tabs) != 2 || listPayload.Tabs[1].TargetID == "" {
		t.Fatalf("expected tracked target ids in act tabs payload, got %#v", listPayload)
	}

	targetID := listPayload.Tabs[1].TargetID
	out, err = reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: fmt.Sprintf(`{"kind":"click","target":"target:%s","selector":"button.buy"}`, targetID),
	})
	if err != nil {
		t.Fatalf("browser_act click: %v", err)
	}
	if len(backend.clickReqs) != 1 || backend.clickReqs[0].TabIndex != 2 {
		t.Fatalf("expected act target handle to resolve to tab 2, got %#v", backend.clickReqs)
	}
	var clickPayload struct {
		Target   string `json:"target"`
		TargetID string `json:"target_id"`
	}
	if err := json.Unmarshal([]byte(out), &clickPayload); err != nil {
		t.Fatalf("decode act click output: %v", err)
	}
	if clickPayload.Target != "target:"+targetID || clickPayload.TargetID != targetID {
		t.Fatalf("unexpected act click target payload: %#v", clickPayload)
	}
}
