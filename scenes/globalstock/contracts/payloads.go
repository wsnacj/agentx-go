package contracts

// MetricValue represents a source-backed scalar value without forcing early parsing.
type MetricValue struct {
	Field          string         `json:"field,omitempty"`
	Value          string         `json:"value,omitempty"`
	Unit           string         `json:"unit,omitempty"`
	Currency       string         `json:"currency,omitempty"`
	AsOf           string         `json:"as_of,omitempty"`
	Evidence       SourceEvidence `json:"evidence,omitempty"`
	Confidence     float64        `json:"confidence,omitempty"`
	ReviewRequired bool           `json:"review_required,omitempty"`
}

// MetricUnitHundredMillion is the provider scale used by HK/US market-cap
// fields. Keep it language-neutral so consumers do not mistake it for billions.
const MetricUnitHundredMillion = "100_million"

// TaskKind describes the high-level global-stock task frame selected by the model.
type TaskKind string

const (
	TaskKindQuoteSnapshot        TaskKind = "quote_snapshot"
	TaskKindValuationSnapshot    TaskKind = "valuation_snapshot"
	TaskKindResearchLookup       TaskKind = "research_lookup"
	TaskKindAnnouncement         TaskKind = "announcement_lookup"
	TaskKindProfileLookup        TaskKind = "profile_lookup"
	TaskKindSignalLookup         TaskKind = "signal_lookup"
	TaskKindScreening            TaskKind = "screening"
	TaskKindComparison           TaskKind = "comparison"
	TaskKindFullInvestigation    TaskKind = "full_investigation"
	TaskKindFinanceReportHandoff TaskKind = "finance_report_handoff"
)

// SignalType identifies a source-backed HK/US market signal family.
type SignalType string

const (
	SignalTypeHKDisclosure    SignalType = "hk_disclosure"
	SignalTypeHKBuyback       SignalType = "hk_buyback"
	SignalTypeHKBoardMeeting  SignalType = "hk_board_meeting"
	SignalTypeHKShortSelling  SignalType = "hk_short_selling"
	SignalTypeHKSouthbound    SignalType = "hk_southbound_flow"
	SignalTypeUSForm4         SignalType = "us_form_4"
	SignalTypeUS8K            SignalType = "us_8k"
	SignalTypeUS13F           SignalType = "us_13f"
	SignalTypeUSEarningsEvent SignalType = "us_earnings_event"
)

// Subject identifies a global security after adapter verification.
type Subject struct {
	EntityName  string         `json:"entity_name,omitempty"`
	DisplayName string         `json:"display_name,omitempty"`
	StockCode   string         `json:"stock_code,omitempty"`
	Market      Market         `json:"market,omitempty"`
	Exchange    string         `json:"exchange,omitempty"`
	Currency    string         `json:"currency,omitempty"`
	Verified    bool           `json:"verified,omitempty"`
	Evidence    SourceEvidence `json:"evidence,omitempty"`
}

// IdentityResolution records how a natural-language subject was resolved.
// It is diagnostic-only: selected facts still need source-backed payload readiness.
type IdentityResolution struct {
	InputTerm         string                        `json:"input_term,omitempty"`
	PreferredMarket   Market                        `json:"preferred_market,omitempty"`
	Strategy          string                        `json:"strategy,omitempty"`
	SelectedReason    string                        `json:"selected_reason,omitempty"`
	SelectedCandidate *IdentityResolutionCandidate  `json:"selected_candidate,omitempty"`
	QueryVariants     []IdentityResolutionQuery     `json:"query_variants,omitempty"`
	Candidates        []IdentityResolutionCandidate `json:"candidates,omitempty"`
	Warnings          []string                      `json:"warnings,omitempty"`
}

// IdentityResolutionQuery records one provider query expansion attempt.
type IdentityResolutionQuery struct {
	Term           string `json:"term,omitempty"`
	Reason         string `json:"reason,omitempty"`
	Priority       int    `json:"priority,omitempty"`
	Provider       string `json:"provider,omitempty"`
	Status         string `json:"status,omitempty"`
	CandidateCount int    `json:"candidate_count,omitempty"`
	Message        string `json:"message,omitempty"`
}

