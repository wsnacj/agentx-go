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

func TestRegisterBrowserTools_RuntimeSessions(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionRunRegistry := newTestSessionRunRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-session")
	sessionRegistry.TrackTab("browser-runtime-session", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/a",
		Title:      "Node A",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	sessionRegistry.TrackTab("browser-runtime-session", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://host.example/b",
		Title:      "Host B",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, true)
	sessionRunRegistry.Record("browser-runtime-session", BrowserSessionRunInfo{
		RunID:    "run-77",
		NodeID:   "mac-alpha",
		Status:   "running",
		Provider: "gateway",
		Action:   "run_status",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-session", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		SessionRunRegistry:   sessionRunRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions: %v", err)
	}
	var payload struct {
		Action             string   `json:"action"`
		Status             string   `json:"status"`
		SessionID          string   `json:"session_id"`
		SessionTargetCount int      `json:"session_target_count"`
		RuntimeActions     []string `json:"runtime_actions"`
		SessionBinding     struct {
			NodeRunCount               int            `json:"node_run_count"`
			ActiveNodeRunID            string         `json:"active_node_run_id"`
			NodeRunStatusCounts        map[string]int `json:"node_run_status_counts"`
			BrowserProfileCount        int            `json:"browser_profile_count"`
			ActiveBrowserProfile       string         `json:"active_browser_profile"`
			BrowserProfileStatusCounts map[string]int `json:"browser_profile_status_counts"`
			Coordination               struct {
				State                     string   `json:"state"`
				BrowserOnNode             bool     `json:"browser_on_node"`
				HasActiveNodeRun          bool     `json:"has_active_node_run"`
				HasRunningBrowserProfile  bool     `json:"has_running_browser_profile"`
				SyncBrowserAction         string   `json:"sync_browser_action"`
				RecommendedBrowserActions []string `json:"recommended_browser_actions"`
				PrimaryBrowserAction      string   `json:"primary_browser_action"`
				PrimaryNodeAction         string   `json:"primary_node_action"`
				NextStep                  string   `json:"next_step"`
				RecommendedNodeActions    []string `json:"recommended_node_actions"`
			} `json:"coordination"`
		} `json:"session_binding"`
		SessionRuns []struct {
			RunID  string `json:"run_id"`
			Status string `json:"status"`
			Action string `json:"action"`
		} `json:"session_runs"`
		SessionProfiles []struct {
			Backend       string `json:"backend"`
			Profile       string `json:"profile"`
			RuntimeTarget string `json:"runtime_target"`
			Status        string `json:"status"`
		} `json:"session_profiles"`
		SessionRoutes []struct {
			Backend         string `json:"backend"`
			Profile         string `json:"profile"`
			RuntimeTarget   string `json:"runtime_target"`
			CurrentTargetID string `json:"current_target_id"`
			Targets         []struct {
				ID            string `json:"id"`
				TabIndex      int    `json:"tab_index"`
				Current       bool   `json:"current"`
				RuntimeTarget string `json:"runtime_target"`
			} `json:"targets"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Action != "sessions" || payload.Status != "ok" || payload.SessionID != "browser-runtime-session" {
		t.Fatalf("unexpected runtime sessions payload: %#v", payload)
	}
	if !browserStringSliceContains(payload.RuntimeActions, "sessions") {
		t.Fatalf("expected runtime actions to include sessions, got %#v", payload.RuntimeActions)
	}
	if payload.SessionTargetCount != 1 || len(payload.SessionRoutes) != 1 {
		t.Fatalf("expected filtered node session view, got %#v", payload)
	}
	if payload.SessionBinding.NodeRunCount != 1 || payload.SessionBinding.ActiveNodeRunID != "run-77" || payload.SessionBinding.NodeRunStatusCounts["running"] != 1 || payload.SessionBinding.BrowserProfileCount != 1 || payload.SessionBinding.ActiveBrowserProfile != "isolated" || payload.SessionBinding.BrowserProfileStatusCounts["running"] != 1 || len(payload.SessionRuns) != 1 || payload.SessionRuns[0].RunID != "run-77" || payload.SessionRuns[0].Action != "run_status" {
		t.Fatalf("expected node runs in runtime sessions payload, got %#v", payload)
	}
	if payload.SessionBinding.Coordination.State != "coordinated" || !payload.SessionBinding.Coordination.BrowserOnNode || !payload.SessionBinding.Coordination.HasActiveNodeRun || !payload.SessionBinding.Coordination.HasRunningBrowserProfile || !browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedNodeActions, "nodes action=run_wait") {
		t.Fatalf("expected coordinated session binding guidance, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.SessionBinding.Coordination.SyncBrowserAction != "browser_runtime action=coordinate coordination_goal=sync" || !browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=coordinate coordination_goal=sync") {
		t.Fatalf("expected coordinated sync action hint, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser_runtime action=coordinate coordination_goal=sync" || payload.SessionBinding.Coordination.PrimaryNodeAction != "nodes action=run_status" || payload.SessionBinding.Coordination.NextStep != "browser_runtime action=coordinate coordination_goal=sync" {
		t.Fatalf("expected coordinated session binding plan, got %#v", payload.SessionBinding.Coordination)
	}
	if len(payload.SessionProfiles) != 1 || payload.SessionProfiles[0].Backend != "proxy" || payload.SessionProfiles[0].Profile != "isolated" || payload.SessionProfiles[0].RuntimeTarget != "node" || payload.SessionProfiles[0].Status != "running" {
		t.Fatalf("expected browser profile session state in runtime sessions payload, got %#v", payload.SessionProfiles)
	}
	if payload.SessionRoutes[0].Backend != "proxy" || payload.SessionRoutes[0].Profile != "isolated" || payload.SessionRoutes[0].RuntimeTarget != "node" {
		t.Fatalf("unexpected session route payload: %#v", payload.SessionRoutes[0])
	}
	if len(payload.SessionRoutes[0].Targets) != 1 || payload.SessionRoutes[0].Targets[0].TabIndex != 1 || !payload.SessionRoutes[0].Targets[0].Current {
		t.Fatalf("unexpected session targets payload: %#v", payload.SessionRoutes[0].Targets)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsIncludesPendingPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-session-pending-popup")

	sessionRegistry.TrackTab("browser-runtime-session-pending-popup", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/home",
		Title:      "Home",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)

	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "proxy",
			BrowserApp:  "Chromium",
			Action:      "list",
			Status:      "ok",
			ActiveIndex: 3,
			Tabs: []BrowserTab{
				{Index: 1, Title: "Home", URL: "https://node.example/home", Active: false},
				{Index: 3, Title: "Popup", URL: "https://popup.example/offer", Active: true},
			},
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}

	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          backend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_tabs", "browser_runtime"},
	})

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"list","runtime_target":"node","profile":"isolated","remember_target":true}`,
	}); err != nil {
		t.Fatalf("browser_tabs list remember_target popup review: %v", err)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions after popup review: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			PendingTargetReviewCount    int `json:"pending_target_review_count"`
			BlockedAutoFollowRouteCount int `json:"blocked_auto_follow_route_count"`
		} `json:"session_binding"`
		SessionRoutes []struct {
			CurrentTargetID     string `json:"current_target_id"`
			CurrentTargetSource string `json:"current_target_source"`
			FollowPolicyState   string `json:"follow_policy_state"`
			FollowPolicyReason  string `json:"follow_policy_reason"`
			PendingTargetReview *struct {
				ID       string `json:"id"`
				TabIndex int    `json:"tab_index"`
				Decision string `json:"decision"`
				Reason   string `json:"reason"`
			} `json:"pending_target_review"`
			Targets []struct {
				ID       string `json:"id"`
				TabIndex int    `json:"tab_index"`
				Current  bool   `json:"current"`
			} `json:"targets"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime sessions pending popup output: %v", err)
	}
	if payload.SessionBinding.PendingTargetReviewCount != 1 || payload.SessionBinding.BlockedAutoFollowRouteCount != 1 || len(payload.SessionRoutes) != 1 {
		t.Fatalf("expected one pending popup review in runtime sessions payload, got %#v", payload)
	}
	if payload.SessionRoutes[0].CurrentTargetSource != "tracked_active_tab" || payload.SessionRoutes[0].PendingTargetReview == nil {
		t.Fatalf("expected current target and pending popup review, got %#v", payload.SessionRoutes[0])
	}
	if payload.SessionRoutes[0].FollowPolicyState != "popup_review_required" || !strings.Contains(payload.SessionRoutes[0].FollowPolicyReason, "pending popup target") {
		t.Fatalf("expected popup follow policy state, got %#v", payload.SessionRoutes[0])
	}
	if payload.SessionRoutes[0].PendingTargetReview.TabIndex != 3 || payload.SessionRoutes[0].PendingTargetReview.Decision != "session_target_popup_review_required" || !strings.Contains(payload.SessionRoutes[0].PendingTargetReview.Reason, "newly opened active tab") {
		t.Fatalf("unexpected pending popup review payload: %#v", payload.SessionRoutes[0].PendingTargetReview)
	}
	if len(payload.SessionRoutes[0].Targets) != 2 || !payload.SessionRoutes[0].Targets[0].Current || payload.SessionRoutes[0].Targets[0].TabIndex != 1 {
		t.Fatalf("expected prior current target to remain selected after popup review, got %#v", payload.SessionRoutes[0].Targets)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsListTabsEmptyClearsOrphanRouteState(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-session-empty-tabs")

	tracked := sessionRegistry.TrackTabs("browser-runtime-session-empty-tabs", []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example/home", Title: "Home", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 3, URL: "https://popup.example/offer", Title: "Offer", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	if len(tracked) != 2 {
		t.Fatalf("expected tracked tabs, got %#v", tracked)
	}
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-session-empty-tabs", agentxbrowserruntime.BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
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

	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{
		tabsResult: BrowserTabsResult{
			Backend:     "proxy",
			BrowserApp:  "Chromium",
			Action:      "list",
			Status:      "ok",
			ActiveIndex: 0,
			Tabs:        nil,
		},
	}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		NodeBackend:     backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_tabs", "browser_runtime"},
	})

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_tabs",
		Arguments: `{"action":"list","runtime_target":"node","profile":"isolated"}`,
	}); err != nil {
		t.Fatalf("browser_tabs list empty cleanup: %v", err)
	}

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions after empty list cleanup: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			PendingTargetReviewCount    int `json:"pending_target_review_count"`
			BlockedAutoFollowRouteCount int `json:"blocked_auto_follow_route_count"`
			PopupStormRouteCount        int `json:"popup_storm_route_count"`
		} `json:"session_binding"`
		SessionRoutes []any `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime sessions empty list cleanup output: %v", err)
	}
	if payload.SessionBinding.PendingTargetReviewCount != 0 || payload.SessionBinding.BlockedAutoFollowRouteCount != 0 || payload.SessionBinding.PopupStormRouteCount != 0 {
		t.Fatalf("expected empty list cleanup to clear pending popup state, got %#v", payload.SessionBinding)
	}
	if len(payload.SessionRoutes) != 0 {
		t.Fatalf("expected empty list cleanup to clear route snapshot, got %#v", payload.SessionRoutes)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsPrunesStalePendingReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-session-prune-stale-review")

	tracked := sessionRegistry.TrackTabs("browser-runtime-session-prune-stale-review", []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example/home", Title: "Home", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	if len(tracked) != 1 {
		t.Fatalf("expected tracked tab, got %#v", tracked)
	}
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-session-prune-stale-review", agentxbrowserruntime.BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         "missing-popup-target",
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		NodeBackend:     backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions after stale review prune: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			PendingTargetReviewCount    int `json:"pending_target_review_count"`
			BlockedAutoFollowRouteCount int `json:"blocked_auto_follow_route_count"`
		} `json:"session_binding"`
		SessionRoutes []struct {
			CurrentTargetID          string `json:"current_target_id"`
			FollowPolicyState        string `json:"follow_policy_state"`
			PendingTargetReviewCount int    `json:"pending_target_review_count"`
			PendingTargetReview      any    `json:"pending_target_review"`
			Targets                  []struct {
				ID      string `json:"id"`
				Current bool   `json:"current"`
			} `json:"targets"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime sessions stale review prune output: %v", err)
	}
	if payload.SessionBinding.PendingTargetReviewCount != 0 || payload.SessionBinding.BlockedAutoFollowRouteCount != 0 {
		t.Fatalf("expected stale pending review to be pruned from session binding, got %#v", payload.SessionBinding)
	}
	if len(payload.SessionRoutes) != 1 {
		t.Fatalf("expected single surviving route after stale review prune, got %#v", payload.SessionRoutes)
	}
	if payload.SessionRoutes[0].PendingTargetReview != nil || payload.SessionRoutes[0].PendingTargetReviewCount != 0 || payload.SessionRoutes[0].FollowPolicyState != "auto_follow_allowed" {
		t.Fatalf("expected stale pending review to be removed from route state, got %#v", payload.SessionRoutes[0])
	}
	if payload.SessionRoutes[0].CurrentTargetID != tracked[0].ID || len(payload.SessionRoutes[0].Targets) != 1 || !payload.SessionRoutes[0].Targets[0].Current {
		t.Fatalf("expected surviving tracked target to remain current, got %#v", payload.SessionRoutes[0])
	}
}

func TestRegisterBrowserTools_RuntimeSessionsIncludesPopupStormPosture(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-session-popup-storm")

	tracked := sessionRegistry.TrackTabs("browser-runtime-session-popup-storm", []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example/home", Title: "Home", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 3, URL: "https://popup.example/offer", Title: "Offer", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 4, URL: "https://popup.example/bonus", Title: "Bonus", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	if len(tracked) != 3 {
		t.Fatalf("expected tracked tabs, got %#v", tracked)
	}
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-session-popup-storm", agentxbrowserruntime.BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
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
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-session-popup-storm", agentxbrowserruntime.BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
		BrowserApp: "Chromium",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
		ID:         tracked[2].ID,
		TabIndex:   tracked[2].TabIndex,
		URL:        tracked[2].URL,
		Title:      tracked[2].Title,
		BrowserApp: tracked[2].BrowserApp,
		Backend:    tracked[2].Backend,
		Profile:    tracked[2].Profile,
		Target:     tracked[2].Target,
		Decision:   "session_target_popup_review_required",
		Reason:     "pending popup review",
	})

	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		NodeBackend:     backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions after popup storm: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			PendingTargetReviewCount    int `json:"pending_target_review_count"`
			BlockedAutoFollowRouteCount int `json:"blocked_auto_follow_route_count"`
			PopupStormRouteCount        int `json:"popup_storm_route_count"`
		} `json:"session_binding"`
		SessionRoutes []struct {
			PendingTargetReviewCount int    `json:"pending_target_review_count"`
			FollowPolicyState        string `json:"follow_policy_state"`
			FollowPolicyReason       string `json:"follow_policy_reason"`
			PopupPolicyState         string `json:"popup_policy_state"`
			PopupPolicyReason        string `json:"popup_policy_reason"`
			PendingTargetReview      *struct {
				ID       string `json:"id"`
				TabIndex int    `json:"tab_index"`
			} `json:"pending_target_review"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime sessions popup storm output: %v", err)
	}
	if payload.SessionBinding.PendingTargetReviewCount != 1 || payload.SessionBinding.BlockedAutoFollowRouteCount != 1 || payload.SessionBinding.PopupStormRouteCount != 1 || len(payload.SessionRoutes) != 1 {
		t.Fatalf("expected popup storm summary in runtime sessions payload, got %#v", payload)
	}
	if payload.SessionRoutes[0].PendingTargetReview == nil || payload.SessionRoutes[0].PendingTargetReview.ID != tracked[2].ID || payload.SessionRoutes[0].PendingTargetReviewCount != 2 {
		t.Fatalf("expected latest pending popup review with count 2, got %#v", payload.SessionRoutes[0])
	}
	if payload.SessionRoutes[0].FollowPolicyState != "popup_storm_review_required" || !strings.Contains(payload.SessionRoutes[0].FollowPolicyReason, "accumulated 2 pending popup targets") {
		t.Fatalf("expected popup storm follow policy, got %#v", payload.SessionRoutes[0])
	}
	if payload.SessionRoutes[0].PopupPolicyState != "popup_storm_review_required" || !strings.Contains(payload.SessionRoutes[0].PopupPolicyReason, "accumulated 2 pending popup targets") {
		t.Fatalf("expected popup storm posture, got %#v", payload.SessionRoutes[0])
	}
}

func TestRegisterBrowserTools_RuntimeSessionsPopupStormCleanupHints(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-session-popup-storm-hints")

	tracked := sessionRegistry.TrackTabs("browser-runtime-session-popup-storm-hints", []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example/home", Title: "Home", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 3, URL: "https://popup.example/offer", Title: "Offer", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 4, URL: "https://popup.example/bonus", Title: "Bonus", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	if len(tracked) != 3 {
		t.Fatalf("expected tracked tabs, got %#v", tracked)
	}
	route := agentxbrowserruntime.BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	for _, item := range tracked[1:] {
		sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-session-popup-storm-hints", route, agentxbrowserruntime.BrowserSessionTargetReview{
			ID:         item.ID,
			TabIndex:   item.TabIndex,
			URL:        item.URL,
			Title:      item.Title,
			BrowserApp: item.BrowserApp,
			Backend:    item.Backend,
			Profile:    item.Profile,
			Target:     item.Target,
			Decision:   "session_target_popup_review_required",
			Reason:     "pending popup review",
		})
	}

	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		NodeBackend:     backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions popup storm hints: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			Coordination struct {
				PrimaryBrowserAction      string   `json:"primary_browser_action"`
				NextStep                  string   `json:"next_step"`
				RecommendedBrowserActions []string `json:"recommended_browser_actions"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime sessions popup storm hints output: %v", err)
	}
	if payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser_runtime action=sessions" || payload.SessionBinding.Coordination.NextStep != "browser_runtime action=sessions" {
		t.Fatalf("expected popup storm to prefer close cleanup, got %#v", payload.SessionBinding.Coordination)
	}
	if !browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=sessions") ||
		!browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=select_target") {
		t.Fatalf("expected popup storm cleanup guidance in recommended actions, got %#v", payload.SessionBinding.Coordination)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsPopupStormWithoutLiveNodePrefersReset(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-session-popup-storm-stale-node")

	tracked := sessionRegistry.TrackTabs("browser-runtime-session-popup-storm-stale-node", []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example/home", Title: "Home", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 3, URL: "https://popup.example/offer", Title: "Offer", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 4, URL: "https://popup.example/bonus", Title: "Bonus", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	if len(tracked) != 3 {
		t.Fatalf("expected tracked tabs, got %#v", tracked)
	}
	route := agentxbrowserruntime.BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node", BrowserApp: "Chromium"}
	for _, item := range tracked[1:] {
		sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-session-popup-storm-stale-node", route, agentxbrowserruntime.BrowserSessionTargetReview{
			ID:         item.ID,
			TabIndex:   item.TabIndex,
			URL:        item.URL,
			Title:      item.Title,
			BrowserApp: item.BrowserApp,
			Backend:    item.Backend,
			Profile:    item.Profile,
			Target:     item.Target,
			Decision:   "session_target_popup_review_required",
			Reason:     "pending popup review",
		})
	}
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-session-popup-storm-stale-node", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "stopped",
	})

	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          backend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions popup storm stale node: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			Coordination struct {
				State                     string   `json:"state"`
				PrimaryBrowserAction      string   `json:"primary_browser_action"`
				NextStep                  string   `json:"next_step"`
				RecommendedBrowserActions []string `json:"recommended_browser_actions"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime sessions popup storm stale node output: %v", err)
	}
	if payload.SessionBinding.Coordination.State != "browser_attached" {
		t.Fatalf("expected browser_attached stale node posture, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser_runtime action=clear_session" || payload.SessionBinding.Coordination.NextStep != "browser_runtime action=clear_session" {
		t.Fatalf("expected stale node popup storm to prefer reset, got %#v", payload.SessionBinding.Coordination)
	}
	if !browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=clear_session") ||
		!browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=coordinate coordination_goal=ensure") ||
		browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=sessions") {
		t.Fatalf("expected stale node popup storm guidance to prioritize reset/ensure without live popup actions, got %#v", payload.SessionBinding.Coordination)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsRedirectReviewHints(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-session-redirect-hints")

	tracked := sessionRegistry.TrackTabs("browser-runtime-session-redirect-hints", []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example/home", Title: "Home", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 2, URL: "https://redirected.example/dashboard", Title: "Redirected", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	if len(tracked) != 2 {
		t.Fatalf("expected tracked tabs, got %#v", tracked)
	}
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-session-redirect-hints", agentxbrowserruntime.BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
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
		Decision:   "session_target_redirect_review_required",
		Reason:     "redirect review",
	})

	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		NodeBackend:     backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions redirect hints: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			Coordination struct {
				PrimaryBrowserAction      string   `json:"primary_browser_action"`
				NextStep                  string   `json:"next_step"`
				RecommendedBrowserActions []string `json:"recommended_browser_actions"`
			} `json:"coordination"`
		} `json:"session_binding"`
		SessionRoutes []struct {
			FollowPolicyState string `json:"follow_policy_state"`
		} `json:"session_routes"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime sessions redirect hints output: %v", err)
	}
	if len(payload.SessionRoutes) != 1 || payload.SessionRoutes[0].FollowPolicyState != "redirect_review_required" {
		t.Fatalf("expected redirect review route summary, got %#v", payload.SessionRoutes)
	}
	if payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser_runtime action=select_target" || payload.SessionBinding.Coordination.NextStep != "browser_runtime action=select_target" {
		t.Fatalf("expected redirect review to prefer explicit target adoption, got %#v", payload.SessionBinding.Coordination)
	}
	if !browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=select_target") ||
		!browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=sessions") {
		t.Fatalf("expected redirect review guidance in recommended actions, got %#v", payload.SessionBinding.Coordination)
	}
}

func TestRegisterBrowserTools_RuntimeWorkbenchRedirectReviewWithoutLiveNodePrefersReset(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench-redirect-stale-node")

	tracked := sessionRegistry.TrackTabs("browser-runtime-workbench-redirect-stale-node", []BrowserSessionTarget{
		{TabIndex: 1, URL: "https://node.example/home", Title: "Home", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
		{TabIndex: 2, URL: "https://redirected.example/dashboard", Title: "Redirected", BrowserApp: "Chromium", Backend: "proxy", Profile: "isolated", Target: "node"},
	}, 1)
	if len(tracked) != 2 {
		t.Fatalf("expected tracked tabs, got %#v", tracked)
	}
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-runtime-workbench-redirect-stale-node", agentxbrowserruntime.BrowserSessionRoute{
		Backend:    "proxy",
		Profile:    "isolated",
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
		Decision:   "session_target_redirect_review_required",
		Reason:     "redirect review",
	})
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-workbench-redirect-stale-node", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "stopped",
	})

	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "stopped",
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "stopped"},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"workbench","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench redirect stale node: %v", err)
	}
	var payload struct {
		WorkbenchPrimaryBrowserAction      string   `json:"workbench_primary_browser_action"`
		WorkbenchNextStep                  string   `json:"workbench_next_step"`
		WorkbenchRecommendedBrowserActions []string `json:"workbench_recommended_browser_actions"`
		SessionBinding                     struct {
			Coordination struct {
				State string `json:"state"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime workbench redirect stale node output: %v", err)
	}
	if payload.SessionBinding.Coordination.State != "browser_attached" {
		t.Fatalf("expected browser_attached stale node posture, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.WorkbenchPrimaryBrowserAction != "browser_runtime action=clear_session" || payload.WorkbenchNextStep != "browser_runtime action=clear_session" {
		t.Fatalf("expected stale node redirect review to prefer reset, got %#v", payload)
	}
	if !browserStringSliceContains(payload.WorkbenchRecommendedBrowserActions, "browser_runtime action=clear_session") ||
		!browserStringSliceContains(payload.WorkbenchRecommendedBrowserActions, "browser_runtime action=coordinate coordination_goal=ensure") ||
		browserStringSliceContains(payload.WorkbenchRecommendedBrowserActions, "browser action=pin_target") {
		t.Fatalf("expected stale node redirect guidance to prioritize reset/ensure without live follow actions, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsStaleRouteRecoveryHints(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-session-stale-route-recovery")

	sessionRegistry.TrackTab("browser-runtime-session-stale-route-recovery", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/home",
		Title:      "Home",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-session-stale-route-recovery", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "stopped",
	})

	backend := &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		NodeBackend:          backend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions stale route recovery: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			Coordination struct {
				State                     string   `json:"state"`
				PrimaryBrowserAction      string   `json:"primary_browser_action"`
				NextStep                  string   `json:"next_step"`
				RecommendedBrowserActions []string `json:"recommended_browser_actions"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime sessions stale route recovery output: %v", err)
	}
	if payload.SessionBinding.Coordination.State != "browser_attached" {
		t.Fatalf("expected browser_attached stale route posture, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser_runtime action=clear_session" || payload.SessionBinding.Coordination.NextStep != "browser_runtime action=clear_session" {
		t.Fatalf("expected stale route recovery to prefer reset, got %#v", payload.SessionBinding.Coordination)
	}
	if !browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=clear_session") ||
		!browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=coordinate coordination_goal=ensure") {
		t.Fatalf("expected stale route recovery guidance in recommended actions, got %#v", payload.SessionBinding.Coordination)
	}
}

func TestRegisterBrowserTools_RuntimeWorkbenchStaleRouteRecoveryHints(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench-stale-route-recovery")

	sessionRegistry.TrackTab("browser-runtime-workbench-stale-route-recovery", BrowserSessionTarget{
		TabIndex:   1,
		URL:        "https://node.example/home",
		Title:      "Home",
		BrowserApp: "Chromium",
		Backend:    "proxy",
		Profile:    "isolated",
		Target:     "node",
	}, true)
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-workbench-stale-route-recovery", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "stopped",
	})
	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "stopped",
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "stopped"},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionRegistry:      sessionRegistry,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"workbench","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench stale route recovery: %v", err)
	}
	var payload struct {
		WorkbenchPrimaryBrowserAction      string   `json:"workbench_primary_browser_action"`
		WorkbenchNextStep                  string   `json:"workbench_next_step"`
		WorkbenchRecommendedBrowserActions []string `json:"workbench_recommended_browser_actions"`
		SessionBinding                     struct {
			Coordination struct {
				State string `json:"state"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_runtime workbench stale route recovery output: %v", err)
	}
	if payload.SessionBinding.Coordination.State != "browser_attached" {
		t.Fatalf("expected browser_attached stale route posture, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.WorkbenchPrimaryBrowserAction != "browser_runtime action=clear_session" || payload.WorkbenchNextStep != "browser_runtime action=clear_session" {
		t.Fatalf("expected workbench stale route recovery to prefer reset, got %#v", payload)
	}
	if !browserStringSliceContains(payload.WorkbenchRecommendedBrowserActions, "browser_runtime action=clear_session") ||
		!browserStringSliceContains(payload.WorkbenchRecommendedBrowserActions, "browser_runtime action=coordinate coordination_goal=ensure") {
		t.Fatalf("expected workbench stale route recovery guidance in recommended actions, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsProfileDisconnectedHealthHints(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profile-disconnected")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profile-disconnected", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       true,
		Connected:     false,
		Note:          "cdp transport closed",
	})

	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}},
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions profile disconnected: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			SessionHealthState          string `json:"session_health_state"`
			SessionHealthReason         string `json:"session_health_reason"`
			SessionHealthRecoveryAction string `json:"session_health_recovery_action"`
			Coordination                struct {
				State                     string   `json:"state"`
				PrimaryBrowserAction      string   `json:"primary_browser_action"`
				NextStep                  string   `json:"next_step"`
				RecommendedBrowserActions []string `json:"recommended_browser_actions"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.SessionBinding.SessionHealthState != "profile_disconnected" || !strings.Contains(payload.SessionBinding.SessionHealthReason, "disconnected") || payload.SessionBinding.SessionHealthRecoveryAction != "browser_runtime action=refresh" {
		t.Fatalf("expected profile_disconnected health posture, got %#v", payload.SessionBinding)
	}
	if payload.SessionBinding.Coordination.State != "browser_ready" {
		t.Fatalf("expected browser_ready coordination state, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser_runtime action=refresh" || payload.SessionBinding.Coordination.NextStep != "browser_runtime action=refresh" {
		t.Fatalf("expected disconnected profile to prefer refresh, got %#v", payload.SessionBinding.Coordination)
	}
	if !browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=refresh") {
		t.Fatalf("expected refresh to appear in recommended browser actions, got %#v", payload.SessionBinding.Coordination)
	}
}

func TestRegisterBrowserTools_RuntimeWorkbenchProfileDisconnectedHealthHints(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-workbench-profile-disconnected")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-workbench-profile-disconnected", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "disconnected",
		Running:       true,
		Connected:     false,
		Note:          "cdp transport closed",
	})

	nodeBackend := &runtimeControlBrowserBackend{runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			runtimeStatusResult: BrowserProfileStatusResult{
				Backend:    "proxy",
				BrowserApp: "Chromium",
				Profile:    "isolated",
				Status:     "disconnected",
				Running:    true,
				Connected:  false,
				Note:       "cdp transport closed",
			},
			runtimeProfilesResult: BrowserProfilesResult{
				Backend:        "proxy",
				DefaultProfile: "isolated",
				Profiles: []BrowserProfileInfo{
					{Profile: "isolated", BrowserApp: "Chromium", Status: "disconnected", Running: true, Connected: false, Note: "cdp transport closed"},
				},
			},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          nodeBackend,
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"workbench","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime workbench profile disconnected: %v", err)
	}
	var payload struct {
		WorkbenchPrimaryBrowserAction      string   `json:"workbench_primary_browser_action"`
		WorkbenchNextStep                  string   `json:"workbench_next_step"`
		WorkbenchRecommendedBrowserActions []string `json:"workbench_recommended_browser_actions"`
		SessionBinding                     struct {
			SessionHealthState          string `json:"session_health_state"`
			SessionHealthRecoveryAction string `json:"session_health_recovery_action"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.SessionBinding.SessionHealthState != "profile_disconnected" || payload.SessionBinding.SessionHealthRecoveryAction != "browser_runtime action=refresh" {
		t.Fatalf("expected disconnected session health on workbench payload, got %#v", payload.SessionBinding)
	}
	if payload.WorkbenchPrimaryBrowserAction != "browser_runtime action=refresh" || payload.WorkbenchNextStep != "browser_runtime action=refresh" {
		t.Fatalf("expected workbench to prefer refresh for disconnected profile, got %#v", payload)
	}
	if !browserStringSliceContains(payload.WorkbenchRecommendedBrowserActions, "browser_runtime action=refresh") {
		t.Fatalf("expected refresh in workbench recommended actions, got %#v", payload)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsProfileStoppedHealthHints(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-profile-stopped")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-profile-stopped", agentxbrowserruntime.SharedSessionBrowserProfileState{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node",
		BrowserApp:    "Chromium",
		Status:        "stopped",
		Running:       false,
		Connected:     false,
	})

	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                 t.TempDir(),
		Backend:              &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"}},
		NodeBackend:          &runtimeInfoBrowserBackend{fakeBrowserBackend: &fakeBrowserBackend{}, runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"}},
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node","profile":"isolated"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions profile stopped: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			SessionHealthState          string `json:"session_health_state"`
			SessionHealthRecoveryAction string `json:"session_health_recovery_action"`
			Coordination                struct {
				State                string `json:"state"`
				PrimaryBrowserAction string `json:"primary_browser_action"`
				NextStep             string `json:"next_step"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.SessionBinding.SessionHealthState != "profile_stopped" || payload.SessionBinding.SessionHealthRecoveryAction != "browser_runtime action=coordinate coordination_goal=ensure" {
		t.Fatalf("expected profile_stopped health posture, got %#v", payload.SessionBinding)
	}
	if payload.SessionBinding.Coordination.State != "browser_attached" || payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser_runtime action=coordinate coordination_goal=ensure" || payload.SessionBinding.Coordination.NextStep != "browser_runtime action=coordinate coordination_goal=ensure" {
		t.Fatalf("expected stopped profile to prefer ensure, got %#v", payload.SessionBinding.Coordination)
	}
}

func TestRegisterBrowserTools_RuntimeSessionsBrowserReadyControlHints(t *testing.T) {
	reg := llmxtools.NewRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-runtime-browser-ready")
	sessionStateRegistry := NewBrowserSessionStateRegistry()
	sessionStateRegistry.RecordBrowserProfileState("browser-runtime-browser-ready", agentxbrowserruntime.SharedSessionBrowserProfileState{
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
		SessionStateRegistry: sessionStateRegistry,
		EnabledTools:         []string{"browser_runtime"},
	})

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions browser_ready: %v", err)
	}
	var payload struct {
		SessionBinding struct {
			Coordination struct {
				RecommendedBrowserActions []string `json:"recommended_browser_actions"`
				SyncBrowserAction         string   `json:"sync_browser_action"`
				State                     string   `json:"state"`
				PrepareBrowserAction      string   `json:"prepare_browser_action"`
				RestartBrowserAction      string   `json:"restart_browser_action"`
				TeardownBrowserAction     string   `json:"teardown_browser_action"`
				PrimaryBrowserAction      string   `json:"primary_browser_action"`
				PrimaryNodeAction         string   `json:"primary_node_action"`
				NextStep                  string   `json:"next_step"`
			} `json:"coordination"`
		} `json:"session_binding"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.SessionBinding.Coordination.State != "browser_ready" {
		t.Fatalf("expected browser_ready coordination state, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.SessionBinding.Coordination.SyncBrowserAction != "browser_runtime action=coordinate coordination_goal=sync" {
		t.Fatalf("expected coordinate sync action hint for browser_ready coordination, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.SessionBinding.Coordination.PrepareBrowserAction != "" {
		t.Fatalf("expected no prepare action when browser profile is already running, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.SessionBinding.Coordination.RestartBrowserAction != "browser_runtime action=refresh" {
		t.Fatalf("expected restart action hint for browser_ready coordination, got %#v", payload.SessionBinding.Coordination)
	}
	if !browserStringSliceContains(payload.SessionBinding.Coordination.RecommendedBrowserActions, "browser_runtime action=refresh") {
		t.Fatalf("expected refresh to appear in browser_ready recommended actions, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.SessionBinding.Coordination.TeardownBrowserAction != "browser_runtime action=coordinate coordination_goal=teardown" {
		t.Fatalf("expected teardown action hint for idle browser profile, got %#v", payload.SessionBinding.Coordination)
	}
	if payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser_runtime action=coordinate coordination_goal=sync" || payload.SessionBinding.Coordination.PrimaryNodeAction != "nodes action=run" || payload.SessionBinding.Coordination.NextStep != "browser_runtime action=coordinate coordination_goal=sync" {
		t.Fatalf("unexpected browser_ready primary actions, got %#v", payload.SessionBinding.Coordination)
	}
}
