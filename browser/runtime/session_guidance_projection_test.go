package browserruntime

import "testing"

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsResolverAliases(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			ResolverBlockedBy:        "multiple_candidates_filtered",
			ResolverAmbiguityClass:   "filtered_residual",
			ResolverCandidateKind:    "role_label",
			ResolverRetryDisposition: "manual_only",
			ResolverManualRetryHint:  "add_ordinal",
			ResolverNextStepAlias:    "snapshot",
		},
	)

	if projection.ResolverExplanation == nil ||
		projection.ResolverExplanation.Category != "resolver" ||
		projection.ResolverExplanation.State != "manual_resolution_required" ||
		projection.ResolverExplanation.SummaryCode != "role_label_filtered_residual" ||
		projection.ResolverExplanation.NextStepAlias != "snapshot" ||
		projection.ResolverExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("expected shared guidance projection to surface resolver explanation, got %#v", projection.ResolverExplanation)
	}
	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "resolver" ||
		projection.DiagnosticsExplanation.SummaryCode != "role_label_filtered_residual" {
		t.Fatalf("expected shared guidance projection to surface diagnostics explanation, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.Explanation == nil || projection.Diagnostics == nil || projection.Summary == nil || projection.Display == nil {
		t.Fatalf("expected shared guidance projection to build top-level aliases, got %#v", projection)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsWorkbenchCoordinationAliases(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			IncludeWorkbenchSurface:       true,
			WorkbenchReady:                true,
			WorkbenchSections:             []string{"route", "coordination"},
			WorkbenchPrimaryBrowserAction: "browser action=refresh",
			WorkbenchPrimaryNodeAction:    "nodes action=run",
			WorkbenchNextStep:             "browser action=refresh",
		},
	)

	if projection.WorkbenchDiagnostics == nil ||
		projection.WorkbenchDiagnostics.Category != "coordination" ||
		projection.WorkbenchDiagnostics.State != "action_plan_available" ||
		projection.WorkbenchDiagnostics.SummaryCode != "workbench_action_plan" ||
		projection.WorkbenchDiagnostics.PrimaryBrowserAction != "browser action=refresh" ||
		projection.WorkbenchDiagnostics.PrimaryNodeAction != "nodes action=run" ||
		projection.WorkbenchDiagnostics.NextStep != "browser action=refresh" {
		t.Fatalf("expected shared guidance projection to build workbench coordination diagnostics, got %#v", projection.WorkbenchDiagnostics)
	}
	if projection.WorkbenchSummary == nil || projection.Diagnostics == nil || projection.Summary == nil {
		t.Fatalf("expected shared guidance projection to mirror workbench diagnostics into aliases, got %#v", projection)
	}
	if projection.WorkbenchDisplay == nil ||
		!projection.WorkbenchDisplay.Ready ||
		len(projection.WorkbenchDisplay.Sections) != 2 ||
		projection.WorkbenchDisplay.PrimaryBrowserAction != "browser action=refresh" {
		t.Fatalf("expected shared guidance projection to build workbench display, got %#v", projection.WorkbenchDisplay)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsActionSuccessAliases(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			ActionKind:   "list_tabs",
			ActionStatus: "ok",
		},
	)

	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "tabs" ||
		projection.DiagnosticsExplanation.State != "completed" ||
		projection.DiagnosticsExplanation.SummaryCode != "list_tabs_completed" ||
		projection.DiagnosticsExplanation.PrimaryBrowserAction != "browser action=tabs" ||
		projection.DiagnosticsExplanation.NextStep != "browser action=tabs" {
		t.Fatalf("expected shared guidance projection to build action success diagnostics, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.Explanation == nil ||
		projection.Diagnostics == nil ||
		projection.Summary == nil ||
		projection.Display == nil ||
		projection.Display.Category != "tabs" ||
		projection.Display.SummaryCode != "list_tabs_completed" ||
		projection.Display.PrimaryBrowserAction != "browser action=tabs" ||
		projection.Display.NextStep != "browser action=tabs" {
		t.Fatalf("expected shared guidance projection to mirror action success into top-level aliases, got %#v", projection)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsRuntimeInspectionSuccessAliases(t *testing.T) {
	tests := []struct {
		name        string
		actionKind  string
		summaryCode string
	}{
		{name: "profiles", actionKind: "profiles", summaryCode: "profiles_completed"},
		{name: "sessions", actionKind: "sessions", summaryCode: "sessions_completed"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			projection := BuildSharedSessionBrowserGuidanceProjection(
				SharedSessionBrowserGuidanceProjectionRequest{
					ActionKind:   tc.actionKind,
					ActionStatus: "ok",
				},
			)
			action := "browser action=" + tc.actionKind
			if projection.DiagnosticsExplanation == nil ||
				projection.DiagnosticsExplanation.Category != "inspection" ||
				projection.DiagnosticsExplanation.State != "completed" ||
				projection.DiagnosticsExplanation.SummaryCode != tc.summaryCode ||
				projection.DiagnosticsExplanation.PrimaryBrowserAction != action ||
				projection.DiagnosticsExplanation.NextStep != action {
				t.Fatalf("expected shared guidance projection to build %s success diagnostics, got %#v", tc.actionKind, projection.DiagnosticsExplanation)
			}
			if projection.Explanation == nil ||
				projection.Diagnostics == nil ||
				projection.Summary == nil ||
				projection.Display == nil ||
				projection.Display.Category != "inspection" ||
				projection.Display.SummaryCode != tc.summaryCode ||
				projection.Display.PrimaryBrowserAction != action ||
				projection.Display.NextStep != action {
				t.Fatalf("expected shared guidance projection to mirror %s success into top-level aliases, got %#v", tc.actionKind, projection)
			}
		})
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsHighlightSuccessAliases(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			ActionKind:   "highlight",
			ActionStatus: "highlighted",
		},
	)

	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "interaction" ||
		projection.DiagnosticsExplanation.State != "completed" ||
		projection.DiagnosticsExplanation.SummaryCode != "highlight_completed" ||
		projection.DiagnosticsExplanation.PrimaryBrowserAction != "browser action=highlight" ||
		projection.DiagnosticsExplanation.NextStep != "browser action=highlight" {
		t.Fatalf("expected shared guidance projection to build highlight success diagnostics, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.Explanation == nil ||
		projection.Diagnostics == nil ||
		projection.Summary == nil ||
		projection.Display == nil ||
		projection.Display.Category != "interaction" ||
		projection.Display.SummaryCode != "highlight_completed" ||
		projection.Display.PrimaryBrowserAction != "browser action=highlight" ||
		projection.Display.NextStep != "browser action=highlight" ||
		projection.Summary.Category != "interaction" ||
		projection.Summary.SummaryCode != "highlight_completed" ||
		projection.Diagnostics.Category != "interaction" ||
		projection.Diagnostics.SummaryCode != "highlight_completed" ||
		projection.Explanation.Category != "interaction" ||
		projection.Explanation.SummaryCode != "highlight_completed" {
		t.Fatalf("expected shared guidance projection to mirror highlight success into top-level aliases, got %#v", projection)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsStorageMutationSuccessAliases(t *testing.T) {
	tests := []struct {
		name         string
		actionKind   string
		actionStatus string
		action       string
		category     string
		summaryCode  string
	}{
		{name: "storage_set", actionKind: "storage_set", actionStatus: "updated", action: "browser action=storage_set", category: "storage", summaryCode: "storage_set_completed"},
		{name: "storage_clear", actionKind: "storage_clear", actionStatus: "cleared", action: "browser action=storage_clear", category: "storage", summaryCode: "storage_clear_completed"},
		{name: "cookies_set", actionKind: "cookies_set", actionStatus: "updated", action: "browser action=cookies_set", category: "storage", summaryCode: "cookies_set_completed"},
		{name: "cookies_clear", actionKind: "cookies_clear", actionStatus: "cleared", action: "browser action=cookies_clear", category: "storage", summaryCode: "cookies_clear_completed"},
		{name: "headers_updated", actionKind: "headers", actionStatus: "updated", action: "browser action=headers", category: "network", summaryCode: "headers_updated"},
		{name: "headers_cleared", actionKind: "headers", actionStatus: "cleared", action: "browser action=headers", category: "network", summaryCode: "headers_cleared"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			projection := BuildSharedSessionBrowserGuidanceProjection(
				SharedSessionBrowserGuidanceProjectionRequest{
					ActionKind:   tc.actionKind,
					ActionStatus: tc.actionStatus,
				},
			)

			if projection.DiagnosticsExplanation == nil ||
				projection.DiagnosticsExplanation.Category != tc.category ||
				projection.DiagnosticsExplanation.State != "completed" ||
				projection.DiagnosticsExplanation.SummaryCode != tc.summaryCode ||
				projection.DiagnosticsExplanation.PrimaryBrowserAction != tc.action ||
				projection.DiagnosticsExplanation.NextStep != tc.action {
				t.Fatalf("expected shared guidance projection to build %s success diagnostics, got %#v", tc.actionKind, projection.DiagnosticsExplanation)
			}
			if projection.Explanation == nil ||
				projection.Diagnostics == nil ||
				projection.Summary == nil ||
				projection.Display == nil ||
				projection.Display.Category != tc.category ||
				projection.Display.SummaryCode != tc.summaryCode ||
				projection.Display.PrimaryBrowserAction != tc.action ||
				projection.Display.NextStep != tc.action ||
				projection.Summary.Category != tc.category ||
				projection.Summary.SummaryCode != tc.summaryCode ||
				projection.Diagnostics.Category != tc.category ||
				projection.Diagnostics.SummaryCode != tc.summaryCode ||
				projection.Explanation.Category != tc.category ||
				projection.Explanation.SummaryCode != tc.summaryCode {
				t.Fatalf("expected shared guidance projection to mirror %s success into top-level aliases, got %#v", tc.actionKind, projection)
			}
		})
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsBrowserReadSuccessAliases(t *testing.T) {
	tests := []struct {
		name         string
		actionKind   string
		actionStatus string
		action       string
		category     string
		summaryCode  string
	}{
		{name: "response_body", actionKind: "response_body", actionStatus: "ok", action: "browser action=response_body", category: "content", summaryCode: "response_body_collected"},
		{name: "errors", actionKind: "errors", actionStatus: "ok", action: "browser action=errors", category: "observability", summaryCode: "errors_collected"},
		{name: "errors_clear", actionKind: "errors", actionStatus: "cleared", action: "browser action=errors", category: "observability", summaryCode: "errors_cleared"},
		{name: "cookies", actionKind: "cookies", actionStatus: "ok", action: "browser action=cookies", category: "storage", summaryCode: "cookies_collected"},
		{name: "storage", actionKind: "storage", actionStatus: "ok", action: "browser action=storage", category: "storage", summaryCode: "storage_collected"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			projection := BuildSharedSessionBrowserGuidanceProjection(
				SharedSessionBrowserGuidanceProjectionRequest{
					ActionKind:   tc.actionKind,
					ActionStatus: tc.actionStatus,
				},
			)
			if projection.DiagnosticsExplanation == nil ||
				projection.DiagnosticsExplanation.Category != tc.category ||
				projection.DiagnosticsExplanation.State != "completed" ||
				projection.DiagnosticsExplanation.SummaryCode != tc.summaryCode ||
				projection.DiagnosticsExplanation.PrimaryBrowserAction != tc.action ||
				projection.DiagnosticsExplanation.NextStep != tc.action {
				t.Fatalf("expected shared guidance projection to build %s success diagnostics, got %#v", tc.actionKind, projection.DiagnosticsExplanation)
			}
			if projection.Explanation == nil ||
				projection.Diagnostics == nil ||
				projection.Summary == nil ||
				projection.Display == nil ||
				projection.Display.Category != tc.category ||
				projection.Display.SummaryCode != tc.summaryCode ||
				projection.Display.PrimaryBrowserAction != tc.action ||
				projection.Display.NextStep != tc.action ||
				projection.Summary.Category != tc.category ||
				projection.Summary.SummaryCode != tc.summaryCode ||
				projection.Diagnostics.Category != tc.category ||
				projection.Diagnostics.SummaryCode != tc.summaryCode ||
				projection.Explanation.Category != tc.category ||
				projection.Explanation.SummaryCode != tc.summaryCode {
				t.Fatalf("expected shared guidance projection to mirror %s success into top-level aliases, got %#v", tc.actionKind, projection)
			}
		})
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsDialogArmedSuccessAliases(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			ActionKind:   "dialog",
			ActionStatus: "armed",
		},
	)
	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "interaction" ||
		projection.DiagnosticsExplanation.State != "started" ||
		projection.DiagnosticsExplanation.SummaryCode != "dialog_armed" ||
		projection.DiagnosticsExplanation.PrimaryBrowserAction != "browser action=dialog" ||
		projection.DiagnosticsExplanation.NextStep != "browser action=dialog" {
		t.Fatalf("expected shared guidance projection to build dialog armed diagnostics, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.Explanation == nil ||
		projection.Diagnostics == nil ||
		projection.Summary == nil ||
		projection.Display == nil ||
		projection.Display.Category != "interaction" ||
		projection.Display.State != "started" ||
		projection.Display.SummaryCode != "dialog_armed" ||
		projection.Display.PrimaryBrowserAction != "browser action=dialog" ||
		projection.Display.NextStep != "browser action=dialog" ||
		projection.Summary.Category != "interaction" ||
		projection.Summary.State != "started" ||
		projection.Summary.SummaryCode != "dialog_armed" ||
		projection.Diagnostics.Category != "interaction" ||
		projection.Diagnostics.State != "started" ||
		projection.Diagnostics.SummaryCode != "dialog_armed" ||
		projection.Explanation.Category != "interaction" ||
		projection.Explanation.State != "started" ||
		projection.Explanation.SummaryCode != "dialog_armed" {
		t.Fatalf("expected shared guidance projection to mirror dialog armed success into top-level aliases, got %#v", projection)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsRepairSuccessAliases(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			ActionKind:   "repair",
			ActionStatus: "ok",
		},
	)
	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "coordination" ||
		projection.DiagnosticsExplanation.State != "completed" ||
		projection.DiagnosticsExplanation.SummaryCode != "repair_completed" ||
		projection.DiagnosticsExplanation.PrimaryBrowserAction != "browser action=repair" ||
		projection.DiagnosticsExplanation.NextStep != "browser action=repair" {
		t.Fatalf("expected shared guidance projection to build repair success diagnostics, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.Explanation == nil ||
		projection.Diagnostics == nil ||
		projection.Summary == nil ||
		projection.Display == nil ||
		projection.Display.Category != "coordination" ||
		projection.Display.State != "completed" ||
		projection.Display.SummaryCode != "repair_completed" ||
		projection.Display.PrimaryBrowserAction != "browser action=repair" ||
		projection.Display.NextStep != "browser action=repair" ||
		projection.Summary.Category != "coordination" ||
		projection.Summary.State != "completed" ||
		projection.Summary.SummaryCode != "repair_completed" ||
		projection.Diagnostics.Category != "coordination" ||
		projection.Diagnostics.SummaryCode != "repair_completed" ||
		projection.Explanation.Category != "coordination" ||
		projection.Explanation.SummaryCode != "repair_completed" {
		t.Fatalf("expected shared guidance projection to mirror repair success into top-level aliases, got %#v", projection)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsRuntimeSessionAndLifecycleSuccessAliases(t *testing.T) {
	for _, contract := range sharedSessionBrowserRuntimeSessionActionSuccessContracts {
		t.Run(contract.ActionKind, func(t *testing.T) {
			projection := BuildSharedSessionBrowserGuidanceProjection(
				SharedSessionBrowserGuidanceProjectionRequest{
					ActionKind:   contract.ActionKind,
					ActionStatus: contract.ActionStatus,
				},
			)
			action := "browser action=" + sharedSessionBrowserActionLabelForKind(contract.ActionKind)
			if projection.DiagnosticsExplanation == nil ||
				projection.DiagnosticsExplanation.Category != contract.Category ||
				projection.DiagnosticsExplanation.State != contract.State ||
				projection.DiagnosticsExplanation.SummaryCode != contract.SummaryCode ||
				projection.DiagnosticsExplanation.PrimaryBrowserAction != action ||
				projection.DiagnosticsExplanation.NextStep != action {
				t.Fatalf("expected shared guidance projection to build %s success diagnostics, got %#v", contract.ActionKind, projection.DiagnosticsExplanation)
			}
			if projection.Explanation == nil ||
				projection.Diagnostics == nil ||
				projection.Summary == nil ||
				projection.Display == nil ||
				projection.Display.Category != contract.Category ||
				projection.Display.State != contract.State ||
				projection.Display.SummaryCode != contract.SummaryCode ||
				projection.Display.PrimaryBrowserAction != action ||
				projection.Display.NextStep != action ||
				projection.Summary.Category != contract.Category ||
				projection.Summary.State != contract.State ||
				projection.Summary.SummaryCode != contract.SummaryCode ||
				projection.Diagnostics.Category != contract.Category ||
				projection.Diagnostics.SummaryCode != contract.SummaryCode ||
				projection.Explanation.Category != contract.Category ||
				projection.Explanation.SummaryCode != contract.SummaryCode {
				t.Fatalf("expected shared guidance projection to mirror %s success into top-level aliases, got %#v", contract.ActionKind, projection)
			}
		})
	}
}

func TestRuntimeActionsAreClassifiedForActionSuccessProjection(t *testing.T) {
	allActions := BrowserCapabilities{
		RuntimeWorkbench:    true,
		RuntimePrepare:      true,
		RuntimeCoordinate:   true,
		RuntimeStart:        true,
		RuntimeRestart:      true,
		RuntimeStop:         true,
		RuntimeCreate:       true,
		RuntimeDelete:       true,
		RuntimeSelect:       true,
		RuntimeClear:        true,
		RuntimeClearSession: true,
		RuntimeSyncSession:  true,
		RuntimeSelectTarget: true,
		RuntimeClearTarget:  true,
		RuntimeList:         true,
		RuntimeSessions:     true,
	}.SupportedRuntimeActions()
	allActions = append(allActions, "doctor", "repair")

	success := map[string]sharedSessionBrowserActionSuccessContract{}
	for _, contract := range sharedSessionBrowserRuntimeSessionActionSuccessContracts {
		if _, exists := success[contract.ActionKind]; exists {
			t.Fatalf("duplicate runtime/session action success contract for %q", contract.ActionKind)
		}
		success[contract.ActionKind] = contract
	}
	observationOnly := map[string]bool{}
	for _, action := range sharedSessionBrowserRuntimeObservationOnlyActionKinds {
		if _, conflicts := success[action]; conflicts {
			t.Fatalf("runtime action %q cannot be both success-projected and observation-only", action)
		}
		observationOnly[action] = true
	}

	for _, action := range allActions {
		contract, hasSuccess := success[action]
		switch {
		case hasSuccess:
			projection := BuildSharedSessionBrowserGuidanceProjection(
				SharedSessionBrowserGuidanceProjectionRequest{
					ActionKind:   action,
					ActionStatus: contract.ActionStatus,
				},
			)
			if projection.Summary == nil || projection.Display == nil || projection.Diagnostics == nil ||
				projection.Summary.SummaryCode != contract.SummaryCode ||
				projection.Display.SummaryCode != contract.SummaryCode ||
				projection.Diagnostics.SummaryCode != contract.SummaryCode {
				t.Fatalf("expected runtime action %q to project stable success aliases from contract %#v, got %#v", action, contract, projection)
			}
		case observationOnly[action]:
			projection := BuildSharedSessionBrowserGuidanceProjection(
				SharedSessionBrowserGuidanceProjectionRequest{
					ActionKind:   action,
					ActionStatus: "ok",
				},
			)
			if projection.Summary != nil || projection.Display != nil || projection.Diagnostics != nil || projection.DiagnosticsExplanation != nil {
				t.Fatalf("expected runtime action %q to stay observation-only without action-success shells, got %#v", action, projection)
			}
		default:
			t.Fatalf("runtime action %q must be classified as success-projected or observation-only", action)
		}
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsSettingsActionSuccessAliases(t *testing.T) {
	testCases := []struct {
		name         string
		actionKind   string
		actionStatus string
		category     string
		summaryCode  string
	}{
		{name: "offline_updated", actionKind: "offline", actionStatus: "updated", category: "network", summaryCode: "offline_updated"},
		{name: "credentials_cleared", actionKind: "credentials", actionStatus: "cleared", category: "auth", summaryCode: "credentials_cleared"},
		{name: "geolocation_updated", actionKind: "geolocation", actionStatus: "updated", category: "settings", summaryCode: "geolocation_updated"},
		{name: "device_cleared", actionKind: "device", actionStatus: "cleared", category: "settings", summaryCode: "device_cleared"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			projection := BuildSharedSessionBrowserGuidanceProjection(
				SharedSessionBrowserGuidanceProjectionRequest{
					ActionKind:   tc.actionKind,
					ActionStatus: tc.actionStatus,
				},
			)
			if projection.DiagnosticsExplanation == nil ||
				projection.DiagnosticsExplanation.Category != tc.category ||
				projection.DiagnosticsExplanation.State != "completed" ||
				projection.DiagnosticsExplanation.SummaryCode != tc.summaryCode ||
				projection.DiagnosticsExplanation.PrimaryBrowserAction != "browser action="+tc.actionKind ||
				projection.DiagnosticsExplanation.NextStep != "browser action="+tc.actionKind {
				t.Fatalf("expected shared guidance projection to build %s diagnostics, got %#v", tc.name, projection.DiagnosticsExplanation)
			}
			if projection.Explanation == nil ||
				projection.Diagnostics == nil ||
				projection.Summary == nil ||
				projection.Display == nil ||
				projection.Display.Category != tc.category ||
				projection.Display.SummaryCode != tc.summaryCode ||
				projection.Summary.Category != tc.category ||
				projection.Summary.SummaryCode != tc.summaryCode ||
				projection.Diagnostics.Category != tc.category ||
				projection.Diagnostics.SummaryCode != tc.summaryCode ||
				projection.Explanation.Category != tc.category ||
				projection.Explanation.SummaryCode != tc.summaryCode {
				t.Fatalf("expected shared guidance projection to mirror %s into top-level aliases, got %#v", tc.name, projection)
			}
		})
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionPrefersReviewExplanation(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			IncludeWorkbenchSurface:       true,
			WorkbenchReady:                true,
			WorkbenchSections:             []string{"route", "coordination"},
			WorkbenchPrimaryBrowserAction: "browser action=tabs",
			WorkbenchPrimaryNodeAction:    "nodes action=run_status",
			WorkbenchNextStep:             "browser action=tabs",
			Routes: []SharedSessionBrowserRouteCoordinationInput{
				{FollowPolicyState: "popup_review_required"},
			},
			ResolverBlockedBy:        "multiple_candidates_filtered",
			ResolverAmbiguityClass:   "filtered_residual",
			ResolverCandidateKind:    "label",
			ResolverRetryDisposition: "manual_only",
			ResolverManualRetryHint:  "add_ordinal",
			ResolverNextStepAlias:    "snapshot",
		},
	)

	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "review" ||
		projection.DiagnosticsExplanation.State != "manual_confirmation_required" ||
		projection.DiagnosticsExplanation.SummaryCode != "popup_review_required" ||
		projection.DiagnosticsExplanation.NextStepAlias != "tabs" ||
		projection.DiagnosticsExplanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("expected shared guidance projection to prefer review explanation, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.WorkbenchSummary == nil ||
		projection.WorkbenchSummary.Category != "review" ||
		projection.WorkbenchSummary.PrimaryBrowserAction != "browser action=tabs" ||
		projection.WorkbenchSummary.PrimaryNodeAction != "nodes action=run_status" {
		t.Fatalf("expected shared guidance projection to combine review explanation with workbench actions, got %#v", projection.WorkbenchSummary)
	}
	if projection.Display == nil ||
		projection.Display.Category != "review" ||
		projection.Display.NextStepAlias != "tabs" {
		t.Fatalf("expected shared guidance projection to mirror review into top-level display, got %#v", projection.Display)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionPrefersResolverOverActionSuccess(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			ActionKind:               "open",
			ActionStatus:             "opened",
			ResolverBlockedBy:        "multiple_candidates_filtered",
			ResolverAmbiguityClass:   "filtered_residual",
			ResolverCandidateKind:    "label",
			ResolverRetryDisposition: "manual_only",
			ResolverManualRetryHint:  "add_ordinal",
			ResolverNextStepAlias:    "snapshot",
		},
	)

	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "resolver" ||
		projection.DiagnosticsExplanation.SummaryCode != "label_filtered_residual" ||
		projection.DiagnosticsExplanation.NextStepAlias != "snapshot" {
		t.Fatalf("expected resolver guidance to take precedence over action success, got %#v", projection.DiagnosticsExplanation)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsActionabilityFailureAliases(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			ActionKind:                   "click",
			ActionStatus:                 "failed",
			ActionabilityStatus:          BrowserActionabilityStatusFailed,
			ActionabilityFailedCheck:     "visible",
			ActionabilityFailureReason:   "actionability_visible_failed",
			ActionabilityManualRetryHint: "choose_visible_target",
			ActionabilityRecoveryAction:  "browser action=snapshot",
		},
	)

	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "actionability" ||
		projection.DiagnosticsExplanation.State != "actionability_failed" ||
		projection.DiagnosticsExplanation.SummaryCode != "actionability_visible_failed" ||
		projection.DiagnosticsExplanation.NextStepAlias != "snapshot" ||
		projection.DiagnosticsExplanation.ManualRetryHint != "choose_visible_target" ||
		projection.DiagnosticsExplanation.PrimaryBrowserAction != "browser action=snapshot" ||
		projection.DiagnosticsExplanation.NextStep != "browser action=snapshot" {
		t.Fatalf("expected actionability failure diagnostics, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.Explanation == nil ||
		projection.Diagnostics == nil ||
		projection.Summary == nil ||
		projection.Display == nil ||
		projection.Display.Category != "actionability" ||
		projection.Display.State != "actionability_failed" ||
		projection.Display.SummaryCode != "actionability_visible_failed" ||
		projection.Display.PrimaryBrowserAction != "browser action=snapshot" ||
		projection.Display.NextStep != "browser action=snapshot" {
		t.Fatalf("expected actionability failure to mirror into top-level aliases, got %#v", projection)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsPostActionWaitFailureAliases(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			IncludeWorkbenchSurface:      true,
			WorkbenchReady:               true,
			WorkbenchSections:            []string{"diagnostics"},
			ActionKind:                   "click",
			ActionStatus:                 "failed",
			ActionabilityStatus:          BrowserActionabilityStatusFailed,
			ActionabilityFailedCheck:     "navigation_wait",
			ActionabilityFailureReason:   "actionability_navigation_wait_failed",
			ActionabilityManualRetryHint: "wait_for_navigation_or_snapshot",
		},
	)

	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "actionability" ||
		projection.DiagnosticsExplanation.State != "post_action_wait_failed" ||
		projection.DiagnosticsExplanation.SummaryCode != "actionability_navigation_wait_failed" ||
		projection.DiagnosticsExplanation.NextStepAlias != "wait" ||
		projection.DiagnosticsExplanation.ManualRetryHint != "wait_for_navigation_or_snapshot" ||
		projection.DiagnosticsExplanation.PrimaryBrowserAction != "browser action=wait" ||
		projection.DiagnosticsExplanation.NextStep != "browser action=wait" {
		t.Fatalf("expected post-action wait diagnostics, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.WorkbenchDiagnostics == nil ||
		projection.WorkbenchDiagnostics.Category != "actionability" ||
		projection.WorkbenchDiagnostics.State != "post_action_wait_failed" ||
		projection.WorkbenchDiagnostics.PrimaryBrowserAction != "browser action=wait" ||
		projection.WorkbenchDiagnostics.NextStep != "browser action=wait" {
		t.Fatalf("expected workbench diagnostics to preserve post-action wait guidance, got %#v", projection.WorkbenchDiagnostics)
	}
	if projection.Display == nil ||
		!projection.Display.Ready ||
		projection.Display.State != "post_action_wait_failed" ||
		projection.Display.PrimaryBrowserAction != "browser action=wait" {
		t.Fatalf("expected display to mirror post-action wait guidance, got %#v", projection.Display)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionPrefersActionabilityOverResolverGuidance(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			ActionKind:                  "click",
			ActionStatus:                "failed",
			ActionabilityStatus:         BrowserActionabilityStatusFailed,
			ActionabilityFailedCheck:    "receives_events",
			ActionabilityFailureReason:  "actionability_receives_events_failed",
			ActionabilityRecoveryAction: "browser action=snapshot",
			ResolverBlockedBy:           "multiple_candidates_filtered",
			ResolverAmbiguityClass:      "filtered_residual",
			ResolverCandidateKind:       "label",
			ResolverRetryDisposition:    "manual_only",
			ResolverManualRetryHint:     "add_ordinal",
			ResolverNextStepAlias:       "snapshot",
		},
	)

	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "actionability" ||
		projection.DiagnosticsExplanation.SummaryCode != "actionability_receives_events_failed" {
		t.Fatalf("expected concrete actionability failure to take precedence over resolver guidance, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.ResolverExplanation == nil ||
		projection.ResolverExplanation.Category != "resolver" ||
		projection.ResolverExplanation.SummaryCode != "label_filtered_residual" {
		t.Fatalf("expected resolver explanation to remain available, got %#v", projection.ResolverExplanation)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionPrefersReviewOverActionability(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			ReviewStatus:                "review_required",
			ReviewDecision:              "session_target_popup_review_required",
			ActionabilityStatus:         BrowserActionabilityStatusFailed,
			ActionabilityFailedCheck:    "visible",
			ActionabilityFailureReason:  "actionability_visible_failed",
			ActionabilityRecoveryAction: "browser action=snapshot",
		},
	)

	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "review" ||
		projection.DiagnosticsExplanation.SummaryCode != "popup_review_required" {
		t.Fatalf("expected explicit review to take precedence over actionability diagnostics, got %#v", projection.DiagnosticsExplanation)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionBuildsExplicitReviewAliases(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			ReviewStatus:   "review_required",
			ReviewDecision: "session_target_popup_review_required",
		},
	)

	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.Category != "review" ||
		projection.DiagnosticsExplanation.State != "manual_confirmation_required" ||
		projection.DiagnosticsExplanation.SummaryCode != "popup_review_required" ||
		projection.DiagnosticsExplanation.NextStepAlias != "" ||
		projection.DiagnosticsExplanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("expected shared guidance projection to build explicit review diagnostics, got %#v", projection.DiagnosticsExplanation)
	}
	if projection.Explanation == nil ||
		projection.Diagnostics == nil ||
		projection.Summary == nil ||
		projection.Display == nil ||
		projection.Display.Category != "review" ||
		projection.Display.SummaryCode != "popup_review_required" {
		t.Fatalf("expected shared guidance projection to mirror explicit review into top-level aliases, got %#v", projection)
	}
}

func TestBuildSharedSessionBrowserGuidanceProjectionPrefersExplicitReviewOverRoutes(t *testing.T) {
	projection := BuildSharedSessionBrowserGuidanceProjection(
		SharedSessionBrowserGuidanceProjectionRequest{
			ReviewStatus:   "review_required",
			ReviewDecision: "navigate_redirect_review_required",
			Routes: []SharedSessionBrowserRouteCoordinationInput{
				{FollowPolicyState: "popup_review_required"},
			},
		},
	)

	if projection.DiagnosticsExplanation == nil ||
		projection.DiagnosticsExplanation.SummaryCode != "redirect_review_required" ||
		projection.DiagnosticsExplanation.NextStepAlias != "" {
		t.Fatalf("expected explicit review to take precedence over route review explanation, got %#v", projection.DiagnosticsExplanation)
	}
}

func TestBuildSharedSessionBrowserExplanationAliasFromRequestPrefersWorkbench(t *testing.T) {
	summary := BuildSharedSessionBrowserExplanationAliasFromRequest(
		SharedSessionBrowserExplanationAliasRequest{
			Workbench: &SharedSessionBrowserSummary{
				Category:    "resolver",
				State:       "manual_resolution_required",
				SummaryCode: "workbench_override",
			},
			Diagnostics: &SharedSessionBrowserSummary{
				Category:    "resolver",
				State:       "manual_resolution_required",
				SummaryCode: "diagnostics_fallback",
			},
			Resolver: &SharedSessionBrowserSummary{
				Category:    "resolver",
				State:       "manual_resolution_required",
				SummaryCode: "resolver_fallback",
			},
		},
	)
	if summary == nil {
		t.Fatalf("expected shared explanation alias")
	}
	if summary.SummaryCode != "workbench_override" {
		t.Fatalf("expected workbench explanation to remain primary, got %#v", summary)
	}
}

func TestBuildSharedSessionBrowserDiagnosticsAliasFromRequestFallsBackToSummary(t *testing.T) {
	summary := BuildSharedSessionBrowserDiagnosticsAliasFromRequest(
		SharedSessionBrowserDiagnosticsAliasRequest{
			Summary: &SharedSessionBrowserSummary{
				Category:    "resolver",
				State:       "manual_resolution_required",
				SummaryCode: "summary_fallback",
			},
		},
	)
	if summary == nil {
		t.Fatalf("expected shared diagnostics alias")
	}
	if summary.SummaryCode != "summary_fallback" {
		t.Fatalf("expected diagnostics alias to fall back to summary, got %#v", summary)
	}
}
