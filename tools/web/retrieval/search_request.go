package retrieval

import (
	"fmt"
	"strings"
	"time"
)

type SearchPrepareOptions struct {
	ConfiguredProvider                    string
	RequestedProvider                     string
	Query                                 string
	Count                                 int
	Country                               string
	SearchLang                            string
	UILang                                string
	Freshness                             string
	DateAfter                             string
	DateBefore                            string
	DomainFilter                          []string
	AuthoritativeOnly                     bool
	QueryRewrite                          bool
	CanUseProvider                        func(provider string) bool
	CanUsePerplexityStructuredDateFilters bool
	Now                                   func() time.Time
}

type SearchValidationError struct {
	Code    string
	Message string
}

type PreparedSearchRequest struct {
	Query                 string
	Count                 int
	Country               string
	SearchLang            string
	UILang                string
	RequestedProvider     string
	EffectiveProvider     string
	ProviderNote          string
	Freshness             string
	DateAfter             string
	DateBefore            string
	DomainFilter          []string
	DoubaoCustomTimeRange string
	// Deprecated: use DoubaoCustomTimeRange. Retained for source compatibility.
	ArkTimeRange         string
	DoubaoSites          []string
	DoubaoBlockedHosts   []string
	AuthoritativeOnly    bool
	QueryRewrite         bool
	BaiduRecency         string
	PerplexityRecency    string
	PerplexityDateAfter  string
	PerplexityDateBefore string
}

