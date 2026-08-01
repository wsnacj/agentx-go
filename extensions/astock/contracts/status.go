package contracts

// AdapterStatus describes the evidence readiness reported by a data adapter.
type AdapterStatus string

const (
	AdapterStatusOK                 AdapterStatus = "ok"
	AdapterStatusNeedsReview        AdapterStatus = "needs_review"
	AdapterStatusEvidenceIncomplete AdapterStatus = "evidence_incomplete"
	AdapterStatusUnsupported        AdapterStatus = "unsupported"
	AdapterStatusUnavailable        AdapterStatus = "unavailable"
)

// FailureCode describes source-neutral failure categories for A-share tools.
type FailureCode string

const (
	FailureCodeNone                 FailureCode = ""
	FailureCodeIdentityNotFound     FailureCode = "identity_not_found"
	FailureCodeSourceUnavailable    FailureCode = "source_unavailable"
	FailureCodeProviderUnconfigured FailureCode = "provider_unconfigured"
	FailureCodeProviderUnavailable  FailureCode = "provider_unavailable"
	FailureCodeRateLimited          FailureCode = "rate_limited"
	FailureCodeTimeout              FailureCode = "timeout"
	FailureCodeStaleData            FailureCode = "stale_data"
	FailureCodeMissingFields        FailureCode = "missing_fields"
	FailureCodeReviewRequired       FailureCode = "review_required"
	FailureCodeUnsupported          FailureCode = "unsupported"
)

// TradingSession describes the market-time status of an A-share data payload.
type TradingSession string

const (
	TradingSessionUnknown      TradingSession = "unknown"
	TradingSessionPreMarket    TradingSession = "pre_market"
	TradingSessionTrading      TradingSession = "trading"
	TradingSessionMiddayBreak  TradingSession = "midday_break"
	TradingSessionClosed       TradingSession = "closed"
	TradingSessionNonTradeDate TradingSession = "non_trade_date"
)

// FreshnessMode describes whether data is realtime, intraday, EOD, historical, or cached.
type FreshnessMode string

const (
	FreshnessModeUnknown    FreshnessMode = "unknown"
	FreshnessModeRealtime   FreshnessMode = "realtime"
	FreshnessModeIntraday   FreshnessMode = "intraday"
	FreshnessModeEOD        FreshnessMode = "eod"
	FreshnessModeHistorical FreshnessMode = "historical"
	FreshnessModeCached     FreshnessMode = "cached"
)

// Freshness captures user-requested and source-confirmed time constraints.
type Freshness struct {
	Mode                    FreshnessMode  `json:"mode,omitempty"`
	RelativeDateHint        string         `json:"relative_date_hint,omitempty"`
	TradeDate               string         `json:"trade_date,omitempty"`
	AsOf                    string         `json:"as_of,omitempty"`
	TradingSession          TradingSession `json:"trading_session,omitempty"`
	RequireRealtime         bool           `json:"require_realtime,omitempty"`
	RequireLatestTradingDay bool           `json:"require_latest_trading_day,omitempty"`
	CacheStatus             string         `json:"cache_status,omitempty"`
}

// SourceEvidence captures source-level provenance common to all A-share tools.
type SourceEvidence struct {
	Provider       string         `json:"provider,omitempty"`
	Source         string         `json:"source,omitempty"`
	SourceURL      string         `json:"source_url,omitempty"`
	AsOf           string         `json:"as_of,omitempty"`
	TradingSession TradingSession `json:"trading_session,omitempty"`
	Freshness      FreshnessMode  `json:"freshness,omitempty"`
	Confidence     float64        `json:"confidence,omitempty"`
	Warnings       []string       `json:"warnings,omitempty"`
}

// ProviderAttempt records one attempted provider in a host-owned fallback chain.
type ProviderAttempt struct {
	Provider     string      `json:"provider,omitempty"`
	ProviderType string      `json:"provider_type,omitempty"`
	AdapterID    string      `json:"adapter_id,omitempty"`
	Status       string      `json:"status,omitempty"`
	FailureCode  FailureCode `json:"failure_code,omitempty"`
	Message      string      `json:"message,omitempty"`
	Credentialed bool        `json:"credentialed,omitempty"`
	Selected     bool        `json:"selected,omitempty"`
}

// Subject identifies an A-share security after adapter verification.
type Subject struct {
	EntityName string         `json:"entity_name,omitempty"`
	StockCode  string         `json:"stock_code,omitempty"`
	Market     Market         `json:"market,omitempty"`
	Verified   bool           `json:"verified,omitempty"`
	Evidence   SourceEvidence `json:"evidence,omitempty"`
}

