package profile

import (
	"strings"
	"testing"
)

func TestPromoteReviewedProposalBuildsVerifiedProfileAndRegression(t *testing.T) {
	artifact, err := PromoteReviewedProposal(ReviewDecision{
		Decision:     ReviewDecisionApprove,
		ReviewID:     "review-001",
		ReviewedBy:   "host-reviewer",
		ReviewedAt:   "2026-05-26T08:00:00Z",
		Proposal:     reviewRequiredProposal(),
		ProfileID:    "commercial-invoice-v1",
		DocumentType: "commercial_invoice",
		Version:      "v1",
		Description:  "Commercial invoice fields approved by host review.",
		FieldKeys:    []string{"amount", "invoice_date", "amount"},
		RouteHints:   []string{"ocrx_html_llm"},
		Metadata:     map[string]any{"owner": "host"},
		RegressionCase: RegressionCaseSeed{
			CaseID:                "commercial_invoice_regression_001",
			CaseType:              "document.verify_fields",
			NaturalLanguagePrompt: "从这份商业发票里抽金额和开票日期，并给证据位置。",
			DocumentRef:           "host-redacted/commercial-invoice/sample-001",
			RequiredFields:        []string{"amount", "invoice_date"},
			MinFieldCount:         2,
		},
	})
	if err != nil {
		t.Fatalf("PromoteReviewedProposal returned error: %v", err)
	}
	if artifact.Status != PromotionStatusApproved || artifact.Profile == nil || artifact.RegressionCase == nil {
		t.Fatalf("unexpected artifact: %#v", artifact)
	}
	if artifact.Profile.ID != "commercial-invoice-v1" || artifact.Profile.DocumentType != "commercial_invoice" {
		t.Fatalf("unexpected profile: %#v", artifact.Profile)
	}
	if got := artifact.Profile.FieldKeys; len(got) != 2 || got[0] != "amount" || got[1] != "invoice_date" {
		t.Fatalf("field keys = %#v, want normalized unique keys", got)
	}
	if artifact.Profile.Metadata["promotion_source"] != "host_review" ||
		artifact.Profile.Metadata["review_id"] != "review-001" ||
		artifact.Profile.Metadata["proposal_source"] != "referencehost" {
		t.Fatalf("profile metadata missing audit data: %#v", artifact.Profile.Metadata)
	}
	if artifact.RegressionCase.CaseID != "commercial_invoice_regression_001" {
		t.Fatalf("unexpected regression case: %#v", artifact.RegressionCase)
	}
}

func TestPromoteReviewedProposalDoesNotPromoteSuggestionsWithoutExplicitApproval(t *testing.T) {
	_, err := PromoteReviewedProposal(ReviewDecision{
		Decision:   ReviewDecisionApprove,
		ReviewedBy: "host-reviewer",
		Proposal: Proposal{
			Source:             "referencehost",
			SuggestedProfileID: "model-suggested-invoice",
			SuggestedFields:    []string{"amount", "invoice_date"},
			ReviewRequired:     true,
		},
		RegressionCase: RegressionCaseSeed{
			CaseID:                "suggestion_only",
			NaturalLanguagePrompt: "抽金额。",
			DocumentRef:           "host-redacted/doc/sample-001",
			RequiredFields:        []string{"amount"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "profile_id_required") {
		t.Fatalf("expected explicit profile_id requirement, got %v", err)
	}
}

func TestPromoteReviewedProposalRejectsUnreviewedProposal(t *testing.T) {
	_, err := PromoteReviewedProposal(ReviewDecision{
		Decision:     ReviewDecisionApprove,
		ReviewedBy:   "host-reviewer",
		Proposal:     Proposal{Source: "referencehost", ReviewRequired: false},
		ProfileID:    "invoice-v1",
		DocumentType: "invoice",
		FieldKeys:    []string{"amount"},
		RegressionCase: RegressionCaseSeed{
			CaseID:                "invoice_regression",
			NaturalLanguagePrompt: "抽金额。",
			DocumentRef:           "host-redacted/invoice/sample-001",
			RequiredFields:        []string{"amount"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "proposal_review_required_missing") {
		t.Fatalf("expected review-required proposal guard, got %v", err)
	}
}

func TestPromoteReviewedProposalRejectsMissingRegressionCase(t *testing.T) {
	_, err := PromoteReviewedProposal(ReviewDecision{
		Decision:     ReviewDecisionApprove,
		ReviewedBy:   "host-reviewer",
		Proposal:     reviewRequiredProposal(),
		ProfileID:    "invoice-v1",
		DocumentType: "invoice",
		FieldKeys:    []string{"amount"},
	})
	if err == nil || !strings.Contains(err.Error(), "regression_case_id_required") {
		t.Fatalf("expected regression case requirement, got %v", err)
	}
}

func TestPromoteReviewedProposalRecordsRejectDecision(t *testing.T) {
	artifact, err := PromoteReviewedProposal(ReviewDecision{
		Decision:   ReviewDecisionReject,
		ReviewID:   "review-reject-001",
		ReviewedBy: "host-reviewer",
		Proposal:   reviewRequiredProposal(),
	})
	if err != nil {
		t.Fatalf("PromoteReviewedProposal reject returned error: %v", err)
	}
	if artifact.Status != PromotionStatusRejected || artifact.Profile != nil || artifact.RegressionCase != nil {
		t.Fatalf("unexpected reject artifact: %#v", artifact)
	}
}

func reviewRequiredProposal() Proposal {
	return Proposal{
		Reason:          "unknown document profile",
		Source:          "referencehost",
		RequestedFields: []string{"amount", "invoice_date"},
		EvidenceSnippets: []string{
			"Commercial Invoice",
			"Invoice Date 2026-05-20",
		},
		ReviewRequired: true,
	}
}
