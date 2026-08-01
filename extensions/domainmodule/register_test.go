package domainmodule

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeManifestAndConfig(t *testing.T) {
	manifest, err := NormalizeManifest(Manifest{ID: " Demo ", Skills: []string{" one ", "one", ""}, Tools: []string{" tool "}})
	if err != nil {
		t.Fatalf("NormalizeManifest: %v", err)
	}
	if manifest.ID != "demo" || manifest.Name != "demo" || !reflect.DeepEqual(manifest.Skills, []string{"one"}) || !reflect.DeepEqual(manifest.Tools, []string{"tool"}) {
		t.Fatalf("unexpected manifest: %#v", manifest)
	}
	cfg := Config{}.With(" Demo ", map[string]string{"token": "value"})
	if !cfg.Has("DEMO") || cfg.Value("demo") == nil {
		t.Fatalf("canonical config lookup failed: %#v", cfg)
	}
}

func TestRegisterAllRejectsDuplicateBeforeMutation(t *testing.T) {
	called := 0
	report, err := RegisterAll(context.Background(), []Registration{
		{Manifest: Manifest{ID: "Demo", Name: "first"}, Apply: func(context.Context, Manifest, Config) (Diagnostics, error) { called++; return nil, nil }},
		{Manifest: Manifest{ID: " demo ", Name: "second"}, Apply: func(context.Context, Manifest, Config) (Diagnostics, error) { called++; return nil, nil }},
	}, RegisterOptions{})
	if err == nil || err.Error() != `duplicate domain module id "demo"` {
		t.Fatalf("unexpected duplicate error: %v", err)
	}
	if called != 0 {
		t.Fatalf("apply called %d times before duplicate rejection", called)
	}
	if !report.HasErrors() || len(report.Modules) != 1 || report.Modules[0].Manifest.ID != "demo" {
		t.Fatalf("unexpected duplicate report: %#v", report)
	}
}

func TestRegisterAllPreservesOrderAndPartialMutation(t *testing.T) {
	wantErr := errors.New("second failed")
	var order []string
	report, err := RegisterAll(context.Background(), []Registration{
		{Manifest: Manifest{ID: "first"}, Apply: func(_ context.Context, manifest Manifest, _ Config) (Diagnostics, error) {
			order = append(order, manifest.ID)
			return Diagnostics{NewDiagnostic("", SeverityWarning, "first_warning", "warning", nil)}, nil
		}},
		{Manifest: Manifest{ID: "second"}, Apply: func(_ context.Context, manifest Manifest, _ Config) (Diagnostics, error) {
			order = append(order, manifest.ID)
			return Diagnostics{NewDiagnostic("", SeverityError, "second_error", "failed", nil)}, wantErr
		}},
		{Manifest: Manifest{ID: "third"}, Apply: func(_ context.Context, manifest Manifest, _ Config) (Diagnostics, error) {
			order = append(order, manifest.ID)
			return nil, nil
		}},
	}, RegisterOptions{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("RegisterAll error = %v", err)
	}
	if !reflect.DeepEqual(order, []string{"first", "second"}) {
		t.Fatalf("apply order = %#v", order)
	}
	if got := report.Diagnostics(); len(got) != 3 || got[0].Code != "first_warning" || got[1].Code != DiagnosticModuleRegistered || got[2].Code != "second_error" {
		t.Fatalf("diagnostic order = %#v", got)
	}
}

func TestRegisterAllResolvesConfigAndReportsMissingRequirement(t *testing.T) {
	type moduleConfig struct {
		Token string `json:"token"`
	}
	var resolved Config
	report, err := RegisterAll(context.Background(), []Registration{
		{Manifest: Manifest{ID: "resolved", RequiredConfig: []ConfigRequirement{{Key: "token", Required: true}}}},
		{Manifest: Manifest{ID: "missing", RequiredConfig: []ConfigRequirement{{Key: "token", Description: "API token", Required: true}}}},
	}, RegisterOptions{
		ConfigResolvers: []ConfigResolver{func(_ context.Context, manifest Manifest, _ Config) (any, Diagnostics, error) {
			if manifest.ID == "resolved" {
				return moduleConfig{Token: "value"}, nil, nil
			}
			return nil, nil, nil
		}},
		ConfigResolved: func(value Config) { resolved = value },
	})
	if err != nil {
		t.Fatalf("RegisterAll: %v", err)
	}
	if !resolved.Has("resolved") {
		t.Fatalf("resolved config not delivered: %#v", resolved)
	}
	var resolvedCode, missingCode bool
	for _, diagnostic := range report.Diagnostics() {
		resolvedCode = resolvedCode || diagnostic.Code == DiagnosticConfigResolved
		missingCode = missingCode || diagnostic.Code == DiagnosticMissingConfig
	}
	if !resolvedCode || !missingCode {
		t.Fatalf("expected config diagnostics: %#v", report.Diagnostics())
	}
}

func TestRegisterAllPreflightRunsBeforeResolver(t *testing.T) {
	preflightErr := errors.New("host preflight failed")
	resolverCalled := false
	_, err := RegisterAll(context.Background(), []Registration{{Manifest: Manifest{ID: "demo"}}}, RegisterOptions{
		Preflight: func(manifests []Manifest) error {
			if len(manifests) != 1 || manifests[0].ID != "demo" {
				t.Fatalf("unexpected manifests: %#v", manifests)
			}
			return preflightErr
		},
		ConfigResolvers: []ConfigResolver{func(context.Context, Manifest, Config) (any, Diagnostics, error) {
			resolverCalled = true
			return nil, nil, nil
		}},
	})
	if !errors.Is(err, preflightErr) || resolverCalled {
		t.Fatalf("err=%v resolverCalled=%v", err, resolverCalled)
	}
}

func TestDiagnosticJSONAndNormalization(t *testing.T) {
	diagnostic := NewDiagnostic(" Demo ", Severity(" WARNING "), " code ", " message ", map[string]string{" key ": " value ", "": "drop"})
	encoded, err := json.Marshal(diagnostic)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	text := string(encoded)
	for _, fragment := range []string{`"module_id":"demo"`, `"severity":"warning"`, `"code":"code"`, `"key":"value"`} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("JSON %s missing %s", text, fragment)
		}
	}
}
