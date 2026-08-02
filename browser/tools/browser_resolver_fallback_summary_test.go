package tools

import (
	"encoding/json"
	"reflect"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func TestBrowserActResultWithResolverFallbackSummary(t *testing.T) {
	outcome := &agentxbrowserruntime.BrowserElementResolverOutcome{
		Status:                        "matched",
		FallbackFromKind:              "label",
		FallbackFromIndex:             0,
		FallbackFromCandidateStrength: "medium",
		FallbackFromSpecificityFields: []string{"tag", "type"},
		BlockedBy:                     "multiple_candidates_filtered",
		AmbiguityClass:                "filtered_residual",
		CandidateKind:                 "label",
		CandidateStrength:             "medium",
		RetryDisposition:              "manual_only",
		ManualRetryHint:               "add_ordinal",
		NextStepAlias:                 "snapshot",
	}
	result := browserActResultWithResolverFallbackSummary(BrowserActResult{
		Kind:            "type",
		ResolverOutcome: outcome,
	})
	if !result.ResolvedViaFallback {
		t.Fatalf("expected resolved_via_fallback to be true: %#v", result)
	}
	if result.ResolverFallbackKind != "label" {
		t.Fatalf("ResolverFallbackKind = %q, want label", result.ResolverFallbackKind)
	}
	if result.ResolverFallbackIndex == nil || *result.ResolverFallbackIndex != 0 {
		t.Fatalf("ResolverFallbackIndex = %#v, want 0", result.ResolverFallbackIndex)
	}
	if result.ResolverFallbackCandidateStrength != "medium" {
		t.Fatalf("ResolverFallbackCandidateStrength = %q, want medium", result.ResolverFallbackCandidateStrength)
	}
	if result.ResolverFallbackBlockedBy != "" ||
		result.ResolverFallbackAmbiguityClass != "" ||
		result.ResolverFallbackManualRetryHint != "" {
		t.Fatalf("expected fallback diagnostics to stay empty without fallback_from_* detail, got %#v", result)
	}
	if !reflect.DeepEqual(result.ResolverFallbackSpecificityFields, []string{"tag", "type"}) {
		t.Fatalf("ResolverFallbackSpecificityFields = %#v", result.ResolverFallbackSpecificityFields)
	}
	if result.ResolverBlockedBy != "multiple_candidates_filtered" ||
		result.ResolverAmbiguityClass != "filtered_residual" ||
		result.ResolverCandidateKind != "label" ||
		result.ResolverCandidateStrength != "medium" ||
		result.ResolverRetryDisposition != "manual_only" ||
		result.ResolverManualRetryHint != "add_ordinal" ||
		result.ResolverNextStepAlias != "snapshot" {
		t.Fatalf("unexpected resolver guidance aliases: %#v", result)
	}
	if result.Actionability == nil || result.Actionability.Status != agentxbrowserruntime.BrowserActionabilityStatusPartial {
		t.Fatalf("expected resolver fallback summary to attach partial actionability, got %#v", result.Actionability)
	}
	if result.FailureEvidence != nil {
		t.Fatalf("expected matched action not to attach failure evidence, got %#v", result.FailureEvidence)
	}

	outcome.FallbackFromSpecificityFields[0] = "mutated"
	if reflect.DeepEqual(result.ResolverFallbackSpecificityFields, outcome.FallbackFromSpecificityFields) {
		t.Fatalf("expected resolver fallback specificity fields to be copied, got %#v", result.ResolverFallbackSpecificityFields)
	}
}

func TestMarshalBrowserActPayloadAddsActionabilityFailureEvidence(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:           "click",
		Status:         "unresolved",
		Ref:            "css1:#submit",
		Snapshot:       "button Submit",
		SnapshotFormat: "role",
		SnapshotRefs:   "role",
		Elements:       []BrowserSnapshotElement{{Index: 1, Role: "button", Label: "Submit"}},
		ResolverOutcome: &agentxbrowserruntime.BrowserElementResolverOutcome{
			Status:         "unresolved",
			BlockedBy:      "multiple_candidates",
			RecoveryAction: "browser action=snapshot",
			Note:           "multiple submit buttons matched",
		},
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload: %v", err)
	}
	var payload struct {
		Actionability   *agentxbrowserruntime.BrowserActionabilityReport   `json:"actionability"`
		FailureEvidence *agentxbrowserruntime.BrowserActionFailureEvidence `json:"failure_evidence"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_act payload: %v", err)
	}
	if payload.Actionability == nil ||
		payload.Actionability.Status != agentxbrowserruntime.BrowserActionabilityStatusFailed ||
		payload.Actionability.FailedCheck != "resolve_target" {
		t.Fatalf("expected failed actionability in payload, got %#v", payload.Actionability)
	}
	if payload.FailureEvidence == nil ||
		payload.FailureEvidence.ReasonCode != "resolver_unresolved_multiple_candidates" ||
		payload.FailureEvidence.RecoveryAction != "browser action=snapshot" ||
		!payload.FailureEvidence.SnapshotAvailable ||
		payload.FailureEvidence.SnapshotElementCount != 1 {
		t.Fatalf("unexpected failure evidence in payload: %#v", payload.FailureEvidence)
	}
	if payload.FailureEvidence.Artifact == nil ||
		payload.FailureEvidence.Artifact.Kind != "trace_like" ||
		payload.FailureEvidence.Artifact.ReasonCode != "resolver_unresolved_multiple_candidates" ||
		payload.FailureEvidence.Artifact.TargetKind != "ref" ||
		payload.FailureEvidence.Artifact.Target != "css1:#submit" ||
		payload.FailureEvidence.Artifact.SnapshotRefs != "role" ||
		payload.FailureEvidence.Artifact.SnapshotElementCount != 1 ||
		payload.FailureEvidence.Artifact.ResolverOutcome == nil {
		t.Fatalf("unexpected trace-like failure artifact in payload: %#v", payload.FailureEvidence.Artifact)
	}
}

func TestMarshalBrowserActPayloadBuildsActionabilityFailureDiagnostics(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:   "click",
		Status: "failed",
		Ref:    "css1:#submit",
		Actionability: &agentxbrowserruntime.BrowserActionabilityReport{
			Action:          "click",
			Status:          agentxbrowserruntime.BrowserActionabilityStatusFailed,
			TargetKind:      "ref",
			Target:          "css1:#submit",
			FailedCheck:     "visible",
			FailureReason:   "actionability_visible_failed",
			ManualRetryHint: "choose_visible_target",
			RecoveryAction:  "browser action=snapshot",
		},
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_act payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "actionability" ||
		payload.DiagnosticsExplanation.State != "actionability_failed" ||
		payload.DiagnosticsExplanation.SummaryCode != "actionability_visible_failed" ||
		payload.DiagnosticsExplanation.NextStepAlias != "snapshot" ||
		payload.DiagnosticsExplanation.ManualRetryHint != "choose_visible_target" {
		t.Fatalf("unexpected actionability diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "actionability" ||
		payload.Summary.State != "actionability_failed" ||
		payload.Summary.PrimaryBrowserAction != "browser action=snapshot" ||
		payload.Summary.NextStep != "browser action=snapshot" {
		t.Fatalf("unexpected actionability top-level summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "actionability" ||
		payload.Display.PrimaryBrowserAction != "browser action=snapshot" ||
		payload.Display.NextStep != "browser action=snapshot" {
		t.Fatalf("unexpected actionability display: %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "actionability" ||
		payload.Surface.SummaryCode != "actionability_visible_failed" ||
		payload.Surface.PrimaryBrowserAction != "browser action=snapshot" {
		t.Fatalf("unexpected actionability surface: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Category != "actionability" ||
		payload.View.SummaryCode != "actionability_visible_failed" ||
		payload.View.PrimaryBrowserAction != "browser action=snapshot" {
		t.Fatalf("unexpected actionability view: %#v", payload.View)
	}
}

func TestMarshalBrowserActPayloadBuildsPostActionWaitDiagnostics(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:   "click",
		Status: "timeout",
		Ref:    "css1:#submit",
		Actionability: &agentxbrowserruntime.BrowserActionabilityReport{
			Action:          "click",
			Status:          agentxbrowserruntime.BrowserActionabilityStatusFailed,
			TargetKind:      "ref",
			Target:          "css1:#submit",
			FailedCheck:     "navigation_wait",
			FailureReason:   "actionability_navigation_wait_failed",
			ManualRetryHint: "wait_for_navigation_or_snapshot",
		},
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_act payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "actionability" ||
		payload.DiagnosticsExplanation.State != "post_action_wait_failed" ||
		payload.DiagnosticsExplanation.SummaryCode != "actionability_navigation_wait_failed" ||
		payload.DiagnosticsExplanation.NextStepAlias != "wait" ||
		payload.DiagnosticsExplanation.ManualRetryHint != "wait_for_navigation_or_snapshot" {
		t.Fatalf("unexpected post-action wait diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil ||
		payload.Summary.State != "post_action_wait_failed" ||
		payload.Summary.PrimaryBrowserAction != "browser action=wait" ||
		payload.Summary.NextStep != "browser action=wait" {
		t.Fatalf("unexpected post-action wait summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.State != "post_action_wait_failed" ||
		payload.Display.PrimaryBrowserAction != "browser action=wait" {
		t.Fatalf("unexpected post-action wait display: %#v", payload.Display)
	}
}

func TestBrowserActResultWithResolverFallbackSummaryClearsEmptyFallback(t *testing.T) {
	index := 2
	result := browserActResultWithResolverFallbackSummary(BrowserActResult{
		Kind:                              "click",
		ResolvedViaFallback:               true,
		ResolverFallbackKind:              "href",
		ResolverFallbackIndex:             &index,
		ResolverFallbackCandidateStrength: "strong",
		ResolverFallbackSpecificityFields: []string{"href"},
	})
	if result.ResolvedViaFallback {
		t.Fatalf("expected empty fallback summary to clear resolved_via_fallback: %#v", result)
	}
	if result.ResolverFallbackKind != "" ||
		result.ResolverFallbackIndex != nil ||
		result.ResolverFallbackCandidateStrength != "" ||
		result.ResolverFallbackBlockedBy != "" ||
		result.ResolverFallbackAmbiguityClass != "" ||
		result.ResolverFallbackManualRetryHint != "" ||
		result.ResolverFallbackSpecificityFields != nil {
		t.Fatalf("expected empty fallback summary to clear alias fields: %#v", result)
	}
	if result.ResolverBlockedBy != "" ||
		result.ResolverAmbiguityClass != "" ||
		result.ResolverCandidateKind != "" ||
		result.ResolverCandidateStrength != "" ||
		result.ResolverRetryDisposition != "" ||
		result.ResolverManualRetryHint != "" ||
		result.ResolverNextStepAlias != "" {
		t.Fatalf("expected empty fallback summary to clear resolver guidance aliases: %#v", result)
	}
}

func TestMarshalBrowserOpenPayloadAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserOpenPayload(browserOpenToolPayload{
		URL:     "https://93.184.216.34",
		Backend: "proxy-open",
		Status:  "opened",
	})
	if err != nil {
		t.Fatalf("marshalBrowserOpenPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_open payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "navigation" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_open diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "navigation" ||
		payload.Explanation.State != "completed" ||
		payload.Explanation.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_open explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "navigation" ||
		payload.Diagnostics.State != "completed" ||
		payload.Diagnostics.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_open diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "navigation" ||
		payload.Summary.State != "completed" ||
		payload.Summary.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_open summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "navigation" ||
		payload.Display.State != "completed" ||
		payload.Display.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_open display: %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "navigation" ||
		payload.Surface.State != "completed" ||
		payload.Surface.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_open surface: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Category != "navigation" ||
		payload.View.State != "completed" ||
		payload.View.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_open view: %#v", payload.View)
	}
}

func TestMarshalBrowserNavigatePayloadAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserNavigatePayload(browserNavigateToolPayload{
		URL:     "https://93.184.216.34",
		Backend: "proxy-navigate",
		Status:  "navigated",
	})
	if err != nil {
		t.Fatalf("marshalBrowserNavigatePayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_navigate payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "navigation" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "navigate_completed" {
		t.Fatalf("unexpected browser_navigate diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "navigation" ||
		payload.Explanation.State != "completed" ||
		payload.Explanation.SummaryCode != "navigate_completed" {
		t.Fatalf("unexpected browser_navigate explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "navigation" ||
		payload.Diagnostics.State != "completed" ||
		payload.Diagnostics.SummaryCode != "navigate_completed" {
		t.Fatalf("unexpected browser_navigate diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "navigation" ||
		payload.Summary.State != "completed" ||
		payload.Summary.SummaryCode != "navigate_completed" {
		t.Fatalf("unexpected browser_navigate summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "navigation" ||
		payload.Display.State != "completed" ||
		payload.Display.SummaryCode != "navigate_completed" {
		t.Fatalf("unexpected browser_navigate display: %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "navigation" ||
		payload.Surface.State != "completed" ||
		payload.Surface.SummaryCode != "navigate_completed" {
		t.Fatalf("unexpected browser_navigate surface: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Category != "navigation" ||
		payload.View.State != "completed" ||
		payload.View.SummaryCode != "navigate_completed" {
		t.Fatalf("unexpected browser_navigate view: %#v", payload.View)
	}
}

func TestMarshalBrowserExtractPayloadAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserExtractPayload(browserExtractToolPayload{
		Backend: "proxy-extract",
		Status:  "extracted",
		Content: "hello",
	})
	if err != nil {
		t.Fatalf("marshalBrowserExtractPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_extract payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "content" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "extract_completed" {
		t.Fatalf("unexpected browser_extract diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "content" ||
		payload.Explanation.State != "completed" ||
		payload.Explanation.SummaryCode != "extract_completed" {
		t.Fatalf("unexpected browser_extract explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "content" ||
		payload.Diagnostics.State != "completed" ||
		payload.Diagnostics.SummaryCode != "extract_completed" {
		t.Fatalf("unexpected browser_extract diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "content" ||
		payload.Summary.State != "completed" ||
		payload.Summary.SummaryCode != "extract_completed" {
		t.Fatalf("unexpected browser_extract summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "content" ||
		payload.Display.State != "completed" ||
		payload.Display.SummaryCode != "extract_completed" {
		t.Fatalf("unexpected browser_extract display: %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "content" ||
		payload.Surface.State != "completed" ||
		payload.Surface.SummaryCode != "extract_completed" {
		t.Fatalf("unexpected browser_extract surface: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Category != "content" ||
		payload.View.State != "completed" ||
		payload.View.SummaryCode != "extract_completed" {
		t.Fatalf("unexpected browser_extract view: %#v", payload.View)
	}
}

func TestMarshalBrowserScreenshotPayloadAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserScreenshotPayload(browserScreenshotToolPayload{
		Backend: "proxy-screenshot",
		Status:  "captured",
		Path:    "shots/example.png",
	})
	if err != nil {
		t.Fatalf("marshalBrowserScreenshotPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_screenshot payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "capture" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "screenshot_completed" {
		t.Fatalf("unexpected browser_screenshot diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "capture" || payload.Summary.SummaryCode != "screenshot_completed" {
		t.Fatalf("unexpected browser_screenshot summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "capture" || payload.Display.SummaryCode != "screenshot_completed" {
		t.Fatalf("unexpected browser_screenshot display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "capture" || payload.Surface.SummaryCode != "screenshot_completed" {
		t.Fatalf("unexpected browser_screenshot surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "capture" || payload.View.SummaryCode != "screenshot_completed" {
		t.Fatalf("unexpected browser_screenshot view: %#v", payload.View)
	}
}

func TestMarshalBrowserEvalPayloadAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserEvalPayload(browserEvalToolPayload{
		Backend: "proxy-eval",
		Status:  "evaluated",
		Result:  "ok",
	})
	if err != nil {
		t.Fatalf("marshalBrowserEvalPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_eval payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "script" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected browser_eval diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "script" || payload.Summary.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected browser_eval summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "script" || payload.Display.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected browser_eval display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "script" || payload.Surface.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected browser_eval surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "script" || payload.View.SummaryCode != "evaluate_completed" {
		t.Fatalf("unexpected browser_eval view: %#v", payload.View)
	}
}

func TestMarshalBrowserTabsPayloadAppliesSuccessSummaryAliases(t *testing.T) {
	tests := []struct {
		name        string
		action      string
		status      string
		summaryCode string
	}{
		{name: "list", action: "list", status: "listed", summaryCode: "list_tabs_completed"},
		{name: "list ok", action: "list", status: "ok", summaryCode: "list_tabs_completed"},
		{name: "focus", action: "focus", status: "focused", summaryCode: "focus_tab_completed"},
		{name: "close", action: "close", status: "closed", summaryCode: "close_tab_completed"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := marshalBrowserTabsPayload(browserTabsToolPayload{
				Backend: "proxy-tabs",
				Action:  tc.action,
				Status:  tc.status,
			})
			if err != nil {
				t.Fatalf("marshalBrowserTabsPayload: %v", err)
			}
			var payload struct {
				DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
				Explanation            *browserTopLevelSummary               `json:"explanation"`
				Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
				Summary                *browserTopLevelSummary               `json:"summary"`
				Display                *browserTopLevelDisplaySummary        `json:"display"`
				Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
				View                   *browserTopLevelViewSummary           `json:"view"`
			}
			if err := json.Unmarshal([]byte(out), &payload); err != nil {
				t.Fatalf("unmarshal browser_tabs payload: %v", err)
			}
			if payload.DiagnosticsExplanation == nil ||
				payload.DiagnosticsExplanation.Category != "tabs" ||
				payload.DiagnosticsExplanation.State != "completed" ||
				payload.DiagnosticsExplanation.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_tabs diagnostics explanation: %#v", payload.DiagnosticsExplanation)
			}
			if payload.Summary == nil || payload.Summary.Category != "tabs" || payload.Summary.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_tabs summary: %#v", payload.Summary)
			}
			if payload.Display == nil || payload.Display.Category != "tabs" || payload.Display.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_tabs display: %#v", payload.Display)
			}
			if payload.Surface == nil || payload.Surface.Category != "tabs" || payload.Surface.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_tabs surface: %#v", payload.Surface)
			}
			if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "tabs" || payload.View.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_tabs view: %#v", payload.View)
			}
		})
	}
}

func TestMarshalBrowserActPayloadOpenAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:   "open",
		Status: "opened",
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_act payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "navigation" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_act open diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "navigation" ||
		payload.Explanation.State != "completed" ||
		payload.Explanation.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_act open explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "navigation" ||
		payload.Diagnostics.State != "completed" ||
		payload.Diagnostics.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_act open diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "navigation" ||
		payload.Summary.State != "completed" ||
		payload.Summary.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_act open summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "navigation" ||
		payload.Display.State != "completed" ||
		payload.Display.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_act open display: %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "navigation" ||
		payload.Surface.State != "completed" ||
		payload.Surface.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_act open surface: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Category != "navigation" ||
		payload.View.State != "completed" ||
		payload.View.SummaryCode != "open_completed" {
		t.Fatalf("unexpected browser_act open view: %#v", payload.View)
	}
}

func TestMarshalBrowserActPayloadSnapshotAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:   "snapshot",
		Status: "snapshotted",
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_act payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "content" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected browser_act snapshot diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "content" ||
		payload.Explanation.State != "completed" ||
		payload.Explanation.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected browser_act snapshot explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "content" ||
		payload.Diagnostics.State != "completed" ||
		payload.Diagnostics.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected browser_act snapshot diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "content" ||
		payload.Summary.State != "completed" ||
		payload.Summary.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected browser_act snapshot summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "content" ||
		payload.Display.State != "completed" ||
		payload.Display.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected browser_act snapshot display: %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "content" ||
		payload.Surface.State != "completed" ||
		payload.Surface.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected browser_act snapshot surface: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Category != "content" ||
		payload.View.State != "completed" ||
		payload.View.SummaryCode != "snapshot_completed" {
		t.Fatalf("unexpected browser_act snapshot view: %#v", payload.View)
	}
}

func TestMarshalBrowserActPayloadDownloadAppliesSuccessSummaryAliases(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		status      string
		summaryCode string
	}{
		{name: "download", kind: "download", status: "downloaded", summaryCode: "download_completed"},
		{name: "wait_download", kind: "wait_download", status: "downloaded", summaryCode: "wait_download_completed"},
		{name: "save_pdf", kind: "save_pdf", status: "saved", summaryCode: "save_pdf_completed"},
		{name: "save_html", kind: "save_html", status: "saved", summaryCode: "save_html_completed"},
		{name: "trace_stop", kind: "trace_stop", status: "saved", summaryCode: "trace_stop_completed"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := marshalBrowserActPayload(browserActToolPayload{
				Kind:   tc.kind,
				Status: tc.status,
			})
			if err != nil {
				t.Fatalf("marshalBrowserActPayload: %v", err)
			}
			var payload struct {
				DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
				Explanation            *browserTopLevelSummary               `json:"explanation"`
				Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
				Summary                *browserTopLevelSummary               `json:"summary"`
				Display                *browserTopLevelDisplaySummary        `json:"display"`
				Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
				View                   *browserTopLevelViewSummary           `json:"view"`
			}
			if err := json.Unmarshal([]byte(out), &payload); err != nil {
				t.Fatalf("unmarshal browser_act payload: %v", err)
			}
			if payload.DiagnosticsExplanation == nil ||
				payload.DiagnosticsExplanation.Category != "artifact" ||
				payload.DiagnosticsExplanation.State != "completed" ||
				payload.DiagnosticsExplanation.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s diagnostics explanation: %#v", tc.kind, payload.DiagnosticsExplanation)
			}
			if payload.Summary == nil || payload.Summary.Category != "artifact" || payload.Summary.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s summary: %#v", tc.kind, payload.Summary)
			}
			if payload.Display == nil || payload.Display.Category != "artifact" || payload.Display.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s display: %#v", tc.kind, payload.Display)
			}
			if payload.Surface == nil || payload.Surface.Category != "artifact" || payload.Surface.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s surface: %#v", tc.kind, payload.Surface)
			}
			if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "artifact" || payload.View.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s view: %#v", tc.kind, payload.View)
			}
		})
	}
}

func TestMarshalBrowserActPayloadTraceStartAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:   "trace_start",
		Status: "started",
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_act payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "trace" ||
		payload.DiagnosticsExplanation.State != "started" ||
		payload.DiagnosticsExplanation.SummaryCode != "trace_start_started" {
		t.Fatalf("unexpected browser_act trace_start diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "trace" ||
		payload.Summary.State != "started" ||
		payload.Summary.SummaryCode != "trace_start_started" {
		t.Fatalf("unexpected browser_act trace_start summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "trace" ||
		payload.Display.State != "started" ||
		payload.Display.SummaryCode != "trace_start_started" {
		t.Fatalf("unexpected browser_act trace_start display: %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "trace" ||
		payload.Surface.State != "started" ||
		payload.Surface.SummaryCode != "trace_start_started" {
		t.Fatalf("unexpected browser_act trace_start surface: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Category != "trace" ||
		payload.View.State != "started" ||
		payload.View.SummaryCode != "trace_start_started" {
		t.Fatalf("unexpected browser_act trace_start view: %#v", payload.View)
	}
}

func TestMarshalBrowserActPayloadDialogArmedAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:   "dialog",
		Status: "armed",
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_act payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "interaction" ||
		payload.DiagnosticsExplanation.State != "started" ||
		payload.DiagnosticsExplanation.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "interaction" ||
		payload.Explanation.State != "started" ||
		payload.Explanation.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "interaction" ||
		payload.Diagnostics.State != "started" ||
		payload.Diagnostics.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "interaction" ||
		payload.Summary.State != "started" ||
		payload.Summary.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "interaction" ||
		payload.Display.State != "started" ||
		payload.Display.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog display: %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "interaction" ||
		payload.Surface.State != "started" ||
		payload.Surface.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog surface: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Category != "interaction" ||
		payload.View.State != "started" ||
		payload.View.SummaryCode != "dialog_armed" {
		t.Fatalf("unexpected browser_act dialog view: %#v", payload.View)
	}
}

func TestMarshalBrowserActPayloadWaitAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:   "wait",
		Status: "waited",
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_act payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "timing" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "wait_completed" {
		t.Fatalf("unexpected browser_act wait diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "timing" || payload.Summary.SummaryCode != "wait_completed" {
		t.Fatalf("unexpected browser_act wait summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "timing" || payload.Display.SummaryCode != "wait_completed" {
		t.Fatalf("unexpected browser_act wait display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "timing" || payload.Surface.SummaryCode != "wait_completed" {
		t.Fatalf("unexpected browser_act wait surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "timing" || payload.View.SummaryCode != "wait_completed" {
		t.Fatalf("unexpected browser_act wait view: %#v", payload.View)
	}
}

func TestMarshalBrowserActPayloadObservabilityAppliesSuccessSummaryAliases(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		status      string
		summaryCode string
	}{
		{name: "console", kind: "console", status: "ok", summaryCode: "console_collected"},
		{name: "requests", kind: "requests", status: "ok", summaryCode: "requests_collected"},
		{name: "requests cleared", kind: "requests", status: "cleared", summaryCode: "requests_cleared"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := marshalBrowserActPayload(browserActToolPayload{
				Kind:   tc.kind,
				Status: tc.status,
			})
			if err != nil {
				t.Fatalf("marshalBrowserActPayload: %v", err)
			}
			var payload struct {
				DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
				Summary                *browserTopLevelSummary               `json:"summary"`
				Display                *browserTopLevelDisplaySummary        `json:"display"`
				Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
				View                   *browserTopLevelViewSummary           `json:"view"`
			}
			if err := json.Unmarshal([]byte(out), &payload); err != nil {
				t.Fatalf("unmarshal browser_act payload: %v", err)
			}
			if payload.DiagnosticsExplanation == nil ||
				payload.DiagnosticsExplanation.Category != "observability" ||
				payload.DiagnosticsExplanation.State != "completed" ||
				payload.DiagnosticsExplanation.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s diagnostics explanation: %#v", tc.kind, payload.DiagnosticsExplanation)
			}
			if payload.Summary == nil || payload.Summary.Category != "observability" || payload.Summary.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s summary: %#v", tc.kind, payload.Summary)
			}
			if payload.Display == nil || payload.Display.Category != "observability" || payload.Display.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s display: %#v", tc.kind, payload.Display)
			}
			if payload.Surface == nil || payload.Surface.Category != "observability" || payload.Surface.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s surface: %#v", tc.kind, payload.Surface)
			}
			if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "observability" || payload.View.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s view: %#v", tc.kind, payload.View)
			}
		})
	}
}

func TestMarshalBrowserActPayloadUploadAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:   "upload",
		Status: "uploaded",
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_act payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "form" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "upload_completed" {
		t.Fatalf("unexpected browser_act upload diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "form" || payload.Summary.SummaryCode != "upload_completed" {
		t.Fatalf("unexpected browser_act upload summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "form" || payload.Display.SummaryCode != "upload_completed" {
		t.Fatalf("unexpected browser_act upload display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "form" || payload.Surface.SummaryCode != "upload_completed" {
		t.Fatalf("unexpected browser_act upload surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "form" || payload.View.SummaryCode != "upload_completed" {
		t.Fatalf("unexpected browser_act upload view: %#v", payload.View)
	}
}

func TestMarshalBrowserActPayloadResizeAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:   "resize",
		Status: "resized",
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_act payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "viewport" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "resize_completed" {
		t.Fatalf("unexpected browser_act resize diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "viewport" || payload.Summary.SummaryCode != "resize_completed" {
		t.Fatalf("unexpected browser_act resize summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "viewport" || payload.Display.SummaryCode != "resize_completed" {
		t.Fatalf("unexpected browser_act resize display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "viewport" || payload.Surface.SummaryCode != "resize_completed" {
		t.Fatalf("unexpected browser_act resize surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "viewport" || payload.View.SummaryCode != "resize_completed" {
		t.Fatalf("unexpected browser_act resize view: %#v", payload.View)
	}
}

func TestMarshalBrowserActPayloadInteractionAppliesSuccessSummaryAliases(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		status      string
		summaryCode string
	}{
		{name: "click", kind: "click", status: "clicked", summaryCode: "click_completed"},
		{name: "press", kind: "press", status: "pressed", summaryCode: "press_completed"},
		{name: "highlight", kind: "highlight", status: "highlighted", summaryCode: "highlight_completed"},
		{name: "hover", kind: "hover", status: "hovered", summaryCode: "hover_completed"},
		{name: "drag", kind: "drag", status: "dragged", summaryCode: "drag_completed"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := marshalBrowserActPayload(browserActToolPayload{
				Kind:   tc.kind,
				Status: tc.status,
			})
			if err != nil {
				t.Fatalf("marshalBrowserActPayload: %v", err)
			}
			var payload struct {
				DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
				Summary                *browserTopLevelSummary               `json:"summary"`
				Display                *browserTopLevelDisplaySummary        `json:"display"`
				Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
				View                   *browserTopLevelViewSummary           `json:"view"`
			}
			if err := json.Unmarshal([]byte(out), &payload); err != nil {
				t.Fatalf("unmarshal browser_act payload: %v", err)
			}
			if payload.DiagnosticsExplanation == nil ||
				payload.DiagnosticsExplanation.Category != "interaction" ||
				payload.DiagnosticsExplanation.State != "completed" ||
				payload.DiagnosticsExplanation.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s diagnostics explanation: %#v", tc.kind, payload.DiagnosticsExplanation)
			}
			if payload.Summary == nil || payload.Summary.Category != "interaction" || payload.Summary.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s summary: %#v", tc.kind, payload.Summary)
			}
			if payload.Display == nil || payload.Display.Category != "interaction" || payload.Display.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s display: %#v", tc.kind, payload.Display)
			}
			if payload.Surface == nil || payload.Surface.Category != "interaction" || payload.Surface.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s surface: %#v", tc.kind, payload.Surface)
			}
			if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "interaction" || payload.View.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s view: %#v", tc.kind, payload.View)
			}
		})
	}
}

func TestMarshalBrowserActPayloadFormAppliesSuccessSummaryAliases(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		status      string
		summaryCode string
	}{
		{name: "type", kind: "type", status: "typed", summaryCode: "type_completed"},
		{name: "select", kind: "select", status: "selected", summaryCode: "select_completed"},
		{name: "fill", kind: "fill", status: "filled", summaryCode: "fill_completed"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := marshalBrowserActPayload(browserActToolPayload{
				Kind:   tc.kind,
				Status: tc.status,
			})
			if err != nil {
				t.Fatalf("marshalBrowserActPayload: %v", err)
			}
			var payload struct {
				DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
				Summary                *browserTopLevelSummary               `json:"summary"`
				Display                *browserTopLevelDisplaySummary        `json:"display"`
				Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
				View                   *browserTopLevelViewSummary           `json:"view"`
			}
			if err := json.Unmarshal([]byte(out), &payload); err != nil {
				t.Fatalf("unmarshal browser_act payload: %v", err)
			}
			if payload.DiagnosticsExplanation == nil ||
				payload.DiagnosticsExplanation.Category != "form" ||
				payload.DiagnosticsExplanation.State != "completed" ||
				payload.DiagnosticsExplanation.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s diagnostics explanation: %#v", tc.kind, payload.DiagnosticsExplanation)
			}
			if payload.Summary == nil || payload.Summary.Category != "form" || payload.Summary.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s summary: %#v", tc.kind, payload.Summary)
			}
			if payload.Display == nil || payload.Display.Category != "form" || payload.Display.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s display: %#v", tc.kind, payload.Display)
			}
			if payload.Surface == nil || payload.Surface.Category != "form" || payload.Surface.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s surface: %#v", tc.kind, payload.Surface)
			}
			if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "form" || payload.View.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s view: %#v", tc.kind, payload.View)
			}
		})
	}
}

func TestMarshalBrowserActPayloadStorageAppliesSuccessSummaryAliases(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		status      string
		category    string
		summaryCode string
	}{
		{name: "storage_set", kind: "storage_set", status: "updated", category: "storage", summaryCode: "storage_set_completed"},
		{name: "storage_clear", kind: "storage_clear", status: "cleared", category: "storage", summaryCode: "storage_clear_completed"},
		{name: "cookies_set", kind: "cookies_set", status: "updated", category: "storage", summaryCode: "cookies_set_completed"},
		{name: "cookies_clear", kind: "cookies_clear", status: "cleared", category: "storage", summaryCode: "cookies_clear_completed"},
		{name: "headers_updated", kind: "headers", status: "updated", category: "network", summaryCode: "headers_updated"},
		{name: "headers_cleared", kind: "headers", status: "cleared", category: "network", summaryCode: "headers_cleared"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := marshalBrowserActPayload(browserActToolPayload{
				Kind:   tc.kind,
				Status: tc.status,
			})
			if err != nil {
				t.Fatalf("marshalBrowserActPayload: %v", err)
			}
			var payload struct {
				DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
				Summary                *browserTopLevelSummary               `json:"summary"`
				Display                *browserTopLevelDisplaySummary        `json:"display"`
				Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
				View                   *browserTopLevelViewSummary           `json:"view"`
			}
			if err := json.Unmarshal([]byte(out), &payload); err != nil {
				t.Fatalf("unmarshal browser_act payload: %v", err)
			}
			if payload.DiagnosticsExplanation == nil ||
				payload.DiagnosticsExplanation.Category != tc.category ||
				payload.DiagnosticsExplanation.State != "completed" ||
				payload.DiagnosticsExplanation.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s diagnostics explanation: %#v", tc.kind, payload.DiagnosticsExplanation)
			}
			if payload.Summary == nil || payload.Summary.Category != tc.category || payload.Summary.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s summary: %#v", tc.kind, payload.Summary)
			}
			if payload.Display == nil || payload.Display.Category != tc.category || payload.Display.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s display: %#v", tc.kind, payload.Display)
			}
			if payload.Surface == nil || payload.Surface.Category != tc.category || payload.Surface.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s surface: %#v", tc.kind, payload.Surface)
			}
			if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != tc.category || payload.View.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s view: %#v", tc.kind, payload.View)
			}
		})
	}
}

func TestMarshalBrowserActPayloadSettingsAppliesSuccessSummaryAliases(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		status      string
		category    string
		summaryCode string
	}{
		{name: "offline_updated", kind: "offline", status: "updated", category: "network", summaryCode: "offline_updated"},
		{name: "offline_cleared", kind: "offline", status: "cleared", category: "network", summaryCode: "offline_cleared"},
		{name: "credentials_updated", kind: "credentials", status: "updated", category: "auth", summaryCode: "credentials_updated"},
		{name: "credentials_cleared", kind: "credentials", status: "cleared", category: "auth", summaryCode: "credentials_cleared"},
		{name: "geolocation_updated", kind: "geolocation", status: "updated", category: "settings", summaryCode: "geolocation_updated"},
		{name: "media_cleared", kind: "media", status: "cleared", category: "settings", summaryCode: "media_cleared"},
		{name: "timezone_updated", kind: "timezone", status: "updated", category: "settings", summaryCode: "timezone_updated"},
		{name: "locale_cleared", kind: "locale", status: "cleared", category: "settings", summaryCode: "locale_cleared"},
		{name: "device_updated", kind: "device", status: "updated", category: "settings", summaryCode: "device_updated"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := marshalBrowserActPayload(browserActToolPayload{
				Kind:   tc.kind,
				Status: tc.status,
			})
			if err != nil {
				t.Fatalf("marshalBrowserActPayload: %v", err)
			}
			var payload struct {
				DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
				Summary                *browserTopLevelSummary               `json:"summary"`
				Display                *browserTopLevelDisplaySummary        `json:"display"`
				Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
				View                   *browserTopLevelViewSummary           `json:"view"`
			}
			if err := json.Unmarshal([]byte(out), &payload); err != nil {
				t.Fatalf("unmarshal browser_act payload: %v", err)
			}
			if payload.DiagnosticsExplanation == nil ||
				payload.DiagnosticsExplanation.Category != tc.category ||
				payload.DiagnosticsExplanation.State != "completed" ||
				payload.DiagnosticsExplanation.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s diagnostics explanation: %#v", tc.kind, payload.DiagnosticsExplanation)
			}
			if payload.Summary == nil || payload.Summary.Category != tc.category || payload.Summary.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s summary: %#v", tc.kind, payload.Summary)
			}
			if payload.Display == nil || payload.Display.Category != tc.category || payload.Display.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s display: %#v", tc.kind, payload.Display)
			}
			if payload.Surface == nil || payload.Surface.Category != tc.category || payload.Surface.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s surface: %#v", tc.kind, payload.Surface)
			}
			if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != tc.category || payload.View.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s view: %#v", tc.kind, payload.View)
			}
		})
	}
}

func TestMarshalBrowserActPayloadBrowserReadAppliesSuccessSummaryAliases(t *testing.T) {
	tests := []struct {
		name        string
		kind        string
		status      string
		category    string
		summaryCode string
	}{
		{name: "response_body", kind: "response_body", status: "ok", category: "content", summaryCode: "response_body_collected"},
		{name: "errors", kind: "errors", status: "ok", category: "observability", summaryCode: "errors_collected"},
		{name: "errors_cleared", kind: "errors", status: "cleared", category: "observability", summaryCode: "errors_cleared"},
		{name: "cookies", kind: "cookies", status: "ok", category: "storage", summaryCode: "cookies_collected"},
		{name: "storage", kind: "storage", status: "ok", category: "storage", summaryCode: "storage_collected"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := marshalBrowserActPayload(browserActToolPayload{
				Kind:   tc.kind,
				Status: tc.status,
			})
			if err != nil {
				t.Fatalf("marshalBrowserActPayload: %v", err)
			}
			var payload struct {
				DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
				Summary                *browserTopLevelSummary               `json:"summary"`
				Display                *browserTopLevelDisplaySummary        `json:"display"`
				Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
				View                   *browserTopLevelViewSummary           `json:"view"`
			}
			if err := json.Unmarshal([]byte(out), &payload); err != nil {
				t.Fatalf("unmarshal browser_act payload: %v", err)
			}
			if payload.DiagnosticsExplanation == nil ||
				payload.DiagnosticsExplanation.Category != tc.category ||
				payload.DiagnosticsExplanation.State != "completed" ||
				payload.DiagnosticsExplanation.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s diagnostics explanation: %#v", tc.kind, payload.DiagnosticsExplanation)
			}
			if payload.Summary == nil || payload.Summary.Category != tc.category || payload.Summary.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s summary: %#v", tc.kind, payload.Summary)
			}
			if payload.Display == nil || payload.Display.Category != tc.category || payload.Display.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s display: %#v", tc.kind, payload.Display)
			}
			if payload.Surface == nil || payload.Surface.Category != tc.category || payload.Surface.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s surface: %#v", tc.kind, payload.Surface)
			}
			if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != tc.category || payload.View.SummaryCode != tc.summaryCode {
				t.Fatalf("unexpected browser_act %s view: %#v", tc.kind, payload.View)
			}
		})
	}
}

func TestMarshalBrowserClickPayloadAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserClickPayload(browserClickToolPayload{
		Status: "clicked",
	})
	if err != nil {
		t.Fatalf("marshalBrowserClickPayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_click payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "interaction" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "click_completed" {
		t.Fatalf("unexpected browser_click diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "interaction" || payload.Summary.SummaryCode != "click_completed" {
		t.Fatalf("unexpected browser_click summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "interaction" || payload.Display.SummaryCode != "click_completed" {
		t.Fatalf("unexpected browser_click display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "interaction" || payload.Surface.SummaryCode != "click_completed" {
		t.Fatalf("unexpected browser_click surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "interaction" || payload.View.SummaryCode != "click_completed" {
		t.Fatalf("unexpected browser_click view: %#v", payload.View)
	}
}

func TestMarshalBrowserTypePayloadAppliesSuccessSummaryAliases(t *testing.T) {
	out, err := marshalBrowserTypePayload(browserTypeToolPayload{
		Status: "typed",
	})
	if err != nil {
		t.Fatalf("marshalBrowserTypePayload: %v", err)
	}
	var payload struct {
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
		View                   *browserTopLevelViewSummary           `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_type payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "form" ||
		payload.DiagnosticsExplanation.State != "completed" ||
		payload.DiagnosticsExplanation.SummaryCode != "type_completed" {
		t.Fatalf("unexpected browser_type diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Summary == nil || payload.Summary.Category != "form" || payload.Summary.SummaryCode != "type_completed" {
		t.Fatalf("unexpected browser_type summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.Category != "form" || payload.Display.SummaryCode != "type_completed" {
		t.Fatalf("unexpected browser_type display: %#v", payload.Display)
	}
	if payload.Surface == nil || payload.Surface.Category != "form" || payload.Surface.SummaryCode != "type_completed" {
		t.Fatalf("unexpected browser_type surface: %#v", payload.Surface)
	}
	if payload.View == nil || payload.View.Kind != "result" || payload.View.Category != "form" || payload.View.SummaryCode != "type_completed" {
		t.Fatalf("unexpected browser_type view: %#v", payload.View)
	}
}

func TestMarshalBrowserActPayloadPrefersExplicitResolverFallbackSummary(t *testing.T) {
	index := 3
	outcome := &agentxbrowserruntime.BrowserElementResolverOutcome{
		Status:                        "matched",
		FallbackFromKind:              "placeholder",
		FallbackFromIndex:             1,
		FallbackFromCandidateStrength: "weak",
		FallbackFromSpecificityFields: []string{"tag"},
	}
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:                              "fill",
		Status:                            "filled",
		ResolverOutcome:                   outcome,
		ResolvedViaFallback:               true,
		ResolverFallbackKind:              "label",
		ResolverFallbackIndex:             &index,
		ResolverFallbackStrength:          "medium",
		ResolverFallbackBlockedBy:         "multiple_candidates_filtered",
		ResolverFallbackAmbiguityClass:    "filtered_residual",
		ResolverFallbackManualRetryHint:   "add_ordinal",
		ResolverFallbackSpecificityFields: []string{"tag", "type"},
		ResolverBlockedBy:                 "multiple_candidates_filtered",
		ResolverAmbiguityClass:            "filtered_residual",
		ResolverCandidateKind:             "label",
		ResolverCandidateStrength:         "medium",
		ResolverRetryDisposition:          "manual_only",
		ResolverManualRetryHint:           "add_ordinal",
		ResolverNextStepAlias:             "snapshot",
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload: %v", err)
	}
	var payload struct {
		ResolvedViaFallback               bool                                       `json:"resolved_via_fallback"`
		ResolverFallbackKind              string                                     `json:"resolver_fallback_kind"`
		ResolverFallbackIndex             *int                                       `json:"resolver_fallback_index"`
		ResolverFallbackCandidateStrength string                                     `json:"resolver_fallback_candidate_strength"`
		ResolverFallbackBlockedBy         string                                     `json:"resolver_fallback_blocked_by"`
		ResolverFallbackAmbiguityClass    string                                     `json:"resolver_fallback_ambiguity_class"`
		ResolverFallbackManualRetryHint   string                                     `json:"resolver_fallback_manual_retry_hint"`
		ResolverFallbackSpecificityFields []string                                   `json:"resolver_fallback_specificity_fields"`
		ResolverBlockedBy                 string                                     `json:"resolver_blocked_by"`
		ResolverAmbiguityClass            string                                     `json:"resolver_ambiguity_class"`
		ResolverCandidateKind             string                                     `json:"resolver_candidate_kind"`
		ResolverCandidateStrength         string                                     `json:"resolver_candidate_strength"`
		ResolverRetryDisposition          string                                     `json:"resolver_retry_disposition"`
		ResolverManualRetryHint           string                                     `json:"resolver_manual_retry_hint"`
		ResolverNextStepAlias             string                                     `json:"resolver_next_step_alias"`
		ResolverFallbackExplanation       *browserResolverFallbackExplanationSummary `json:"resolver_fallback_explanation"`
		ResolverExplanation               *browserRuntimeResolverExplanationSummary  `json:"resolver_explanation"`
		DiagnosticsExplanation            *browserDiagnosticsExplanationSummary      `json:"diagnostics_explanation"`
		Explanation                       *browserTopLevelSummary                    `json:"explanation"`
		Diagnostics                       *browserTopLevelSummary                    `json:"diagnostics"`
		Summary                           *browserTopLevelSummary                    `json:"summary"`
		Display                           *browserTopLevelDisplaySummary             `json:"display"`
		Surface                           *browserTopLevelSurfaceSummary             `json:"surface"`
		View                              *browserTopLevelViewSummary                `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_act payload: %v", err)
	}
	if !payload.ResolvedViaFallback ||
		payload.ResolverFallbackKind != "label" ||
		payload.ResolverFallbackIndex == nil ||
		*payload.ResolverFallbackIndex != 3 ||
		payload.ResolverFallbackCandidateStrength != "medium" ||
		payload.ResolverFallbackBlockedBy != "multiple_candidates_filtered" ||
		payload.ResolverFallbackAmbiguityClass != "filtered_residual" ||
		payload.ResolverFallbackManualRetryHint != "add_ordinal" ||
		!reflect.DeepEqual(payload.ResolverFallbackSpecificityFields, []string{"tag", "type"}) {
		t.Fatalf("unexpected explicit fallback summary in browser_act payload: %#v", payload)
	}
	if payload.ResolverBlockedBy != "multiple_candidates_filtered" ||
		payload.ResolverAmbiguityClass != "filtered_residual" ||
		payload.ResolverCandidateKind != "label" ||
		payload.ResolverCandidateStrength != "medium" ||
		payload.ResolverRetryDisposition != "manual_only" ||
		payload.ResolverManualRetryHint != "add_ordinal" ||
		payload.ResolverNextStepAlias != "snapshot" {
		t.Fatalf("unexpected explicit resolver guidance summary in browser_act payload: %#v", payload)
	}
	if payload.ResolverFallbackExplanation == nil ||
		payload.ResolverFallbackExplanation.State != "resolved_via_fallback" ||
		payload.ResolverFallbackExplanation.SummaryCode != "label_filtered_residual" ||
		payload.ResolverFallbackExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected explicit resolver fallback explanation in browser_act payload: %#v", payload.ResolverFallbackExplanation)
	}
	if payload.ResolverExplanation == nil ||
		payload.ResolverExplanation.State != "resolved_via_fallback" ||
		payload.ResolverExplanation.SummaryCode != "label_filtered_residual" ||
		payload.ResolverExplanation.NextStepAlias != "" ||
		payload.ResolverExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected explicit resolver explanation in browser_act payload: %#v", payload.ResolverExplanation)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "resolver_fallback" ||
		payload.DiagnosticsExplanation.State != "resolved_via_fallback" ||
		payload.DiagnosticsExplanation.SummaryCode != "label_filtered_residual" ||
		payload.DiagnosticsExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected explicit diagnostics explanation in browser_act payload: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "resolver_fallback" ||
		payload.Explanation.State != "resolved_via_fallback" ||
		payload.Explanation.SummaryCode != "label_filtered_residual" ||
		payload.Explanation.ManualRetryHint != "add_ordinal" ||
		!payload.Explanation.ResolvedViaFallback {
		t.Fatalf("unexpected explicit top-level explanation in browser_act payload: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "resolver_fallback" ||
		payload.Diagnostics.State != "resolved_via_fallback" ||
		payload.Diagnostics.SummaryCode != "label_filtered_residual" ||
		payload.Diagnostics.ManualRetryHint != "add_ordinal" ||
		!payload.Diagnostics.ResolvedViaFallback {
		t.Fatalf("unexpected explicit top-level diagnostics in browser_act payload: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "resolver_fallback" ||
		payload.Summary.State != "resolved_via_fallback" ||
		payload.Summary.SummaryCode != "label_filtered_residual" ||
		payload.Summary.NextStepAlias != "" ||
		payload.Summary.ManualRetryHint != "add_ordinal" ||
		!payload.Summary.ResolvedViaFallback {
		t.Fatalf("unexpected explicit top-level summary in browser_act payload: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Ready ||
		len(payload.Display.Sections) != 0 ||
		payload.Display.Category != "resolver_fallback" ||
		payload.Display.State != "resolved_via_fallback" ||
		payload.Display.SummaryCode != "label_filtered_residual" ||
		payload.Display.ManualRetryHint != "add_ordinal" ||
		!payload.Display.ResolvedViaFallback {
		t.Fatalf("unexpected explicit top-level display in browser_act payload: %#v", payload.Display)
	}
	if payload.Surface == nil ||
		payload.Surface.Ready ||
		len(payload.Surface.Sections) != 0 ||
		payload.Surface.Category != "resolver_fallback" ||
		payload.Surface.State != "resolved_via_fallback" ||
		payload.Surface.SummaryCode != "label_filtered_residual" ||
		payload.Surface.ManualRetryHint != "add_ordinal" ||
		!payload.Surface.ResolvedViaFallback ||
		payload.Surface.ReviewPolicyState != "" ||
		payload.Surface.ReviewDecision != "" ||
		payload.Surface.ReviewReady {
		t.Fatalf("unexpected explicit top-level surface in browser_act payload: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "result" ||
		payload.View.Ready ||
		len(payload.View.Sections) != 0 ||
		payload.View.Category != "resolver_fallback" ||
		payload.View.State != "resolved_via_fallback" ||
		payload.View.SummaryCode != "label_filtered_residual" ||
		payload.View.ManualRetryHint != "add_ordinal" ||
		!payload.View.ResolvedViaFallback ||
		payload.View.Review != nil {
		t.Fatalf("unexpected explicit top-level view in browser_act payload: %#v", payload.View)
	}
}

func TestMarshalBrowserActPayloadBuildsDiagnosticsExplanationFromResolverGuidance(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:                      "click",
		Status:                    "unresolved",
		ResolverBlockedBy:         "multiple_candidates_filtered",
		ResolverAmbiguityClass:    "filtered_residual",
		ResolverCandidateKind:     "label",
		ResolverCandidateStrength: "medium",
		ResolverRetryDisposition:  "manual_only",
		ResolverManualRetryHint:   "add_ordinal",
		ResolverNextStepAlias:     "snapshot",
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload guidance-only: %v", err)
	}
	var payload struct {
		ResolverExplanation         *browserRuntimeResolverExplanationSummary  `json:"resolver_explanation"`
		ResolverFallbackExplanation *browserResolverFallbackExplanationSummary `json:"resolver_fallback_explanation"`
		DiagnosticsExplanation      *browserDiagnosticsExplanationSummary      `json:"diagnostics_explanation"`
		Explanation                 *browserTopLevelSummary                    `json:"explanation"`
		Diagnostics                 *browserTopLevelSummary                    `json:"diagnostics"`
		Summary                     *browserTopLevelSummary                    `json:"summary"`
		Display                     *browserTopLevelDisplaySummary             `json:"display"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal guidance-only browser_act payload: %v", err)
	}
	if payload.ResolverFallbackExplanation != nil {
		t.Fatalf("expected no fallback explanation for guidance-only payload, got %#v", payload.ResolverFallbackExplanation)
	}
	if payload.ResolverExplanation == nil ||
		payload.ResolverExplanation.State != "manual_resolution_required" ||
		payload.ResolverExplanation.SummaryCode != "label_filtered_residual" ||
		payload.ResolverExplanation.NextStepAlias != "snapshot" ||
		payload.ResolverExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected resolver explanation for guidance-only payload: %#v", payload.ResolverExplanation)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "resolver" ||
		payload.DiagnosticsExplanation.State != "manual_resolution_required" ||
		payload.DiagnosticsExplanation.SummaryCode != "label_filtered_residual" ||
		payload.DiagnosticsExplanation.NextStepAlias != "snapshot" ||
		payload.DiagnosticsExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected diagnostics explanation for guidance-only payload: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "resolver" ||
		payload.Explanation.State != "manual_resolution_required" ||
		payload.Explanation.SummaryCode != "label_filtered_residual" ||
		payload.Explanation.NextStepAlias != "snapshot" ||
		payload.Explanation.ManualRetryHint != "add_ordinal" ||
		payload.Explanation.ResolvedViaFallback {
		t.Fatalf("unexpected top-level explanation for guidance-only payload: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "resolver" ||
		payload.Diagnostics.State != "manual_resolution_required" ||
		payload.Diagnostics.SummaryCode != "label_filtered_residual" ||
		payload.Diagnostics.NextStepAlias != "snapshot" ||
		payload.Diagnostics.ManualRetryHint != "add_ordinal" ||
		payload.Diagnostics.ResolvedViaFallback {
		t.Fatalf("unexpected top-level diagnostics for guidance-only payload: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "resolver" ||
		payload.Summary.State != "manual_resolution_required" ||
		payload.Summary.SummaryCode != "label_filtered_residual" ||
		payload.Summary.NextStepAlias != "snapshot" ||
		payload.Summary.ManualRetryHint != "add_ordinal" ||
		payload.Summary.ResolvedViaFallback {
		t.Fatalf("unexpected top-level summary for guidance-only payload: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Ready ||
		len(payload.Display.Sections) != 0 ||
		payload.Display.Category != "resolver" ||
		payload.Display.State != "manual_resolution_required" ||
		payload.Display.SummaryCode != "label_filtered_residual" ||
		payload.Display.NextStepAlias != "snapshot" ||
		payload.Display.ManualRetryHint != "add_ordinal" ||
		payload.Display.ResolvedViaFallback {
		t.Fatalf("unexpected top-level display for guidance-only payload: %#v", payload.Display)
	}
}

func TestMarshalBrowserActPayloadBuildsReviewDiagnosticsFromReviewDecision(t *testing.T) {
	out, err := marshalBrowserActPayload(browserActToolPayload{
		Kind:           "click",
		Status:         "review_required",
		ReviewDecision: "session_target_popup_review_required",
	})
	if err != nil {
		t.Fatalf("marshalBrowserActPayload review-required: %v", err)
	}
	var payload struct {
		ReviewDecision         string                                    `json:"review_decision"`
		Review                 *browserReviewSurfaceSummary              `json:"review"`
		ResolverExplanation    *browserRuntimeResolverExplanationSummary `json:"resolver_explanation"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary     `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary                   `json:"explanation"`
		Diagnostics            *browserTopLevelSummary                   `json:"diagnostics"`
		Summary                *browserTopLevelSummary                   `json:"summary"`
		Display                *browserTopLevelDisplaySummary            `json:"display"`
		Surface                *browserTopLevelSurfaceSummary            `json:"surface"`
		View                   *browserTopLevelViewSummary               `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal review-required browser_act payload: %v", err)
	}
	if payload.ReviewDecision != "session_target_popup_review_required" {
		t.Fatalf("unexpected review decision: %#v", payload)
	}
	if payload.ResolverExplanation != nil {
		t.Fatalf("expected review-required payload not to synthesize resolver explanation, got %#v", payload.ResolverExplanation)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "review" ||
		payload.DiagnosticsExplanation.State != "manual_confirmation_required" ||
		payload.DiagnosticsExplanation.SummaryCode != "popup_review_required" ||
		payload.DiagnosticsExplanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected review diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "review" ||
		payload.Explanation.State != "manual_confirmation_required" ||
		payload.Explanation.SummaryCode != "popup_review_required" ||
		payload.Explanation.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected review explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "review" ||
		payload.Diagnostics.State != "manual_confirmation_required" ||
		payload.Diagnostics.SummaryCode != "popup_review_required" ||
		payload.Diagnostics.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected review diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "review" ||
		payload.Summary.State != "manual_confirmation_required" ||
		payload.Summary.SummaryCode != "popup_review_required" ||
		payload.Summary.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected review summary: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Category != "review" ||
		payload.Display.State != "manual_confirmation_required" ||
		payload.Display.SummaryCode != "popup_review_required" ||
		payload.Display.ManualRetryHint != "rerun_with_force" {
		t.Fatalf("unexpected review display: %#v", payload.Display)
	}
	if payload.Review == nil ||
		payload.Review.PolicyState != "popup_review_required" ||
		payload.Review.Decision != "session_target_popup_review_required" ||
		payload.Review.Ready ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.SummaryCode != "popup_review_required" ||
		payload.Review.Display == nil ||
		payload.Review.Display.SummaryCode != "popup_review_required" {
		t.Fatalf("unexpected review surface: %#v", payload.Review)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "review" ||
		payload.Surface.State != "manual_confirmation_required" ||
		payload.Surface.SummaryCode != "popup_review_required" ||
		payload.Surface.ManualRetryHint != "rerun_with_force" ||
		payload.Surface.ReviewPolicyState != "popup_review_required" ||
		payload.Surface.ReviewDecision != "session_target_popup_review_required" ||
		payload.Surface.ReviewReady {
		t.Fatalf("unexpected top-level review surface alias: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "review" ||
		payload.View.Category != "review" ||
		payload.View.State != "manual_confirmation_required" ||
		payload.View.SummaryCode != "popup_review_required" ||
		payload.View.ManualRetryHint != "rerun_with_force" ||
		payload.View.Review == nil ||
		payload.View.Review.PolicyState != "popup_review_required" ||
		payload.View.Review.Decision != "session_target_popup_review_required" ||
		payload.View.Review.Ready {
		t.Fatalf("unexpected top-level review view alias: %#v", payload.View)
	}
}

func TestMarshalBrowserNavigatePayloadBuildsReviewSummaryFromReviewDecision(t *testing.T) {
	out, err := marshalBrowserNavigatePayload(browserNavigateToolPayload{
		URL:            "https://93.184.216.34/start",
		FinalURL:       "https://93.184.216.34/landing",
		Backend:        "proxy",
		Status:         "review_required",
		ReviewDecision: "navigate_redirect_review_required",
	})
	if err != nil {
		t.Fatalf("marshalBrowserNavigatePayload review-required: %v", err)
	}
	var payload struct {
		Review                 *browserReviewSurfaceSummary          `json:"review"`
		DiagnosticsExplanation *browserDiagnosticsExplanationSummary `json:"diagnostics_explanation"`
		Explanation            *browserTopLevelSummary               `json:"explanation"`
		Diagnostics            *browserTopLevelSummary               `json:"diagnostics"`
		Summary                *browserTopLevelSummary               `json:"summary"`
		Display                *browserTopLevelDisplaySummary        `json:"display"`
		Surface                *browserTopLevelSurfaceSummary        `json:"surface"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal review-required browser_navigate payload: %v", err)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "review" ||
		payload.DiagnosticsExplanation.SummaryCode != "redirect_review_required" {
		t.Fatalf("unexpected navigate review diagnostics explanation: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil || payload.Explanation.SummaryCode != "redirect_review_required" {
		t.Fatalf("unexpected navigate review explanation: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil || payload.Diagnostics.SummaryCode != "redirect_review_required" {
		t.Fatalf("unexpected navigate review diagnostics: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "redirect_review_required" {
		t.Fatalf("unexpected navigate review summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.SummaryCode != "redirect_review_required" {
		t.Fatalf("unexpected navigate review display: %#v", payload.Display)
	}
	if payload.Review == nil ||
		payload.Review.PolicyState != "redirect_review_required" ||
		payload.Review.Decision != "navigate_redirect_review_required" ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.SummaryCode != "redirect_review_required" {
		t.Fatalf("unexpected navigate review surface: %#v", payload.Review)
	}
	if payload.Surface == nil ||
		payload.Surface.Category != "review" ||
		payload.Surface.State != "manual_confirmation_required" ||
		payload.Surface.SummaryCode != "redirect_review_required" ||
		payload.Surface.ManualRetryHint != "rerun_with_force" ||
		payload.Surface.ReviewPolicyState != "redirect_review_required" ||
		payload.Surface.ReviewDecision != "navigate_redirect_review_required" ||
		payload.Surface.ReviewReady {
		t.Fatalf("unexpected navigate review surface alias: %#v", payload.Surface)
	}
}

func TestMarshalBrowserTabsPayloadCarriesSuccessSummaryIntoConfirmedReviewSurface(t *testing.T) {
	out, err := marshalBrowserTabsPayload(browserTabsToolPayload{
		Backend:          "proxy-tabs",
		Action:           "list",
		Status:           "ok",
		RememberDecision: "session_target_popup_review_confirmed",
		RememberReady:    true,
	})
	if err != nil {
		t.Fatalf("marshalBrowserTabsPayload confirmed review success: %v", err)
	}
	var payload struct {
		Review  *browserReviewSurfaceSummary   `json:"review"`
		Summary *browserTopLevelSummary        `json:"summary"`
		Display *browserTopLevelDisplaySummary `json:"display"`
		Surface *browserTopLevelSurfaceSummary `json:"surface"`
		View    *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal confirmed-review browser_tabs payload: %v", err)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "list_tabs_completed" {
		t.Fatalf("unexpected top-level tabs summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.SummaryCode != "list_tabs_completed" {
		t.Fatalf("unexpected top-level tabs display: %#v", payload.Display)
	}
	if payload.Review == nil ||
		payload.Review.PolicyState != "session_target_popup_review_confirmed" ||
		payload.Review.Decision != "session_target_popup_review_confirmed" ||
		!payload.Review.Ready ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.SummaryCode != "list_tabs_completed" ||
		payload.Review.Display == nil ||
		payload.Review.Display.SummaryCode != "list_tabs_completed" {
		t.Fatalf("unexpected confirmed review surface: %#v", payload.Review)
	}
	if payload.Surface == nil ||
		payload.Surface.SummaryCode != "list_tabs_completed" ||
		payload.Surface.ReviewPolicyState != "session_target_popup_review_confirmed" ||
		payload.Surface.ReviewDecision != "session_target_popup_review_confirmed" ||
		!payload.Surface.ReviewReady {
		t.Fatalf("unexpected confirmed review surface alias: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "review" ||
		payload.View.SummaryCode != "list_tabs_completed" ||
		payload.View.Review == nil ||
		payload.View.Review.PolicyState != "session_target_popup_review_confirmed" ||
		payload.View.Review.Summary == nil ||
		payload.View.Review.Summary.SummaryCode != "list_tabs_completed" {
		t.Fatalf("unexpected confirmed review view alias: %#v", payload.View)
	}
}

func TestMarshalBrowserTabsPayloadCarriesCloseSuccessSummaryIntoConfirmedReviewSurface(t *testing.T) {
	out, err := marshalBrowserTabsPayload(browserTabsToolPayload{
		Backend:          "proxy-tabs",
		Action:           "close",
		Status:           "closed",
		RememberDecision: "session_target_popup_review_confirmed",
		RememberReady:    true,
	})
	if err != nil {
		t.Fatalf("marshalBrowserTabsPayload confirmed close review success: %v", err)
	}
	var payload struct {
		Review  *browserReviewSurfaceSummary   `json:"review"`
		Summary *browserTopLevelSummary        `json:"summary"`
		Display *browserTopLevelDisplaySummary `json:"display"`
		Surface *browserTopLevelSurfaceSummary `json:"surface"`
		View    *browserTopLevelViewSummary    `json:"view"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal confirmed close-review browser_tabs payload: %v", err)
	}
	if payload.Summary == nil || payload.Summary.SummaryCode != "close_tab_completed" {
		t.Fatalf("unexpected top-level close summary: %#v", payload.Summary)
	}
	if payload.Display == nil || payload.Display.SummaryCode != "close_tab_completed" {
		t.Fatalf("unexpected top-level close display: %#v", payload.Display)
	}
	if payload.Review == nil ||
		payload.Review.PolicyState != "session_target_popup_review_confirmed" ||
		payload.Review.Decision != "session_target_popup_review_confirmed" ||
		!payload.Review.Ready ||
		payload.Review.Summary == nil ||
		payload.Review.Summary.SummaryCode != "close_tab_completed" ||
		payload.Review.Display == nil ||
		payload.Review.Display.SummaryCode != "close_tab_completed" {
		t.Fatalf("unexpected confirmed close review surface: %#v", payload.Review)
	}
	if payload.Surface == nil ||
		payload.Surface.SummaryCode != "close_tab_completed" ||
		payload.Surface.ReviewPolicyState != "session_target_popup_review_confirmed" ||
		payload.Surface.ReviewDecision != "session_target_popup_review_confirmed" ||
		!payload.Surface.ReviewReady {
		t.Fatalf("unexpected confirmed close review surface alias: %#v", payload.Surface)
	}
	if payload.View == nil ||
		payload.View.Kind != "review" ||
		payload.View.SummaryCode != "close_tab_completed" ||
		payload.View.Review == nil ||
		payload.View.Review.PolicyState != "session_target_popup_review_confirmed" ||
		payload.View.Review.Summary == nil ||
		payload.View.Review.Summary.SummaryCode != "close_tab_completed" {
		t.Fatalf("unexpected confirmed close review view alias: %#v", payload.View)
	}
}

func TestMarshalBrowserScreenshotPayloadBuildsTopLevelSummaryFromFallback(t *testing.T) {
	index := 2
	out, err := marshalBrowserScreenshotPayload(browserScreenshotToolPayload{
		Path:                            ".agentx/browser/example.png",
		Status:                          "captured",
		ResolvedViaFallback:             true,
		ResolverFallbackKind:            "label",
		ResolverFallbackIndex:           &index,
		ResolverFallbackStrength:        "medium",
		ResolverFallbackBlockedBy:       "multiple_candidates_filtered",
		ResolverFallbackAmbiguityClass:  "filtered_residual",
		ResolverFallbackManualRetryHint: "add_ordinal",
	})
	if err != nil {
		t.Fatalf("marshalBrowserScreenshotPayload: %v", err)
	}
	var payload struct {
		ResolverFallbackExplanation *browserResolverFallbackExplanationSummary `json:"resolver_fallback_explanation"`
		DiagnosticsExplanation      *browserDiagnosticsExplanationSummary      `json:"diagnostics_explanation"`
		Explanation                 *browserTopLevelSummary                    `json:"explanation"`
		Diagnostics                 *browserTopLevelSummary                    `json:"diagnostics"`
		Summary                     *browserTopLevelSummary                    `json:"summary"`
		Display                     *browserTopLevelDisplaySummary             `json:"display"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("unmarshal browser_screenshot payload: %v", err)
	}
	if payload.ResolverFallbackExplanation == nil ||
		payload.ResolverFallbackExplanation.State != "resolved_via_fallback" ||
		payload.ResolverFallbackExplanation.SummaryCode != "label_filtered_residual" ||
		payload.ResolverFallbackExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected fallback explanation in browser_screenshot payload: %#v", payload.ResolverFallbackExplanation)
	}
	if payload.DiagnosticsExplanation == nil ||
		payload.DiagnosticsExplanation.Category != "resolver_fallback" ||
		payload.DiagnosticsExplanation.State != "resolved_via_fallback" ||
		payload.DiagnosticsExplanation.SummaryCode != "label_filtered_residual" ||
		payload.DiagnosticsExplanation.ManualRetryHint != "add_ordinal" {
		t.Fatalf("unexpected diagnostics explanation in browser_screenshot payload: %#v", payload.DiagnosticsExplanation)
	}
	if payload.Explanation == nil ||
		payload.Explanation.Category != "resolver_fallback" ||
		payload.Explanation.State != "resolved_via_fallback" ||
		payload.Explanation.SummaryCode != "label_filtered_residual" ||
		payload.Explanation.ManualRetryHint != "add_ordinal" ||
		!payload.Explanation.ResolvedViaFallback {
		t.Fatalf("unexpected top-level explanation in browser_screenshot payload: %#v", payload.Explanation)
	}
	if payload.Diagnostics == nil ||
		payload.Diagnostics.Category != "resolver_fallback" ||
		payload.Diagnostics.State != "resolved_via_fallback" ||
		payload.Diagnostics.SummaryCode != "label_filtered_residual" ||
		payload.Diagnostics.ManualRetryHint != "add_ordinal" ||
		!payload.Diagnostics.ResolvedViaFallback {
		t.Fatalf("unexpected top-level diagnostics in browser_screenshot payload: %#v", payload.Diagnostics)
	}
	if payload.Summary == nil ||
		payload.Summary.Category != "resolver_fallback" ||
		payload.Summary.State != "resolved_via_fallback" ||
		payload.Summary.SummaryCode != "label_filtered_residual" ||
		payload.Summary.NextStepAlias != "" ||
		payload.Summary.ManualRetryHint != "add_ordinal" ||
		!payload.Summary.ResolvedViaFallback {
		t.Fatalf("unexpected top-level summary in browser_screenshot payload: %#v", payload.Summary)
	}
	if payload.Display == nil ||
		payload.Display.Ready ||
		len(payload.Display.Sections) != 0 ||
		payload.Display.Category != "resolver_fallback" ||
		payload.Display.State != "resolved_via_fallback" ||
		payload.Display.SummaryCode != "label_filtered_residual" ||
		payload.Display.ManualRetryHint != "add_ordinal" ||
		!payload.Display.ResolvedViaFallback {
		t.Fatalf("unexpected top-level display in browser_screenshot payload: %#v", payload.Display)
	}
}
