package browserruntime_test

import (
	"reflect"
	"testing"

	browserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func TestExternalConsumerTracksTargetsAndToolSurface(t *testing.T) {
	registry := browserruntime.NewBrowserSessionRegistry()
	tracked := registry.TrackTab("session-1", browserruntime.BrowserSessionTarget{
		TabIndex: 1, URL: "https://example.test/", Title: "Fixture",
		Backend: "fixture", Profile: "isolated",
	}, true)
	current, ok := registry.CurrentTarget("session-1")
	if !ok || current.ID == "" || current.ID != tracked.ID || current.URL != "https://example.test/" {
		t.Fatalf("tracked=%#v current=%#v ok=%v", tracked, current, ok)
	}

	wantNames := []string{
		"browser", "browser_runtime", "browser_act", "browser_open", "browser_navigate",
		"browser_tabs", "browser_extract", "browser_screenshot", "browser_click", "browser_type",
		"browser_eval",
	}
	if got := browserruntime.BrowserAllToolNames(); !reflect.DeepEqual(got, wantNames) {
		t.Fatalf("tool names=%v want=%v", got, wantNames)
	}
	if action := browserruntime.BrowserRuntimeActionForToolCall("browser", map[string]any{"action": "navigate"}); action != "navigate" {
		t.Fatalf("action=%q", action)
	}
}
