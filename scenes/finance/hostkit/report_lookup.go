package hostkit

import (
	"context"
	"encoding/json"
	"strings"

	financereports "github.com/wsnacj/agentx-go/scenes/finance"
	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

// MetricsCandidatesHandler builds verified public-web or disclosure candidates.
// Hosts own the source adapters and source policy; hostkit only coordinates the
// standard finance report lookup flow.
type MetricsCandidatesHandler func(context.Context, map[string]any) (financereports.MetricsCandidatesPayload, error)

// MetricsHandler extracts or guards standard finance metric evidence.
type MetricsHandler func(context.Context, map[string]any) (financereports.MetricsToolPayload, error)

// BriefHandler extracts or guards standard finance report brief evidence.
type BriefHandler func(context.Context, map[string]any) (financereports.BriefToolPayload, error)

type FinanceReportLookupHandlers struct {
	Candidates     MetricsCandidatesHandler
	MetricsExtract MetricsHandler
	MetricsGuard   MetricsHandler
	BriefExtract   BriefHandler
	BriefGuard     BriefHandler
}

type FinanceReportLookupConfig struct {
	Source              string
	PackID              string
	WorkflowID          string
	SourcePolicyDefault string

	CaseType func(financereports.FinanceReportLookupIntent) string

	// BriefRequested lets a host refine when expensive full-report brief
	// extraction should run. If unset, hostkit uses only structured tool intent.
	BriefRequested func(map[string]any, financereports.FinanceReportLookupIntent) bool

	Handlers FinanceReportLookupHandlers
}

func BuildFinanceReportLookupHandler(cfg FinanceReportLookupConfig) financereports.ToolPayloadHandler {
	return func(ctx context.Context, params map[string]any) (any, error) {
		return BuildFinanceReportLookupPayload(ctx, cfg, params)
	}
}

func BuildFinanceReportLookupPayload(ctx context.Context, cfg FinanceReportLookupConfig, params map[string]any) (financereports.FinanceReportLookupPayload, error) {
	working := copyParams(params)
	if strings.TrimSpace(stringArg(working["source_policy"])) == "" && strings.TrimSpace(cfg.SourcePolicyDefault) != "" {
		working["source_policy"] = strings.TrimSpace(cfg.SourcePolicyDefault)
	}

	intent := LookupIntentFromParams(working)
	candidates, err := runFinanceReportLookupCandidates(ctx, cfg.Handlers.Candidates, working)
	if err != nil {
		return financereports.FinanceReportLookupPayload{}, err
	}
	financereports.EnsureMetricsCandidatesIdentityResolution(&candidates)
	working = paramsWithLookupCandidates(working, candidates)
	intent = intentWithLookupCandidates(intent, candidates)
	working = paramsWithLookupIntent(working, intent)

	metrics, err := runFinanceReportLookupMetrics(ctx, cfg, working)
	if err != nil {
		return financereports.FinanceReportLookupPayload{}, err
	}

	var brief *financereports.BriefToolPayload
	if financeReportLookupBriefRequested(cfg, working, intent) {
		briefPayload, err := runFinanceReportLookupBrief(ctx, cfg, working)
		if err != nil {
			return financereports.FinanceReportLookupPayload{}, err
		}
		brief = &briefPayload
	}

	payload := financereports.FinanceReportLookupPayload{
		Tool:               financereports.ToolFinanceReportLookup,
		Source:             firstNonEmpty(cfg.Source, "agentx_finance_hostkit"),
		PackID:             firstNonEmpty(cfg.PackID, financialreportmetrics.PackID),
		CaseType:           financeReportLookupCaseType(cfg, intent),
		WorkflowID:         firstNonEmpty(cfg.WorkflowID, financialreportmetrics.DefaultWorkflow),
		AdapterID:          financeReportLookupAdapterID(candidates, metrics, brief),
		AdapterStatus:      financeReportLookupAdapterStatus(candidates, metrics, brief),
		FailureCode:        financeReportLookupFailureCode(candidates, metrics, brief),
		GuardStatus:        financeReportLookupGuardStatus(metrics, brief),
		Intent:             intent,
		Candidates:         &candidates,
		Metrics:            &metrics,
		Brief:              brief,
		Warnings:           financeReportLookupWarnings(candidates, metrics, brief),
		IdentityResolution: financereports.CloneReportIdentityResolution(candidates.IdentityResolution),
	}
	payload.Intent.NeedsSourceVerify = financeReportLookupNeedsSourceVerify(payload)
	readiness := financereports.FinanceReportLookupAnswerReadiness(payload)
	payload.AnswerReady = &readiness
	payload.AnswerContract = financereports.FinanceReportLookupAnswerContract(payload)
	if assessment, ok := financereports.FinanceReportAssessmentFromPayload(payload); ok {
		payload.Assessment = &assessment
	}
	return payload, nil
}

func LookupIntentFromParams(params map[string]any) financereports.FinanceReportLookupIntent {
	intent := financereports.FinanceReportLookupIntent{
		UserMessage:      strings.TrimSpace(stringArg(params["user_message"])),
		TaskKind:         strings.TrimSpace(stringArg(params["task_kind"])),
		ReportKind:       strings.TrimSpace(stringArg(params["report_kind"])),
		EntityName:       strings.TrimSpace(stringArg(params["entity_name"])),
		EntityMentions:   stringListArg(params["entity_mentions"]),
		StockCode:        strings.TrimSpace(stringArg(params["stock_code"])),
		Ticker:           strings.TrimSpace(stringArg(params["ticker"])),
		RequestedMetrics: normalizeStringList(stringListArg(params["requested_metrics"])),
		RequestedOutputs: normalizeLookupRequestedOutputs(stringListArg(params["requested_outputs"])),
		Assessment:       objectArg(params["assessment"]),
		PeriodScope:      periodScopeArg(params),
		Freshness:        freshnessArg(params),
		SourceHint:       strings.TrimSpace(stringArg(params["source_hint"])),
		SourcePolicy:     strings.TrimSpace(stringArg(params["source_policy"])),
		OriginalIntent:   strings.TrimSpace(stringArg(params["original_intent"])),
		StopCondition:    strings.TrimSpace(stringArg(params["stop_condition"])),
	}
	if intent.UserMessage == "" {
		intent.UserMessage = intent.OriginalIntent
	}
	if intent.OriginalIntent == "" {
		intent.OriginalIntent = intent.UserMessage
	}
	intent.RequestedOutputs = lookupRequestedOutputsForIntent(intent)
	intent.Assessment = lookupAssessmentForIntent(intent)
	return intent
}

func runFinanceReportLookupCandidates(ctx context.Context, handler MetricsCandidatesHandler, params map[string]any) (financereports.MetricsCandidatesPayload, error) {
	if handler == nil {
		return financereports.MetricsCandidatesPayload{
			Tool:          financereports.ToolReportMetricsCandidates,
			AdapterStatus: "unsupported",
			FailureCode:   "metrics_candidates_adapter_not_configured",
		}, nil
	}
	return handler(ctx, params)
}

func runFinanceReportLookupMetrics(ctx context.Context, cfg FinanceReportLookupConfig, params map[string]any) (financereports.MetricsToolPayload, error) {
	working := copyParams(params)
	var extracted *financereports.MetricsToolPayload
	if cfg.Handlers.MetricsExtract != nil {
		payload, err := cfg.Handlers.MetricsExtract(ctx, working)
		if err != nil {
			return financereports.MetricsToolPayload{}, err
		}
		extracted = &payload
		working = paramsWithMetricsPayload(working, payload)
	}
	if cfg.Handlers.MetricsGuard != nil {
		return cfg.Handlers.MetricsGuard(ctx, working)
	}
	if extracted != nil {
		extracted.GuardStatus = firstNonEmpty(extracted.GuardStatus, "needs_review")
		extracted.FailureCode = firstNonEmpty(extracted.FailureCode, "metrics_guard_adapter_not_configured")
		return *extracted, nil
	}
	return financereports.MetricsToolPayload{
		Tool:          financereports.ToolReportMetricsGuard,
		AdapterStatus: "unsupported",
		FailureCode:   "metrics_guard_adapter_not_configured",
		GuardStatus:   "needs_review",
	}, nil
}

func runFinanceReportLookupBrief(ctx context.Context, cfg FinanceReportLookupConfig, params map[string]any) (financereports.BriefToolPayload, error) {
	working := copyParams(params)
	var extracted *financereports.BriefToolPayload
	if cfg.Handlers.BriefExtract != nil {
		payload, err := cfg.Handlers.BriefExtract(ctx, working)
		if err != nil {
			return financereports.BriefToolPayload{}, err
		}
		extracted = &payload
		working = paramsWithBriefPayload(working, payload)
	}
	if cfg.Handlers.BriefGuard != nil {
		return cfg.Handlers.BriefGuard(ctx, working)
	}
	if extracted != nil {
		extracted.GuardStatus = firstNonEmpty(extracted.GuardStatus, "needs_review")
		extracted.FailureCode = firstNonEmpty(extracted.FailureCode, "brief_guard_adapter_not_configured")
		return *extracted, nil
	}
	return financereports.BriefToolPayload{
		Tool:          financereports.ToolReportBriefGuard,
		AdapterStatus: "unsupported",
		FailureCode:   "brief_guard_adapter_not_configured",
		GuardStatus:   "needs_review",
	}, nil
}

func paramsWithLookupCandidates(params map[string]any, candidates financereports.MetricsCandidatesPayload) map[string]any {
	out := copyParams(params)
	out["entity_name"] = firstNonEmpty(stringArg(out["entity_name"]), candidates.ResolvedCompany)
	out["stock_code"] = firstNonEmpty(stringArg(out["stock_code"]), candidates.ResolvedCode)
	out["ticker"] = firstNonEmpty(stringArg(out["ticker"]), candidates.ResolvedCode)
	out["source_url"] = firstNonEmpty(stringArg(out["source_url"]), candidates.PrimaryURL)
	out["title"] = firstNonEmpty(stringArg(out["title"]), candidates.PageTitle, candidates.PrimaryKind)
	if len(candidates.ResolvedEntities) > 0 {
		entity := candidates.ResolvedEntities[0]
		out["entity_name"] = firstNonEmpty(stringArg(out["entity_name"]), entity.EntityName)
		out["stock_code"] = firstNonEmpty(stringArg(out["stock_code"]), entity.CodeOrTicker)
		out["ticker"] = firstNonEmpty(stringArg(out["ticker"]), entity.CodeOrTicker)
	}
	if strings.TrimSpace(stringArg(out["source_url"])) == "" && len(candidates.Candidates) > 0 {
		out["source_url"] = strings.TrimSpace(candidates.Candidates[0].URL)
	}
	return out
}

func intentWithLookupCandidates(intent financereports.FinanceReportLookupIntent, candidates financereports.MetricsCandidatesPayload) financereports.FinanceReportLookupIntent {
	intent.EntityName = firstNonEmpty(intent.EntityName, candidates.ResolvedCompany)
	intent.StockCode = firstNonEmpty(intent.StockCode, candidates.ResolvedCode)
	intent.Ticker = firstNonEmpty(intent.Ticker, candidates.ResolvedCode)
	if len(candidates.ResolvedEntities) > 0 {
		entity := candidates.ResolvedEntities[0]
		intent.EntityName = firstNonEmpty(intent.EntityName, entity.EntityName)
		intent.StockCode = firstNonEmpty(intent.StockCode, entity.CodeOrTicker)
		intent.Ticker = firstNonEmpty(intent.Ticker, entity.CodeOrTicker)
	}
	return intent
}

func paramsWithLookupIntent(params map[string]any, intent financereports.FinanceReportLookupIntent) map[string]any {
	out := copyParams(params)
	out["user_message"] = firstNonEmpty(intent.UserMessage, stringArg(out["user_message"]))
	out["task_kind"] = firstNonEmpty(intent.TaskKind, stringArg(out["task_kind"]))
	out["report_kind"] = firstNonEmpty(intent.ReportKind, stringArg(out["report_kind"]))
	out["entity_name"] = firstNonEmpty(intent.EntityName, stringArg(out["entity_name"]))
	out["entity_mentions"] = intent.EntityMentions
	out["stock_code"] = firstNonEmpty(intent.StockCode, stringArg(out["stock_code"]))
	out["ticker"] = firstNonEmpty(intent.Ticker, stringArg(out["ticker"]))
	out["requested_metrics"] = intent.RequestedMetrics
	out["requested_outputs"] = intent.RequestedOutputs
	out["assessment"] = intent.Assessment
	out["period_scope"] = firstNonEmpty(intent.PeriodScope, stringArg(out["period_scope"]))
	out["source_hint"] = firstNonEmpty(intent.SourceHint, stringArg(out["source_hint"]))
	out["source_policy"] = firstNonEmpty(intent.SourcePolicy, stringArg(out["source_policy"]))
	out["original_intent"] = firstNonEmpty(intent.OriginalIntent, stringArg(out["original_intent"]))
	out["stop_condition"] = firstNonEmpty(intent.StopCondition, stringArg(out["stop_condition"]))
	if len(intent.Freshness) > 0 {
		out["freshness"] = intent.Freshness
	}
	return out
}

func paramsWithMetricsPayload(params map[string]any, payload financereports.MetricsToolPayload) map[string]any {
	out := copyParams(params)
	out["adapter_id"] = firstNonEmpty(stringArg(out["adapter_id"]), payload.AdapterID)
	out["source_url"] = firstNonEmpty(stringArg(out["source_url"]), payload.FinalURL, payload.Evidence.OfficialSource)
	out["title"] = firstNonEmpty(stringArg(out["title"]), payload.PageTitle, payload.Evidence.PageTitle)
	out["company_name"] = firstNonEmpty(stringArg(out["company_name"]), payload.Evidence.CompanyName)
	out["stock_code"] = firstNonEmpty(stringArg(out["stock_code"]), payload.Evidence.StockCode)
	out["selection_reason"] = firstNonEmpty(stringArg(out["selection_reason"]), payload.Evidence.SelectionReason)
	out["official_source"] = firstNonEmpty(stringArg(out["official_source"]), payload.Evidence.OfficialSource)
	out["report_period"] = firstNonEmpty(stringArg(out["report_period"]), payload.Evidence.ReportPeriod)
	out["revenue"] = firstNonEmpty(stringArg(out["revenue"]), payload.Evidence.Revenue)
	out["revenue_growth"] = firstNonEmpty(stringArg(out["revenue_growth"]), payload.Evidence.RevenueGrowth)
	out["net_profit"] = firstNonEmpty(stringArg(out["net_profit"]), payload.Evidence.NetProfit)
	out["net_profit_growth"] = firstNonEmpty(stringArg(out["net_profit_growth"]), payload.Evidence.NetProfitGrowth)
	out["operating_cash_flow"] = firstNonEmpty(stringArg(out["operating_cash_flow"]), payload.Evidence.OperatingCashFlow)
	if len(payload.Evidence.MetricEvidence) > 0 {
		out["metric_evidence"] = payload.Evidence.MetricEvidence
	}
	if len(payload.Evidence.TrendSeries) > 0 {
		out["trend_series"] = payload.Evidence.TrendSeries
	}
	return out
}

func paramsWithBriefPayload(params map[string]any, payload financereports.BriefToolPayload) map[string]any {
	out := copyParams(params)
	out["evidence"] = payload.Evidence
	return out
}

func financeReportLookupCaseType(cfg FinanceReportLookupConfig, intent financereports.FinanceReportLookupIntent) string {
	if cfg.CaseType != nil {
		if caseType := strings.TrimSpace(cfg.CaseType(intent)); caseType != "" {
			return caseType
		}
	}
	if financialreportmetrics.MetricsTrendRequested(intent.UserMessage, intent.PeriodScope) {
		return financialreportmetrics.CaseTypeTrend
	}
	return financialreportmetrics.CaseTypeLatest
}

func financeReportLookupBriefRequested(cfg FinanceReportLookupConfig, params map[string]any, intent financereports.FinanceReportLookupIntent) bool {
	if cfg.BriefRequested != nil {
		return cfg.BriefRequested(params, intent)
	}
	for _, output := range intent.RequestedOutputs {
		switch normalizeLookupOutput(output) {
		case "brief", "report_brief", "summary":
			return true
		}
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(intent.TaskKind)), "brief")
}

