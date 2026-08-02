package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ActPressReturnsStatus(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			pressResult: BrowserPressResult{
				Backend:    "proxy-press",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Keyboard",
				Key:        "Enter",
				Status:     "pressed",
			},
		},
		capabilities: BrowserCapabilities{
			Press: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"press","tab_index":2,"key":"Enter","delay_ms":40}`,
	})
	if err != nil {
		t.Fatalf("browser_act press: %v", err)
	}
	if len(backend.pressReqs) != 1 {
		t.Fatalf("expected one press request, got %#v", backend.pressReqs)
	}
	req := backend.pressReqs[0]
	if req.TabIndex != 2 || req.Key != "Enter" || req.DelayMs != 40 || req.PostWaitMs != 250 {
		t.Fatalf("unexpected press request: %#v", req)
	}
	var payload struct {
		Kind     string                         `json:"kind"`
		Backend  string                         `json:"backend"`
		FinalURL string                         `json:"final_url"`
		Key      string                         `json:"key"`
		Status   string                         `json:"status"`
		TabIndex int                            `json:"tab_index"`
		Summary  *browserTopLevelSummary        `json:"summary"`
		Display  *browserTopLevelDisplaySummary `json:"display"`
		Surface  *browserTopLevelSurfaceSummary `json:"surface"`
		View     *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "press" || payload.Backend != "proxy-press" || payload.FinalURL != "https://93.184.216.34/app" || payload.Key != "Enter" || payload.Status != "pressed" || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act press payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "interaction" || payload.Summary.SummaryCode != "press_completed" {
		t.Fatalf("unexpected browser_act press summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "interaction" || payload.Display.SummaryCode != "press_completed" {
		t.Fatalf("unexpected browser_act press display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "interaction" || payload.Surface.SummaryCode != "press_completed" {
		t.Fatalf("unexpected browser_act press surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "interaction" || payload.View.SummaryCode != "press_completed" {
		t.Fatalf("unexpected browser_act press view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_ActHoverReturnsStatus(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			hoverResult: BrowserHoverResult{
				Backend:    "proxy-hover",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Hover",
				Status:     "hovered",
				ResolverOutcome: &agentxbrowserruntime.BrowserElementResolverOutcome{
					Status:         "matched",
					ResolutionMode: "selector_first",
					PrimaryKind:    "selector",
					AttemptCount:   1,
					MatchedKind:    "selector",
					MatchedIndex:   0,
				},
			},
		},
		capabilities: BrowserCapabilities{
			Hover: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"hover","tab_index":2,"selector":"button.menu"}`,
	})
	if err != nil {
		t.Fatalf("browser_act hover: %v", err)
	}
	if len(backend.hoverReqs) != 1 {
		t.Fatalf("expected one hover request, got %#v", backend.hoverReqs)
	}
	req := backend.hoverReqs[0]
	if req.TabIndex != 2 || req.Selector != "button.menu" || req.ElementRef == "" || req.PostWaitMs != 250 {
		t.Fatalf("unexpected hover request: %#v", req)
	}
	var payload struct {
		Kind            string                                              `json:"kind"`
		Backend         string                                              `json:"backend"`
		FinalURL        string                                              `json:"final_url"`
		Selector        string                                              `json:"selector"`
		Status          string                                              `json:"status"`
		ResolverOutcome *agentxbrowserruntime.BrowserElementResolverOutcome `json:"resolver_outcome"`
		Summary         *browserTopLevelSummary                             `json:"summary"`
		Display         *browserTopLevelDisplaySummary                      `json:"display"`
		Surface         *browserTopLevelSurfaceSummary                      `json:"surface"`
		View            *browserTopLevelViewSummary                         `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "hover" || payload.Backend != "proxy-hover" || payload.FinalURL != "https://93.184.216.34/app" || payload.Selector != "button.menu" || payload.Status != "hovered" {
		t.Fatalf("unexpected browser_act hover payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "interaction" || payload.Summary.SummaryCode != "hover_completed" {
		t.Fatalf("unexpected browser_act hover summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "interaction" || payload.Display.SummaryCode != "hover_completed" {
		t.Fatalf("unexpected browser_act hover display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "interaction" || payload.Surface.SummaryCode != "hover_completed" {
		t.Fatalf("unexpected browser_act hover surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "interaction" || payload.View.SummaryCode != "hover_completed" {
		t.Fatalf("unexpected browser_act hover view: %#v", payload.View)
	}
	if payload.ResolverOutcome == nil || payload.ResolverOutcome.Status != "matched" || payload.ResolverOutcome.MatchedKind != "selector" || payload.ResolverOutcome.MatchedIndex != 0 {
		t.Fatalf("unexpected browser_act hover resolver outcome: %#v", payload.ResolverOutcome)
	}
}

func TestRegisterBrowserTools_ActHoverReturnsStructuredMissingLocatorErrorForTextHint(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Hover: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"hover","text":"Open menu"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act hover missing locator error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_locator" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: selector or ref is required for kind hover" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"selector_or_ref"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_field", "use_declared_hint"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActHoverAcceptsElementLabelHint(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			hoverResult: BrowserHoverResult{
				Backend:    "proxy-hover",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Hover",
				Status:     "hovered",
			},
		},
		capabilities: BrowserCapabilities{
			Hover: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"hover","element":"Open menu"}`,
	})
	if err != nil {
		t.Fatalf("browser_act hover element label: %v", err)
	}
	if len(backend.hoverReqs) != 1 || backend.hoverReqs[0].Selector != "" || backend.hoverReqs[0].ElementRef == "" {
		t.Fatalf("unexpected browser_act element-label hover request: %#v", backend.hoverReqs)
	}
	if backend.hoverReqs[0].ElementHint == nil || backend.hoverReqs[0].ElementHint.Label != "Open menu" || len(backend.hoverReqs[0].ElementHint.LocatorPlan) != 1 || backend.hoverReqs[0].ElementHint.LocatorPlan[0].Kind != "label" {
		t.Fatalf("unexpected browser_act element-label hover hint: %#v", backend.hoverReqs[0].ElementHint)
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
	if payload.Kind != "hover" || payload.Status != "hovered" || payload.Ref == "" || payload.Selector != "" {
		t.Fatalf("unexpected browser_act element-label hover payload: %#v", payload)
	}
}

