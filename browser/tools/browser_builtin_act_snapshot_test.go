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

func TestRegisterBrowserTools_ActSnapshot(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		snapshotResult: BrowserSnapshotResult{
			Backend:     "fake-snapshot",
			BrowserApp:  "Safari",
			FinalURL:    "https://93.184.216.34/final",
			Title:       "Example",
			Snapshot:    strings.Repeat("s", 80),
			Format:      "aria",
			Mode:        "efficient",
			Refs:        "role",
			Interactive: true,
			Compact:     true,
			Depth:       6,
			Selector:    "#main",
			Frame:       "iframe#main",
			Elements: []BrowserSnapshotElement{
				{Index: 1, Role: "button", Label: "Search", Ref: browserElementRefForSelector(`button[aria-label="Search"]`), Selector: `button[aria-label="Search"]`},
			},
			Truncated: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		MaxChars:     64,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"snapshot","tab_index":1,"max_chars":24,"max_elements":6,"format":"aria","mode":"efficient","interactive":true,"compact":true,"depth":6,"selector":"#main","frame":"iframe#main"}`,
	})
	if err != nil {
		t.Fatalf("browser_act snapshot: %v", err)
	}
	if len(backend.snapshotReqs) != 1 ||
		backend.snapshotReqs[0].URL != "" ||
		backend.snapshotReqs[0].MaxChars != 24 ||
		backend.snapshotReqs[0].MaxElements != 6 ||
		backend.snapshotReqs[0].TabIndex != 1 ||
		backend.snapshotReqs[0].WaitMs != browserTabTargetWaitMs ||
		backend.snapshotReqs[0].Format != "aria" ||
		backend.snapshotReqs[0].Mode != "efficient" ||
		backend.snapshotReqs[0].Refs != "role" ||
		!backend.snapshotReqs[0].Interactive ||
		!backend.snapshotReqs[0].Compact ||
		backend.snapshotReqs[0].Depth != 6 ||
		backend.snapshotReqs[0].Selector != "#main" ||
		backend.snapshotReqs[0].Frame != "iframe#main" {
		t.Fatalf("unexpected browser_act snapshot dispatch: %#v", backend.snapshotReqs)
	}
	var payload struct {
		Kind                string                         `json:"kind"`
		Backend             string                         `json:"backend"`
		FinalURL            string                         `json:"final_url"`
		Snapshot            string                         `json:"snapshot"`
		SnapshotFormat      string                         `json:"snapshot_format"`
		SnapshotMode        string                         `json:"snapshot_mode"`
		SnapshotRefs        string                         `json:"snapshot_refs"`
		SnapshotInteractive bool                           `json:"snapshot_interactive"`
		SnapshotCompact     bool                           `json:"snapshot_compact"`
		SnapshotDepth       int                            `json:"snapshot_depth"`
		SnapshotFrame       string                         `json:"snapshot_frame"`
		Selector            string                         `json:"selector"`
		Elements            []BrowserSnapshotElement       `json:"elements"`
		Truncated           bool                           `json:"truncated"`
		Status              string                         `json:"status"`
		TabIndex            int                            `json:"tab_index"`
		Summary             *browserTopLevelSummary        `json:"summary"`
		Display             *browserTopLevelDisplaySummary `json:"display"`
		Surface             *browserTopLevelSurfaceSummary `json:"surface"`
		View                *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "snapshot" ||
		payload.Backend != "fake-snapshot" ||
		payload.FinalURL != "https://93.184.216.34/final" ||
		len(payload.Snapshot) != 24 ||
		payload.SnapshotFormat != "aria" ||
		payload.SnapshotMode != "efficient" ||
		payload.SnapshotRefs != "role" ||
		!payload.SnapshotInteractive ||
		!payload.SnapshotCompact ||
		payload.SnapshotDepth != 6 ||
		payload.SnapshotFrame != "iframe#main" ||
		payload.Selector != "#main" ||
		len(payload.Elements) != 1 ||
		!payload.Truncated ||
		payload.Status != "snapshotted" ||
		payload.TabIndex != 1 {
		t.Fatalf("unexpected browser_act snapshot output: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "content" || payload.Summary.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected browser_act snapshot summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "content" || payload.Display.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected browser_act snapshot display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "content" || payload.Surface.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected browser_act snapshot surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "content" || payload.View.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected browser_act snapshot view: %#v", payload.View)
	}
	if payload.Elements[0].Ref == "" {
		t.Fatalf("expected snapshot element ref, got %#v", payload.Elements[0])
	}
}

func TestRegisterBrowserTools_ActSnapshotNormalizesElementRefs(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		snapshotResult: BrowserSnapshotResult{
			Backend:    "fake-snapshot",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.34/search?q=agentx#top",
			Title:      "Example Search",
			Snapshot:   "snapshot",
			Elements: []BrowserSnapshotElement{
				{Role: "button", Label: "Search", Ref: "e12", Selector: `button[aria-label="Search"]`},
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
		Arguments: `{"kind":"snapshot","tab_index":1}`,
	})
	if err != nil {
		t.Fatalf("browser_act snapshot normalize refs: %v", err)
	}
	var payload struct {
		Elements []BrowserSnapshotElement `json:"elements"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if len(payload.Elements) != 1 {
		t.Fatalf("expected one snapshot element, got %#v", payload.Elements)
	}
	if payload.Elements[0].Index != 1 || payload.Elements[0].Ref == "" {
		t.Fatalf("expected normalized snapshot element index/ref, got %#v", payload.Elements[0])
	}
	if payload.Elements[0].Ref == "e12" {
		t.Fatalf("expected backend-native snapshot ref to be canonicalized, got %#v", payload.Elements[0])
	}
	refPayload, err := browserDecodeElementRef(payload.Elements[0].Ref)
	if err != nil {
		t.Fatalf("decode normalized snapshot ref: %v", err)
	}
	if refPayload.PageURL != "https://93.184.216.34/search?q=agentx" || refPayload.PageTitle != "Example Search" || refPayload.Selector != `button[aria-label="Search"]` {
		t.Fatalf("unexpected normalized snapshot ref payload: %#v", refPayload)
	}
}

