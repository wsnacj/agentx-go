package gemini_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers"
	"github.com/wsnacj/agentx-go/providers/gemini"
	"github.com/wsnacj/agentx-go/providers/transport"
)

type fixtureDoer struct {
	request *http.Request
	body    map[string]any
}

type toolSequenceDoer struct {
	bodies    []map[string]any
	responses []string
}

func TestProviderChatAndVisionContract(t *testing.T) {
	doer := &fixtureDoer{}
	provider := gemini.NewProvider(gemini.Config{
		BaseURL:    "https://example.invalid/v1beta",
		HTTPClient: doer,
		ResolveMedia: func(_ context.Context, source string) (gemini.ResolvedMedia, error) {
			if source != "fixture.png" {
				t.Fatalf("source = %q", source)
			}
			return gemini.ResolvedMedia{MIMEType: "image/png", Base64Data: "Zml4dHVyZQ=="}, nil
		},
	})
	model := gemini.ModelConfig{Model: "gemini-fixture", MaxCompletion: 64, Capability: gemini.Capability{Vision: true, LocalFiles: true}}
	chat, _, err := provider.Chat(context.Background(), model, llm.ChatRequest{Messages: llm.Conversation{{Role: "user", Content: "hello"}}})
	if err != nil {
		t.Fatal(err)
	}
	if chat.Content != "ready" {
		t.Fatalf("chat = %#v", chat)
	}
	vision, _, err := provider.Vision(context.Background(), model, llm.VisualRequest{
		Messages: llm.Conversation{{Role: "user", Content: "inspect"}},
		Visual:   []llm.VisualContent{{Type: "image_url", ImageURL: "fixture.png"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if vision.Content != "ready" {
		t.Fatalf("vision = %#v", vision)
	}
	contents, ok := doer.body["contents"].([]any)
	if !ok || len(contents) != 1 {
		t.Fatalf("contents = %#v", doer.body["contents"])
	}
	content, _ := contents[0].(map[string]any)
	parts, _ := content["parts"].([]any)
	if len(parts) != 2 {
		t.Fatalf("parts = %#v", parts)
	}
	media, _ := parts[1].(map[string]any)
	inline, _ := media["inlineData"].(map[string]any)
	if inline["mimeType"] != "image/png" || inline["data"] != "Zml4dHVyZQ==" {
		t.Fatalf("inlineData = %#v", inline)
	}

	_, _, err = provider.Chat(context.Background(), model, llm.ChatRequest{Tools: []llm.Tool{{Type: "function"}}})
	if err != providers.ErrUnsupported {
		t.Fatalf("tool error = %v", err)
	}
}

func TestProviderChatToolContinuationContract(t *testing.T) {
	doer := &toolSequenceDoer{responses: []string{
		`{"candidates":[{"content":{"role":"model","parts":[{"functionCall":{"name":"lookup","args":{"id":"fixture"}}}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":2,"candidatesTokenCount":3,"totalTokenCount":5}}`,
		`{"candidates":[{"content":{"role":"model","parts":[{"text":"AGENTX_GEMINI_TOOL_OK"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":4,"candidatesTokenCount":2,"totalTokenCount":6}}`,
	}}
	provider := gemini.NewProvider(gemini.Config{BaseURL: "https://example.invalid/v1beta", HTTPClient: doer})
	model := gemini.ModelConfig{Model: "gemini-fixture", MaxCompletion: 64, Capability: gemini.Capability{ToolCalling: true}}
	tool := llm.Tool{Type: "function", Function: llm.Function{Name: "lookup", Parameters: map[string]any{"type": "object"}}}
	first, usage, err := provider.Chat(context.Background(), model, llm.ChatRequest{
		Messages: llm.Conversation{{Role: "user", Content: "lookup fixture"}}, Tools: []llm.Tool{tool},
		ToolChoice: &llm.ToolChoice{Type: "required"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Calls) != 1 || first.Calls[0].Name != "lookup" || usage == nil || usage.TotalTokens != 5 || first.Usage != usage {
		t.Fatalf("first = %#v usage=%#v", first, usage)
	}
	second, _, err := provider.Chat(context.Background(), model, llm.ChatRequest{
		Messages: llm.Conversation{
			{Role: "user", Content: "lookup fixture"},
			{Role: "assistant", ToolCalls: first.Calls},
			{Role: "tool", ToolName: "lookup", Content: `{"value":"ready"}`},
		},
		Tools: []llm.Tool{tool}, ToolChoice: &llm.ToolChoice{Type: "none"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if second.Content != "AGENTX_GEMINI_TOOL_OK" || len(second.Calls) != 0 || len(doer.bodies) != 2 {
		t.Fatalf("second = %#v bodies=%d", second, len(doer.bodies))
	}
	contents := doer.bodies[1]["contents"].([]any)
	if len(contents) != 3 {
		t.Fatalf("continuation contents = %#v", contents)
	}
}

func (d *toolSequenceDoer) Do(request *http.Request) (*http.Response, error) {
	var body map[string]any
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		return nil, err
	}
	d.bodies = append(d.bodies, body)
	response := d.responses[len(d.bodies)-1]
	return &http.Response{
		StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(response)), Request: request,
	}, nil
}

func (d *fixtureDoer) Do(request *http.Request) (*http.Response, error) {
	d.request = request
	if request.Body != nil {
		if err := json.NewDecoder(request.Body).Decode(&d.body); err != nil {
			return nil, err
		}
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"candidates":[{"content":{"parts":[{"text":"ready"}]}}]}`)),
		Request:    request,
	}, nil
}

func TestClientGenerateContentContract(t *testing.T) {
	doer := &fixtureDoer{}
	client := gemini.New(gemini.Config{
		BaseURL: "https://example.invalid/v1beta",
		Transport: transport.Config{Headers: map[string]string{
			"x-agentx-client": "default",
		}},
		Authorize: func(_ context.Context, headers http.Header) error {
			headers.Set("x-goog-api-key", "fixture-token")
			return nil
		},
		HTTPClient: doer,
	})
	ctx := transport.WithRequestOptions(context.Background(), llm.RequestOptions{
		Headers: map[string]string{"x-agentx-client": "request"},
	})
	response, err := client.GenerateContent(ctx, "gemini-fixture", &gemini.GenerateContentRequest{
		Contents: []gemini.Content{gemini.NewContent(gemini.NewTextPart("hello"))},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := gemini.ResponseText(response); got != "ready" {
		t.Fatalf("ResponseText() = %q", got)
	}
	if doer.request.URL.Path != "/v1beta/models/gemini-fixture:generateContent" {
		t.Fatalf("path = %q", doer.request.URL.Path)
	}
	if doer.request.Header.Get("x-goog-api-key") != "fixture-token" || doer.request.Header.Get("x-agentx-client") != "request" {
		t.Fatalf("headers = %#v", doer.request.Header)
	}
	if _, ok := doer.body["contents"]; !ok {
		t.Fatalf("payload = %#v", doer.body)
	}
}
