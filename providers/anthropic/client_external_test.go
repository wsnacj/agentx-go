package anthropic_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers/anthropic"
)

type fixtureDoer struct {
	request *http.Request
	body    map[string]any
}

func (d *fixtureDoer) Do(request *http.Request) (*http.Response, error) {
	d.request = request
	if err := json.NewDecoder(request.Body).Decode(&d.body); err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(
			`{"content":[{"type":"text","text":"ready"},{"type":"tool_use","id":"call_1","name":"lookup","input":{"id":"fixture"}}],"usage":{"input_tokens":2,"output_tokens":3}}`,
		)),
		Request: request,
	}, nil
}

func TestClientChatContract(t *testing.T) {
	doer := &fixtureDoer{}
	client, err := anthropic.New(anthropic.Config{
		BaseURL: "https://example.invalid", HTTPClient: doer,
		Authorize: func(_ context.Context, headers http.Header) error {
			headers.Set("x-api-key", "fixture-token")
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	response, usage, err := client.Chat(context.Background(), anthropic.ModelConfig{Model: "fixture", MaxCompletion: 64}, llm.ChatRequest{
		Messages: llm.Conversation{{Role: "user", Content: "lookup"}},
		Tools:    []llm.Tool{{Type: "function", Function: llm.Function{Name: "lookup", Parameters: map[string]any{"type": "object"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Content != "ready" || len(response.Calls) != 1 || response.Calls[0].Name != "lookup" {
		t.Fatalf("response = %#v", response)
	}
	if usage.TotalTokens != 5 || response.Usage != usage {
		t.Fatalf("usage = %#v response usage = %#v", usage, response.Usage)
	}
	if doer.request.Header.Get("x-api-key") != "fixture-token" || doer.request.Header.Get("anthropic-version") != "2023-06-01" {
		t.Fatalf("headers = %#v", doer.request.Header)
	}
	if doer.body["model"] != "fixture" || doer.body["max_tokens"] != float64(64) {
		t.Fatalf("payload = %#v", doer.body)
	}
}