func TestRegisterBrowserTools_ActSnapshotAcceptsQuotedRefsTokens(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		snapshotResult: BrowserSnapshotResult{
			Backend:    "fake-snapshot",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.34/final",
			Title:      "Example",
			Snapshot:   "snapshot",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: "{\"kind\":\"snapshot\",\"tab_index\":1,\"refs\":\" \\n\\t\\\"role\\\"\\t\",\"format\":\"\\n\\\"ai\\\"\",\"mode\":\"\\n\\\"efficient\\\"\"}",
	})
	if err != nil {
		t.Fatalf("browser_act snapshot quoted refs tokens: %v", err)
	}
	var payload struct {
		SnapshotRefs   string `json:"snapshot_refs"`
		SnapshotFormat string `json:"snapshot_format"`
		SnapshotMode   string `json:"snapshot_mode"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.SnapshotRefs != "role" || payload.SnapshotFormat != "ai" || payload.SnapshotMode != "efficient" {
		t.Fatalf("unexpected normalized snapshot tokens: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActSnapshotCanonicalizedRefDrivesLaterClick(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &fakeBrowserBackend{
		snapshotResult: BrowserSnapshotResult{
			Backend:    "fake-snapshot",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.34/search?q=agentx#top",
			Title:      "Example Search",
			Snapshot:   "snapshot",
			Elements: []BrowserSnapshotElement{
				{Role: "button", Label: "Search", Ref: "e12", Selector: `button[aria-label="Search"]`},
			},
		},
		clickResult: BrowserClickResult{
			Backend:    "fake-click",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.34/search?q=agentx",
			Status:     "clicked",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	snapshotOut, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"snapshot","target":"current"}`,
	})
	if err != nil {
		t.Fatalf("browser_act snapshot before click: %v", err)
	}
	var snapshotPayload struct {
		Elements []BrowserSnapshotElement `json:"elements"`
	}
	if err := json.Unmarshal([]byte(snapshotOut), &snapshotPayload); err != nil {
		t.Fatalf("decode snapshot output: %v", err)
	}
	if len(snapshotPayload.Elements) != 1 || snapshotPayload.Elements[0].Ref == "" {
		t.Fatalf("expected canonicalized snapshot element, got %#v", snapshotPayload.Elements)
	}
	if snapshotPayload.Elements[0].Ref == "e12" {
		t.Fatalf("expected canonicalized snapshot ref, got %#v", snapshotPayload.Elements[0])
	}

	clickArgs, err := json.Marshal(map[string]any{
		"kind":   "click",
		"target": "current",
		"ref":    snapshotPayload.Elements[0].Ref,
	})
	if err != nil {
		t.Fatalf("marshal click args: %v", err)
	}
	clickOut, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: string(clickArgs),
	})
	if err != nil {
		t.Fatalf("browser_act click canonicalized ref: %v", err)
	}
	if len(backend.clickReqs) != 1 {
		t.Fatalf("expected click dispatch, got %#v", backend.clickReqs)
	}
	if backend.clickReqs[0].ElementRef != snapshotPayload.Elements[0].Ref || backend.clickReqs[0].Selector != `button[aria-label="Search"]` {
		t.Fatalf("expected click to use canonicalized ref and selector, got %#v", backend.clickReqs[0])
	}
	var clickPayload struct {
		Kind     string `json:"kind"`
		Status   string `json:"status"`
		Ref      string `json:"ref"`
		Selector string `json:"selector"`
	}
	if err := json.Unmarshal([]byte(clickOut), &clickPayload); err != nil {
		t.Fatalf("decode click output: %v", err)
	}
	if clickPayload.Kind != "click" || clickPayload.Status != "clicked" || clickPayload.Ref != snapshotPayload.Elements[0].Ref || clickPayload.Selector != `button[aria-label="Search"]` {
		t.Fatalf("unexpected click payload after canonicalized snapshot ref: %#v", clickPayload)
	}
}

