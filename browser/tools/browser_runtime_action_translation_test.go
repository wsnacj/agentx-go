package tools

import "testing"

func TestBrowserRuntimeToolAwareActionCommandUsesRuntimeSurfaceWhenUnifiedBrowserDisabled(t *testing.T) {
	ctx := browserRegistrationContext{
		enabledTools: map[string]bool{
			"browser_runtime": true,
			"browser_tabs":    true,
		},
	}
	cases := map[string]string{
		"browser action=refresh":    "browser_runtime action=refresh",
		"browser action=ensure":     "browser_runtime action=coordinate coordination_goal=ensure",
		"browser action=sync":       "browser_runtime action=coordinate coordination_goal=sync",
		"browser action=teardown":   "browser_runtime action=coordinate coordination_goal=teardown",
		"browser action=reset":      "browser_runtime action=clear_session",
		"browser action=launch":     "browser_runtime action=start",
		"browser action=pin_target": "browser_runtime action=select_target",
		"browser action=tabs":       "browser_tabs action=list",
		"browser action=focus":      "browser_tabs action=focus",
		"browser action=close":      "browser_tabs action=close",
	}
	for input, want := range cases {
		if got := browserRuntimeToolAwareActionCommand(ctx, input); got != want {
			t.Fatalf("unexpected translation for %q: got %q want %q", input, got, want)
		}
	}
}

func TestBrowserRuntimeToolAwareNextStepAliasUsesCompatTabsWhenEnabled(t *testing.T) {
	ctx := browserRegistrationContext{
		enabledTools: map[string]bool{
			"browser_runtime": true,
			"browser_tabs":    true,
		},
	}
	cases := map[string]string{
		"tabs":  "list",
		"focus": "focus",
		"close": "close",
	}
	for input, want := range cases {
		if got := browserRuntimeToolAwareNextStepAlias(ctx, input); got != want {
			t.Fatalf("unexpected compat tabs alias translation for %q: got %q want %q", input, got, want)
		}
	}
}

func TestBrowserRuntimeToolAwareActionCommandFallsBackToRuntimeInspectionWithoutBrowserTabs(t *testing.T) {
	ctx := browserRegistrationContext{
		enabledTools: map[string]bool{
			"browser_runtime": true,
		},
	}
	cases := map[string]string{
		"browser":              "browser_runtime action=workbench",
		"browser action=tabs":  "browser_runtime action=sessions",
		"browser action=focus": "browser_runtime action=sessions",
		"browser action=close": "browser_runtime action=sessions",
	}
	for input, want := range cases {
		if got := browserRuntimeToolAwareActionCommand(ctx, input); got != want {
			t.Fatalf("unexpected fallback translation for %q: got %q want %q", input, got, want)
		}
	}
}

func TestBrowserRuntimeToolAwareNextStepAliasFallsBackToRuntimeInspectionWithoutBrowserTabs(t *testing.T) {
	ctx := browserRegistrationContext{
		enabledTools: map[string]bool{
			"browser_runtime": true,
		},
	}
	cases := map[string]string{
		"browser": "workbench",
		"tabs":    "sessions",
		"focus":   "sessions",
		"close":   "sessions",
	}
	for input, want := range cases {
		if got := browserRuntimeToolAwareNextStepAlias(ctx, input); got != want {
			t.Fatalf("unexpected fallback alias translation for %q: got %q want %q", input, got, want)
		}
	}
}

func TestBrowserRuntimeApplyToolAwareActionCommandsRewritesPayloadCommands(t *testing.T) {
	ctx := browserRegistrationContext{
		enabledTools: map[string]bool{"browser_runtime": true},
	}
	payload := browserRuntimePayload{
		WorkbenchPrimaryBrowserAction:      "browser action=reset",
		WorkbenchNextStep:                  "browser action=ensure",
		WorkbenchRecommendedBrowserActions: []string{"browser action=reset", "browser action=ensure", "browser action=refresh"},
		Diagnostics: &browserTopLevelSummary{
			NextStepAlias:        "launch",
			PrimaryBrowserAction: "browser action=refresh",
			NextStep:             "browser action=launch",
		},
		SessionBinding: &browserRuntimeSessionBinding{
			SessionHealthRecoveryAction: "browser action=ensure",
			Coordination: &browserRuntimeCoordination{
				SyncBrowserAction:         "browser action=sync",
				RestartBrowserAction:      "browser action=refresh",
				TeardownBrowserAction:     "browser action=teardown",
				PrimaryBrowserAction:      "browser action=reset",
				NextStep:                  "browser action=ensure",
				RecommendedBrowserActions: []string{"browser action=sync", "browser action=reset"},
			},
		},
	}

	browserRuntimeApplyToolAwareActionCommands(ctx, &payload)

	if payload.WorkbenchPrimaryBrowserAction != "browser_runtime action=clear_session" ||
		payload.WorkbenchNextStep != "browser_runtime action=coordinate coordination_goal=ensure" {
		t.Fatalf("unexpected translated workbench actions: %#v", payload)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.NextStepAlias != "start" ||
		payload.Diagnostics.PrimaryBrowserAction != "browser_runtime action=refresh" ||
		payload.Diagnostics.NextStep != "browser_runtime action=start" {
		t.Fatalf("unexpected translated diagnostics summary: %#v", payload.Diagnostics)
	}
	if payload.SessionBinding == nil ||
		payload.SessionBinding.SessionHealthRecoveryAction != "browser_runtime action=coordinate coordination_goal=ensure" ||
		payload.SessionBinding.Coordination == nil ||
		payload.SessionBinding.Coordination.SyncBrowserAction != "browser_runtime action=coordinate coordination_goal=sync" ||
		payload.SessionBinding.Coordination.RestartBrowserAction != "browser_runtime action=refresh" ||
		payload.SessionBinding.Coordination.TeardownBrowserAction != "browser_runtime action=coordinate coordination_goal=teardown" ||
		payload.SessionBinding.Coordination.PrimaryBrowserAction != "browser_runtime action=clear_session" ||
		payload.SessionBinding.Coordination.NextStep != "browser_runtime action=coordinate coordination_goal=ensure" {
		t.Fatalf("unexpected translated session binding: %#v", payload.SessionBinding)
	}
}
