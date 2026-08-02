package browserruntime

import "testing"

func TestBuildBrowserStepRecoveryTranscriptFromFailureEvidence(t *testing.T) {
	actionability := &BrowserActionabilityReport{
		Action:           "click",
		Status:           BrowserActionabilityStatusFailed,
		TargetKind:       "ref",
		Target:           "elem1:submit",
		FailedCheck:      "resolve_target",
		FailureReason:    "resolver_unresolved_multiple_candidates",
		RecoveryAction:   "browser action=snapshot",
		ManualRetryHint:  "add_ordinal",
		RetryDisposition: "manual_only",
	}
	outcome := &BrowserElementResolverOutcome{
		Status:          "unresolved",
		PrimaryKind:     "native_ref",
		BlockedBy:       "multiple_candidates_filtered",
		RecoveryAction:  "browser action=snapshot",
		NextStepAlias:   "snapshot",
		ManualRetryHint: "add_ordinal",
	}
	freshness := &BrowserSnapshotFreshness{
		State:              BrowserSnapshotFreshnessStateStale,
		Source:             BrowserSnapshotFreshnessSourceResolverOutcome,
		RefreshRecommended: true,
		RefreshReason:      BrowserSnapshotRefreshReasonAmbiguousTarget,
		RecoveryAction:     "browser action=snapshot",
		NextStepAlias:      "snapshot",
	}

	transcript := BuildBrowserStepRecoveryTranscript(BrowserStepRecoveryTranscriptRequest{
		Action:            "click",
		Status:            "action_failed",
		ReasonCode:        "resolver_unresolved_multiple_candidates",
		Message:           "multiple matching buttons",
		RecoveryAction:    "browser action=snapshot",
		Retryable:         true,
		TargetKind:        "ref",
		Target:            "elem1:submit",
		ResolverOutcome:   outcome,
		Actionability:     actionability,
		SnapshotFreshness: freshness,
	})
	if transcript == nil {
		t.Fatal("expected recovery transcript")
	}
	if transcript.Kind != BrowserStepRecoveryTranscriptKind ||
		transcript.Action != "click" ||
		transcript.Status != "action_failed" ||
		transcript.ReasonCode != "resolver_unresolved_multiple_candidates" ||
		transcript.TargetKind != "ref" ||
		transcript.Target != "elem1:submit" ||
		transcript.FailedCheck != "resolve_target" ||
		transcript.SnapshotState != BrowserSnapshotFreshnessStateStale ||
		transcript.RecoveryAction != "browser action=snapshot" ||
		transcript.NextStepAlias != "snapshot" ||
		transcript.ManualRetryHint != "add_ordinal" ||
		!transcript.Retryable {
		t.Fatalf("unexpected transcript: %#v", transcript)
	}
	if len(transcript.Steps) != 5 {
		t.Fatalf("expected 5 compact transcript steps, got %#v", transcript.Steps)
	}
	expectedPhases := []string{
		BrowserStepRecoveryTranscriptPhaseFailedAction,
		BrowserStepRecoveryTranscriptPhaseActionability,
		BrowserStepRecoveryTranscriptPhaseTargetResolution,
		BrowserStepRecoveryTranscriptPhaseSnapshotFreshness,
		BrowserStepRecoveryTranscriptPhaseRecommendedNextStep,
	}
	for i, expected := range expectedPhases {
		step := transcript.Steps[i]
		if step.Index != i+1 || step.Phase != expected {
			t.Fatalf("unexpected step %d: %#v", i, step)
		}
	}
	if transcript.Steps[2].EvidenceSource != "resolver_outcome" ||
		transcript.Steps[2].Reason != "multiple_candidates_filtered" ||
		transcript.Steps[4].EvidenceSource != "recovery_guidance" {
		t.Fatalf("unexpected transcript evidence sources: %#v", transcript.Steps)
	}
}

func TestBuildBrowserActionFailureEvidenceCarriesRecoveryTranscript(t *testing.T) {
	outcome := &BrowserElementResolverOutcome{
		Status:         "page_binding_blocked",
		BlockedBy:      "page_url",
		RecoveryAction: "browser action=snapshot",
		Note:           "snapshot ref belongs to a different page",
	}
	evidence := BuildBrowserActionFailureEvidence(BrowserActionFailureEvidenceRequest{
		Action:          "type",
		Status:          "unresolved",
		Selector:        "input[name=email]",
		ResolverOutcome: outcome,
		SnapshotText:    "textbox Email",
		SnapshotFormat:  "role",
		SnapshotRefs:    "role",
		ElementCount:    3,
	})
	if evidence == nil || evidence.RecoveryTranscript == nil {
		t.Fatalf("expected failure evidence recovery transcript, got %#v", evidence)
	}
	if evidence.RecoveryTranscript.RecoveryAction != BrowserSnapshotRecoveryUseRecoverySnapshot ||
		evidence.RecoveryTranscript.SnapshotState != BrowserSnapshotFreshnessStateFresh ||
		evidence.RecoveryTranscript.NextStepAlias != "snapshot" {
		t.Fatalf("unexpected recovery transcript: %#v", evidence.RecoveryTranscript)
	}
	if evidence.Artifact == nil || evidence.Artifact.RecoveryTranscript == nil {
		t.Fatalf("expected artifact recovery transcript, got %#v", evidence.Artifact)
	}
	evidence.RecoveryTranscript.Steps[0].Phase = "mutated"
	if evidence.Artifact.RecoveryTranscript.Steps[0].Phase == "mutated" {
		t.Fatalf("expected artifact transcript clone, got %#v", evidence.Artifact.RecoveryTranscript)
	}
}
