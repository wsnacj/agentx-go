package pipeline_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/wsnacj/agentx-go/document/pipeline"
	"github.com/wsnacj/agentx-go/document/pipeline/section"
)

func TestRuntimeRunUsesExplicitAdapters(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(docPath, []byte("amount 42"), 0o600); err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(dir, "main.yaml")
	spec := `meta:
  doc_type: canonical
  version: v1
  header_footer_cleanup: none
chapters:
  - key: summary
    fields:
      - key: amount
        type: number
        normalize: number
        extractors:
          - type: llm
validations:
  - name: positive
    expr: summary.amount > 0
    severity: error
`
	if err := os.WriteFile(specPath, []byte(spec), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sections.yaml"), []byte("rules: []\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var modelCalls atomic.Int32
	runtime, err := pipeline.New(pipeline.Dependencies{
		Loader: pipeline.DocumentLoaderFunc(func(ctx context.Context, req pipeline.ExtractRequest) (*pipeline.ExtractedDocument, error) {
			if req.Path != docPath {
				t.Fatalf("path = %q", req.Path)
			}
			return &pipeline.ExtractedDocument{Pages: []string{"amount 42"}, TextSource: "test"}, nil
		}),
		Sectioner: pipeline.SectionerFunc(func(ctx context.Context, req pipeline.SectionRequest) ([]*section.Node, error) {
			return []*section.Node{{Name: "summary", Pages: append([]string{}, req.Pages...)}}, nil
		}),
		Model: pipeline.ModelFunc(func(ctx context.Context, req pipeline.ModelRequest) (string, error) {
			modelCalls.Add(1)
			return `{"amount":"42"}`, nil
		}),
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := runtime.Run(context.Background(), pipeline.ParseRequest{
		DocPath:        docPath,
		SpecPath:       specPath,
		ModelName:      "host-model",
		ArtifactPolicy: pipeline.ArtifactPolicyNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if modelCalls.Load() != 1 {
		t.Fatalf("model calls = %d, want 1", modelCalls.Load())
	}
	chapter := result.Chapters["summary"]
	if chapter == nil {
		t.Fatal("summary chapter is missing")
	}
	if got := chapter.Fields["amount"].Value; got != float64(42) {
		t.Fatalf("amount = %#v, want 42", got)
	}
	if len(result.Validations) != 1 || !result.Validations[0].Passed {
		t.Fatalf("validations = %#v", result.Validations)
	}
	if result.Diagnostics == nil || result.Diagnostics.TextSource != "test" {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
}

func TestRuntimeRunHonorsPreCanceledContext(t *testing.T) {
	runtime, err := pipeline.New(pipeline.Dependencies{
		Loader: pipeline.DocumentLoaderFunc(func(context.Context, pipeline.ExtractRequest) (*pipeline.ExtractedDocument, error) {
			t.Fatal("loader must not run")
			return nil, nil
		}),
		Sectioner: pipeline.SectionerFunc(func(context.Context, pipeline.SectionRequest) ([]*section.Node, error) {
			t.Fatal("sectioner must not run")
			return nil, nil
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = runtime.Run(ctx, pipeline.ParseRequest{DocPath: "input", SpecPath: "spec"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}
