package hostkit

import (
	"context"
	"fmt"
	"strings"

	astockcontracts "github.com/wsnacj/agentx-go/extensions/astock/contracts"
	agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
)

// ToolPayloadHandler is the provider-neutral handler shape used by Host wiring.
type ToolPayloadHandler func(context.Context, map[string]any) (any, error)

type InvestigationConfig struct {
	Source              string
	SourcePolicyDefault string
	Handlers            InvestigationHandlers
}

type InvestigationHandlers struct {
	Quote        func(context.Context, map[string]any) (astockcontracts.QuotePayload, error)
	Research     func(context.Context, map[string]any) (astockcontracts.ResearchPayload, error)
	Signal       func(context.Context, map[string]any) (astockcontracts.SignalPayload, error)
	Announcement func(context.Context, map[string]any) (astockcontracts.AnnouncementPayload, error)
	Profile      func(context.Context, map[string]any) (astockcontracts.ProfilePayload, error)
}

// BuildAStockInvestigationHandler builds the high-level tool handler used by hosts.
func BuildAStockInvestigationHandler(cfg InvestigationConfig) ToolPayloadHandler {
	return func(ctx context.Context, params map[string]any) (any, error) {
		return BuildAStockInvestigationPayload(ctx, cfg, params)
	}
}

// BuildAStockInvestigationPayload dispatches the structured task frame to host-owned handlers.
func BuildAStockInvestigationPayload(ctx context.Context, cfg InvestigationConfig, params map[string]any) (astockcontracts.InvestigationPayload, error) {
	intent := IntentFromParams(params)
	if intent.SourcePolicy == "" {
		intent.SourcePolicy = cfg.SourcePolicyDefault
	}
	steps := plannedSteps(intent)
	taskParams := ParamsFromIntent(params, intent)
	payload := astockcontracts.InvestigationPayload{
		Tool:   astockcontracts.ToolAStockInvestigation,
		Source: firstNonEmpty(cfg.Source, "agentx_a_stock_hostkit"),
		Intent: intent,
		Evidence: astockcontracts.SourceEvidence{
			Provider:   "agentx_a_stock",
			Source:     firstNonEmpty(cfg.Source, "agentx_a_stock_hostkit"),
			Freshness:  intent.Freshness.Mode,
			AsOf:       intent.Freshness.AsOf,
			Confidence: 0.7,
		},
	}
	if len(steps) == 0 {
		payload.AdapterStatus = astockcontracts.AdapterStatusUnsupported
		payload.FailureCode = astockcontracts.FailureCodeUnsupported
		payload.Readiness = astockcontracts.BuildReadiness(payload.AdapterStatus, payload.FailureCode, false, nil, nil)
		payload.Warnings = append(payload.Warnings, "a_stock_investigation_no_supported_task_kind")
		return payload, nil
	}
	if isMultiEntityQuoteComparison(intent) {
		return runMultiEntityQuoteComparison(ctx, cfg, payload, taskParams, intent)
	}
	readiness := []astockcontracts.Readiness{}
	for _, step := range steps {
		switch step {
		case astockcontracts.TaskKindQuoteSnapshot, astockcontracts.TaskKindValuationSnapshot:
			if cfg.Handlers.Quote == nil {
				payload.Warnings = append(payload.Warnings, "a_stock_quote_lookup_handler_missing")
				readiness = append(readiness, missingReadiness())
				continue
			}
			quote, err := cfg.Handlers.Quote(ctx, taskParams)
			if err != nil {
				payload.Warnings = append(payload.Warnings, "a_stock_quote_lookup_error:"+aStockInvestigationError(err, "quote_lookup_failed"))
				readiness = append(readiness, sourceUnavailableReadiness())
				continue
			}
			payload.Quote = &quote
			projectIdentityResolution(&payload, quote.IdentityResolution)
			readiness = append(readiness, quote.Readiness)
		case astockcontracts.TaskKindResearchLookup:
			if cfg.Handlers.Research == nil {
				payload.Warnings = append(payload.Warnings, "a_stock_research_lookup_handler_missing")
				readiness = append(readiness, missingReadiness())
				continue
			}
			research, err := cfg.Handlers.Research(ctx, taskParams)
			if err != nil {
				payload.Warnings = append(payload.Warnings, "a_stock_research_lookup_error:"+aStockInvestigationError(err, "research_lookup_failed"))
				readiness = append(readiness, sourceUnavailableReadiness())
				continue
			}
			payload.Research = &research
			projectIdentityResolution(&payload, research.IdentityResolution)
			readiness = append(readiness, research.Readiness)
		case astockcontracts.TaskKindSignalLookup:
			signal, err := runSignal(ctx, cfg, taskParams)
			if err != nil {
				payload.Warnings = append(payload.Warnings, "a_stock_signal_lookup_error:"+aStockInvestigationError(err, "signal_lookup_failed"))
				readiness = append(readiness, sourceUnavailableReadiness())
				continue
			}
			payload.Signal = &signal
			projectIdentityResolution(&payload, signal.IdentityResolution)
			readiness = append(readiness, signal.Readiness)
		case astockcontracts.TaskKindAnnouncement:
			if cfg.Handlers.Announcement == nil {
				payload.Warnings = append(payload.Warnings, "a_stock_announcement_lookup_handler_missing")
				readiness = append(readiness, missingReadiness())
				continue
			}
			announcement, err := cfg.Handlers.Announcement(ctx, taskParams)
			if err != nil {
				payload.Warnings = append(payload.Warnings, "a_stock_announcement_lookup_error:"+aStockInvestigationError(err, "announcement_lookup_failed"))
				readiness = append(readiness, sourceUnavailableReadiness())
				continue
			}
			payload.Announcement = &announcement
			projectIdentityResolution(&payload, announcement.IdentityResolution)
			readiness = append(readiness, announcement.Readiness)
		case astockcontracts.TaskKindProfileLookup:
			if cfg.Handlers.Profile == nil {
				payload.Warnings = append(payload.Warnings, "a_stock_profile_lookup_handler_missing")
				readiness = append(readiness, missingReadiness())
				continue
			}
			profile, err := cfg.Handlers.Profile(ctx, taskParams)
			if err != nil {
				payload.Warnings = append(payload.Warnings, "a_stock_profile_lookup_error:"+aStockInvestigationError(err, "profile_lookup_failed"))
				readiness = append(readiness, sourceUnavailableReadiness())
				continue
			}
			payload.Profile = &profile
			projectIdentityResolution(&payload, profile.IdentityResolution)
			readiness = append(readiness, profile.Readiness)
		}
	}
	payload.AdapterStatus, payload.FailureCode, payload.Readiness = aggregateReadiness(readiness)
	return payload, nil
}

