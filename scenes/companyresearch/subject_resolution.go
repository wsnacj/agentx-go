package companyresearch

type SubjectResolutionRequest struct {
	UserMessage         string         `json:"user_message,omitempty"`
	EntityName          string         `json:"entity_name,omitempty"`
	EntityMentions      []string       `json:"entity_mentions,omitempty"`
	MarketHint          string         `json:"market_hint,omitempty"`
	RequestedDimensions []string       `json:"requested_dimensions,omitempty"`
	RequestedOutputs    []string       `json:"requested_outputs,omitempty"`
	Freshness           map[string]any `json:"freshness,omitempty"`
	SourcePolicy        string         `json:"source_policy,omitempty"`
	OriginalIntent      string         `json:"original_intent,omitempty"`
}

type SubjectResolution struct {
	AdapterID         string                       `json:"adapter_id,omitempty"`
	AdapterStatus     string                       `json:"adapter_status,omitempty"`
	FailureCode       string                       `json:"failure_code,omitempty"`
	InputTerm         string                       `json:"input_term,omitempty"`
	PreferredMarket   string                       `json:"preferred_market,omitempty"`
	Strategy          string                       `json:"strategy,omitempty"`
	SelectedReason    string                       `json:"selected_reason,omitempty"`
	SelectedCandidate *SubjectResolutionCandidate  `json:"selected_candidate,omitempty"`
	Candidates        []SubjectResolutionCandidate `json:"candidates,omitempty"`
	QueryVariants     []SubjectResolutionQuery     `json:"query_variants,omitempty"`
	Warnings          []string                     `json:"warnings,omitempty"`
}

type SubjectResolutionCandidate struct {
	EntityName     string  `json:"entity_name,omitempty"`
	DisplayName    string  `json:"display_name,omitempty"`
	StockCode      string  `json:"stock_code,omitempty"`
	Ticker         string  `json:"ticker,omitempty"`
	Market         string  `json:"market,omitempty"`
	Exchange       string  `json:"exchange,omitempty"`
	Source         string  `json:"source,omitempty"`
	EvidenceURL    string  `json:"evidence_url,omitempty"`
	Confidence     float64 `json:"confidence,omitempty"`
	Verified       bool    `json:"verified,omitempty"`
	MatchReason    string  `json:"match_reason,omitempty"`
	MismatchReason string  `json:"mismatch_reason,omitempty"`
}

type SubjectResolutionQuery struct {
	Term           string `json:"term,omitempty"`
	Reason         string `json:"reason,omitempty"`
	Priority       int    `json:"priority,omitempty"`
	Provider       string `json:"provider,omitempty"`
	Status         string `json:"status,omitempty"`
	CandidateCount int    `json:"candidate_count,omitempty"`
	Message        string `json:"message,omitempty"`
}

func SubjectResolutionRequestFromIntent(intent CompanyResearchIntent) SubjectResolutionRequest {
	return SubjectResolutionRequest{
		UserMessage:         intent.UserMessage,
		EntityName:          intent.EntityName,
		EntityMentions:      append([]string(nil), intent.EntityMentions...),
		MarketHint:          intent.MarketHint,
		RequestedDimensions: append([]string(nil), intent.RequestedDimensions...),
		RequestedOutputs:    append([]string(nil), intent.RequestedOutputs...),
		Freshness:           intent.Freshness,
		SourcePolicy:        intent.SourcePolicy,
		OriginalIntent:      intent.OriginalIntent,
	}
}
