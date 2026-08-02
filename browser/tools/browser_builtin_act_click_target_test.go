package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ActClickCanonicalRefIncludesElementHint(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		clickResult: BrowserClickResult{Backend: "fake-click", BrowserApp: "Safari", FinalURL: "https://93.184.216.34/search?q=agentx", Status: "clicked"},
	}
	ref := browserElementRefForSnapshotElement(BrowserSnapshotElement{
		Selector: `button[aria-label="Search"]`,
		Ref:      "e12",
		Role:     "button",
		Label:    "Search",
		Href:     "https://93.184.216.34/search?q=agentx",
	}, "https://93.184.216.34/search?q=agentx#top", "Example Search")
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		NodeBackend:  backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","target":"current","ref":"` + ref + `"}`,
	})
	if err != nil {
		t.Fatalf("browser_act click canonical ref: %v", err)
	}
	if len(backend.clickReqs) != 1 {
		t.Fatalf("expected one click request, got %#v", backend.clickReqs)
	}
	req := backend.clickReqs[0]
	if req.Selector != `button[aria-label="Search"]` || req.ElementRef != ref {
		t.Fatalf("unexpected canonical click request: %#v", req)
	}
	if req.ElementHint == nil ||
		req.ElementHint.Selector != `button[aria-label="Search"]` ||
		req.ElementHint.NativeRef != "e12" ||
		req.ElementHint.Role != "button" ||
		req.ElementHint.Label != "Search" ||
		req.ElementHint.Href != "https://93.184.216.34/search?q=agentx" ||
		req.ElementHint.PageURL != "https://93.184.216.34/search?q=agentx" ||
		req.ElementHint.PageTitle != "Example Search" {
		t.Fatalf("expected element hint from canonical ref, got %#v", req.ElementHint)
	}
	var payload struct {
		Kind     string `json:"kind"`
		Status   string `json:"status"`
		Ref      string `json:"ref"`
		Selector string `json:"selector"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "click" || payload.Status != "clicked" || payload.Ref != ref || payload.Selector != `button[aria-label="Search"]` {
		t.Fatalf("unexpected browser_act click canonical payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActClickAcceptsElementLabelHint(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		clickResult: BrowserClickResult{Backend: "fake-click", BrowserApp: "Safari", FinalURL: "https://93.184.216.34/investors", Status: "clicked"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		NodeBackend:  backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","target":"current","element":"2025 年度报告"}`,
	})
	if err != nil {
		t.Fatalf("browser_act click element label: %v", err)
	}
	if len(backend.clickReqs) != 1 || backend.clickReqs[0].Selector != "" || backend.clickReqs[0].ElementRef == "" || backend.clickReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs {
		t.Fatalf("unexpected browser_act element-label click request: %#v", backend.clickReqs)
	}
	assertBrowserClickLabelSemanticHint(t, backend.clickReqs[0].ElementHint, "2025 年度报告", "")
	var payload struct {
		Kind     string `json:"kind"`
		Status   string `json:"status"`
		Ref      string `json:"ref"`
		Selector string `json:"selector"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "click" || payload.Status != "clicked" || payload.Ref == "" || payload.Selector != "" {
		t.Fatalf("unexpected browser_act element-label click payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActClickAcceptsTextLabelHint(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		clickResult: BrowserClickResult{Backend: "fake-click", BrowserApp: "Safari", FinalURL: "https://93.184.216.34/investors", Status: "clicked"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		NodeBackend:  backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","target":"current","text":"2025 年度报告"}`,
	})
	if err != nil {
		t.Fatalf("browser_act click text label: %v", err)
	}
	if len(backend.clickReqs) != 1 || backend.clickReqs[0].Selector != "" || backend.clickReqs[0].ElementRef == "" {
		t.Fatalf("unexpected browser_act text-label click request: %#v", backend.clickReqs)
	}
	assertBrowserClickLabelSemanticHint(t, backend.clickReqs[0].ElementHint, "2025 年度报告", "")
	var payload struct {
		Kind     string `json:"kind"`
		Status   string `json:"status"`
		Ref      string `json:"ref"`
		Selector string `json:"selector"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "click" || payload.Status != "clicked" || payload.Ref == "" || payload.Selector != "" {
		t.Fatalf("unexpected browser_act text-label click payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActClickRejectsRefBoundToDifferentRequestedPage(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		clickResult: BrowserClickResult{Backend: "fake-click", BrowserApp: "Safari", FinalURL: "https://93.184.216.34/cart", Status: "clicked"},
	}
	ref := browserElementRefForSnapshotElement(BrowserSnapshotElement{
		Selector: `button[aria-label="Buy"]`,
		Label:    "Buy",
		Role:     "button",
	}, "https://93.184.216.34/search?q=agentx", "Search")
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		NodeBackend:  backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","url":"https://93.184.216.34/cart","ref":"` + ref + `"}`,
	})
	if err == nil || !strings.Contains(err.Error(), "element ref page binding differs from requested page") {
		t.Fatalf("expected requested page-bound ref error, got %v", err)
	}
	if len(backend.clickReqs) != 0 {
		t.Fatalf("expected requested page-bound ref mismatch to block backend dispatch, got %#v", backend.clickReqs)
	}
}

