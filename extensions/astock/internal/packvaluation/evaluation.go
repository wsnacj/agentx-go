package packvaluation

import (
	"net/url"
	"strings"

	astockcontracts "github.com/wsnacj/agentx-go/extensions/astock/contracts"
)

type ValuationEvaluationInput struct {
	ExpectedEntityName        string
	ExpectedStockCode         string
	EvidenceEntityName        string
	EvidenceStockCode         string
	AdapterStatus             astockcontracts.AdapterStatus
	FailureCode               astockcontracts.FailureCode
	AnswerReady               bool
	RequestedFields           []string
	FieldValues               map[string]string
	AsOf                      string
	SourceURL                 string
	MissingRequestedFields    []string
	ReviewRequiredFields      []string
	InvestmentAdviceRequested bool
	AdviceBoundaryStated      bool
}

type ValuationEvaluation struct {
	Passed                  bool     `json:"passed"`
	SubjectCorrect          bool     `json:"subject_correct"`
	FreshnessAccepted       bool     `json:"freshness_accepted"`
	FieldsReady             bool     `json:"fields_ready"`
	SourceAccepted          bool     `json:"source_accepted"`
	AdviceBoundaryRespected bool     `json:"advice_boundary_respected"`
	MissingRequestedFields  []string `json:"missing_requested_fields,omitempty"`
	ReviewRequiredFields    []string `json:"review_required_fields,omitempty"`
	RequestedFields         []string `json:"requested_fields,omitempty"`
	SourceURL               string   `json:"source_url,omitempty"`
	AdapterStatus           string   `json:"adapter_status,omitempty"`
	FailureCode             string   `json:"failure_code,omitempty"`
	FailureReason           string   `json:"failure_reason,omitempty"`
}

func EvaluateValuationEvidence(input ValuationEvaluationInput) ValuationEvaluation {
	missing := normalizeStringList(input.MissingRequestedFields)
	review := normalizeStringList(input.ReviewRequiredFields)
	requested := normalizeStringList(input.RequestedFields)
	out := ValuationEvaluation{
		SubjectCorrect:          valuationSubjectCorrect(input),
		FreshnessAccepted:       strings.TrimSpace(input.AsOf) != "",
		FieldsReady:             input.AnswerReady && len(missing) == 0 && len(review) == 0 && valuationFieldsCovered(requested, input.FieldValues),
		SourceAccepted:          valuationSourceAccepted(input.SourceURL),
		AdviceBoundaryRespected: !input.InvestmentAdviceRequested || input.AdviceBoundaryStated,
		MissingRequestedFields:  missing,
		ReviewRequiredFields:    review,
		RequestedFields:         requested,
		SourceURL:               strings.TrimSpace(input.SourceURL),
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
		if !valuationFieldsCovered(requested, input.FieldValues) {
			reasons = append(reasons, "requested_quote_fields_missing")
		}
		if len(missing) == 0 && len(review) == 0 && valuationFieldsCovered(requested, input.FieldValues) {
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

func valuationSubjectCorrect(input ValuationEvaluationInput) bool {
	expectedCode, _, expectedOK := astockcontracts.NormalizeAStockCode(input.ExpectedStockCode)
	evidenceCode, _, evidenceOK := astockcontracts.NormalizeAStockCode(input.EvidenceStockCode)
	if expectedOK {
		return evidenceOK && evidenceCode == expectedCode
	}
	expectedName := normalizeEntityName(input.ExpectedEntityName)
	evidenceName := normalizeEntityName(input.EvidenceEntityName)
	return expectedName != "" && evidenceName != "" && (strings.Contains(evidenceName, expectedName) || strings.Contains(expectedName, evidenceName))
}

func valuationFieldsCovered(requested []string, values map[string]string) bool {
	if len(values) == 0 {
		return false
	}
	if len(requested) == 0 {
		for _, field := range []string{"price", "pe_ttm", "pb", "market_cap"} {
			if fieldValueReady(values[field]) {
				return true
			}
		}
		return false
	}
	for _, field := range requested {
		if !fieldValueReady(valuationFieldValue(values, field)) {
			return false
		}
	}
	return true
}

func valuationFieldValue(values map[string]string, field string) string {
	field = strings.TrimSpace(field)
	if value := values[field]; strings.TrimSpace(value) != "" {
		return value
	}
	switch field {
	case "change_pct":
		return values["change_percent"]
	case "turnover":
		return values["turnover_percent"]
	default:
		return ""
	}
}

func fieldValueReady(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && !strings.EqualFold(value, "unknown") && value != "--"
}

func valuationSourceAccepted(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
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
