package hostkit

import (
	"strings"
	"unicode"

	research "github.com/wsnacj/agentx-go/scenes/companyresearch"
)

type financeIdentityHint struct {
	CompanyName string
	StockCode   string
	MarketHint  string
}

func marketIntentWithFinanceIdentity(intent research.CompanyResearchIntent, finance map[string]any) research.CompanyResearchIntent {
	hint := financeResolvedIdentityHint(finance)
	if hint.MarketHint == "" {
		return intent
	}
	if strings.TrimSpace(intent.MarketHint) == "" || !marketHintHasSpecificExchange(intent.MarketHint) {
		intent.MarketHint = hint.MarketHint
	}
	return intent
}

func downstreamParamsWithFinanceIdentity(intent research.CompanyResearchIntent, target string, finance map[string]any) map[string]any {
	out := downstreamParams(intent, target)
	hint := financeResolvedIdentityHint(finance)
	if hint.StockCode != "" {
		out["stock_code"] = hint.StockCode
		out["ticker"] = hint.StockCode
	}
	if hint.MarketHint != "" {
		out["market_hint"] = hint.MarketHint
		out["market"] = hint.MarketHint
	}
	if hint.CompanyName != "" {
		if strings.TrimSpace(research.StringArg(out["entity_name"])) == "" {
			out["entity_name"] = hint.CompanyName
		}
	}
	return out
}

func downstreamNewsParamsWithFinanceIdentity(intent research.CompanyResearchIntent, finance map[string]any) map[string]any {
	out := downstreamParams(intent, "news")
	hint := financeResolvedIdentityHint(finance)
	if hint.CompanyName != "" {
		out["entity_name"] = hint.CompanyName
		out["entity_mentions"] = []string{hint.CompanyName}
	}
	if hint.StockCode != "" {
		out["stock_code"] = hint.StockCode
		out["ticker"] = hint.StockCode
	}
	if hint.MarketHint != "" {
		out["market_hint"] = hint.MarketHint
		out["market"] = hint.MarketHint
	}
	return out
}

func financeResolvedIdentityHint(finance map[string]any) financeIdentityHint {
	if len(finance) == 0 {
		return financeIdentityHint{}
	}
	stockCode := firstNonEmpty(
		deepString(finance, "candidates", "resolved_code"),
		deepString(finance, "candidates", "identity_resolution", "selected_candidate", "code_or_ticker"),
		deepString(finance, "candidates", "identity_resolution", "selected_candidate", "stock_code"),
		deepString(finance, "candidates", "identity_resolution", "selected_candidate", "ticker"),
		deepString(finance, "metrics", "stock_code"),
		deepString(finance, "metrics", "evidence", "stock_code"),
	)
	companyName := financeResolvedCompanyName(finance)
	marketHint := normalizeFinanceMarketHint(firstNonEmpty(
		deepString(finance, "candidates", "resolved_market"),
		deepString(finance, "candidates", "identity_resolution", "selected_candidate", "market"),
		deepString(finance, "metrics", "evidence", "market"),
		marketHintFromCode(stockCode),
	))
	return financeIdentityHint{
		CompanyName: companyName,
		StockCode:   normalizeStockCodeHint(stockCode),
		MarketHint:  marketHint,
	}
}

func financeResolvedCompanyName(finance map[string]any) string {
	if financeEvidenceReady(finance) {
		if name := firstMeaningfulCompanyName(
			deepString(finance, "metrics", "evidence", "company_name"),
			deepString(finance, "metrics", "company_name"),
			deepString(finance, "brief", "evidence", "company_name"),
		); name != "" {
			return name
		}
	}
	return firstMeaningfulCompanyName(
		deepString(finance, "candidates", "resolved_company"),
		deepString(finance, "candidates", "identity_resolution", "selected_candidate", "entity_name"),
	)
}

func firstMeaningfulCompanyName(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		switch strings.ToLower(value) {
		case "", "unknown", "n/a", "null":
			continue
		default:
			return value
		}
	}
	return ""
}

func financeSubjectMatchesIntent(intent research.CompanyResearchIntent, finance map[string]any) bool {
	if len(finance) == 0 {
		return true
	}
	hint := financeResolvedIdentityHint(finance)
	if hint.CompanyName == "" && hint.StockCode == "" && hint.MarketHint == "" {
		return true
	}
	if !marketHintsOverlap(intent.MarketHint, hint.MarketHint) {
		return false
	}
	if companyNameMatchesIntent(intent, hint.CompanyName) {
		return true
	}
	return identityCodeMatchesIntent(intent, hint.StockCode)
}

func marketHintsOverlap(expected, observed string) bool {
	expectedCategories := marketHintCategories(expected)
	observedCategories := marketHintCategories(observed)
	if len(expectedCategories) == 0 || len(observedCategories) == 0 {
		return true
	}
	if expectedCategories["global"] || observedCategories["global"] {
		return true
	}
	for category := range expectedCategories {
		if observedCategories[category] {
			return true
		}
	}
	return false
}

