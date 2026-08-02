package ocrx_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	ocrx "github.com/wsnacj/agentx-go/document/ocr"
	"github.com/wsnacj/agentx-go/document/ocr/config"
	"github.com/wsnacj/agentx-go/document/ocr/model"
	"github.com/wsnacj/agentx-go/document/ocr/processor"
	"github.com/wsnacj/agentx-go/document/ocr/provider"
	"github.com/wsnacj/agentx-go/document/ocr/splitter"
)

func TestExternalConsumerUsesExplicitAdapters(t *testing.T) {
	input := writeInput(t, "agentx document")
	client := newFakeClient(t, fakeProvider{raw: []byte("agentx document")})

	result, err := client.RecognizeText(input, ocrx.WithContext(context.Background()))
	if err != nil {
		t.Fatalf("RecognizeText() error = %v", err)
	}
	if result.Text != "agentx document" {
		t.Fatalf("RecognizeText() text = %q", result.Text)
	}
}

func TestExternalConsumerPropagatesCancellation(t *testing.T) {
	input := writeInput(t, "cancel")
	entered := make(chan struct{})
	client := newFakeClient(t, fakeProvider{wait: true, entered: entered})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-entered
		cancel()
	}()

	_, err := client.RecognizeText(input, ocrx.WithContext(ctx))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RecognizeText() error = %v, want context.Canceled", err)
	}
}

func newFakeClient(t *testing.T, upstream provider.Provider) *ocrx.Client {
	t.Helper()
	cfg := config.ServiceConfig{Pipelines: map[string]config.PipelineConfig{
		string(model.OperationKindOCR): {
			Provider: config.ProviderConfig{Kind: "fake", Timeout: time.Second},
			Splitter: config.SplitterConfig{Kind: "passthrough"},
			Worker:   config.WorkerConfig{MaxConcurrent: 1},
		},
	}}
	service, err := ocrx.NewService(cfg, ocrx.Dependencies{
		ProviderFactories: map[string]provider.Factory{
			"fake": func(config.ProviderConfig) (provider.Provider, error) { return upstream, nil },
		},
		SplitterFactories: map[string]splitter.Factory{
			"passthrough": func(config.SplitterConfig) (splitter.Splitter, error) {
				return passthroughSplitter{}, nil
			},
		},
		ProcessorFactories: processor.Registry{
			"fake": func(config.ProviderConfig) (processor.ProviderProcessor, error) {
				return fakeProviderProcessor{}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return ocrx.NewClientFromService(service)
}

type fakeProvider struct {
	raw     []byte
	wait    bool
	entered chan<- struct{}
}

func (p fakeProvider) Call(ctx context.Context, _ provider.Request) (provider.Response, error) {
	if p.wait {
		if p.entered != nil {
			close(p.entered)
		}
		<-ctx.Done()
		return provider.Response{}, ctx.Err()
	}
	return provider.Response{Raw: append([]byte(nil), p.raw...), MediaType: "text/plain"}, nil
}

type passthroughSplitter struct{}

func (passthroughSplitter) Split(_ context.Context, req splitter.Request) (splitter.Result, error) {
	return splitter.Result{Images: []string{req.Path}}, nil
}

type fakeProviderProcessor struct{}

func (fakeProviderProcessor) For(kind model.OperationKind) (processor.OperationProcessor, error) {
	if kind != model.OperationKindOCR {
		return nil, errors.New("unsupported fake operation")
	}
	return fakeOperationProcessor{}, nil
}

type fakeOperationProcessor struct{}

func (fakeOperationProcessor) Build(raw [][]byte, files []string) (any, error) {
	text := ""
	if len(raw) > 0 {
		text = string(raw[0])
	}
	return model.OCRPayload{
		Files:          append([]string(nil), files...),
		RawResponses:   raw,
		RecognizedText: text,
		PageTexts:      []string{text},
	}, nil
}

func (fakeOperationProcessor) Diff([][]byte, []byte, int) (*model.DiffSummary, error) {
	return nil, nil
}

func writeInput(t *testing.T, content string) string {
	t.Helper()
	path := t.TempDir() + "/input.txt"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
