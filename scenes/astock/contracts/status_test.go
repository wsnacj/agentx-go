package contracts

import "testing"

func TestBuildReadinessPassed(t *testing.T) {
	got := BuildReadiness(AdapterStatusOK, FailureCodeNone, true, nil, nil)
	if !got.AnswerReady || got.Degraded || got.DegradeReason != "" || got.NextRepairHint != "" {
		t.Fatalf("expected answer-ready readiness, got %#v", got)
	}
}

func TestBuildReadinessMissingFieldsCanDegrade(t *testing.T) {
	got := BuildReadiness(AdapterStatusEvidenceIncomplete, FailureCodeMissingFields, false, []string{"pe_ttm"}, nil)
	if got.AnswerReady || !got.Degraded || got.DegradeReason != string(FailureCodeMissingFields) || got.NextRepairHint != "fetch_missing_fields" {
		t.Fatalf("expected missing-field degraded readiness, got %#v", got)
	}
	if len(got.MissingFields) != 1 || got.MissingFields[0] != "pe_ttm" {
		t.Fatalf("expected missing fields to be preserved, got %#v", got.MissingFields)
	}
}

func TestBuildReadinessIdentityNotFoundBlocksAnswer(t *testing.T) {
	got := BuildReadiness(AdapterStatusUnavailable, FailureCodeIdentityNotFound, false, nil, nil)
	if got.AnswerReady || got.Degraded || got.DegradeReason != string(FailureCodeIdentityNotFound) || got.NextRepairHint != "resolve_verified_a_share_subject" {
		t.Fatalf("expected identity-not-found blocking readiness, got %#v", got)
	}
}

func TestBuildReadinessReviewRequired(t *testing.T) {
	got := BuildReadiness(AdapterStatusNeedsReview, FailureCodeReviewRequired, true, nil, []string{"freshness"})
	if got.AnswerReady || !got.Degraded || got.DegradeReason != string(FailureCodeReviewRequired) || got.NextRepairHint != "verify_review_required_fields" {
		t.Fatalf("expected review-required readiness, got %#v", got)
	}
}
