package tools_test

import (
	"encoding/json"
	"reflect"
	"testing"

	browsertools "github.com/wsnacj/agentx-go/browser/tools"
)

func TestAttemptConstrainedBrowserArgumentRepairUsesDeclaredSelectorAlias(t *testing.T) {
	repaired, kinds, applied, err := browsertools.AttemptConstrainedBrowserArgumentRepair(
		"browser_type",
		`{"element":"button.buy"}`,
		&browsertools.ToolArgumentError{
			Code:           "missing_locator",
			Repairable:     true,
			SafeAutorepair: true,
			AllowedRepairs: []browsertools.ToolArgumentRepair{{
				Kind: "use_alias_field",
				From: "element",
				To:   "selector_or_ref",
			}},
		},
	)
	if err != nil {
		t.Fatalf("AttemptConstrainedBrowserArgumentRepair: %v", err)
	}
	if !applied || !reflect.DeepEqual(kinds, []string{"use_alias_field"}) {
		t.Fatalf("unexpected repair result: applied=%v kinds=%#v", applied, kinds)
	}
	var args map[string]any
	if err := json.Unmarshal([]byte(repaired), &args); err != nil {
		t.Fatalf("decode repaired arguments: %v", err)
	}
	if got := args["selector"]; got != "button.buy" {
		t.Fatalf("selector=%#v, want button.buy", got)
	}
}

func TestAttemptConstrainedBrowserArgumentRepairRejectsNonBrowserAndUnsafeRepair(t *testing.T) {
	for _, tc := range []struct {
		name     string
		toolName string
		safe     bool
	}{
		{name: "non browser", toolName: "exec", safe: true},
		{name: "unsafe", toolName: "browser_click", safe: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repaired, kinds, applied, err := browsertools.AttemptConstrainedBrowserArgumentRepair(
				tc.toolName,
				`{"text":"Buy now"}`,
				&browsertools.ToolArgumentError{
					Code:           "missing_locator",
					Repairable:     true,
					SafeAutorepair: tc.safe,
					AllowedRepairs: []browsertools.ToolArgumentRepair{{
						Kind: "use_declared_hint",
						From: "text",
					}},
				},
			)
			if err != nil {
				t.Fatalf("AttemptConstrainedBrowserArgumentRepair: %v", err)
			}
			if applied || repaired != "" || len(kinds) != 0 {
				t.Fatalf("unexpected repair: applied=%v repaired=%q kinds=%#v", applied, repaired, kinds)
			}
		})
	}
}