func lookupRequestedOutputsForIntent(intent financereports.FinanceReportLookupIntent) []string {
	outputs := normalizeLookupRequestedOutputs(intent.RequestedOutputs)
	outputs = appendLookupOutputIfMissing(outputs, financialreportmetrics.RequestedOutputMetrics)
	switch lookupAssessmentKind(intent) {
	case financialreportmetrics.AssessmentKindBusinessPerformance:
		outputs = appendLookupOutputIfMissing(outputs, financialreportmetrics.RequestedOutputPerformanceAssessment)
	case financialreportmetrics.AssessmentKindInvestmentRisk:
		outputs = appendLookupOutputIfMissing(outputs, financialreportmetrics.RequestedOutputInvestmentAssessment)
	}
	return outputs
}

func lookupAssessmentForIntent(intent financereports.FinanceReportLookupIntent) map[string]any {
	assessment := copyObject(intent.Assessment)
	kind := lookupAssessmentKind(intent)
	if kind == financialreportmetrics.AssessmentKindNone {
		if len(assessment) == 0 {
			return nil
		}
		return assessment
	}
	if assessment == nil {
		assessment = map[string]any{}
	}
	assessment["kind"] = kind
	if strings.TrimSpace(stringArg(assessment["scope"])) == "" {
		assessment["scope"] = financialreportmetrics.AssessmentScopeMetricsOnly
	}
	if kind == financialreportmetrics.AssessmentKindInvestmentRisk {
		assessment["requires_valuation"] = true
	}
	return assessment
}

