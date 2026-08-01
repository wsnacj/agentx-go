package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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

func main() {
	report, err := register(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
