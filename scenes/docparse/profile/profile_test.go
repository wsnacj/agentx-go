package profile

import "testing"

func TestMatcherVerifiesExplicitProfileID(t *testing.T) {
	reg := NewRegistry(ExtractionProfile{ID: "invoice-v1", DocumentType: "invoice"})
	got := NewMatcher(reg).Match(MatchInput{ExplicitProfileID: " invoice_v1 "})
	if got.Status != MatchStatusVerified || got.Profile == nil || got.Profile.ID != "invoice-v1" {
		t.Fatalf("unexpected match: %#v", got)
	}
}

func TestMatcherReturnsCandidateForExplicitDocumentType(t *testing.T) {
	reg := NewRegistry(
		ExtractionProfile{ID: "invoice-v1", DocumentType: "invoice"},
		ExtractionProfile{ID: "contract-v1", DocumentType: "contract"},
	)
	got := NewMatcher(reg).Match(MatchInput{ExplicitDocumentType: "Invoice"})
	if got.Status != MatchStatusCandidate || len(got.Candidates) != 1 || got.Candidates[0].Profile.ID != "invoice-v1" {
		t.Fatalf("unexpected candidates: %#v", got)
	}
}

func TestMatcherKeepsNonLatinDocumentTypes(t *testing.T) {
	reg := NewRegistry(ExtractionProfile{ID: "cn-invoice-v1", DocumentType: "发票"})
	got := NewMatcher(reg).Match(MatchInput{ExplicitDocumentType: "发票"})
	if got.Status != MatchStatusCandidate || len(got.Candidates) != 1 {
		t.Fatalf("expected Chinese document type candidate: %#v", got)
	}
}

func TestMatcherUnknownProducesProposalWithoutBusinessGuess(t *testing.T) {
	got := NewMatcher(NewRegistry()).Match(MatchInput{RequestedFields: []string{"amount", "date"}})
	if got.Status != MatchStatusUnknown || got.Proposal == nil {
		t.Fatalf("expected unknown proposal: %#v", got)
	}
	if !got.Proposal.ReviewRequired || got.Proposal.Source != "profile_matcher" {
		t.Fatalf("expected review-required matcher proposal: %#v", got.Proposal)
	}
	if len(got.Proposal.RequestedFields) != 2 {
		t.Fatalf("expected requested fields to be preserved: %#v", got.Proposal)
	}
}