func lookupAssessmentKind(intent financereports.FinanceReportLookupIntent) string {
	if assessment := intent.Assessment; len(assessment) > 0 {
		if kind := normalizeLookupAssessmentKind(stringArg(assessment["kind"])); kind != financialreportmetrics.AssessmentKindNone {
			return kind
		}
	}
	switch strings.ToLower(strings.TrimSpace(intent.TaskKind)) {
	case "business_performance_assessment":
		return financialreportmetrics.AssessmentKindBusinessPerformance
	case "investment_assessment":
		return financialreportmetrics.AssessmentKindInvestmentRisk
	}
	for _, output := range intent.RequestedOutputs {
		switch normalizeLookupOutput(output) {
		case financialreportmetrics.RequestedOutputPerformanceAssessment:
			return financialreportmetrics.AssessmentKindBusinessPerformance
		case financialreportmetrics.RequestedOutputInvestmentAssessment:
			return financialreportmetrics.AssessmentKindInvestmentRisk
		}
	}
	return financialreportmetrics.AssessmentKindNone
}

func normalizeLookupAssessmentKind(value string) string {
	value = normalizeLookupOutput(value)
	switch value {
	case "", "none", "no_assessment":
		return financialreportmetrics.AssessmentKindNone
	case financialreportmetrics.AssessmentKindBusinessPerformance, "performance_assessment", "performance", "financial_performance":
		return financialreportmetrics.AssessmentKindBusinessPerformance
	case financialreportmetrics.AssessmentKindInvestmentRisk, "investment", "investment_judgment":
		return financialreportmetrics.AssessmentKindInvestmentRisk
	default:
		return financialreportmetrics.AssessmentKindNone
	}
}

