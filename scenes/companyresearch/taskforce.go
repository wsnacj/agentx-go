package companyresearch

import "strings"

type CompanyResearchTaskRole string

const (
	CompanyResearchRoleSubjectResolver CompanyResearchTaskRole = "subject_resolver"
	CompanyResearchRoleFinanceAnalyst  CompanyResearchTaskRole = "finance_analyst"
	CompanyResearchRoleMarketAnalyst   CompanyResearchTaskRole = "market_analyst"
	CompanyResearchRoleNewsAnalyst     CompanyResearchTaskRole = "news_analyst"
	CompanyResearchRoleRiskReviewer    CompanyResearchTaskRole = "risk_reviewer"
	CompanyResearchRoleEvidenceGuard   CompanyResearchTaskRole = "evidence_guard"
	CompanyResearchRoleSynthesisEditor CompanyResearchTaskRole = "synthesis_editor"
)

type CompanyResearchTaskStatus string

const (
	CompanyResearchTaskStatusPending  CompanyResearchTaskStatus = "pending"
	CompanyResearchTaskStatusReady    CompanyResearchTaskStatus = "ready"
	CompanyResearchTaskStatusDegraded CompanyResearchTaskStatus = "degraded"
	CompanyResearchTaskStatusSkipped  CompanyResearchTaskStatus = "skipped"
	CompanyResearchTaskStatusFailed   CompanyResearchTaskStatus = "failed"
)

type CompanyResearchTaskPlanInput struct {
	CaseType          string
	WorkflowID        string
	Intent            CompanyResearchIntent
	SubjectResolution *SubjectResolution
}

type CompanyResearchTaskPlan struct {
	SchemaVersion string                    `json:"schema_version,omitempty"`
	PlanID        string                    `json:"plan_id,omitempty"`
	CaseType      string                    `json:"case_type,omitempty"`
	WorkflowID    string                    `json:"workflow_id,omitempty"`
	TaskKind      string                    `json:"task_kind,omitempty"`
	UserMessage   string                    `json:"user_message,omitempty"`
	Subject       CompanyResearchSubjectRef `json:"subject,omitempty"`
	Dimensions    []string                  `json:"dimensions,omitempty"`
	Outputs       []string                  `json:"outputs,omitempty"`
	Tasks         []CompanyResearchTaskSpec `json:"tasks,omitempty"`
	Warnings      []string                  `json:"warnings,omitempty"`
}

type CompanyResearchSubjectRef struct {
	InputTerm       string   `json:"input_term,omitempty"`
	CanonicalName   string   `json:"canonical_name,omitempty"`
	EntityMentions  []string `json:"entity_mentions,omitempty"`
	MarketHint      string   `json:"market_hint,omitempty"`
	StockCode       string   `json:"stock_code,omitempty"`
	Ticker          string   `json:"ticker,omitempty"`
	Source          string   `json:"source,omitempty"`
	EvidenceURL     string   `json:"evidence_url,omitempty"`
	Verified        bool     `json:"verified,omitempty"`
	Confidence      float64  `json:"confidence,omitempty"`
	FailureCode     string   `json:"failure_code,omitempty"`
	ResolutionState string   `json:"resolution_state,omitempty"`
}

type CompanyResearchTaskSpec struct {
	ID           string                  `json:"id,omitempty"`
	Role         CompanyResearchTaskRole `json:"role,omitempty"`
	Title        string                  `json:"title,omitempty"`
	Description  string                  `json:"description,omitempty"`
	Dimensions   []string                `json:"dimensions,omitempty"`
	Required     bool                    `json:"required,omitempty"`
	DependsOn    []string                `json:"depends_on,omitempty"`
	ToolHints    []string                `json:"tool_hints,omitempty"`
	ExpectedKeys []string                `json:"expected_keys,omitempty"`
}

