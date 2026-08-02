package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	types "github.com/wsnacj/agentx-go/components/llm"
	llmxtools "github.com/wsnacj/agentx-go/tools"
)

func TestRegisterBrowserTools_ActCookiesReturnsEntries(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			cookiesResult: BrowserCookiesResult{
				Backend:    "proxy-cookies",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Cookies",
				Cookies: []BrowserCookieEntry{
					{Name: "session", Value: "abc123", Domain: ".example.com", Path: "/", HTTPOnly: true, Secure: true},
				},
				Note: "captured",
			},
		},
		capabilities: BrowserCapabilities{
			Cookies: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"cookies","tab_index":2,"filter":"session"}`,
	})
	if err != nil {
		t.Fatalf("browser_act cookies: %v", err)
	}
	if len(backend.cookiesReqs) != 1 {
		t.Fatalf("expected one cookies request, got %#v", backend.cookiesReqs)
	}
	req := backend.cookiesReqs[0]
	if req.TabIndex != 2 || req.Filter != "session" {
		t.Fatalf("unexpected cookies request: %#v", req)
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
		Cookies                []struct {
			Name     string `json:"name"`
			Value    string `json:"value"`
			Domain   string `json:"domain"`
			Path     string `json:"path"`
			HTTPOnly bool   `json:"http_only"`
			Secure   bool   `json:"secure"`
		} `json:"cookies"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "cookies" || payload.Backend != "proxy-cookies" || payload.FinalURL != "https://93.184.216.34/app" || payload.Status != "ok" || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act cookies payload: %#v", payload)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "storage" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "cookies_collected" {
		t.Fatalf("unexpected browser_act cookies diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "storage" || payload.Summary.SummaryCode != "cookies_collected" {
		t.Fatalf("unexpected browser_act cookies summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "storage" || payload.Display.SummaryCode != "cookies_collected" {
		t.Fatalf("unexpected browser_act cookies display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "storage" || payload.Surface.SummaryCode != "cookies_collected" {
		t.Fatalf("unexpected browser_act cookies surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "storage" || payload.View.SummaryCode != "cookies_collected" {
		t.Fatalf("unexpected browser_act cookies view: %#v", payload.View)
	}
	if len(payload.Cookies) != 1 || payload.Cookies[0].Name != "session" || payload.Cookies[0].Value != "abc123" || payload.Cookies[0].Domain != ".example.com" || payload.Cookies[0].Path != "/" || !payload.Cookies[0].HTTPOnly || !payload.Cookies[0].Secure {
		t.Fatalf("unexpected browser_act cookies entries: %#v", payload.Cookies)
	}
}

func TestRegisterBrowserTools_ActCookiesSetReturnsEntries(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			cookiesSetResult: BrowserCookiesResult{
				Backend:    "proxy-cookies-set",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Cookies",
				Status:     "updated",
				Cookies: []BrowserCookieEntry{
					{Name: "session", Value: "xyz789", Domain: ".example.com", Path: "/", HTTPOnly: true, Secure: true},
				},
			},
		},
		capabilities: BrowserCapabilities{
			CookiesSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"cookies_set","tab_index":2,"url":"https://93.184.216.34/app","name":"session","value":"xyz789","domain":".example.com","path":"/","http_only":true,"secure":true}`,
	})
	if err != nil {
		t.Fatalf("browser_act cookies_set: %v", err)
	}
	if len(backend.cookiesSetReqs) != 1 {
		t.Fatalf("expected one cookies_set request, got %#v", backend.cookiesSetReqs)
	}
	req := backend.cookiesSetReqs[0]
	if req.TabIndex != 2 || req.URL != "https://93.184.216.34/app" || len(req.Cookies) != 1 || req.Cookies[0].Name != "session" || req.Cookies[0].Value != "xyz789" {
		t.Fatalf("unexpected cookies_set request: %#v", req)
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
		Cookies                []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"cookies"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "cookies_set" || payload.Backend != "proxy-cookies-set" || payload.FinalURL != "https://93.184.216.34/app" || payload.Status != "updated" || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act cookies_set payload: %#v", payload)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "storage" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "cookies_set_completed" {
		t.Fatalf("unexpected browser_act cookies_set diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "storage" || payload.Summary.SummaryCode != "cookies_set_completed" {
		t.Fatalf("unexpected browser_act cookies_set summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "storage" || payload.Display.SummaryCode != "cookies_set_completed" {
		t.Fatalf("unexpected browser_act cookies_set display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "storage" || payload.Surface.SummaryCode != "cookies_set_completed" {
		t.Fatalf("unexpected browser_act cookies_set surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "storage" || payload.View.SummaryCode != "cookies_set_completed" {
		t.Fatalf("unexpected browser_act cookies_set view: %#v", payload.View)
	}
	if len(payload.Cookies) != 1 || payload.Cookies[0].Name != "session" || payload.Cookies[0].Value != "xyz789" {
		t.Fatalf("unexpected browser_act cookies_set entries: %#v", payload.Cookies)
	}
}

func TestRegisterBrowserTools_ActCookiesClearReturnsStatus(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			cookiesClearResult: BrowserCookiesResult{
				Backend:    "proxy-cookies-clear",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Cookies",
				Status:     "cleared",
			},
		},
		capabilities: BrowserCapabilities{
			CookiesClear: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"cookies_clear","tab_index":2,"filter":"session"}`,
	})
	if err != nil {
		t.Fatalf("browser_act cookies_clear: %v", err)
	}
	if len(backend.cookiesClearReqs) != 1 {
		t.Fatalf("expected one cookies_clear request, got %#v", backend.cookiesClearReqs)
	}
	req := backend.cookiesClearReqs[0]
	if req.TabIndex != 2 || req.Filter != "session" {
		t.Fatalf("unexpected cookies_clear request: %#v", req)
	}
	var payload struct {
		Kind                   string                                `json:"kind"`
		Backend                string                                `json:"backend"`
		Status                 string                                `json:"status"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "cookies_clear" || payload.Backend != "proxy-cookies-clear" || payload.Status != "cleared" {
		t.Fatalf("unexpected browser_act cookies_clear payload: %#v", payload)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "storage" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "cookies_clear_completed" {
		t.Fatalf("unexpected browser_act cookies_clear diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "storage" || payload.Summary.SummaryCode != "cookies_clear_completed" {
		t.Fatalf("unexpected browser_act cookies_clear summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "storage" || payload.Display.SummaryCode != "cookies_clear_completed" {
		t.Fatalf("unexpected browser_act cookies_clear display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "storage" || payload.Surface.SummaryCode != "cookies_clear_completed" {
		t.Fatalf("unexpected browser_act cookies_clear surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "storage" || payload.View.SummaryCode != "cookies_clear_completed" {
		t.Fatalf("unexpected browser_act cookies_clear view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_ActStorageReturnsEntries(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			storageResult: BrowserStorageResult{
				Backend:    "proxy-storage",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Storage",
				Kind:       "session",
				Entries: []BrowserStorageEntry{
					{Key: "theme", Value: "dark"},
				},
				Note: "captured",
			},
		},
		capabilities: BrowserCapabilities{
			Storage: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"storage","tab_index":2,"storage_kind":"session","filter":"theme"}`,
	})
	if err != nil {
		t.Fatalf("browser_act storage: %v", err)
	}
	if len(backend.storageReqs) != 1 {
		t.Fatalf("expected one storage request, got %#v", backend.storageReqs)
	}
	req := backend.storageReqs[0]
	if req.TabIndex != 2 || req.Kind != "session" || req.Filter != "theme" {
		t.Fatalf("unexpected storage request: %#v", req)
	}
	var payload struct {
		Kind                   string                                `json:"kind"`
		Backend                string                                `json:"backend"`
		FinalURL               string                                `json:"final_url"`
		Status                 string                                `json:"status"`
		TabIndex               int                                   `json:"tab_index"`
		StorageKind            string                                `json:"storage_kind"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
		Storage                []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"storage"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "storage" || payload.Backend != "proxy-storage" || payload.FinalURL != "https://93.184.216.34/app" || payload.Status != "ok" || payload.TabIndex != 2 || payload.StorageKind != "session" {
		t.Fatalf("unexpected browser_act storage payload: %#v", payload)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "storage" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "storage_collected" {
		t.Fatalf("unexpected browser_act storage diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "storage" || payload.Summary.SummaryCode != "storage_collected" {
		t.Fatalf("unexpected browser_act storage summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "storage" || payload.Display.SummaryCode != "storage_collected" {
		t.Fatalf("unexpected browser_act storage display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "storage" || payload.Surface.SummaryCode != "storage_collected" {
		t.Fatalf("unexpected browser_act storage surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "storage" || payload.View.SummaryCode != "storage_collected" {
		t.Fatalf("unexpected browser_act storage view: %#v", payload.View)
	}
	if len(payload.Storage) != 1 || payload.Storage[0].Key != "theme" || payload.Storage[0].Value != "dark" {
		t.Fatalf("unexpected browser_act storage entries: %#v", payload.Storage)
	}
}

func TestRegisterBrowserTools_ActStorageSetReturnsEntries(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			storageSetResult: BrowserStorageResult{
				Backend:    "proxy-storage-set",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Storage",
				Kind:       "session",
				Status:     "updated",
				Entries: []BrowserStorageEntry{
					{Key: "theme", Value: "dark"},
				},
			},
		},
		capabilities: BrowserCapabilities{
			StorageSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"storage_set","tab_index":2,"storage_kind":"session","key":"theme","value":"dark"}`,
	})
	if err != nil {
		t.Fatalf("browser_act storage_set: %v", err)
	}
	if len(backend.storageSetReqs) != 1 {
		t.Fatalf("expected one storage_set request, got %#v", backend.storageSetReqs)
	}
	req := backend.storageSetReqs[0]
	if req.TabIndex != 2 || req.Kind != "session" || len(req.Entries) != 1 || req.Entries[0].Key != "theme" || req.Entries[0].Value != "dark" {
		t.Fatalf("unexpected storage_set request: %#v", req)
	}
	var payload struct {
		Kind                   string                                `json:"kind"`
		Backend                string                                `json:"backend"`
		Status                 string                                `json:"status"`
		StorageKind            string                                `json:"storage_kind"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
		Storage                []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"storage"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "storage_set" || payload.Backend != "proxy-storage-set" || payload.Status != "updated" || payload.StorageKind != "session" {
		t.Fatalf("unexpected browser_act storage_set payload: %#v", payload)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "storage" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "storage_set_completed" {
		t.Fatalf("unexpected browser_act storage_set diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "storage" || payload.Summary.SummaryCode != "storage_set_completed" {
		t.Fatalf("unexpected browser_act storage_set summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "storage" || payload.Display.SummaryCode != "storage_set_completed" {
		t.Fatalf("unexpected browser_act storage_set display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "storage" || payload.Surface.SummaryCode != "storage_set_completed" {
		t.Fatalf("unexpected browser_act storage_set surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "storage" || payload.View.SummaryCode != "storage_set_completed" {
		t.Fatalf("unexpected browser_act storage_set view: %#v", payload.View)
	}
	if len(payload.Storage) != 1 || payload.Storage[0].Key != "theme" || payload.Storage[0].Value != "dark" {
		t.Fatalf("unexpected browser_act storage_set entries: %#v", payload.Storage)
	}
}

func TestRegisterBrowserTools_ActStorageSetReturnsStructuredMissingStorageEntriesErrorForSingularEntryAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			StorageSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"storage_set","storage_kind":"session","entry":{"key":"theme","value":"dark"}}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act storage_set missing entries error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_storage_entries" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: kind storage_set requires key/value or entries[]" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"key_value_or_entries"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"promote_singular_entry"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActStorageSetReturnsStructuredMissingStorageEntriesErrorForStringifiedSingularEntryAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			StorageSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: "{\"kind\":\"storage_set\",\"entry\":\"{\\\"key\\\":\\\"theme\\\",\\\"value\\\":\\\"dark\\\"}\"}",
	})
	if err == nil {
		t.Fatalf("expected browser_act storage_set stringified singular entry alias error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_storage_entries" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"promote_singular_entry"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActStorageSetReturnsStructuredInvalidEntriesShapeError(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			StorageSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"storage_set","storage_kind":"session","entries":{"key":"theme","value":"dark"}}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act storage_set invalid entries shape error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "invalid_storage_entries_shape" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: entries must be an array of objects" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"entries_array"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"wrap_singleton_entry"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActStorageSetReturnsStructuredInvalidEntriesShapeErrorForStringifiedEntries(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			StorageSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: "{\"kind\":\"storage_set\",\"entries\":\"[{\\\"key\\\":\\\"theme\\\",\\\"value\\\":\\\"dark\\\"}]\"}",
	})
	if err == nil {
		t.Fatalf("expected browser_act storage_set invalid entries shape error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "invalid_storage_entries_shape" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"parse_stringified_entries"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActStorageSetReturnsStructuredInvalidEntriesShapeErrorForStringifiedSingleEntry(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			StorageSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: "{\"kind\":\"storage_set\",\"entries\":\"{\\\"key\\\":\\\"theme\\\",\\\"value\\\":\\\"dark\\\"}\"}",
	})
	if err == nil {
		t.Fatalf("expected browser_act storage_set invalid entries shape error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "invalid_storage_entries_shape" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"parse_stringified_entries"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActCookiesSetReturnsStructuredMissingCookieEntriesErrorForSingularCookieAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			CookiesSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"cookies_set","cookie":{"name":"session","value":"xyz789"}}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act cookies_set missing cookie entries error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_cookie_entries" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: kind cookies_set requires cookie fields or cookies[]" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"cookie_fields_or_cookies"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"promote_singular_cookie"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActCookiesSetReturnsStructuredMissingCookieEntriesErrorForStringifiedSingularCookieAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			CookiesSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: "{\"kind\":\"cookies_set\",\"cookie\":\"{\\\"name\\\":\\\"session\\\",\\\"value\\\":\\\"xyz789\\\"}\"}",
	})
	if err == nil {
		t.Fatalf("expected browser_act cookies_set stringified singular cookie alias error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_cookie_entries" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"promote_singular_cookie"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActCookiesSetReturnsStructuredInvalidCookiesShapeError(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			CookiesSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"cookies_set","cookies":{"name":"session","value":"xyz789"}}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act cookies_set invalid cookies shape error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "invalid_cookie_entries_shape" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: cookies must be an array of objects" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"cookies_array"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"wrap_singleton_cookie"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActCookiesSetReturnsStructuredInvalidCookiesShapeErrorForStringifiedCookies(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			CookiesSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: "{\"kind\":\"cookies_set\",\"cookies\":\"[{\\\"name\\\":\\\"session\\\",\\\"value\\\":\\\"xyz789\\\"}]\"}",
	})
	if err == nil {
		t.Fatalf("expected browser_act cookies_set invalid cookies shape error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "invalid_cookie_entries_shape" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"parse_stringified_cookies"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActCookiesSetReturnsStructuredInvalidCookiesShapeErrorForStringifiedSingleCookie(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			CookiesSet: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: "{\"kind\":\"cookies_set\",\"cookies\":\"{\\\"name\\\":\\\"session\\\",\\\"value\\\":\\\"xyz789\\\"}\"}",
	})
	if err == nil {
		t.Fatalf("expected browser_act cookies_set invalid cookies shape error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "invalid_cookie_entries_shape" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"parse_stringified_cookies"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActHeadersReturnsStructuredMissingHeadersErrorForSingularHeaderAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Headers: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"headers","header":{"X-Test":"1","X-Trace":"abc"}}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act headers missing headers error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_headers" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: kind headers requires headers, headers_json, or clear=true" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"headers_or_headers_json_or_clear"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"promote_singular_header"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActHeadersReturnsStructuredMissingHeadersErrorForHeaderJSONMapAlias(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Headers: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"headers","headers_json":{"X-Test":"1","X-Trace":"abc"}}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act headers missing headers error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_headers" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: kind headers requires headers, headers_json, or clear=true" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"headers_or_headers_json_or_clear"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"promote_header_json_map"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActHeadersReturnsStructuredMissingHeadersErrorForStringifiedHeaders(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{},
		capabilities: BrowserCapabilities{
			Headers: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	_, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"headers","headers":"{\"X-Test\":\"1\",\"X-Trace\":\"abc\"}"}`,
	})
	if err == nil {
		t.Fatalf("expected browser_act headers missing headers error")
	}
	argErr, ok := AsToolArgumentError(err)
	if !ok {
		t.Fatalf("expected structured argument error, got %T %v", err, err)
	}
	if argErr.Code != "missing_headers" || !argErr.Repairable || !argErr.SafeAutorepair {
		t.Fatalf("unexpected structured argument error: %#v", argErr)
	}
	if argErr.Error() != "browser_act: kind headers requires headers, headers_json, or clear=true" {
		t.Fatalf("unexpected error detail: %q", argErr.Error())
	}
	if !reflect.DeepEqual(argErr.MissingFields, []string{"headers_or_headers_json_or_clear"}) {
		t.Fatalf("unexpected missing fields: %#v", argErr.MissingFields)
	}
	if got, want := browserRepairAdviceKinds(argErr.AllowedRepairs), []string{"parse_stringified_headers"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected repair advice: got=%#v want=%#v", got, want)
	}
}

