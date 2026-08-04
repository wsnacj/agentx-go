// document 展示无文件、无网络的 PDF Tool Host 组合。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	llm "github.com/wsnacj/agentx-go/components/llm"
	documenttools "github.com/wsnacj/agentx-go/document/tools"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

type documentResult struct {
	Status string `json:"status"`
	Answer string `json:"answer"`
}

func run(ctx context.Context) (documentResult, error) {
	registry := agentxtools.NewRegistry()
	err := documenttools.RegisterPDFTools(registry, documenttools.PDFToolOptions{
		EnabledTools: []string{"pdf"},
		Backend:      memoryPDF{},
		Models: []documenttools.PDFModelCandidate{{
			Name: "fixture", Client: "memory", Model: "fixture", ConfigKey: "fixture",
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
					return nil, fmt.Errorf("fixture expected grounded PDF context")
				}
				return &llm.ChatResponse{Content: "Revenue is 42 [p1]."}, nil
			},
		},
	})
	if err != nil {
		return documentResult{}, err
	}
	raw, err := registry.Execute(ctx, llm.FunctionCall{
		Name: "pdf", Arguments: `{"pdf":"memory.pdf","prompt":"What is revenue?"}`,
	})
	if err != nil {
		return documentResult{}, err
	}
	var result documentResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return documentResult{}, err
	}
	return result, nil
}

func main() {
	result, err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoded, _ := json.Marshal(result)
	fmt.Println(string(encoded))
}

type memoryPDF struct{}

func (memoryPDF) Name() string { return "memory" }
func (memoryPDF) Availability(context.Context) documenttools.PDFBackendAvailability {
	return documenttools.PDFBackendAvailability{Name: "memory", Available: true}
}
func (memoryPDF) ExtractAllText(context.Context, string) (documenttools.PDFTextResult, error) {
	return documenttools.PDFTextResult{Pages: []documenttools.PDFPageText{{Page: 1, Text: "Revenue 42"}}}, nil
}
func (memoryPDF) ExtractPageText(context.Context, string, []int) (documenttools.PDFTextResult, error) {
	return documenttools.PDFTextResult{Pages: []documenttools.PDFPageText{{Page: 1, Text: "Revenue 42"}}}, nil
}
func (memoryPDF) ReadMetadata(context.Context, string, bool) (documenttools.PDFMetadataResult, error) {
	return documenttools.PDFMetadataResult{PageCount: 1}, nil
}
