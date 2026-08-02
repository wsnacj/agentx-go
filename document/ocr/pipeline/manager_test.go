package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/wsnacj/agentx-go/document/ocr/cache"
	"github.com/wsnacj/agentx-go/document/ocr/config"
	"github.com/wsnacj/agentx-go/document/ocr/model"
	"github.com/wsnacj/agentx-go/document/ocr/processor/textin"
	"github.com/wsnacj/agentx-go/document/ocr/provider"
	"github.com/wsnacj/agentx-go/document/ocr/splitter"
	"github.com/wsnacj/agentx-go/document/ocr/worker"
)

type stubProvider struct {
	resp provider.Response
}

func (s stubProvider) Call(ctx context.Context, req provider.Request) (provider.Response, error) {
	return s.resp, nil
}

type countingProvider struct {
	resp  provider.Response
	mu    sync.Mutex
	calls int
}

func (c *countingProvider) Call(ctx context.Context, req provider.Request) (provider.Response, error) {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()
	return c.resp, nil
}

func (c *countingProvider) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

type cancelAwareProvider struct {
	resp        provider.Response
	waitStarted chan struct{}
	mu          sync.Mutex
	canceled    bool
}

func (c *cancelAwareProvider) Call(ctx context.Context, req provider.Request) (provider.Response, error) {
	if req.FilePath == "http://example.com/wait" {
		close(c.waitStarted)
		select {
		case <-ctx.Done():
			c.mu.Lock()
			c.canceled = true
			c.mu.Unlock()
			return provider.Response{}, ctx.Err()
		case <-time.After(2 * time.Second):
			return c.resp, nil
		}
	}
	<-c.waitStarted
	return provider.Response{}, errors.New("boom")
}

func (c *cancelAwareProvider) SawCancel() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.canceled
}

type capturingProvider struct {
	resp provider.Response
	mu   sync.Mutex
	last provider.Request
}

func (c *capturingProvider) Call(ctx context.Context, req provider.Request) (provider.Response, error) {
	c.mu.Lock()
	c.last = req
	c.mu.Unlock()
	return c.resp, nil
}

func (c *capturingProvider) Last() provider.Request {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.last
}

type noopSplitter struct{}

func (noopSplitter) Split(ctx context.Context, req splitter.Request) (splitter.Result, error) {
	return splitter.Result{Images: []string{req.Path}}, nil
}

type remoteSplitterStub struct {
	images []string
}

func (r remoteSplitterStub) Split(ctx context.Context, req splitter.Request) (splitter.Result, error) {
	return splitter.Result{Images: append([]string(nil), r.images...)}, nil
}

type memoryStore struct{}

func (memoryStore) Get(ctx context.Context, key string) (cache.Entry, bool, error) {
	return cache.Entry{}, false, nil
}

func (memoryStore) Set(ctx context.Context, key string, entry cache.Entry) error {
	return nil
}

type mapStore struct {
	mu    sync.Mutex
	items map[string]cache.Entry
}

func newMapStore() *mapStore {
	return &mapStore{items: make(map[string]cache.Entry)}
}

func (m *mapStore) Get(ctx context.Context, key string) (cache.Entry, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.items[key]
	return entry, ok, nil
}

func (m *mapStore) Set(ctx context.Context, key string, entry cache.Entry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = entry
	return nil
}