type CompanyResearchTaskResult struct {
	TaskID        string                    `json:"task_id,omitempty"`
	Role          CompanyResearchTaskRole   `json:"role,omitempty"`
	Status        CompanyResearchTaskStatus `json:"status,omitempty"`
	Dimensions    []string                  `json:"dimensions,omitempty"`
	EvidenceReady bool                      `json:"evidence_ready,omitempty"`
	AdapterStatus string                    `json:"adapter_status,omitempty"`
	FailureCode   string                    `json:"failure_code,omitempty"`
	FailureClass  string                    `json:"failure_class,omitempty"`
	ExecutorID    string                    `json:"executor_id,omitempty"`
	Summary       string                    `json:"summary,omitempty"`
	Diagnostics   map[string]string         `json:"diagnostics,omitempty"`
	Warnings      []string                  `json:"warnings,omitempty"`
}

func BuildCompanyResearchTaskPlan(input CompanyResearchTaskPlanInput) CompanyResearchTaskPlan {
	intent := input.Intent
	dimensions := NormalizeStrings(intent.RequestedDimensions)
	if len(dimensions) == 0 {
		dimensions = []string{"financials", "market_data", "news", "risk"}
	}
	outputs := NormalizeStrings(intent.RequestedOutputs)
	if len(outputs) == 0 {
		outputs = []string{"brief", "risk_summary", "investment_boundary"}
	}
	caseType := strings.TrimSpace(input.CaseType)
	if caseType == "" {
		caseType = "company_research.single_company"
	}
	workflowID := strings.TrimSpace(input.WorkflowID)
	if workflowID == "" {
		workflowID = WorkflowID
	}
	plan := CompanyResearchTaskPlan{
		SchemaVersion: "company_research.taskforce.v1",
		PlanID:        stablePlanID(caseType, workflowID),
		CaseType:      caseType,
		WorkflowID:    workflowID,
		TaskKind:      intent.TaskKind,
		UserMessage:   intent.UserMessage,
		Subject:       taskSubjectRef(intent, input.SubjectResolution),
		Dimensions:    dimensions,
		Outputs:       outputs,
	}
	plan.Tasks = append(plan.Tasks, CompanyResearchTaskSpec{
		ID:           "subject_resolution",
		Role:         CompanyResearchRoleSubjectResolver,
		Title:        "Resolve company subject",
		Description:  "Resolve the customer-provided company, product, app, alias, or ticker mention into a verified company/security candidate when host evidence is available.",
		Required:     false,
		ToolHints:    []string{"host_subject_resolver"},
		ExpectedKeys: []string{"canonical_name", "market_hint", "stock_code", "ticker", "evidence_url", "verified"},
	})
	evidenceTaskIDs := []string{}
	if containsDimension(dimensions, "financials") {
		plan.Tasks = append(plan.Tasks, CompanyResearchTaskSpec{
			ID:           "finance_analysis",
			Role:         CompanyResearchRoleFinanceAnalyst,
			Title:        "Analyze public financial evidence",
			Description:  "Collect source-backed financial report, metrics, report-period, and finance-summary evidence through host-owned finance adapters.",
			Dimensions:   []string{"financials"},
			Required:     true,
			DependsOn:    []string{"subject_resolution"},
			ToolHints:    []string{"finance_report_lookup"},
			ExpectedKeys: []string{"adapter_status", "brief", "metrics", "candidates"},
		})
		evidenceTaskIDs = append(evidenceTaskIDs, "finance_analysis")
	}
	if containsDimension(dimensions, "market_data", "valuation", "announcements", "research") {
		plan.Tasks = append(plan.Tasks, CompanyResearchTaskSpec{
			ID:           "market_analysis",
			Role:         CompanyResearchRoleMarketAnalyst,
			Title:        "Analyze market and valuation evidence",
			Description:  "Collect quote, valuation, announcement, profile, and market signal evidence through host-owned stock adapters.",
			Dimensions:   intersectDimensions(dimensions, "market_data", "valuation", "announcements", "research"),
			Required:     true,
			DependsOn:    []string{"subject_resolution"},
			ToolHints:    []string{"a_stock_investigation", "global_stock_investigation"},
			ExpectedKeys: []string{"adapter_status", "quote", "subject", "readiness"},
		})
		evidenceTaskIDs = append(evidenceTaskIDs, "market_analysis")
	}
	if containsDimension(dimensions, "news", "risk") {
		plan.Tasks = append(plan.Tasks, CompanyResearchTaskSpec{
			ID:           "news_analysis",
			Role:         CompanyResearchRoleNewsAnalyst,
			Title:        "Analyze latest public news evidence",
			Description:  "Collect public-news and current-event evidence through host-owned search, browser, and source policy.",
			Dimensions:   intersectDimensions(dimensions, "news", "risk"),
			Required:     containsDimension(dimensions, "news"),
			DependsOn:    []string{"subject_resolution"},
			ToolHints:    []string{"latest_news_lookup"},
			ExpectedKeys: []string{"adapter_status", "sources", "primary_source", "evaluator_report"},
		})
		evidenceTaskIDs = append(evidenceTaskIDs, "news_analysis")
	}
	if containsDimension(dimensions, "risk") {
		deps := append([]string{}, evidenceTaskIDs...)
		if len(deps) == 0 {
			deps = []string{"subject_resolution"}
		}
		plan.Tasks = append(plan.Tasks, CompanyResearchTaskSpec{
			ID:          "risk_review",
			Role:        CompanyResearchRoleRiskReviewer,
			Title:       "Review company risk boundary",
			Description: "Review source-backed finance, market, and news evidence for bounded risk statements and missing-evidence warnings.",
			Dimensions:  []string{"risk"},
			Required:    true,
			DependsOn:   deps,
			ToolHints:   []string{ToolCompanyResearchGuard},
		})
		evidenceTaskIDs = append(evidenceTaskIDs, "risk_review")
	}
	guardDeps := append([]string{}, evidenceTaskIDs...)
	if len(guardDeps) == 0 {
		guardDeps = []string{"subject_resolution"}
	}
	plan.Tasks = append(plan.Tasks, CompanyResearchTaskSpec{
		ID:          "evidence_guard",
		Role:        CompanyResearchRoleEvidenceGuard,
		Title:       "Guard evidence readiness",
		Description: "Check task-level evidence readiness, missing dimensions, identity mismatch, and answer-boundary risk.",
		Required:    true,
		DependsOn:   guardDeps,
		ToolHints:   []string{ToolCompanyResearchGuard},
	})
	plan.Tasks = append(plan.Tasks, CompanyResearchTaskSpec{
		ID:          "synthesis",
		Role:        CompanyResearchRoleSynthesisEditor,
		Title:       "Synthesize bounded answer",
		Description: "Produce a final answer only within the allowed evidence scope and degrade explicitly when requested dimensions are missing.",
		Required:    true,
		DependsOn:   []string{"evidence_guard"},
	})
	if !plan.Subject.Verified {
		plan.Warnings = append(plan.Warnings, "subject_not_verified_by_task_plan")
	}
	return plan
}

