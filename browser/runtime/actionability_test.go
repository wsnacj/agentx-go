package browserruntime

import "testing"

func TestBuildBrowserActionabilityReportFailedResolver(t *testing.T) {
	report := BuildBrowserActionabilityReport(BrowserActionabilityReportRequest{
		Action:     "click",
		ElementRef: "css1:#submit",
		ResolverOutcome: &BrowserElementResolverOutcome{
			Status:           "unresolved",
			BlockedBy:        "multiple_candidates",
			RetryDisposition: "manual_only",
			ManualRetryHint:  "add_specificity",
			RecoveryAction:   "browser action=snapshot",
			Note:             "multiple matching buttons",
		},
	})
	if report == nil {
		t.Fatal("expected actionability report")
	}
	if report.Status != BrowserActionabilityStatusFailed ||
		report.FailedCheck != "resolve_target" ||
		report.FailureReason != "resolver_unresolved_multiple_candidates" ||
		report.ManualRetryHint != "add_specificity" ||
		report.RecoveryAction != "browser action=snapshot" {
		t.Fatalf("unexpected failed report: %#v", report)
	}
	if len(report.Checks) < 2 || report.Checks[0].Name != "resolve_target" || report.Checks[0].Status != BrowserActionabilityStatusFailed {
		t.Fatalf("expected failed resolve_target check, got %#v", report.Checks)
	}
	for _, check := range report.Checks[1:] {
		if check.Required && check.Status != BrowserActionabilityStatusSkipped {
			t.Fatalf("expected checks after failed resolution to be skipped, got %#v", report.Checks)
		}
	}
}

func TestBuildBrowserActionFailureEvidenceIncludesSnapshotSummary(t *testing.T) {
	outcome := &BrowserElementResolverOutcome{
		Status:         "page_binding_blocked",
		BlockedBy:      "page_url",
		RecoveryAction: "browser action=snapshot",
		Note:           "snapshot ref belongs to a different page",
	}
	actionability := BuildBrowserActionabilityReport(BrowserActionabilityReportRequest{
		Action:          "type",
		Selector:        "input[name=email]",
		ResolverOutcome: outcome,
	})
	evidence := BuildBrowserActionFailureEvidence(BrowserActionFailureEvidenceRequest{
		Action:            "type",
		Status:            "unresolved",
		FinalURL:          "https://example.com/form",
		Title:             "Example Form",
		Selector:          "input[name=email]",
		ResolverOutcome:   outcome,
		Actionability:     actionability,
		SnapshotText:      "textbox Email",
		SnapshotFormat:    "role",
		SnapshotRefs:      "role",
		ElementCount:      3,
		SnapshotTruncated: true,
		ConsoleMessages: []BrowserConsoleMessage{
			{Level: "warn", Text: "slow input handler", Source: "console"},
		},
		Errors: []BrowserErrorEntry{
			{Event: "page_error", Category: "script", Severity: "error", Message: "client error"},
		},
	})
	if evidence == nil {
		t.Fatal("expected failure evidence")
	}
	if evidence.ReasonCode != "resolver_page_binding_blocked_page_url" ||
		evidence.RecoveryAction != "browser action=snapshot" ||
		!evidence.Retryable {
		t.Fatalf("unexpected evidence: %#v", evidence)
	}
	if !evidence.SnapshotAvailable || evidence.SnapshotElementCount != 3 || !evidence.SnapshotTruncated {
		t.Fatalf("expected snapshot summary in evidence, got %#v", evidence)
	}
	if evidence.SnapshotFreshness == nil ||
		evidence.SnapshotFreshness.State != BrowserSnapshotFreshnessStateFresh ||
		evidence.SnapshotFreshness.Source != BrowserSnapshotFreshnessSourceRecoverySnapshot ||
		evidence.SnapshotFreshness.RefreshRecommended ||
		evidence.SnapshotFreshness.RefreshReason != BrowserSnapshotRefreshReasonPageChanged ||
		evidence.SnapshotFreshness.RecoveryAction != BrowserSnapshotRecoveryUseRecoverySnapshot {
		t.Fatalf("expected recovery snapshot freshness in evidence, got %#v", evidence.SnapshotFreshness)
	}
	if evidence.Actionability == nil || evidence.Actionability.Status != BrowserActionabilityStatusFailed {
		t.Fatalf("expected failed actionability in evidence, got %#v", evidence.Actionability)
	}
	if evidence.Artifact == nil ||
		evidence.Artifact.Kind != "trace_like" ||
		evidence.Artifact.Action != "type" ||
		evidence.Artifact.ReasonCode != "resolver_page_binding_blocked_page_url" ||
		evidence.Artifact.TargetKind != "selector" ||
		evidence.Artifact.Target != "input[name=email]" ||
		evidence.Artifact.FailedCheck != "resolve_target" ||
		evidence.Artifact.FinalURL != "https://example.com/form" ||
		evidence.Artifact.Title != "Example Form" ||
		!evidence.Artifact.SnapshotAvailable ||
		evidence.Artifact.SnapshotRefs != "role" ||
		evidence.Artifact.SnapshotElementCount != 3 ||
		evidence.Artifact.SnapshotFreshness == nil ||
		evidence.Artifact.SnapshotFreshness.Source != BrowserSnapshotFreshnessSourceRecoverySnapshot ||
		evidence.Artifact.ResolverOutcome == nil ||
		evidence.Artifact.ConsoleMessageCount != 1 ||
		evidence.Artifact.ErrorCount != 1 ||
		len(evidence.Artifact.ConsoleMessages) != 1 ||
		len(evidence.Artifact.Errors) != 1 {
		t.Fatalf("expected trace-like failure artifact, got %#v", evidence.Artifact)
	}
}

