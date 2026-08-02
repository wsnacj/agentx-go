package tools

import "testing"

func TestBrowserCompatForceConfirmationNeedsGuardianReview(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		params     map[string]any
		wantReview bool
	}{
		{
			name:       "navigate uses descriptor flag",
			toolName:   "browser_navigate",
			params:     map[string]any{"force": true},
			wantReview: true,
		},
		{
			name:       "tabs focus uses descriptor action owner",
			toolName:   "browser_tabs",
			params:     map[string]any{"action": "focus", "force": true},
			wantReview: true,
		},
		{
			name:       "tabs list remember uses descriptor remember action owner",
			toolName:   "browser_tabs",
			params:     map[string]any{"action": "list", "remember_target": true, "force": true},
			wantReview: true,
		},
		{
			name:       "tabs close stays cleanup",
			toolName:   "browser_tabs",
			params:     map[string]any{"action": "close", "force": true},
			wantReview: false,
		},
		{
			name:       "tabs list without remember stays read only",
			toolName:   "browser_tabs",
			params:     map[string]any{"action": "list", "force": true},
			wantReview: false,
		},
		{
			name:       "missing force does not review",
			toolName:   "browser_click",
			params:     map[string]any{"selector": "#submit"},
			wantReview: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := browserCompatForceConfirmationNeedsGuardianReview(tc.toolName, tc.params); got != tc.wantReview {
				t.Fatalf("browserCompatForceConfirmationNeedsGuardianReview(%q, %#v) = %v, want %v", tc.toolName, tc.params, got, tc.wantReview)
			}
		})
	}
}

func TestBrowserArtifactSourceForTool(t *testing.T) {
	tests := []struct {
		toolName string
		want     string
	}{
		{toolName: "browser_extract", want: "browser"},
		{toolName: "browser_screenshot", want: "browser"},
		{toolName: "browser_act", want: "browser"},
		{toolName: "browser_runtime", want: "browser"},
		{toolName: "browser_open", want: ""},
		{toolName: "read", want: ""},
	}
	for _, tc := range tests {
		if got := browserArtifactSourceForTool(tc.toolName); got != tc.want {
			t.Fatalf("browserArtifactSourceForTool(%q) = %q, want %q", tc.toolName, got, tc.want)
		}
	}
}

func TestBrowserCompatActorAndEventSource(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		action     string
		wantSource string
		wantActor  string
	}{
		{name: "extract uses tool name", toolName: "browser_extract", wantSource: "browser_extract", wantActor: "browser_extract"},
		{name: "tabs keeps base source", toolName: "browser_tabs", action: "focus", wantSource: "browser_tabs", wantActor: "browser_tabs focus"},
		{name: "tabs list uses action actor", toolName: "browser_tabs", action: "list", wantSource: "browser_tabs", wantActor: "browser_tabs list"},
		{name: "tabs empty action keeps base actor", toolName: "browser_tabs", wantSource: "browser_tabs", wantActor: "browser_tabs"},
		{name: "unknown stays empty", toolName: "browser", wantSource: "", wantActor: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := browserCompatEventSource(tc.toolName); got != tc.wantSource {
				t.Fatalf("browserCompatEventSource(%q) = %q, want %q", tc.toolName, got, tc.wantSource)
			}
			if got := browserCompatActor(tc.toolName, tc.action); got != tc.wantActor {
				t.Fatalf("browserCompatActor(%q, %q) = %q, want %q", tc.toolName, tc.action, got, tc.wantActor)
			}
		})
	}
}

func TestBrowserCompatImplicitLegacyHostFallbackError(t *testing.T) {
	if err := browserCompatImplicitLegacyHostFallbackError("browser_extract", true, false, BrowserRuntimeInfo{Target: "host"}, browserToolTarget{}, "", "https://93.184.216.34"); err == nil {
		t.Fatal("expected browser_extract compat fallback helper to surface host fallback error")
	}
	if err := browserCompatImplicitLegacyHostFallbackError("browser_tabs", false, false, BrowserRuntimeInfo{Target: "host"}, browserToolTarget{}, "list", ""); err != nil {
		t.Fatalf("expected browser_tabs compat fallback helper to allow non-hidden host default list, got %v", err)
	}
	if err := browserCompatImplicitLegacyHostFallbackError("browser_open", true, true, BrowserRuntimeInfo{Target: "host"}, browserToolTarget{}, "", "https://93.184.216.34"); err != nil {
		t.Fatalf("expected explicit host runtime target to satisfy browser_open fallback helper, got %v", err)
	}
}
