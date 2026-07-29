package llm

import (
	"context"
	"testing"
)

func TestRequestOptionsRoundTrip(t *testing.T) {
	in := map[string]any{
		"temperature":         0.2,
		"top_p":               0.8,
		"top_k":               6,
		"response_format":     map[string]any{"type": "json_object"},
		"detail":              "high",
		"thinking":            true,
		"reasoning_effort":    "medium",
		"transport":           "sse",
		"session_id":          "sess-1",
		"parallel_tool_calls": true,
		"max_retry_delay_ms":  1200,
		"headers":             map[string]any{"x-test": "1"},
		"metadata":            map[string]any{"scope": "typed"},
		"extra_body":          map[string]any{"cache_control": "no-store", "custom": "body"},
		"custom_field":        "kept",
	}

	opts := RequestOptionsFromMap(in)
	if opts.Temperature == nil || *opts.Temperature != 0.2 {
		t.Fatalf("expected temperature override, got %+v", opts.Temperature)
	}
	if opts.TopP == nil || *opts.TopP != 0.8 {
		t.Fatalf("expected top_p override, got %+v", opts.TopP)
	}
	if opts.TopK == nil || *opts.TopK != 6 {
		t.Fatalf("expected top_k override, got %+v", opts.TopK)
	}
	if opts.Thinking == nil || !opts.Thinking.Enabled {
		t.Fatalf("expected thinking enabled, got %+v", opts.Thinking)
	}
	if opts.Reasoning == nil || opts.Reasoning.Effort != "medium" {
		t.Fatalf("expected reasoning effort, got %+v", opts.Reasoning)
	}
	if opts.SessionID != "sess-1" {
		t.Fatalf("expected session id, got %q", opts.SessionID)
	}
	if opts.ParallelToolCalls == nil || !*opts.ParallelToolCalls {
		t.Fatalf("expected parallel_tool_calls typed option, got %+v", opts.ParallelToolCalls)
	}
	if _, ok := opts.ProviderFields["parallel_tool_calls"]; ok {
		t.Fatalf("parallel_tool_calls should be consumed as typed option, got provider fields %+v", opts.ProviderFields)
	}
	if opts.MaxRetryDelayMs == nil || *opts.MaxRetryDelayMs != 1200 {
		t.Fatalf("expected max retry delay override, got %+v", opts.MaxRetryDelayMs)
	}
	if opts.CacheControl != "no-store" {
		t.Fatalf("expected cache control no-store, got %#v", opts.CacheControl)
	}
	if opts.ExtraBody["custom"] != "body" {
		t.Fatalf("expected extra_body custom preserved, got %#v", opts.ExtraBody)
	}
	if opts.ProviderFields["custom_field"] != "kept" {
		t.Fatalf("expected provider field preserved, got %+v", opts.ProviderFields)
	}

	out := opts.ToMap()
	if out["custom_field"] != "kept" {
		t.Fatalf("expected custom field in roundtrip map, got %+v", out)
	}
	if _, ok := out["temperature"]; !ok {
		t.Fatalf("expected temperature in roundtrip map")
	}
	if got := out["session_id"]; got != "sess-1" {
		t.Fatalf("expected session_id in roundtrip map, got %#v", got)
	}
	if got := out["parallel_tool_calls"]; got != true {
		t.Fatalf("expected parallel_tool_calls in roundtrip map, got %#v", got)
	}
	if got := out["max_retry_delay_ms"]; got != 1200 {
		t.Fatalf("expected max_retry_delay_ms in roundtrip map, got %#v", got)
	}
	extraBody := out["extra_body"].(map[string]any)
	if extraBody["cache_control"] != "no-store" || extraBody["custom"] != "body" {
		t.Fatalf("expected cache_control and custom extra body preserved, got %#v", extraBody)
	}
}

func TestRequestOptionsClonePreservesHooks(t *testing.T) {
	payloadSeen := false
	responseSeen := false
	retryDelay := 700
	opts := RequestOptions{
		MaxRetryDelayMs: &retryDelay,
		OnPayload: func(_ context.Context, payload any) (any, error) {
			payloadSeen = true
			return payload, nil
		},
		OnResponse: func(_ context.Context, meta ResponseMetadata) error {
			if meta.StatusCode == 204 {
				responseSeen = true
			}
			return nil
		},
	}

	cloned := opts.Clone()
	if cloned.MaxRetryDelayMs == nil || *cloned.MaxRetryDelayMs != retryDelay {
		t.Fatalf("expected cloned retry delay, got %+v", cloned.MaxRetryDelayMs)
	}
	if _, err := cloned.OnPayload(context.Background(), map[string]any{"ok": true}); err != nil {
		t.Fatalf("payload hook returned error: %v", err)
	}
	if err := cloned.OnResponse(context.Background(), ResponseMetadata{StatusCode: 204}); err != nil {
		t.Fatalf("response hook returned error: %v", err)
	}
	if !payloadSeen || !responseSeen {
		t.Fatalf("expected cloned hooks to remain callable")
	}
}

func TestEmbeddingOptionsRoundTrip(t *testing.T) {
	in := map[string]any{
		"dimensions":       1024,
		"encoding_format":  "base64",
		"instructions":     "compress",
		"sparse_embedding": true,
		"path":             "/embeddings/custom",
		"batch_size":       8,
		"video_urls":       []string{"https://example.com/a.mp4"},
	}

	opts := EmbeddingOptionsFromMap(in)
	if opts.Dimensions == nil || *opts.Dimensions != 1024 {
		t.Fatalf("expected dimensions, got %+v", opts.Dimensions)
	}
	if opts.Encoding != "base64" {
		t.Fatalf("expected encoding override, got %q", opts.Encoding)
	}
	if opts.SparseEmbedding == nil || !*opts.SparseEmbedding {
		t.Fatalf("expected sparse_embedding, got %+v", opts.SparseEmbedding)
	}
	if opts.Path != "/embeddings/custom" {
		t.Fatalf("expected path override, got %q", opts.Path)
	}
	if opts.BatchSize == nil || *opts.BatchSize != 8 {
		t.Fatalf("expected batch size, got %+v", opts.BatchSize)
	}
	if _, ok := opts.ProviderFields["video_urls"]; !ok {
		t.Fatalf("expected unknown embedding field preserved")
	}

	out := opts.ToMap()
	if out["video_urls"] == nil {
		t.Fatalf("expected provider field preserved in roundtrip map")
	}
}
