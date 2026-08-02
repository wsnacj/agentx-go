package tools_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
	documenttools "github.com/wsnacj/agentx-go/document/tools"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

func TestPDFSpecialistContractsExecuteThroughCanonicalOwner(t *testing.T) {
	registry := agentxtools.NewRegistry()
	backend := contractPDFBackend{}
	err := documenttools.RegisterPDFTools(registry, documenttools.PDFToolOptions{
		EnabledTools: []string{"pdf_extract", "pdf_read_pages", "pdf_outline", "pdf_analyze", "pdf_extract_structured"},
		Backend:      backend,
		Host: documenttools.PDFHost{Inputs: documenttools.PDFInputResolverFunc(func(ctx context.Context, request documenttools.PDFInputRequest) (documenttools.ResolvedPDFInput, error) {
			if err := ctx.Err(); err != nil {
				return documenttools.ResolvedPDFInput{}, err
			}
			return documenttools.ResolvedPDFInput{Path: request.Reference, Display: request.Reference}, nil
		})},
	})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name      string
		arguments string
		keys      []string
	}{
		{name: "pdf_extract", arguments: `{"pdf":"report.pdf","include_pages":true,"include_outline":true}`, keys: []string{"path", "text", "page_count"}},
		{name: "pdf_read_pages", arguments: `{"pdf":"report.pdf","pages":[2]}`, keys: []string{"path", "pages", "page_count"}},
		{name: "pdf_outline", arguments: `{"pdf":"report.pdf"}`, keys: []string{"path", "outline", "page_count"}},
		{name: "pdf_analyze", arguments: `{"pdf":"report.pdf","query":"revenue","include_page_map":true}`, keys: []string{"path", "analysis_plan", "page_count"}},
		{name: "pdf_extract_structured", arguments: `{"pdf":"report.pdf","query":"revenue","include_page_map":true}`, keys: []string{"path", "structured_batches", "page_count"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := registry.Execute(context.Background(), llm.FunctionCall{Name: tc.name, Arguments: tc.arguments})
			if err != nil {
				t.Fatalf("execute %s: %v", tc.name, err)
			}
			var payload map[string]any
			if err := json.Unmarshal([]byte(raw), &payload); err != nil {
				t.Fatalf("decode %s: %v\n%s", tc.name, err, raw)
			}
			for _, key := range tc.keys {
				if _, ok := payload[key]; !ok {
					t.Fatalf("%s response missing %q: %#v", tc.name, key, payload)
				}
			}
		})
	}
}

func TestPDFUnifiedMultiInputUsesHostResolverAndCleanup(t *testing.T) {
	registry := agentxtools.NewRegistry()
	var cleanups atomic.Int32
	var grounded atomic.Bool
	err := documenttools.RegisterPDFTools(registry, documenttools.PDFToolOptions{
		EnabledTools: []string{"pdf"},
		Backend:      contractPDFBackend{},
		Models: []documenttools.PDFModelCandidate{{
			Name: "memory", Client: "memory", Model: "memory", ConfigKey: "memory",
		}},
		Host: documenttools.PDFHost{
			Inputs: documenttools.PDFInputResolverFunc(func(ctx context.Context, request documenttools.PDFInputRequest) (documenttools.ResolvedPDFInput, error) {
				if err := ctx.Err(); err != nil {
					return documenttools.ResolvedPDFInput{}, err
				}
				return documenttools.ResolvedPDFInput{
					Path: request.Reference, Display: request.Reference,
					Cleanup: func() { cleanups.Add(1) },
				}, nil
			}),
			Chat: func(ctx context.Context, input llm.ChatInput) (*llm.ChatResponse, error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				raw, _ := json.Marshal(input.Messages)
				text := string(raw)
				grounded.Store(strings.Contains(text, "alpha.pdf") && strings.Contains(text, "beta.pdf"))
				return &llm.ChatResponse{Content: "comparison [PDF 1 p1] [PDF 2 p1]"}, nil
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := registry.Execute(context.Background(), llm.FunctionCall{
		Name: "pdf", Arguments: `{"pdfs":["alpha.pdf","beta.pdf"],"prompt":"compare revenue"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(raw, `"status":"success"`) || !grounded.Load() {
		t.Fatalf("unexpected multi-input result: grounded=%t raw=%s", grounded.Load(), raw)
	}
	if got := cleanups.Load(); got != 2 {
		t.Fatalf("expected exactly two cleanups, got %d", got)
	}
}

func TestPDFResolverFailureCleansPreviouslyMaterializedInputs(t *testing.T) {
	registry := agentxtools.NewRegistry()
	var cleanups atomic.Int32
	errBoom := errors.New("resolver boom")
	err := documenttools.RegisterPDFTools(registry, documenttools.PDFToolOptions{
		EnabledTools: []string{"pdf_analyze"}, Backend: contractPDFBackend{},
		Host: documenttools.PDFHost{Inputs: documenttools.PDFInputResolverFunc(func(_ context.Context, request documenttools.PDFInputRequest) (documenttools.ResolvedPDFInput, error) {
			if request.Reference == "bad.pdf" {
				return documenttools.ResolvedPDFInput{}, errBoom
			}
			return documenttools.ResolvedPDFInput{Path: request.Reference, Display: request.Reference, Cleanup: func() { cleanups.Add(1) }}, nil
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = registry.Execute(context.Background(), llm.FunctionCall{Name: "pdf_analyze", Arguments: `{"pdfs":["good.pdf","bad.pdf"]}`})
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected resolver error identity, got %v", err)
	}
	if got := cleanups.Load(); got != 1 {
		t.Fatalf("expected prior input cleanup, got %d", got)
	}
}

type contractPDFBackend struct{}

func (contractPDFBackend) Name() string { return "contract" }
func (contractPDFBackend) Availability(context.Context) documenttools.PDFBackendAvailability {
	return documenttools.PDFBackendAvailability{Name: "contract", Available: true}
}
func (contractPDFBackend) ExtractAllText(_ context.Context, path string) (documenttools.PDFTextResult, error) {
	return documenttools.PDFTextResult{Pages: []documenttools.PDFPageText{
		{Page: 1, Text: "Revenue for " + path + " is 42."},
		{Page: 2, Text: "Operating income is 7."},
	}}, nil
}
func (contractPDFBackend) ExtractPageText(ctx context.Context, path string, selected []int) (documenttools.PDFTextResult, error) {
	all, err := (contractPDFBackend{}).ExtractAllText(ctx, path)
	if err != nil || len(selected) == 0 {
		return all, err
	}
	allowed := make(map[int]struct{}, len(selected))
	for _, page := range selected {
		allowed[page] = struct{}{}
	}
	out := make([]documenttools.PDFPageText, 0, len(selected))
	for _, page := range all.Pages {
		if _, ok := allowed[page.Page]; ok {
			out = append(out, page)
		}
	}
	return documenttools.PDFTextResult{Pages: out}, nil
}
func (contractPDFBackend) ReadMetadata(context.Context, string, bool) (documenttools.PDFMetadataResult, error) {
	return documenttools.PDFMetadataResult{
		PageCount: 2,
		Outline:   &documenttools.PDFOutline{Children: []documenttools.PDFOutline{{Title: "Summary"}}},
	}, nil
}