// IdentityResolution records how a natural-language A-share subject was resolved.
// It is diagnostic-only; answer readiness still depends on source-backed payload fields.
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

// IdentityResolutionCandidate records one source-backed A-share candidate.
type IdentityResolutionCandidate struct {
	Name           string  `json:"name,omitempty"`
	Code           string  `json:"code,omitempty"`
	Market         Market  `json:"market,omitempty"`
	OrgID          string  `json:"org_id,omitempty"`
	Category       string  `json:"category,omitempty"`
	Type           string  `json:"type,omitempty"`
	SourceURL      string  `json:"source_url,omitempty"`
	QueryTerm      string  `json:"query_term,omitempty"`
	QueryReason    string  `json:"query_reason,omitempty"`
	QueryPriority  int     `json:"query_priority,omitempty"`
	Confidence     float64 `json:"confidence,omitempty"`
	Selected       bool    `json:"selected,omitempty"`
	SelectedReason string  `json:"selected_reason,omitempty"`
}

// Readiness summarizes whether a host may answer, must degrade, or should stop.
type Readiness struct {
	AnswerReady          bool          `json:"answer_ready"`
	Degraded             bool          `json:"degraded,omitempty"`
	DegradeReason        string        `json:"degrade_reason,omitempty"`
	AdapterStatus        AdapterStatus `json:"adapter_status,omitempty"`
	FailureCode          FailureCode   `json:"failure_code,omitempty"`
	RequestedFieldsReady bool          `json:"requested_fields_ready"`
	MissingFields        []string      `json:"missing_fields,omitempty"`
	ReviewRequiredFields []string      `json:"review_required_fields,omitempty"`
	NextRepairHint       string        `json:"next_repair_hint,omitempty"`
}

// BuildReadiness projects source-neutral adapter/guard fields into answer readiness.
func BuildReadiness(status AdapterStatus, failure FailureCode, requestedFieldsReady bool, missingFields []string, reviewFields []string) Readiness {
	ready := status == AdapterStatusOK && failure == FailureCodeNone && requestedFieldsReady && len(missingFields) == 0 && len(reviewFields) == 0
	out := Readiness{
		AnswerReady:          ready,
		AdapterStatus:        status,
		FailureCode:          failure,
		RequestedFieldsReady: requestedFieldsReady,
		MissingFields:        append([]string(nil), missingFields...),
		ReviewRequiredFields: append([]string(nil), reviewFields...),
	}
	if ready {
		return out
	}
	out.Degraded = canDegrade(status, failure)
	out.DegradeReason = degradeReason(status, failure, requestedFieldsReady, missingFields, reviewFields)
	out.NextRepairHint = repairHint(failure, missingFields, reviewFields)
	return out
}

func canDegrade(status AdapterStatus, failure FailureCode) bool {
	if failure == FailureCodeIdentityNotFound || failure == FailureCodeProviderUnconfigured || failure == FailureCodeUnsupported {
		return false
	}
	return status == AdapterStatusNeedsReview || status == AdapterStatusEvidenceIncomplete || failure == FailureCodeMissingFields || failure == FailureCodeReviewRequired || failure == FailureCodeStaleData
}

func degradeReason(status AdapterStatus, failure FailureCode, requestedFieldsReady bool, missingFields []string, reviewFields []string) string {
	switch {
	case failure != FailureCodeNone:
		return string(failure)
	case !requestedFieldsReady:
		return "requested_fields_not_ready"
	case len(missingFields) > 0:
		return "missing_fields"
	case len(reviewFields) > 0:
		return "review_required"
	case status != AdapterStatusOK:
		return string(status)
	default:
		return "not_answer_ready"
	}
}

func repairHint(failure FailureCode, missingFields []string, reviewFields []string) string {
	switch failure {
	case FailureCodeIdentityNotFound:
		return "resolve_verified_a_share_subject"
	case FailureCodeProviderUnconfigured, FailureCodeProviderUnavailable:
		return "configure_or_switch_data_provider"
	case FailureCodeRateLimited, FailureCodeTimeout, FailureCodeSourceUnavailable:
		return "retry_or_switch_source"
	case FailureCodeStaleData:
		return "refresh_latest_trading_day_source"
	case FailureCodeUnsupported:
		return "use_supported_a_stock_task_or_adapter"
	}
	if len(missingFields) > 0 {
		return "fetch_missing_fields"
	}
	if len(reviewFields) > 0 {
		return "verify_review_required_fields"
	}
	return ""
}
