package publicnews

import "strings"

// LatestNewsLookupIntent is the structured task frame supplied by the model for
// high-level latest-news requests. It is intent only; host adapters must verify
// source freshness, source independence, and grounded facts.
type LatestNewsLookupIntent struct {
	UserMessage       string         `json:"user_message,omitempty"`
	TaskKind          string         `json:"task_kind,omitempty"`
	Topic             string         `json:"topic,omitempty"`
	EntityMentions    []string       `json:"entity_mentions,omitempty"`
	RequestedFields   []string       `json:"requested_fields,omitempty"`
	RequestedOutputs  []string       `json:"requested_outputs,omitempty"`
	Freshness         map[string]any `json:"freshness,omitempty"`
	SourceHint        string         `json:"source_hint,omitempty"`
	SourcePolicy      string         `json:"source_policy,omitempty"`
	CrossCheckPolicy  string         `json:"cross_check_policy,omitempty"`
	OriginalIntent    string         `json:"original_intent,omitempty"`
	StopCondition     string         `json:"stop_condition,omitempty"`
	NeedsSourceVerify bool           `json:"needs_source_verify,omitempty"`
}

type LatestNewsLookupSource struct {
	PageID      string `json:"page_id,omitempty"`
	Title       string `json:"title,omitempty"`
	SourceURL   string `json:"source_url,omitempty"`
	Text        string `json:"text,omitempty"`
	SourceSite  string `json:"source_site,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
	KeyUpdate   string `json:"key_update,omitempty"`
	Headline    string `json:"headline,omitempty"`
}

type LatestNewsSearchQuery struct {
	Query     string `json:"query,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Language  string `json:"language,omitempty"`
	Freshness string `json:"freshness,omitempty"`
	Priority  int    `json:"priority,omitempty"`
}

type LatestNewsSourcesPayload struct {
	Tool                        string                   `json:"tool,omitempty"`
	Source                      string                   `json:"source,omitempty"`
	PackID                      string                   `json:"pack_id,omitempty"`
	CaseType                    string                   `json:"case_type,omitempty"`
	WorkflowID                  string                   `json:"workflow_id,omitempty"`
	AdapterID                   string                   `json:"adapter_id,omitempty"`
	AdapterStatus               string                   `json:"adapter_status,omitempty"`
	FailureCode                 string                   `json:"failure_code,omitempty"`
	FailureClass                string                   `json:"failure_class,omitempty"`
	ErrorClass                  string                   `json:"error_class,omitempty"`
	Provider                    string                   `json:"provider,omitempty"`
	RequestedProvider           string                   `json:"requested_provider,omitempty"`
	EffectiveProvider           string                   `json:"effective_provider,omitempty"`
	ProviderStatus              string                   `json:"provider_status,omitempty"`
	AvailableProviders          []string                 `json:"available_providers,omitempty"`
	UnavailableProviders        []ProviderStatus         `json:"unavailable_providers,omitempty"`
	FallbackProvider            string                   `json:"fallback_provider,omitempty"`
	FallbackHint                string                   `json:"fallback_hint,omitempty"`
	Retryable                   bool                     `json:"retryable,omitempty"`
	RetrySuppressedReason       string                   `json:"retry_suppressed_reason,omitempty"`
	RetryAttempts               []RetryAttempt           `json:"retry_attempts,omitempty"`
	RetryAttemptCount           int                      `json:"retry_attempt_count,omitempty"`
	RetryExhausted              bool                     `json:"retry_exhausted,omitempty"`
	SearchQueries               []LatestNewsSearchQuery  `json:"search_queries,omitempty"`
	SearchQueryCount            int                      `json:"search_query_count,omitempty"`
	SearchAttemptCount          int                      `json:"search_attempt_count,omitempty"`
	PrimarySearchAttemptCount   int                      `json:"primary_search_attempt_count,omitempty"`
	AlternateSearchAttemptCount int                      `json:"alternate_search_attempt_count,omitempty"`
	CandidateSearchAttemptCount int                      `json:"candidate_search_attempt_count,omitempty"`
	UserMessage                 string                   `json:"user_message,omitempty"`
	Intent                      LatestNewsLookupIntent   `json:"intent,omitempty"`
	PrimarySource               LatestNewsLookupSource   `json:"primary_source,omitempty"`
	SupportingSources           []LatestNewsLookupSource `json:"supporting_sources,omitempty"`
	Warnings                    []string                 `json:"warnings,omitempty"`
}