func TestBuildBrowserActionFailureEvidenceSkipsSuccessfulMatchedAction(t *testing.T) {
	outcome := &BrowserElementResolverOutcome{Status: "matched", MatchedKind: "native_ref"}
	actionability := BuildBrowserActionabilityReport(BrowserActionabilityReportRequest{
		Action:          "click",
		ElementRef:      "css1:#submit",
		ResolverOutcome: outcome,
	})
	if actionability == nil || actionability.Status != BrowserActionabilityStatusPartial {
		t.Fatalf("expected partial actionability when backend checks are not reported, got %#v", actionability)
	}
	evidence := BuildBrowserActionFailureEvidence(BrowserActionFailureEvidenceRequest{
		Action:          "click",
		Status:          "clicked",
		ElementRef:      "css1:#submit",
		ResolverOutcome: outcome,
		Actionability:   actionability,
	})
	if evidence != nil {
		t.Fatalf("expected no failure evidence for matched clicked action, got %#v", evidence)
	}
}

func TestBuildBrowserActionFailureEvidenceArtifactCarriesScreenshotPath(t *testing.T) {
	actionability := &BrowserActionabilityReport{
		Action:        "screenshot",
		Status:        BrowserActionabilityStatusFailed,
		TargetKind:    "selector",
		Target:        "#chart",
		FailedCheck:   "visible",
		FailureReason: "actionability_visible_failed",
	}
	evidence := BuildBrowserActionFailureEvidence(BrowserActionFailureEvidenceRequest{
		Action:        "screenshot",
		Status:        "action_failed",
		Selector:      "#chart",
		Actionability: actionability,
		ArtifactPath:  "/tmp/chart.png",
	})
	if evidence == nil || evidence.Artifact == nil {
		t.Fatalf("expected screenshot failure artifact, got %#v", evidence)
	}
	if evidence.Artifact.ArtifactPath != "/tmp/chart.png" || evidence.Artifact.ScreenshotPath != "/tmp/chart.png" {
		t.Fatalf("expected screenshot path in failure artifact, got %#v", evidence.Artifact)
	}

	actionability.Action = "click"
	evidence = BuildBrowserActionFailureEvidence(BrowserActionFailureEvidenceRequest{
		Action:        "click",
		Status:        "action_failed",
		Selector:      "#chart",
		Actionability: actionability,
		ArtifactPath:  "/tmp/click.json",
	})
	if evidence == nil || evidence.Artifact == nil {
		t.Fatalf("expected click failure artifact, got %#v", evidence)
	}
	if evidence.Artifact.ArtifactPath != "/tmp/click.json" || evidence.Artifact.ScreenshotPath != "" {
		t.Fatalf("expected non-screenshot artifact path without screenshot path, got %#v", evidence.Artifact)
	}
}
