package main

import "testing"

func TestFixedVersionControlContractConsumer(t *testing.T) {
	result, err := run()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	const want = "agentx-controlcontract-ok:ready_for_host_action:1:applied:evidence_weak:objective_graph_ready:objective_verification_recovery_ready"
	if result != want {
		t.Fatalf("result = %q, want %q", result, want)
	}
}