func plannedSteps(intent astockcontracts.InvestigationIntent) []astockcontracts.TaskKind {
	explicitSteps := stepsFromRequested(intent)
	switch intent.TaskKind {
	case astockcontracts.TaskKindQuoteSnapshot, astockcontracts.TaskKindValuationSnapshot:
		return []astockcontracts.TaskKind{intent.TaskKind}
	case astockcontracts.TaskKindResearchLookup:
		return []astockcontracts.TaskKind{astockcontracts.TaskKindResearchLookup}
	case astockcontracts.TaskKindSignalLookup:
		return []astockcontracts.TaskKind{astockcontracts.TaskKindSignalLookup}
	case astockcontracts.TaskKindAnnouncement:
		return []astockcontracts.TaskKind{astockcontracts.TaskKindAnnouncement}
	case astockcontracts.TaskKindProfileLookup:
		return []astockcontracts.TaskKind{astockcontracts.TaskKindProfileLookup}
	case astockcontracts.TaskKindComparison, astockcontracts.TaskKindScreening:
		if len(explicitSteps) > 0 {
			return explicitSteps
		}
		return []astockcontracts.TaskKind{astockcontracts.TaskKindValuationSnapshot}
	case astockcontracts.TaskKindFullInvestigation, "":
		if len(explicitSteps) > 0 {
			return explicitSteps
		}
		return []astockcontracts.TaskKind{
			astockcontracts.TaskKindQuoteSnapshot,
			astockcontracts.TaskKindProfileLookup,
			astockcontracts.TaskKindResearchLookup,
			astockcontracts.TaskKindAnnouncement,
		}
	default:
		return nil
	}
}

func runMultiEntityQuoteComparison(ctx context.Context, cfg InvestigationConfig, payload astockcontracts.InvestigationPayload, taskParams map[string]any, intent astockcontracts.InvestigationIntent) (astockcontracts.InvestigationPayload, error) {
	if cfg.Handlers.Quote == nil {
		payload.Warnings = append(payload.Warnings, "a_stock_quote_lookup_handler_missing")
		payload.AdapterStatus = astockcontracts.AdapterStatusUnsupported
		payload.FailureCode = astockcontracts.FailureCodeUnsupported
		payload.Readiness = astockcontracts.BuildReadiness(payload.AdapterStatus, payload.FailureCode, false, []string{"quotes"}, nil)
		return payload, nil
	}
	readiness := []astockcontracts.Readiness{}
	for _, mention := range uniqueEntityMentions(intent) {
		quote, err := cfg.Handlers.Quote(ctx, paramsForEntityMention(taskParams, mention))
		if err != nil {
			payload.Warnings = append(payload.Warnings, "a_stock_quote_lookup_error:"+mention+":"+aStockInvestigationError(err, "quote_lookup_failed"))
			readiness = append(readiness, sourceUnavailableReadiness())
			continue
		}
		payload.Quotes = append(payload.Quotes, quote)
		readiness = append(readiness, quote.Readiness)
	}
	if len(payload.Quotes) == 0 && len(readiness) == 0 {
		readiness = append(readiness, missingReadiness())
	}
	payload.AdapterStatus, payload.FailureCode, payload.Readiness = aggregateReadiness(readiness)
	return payload, nil
}

