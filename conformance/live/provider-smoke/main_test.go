package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoadConfigRequiresExplicitOptIn(t *testing.T) {
	_, err := loadConfig(func(string) string { return "" })
	if !errors.Is(err, errLiveDisabled) {
		t.Fatalf("loadConfig() error = %v, want errLiveDisabled", err)
	}
}

func TestLoadConfigFailsClosedWhenRequiredValuesAreMissing(t *testing.T) {
	values := map[string]string{envEnable: "1"}
	_, err := loadConfig(func(key string) string { return values[key] })
	if err == nil || !strings.Contains(err.Error(), envBaseURL) || !strings.Contains(err.Error(), envAPIKey) || !strings.Contains(err.Error(), envModel) {
		t.Fatalf("loadConfig() error = %v, want all required environment names", err)
	}
}

func TestLoadConfigUsesExplicitValuesWithoutReadingProviderDefaults(t *testing.T) {
	values := map[string]string{
		envEnable:  "true",
		envBaseURL: "https://provider.example/v1",
		envAPIKey:  "secret",
		envModel:   "model-x",
		envPrompt:  "hello",
		envExpect:  "world",
		envMode:    modeToolLoop,
		envTimeout: "12s",
	}
	cfg, err := loadConfig(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.BaseURL != values[envBaseURL] || cfg.APIKey != values[envAPIKey] || cfg.Model != values[envModel] || cfg.Prompt != "hello" || cfg.Expect != "world" || cfg.Mode != modeToolLoop || cfg.Timeout != 12*time.Second {
		t.Fatalf("loadConfig() = %#v", cfg)
	}
}

func TestTruthy(t *testing.T) {
	for _, value := range []string{"1", "true", "TRUE", "yes", "on"} {
		if !truthy(value) {
			t.Fatalf("truthy(%q) = false", value)
		}
	}
	for _, value := range []string{"", "0", "false", "disabled"} {
		if truthy(value) {
			t.Fatalf("truthy(%q) = true", value)
		}
	}
}

func TestRunComposesCanonicalRuntimeAndProvider(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/chat/completions" {
			t.Fatalf("request path = %q", request.URL.Path)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer test-secret" {
			t.Fatalf("Authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"AGENTX_LIVE_PROVIDER_SMOKE_OK"}}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	value, err := run(context.Background(), config{
		BaseURL: server.URL,
		APIKey:  "test-secret",
		Model:   "fixture-model",
		Prompt:  "smoke",
		Expect:  defaultMarker,
		Mode:    modeChat,
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if value.Status != "completed" || !value.Matched || value.Reply != defaultMarker || value.SessionID != "live-provider-smoke-chat" || value.Mode != modeChat {
		t.Fatalf("run() = %#v", value)
	}
}

func TestRunComposesCanonicalLiveToolLoop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"","tool_calls":[{"id":"call_1","type":"function","function":{"name":"agentx_smoke_marker","arguments":"{\"value\":\"AGENTX_LIVE_PROVIDER_SMOKE_OK\"}"}}]}}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer server.Close()

	value, err := run(context.Background(), config{
		BaseURL: server.URL, APIKey: "test-secret", Model: "fixture-model",
		Prompt: "call the marker", Expect: defaultMarker, Mode: modeToolLoop, Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if value.Status != "completed" || !value.Matched || value.Reply != defaultMarker || value.ToolCalls != 1 || value.Mode != modeToolLoop {
		t.Fatalf("run() = %#v", value)
	}
}
