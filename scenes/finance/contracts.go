package finance

import (
	financialreportbrief "github.com/wsnacj/agentx-go/scenes/finance/brief"
	financialreportmetrics "github.com/wsnacj/agentx-go/scenes/finance/metrics"
)

// MetricsToolPayload is the adapter-neutral result envelope returned by
// report_metrics_extract and report_metrics_guard.
type MetricsToolPayload struct {
	Tool                        string                                          `json:"tool"`
	Source                      string                                          `json:"source"`
	PackID                      string                                          `json:"pack_id,omitempty"`
	CaseType                    string                                          `json:"case_type,omitempty"`
	WorkflowID                  string                                          `json:"workflow_id,omitempty"`
	AdapterID                   string                                          `json:"adapter_id"`
	AdapterStatus               string                                          `json:"adapter_status"`
	SourcePolicy                string                                          `json:"source_policy"`
	FailureCode                 string                                          `json:"failure_code"`
	PageID                      string                                          `json:"page_id,omitempty"`
	FinalURL                    string                                          `json:"final_url,omitempty"`
	PageTitle                   string                                          `json:"page_title,omitempty"`
	RequestedFields             []string                                        `json:"requested_fields,omitempty"`
	RequestedOutputs            []string                                        `json:"requested_outputs,omitempty"`
	AssessmentKind              string                                          `json:"assessment_kind,omitempty"`
	AssessmentScope             string                                          `json:"assessment_scope,omitempty"`
	AssessmentRequiresValuation bool                                            `json:"assessment_requires_valuation,omitempty"`
	RequestedFieldsReady        bool                                            `json:"requested_fields_ready"`
	MissingRequestedFields      []string                                        `json:"missing_requested_fields"`
	ReviewRequiredFields        []string                                        `json:"review_required_fields"`
	Warnings                    []string                                        `json:"warnings,omitempty"`
	GuardStatus                 string                                          `json:"guard_status,omitempty"`
	Evidence                    MetricsEvidence                                 `json:"evidence"`
	Evaluation                  *financialreportmetrics.LatestMetricsEvaluation `json:"evaluation,omitempty"`
}

// MetricsEvidence is the source-neutral financial metric evidence projected by
// host adapters before the guard/evaluator layer decides whether it is usable.
type MetricsEvidence struct {
	CompanyName       string                                                              `json:"company_name,omitempty"`
	StockCode         string                                                              `json:"stock_code,omitempty"`
	SelectionReason   string                                                              `json:"selection_reason,omitempty"`
	OfficialSource    string                                                              `json:"official_source,omitempty"`
	ReportPeriod      string                                                              `json:"report_period,omitempty"`
	Revenue           string                                                              `json:"revenue,omitempty"`
	RevenueGrowth     string                                                              `json:"revenue_growth,omitempty"`
	NetProfit         string                                                              `json:"net_profit,omitempty"`
	NetProfitGrowth   string                                                              `json:"net_profit_growth,omitempty"`
	OperatingCashFlow string                                                              `json:"operating_cash_flow,omitempty"`
	PageTitle         string                                                              `json:"page_title,omitempty"`
	MetricEvidence    map[string]financialreportmetrics.ReportDocumentMetricFieldEvidence `json:"metric_evidence,omitempty"`
	TrendSeries       []MetricsTrendSeriesPoint                                           `json:"trend_series"`
}

type MetricsTrendSeriesPoint struct {
	Period            string   `json:"period,omitempty"`
	Revenue           string   `json:"revenue,omitempty"`
	RevenueGrowth     string   `json:"revenue_growth,omitempty"`
	NetProfit         string   `json:"net_profit,omitempty"`
	NetProfitGrowth   string   `json:"net_profit_growth,omitempty"`
	OperatingCashFlow string   `json:"operating_cash_flow,omitempty"`
	Source            string   `json:"source,omitempty"`
	Confidence        float64  `json:"confidence,omitempty"`
	Calculation       string   `json:"calculation,omitempty"`
	ReviewRequired    bool     `json:"review_required,omitempty"`
	Warnings          []string `json:"warnings,omitempty"`
}