func aStockInvestigationError(err error, code string) string {
	return agentxsafeerror.Summary(agentxsafeerror.Project(err, "a_stock_investigation", code))
}

func isMultiEntityQuoteComparison(intent astockcontracts.InvestigationIntent) bool {
	if len(uniqueEntityMentions(intent)) < 2 {
		return false
	}
	if intent.TaskKind == astockcontracts.TaskKindComparison || intent.TaskKind == astockcontracts.TaskKindScreening {
		return true
	}
	for _, output := range intent.RequestedOutputs {
		if matchAny(output, "comparison", "quote", "valuation", "valuation_snapshot") {
			return true
		}
	}
	for _, field := range intent.RequestedFields {
		if quoteField(field) {
			return true
		}
	}
	return false
}

func uniqueEntityMentions(intent astockcontracts.InvestigationIntent) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range intent.EntityMentions {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	if len(out) == 0 && strings.TrimSpace(intent.EntityName) != "" {
		out = append(out, strings.TrimSpace(intent.EntityName))
	}
	return out
}

func paramsForEntityMention(base map[string]any, mention string) map[string]any {
	out := map[string]any{}
	for key, value := range base {
		out[key] = value
	}
	delete(out, "stock_code")
	delete(out, "market")
	normalized, market, ok := astockcontracts.NormalizeAStockCode(mention)
	if ok {
		out["stock_code"] = normalized
		if market != "" {
			out["market"] = string(market)
		}
		delete(out, "entity_name")
		return out
	}
	out["entity_name"] = strings.TrimSpace(mention)
	return out
}

func stepsFromRequested(intent astockcontracts.InvestigationIntent) []astockcontracts.TaskKind {
	steps := []astockcontracts.TaskKind{}
	for _, output := range intent.RequestedOutputs {
		switch {
		case matchAny(output, "quote", "valuation", "valuation_snapshot"):
			steps = appendUniqueTask(steps, astockcontracts.TaskKindValuationSnapshot)
		case matchAny(output, "research", "research_reports", "rating"):
			steps = appendUniqueTask(steps, astockcontracts.TaskKindResearchLookup)
		case matchAny(output, "announcement", "announcements"):
			steps = appendUniqueTask(steps, astockcontracts.TaskKindAnnouncement)
		case matchAny(output, "signal", "signals"):
			steps = appendUniqueTask(steps, astockcontracts.TaskKindSignalLookup)
		case matchAny(output, "profile", "company_profile"):
			steps = appendUniqueTask(steps, astockcontracts.TaskKindProfileLookup)
		}
	}
	for _, field := range intent.RequestedFields {
		switch {
		case quoteField(field):
			steps = appendUniqueTask(steps, astockcontracts.TaskKindValuationSnapshot)
		case matchAny(field, "research", "rating", "eps_forecast", "profit_forecast", "target_price"):
			steps = appendUniqueTask(steps, astockcontracts.TaskKindResearchLookup)
		case matchAny(field, "announcement", "announcements"):
			steps = appendUniqueTask(steps, astockcontracts.TaskKindAnnouncement)
		case matchAny(field, "hot_reason", "concept_blocks", "fund_flow", "northbound_flow", "dragon_tiger_board", "daily_dragon_tiger", "lockup_expiry", "industry_comparison"):
			steps = appendUniqueTask(steps, astockcontracts.TaskKindSignalLookup)
		case matchAny(field, "industry", "listing_date", "share_capital", "company_profile", "f10_profile"):
			steps = appendUniqueTask(steps, astockcontracts.TaskKindProfileLookup)
		}
	}
	return steps
}

func quoteField(value string) bool {
	return matchAny(value, "price", "change_pct", "turnover", "pe", "pe_ttm", "pe_static", "pb", "market_cap", "float_market_cap", "valuation", "limit_up", "limit_down")
}

func matchAny(value string, candidates ...string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, candidate := range candidates {
		if value == candidate {
			return true
		}
	}
	return false
}

func appendUniqueTask(values []astockcontracts.TaskKind, next astockcontracts.TaskKind) []astockcontracts.TaskKind {
	for _, value := range values {
		if value == next {
			return values
		}
	}
	return append(values, next)
}

