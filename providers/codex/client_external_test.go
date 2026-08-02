package codex_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers"
	"github.com/wsnacj/agentx-go/providers/codex"
)

type fixtureDoer struct {
	status  int
	body    string
	request *http.Request
	payload map[string]any
}

func (d *fixtureDoer) Do(request *http.Request) (*http.Response, error) {
	d.request = request
	if err := json.NewDecoder(request.Body).Decode(&d.payload); err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: d.status, Header: http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(d.body)), Request: request,
	}, nil
}

func TestClientChatCollectsResponsesStream(t *testing.T) {
	doer := &fixtureDoer{status: http.StatusOK, body: "event: response.completed\n" +
		`data: {"type":"response.completed","response":{"status":"completed","output":[{"type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"ready"}]},{"type":"function_call","status":"completed","call_id":"call_1","name":"lookup","arguments":"{\"id\":\"fixture\"}"}],"usage":{"input_tokens":2,"output_tokens":3,"total_tokens":5}}}` + "\n\n"}
	client, err := codex.New(codex.Config{
		BaseURL: "https://example.invalid/codex", HTTPClient: doer,
		Authorize: func(_ context.Context, headers http.Header) error {
			headers.Set("Authorization", "Bearer fixture-token")
			headers.Set("ChatGPT-Account-ID", "acct_fixture")
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	response, usage, err := client.Chat(context.Background(), codex.ModelConfig{Model: "gpt-fixture", ReasoningDefault: "minimal"}, llm.ChatRequest{
		System: "be useful", Messages: llm.Conversation{{Role: "user", Content: "lookup"}},
		Tools: []llm.Tool{{Type: "function", Function: llm.Function{Name: "lookup", Parameters: map[string]any{"type": "object"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.Content != "ready" || len(response.Calls) != 1 || response.Calls[0].Name != "lookup" || usage.TotalTokens != 5 {
		t.Fatalf("response = %#v usage = %#v", response, usage)
	}
	if doer.request.Header.Get("Authorization") != "Bearer fixture-token" || doer.request.Header.Get("ChatGPT-Account-ID") != "acct_fixture" {
		t.Fatalf("headers = %#v", doer.request.Header)
	}
	if doer.request.Header.Get("originator") != "codex_cli_rs" || doer.payload["stream"] != true || doer.payload["model"] != "gpt-fixture" {
		t.Fatalf("headers = %#v payload = %#v", doer.request.Header, doer.payload)
	}
}

func TestClientPreservesTypedAPIError(t *testing.T) {
	doer := &fixtureDoer{status: http.StatusTooManyRequests, body: `{"error":"rate limited"}`}
	client, err := codex.New(codex.Config{BaseURL: "https://example.invalid/codex", HTTPClient: doer})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = client.Chat(context.Background(), codex.ModelConfig{Model: "fixture"}, llm.ChatRequest{})
	var apiErr *providers.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("error = %#v", err)
	}
}
