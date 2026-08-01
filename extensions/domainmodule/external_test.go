package domainmodule_test

import (
	"context"
	"testing"

	"github.com/wsnacj/agentx-go/extensions/domainmodule"
)

func TestExternalHostRegistration(t *testing.T) {
	called := false
	report, err := domainmodule.RegisterAll(context.Background(), []domainmodule.Registration{{
		Manifest: domainmodule.Manifest{ID: "sample", Tools: []string{"lookup"}},
		Apply: func(_ context.Context, manifest domainmodule.Manifest, _ domainmodule.Config) (domainmodule.Diagnostics, error) {
			called = manifest.ID == "sample"
			return nil, nil
		},
	}}, domainmodule.RegisterOptions{})
	if err != nil {
		t.Fatalf("RegisterAll: %v", err)
	}
	if !called || report.HasErrors() {
		t.Fatalf("called=%v report=%#v", called, report)
	}
}
