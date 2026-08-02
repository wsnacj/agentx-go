package tools

import "testing"

func TestBrowserSkipsInteractiveApprovalPrompt(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		arguments string
		want      bool
	}{
		{name: "extract compat", toolName: "browser_extract", arguments: `{"target":"current"}`, want: true},
		{name: "open compat", toolName: "browser_open", arguments: `{"url":"https://93.184.216.34"}`, want: false},
		{name: "screenshot compat", toolName: "browser_screenshot", arguments: `{"target":"current"}`, want: true},
		{name: "tabs default list compat", toolName: "browser_tabs", arguments: `{}`, want: true},
		{name: "tabs list compat", toolName: "browser_tabs", arguments: `{"action":"list"}`, want: true},
		{name: "tabs focus compat", toolName: "browser_tabs", arguments: `{"action":"focus","tab_index":2}`, want: false},
		{name: "browser act snapshot", toolName: "browser_act", arguments: `{"kind":"snapshot","target":"current"}`, want: true},
		{name: "browser act click", toolName: "browser_act", arguments: `{"kind":"click","target":"current"}`, want: false},
		{name: "invalid args", toolName: "browser_tabs", arguments: `{`, want: false},
		{name: "non browser", toolName: "read", arguments: `{}`, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := BrowserSkipsInteractiveApprovalPrompt(tc.toolName, tc.arguments); got != tc.want {
				t.Fatalf("BrowserSkipsInteractiveApprovalPrompt(%q, %q) = %v, want %v", tc.toolName, tc.arguments, got, tc.want)
			}
		})
	}
}
