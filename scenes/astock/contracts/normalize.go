package contracts

import (
	"strings"
	"unicode"
)

// Market is the normalized A-share market prefix.
type Market string

const (
	MarketAuto Market = "auto"
	MarketSH   Market = "sh"
	MarketSZ   Market = "sz"
	MarketBJ   Market = "bj"
)

// NormalizeAStockCode normalizes common A-share ticker formats to a 6-digit code and market.
//
// Supported inputs include 688017, SH688017, sh688017, 688017.SH, SZ000001,
// and BJ832000. The returned market is inferred from explicit prefix/suffix
// when present, otherwise from the code prefix.
func NormalizeAStockCode(input string) (code string, market Market, ok bool) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", "", false
	}
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")

	explicitMarket := Market("")
	for _, prefix := range []struct {
		text   string
		market Market
	}{
		{"sh", MarketSH},
		{"sz", MarketSZ},
		{"bj", MarketBJ},
	} {
		if strings.HasPrefix(s, prefix.text) {
			explicitMarket = prefix.market
			s = strings.TrimPrefix(s, prefix.text)
			break
		}
	}
	for _, suffix := range []struct {
		text   string
		market Market
	}{
		{".sh", MarketSH},
		{".sz", MarketSZ},
		{".bj", MarketBJ},
	} {
		if strings.HasSuffix(s, suffix.text) {
			explicitMarket = suffix.market
			s = strings.TrimSuffix(s, suffix.text)
			break
		}
	}

	if len(s) != 6 || !allDigits(s) {
		return "", "", false
	}
	if explicitMarket != "" {
		return s, explicitMarket, true
	}
	market, ok = InferAStockMarket(s)
	if !ok {
		return "", "", false
	}
	return s, market, true
}

// InferAStockMarket infers the common market prefix for a normalized 6-digit A-share code.
func InferAStockMarket(code string) (Market, bool) {
	if len(code) != 6 || !allDigits(code) {
		return "", false
	}
	switch code[0] {
	case '6', '9':
		return MarketSH, true
	case '0', '2', '3':
		return MarketSZ, true
	case '4', '8':
		return MarketBJ, true
	default:
		return "", false
	}
}

func allDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
