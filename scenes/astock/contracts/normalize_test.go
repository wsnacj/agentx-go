package contracts

import "testing"

func TestNormalizeAStockCode(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantCode   string
		wantMarket Market
		wantOK     bool
	}{
		{name: "plain sh", input: "688017", wantCode: "688017", wantMarket: MarketSH, wantOK: true},
		{name: "prefix sh", input: "SH688017", wantCode: "688017", wantMarket: MarketSH, wantOK: true},
		{name: "suffix sh", input: "688017.SH", wantCode: "688017", wantMarket: MarketSH, wantOK: true},
		{name: "prefix sz", input: "SZ000001", wantCode: "000001", wantMarket: MarketSZ, wantOK: true},
		{name: "suffix sz", input: "300750.sz", wantCode: "300750", wantMarket: MarketSZ, wantOK: true},
		{name: "prefix bj", input: "BJ832000", wantCode: "832000", wantMarket: MarketBJ, wantOK: true},
		{name: "plain bj 4", input: "430047", wantCode: "430047", wantMarket: MarketBJ, wantOK: true},
		{name: "invalid letters", input: "abc", wantOK: false},
		{name: "invalid length", input: "60000", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, gotMarket, gotOK := NormalizeAStockCode(tt.input)
			if gotOK != tt.wantOK || gotCode != tt.wantCode || gotMarket != tt.wantMarket {
				t.Fatalf("NormalizeAStockCode(%q) = code=%q market=%q ok=%v, want code=%q market=%q ok=%v", tt.input, gotCode, gotMarket, gotOK, tt.wantCode, tt.wantMarket, tt.wantOK)
			}
		})
	}
}
