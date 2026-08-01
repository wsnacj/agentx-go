package main

import "testing"

func TestFixedVersionProductShellObservationConsumer(t *testing.T) {
	got, err := run()
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	const want = "agentx-productshell-observation-ok:session-001:completed:1:passed:ready:key=progress.completed status=completed available=true line=kind=progress;status=completed;process_count=1 missing=- blocked=- next=render_log_fields target=log kind=host_diagnostic_operator_line source=external_host_adapter"
	if got != want {
		t.Fatalf("run() = %q, want %q", got, want)
	}
}
