package packresearch

import (
	"net/url"
	"strings"

	astockcontracts "github.com/wsnacj/agentx-go/extensions/astock/contracts"
)

type ResearchEvaluationInput = astockcontracts.ResearchEvaluationInput
type ResearchEvaluation = astockcontracts.ResearchEvaluation

func EvaluateResearchEvidence(input ResearchEvaluationInput) ResearchEvaluation {
	missing := normalizeStringList(input.MissingRequestedFields)
	review := normalizeStringList(input.ReviewRequiredFields)
	requested := normalizeStringList(input.RequestedFields)
	sources := normalizeStringList(input.SourceURLs)
	out := ResearchEvaluation{
		SubjectCorrect:          researchSubjectCorrect(input),
		FreshnessAccepted:       strings.TrimSpace(input.LatestPublishedAt) != "" || strings.TrimSpace(input.AsOf) != "",
		FieldsReady:             input.AnswerReady && input.ReportCount > 0 && len(missing) == 0 && len(review) == 0 && researchFieldsCovered(requested, input.FieldValues, input.ConsensusFields),
		SourceAccepted:          researchSourcesAccepted(sources),
		AdviceBoundaryRespected: !input.InvestmentAdviceRequested || input.AdviceBoundaryStated,
		MissingRequestedFields:  missing,
		ReviewRequiredFields:    review,
		RequestedFields:         requested,
		SourceURLs:              sources,
		AdapterStatus:           string(input.AdapterStatus),
		FailureCode:             string(input.FailureCode),
	}
	reasons := []string{}
	if input.AdapterStatus != astockcontracts.AdapterStatusOK {
		reasons = append(reasons, "adapter_not_ok")
	}
	if input.FailureCode != astockcontracts.FailureCodeNone {
		reasons = append(reasons, "failure_code_present")
	}
	if !input.AnswerReady {
		reasons = append(reasons, "answer_not_ready")
	}
	if input.ReportCount <= 0 {
		reasons = append(reasons, "research_reports_missing")
	}
	if !out.SubjectCorrect {
		reasons = append(reasons, "subject_mismatch")
	}
	if !out.FreshnessAccepted {
		reasons = append(reasons, "freshness_missing")
	}
	if !out.FieldsReady {
		if len(missing) > 0 {
			reasons = append(reasons, "requested_fields_missing")
		}
		if len(review) > 0 {
			reasons = append(reasons, "review_required_fields")
		}
		if !researchFieldsCovered(requested, input.FieldValues, input.ConsensusFields) {
			reasons = append(reasons, "requested_research_fields_missing")
		}
		if len(missing) == 0 && len(review) == 0 && researchFieldsCovered(requested, input.FieldValues, input.ConsensusFields) {
			reasons = append(reasons, "fields_not_ready")
		}
	}
	if !out.SourceAccepted {
		reasons = append(reasons, "source_unaccepted")
	}
	if !out.AdviceBoundaryRespected {
		reasons = append(reasons, "advice_boundary_missing")
	}
	out.Passed = len(reasons) == 0
	out.FailureReason = strings.Join(reasons, ",")
	return out
}

func researchSubjectCorrect(input ResearchEvaluationInput) bool {
	expectedCode, _, expectedOK := astockcontracts.NormalizeAStockCode(input.ExpectedStockCode)
	evidenceCode, _, evidenceOK := astockcontracts.NormalizeAStockCode(input.EvidenceStockCode)
	if expectedOK {
		return evidenceOK && evidenceCode == expectedCode
	}
	expectedName := normalizeEntityName(input.ExpectedEntityName)
	evidenceName := normalizeEntityName(input.EvidenceEntityName)
	return expectedName != "" && evidenceName != "" && (strings.Contains(evidenceName, expectedName) || strings.Contains(expectedName, evidenceName))
}

func researchFieldsCovered(requested []string, values map[string]string, consensus []string) bool {
	if len(requested) == 0 {
		return len(values) > 0 || len(consensus) > 0
	}
	consensusSet := map[string]bool{}
	for _, field := range normalizeStringList(consensus) {
		consensusSet[field] = true
	}
	for _, field := range requested {
		if fieldValueReady(values[field]) || consensusSet[field] {
			continue
		}
		return false
	}
	return true
}

func fieldValueReady(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && !strings.EqualFold(value, "unknown") && value != "--"
}

func researchSourcesAccepted(values []string) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		parsed, err := url.Parse(strings.TrimSpace(value))
		if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return false
		}
	}
	return true
}

func normalizeStringList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func normalizeEntityName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "unknown" {
		return ""
	}
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "", "有限公司", "", "股份", "", "公司", "")
	return replacer.Replace(value)
}