func SubjectResolutionTaskResult(plan CompanyResearchTaskPlan) (CompanyResearchTaskResult, bool) {
	task, ok := plan.TaskByRole(CompanyResearchRoleSubjectResolver)
	if !ok {
		return CompanyResearchTaskResult{}, false
	}
	result := CompanyResearchTaskResult{
		TaskID:     task.ID,
		Role:       task.Role,
		Dimensions: task.Dimensions,
	}
	switch {
	case plan.Subject.Verified:
		result.Status = CompanyResearchTaskStatusReady
		result.EvidenceReady = true
		result.AdapterStatus = "ok"
		result.Summary = "subject resolved from host-backed evidence"
	case plan.Subject.FailureCode != "":
		result.Status = CompanyResearchTaskStatusDegraded
		result.AdapterStatus = plan.Subject.ResolutionState
		result.FailureCode = plan.Subject.FailureCode
		result.Summary = "subject resolver did not verify the input term"
	default:
		result.Status = CompanyResearchTaskStatusSkipped
		result.AdapterStatus = plan.Subject.ResolutionState
		result.Summary = "no verified subject resolution attached; downstream adapters must verify identity"
		result.Warnings = []string{"subject_resolution_not_configured_or_not_run"}
	}
	return result, true
}

func TaskResultFromEvidence(plan CompanyResearchTaskPlan, role CompanyResearchTaskRole, evidence ...map[string]any) (CompanyResearchTaskResult, bool) {
	task, ok := plan.TaskByRole(role)
	if !ok {
		return CompanyResearchTaskResult{}, false
	}
	statuses := []string{}
	failures := []string{}
	failureClasses := []string{}
	hasEvidence := false
	hasReady := false
	hasFailure := false
	for _, item := range evidence {
		if len(item) == 0 {
			continue
		}
		hasEvidence = true
		status := strings.TrimSpace(StringArg(item["adapter_status"]))
		if status != "" {
			statuses = append(statuses, status)
		}
		failure := strings.TrimSpace(StringArg(item["failure_code"]))
		if failure != "" {
			failures = append(failures, failure)
		}
		failureClass := strings.TrimSpace(StringArg(item["failure_class"]))
		if failureClass != "" {
			failureClasses = append(failureClasses, failureClass)
		}
		if EvidencePayloadReady(item) {
			hasReady = true
		}
		if evidencePayloadFailed(item) {
			hasFailure = true
		}
	}
	result := CompanyResearchTaskResult{
		TaskID:        task.ID,
		Role:          task.Role,
		Dimensions:    task.Dimensions,
		EvidenceReady: hasReady,
		AdapterStatus: strings.Join(NormalizeStrings(statuses), ","),
		FailureCode:   strings.Join(NormalizeStrings(failures), ","),
		FailureClass:  strings.Join(NormalizeStrings(failureClasses), ","),
	}
	switch {
	case hasReady:
		result.Status = CompanyResearchTaskStatusReady
		result.Summary = "task evidence is ready"
	case !hasEvidence:
		result.Status = CompanyResearchTaskStatusSkipped
		result.FailureCode = firstTaskResultNonEmpty(result.FailureCode, "task_not_executed")
		result.Summary = "task was not executed or returned no evidence"
	case hasFailure:
		result.Status = CompanyResearchTaskStatusFailed
		result.Summary = "task evidence failed"
	default:
		result.Status = CompanyResearchTaskStatusDegraded
		result.Summary = "task evidence is incomplete or degraded"
	}
	return result, true
}

