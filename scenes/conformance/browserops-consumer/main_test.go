package main

import "testing"

func TestFixedVersionBrowserOpsConsumer(t *testing.T) {
	got, err := run()
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	const want = "agentx-browserops-ok:browser-ops-pack:browser_form_submit_v1:browser_act:true"
	if got != want {
		t.Fatalf("run() = %q, want %q", got, want)
	}
}
