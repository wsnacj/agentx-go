package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ActEvaluateTruncates(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		evalResult: BrowserEvalResult{Backend: "fake-eval", BrowserApp: "Safari", Result: strings.Repeat("x", 60), Status: "evaluated"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                         t.TempDir(),
		Backend:                      backend,
		MaxChars:                     40,
		BrowserCDPEscapeHatchAllowed: boolPtr(true),
		EnabledTools:                 []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"evaluate","tab_index":2,"script":"document.title","max_chars":18}`,
	})
	if err != nil {
		t.Fatalf("browser_act evaluate: %v", err)
	}
	if len(backend.evalReqs) != 1 || backend.evalReqs[0].Script != "document.title" || backend.evalReqs[0].MaxChars != 18 || backend.evalReqs[0].TabIndex != 2 || backend.evalReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs {
		t.Fatalf("unexpected browser_act eval dispatch: %#v", backend.evalReqs)
	}
	var payload struct {
		Kind        string                    `json:"kind"`
		Result      string                    `json:"result"`
		Truncated   bool                      `json:"truncated"`
		TabIndex    int                       `json:"tab_index"`
		SafetyEvent *browserResultSafetyEvent `json:"safety_event"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "evaluate" || len(payload.Result) != 18 || !payload.Truncated || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act evaluate output: %#v", payload)
	}
	if payload.SafetyEvent == nil ||
		payload.SafetyEvent.EventCode != "browser_cdp_escape_hatch_allowed" ||
		payload.SafetyEvent.Decision != "allowed" ||
		payload.SafetyEvent.Source != "browser_act" ||
		payload.SafetyEvent.Action != "browser_act kind=evaluate" ||
		payload.SafetyEvent.Policy != browserCDPEscapeHatchPolicyName ||
		!payload.SafetyEvent.PolicyConfigured {
		t.Fatalf("expected browser_act cdp safety event, got %#v", payload.SafetyEvent)
	}
}

