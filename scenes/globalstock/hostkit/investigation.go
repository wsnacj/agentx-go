package hostkit

import (
	"context"
	"strings"

	agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
	globalstock "github.com/wsnacj/agentx-go/scenes/globalstock"
	globalcontracts "github.com/wsnacj/agentx-go/scenes/globalstock/contracts"
)

type InvestigationConfig struct {
	Source              string
	SourcePolicyDefault string
	Handlers            InvestigationHandlers
	// AnswerContract lets a host attach product-facing answer wording after the
	// portable coordinator has completed source-backed comparison evidence.
	// The canonical kit deliberately does not own investment disclaimers or prose.
	AnswerContract func(globalcontracts.InvestigationPayload) *globalcontracts.InvestigationAnswerContract
}

type InvestigationHandlers struct {
	Quote        func(context.Context, map[string]any) (globalcontracts.QuotePayload, error)
	Profile      func(context.Context, map[string]any) (globalcontracts.ProfilePayload, error)
	Announcement func(context.Context, map[string]any) (globalcontracts.AnnouncementPayload, error)
	Research     func(context.Context, map[string]any) (globalcontracts.ResearchPayload, error)
	Signal       func(context.Context, map[string]any) (globalcontracts.SignalPayload, error)
}

// BuildGlobalStockInvestigationHandler builds the high-level tool handler used by hosts.
func BuildGlobalStockInvestigationHandler(cfg InvestigationConfig) globalstock.ToolPayloadHandler {
	return func(ctx context.Context, params map[string]any) (any, error) {
		return BuildGlobalStockInvestigationPayload(ctx, cfg, params)
	}
}

// BuildGlobalStockInvestigationPayload dispatches the structured task frame to host-owned handlers.
func BuildGlobalStockInvestigationPayload(ctx context.Context, cfg InvestigationConfig, params map[string]any) (globalcontracts.InvestigationPayload, error) {
	intent := IntentFromParams(params)
	if intent.SourcePolicy == "" {
		intent.SourcePolicy = cfg.SourcePolicyDefault
	}
	taskParams := ParamsFromIntent(params, intent)
	payload := globalcontracts.InvestigationPayload{
		Tool:   globalstock.ToolGlobalStockInvestigation,
		Source: firstNonEmpty(cfg.Source, "agentx_global_stock_hostkit"),
		Intent: intent,
		Evidence: globalcontracts.SourceEvidence{
			Provider:   "agentx_global_stock",
			Source:     firstNonEmpty(cfg.Source, "agentx_global_stock_hostkit"),
			Freshness:  intent.Freshness.Mode,
			AsOf:       intent.Freshness.AsOf,
			Confidence: 0.7,
		},
	}
	if isFinanceReportHandoffIntent(intent) {
		return buildFinanceReportHandoffInvestigation(intent, payload), nil
	}
	if intent.TaskKind == globalcontracts.TaskKindProfileLookup {
		return buildProfileInvestigation(ctx, cfg, params, taskParams, intent, payload)
	}
	if isSignalIntent(intent) {
		return buildSignalInvestigation(ctx, cfg, params, taskParams, intent, payload)
	}
	if isAnnouncementIntent(intent) {
		return buildAnnouncementInvestigation(ctx, cfg, params, taskParams, intent, payload)
	}
	if isResearchIntent(intent) {
		return buildResearchInvestigation(ctx, cfg, params, taskParams, intent, payload)
	}
	if cfg.Handlers.Quote == nil {
		payload.AdapterStatus = globalcontracts.AdapterStatusUnsupported
		payload.FailureCode = globalcontracts.FailureCodeUnsupported
		payload.Readiness = globalcontracts.BuildReadiness(payload.AdapterStatus, payload.FailureCode, false, []string{"quote"}, nil)
		payload.Warnings = append(payload.Warnings, "global_stock_quote_lookup_handler_missing")
		return payload, nil
	}
	if isMultiEntityQuoteComparison(intent) {
		readiness := []globalcontracts.Readiness{}
		seenQuotes := map[string]bool{}
		for _, mention := range uniqueEntityMentions(intent) {
			mentionParams := paramsForEntityMention(taskParams, mention)
			if key := quoteParamsIdentityKey(mentionParams); key != "" && seenQuotes[key] {
				continue
			}
			quote, err := cfg.Handlers.Quote(ctx, mentionParams)
			if err != nil {
				payload.Warnings = append(payload.Warnings, "global_stock_quote_lookup_error:"+mention+":"+globalStockInvestigationError(err, "quote_lookup_failed"))
				readiness = append(readiness, sourceUnavailableReadiness())
				continue
			}
			if key := quoteSubjectIdentityKey(quote.Subject); key != "" {
				if seenQuotes[key] {
					continue
				}
				seenQuotes[key] = true
			}
			payload.Quotes = append(payload.Quotes, quote)
			readiness = append(readiness, quote.Readiness)
		}
		if len(payload.Quotes) == 0 && len(readiness) == 0 {
			readiness = append(readiness, missingReadiness())
		}
		payload.AdapterStatus, payload.FailureCode, payload.Readiness = aggregateReadiness(readiness)
		if cfg.AnswerContract != nil {
			payload.AnswerContract = cfg.AnswerContract(payload)
		}
		return payload, nil
	}
	quote, err := cfg.Handlers.Quote(ctx, taskParams)
	if err != nil {
		payload.Warnings = append(payload.Warnings, "global_stock_quote_lookup_error:"+globalStockInvestigationError(err, "quote_lookup_failed"))
		payload.AdapterStatus = globalcontracts.AdapterStatusUnavailable
		payload.FailureCode = globalcontracts.FailureCodeSourceUnavailable
		payload.Readiness = sourceUnavailableReadiness()
		return payload, nil
	}
	payload.Quote = &quote
	payload.IdentityResolution = cloneIdentityResolutionPtr(quote.IdentityResolution)
	payload.AdapterStatus = quote.AdapterStatus
	payload.FailureCode = quote.FailureCode
	payload.Readiness = quote.Readiness
	return payload, nil
}