func TaskResultFromReadiness(plan CompanyResearchTaskPlan, role CompanyResearchTaskRole, readiness CompanyResearchAnswerReadiness) (CompanyResearchTaskResult, bool) {
	task, ok := plan.TaskByRole(role)
	if !ok {
		return CompanyResearchTaskResult{}, false
	}
	result := CompanyResearchTaskResult{
		TaskID:        task.ID,
		Role:          task.Role,
		Dimensions:    task.Dimensions,
		EvidenceReady: readiness.AnswerReady,
		FailureCode:   readiness.FailureCode,
		FailureClass:  readiness.FailureClass,
	}
	switch {
	case readiness.AnswerReady && !readiness.Degraded:
		result.Status = CompanyResearchTaskStatusReady
		result.AdapterStatus = "ok"
		result.Summary = "requested evidence scope is ready"
	case readiness.SafeToAnswer:
		result.Status = CompanyResearchTaskStatusDegraded
		result.AdapterStatus = "degraded"
		result.Summary = firstTaskResultNonEmpty(readiness.DegradeReason, "bounded partial answer is allowed")
	default:
		result.Status = CompanyResearchTaskStatusFailed
		result.AdapterStatus = "failed"
		result.Summary = firstTaskResultNonEmpty(readiness.DegradeReason, "answer is not safe")
	}
	return result, true
}

func TaskResultFromAnswerContract(plan CompanyResearchTaskPlan, contract *CompanyResearchAnswerContract) (CompanyResearchTaskResult, bool) {
	task, ok := plan.TaskByRole(CompanyResearchRoleSynthesisEditor)
	if !ok {
		return CompanyResearchTaskResult{}, false
	}
	result := CompanyResearchTaskResult{
		TaskID:     task.ID,
		Role:       task.Role,
		Dimensions: task.Dimensions,
	}
	if contract == nil {
		result.Status = CompanyResearchTaskStatusSkipped
		result.FailureCode = "answer_contract_missing"
		result.Summary = "answer contract was not produced"
		return result, true
	}
	if contract.FinalAnswerRecommended {
		result.Status = CompanyResearchTaskStatusReady
		result.EvidenceReady = true
		result.AdapterStatus = "ok"
		result.Summary = firstTaskResultNonEmpty(contract.AllowedSummaryScope, "final answer contract is ready")
		return result, true
	}
	result.Status = CompanyResearchTaskStatusDegraded
	result.AdapterStatus = "degraded"
	result.Summary = firstTaskResultNonEmpty(contract.Reason, "final answer is bounded by missing evidence")
	return result, true
}

