package finance

import "strings"

type FinanceReportLookupSummary struct {
	EntityName           string                   `json:"entity_name,omitempty"`
	StockCode            string                   `json:"stock_code,omitempty"`
	ReportPeriod         string                   `json:"report_period,omitempty"`
	Revenue              string                   `json:"revenue,omitempty"`
	RevenueGrowth        string                   `json:"revenue_growth,omitempty"`
	NetProfit            string                   `json:"net_profit,omitempty"`
	NetProfitGrowth      string                   `json:"net_profit_growth,omitempty"`
	OperatingCashFlow    string                   `json:"operating_cash_flow,omitempty"`
	SourceURL            string                   `json:"source_url,omitempty"`
	AdapterID            string                   `json:"adapter_id,omitempty"`
	GuardStatus          string                   `json:"guard_status,omitempty"`
	RequestedFieldsReady bool                     `json:"requested_fields_ready"`
	AnswerReady          bool                     `json:"answer_ready"`
	Degraded             bool                     `json:"degraded,omitempty"`
	DegradeReason        string                   `json:"degrade_reason,omitempty"`
	AllowedSummaryScope  string                   `json:"allowed_summary_scope,omitempty"`
	NextRepairHint       string                   `json:"next_repair_hint,omitempty"`
	FailureCode          string                   `json:"failure_code,omitempty"`
	Warnings             []string                 `json:"warnings,omitempty"`
	Assessment           *FinanceReportAssessment `json:"assessment_projection,omitempty"`
}

func FinanceReportLookupSummaries(payloads []FinanceReportLookupPayload) []FinanceReportLookupSummary {
	out := make([]FinanceReportLookupSummary, 0, len(payloads))
	for _, payload := range payloads {
		summary, ok := FinanceReportLookupSummaryFromPayload(payload)
		if ok {
			out = append(out, summary)
		}
	}
	return out
}

func FinanceReportLookupSummaryFromPayload(payload FinanceReportLookupPayload) (FinanceReportLookupSummary, bool) {
	if payload.Metrics == nil && payload.Brief == nil && payload.Candidates == nil {
		return FinanceReportLookupSummary{}, false
	}
	summary := FinanceReportLookupSummary{
		AdapterID:   strings.TrimSpace(payload.AdapterID),
		GuardStatus: strings.TrimSpace(payload.GuardStatus),
		FailureCode: strings.TrimSpace(payload.FailureCode),
		Warnings:    uniqueLookupSummaryStrings(payload.Warnings),
	}
	readiness := FinanceReportLookupAnswerReadinessFromPayload(payload)
	summary.AnswerReady = readiness.AnswerReady
	summary.Degraded = readiness.Degraded
	summary.DegradeReason = strings.TrimSpace(readiness.DegradeReason)
	summary.AllowedSummaryScope = strings.TrimSpace(readiness.AllowedSummaryScope)
	summary.NextRepairHint = strings.TrimSpace(readiness.NextRepairHint)
	if payload.Assessment != nil {
		summary.Assessment = payload.Assessment
	} else if assessment, ok := FinanceReportAssessmentFromPayload(payload); ok {
		summary.Assessment = &assessment
	}
	if payload.Candidates != nil {
		summary.EntityName = firstLookupSummaryNonEmpty(summary.EntityName, payload.Candidates.ResolvedCompany)
		summary.StockCode = firstLookupSummaryNonEmpty(summary.StockCode, payload.Candidates.ResolvedCode)
		summary.SourceURL = firstLookupSummaryNonEmpty(summary.SourceURL, payload.Candidates.PrimaryURL)
		summary.AdapterID = firstLookupSummaryNonEmpty(summary.AdapterID, payload.Candidates.AdapterID)
	}
	summary.EntityName = firstLookupSummaryNonEmpty(summary.EntityName, payload.Intent.EntityName)
	summary.StockCode = firstLookupSummaryNonEmpty(summary.StockCode, payload.Intent.StockCode, payload.Intent.Ticker)
	if payload.Metrics != nil {
		evidence := payload.Metrics.Evidence
		summary.EntityName = bestLookupSummaryEntityName(summary.EntityName, evidence.CompanyName, summary.StockCode, evidence.StockCode)
		summary.StockCode = firstLookupSummaryNonEmpty(evidence.StockCode, summary.StockCode)
		summary.ReportPeriod = firstLookupSummaryNonEmpty(evidence.ReportPeriod, summary.ReportPeriod)
		summary.Revenue = evidence.Revenue
		summary.RevenueGrowth = evidence.RevenueGrowth
		summary.NetProfit = evidence.NetProfit
		summary.NetProfitGrowth = evidence.NetProfitGrowth
		summary.OperatingCashFlow = evidence.OperatingCashFlow
		summary.SourceURL = firstLookupSummaryNonEmpty(evidence.OfficialSource, payload.Metrics.FinalURL, summary.SourceURL)
		summary.AdapterID = firstLookupSummaryNonEmpty(payload.Metrics.AdapterID, summary.AdapterID)
		summary.GuardStatus = firstLookupSummaryNonEmpty(payload.Metrics.GuardStatus, summary.GuardStatus)
		summary.RequestedFieldsReady = payload.Metrics.RequestedFieldsReady
		summary.FailureCode = firstLookupSummaryNonEmpty(payload.Metrics.FailureCode, summary.FailureCode)
		summary.Warnings = uniqueLookupSummaryStrings(append(summary.Warnings, payload.Metrics.Warnings...))
	}
	if payload.Brief != nil {
		evidence := payload.Brief.Evidence
		summary.EntityName = bestLookupSummaryEntityName(summary.EntityName, evidence.CompanyName, summary.StockCode, evidence.StockCode)
		summary.StockCode = firstLookupSummaryNonEmpty(evidence.StockCode, summary.StockCode)
		summary.ReportPeriod = firstLookupSummaryNonEmpty(evidence.ReportPeriod, summary.ReportPeriod)
		summary.SourceURL = firstLookupSummaryNonEmpty(evidence.SourceURL, summary.SourceURL)
		summary.AdapterID = firstLookupSummaryNonEmpty(payload.Brief.AdapterID, summary.AdapterID)
		summary.GuardStatus = firstLookupSummaryNonEmpty(payload.Brief.GuardStatus, summary.GuardStatus)
		summary.FailureCode = firstLookupSummaryNonEmpty(payload.Brief.FailureCode, summary.FailureCode)
		summary.Warnings = uniqueLookupSummaryStrings(append(summary.Warnings, payload.Brief.Warnings...))
	}
	return summary, true
}

func bestLookupSummaryEntityName(current string, candidate string, currentCode string, candidateCode string) string {
	current = strings.TrimSpace(current)
	candidate = strings.TrimSpace(candidate)
	if candidate == "" || strings.EqualFold(candidate, "unknown") {
		return current
	}
	if current == "" || strings.EqualFold(current, "unknown") {
		return candidate
	}
	currentCode = normalizeLookupSummaryCode(currentCode)
	candidateCode = normalizeLookupSummaryCode(candidateCode)
	normalizedCandidateName := normalizeLookupSummaryCode(candidate)
	if currentCode != "" && normalizedCandidateName == currentCode {
		return current
	}
	if candidateCode != "" && normalizedCandidateName == candidateCode {
		return current
	}
	return candidate
}

func normalizeLookupSummaryCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	replacer := strings.NewReplacer(".", "", "-", "", "_", "", " ", "")
	return replacer.Replace(value)
}

func firstLookupSummaryNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !strings.EqualFold(value, "unknown") {
			return value
		}
	}
	return ""
}

func uniqueLookupSummaryStrings(values []string) []string {
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
