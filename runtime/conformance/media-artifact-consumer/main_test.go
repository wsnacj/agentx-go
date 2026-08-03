package main

import "testing"

func TestRun(t *testing.T) {
	got, err := run()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	const want = "agentx-media-artifact-ok:pdf:rendered_page:page-1.png"
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
