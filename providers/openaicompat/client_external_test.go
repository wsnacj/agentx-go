package openaicompat_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers"
	"github.com/wsnacj/agentx-go/providers/fault"
	"github.com/wsnacj/agentx-go/providers/openaicompat"
	providertransport "github.com/wsnacj/agentx-go/providers/transport"
)

func TestClientChatUsesExplicitHostSeams(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer fixture" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("X-Request"); got != "request" {
			t.Errorf("X-Request = %q", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("X-Observed", "yes")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok","tool_calls":[{"id":"call_1","type":"function","function":{"name":"lookup","arguments":"{}"}}]}}],"usage":{"prompt_tokens":2,"completion_tokens":3,"total_tokens":5}}`))
	}))
	defer server.Close()

	var observed llm.ResponseMetadata
	client, err := openaicompat.New(openaicompat.Config{
		Name: "fixture", BaseURL: server.URL,
		Transport: providertransport.Config{Headers: map[string]string{"X-Request": "default"}},
		Authorize: func(_ context.Context, headers http.Header) error {
			if headers.Get("Authorization") == "" {
				headers.Set("Authorization", "Bearer fixture")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	response, usage, err := client.Chat(context.Background(), openaicompat.ModelConfig{
		Name: "chat", Model: "model-default", MaxCompletion: 128, Temperature: .3,
	}, llm.ChatRequest{
		Model: "model-request", Messages: llm.Conversation{{Role: "user", Content: "hello"}},
		MaxTokens: 77,
		Options: llm.RequestOptions{
			SessionID: "session-1", CacheControl: "no-store",
			Headers:    map[string]string{"X-Request": "request"},
			OnResponse: func(_ context.Context, meta llm.ResponseMetadata) error { observed = meta; return nil },
		}.ToMap(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Content != "ok" || len(response.Calls) != 1 || response.Calls[0].Name != "lookup" {
		t.Fatalf("response = %#v", response)
	}
	if usage == nil || usage.TotalTokens != 5 {
		t.Fatalf("usage = %#v", usage)
	}
	if body["model"] != "model-request" || body["max_completion_tokens"] != float64(77) || body["session_id"] != "session-1" {
		t.Fatalf("payload = %#v", body)
	}
	if observed.StatusCode != http.StatusOK || observed.Headers["X-Observed"][0] != "yes" {
		t.Fatalf("observed = %#v", observed)
	}
}

func TestClientStreamAndErrorContracts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("slow down"))
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"id\":\"resp_1\",\"choices\":[{\"delta\":{\"content\":\"hello\"},\"finish_reason\":\"stop\"}],\"usage\":{\"total_tokens\":4}}\n\ndata: [DONE]\n\n"))
	}))
	defer server.Close()
	client, err := openaicompat.New(openaicompat.Config{BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	stream, err := client.StreamChatEvents(context.Background(), openaicompat.ModelConfig{
		Model: "fixture", Capability: openaicompat.Capability{Streaming: true},
	}, llm.ChatRequest{Messages: llm.Conversation{{Role: "user", Content: "hello"}}})
	if err != nil {
		t.Fatal(err)
	}
	var sawText, sawUsage, sawDone bool
	for event := range stream.Ch {
		sawText = sawText || event.Type == llm.StreamEventTextDelta && event.TextDelta == "hello"
		sawUsage = sawUsage || event.Type == llm.StreamEventUsage && event.Usage != nil && event.Usage.TotalTokens == 4
		sawDone = sawDone || event.Type == llm.StreamEventDone && event.StopReason == llm.StreamStopReasonStop
	}
	if !sawText || !sawUsage || !sawDone {
		t.Fatalf("events text=%t usage=%t done=%t", sawText, sawUsage, sawDone)
	}

	_, _, err = client.Embedding(context.Background(), openaicompat.EmbeddingConfig{Model: "fixture", Path: "/error"}, llm.EmbeddingRequest{Inputs: []string{"x"}})
	var apiErr *providers.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("error = %#v", err)
	}
	classification := fault.Classify(err)
	if classification.Kind != fault.KindRateLimit || !classification.Retryable {
		t.Fatalf("classification = %#v", classification)
	}
}

func TestClientHonorsCallerDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()
	client, err := openaicompat.New(openaicompat.Config{BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, _, err = client.Chat(ctx, openaicompat.ModelConfig{Model: "fixture"}, llm.ChatRequest{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v", err)
	}
}
