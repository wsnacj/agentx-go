package ark_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/wsnacj/agentx-go/providers/ark"
	arktypes "github.com/wsnacj/agentx-go/providers/ark/types"
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
		Body:       io.NopCloser(strings.NewReader(`{"id":"resp_fixture","status":"completed"}`)),
		Request:    request,
	}, nil
}

func TestClientJSONContract(t *testing.T) {
	doer := &fixtureDoer{}
	client := ark.New(ark.Config{
		BaseURL: "https://example.invalid/api/v3",
		Transport: transport.Config{Headers: map[string]string{
			"x-agentx-client": "fixture",
		}},
		Authorize: func(_ context.Context, headers http.Header) error {
			headers.Set("Authorization", "Bearer fixture-token")
			return nil
		},
		HTTPClient: doer,
	})
	request := arktypes.ResponseRequest{
		Model:        "doubao-fixture",
		Input:        arktypes.NewInputText("hello"),
		ExtraHeaders: map[string]string{"x-request-id": "request-1"},
	}
	var response arktypes.Response
	if err := client.DoJSON(context.Background(), http.MethodPost, "/responses", request, &response); err != nil {
		t.Fatal(err)
	}
	if response.ID != "resp_fixture" || response.Status != "completed" {
		t.Fatalf("response = %#v", response)
	}
	if doer.request.URL.Path != "/api/v3/responses" {
		t.Fatalf("path = %q", doer.request.URL.Path)
	}
	if doer.request.Header.Get("Authorization") != "Bearer fixture-token" || doer.request.Header.Get("x-agentx-client") != "fixture" || doer.request.Header.Get("x-request-id") != "request-1" {
		t.Fatalf("headers = %#v", doer.request.Header)
	}
	if doer.body["model"] != "doubao-fixture" || doer.body["input"] != "hello" {
		t.Fatalf("payload = %#v", doer.body)
	}
}
