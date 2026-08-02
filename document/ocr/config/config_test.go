package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestServiceConfigValidateRequiresPipelines(t *testing.T) {
	err := (ServiceConfig{}).Validate()
	if err == nil || !strings.Contains(err.Error(), "no pipelines configured") {
		t.Fatalf("expected missing pipelines error, got %v", err)
	}
}

func TestServiceConfigValidateWrapsNestedErrors(t *testing.T) {
	cfg := ServiceConfig{
		Pipelines: map[string]PipelineConfig{
			"ocr": {
				Provider: ProviderConfig{Kind: "textin"},
				Splitter: SplitterConfig{Kind: "poppler", DPI: -1},
				Cache:    CacheConfig{},
				Worker:   WorkerConfig{},
				Limits:   LimitConfig{},
				Retry:    RetryConfig{},
				Diff:     DiffConfig{},
			},
		},
	}

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "pipeline ocr: splitter: dpi cannot be negative") {
		t.Fatalf("expected wrapped splitter error, got %v", err)
	}
}

func TestLoadParsesYAMLAndValidates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ocrx.yaml")
	content := `
pipelines:
  ocr:
    provider:
      kind: textin
      base_url: https://example.com/ocr
      auth:
        app_id: demo
        secret_code: secret
      timeout: 3s
    splitter:
      kind: poppler
      dpi: 300
      batch_pages: 2
      max_parallel: 3
    cache:
      enabled: true
      kind: fs
      ttl: 1h
      base_dir: ./tmp/cache
      max_size_mb: 128
    worker:
      max_concurrent: 2
      queue_size: 1
    limits:
      max_pages: 10
      max_files: 5
      max_per_file_page: 3
    retry:
      max_attempts: 2
      backoff: 1s
      max_backoff: 5s
    diff:
      enabled: true
      baseline: ./baseline.json
      preview: 4
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	pipe := cfg.Pipelines["ocr"]
	if pipe.Provider.Timeout != 3*time.Second {
		t.Fatalf("unexpected provider timeout: %s", pipe.Provider.Timeout)
	}
	if pipe.Cache.TTL != time.Hour {
		t.Fatalf("unexpected cache ttl: %s", pipe.Cache.TTL)
	}
	if pipe.Splitter.BatchPages != 2 || pipe.Splitter.MaxParallel != 3 {
		t.Fatalf("unexpected splitter config: %+v", pipe.Splitter)
	}
	if !pipe.Diff.Enabled || pipe.Diff.Preview != 4 {
		t.Fatalf("unexpected diff config: %+v", pipe.Diff)
	}
}

func TestLoadReturnsValidationError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
	content := `
pipelines:
  ocr:
    provider:
      kind: textin
    splitter:
      kind: poppler
      dpi: 300
    cache:
      enabled: true
      ttl: -1s
    worker:
      max_concurrent: 1
    limits:
      max_pages: 1
    retry:
      max_attempts: 1
    diff:
      preview: 1
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "cache: ttl cannot be negative") {
		t.Fatalf("expected cache validation error, got %v", err)
	}
}

func TestDefaultTextInConfigBuildsIndependentPipelines(t *testing.T) {
	cfg, err := DefaultTextInConfig(" app ", " secret ")
	if err != nil {
		t.Fatalf("default textin config: %v", err)
	}

	if len(cfg.Pipelines) != 3 {
		t.Fatalf("expected 3 pipelines, got %d", len(cfg.Pipelines))
	}

	ocrPipe := cfg.Pipelines["ocr"]
	tablePipe := cfg.Pipelines["table"]
	stampPipe := cfg.Pipelines["stamp"]
	if ocrPipe.Provider.BaseURL != textInOCREndpoint {
		t.Fatalf("unexpected ocr endpoint: %s", ocrPipe.Provider.BaseURL)
	}
	if tablePipe.Provider.BaseURL != textInTableEndpoint {
		t.Fatalf("unexpected table endpoint: %s", tablePipe.Provider.BaseURL)
	}
	if stampPipe.Provider.BaseURL != textInStampEndpoint {
		t.Fatalf("unexpected stamp endpoint: %s", stampPipe.Provider.BaseURL)
	}

	ocrPipe.Provider.Auth["app_id"] = "changed"
	if tablePipe.Provider.Auth["app_id"] != " app " {
		t.Fatalf("expected auth maps to be independent, got %+v", tablePipe.Provider.Auth)
	}
	if stampPipe.Cache.Enabled || stampPipe.Cache.BaseDir != "" {
		t.Fatalf("expected environment-derived config to avoid implicit filesystem cache, got %+v", stampPipe.Cache)
	}
	if stampPipe.Worker.MaxConcurrent != 5 {
		t.Fatalf("unexpected default worker config: %+v", stampPipe.Worker)
	}
}

func TestDefaultTextInConfigRequiresCredentials(t *testing.T) {
	_, err := DefaultTextInConfig("", "secret")
	if err == nil || !strings.Contains(err.Error(), "requires appID and secret") {
		t.Fatalf("expected credentials error, got %v", err)
	}
}

func TestExampleConfigDoesNotContainUsableTextInCredentials(t *testing.T) {
	cfg, err := Load("example.yaml")
	if err != nil {
		t.Fatalf("load example config: %v", err)
	}
	for name, pipeline := range cfg.Pipelines {
		if strings.TrimSpace(pipeline.Provider.Auth["app_id"]) != "" ||
			strings.TrimSpace(pipeline.Provider.Auth["secret_code"]) != "" {
			t.Fatalf("example pipeline %s must not contain usable TextIn credentials", name)
		}
	}
}
