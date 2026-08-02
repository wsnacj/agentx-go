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
	"github.com/wsnacj/agentx-go/providers/openaicompat"
)

type fixtureDoer struct {
	authorization string
}

func (d *fixtureDoer) Do(request *http.Request) (*http.Response, error) {
	d.authorization = request.Header.Get("Authorization")
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body: io.NopCloser(strings.NewReader(
			`{"choices":[{"message":{"role":"assistant","content":"ready","tool_calls":[{"id":"call_1","type":"function","function":{"name":"lookup","arguments":"{\"id\":\"fixture\"}"}}]}}],"usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}}`,
		)),
		Request: request,
	}, nil
}

type result struct {
	Content       string `json:"content"`
	Tool          string `json:"tool"`
	TotalTokens   int    `json:"total_tokens"`
	Authorization string `json:"authorization"`
}

func run(ctx context.Context) (result, error) {
	doer := &fixtureDoer{}
	client, err := openaicompat.New(openaicompat.Config{
		Name: "fixture", BaseURL: "https://example.invalid/v1", HTTPClient: doer,
		Authorize: func(_ context.Context, headers http.Header) error {
			headers.Set("Authorization", "Bearer fixture-token")
			return nil
		},
	})
	if err != nil {
		return result{}, err
	}
	response, usage, err := client.Chat(ctx, openaicompat.ModelConfig{Model: "fixture-model"}, llm.ChatRequest{
		Messages: llm.Conversation{{Role: "user", Content: "lookup fixture"}},
		Tools:    []llm.Tool{{Type: "function", Function: llm.Function{Name: "lookup", Parameters: map[string]any{"type": "object"}}}},
	})
	if err != nil {
		return result{}, err
	}
	if response == nil || usage == nil || len(response.Calls) != 1 {
		return result{}, fmt.Errorf("incomplete provider response")
	}
	return result{Content: response.Content, Tool: response.Calls[0].Name, TotalTokens: usage.TotalTokens, Authorization: doer.authorization}, nil
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
