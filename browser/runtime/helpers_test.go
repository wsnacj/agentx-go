package browserruntime

import (
	"reflect"
	"testing"
)

func TestBrowserCapabilitiesSupportTables(t *testing.T) {
	caps := BrowserCapabilities{
		Navigate:     true,
		Tabs:         true,
		Snapshot:     true,
		Console:      true,
		Requests:     true,
		ResponseBody: true,
		Errors:       true,
		Cookies:      true,
		CookiesSet:   true,
		CookiesClear: true,
		Storage:      true,
		StorageSet:   true,
		StorageClear: true,
		Offline:      true,
		Headers:      true,
		Credentials:  true,
		Geolocation:  true,
		Media:        true,
		Timezone:     true,
		Locale:       true,
		Device:       true,
		Highlight:    true,
		TraceStart:   true,
		TraceStop:    true,
		Download:     true,
		WaitDownload: true,
		SavePDF:      true,
		SaveHTML:     true,
		Dialog:       true,
		Upload:       true,
		Press:        true,
		Hover:        true,
		Drag:         true,
		Select:       true,
		Fill:         true,
		Resize:       true,
		Click:        true,
		TypeText:     true,
		Evaluate:     true,
		Wait:         true,
	}

	toolCases := map[string]bool{
		"browser":            true,
		"browser_runtime":    true,
		"browser_open":       false,
		"browser_navigate":   true,
		"browser_tabs":       true,
		"browser_extract":    false,
		"browser_screenshot": false,
		"browser_click":      true,
		"browser_type":       true,
		"browser_eval":       true,
		"browser_act":        true,
		"missing":            false,
	}
	for name, want := range toolCases {
		if got := caps.SupportsTool(name); got != want {
			t.Fatalf("SupportsTool(%q)=%v want %v", name, got, want)
		}
	}

	actCases := map[string]bool{
		"open":          false,
		"navigate":      true,
		"snapshot":      true,
		"console":       true,
		"requests":      true,
		"response_body": true,
		"errors":        true,
		"cookies":       true,
		"cookies_set":   true,
		"cookies_clear": true,
		"storage":       true,
		"storage_set":   true,
		"storage_clear": true,
		"offline":       true,
		"headers":       true,
		"credentials":   true,
		"geolocation":   true,
		"media":         true,
		"timezone":      true,
		"locale":        true,
		"device":        true,
		"highlight":     true,
		"trace_start":   true,
		"trace_stop":    true,
		"download":      true,
		"wait_download": true,
		"save_pdf":      true,
		"save_html":     true,
		"dialog":        true,
		"upload":        true,
		"press":         true,
		"hover":         true,
		"drag":          true,
		"select":        true,
		"fill":          true,
		"resize":        true,
		"click":         true,
		"type":          true,
		"evaluate":      true,
		"wait":          true,
		"list_tabs":     true,
		"focus_tab":     true,
		"close_tab":     true,
		"unknown":       false,
	}
	for kind, want := range actCases {
		if got := caps.SupportsActKind(kind); got != want {
			t.Fatalf("SupportsActKind(%q)=%v want %v", kind, got, want)
		}
	}

	if got, want := caps.SupportedActKinds(), []string{
		"navigate",
		"snapshot",
		"console",
		"requests",
		"response_body",
		"errors",
		"cookies",
		"cookies_set",
		"cookies_clear",
		"storage",
		"storage_set",
		"storage_clear",
		"offline",
		"headers",
		"credentials",
		"geolocation",
		"media",
		"timezone",
		"locale",
		"device",
		"highlight",
		"trace_start",
		"trace_stop",
		"download",
		"wait_download",
		"save_pdf",
		"save_html",
		"dialog",
		"upload",
		"press",
		"hover",
		"drag",
		"select",
		"fill",
		"resize",
		"click",
		"type",
		"evaluate",
		"wait",
		"list_tabs",
		"focus_tab",
		"close_tab",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected SupportedActKinds ordering: got=%v want=%v", got, want)
	}
}

