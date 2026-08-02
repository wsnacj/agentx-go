package main

import "testing"

func TestFixedVersionAStockConsumer(t *testing.T) {
	got, err := run()
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	const want = "agentx-astock-ok:agentx_a_stock:a-stock-valuation-pack:a_stock_valuation_lookup_v1:7:true:true"
	if got != want {
		t.Fatalf("run() = %q, want %q", got, want)
	}
}
