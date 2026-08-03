package finance

import (
	"strings"

	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

const (
	FinanceAssessmentStatusReady        = "ready"
	FinanceAssessmentStatusPartial      = "partial"
	FinanceAssessmentStatusInsufficient = "insufficient"

	FinanceAssessmentCashFlowAvailable    = "available"
	FinanceAssessmentCashFlowMissing      = "missing"
	FinanceAssessmentCashFlowNotRequested = "not_requested"
)

// FinanceReportAssessment is a source-neutral projection for assessment-style
// answers. It is intentionally conservative: it exposes evidence boundaries
// and missing inputs, but does not make a buy/sell recommendation.
type FinanceReportAssessment struct {
	Requested           bool                         `json:"requested"`
	Kind                string                       `json:"kind,omitempty"`
	Scope               string                       `json:"scope,omitempty"`
	Status              string                       `json:"status,omitempty"`
	VerifiedFacts       []FinanceReportVerifiedFact  `json:"verified_facts,omitempty"`
	PositiveFactors     []FinanceReportAssessmentRef `json:"positive_factors,omitempty"`
	RiskFactors         []FinanceReportAssessmentRef `json:"risk_factors,omitempty"`
	MissingInputs       []string                     `json:"missing_inputs,omitempty"`
	CashFlowStatus      string                       `json:"cash_flow_status,omitempty"`
	CashFlowValue       string                       `json:"cash_flow_value,omitempty"`
	AdviceBoundary      string                       `json:"advice_boundary,omitempty"`
	NonAdviceDisclaimer string                       `json:"non_advice_disclaimer,omitempty"`
}

type FinanceReportVerifiedFact struct {
	Field  string `json:"field"`
	Value  string `json:"value"`
	Period string `json:"period,omitempty"`
	Source string `json:"source,omitempty"`
}

type FinanceReportAssessmentRef struct {
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

// FinanceReportAssessmentFromPayload projects lookup evidence into reusable
// assessment boundaries for host renderers and evaluators.
func FinanceReportAssessmentFromPayload(payload FinanceReportLookupPayload) (FinanceReportAssessment, bool) {
	kind, scope, requiresValuation := lookupAssessmentIntent(payload)
	requested := kind != "" && kind != financialreportmetrics.AssessmentKindNone
	if !requested {
		return FinanceReportAssessment{}, false
	}

	readiness := FinanceReportLookupAnswerReadinessFromPayload(payload)
	cashFlowRequired := assessmentRequiresCashFlow(payload, kind, readiness)
	out := FinanceReportAssessment{
		Requested:           true,
		Kind:                kind,
		Scope:               firstAssessmentNonEmpty(scope, financialreportmetrics.AssessmentScopeMetricsOnly),
		Status:              FinanceAssessmentStatusInsufficient,
		CashFlowStatus:      FinanceAssessmentCashFlowNotRequested,
		AdviceBoundary:      "assessment_requires_guarded_financial_evidence",
		NonAdviceDisclaimer: "This projection is evidence support for analysis, not investment advice.",
	}
	if payload.Metrics != nil {
		evidence := payload.Metrics.Evidence
		source := firstAssessmentNonEmpty(evidence.OfficialSource, payload.Metrics.FinalURL)
		out.VerifiedFacts = lookupAssessmentVerifiedFacts(evidence, source)
		out.CashFlowValue = strings.TrimSpace(evidence.OperatingCashFlow)
		if isAssessmentKnownValue(out.CashFlowValue) {
			out.CashFlowStatus = FinanceAssessmentCashFlowAvailable
		}
		out.PositiveFactors = append(out.PositiveFactors, lookupAssessmentPositiveFactors(evidence)...)
		out.RiskFactors = append(out.RiskFactors, lookupAssessmentRiskFactors(evidence)...)
	}
	out.MissingInputs = uniqueAssessmentStrings(append(out.MissingInputs, readiness.MissingFields...))
	out.MissingInputs = uniqueAssessmentStrings(append(out.MissingInputs, readiness.ReviewRequiredFields...))
	if cashFlowRequired && out.CashFlowStatus != FinanceAssessmentCashFlowAvailable {
		out.CashFlowStatus = FinanceAssessmentCashFlowMissing
		out.MissingInputs = uniqueAssessmentStrings(append(out.MissingInputs, "operating_cash_flow"))
		out.RiskFactors = append(out.RiskFactors, FinanceReportAssessmentRef{
			Code:   "cash_flow_unavailable",
			Detail: "operating cash flow is unavailable or not guard-passed",
		})
	}
	if requiresValuation {
		out.MissingInputs = uniqueAssessmentStrings(append(out.MissingInputs, "valuation_metrics"))
		out.RiskFactors = append(out.RiskFactors, FinanceReportAssessmentRef{
			Code:   "valuation_unavailable",
			Detail: "investment-risk assessment requested valuation context, but valuation evidence is not part of the verified report metrics",
		})
	}
	if readiness.AnswerReady && len(out.MissingInputs) == 0 {
		out.Status = FinanceAssessmentStatusReady
		out.AdviceBoundary = "facts_support_metrics_level_assessment_only"
	} else if len(out.VerifiedFacts) > 0 {
		out.Status = FinanceAssessmentStatusPartial
		out.AdviceBoundary = "partial_assessment_only_missing_inputs_must_be_disclosed"
	}
	out.PositiveFactors = uniqueAssessmentRefs(out.PositiveFactors)
	out.RiskFactors = uniqueAssessmentRefs(out.RiskFactors)
	return out, true
}

func assessmentRequiresCashFlow(payload FinanceReportLookupPayload, kind string, readiness FinanceReportAnswerReadiness) bool {
	if kind == financialreportmetrics.AssessmentKindInvestmentRisk {
		return true
	}
	if assessmentFieldsContain(readiness.MissingFields, "operating_cash_flow") || assessmentFieldsContain(readiness.ReviewRequiredFields, "operating_cash_flow") {
		return true
	}
	for _, field := range payload.Intent.RequestedMetrics {
		if strings.EqualFold(strings.TrimSpace(field), "operating_cash_flow") {
			return true
		}
	}
	if payload.Metrics != nil {
		for _, field := range payload.Metrics.RequestedFields {
			if strings.EqualFold(strings.TrimSpace(field), "operating_cash_flow") {
				return true
			}
		}
	}
	return false
}

func lookupAssessmentIntent(payload FinanceReportLookupPayload) (string, string, bool) {
	kind := normalizeAssessmentKindString(stringFromAny(payload.Intent.Assessment["kind"]))
	scope := strings.TrimSpace(stringFromAny(payload.Intent.Assessment["scope"]))
	requiresValuation := boolFromAny(payload.Intent.Assessment["requires_valuation"])
	for _, output := range payload.Intent.RequestedOutputs {
		switch normalizeAssessmentKindString(output) {
		case financialreportmetrics.AssessmentKindBusinessPerformance:
			if kind == "" || kind == financialreportmetrics.AssessmentKindNone {
				kind = financialreportmetrics.AssessmentKindBusinessPerformance
			}
		case financialreportmetrics.AssessmentKindInvestmentRisk:
			kind = financialreportmetrics.AssessmentKindInvestmentRisk
			requiresValuation = true
		}
	}
	if payload.Metrics != nil {
		if kind == "" || kind == financialreportmetrics.AssessmentKindNone {
			kind = normalizeAssessmentKindString(payload.Metrics.AssessmentKind)
		}
		if strings.TrimSpace(scope) == "" {
			scope = strings.TrimSpace(payload.Metrics.AssessmentScope)
		}
		requiresValuation = requiresValuation || payload.Metrics.AssessmentRequiresValuation
	}
	if kind == "" {
		kind = financialreportmetrics.AssessmentKindNone
	}
	return kind, scope, requiresValuation
}

func lookupAssessmentVerifiedFacts(evidence MetricsEvidence, source string) []FinanceReportVerifiedFact {
	period := strings.TrimSpace(evidence.ReportPeriod)
	facts := []FinanceReportVerifiedFact{}
	for _, item := range []struct {
		field string
		value string
	}{
		{field: "revenue", value: evidence.Revenue},
		{field: "revenue_growth", value: evidence.RevenueGrowth},
		{field: "net_profit", value: evidence.NetProfit},
		{field: "net_profit_growth", value: evidence.NetProfitGrowth},
		{field: "operating_cash_flow", value: evidence.OperatingCashFlow},
	} {
		if !isAssessmentKnownValue(item.value) {
			continue
		}
		factPeriod := period
		if fieldEvidence, ok := evidence.MetricEvidence[item.field]; ok && strings.TrimSpace(fieldEvidence.Period) != "" {
			factPeriod = strings.TrimSpace(fieldEvidence.Period)
		}
		facts = append(facts, FinanceReportVerifiedFact{
			Field:  item.field,
			Value:  strings.TrimSpace(item.value),
			Period: factPeriod,
			Source: strings.TrimSpace(source),
		})
	}
	return facts
}

func lookupAssessmentPositiveFactors(evidence MetricsEvidence) []FinanceReportAssessmentRef {
	out := []FinanceReportAssessmentRef{}
	if isAssessmentKnownValue(evidence.Revenue) && isAssessmentKnownValue(evidence.NetProfit) {
		out = append(out, FinanceReportAssessmentRef{Code: "core_metrics_available", Detail: "revenue and net profit evidence are available"})
	}
	if isAssessmentKnownValue(evidence.RevenueGrowth) || isAssessmentKnownValue(evidence.NetProfitGrowth) {
		out = append(out, FinanceReportAssessmentRef{Code: "growth_evidence_available", Detail: "growth evidence is available"})
	}
	if isAssessmentKnownValue(evidence.OperatingCashFlow) {
		out = append(out, FinanceReportAssessmentRef{Code: "cash_flow_available", Detail: "operating cash flow evidence is available"})
	}
	return out
}

func lookupAssessmentRiskFactors(evidence MetricsEvidence) []FinanceReportAssessmentRef {
	out := []FinanceReportAssessmentRef{}
	if !isAssessmentKnownValue(evidence.Revenue) || !isAssessmentKnownValue(evidence.NetProfit) {
		out = append(out, FinanceReportAssessmentRef{Code: "core_metrics_incomplete", Detail: "revenue or net profit evidence is missing"})
	}
	if !isAssessmentKnownValue(evidence.RevenueGrowth) || !isAssessmentKnownValue(evidence.NetProfitGrowth) {
		out = append(out, FinanceReportAssessmentRef{Code: "growth_evidence_incomplete", Detail: "revenue or net profit growth evidence is missing"})
	}
	return out
}

func normalizeAssessmentKindString(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "none", "no_assessment":
		return financialreportmetrics.AssessmentKindNone
	case financialreportmetrics.AssessmentKindBusinessPerformance, "performance_assessment", "performance", "financial_performance":
		return financialreportmetrics.AssessmentKindBusinessPerformance
	case financialreportmetrics.AssessmentKindInvestmentRisk, "investment", "investment_assessment", "investment_judgment":
		return financialreportmetrics.AssessmentKindInvestmentRisk
	default:
		return ""
	}
}

func isAssessmentKnownValue(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && !strings.EqualFold(value, "unknown")
}

func assessmentFieldsContain(values []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == target || strings.HasSuffix(value, "."+target) {
			return true
		}
	}
	return false
}

func uniqueAssessmentStrings(values []string) []string {
	out := []string{}
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

func uniqueAssessmentRefs(values []FinanceReportAssessmentRef) []FinanceReportAssessmentRef {
	out := []FinanceReportAssessmentRef{}
	seen := map[string]bool{}
	for _, value := range values {
		value.Code = strings.TrimSpace(value.Code)
		value.Detail = strings.TrimSpace(value.Detail)
		if value.Code == "" || seen[value.Code] {
			continue
		}
		seen[value.Code] = true
		out = append(out, value)
	}
	return out
}

func firstAssessmentNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return ""
	}
}

func boolFromAny(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "1", "yes":
			return true
		default:
			return false
		}
	default:
		return false
	}
}