func TestRegisterBrowserTools_ActStorageClearReturnsStatus(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			storageClearResult: BrowserStorageResult{
				Backend:    "proxy-storage-clear",
				BrowserApp: "Chromium",
				FinalURL:   "https://93.184.216.34/app",
				Title:      "Storage",
				Kind:       "local",
				Status:     "cleared",
			},
		},
		capabilities: BrowserCapabilities{
			StorageClear: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"storage_clear","tab_index":2,"storage_kind":"local","key":"theme"}`,
	})
	if err != nil {
		t.Fatalf("browser_act storage_clear: %v", err)
	}
	if len(backend.storageClearReqs) != 1 {
		t.Fatalf("expected one storage_clear request, got %#v", backend.storageClearReqs)
	}
	req := backend.storageClearReqs[0]
	if req.TabIndex != 2 || req.Kind != "local" || req.Key != "theme" {
		t.Fatalf("unexpected storage_clear request: %#v", req)
	}
	var payload struct {
		Kind                   string                                `json:"kind"`
		Backend                string                                `json:"backend"`
		Status                 string                                `json:"status"`
		StorageKind            string                                `json:"storage_kind"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "storage_clear" || payload.Backend != "proxy-storage-clear" || payload.Status != "cleared" || payload.StorageKind != "local" {
		t.Fatalf("unexpected browser_act storage_clear payload: %#v", payload)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "storage" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "storage_clear_completed" {
		t.Fatalf("unexpected browser_act storage_clear diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "storage" || payload.Summary.SummaryCode != "storage_clear_completed" {
		t.Fatalf("unexpected browser_act storage_clear summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "storage" || payload.Display.SummaryCode != "storage_clear_completed" {
		t.Fatalf("unexpected browser_act storage_clear display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "storage" || payload.Surface.SummaryCode != "storage_clear_completed" {
		t.Fatalf("unexpected browser_act storage_clear surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "storage" || payload.View.SummaryCode != "storage_clear_completed" {
		t.Fatalf("unexpected browser_act storage_clear view: %#v", payload.View)
	}
}

func TestRegisterBrowserTools_ActHeadersReturnsStatus(t *testing.T) {
	reg := llmxtools.NewRegistry()
	backend := &capabilityBrowserBackend{
		fakeBrowserBackend: &fakeBrowserBackend{
			headersResult: BrowserHeadersResult{
				Backend:     "proxy-headers",
				BrowserApp:  "Chromium",
				FinalURL:    "https://93.184.216.34/app",
				Title:       "Headers",
				Status:      "updated",
				HeaderNames: []string{"X-Test", "X-Trace"},
				HeaderCount: 2,
			},
		},
		capabilities: BrowserCapabilities{
			Headers: true,
		},
	}
	RegisterBrowserTools(reg, BrowserToolOptions{
		Root:         t.TempDir(),
		Backend:      backend,
		EnabledTools: []string{"browser_act"},
	})

	out, err := reg.Execute(context.Background(), types.FunctionCall{
		Name:      "browser_act",
		Arguments: `{"kind":"headers","tab_index":2,"headers":{"X-Test":"1","X-Trace":"abc"}}`,
	})
	if err != nil {
		t.Fatalf("browser_act headers: %v", err)
	}
	if len(backend.headersReqs) != 1 {
		t.Fatalf("expected one headers request, got %#v", backend.headersReqs)
	}
	req := backend.headersReqs[0]
	if req.TabIndex != 2 || req.Clear || len(req.Headers) != 2 || req.Headers["X-Test"] != "1" || req.Headers["X-Trace"] != "abc" {
		t.Fatalf("unexpected headers request: %#v", req)
	}
	var payload struct {
		Kind                   string                                `json:"kind"`
		Backend                string                                `json:"backend"`
		FinalURL               string                                `json:"final_url"`
		Status                 string                                `json:"status"`
		TabIndex               int                                   `json:"tab_index"`
		HeaderNames            []string                              `json:"header_names"`
		HeaderCount            int                                   `json:"header_count"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if payload.Kind != "headers" || payload.Backend != "proxy-headers" || payload.FinalURL != "https://93.184.216.34/app" || payload.Status != "updated" || payload.TabIndex != 2 {
		t.Fatalf("unexpected browser_act headers payload: %#v", payload)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "network" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "headers_updated" {
		t.Fatalf("unexpected browser_act headers diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "network" || payload.Summary.SummaryCode != "headers_updated" {
		t.Fatalf("unexpected browser_act headers summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "network" || payload.Display.SummaryCode != "headers_updated" {
		t.Fatalf("unexpected browser_act headers display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "network" || payload.Surface.SummaryCode != "headers_updated" {
		t.Fatalf("unexpected browser_act headers surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "network" || payload.View.SummaryCode != "headers_updated" {
		t.Fatalf("unexpected browser_act headers view: %#v", payload.View)
	}
	if payload.HeaderCount != 2 || !reflect.DeepEqual(payload.HeaderNames, []string{"x-test", "x-trace"}) {
		t.Fatalf("unexpected browser_act headers metadata: names=%#v count=%d", payload.HeaderNames, payload.HeaderCount)
	}
}
