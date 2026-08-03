package web_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/wsnacj/agentx-go/tools/httprequest"
	"github.com/wsnacj/agentx-go/tools/web"
	"github.com/wsnacj/agentx-go/tools/web/retrieval"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (fn doerFunc) Do(request *http.Request) (*http.Response, error) { return fn(request) }

func fakePreparer(t *testing.T, do doerFunc) retrieval.Preparer {
	t.Helper()
	return func(_ context.Context, input httprequest.PrepareInput) (httprequest.PreparedRequest, error) {
		parsed, err := url.Parse(input.RawURL)
		if err != nil {
			return httprequest.PreparedRequest{}, err
		}
		return httprequest.PreparedRequest{URL: parsed, Doer: do}, nil
	}
}

func TestRunSearchUsesCanonicalProviderProtocol(t *testing.T) {
	credentialSensitive := false
	basePrepare := fakePreparer(t, func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("X-Subscription-Token") != "explicit-key" || request.URL.Query().Get("q") != "agent runtime" {
			t.Fatalf("request = %#v", request)
		}
		body := `{"web":{"results":[{"title":"AgentX","url":"https://example.test/agentx","description":"portable runtime"}]}}`
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})
	prepare := func(ctx context.Context, input httprequest.PrepareInput) (httprequest.PreparedRequest, error) {
		credentialSensitive = input.CredentialSensitive
		return basePrepare(ctx, input)
	}
	payload, err := web.RunSearch(context.Background(), web.SearchRequest{Query: "agent runtime", MaxResults: 3}, web.SearchOptions{
		DefaultProvider: "brave", Prepare: prepare,
		Providers: map[string]web.ProviderConfig{"brave": {APIKey: "explicit-key", Endpoint: "https://search.example.test/api"}},
	})
	if err != nil {
		t.Fatalf("RunSearch: %v", err)
	}
	if !credentialSensitive || payload.Provider != "brave" || len(payload.Results) != 1 || payload.Results[0].Title != "AgentX" {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestOpenPageAndFindInPageShareCanonicalCache(t *testing.T) {
	retrieval.ResetOpenPageCache()
	prepare := fakePreparer(t, func(request *http.Request) (*http.Response, error) {
		body := `<html><head><title>Guide</title></head><body><main><h1>AgentX Guide</h1><p>Portable retrieval keeps network policy in the Host.</p></main></body></html>`
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"text/html; charset=utf-8"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})
	page, cached, err := web.RunOpenPage(context.Background(), web.FetchRequest{URL: "https://docs.example.test/guide"}, web.FetchOptions{
		Prepare: prepare, CacheTTL: time.Minute, MaxChars: 4096, MaxResponseBytes: 64_000,
	})
	if err != nil || cached || page.PageID == "" || !strings.Contains(page.Text, "network policy") {
		t.Fatalf("page=%#v cached=%v err=%v", page, cached, err)
	}
	result := web.RunFindInPage(web.FindRequest{PageID: page.PageID, Query: "network policy"})
	if result.MatchCount != 1 || len(result.Matches) != 1 {
		t.Fatalf("find result = %#v", result)
	}
	_, cached, err = web.RunOpenPage(context.Background(), web.FetchRequest{URL: "https://docs.example.test/guide"}, web.FetchOptions{
		Prepare: prepare, CacheTTL: time.Minute, MaxChars: 4096, MaxResponseBytes: 64_000,
	})
	if err != nil || !cached {
		t.Fatalf("second open cached=%v err=%v", cached, err)
	}
}

func TestCancellationAndHostPolicyErrorsRemainObservable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	prepare := fakePreparer(t, func(request *http.Request) (*http.Response, error) { return nil, request.Context().Err() })
	_, err := web.RunFetch(ctx, web.FetchRequest{URL: "https://example.test/"}, web.FetchOptions{Prepare: prepare})
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error = %v", err)
	}

	blocked := errors.New("host policy blocked URL")
	_, err = web.RunFetch(context.Background(), web.FetchRequest{URL: "http://127.0.0.1/private"}, web.FetchOptions{Prepare: func(context.Context, httprequest.PrepareInput) (httprequest.PreparedRequest, error) {
		return httprequest.PreparedRequest{}, blocked
	}})
	if err == nil || !strings.Contains(err.Error(), "invalid_or_blocked_url") || !errors.Is(err, blocked) {
		t.Fatalf("policy error = %v", err)
	}
}

func TestDefinitionsExposeFourToolCapabilities(t *testing.T) {
	want := map[string]bool{"search": true, "web_search": true, "web_fetch": true, "open_page": true, "find_in_page": true}
	for _, definition := range []struct{ name string }{{web.SearchDefinition().Function.Name}, {web.WebSearchDefinition().Function.Name}, {web.WebFetchDefinition().Function.Name}, {web.OpenPageDefinition().Function.Name}, {web.FindInPageDefinition().Function.Name}} {
		delete(want, definition.name)
	}
	if len(want) != 0 {
		t.Fatalf("missing definitions: %#v", want)
	}
}
