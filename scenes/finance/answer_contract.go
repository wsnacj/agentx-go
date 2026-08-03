package finance

import (
	"strings"
)

// FinanceReportLookupAnswerContract projects a lookup payload into a
// model-facing final-answer handoff. It only emits a contract when the finance
// package has enough evidence to define a safe answer boundary.
func FinanceReportLookupAnswerContract(payload FinanceReportLookupPayload) *FinanceReportAnswerContract {
	readiness := FinanceReportLookupAnswerReadinessFromPayload(payload)
	if !readiness.StopRecommended {
		return nil
	}
	switch readiness.DegradeReason {
	case AnswerDegradeOfficialSourceMissing, AnswerDegradeSourceDownloadBlocked, AnswerDegradeSourceReturnedNonPDF:
		return sourceDownloadAnswerContract(payload, readiness)
	default:
		return nil
	}
}

func sourceDownloadAnswerContract(payload FinanceReportLookupPayload, readiness FinanceReportAnswerReadiness) *FinanceReportAnswerContract {
	summary, _ := FinanceReportLookupSummaryFromPayload(payload)
	entity := firstNonEmptyString(
		payload.Intent.EntityName,
		summary.EntityName,
		payload.Intent.StockCode,
		payload.Intent.Ticker,
		"该主体",
	)
	code := firstNonEmptyString(summary.StockCode, payload.Intent.StockCode, payload.Intent.Ticker)
	sourceURL := firstNonEmptyString(summary.SourceURL, metricsOfficialSource(payload), candidatesPrimaryURL(payload), "unknown")
	reportPeriod := firstNonEmptyString(summary.ReportPeriod, metricsReportPeriod(payload), "unknown")
	fields := append([]string{}, readiness.MissingFields...)
	if len(fields) == 0 {
		fields = append([]string{}, payload.Intent.RequestedMetrics...)
	}
	fieldText := strings.Join(cleanAnswerContractStrings(fields), "、")
	if fieldText == "" {
		fieldText = "原请求字段"
	}
	reasonText := sourceDownloadReasonText(readiness.DegradeReason)

	draftLines := []string{
		"当前不能把这次查询包装成已核实的完整财报结果。",
		"已定位到公开披露来源，但" + reasonText + "，因此无法从该报告 artifact 中核验 " + fieldText + "。",
		"可确认的来源元数据：主体=" + entity + "，代码=" + firstNonEmptyString(code, "unknown") + "，报告期=" + reportPeriod + "，来源=" + sourceURL + "。",
		"回答边界：只能说明来源已定位但下载/解析受阻；如需继续提取字段，需要宿主配置可用的替代官方来源、提供可下载 PDF，或提供本地报告文件。",
	}
	return &FinanceReportAnswerContract{
		FinalAnswerRecommended: true,
		Reason:                 readiness.StopReason,
		AllowedSummaryScope:    readiness.AllowedSummaryScope,
		DoNotRetryTools:        []string{"pdf", "browser", "search", "open_page", "find_in_page"},
		FinalAnswerDraft:       strings.Join(draftLines, "\n"),
	}
}

func sourceDownloadReasonText(reason string) string {
	switch reason {
	case AnswerDegradeOfficialSourceMissing:
		return "当前已配置/请求的官方来源没有提供可直接解析的报告文件或解析能力未配置"
	case AnswerDegradeSourceDownloadBlocked:
		return "当前环境访问该报告文件被来源站点拦截"
	case AnswerDegradeSourceReturnedNonPDF:
		return "当前环境下载到的内容不是有效 PDF"
	default:
		return "当前环境无法取得可解析报告文件"
	}
}

func metricsOfficialSource(payload FinanceReportLookupPayload) string {
	if payload.Metrics == nil {
		return ""
	}
	return strings.TrimSpace(payload.Metrics.Evidence.OfficialSource)
}

func metricsReportPeriod(payload FinanceReportLookupPayload) string {
	if payload.Metrics == nil {
		return ""
	}
	return strings.TrimSpace(payload.Metrics.Evidence.ReportPeriod)
}

func candidatesPrimaryURL(payload FinanceReportLookupPayload) string {
	if payload.Candidates == nil {
		return ""
	}
	return strings.TrimSpace(payload.Candidates.PrimaryURL)
}

func cleanAnswerContractStrings(values []string) []string {
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