func TestBrowserSessionRegistryHelperFunctions(t *testing.T) {
	target := normalizeBrowserSessionTarget(BrowserSessionTarget{
		ID:         " target-1 ",
		TabIndex:   -3,
		URL:        " https://example.com ",
		Title:      " Title ",
		BrowserApp: " Safari ",
		Backend:    " Proxy-Tabs ",
		Profile:    " Workbench ",
		Target:     " Node ",
	})
	if target.TabIndex != 0 || target.ID != "target-1" || target.Backend != "Proxy-Tabs" {
		t.Fatalf("unexpected normalized target: %#v", target)
	}

	if got := firstBrowserSessionTargetSource("", "  ", "remember_target", "select_target"); got != "remember_target" {
		t.Fatalf("expected first non-empty target source, got %q", got)
	}
	if got := firstBrowserSessionTargetSource("", " "); got != "select_target" {
		t.Fatalf("expected default target source fallback, got %q", got)
	}

	if got := browserSessionCanonicalBackend(" Proxy-Tabs "); got != "proxy" {
		t.Fatalf("expected proxy family canonical backend, got %q", got)
	}
	if got := browserSessionCanonicalBackend("proxy_tabs"); got != "proxy" {
		t.Fatalf("expected proxy underscore family canonical backend, got %q", got)
	}
	if got := browserSessionCanonicalBackend("SYSTEM-cdp"); got != "system" {
		t.Fatalf("expected system family canonical backend, got %q", got)
	}
	if got := browserSessionCanonicalBackend("system_open"); got != "system" {
		t.Fatalf("expected system underscore family canonical backend, got %q", got)
	}
	if got := browserSessionCanonicalBackend("safari_tabs"); got != "system" {
		t.Fatalf("expected safari family canonical backend, got %q", got)
	}
	if got := browserSessionCanonicalBackend("http_extract_fallback"); got != "system" {
		t.Fatalf("expected http family canonical backend, got %q", got)
	}
	if !browserSessionBackendMatches("proxy-tabs", "proxy") {
		t.Fatalf("expected proxy backend family match")
	}
	if !browserSessionBackendMatches("system_open", "system") {
		t.Fatalf("expected system backend family match")
	}
	if browserSessionBackendMatches("proxy", "system") {
		t.Fatalf("expected mismatched backend families not to match")
	}

	route := browserSessionRouteFromKey("proxy\x00workbench\x00node\x00chromium")
	if route.Backend != "proxy" || route.Profile != "workbench" || route.Target != "node" || route.BrowserApp != "chromium" {
		t.Fatalf("unexpected route from key: %#v", route)
	}
	if got := browserSessionRouteFromKey("__default__"); got != (BrowserSessionRoute{}) {
		t.Fatalf("expected default route key to decode as zero route, got %#v", got)
	}

	if !browserSessionRouteMatchesFilter(
		BrowserSessionRoute{Backend: "proxy-tabs", Profile: "workbench", Target: "node", BrowserApp: "chromium"},
		BrowserSessionRoute{Backend: "proxy", Profile: "workbench", Target: "node"},
	) {
		t.Fatalf("expected route filter to match proxy backend family")
	}
	if browserSessionRouteMatchesFilter(
		BrowserSessionRoute{Backend: "proxy-tabs", Profile: "workbench", Target: "node"},
		BrowserSessionRoute{Backend: "proxy", Profile: "isolated", Target: "node"},
	) {
		t.Fatalf("expected route filter profile mismatch to fail")
	}
}

func TestBrowserSessionStateRegistrySnapshotSelectedBrowserProfilesSorts(t *testing.T) {
	registry := NewBrowserSessionStateRegistry()
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "isolated",
		RuntimeTarget: "node-b",
		BrowserApp:    "Chromium",
	})
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Backend:       "system",
		Profile:       "default",
		RuntimeTarget: "host",
		BrowserApp:    "Safari",
	})
	registry.SelectBrowserProfile("s1", SharedSessionBrowserProfileSelection{
		Backend:       "proxy",
		Profile:       "workbench",
		RuntimeTarget: "node-a",
		BrowserApp:    "Chromium",
	})

	snapshot := registry.SnapshotSelectedBrowserProfiles("s1")
	if got, want := len(snapshot), 3; got != want {
		t.Fatalf("expected 3 selected browser profiles, got %#v", snapshot)
	}
	if snapshot[0].RuntimeTarget != "host" || snapshot[1].RuntimeTarget != "node-a" || snapshot[2].RuntimeTarget != "node-b" {
		t.Fatalf("expected snapshot to sort by runtime target/backend/profile/browser app, got %#v", snapshot)
	}
}
