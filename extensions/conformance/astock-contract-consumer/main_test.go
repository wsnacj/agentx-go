package main

import "testing"

func TestFixedVersionConsumerUsesCanonicalAssetFSAndAStockContracts(t *testing.T) {
	got, err := runConsumer()
	if err != nil {
		t.Fatalf("runConsumer() error = %v", err)
	}
	const want = "agentx-extension-astock-contract-ok:000001:sz:true"
	if got != want {
		t.Fatalf("runConsumer() = %q, want %q", got, want)
	}
}
