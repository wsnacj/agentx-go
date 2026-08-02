package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_EvalTruncatesResult(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		evalResult: BrowserEvalResult{Backend: "fake-eval", BrowserApp: "Safari", FinalURL: "https://93.184.216.34", Result: strings.Repeat("z", 50), Status: "evaluated"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                         t.TempDir(),
		Backend:                      backend,
		MaxChars:                     40,
		BrowserCDPEscapeHatchAllowed: boolPtr(true),
		EnabledTools:                 []string{"browser_eval"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_eval",
		Arguments: `{"tab_index":5,"script":"document.title","max_chars":24}`,
	})
	if err != nil {
		t.Fatalf("browser_eval: %v", err)
	}
	if len(backend.evalReqs) != 1 || backend.evalReqs[0].Script != "document.title" || backend.evalReqs[0].MaxChars != 24 || backend.evalReqs[0].TabIndex != 5 || backend.evalReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("unexpected eval request: %#v", backend.evalReqs)
	}
	var payload struct {
		Backend     string                    `json:"backend"`
		Result      string                    `json:"result"`
		Truncated   bool                      `json:"truncated"`
		TabIndex    int                       `json:"tab_index"`
		SafetyEvent *browserResultSafetyEvent `json:"safety_event"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Backend != "fake-eval" || len(payload.Result) != 24 || !payload.Truncated || payload.TabIndex != 5 {
		t.Fatalf("unexpected eval output: %#v", payload)
	}
	if payload.SafetyEvent == nil ||
		payload.SafetyEvent.EventCode != "browser_cdp_escape_hatch_allowed" ||
		payload.SafetyEvent.Decision != "allowed" ||
		payload.SafetyEvent.Source != "browser_eval" ||
		payload.SafetyEvent.Policy != browserCDPEscapeHatchPolicyName ||
		!payload.SafetyEvent.PolicyConfigured {
		t.Fatalf("expected browser_eval cdp safety event, got %#v", payload.SafetyEvent)
	}
}

func TestRegisterBrowserTools_EvalBlocksCDPEscapeHatchWhenPolicyDenied(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		evalResult: BrowserEvalResult{Backend: "fake-eval", BrowserApp: "Safari", Result: "ok", Status: "evaluated"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:                         t.TempDir(),
		Backend:                      backend,
		BrowserCDPEscapeHatchAllowed: boolPtr(false),
		EnabledTools:                 []string{"browser_eval"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_eval",
		Arguments: `{"script":"document.title"}`,
	})
	if err == nil || !strings.Contains(err.Error(), "browser_eval: cdp_escape_hatch_blocked") {
		t.Fatalf("expected browser_eval cdp escape hatch policy denial, got %v", err)
	}
	if len(backend.evalReqs) != 0 {
		t.Fatalf("expected browser_eval cdp policy denial to block backend dispatch, got %#v", backend.evalReqs)
	}
}

func TestRegisterBrowserTools_EvalRequiresPendingPopupReview(t *testing.T) {
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
		EnabledTools:    []string{"browser_eval"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-eval-popup-review")
	popup := seedBrowserPendingPopupReview(sessionRegistry, "browser-eval-popup-review")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_eval",
		Arguments: `{"target":"tab:3","script":"document.title","max_chars":24}`,
	})
	if err != nil {
		t.Fatalf("browser_eval popup review: %v", err)
	}
	if len(backend.evalReqs) != 0 {
		t.Fatalf("expected browser_eval popup review to block before backend dispatch, got %#v", backend.evalReqs)
	}
	var payload struct {
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
		t.Fatalf("decode browser_eval popup review output: %v", err)
	}
	if payload.Status != "review_required" || payload.Force || payload.ReviewDecision != "session_target_popup_review_required" || payload.ReviewReady || payload.Target != "tab:3" || payload.TargetID != popup.ID || payload.TabIndex != 3 || !strings.Contains(payload.Note, "pending popup target") {
		t.Fatalf("unexpected browser_eval popup review payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_EvalConfirmsPendingPopupWithForce(t *testing.T) {
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
		EnabledTools:    []string{"browser_eval"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-eval-popup-force")
	popup := seedBrowserPendingPopupReview(sessionRegistry, "browser-eval-popup-force")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_eval",
		Arguments: `{"target":"tab:3","script":"document.title","max_chars":24,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_eval popup force: %v", err)
	}
	if len(backend.evalReqs) != 1 || backend.evalReqs[0].Script != "document.title" || backend.evalReqs[0].TabIndex != 3 || backend.evalReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("unexpected browser_eval popup force dispatch: %#v", backend.evalReqs)
	}
	var payload struct {
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
		t.Fatalf("decode browser_eval popup force output: %v", err)
	}
	if payload.Status != "evaluated" || !payload.Force || payload.ReviewDecision != "session_target_popup_review_confirmed" || !payload.ReviewReady || payload.Target != "tab:3" || payload.TargetID != popup.ID || payload.TabIndex != 3 || payload.FinalURL != "https://popup.example/offer" || !strings.Contains(payload.Note, "force=true") {
		t.Fatalf("unexpected browser_eval popup force payload: %#v", payload)
	}
}
