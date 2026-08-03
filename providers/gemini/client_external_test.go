package gemini_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers/gemini"
	"github.com/wsnacj/agentx-go/providers/transport"
)

type fixtureDoer struct {
	request *http.Request
	body    map[string]any
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
