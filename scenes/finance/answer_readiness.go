package finance

import (
	"strings"

	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

const (
	AnswerDegradeNone                  = ""
	AnswerDegradeMissingEvidence       = "missing_evidence"
	AnswerDegradeGuardNotPassed        = "guard_not_passed"
	AnswerDegradeMissingRequested      = "missing_requested_fields"
	AnswerDegradeReviewRequired        = "review_required"
	AnswerDegradePeriodScopeReview     = "period_scope_review_required"
	AnswerDegradeBriefNotReady         = "brief_not_ready"
	AnswerDegradeOfficialSourceMissing = "official_source_missing_or_unparseable"
	AnswerDegradeSourceDownloadBlocked = "source_download_blocked"
	AnswerDegradeSourceReturnedNonPDF  = "source_returned_non_pdf"
	AnswerScopeRequested               = "requested_scope"
	AnswerScopeAvailableVerified       = "available_verified_scope"
	AnswerScopeAvailablePeriodOnly     = "available_verified_period_only"
	AnswerScopePartialVerifiedMetrics  = "partial_verified_metrics"
	AnswerScopeSourceMetadataOnly      = "source_metadata_only"
	AnswerScopeNone                    = "none"
)

// FinanceReportAnswerReadiness is a reusable final-answer gate projection.
// Hosts own wording, but should not need to rediscover guard-stop semantics.
type FinanceReportAnswerReadiness struct {
	AnswerReady              bool     `json:"answer_ready"`
	Degraded                 bool     `json:"degraded,omitempty"`
	DegradeReason            string   `json:"degrade_reason,omitempty"`
	AllowedSummaryScope      string   `json:"allowed_summary_scope,omitempty"`
	PresentationRequirements []string `json:"presentation_requirements,omitempty"`
	NextRepairHint           string   `json:"next_repair_hint,omitempty"`
	StopRecommended          bool     `json:"stop_recommended,omitempty"`
	StopReason               string   `json:"stop_reason,omitempty"`
	GuardStatus              string   `json:"guard_status,omitempty"`
	RequestedFieldsReady     bool     `json:"requested_fields_ready"`
	BriefReady               bool     `json:"brief_ready,omitempty"`
	FailureCode              string   `json:"failure_code,omitempty"`
	AdapterStatus            string   `json:"adapter_status,omitempty"`
	MissingFields            []string `json:"missing_fields,omitempty"`
	ReviewRequiredFields     []string `json:"review_required_fields,omitempty"`
}

// FinanceReportLookupAnswerReadiness projects a lookup payload into answer gate semantics.
func FinanceReportLookupAnswerReadiness(payload FinanceReportLookupPayload) FinanceReportAnswerReadiness {
	out := FinanceReportAnswerReadiness{
		GuardStatus:         strings.TrimSpace(payload.GuardStatus),
		FailureCode:         strings.TrimSpace(payload.FailureCode),
		AdapterStatus:       strings.TrimSpace(payload.AdapterStatus),
		AllowedSummaryScope: AnswerScopeNone,
	}
	if payload.Metrics != nil {
		metrics := payload.Metrics
		out.GuardStatus = firstLookupReadinessNonEmpty(metrics.GuardStatus, out.GuardStatus)
		out.FailureCode = firstLookupReadinessNonEmpty(metrics.FailureCode, out.FailureCode)
		out.AdapterStatus = firstLookupReadinessNonEmpty(metrics.AdapterStatus, out.AdapterStatus)
		out.RequestedFieldsReady = metrics.RequestedFieldsReady
		out.MissingFields = uniqueLookupReadinessStrings(append(out.MissingFields, metrics.MissingRequestedFields...))
		out.ReviewRequiredFields = uniqueLookupReadinessStrings(append(out.ReviewRequiredFields, metrics.ReviewRequiredFields...))
		out.PresentationRequirements = lookupAnswerPresentationRequirements(metrics.Evidence)
		if lookupReadinessNeedsOperatingCashFlow(payload) && !lookupReadinessKnownValue(metrics.Evidence.OperatingCashFlow) {
			out.RequestedFieldsReady = false
			out.MissingFields = uniqueLookupReadinessStrings(append(out.MissingFields, "operating_cash_flow"))
		}
	}
	if payload.Brief != nil {
		brief := payload.Brief
		out.GuardStatus = firstLookupReadinessNonEmpty(brief.GuardStatus, out.GuardStatus)
		out.FailureCode = firstLookupReadinessNonEmpty(brief.FailureCode, out.FailureCode)
		out.AdapterStatus = firstLookupReadinessNonEmpty(brief.AdapterStatus, out.AdapterStatus)
		out.BriefReady = brief.BriefReady
		if !brief.BriefReady {
			out.MissingFields = uniqueLookupReadinessStrings(append(out.MissingFields, "brief"))
		}
	}
	out.AnswerReady = lookupAnswerReady(out, payload)
	if out.AnswerReady {
		out.AllowedSummaryScope = AnswerScopeRequested
		return out
	}
	out.DegradeReason = lookupAnswerDegradeReason(out, payload)
	out.Degraded = lookupAnswerCanDegrade(out)
	out.AllowedSummaryScope = lookupAnswerAllowedScope(out)
	out.NextRepairHint = lookupAnswerRepairHint(out)
	out.StopRecommended = lookupAnswerStopRecommended(out)
	if out.StopRecommended {
		out.StopReason = out.DegradeReason
	}
	return out
}

func lookupAnswerPresentationRequirements(evidence MetricsEvidence) []string {
	if !NetProfitRequiresAttributableScope(evidence) {
		return nil
	}
	return []string{
		"net_profit 必须按字段证据标注为归属于权益持有人/归母的净利润或净亏损，不得简化为未限定口径的公司净利润",
	}
}

// NetProfitRequiresAttributableScope reports whether field provenance narrows
// net profit to holders/owners of the issuer or parent company.
func NetProfitRequiresAttributableScope(evidence MetricsEvidence) bool {
	profitEvidence, ok := evidence.MetricEvidence["net_profit"]
	if !ok {
		return false
	}
	basis := strings.Join([]string{
		profitEvidence.Evidence,
		profitEvidence.SelectionReason,
	}, "\n")
	return containsAnyFoldString(basis,
		"holder_profit",
		"parent_netprofit",
		"profit attributable",
		"equity holders",
		"owners of the company",
		"归属于",
		"归母",
	)
}

// FinanceReportLookupAnswerReadinessFromPayload returns the payload-provided
// readiness projection when present, otherwise derives it from the payload.
func FinanceReportLookupAnswerReadinessFromPayload(payload FinanceReportLookupPayload) FinanceReportAnswerReadiness {
	if payload.AnswerReady != nil {
		return *payload.AnswerReady
	}
	return FinanceReportLookupAnswerReadiness(payload)
}

func lookupAnswerReady(readiness FinanceReportAnswerReadiness, payload FinanceReportLookupPayload) bool {
	if readiness.GuardStatus != "passed" {
		return false
	}
	if len(readiness.ReviewRequiredFields) > 0 || len(readiness.MissingFields) > 0 {
		return false
	}
	if payload.Brief != nil {
		return readiness.BriefReady
	}
	if payload.Metrics != nil {
		return readiness.RequestedFieldsReady
	}
	return false
}

func lookupAnswerCanDegrade(readiness FinanceReportAnswerReadiness) bool {
	if strings.Contains(strings.ToLower(readiness.FailureCode), "identity") {
		return false
	}
	if readiness.DegradeReason == AnswerDegradeMissingEvidence {
		return false
	}
	return true
}

func lookupAnswerDegradeReason(readiness FinanceReportAnswerReadiness, payload FinanceReportLookupPayload) string {
	if sourceReason := lookupAnswerSourceDownloadDegradeReason(payload); sourceReason != "" {
		return sourceReason
	}
	switch {
	case payload.Metrics == nil && payload.Brief == nil:
		return AnswerDegradeMissingEvidence
	case containsLookupReadinessField(readiness.ReviewRequiredFields, MetricPeriodScopeReviewField):
		return AnswerDegradePeriodScopeReview
	case len(readiness.ReviewRequiredFields) > 0:
		return AnswerDegradeReviewRequired
	case len(readiness.MissingFields) > 0:
		if containsLookupReadinessField(readiness.MissingFields, "brief") {
			return AnswerDegradeBriefNotReady
		}
		return AnswerDegradeMissingRequested
	case readiness.GuardStatus != "" && readiness.GuardStatus != "passed":
		return AnswerDegradeGuardNotPassed
	default:
		return AnswerDegradeMissingEvidence
	}
}

func lookupAnswerAllowedScope(readiness FinanceReportAnswerReadiness) string {
	switch readiness.DegradeReason {
	case AnswerDegradeOfficialSourceMissing, AnswerDegradeSourceDownloadBlocked, AnswerDegradeSourceReturnedNonPDF:
		return AnswerScopeSourceMetadataOnly
	case AnswerDegradePeriodScopeReview:
		return AnswerScopeAvailablePeriodOnly
	case AnswerDegradeMissingRequested, AnswerDegradeBriefNotReady, AnswerDegradeReviewRequired, AnswerDegradeGuardNotPassed:
		return AnswerScopePartialVerifiedMetrics
	case AnswerDegradeMissingEvidence:
		return AnswerScopeNone
	default:
		return AnswerScopeAvailableVerified
	}
}

func lookupAnswerRepairHint(readiness FinanceReportAnswerReadiness) string {
	switch readiness.DegradeReason {
	case AnswerDegradeOfficialSourceMissing, AnswerDegradeSourceDownloadBlocked, AnswerDegradeSourceReturnedNonPDF:
		return "stop_without_more_tools_or_use_configured_alternate_source"
	case AnswerDegradePeriodScopeReview:
		return "fetch_requested_period_source"
	case AnswerDegradeMissingRequested:
		return "fetch_missing_requested_fields"
	case AnswerDegradeBriefNotReady:
		return "fetch_or_parse_full_report_brief"
	case AnswerDegradeReviewRequired:
		return "verify_review_required_fields"
	case AnswerDegradeGuardNotPassed:
		return "repair_guard_evidence_or_degrade_scope"
	case AnswerDegradeMissingEvidence:
		return "run_source_lookup_before_answer"
	default:
		return ""
	}
}

func lookupAnswerStopRecommended(readiness FinanceReportAnswerReadiness) bool {
	switch readiness.DegradeReason {
	case AnswerDegradeOfficialSourceMissing, AnswerDegradeSourceDownloadBlocked, AnswerDegradeSourceReturnedNonPDF:
		return true
	default:
		return false
	}
}

func lookupAnswerSourceDownloadDegradeReason(payload FinanceReportLookupPayload) string {
	text := strings.ToLower(strings.Join(lookupAnswerFailureSignals(payload), "\n"))
	switch {
	case strings.Contains(text, "pdf_download_blocked_by_source") ||
		strings.Contains(text, "download_blocked") ||
		strings.Contains(text, "block_hint=bot_denied") ||
		strings.Contains(text, "denied by bot") ||
		strings.Contains(text, "captcha") ||
		strings.Contains(text, "access denied") ||
		strings.Contains(text, "status 403") ||
		strings.Contains(text, "forbidden"):
		return AnswerDegradeSourceDownloadBlocked
	case strings.Contains(text, "pdf_download_returned_non_pdf") ||
		strings.Contains(text, "returned_non_pdf") ||
		strings.Contains(text, "not a valid pdf") ||
		strings.Contains(text, "content_type=text/html") ||
		strings.Contains(text, "artifact_kind=html") ||
		strings.Contains(text, "direct download invalid"):
		return AnswerDegradeSourceReturnedNonPDF
	case strings.Contains(text, "annual_report_pdf_not_found") ||
		strings.Contains(text, "pdf_link_not_found") ||
		strings.Contains(text, "official_disclosure_pdf_adapter_required") ||
		strings.Contains(text, "official_ir_report_url_missing") ||
		strings.Contains(text, "official_ir_pdf_download_not_configured") ||
		strings.Contains(text, "official_ir_pdf_docparse_not_configured"):
		return AnswerDegradeOfficialSourceMissing
	default:
		return ""
	}
}

func lookupAnswerFailureSignals(payload FinanceReportLookupPayload) []string {
	signals := []string{payload.FailureCode}
	signals = append(signals, payload.Warnings...)
	if payload.Candidates != nil {
		signals = append(signals, payload.Candidates.FailureCode)
		signals = append(signals, payload.Candidates.Warnings...)
	}
	if payload.Metrics != nil {
		signals = append(signals, payload.Metrics.FailureCode)
		signals = append(signals, payload.Metrics.Warnings...)
	}
	if payload.Brief != nil {
		signals = append(signals, payload.Brief.FailureCode)
		signals = append(signals, payload.Brief.Warnings...)
	}
	return signals
}

func containsLookupReadinessField(fields []string, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	for _, field := range fields {
		field = strings.ToLower(strings.TrimSpace(field))
		if field == target || strings.HasSuffix(field, "."+target) {
			return true
		}
	}
	return false
}

func firstLookupReadinessNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func uniqueLookupReadinessStrings(values []string) []string {
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

func lookupReadinessNeedsOperatingCashFlow(payload FinanceReportLookupPayload) bool {
	kind := normalizeAssessmentKindString(lookupReadinessAssessmentKind(payload))
	if kind == financialreportmetrics.AssessmentKindInvestmentRisk {
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

func lookupReadinessKnownValue(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && !strings.EqualFold(value, "unknown")
}

func lookupReadinessAssessmentKind(payload FinanceReportLookupPayload) string {
	if payload.Metrics != nil && strings.TrimSpace(payload.Metrics.AssessmentKind) != "" {
		kind := normalizeAssessmentKindString(payload.Metrics.AssessmentKind)
		if kind != "" && kind != financialreportmetrics.AssessmentKindNone {
			return kind
		}
	}
	if kind, ok := payload.Intent.Assessment["kind"].(string); ok {
		return strings.TrimSpace(kind)
	}
	for _, output := range payload.Intent.RequestedOutputs {
		switch strings.ToLower(strings.TrimSpace(output)) {
		case financialreportmetrics.RequestedOutputInvestmentAssessment, "investment", "investment_judgment":
			return financialreportmetrics.AssessmentKindInvestmentRisk
		}
	}
	return ""
}