// MetricsCandidatesPayload is the adapter-neutral candidate-generation result
// envelope returned by report_metrics_candidates.
type MetricsCandidatesPayload struct {
	Tool                        string                    `json:"tool"`
	Source                      string                    `json:"source"`
	PackID                      string                    `json:"pack_id,omitempty"`
	CaseType                    string                    `json:"case_type,omitempty"`
	WorkflowID                  string                    `json:"workflow_id,omitempty"`
	AdapterID                   string                    `json:"adapter_id"`
	AdapterStatus               string                    `json:"adapter_status"`
	SourcePolicy                string                    `json:"source_policy"`
	FailureCode                 string                    `json:"failure_code"`
	UserMessage                 string                    `json:"user_message,omitempty"`
	EntityName                  string                    `json:"entity_name,omitempty"`
	EntityMentions              []string                  `json:"entity_mentions,omitempty"`
	RequestedMetrics            []string                  `json:"requested_metrics,omitempty"`
	RequestedOutputs            []string                  `json:"requested_outputs,omitempty"`
	AssessmentKind              string                    `json:"assessment_kind,omitempty"`
	AssessmentScope             string                    `json:"assessment_scope,omitempty"`
	AssessmentRequiresValuation bool                      `json:"assessment_requires_valuation,omitempty"`
	PeriodScope                 string                    `json:"period_scope,omitempty"`
	SourceHint                  string                    `json:"source_hint,omitempty"`
	PageID                      string                    `json:"page_id,omitempty"`
	PageTitle                   string                    `json:"page_title,omitempty"`
	SourceURL                   string                    `json:"source_url,omitempty"`
	ResolvedCode                string                    `json:"resolved_code,omitempty"`
	ResolvedMarket              string                    `json:"resolved_market,omitempty"`
	ResolvedCompany             string                    `json:"resolved_company,omitempty"`
	ResolvedEntities            []ResolvedEntityCandidate `json:"resolved_entities,omitempty"`
	PrimaryURL                  string                    `json:"primary_url,omitempty"`
	PrimaryKind                 string                    `json:"primary_source_kind,omitempty"`
	Candidates                  []MetricsCandidate        `json:"candidates,omitempty"`
	Warnings                    []string                  `json:"warnings,omitempty"`
	IdentityResolution          *ReportIdentityResolution `json:"identity_resolution,omitempty"`
}

// ReportIdentityResolution records how a report lookup subject was resolved.
// It is diagnostic-only; report metrics and briefs still require guarded source evidence.
type ReportIdentityResolution struct {
	InputTerm         string                              `json:"input_term,omitempty"`
	PreferredMarket   string                              `json:"preferred_market,omitempty"`
	Strategy          string                              `json:"strategy,omitempty"`
	SelectedReason    string                              `json:"selected_reason,omitempty"`
	SelectedCandidate *ReportIdentityResolutionCandidate  `json:"selected_candidate,omitempty"`
	QueryVariants     []ReportIdentityResolutionQuery     `json:"query_variants,omitempty"`
	Candidates        []ReportIdentityResolutionCandidate `json:"candidates,omitempty"`
	Warnings          []string                            `json:"warnings,omitempty"`
}

// ReportIdentityResolutionQuery records one candidate-generation attempt.
type ReportIdentityResolutionQuery struct {
	Term           string `json:"term,omitempty"`
	Reason         string `json:"reason,omitempty"`
	Priority       int    `json:"priority,omitempty"`
	Provider       string `json:"provider,omitempty"`
	Status         string `json:"status,omitempty"`
	CandidateCount int    `json:"candidate_count,omitempty"`
	Message        string `json:"message,omitempty"`
}

// ReportIdentityResolutionCandidate records one source-backed issuer candidate.
type ReportIdentityResolutionCandidate struct {
	EntityName     string  `json:"entity_name,omitempty"`
	CodeOrTicker   string  `json:"code_or_ticker,omitempty"`
	Market         string  `json:"market,omitempty"`
	Source         string  `json:"source,omitempty"`
	EvidenceURL    string  `json:"evidence_url,omitempty"`
	Confidence     float64 `json:"confidence,omitempty"`
	MatchReason    string  `json:"match_reason,omitempty"`
	MismatchReason string  `json:"mismatch_reason,omitempty"`
	Selected       bool    `json:"selected,omitempty"`
	SelectedReason string  `json:"selected_reason,omitempty"`
}

type ResolvedEntityCandidate struct {
	EntityName     string                        `json:"entity_name,omitempty"`
	CodeOrTicker   string                        `json:"code_or_ticker,omitempty"`
	Market         string                        `json:"market,omitempty"`
	Source         string                        `json:"source,omitempty"`
	EvidenceURL    string                        `json:"evidence_url,omitempty"`
	Confidence     float64                       `json:"confidence"`
	MatchReason    string                        `json:"match_reason,omitempty"`
	MismatchReason string                        `json:"mismatch_reason,omitempty"`
	Provenance     []ResolvedEntityProvenanceRef `json:"provenance,omitempty"`
}

type ResolvedEntityProvenanceRef struct {
	Source string `json:"source,omitempty"`
	URL    string `json:"url,omitempty"`
	Field  string `json:"field,omitempty"`
	Value  string `json:"value,omitempty"`
}

