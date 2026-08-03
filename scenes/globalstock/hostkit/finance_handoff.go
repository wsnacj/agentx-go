package hostkit

import (
	"strings"

	globalcontracts "github.com/wsnacj/agentx-go/scenes/globalstock/contracts"
)

const financeReportLookupTool = "finance_report_lookup"

func isFinanceReportHandoffIntent(intent globalcontracts.InvestigationIntent) bool {
	if intent.TaskKind == globalcontracts.TaskKindFinanceReportHandoff {
		return true
	}
	for _, output := range intent.RequestedOutputs {
		if matchAny(output,
			"report_metrics",
			"financial_report_metrics",
			"financial_report",
			"report_brief",
			"financial_report_brief",
			"business_performance_assessment",
			"investment_assessment",
		) {
			return true
		}
	}
	for _, field := range intent.RequestedFields {
		if isFinanceMetricField(field) {
			return true
		}
	}
	return false
}

func buildFinanceReportHandoffInvestigation(intent globalcontracts.InvestigationIntent, payload globalcontracts.InvestigationPayload) globalcontracts.InvestigationPayload {
	payload.AdapterStatus = globalcontracts.AdapterStatusUnsupported
	payload.FailureCode = globalcontracts.FailureCodeUnsupported
	payload.Readiness = globalcontracts.BuildReadiness(
		payload.AdapterStatus,
		payload.FailureCode,
		false,
		[]string{financeReportLookupTool},
		nil,
	)
	payload.Readiness.NextRepairHint = "call_finance_report_lookup"
	payload.Handoff = &globalcontracts.ToolHandoff{
		TargetPackage: "agentx_finance",
		TargetTool:    financeReportLookupTool,
		Reason:        "financial_report_metrics_are_owned_by_agentx_finance",
		Required:      true,
		Arguments:     financeReportLookupArguments(intent),
		Boundary:      "agentx_global_stock only provides HK/US market-data evidence; report-derived metrics, briefs, and assessments must be verified by agentx_finance.",
	}
	payload.Warnings = append(payload.Warnings, "finance_report_lookup_handoff_required")
	return payload
}

func financeReportLookupArguments(intent globalcontracts.InvestigationIntent) map[string]any {
	args := map[string]any{
		"user_message":      intent.UserMessage,
		"task_kind":         financeReportTaskKind(intent),
		"report_kind":       "auto",
		"entity_name":       intent.EntityName,
		"entity_mentions":   append([]string(nil), intent.EntityMentions...),
		"requested_metrics": financeMetricFields(intent.RequestedFields),
		"requested_outputs": financeRequestedOutputs(intent.RequestedOutputs),
		"period_scope":      financePeriodScope(intent),
		"freshness":         financeFreshness(intent.Freshness),
		"source_hint":       intent.SourceHint,
		"source_policy":     "public_financial_report_sources",
		"original_intent":   firstNonEmpty(intent.OriginalIntent, intent.UserMessage),
		"stop_condition":    firstNonEmpty(intent.StopCondition, "finance_report_lookup_answer_ready_or_explicit_degradation"),
	}
	if intent.StockCode != "" {
		if intent.Market == globalcontracts.MarketUS {
			args["ticker"] = intent.StockCode
		} else {
			args["stock_code"] = intent.StockCode
		}
	}
	return args
}

func financeReportTaskKind(intent globalcontracts.InvestigationIntent) string {
	for _, output := range intent.RequestedOutputs {
		switch strings.ToLower(strings.TrimSpace(output)) {
		case "report_brief", "financial_report_brief":
			return "latest_report_brief"
		case "business_performance_assessment":
			return "business_performance_assessment"
		case "investment_assessment":
			return "investment_assessment"
		case "trend", "report_metrics_trend":
			return "report_metrics_trend"
		}
	}
	return "latest_report_metrics"
}

func financePeriodScope(intent globalcontracts.InvestigationIntent) string {
	for _, output := range intent.RequestedOutputs {
		if matchAny(output, "trend", "report_metrics_trend") {
			return "recent_years"
		}
	}
	return "latest_disclosed_report"
}

func financeRequestedOutputs(outputs []string) []string {
	out := []string{}
	for _, output := range outputs {
		switch strings.ToLower(strings.TrimSpace(output)) {
		case "report_brief", "financial_report_brief":
			out = appendUnique(out, "brief")
		case "business_performance_assessment":
			out = appendUnique(out, "assessment")
		case "investment_assessment":
			out = appendUnique(out, "assessment")
		case "trend", "report_metrics_trend":
			out = appendUnique(out, "trend")
		case "report_metrics", "financial_report_metrics", "financial_report":
			out = appendUnique(out, "metrics")
		}
	}
	if len(out) == 0 {
		out = append(out, "metrics")
	}
	return out
}

func financeMetricFields(fields []string) []string {
	out := []string{}
	for _, field := range fields {
		switch strings.ToLower(strings.TrimSpace(field)) {
		case "revenue", "revenue_growth", "net_profit", "net_profit_growth", "operating_cash_flow":
			out = appendUnique(out, strings.ToLower(strings.TrimSpace(field)))
		}
	}
	return out
}

func isFinanceMetricField(field string) bool {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "revenue", "revenue_growth", "net_profit", "net_profit_growth", "operating_cash_flow":
		return true
	default:
		return false
	}
}

func financeFreshness(freshness globalcontracts.Freshness) map[string]any {
	out := map[string]any{}
	if freshness.Mode != "" {
		out["mode"] = string(freshness.Mode)
	}
	if freshness.RelativeDateHint != "" {
		out["relative_date_hint"] = freshness.RelativeDateHint
	}
	if freshness.AsOf != "" {
		out["as_of"] = freshness.AsOf
	}
	if freshness.RequireLatestTradingDay {
		out["require_latest"] = true
	}
	return out
}

func appendUnique(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
