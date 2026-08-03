package contracts

import "testing"

func TestNormalizeSecurityCodeHKAndUS(t *testing.T) {
	cases := []struct {
		in     string
		market Market
		code   string
		ok     bool
	}{
		{in: "00700", market: MarketHK, code: "00700", ok: true},
		{in: "700.HK", market: MarketHK, code: "00700", ok: true},
		{in: "HK.1810", market: MarketHK, code: "01810", ok: true},
		{in: "NASDAQ:AAPL", market: MarketUS, code: "AAPL", ok: true},
		{in: "aapl.us", market: MarketUS, code: "AAPL", ok: true},
	}
	for _, tc := range cases {
		code, market, ok := NormalizeSecurityCode(tc.in, MarketAuto)
		if ok != tc.ok || code != tc.code || market != tc.market {
			t.Fatalf("NormalizeSecurityCode(%q) = code=%q market=%q ok=%v", tc.in, code, market, ok)
		}
	}
}
