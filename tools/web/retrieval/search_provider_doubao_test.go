package retrieval

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/wsnacj/agentx-go/tools/httprequest"
)

type searchDoerFunc func(*http.Request) (*http.Response, error)

func (fn searchDoerFunc) Do(request *http.Request) (*http.Response, error) { return fn(request) }

func searchPreparer(t *testing.T, do searchDoerFunc) Preparer {
	t.Helper()
	return func(_ context.Context, input httprequest.PrepareInput) (httprequest.PreparedRequest, error) {
		parsed, err := url.Parse(input.RawURL)
		if err != nil {
			return httprequest.PreparedRequest{}, err
		}
		return httprequest.PreparedRequest{URL: parsed, Doer: do}, nil
	}
}

func TestPrepareDoubaoCustomSearchRequest(t *testing.T) {
	t.Parallel()
	plan, validation := PrepareSearchRequest(SearchPrepareOptions{
		ConfiguredProvider: "ark",
		Query:              "AgentX",
		Count:              5,
		DateAfter:          "2026-08-01",
		DateBefore:         "2026-08-18",
		DomainFilter:       []string{"docs.example.com", "-spam.example.com"},
		AuthoritativeOnly:  true,
		QueryRewrite:       true,
	})
	if validation != nil {
		t.Fatalf("validation = %#v", validation)
	}
	if plan.EffectiveProvider != SearchProviderDoubaoCustom || plan.DoubaoCustomTimeRange != "2026-08-01..2026-08-18" {
		t.Fatalf("plan = %#v", plan)
	}
	if len(plan.DoubaoSites) != 1 || plan.DoubaoSites[0] != "docs.example.com" || len(plan.DoubaoBlockedHosts) != 1 || plan.DoubaoBlockedHosts[0] != "spam.example.com" {
		t.Fatalf("domain mapping = %#v / %#v", plan.DoubaoSites, plan.DoubaoBlockedHosts)
	}
}

func TestRunDoubaoCustomSearchProtocol(t *testing.T) {
	t.Parallel()
	prepare := searchPreparer(t, func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Authorization") != "Bearer search-key" {
			t.Fatalf("authorization header was not populated")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		filter, _ := body["Filter"].(map[string]any)
		control, _ := body["QueryControl"].(map[string]any)
		if body["SearchType"] != "web" || body["TimeRange"] != "OneWeek" || filter["NeedUrl"] != true || filter["Sites"] != "docs.example.com" || filter["BlockHosts"] != "spam.example.com" || filter["AuthInfoLevel"] != float64(1) || control["QueryRewrite"] != true {
			t.Fatalf("request body = %#v", body)
		}
		response := `{"ResponseMetadata":{"RequestId":"request:custom"},"Result":{"ResultCount":1,"WebResults":[{"Title":"AgentX","Url":"https://docs.example.com/agentx","Snippet":"short","Summary":"model-ready summary","PublishTime":"2026-08-18T10:00:00+08:00","SiteName":"Example Docs","RankScore":0.97,"AuthInfoDes":"非常权威","AuthInfoLevel":1}]}}`
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response)), Request: request}, nil
	})
	payload, err := RunSearch(context.Background(), SearchRunOptions{
		Provider: SearchProviderDoubaoCustom, Query: "AgentX", Count: 2,
		DoubaoCustomEndpoint: DefaultSearchDoubaoCustomURL, DoubaoCustomAPIKey: "search-key", DoubaoCustomTimeRange: "OneWeek",
		DoubaoSites: []string{"docs.example.com"}, DoubaoBlockedHosts: []string{"spam.example.com"},
		AuthoritativeOnly: true, QueryRewrite: true, Prepare: prepare,
	})
	if err != nil {
		t.Fatalf("RunSearch: %v", err)
	}
	if payload.Provider != SearchProviderDoubaoCustom || payload.RequestID != "request:custom" || len(payload.Results) != 1 {
		t.Fatalf("payload = %#v", payload)
	}
	result := payload.Results[0]
	if result.Description != "model-ready summary" || result.Score != 0.97 || result.AuthorityLevel != 1 || result.Authority != "非常权威" {
		t.Fatalf("result = %#v", result)
	}
}

