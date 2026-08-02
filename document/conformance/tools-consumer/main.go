package main

import (
	"context"
	"encoding/json"
	"fmt"

	llm "github.com/wsnacj/agentx-go/components/llm"
	documenttools "github.com/wsnacj/agentx-go/document/tools"
	agentxtools "github.com/wsnacj/agentx-go/tools"
)

type output struct {
	Status string `json:"status"`
	Answer string `json:"answer"`
}

func run(ctx context.Context) (output, error) {
	registry := agentxtools.NewRegistry()
	err := documenttools.RegisterPDFTools(registry, documenttools.PDFToolOptions{
		EnabledTools: []string{"pdf"},
		Backend:      memoryBackend{},
		Models: []documenttools.PDFModelCandidate{{
			Name: "memory-model", Client: "memory", Model: "memory-model", ConfigKey: "memory-model",
		}},
		Host: documenttools.PDFHost{
			Inputs: documenttools.PDFInputResolverFunc(func(ctx context.Context, request documenttools.PDFInputRequest) (documenttools.ResolvedPDFInput, error) {
				if err := ctx.Err(); err != nil {
					return documenttools.ResolvedPDFInput{}, err
				}
				return documenttools.ResolvedPDFInput{Path: request.Reference, Display: request.Reference}, nil
			}),
			Chat: func(ctx context.Context, _ llm.ChatInput) (*llm.ChatResponse, error) {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				return &llm.ChatResponse{Content: "canonical document tools [p1]"}, nil
			},
		},
	})
	if err != nil {
		return output{}, err
	}
	raw, err := registry.Execute(ctx, llm.FunctionCall{Name: "pdf", Arguments: `{"pdf":"memory.pdf","prompt":"summarize"}`})
	if err != nil {
		return output{}, err
	}
	var payload struct {
		Status string `json:"status"`
		Answer string `json:"answer"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return output{}, err
	}
	return output{Status: payload.Status, Answer: payload.Answer}, nil
}

func main() {
	result, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(payload))
}

type memoryBackend struct{}

func (memoryBackend) Name() string { return "memory" }
func (memoryBackend) Availability(context.Context) documenttools.PDFBackendAvailability {
	return documenttools.PDFBackendAvailability{Name: "memory", Available: true}
}
func (memoryBackend) ExtractAllText(context.Context, string) (documenttools.PDFTextResult, error) {
	return documenttools.PDFTextResult{Pages: []documenttools.PDFPageText{{Page: 1, Text: "portable document tools"}}}, nil
}
func (memoryBackend) ExtractPageText(context.Context, string, []int) (documenttools.PDFTextResult, error) {
	return documenttools.PDFTextResult{Pages: []documenttools.PDFPageText{{Page: 1, Text: "portable document tools"}}}, nil
}
func (memoryBackend) ReadMetadata(context.Context, string, bool) (documenttools.PDFMetadataResult, error) {
	return documenttools.PDFMetadataResult{PageCount: 1}, nil
}
