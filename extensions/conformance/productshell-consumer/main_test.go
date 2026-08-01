package main

import "testing"

func TestFixedVersionProductShellConsumer(t *testing.T) {
	got, err := run()
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	const want = "agentx-productshell-ok:portable-research:research.lookup:collect-v1:portable-review:case-001:AgentX"
	if got != want {
		t.Fatalf("run() = %q, want %q", got, want)
	}
}