type ProviderStatus struct {
	Provider    string   `json:"provider,omitempty"`
	Available   bool     `json:"available,omitempty"`
	Reason      string   `json:"reason,omitempty"`
	RequiresEnv []string `json:"requires_env,omitempty"`
}

type RetryAttempt struct {
	Attempt        int    `json:"attempt,omitempty"`
	AdapterStatus  string `json:"adapter_status,omitempty"`
	FailureCode    string `json:"failure_code,omitempty"`
	FailureClass   string `json:"failure_class,omitempty"`
	Provider       string `json:"provider,omitempty"`
	ProviderStatus string `json:"provider_status,omitempty"`
	Retryable      bool   `json:"retryable,omitempty"`
}

type LatestNewsLookupPayload struct {
	Tool                  string                     `json:"tool"`
	Source                string                     `json:"source"`
	PackID                string                     `json:"pack_id,omitempty"`
	CaseType              string                     `json:"case_type,omitempty"`
	WorkflowID            string                     `json:"workflow_id,omitempty"`
	AdapterID             string                     `json:"adapter_id,omitempty"`
	AdapterStatus         string                     `json:"adapter_status,omitempty"`
	FailureCode           string                     `json:"failure_code"`
	FailureClass          string                     `json:"failure_class,omitempty"`
	GuardStatus           string                     `json:"guard_status,omitempty"`
	NewsFieldsReady       bool                       `json:"news_fields_ready"`
	CrossCheckReady       bool                       `json:"cross_check_ready"`
	Passed                bool                       `json:"passed"`
	FreshnessConfirmed    bool                       `json:"freshness_confirmed"`
	SourceAccepted        bool                       `json:"source_accepted"`
	StopAfterGuardPassed  bool                       `json:"stop_after_guard_passed"`
	MissingNewsFields     []string                   `json:"missing_news_fields"`
	ReviewReasons         []string                   `json:"review_reasons"`
	SourceURL             string                     `json:"source_url"`
	PublishedAt           string                     `json:"published_at"`
	Summary               string                     `json:"summary"`
	Intent                LatestNewsLookupIntent     `json:"intent"`
	Sources               *LatestNewsSourcesPayload  `json:"sources,omitempty"`
	Extract               *Payload                   `json:"extract,omitempty"`
	Guard                 *Payload                   `json:"guard,omitempty"`
	EvidenceReview        *EvidenceReviewResult      `json:"evidence_review,omitempty"`
	EvaluatorReport       *LatestNewsEvaluatorReport `json:"evaluator_report,omitempty"`
	AnswerReadiness       LatestNewsAnswerReadiness  `json:"answer_readiness,omitempty"`
	AnswerContract        *LatestNewsAnswerContract  `json:"answer_contract,omitempty"`
	Warnings              []string                   `json:"warnings,omitempty"`
	RetryAttempts         []RetryAttempt             `json:"retry_attempts,omitempty"`
	RetryAttemptCount     int                        `json:"retry_attempt_count,omitempty"`
	RetryExhausted        bool                       `json:"retry_exhausted,omitempty"`
	RetrySuppressedReason string                     `json:"retry_suppressed_reason,omitempty"`
}

