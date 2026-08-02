package packsignal

import (
	"net/url"
	"strings"

	astockcontracts "github.com/wsnacj/agentx-go/extensions/astock/contracts"
)

type SignalEvaluationInput = astockcontracts.SignalEvaluationInput
type SignalEvaluation = astockcontracts.SignalEvaluation

func EvaluateSignalEvidence(input SignalEvaluationInput) SignalEvaluation {
	missing := normalizeStringList(input.MissingRequestedFields)
	review := normalizeStringList(input.ReviewRequiredFields)
	requested := normalizeStringList(input.RequestedSignalTypes)
	returned := normalizeStringList(input.ReturnedSignalTypes)
	sources := normalizeStringList(input.SourceURLs)
	out := SignalEvaluation{
		SubjectCorrect:          signalSubjectCorrect(input),
		FreshnessAccepted:       strings.TrimSpace(input.TradeDate) != "" || strings.TrimSpace(input.AsOf) != "",
		FieldsReady:             input.AnswerReady && len(missing) == 0 && len(review) == 0 && signalTypesCovered(requested, returned),
		SourceAccepted:          signalSourcesAccepted(sources),
		AdviceBoundaryRespected: !input.InvestmentAdviceRequested || input.AdviceBoundaryStated,
		MissingRequestedFields:  missing,
		ReviewRequiredFields:    review,
		RequestedSignalTypes:    requested,
		ReturnedSignalTypes:     returned,
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
		if !signalTypesCovered(requested, returned) {
			reasons = append(reasons, "requested_signal_types_missing")
		}
		if len(missing) == 0 && len(review) == 0 && signalTypesCovered(requested, returned) {
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

func signalSubjectCorrect(input SignalEvaluationInput) bool {
	expectedCode, _, expectedOK := astockcontracts.NormalizeAStockCode(input.ExpectedStockCode)
	evidenceCode, _, evidenceOK := astockcontracts.NormalizeAStockCode(input.EvidenceStockCode)
	if expectedOK {
		return evidenceOK && evidenceCode == expectedCode
	}
	expectedName := normalizeEntityName(input.ExpectedEntityName)
	if expectedName == "" {
		return true
	}
	evidenceName := normalizeEntityName(input.EvidenceEntityName)
	return evidenceName != "" && (strings.Contains(evidenceName, expectedName) || strings.Contains(expectedName, evidenceName))
}

func signalTypesCovered(requested []string, returned []string) bool {
	if len(requested) == 0 {
		return len(returned) > 0
	}
	returnedSet := map[string]bool{}
	for _, item := range returned {
		returnedSet[item] = true
	}
	for _, item := range requested {
		if !returnedSet[item] {
			return false
		}
	}
	return true
}

func signalSourcesAccepted(values []string) bool {
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
