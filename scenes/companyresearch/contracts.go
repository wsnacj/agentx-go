package companyresearch

type CompanySubject struct {
	EntityName     string   `json:"entity_name,omitempty"`
	EntityMentions []string `json:"entity_mentions,omitempty"`
	MarketHint     string   `json:"market_hint,omitempty"`
}

type CompanyResearchIntent struct {
	UserMessage         string           `json:"user_message,omitempty"`
	TaskKind            string           `json:"task_kind,omitempty"`
	EntityName          string           `json:"entity_name,omitempty"`
	EntityMentions      []string         `json:"entity_mentions,omitempty"`
	MarketHint          string           `json:"market_hint,omitempty"`
	ComparisonSubjects  []CompanySubject `json:"comparison_subjects,omitempty"`
	RequestedDimensions []string         `json:"requested_dimensions,omitempty"`
	RequestedOutputs    []string         `json:"requested_outputs,omitempty"`
	Freshness           map[string]any   `json:"freshness,omitempty"`
	RiskScope           string           `json:"risk_scope,omitempty"`
	SourcePolicy        string           `json:"source_policy,omitempty"`
	OriginalIntent      string           `json:"original_intent,omitempty"`
	StopCondition       string           `json:"stop_condition,omitempty"`
}

type CompanyResearchEvidence struct {
	Finance     map[string]any `json:"finance,omitempty"`
	AStock      map[string]any `json:"a_stock,omitempty"`
	GlobalStock map[string]any `json:"global_stock,omitempty"`
	News        map[string]any `json:"news,omitempty"`
	Guard       map[string]any `json:"guard,omitempty"`
}

type CompanyResearchAnswerReadiness struct {
	AnswerReady       bool     `json:"answer_ready"`
	SafeToAnswer      bool     `json:"safe_to_answer"`
	Degraded          bool     `json:"degraded,omitempty"`
	DegradeReason     string   `json:"degrade_reason,omitempty"`
	AllowedScope      string   `json:"allowed_summary_scope,omitempty"`
	MissingDimensions []string `json:"missing_dimensions,omitempty"`
	ReadyDimensions   []string `json:"ready_dimensions,omitempty"`
	FailureCode       string   `json:"failure_code,omitempty"`
	FailureClass      string   `json:"failure_class,omitempty"`
}

type CompanyResearchPayload struct {
	Tool              string                         `json:"tool,omitempty"`
	Source            string                         `json:"source,omitempty"`
	PackID            string                         `json:"pack_id,omitempty"`
	CaseType          string                         `json:"case_type,omitempty"`
	WorkflowID        string                         `json:"workflow_id,omitempty"`
	AdapterID         string                         `json:"adapter_id,omitempty"`
	AdapterStatus     string                         `json:"adapter_status,omitempty"`
	FailureCode       string                         `json:"failure_code,omitempty"`
	FailureClass      string                         `json:"failure_class,omitempty"`
	GuardStatus       string                         `json:"guard_status,omitempty"`
	Intent            CompanyResearchIntent          `json:"intent,omitempty"`
	SubjectResolution *SubjectResolution             `json:"subject_resolution,omitempty"`
	TaskPlan          *CompanyResearchTaskPlan       `json:"task_plan,omitempty"`
	TaskResults       []CompanyResearchTaskResult    `json:"task_results,omitempty"`
	TaskSummary       *CompanyResearchTaskSummary    `json:"task_summary,omitempty"`
	Evidence          CompanyResearchEvidence        `json:"evidence,omitempty"`
	Subjects          []CompanyResearchPayload       `json:"subjects,omitempty"`
	AnswerReadiness   CompanyResearchAnswerReadiness `json:"answer_readiness,omitempty"`
	AnswerContract    *CompanyResearchAnswerContract `json:"answer_contract,omitempty"`
	Warnings          []string                       `json:"warnings,omitempty"`
}

type CompanyResearchAnswerContract struct {
	FinalAnswerRecommended bool                            `json:"final_answer_recommended"`
	Reason                 string                          `json:"reason,omitempty"`
	AllowedSummaryScope    string                          `json:"allowed_summary_scope,omitempty"`
	DoNotRetryTools        []string                        `json:"do_not_retry_tools,omitempty"`
	RecoveryRecommended    bool                            `json:"recovery_recommended,omitempty"`
	RecoveryReason         string                          `json:"recovery_reason,omitempty"`
	SuggestedRecoveryTools []string                        `json:"suggested_recovery_tools,omitempty"`
	RecoveryTargets        []CompanyResearchRecoveryTarget `json:"recovery_targets,omitempty"`
	PossibleImpact         string                          `json:"possible_impact,omitempty"`
	RiskBoundary           string                          `json:"risk_boundary,omitempty"`
	FinalAnswerDraft       string                          `json:"final_answer_draft,omitempty"`
	SubjectSummaries       []CompanyResearchSubjectSummary `json:"subject_summaries,omitempty"`
	SubjectBudget          *CompanyResearchSubjectBudget   `json:"subject_budget,omitempty"`
	Truncated              bool                            `json:"truncated"`
}

type CompanyResearchSubjectSummary struct {
	EntityName        string   `json:"entity_name"`
	AnswerReady       bool     `json:"answer_ready"`
	ReadyDimensions   []string `json:"ready_dimensions,omitempty"`
	MissingDimensions []string `json:"missing_dimensions,omitempty"`
	Summary           string   `json:"summary,omitempty"`
	Truncated         bool     `json:"truncated"`
}

type CompanyResearchSubjectBudget struct {
	MaxSubjects        int `json:"max_subjects"`
	MaxRunesPerSubject int `json:"max_runes_per_subject"`
	TotalSubjects      int `json:"total_subjects"`
	ReturnedSubjects   int `json:"returned_subjects"`
}

type CompanyResearchRecoveryTarget struct {
	EntityName       string   `json:"entity_name,omitempty"`
	MissingDimension string   `json:"missing_dimension,omitempty"`
	FailureClass     string   `json:"failure_class,omitempty"`
	FailureCode      string   `json:"failure_code,omitempty"`
	SuggestedTools   []string `json:"suggested_tools,omitempty"`
}

type CompanyResearchTaskSummary struct {
	ReadyRoles        []CompanyResearchTaskRole     `json:"ready_roles,omitempty"`
	DegradedRoles     []CompanyResearchTaskRole     `json:"degraded_roles,omitempty"`
	FailedRoles       []CompanyResearchTaskRole     `json:"failed_roles,omitempty"`
	SkippedRoles      []CompanyResearchTaskRole     `json:"skipped_roles,omitempty"`
	ReadyDimensions   []string                      `json:"ready_dimensions,omitempty"`
	MissingDimensions []string                      `json:"missing_dimensions,omitempty"`
	Conflicts         []CompanyResearchTaskConflict `json:"conflicts,omitempty"`
	Warnings          []string                      `json:"warnings,omitempty"`
}

type CompanyResearchTaskConflict struct {
	Code      string                  `json:"code,omitempty"`
	Subject   string                  `json:"subject,omitempty"`
	Role      CompanyResearchTaskRole `json:"role,omitempty"`
	OtherRole CompanyResearchTaskRole `json:"other_role,omitempty"`
	Dimension string                  `json:"dimension,omitempty"`
	Expected  string                  `json:"expected,omitempty"`
	Observed  string                  `json:"observed,omitempty"`
	Summary   string                  `json:"summary,omitempty"`
}
