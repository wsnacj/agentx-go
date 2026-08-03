package main

import (
	"context"
	"testing"
)

func TestFixedVersionConsumer(t *testing.T) {
	got, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := "agentx-finance-ok:global-stock-quote-pack:financial-report-metrics-pack:1:2"
	if got != want {
		t.Fatalf("output=%q want=%q", got, want)
	}
}