func companyNameMatchesIntent(intent research.CompanyResearchIntent, observed string) bool {
	observedKey := normalizeSubjectIdentity(observed)
	if observedKey == "" {
		return true
	}
	candidates := append([]string{}, intent.EntityName)
	candidates = append(candidates, intent.EntityMentions...)
	hasSpecificCJKCandidate := false
	for _, candidate := range candidates {
		if len(hanRunes(candidate)) >= 3 {
			hasSpecificCJKCandidate = true
			break
		}
	}
	for _, candidate := range candidates {
		if subjectIdentityCandidateMatches(candidate, observed, hasSpecificCJKCandidate) {
			return true
		}
	}
	return false
}

func subjectIdentityCandidateMatches(candidate string, observed string, hasSpecificCJKCandidate bool) bool {
	expectedKey := normalizeSubjectIdentity(candidate)
	observedKey := normalizeSubjectIdentity(observed)
	if expectedKey == "" || observedKey == "" {
		return false
	}
	if expectedKey == observedKey {
		return true
	}
	if identityContainmentAllowed(candidate, observed, hasSpecificCJKCandidate) &&
		(strings.Contains(observedKey, expectedKey) || strings.Contains(expectedKey, observedKey)) {
		return true
	}
	cjkKey := cjkSubjectPrefix(candidate)
	if cjkKey != "" &&
		identityContainmentAllowed(cjkKey, observed, hasSpecificCJKCandidate) &&
		strings.Contains(observedKey, normalizeSubjectIdentity(cjkKey)) {
		return true
	}
	return false
}

func identityContainmentAllowed(candidate string, observed string, hasSpecificCJKCandidate bool) bool {
	candidateHan := hanRunes(candidate)
	if len(candidateHan) > 0 {
		if len(candidateHan) < 3 {
			if observedHasCorporateSuffixAfterPrefix(string(candidateHan), observed) {
				return true
			}
			if hasSpecificCJKCandidate {
				return false
			}
			return false
		}
		return true
	}
	candidateKey := normalizeSubjectIdentity(candidate)
	return len([]rune(candidateKey)) > 3
}

func observedHasCorporateSuffixAfterPrefix(prefix string, observed string) bool {
	prefixRunes := hanRunes(prefix)
	observedRunes := hanRunes(observed)
	if len(prefixRunes) == 0 || len(observedRunes) <= len(prefixRunes) {
		return false
	}
	for idx, r := range prefixRunes {
		if observedRunes[idx] != r {
			return false
		}
	}
	tail := string(observedRunes[len(prefixRunes):])
	for _, suffix := range []string{"控股", "集团", "股份", "公司", "有限", "银行", "证券", "保险"} {
		if strings.HasPrefix(tail, suffix) {
			return true
		}
	}
	return false
}

func hanRunes(value string) []rune {
	out := []rune{}
	for _, r := range value {
		if unicode.Is(unicode.Han, r) {
			out = append(out, r)
		}
	}
	return out
}

func normalizeSubjectIdentity(value string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || unicode.Is(unicode.Han, r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func cjkSubjectPrefix(value string) string {
	chars := []rune{}
	for _, r := range value {
		if unicode.Is(unicode.Han, r) {
			chars = append(chars, r)
		}
	}
	if len(chars) >= 3 {
		prefix := string(chars[:2])
		if prefix == "中国" || prefix == "上海" || prefix == "北京" || prefix == "香港" {
			return string(chars[:3])
		}
	}
	if len(chars) >= 2 {
		return string(chars[:2])
	}
	return ""
}

func deepString(object map[string]any, path ...string) string {
	var current any = object
	for _, key := range path {
		next, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = next[key]
	}
	return strings.TrimSpace(research.StringArg(current))
}

func normalizeFinanceMarketHint(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.TrimPrefix(normalized, ".")
	switch normalized {
	case "hk", "hkg", "hongkong", "hong kong", "港股", "香港":
		return "hk"
	case "us", "usa", "nyse", "nasdaq", "amex", "美股", "美国":
		return "us"
	case "a", "cn", "china", "ashare", "a-share", "a_share", "sh", "sha", "sse", "sz", "sza", "szse", "沪市", "深市", "a股":
		return "A-share"
	default:
		return strings.TrimSpace(value)
	}
}

func normalizeStockCodeHint(value string) string {
	value = strings.TrimSpace(value)
	upper := strings.ToUpper(value)
	for _, suffix := range []string{".HK", ".US", ".SH", ".SZ", ".SS", ".O", ".N"} {
		if strings.HasSuffix(upper, suffix) {
			return strings.TrimSpace(value[:len(value)-len(suffix)])
		}
	}
	return value
}

func marketHintFromCode(value string) string {
	upper := strings.ToUpper(strings.TrimSpace(value))
	switch {
	case strings.HasSuffix(upper, ".HK"):
		return "hk"
	case strings.HasSuffix(upper, ".US"), strings.HasSuffix(upper, ".O"), strings.HasSuffix(upper, ".N"):
		return "us"
	case strings.HasSuffix(upper, ".SH"), strings.HasSuffix(upper, ".SZ"), strings.HasSuffix(upper, ".SS"):
		return "A-share"
	default:
		return ""
	}
}