func TestRegisterBrowserTools_ActEvaluateBlocksCDPEscapeHatchWhenPolicyDenied(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		evalResult: BrowserEvalResult{Backend: "fake-eval", BrowserApp: "Safari", Result: "ok", Status: "evaluated"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                         t.TempDir(),
		Backend:                      backend,
		BrowserCDPEscapeHatchAllowed: boolPtr(false),
		EnabledTools:                 []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"evaluate","script":"document.title"}`,
	})
	if err == nil || !strings.Contains(err.Error(), "browser_act: cdp_escape_hatch_blocked") {
		t.Fatalf("expected browser_act cdp escape hatch policy denial, got %v", err)
	}
	if len(backend.evalReqs) != 0 {
		t.Fatalf("expected browser_act cdp policy denial to block backend dispatch, got %#v", backend.evalReqs)
	}
}

func TestRegisterBrowserTools_ActEvaluateRequiresPendingPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		evalResult: BrowserEvalResult{Backend: "fake-eval", BrowserApp: "Safari", FinalURL: "https://popup.example/offer", Result: "ok", Status: "evaluated"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		MaxChars:        40,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-evaluate-popup-review")
	popup := seedBrowserPendingPopupReview(sessionRegistry, "browser-act-evaluate-popup-review")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"evaluate","target":"tab:3","script":"document.title","max_chars":24}`,
	})
	if err != nil {
		t.Fatalf("browser_act evaluate popup review: %v", err)
	}
	if len(backend.evalReqs) != 0 {
		t.Fatalf("expected browser_act evaluate popup review to block before backend dispatch, got %#v", backend.evalReqs)
	}
	var payload struct {
		Kind           string `json:"kind"`
		Status         string `json:"status"`
		Force          bool   `json:"force"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		Target         string `json:"target"`
		TargetID       string `json:"target_id"`
		TabIndex       int    `json:"tab_index"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act evaluate popup review output: %v", err)
	}
	if payload.Kind != "evaluate" || payload.Status != "review_required" || payload.Force || payload.ReviewDecision != "session_target_popup_review_required" || payload.ReviewReady || payload.Target != "tab:3" || payload.TargetID != popup.ID || payload.TabIndex != 3 || !strings.Contains(payload.Note, "pending popup target") {
		t.Fatalf("unexpected browser_act evaluate popup review payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActEvaluateConfirmsPendingPopupWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		evalResult: BrowserEvalResult{Backend: "fake-eval", BrowserApp: "Safari", FinalURL: "https://popup.example/offer", Result: "forced result", Status: "evaluated"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		MaxChars:        40,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-evaluate-popup-force")
	popup := seedBrowserPendingPopupReview(sessionRegistry, "browser-act-evaluate-popup-force")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"evaluate","target":"tab:3","script":"document.title","max_chars":24,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act evaluate popup force: %v", err)
	}
	if len(backend.evalReqs) != 1 || backend.evalReqs[0].Script != "document.title" || backend.evalReqs[0].TabIndex != 3 || backend.evalReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs {
		t.Fatalf("unexpected browser_act evaluate popup force dispatch: %#v", backend.evalReqs)
	}
	var payload struct {
		Kind           string `json:"kind"`
		Status         string `json:"status"`
		Force          bool   `json:"force"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		Target         string `json:"target"`
		TargetID       string `json:"target_id"`
		TabIndex       int    `json:"tab_index"`
		FinalURL       string `json:"final_url"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act evaluate popup force output: %v", err)
	}
	if payload.Kind != "evaluate" || payload.Status != "evaluated" || !payload.Force || payload.ReviewDecision != "session_target_popup_review_confirmed" || !payload.ReviewReady || payload.Target != "tab:3" || payload.TargetID != popup.ID || payload.TabIndex != 3 || payload.FinalURL != "https://popup.example/offer" || !strings.Contains(payload.Note, "force=true") {
		t.Fatalf("unexpected browser_act evaluate popup force payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActRejectsInvalidTarget(t *testing.T) {
	reg := llmxtools.NewRegistry()
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      &fakeBrowserBackend{},
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"evaluate","target":"tab:nope","script":"document.title"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `target must be "current", "tab:<n>", or "target:<id>"`) {
		t.Fatalf("expected invalid target error, got %v", err)
	}
}

func TestRegisterBrowserTools_ActIncludesRuntimeRoute(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	backend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "fake-open", BrowserApp: "Safari", Status: "opened"},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		NodeBackend:  backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"open","url":"` + publicExampleIPURL + `","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_act: %v", err)
	}
	var payload struct {
		Kind          string `json:"kind"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "open" || payload.Profile != "isolated" || payload.RuntimeTarget != "node" || payload.Status != "opened" {
		t.Fatalf("unexpected runtime route payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActRejectsUnsupportedRuntimeTarget(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "fake-open", BrowserApp: "Safari", Status: "opened"},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"open","url":"https://93.184.216.34","runtime_target":"node"}`,
	})
	if err == nil || !strings.Contains(err.Error(), `runtime_target "node" is unsupported`) {
		t.Fatalf("expected unsupported runtime_target error, got %v", err)
	}
}

func TestRegisterBrowserTools_OpenRoutesToConfiguredNodeBackend(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "host-open", BrowserApp: "Safari", Status: "opened"},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "node-open", BrowserApp: "Chromium", Status: "opened"},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_open"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_open",
		Arguments: `{"url":"` + publicExampleIPURL + `","profile":"isolated","runtime_target":"node"}`,
	})
	if err != nil {
		t.Fatalf("browser_open: %v", err)
	}
	if len(hostBackend.openReqs) != 0 {
		t.Fatalf("expected host backend not to be used, got %#v", hostBackend.openReqs)
	}
	if len(nodeBackend.openReqs) != 1 || nodeBackend.openReqs[0].URL != publicExampleIPURL {
		t.Fatalf("expected node backend open request, got %#v", nodeBackend.openReqs)
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Backend != "node-open" || payload.BrowserApp != "Chromium" || payload.Profile != "isolated" || payload.RuntimeTarget != "node" || payload.Status != "opened" {
		t.Fatalf("unexpected routed open output: %#v", payload)
	}
}

