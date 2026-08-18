package retrieval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/wsnacj/agentx-go/tools/httprequest"
)

const defaultSearchTimeoutMs = 15_000

type SearchPayload struct {
	Query                  string               `json:"query"`
	Provider               string               `json:"provider"`
	ProviderKind           string               `json:"provider_kind,omitempty"`
	ResultKind             string               `json:"result_kind,omitempty"`
	RequestedProvider      string               `json:"requested_provider,omitempty"`
	ProviderNote           string               `json:"provider_note,omitempty"`
	ProviderFallbackFrom   string               `json:"provider_fallback_from,omitempty"`
	ProviderFallbackReason string               `json:"provider_fallback_reason,omitempty"`
	ProviderDiagnostics    *ProviderDiagnostics `json:"provider_diagnostics,omitempty"`
	Count                  int                  `json:"count"`
	TookMs                 int64                `json:"took_ms"`
	RequestID              string               `json:"request_id,omitempty"`
	Results                []SearchResult       `json:"results,omitempty"`
	Cached                 bool                 `json:"cached,omitempty"`
}

type SearchResult struct {
	Title          string  `json:"title"`
	URL            string  `json:"url"`
	Description    string  `json:"description"`
	Published      string  `json:"published,omitempty"`
	SiteName       string  `json:"site_name,omitempty"`
	Score          float64 `json:"score,omitempty"`
	Authority      string  `json:"authority,omitempty"`
	AuthorityLevel int     `json:"authority_level,omitempty"`
}

type SearchRunOptions struct {
	Provider                     string
	Query                        string
	Count                        int
	Country                      string
	SearchLang                   string
	UILang                       string
	Freshness                    string
	TimeoutMs                    int
	BraveEndpoint                string
	BraveAPIKey                  string
	DoubaoCustomEndpoint         string
	DoubaoCustomAPIKey           string
	DoubaoCustomTimeRange        string
	DoubaoGlobalEndpoint         string
	DoubaoGlobalAPIKey           string
	DoubaoGlobalMaxSnippetTokens int
	DoubaoGlobalICPHostOnly      bool
	// Deprecated: use DoubaoCustomEndpoint. Retained for source compatibility.
	ArkEndpoint string
	// Deprecated: use DoubaoCustomAPIKey. Retained for source compatibility.
	ArkAPIKey string
	// Deprecated: use DoubaoCustomTimeRange. Retained for source compatibility.
	ArkTimeRange           string
	DoubaoSites            []string
	DoubaoBlockedHosts     []string
	AuthoritativeOnly      bool
	QueryRewrite           bool
	BaiduEndpoint          string
	BaiduAPIKey            string
	BaiduRecency           string
	PerplexityAPIKey       string
	PerplexityBaseURL      string
	PerplexityModel        string
	PerplexityRecency      string
	PerplexityDateAfter    string
	PerplexityDateBefore   string
	PerplexityDomainFilter []string
	Prepare                Preparer
	ClassifyNetworkError   NetworkErrorClassifier
}

type braveSearchResponse struct {
	Mixed struct {
		Main []struct {
			Type  string `json:"type"`
			Index *int   `json:"index,omitempty"`
			All   bool   `json:"all,omitempty"`
		} `json:"main"`
	} `json:"mixed"`
	Web struct {
		Results []braveSearchResult `json:"results"`
	} `json:"web"`
	News struct {
		Results []braveSearchResult `json:"results"`
	} `json:"news"`
}

type braveSearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Age         string `json:"age"`
	PageAge     string `json:"page_age"`
	MetaURL     struct {
		Hostname string `json:"hostname"`
	} `json:"meta_url"`
}

