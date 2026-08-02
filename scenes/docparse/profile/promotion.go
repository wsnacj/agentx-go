package profile

import (
	"fmt"
	"strings"
)

const (
	ReviewDecisionApprove           = "approve"
	ReviewDecisionReject            = "reject"
	ReviewDecisionNeedsMoreEvidence = "needs_more_evidence"

	PromotionStatusApproved          = "approved_profile"
	PromotionStatusRejected          = "rejected_profile"
	PromotionStatusNeedsMoreEvidence = "needs_more_evidence"
)

// RegressionCaseSeed is the host-reviewed regression case attached to a newly
// approved profile. It intentionally avoids parser execution or private storage
// concerns; hosts decide where artifacts and parse results live.
type RegressionCaseSeed struct {
	CaseID                   string           `json:"case_id"`
	CaseType                 string           `json:"case_type,omitempty"`
	NaturalLanguagePrompt    string           `json:"natural_language_prompt"`
	Scenario                 string           `json:"scenario,omitempty"`
	DocumentRef              string           `json:"document_ref,omitempty"`
	SpecRef                  string           `json:"spec_ref,omitempty"`
	PageRange                string           `json:"page_range,omitempty"`
	ResultPath               string           `json:"result_path,omitempty"`
	ParseResult              map[string]any   `json:"parse_result,omitempty"`
	RequiredFields           []string         `json:"required_fields,omitempty"`
	ExpectedFields           []map[string]any `json:"expected_fields,omitempty"`
	ExpectedTables           []map[string]any `json:"expected_tables,omitempty"`
	RequiredEvidence         map[string]any   `json:"required_evidence,omitempty"`
	RequirePageRefs          bool             `json:"require_page_refs,omitempty"`
	RequireBoundingBoxes     bool             `json:"require_bounding_boxes,omitempty"`
	RequireTableCells        bool             `json:"require_table_cells,omitempty"`
	RequireCompleteTableRows bool             `json:"require_complete_table_rows,omitempty"`
	MinFieldAccuracy         float64          `json:"min_field_accuracy,omitempty"`
	MinEvidenceHitRate       float64          `json:"min_evidence_hit_rate,omitempty"`
	MinFieldCount            int              `json:"min_field_count,omitempty"`
	MinTableCount            int              `json:"min_table_count,omitempty"`
}

// ReviewDecision is the explicit host decision for an unknown profile proposal.
//
// Suggested profile IDs or fields from a model proposal are not promoted
// automatically. Approval requires explicit host-reviewed profile data.
type ReviewDecision struct {
	Decision       string             `json:"decision"`
	ReviewID       string             `json:"review_id,omitempty"`
	ReviewedBy     string             `json:"reviewed_by"`
	ReviewedAt     string             `json:"reviewed_at,omitempty"`
	Proposal       Proposal           `json:"proposal"`
	ProfileID      string             `json:"profile_id,omitempty"`
	DocumentType   string             `json:"document_type,omitempty"`
	Version        string             `json:"version,omitempty"`
	Description    string             `json:"description,omitempty"`
	SpecPath       string             `json:"spec_path,omitempty"`
	FieldKeys      []string           `json:"field_keys,omitempty"`
	TableKeys      []string           `json:"table_keys,omitempty"`
	RouteHints     []string           `json:"route_hints,omitempty"`
	Tags           []string           `json:"tags,omitempty"`
	Metadata       map[string]any     `json:"metadata,omitempty"`
	RegressionCase RegressionCaseSeed `json:"regression_case,omitempty"`
}

// PromotionArtifact is the auditable output of a host review decision.
type PromotionArtifact struct {
	Status         string              `json:"status"`
	Decision       string              `json:"decision"`
	ReviewID       string              `json:"review_id,omitempty"`
	ReviewedBy     string              `json:"reviewed_by,omitempty"`
	ReviewedAt     string              `json:"reviewed_at,omitempty"`
	Proposal       Proposal            `json:"proposal,omitempty"`
	Profile        *ExtractionProfile  `json:"profile,omitempty"`
	RegressionCase *RegressionCaseSeed `json:"regression_case,omitempty"`
	Warnings       []string            `json:"warnings,omitempty"`
}

// PromoteReviewedProposal applies a host review decision to an unknown-profile
// proposal. It never promotes model suggestions without explicit approval data.
func PromoteReviewedProposal(decision ReviewDecision) (PromotionArtifact, error) {
	action := normalizeReviewDecision(decision.Decision)
	if action == "" {
		return PromotionArtifact{}, fmt.Errorf("review_decision_required")
	}
	reviewedBy := strings.TrimSpace(decision.ReviewedBy)
	if reviewedBy == "" {
		return PromotionArtifact{}, fmt.Errorf("reviewed_by_required")
	}
	artifact := PromotionArtifact{
		Status:     promotionStatusForDecision(action),
		Decision:   action,
		ReviewID:   strings.TrimSpace(decision.ReviewID),
		ReviewedBy: reviewedBy,
		ReviewedAt: strings.TrimSpace(decision.ReviewedAt),
		Proposal:   cloneProposal(decision.Proposal),
	}
	switch action {
	case ReviewDecisionReject, ReviewDecisionNeedsMoreEvidence:
		return artifact, nil
	case ReviewDecisionApprove:
	default:
		return PromotionArtifact{}, fmt.Errorf("review_decision_unsupported:%s", decision.Decision)
	}
	if !decision.Proposal.ReviewRequired {
		return PromotionArtifact{}, fmt.Errorf("proposal_review_required_missing")
	}
	profile, err := reviewedProfile(decision)
	if err != nil {
		return PromotionArtifact{}, err
	}
	regressionCase, err := reviewedRegressionCase(decision.RegressionCase)
	if err != nil {
		return PromotionArtifact{}, err
	}
	artifact.Profile = &profile
	artifact.RegressionCase = &regressionCase
	return artifact, nil
}

