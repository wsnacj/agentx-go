package contracts

// ResearchEvaluationInput contains the portable evidence used by the research evaluator.
// It does not trigger provider access or investment-advice generation.
type ResearchEvaluationInput struct {
	ExpectedEntityName        string
	ExpectedStockCode         string
	EvidenceEntityName        string
	EvidenceStockCode         string
	AdapterStatus             AdapterStatus
	FailureCode               FailureCode
	AnswerReady               bool
	RequestedFields           []string
	FieldValues               map[string]string
	ConsensusFields           []string
	ReportCount               int
	LatestPublishedAt         string
	AsOf                      string
	SourceURLs                []string
	MissingRequestedFields    []string
	ReviewRequiredFields      []string
	InvestmentAdviceRequested bool
	AdviceBoundaryStated      bool
}

// ResearchEvaluation is the deterministic result of evaluating research evidence.
type ResearchEvaluation struct {
	Passed                  bool     `json:"passed"`
	SubjectCorrect          bool     `json:"subject_correct"`
	FreshnessAccepted       bool     `json:"freshness_accepted"`
	FieldsReady             bool     `json:"fields_ready"`
	SourceAccepted          bool     `json:"source_accepted"`
	AdviceBoundaryRespected bool     `json:"advice_boundary_respected"`
	MissingRequestedFields  []string `json:"missing_requested_fields,omitempty"`
	ReviewRequiredFields    []string `json:"review_required_fields,omitempty"`
	RequestedFields         []string `json:"requested_fields,omitempty"`
	SourceURLs              []string `json:"source_urls,omitempty"`
	AdapterStatus           string   `json:"adapter_status,omitempty"`
	FailureCode             string   `json:"failure_code,omitempty"`
	FailureReason           string   `json:"failure_reason,omitempty"`
}

// SignalEvaluationInput contains the portable evidence used by the signal evaluator.
// It does not trigger provider access or investment-advice generation.
type SignalEvaluationInput struct {
	ExpectedEntityName        string
	ExpectedStockCode         string
	EvidenceEntityName        string
	EvidenceStockCode         string
	AdapterStatus             AdapterStatus
	FailureCode               FailureCode
	AnswerReady               bool
	Degraded                  bool
	RequestedSignalTypes      []string
	ReturnedSignalTypes       []string
	TradeDate                 string
	AsOf                      string
	MissingRequestedFields    []string
	ReviewRequiredFields      []string
	SourceURLs                []string
	InvestmentAdviceRequested bool
	AdviceBoundaryStated      bool
}

// SignalEvaluation is the deterministic result of evaluating signal evidence.
type SignalEvaluation struct {
	Passed                  bool     `json:"passed"`
	SubjectCorrect          bool     `json:"subject_correct"`
	FreshnessAccepted       bool     `json:"freshness_accepted"`
	FieldsReady             bool     `json:"fields_ready"`
	SourceAccepted          bool     `json:"source_accepted"`
	AdviceBoundaryRespected bool     `json:"advice_boundary_respected"`
	MissingRequestedFields  []string `json:"missing_requested_fields,omitempty"`
	ReviewRequiredFields    []string `json:"review_required_fields,omitempty"`
	RequestedSignalTypes    []string `json:"requested_signal_types,omitempty"`
	ReturnedSignalTypes     []string `json:"returned_signal_types,omitempty"`
	SourceURLs              []string `json:"source_urls,omitempty"`
	AdapterStatus           string   `json:"adapter_status,omitempty"`
	FailureCode             string   `json:"failure_code,omitempty"`
	FailureReason           string   `json:"failure_reason,omitempty"`
}

// ValuationEvaluationInput contains the portable evidence used by the valuation evaluator.
// It does not trigger provider access or investment-advice generation.
type ValuationEvaluationInput struct {
	ExpectedEntityName        string
	ExpectedStockCode         string
	EvidenceEntityName        string
	EvidenceStockCode         string
	AdapterStatus             AdapterStatus
	FailureCode               FailureCode
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

// ValuationEvaluation is the deterministic result of evaluating valuation evidence.
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
