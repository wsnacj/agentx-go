package main

import (
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	output, err := run()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(output, "agentx-astock-ok:") {
		t.Fatalf("unexpected output: %s", output)
	}
}
