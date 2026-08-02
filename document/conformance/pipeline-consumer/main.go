package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wsnacj/agentx-go/document/pipeline"
	"github.com/wsnacj/agentx-go/document/pipeline/section"
)

func run(ctx context.Context) (string, error) {
	dir, err := os.MkdirTemp("", "agentx-document-consumer-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	docPath := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(docPath, []byte("Revenue: 42"), 0o600); err != nil {
		return "", err
	}
	specPath := filepath.Join(dir, "main.yaml")
	spec := `meta:
  doc_type: conformance
  version: v1
  header_footer_cleanup: none
chapters:
  - key: summary
    fields:
      - key: revenue
        type: number
        normalize: number
        extractors:
          - type: regex
            pattern: 'Revenue:\s*([0-9]+)'
`
	if err := os.WriteFile(specPath, []byte(spec), 0o600); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "sections.yaml"), []byte("rules: []\n"), 0o600); err != nil {
		return "", err
	}

	runtime, err := pipeline.New(pipeline.Dependencies{
		Loader: pipeline.DocumentLoaderFunc(func(context.Context, pipeline.ExtractRequest) (*pipeline.ExtractedDocument, error) {
			return &pipeline.ExtractedDocument{Pages: []string{"Revenue: 42"}, TextSource: "consumer"}, nil
		}),
		Sectioner: pipeline.SectionerFunc(func(_ context.Context, req pipeline.SectionRequest) ([]*section.Node, error) {
			return []*section.Node{{Name: "summary", Pages: append([]string{}, req.Pages...)}}, nil
		}),
	})
	if err != nil {
		return "", err
	}
	result, err := runtime.Run(ctx, pipeline.ParseRequest{
		DocPath:        docPath,
		SpecPath:       specPath,
		ArtifactPolicy: pipeline.ArtifactPolicyNone,
	})
	if err != nil {
		return "", err
	}
	value := result.Chapters["summary"].Fields["revenue"].Value
	payload, err := json.Marshal(map[string]any{"status": "parsed", "revenue": value})
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func main() {
	output, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}