func TestRegisterBrowserTools_ActSelectReturnsStructuredMissingLocatorErrorForTextHint(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Select: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"select","text":"Country","value":"CN"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act select missing locator error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_locator" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: selector or ref is required for kind select" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"selector_or_ref"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"use_alias_field", "use_declared_hint"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActSelectAcceptsElementLabelHint(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			selectResult: BrowserSelectResult{
				Backend:    "proxy-select",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Select",
				Values:     []string{"shanghai"},
				Status:     "selected",
			},
		},
		capabilities: BrowserCapabilities{
			Select: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"select","element":"City","value":"shanghai"}`,
	})
	if err != nil {
		t.Fatalf("browser_act select element label: %v", err)
	}
	if len(backend.selectReqs) != 1 || backend.selectReqs[0].Selector != "" || backend.selectReqs[0].ElementRef == "" {
		t.Fatalf("unexpected browser_act element-label select request: %#v", backend.selectReqs)
	}
	if backend.selectReqs[0].ElementHint == nil || backend.selectReqs[0].ElementHint.Label != "City" || len(backend.selectReqs[0].ElementHint.LocatorPlan) != 1 || backend.selectReqs[0].ElementHint.LocatorPlan[0].Kind != "label" {
		t.Fatalf("unexpected browser_act element-label select hint: %#v", backend.selectReqs[0].ElementHint)
	}
	if !reflect.DeepEqual(backend.selectReqs[0].Values, []string{"shanghai"}) {
		t.Fatalf("unexpected browser_act element-label select values: %#v", backend.selectReqs[0].Values)
	}
	var payload struct {
		Kind     string   `json:"kind"`
		Status   string   `json:"status"`
		Ref      string   `json:"ref"`
		Selector string   `json:"selector"`
		Values   []string `json:"values"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "select" || payload.Status != "selected" || payload.Ref == "" || payload.Selector != "" || !reflect.DeepEqual(payload.Values, []string{"shanghai"}) {
		t.Fatalf("unexpected browser_act element-label select payload: %#v", payload)
	}
}
