package tools

import "testing"

func TestBrowserActTabsAction(t *testing.T) {
	tests := map[string]string{
		"list_tabs": "list",
		"focus_tab": "focus",
		"close_tab": "close",
		"something": "",
		"":          "",
	}
	for kind, want := range tests {
		if got := browserActTabsAction(kind); got != want {
			t.Fatalf("browserActTabsAction(%q) = %q, want %q", kind, got, want)
		}
	}
}

func TestBrowserActTabsWaitMs(t *testing.T) {
	if got := browserActTabsWaitMs(135); got != 135 {
		t.Fatalf("browserActTabsWaitMs(explicit) = %d, want 135", got)
	}
	if got := browserActTabsWaitMs(0); got != 200 {
		t.Fatalf("browserActTabsWaitMs(default) = %d, want 200", got)
	}
}