type LatestNewsAnswerReadiness struct {
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

// LatestNewsAnswerContract is a runtime-facing final-answer handoff. It is
// emitted only by trusted host/tool code, never inferred from page text.
type LatestNewsAnswerContract struct {
	FinalAnswerRecommended bool     `json:"final_answer_recommended"`
	Reason                 string   `json:"reason,omitempty"`
	AllowedSummaryScope    string   `json:"allowed_summary_scope,omitempty"`
	DoNotRetryTools        []string `json:"do_not_retry_tools,omitempty"`
	PossibleImpact         string   `json:"possible_impact,omitempty"`
	RiskBoundary           string   `json:"risk_boundary,omitempty"`
	FinalAnswerDraft       string   `json:"final_answer_draft,omitempty"`
}

func LatestNewsLookupIntentFromParams(params map[string]any) LatestNewsLookupIntent {
	intent := LatestNewsLookupIntent{
		UserMessage:      strings.TrimSpace(StringArg(params["user_message"])),
		TaskKind:         strings.TrimSpace(StringArg(params["task_kind"])),
		Topic:            strings.TrimSpace(StringArg(params["topic"])),
		EntityMentions:   StringListArg(params["entity_mentions"]),
		RequestedFields:  normalizeStringList(StringListArg(params["requested_fields"])),
		RequestedOutputs: normalizeStringList(StringListArg(params["requested_outputs"])),
		Freshness:        objectArg(params["freshness"]),
		SourceHint:       strings.TrimSpace(StringArg(params["source_hint"])),
		SourcePolicy:     strings.TrimSpace(StringArg(params["source_policy"])),
		CrossCheckPolicy: strings.TrimSpace(StringArg(params["cross_check_policy"])),
		OriginalIntent:   strings.TrimSpace(StringArg(params["original_intent"])),
		StopCondition:    strings.TrimSpace(StringArg(params["stop_condition"])),
	}
	if intent.UserMessage == "" {
		intent.UserMessage = intent.OriginalIntent
	}
	if intent.OriginalIntent == "" {
		intent.OriginalIntent = intent.UserMessage
	}
	if intent.TaskKind == "" {
		intent.TaskKind = "latest_news_brief"
	}
	if len(intent.RequestedFields) == 0 {
		intent.RequestedFields = []string{"headline", "published_at", "key_update", "source_url"}
	}
	if len(intent.RequestedOutputs) == 0 {
		intent.RequestedOutputs = []string{"brief", "source_verification"}
	}
	return intent
}

func ContextFromLookupSource(source LatestNewsLookupSource, userMessage string) Context {
	return Context{
		UserMessage: strings.TrimSpace(userMessage),
		PageID:      strings.TrimSpace(source.PageID),
		Title:       firstNonEmpty(source.Title, source.Headline),
		SourceURL:   strings.TrimSpace(source.SourceURL),
		Text:        strings.TrimSpace(source.Text),
	}
}

func LookupSourceFromContext(ctx Context) LatestNewsLookupSource {
	return LatestNewsLookupSource{
		PageID:    strings.TrimSpace(ctx.PageID),
		Title:     strings.TrimSpace(ctx.Title),
		SourceURL: strings.TrimSpace(ctx.SourceURL),
		Text:      strings.TrimSpace(ctx.Text),
	}
}

func LookupSourceFromParams(params map[string]any) LatestNewsLookupSource {
	return LatestNewsLookupSource{
		PageID:      strings.TrimSpace(StringArg(params["page_id"])),
		Title:       strings.TrimSpace(StringArg(params["title"])),
		SourceURL:   strings.TrimSpace(StringArg(params["source_url"])),
		Text:        strings.TrimSpace(StringArg(params["text"])),
		SourceSite:  strings.TrimSpace(StringArg(params["source_site"])),
		PublishedAt: strings.TrimSpace(StringArg(params["published_at"])),
		KeyUpdate:   strings.TrimSpace(StringArg(params["key_update"])),
		Headline:    strings.TrimSpace(StringArg(params["headline"])),
	}
}

func StringListArg(raw any) []string {
	if typed, ok := raw.([]string); ok {
		return normalizeStringList(typed)
	}
	values, ok := anySlice(raw)
	if !ok {
		if value := strings.TrimSpace(StringArg(raw)); value != "" {
			return []string{value}
		}
		return nil
	}
	out := make([]string, 0, len(values))
	for _, item := range values {
		if value := strings.TrimSpace(StringArg(item)); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func normalizeStringList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

func objectArg(raw any) map[string]any {
	switch value := raw.(type) {
	case map[string]any:
		return value
	default:
		return nil
	}
}