func reviewedProfile(decision ReviewDecision) (ExtractionProfile, error) {
	profile := ExtractionProfile{
		ID:           strings.TrimSpace(decision.ProfileID),
		DocumentType: strings.TrimSpace(decision.DocumentType),
		Version:      strings.TrimSpace(decision.Version),
		Description:  strings.TrimSpace(decision.Description),
		SpecPath:     strings.TrimSpace(decision.SpecPath),
		FieldKeys:    uniqueStrings(decision.FieldKeys),
		TableKeys:    uniqueStrings(decision.TableKeys),
		RouteHints:   uniqueStrings(decision.RouteHints),
		Tags:         uniqueStrings(decision.Tags),
		Metadata:     cloneMap(decision.Metadata),
	}
	if profile.ID == "" {
		return ExtractionProfile{}, fmt.Errorf("profile_id_required")
	}
	if profile.DocumentType == "" {
		return ExtractionProfile{}, fmt.Errorf("document_type_required")
	}
	if profile.SpecPath == "" && len(profile.FieldKeys) == 0 && len(profile.TableKeys) == 0 && len(profile.RouteHints) == 0 {
		return ExtractionProfile{}, fmt.Errorf("profile_contract_required")
	}
	if profile.Metadata == nil {
		profile.Metadata = map[string]any{}
	}
	profile.Metadata["promotion_source"] = "host_review"
	if reviewID := strings.TrimSpace(decision.ReviewID); reviewID != "" {
		profile.Metadata["review_id"] = reviewID
	}
	profile.Metadata["reviewed_by"] = strings.TrimSpace(decision.ReviewedBy)
	if proposalSource := strings.TrimSpace(decision.Proposal.Source); proposalSource != "" {
		profile.Metadata["proposal_source"] = proposalSource
	}
	return normalizeProfile(profile), nil
}

func reviewedRegressionCase(seed RegressionCaseSeed) (RegressionCaseSeed, error) {
	seed.CaseID = strings.TrimSpace(seed.CaseID)
	seed.CaseType = strings.TrimSpace(seed.CaseType)
	seed.NaturalLanguagePrompt = strings.TrimSpace(seed.NaturalLanguagePrompt)
	seed.Scenario = strings.TrimSpace(seed.Scenario)
	seed.DocumentRef = strings.TrimSpace(seed.DocumentRef)
	seed.SpecRef = strings.TrimSpace(seed.SpecRef)
	seed.PageRange = strings.TrimSpace(seed.PageRange)
	seed.ResultPath = strings.TrimSpace(seed.ResultPath)
	seed.RequiredFields = uniqueStrings(seed.RequiredFields)
	seed.ExpectedFields = cloneObjectSlice(seed.ExpectedFields)
	seed.ExpectedTables = cloneObjectSlice(seed.ExpectedTables)
	seed.RequiredEvidence = cloneMap(seed.RequiredEvidence)
	if seed.CaseID == "" {
		return RegressionCaseSeed{}, fmt.Errorf("regression_case_id_required")
	}
	if seed.NaturalLanguagePrompt == "" {
		return RegressionCaseSeed{}, fmt.Errorf("regression_case_prompt_required")
	}
	if seed.DocumentRef == "" && seed.ResultPath == "" && len(seed.ParseResult) == 0 {
		return RegressionCaseSeed{}, fmt.Errorf("regression_case_artifact_required")
	}
	if len(seed.RequiredFields) == 0 && len(seed.ExpectedFields) == 0 && len(seed.ExpectedTables) == 0 && seed.MinFieldCount == 0 && seed.MinTableCount == 0 {
		return RegressionCaseSeed{}, fmt.Errorf("regression_case_assertion_required")
	}
	seed.ParseResult = cloneMap(seed.ParseResult)
	return seed, nil
}

func normalizeReviewDecision(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ReviewDecisionApprove, "approved":
		return ReviewDecisionApprove
	case ReviewDecisionReject, "rejected":
		return ReviewDecisionReject
	case ReviewDecisionNeedsMoreEvidence, "needs_more", "needs_review":
		return ReviewDecisionNeedsMoreEvidence
	default:
		return strings.TrimSpace(value)
	}
}

func promotionStatusForDecision(decision string) string {
	switch decision {
	case ReviewDecisionApprove:
		return PromotionStatusApproved
	case ReviewDecisionReject:
		return PromotionStatusRejected
	case ReviewDecisionNeedsMoreEvidence:
		return PromotionStatusNeedsMoreEvidence
	default:
		return ""
	}
}

func cloneProposal(proposal Proposal) Proposal {
	proposal.RequestedFields = append([]string(nil), proposal.RequestedFields...)
	proposal.SuggestedFields = append([]string(nil), proposal.SuggestedFields...)
	proposal.SuggestedRouteHints = append([]string(nil), proposal.SuggestedRouteHints...)
	proposal.CandidateDocumentTypes = append([]CandidateDocumentType(nil), proposal.CandidateDocumentTypes...)
	proposal.EvidenceSnippets = append([]string(nil), proposal.EvidenceSnippets...)
	return proposal
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		key = strings.TrimSpace(key)
		if key != "" {
			out[key] = value
		}
	}
	return out
}

func cloneObjectSlice(items []map[string]any) []map[string]any {
	if len(items) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if cloned := cloneMap(item); len(cloned) > 0 {
			out = append(out, cloned)
		}
	}
	return out
}
