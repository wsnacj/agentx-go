package docparse_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"testing"

	"github.com/wsnacj/agentx-go/scenes/docparse"
	"github.com/wsnacj/agentx-go/scenes/docparse/adapters"
	"github.com/wsnacj/agentx-go/scenes/docparse/fusion"
	"github.com/wsnacj/agentx-go/scenes/docparse/hostkit"
	"github.com/wsnacj/agentx-go/scenes/docparse/planner"
	"github.com/wsnacj/agentx-go/scenes/docparse/profile"
	"github.com/wsnacj/agentx-go/scenes/docparse/representation"
	"github.com/wsnacj/agentx-go/scenes/docparse/understanding"
)

func TestRecommendedSurfaceAndDeterministicUnderstanding(t *testing.T) {
	def := docparse.PackDefinition()
	if def.Manifest.ID != docparse.PackID || len(docparse.ToolNames()) != 7 {
		t.Fatalf("unexpected portable surface: id=%q tools=%v", def.Manifest.ID, docparse.ToolNames())
	}
	if _, err := fs.ReadFile(docparse.ExtensionFS(), "skills/document-operations/SKILL.md"); err != nil {
		t.Fatalf("read embedded skill: %v", err)
	}

	document, err := representation.FromTextPages("host://invoice", []string{"invoice total amount 10 with page evidence"}, representation.TextSourcePlainText)
	if err != nil {
		t.Fatal(err)
	}
	engine := understanding.New(understanding.Options{
		Profiles: profile.NewRegistry(profile.ExtractionProfile{ID: "invoice-v1", DocumentType: "invoice"}),
		Adapters: adapters.NewRegistry(adapters.Func{
			AdapterID: "host-invoice",
			RouteKind: planner.RouteHostProfile,
			Run: func(context.Context, adapters.Input) (adapters.Output, error) {
				return adapters.Output{Status: "success", Fields: []map[string]any{{"key": "amount", "value": "10", "page_refs": []int{1}}}}, nil
			},
		}),
	})
	out, err := engine.Run(context.Background(), understanding.Input{
		Document: document, Params: map[string]any{"profile_id": "invoice-v1"}, HasHostProfileAdapter: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Fusion == nil || out.Fusion.Status != fusion.StatusAnswerReady || out.Fusion.FieldCount != 1 {
		t.Fatalf("unexpected understanding result: %#v", out)
	}
}

func TestHostKitProjectsInlineResultWithoutFilesystem(t *testing.T) {
	kit := hostkit.New(hostkit.Config{Source: "external-test"})
	out, err := kit.ExtractFields(context.Background(), map[string]any{
		"parse_result": map[string]any{
			"status": "success",
			"fields": []any{map[string]any{"key": "amount", "value": "10", "page_refs": []any{1}}},
		},
		"requested_fields": []string{"amount"},
	})
	if err != nil {
		t.Fatal(err)
	}
	blob, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(blob, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["status"] != "success" || payload["source"] != "external-test" {
		t.Fatalf("unexpected hostkit payload: %s", blob)
	}
}

func TestUnderstandingPreservesCancellationIdentity(t *testing.T) {
	document, err := representation.FromTextPages("host://cancel", []string{"document text long enough to execute adapter"}, "text")
	if err != nil {
		t.Fatal(err)
	}
	engine := understanding.New(understanding.Options{Adapters: adapters.NewRegistry(adapters.Func{
		AdapterID: "cancelled",
		RouteKind: planner.RouteGenericText,
		Run: func(ctx context.Context, _ adapters.Input) (adapters.Output, error) {
			<-ctx.Done()
			return adapters.Output{}, ctx.Err()
		},
	})})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = engine.Run(ctx, understanding.Input{Document: document})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation identity lost: %v", err)
	}
}