// IdentityResolutionCandidate records one source-backed resolver candidate.
type IdentityResolutionCandidate struct {
	Name           string `json:"name,omitempty"`
	Code           string `json:"code,omitempty"`
	Market         Market `json:"market,omitempty"`
	Exchange       string `json:"exchange,omitempty"`
	QuoteID        string `json:"quote_id,omitempty"`
	SourceURL      string `json:"source_url,omitempty"`
	QueryTerm      string `json:"query_term,omitempty"`
	QueryReason    string `json:"query_reason,omitempty"`
	QueryPriority  int    `json:"query_priority,omitempty"`
	Score          int    `json:"score,omitempty"`
	Selected       bool   `json:"selected,omitempty"`
	SelectedReason string `json:"selected_reason,omitempty"`
}

// QuotePayload contains realtime or delayed quote and valuation snapshot evidence.
type QuotePayload struct {
	Tool               string              `json:"tool,omitempty"`
	Source             string              `json:"source,omitempty"`
	AdapterID          string              `json:"adapter_id,omitempty"`
	AdapterStatus      AdapterStatus       `json:"adapter_status,omitempty"`
	FailureCode        FailureCode         `json:"failure_code,omitempty"`
	Subject            Subject             `json:"subject,omitempty"`
	Freshness          Freshness           `json:"freshness,omitempty"`
	Evidence           SourceEvidence      `json:"evidence,omitempty"`
	Readiness          Readiness           `json:"readiness,omitempty"`
	Quote              QuoteSnapshot       `json:"quote,omitempty"`
	Warnings           []string            `json:"warnings,omitempty"`
	ProviderChain      []ProviderAttempt   `json:"provider_chain,omitempty"`
	IdentityResolution *IdentityResolution `json:"identity_resolution,omitempty"`
}

// QuoteSnapshot contains common global stock quote and valuation fields.
type QuoteSnapshot struct {
	Price           MetricValue `json:"price,omitempty"`
	LastClose       MetricValue `json:"last_close,omitempty"`
	Open            MetricValue `json:"open,omitempty"`
	High            MetricValue `json:"high,omitempty"`
	Low             MetricValue `json:"low,omitempty"`
	ChangeAmount    MetricValue `json:"change_amount,omitempty"`
	ChangePercent   MetricValue `json:"change_percent,omitempty"`
	Volume          MetricValue `json:"volume,omitempty"`
	Amount          MetricValue `json:"amount,omitempty"`
	TurnoverPercent MetricValue `json:"turnover_percent,omitempty"`
	PETTM           MetricValue `json:"pe_ttm,omitempty"`
	PEStatic        MetricValue `json:"pe_static,omitempty"`
	PB              MetricValue `json:"pb,omitempty"`
	MarketCap       MetricValue `json:"market_cap,omitempty"`
	FloatMarketCap  MetricValue `json:"float_market_cap,omitempty"`
	High52Week      MetricValue `json:"high_52_week,omitempty"`
	Low52Week       MetricValue `json:"low_52_week,omitempty"`
}

// ProfilePayload contains source-backed listed-security profile evidence.
type ProfilePayload struct {
	Tool               string                 `json:"tool,omitempty"`
	Source             string                 `json:"source,omitempty"`
	AdapterID          string                 `json:"adapter_id,omitempty"`
	AdapterStatus      AdapterStatus          `json:"adapter_status,omitempty"`
	FailureCode        FailureCode            `json:"failure_code,omitempty"`
	Subject            Subject                `json:"subject,omitempty"`
	Freshness          Freshness              `json:"freshness,omitempty"`
	Evidence           SourceEvidence         `json:"evidence,omitempty"`
	Readiness          Readiness              `json:"readiness,omitempty"`
	Profile            CompanyProfile         `json:"profile,omitempty"`
	Finance            map[string]MetricValue `json:"finance,omitempty"`
	Warnings           []string               `json:"warnings,omitempty"`
	IdentityResolution *IdentityResolution    `json:"identity_resolution,omitempty"`
}