func TestRunDoubaoCustomSearchReturnsTypedCredentialError(t *testing.T) {
	t.Parallel()
	prepare := searchPreparer(t, func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"message":"invalid_api_key"}`)), Request: request}, nil
	})
	_, err := RunSearch(context.Background(), SearchRunOptions{
		Provider: SearchProviderDoubaoCustom, Query: "AgentX", Count: 1,
		DoubaoCustomEndpoint: DefaultSearchDoubaoCustomURL, DoubaoCustomAPIKey: "wrong-key", Prepare: prepare,
	})
	var providerErr *SearchProviderError
	if !errors.As(err, &providerErr) || providerErr.Health != ProviderHealthCredentialInvalid || providerErr.Retryable {
		t.Fatalf("error = %#v", err)
	}
	if strings.Contains(err.Error(), "wrong-key") || strings.Contains(err.Error(), "invalid_api_key") {
		t.Fatalf("error should remain display-safe: %v", err)
	}
}

func TestMapOpenDoubaoCustomDateRangeUsesProvidedClock(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 18, 23, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	if got := MapDateRangeForDoubaoCustom("2026-08-01", "", now); got != "2026-08-01..2026-08-18" {
		t.Fatalf("date range = %q", got)
	}
}

func TestRunDoubaoGlobalSearchProtocol(t *testing.T) {
	t.Parallel()
	prepare := searchPreparer(t, func(request *http.Request) (*http.Response, error) {
		if request.Header.Get("Authorization") != "Bearer global-key" {
			t.Fatalf("authorization header was not populated")
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		filter, _ := body["Filter"].(map[string]any)
		if body["SearchType"] != "web" || body["DocCount"] != float64(3) ||
			body["MaxSnippetLength"] != float64(800) || body["MaxImageCountPerDoc"] != float64(0) ||
			filter["IcpHostOnly"] != true {
			t.Fatalf("request body = %#v", body)
		}
		response := `{"ResponseMetadata":{"RequestId":"request:global"},"Result":{"TotalDocCount":20,"Documents":[{"Rank":0,"Url":"https://docs.example.com/agentx","Title":"AgentX","Snippet":[{"Type":"text","Text":"first"},{"Type":"image","Image":{"ImageUrl":"https://img.example.com/a.png"}},{"Type":"text","Text":"second"}],"DocumentInfo":{"PublishTime":"2026-08-18"},"HostInfo":{"Hostname":"Example Docs","AuthorityLevel":"very_high"}}],"ErrorCode":0,"ErrorMsg":""}}`
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response)), Request: request}, nil
	})
	payload, err := RunSearch(context.Background(), SearchRunOptions{
		Provider: SearchProviderDoubaoGlobal, Query: "AgentX", Count: 3,
		DoubaoGlobalEndpoint: DefaultSearchDoubaoGlobalURL, DoubaoGlobalAPIKey: "global-key",
		DoubaoGlobalMaxSnippetTokens: 800, DoubaoGlobalICPHostOnly: true, Prepare: prepare,
	})
	if err != nil {
		t.Fatalf("RunSearch: %v", err)
	}
	if payload.Provider != SearchProviderDoubaoGlobal || payload.RequestID != "request:global" || len(payload.Results) != 1 {
		t.Fatalf("payload = %#v", payload)
	}
	result := payload.Results[0]
	if result.Description != "first\nsecond" || result.Published != "2026-08-18" || result.SiteName != "Example Docs" || result.Authority != "very_high" {
		t.Fatalf("result = %#v", result)
	}
}

func TestRunDoubaoGlobalSearchReturnsTypedPlanError(t *testing.T) {
	t.Parallel()
	prepare := searchPreparer(t, func(request *http.Request) (*http.Response, error) {
		response := `{"ResponseMetadata":{"RequestId":"request:global-error"},"Result":{"ErrorCode":10409,"ErrorMsg":"subscription plan unsupported"}}`
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response)), Request: request}, nil
	})
	_, err := RunSearch(context.Background(), SearchRunOptions{
		Provider: SearchProviderDoubaoGlobal, Query: "AgentX", Count: 1,
		DoubaoGlobalEndpoint: DefaultSearchDoubaoGlobalURL, DoubaoGlobalAPIKey: "subscription-key", Prepare: prepare,
	})
	var providerErr *SearchProviderError
	if !errors.As(err, &providerErr) || providerErr.Code != "10409" || providerErr.Health != ProviderHealthUnsupportedFeature || providerErr.Retryable {
		t.Fatalf("error = %#v", err)
	}
	if strings.Contains(err.Error(), "subscription plan unsupported") || strings.Contains(err.Error(), "subscription-key") {
		t.Fatalf("error should remain display-safe: %v", err)
	}
}

func TestPrepareDoubaoGlobalRejectsUnsupportedDateFilter(t *testing.T) {
	t.Parallel()
	_, validation := PrepareSearchRequest(SearchPrepareOptions{
		ConfiguredProvider: SearchProviderDoubaoGlobal, Query: "AgentX", Count: 3, DateAfter: "2026-08-01",
	})
	if validation == nil || validation.Code != "unsupported_date_filter" {
		t.Fatalf("validation = %#v", validation)
	}
}
