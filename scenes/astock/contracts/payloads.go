package contracts

import (
	"bytes"
	"encoding/json"
	"strings"
)

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

// UnmarshalJSON accepts both the canonical metric object and scalar values.
// LLM-mediated formatting calls may compact a metric to "123.4"; preserving it
// as Value is safer than failing an otherwise verified answer-format pass.
func (m *MetricValue) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*m = MetricValue{}
		return nil
	}
	if trimmed[0] == '{' {
		type alias MetricValue
		var value alias
		if err := json.Unmarshal(trimmed, &value); err != nil {
			return err
		}
		*m = MetricValue(value)
		return nil
	}
	var scalar any
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	if err := decoder.Decode(&scalar); err != nil {
		return err
	}
	switch value := scalar.(type) {
	case string:
		m.Value = strings.TrimSpace(value)
	case json.Number:
		m.Value = value.String()
	case bool:
		if value {
			m.Value = "true"
		} else {
			m.Value = "false"
		}
	default:
		m.Value = strings.TrimSpace(string(trimmed))
	}
	return nil
}

// TaskKind describes the high-level A-share task frame selected by the model.
type TaskKind string

const (
	TaskKindQuoteSnapshot     TaskKind = "quote_snapshot"
	TaskKindValuationSnapshot TaskKind = "valuation_snapshot"
	TaskKindResearchLookup    TaskKind = "research_lookup"
	TaskKindSignalLookup      TaskKind = "signal_lookup"
	TaskKindAnnouncement      TaskKind = "announcement_lookup"
	TaskKindProfileLookup     TaskKind = "profile_lookup"
	TaskKindScreening         TaskKind = "screening"
	TaskKindComparison        TaskKind = "comparison"
	TaskKindFullInvestigation TaskKind = "full_investigation"
)

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

// QuoteSnapshot contains common A-share quote and valuation fields.
type QuoteSnapshot struct {
	Price           MetricValue `json:"price,omitempty"`
	LastClose       MetricValue `json:"last_close,omitempty"`
	Open            MetricValue `json:"open,omitempty"`
	High            MetricValue `json:"high,omitempty"`
	Low             MetricValue `json:"low,omitempty"`
	ChangeAmount    MetricValue `json:"change_amount,omitempty"`
	ChangePercent   MetricValue `json:"change_percent,omitempty"`
	TurnoverPercent MetricValue `json:"turnover_percent,omitempty"`
	PETTM           MetricValue `json:"pe_ttm,omitempty"`
	PEStatic        MetricValue `json:"pe_static,omitempty"`
	PB              MetricValue `json:"pb,omitempty"`
	MarketCap       MetricValue `json:"market_cap,omitempty"`
	FloatMarketCap  MetricValue `json:"float_market_cap,omitempty"`
	LimitUp         MetricValue `json:"limit_up,omitempty"`
	LimitDown       MetricValue `json:"limit_down,omitempty"`
	VolumeRatio     MetricValue `json:"volume_ratio,omitempty"`
	OrderBook       *OrderBook  `json:"order_book,omitempty"`
	KLines          []KLine     `json:"klines,omitempty"`
	Transactions    []TradeTick `json:"transactions,omitempty"`
}

// OrderBook contains level-1 to level-5 bid/ask evidence when available.
type OrderBook struct {
	Bids []OrderBookLevel `json:"bids,omitempty"`
	Asks []OrderBookLevel `json:"asks,omitempty"`
	AsOf string           `json:"as_of,omitempty"`
}

// OrderBookLevel contains one bid/ask level.
type OrderBookLevel struct {
	Level  int         `json:"level,omitempty"`
	Price  MetricValue `json:"price,omitempty"`
	Volume MetricValue `json:"volume,omitempty"`
}

// KLine contains one OHLCV bar.
type KLine struct {
	Period string      `json:"period,omitempty"`
	Time   string      `json:"time,omitempty"`
	Open   MetricValue `json:"open,omitempty"`
	Close  MetricValue `json:"close,omitempty"`
	High   MetricValue `json:"high,omitempty"`
	Low    MetricValue `json:"low,omitempty"`
	Volume MetricValue `json:"volume,omitempty"`
	Amount MetricValue `json:"amount,omitempty"`
}

// TradeTick contains one transaction tick.
type TradeTick struct {
	Time      string      `json:"time,omitempty"`
	Price     MetricValue `json:"price,omitempty"`
	Volume    MetricValue `json:"volume,omitempty"`
	Direction string      `json:"direction,omitempty"`
}

// ResearchPayload contains research report and consensus forecast evidence.
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
	IdentityResolution *IdentityResolution `json:"identity_resolution,omitempty"`
}

// ResearchReport contains one source-backed report record.
type ResearchReport struct {
	Title       string         `json:"title,omitempty"`
	Institution string         `json:"institution,omitempty"`
	Analyst     string         `json:"analyst,omitempty"`
	PublishedAt string         `json:"published_at,omitempty"`
	Rating      string         `json:"rating,omitempty"`
	Summary     string         `json:"summary,omitempty"`
	PDFURL      string         `json:"pdf_url,omitempty"`
	SourceURL   string         `json:"source_url,omitempty"`
	Forecasts   []ForecastItem `json:"forecasts,omitempty"`
	Evidence    SourceEvidence `json:"evidence,omitempty"`
}

// ForecastItem contains EPS/profit/target-price forecast evidence.
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

// SignalPayload contains hot-topic, flow, board, lockup, and industry evidence.
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
	IdentityResolution *IdentityResolution `json:"identity_resolution,omitempty"`
}

// SignalType describes supported A-share signal categories.
type SignalType string

