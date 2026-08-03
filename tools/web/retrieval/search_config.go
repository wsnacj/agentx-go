package retrieval

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultSearchEndpoint      = "https://api.search.brave.com/res/v1/web/search"
	DefaultSearchPerplexityURL = "https://api.perplexity.ai/chat/completions"
	DefaultSearchOpenRouterURL = "https://openrouter.ai/api/v1/chat/completions"
	DefaultSearchArkURL        = "https://open.feedcoopapi.com/search_api/web_search"
	DefaultSearchBaiduURL      = "https://qianfan.baidubce.com/v2/ai_search/web_search"
	DefaultSearchProvider      = "brave"
)

var (
	searchFreshnessShortcut = map[string]bool{
		"pd": true,
		"pw": true,
		"pm": true,
		"py": true,
	}
)

func DefaultSearchEndpointForProvider(provider string) string {
	switch provider {
	case "ark":
		return DefaultSearchArkURL
	case "baidu":
		return DefaultSearchBaiduURL
	default:
		return DefaultSearchEndpoint
	}
}

func NormalizeSearchProvider(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return DefaultSearchProvider
	}
	switch normalized {
	case "arksearch":
		return "ark"
	case "bidu", "bidusearch":
		return "baidu"
	}
	return normalized
}

func IsSupportedSearchProvider(provider string) bool {
	switch NormalizeSearchProvider(provider) {
	case "brave", "perplexity", "openrouter", "ark", "baidu":
		return true
	default:
		return false
	}
}

func NormalizeFreshness(value string) string {
	trimmed := normalizeFreshnessToken(value)
	if trimmed == "" {
		return ""
	}
	if searchFreshnessShortcut[trimmed] {
		return trimmed
	}
	parts := strings.Split(trimmed, "to")
	if len(parts) != 2 {
		return ""
	}
	start := strings.TrimSpace(parts[0])
	end := strings.TrimSpace(parts[1])
	if !IsValidISODate(start) || !IsValidISODate(end) {
		return ""
	}
	if start > end {
		return ""
	}
	return start + "to" + end
}

func MapFreshnessForArk(value string) (string, bool) {
	switch normalizeFreshnessToken(value) {
	case "pd":
		return "OneDay", true
	case "pw":
		return "OneWeek", true
	case "pm":
		return "OneMonth", true
	case "py":
		return "OneYear", true
	default:
		return "", false
	}
}

func MapFreshnessForBaidu(value string) (string, bool) {
	switch normalizeFreshnessToken(value) {
	case "pd", "pw":
		return "week", true
	case "pm":
		return "month", true
	case "py":
		return "year", true
	default:
		return "", false
	}
}

func MapFreshnessForPerplexity(value string) (string, bool) {
	switch normalizeFreshnessToken(value) {
	case "pd":
		return "day", true
	case "pw":
		return "week", true
	case "pm":
		return "month", true
	case "py":
		return "year", true
	default:
		return "", false
	}
}

func BuildFreshnessFromDateRange(dateAfter string, dateBefore string, now time.Time) string {
	dateAfter = strings.TrimSpace(dateAfter)
	dateBefore = strings.TrimSpace(dateBefore)
	switch {
	case dateAfter != "" && dateBefore != "":
		return dateAfter + "to" + dateBefore
	case dateAfter != "":
		return dateAfter + "to" + now.UTC().Format("2006-01-02")
	case dateBefore != "":
		return "1970-01-01to" + dateBefore
	default:
		return ""
	}
}

func NormalizeSearchDomainFilter(items []string) ([]string, string) {
	if len(items) == 0 {
		return nil, ""
	}
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	hasAllowlist := false
	hasDenylist := false
	for _, item := range items {
		trimmed := strings.ToLower(strings.TrimSpace(item))
		if trimmed == "" {
			continue
		}
		deny := strings.HasPrefix(trimmed, "-")
		if deny {
			hasDenylist = true
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
		} else {
			hasAllowlist = true
		}
		if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
			return nil, "domain_filter must contain domains or suffixes, not full URLs."
		}
		if !isSearchDomainToken(trimmed) {
			return nil, fmt.Sprintf("domain_filter contains invalid entry %q. Use domain names like nature.com or suffixes like .gov.", item)
		}
		normalized := trimmed
		if deny {
			normalized = "-" + trimmed
		}
		if seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	if len(out) == 0 {
		return nil, ""
	}
	if hasAllowlist && hasDenylist {
		return nil, "domain_filter cannot mix allowlist and denylist entries. Use either all positive entries or all entries prefixed with '-'."
	}
	if len(out) > 20 {
		return nil, "domain_filter supports a maximum of 20 domains."
	}
	return out, ""
}

func FormatDateForPerplexity(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return ""
	}
	return parsed.Format("1/2/2006")
}

func ResolveDateFilterProvider(currentProvider string, configuredProvider string, configuredProviderAPIKey string, configuredPerplexityAPIKey string, configuredPerplexityURL string, isConfigured func(provider string, providerAPIKey string, perplexityAPIKey string) bool) string {
	currentProvider = NormalizeSearchProvider(currentProvider)
	switch currentProvider {
	case "brave", "perplexity":
		return currentProvider
	}
	if CanUsePerplexityStructuredDateFilters(configuredPerplexityAPIKey, configuredPerplexityURL) {
		return "perplexity"
	}
	braveKey := ""
	if NormalizeSearchProvider(configuredProvider) == "brave" {
		braveKey = configuredProviderAPIKey
	}
	if isConfigured != nil && isConfigured("brave", braveKey, "") {
		return "brave"
	}
	return ""
}

func CanUsePerplexityStructuredDateFilters(apiKey string, baseURL string) bool {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return false
	}
	baseURL = strings.ToLower(strings.TrimSpace(baseURL))
	if baseURL == "" {
		baseURL = strings.ToLower(DefaultSearchPerplexityURL)
	}
	return !strings.Contains(baseURL, "openrouter")
}

func IsValidISODate(value string) bool {
	if len(value) != 10 {
		return false
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return false
	}
	return parsed.Format("2006-01-02") == value
}

func ResolveSearchSiteName(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Hostname())
}

func normalizeFreshnessToken(value string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	switch trimmed {
	case "d", "24h", "1d", "day", "today", "last24h", "last_24h":
		return "pd"
	case "w", "7d", "1w", "week", "last7d", "last_7d":
		return "pw"
	case "m", "30d", "1m", "month", "last30d", "last_30d":
		return "pm"
	case "y", "365d", "1y", "year", "last365d", "last_365d":
		return "py"
	default:
		return trimmed
	}
}

func isSearchDomainToken(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	trimmed = strings.TrimPrefix(trimmed, ".")
	if trimmed == "" {
		return false
	}
	for i, r := range trimmed {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		switch {
		case isAlphaNum:
		case r == '.' || r == '-':
		default:
			return false
		}
		if i == 0 && !isAlphaNum {
			return false
		}
	}
	return true
}
