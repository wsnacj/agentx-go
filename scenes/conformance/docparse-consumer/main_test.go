package main

import "testing"

func TestFixedVersionDocparseConsumer(t *testing.T) {
	got, err := result()
	if err != nil {
		t.Fatal(err)
	}
	const want = "agentx-docparse-ok:docparse-document-pack:7:success:true"
	if got != want {
		t.Fatalf("result = %q, want %q", got, want)
	}
}