type doubaoCustomSearchResponse struct {
	Code             json.RawMessage `json:"code,omitempty"`
	Message          string          `json:"message,omitempty"`
	ResponseMetadata struct {
		RequestID string `json:"RequestId"`
		Error     *struct {
			CodeN   int    `json:"CodeN"`
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error,omitempty"`
	} `json:"ResponseMetadata"`
	Result *struct {
		ResultCount int `json:"ResultCount"`
		WebResults  []struct {
			Title         string  `json:"Title"`
			URL           string  `json:"Url"`
			Snippet       string  `json:"Snippet"`
			Summary       string  `json:"Summary"`
			PublishTime   string  `json:"PublishTime"`
			SiteName      string  `json:"SiteName"`
			RankScore     float64 `json:"RankScore"`
			AuthInfoDes   string  `json:"AuthInfoDes"`
			AuthInfoLevel int     `json:"AuthInfoLevel"`
		} `json:"WebResults"`
	} `json:"Result"`
}

type doubaoGlobalSearchResponse struct {
	Code             json.RawMessage `json:"code,omitempty"`
	Message          string          `json:"message,omitempty"`
	ResponseMetadata struct {
		RequestID string `json:"RequestId"`
		Error     *struct {
			Code    string `json:"Code"`
			CodeN   int    `json:"CodeN"`
			Message string `json:"Message"`
		} `json:"Error,omitempty"`
	} `json:"ResponseMetadata"`
	Result *struct {
		TotalDocCount int    `json:"TotalDocCount"`
		ErrorCode     int    `json:"ErrorCode"`
		ErrorMsg      string `json:"ErrorMsg"`
		Documents     []struct {
			Rank    int    `json:"Rank"`
			URL     string `json:"Url"`
			Title   string `json:"Title"`
			Snippet []struct {
				Type string `json:"Type"`
				Text string `json:"Text"`
			} `json:"Snippet"`
			DocumentInfo struct {
				PublishTime string `json:"PublishTime"`
			} `json:"DocumentInfo"`
			HostInfo struct {
				Hostname       string `json:"Hostname"`
				AuthorityLevel string `json:"AuthorityLevel"`
			} `json:"HostInfo"`
		} `json:"Documents"`
	} `json:"Result"`
}

type baiduSearchResponse struct {
	RequestID  string `json:"request_id"`
	Code       *int   `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	References []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Website string `json:"website,omitempty"`
		Content string `json:"content,omitempty"`
		Date    string `json:"date,omitempty"`
	} `json:"references,omitempty"`
}

type perplexitySearchResponse struct {
	Citations     any `json:"citations"`
	SearchResults []struct {
		Title    string `json:"title"`
		URL      string `json:"url"`
		Snippet  string `json:"snippet"`
		Date     string `json:"date"`
		SiteName string `json:"site_name"`
	} `json:"search_results"`
	Choices []struct {
		Message struct {
			Content any `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func RunSearch(ctx context.Context, opts SearchRunOptions) (SearchPayload, error) {
	switch opts.Provider {
	case "perplexity", "openrouter":
		return runPerplexitySearch(ctx, opts)
	case "brave":
		return runBraveSearch(ctx, opts)
	case SearchProviderDoubaoCustom:
		return runDoubaoCustomSearch(ctx, opts)
	case SearchProviderDoubaoGlobal:
		return runDoubaoGlobalSearch(ctx, opts)
	case "baidu":
		return runBaiduSearch(ctx, opts)
	default:
		return SearchPayload{}, fmt.Errorf("web_search: unsupported provider %q", opts.Provider)
	}
}

func prepareSearchEndpoint(ctx context.Context, prepare Preparer, rawURL string, timeoutMs int) (httprequest.PreparedRequest, error) {
	if timeoutMs <= 0 {
		timeoutMs = defaultSearchTimeoutMs
	}
	if prepare == nil {
		return httprequest.PreparedRequest{}, fmt.Errorf("request preparer is unavailable")
	}
	prepared, err := prepare(ctx, httprequest.PrepareInput{RawURL: rawURL, TimeoutMs: timeoutMs, FollowRedirects: true, MaxRedirects: 5, CredentialSensitive: true})
	if err != nil {
		return httprequest.PreparedRequest{}, err
	}
	if prepared.URL == nil || prepared.Doer == nil {
		if prepared.Close != nil {
			prepared.Close()
		}
		return httprequest.PreparedRequest{}, fmt.Errorf("request preparer returned an incomplete client")
	}
	return prepared, nil
}

func runBraveSearch(ctx context.Context, opts SearchRunOptions) (SearchPayload, error) {
	started := time.Now()
	prepared, err := prepareSearchEndpoint(ctx, opts.Prepare, opts.BraveEndpoint, opts.TimeoutMs)
	if err != nil {
		return SearchPayload{}, fmt.Errorf("web_search: invalid endpoint: %w", err)
	}
	if prepared.Close != nil {
		defer prepared.Close()
	}
	endpoint := prepared.URL
	q := endpoint.Query()
	q.Set("q", opts.Query)
	q.Set("count", strconv.Itoa(opts.Count))
	if opts.Country != "" {
		q.Set("country", opts.Country)
	}
	if opts.SearchLang != "" {
		q.Set("search_lang", opts.SearchLang)
	}
	if opts.UILang != "" {
		q.Set("ui_lang", opts.UILang)
	}
	if opts.Freshness != "" {
		q.Set("freshness", opts.Freshness)
	}
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return SearchPayload{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", opts.BraveAPIKey)
	resp, err := prepared.Doer.Do(req)
	if err != nil {
		return SearchPayload{}, formatSearchRequestError("brave", endpoint.String(), err, opts.ClassifyNetworkError)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		message := strings.TrimSpace(string(detail))
		if message == "" {
			message = resp.Status
		}
		return SearchPayload{}, fmt.Errorf("web_search: brave api error (%d): %s", resp.StatusCode, truncate(message, 320))
	}
	var parsed braveSearchResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&parsed); err != nil {
		return SearchPayload{}, fmt.Errorf("web_search: decode response: %w", err)
	}
	results := normalizeBraveSearchResults(parsed, opts.Count)
	return SearchPayload{
		Query:        opts.Query,
		Provider:     "brave",
		ProviderKind: searchProviderKindString(ResolveSearchProviderKind("brave")),
		ResultKind:   searchResultKindString(ResolveSearchResultKind("brave")),
		Count:        len(results),
		TookMs:       time.Since(started).Milliseconds(),
		Results:      results,
	}, nil
}

func normalizeBraveSearchResults(parsed braveSearchResponse, maxResults int) []SearchResult {
	limit := maxResults
	if limit <= 0 {
		limit = len(parsed.Web.Results) + len(parsed.News.Results)
	}
	if limit <= 0 {
		return nil
	}
	results := make([]SearchResult, 0, limit)
	seen := map[string]bool{}
	appendItem := func(item braveSearchResult) {
		if len(results) >= limit {
			return
		}
		current := SearchResult{
			Title:       strings.TrimSpace(item.Title),
			URL:         strings.TrimSpace(item.URL),
			Description: strings.TrimSpace(item.Description),
			Published:   firstSearchNonEmpty(item.PageAge, item.Age),
			SiteName:    strings.TrimSpace(item.MetaURL.Hostname),
		}
		if current.URL == "" {
			return
		}
		key := strings.ToLower(current.URL)
		if seen[key] {
			return
		}
		seen[key] = true
		if current.SiteName == "" {
			current.SiteName = resolveSiteName(current.URL)
		}
		results = append(results, current)
	}
	if len(parsed.Mixed.Main) > 0 {
		for _, item := range parsed.Mixed.Main {
			switch strings.ToLower(strings.TrimSpace(item.Type)) {
			case "web":
				if item.Index != nil && *item.Index >= 0 && *item.Index < len(parsed.Web.Results) {
					appendItem(parsed.Web.Results[*item.Index])
				} else if item.All {
					for _, candidate := range parsed.Web.Results {
						appendItem(candidate)
					}
				}
			case "news":
				if item.Index != nil && *item.Index >= 0 && *item.Index < len(parsed.News.Results) {
					appendItem(parsed.News.Results[*item.Index])
				} else if item.All {
					for _, candidate := range parsed.News.Results {
						appendItem(candidate)
					}
				}
			}
			if len(results) >= limit {
				return results
			}
		}
	}
	for _, item := range parsed.Web.Results {
		appendItem(item)
	}
	for _, item := range parsed.News.Results {
		appendItem(item)
	}
	return results
}

func runDoubaoCustomSearch(ctx context.Context, opts SearchRunOptions) (SearchPayload, error) {
	started := time.Now()
	endpointValue := firstNonEmpty(opts.DoubaoCustomEndpoint, opts.ArkEndpoint)
	apiKey := firstNonEmpty(opts.DoubaoCustomAPIKey, opts.ArkAPIKey)
	timeRange := firstNonEmpty(opts.DoubaoCustomTimeRange, opts.ArkTimeRange)
	prepared, err := prepareSearchEndpoint(ctx, opts.Prepare, endpointValue, opts.TimeoutMs)
	if err != nil {
		return SearchPayload{}, fmt.Errorf("web_search: invalid endpoint: %w", err)
	}
	if prepared.Close != nil {
		defer prepared.Close()
	}
	endpoint := prepared.URL
	reqBody := map[string]any{
		"Query":      opts.Query,
		"SearchType": "web",
	}
	if opts.Count > 0 {
		reqBody["Count"] = opts.Count
	}
	if strings.TrimSpace(timeRange) != "" {
		reqBody["TimeRange"] = strings.TrimSpace(timeRange)
	}
	filter := map[string]any{"NeedUrl": true, "NeedContent": false}
	if len(opts.DoubaoSites) > 0 {
		filter["Sites"] = strings.Join(opts.DoubaoSites, "|")
	}
	if len(opts.DoubaoBlockedHosts) > 0 {
		filter["BlockHosts"] = strings.Join(opts.DoubaoBlockedHosts, "|")
	}
	if opts.AuthoritativeOnly {
		filter["AuthInfoLevel"] = 1
	}
	reqBody["Filter"] = filter
	if opts.QueryRewrite {
		reqBody["QueryControl"] = map[string]any{"QueryRewrite": true}
	}
	blob, err := json.Marshal(reqBody)
	if err != nil {
		return SearchPayload{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(string(blob)))
	if err != nil {
		return SearchPayload{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := prepared.Doer.Do(req)
	if err != nil {
		return SearchPayload{}, formatSearchRequestError(SearchProviderDoubaoCustom, endpoint.String(), err, opts.ClassifyNetworkError)
	}
	defer resp.Body.Close()
	rawBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if readErr != nil {
		return SearchPayload{}, fmt.Errorf("web_search: decode response: %w", readErr)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code, message := doubaoTopLevelError(rawBody)
		return SearchPayload{}, newDoubaoProviderError(SearchProviderDoubaoCustom, code, message, resp.StatusCode)
	}
	var parsed doubaoCustomSearchResponse
	if err := json.Unmarshal(rawBody, &parsed); err != nil {
		return SearchPayload{}, fmt.Errorf("web_search: decode response: %w", err)
	}
	if parsed.ResponseMetadata.Error != nil {
		e := parsed.ResponseMetadata.Error
		code := strings.TrimSpace(e.Code)
		if code == "" && e.CodeN != 0 {
			code = strconv.Itoa(e.CodeN)
		}
		return SearchPayload{}, newDoubaoProviderError(SearchProviderDoubaoCustom, code, e.Message, resp.StatusCode)
	}
	if code := rawJSONScalarString(parsed.Code); code != "" || (parsed.Result == nil && strings.TrimSpace(parsed.Message) != "") {
		return SearchPayload{}, newDoubaoProviderError(SearchProviderDoubaoCustom, code, parsed.Message, resp.StatusCode)
	}
	results := make([]SearchResult, 0)
	if parsed.Result != nil {
		for _, item := range parsed.Result.WebResults {
			if len(results) >= opts.Count {
				break
			}
			description := strings.TrimSpace(item.Summary)
			if description == "" {
				description = strings.TrimSpace(item.Snippet)
			}
			current := SearchResult{
				Title:          strings.TrimSpace(item.Title),
				URL:            strings.TrimSpace(item.URL),
				Description:    description,
				Published:      strings.TrimSpace(item.PublishTime),
				SiteName:       strings.TrimSpace(item.SiteName),
				Score:          item.RankScore,
				Authority:      strings.TrimSpace(item.AuthInfoDes),
				AuthorityLevel: item.AuthInfoLevel,
			}
			if current.SiteName == "" {
				current.SiteName = resolveSiteName(current.URL)
			}
			results = append(results, current)
		}
	}
	return SearchPayload{
		Query:        opts.Query,
		Provider:     SearchProviderDoubaoCustom,
		ProviderKind: searchProviderKindString(ResolveSearchProviderKind(SearchProviderDoubaoCustom)),
		ResultKind:   searchResultKindString(ResolveSearchResultKind(SearchProviderDoubaoCustom)),
		Count:        len(results),
		TookMs:       time.Since(started).Milliseconds(),
		RequestID:    strings.TrimSpace(parsed.ResponseMetadata.RequestID),
		Results:      results,
	}, nil
}

func runDoubaoGlobalSearch(ctx context.Context, opts SearchRunOptions) (SearchPayload, error) {
	started := time.Now()
	prepared, err := prepareSearchEndpoint(ctx, opts.Prepare, opts.DoubaoGlobalEndpoint, opts.TimeoutMs)
	if err != nil {
		return SearchPayload{}, err
	}
	endpoint := prepared.URL
	requestBody := map[string]any{
		"Query":      opts.Query,
		"SearchType": "web",
		"DocCount":   opts.Count,
	}
	maxSnippetTokens := opts.DoubaoGlobalMaxSnippetTokens
	if maxSnippetTokens <= 0 {
		maxSnippetTokens = 500
	}
	if maxSnippetTokens > 3000 {
		return SearchPayload{}, fmt.Errorf("web_search: doubao_global max snippet tokens must be between 1 and 3000")
	}
	requestBody["MaxSnippetLength"] = maxSnippetTokens
	requestBody["MaxImageCountPerDoc"] = 0
	if opts.DoubaoGlobalICPHostOnly {
		requestBody["Filter"] = map[string]any{"IcpHostOnly": true}
	}
	blob, err := json.Marshal(requestBody)
	if err != nil {
		return SearchPayload{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(blob))
	if err != nil {
		return SearchPayload{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+opts.DoubaoGlobalAPIKey)
	resp, err := prepared.Doer.Do(req)
	if err != nil {
		return SearchPayload{}, formatSearchRequestError(SearchProviderDoubaoGlobal, endpoint.String(), err, opts.ClassifyNetworkError)
	}
	defer resp.Body.Close()
	rawBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if readErr != nil {
		return SearchPayload{}, fmt.Errorf("web_search: decode response: %w", readErr)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		code, message := doubaoTopLevelError(rawBody)
		return SearchPayload{}, newDoubaoProviderError(SearchProviderDoubaoGlobal, code, message, resp.StatusCode)
	}
	var parsed doubaoGlobalSearchResponse
	if err := json.Unmarshal(rawBody, &parsed); err != nil {
		return SearchPayload{}, fmt.Errorf("web_search: decode response: %w", err)
	}
	if parsed.ResponseMetadata.Error != nil {
		providerError := parsed.ResponseMetadata.Error
		code := strings.TrimSpace(providerError.Code)
		if code == "" && providerError.CodeN != 0 {
			code = strconv.Itoa(providerError.CodeN)
		}
		return SearchPayload{}, newDoubaoProviderError(SearchProviderDoubaoGlobal, code, providerError.Message, resp.StatusCode)
	}
	if code := rawJSONScalarString(parsed.Code); code != "" || (parsed.Result == nil && strings.TrimSpace(parsed.Message) != "") {
		return SearchPayload{}, newDoubaoProviderError(SearchProviderDoubaoGlobal, code, parsed.Message, resp.StatusCode)
	}
	if parsed.Result != nil && parsed.Result.ErrorCode != 0 {
		return SearchPayload{}, newDoubaoProviderError(SearchProviderDoubaoGlobal, strconv.Itoa(parsed.Result.ErrorCode), parsed.Result.ErrorMsg, resp.StatusCode)
	}
	results := make([]SearchResult, 0)
	if parsed.Result != nil {
		for _, document := range parsed.Result.Documents {
			if len(results) >= opts.Count {
				break
			}
			text := make([]string, 0, len(document.Snippet))
			for _, snippet := range document.Snippet {
				if strings.EqualFold(strings.TrimSpace(snippet.Type), "text") && strings.TrimSpace(snippet.Text) != "" {
					text = append(text, strings.TrimSpace(snippet.Text))
				}
			}
			current := SearchResult{
				Title:       strings.TrimSpace(document.Title),
				URL:         strings.TrimSpace(document.URL),
				Description: strings.Join(text, "\n"),
				Published:   strings.TrimSpace(document.DocumentInfo.PublishTime),
				SiteName:    strings.TrimSpace(document.HostInfo.Hostname),
				Authority:   strings.TrimSpace(document.HostInfo.AuthorityLevel),
			}
			if current.SiteName == "" {
				current.SiteName = resolveSiteName(current.URL)
			}
			results = append(results, current)
		}
	}
	return SearchPayload{
		Query: opts.Query, Provider: SearchProviderDoubaoGlobal,
		ProviderKind: searchProviderKindString(ResolveSearchProviderKind(SearchProviderDoubaoGlobal)),
		ResultKind:   searchResultKindString(ResolveSearchResultKind(SearchProviderDoubaoGlobal)),
		Count:        len(results), TookMs: time.Since(started).Milliseconds(),
		RequestID: strings.TrimSpace(parsed.ResponseMetadata.RequestID), Results: results,
	}, nil
}

func doubaoTopLevelError(raw []byte) (string, string) {
	var payload struct {
		Code    json.RawMessage `json:"code"`
		Message string          `json:"message"`
	}
	if json.Unmarshal(raw, &payload) != nil {
		return "", ""
	}
	return rawJSONScalarString(payload.Code), strings.TrimSpace(payload.Message)
}

func rawJSONScalarString(raw json.RawMessage) string {
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(string(raw))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func runBaiduSearch(ctx context.Context, opts SearchRunOptions) (SearchPayload, error) {
	started := time.Now()
	prepared, err := prepareSearchEndpoint(ctx, opts.Prepare, opts.BaiduEndpoint, opts.TimeoutMs)
	if err != nil {
		return SearchPayload{}, fmt.Errorf("web_search: invalid endpoint: %w", err)
	}
	if prepared.Close != nil {
		defer prepared.Close()
	}
	endpoint := prepared.URL
	reqBody := map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": opts.Query},
		},
		"search_source": "baidu_search_v2",
		"resource_type_filter": []map[string]any{
			{"type": "web", "top_k": opts.Count},
			{"type": "video", "top_k": 0},
			{"type": "image", "top_k": 0},
		},
	}
	if strings.TrimSpace(opts.BaiduRecency) != "" {
		reqBody["search_recency_filter"] = strings.TrimSpace(opts.BaiduRecency)
	}
	requestBody, err := json.Marshal(reqBody)
	if err != nil {
		return SearchPayload{}, err
	}
	parsed, rawBody, statusCode, err := doBaiduSearchOnce(ctx, prepared.Doer, endpoint.String(), requestBody, "Authorization", "Bearer "+opts.BaiduAPIKey)
	if err != nil {
		return SearchPayload{}, formatSearchRequestError("baidu", endpoint.String(), err, opts.ClassifyNetworkError)
	}
	if statusCode == http.StatusUnauthorized || hasInvalidBaiduAuthResponse(rawBody) {
		parsed, rawBody, statusCode, err = doBaiduSearchOnce(ctx, prepared.Doer, endpoint.String(), requestBody, "X-Appbuilder-Authorization", "Bearer "+opts.BaiduAPIKey)
		if err != nil {
			return SearchPayload{}, formatSearchRequestError("baidu", endpoint.String(), err, opts.ClassifyNetworkError)
		}
	}
	if statusCode < 200 || statusCode >= 300 {
		message := strings.TrimSpace(string(rawBody))
		if message == "" {
			message = http.StatusText(statusCode)
		}
		return SearchPayload{}, fmt.Errorf("web_search: baidu api error (%d): %s", statusCode, truncate(message, 320))
	}
	if parsed.Code != nil && *parsed.Code != 0 {
		return SearchPayload{}, fmt.Errorf("web_search: baidu api error (code=%d): %s", *parsed.Code, truncate(strings.TrimSpace(parsed.Message), 320))
	}
	results := make([]SearchResult, 0, len(parsed.References))
	for _, item := range parsed.References {
		if len(results) >= opts.Count {
			break
		}
		current := SearchResult{
			Title:       strings.TrimSpace(item.Title),
			URL:         strings.TrimSpace(item.URL),
			Description: truncate(strings.TrimSpace(item.Content), 280),
			Published:   strings.TrimSpace(item.Date),
			SiteName:    strings.TrimSpace(item.Website),
		}
		if current.SiteName == "" {
			current.SiteName = resolveSiteName(current.URL)
		}
		results = append(results, current)
	}
	return SearchPayload{
		Query:        opts.Query,
		Provider:     "baidu",
		ProviderKind: searchProviderKindString(ResolveSearchProviderKind("baidu")),
		ResultKind:   searchResultKindString(ResolveSearchResultKind("baidu")),
		Count:        len(results),
		TookMs:       time.Since(started).Milliseconds(),
		Results:      results,
	}, nil
}

func doBaiduSearchOnce(ctx context.Context, client httprequest.HTTPDoer, endpoint string, requestBody []byte, authHeader string, authValue string) (baiduSearchResponse, []byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(requestBody)))
	if err != nil {
		return baiduSearchResponse{}, nil, 0, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "agentx-web-search/1.0")
	req.Header.Set(authHeader, authValue)

	resp, err := client.Do(req)
	if err != nil {
		return baiduSearchResponse{}, nil, 0, err
	}
	defer resp.Body.Close()
	rawBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	parsed := baiduSearchResponse{}
	_ = json.Unmarshal(rawBody, &parsed)
	return parsed, rawBody, resp.StatusCode, nil
}

func runPerplexitySearch(ctx context.Context, opts SearchRunOptions) (SearchPayload, error) {
	started := time.Now()
	prepared, err := prepareSearchEndpoint(ctx, opts.Prepare, opts.PerplexityBaseURL, opts.TimeoutMs)
	if err != nil {
		return SearchPayload{}, fmt.Errorf("web_search: invalid endpoint: %w", err)
	}
	if prepared.Close != nil {
		defer prepared.Close()
	}
	endpoint := prepared.URL
	userQuery := buildPerplexityUserQuery(opts)
	body := map[string]any{
		"model": opts.PerplexityModel,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a search assistant. Prefer concise factual output and include source-backed claims.",
			},
			{
				"role":    "user",
				"content": userQuery,
			},
		},
	}
	if opts.PerplexityRecency != "" {
		body["search_recency_filter"] = opts.PerplexityRecency
	}
	if opts.Provider == "perplexity" {
		if opts.Country != "" {
			body["country"] = opts.Country
		}
		if opts.SearchLang != "" {
			body["search_language_filter"] = []string{opts.SearchLang}
		}
		if opts.PerplexityDateAfter != "" {
			body["search_after_date_filter"] = opts.PerplexityDateAfter
		}
		if opts.PerplexityDateBefore != "" {
			body["search_before_date_filter"] = opts.PerplexityDateBefore
		}
		if len(opts.PerplexityDomainFilter) > 0 {
			body["search_domain_filter"] = opts.PerplexityDomainFilter
		}
	}
	requestBody, err := json.Marshal(body)
	if err != nil {
		return SearchPayload{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(string(requestBody)))
	if err != nil {
		return SearchPayload{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+opts.PerplexityAPIKey)
	if opts.Provider == "openrouter" {
		req.Header.Set("HTTP-Referer", "https://agentx.local")
		req.Header.Set("X-Title", "agentx-web-search")
	}
	resp, err := prepared.Doer.Do(req)
	if err != nil {
		return SearchPayload{}, formatSearchRequestError(opts.Provider, endpoint.String(), err, opts.ClassifyNetworkError)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		message := strings.TrimSpace(string(detail))
		if message == "" {
			message = resp.Status
		}
		return SearchPayload{}, fmt.Errorf("web_search: %s api error (%d): %s", opts.Provider, resp.StatusCode, truncate(message, 320))
	}
	var parsed perplexitySearchResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&parsed); err != nil {
		return SearchPayload{}, fmt.Errorf("web_search: decode response: %w", err)
	}
	answer := ""
	if len(parsed.Choices) > 0 {
		answer = extractPerplexityAnswerText(parsed.Choices[0].Message.Content)
	}
	results := make([]SearchResult, 0, maxInt(len(parsed.SearchResults), opts.Count))
	for _, item := range parsed.SearchResults {
		if len(results) >= opts.Count {
			break
		}
		current := SearchResult{
			Title:       strings.TrimSpace(item.Title),
			URL:         strings.TrimSpace(item.URL),
			Description: strings.TrimSpace(item.Snippet),
			Published:   strings.TrimSpace(item.Date),
			SiteName:    strings.TrimSpace(item.SiteName),
		}
		if current.Title == "" {
			current.Title = buildCitationTitle(current.URL)
		}
		if current.SiteName == "" {
			current.SiteName = resolveSiteName(current.URL)
		}
		if current.Description == "" {
			current.Description = truncate(answer, 280)
		}
		results = append(results, current)
	}
	if len(results) == 0 {
		citations := extractPerplexityCitations(parsed.Citations)
		for _, citeURL := range citations {
			if len(results) >= opts.Count {
				break
			}
			item := SearchResult{
				Title:       buildCitationTitle(citeURL),
				URL:         citeURL,
				Description: truncate(answer, 280),
			}
			if host := resolveSiteName(citeURL); host != "" {
				item.SiteName = host
			}
			results = append(results, item)
		}
	}
	if len(results) == 0 && strings.TrimSpace(answer) != "" {
		results = append(results, SearchResult{
			Title:       "answer",
			URL:         "",
			Description: truncate(answer, 280),
		})
	}
	return SearchPayload{
		Query:        opts.Query,
		Provider:     opts.Provider,
		ProviderKind: searchProviderKindString(ResolveSearchProviderKind(opts.Provider)),
		ResultKind:   searchResultKindString(ResolveSearchResultKind(opts.Provider)),
		Count:        len(results),
		TookMs:       time.Since(started).Milliseconds(),
		Results:      results,
	}, nil
}

func buildPerplexityUserQuery(opts SearchRunOptions) string {
	parts := []string{
		strings.TrimSpace(opts.Query),
	}
	if opts.Count > 0 {
		parts = append(parts, fmt.Sprintf("Return up to %d source links.", opts.Count))
	}
	if opts.Country != "" {
		parts = append(parts, "Country preference: "+opts.Country)
	}
	if opts.SearchLang != "" {
		parts = append(parts, "Search language: "+opts.SearchLang)
	}
	if opts.UILang != "" {
		parts = append(parts, "UI language: "+opts.UILang)
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func buildCitationTitle(rawURL string) string {
	site := resolveSiteName(rawURL)
	if site == "" {
		return "source"
	}
	return site
}

func extractPerplexityAnswerText(content any) string {
	switch typed := content.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			switch block := item.(type) {
			case string:
				text := strings.TrimSpace(block)
				if text != "" {
					parts = append(parts, text)
				}
			case map[string]any:
				text := stringFromAny(block["text"])
				if text == "" {
					text = stringFromAny(block["content"])
				}
				text = strings.TrimSpace(text)
				if text != "" {
					parts = append(parts, text)
				}
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	default:
		return ""
	}
}

func extractPerplexityCitations(raw any) []string {
	switch typed := raw.(type) {
	case []any:
		seen := map[string]bool{}
		results := make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(stringFromAny(item))
			if text == "" || seen[text] {
				continue
			}
			parsed, err := url.Parse(text)
			if err != nil || parsed.Scheme == "" || parsed.Host == "" {
				continue
			}
			seen[text] = true
			results = append(results, text)
		}
		return results
	default:
		return nil
	}
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		if typed == nil {
			return ""
		}
		blob, err := json.Marshal(typed)
		if err != nil {
			return ""
		}
		if string(blob) == "null" {
			return ""
		}
		return strings.TrimSpace(string(blob))
	}
}

func formatSearchRequestError(provider string, endpoint string, err error, classify NetworkErrorClassifier) error {
	if err == nil {
		return nil
	}
	if classify != nil && classify(err) == NetworkErrorProxyConfig {
		return fmt.Errorf("web_search: proxy_config_invalid provider=%s endpoint=%s hint=check tools.webSearchTrustedEnvProxy and HTTP_PROXY/HTTPS_PROXY/NO_PROXY: %w", strings.TrimSpace(provider), strings.TrimSpace(endpoint), err)
	}
	return fmt.Errorf("web_search: request_failed provider=%s endpoint=%s: %w", strings.TrimSpace(provider), strings.TrimSpace(endpoint), err)
}

func hasInvalidBaiduAuthResponse(raw []byte) bool {
	lower := strings.ToLower(strings.TrimSpace(string(raw)))
	if lower == "" {
		return false
	}
	return strings.Contains(lower, "invalidhttpauthheader") ||
		strings.Contains(lower, "fail to parse apikey authorization")
}

func resolveSiteName(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Hostname())
}

func truncate(value string, max int) string {
	return truncateWithEllipsis(value, max)
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func firstSearchNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