// AnnouncementPayload contains source-backed HKEX/SEC filing or announcement metadata.
type AnnouncementPayload struct {
	Tool               string               `json:"tool,omitempty"`
	Source             string               `json:"source,omitempty"`
	AdapterID          string               `json:"adapter_id,omitempty"`
	AdapterStatus      AdapterStatus        `json:"adapter_status,omitempty"`
	FailureCode        FailureCode          `json:"failure_code,omitempty"`
	Subject            Subject              `json:"subject,omitempty"`
	Freshness          Freshness            `json:"freshness,omitempty"`
	Evidence           SourceEvidence       `json:"evidence,omitempty"`
	Readiness          Readiness            `json:"readiness,omitempty"`
	Announcements      []AnnouncementRecord `json:"announcements,omitempty"`
	Warnings           []string             `json:"warnings,omitempty"`
	ProviderChain      []ProviderAttempt    `json:"provider_chain,omitempty"`
	IdentityResolution *IdentityResolution  `json:"identity_resolution,omitempty"`
}

// AnnouncementRecord is one public filing/announcement item.
type AnnouncementRecord struct {
	Title             string         `json:"title,omitempty"`
	Type              string         `json:"type,omitempty"`
	Form              string         `json:"form,omitempty"`
	PublishedAt       string         `json:"published_at,omitempty"`
	FilingDate        string         `json:"filing_date,omitempty"`
	AccessionNumber   string         `json:"accession_number,omitempty"`
	PrimaryDocument   string         `json:"primary_document,omitempty"`
	URL               string         `json:"url,omitempty"`
	FullTextAvailable bool           `json:"full_text_available,omitempty"`
	Summary           string         `json:"summary,omitempty"`
	Evidence          SourceEvidence `json:"evidence,omitempty"`
}

// ResearchPayload contains source-backed analyst rating and research evidence.
type ResearchPayload struct {
	Tool               string              `json:"tool,omitempty"`
	Source             string              `json:"source,omitempty"`
	AdapterID          string              `json:"adapter_id,omitempty"`
	AdapterStatus      AdapterStatus       `json:"adapter_status,omitempty"`
	FailureCode        FailureCode         `json:"failure_code,omitempty"`
	Subject            Subject             `json:"subject,omitempty"`
	Query              string              `json:"query,omitempty"`
	Freshness          Freshness           `json:"freshness,omitempty"`
	Evidence           SourceEvidence      `json:"evidence,omitempty"`
	Readiness          Readiness           `json:"readiness,omitempty"`
	Reports            []ResearchReport    `json:"reports,omitempty"`
	Consensus          []ForecastItem      `json:"consensus,omitempty"`
	Warnings           []string            `json:"warnings,omitempty"`
	ProviderChain      []ProviderAttempt   `json:"provider_chain,omitempty"`
	IdentityResolution *IdentityResolution `json:"identity_resolution,omitempty"`
}

// SignalPayload contains source-backed HK/US disclosure-style market signals.
type SignalPayload struct {
	Tool               string              `json:"tool,omitempty"`
	Source             string              `json:"source,omitempty"`
	AdapterID          string              `json:"adapter_id,omitempty"`
	AdapterStatus      AdapterStatus       `json:"adapter_status,omitempty"`
	FailureCode        FailureCode         `json:"failure_code,omitempty"`
	Subject            Subject             `json:"subject,omitempty"`
	Freshness          Freshness           `json:"freshness,omitempty"`
	Evidence           SourceEvidence      `json:"evidence,omitempty"`
	Readiness          Readiness           `json:"readiness,omitempty"`
	Signals            []SignalEvent       `json:"signals,omitempty"`
	Warnings           []string            `json:"warnings,omitempty"`
	ProviderChain      []ProviderAttempt   `json:"provider_chain,omitempty"`
	IdentityResolution *IdentityResolution `json:"identity_resolution,omitempty"`
}