func normalizeLookupRequestedOutputs(values []string) []string {
	out := []string{}
	for _, value := range values {
		out = appendLookupOutputIfMissing(out, value)
	}
	return out
}

func appendLookupOutputIfMissing(values []string, value string) []string {
	value = normalizeLookupOutput(value)
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

func normalizeLookupOutput(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	return value
}

func freshnessArg(params map[string]any) map[string]any {
	out := map[string]any{}
	if object := objectArg(params["freshness"]); len(object) > 0 {
		for key, value := range object {
			if strings.TrimSpace(key) != "" && value != nil {
				out[key] = value
			}
		}
	}
	addStringIfPresent(out, "mode", params["freshness_mode"])
	addStringIfPresent(out, "relative_date_hint", params["relative_date_hint"])
	addStringIfPresent(out, "published_after", params["published_after"])
	addStringIfPresent(out, "published_before", params["published_before"])
	if len(out) == 0 {
		return nil
	}
	return out
}

func periodScopeArg(params map[string]any) string {
	if text := strings.TrimSpace(stringArg(params["period_scope_mode"])); text != "" {
		return text
	}
	raw := params["period_scope"]
	if text := strings.TrimSpace(stringArg(raw)); text != "" && !strings.HasPrefix(text, "map[") {
		return text
	}
	if object := objectArg(raw); len(object) > 0 {
		return firstNonEmpty(
			stringArg(object["mode"]),
			stringArg(object["scope"]),
			stringArg(object["period"]),
			stringArg(object["year"]),
		)
	}
	return ""
}

func addStringIfPresent(out map[string]any, key string, value any) {
	if text := strings.TrimSpace(stringArg(value)); text != "" {
		out[key] = text
	}
}

func copyParams(params map[string]any) map[string]any {
	out := make(map[string]any, len(params))
	for key, value := range params {
		out[key] = value
	}
	return out
}

func copyObject(params map[string]any) map[string]any {
	if len(params) == 0 {
		return nil
	}
	out := make(map[string]any, len(params))
	for key, value := range params {
		out[key] = value
	}
	return out
}

func objectArg(value any) map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		return typed
	case map[string]string:
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			out[key] = value
		}
		return out
	default:
		return nil
	}
}

