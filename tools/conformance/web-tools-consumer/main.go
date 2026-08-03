package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/tools"
	"github.com/wsnacj/agentx-go/tools/httprequest"
	"github.com/wsnacj/agentx-go/tools/web"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (fn doerFunc) Do(request *http.Request) (*http.Response, error) { return fn(request) }

type result struct {
	Registered int  `json:"registered"`
	SearchOK   bool `json:"search_ok"`
	OpenOK     bool `json:"open_ok"`
	FindOK     bool `json:"find_ok"`
	Verified   bool `json:"verified"`
}

func run(ctx context.Context) (result, error) {
	prepare := func(_ context.Context, input httprequest.PrepareInput) (httprequest.PreparedRequest, error) {
		parsed, err := url.Parse(input.RawURL)
		if err != nil {
			return httprequest.PreparedRequest{}, err
		}
		return httprequest.PreparedRequest{URL: parsed, Doer: doerFunc(func(request *http.Request) (*http.Response, error) {
			body := `<html><head><title>Portable Web</title></head><body><main><p>AgentX keeps outbound policy in the Host.</p></main></body></html>`
			contentType := "text/html"
			if request.URL.Host == "search.fixture.test" {
				body = `{"web":{"results":[{"title":"Portable Web","url":"https://docs.fixture.test/web","description":"canonical search"}]}}`
				contentType = "application/json"
			}
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
		})}, nil
	}
	registry := tools.NewRegistry()
	web.Register(registry, web.Options{
		Search: web.SearchOptions{DefaultProvider: "brave", Prepare: prepare, Providers: map[string]web.ProviderConfig{"brave": {APIKey: "fixture", Endpoint: "https://search.fixture.test/api"}}},
		Fetch:  web.FetchOptions{Prepare: prepare, CacheTTL: time.Minute},
	})
	searchRaw, err := registry.Execute(ctx, toolcontract.Call{Name: web.SearchName, Arguments: `{"query":"agentx"}`})
	if err != nil {
		return result{}, err
	}
	var searchPayload struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal([]byte(searchRaw), &searchPayload); err != nil {
		return result{}, err
	}
	openRaw, err := registry.Execute(ctx, toolcontract.Call{Name: web.OpenPageName, Arguments: `{"url":"https://docs.fixture.test/web"}`})
	if err != nil {
		return result{}, err
	}
	var page struct {
		PageID string `json:"page_id"`
		Text   string `json:"text"`
	}
	if err := json.Unmarshal([]byte(openRaw), &page); err != nil {
		return result{}, err
	}
	findRaw, err := registry.Execute(ctx, toolcontract.Call{Name: web.FindInPageName, Arguments: fmt.Sprintf(`{"page_id":%q,"query":"outbound policy"}`, page.PageID)})
	if err != nil {
		return result{}, err
	}
	var findPayload struct {
		MatchCount int `json:"match_count"`
	}
	if err := json.Unmarshal([]byte(findRaw), &findPayload); err != nil {
		return result{}, err
	}
	value := result{Registered: len(registry.Definitions()), SearchOK: searchPayload.Count == 1, OpenOK: page.PageID != "" && strings.Contains(page.Text, "outbound policy"), FindOK: findPayload.MatchCount == 1}
	value.Verified = value.Registered == 5 && value.SearchOK && value.OpenOK && value.FindOK
	return value, nil
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
