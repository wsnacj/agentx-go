package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers/anthropic"
	"github.com/wsnacj/agentx-go/providers/codex"
)

type fixtureDoer struct {
	body          string
	contentType   string
	authorization string
	accountID     string
}

func (d *fixtureDoer) Do(request *http.Request) (*http.Response, error) {
	d.authorization = request.Header.Get("Authorization")
	d.accountID = request.Header.Get("ChatGPT-Account-ID")
	return &http.Response{
		StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{d.contentType}},
		Body: io.NopCloser(strings.NewReader(d.body)), Request: request,
	}, nil
}

type result struct {
	AnthropicContent string `json:"anthropic_content"`
	AnthropicTokens  int    `json:"anthropic_tokens"`
	CodexContent     string `json:"codex_content"`
	CodexTool        string `json:"codex_tool"`
	CodexTokens      int    `json:"codex_tokens"`
	Authorization    string `json:"authorization"`
	AccountID        string `json:"account_id"`
}

func run(ctx context.Context) (result, error) {
	anthropicDoer := &fixtureDoer{contentType: "application/json", body: `{"content":[{"type":"text","text":"anthropic-ready"}],"usage":{"input_tokens":2,"output_tokens":3}}`}
	anthropicClient, err := anthropic.New(anthropic.Config{
		BaseURL: "https://example.invalid/anthropic", HTTPClient: anthropicDoer,
		Authorize: func(_ context.Context, headers http.Header) error {
			headers.Set("x-api-key", "anthropic-fixture")
			return nil
		},
	})
	if err != nil {
		return result{}, err
	}
	anthropicResponse, anthropicUsage, err := anthropicClient.Chat(ctx, anthropic.ModelConfig{Model: "claude-fixture"}, llm.ChatRequest{Messages: llm.Conversation{{Role: "user", Content: "ping"}}})
	if err != nil {
		return result{}, err
	}

	codexDoer := &fixtureDoer{contentType: "text/event-stream", body: "event: response.completed\n" +
		`data: {"type":"response.completed","response":{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"codex-ready"}]},{"type":"function_call","status":"completed","call_id":"call_1","name":"lookup","arguments":"{}"}],"usage":{"input_tokens":4,"output_tokens":2,"total_tokens":6}}}` + "\n\n"}
	codexClient, err := codex.New(codex.Config{
		BaseURL: "https://example.invalid/codex", HTTPClient: codexDoer,
		Authorize: func(_ context.Context, headers http.Header) error {
			headers.Set("Authorization", "Bearer codex-fixture")
			headers.Set("ChatGPT-Account-ID", "acct-fixture")
			return nil
		},
	})
	if err != nil {
		return result{}, err
	}
	codexResponse, codexUsage, err := codexClient.Chat(ctx, codex.ModelConfig{Model: "gpt-fixture"}, llm.ChatRequest{Messages: llm.Conversation{{Role: "user", Content: "lookup"}}})
	if err != nil {
		return result{}, err
	}
	if anthropicResponse == nil || anthropicUsage == nil || codexResponse == nil || codexUsage == nil || len(codexResponse.Calls) != 1 {
		return result{}, fmt.Errorf("incomplete provider cohort response")
	}
	return result{
		AnthropicContent: anthropicResponse.Content, AnthropicTokens: anthropicUsage.TotalTokens,
		CodexContent: codexResponse.Content, CodexTool: codexResponse.Calls[0].Name,
		CodexTokens: codexUsage.TotalTokens, Authorization: codexDoer.authorization, AccountID: codexDoer.accountID,
	}, nil
}

func main() {
	value, err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