func PrepareSearchRequest(opts SearchPrepareOptions) (PreparedSearchRequest, *SearchValidationError) {
	plan := PreparedSearchRequest{
		Query:             strings.TrimSpace(opts.Query),
		Count:             opts.Count,
		Country:           strings.TrimSpace(opts.Country),
		SearchLang:        strings.TrimSpace(opts.SearchLang),
		UILang:            strings.TrimSpace(opts.UILang),
		EffectiveProvider: NormalizeSearchProvider(opts.ConfiguredProvider),
		Freshness:         strings.TrimSpace(opts.Freshness),
		DateAfter:         strings.TrimSpace(opts.DateAfter),
		DateBefore:        strings.TrimSpace(opts.DateBefore),
		AuthoritativeOnly: opts.AuthoritativeOnly,
		QueryRewrite:      opts.QueryRewrite,
	}
	if requested := strings.TrimSpace(opts.RequestedProvider); requested != "" {
		plan.RequestedProvider = NormalizeSearchProvider(requested)
	}
	canUseProvider := opts.CanUseProvider
	if canUseProvider == nil {
		canUseProvider = func(string) bool { return false }
	}
	if plan.EffectiveProvider == "" {
		plan.EffectiveProvider = DefaultSearchProvider
	}
	if plan.RequestedProvider != "" && !IsSupportedSearchProvider(plan.RequestedProvider) {
		return PreparedSearchRequest{}, &SearchValidationError{
			Code:    "unsupported_provider",
			Message: "provider must be brave, perplexity, openrouter, doubao_custom, doubao_global, ark, or baidu",
		}
	}
	if plan.RequestedProvider != "" && plan.RequestedProvider != plan.EffectiveProvider {
		if canUseProvider(plan.RequestedProvider) {
			plan.EffectiveProvider = plan.RequestedProvider
			plan.ProviderNote = AppendSearchProviderNote(plan.ProviderNote, fmt.Sprintf("provider override accepted: %s", plan.RequestedProvider))
		} else {
			plan.ProviderNote = AppendSearchProviderNote(plan.ProviderNote, fmt.Sprintf("provider override %s ignored: missing credentials, fallback to %s", plan.RequestedProvider, plan.EffectiveProvider))
		}
	}
	domainFilter, domainFilterErr := normalizeSearchDomainFilter(opts.DomainFilter, plan.EffectiveProvider == SearchProviderDoubaoCustom)
	if domainFilterErr != "" {
		return PreparedSearchRequest{}, &SearchValidationError{
			Code:    "invalid_domain_filter",
			Message: domainFilterErr,
		}
	}
	plan.DomainFilter = domainFilter
	if plan.Freshness != "" && (plan.DateAfter != "" || plan.DateBefore != "") {
		return PreparedSearchRequest{}, &SearchValidationError{
			Code:    "conflicting_time_filters",
			Message: "freshness and date_after/date_before cannot be used together. Use either freshness or a date range.",
		}
	}
	if plan.DateAfter != "" && !IsValidISODate(plan.DateAfter) {
		return PreparedSearchRequest{}, &SearchValidationError{
			Code:    "invalid_date",
			Message: "date_after must be YYYY-MM-DD format.",
		}
	}
	if plan.DateBefore != "" && !IsValidISODate(plan.DateBefore) {
		return PreparedSearchRequest{}, &SearchValidationError{
			Code:    "invalid_date",
			Message: "date_before must be YYYY-MM-DD format.",
		}
	}
	if plan.DateAfter != "" && plan.DateBefore != "" && plan.DateAfter > plan.DateBefore {
		return PreparedSearchRequest{}, &SearchValidationError{
			Code:    "invalid_date_range",
			Message: "date_after must be before date_before.",
		}
	}
	if plan.DateAfter != "" || plan.DateBefore != "" {
		compatibleProvider := resolveDateFilterProviderWithPredicate(plan.EffectiveProvider, opts.CanUsePerplexityStructuredDateFilters, canUseProvider)
		if compatibleProvider != "" && compatibleProvider != plan.EffectiveProvider {
			plan.ProviderNote = AppendSearchProviderNote(plan.ProviderNote, fmt.Sprintf("provider %s switched to %s because structured date filters require brave, perplexity, or doubao_custom", plan.EffectiveProvider, compatibleProvider))
			plan.EffectiveProvider = compatibleProvider
		}
		switch plan.EffectiveProvider {
		case "brave":
			now := time.Now
			if opts.Now != nil {
				now = opts.Now
			}
			plan.Freshness = BuildFreshnessFromDateRange(plan.DateAfter, plan.DateBefore, now().UTC())
		case SearchProviderDoubaoCustom:
			now := time.Now
			if opts.Now != nil {
				now = opts.Now
			}
			plan.DoubaoCustomTimeRange = MapDateRangeForDoubaoCustom(plan.DateAfter, plan.DateBefore, now().UTC())
			plan.ArkTimeRange = plan.DoubaoCustomTimeRange
			plan.Freshness = plan.DoubaoCustomTimeRange
		case "perplexity":
			plan.PerplexityDateAfter = FormatDateForPerplexity(plan.DateAfter)
			plan.PerplexityDateBefore = FormatDateForPerplexity(plan.DateBefore)
		case "openrouter":
			return PreparedSearchRequest{}, &SearchValidationError{
				Code:    "unsupported_date_filter",
				Message: "date_after/date_before filtering is not supported by the openrouter compatibility path. Use provider=perplexity for structured date filters.",
			}
		case SearchProviderDoubaoGlobal:
			return PreparedSearchRequest{}, &SearchValidationError{
				Code:    "unsupported_date_filter",
				Message: "date_after/date_before filtering is not supported by doubao_global.",
			}
		default:
			plan.ProviderNote = AppendSearchProviderNote(plan.ProviderNote, fmt.Sprintf("date_after/date_before were ignored because provider %s does not support structured date filters", plan.EffectiveProvider))
			plan.DateAfter = ""
			plan.DateBefore = ""
		}
	}
	if plan.Freshness != "" {
		switch plan.EffectiveProvider {
		case "brave":
			normalized := NormalizeFreshness(plan.Freshness)
			if normalized == "" {
				return PreparedSearchRequest{}, &SearchValidationError{
					Code:    "invalid_freshness",
					Message: "freshness must be one of pd, pw, pm, py, or YYYY-MM-DDtoYYYY-MM-DD",
				}
			}
			plan.Freshness = normalized
		case SearchProviderDoubaoCustom:
			if plan.DoubaoCustomTimeRange == "" {
				mapped, ok := MapFreshnessForArk(plan.Freshness)
				if !ok {
					return PreparedSearchRequest{}, &SearchValidationError{
						Code:    "invalid_freshness",
						Message: "doubao_custom provider freshness supports only pd, pw, pm, py.",
					}
				}
				plan.DoubaoCustomTimeRange = mapped
				plan.ArkTimeRange = mapped
				plan.Freshness = mapped
			}
		case "baidu":
			mapped, ok := MapFreshnessForBaidu(plan.Freshness)
			if !ok {
				return PreparedSearchRequest{}, &SearchValidationError{
					Code:    "invalid_freshness",
					Message: "baidu provider freshness supports only pd, pw, pm, py.",
				}
			}
			plan.BaiduRecency = mapped
			plan.Freshness = mapped
		case "perplexity", "openrouter":
			mapped, ok := MapFreshnessForPerplexity(plan.Freshness)
			if !ok {
				return PreparedSearchRequest{}, &SearchValidationError{
					Code:    "invalid_freshness",
					Message: "perplexity/openrouter freshness supports only pd, pw, pm, py.",
				}
			}
			plan.PerplexityRecency = mapped
			plan.Freshness = mapped
		default:
			return PreparedSearchRequest{}, &SearchValidationError{
				Code:    "unsupported_freshness",
				Message: "freshness is only supported for brave/perplexity/openrouter/doubao_custom/baidu providers.",
			}
		}
	}
	if len(plan.DomainFilter) > 0 && plan.EffectiveProvider != "perplexity" && plan.EffectiveProvider != SearchProviderDoubaoCustom {
		message := fmt.Sprintf("domain_filter is not supported by the %s provider. Use Perplexity or Doubao Custom for domain filtering.", plan.EffectiveProvider)
		if plan.EffectiveProvider == "openrouter" {
			message = "domain_filter is not supported by the openrouter compatibility path. Use provider=perplexity for native Perplexity domain filtering."
		}
		return PreparedSearchRequest{}, &SearchValidationError{
			Code:    "unsupported_domain_filter",
			Message: message,
		}
	}
	if plan.EffectiveProvider == SearchProviderDoubaoCustom {
		plan.DoubaoSites, plan.DoubaoBlockedHosts = SplitDomainFilterForDoubaoCustom(plan.DomainFilter)
		if len(plan.DoubaoSites) > 20 || len(plan.DoubaoBlockedHosts) > 5 {
			return PreparedSearchRequest{}, &SearchValidationError{
				Code:    "domain_filter_limit",
				Message: "doubao_custom supports at most 20 included domains and 5 excluded domains.",
			}
		}
	} else if plan.AuthoritativeOnly {
		return PreparedSearchRequest{}, &SearchValidationError{
			Code:    "unsupported_authority_filter",
			Message: fmt.Sprintf("authoritative_only is not supported by the %s provider.", plan.EffectiveProvider),
		}
	} else if plan.QueryRewrite {
		return PreparedSearchRequest{}, &SearchValidationError{
			Code:    "unsupported_query_rewrite",
			Message: fmt.Sprintf("query_rewrite is not supported by the %s provider.", plan.EffectiveProvider),
		}
	}
	return plan, nil
}

func AppendSearchProviderNote(current string, note string) string {
	current = strings.TrimSpace(current)
	note = strings.TrimSpace(note)
	switch {
	case note == "":
		return current
	case current == "":
		return note
	case strings.Contains(strings.ToLower(current), strings.ToLower(note)):
		return current
	default:
		return current + "; " + note
	}
}

func resolveDateFilterProviderWithPredicate(currentProvider string, canUseStructuredPerplexity bool, canUseProvider func(provider string) bool) string {
	currentProvider = NormalizeSearchProvider(currentProvider)
	switch currentProvider {
	case "brave", "perplexity", SearchProviderDoubaoCustom:
		return currentProvider
	}
	if canUseStructuredPerplexity {
		return "perplexity"
	}
	if canUseProvider != nil && canUseProvider("brave") {
		return "brave"
	}
	return ""
}
