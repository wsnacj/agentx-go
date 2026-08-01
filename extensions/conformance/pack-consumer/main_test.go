package main

import "testing"

func TestFixedVersionConsumer(t *testing.T) {
	got, err := run()
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	const want = "agentx-pack-ok:portable-research:collect-v1:host_collect"
	if got != want {
		t.Fatalf("run() = %q, want %q", got, want)
	}
}
