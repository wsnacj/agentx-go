package main

import (
	"context"
	"testing"

	"github.com/wsnacj/agentx-go/extensions/domainmodule"
)

func TestFixedVersionConsumer(t *testing.T) {
	report, err := register(context.Background())
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if report.HasErrors() {
		t.Fatalf("unexpected report errors: %#v", report)
	}
	want := map[string]bool{
		domainmodule.DiagnosticConfigResolved:   false,
		domainmodule.DiagnosticModuleRegistered: false,
		"sample_ready":                          false,
	}
	for _, diagnostic := range report.Diagnostics() {
		if _, ok := want[diagnostic.Code]; ok {
			want[diagnostic.Code] = true
		}
	}
	for code, found := range want {
		if !found {
			t.Fatalf("missing diagnostic %q in %#v", code, report.Diagnostics())
		}
	}
}
