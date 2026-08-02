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

func TestRegisterBrowserTools_ExtractRequiresPendingPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser_extract"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-extract-popup-review")
	popup := sessionRegistry.TrackTab("browser-extract-popup-review", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-extract-popup-review", BrowserSessionRoute{
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
		Name:      "browser_extract",
		Arguments: `{"target":"tab:3","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_extract popup review: %v", err)
	}
	if len(backend.extractReqs) != 0 {
		t.Fatalf("expected browser_extract popup review to block before backend dispatch, got %#v", backend.extractReqs)
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
		t.Fatalf("decode popup review output: %v", err)
	}
	if payload.Status != "review_required" || payload.Force || payload.ReviewDecision != "session_target_popup_review_required" || payload.ReviewReady || payload.Target != "tab:3" || payload.TargetID != popup.ID || payload.TabIndex != 3 || !strings.Contains(payload.Note, "pending popup target") {
		t.Fatalf("unexpected browser_extract popup review payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ExtractRequiresPopupStormReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser_extract"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-extract-popup-storm")
	popup := seedBrowserPendingPopupStormReview(sessionRegistry, "browser-extract-popup-storm")

	out, err := reg.Execute(callCtx, types.FunctionCall{
		Name:      "browser_extract",
		Arguments: `{"target":"tab:4","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_extract popup storm review: %v", err)
	}
	if len(backend.extractReqs) != 0 {
		t.Fatalf("expected browser_extract popup storm to block before backend dispatch, got %#v", backend.extractReqs)
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
		t.Fatalf("decode browser_extract popup storm output: %v", err)
	}
	if payload.Status != "review_required" || payload.Force || payload.ReviewDecision != "session_target_popup_review_required" || payload.ReviewReady || payload.Target != "tab:4" || payload.TargetID != popup.ID || payload.TabIndex != 4 || !strings.Contains(payload.Note, "accumulated 2 pending popup targets") {
		t.Fatalf("unexpected browser_extract popup storm payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ExtractConfirmsPendingPopupWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		extractResult: BrowserExtractResult{
			Backend:     "fake-extract",
			BrowserApp:  "Safari",
			Title:       "Offer",
			Content:     "popup",
			FinalURL:    "https://popup.example/offer",
			ContentType: "text/plain",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser_extract"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-extract-popup-force")
	popup := sessionRegistry.TrackTab("browser-extract-popup-force", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-extract-popup-force", BrowserSessionRoute{
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
		Name:      "browser_extract",
		Arguments: `{"target":"tab:3","max_chars":20,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_extract popup force: %v", err)
	}
	if len(backend.extractReqs) != 1 || backend.extractReqs[0].TabIndex != 3 || backend.extractReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("unexpected browser_extract popup force dispatch: %#v", backend.extractReqs)
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
		t.Fatalf("decode popup force output: %v", err)
	}
	if payload.Status != "extracted" || !payload.Force || payload.ReviewDecision != "session_target_popup_review_confirmed" || !payload.ReviewReady || payload.Target != "tab:3" || payload.TargetID != popup.ID || payload.TabIndex != 3 || payload.FinalURL != "https://popup.example/offer" || !strings.Contains(payload.Note, "force=true") {
		t.Fatalf("unexpected browser_extract popup force payload: %#v", payload)
	}
}

func seedBrowserPendingPopupReview(sessionRegistry *BrowserSessionRegistry, sessionID string) BrowserSessionTarget {
	popup := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute(sessionID, BrowserSessionRoute{
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
	return popup
}

func seedBrowserPendingPopupStormReview(sessionRegistry *BrowserSessionRegistry, sessionID string) BrowserSessionTarget {
	seedBrowserPendingPopupReview(sessionRegistry, sessionID)
	second := sessionRegistry.TrackTab(sessionID, BrowserSessionTarget{
		TabIndex:   4,
		URL:        "https://popup.example/bonus",
		Title:      "Bonus",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute(sessionID, BrowserSessionRoute{
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
		BrowserApp: "Safari",
	}, agentxbrowserruntime.BrowserSessionTargetReview{
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
	return second
}