func buildSignalInvestigation(ctx context.Context, cfg InvestigationConfig, params map[string]any, taskParams map[string]any, intent globalcontracts.InvestigationIntent, payload globalcontracts.InvestigationPayload) (globalcontracts.InvestigationPayload, error) {
	if cfg.Handlers.Signal == nil {
		payload.AdapterStatus = globalcontracts.AdapterStatusUnsupported
		payload.FailureCode = globalcontracts.FailureCodeUnsupported
		payload.Readiness = globalcontracts.BuildReadiness(payload.AdapterStatus, payload.FailureCode, false, []string{"signals"}, nil)
		payload.Warnings = append(payload.Warnings, "global_stock_signal_lookup_handler_missing")
		return payload, nil
	}
	signal, err := cfg.Handlers.Signal(ctx, taskParams)
	if err != nil {
		payload.Warnings = append(payload.Warnings, "global_stock_signal_lookup_error:"+globalStockInvestigationError(err, "signal_lookup_failed"))
		payload.AdapterStatus = globalcontracts.AdapterStatusUnavailable
		payload.FailureCode = globalcontracts.FailureCodeSourceUnavailable
		payload.Readiness = globalcontracts.BuildReadiness(payload.AdapterStatus, payload.FailureCode, false, []string{"signals"}, nil)
		return payload, nil
	}
	payload.Signal = &signal
	payload.IdentityResolution = cloneIdentityResolutionPtr(signal.IdentityResolution)
	payload.AdapterStatus = signal.AdapterStatus
	payload.FailureCode = signal.FailureCode
	payload.Readiness = signal.Readiness
	return payload, nil
}

func buildResearchInvestigation(ctx context.Context, cfg InvestigationConfig, params map[string]any, taskParams map[string]any, intent globalcontracts.InvestigationIntent, payload globalcontracts.InvestigationPayload) (globalcontracts.InvestigationPayload, error) {
	if cfg.Handlers.Research == nil {
		payload.AdapterStatus = globalcontracts.AdapterStatusUnsupported
		payload.FailureCode = globalcontracts.FailureCodeUnsupported
		payload.Readiness = globalcontracts.BuildReadiness(payload.AdapterStatus, payload.FailureCode, false, []string{"research_reports"}, nil)
		payload.Warnings = append(payload.Warnings, "global_stock_research_lookup_handler_missing")
		return payload, nil
	}
	research, err := cfg.Handlers.Research(ctx, taskParams)
	if err != nil {
		payload.Warnings = append(payload.Warnings, "global_stock_research_lookup_error:"+globalStockInvestigationError(err, "research_lookup_failed"))
		payload.AdapterStatus = globalcontracts.AdapterStatusUnavailable
		payload.FailureCode = globalcontracts.FailureCodeSourceUnavailable
		payload.Readiness = globalcontracts.BuildReadiness(payload.AdapterStatus, payload.FailureCode, false, []string{"research_reports"}, nil)
		return payload, nil
	}
	payload.Research = &research
	payload.IdentityResolution = cloneIdentityResolutionPtr(research.IdentityResolution)
	payload.AdapterStatus = research.AdapterStatus
	payload.FailureCode = research.FailureCode
	payload.Readiness = research.Readiness
	return payload, nil
}