// SignalEvent is one source-backed disclosure or filing signal.
type SignalEvent struct {
	Type              SignalType          `json:"type,omitempty"`
	Title             string              `json:"title,omitempty"`
	Form              string              `json:"form,omitempty"`
	PublishedAt       string              `json:"published_at,omitempty"`
	FilingDate        string              `json:"filing_date,omitempty"`
	URL               string              `json:"url,omitempty"`
	Summary           string              `json:"summary,omitempty"`
	OwnerName         string              `json:"owner_name,omitempty"`
	SecurityTitle     string              `json:"security_title,omitempty"`
	TransactionCode   string              `json:"transaction_code,omitempty"`
	TransactionDate   string              `json:"transaction_date,omitempty"`
	AcquiredDisposed  string              `json:"acquired_disposed,omitempty"`
	Shares            MetricValue         `json:"shares,omitempty"`
	Price             MetricValue         `json:"price,omitempty"`
	Metrics           []MetricValue       `json:"metrics,omitempty"`
	Transactions      []SignalTransaction `json:"transactions,omitempty"`
	Holdings          []SignalHolding     `json:"holdings,omitempty"`
	FullTextAvailable bool                `json:"full_text_available,omitempty"`
	Evidence          SourceEvidence      `json:"evidence,omitempty"`
}

// SignalTransaction is one parsed transaction inside a disclosure signal.
type SignalTransaction struct {
	SecurityTitle    string         `json:"security_title,omitempty"`
	TransactionCode  string         `json:"transaction_code,omitempty"`
	TransactionDate  string         `json:"transaction_date,omitempty"`
	AcquiredDisposed string         `json:"acquired_disposed,omitempty"`
	Shares           MetricValue    `json:"shares,omitempty"`
	Price            MetricValue    `json:"price,omitempty"`
	Evidence         SourceEvidence `json:"evidence,omitempty"`
}

// SignalHolding is one parsed portfolio holding inside a disclosure signal.
type SignalHolding struct {
	IssuerName           string         `json:"issuer_name,omitempty"`
	TitleOfClass         string         `json:"title_of_class,omitempty"`
	CUSIP                string         `json:"cusip,omitempty"`
	Value                MetricValue    `json:"value,omitempty"`
	Shares               MetricValue    `json:"shares,omitempty"`
	PutCall              string         `json:"put_call,omitempty"`
	InvestmentDiscretion string         `json:"investment_discretion,omitempty"`
	VotingSole           string         `json:"voting_sole,omitempty"`
	VotingShared         string         `json:"voting_shared,omitempty"`
	VotingNone           string         `json:"voting_none,omitempty"`
	Evidence             SourceEvidence `json:"evidence,omitempty"`
}

// ResearchReport contains one source-backed analyst rating or report record.
type ResearchReport struct {
	Title       string         `json:"title,omitempty"`
	Institution string         `json:"institution,omitempty"`
	Analyst     string         `json:"analyst,omitempty"`
	PublishedAt string         `json:"published_at,omitempty"`
	Rating      string         `json:"rating,omitempty"`
	Action      string         `json:"action,omitempty"`
	Summary     string         `json:"summary,omitempty"`
	PDFURL      string         `json:"pdf_url,omitempty"`
	SourceURL   string         `json:"source_url,omitempty"`
	Forecasts   []ForecastItem `json:"forecasts,omitempty"`
	Evidence    SourceEvidence `json:"evidence,omitempty"`
}

// ForecastItem contains target-price, upside, or consensus evidence.
type ForecastItem struct {
	Field      string         `json:"field,omitempty"`
	Year       string         `json:"year,omitempty"`
	Value      MetricValue    `json:"value,omitempty"`
	Min        MetricValue    `json:"min,omitempty"`
	Mean       MetricValue    `json:"mean,omitempty"`
	Max        MetricValue    `json:"max,omitempty"`
	Count      int            `json:"count,omitempty"`
	Evidence   SourceEvidence `json:"evidence,omitempty"`
	Confidence float64        `json:"confidence,omitempty"`
}

// ToolHandoff describes a source-boundary handoff to another domain package.
// The emitting package must not execute or fabricate facts for the target tool;
// it only provides a model/host-readable task frame for the next tool call.
type ToolHandoff struct {
	TargetPackage string         `json:"target_package,omitempty"`
	TargetTool    string         `json:"target_tool,omitempty"`
	Reason        string         `json:"reason,omitempty"`
	Required      bool           `json:"required,omitempty"`
	Arguments     map[string]any `json:"arguments,omitempty"`
	Boundary      string         `json:"boundary,omitempty"`
}

