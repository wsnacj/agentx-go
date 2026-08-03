package contracts

import (
	"regexp"
	"strings"
	"unicode"
)

// Market is the normalized global stock market identifier supported by this package.
type Market string

const (
	MarketAuto Market = "auto"
	MarketHK   Market = "hk"
	MarketUS   Market = "us"
)

// NormalizeSecurityCode normalizes common HK/US ticker formats.
func NormalizeSecurityCode(input string, preferred Market) (code string, market Market, ok bool) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", "", false
	}
	s = strings.TrimSpace(strings.ReplaceAll(s, "：", ":"))
	lower := strings.ToLower(s)

	switch {
	case strings.HasPrefix(lower, "hk."):
		return normalizeHKCode(s[3:])
	case strings.HasPrefix(lower, "hk:"):
		return normalizeHKCode(s[3:])
	case strings.HasPrefix(lower, "hk"):
		if code, market, ok := normalizeHKCode(s[2:]); ok {
			return code, market, ok
		}
	case strings.HasSuffix(lower, ".hk"):
		return normalizeHKCode(s[:len(s)-3])
	case strings.HasSuffix(lower, ":hk"):
		return normalizeHKCode(s[:len(s)-3])
	case strings.HasPrefix(lower, "us."):
		return normalizeUSCode(s[3:])
	case strings.HasPrefix(lower, "us:"):
		return normalizeUSCode(s[3:])
	case strings.HasSuffix(lower, ".us"):
		return normalizeUSCode(s[:len(s)-3])
	case strings.Contains(lower, ":"):
		left, right, _ := strings.Cut(s, ":")
		switch strings.ToLower(strings.TrimSpace(left)) {
		case "nasdaq", "nyse", "amex", "us":
			return normalizeUSCode(right)
		case "hkg", "hkex", "sehk", "hk":
			return normalizeHKCode(right)
		}
	}

	if preferred == MarketHK {
		return normalizeHKCode(s)
	}
	if preferred == MarketUS {
		return normalizeUSCode(s)
	}
	if code, market, ok := normalizeHKCode(s); ok {
		return code, market, ok
	}
	return normalizeUSCode(s)
}

func normalizeHKCode(input string) (string, Market, bool) {
	s := strings.TrimSpace(input)
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return "", "", false
	}
	if len(s) > 5 || !allDigits(s) {
		return "", "", false
	}
	return leftPad(s, 5), MarketHK, true
}

var usTickerPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9.-]{0,9}$`)

func normalizeUSCode(input string) (string, Market, bool) {
	s := strings.ToUpper(strings.TrimSpace(input))
	s = strings.TrimSuffix(s, ".OQ")
	s = strings.TrimSuffix(s, ".N")
	s = strings.TrimSuffix(s, ".A")
	if s == "" || !usTickerPattern.MatchString(s) {
		return "", "", false
	}
	return s, MarketUS, true
}

func allDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func leftPad(s string, width int) string {
	for len(s) < width {
		s = "0" + s
	}
	return s
}
