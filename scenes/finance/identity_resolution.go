package finance

import "strings"

// EnsureMetricsCandidatesIdentityResolution fills candidate diagnostics when an
// adapter already verified issuer candidates but did not provide an explicit
// resolution payload.
func EnsureMetricsCandidatesIdentityResolution(payload *MetricsCandidatesPayload) {
	if payload == nil || payload.IdentityResolution != nil {
		return
	}
	payload.IdentityResolution = IdentityResolutionFromMetricsCandidates(*payload)
}

// IdentityResolutionFromMetricsCandidates projects existing report candidate
// evidence into a diagnostic-only identity-resolution contract.
func IdentityResolutionFromMetricsCandidates(payload MetricsCandidatesPayload) *ReportIdentityResolution {
	candidates := make([]ReportIdentityResolutionCandidate, 0, len(payload.ResolvedEntities))
	for _, entity := range payload.ResolvedEntities {
		candidates = append(candidates, reportResolutionCandidateFromResolvedEntity(entity))
	}
	selected := selectReportResolutionCandidate(candidates)
	if selected == nil && strings.TrimSpace(payload.ResolvedCode) != "" {
		candidate := ReportIdentityResolutionCandidate{
			EntityName:     strings.TrimSpace(payload.ResolvedCompany),
			CodeOrTicker:   strings.TrimSpace(payload.ResolvedCode),
			Market:         strings.TrimSpace(payload.ResolvedMarket),
			Source:         strings.TrimSpace(payload.AdapterID),
			EvidenceURL:    firstNonEmptyIdentityResolution(payload.PrimaryURL, payload.SourceURL),
			Confidence:     0.75,
			MatchReason:    "resolved_from_metrics_candidates_payload",
			Selected:       true,
			SelectedReason: "resolved_code_selected",
		}
		candidates = append(candidates, candidate)
		selected = &candidates[len(candidates)-1]
	}
	if len(candidates) == 0 && strings.TrimSpace(payload.EntityName) == "" && strings.TrimSpace(payload.ResolvedCompany) == "" {
		return nil
	}
	status := strings.TrimSpace(payload.AdapterStatus)
	if status == "" {
		status = "unknown"
	}
	inputTerm := firstNonEmptyIdentityResolution(payload.EntityName, firstStringIdentityResolution(payload.EntityMentions), payload.ResolvedCompany, payload.ResolvedCode)
	resolution := ReportIdentityResolution{
		InputTerm:       inputTerm,
		PreferredMarket: strings.TrimSpace(payload.ResolvedMarket),
		Strategy:        firstNonEmptyIdentityResolution(payload.AdapterID, "report_metrics_candidates"),
		QueryVariants: []ReportIdentityResolutionQuery{{
			Term:           inputTerm,
			Reason:         "report_candidate_generation",
			Provider:       firstNonEmptyIdentityResolution(payload.AdapterID, payload.Source),
			Status:         status,
			CandidateCount: len(candidates),
			Message:        strings.TrimSpace(payload.FailureCode),
		}},
		Candidates: candidates,
		Warnings:   append([]string(nil), payload.Warnings...),
	}
	if selected != nil {
		selectedCopy := *selected
		selectedCopy.Selected = true
		if strings.TrimSpace(selectedCopy.SelectedReason) == "" {
			selectedCopy.SelectedReason = firstNonEmptyIdentityResolution(selectedCopy.MatchReason, "report_candidate_selected")
		}
		resolution.SelectedReason = selectedCopy.SelectedReason
		resolution.SelectedCandidate = &selectedCopy
		for i := range resolution.Candidates {
			if sameReportResolutionCandidate(resolution.Candidates[i], selectedCopy) {
				resolution.Candidates[i].Selected = true
				resolution.Candidates[i].SelectedReason = selectedCopy.SelectedReason
			}
		}
	}
	return &resolution
}

func CloneReportIdentityResolution(resolution *ReportIdentityResolution) *ReportIdentityResolution {
	if resolution == nil {
		return nil
	}
	out := *resolution
	out.QueryVariants = append([]ReportIdentityResolutionQuery(nil), resolution.QueryVariants...)
	out.Candidates = append([]ReportIdentityResolutionCandidate(nil), resolution.Candidates...)
	out.Warnings = append([]string(nil), resolution.Warnings...)
	if resolution.SelectedCandidate != nil {
		selected := *resolution.SelectedCandidate
		out.SelectedCandidate = &selected
	}
	return &out
}

func reportResolutionCandidateFromResolvedEntity(entity ResolvedEntityCandidate) ReportIdentityResolutionCandidate {
	return ReportIdentityResolutionCandidate{
		EntityName:     strings.TrimSpace(entity.EntityName),
		CodeOrTicker:   strings.TrimSpace(entity.CodeOrTicker),
		Market:         strings.TrimSpace(entity.Market),
		Source:         strings.TrimSpace(entity.Source),
		EvidenceURL:    strings.TrimSpace(entity.EvidenceURL),
		Confidence:     entity.Confidence,
		MatchReason:    strings.TrimSpace(entity.MatchReason),
		MismatchReason: strings.TrimSpace(entity.MismatchReason),
	}
}

func selectReportResolutionCandidate(candidates []ReportIdentityResolutionCandidate) *ReportIdentityResolutionCandidate {
	for i := range candidates {
		if strings.TrimSpace(candidates[i].MismatchReason) == "" &&
			(strings.TrimSpace(candidates[i].CodeOrTicker) != "" || strings.TrimSpace(candidates[i].EntityName) != "") {
			candidate := candidates[i]
			candidate.Selected = true
			candidate.SelectedReason = firstNonEmptyIdentityResolution(candidate.MatchReason, "report_candidate_selected")
			return &candidate
		}
	}
	return nil
}

func sameReportResolutionCandidate(left ReportIdentityResolutionCandidate, right ReportIdentityResolutionCandidate) bool {
	return strings.EqualFold(strings.TrimSpace(left.CodeOrTicker), strings.TrimSpace(right.CodeOrTicker)) &&
		strings.EqualFold(strings.TrimSpace(left.Market), strings.TrimSpace(right.Market)) &&
		strings.EqualFold(strings.TrimSpace(left.EntityName), strings.TrimSpace(right.EntityName))
}

func firstStringIdentityResolution(values []string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstNonEmptyIdentityResolution(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
