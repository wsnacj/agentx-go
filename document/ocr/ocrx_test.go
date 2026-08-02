package ocrx

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	ocrxconfig "github.com/wsnacj/agentx-go/document/ocr/config"
	"github.com/wsnacj/agentx-go/document/ocr/processor"
	textinproc "github.com/wsnacj/agentx-go/document/ocr/processor/textin"
	"github.com/wsnacj/agentx-go/document/ocr/provider"
	"github.com/wsnacj/agentx-go/document/ocr/splitter"
)

const (
	stubOCRJSON   = `{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"Hello","char_candidates":[["H"]],"char_candidates_score":[[0.9]],"char_positions":[[0,0,1,1]]}]},{"angle":0,"width":100,"height":100,"lines":[{"text":"World","char_candidates":[["W"]],"char_candidates_score":[[0.9]],"char_positions":[[0,0,1,1]]}]}]}}`
	stubTableJSON = `{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"tables":[{"lines":[{"text":"Row","char_candidates":[["R"],["o"],["w"]],"char_candidates_score":[[0.9],[0.9],[0.9]],"char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1]]}],"table_cells":[]}]}]}}`
	stubStampJSON = `{"code":200,"message":"OK","result":{"details":{"stamp":[{"value":"seal","type":"official","stamp_shape":"round","color":"red","position":[0,0,10,10]}]}}}`
)

func TestRecognizeText(t *testing.T) {
	client, file := newStubClient(t, ModeOCR, stubOCRJSON)
	res, err := client.RecognizeText(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Text == "" {
		t.Fatalf("expected text, got empty")
	}
}

func TestRecognizeHTML(t *testing.T) {
	client, file := newStubClient(t, ModeOCR, stubOCRJSON)
	res, err := client.RecognizeHTML(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Pages) == 0 || res.HTML == "" {
		t.Fatalf("unexpected html result: %+v", res)
	}
}

func TestRecognizeTextPages(t *testing.T) {
	client, file := newStubClient(t, ModeOCR, stubOCRJSON)
	res, err := client.RecognizeTextPages(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Pages) != 2 || res.Pages[0].Text == "" {
		t.Fatalf("unexpected pages: %+v", res.Pages)
	}
}

func TestRecognizeTable(t *testing.T) {
	client, file := newStubClient(t, ModeTable, stubTableJSON)
	res, err := client.RecognizeTable(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Pages) != 1 || res.Pages[0].Text == "" {
		t.Fatalf("unexpected table result: %+v", res)
	}
}

func TestRecognizeTableHonorsMaxPagesOption(t *testing.T) {
	gotMaxPages := 0
	client, file := newStubClientWithSplitter(t, ModeTable, stubTableJSON, captureSplitter{maxPages: &gotMaxPages}, ".pdf")
	if _, err := client.RecognizeTable(file, WithMaxPages(2)); err != nil {
		t.Fatalf("RecognizeTable returned error: %v", err)
	}
	if gotMaxPages != 2 {
		t.Fatalf("expected splitter max pages 2, got %d", gotMaxPages)
	}
}

func TestRecognizeStamp(t *testing.T) {
	client, file := newStubClient(t, ModeStamp, stubStampJSON)
	res, err := client.RecognizeStamp(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Pages) == 0 || len(res.Pages[0].Stamps) == 0 {
		t.Fatalf("unexpected stamp result: %+v", res)
	}
}

func TestRecognizeFiles(t *testing.T) {
	client, file := newStubClient(t, ModeOCR, stubOCRJSON)
	results := client.RecognizeFiles([]string{file}, ModeOCR)
	if len(results) != 1 || results[0].Err != nil {
		t.Fatalf("unexpected batch result: %+v", results)
	}
}

func TestPreviewAndWriteHelpers(t *testing.T) {
	outDir := t.TempDir()
	if err := WriteText(filepath.Join(outDir, "text.txt"), "hello"); err != nil {
		t.Fatalf("WriteText failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "text.txt")); err != nil {
		t.Fatalf("expected file: %v", err)
	}
	if err := WriteJSONResult(filepath.Join(outDir, "result.json"), map[string]string{"a": "b"}); err != nil {
		t.Fatalf("WriteJSONResult failed: %v", err)
	}
	prev := PreviewText("hello world", 4)
	if prev != "hell..." {
		t.Fatalf("unexpected preview: %s", prev)
	}
}

func TestSupportedExt(t *testing.T) {
	cases := []struct {
		ext string
		ok  bool
	}{
		{".pdf", true},
		{".png", true},
		{".gif", true},
		{".webp", true},
		{".bmp", true},
		{".unknown", false},
	}
	for _, c := range cases {
		if supportedExt(c.ext) != c.ok {
			t.Fatalf("supportedExt(%s) expected %v", c.ext, c.ok)
		}
	}
}

// helper creating client with stub provider.
func newStubClient(t *testing.T, mode Mode, respJSON string) (*Client, string) {
	t.Helper()
	return newStubClientWithSplitter(t, mode, respJSON, stubSplitter{}, ".png")
}

func newStubClientWithSplitter(t *testing.T, mode Mode, respJSON string, split splitter.Splitter, ext string) (*Client, string) {
	t.Helper()
	cfg := ocrxconfig.ServiceConfig{Pipelines: map[string]ocrxconfig.PipelineConfig{
		string(mode): {
			Provider: ocrxconfig.ProviderConfig{Kind: "stub"},
			Splitter: ocrxconfig.SplitterConfig{Kind: "noop"},
			Cache:    ocrxconfig.CacheConfig{Enabled: false},
			Worker:   ocrxconfig.WorkerConfig{MaxConcurrent: 1},
			Limits:   ocrxconfig.LimitConfig{MaxPages: 10, MaxFiles: 10},
			Retry:    ocrxconfig.RetryConfig{MaxAttempts: 1},
		},
	}}

	deps := Dependencies{
		ProviderFactories: map[string]provider.Factory{
			"stub": func(cfg ocrxconfig.ProviderConfig) (provider.Provider, error) {
				return stubProvider{data: []byte(respJSON)}, nil
			},
		},
		SplitterFactories: map[string]splitter.Factory{
			"noop": func(cfg ocrxconfig.SplitterConfig) (splitter.Splitter, error) {
				return split, nil
			},
		},
		ProcessorFactories: processor.Registry{
			"stub": func(cfg ocrxconfig.ProviderConfig) (processor.ProviderProcessor, error) {
				return textinproc.NewProcessor(ocrxconfig.ProviderConfig{Kind: "textin"})
			},
		},
	}

	svc, err := NewService(cfg, deps)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	client := NewClientFromService(svc)

	tmpFile := createTempFile(t, ext)
	return client, tmpFile
}

func createTempFile(t *testing.T, ext string) string {
	f, err := os.CreateTemp(t.TempDir(), "*"+ext)
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.Write([]byte("stub")); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	return f.Name()
}

type stubProvider struct {
	data []byte
}

func (s stubProvider) Call(ctx context.Context, req provider.Request) (provider.Response, error) {
	return provider.Response{Raw: append([]byte(nil), s.data...)}, nil
}

type stubSplitter struct{}

func (stubSplitter) Split(ctx context.Context, req splitter.Request) (splitter.Result, error) {
	return splitter.Result{Images: []string{req.Path}}, nil
}

type captureSplitter struct {
	maxPages *int
}

func (s captureSplitter) Split(ctx context.Context, req splitter.Request) (splitter.Result, error) {
	if s.maxPages != nil {
		*s.maxPages = req.MaxPages
	}
	return splitter.Result{Images: []string{req.Path}}, nil
}