func stringArg(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return ""
	}
}

func stringListArg(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string{}, typed...)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := strings.TrimSpace(stringArg(item)); text != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func normalizeStringList(values []string) []string {
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func financeReportLookupAdapterID(candidates financereports.MetricsCandidatesPayload, metrics financereports.MetricsToolPayload, brief *financereports.BriefToolPayload) string {
	if brief != nil && strings.TrimSpace(brief.AdapterID) != "" && strings.TrimSpace(metrics.AdapterID) == "" {
		return brief.AdapterID
	}
	return firstNonEmpty(metrics.AdapterID, candidates.AdapterID, financeReportLookupBriefAdapterID(brief))
}

func financeReportLookupBriefAdapterID(brief *financereports.BriefToolPayload) string {
	if brief == nil {
		return ""
	}
	return brief.AdapterID
}

func financeReportLookupAdapterStatus(candidates financereports.MetricsCandidatesPayload, metrics financereports.MetricsToolPayload, brief *financereports.BriefToolPayload) string {
	return firstNonEmpty(metrics.AdapterStatus, financeReportLookupBriefAdapterStatus(brief), candidates.AdapterStatus, "unsupported")
}

func financeReportLookupBriefAdapterStatus(brief *financereports.BriefToolPayload) string {
	if brief == nil {
		return ""
	}
	return brief.AdapterStatus
}

func financeReportLookupFailureCode(candidates financereports.MetricsCandidatesPayload, metrics financereports.MetricsToolPayload, brief *financereports.BriefToolPayload) string {
	return firstNonEmpty(metrics.FailureCode, financeReportLookupBriefFailureCode(brief), candidates.FailureCode)
}

func financeReportLookupBriefFailureCode(brief *financereports.BriefToolPayload) string {
	if brief == nil {
		return ""
	}
	return brief.FailureCode
}

func financeReportLookupGuardStatus(metrics financereports.MetricsToolPayload, brief *financereports.BriefToolPayload) string {
	if strings.TrimSpace(metrics.GuardStatus) != "" {
		return metrics.GuardStatus
	}
	if brief == nil {
		return ""
	}
	return brief.GuardStatus
}

func financeReportLookupWarnings(candidates financereports.MetricsCandidatesPayload, metrics financereports.MetricsToolPayload, brief *financereports.BriefToolPayload) []string {
	out := []string{}
	out = append(out, candidates.Warnings...)
	out = append(out, metrics.Warnings...)
	if brief != nil {
		out = append(out, brief.Warnings...)
	}
	if metrics.GuardStatus != "" && metrics.GuardStatus != "passed" {
		out = append(out, "finance_report_lookup_metrics_guard_not_passed:"+metrics.GuardStatus)
	}
	if brief != nil && brief.GuardStatus != "" && brief.GuardStatus != "passed" {
		out = append(out, "finance_report_lookup_brief_guard_not_passed:"+brief.GuardStatus)
	}
	return uniqueStrings(out)
}

func financeReportLookupNeedsSourceVerify(payload financereports.FinanceReportLookupPayload) bool {
	if payload.Metrics != nil && payload.Metrics.GuardStatus != "" && payload.Metrics.GuardStatus != "passed" {
		return true
	}
	if payload.Brief != nil && payload.Brief.GuardStatus != "" && payload.Brief.GuardStatus != "passed" {
		return true
	}
	return false
}

func uniqueStrings(values []string) []string {
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

// DecodeFinanceReportLookupPayload is a small host helper for reply renderers
// that consume the serialized finance_report_lookup tool output.
func DecodeFinanceReportLookupPayload(raw string) (financereports.FinanceReportLookupPayload, bool) {
	var payload financereports.FinanceReportLookupPayload
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &payload); err != nil {
		return financereports.FinanceReportLookupPayload{}, false
	}
	return payload, true
}