func TestRegisterBrowserTools_ActSnapshotUsesTrackedSafariBrowserApp(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	sessionID := "browser-act-snapshot-tracked-safari"
	tracked := sessionRegistry.TrackCurrentTarget(sessionID, BrowserSessionTarget{
		URL:        "https://93.184.216.34",
		Title:      "Example",
		BrowserApp: "Safari",
		Backend:    "safari_snapshot",
		Profile:    "default",
		Target:     "host",
	})
	backend := &fakeBrowserBackend{
		snapshotResult: BrowserSnapshotResult{
			Backend:    "fake-snapshot",
			BrowserApp: "Safari",
			FinalURL:   "https://93.184.216.34",
			Title:      "Example",
			Snapshot:   "snapshot body",
			Format:     "ai",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser_act"},
	})

	_, err := reg.Execute(WithToolSessionID(context.Background(), sessionID), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"snapshot","target":"target:` + tracked.ID + `","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_act snapshot with tracked safari target: %v", err)
	}
	if len(backend.snapshotReqs) != 1 {
		t.Fatalf("expected one snapshot request, got %#v", backend.snapshotReqs)
	}
	if backend.snapshotReqs[0].BrowserApp != "Safari" || backend.snapshotReqs[0].URL != "https://93.184.216.34" {
		t.Fatalf("expected tracked safari target to infer browser app and url, got %#v", backend.snapshotReqs[0])
	}
}

func TestRegisterBrowserTools_ActSnapshotRequiresPendingPopupReview(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		snapshotResult: BrowserSnapshotResult{
			Backend:    "fake-snapshot",
			BrowserApp: "Safari",
			FinalURL:   "https://popup.example/offer",
			Title:      "Offer",
			Snapshot:   "popup snapshot",
			Format:     "aria",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-snapshot-popup-review")
	popup := sessionRegistry.TrackTab("browser-act-snapshot-popup-review", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-act-snapshot-popup-review", BrowserSessionRoute{
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
		Arguments: `{"kind":"snapshot","target":"tab:3","max_chars":20}`,
	})
	if err != nil {
		t.Fatalf("browser_act snapshot popup review: %v", err)
	}
	if len(backend.snapshotReqs) != 0 {
		t.Fatalf("expected browser_act snapshot popup review to block before backend dispatch, got %#v", backend.snapshotReqs)
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
		t.Fatalf("decode browser_act snapshot popup review output: %v", err)
	}
	if payload.Kind != "snapshot" || payload.Status != "review_required" || payload.Force || payload.ReviewDecision != "session_target_popup_review_required" || payload.ReviewReady || payload.Target != "tab:3" || payload.TargetID != popup.ID || payload.TabIndex != 3 || !strings.Contains(payload.Note, "pending popup target") {
		t.Fatalf("unexpected browser_act snapshot popup review payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActSnapshotConfirmsPendingPopupWithForce(t *testing.T) {
	reg := llmxtools.NewRegistry()
	sessionRegistry := NewBrowserSessionRegistry()
	backend := &fakeBrowserBackend{
		snapshotResult: BrowserSnapshotResult{
			Backend:    "fake-snapshot",
			BrowserApp: "Safari",
			FinalURL:   "https://popup.example/offer",
			Title:      "Offer",
			Snapshot:   "popup snapshot",
			Format:     "aria",
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:            t.TempDir(),
		Backend:         backend,
		SessionRegistry: sessionRegistry,
		MaxChars:        64,
		EnabledTools:    []string{"browser_act"},
	})

	callCtx := WithToolSessionID(context.Background(), "browser-act-snapshot-popup-force")
	popup := sessionRegistry.TrackTab("browser-act-snapshot-popup-force", BrowserSessionTarget{
		TabIndex:   3,
		URL:        "https://popup.example/offer",
		Title:      "Offer",
		BrowserApp: "Safari",
		Backend:    "system",
		Profile:    "default",
		Target:     "host",
	}, false)
	sessionRegistry.RecordPendingTargetReviewForRoute("browser-act-snapshot-popup-force", BrowserSessionRoute{
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
		Arguments: `{"kind":"snapshot","target":"tab:3","max_chars":20,"force":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act snapshot popup force: %v", err)
	}
	if len(backend.snapshotReqs) != 1 || backend.snapshotReqs[0].TabIndex != 3 || backend.snapshotReqs[0].WaitMs != browserTabTargetWaitMs {
		t.Fatalf("unexpected browser_act snapshot popup force dispatch: %#v", backend.snapshotReqs)
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
		t.Fatalf("decode browser_act snapshot popup force output: %v", err)
	}
	if payload.Kind != "snapshot" || payload.Status != "snapshotted" || !payload.Force || payload.ReviewDecision != "session_target_popup_review_confirmed" || !payload.ReviewReady || payload.Target != "tab:3" || payload.TargetID != popup.ID || payload.TabIndex != 3 || payload.FinalURL != "https://popup.example/offer" || !strings.Contains(payload.Note, "force=true") {
		t.Fatalf("unexpected browser_act snapshot popup force payload: %#v", payload)
	}
}