func buildAnnouncementInvestigation(ctx context.Context, cfg InvestigationConfig, params map[string]any, taskParams map[string]any, intent globalcontracts.InvestigationIntent, payload globalcontracts.InvestigationPayload) (globalcontracts.InvestigationPayload, error) {
	if cfg.Handlers.Announcement == nil {
		payload.AdapterStatus = globalcontracts.AdapterStatusUnsupported
		payload.FailureCode = globalcontracts.FailureCodeUnsupported
		payload.Readiness = globalcontracts.BuildReadiness(payload.AdapterStatus, payload.FailureCode, false, []string{"announcements"}, nil)
		payload.Warnings = append(payload.Warnings, "global_stock_announcement_lookup_handler_missing")
		return payload, nil
	}
	announcement, err := cfg.Handlers.Announcement(ctx, taskParams)
	if err != nil {
		payload.Warnings = append(payload.Warnings, "global_stock_announcement_lookup_error:"+globalStockInvestigationError(err, "announcement_lookup_failed"))
		payload.AdapterStatus = globalcontracts.AdapterStatusUnavailable
		payload.FailureCode = globalcontracts.FailureCodeSourceUnavailable
		payload.Readiness = globalcontracts.BuildReadiness(payload.AdapterStatus, payload.FailureCode, false, []string{"announcements"}, nil)
		return payload, nil
	}
	payload.Announcement = &announcement
	payload.IdentityResolution = cloneIdentityResolutionPtr(announcement.IdentityResolution)
	payload.AdapterStatus = announcement.AdapterStatus
	payload.FailureCode = announcement.FailureCode
	payload.Readiness = announcement.Readiness
	return payload, nil
}

func isSignalIntent(intent globalcontracts.InvestigationIntent) bool {
	if intent.TaskKind == globalcontracts.TaskKindSignalLookup {
		return true
	}
	for _, output := range intent.RequestedOutputs {
		if matchAny(output, "signal", "signals", "disclosure_signal", "filing_signal") {
			return true
		}
	}
	for _, field := range intent.RequestedFields {
		if matchAny(field, "signal", "signals", "hk_buyback", "hk_board_meeting", "us_form_4", "us_8k", "us_13f", "us_earnings_event") {
			return true
		}
	}
	return false
}

func isAnnouncementIntent(intent globalcontracts.InvestigationIntent) bool {
	if intent.TaskKind == globalcontracts.TaskKindAnnouncement {
		return true
	}
	for _, output := range intent.RequestedOutputs {
		if matchAny(output, "announcement", "announcements", "filing", "filings", "disclosure") {
			return true
		}
	}
	return false
}

func isResearchIntent(intent globalcontracts.InvestigationIntent) bool {
	if intent.TaskKind == globalcontracts.TaskKindResearchLookup {
		return true
	}
	for _, output := range intent.RequestedOutputs {
		if matchAny(output, "research", "rating", "analyst_rating", "target_price", "consensus_rating") {
			return true
		}
	}
	for _, field := range intent.RequestedFields {
		if matchAny(field, "research", "rating", "rating_change", "target_price", "consensus_rating", "analyst_count") {
			return true
		}
	}
	return false
}

func buildProfileInvestigation(ctx context.Context, cfg InvestigationConfig, params map[string]any, taskParams map[string]any, intent globalcontracts.InvestigationIntent, payload globalcontracts.InvestigationPayload) (globalcontracts.InvestigationPayload, error) {
	if cfg.Handlers.Profile == nil {
		payload.AdapterStatus = globalcontracts.AdapterStatusUnsupported
		payload.FailureCode = globalcontracts.FailureCodeUnsupported
		payload.Readiness = globalcontracts.BuildReadiness(payload.AdapterStatus, payload.FailureCode, false, []string{"profile"}, nil)
		payload.Warnings = append(payload.Warnings, "global_stock_profile_lookup_handler_missing")
		return payload, nil
	}
	profile, err := cfg.Handlers.Profile(ctx, taskParams)
	if err != nil {
		payload.Warnings = append(payload.Warnings, "global_stock_profile_lookup_error:"+globalStockInvestigationError(err, "profile_lookup_failed"))
		payload.AdapterStatus = globalcontracts.AdapterStatusUnavailable
		payload.FailureCode = globalcontracts.FailureCodeSourceUnavailable
		payload.Readiness = globalcontracts.BuildReadiness(payload.AdapterStatus, payload.FailureCode, false, []string{"profile"}, nil)
		return payload, nil
	}
	payload.Profile = &profile
	payload.IdentityResolution = cloneIdentityResolutionPtr(profile.IdentityResolution)
	payload.AdapterStatus = profile.AdapterStatus
	payload.FailureCode = profile.FailureCode
	payload.Readiness = profile.Readiness
	return payload, nil
}

