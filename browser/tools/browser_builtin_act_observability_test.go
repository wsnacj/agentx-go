package tools

import (
	"context"
	"encoding/json"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ActConsoleReturnsMessages(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			consoleResult: BrowserConsoleResult{
				Backend:    "proxy-console",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Console",
				Messages: []BrowserConsoleMessage{
					{Level: "error", Text: "Console payload here.", Source: "app"},
				},
				Note: "captured",
			},
		},
		capabilities: BrowserCapabilities{
			Console: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"console","tab_index":2,"level":"error"}`,
	})
	if err != nil {
		t.Fatalf("browser_act console: %v", err)
	}
	if len(backend.consoleReqs) != 1 {
		t.Fatalf("expected one console request, got %#v", backend.consoleReqs)
	}
	req := backend.consoleReqs[0]
	if req.Level != "error" || req.TabIndex != 2 {
		t.Fatalf("unexpected console request: %#v", req)
	}
	var payload struct {
		Kind     string                         `json:"kind"`
		Backend  string                         `json:"backend"`
		FinalURL string                         `json:"final_url"`
		Status   string                         `json:"status"`
		TabIndex int                            `json:"tab_index"`
		Summary  *browserTopLevelSummary        `json:"summary"`
		Display  *browserTopLevelDisplaySummary `json:"display"`
		Surface  *browserTopLevelSurfaceSummary `json:"surface"`
		View     *browserTopLevelViewSummary    `json:"view"`
		Messages []struct {
			Level  string `json:"level"`
			Text   string `json:"text"`
			Source string `json:"source"`
		} `json:"messages"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "console" || payload.Backend != "proxy-console" || payload.FinalURL != "https://93.184.216.34/app" || payload.Status != "ok" || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act console payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "console_collected" {
		t.Fatalf("unexpected browser_act console summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "console_collected" {
		t.Fatalf("unexpected browser_act console display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "console_collected" {
		t.Fatalf("unexpected browser_act console surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "console_collected" {
		t.Fatalf("unexpected browser_act console view: %#v", payload.View)
	}
	if len(payload.Messages) != 1 || payload.Messages[0].Level != "error" || payload.Messages[0].Text != "Console payload here." || payload.Messages[0].Source != "app" {
		t.Fatalf("unexpected browser_act console messages: %#v", payload.Messages)
	}
}

func TestRegisterBrowserTools_ActRequestsReturnsEntries(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			requestsResult: BrowserRequestsResult{
				Backend:    "proxy-requests",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Requests",
				Requests: []BrowserRequestEntry{
					{Method: "GET", URL: "https://93.184.216.34/api/items", Status: 200, ResourceType: "xhr"},
				},
				Note: "captured",
			},
		},
		capabilities: BrowserCapabilities{
			Requests: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"requests","tab_index":2,"filter":"api"}`,
	})
	if err != nil {
		t.Fatalf("browser_act requests: %v", err)
	}
	if len(backend.requestsReqs) != 1 {
		t.Fatalf("expected one requests request, got %#v", backend.requestsReqs)
	}
	req := backend.requestsReqs[0]
	if req.Filter != "api" || req.TabIndex != 2 {
		t.Fatalf("unexpected requests request: %#v", req)
	}
	var payload struct {
		Kind     string                         `json:"kind"`
		Backend  string                         `json:"backend"`
		FinalURL string                         `json:"final_url"`
		Status   string                         `json:"status"`
		TabIndex int                            `json:"tab_index"`
		Summary  *browserTopLevelSummary        `json:"summary"`
		Display  *browserTopLevelDisplaySummary `json:"display"`
		Surface  *browserTopLevelSurfaceSummary `json:"surface"`
		View     *browserTopLevelViewSummary    `json:"view"`
		Requests []struct {
			Method       string `json:"method"`
			URL          string `json:"url"`
			Status       int    `json:"status"`
			ResourceType string `json:"resource_type"`
		} `json:"requests"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "requests" || payload.Backend != "proxy-requests" || payload.FinalURL != "https://93.184.216.34/app" || payload.Status != "ok" || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act requests payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected browser_act requests summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected browser_act requests display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected browser_act requests surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "requests_collected" {
		t.Fatalf("unexpected browser_act requests view: %#v", payload.View)
	}
	if len(payload.Requests) != 1 || payload.Requests[0].Method != "GET" || payload.Requests[0].URL != "https://93.184.216.34/api/items" || payload.Requests[0].Status != 200 || payload.Requests[0].ResourceType != "xhr" {
		t.Fatalf("unexpected browser_act requests entries: %#v", payload.Requests)
	}
}

func TestRegisterBrowserTools_ActRequestsClearReturnsClearedStatus(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			requestsResult: BrowserRequestsResult{
				Backend:    "proxy-requests",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Requests",
				Requests: []BrowserRequestEntry{
					{Method: "GET", URL: "https://93.184.216.34/api/items", Status: 200, ResourceType: "xhr"},
				},
				Note: "cleared",
			},
		},
		capabilities: BrowserCapabilities{
			Requests: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"requests","tab_index":2,"filter":"api","clear":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act requests clear: %v", err)
	}
	if len(backend.requestsReqs) != 1 {
		t.Fatalf("expected one requests request, got %#v", backend.requestsReqs)
	}
	req := backend.requestsReqs[0]
	if req.Filter != "api" || req.TabIndex != 2 || !req.Clear {
		t.Fatalf("unexpected requests clear request: %#v", req)
	}
	var payload struct {
		Kind     string                         `json:"kind"`
		Status   string                         `json:"status"`
		Summary  *browserTopLevelSummary        `json:"summary"`
		Display  *browserTopLevelDisplaySummary `json:"display"`
		Surface  *browserTopLevelSurfaceSummary `json:"surface"`
		View     *browserTopLevelViewSummary    `json:"view"`
		Requests []struct {
			URL string `json:"url"`
		} `json:"requests"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "requests" || payload.Status != "cleared" {
		t.Fatalf("unexpected browser_act requests clear payload: %#v", payload)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "requests_cleared" {
		t.Fatalf("unexpected browser_act requests clear summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "requests_cleared" {
		t.Fatalf("unexpected browser_act requests clear display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "requests_cleared" {
		t.Fatalf("unexpected browser_act requests clear surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "requests_cleared" {
		t.Fatalf("unexpected browser_act requests clear view: %#v", payload.View)
	}
	if len(payload.Requests) != 1 || payload.Requests[0].URL != "https://93.184.216.34/api/items" {
		t.Fatalf("unexpected browser_act requests clear entries: %#v", payload.Requests)
	}
}

func TestRegisterBrowserTools_ActResponseBodyReturnsContent(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			responseBodyResult: BrowserResponseBodyResult{
				Backend:     "proxy-response-body",
				BrowserApp:  "Chromium",
				FinalURL:    "https://93.184.216.34/app",
				Title:       "Requests",
				URL:         "https://93.184.216.34/api/items",
				Method:      "POST",
				StatusCode:  201,
				ContentType: "application/json",
				Body:        `{"ok":true}`,
				Truncated:   true,
				Note:        "captured",
			},
		},
		capabilities: BrowserCapabilities{
			ResponseBody: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"response_body","tab_index":2,"filter":"items","request_url":"https://93.184.216.34/api/items","max_chars":256}`,
	})
	if err != nil {
		t.Fatalf("browser_act response_body: %v", err)
	}
	if len(backend.responseBodyReqs) != 1 {
		t.Fatalf("expected one response_body request, got %#v", backend.responseBodyReqs)
	}
	req := backend.responseBodyReqs[0]
	if req.Filter != "items" || req.URL != "https://93.184.216.34/api/items" || req.MaxChars != 256 || req.TabIndex != 2 {
		t.Fatalf("unexpected response_body request: %#v", req)
	}
	var payload struct {
		Kind                   string                                `json:"kind"`
		Backend                string                                `json:"backend"`
		FinalURL               string                                `json:"final_url"`
		Status                 string                                `json:"status"`
		TabIndex               int                                   `json:"tab_index"`
		RequestURL             string                                `json:"request_url"`
		RequestMethod          string                                `json:"request_method"`
		ResponseStatusCode     int                                   `json:"response_status_code"`
		Content                string                                `json:"content"`
		ContentType            string                                `json:"content_type"`
		Truncated              bool                                  `json:"truncated"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "response_body" ||
		payload.Backend != "proxy-response-body" ||
		payload.FinalURL != "https://93.184.216.34/app" ||
		payload.Status != "ok" ||
		payload.TabIndex != 2 ||
		payload.RequestURL != "https://93.184.216.34/api/items" ||
		payload.RequestMethod != "POST" ||
		payload.ResponseStatusCode != 201 ||
		payload.Content != `{"ok":true}` ||
		payload.ContentType != "application/json" ||
		!payload.Truncated {
		t.Fatalf("unexpected browser_act response_body payload: %#v", payload)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "content" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "response_body_collected" {
		t.Fatalf("unexpected browser_act response_body diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "content" || payload.Summary.SummaryCode != "response_body_collected" {
		t.Fatalf("unexpected browser_act response_body summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "content" || payload.Display.SummaryCode != "response_body_collected" {
		t.Fatalf("unexpected browser_act response_body display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "content" || payload.Surface.SummaryCode != "response_body_collected" {
		t.Fatalf("unexpected browser_act response_body surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "content" || payload.View.SummaryCode != "response_body_collected" {
		t.Fatalf("unexpected browser_act response_body view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_ActErrorsReturnsEntries(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			errorsResult: BrowserErrorsResult{
				Backend:    "proxy-errors",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Errors",
				Errors: []BrowserErrorEntry{
					{Message: "Unhandled exception", Source: "app.js", URL: "https://93.184.216.34/app.js", ResolverStatus: "page_binding_blocked", BlockedBy: "page_url"},
				},
				Note: "captured",
			},
		},
		capabilities: BrowserCapabilities{
			Errors: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"errors","tab_index":2}`,
	})
	if err != nil {
		t.Fatalf("browser_act errors: %v", err)
	}
	if len(backend.errorsReqs) != 1 {
		t.Fatalf("expected one errors request, got %#v", backend.errorsReqs)
	}
	req := backend.errorsReqs[0]
	if req.TabIndex != 2 {
		t.Fatalf("unexpected errors request: %#v", req)
	}
	var payload struct {
		Kind                   string                                `json:"kind"`
		Backend                string                                `json:"backend"`
		FinalURL               string                                `json:"final_url"`
		Status                 string                                `json:"status"`
		TabIndex               int                                   `json:"tab_index"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
		Errors                 []struct {
			Message        string `json:"message"`
			Source         string `json:"source"`
			URL            string `json:"url"`
			ResolverStatus string `json:"resolver_status"`
			BlockedBy      string `json:"blocked_by"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "errors" || payload.Backend != "proxy-errors" || payload.FinalURL != "https://93.184.216.34/app" || payload.Status != "ok" || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act errors payload: %#v", payload)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "observability" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "errors_collected" {
		t.Fatalf("unexpected browser_act errors diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "errors_collected" {
		t.Fatalf("unexpected browser_act errors summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "errors_collected" {
		t.Fatalf("unexpected browser_act errors display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "errors_collected" {
		t.Fatalf("unexpected browser_act errors surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "errors_collected" {
		t.Fatalf("unexpected browser_act errors view: %#v", payload.View)
	}
	if len(payload.Errors) != 1 ||
		payload.Errors[0].Message != "Unhandled exception" ||
		payload.Errors[0].Source != "app.js" ||
		payload.Errors[0].URL != "https://93.184.216.34/app.js" ||
		payload.Errors[0].ResolverStatus != "page_binding_blocked" ||
		payload.Errors[0].BlockedBy != "page_url" {
		t.Fatalf("unexpected browser_act errors entries: %#v", payload.Errors)
	}
}

func TestRegisterBrowserTools_ActErrorsClearReturnsClearedStatus(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			errorsResult: BrowserErrorsResult{
				Backend:    "proxy-errors",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Errors",
				Errors: []BrowserErrorEntry{
					{Message: "Unhandled exception", Source: "app.js", URL: "https://93.184.216.34/app.js"},
				},
				Note: "cleared",
			},
		},
		capabilities: BrowserCapabilities{
			Errors: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"errors","tab_index":2,"clear":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act errors clear: %v", err)
	}
	if len(backend.errorsReqs) != 1 {
		t.Fatalf("expected one errors request, got %#v", backend.errorsReqs)
	}
	req := backend.errorsReqs[0]
	if req.TabIndex != 2 || !req.Clear {
		t.Fatalf("unexpected errors clear request: %#v", req)
	}
	var payload struct {
		Kind                   string                                `json:"kind"`
		Status                 string                                `json:"status"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
		Errors                 []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "errors" || payload.Status != "cleared" {
		t.Fatalf("unexpected browser_act errors clear payload: %#v", payload)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "observability" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "errors_cleared" {
		t.Fatalf("unexpected browser_act errors clear diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != "errors_cleared" {
		t.Fatalf("unexpected browser_act errors clear summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != "errors_cleared" {
		t.Fatalf("unexpected browser_act errors clear display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != "errors_cleared" {
		t.Fatalf("unexpected browser_act errors clear surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != "errors_cleared" {
		t.Fatalf("unexpected browser_act errors clear view: %#v", payload.View)
	}
	if len(payload.Errors) != 1 || payload.Errors[0].Message != "Unhandled exception" {
		t.Fatalf("unexpected browser_act errors clear entries: %#v", payload.Errors)
	}
}
