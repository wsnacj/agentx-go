package tools

import (
	"context"
	"encoding/json"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ActDialogArmsNextModal(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			dialogResult: BrowserDialogResult{
				Backend:    "proxy-dialog",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/form",
				Title:      "Confirm Upload",
				Status:     "armed",
			},
		},
		capabilities: BrowserCapabilities{
			Dialog: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"dialog","tab_index":2,"action":"accept","prompt_text":"Ada","force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act dialog: %v", err)
	}
	if len(backend.dialogReqs) != 1 {
		t.Fatalf("expected one dialog request, got %#v", backend.dialogReqs)
	}
	req := backend.dialogReqs[0]
	if req.Action != "accept" || req.PromptText != "Ada" || req.TabIndex != 2 {
		t.Fatalf("unexpected dialog request: %#v", req)
	}
	var payload struct {
		Kind           string                                `json:"kind"`
		Action         string                                `json:"action"`
		Backend        string                                `json:"backend"`
		FinalURL       string                                `json:"final_url"`
		Status         string                                `json:"status"`
		Force          bool                                  `json:"force"`
		ReviewDecision string                                `json:"review_decision"`
		ReviewReady    bool                                  `json:"review_ready"`
		Explanation    *browserTopLevelSummary               `json:"explanation"`
		Diagnostics    *browserTopLevelSummary               `json:"diagnostics"`
		Summary        *browserTopLevelSummary               `json:"summary"`
		Display        *browserTopLevelDisplaySummary        `json:"display"`
		Surface        *browserTopLevelSurfaceSummary        `json:"surface"`
		View           *browserTopLevelViewSummary           `json:"view"`
		DiagExplain    *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "dialog" || payload.Action != "accept" || payload.Backend != "proxy-dialog" || payload.FinalURL != "https://93.184.216.34/form" || payload.Status != "armed" || !payload.Force || payload.ReviewDecision != "dialog_review_confirmed" || !payload.ReviewReady {
		t.Fatalf("unexpected browser_act dialog payload: %#v", payload)
	}
	if payload.DiagExplain == nil || payload.DiagExplain.Category != "interaction" || payload.DiagExplain.State != "started" || payload.DiagExplain.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog diagnostics explanation: %#v", payload.DiagExplain)
	}
	if payload.Explanation == nil || payload.Explanation.Category != "interaction" || payload.Explanation.State != "started" || payload.Explanation.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil || payload.Diagnostics.Category != "interaction" || payload.Diagnostics.State != "started" || payload.Diagnostics.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil || payload.Summary.Category != "interaction" || payload.Summary.State != "started" || payload.Summary.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "interaction" || payload.Display.State != "started" || payload.Display.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "interaction" || payload.Surface.State != "started" || payload.Surface.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "review" || payload.View.Category != "interaction" || payload.View.State != "started" || payload.View.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_ActDialogAcceptRequiresReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Dialog: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"dialog","tab_index":2,"action":"accept","prompt_text":"Ada"}`,
	})
	if err != nil {
		t.Fatalf("browser_act dialog review gate: %v", err)
	}
	if len(backend.dialogReqs) != 0 {
		t.Fatalf("expected dialog accept to be blocked before backend call, got %#v", backend.dialogReqs)
	}
	var payload struct {
		Kind           string `json:"kind"`
		Action         string `json:"action"`
		Status         string `json:"status"`
		ReviewDecision string `json:"review_decision"`
		ReviewReady    bool   `json:"review_ready"`
		Force          bool   `json:"force"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "dialog" || payload.Action != "accept" || payload.Status != "review_required" || payload.ReviewDecision != "dialog_review_required" || payload.ReviewReady || payload.Force {
		t.Fatalf("unexpected browser_act dialog review payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActDialogDismissDoesNotRequireReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			dialogResult: BrowserDialogResult{
				Backend:    "proxy-dialog",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/form",
				Title:      "Confirm Upload",
				Status:     "armed",
			},
		},
		capabilities: BrowserCapabilities{
			Dialog: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"dialog","tab_index":2,"action":"dismiss"}`,
	})
	if err != nil {
		t.Fatalf("browser_act dialog dismiss: %v", err)
	}
	if len(backend.dialogReqs) != 1 {
		t.Fatalf("expected one dialog dismiss request, got %#v", backend.dialogReqs)
	}
	req := backend.dialogReqs[0]
	if req.Action != "dismiss" || req.PromptText != "" || req.TabIndex != 2 {
		t.Fatalf("unexpected dismiss dialog request: %#v", req)
	}
	var payload struct {
		Kind           string                                `json:"kind"`
		Action         string                                `json:"action"`
		Status         string                                `json:"status"`
		Force          bool                                  `json:"force"`
		ReviewDecision string                                `json:"review_decision"`
		ReviewReady    bool                                  `json:"review_ready"`
		Explanation    *browserTopLevelSummary               `json:"explanation"`
		Diagnostics    *browserTopLevelSummary               `json:"diagnostics"`
		Summary        *browserTopLevelSummary               `json:"summary"`
		Display        *browserTopLevelDisplaySummary        `json:"display"`
		Surface        *browserTopLevelSurfaceSummary        `json:"surface"`
		View           *browserTopLevelViewSummary           `json:"view"`
		DiagExplain    *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "dialog" || payload.Action != "dismiss" || payload.Status != "armed" || payload.Force || payload.ReviewDecision != "" || payload.ReviewReady {
		t.Fatalf("unexpected browser_act dialog dismiss payload: %#v", payload)
	}
	if payload.DiagExplain == nil || payload.DiagExplain.Category != "interaction" || payload.DiagExplain.State != "started" || payload.DiagExplain.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog dismiss diagnostics explanation: %#v", payload.DiagExplain)
	}
	if payload.Explanation == nil || payload.Explanation.Category != "interaction" || payload.Explanation.State != "started" || payload.Explanation.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog dismiss explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil || payload.Diagnostics.Category != "interaction" || payload.Diagnostics.State != "started" || payload.Diagnostics.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog dismiss diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil || payload.Summary.Category != "interaction" || payload.Summary.State != "started" || payload.Summary.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog dismiss summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "interaction" || payload.Display.State != "started" || payload.Display.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog dismiss display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "interaction" || payload.Surface.State != "started" || payload.Surface.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog dismiss surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "interaction" || payload.View.State != "started" || payload.View.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog dismiss view: %#v", payload.View)
	}
}