func globalStockInvestigationError(err error, code string) string {
	return agentxsafeerror.Summary(agentxsafeerror.Project(err, "global_stock_investigation", code))
}

func isMultiEntityQuoteComparison(intent globalcontracts.InvestigationIntent) bool {
	if strings.TrimSpace(intent.StockCode) != "" {
		return false
	}
	if len(uniqueEntityMentions(intent)) < 2 {
		return false
	}
	if intent.TaskKind == globalcontracts.TaskKindComparison || intent.TaskKind == globalcontracts.TaskKindScreening {
		return true
	}
	for _, output := range intent.RequestedOutputs {
		if matchAny(output, "comparison", "quote", "valuation", "valuation_snapshot") {
			return true
		}
	}
	for _, field := range intent.RequestedFields {
		if matchAny(field, "price", "change_pct", "pe_ttm", "pb", "market_cap") {
			return true
		}
	}
	return false
}

func uniqueEntityMentions(intent globalcontracts.InvestigationIntent) []string {
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

func quoteParamsIdentityKey(params map[string]any) string {
	code := strings.ToUpper(strings.TrimSpace(firstNonEmptyString(params["stock_code"])))
	market := strings.ToLower(strings.TrimSpace(firstNonEmptyString(params["market"])))
	if code == "" || market == "" || market == string(globalcontracts.MarketAuto) {
		return ""
	}
	return market + ":" + code
}

func quoteSubjectIdentityKey(subject globalcontracts.Subject) string {
	if !subject.Verified {
		return ""
	}
	code := strings.ToUpper(strings.TrimSpace(subject.StockCode))
	market := strings.ToLower(strings.TrimSpace(string(subject.Market)))
	if code == "" || market == "" || market == string(globalcontracts.MarketAuto) {
		return ""
	}
	return market + ":" + code
}

func firstNonEmptyString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return ""
	}
}

func paramsForEntityMention(base map[string]any, mention string) map[string]any {
	out := map[string]any{}
	for key, value := range base {
		out[key] = value
	}
	delete(out, "stock_code")
	delete(out, "market")
	if code, market, ok := normalizeExplicitSecurityCodeMention(mention); ok {
		out["stock_code"] = code
		out["market"] = string(market)
		delete(out, "entity_name")
		return out
	}
	if entity, market, ok := splitMarketQualifiedMention(mention); ok {
		out["entity_name"] = entity
		out["market"] = string(market)
		return out
	}
	out["entity_name"] = strings.TrimSpace(mention)
	return out
}

func splitMarketQualifiedMention(mention string) (string, globalcontracts.Market, bool) {
	term := normalizeMentionWhitespace(strings.TrimSpace(mention))
	if term == "" {
		return "", "", false
	}
	lower := strings.ToLower(term)
	qualifiers := []struct {
		market   globalcontracts.Market
		suffixes []string
		prefixes []string
	}{
		{
			market: globalcontracts.MarketHK,
			suffixes: []string{
				"港股", "香港上市", "香港",
				" hk", ".hk", "-hk", "_hk", "(hk)", "（hk）",
				" hkex", ".hkex", "-hkex", "_hkex", "(hkex)", "（hkex）",
			},
			prefixes: []string{"港股", "香港上市", "香港", "hk ", "hkex "},
		},
		{
			market: globalcontracts.MarketUS,
			suffixes: []string{
				"美股", "美国上市", "美国", "纳斯达克", "纽交所",
				" us", ".us", "-us", "_us", "(us)", "（us）",
				" nyse", ".nyse", "-nyse", "_nyse", "(nyse)", "（nyse）",
				" nasdaq", ".nasdaq", "-nasdaq", "_nasdaq", "(nasdaq)", "（nasdaq）",
			},
			prefixes: []string{"美股", "美国上市", "美国", "纳斯达克", "纽交所", "us ", "nyse ", "nasdaq "},
		},
	}
	for _, qualifier := range qualifiers {
		for _, suffix := range qualifier.suffixes {
			if !strings.HasSuffix(lower, strings.ToLower(suffix)) {
				continue
			}
			entity := strings.TrimSpace(term[:len(term)-len(suffix)])
			entity = trimMarketQualifierSeparators(entity)
			if entity != "" {
				return entity, qualifier.market, true
			}
		}
		for _, prefix := range qualifier.prefixes {
			if !strings.HasPrefix(lower, strings.ToLower(prefix)) {
				continue
			}
			entity := strings.TrimSpace(term[len(prefix):])
			entity = trimMarketQualifierSeparators(entity)
			if entity != "" {
				return entity, qualifier.market, true
			}
		}
	}
	return "", "", false
}

