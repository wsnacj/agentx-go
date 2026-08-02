package publicnews

import (
	"strings"
	"testing"
)

func TestEvaluateLatestNewsBriefEvidencePasses(t *testing.T) {
	eval := EvaluateLatestNewsBriefEvidence(LatestNewsBriefEvaluationInput{
		SourceURL:              "https://example.com/news/latest",
		PublishedAt:            "2026-04-24 10:00",
		NewsFieldsReady:        true,
		GuardStatus:            "passed",
		CrossCheckReady:        true,
		SourceEvidenceAccepted: true,
		StopAfterGuardPassed:   true,
	})
	if !eval.Passed {
		t.Fatalf("expected evaluation to pass: %#v", eval)
	}
	if !eval.FreshnessConfirmed || !eval.SourceAccepted || !eval.CrossCheckReady || !eval.StopAfterGuardPassed {
		t.Fatalf("expected all guard dimensions to be true: %#v", eval)
	}
}

func TestEvaluateLatestNewsBriefEvidenceRejectsPublishedBeforeRequestedWindow(t *testing.T) {
	eval := EvaluateLatestNewsBriefEvidence(LatestNewsBriefEvaluationInput{
		SourceURL:              "https://example.com/news/old",
		PublishedAt:            "2026-07-01T13:25:00+08:00",
		PublishedAfter:         "2026-07-07T00:00:00+08:00",
		NewsFieldsReady:        true,
		GuardStatus:            "passed",
		CrossCheckReady:        true,
		SourceEvidenceAccepted: true,
		StopAfterGuardPassed:   true,
	})
	if eval.FreshnessConfirmed || eval.Passed {
		t.Fatalf("expected stale primary evidence to fail freshness, got %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "published_at_outside_freshness_window") {
		t.Fatalf("expected freshness-window failure, got %q", eval.FailureReason)
	}
}

func TestEvaluateLatestNewsBriefEvidenceFailsOnSingleSourceAndMissingTime(t *testing.T) {
	eval := EvaluateLatestNewsBriefEvidence(LatestNewsBriefEvaluationInput{
		SourceURL:              "https://example.com/news/latest",
		PublishedAt:            "",
		NewsFieldsReady:        true,
		GuardStatus:            "passed",
		CrossCheckReady:        false,
		ObservedSourceCount:    1,
		SourceEvidenceAccepted: true,
		StopAfterGuardPassed:   true,
	})
	if eval.Passed {
		t.Fatalf("expected evaluation to fail: %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "published_at_missing") {
		t.Fatalf("expected published_at_missing failure, got %q", eval.FailureReason)
	}
	if !strings.Contains(eval.FailureReason, "cross_check_not_ready") {
		t.Fatalf("expected cross_check_not_ready failure, got %q", eval.FailureReason)
	}
}

func TestEvaluateLatestNewsBriefEvidenceRequiresExplicitCrossCheckReady(t *testing.T) {
	eval := EvaluateLatestNewsBriefEvidence(LatestNewsBriefEvaluationInput{
		SourceURL:              "https://example.com/news/latest",
		PublishedAt:            "2026-04-24",
		NewsFieldsReady:        true,
		GuardStatus:            "passed",
		ObservedSourceCount:    2,
		SourceEvidenceAccepted: true,
		StopAfterGuardPassed:   true,
	})
	if eval.CrossCheckReady || eval.Passed {
		t.Fatalf("expected observed source count without explicit cross_check_ready not to pass: %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "cross_check_not_ready") {
		t.Fatalf("expected cross_check_not_ready failure, got %q", eval.FailureReason)
	}
}

func TestEvaluateLatestNewsBriefEvidenceRejectsUngroundedSourceEvidence(t *testing.T) {
	eval := EvaluateLatestNewsBriefEvidence(LatestNewsBriefEvaluationInput{
		SourceURL:              "https://example.com/news/candidate",
		PublishedAt:            "2026-07-15T13:00:00+08:00",
		NewsFieldsReady:        false,
		GuardStatus:            "missing_news_fields",
		CrossCheckReady:        false,
		SourceEvidenceAccepted: false,
		StopAfterGuardPassed:   false,
	})
	if eval.SourceAccepted || eval.Passed {
		t.Fatalf("expected URL-only candidate not to be accepted evidence: %#v", eval)
	}
	if !strings.Contains(eval.FailureReason, "source_unaccepted") {
		t.Fatalf("expected source_unaccepted failure, got %q", eval.FailureReason)
	}
}
