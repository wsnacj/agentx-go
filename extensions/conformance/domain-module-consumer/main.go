package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wsnacj/agentx-go/extensions/domainkit"
	"github.com/wsnacj/agentx-go/extensions/domainmodule"
)

type sampleConfig struct {
	Token string `json:"token"`
}

func register(ctx context.Context) (domainmodule.Report, error) {
	return domainmodule.RegisterAll(ctx, []domainmodule.Registration{{
		Manifest: domainmodule.Manifest{
			ID: "sample", Name: "Sample", Tools: []string{"lookup"},
			RequiredConfig: []domainmodule.ConfigRequirement{{Key: "token", Required: true}},
		},
		Apply: func(_ context.Context, manifest domainmodule.Manifest, cfg domainmodule.Config) (domainmodule.Diagnostics, error) {
			value, ok := cfg.Value(manifest.ID).(sampleConfig)
			if !ok || value.Token == "" {
				return nil, fmt.Errorf("sample config is unavailable")
			}
			return domainmodule.Diagnostics{
				domainmodule.NewDiagnostic(manifest.ID, domainmodule.SeverityInfo, "sample_ready", "sample adapter is ready", nil),
			}, nil
		},
	}}, domainmodule.RegisterOptions{
		ConfigResolvers: []domainmodule.ConfigResolver{func(_ context.Context, manifest domainmodule.Manifest, _ domainmodule.Config) (any, domainmodule.Diagnostics, error) {
			if manifest.ID != "sample" {
				return nil, nil, nil
			}
			return sampleConfig{Token: "fixture-token"}, nil, nil
		}},
	})
}

func execute(ctx context.Context) (domainkit.RunResult, error) {
	runtime, err := domainkit.New(domainkit.Config{Modules: []domainkit.Module{{
		Manifest: domainmodule.Manifest{ID: "sample", Name: "Sample", Tools: []string{"lookup"}},
		Handlers: map[string]domainkit.Handler{
			"lookup": func(_ context.Context, args map[string]any) (any, error) {
				return map[string]any{"id": args["id"], "status": "ready"}, nil
			},
		},
	}}})
	if err != nil {
		return domainkit.RunResult{}, err
	}
	return runtime.Run(ctx, domainkit.RunRequest{
		ModuleID:  "sample",
		CaseID:    "lookup-fixture",
		Tool:      "lookup",
		Arguments: map[string]any{"id": "fixture-1"},
	})
}

func main() {
	report, err := register(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	result, err := execute(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Registration domainmodule.Report `json:"registration"`
		Execution    domainkit.RunResult `json:"execution"`
	}{Registration: report, Execution: result}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