func TestRegisterBrowserTools_OpenDefaultsToPromotedNodeBackend(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "host-open", BrowserApp: "Safari", Status: "opened"},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "node-open", BrowserApp: "Chromium", Status: "opened"},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      hostBackend,
		NodeBackend:  nodeBackend,
		EnabledTools: []string{"browser_open"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_open",
		Arguments: `{"url":"` + publicExampleIPURL + `"}`,
	})
	if err != nil {
		t.Fatalf("browser_open: %v", err)
	}
	if len(hostBackend.openReqs) != 0 {
		t.Fatalf("expected host backend to stay unused when default route promotes to node, got %#v", hostBackend.openReqs)
	}
	if len(nodeBackend.openReqs) != 1 || nodeBackend.openReqs[0].URL != publicExampleIPURL {
		t.Fatalf("expected promoted default node backend open request, got %#v", nodeBackend.openReqs)
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Backend != "node-open" || payload.BrowserApp != "Chromium" || payload.Profile != "isolated" || payload.RuntimeTarget != "node" || payload.Status != "opened" {
		t.Fatalf("unexpected promoted default open output: %#v", payload)
	}
}

func TestRegisterBrowserTools_OpenFallsBackToLocalBackendForLocalURLWhenDefaultIsRemote(t *testing.T) {
	reg := llmxtools.NewRegistry()
	hostBackend := &runtimeInfoBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			openResult: BrowserOpenResult{Backend: "host-open", BrowserApp: "Safari", Status: "opened"},
		},
		runtimeInfo: BrowserRuntimeInfo{Backend: "system", Profile: "default", Target: "host"},
	}
	nodeBackend := &remoteTargetGuardRuntimeInfoBrowserBackend{
		runtimeInfoBrowserBackend: &runtimeInfoBrowserBackend{
			fakeBrowserBackend: &fakeBrowserBackend{
				openResult: BrowserOpenResult{Backend: "node-open", BrowserApp: "Chromium", Status: "opened"},
			},
			runtimeInfo: BrowserRuntimeInfo{Backend: "proxy", Profile: "isolated", Target: "node"},
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:              t.TempDir(),
		AllowPrivateHosts: true,
		Backend:           hostBackend,
		NodeBackend:       nodeBackend,
		EnabledTools:      []string{"browser_open"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_open",
		Arguments: `{"url":"http://127.0.0.1:3000"}`,
	})
	if err != nil {
		t.Fatalf("browser_open local fallback: %v", err)
	}
	if len(nodeBackend.openReqs) != 0 {
		t.Fatalf("expected remote node backend not to receive local URL, got %#v", nodeBackend.openReqs)
	}
	if len(hostBackend.openReqs) != 1 || hostBackend.openReqs[0].URL != "http://127.0.0.1:3000" {
		t.Fatalf("expected local URL to fallback to host backend, got %#v", hostBackend.openReqs)
	}
	var payload struct {
		Backend       string `json:"backend"`
		BrowserApp    string `json:"browser_app"`
		Profile       string `json:"profile"`
		RuntimeTarget string `json:"runtime_target"`
		Status        string `json:"status"`
		Note          string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Backend != "host-open" ||
		payload.BrowserApp != "Safari" ||
		payload.Profile != "default" ||
		payload.RuntimeTarget != "host" ||
		payload.Status != "opened" ||
		!strings.Contains(payload.Note, "route_fallback_reason="+browserRouteFallbackRemoteLocalURLReason) ||
		!strings.Contains(payload.Note, "original_runtime_target=node") ||
		!strings.Contains(payload.Note, "selected_runtime_target=host") {
		t.Fatalf("unexpected local fallback output: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActOpen(t *testing.T) {
	const publicExampleIPURL = "https://93.184.216.34"
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		openResult: BrowserOpenResult{Backend: "fake-open", BrowserApp: "Safari", Status: "opened"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"open","url":"` + publicExampleIPURL + `","wait_ms":700}`,
	})
	if err != nil {
		t.Fatalf("browser_act open: %v", err)
	}
	if len(backend.openReqs) != 1 || backend.openReqs[0].URL != publicExampleIPURL || backend.openReqs[0].WaitMs != 700 {
		t.Fatalf("unexpected browser_act open dispatch: %#v", backend.openReqs)
	}
	var payload struct {
		Kind    string `json:"kind"`
		Backend string `json:"backend"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "open" || payload.Backend != "fake-open" || payload.Status != "opened" {
		t.Fatalf("unexpected browser_act open output: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActNavigateRedirectRequiresReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		navigateResult: BrowserNavigateResult{
			Backend:    "fake-navigate",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.35/landing",
			Title:      "Redirected",
			Status:     "navigated",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"navigate","url":"https://93.184.216.34","wait_ms":700,"target":"tab:2"}`,
	})
	if err != nil {
		t.Fatalf("browser_act navigate redirect review: %v", err)
	}
	if len(backend.navigateReqs) != 1 || backend.navigateReqs[0].URL != "https://93.184.216.34" || backend.navigateReqs[0].WaitMs != 700 || backend.navigateReqs[0].TabIndex != 2 {
		t.Fatalf("unexpected browser_act navigate dispatch: %#v", backend.navigateReqs)
	}
	var payload struct {
		Kind           string `json:"kind"`
		FinalURL       string `json:"final_url"`
		Status         string `json:"status"`
		Force          bool   `json:"force"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TabIndex       int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "navigate" || payload.FinalURL != "https://93.184.216.35/landing" || payload.Status != "review_required" || payload.Force || payload.ReviewDecision != "navigate_redirect_review_required" || payload.ReviewReady || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act navigate redirect review output: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActNavigateRedirectReviewBlocksLaterExtract(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-act-navigate-redirect-pending")
	sessionRegistry.TrackTab("browser-act-navigate-redirect-pending", BrowserSessionTarget{
		TabIndex:   2,
		URL:        "https://93.184.216.34/start",
		Title:      "Start",
		BrowserApp: "Safari",
		Backend:    "system",
		Target:     "host",
	}, true)
	backend := &fakeBrowserBackend{
		navigateResult: BrowserNavigateResult{
			Backend:    "fake-navigate",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.35/landing",
			Title:      "Redirected",
			Status:     "navigated",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"navigate","url":"https://93.184.216.34","target":"tab:2"}`,
	}); err != nil {
		t.Fatalf("browser_act navigate redirect review: %v", err)
	}
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract","target":"tab:2"}`,
	})
	if err != nil {
		t.Fatalf("browser_act extract after redirect review: %v", err)
	}
	var payload struct {
		Kind           string `json:"kind"`
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TabIndex       int    `json:"tab_index"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act extract after redirect review output: %v", err)
	}
	if payload.Kind != "extract" || payload.Status != "review_required" || payload.ReviewDecision != "session_target_redirect_review_required" || payload.ReviewReady || payload.TabIndex != 2 || !strings.Contains(payload.Note, "redirected target") {
		t.Fatalf("unexpected browser_act extract payload after redirect review: %#v", payload)
	}
	if len(backend.extractReqs) != 0 {
		t.Fatalf("expected browser_act extract not to follow redirected target before review, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_ActNavigateRedirectReviewWithoutPriorSelectionBlocksImplicitFollow(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	callCtx := WithToolSessionID(context.Background(), "browser-act-navigate-redirect-implicit-follow")
	backend := &fakeBrowserBackend{
		navigateResult: BrowserNavigateResult{
			Backend:    "fake-navigate",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.35/landing",
			Title:      "Redirected",
			Status:     "navigated",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act", "browser_runtime"},
	})

	if _, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"navigate","url":"https://93.184.216.34","target":"tab:2"}`,
	}); err != nil {
		t.Fatalf("browser_act navigate redirect review without prior selection: %v", err)
	}

	sessionsOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_runtime",
		Arguments: `{"action":"sessions"}`,
	})
	if err != nil {
		t.Fatalf("browser_runtime sessions after browser_act redirect review: %v", err)
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
		t.Fatalf("decode browser_runtime sessions after browser_act redirect review output: %v", err)
	}
	if sessionsPayload.SessionBinding.PendingTargetReviewCount != 1 || sessionsPayload.SessionBinding.BlockedAutoFollowRouteCount != 1 || len(sessionsPayload.SessionRoutes) != 1 {
		t.Fatalf("expected one pending redirect review route, got %#v", sessionsPayload)
	}
	if sessionsPayload.SessionRoutes[0].CurrentTargetID != "" || sessionsPayload.SessionRoutes[0].CurrentTargetSource != "" {
		t.Fatalf("expected browser_act redirect review not to create current target selection on previously unselected route, got %#v", sessionsPayload.SessionRoutes[0])
	}
	if sessionsPayload.SessionRoutes[0].FollowPolicyState != "redirect_review_required" || !strings.Contains(sessionsPayload.SessionRoutes[0].FollowPolicyReason, "redirected target") || sessionsPayload.SessionRoutes[0].PopupPolicyState != "" {
		t.Fatalf("expected browser_act redirect follow policy without popup posture, got %#v", sessionsPayload.SessionRoutes[0])
	}
	if sessionsPayload.SessionRoutes[0].PendingTargetReview == nil || sessionsPayload.SessionRoutes[0].PendingTargetReview.Decision != "session_target_redirect_review_required" {
		t.Fatalf("expected pending redirect review in browser_act sessions payload, got %#v", sessionsPayload.SessionRoutes[0])
	}
	if len(sessionsPayload.SessionRoutes[0].Targets) != 1 || sessionsPayload.SessionRoutes[0].Targets[0].Current {
		t.Fatalf("expected redirected browser_act target to stay unselected until confirmed, got %#v", sessionsPayload.SessionRoutes[0].Targets)
	}

	extractOut, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"extract"}`,
	})
	if err != nil {
		t.Fatalf("browser_act implicit extract after redirect review: %v", err)
	}
	var extractPayload struct {
		Kind           string `json:"kind"`
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TargetID       string `json:"target_id"`
		Note           string `json:"note"`
	}
	if err := json.Unmarshal([]byte(extractOut), &extractPayload); err != nil {
		t.Fatalf("decode browser_act implicit extract after redirect review output: %v", err)
	}
	if extractPayload.Kind != "extract" || extractPayload.Status != "review_required" || extractPayload.ReviewDecision != "session_target_redirect_review_required" || extractPayload.ReviewReady || strings.TrimSpace(extractPayload.TargetID) == "" || !strings.Contains(extractPayload.Note, "redirected target") {
		t.Fatalf("unexpected browser_act implicit extract payload after redirect review: %#v", extractPayload)
	}
	if len(backend.extractReqs) != 0 {
		t.Fatalf("expected browser_act implicit extract not to follow redirected target before review, got %#v", backend.extractReqs)
	}
}

func TestRegisterBrowserTools_ActNavigateRedirectConfirmedWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		navigateResult: BrowserNavigateResult{
			Backend:    "fake-navigate",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.35/landing",
			Title:      "Redirected",
			Status:     "navigated",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"navigate","url":"https://93.184.216.34","wait_ms":700,"target":"tab:2","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act navigate redirect force: %v", err)
	}
	var payload struct {
		Kind           string `json:"kind"`
		FinalURL       string `json:"final_url"`
		Status         string `json:"status"`
		Force          bool   `json:"force"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		TabIndex       int    `json:"tab_index"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "navigate" || payload.FinalURL != "https://93.184.216.35/landing" || payload.Status != "navigated" || !payload.Force || payload.ReviewDecision != "navigate_redirect_review_confirmed" || !payload.ReviewReady || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act navigate redirect force output: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActNavigateRejectsFinalURLPolicyViolation(t *testing.T) {
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
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"navigate","url":"https://93.184.216.34","wait_ms":700,"target":"tab:2"}`,
	})
	if err == nil || !strings.Contains(err.Error(), "final_url") {
		t.Fatalf("expected browser_act final_url policy error, got %v", err)
	}
}