func TestManagerRunWithStubProvider(t *testing.T) {
	ocrPayload := map[string]any{
		"code":    200,
		"message": "OK",
		"result": map[string]any{
			"pages": []map[string]any{
				{
					"angle":  0,
					"width":  100,
					"height": 200,
					"lines": []map[string]any{
						{
							"text":                  "stub",
							"char_candidates":       [][]string{{"s"}},
							"char_candidates_score": [][]float64{{0.9}},
							"char_positions":        [][]int{{0, 0, 2, 2}},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(ocrPayload)
	if err != nil {
		t.Fatalf("marshal stub payload: %v", err)
	}

	mgrCfg := config.PipelineConfig{
		Provider: config.ProviderConfig{Kind: "textin"},
		Splitter: config.SplitterConfig{Kind: "noop"},
		Cache:    config.CacheConfig{Enabled: false},
		Worker:   config.WorkerConfig{MaxConcurrent: 2},
		Limits:   config.LimitConfig{MaxPages: 5, MaxFiles: 5},
		Retry:    config.RetryConfig{MaxAttempts: 1},
		Diff:     config.DiffConfig{Enabled: false},
	}

	providerStub := stubProvider{resp: provider.Response{Raw: data}}
	splitterStub := noopSplitter{}
	store := memoryStore{}
	pool := worker.NewPool(mgrCfg.Worker)
	procProvider, err := textin.NewProcessor(config.ProviderConfig{Kind: "textin"})
	if err != nil {
		t.Fatalf("new textin processor: %v", err)
	}
	opProc, err := procProvider.For(model.OperationKindOCR)
	if err != nil {
		t.Fatalf("operation processor: %v", err)
	}

	mgr, err := NewManager(model.OperationKindOCR, mgrCfg, providerStub, splitterStub, store, pool, opProc)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := mgr.Run(ctx, model.Request{
		Paths:     []string{"http://example.com/page.png"},
		Options:   map[string]any{"need_character": true},
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("run manager: %v", err)
	}

	ocrResp, ok := resp.Payload.(model.OCRPayload)
	if !ok {
		t.Fatalf("unexpected payload type %T", resp.Payload)
	}
	if ocrResp.RecognizedText == "" {
		t.Fatal("recognized text should not be empty")
	}
	if len(resp.Meta.ProcessedPaths) == 0 {
		t.Fatal("meta processed paths should not be empty")
	}
}

func TestManagerDiffSummary(t *testing.T) {
	baseline := []byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"Hello","char_candidates":[["H"],["e"],["l"],["l"],["o"]],"char_candidates_score":[[0.9],[0.9],[0.9],[0.9],[0.9]],"char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1],[3,0,4,1],[4,0,5,1]]}]}]}}`)
	current := []byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"Hello world","char_candidates":[["H"],["e"],["l"],["l"],["o"],[" "],["w"],["o"],["r"],["l"],["d"]],"char_candidates_score":[[0.9],[0.9],[0.9],[0.9],[0.9],[1],[0.9],[0.9],[0.9],[0.9],[0.9]],"char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1],[3,0,4,1],[4,0,5,1],[5,0,6,1],[6,0,7,1],[7,0,8,1],[8,0,9,1],[9,0,10,1],[10,0,11,1]]}]}]}}`)
	tmp, err := os.CreateTemp("", "baseline-*.json")
	if err != nil {
		t.Fatalf("create temp baseline: %v", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(baseline); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("close baseline: %v", err)
	}

	mgrCfg := config.PipelineConfig{
		Provider: config.ProviderConfig{Kind: "textin"},
		Splitter: config.SplitterConfig{Kind: "noop"},
		Cache:    config.CacheConfig{Enabled: false},
		Worker:   config.WorkerConfig{MaxConcurrent: 1},
		Limits:   config.LimitConfig{MaxPages: 5, MaxFiles: 5},
		Retry:    config.RetryConfig{MaxAttempts: 1},
		Diff:     config.DiffConfig{Enabled: true, Baseline: tmp.Name(), Preview: 2},
	}

	providerStub := stubProvider{resp: provider.Response{Raw: current}}
	procProvider, err := textin.NewProcessor(config.ProviderConfig{Kind: "textin"})
	if err != nil {
		t.Fatalf("new textin processor: %v", err)
	}
	opProc, err := procProvider.For(model.OperationKindOCR)
	if err != nil {
		t.Fatalf("operation processor: %v", err)
	}
	mgr, err := NewManager(model.OperationKindOCR, mgrCfg, providerStub, noopSplitter{}, memoryStore{}, worker.NewPool(mgrCfg.Worker), opProc)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := mgr.Run(ctx, model.Request{
		Paths: []string{"http://example.com/doc"},
	})
	if err != nil {
		t.Fatalf("run manager: %v", err)
	}
	if resp.Diff == nil || resp.Diff.OCRDiff == nil {
		t.Fatalf("expected diff summary")
	}
	if !resp.Diff.OCRDiff.HasDiff {
		t.Fatalf("expected diff to be detected")
	}
}

func TestManagerCachesLocalFileResults(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "sample.png")
	if err := os.WriteFile(imgPath, []byte("fake"), 0o644); err != nil {
		t.Fatalf("write temp image: %v", err)
	}

	ocrPayload := map[string]any{
		"code":    200,
		"message": "OK",
		"result": map[string]any{
			"pages": []map[string]any{
				{
					"angle":  0,
					"width":  100,
					"height": 200,
					"lines": []map[string]any{
						{
							"text":                  "stub",
							"char_candidates":       [][]string{{"s"}},
							"char_candidates_score": [][]float64{{0.9}},
							"char_positions":        [][]int{{0, 0, 2, 2}},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(ocrPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	mgrCfg := config.PipelineConfig{
		Provider: config.ProviderConfig{Kind: "textin"},
		Splitter: config.SplitterConfig{Kind: "noop"},
		Cache:    config.CacheConfig{Enabled: true},
		Worker:   config.WorkerConfig{MaxConcurrent: 1},
		Limits:   config.LimitConfig{MaxPages: 5, MaxFiles: 5},
		Retry:    config.RetryConfig{MaxAttempts: 1},
		Diff:     config.DiffConfig{Enabled: false},
	}

	provider := &countingProvider{resp: provider.Response{Raw: data}}
	procProvider, err := textin.NewProcessor(config.ProviderConfig{Kind: "textin"})
	if err != nil {
		t.Fatalf("new textin processor: %v", err)
	}
	opProc, err := procProvider.For(model.OperationKindOCR)
	if err != nil {
		t.Fatalf("operation processor: %v", err)
	}
	mgr, err := NewManager(model.OperationKindOCR, mgrCfg, provider, noopSplitter{}, newMapStore(), worker.NewPool(mgrCfg.Worker), opProc)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	req := model.Request{Paths: []string{imgPath}, Options: map[string]any{"need_character": true}, CreatedAt: time.Now()}
	ctx := context.Background()
	if _, err := mgr.Run(ctx, req); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if _, err := mgr.Run(ctx, req); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if got := provider.Count(); got != 1 {
		t.Fatalf("expected provider to be invoked once, got %d", got)
	}
}

func TestManagerInvalidatesLocalCacheWhenContentChangesWithoutMetadataChange(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "sample.png")
	original := []byte("fake")
	if err := os.WriteFile(imgPath, original, 0o644); err != nil {
		t.Fatalf("write temp image: %v", err)
	}
	info, err := os.Stat(imgPath)
	if err != nil {
		t.Fatalf("stat temp image: %v", err)
	}

	ocrPayload := map[string]any{
		"code":    200,
		"message": "OK",
		"result": map[string]any{
			"pages": []map[string]any{
				{
					"angle":  0,
					"width":  100,
					"height": 200,
					"lines": []map[string]any{
						{
							"text":                  "stub",
							"char_candidates":       [][]string{{"s"}},
							"char_candidates_score": [][]float64{{0.9}},
							"char_positions":        [][]int{{0, 0, 2, 2}},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(ocrPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	mgrCfg := config.PipelineConfig{
		Provider: config.ProviderConfig{Kind: "textin"},
		Splitter: config.SplitterConfig{Kind: "noop"},
		Cache:    config.CacheConfig{Enabled: true},
		Worker:   config.WorkerConfig{MaxConcurrent: 1},
		Limits:   config.LimitConfig{MaxPages: 5, MaxFiles: 5},
		Retry:    config.RetryConfig{MaxAttempts: 1},
		Diff:     config.DiffConfig{Enabled: false},
	}

	provider := &countingProvider{resp: provider.Response{Raw: data}}
	procProvider, err := textin.NewProcessor(config.ProviderConfig{Kind: "textin"})
	if err != nil {
		t.Fatalf("new textin processor: %v", err)
	}
	opProc, err := procProvider.For(model.OperationKindOCR)
	if err != nil {
		t.Fatalf("operation processor: %v", err)
	}
	mgr, err := NewManager(model.OperationKindOCR, mgrCfg, provider, noopSplitter{}, newMapStore(), worker.NewPool(mgrCfg.Worker), opProc)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	req := model.Request{Paths: []string{imgPath}, Options: map[string]any{"need_character": true}, CreatedAt: time.Now()}
	ctx := context.Background()
	if _, err := mgr.Run(ctx, req); err != nil {
		t.Fatalf("first run: %v", err)
	}

	updated := []byte("FakE")
	if len(updated) != len(original) {
		t.Fatalf("test fixture expects equal length")
	}
	if err := os.WriteFile(imgPath, updated, 0o644); err != nil {
		t.Fatalf("rewrite temp image: %v", err)
	}
	if err := os.Chtimes(imgPath, info.ModTime(), info.ModTime()); err != nil {
		t.Fatalf("restore mod time: %v", err)
	}

	if _, err := mgr.Run(ctx, req); err != nil {
		t.Fatalf("second run: %v", err)
	}
	if got := provider.Count(); got != 2 {
		t.Fatalf("expected provider to be invoked twice after content change, got %d", got)
	}
}

func TestManagerSeparatesCacheAcrossOperationKinds(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "sample.png")
	if err := os.WriteFile(imgPath, []byte("fake"), 0o644); err != nil {
		t.Fatalf("write temp image: %v", err)
	}

	ocrPayload := map[string]any{
		"code":    200,
		"message": "OK",
		"result": map[string]any{
			"pages": []map[string]any{
				{
					"angle":  0,
					"width":  100,
					"height": 200,
					"lines": []map[string]any{
						{
							"text":                  "stub",
							"char_candidates":       [][]string{{"s"}},
							"char_candidates_score": [][]float64{{0.9}},
							"char_positions":        [][]int{{0, 0, 2, 2}},
						},
					},
				},
			},
		},
	}
	stampPayload := map[string]any{
		"code":    200,
		"message": "OK",
		"result": map[string]any{
			"details": map[string]any{
				"stamp": []map[string]any{
					{
						"value":       "seal",
						"type":        "official",
						"stamp_shape": "round",
						"color":       "red",
						"position":    []int{0, 0, 10, 0, 10, 10, 0, 10},
					},
				},
			},
		},
	}

	ocrData, err := json.Marshal(ocrPayload)
	if err != nil {
		t.Fatalf("marshal ocr payload: %v", err)
	}
	stampData, err := json.Marshal(stampPayload)
	if err != nil {
		t.Fatalf("marshal stamp payload: %v", err)
	}

	mgrCfg := config.PipelineConfig{
		Provider: config.ProviderConfig{
			Kind:    "textin",
			BaseURL: "https://example.com/shared-provider",
		},
		Splitter: config.SplitterConfig{Kind: "noop"},
		Cache:    config.CacheConfig{Enabled: true},
		Worker:   config.WorkerConfig{MaxConcurrent: 1},
		Limits:   config.LimitConfig{MaxPages: 5, MaxFiles: 5},
		Retry:    config.RetryConfig{MaxAttempts: 1},
		Diff:     config.DiffConfig{Enabled: false},
	}

	procProvider, err := textin.NewProcessor(config.ProviderConfig{Kind: "textin"})
	if err != nil {
		t.Fatalf("new textin processor: %v", err)
	}
	ocrProc, err := procProvider.For(model.OperationKindOCR)
	if err != nil {
		t.Fatalf("ocr processor: %v", err)
	}
	stampProc, err := procProvider.For(model.OperationKindStamp)
	if err != nil {
		t.Fatalf("stamp processor: %v", err)
	}

	store := newMapStore()
	ocrProvider := &countingProvider{resp: provider.Response{Raw: ocrData}}
	stampProvider := &countingProvider{resp: provider.Response{Raw: stampData}}

	ocrMgr, err := NewManager(model.OperationKindOCR, mgrCfg, ocrProvider, noopSplitter{}, store, worker.NewPool(mgrCfg.Worker), ocrProc)
	if err != nil {
		t.Fatalf("new ocr manager: %v", err)
	}
	stampMgr, err := NewManager(model.OperationKindStamp, mgrCfg, stampProvider, noopSplitter{}, store, worker.NewPool(mgrCfg.Worker), stampProc)
	if err != nil {
		t.Fatalf("new stamp manager: %v", err)
	}

	req := model.Request{
		Paths:     []string{imgPath},
		Options:   map[string]any{"need_character": true},
		CreatedAt: time.Now(),
	}
	ctx := context.Background()

	if _, err := ocrMgr.Run(ctx, req); err != nil {
		t.Fatalf("ocr run: %v", err)
	}
	if got := ocrProvider.Count(); got != 1 {
		t.Fatalf("expected ocr provider to be invoked once, got %d", got)
	}

	resp, err := stampMgr.Run(ctx, req)
	if err != nil {
		t.Fatalf("stamp run: %v", err)
	}
	if got := stampProvider.Count(); got != 1 {
		t.Fatalf("expected stamp provider to be invoked once, got %d", got)
	}

	stampResp, ok := resp.Payload.(model.StampPayload)
	if !ok {
		t.Fatalf("unexpected stamp payload type %T", resp.Payload)
	}
	if len(stampResp.Pages) != 1 || len(stampResp.Pages[0].Stamp) != 1 {
		t.Fatalf("expected stamp detection to come from stamp provider, got %+v", stampResp.Pages)
	}
	if stampResp.Pages[0].Stamp[0].Text != "seal" {
		t.Fatalf("unexpected stamp text: %+v", stampResp.Pages[0].Stamp[0])
	}
}

func TestManagerDiffOverride(t *testing.T) {
	baseline := []byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"Hello","char_candidates":[["H"],["e"],["l"],["l"],["o"]],"char_candidates_score":[[0.9],[0.9],[0.9],[0.9],[0.9]],"char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1],[3,0,4,1],[4,0,5,1]]}]}]}}`)
	current := []byte(`{"code":200,"message":"OK","result":{"pages":[{"angle":0,"width":100,"height":100,"lines":[{"text":"Hello world","char_candidates":[["H"],["e"],["l"],["l"],["o"],[" "],["w"],["o"],["r"],["l"],["d"]],"char_candidates_score":[[0.9],[0.9],[0.9],[0.9],[0.9],[1],[0.9],[0.9],[0.9],[0.9],[0.9]],"char_positions":[[0,0,1,1],[1,0,2,1],[2,0,3,1],[3,0,4,1],[4,0,5,1],[5,0,6,1],[6,0,7,1],[7,0,8,1],[8,0,9,1],[9,0,10,1],[10,0,11,1]]}]}]}}`)
	tmp, err := os.CreateTemp(t.TempDir(), "baseline-*.json")
	if err != nil {
		t.Fatalf("create temp baseline: %v", err)
	}
	if _, err := tmp.Write(baseline); err != nil {
		t.Fatalf("write baseline: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("close baseline: %v", err)
	}

	mgrCfg := config.PipelineConfig{
		Provider: config.ProviderConfig{Kind: "textin"},
		Splitter: config.SplitterConfig{Kind: "noop"},
		Cache:    config.CacheConfig{Enabled: false},
		Worker:   config.WorkerConfig{MaxConcurrent: 1},
		Limits:   config.LimitConfig{MaxPages: 5, MaxFiles: 5},
		Retry:    config.RetryConfig{MaxAttempts: 1},
		Diff:     config.DiffConfig{Enabled: false, Preview: 2},
	}

	providerStub := stubProvider{resp: provider.Response{Raw: current}}
	procProvider, err := textin.NewProcessor(config.ProviderConfig{Kind: "textin"})
	if err != nil {
		t.Fatalf("new textin processor: %v", err)
	}
	opProc, err := procProvider.For(model.OperationKindOCR)
	if err != nil {
		t.Fatalf("operation processor: %v", err)
	}
	mgr, err := NewManager(model.OperationKindOCR, mgrCfg, providerStub, noopSplitter{}, memoryStore{}, worker.NewPool(mgrCfg.Worker), opProc)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	req := model.Request{Paths: []string{"http://example.com/doc"}, Options: map[string]any{"diff_enabled": true, "diff_baseline": tmp.Name()}}
	resp, err := mgr.Run(context.Background(), req)
	if err != nil {
		t.Fatalf("run with diff override: %v", err)
	}
	if resp.Diff == nil || resp.Diff.OCRDiff == nil {
		t.Fatalf("expected diff summary when override enabled")
	}

	reqDisable := model.Request{Paths: []string{"http://example.com/doc"}, Options: map[string]any{"diff_enabled": false, "diff_baseline": tmp.Name()}}
	respDisable, err := mgr.Run(context.Background(), reqDisable)
	if err != nil {
		t.Fatalf("run with diff disabled: %v", err)
	}
	if respDisable.Diff != nil {
		t.Fatalf("expected diff to be disabled by override")
	}
}

func TestManagerRemoteSplitterMarksRemotePayload(t *testing.T) {
	ocrPayload := map[string]any{
		"code":    200,
		"message": "OK",
		"result": map[string]any{
			"pages": []map[string]any{
				{
					"angle":  0,
					"width":  100,
					"height": 200,
					"lines": []map[string]any{
						{
							"text":                  "remote",
							"char_candidates":       [][]string{{"r"}},
							"char_candidates_score": [][]float64{{0.9}},
							"char_positions":        [][]int{{0, 0, 2, 2}},
						},
					},
				},
			},
		},
	}
	data, err := json.Marshal(ocrPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	providerStub := &capturingProvider{resp: provider.Response{Raw: data}}
	mgrCfg := config.PipelineConfig{
		Provider: config.ProviderConfig{Kind: "textin"},
		Splitter: config.SplitterConfig{Kind: "noop"},
		Cache:    config.CacheConfig{Enabled: false},
		Worker:   config.WorkerConfig{MaxConcurrent: 1},
		Limits:   config.LimitConfig{MaxPages: 5, MaxFiles: 5},
		Retry:    config.RetryConfig{MaxAttempts: 1},
		Diff:     config.DiffConfig{Enabled: false},
	}

	dir := t.TempDir()
	pdfPath := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(pdfPath, []byte("pdf"), 0o644); err != nil {
		t.Fatalf("write pdf: %v", err)
	}

	remoteImages := []string{"https://cdn.example.com/page-1.png"}
	procProvider, err := textin.NewProcessor(config.ProviderConfig{Kind: "textin"})
	if err != nil {
		t.Fatalf("new textin processor: %v", err)
	}
	opProc, err := procProvider.For(model.OperationKindOCR)
	if err != nil {
		t.Fatalf("operation processor: %v", err)
	}
	mgr, err := NewManager(model.OperationKindOCR, mgrCfg, providerStub, remoteSplitterStub{images: remoteImages}, memoryStore{}, worker.NewPool(mgrCfg.Worker), opProc)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	req := model.Request{Paths: []string{pdfPath}, CreatedAt: time.Now()}
	if _, err := mgr.Run(context.Background(), req); err != nil {
		t.Fatalf("run manager: %v", err)
	}
	last := providerStub.Last()
	if !last.IsRemote {
		t.Fatalf("expected remote payload to be marked as remote")
	}
	if last.FilePath != remoteImages[0] {
		t.Fatalf("unexpected payload path: %s", last.FilePath)
	}
}

func TestManagerCancelsSiblingJobsOnError(t *testing.T) {
	ocrPayload := map[string]any{
		"code":    200,
		"message": "OK",
		"result": map[string]any{
			"pages": []map[string]any{
				{
					"angle":  0,
					"width":  100,
					"height": 200,
					"lines": []map[string]any{
						{
							"text":                  "ok",
							"char_candidates":       [][]string{{"o"}},
							"char_candidates_score": [][]float64{{0.9}},
							"char_positions":        [][]int{{0, 0, 2, 2}},
						},
					},
				},
			},
		},
	}
	data, err := json.Marshal(ocrPayload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	providerStub := &cancelAwareProvider{
		resp:        provider.Response{Raw: data},
		waitStarted: make(chan struct{}),
	}
	mgrCfg := config.PipelineConfig{
		Provider: config.ProviderConfig{Kind: "textin"},
		Splitter: config.SplitterConfig{Kind: "noop"},
		Cache:    config.CacheConfig{Enabled: false},
		Worker:   config.WorkerConfig{MaxConcurrent: 2},
		Limits:   config.LimitConfig{MaxPages: 5, MaxFiles: 5},
		Retry:    config.RetryConfig{MaxAttempts: 1},
		Diff:     config.DiffConfig{Enabled: false},
	}

	procProvider, err := textin.NewProcessor(config.ProviderConfig{Kind: "textin"})
	if err != nil {
		t.Fatalf("new textin processor: %v", err)
	}
	opProc, err := procProvider.For(model.OperationKindOCR)
	if err != nil {
		t.Fatalf("operation processor: %v", err)
	}
	mgr, err := NewManager(model.OperationKindOCR, mgrCfg, providerStub, noopSplitter{}, memoryStore{}, worker.NewPool(mgrCfg.Worker), opProc)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	start := time.Now()
	_, err = mgr.Run(ctx, model.Request{
		Paths: []string{
			"http://example.com/wait",
			"http://example.com/fail",
		},
	})
	if err == nil {
		t.Fatalf("expected run to fail")
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Fatalf("expected fail-fast cancellation, took %s", elapsed)
	}
	if !providerStub.SawCancel() {
		t.Fatalf("expected sibling job to observe cancellation")
	}
}