func EvidencePayloadReady(evidence map[string]any) bool {
	if len(evidence) == 0 {
		return false
	}
	status := strings.ToLower(strings.TrimSpace(StringArg(evidence["adapter_status"])))
	if status == "unsupported" || status == "error" || status == "evidence_incomplete" || status == "unavailable" {
		return false
	}
	if strings.TrimSpace(StringArg(evidence["failure_code"])) != "" {
		return false
	}
	return true
}

func (plan CompanyResearchTaskPlan) TaskByRole(role CompanyResearchTaskRole) (CompanyResearchTaskSpec, bool) {
	for _, task := range plan.Tasks {
		if task.Role == role {
			return task, true
		}
	}
	return CompanyResearchTaskSpec{}, false
}

func taskSubjectRef(intent CompanyResearchIntent, resolution *SubjectResolution) CompanyResearchSubjectRef {
	ref := CompanyResearchSubjectRef{
		InputTerm:      firstTaskResultNonEmpty(intent.EntityName, firstTaskResultString(intent.EntityMentions), intent.UserMessage),
		CanonicalName:  intent.EntityName,
		EntityMentions: append([]string(nil), intent.EntityMentions...),
		MarketHint:     intent.MarketHint,
	}
	if resolution == nil {
		ref.ResolutionState = "not_configured_or_not_run"
		return ref
	}
	ref.InputTerm = firstTaskResultNonEmpty(resolution.InputTerm, ref.InputTerm)
	ref.MarketHint = firstTaskResultNonEmpty(resolution.PreferredMarket, ref.MarketHint)
	ref.FailureCode = strings.TrimSpace(resolution.FailureCode)
	ref.ResolutionState = strings.TrimSpace(resolution.AdapterStatus)
	if ref.ResolutionState == "" {
		ref.ResolutionState = "unknown"
	}
	if resolution.SelectedCandidate == nil {
		return ref
	}
	candidate := *resolution.SelectedCandidate
	ref.CanonicalName = firstTaskResultNonEmpty(candidate.EntityName, candidate.DisplayName, ref.CanonicalName)
	ref.MarketHint = firstTaskResultNonEmpty(candidate.Market, ref.MarketHint)
	ref.StockCode = strings.TrimSpace(candidate.StockCode)
	ref.Ticker = strings.TrimSpace(candidate.Ticker)
	ref.Source = strings.TrimSpace(candidate.Source)
	ref.EvidenceURL = strings.TrimSpace(candidate.EvidenceURL)
	ref.Verified = candidate.Verified && strings.TrimSpace(candidate.MismatchReason) == ""
	ref.Confidence = candidate.Confidence
	return ref
}

func stablePlanID(caseType string, workflowID string) string {
	caseType = strings.NewReplacer(".", "_", "-", "_", " ", "_").Replace(strings.ToLower(strings.TrimSpace(caseType)))
	workflowID = strings.NewReplacer(".", "_", "-", "_", " ", "_").Replace(strings.ToLower(strings.TrimSpace(workflowID)))
	return firstTaskResultNonEmpty(caseType+"__"+workflowID, "company_research_taskforce")
}

func containsDimension(values []string, candidates ...string) bool {
	for _, value := range values {
		for _, candidate := range candidates {
			if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(candidate)) {
				return true
			}
		}
	}
	return false
}

func intersectDimensions(values []string, candidates ...string) []string {
	out := []string{}
	for _, value := range values {
		for _, candidate := range candidates {
			if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(candidate)) {
				out = append(out, strings.TrimSpace(value))
				break
			}
		}
	}
	return NormalizeStrings(out)
}

func evidencePayloadFailed(evidence map[string]any) bool {
	status := strings.ToLower(strings.TrimSpace(StringArg(evidence["adapter_status"])))
	return status == "error"
}

func firstTaskResultString(values []string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstTaskResultNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