// CompanyProfile captures profile facts that are safe to answer from public HK/US sources.
type CompanyProfile struct {
	Name           string            `json:"name,omitempty"`
	DisplayName    string            `json:"display_name,omitempty"`
	StockCode      string            `json:"stock_code,omitempty"`
	Market         Market            `json:"market,omitempty"`
	Exchange       string            `json:"exchange,omitempty"`
	Currency       string            `json:"currency,omitempty"`
	SecurityType   string            `json:"security_type,omitempty"`
	Industry       string            `json:"industry,omitempty"`
	ListingDate    string            `json:"listing_date,omitempty"`
	Website        string            `json:"website,omitempty"`
	Description    string            `json:"description,omitempty"`
	MarketCap      MetricValue       `json:"market_cap,omitempty"`
	FloatMarketCap MetricValue       `json:"float_market_cap,omitempty"`
	PETTM          MetricValue       `json:"pe_ttm,omitempty"`
	PB             MetricValue       `json:"pb,omitempty"`
	High52Week     MetricValue       `json:"high_52_week,omitempty"`
	Low52Week      MetricValue       `json:"low_52_week,omitempty"`
	Sections       map[string]string `json:"sections,omitempty"`
	Evidence       SourceEvidence    `json:"evidence,omitempty"`
}

// InvestigationIntent is the model-filled candidate task frame.
type InvestigationIntent struct {
	UserMessage      string    `json:"user_message,omitempty"`
	TaskKind         TaskKind  `json:"task_kind,omitempty"`
	EntityName       string    `json:"entity_name,omitempty"`
	EntityMentions   []string  `json:"entity_mentions,omitempty"`
	StockCode        string    `json:"stock_code,omitempty"`
	Market           Market    `json:"market,omitempty"`
	RequestedFields  []string  `json:"requested_fields,omitempty"`
	RequestedOutputs []string  `json:"requested_outputs,omitempty"`
	Freshness        Freshness `json:"freshness,omitempty"`
	SourceHint       string    `json:"source_hint,omitempty"`
	SourcePolicy     string    `json:"source_policy,omitempty"`
	OriginalIntent   string    `json:"original_intent,omitempty"`
	StopCondition    string    `json:"stop_condition,omitempty"`
}

// InvestigationAnswerContract provides a deterministic final-answer draft when
// verified comparison evidence is complete. Consumers must preserve its values
// and units instead of independently rescaling provider metrics.
type InvestigationAnswerContract struct {
	FinalAnswerRecommended     bool   `json:"final_answer_recommended"`
	Reason                     string `json:"reason,omitempty"`
	NumericConsistencyRequired bool   `json:"numeric_consistency_required,omitempty"`
	FinalAnswerDraft           string `json:"final_answer_draft,omitempty"`
}

// InvestigationPayload is the high-level orchestration payload.
type InvestigationPayload struct {
	Tool               string                       `json:"tool,omitempty"`
	Source             string                       `json:"source,omitempty"`
	AdapterID          string                       `json:"adapter_id,omitempty"`
	AdapterStatus      AdapterStatus                `json:"adapter_status,omitempty"`
	FailureCode        FailureCode                  `json:"failure_code,omitempty"`
	Intent             InvestigationIntent          `json:"intent,omitempty"`
	Evidence           SourceEvidence               `json:"evidence,omitempty"`
	Readiness          Readiness                    `json:"readiness,omitempty"`
	Quote              *QuotePayload                `json:"quote,omitempty"`
	Profile            *ProfilePayload              `json:"profile,omitempty"`
	Announcement       *AnnouncementPayload         `json:"announcement,omitempty"`
	Research           *ResearchPayload             `json:"research,omitempty"`
	Signal             *SignalPayload               `json:"signal,omitempty"`
	Quotes             []QuotePayload               `json:"quotes,omitempty"`
	AnswerContract     *InvestigationAnswerContract `json:"answer_contract,omitempty"`
	Handoff            *ToolHandoff                 `json:"handoff,omitempty"`
	Warnings           []string                     `json:"warnings,omitempty"`
	IdentityResolution *IdentityResolution          `json:"identity_resolution,omitempty"`
}
