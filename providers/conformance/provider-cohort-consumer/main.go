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
	"github.com/wsnacj/agentx-go/providers/ark"
	arktypes "github.com/wsnacj/agentx-go/providers/ark/types"
	"github.com/wsnacj/agentx-go/providers/codex"
	"github.com/wsnacj/agentx-go/providers/gemini"
)

type fixtureDoer struct {
	body          string
	contentType   string
	authorization string
	accountID     string
	headers       http.Header
}

func (d *fixtureDoer) Do(request *http.Request) (*http.Response, error) {
	d.authorization = request.Header.Get("Authorization")
	d.accountID = request.Header.Get("ChatGPT-Account-ID")
	d.headers = request.Header.Clone()
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
	ArkResponseID    string `json:"ark_response_id"`
	ArkAuthorization string `json:"ark_authorization"`
	GeminiContent    string `json:"gemini_content"`
	GeminiAPIKey     string `json:"gemini_api_key"`
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

	arkDoer := &fixtureDoer{contentType: "application/json", body: `{"id":"ark-ready","status":"completed"}`}
	arkClient := ark.New(ark.Config{
		BaseURL: "https://example.invalid/ark", HTTPClient: arkDoer,
		Authorize: func(_ context.Context, headers http.Header) error {
			headers.Set("Authorization", "Bearer ark-fixture")
			return nil
		},
	})
	arkResponse, err := arkClient.CreateResponse(ctx, arktypes.ResponseRequest{Model: "doubao-fixture", Input: arktypes.NewInputText("ping")})
	if err != nil {
		return result{}, err
	}

	geminiDoer := &fixtureDoer{contentType: "application/json", body: `{"candidates":[{"content":{"parts":[{"text":"gemini-ready"}]}}],"usageMetadata":{"promptTokenCount":2,"candidatesTokenCount":1,"totalTokenCount":3}}`}
	geminiProvider := gemini.NewProvider(gemini.Config{
		BaseURL: "https://example.invalid/gemini", HTTPClient: geminiDoer,
		Authorize: func(_ context.Context, headers http.Header) error {
			headers.Set("x-goog-api-key", "gemini-fixture")
			return nil
		},
	})
	geminiResponse, geminiUsage, err := geminiProvider.Chat(ctx, gemini.ModelConfig{Model: "gemini-fixture"}, llm.ChatRequest{Messages: llm.Conversation{{Role: "user", Content: "ping"}}})
	if err != nil {
		return result{}, err
	}
	if arkResponse == nil || geminiResponse == nil || geminiUsage == nil || geminiUsage.TotalTokens != 3 {
		return result{}, fmt.Errorf("incomplete Ark/Gemini provider response")
	}
	return result{
		AnthropicContent: anthropicResponse.Content, AnthropicTokens: anthropicUsage.TotalTokens,
		CodexContent: codexResponse.Content, CodexTool: codexResponse.Calls[0].Name,
		CodexTokens: codexUsage.TotalTokens, Authorization: codexDoer.authorization, AccountID: codexDoer.accountID,
		ArkResponseID: arkResponse.ID, ArkAuthorization: arkDoer.authorization,
		GeminiContent: geminiResponse.Content, GeminiAPIKey: geminiDoer.headers.Get("x-goog-api-key"),
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