func runSignal(ctx context.Context, cfg InvestigationConfig, params map[string]any) (astockcontracts.SignalPayload, error) {
	if cfg.Handlers.Signal != nil {
		return cfg.Handlers.Signal(ctx, params)
	}
	readiness := astockcontracts.BuildReadiness(astockcontracts.AdapterStatusUnsupported, astockcontracts.FailureCodeUnsupported, false, nil, nil)
	return astockcontracts.SignalPayload{
		Tool:          astockcontracts.ToolAStockSignalLookup,
		Source:        firstNonEmpty(cfg.Source, "agentx_a_stock_hostkit"),
		AdapterID:     "agentx_a_stock_signal_not_configured",
		AdapterStatus: astockcontracts.AdapterStatusUnsupported,
		FailureCode:   astockcontracts.FailureCodeUnsupported,
		Readiness:     readiness,
		Warnings:      []string{"a_stock_signal_lookup_handler_missing"},
	}, nil
}

func aggregateReadiness(items []astockcontracts.Readiness) (astockcontracts.AdapterStatus, astockcontracts.FailureCode, astockcontracts.Readiness) {
	if len(items) == 0 {
		status := astockcontracts.AdapterStatusUnsupported
		failure := astockcontracts.FailureCodeUnsupported
		return status, failure, astockcontracts.BuildReadiness(status, failure, false, nil, nil)
	}
	missing := []string{}
	review := []string{}
	status := astockcontracts.AdapterStatusOK
	failure := astockcontracts.FailureCodeNone
	requestedReady := true
	readyCount := 0
	for index, item := range items {
		if item.AnswerReady {
			readyCount++
		} else {
			requestedReady = false
		}
		if item.AdapterStatus != "" && item.AdapterStatus != astockcontracts.AdapterStatusOK && status == astockcontracts.AdapterStatusOK {
			status = item.AdapterStatus
		}
		if item.FailureCode != astockcontracts.FailureCodeNone && failure == astockcontracts.FailureCodeNone {
			failure = item.FailureCode
		}
		for _, field := range item.MissingFields {
			missing = append(missing, fmt.Sprintf("step_%d.%s", index+1, field))
		}
		for _, field := range item.ReviewRequiredFields {
			review = append(review, fmt.Sprintf("step_%d.%s", index+1, field))
		}
	}
	if status == astockcontracts.AdapterStatusOK && !requestedReady {
		status = astockcontracts.AdapterStatusEvidenceIncomplete
	}
	if failure == astockcontracts.FailureCodeNone && !requestedReady {
		failure = astockcontracts.FailureCodeMissingFields
	}
	if len(items) > 1 && readyCount > 0 && !requestedReady {
		return status, failure, partialReadiness(status, failure, missing, review)
	}
	return status, failure, astockcontracts.BuildReadiness(status, failure, requestedReady, missing, review)
}

func partialReadiness(status astockcontracts.AdapterStatus, failure astockcontracts.FailureCode, missing []string, review []string) astockcontracts.Readiness {
	if failure == astockcontracts.FailureCodeNone {
		failure = astockcontracts.FailureCodeMissingFields
	}
	reason := string(failure)
	if reason == "" {
		reason = "partial_evidence"
	}
	return astockcontracts.Readiness{
		AnswerReady:          true,
		Degraded:             true,
		DegradeReason:        reason,
		AdapterStatus:        status,
		FailureCode:          failure,
		RequestedFieldsReady: false,
		MissingFields:        append([]string(nil), missing...),
		ReviewRequiredFields: append([]string(nil), review...),
		NextRepairHint:       "fetch_missing_fields",
	}
}

func missingReadiness() astockcontracts.Readiness {
	return astockcontracts.BuildReadiness(astockcontracts.AdapterStatusUnsupported, astockcontracts.FailureCodeUnsupported, false, nil, nil)
}

func sourceUnavailableReadiness() astockcontracts.Readiness {
	return astockcontracts.BuildReadiness(astockcontracts.AdapterStatusUnavailable, astockcontracts.FailureCodeSourceUnavailable, false, nil, nil)
}

func projectIdentityResolution(payload *astockcontracts.InvestigationPayload, resolution *astockcontracts.IdentityResolution) {
	if payload == nil || payload.IdentityResolution != nil || resolution == nil {
		return
	}
	payload.IdentityResolution = cloneIdentityResolutionPtr(resolution)
}

func cloneIdentityResolutionPtr(resolution *astockcontracts.IdentityResolution) *astockcontracts.IdentityResolution {
	if resolution == nil {
		return nil
	}
	out := *resolution
	out.QueryVariants = append([]astockcontracts.IdentityResolutionQuery(nil), resolution.QueryVariants...)
	out.Candidates = append([]astockcontracts.IdentityResolutionCandidate(nil), resolution.Candidates...)
	out.Warnings = append([]string(nil), resolution.Warnings...)
	if resolution.SelectedCandidate != nil {
		selected := *resolution.SelectedCandidate
		out.SelectedCandidate = &selected
	}
	return &out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
