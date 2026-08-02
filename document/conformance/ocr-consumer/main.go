package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	ocrx "github.com/wsnacj/agentx-go/document/ocr"
	"github.com/wsnacj/agentx-go/document/ocr/config"
	"github.com/wsnacj/agentx-go/document/ocr/model"
	"github.com/wsnacj/agentx-go/document/ocr/processor"
	"github.com/wsnacj/agentx-go/document/ocr/provider"
	"github.com/wsnacj/agentx-go/document/ocr/splitter"
)

type output struct {
	Status string `json:"status"`
	Text   string `json:"text"`
}

func run(ctx context.Context) (output, error) {
	service, err := ocrx.NewService(config.ServiceConfig{Pipelines: map[string]config.PipelineConfig{
		"ocr": {
			Provider: config.ProviderConfig{Kind: "memory"},
			Splitter: config.SplitterConfig{Kind: "passthrough"},
			Worker:   config.WorkerConfig{MaxConcurrent: 1},
		},
	}}, ocrx.Dependencies{
		ProviderFactories: map[string]provider.Factory{
			"memory": func(config.ProviderConfig) (provider.Provider, error) {
				return memoryProvider{}, nil
			},
		},
		SplitterFactories: map[string]splitter.Factory{
			"passthrough": func(config.SplitterConfig) (splitter.Splitter, error) {
				return passthroughSplitter{}, nil
			},
		},
		ProcessorFactories: processor.Registry{
			"memory": func(config.ProviderConfig) (processor.ProviderProcessor, error) {
				return memoryProcessor{}, nil
			},
		},
	})
	if err != nil {
		return output{}, err
	}
	file, err := os.CreateTemp("", "agentx-ocr-consumer-*.png")
	if err != nil {
		return output{}, err
	}
	path := file.Name()
	_ = file.Close()
	defer os.Remove(path)

	result, err := ocrx.NewClientFromService(service).RecognizeText(path, ocrx.WithContext(ctx))
	if err != nil {
		return output{}, err
	}
	return output{Status: "recognized", Text: result.Text}, nil
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

type memoryProvider struct{}

func (memoryProvider) Call(ctx context.Context, _ provider.Request) (provider.Response, error) {
	if err := ctx.Err(); err != nil {
		return provider.Response{}, err
	}
	return provider.Response{Raw: []byte("canonical document")}, nil
}

type passthroughSplitter struct{}

func (passthroughSplitter) Split(_ context.Context, req splitter.Request) (splitter.Result, error) {
	return splitter.Result{Images: []string{req.Path}}, nil
}

type memoryProcessor struct{}

func (memoryProcessor) For(kind model.OperationKind) (processor.OperationProcessor, error) {
	if kind != model.OperationKindOCR {
		return nil, fmt.Errorf("unsupported operation %s", kind)
	}
	return memoryOperation{}, nil
}

type memoryOperation struct{}

func (memoryOperation) Build(raw [][]byte, files []string) (any, error) {
	text := string(raw[0])
	return model.OCRPayload{Files: files, RawResponses: raw, RecognizedText: text, PageTexts: []string{text}}, nil
}

func (memoryOperation) Diff([][]byte, []byte, int) (*model.DiffSummary, error) { return nil, nil }