func TestRegisterBrowserTools_ActClickRequiresPendingPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		clickResult: BrowserClickResult{Backend: "fake-click", BrowserApp: "Safari", FinalURL: "https://popup.example/offer", Title: "Offer", Status: "clicked"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-click-popup-review")
	popup := seedBrowserPendingPopupReview(sessionRegistry, "browser-act-click-popup-review")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","target":"tab:3","selector":"button.buy"}`,
	})
	if err != nil {
		t.Fatalf("browser_act click popup review: %v", err)
	}
	if len(backend.clickReqs) != 0 {
		t.Fatalf("expected browser_act click popup review to block before backend dispatch, got %#v", backend.clickReqs)
	}
	var payload struct {
		Kind                   string                                `json:"kind"`
		Status                 string                                `json:"status"`
		Force                  bool                                  `json:"force"`
		ReviewDecision         string                                `json:"review_decision"`
		ReviewReady            bool                                  `json:"review_ready"`
		Target                 string                                `json:"target"`
		TargetID               string                                `json:"target_id"`
		TabIndex               int                                   `json:"tab_index"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Note                   string                                `json:"note"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode browser_act click popup review output: %v", err)
	}
	if payload.Kind != "click" || payload.Status != "review_required" || payload.Force || payload.ReviewDecision != "session_target_popup_review_required" || payload.ReviewReady || payload.Target != "tab:3" || payload.TargetID != popup.ID || payload.TabIndex != 3 || !strings.Contains(payload.Note, "pending popup target") {
		t.Fatalf("unexpected browser_act click popup review payload: %#v", payload)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "review" ||
		payload.DiagnosticsExplanation.State != "manual_confirmation_required" ||
		payload.DiagnosticsExplanation.SummaryCode != "popup_review_required" ||
		payload.DiagnosticsExplanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected browser_act click review diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "review" ||
		payload.Explanation.State != "manual_confirmation_required" ||
		payload.Explanation.SummaryCode != "popup_review_required" ||
		payload.Explanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected browser_act click review explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "review" ||
		payload.Diagnostics.State != "manual_confirmation_required" ||
		payload.Diagnostics.SummaryCode != "popup_review_required" ||
		payload.Diagnostics.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected browser_act click review diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "review" ||
		payload.Summary.State != "manual_confirmation_required" ||
		payload.Summary.SummaryCode != "popup_review_required" ||
		payload.Summary.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected browser_act click review summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "review" ||
		payload.Display.State != "manual_confirmation_required" ||
		payload.Display.SummaryCode != "popup_review_required" ||
		payload.Display.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected browser_act click review display: %#v", payload.Display)
	}
}

func TestRegisterBrowserTools_ActClickConfirmsPendingPopupWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		clickResult: BrowserClickResult{Backend: "fake-click", BrowserApp: "Safari", FinalURL: "https://popup.example/offer", Title: "Offer", Status: "clicked"},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-click-popup-force")
	popup := seedBrowserPendingPopupReview(sessionRegistry, "browser-act-click-popup-force")
	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"click","target":"tab:3","selector":"button.buy","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act click popup force: %v", err)
	}
	if len(backend.clickReqs) != 1 || backend.clickReqs[0].Selector != "button.buy" || backend.clickReqs[0].TabIndex != 3 || backend.clickReqs[0].WaitMs != defaultBrowserInteractiveActionWaitMs {
		t.Fatalf("unexpected browser_act click popup force dispatch: %#v", backend.clickReqs)
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
		t.Fatalf("decode browser_act click popup force output: %v", err)
	}
	if payload.Kind != "click" || payload.Status != "clicked" || !payload.Force || payload.ReviewDecision != "session_target_popup_review_confirmed" || !payload.ReviewReady || payload.Target != "tab:3" || payload.TargetID != popup.ID || payload.TabIndex != 3 || payload.FinalURL != "https://popup.example/offer" || !strings.Contains(payload.Note, "force=true") {
		t.Fatalf("unexpected browser_act click popup force payload: %#v", payload)
	}
}
