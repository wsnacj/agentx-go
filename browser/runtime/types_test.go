package browserruntime

import (
	"reflect"
	"testing"
)

func TestBrowserCapabilitiesSupportsToolAndActKinds(t *testing.T) {
	caps := BrowserCapabilities{
		Open:        true,
		Tabs:        true,
		TypeText:    true,
		RuntimeList: true,
	}

	if !caps.SupportsTool("browser") {
		t.Fatalf("expected unified browser tool to always be supported")
	}
	if !caps.SupportsTool("browser_runtime") {
		t.Fatalf("expected browser_runtime tool to always be supported")
	}
	if !caps.SupportsTool("browser_open") {
		t.Fatalf("expected browser_open to follow Open capability")
	}
	if caps.SupportsTool("browser_click") {
		t.Fatalf("expected browser_click to be unsupported without Click capability")
	}
	if !caps.SupportsTool("browser_act") {
		t.Fatalf("expected browser_act to be supported when any act kind is available")
	}
	if !caps.SupportsActKind("type") || !caps.SupportsActKind("list_tabs") || !caps.SupportsActKind("focus_tab") {
		t.Fatalf("expected aliases backed by TypeText/Tabs to be supported")
	}
	if caps.SupportsActKind("highlight") {
		t.Fatalf("expected unsupported act kind to stay disabled")
	}
}

func TestBrowserCapabilitiesSupportedToolAndRuntimeActionOrdering(t *testing.T) {
	caps := BrowserCapabilities{
		Open:               true,
		Extract:            true,
		TypeText:           true,
		RuntimeWorkbench:   true,
		RuntimePrepare:     true,
		RuntimeRestart:     true,
		RuntimeCreate:      true,
		RuntimeSelect:      true,
		RuntimeSyncSession: true,
		RuntimeList:        true,
		RuntimeSessions:    true,
	}

	if got, want := caps.SupportedToolNames(), []string{
		"browser_runtime",
		"browser_open",
		"browser_extract",
		"browser_type",
		"browser_act",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected supported tool names: got=%v want=%v", got, want)
	}

	if got, want := caps.SupportedRuntimeActions(), []string{
		"status",
		"workbench",
		"prepare",
		"restart",
		"refresh",
		"create_profile",
		"select_profile",
		"sync_session",
		"profiles",
		"sessions",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected supported runtime actions: got=%v want=%v", got, want)
	}
}

func TestBrowserCapabilitiesSupportedActKindsOrdering(t *testing.T) {
	caps := BrowserCapabilities{
		Open:         true,
		ResponseBody: true,
		WaitDownload: true,
		TypeText:     true,
		Evaluate:     true,
		Tabs:         true,
	}

	if got, want := caps.SupportedActKinds(), []string{
		"open",
		"response_body",
		"wait_download",
		"type",
		"evaluate",
		"list_tabs",
		"focus_tab",
		"close_tab",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected supported act kinds: got=%v want=%v", got, want)
	}
}

func TestBrowserCapabilitiesSupportsAnyActKind(t *testing.T) {
	if (BrowserCapabilities{}).SupportsAnyActKind() {
		t.Fatalf("expected empty capabilities to report no act kinds")
	}
	if !(BrowserCapabilities{Wait: true}).SupportsAnyActKind() {
		t.Fatalf("expected a single act kind to enable browser_act support")
	}
}