type MetricsCandidate struct {
	URL                 string `json:"url"`
	SourceKind          string `json:"source_kind"`
	StockCode           string `json:"stock_code"`
	MarketCode          string `json:"market_code,omitempty"`
	Reason              string `json:"reason"`
	PreferredNextAction string `json:"preferred_next_action"`
}

// FinanceReportLookupPayload is the high-level finance report result envelope
// returned by finance_report_lookup. It lets the main agent loop provide a
// structured task frame in one tool call while host adapters still own source
// discovery, extraction, fallback, and guard policy.
type FinanceReportLookupPayload struct {
	Tool               string                        `json:"tool"`
	Source             string                        `json:"source"`
	PackID             string                        `json:"pack_id,omitempty"`
	CaseType           string                        `json:"case_type,omitempty"`
	WorkflowID         string                        `json:"workflow_id,omitempty"`
	AdapterID          string                        `json:"adapter_id,omitempty"`
	AdapterStatus      string                        `json:"adapter_status"`
	FailureCode        string                        `json:"failure_code,omitempty"`
	GuardStatus        string                        `json:"guard_status,omitempty"`
	Intent             FinanceReportLookupIntent     `json:"intent"`
	Candidates         *MetricsCandidatesPayload     `json:"candidates,omitempty"`
	Metrics            *MetricsToolPayload           `json:"metrics,omitempty"`
	Brief              *BriefToolPayload             `json:"brief,omitempty"`
	AnswerReady        *FinanceReportAnswerReadiness `json:"answer_readiness,omitempty"`
	AnswerContract     *FinanceReportAnswerContract  `json:"answer_contract,omitempty"`
	Assessment         *FinanceReportAssessment      `json:"assessment_projection,omitempty"`
	Warnings           []string                      `json:"warnings,omitempty"`
	IdentityResolution *ReportIdentityResolution     `json:"identity_resolution,omitempty"`
}

// FinanceReportAnswerContract is a model-facing answer handoff contract.
// Runtimes may honor it only when it is emitted as a trusted top-level tool
// result contract; it must not be inferred from nested page text or used to
// encode domain-specific routing policy.
type FinanceReportAnswerContract struct {
	FinalAnswerRecommended bool     `json:"final_answer_recommended"`
	Reason                 string   `json:"reason,omitempty"`
	AllowedSummaryScope    string   `json:"allowed_summary_scope,omitempty"`
	DoNotRetryTools        []string `json:"do_not_retry_tools,omitempty"`
	FinalAnswerDraft       string   `json:"final_answer_draft,omitempty"`
}

type FinanceReportLookupIntent struct {
	UserMessage       string         `json:"user_message,omitempty"`
	TaskKind          string         `json:"task_kind,omitempty"`
	ReportKind        string         `json:"report_kind,omitempty"`
	EntityName        string         `json:"entity_name,omitempty"`
	EntityMentions    []string       `json:"entity_mentions,omitempty"`
	StockCode         string         `json:"stock_code,omitempty"`
	Ticker            string         `json:"ticker,omitempty"`
	RequestedMetrics  []string       `json:"requested_metrics,omitempty"`
	RequestedOutputs  []string       `json:"requested_outputs,omitempty"`
	Assessment        map[string]any `json:"assessment,omitempty"`
	PeriodScope       string         `json:"period_scope,omitempty"`
	Freshness         map[string]any `json:"freshness,omitempty"`
	SourceHint        string         `json:"source_hint,omitempty"`
	SourcePolicy      string         `json:"source_policy,omitempty"`
	OriginalIntent    string         `json:"original_intent,omitempty"`
	StopCondition     string         `json:"stop_condition,omitempty"`
	NeedsSourceVerify bool           `json:"needs_source_verify,omitempty"`
}

// BriefToolPayload is the adapter-neutral result envelope returned by
// report_brief_extract and report_brief_guard.
type BriefToolPayload struct {
	Tool          string                                `json:"tool"`
	Source        string                                `json:"source"`
	PackID        string                                `json:"pack_id,omitempty"`
	CaseType      string                                `json:"case_type,omitempty"`
	WorkflowID    string                                `json:"workflow_id,omitempty"`
	AdapterID     string                                `json:"adapter_id"`
	AdapterStatus string                                `json:"adapter_status"`
	SourcePolicy  string                                `json:"source_policy"`
	FailureCode   string                                `json:"failure_code,omitempty"`
	FailureStage  string                                `json:"failure_stage,omitempty"`
	GuardStatus   string                                `json:"guard_status,omitempty"`
	BriefReady    bool                                  `json:"brief_ready"`
	Evidence      financialreportbrief.BriefEvidence    `json:"evidence"`
	Evaluation    *financialreportbrief.BriefEvaluation `json:"evaluation,omitempty"`
	Warnings      []string                              `json:"warnings,omitempty"`
}