const (
	SignalTypeHotReason        SignalType = "hot_reason"
	SignalTypeConceptBlocks    SignalType = "concept_blocks"
	SignalTypeFundFlow         SignalType = "fund_flow"
	SignalTypeNorthboundFlow   SignalType = "northbound_flow"
	SignalTypeDragonTigerBoard SignalType = "dragon_tiger_board"
	SignalTypeDailyBillboard   SignalType = "daily_dragon_tiger"
	SignalTypeLockupExpiry     SignalType = "lockup_expiry"
	SignalTypeIndustryCompare  SignalType = "industry_comparison"
)

// SignalEvent contains one source-backed market signal.
type SignalEvent struct {
	Type        SignalType     `json:"type,omitempty"`
	Title       string         `json:"title,omitempty"`
	Reason      string         `json:"reason,omitempty"`
	TradeDate   string         `json:"trade_date,omitempty"`
	Rank        int            `json:"rank,omitempty"`
	Industry    string         `json:"industry,omitempty"`
	Concept     string         `json:"concept,omitempty"`
	NetBuy      MetricValue    `json:"net_buy,omitempty"`
	Amount      MetricValue    `json:"amount,omitempty"`
	Ratio       MetricValue    `json:"ratio,omitempty"`
	SourceURL   string         `json:"source_url,omitempty"`
	Evidence    SourceEvidence `json:"evidence,omitempty"`
	ReviewNotes []string       `json:"review_notes,omitempty"`
}

// AnnouncementPayload contains source-backed announcement records.
type AnnouncementPayload struct {
	Tool               string              `json:"tool,omitempty"`
	Source             string              `json:"source,omitempty"`
	AdapterID          string              `json:"adapter_id,omitempty"`
	AdapterStatus      AdapterStatus       `json:"adapter_status,omitempty"`
	FailureCode        FailureCode         `json:"failure_code,omitempty"`
	Subject            Subject             `json:"subject,omitempty"`
	Freshness          Freshness           `json:"freshness,omitempty"`
	Evidence           SourceEvidence      `json:"evidence,omitempty"`
	Readiness          Readiness           `json:"readiness,omitempty"`
	Announcements      []Announcement      `json:"announcements,omitempty"`
	Warnings           []string            `json:"warnings,omitempty"`
	IdentityResolution *IdentityResolution `json:"identity_resolution,omitempty"`
}

// Announcement contains one CNINFO/F10-style disclosure record.
type Announcement struct {
	Title             string         `json:"title,omitempty"`
	Type              string         `json:"type,omitempty"`
	PublishedAt       string         `json:"published_at,omitempty"`
	URL               string         `json:"url,omitempty"`
	FullTextAvailable bool           `json:"full_text_available,omitempty"`
	Summary           string         `json:"summary,omitempty"`
	Evidence          SourceEvidence `json:"evidence,omitempty"`
}

// ProfilePayload contains A-share company profile and basic finance snapshot evidence.
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

// CompanyProfile contains source-backed company metadata.
type CompanyProfile struct {
	Name         string            `json:"name,omitempty"`
	Industry     string            `json:"industry,omitempty"`
	Region       string            `json:"region,omitempty"`
	ListingDate  string            `json:"listing_date,omitempty"`
	ShareCapital MetricValue       `json:"share_capital,omitempty"`
	MarketCap    MetricValue       `json:"market_cap,omitempty"`
	F10Sections  map[string]string `json:"f10_sections,omitempty"`
	Evidence     SourceEvidence    `json:"evidence,omitempty"`
}

// InvestigationIntent captures the model-supplied structured A-share task frame.
type InvestigationIntent struct {
	UserMessage      string    `json:"user_message,omitempty"`
	TaskKind         TaskKind  `json:"task_kind,omitempty"`
	EntityName       string    `json:"entity_name,omitempty"`
	EntityMentions   []string  `json:"entity_mentions,omitempty"`
	StockCode        string    `json:"stock_code,omitempty"`
	Market           Market    `json:"market,omitempty"`
	RequestedFields  []string  `json:"requested_fields,omitempty"`
	RequestedOutputs []string  `json:"requested_outputs,omitempty"`
	Assessment       string    `json:"assessment,omitempty"`
	SourceHint       string    `json:"source_hint,omitempty"`
	SourcePolicy     string    `json:"source_policy,omitempty"`
	Freshness        Freshness `json:"freshness,omitempty"`
}

// InvestigationPayload aggregates task-level A-share evidence for high-level flows.
type InvestigationPayload struct {
	Tool               string               `json:"tool,omitempty"`
	Source             string               `json:"source,omitempty"`
	AdapterStatus      AdapterStatus        `json:"adapter_status,omitempty"`
	FailureCode        FailureCode          `json:"failure_code,omitempty"`
	Intent             InvestigationIntent  `json:"intent,omitempty"`
	Evidence           SourceEvidence       `json:"evidence,omitempty"`
	Readiness          Readiness            `json:"readiness,omitempty"`
	Quote              *QuotePayload        `json:"quote,omitempty"`
	Quotes             []QuotePayload       `json:"quotes,omitempty"`
	Research           *ResearchPayload     `json:"research,omitempty"`
	Signal             *SignalPayload       `json:"signal,omitempty"`
	Announcement       *AnnouncementPayload `json:"announcement,omitempty"`
	Profile            *ProfilePayload      `json:"profile,omitempty"`
	Warnings           []string             `json:"warnings,omitempty"`
	Diagnostics        []string             `json:"diagnostics,omitempty"`
	IdentityResolution *IdentityResolution  `json:"identity_resolution,omitempty"`
}
