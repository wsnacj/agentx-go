package tools_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/document/pipeline"
	pipelinetypes "github.com/wsnacj/agentx-go/document/pipeline/types"
	documenttools "github.com/wsnacj/agentx-go/document/tools"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

func TestDocumentParseAndPDFToolsUseExplicitHosts(t *testing.T) {
	registry := agentxtools.NewRegistry()
	documentHost := documenttools.DocumentHost{
		Runtime: documenttools.DocumentParserFunc(func(ctx context.Context, request pipeline.ParseRequest) (*pipelinetypes.DocumentResult, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			return &pipelinetypes.DocumentResult{
				ChapterOrder: []string{"summary"},
				Chapters: map[string]*pipelinetypes.ChapterResult{
					"summary": {Key: "summary", Fields: map[string]pipelinetypes.FieldResult{
						"revenue": {Key: "revenue", Value: 42, Source: "regex", Confidence: 1},
					}},
				},
				Diagnostics: &pipelinetypes.DocumentDiagnostics{PageCount: 1, TextSource: "fake"},
			}, nil
		}),
		Paths: documenttools.PathResolverFunc(func(ctx context.Context, request documenttools.PathRequest) (documenttools.ResolvedPath, error) {
			if err := ctx.Err(); err != nil {
				return documenttools.ResolvedPath{}, err
			}
			return documenttools.ResolvedPath{Path: request.Value, Display: request.Value}, nil
		}),
	}
	if err := documenttools.RegisterDocumentParseTools(registry, documenttools.DocumentParseToolOptions{
		EnabledTools: []string{"document_parse"}, Host: documentHost,
	}); err != nil {
		t.Fatal(err)
	}

	documentResult, err := registry.Execute(context.Background(), llm.FunctionCall{
		Name: "document_parse", Arguments: `{"document_path":"report.txt","spec_path":"spec","artifact_policy":"none"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	var documentPayload map[string]any
	if err := json.Unmarshal([]byte(documentResult), &documentPayload); err != nil {
		t.Fatal(err)
	}
	if documentPayload["status"] != "success" || documentPayload["field_count"] != float64(1) {
		t.Fatalf("unexpected document payload: %#v", documentPayload)
	}

	pdfBackend := memoryPDFBackend{}
	if err := documenttools.RegisterPDFTools(registry, documenttools.PDFToolOptions{
		EnabledTools: []string{"pdf"},
		Backend:      pdfBackend,
		Models: []documenttools.PDFModelCandidate{{
			Name: "text-model", Client: "fake", Model: "text-model", ConfigKey: "text-model",
		}},
		Host: documenttools.PDFHost{
			Inputs: documenttools.PDFInputResolverFunc(func(ctx context.Context, request documenttools.PDFInputRequest) (documenttools.ResolvedPDFInput, error) {
				if err := ctx.Err(); err != nil {
					return documenttools.ResolvedPDFInput{}, err
				}
				return documenttools.ResolvedPDFInput{Path: request.Reference, Display: request.Reference}, nil
			}),
			Chat: func(ctx context.Context, input llm.ChatInput) (*llm.ChatResponse, error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				if len(input.Messages) == 0 {
					t.Fatal("expected grounded PDF context")
				}
				return &llm.ChatResponse{Content: "Revenue is 42 [p1]."}, nil
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	pdfResult, err := registry.Execute(context.Background(), llm.FunctionCall{
		Name: "pdf", Arguments: `{"pdf":"report.pdf","prompt":"What is revenue?"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(pdfResult, `"status":"success"`) || !strings.Contains(pdfResult, "Revenue is 42") {
		t.Fatalf("unexpected pdf payload: %s", pdfResult)
	}
}

func TestPDFInputCancellationPropagates(t *testing.T) {
	registry := agentxtools.NewRegistry()
	if err := documenttools.RegisterPDFTools(registry, documenttools.PDFToolOptions{
		EnabledTools: []string{"pdf_extract"},
		Backend:      memoryPDFBackend{},
		Host: documenttools.PDFHost{Inputs: documenttools.PDFInputResolverFunc(func(ctx context.Context, _ documenttools.PDFInputRequest) (documenttools.ResolvedPDFInput, error) {
			<-ctx.Done()
			return documenttools.ResolvedPDFInput{}, ctx.Err()
		})},
	}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := registry.Execute(ctx, llm.FunctionCall{Name: "pdf_extract", Arguments: `{"pdf":"report.pdf"}`})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestDocumentParseUsesHostErrorProjector(t *testing.T) {
	registry := agentxtools.NewRegistry()
	rawErr := errors.New("provider secret sentinel")
	host := documenttools.DocumentHost{
		Runtime: documenttools.DocumentParserFunc(func(context.Context, pipeline.ParseRequest) (*pipelinetypes.DocumentResult, error) {
			return nil, rawErr
		}),
		Paths: documenttools.PathResolverFunc(func(_ context.Context, request documenttools.PathRequest) (documenttools.ResolvedPath, error) {
			return documenttools.ResolvedPath{Path: request.Value, Display: request.Value}, nil
		}),
		Errors: documenttools.ErrorProjectorFuncs{
			ClassifyFunc: func(err error) string {
				if errors.Is(err, rawErr) {
					return "provider_failure"
				}
				return "unknown"
			},
			DisplayFunc: func(error, string, string) string { return "safe projected failure" },
		},
	}
	if err := documenttools.RegisterDocumentParseTools(registry, documenttools.DocumentParseToolOptions{
		EnabledTools: []string{"document_parse"}, Host: host,
	}); err != nil {
		t.Fatal(err)
	}
	raw, err := registry.Execute(context.Background(), llm.FunctionCall{
		Name: "document_parse", Arguments: `{"document_path":"report.txt","spec_path":"spec","artifact_policy":"none"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(raw, rawErr.Error()) || !strings.Contains(raw, `"error_class":"provider_failure"`) || !strings.Contains(raw, "safe projected failure") {
		t.Fatalf("unexpected projected payload: %s", raw)
	}
}

type memoryPDFBackend struct{}

func (memoryPDFBackend) Name() string { return "memory" }
func (memoryPDFBackend) Availability(context.Context) documenttools.PDFBackendAvailability {
	return documenttools.PDFBackendAvailability{Name: "memory", Available: true}
}
func (memoryPDFBackend) ExtractAllText(context.Context, string) (documenttools.PDFTextResult, error) {
	return documenttools.PDFTextResult{Pages: []documenttools.PDFPageText{{Page: 1, Text: "Revenue 42"}}}, nil
}
func (memoryPDFBackend) ExtractPageText(_ context.Context, _ string, pages []int) (documenttools.PDFTextResult, error) {
	if len(pages) == 0 {
		pages = []int{1}
	}
	out := make([]documenttools.PDFPageText, 0, len(pages))
	for _, page := range pages {
		out = append(out, documenttools.PDFPageText{Page: page, Text: "Revenue 42"})
	}
	return documenttools.PDFTextResult{Pages: out}, nil
}
func (memoryPDFBackend) ReadMetadata(context.Context, string, bool) (documenttools.PDFMetadataResult, error) {
	return documenttools.PDFMetadataResult{PageCount: 1}, nil
}