func normalizeMentionWhitespace(value string) string {
	value = strings.ReplaceAll(value, "　", " ")
	return strings.Join(strings.Fields(value), " ")
}

func trimMarketQualifierSeparators(value string) string {
	return strings.Trim(strings.TrimSpace(value), " -_./()（）[]【】")
}

func normalizeExplicitSecurityCodeMention(mention string) (string, globalcontracts.Market, bool) {
	term := strings.TrimSpace(strings.ReplaceAll(mention, "：", ":"))
	if term == "" {
		return "", "", false
	}
	lower := strings.ToLower(term)
	explicit := strings.Contains(term, ":") ||
		strings.HasPrefix(lower, "hk.") ||
		strings.HasPrefix(lower, "hk") ||
		strings.HasSuffix(lower, ".hk") ||
		strings.HasPrefix(lower, "us.") ||
		strings.HasSuffix(lower, ".us") ||
		isDigitsOnly(term)
	if !explicit && term == strings.ToUpper(term) && hasASCIIAlpha(term) && looksLikeBareUSTickerMention(term) {
		explicit = true
	}
	if !explicit {
		return "", "", false
	}
	return globalcontracts.NormalizeSecurityCode(term, globalcontracts.MarketAuto)
}

func looksLikeBareUSTickerMention(term string) bool {
	term = strings.TrimSpace(term)
	return term != "" && len(term) <= 5
}

func isDigitsOnly(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func hasASCIIAlpha(value string) bool {
	for _, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			return true
		}
	}
	return false
}

func aggregateReadiness(values []globalcontracts.Readiness) (globalcontracts.AdapterStatus, globalcontracts.FailureCode, globalcontracts.Readiness) {
	if len(values) == 0 {
		return globalcontracts.AdapterStatusUnavailable, globalcontracts.FailureCodeMissingFields, missingReadiness()
	}
	allReady := true
	partialReady := false
	for _, item := range values {
		if item.AnswerReady {
			partialReady = true
			continue
		}
		allReady = false
	}
	if allReady {
		return globalcontracts.AdapterStatusOK, globalcontracts.FailureCodeNone, globalcontracts.BuildReadiness(globalcontracts.AdapterStatusOK, globalcontracts.FailureCodeNone, true, nil, nil)
	}
	if partialReady {
		return globalcontracts.AdapterStatusEvidenceIncomplete, globalcontracts.FailureCodeMissingFields, globalcontracts.BuildReadiness(globalcontracts.AdapterStatusEvidenceIncomplete, globalcontracts.FailureCodeMissingFields, false, []string{"partial_quotes"}, nil)
	}
	return values[0].AdapterStatus, values[0].FailureCode, values[0]
}

func missingReadiness() globalcontracts.Readiness {
	return globalcontracts.BuildReadiness(globalcontracts.AdapterStatusUnavailable, globalcontracts.FailureCodeMissingFields, false, []string{"quote"}, nil)
}

func sourceUnavailableReadiness() globalcontracts.Readiness {
	return globalcontracts.BuildReadiness(globalcontracts.AdapterStatusUnavailable, globalcontracts.FailureCodeSourceUnavailable, false, []string{"quote"}, nil)
}

func cloneIdentityResolutionPtr(resolution *globalcontracts.IdentityResolution) *globalcontracts.IdentityResolution {
	if resolution == nil {
		return nil
	}
	out := *resolution
	out.QueryVariants = append([]globalcontracts.IdentityResolutionQuery(nil), resolution.QueryVariants...)
	out.Candidates = append([]globalcontracts.IdentityResolutionCandidate(nil), resolution.Candidates...)
	out.Warnings = append([]string(nil), resolution.Warnings...)
	if resolution.SelectedCandidate != nil {
		selected := *resolution.SelectedCandidate
		out.SelectedCandidate = &selected
	}
	return &out
}

func matchAny(value string, needles ...string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, needle := range needles {
		if value == strings.ToLower(strings.TrimSpace(needle)) {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
