package browserops

import (
	"strings"
	"testing"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

func TestEvaluateBrowserActionFailurePayloadEvidencePassesRequiredChecks(t *testing.T) {
	required := []string{"visible", "enabled", "editable", "receives_events", "attached"}
	payloads := make([]BrowserActionFailurePayloadEvidence, 0, len(required))
	for _, failedCheck := range required {
		action := "click"
		if failedCheck == "editable" {
			action = "type"
		}
		payloads = append(payloads, browserActionFailurePayloadFixture(action, failedCheck))
	}

	result := EvaluateBrowserActionFailurePayloadEvidence(BrowserActionFailurePayloadEvaluationInput{
		Payloads:                payloads,
		RequiredFailedChecks:    required,
		RequireTraceArtifact:    true,
		RequireSnapshotEvidence: true,
		MinDistinctFailedChecks: len(required),
	})

	if !result.Passed ||
		result.ValidPayloadCount != len(required) ||
		result.DistinctFailedCheckCount != len(required) ||
		len(result.MissingFailedChecks) != 0 ||
		!result.TraceArtifactReady ||
		!result.SnapshotEvidenceReady {
		t.Fatalf("expected failed browser action payload gate to pass, got %#v", result)
	}
	for _, failedCheck := range required {
		if !containsBrowserActionFailureCheck(result.DistinctFailedChecks, failedCheck) {
			t.Fatalf("expected failed_check %q in %#v", failedCheck, result.DistinctFailedChecks)
		}
	}
}

func TestEvaluateBrowserActionFailurePayloadEvidenceFailsMissingArtifactAndSnapshot(t *testing.T) {
	payload := browserActionFailurePayloadFixture("click", "visible")
	payload.FailureEvidence.Artifact = nil
	payload.FailureEvidence.SnapshotAvailable = false
	payload.FailureEvidence.SnapshotElementCount = 0
	payload.FailureEvidence.SnapshotRefs = ""
	payload.FailureEvidence.SnapshotFormat = ""

	result := EvaluateBrowserActionFailurePayloadEvidence(BrowserActionFailurePayloadEvaluationInput{
		Payloads:                []BrowserActionFailurePayloadEvidence{payload},
		RequiredFailedChecks:    []string{"visible"},
		RequireTraceArtifact:    true,
		RequireSnapshotEvidence: true,
	})

	if result.Passed || result.ValidPayloadCount != 0 {
		t.Fatalf("expected missing artifact/snapshot to fail, got %#v", result)
	}
	for _, reason := range []string{"payload_0:trace_artifact_missing", "payload_0:snapshot_evidence_missing"} {
		if !containsBrowserFailureReason(result.FailureReasons, reason) {
			t.Fatalf("expected reason %q in %#v", reason, result.FailureReasons)
		}
	}
}

func TestEvaluateBrowserActionFailurePayloadEvidenceFailsReasonMismatch(t *testing.T) {
	payload := browserActionFailurePayloadFixture("click", "receives_events")
	payload.FailureEvidence.ReasonCode = "action_failed_without_actionability"
	if payload.FailureEvidence.Artifact != nil {
		payload.FailureEvidence.Artifact.ReasonCode = "action_failed_without_actionability"
	}

	result := EvaluateBrowserActionFailurePayloadEvidence(BrowserActionFailurePayloadEvaluationInput{
		Payloads:                []BrowserActionFailurePayloadEvidence{payload},
		RequiredFailedChecks:    []string{"receives_events"},
		RequireTraceArtifact:    true,
		RequireSnapshotEvidence: true,
	})

	if result.Passed {
		t.Fatalf("expected reason mismatch to fail, got %#v", result)
	}
	if !containsBrowserFailureReason(result.FailureReasons, "payload_0:reason_code_mismatch") {
		t.Fatalf("expected reason_code_mismatch in %#v", result.FailureReasons)
	}
	if !strings.Contains(result.Summary, "validated 0/1") {
		t.Fatalf("expected failure summary to include payload count, got %q", result.Summary)
	}
}

func browserActionFailurePayloadFixture(action string, failedCheck string) BrowserActionFailurePayloadEvidence {
	reasonCode := "actionability_" + failedCheck + "_failed"
	report := &agentxbrowserruntime.BrowserActionabilityReport{
		Action:           action,
		Status:           agentxbrowserruntime.BrowserActionabilityStatusFailed,
		TargetKind:       "selector",
		Target:           "#target",
		FailedCheck:      failedCheck,
		FailureReason:    reasonCode,
		RetryDisposition: "retry_after_state_refresh",
		ManualRetryHint:  "browser_actionability_fixture",
		RecoveryAction:   "browser action=snapshot",
		Checks: []agentxbrowserruntime.BrowserActionabilityCheck{
			{Name: "resolve_target", Status: agentxbrowserruntime.BrowserActionabilityStatusPassed, Required: true},
			{Name: failedCheck, Status: agentxbrowserruntime.BrowserActionabilityStatusFailed, Required: true},
		},
	}
	evidence := agentxbrowserruntime.BuildBrowserActionFailureEvidence(agentxbrowserruntime.BrowserActionFailureEvidenceRequest{
		Action:            action,
		Status:            "action_failed",
		Note:              reasonCode,
		RecoveryAction:    "browser action=snapshot",
		Selector:          "#target",
		Actionability:     report,
		SnapshotFormat:    "aria",
		SnapshotRefs:      "role",
		ElementCount:      1,
		SnapshotText:      "button Target",
		ArtifactPath:      "data/tmp/browser-action-failure.png",
		FinalURL:          "https://example.com/form",
		Title:             "Browser actionability fixture",
		ConsoleMessages:   []agentxbrowserruntime.BrowserConsoleMessage{{Text: "covered by overlay"}},
		Errors:            []agentxbrowserruntime.BrowserErrorEntry{{Message: reasonCode}},
		SnapshotTruncated: false,
	})
	return BrowserActionFailurePayloadEvidence{
		Action:          action,
		Status:          "action_failed",
		Actionability:   report,
		FailureEvidence: evidence,
	}
}

func containsBrowserFailureReason(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
