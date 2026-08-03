package retrieval

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/wsnacj/agentx-go/tools/httprequest"
)

const (
	webFetchDefaultMaxChars = 50_000
	webFetchHardMaxChars    = 80_000
	webFetchDefaultMaxBytes = 2_000_000
	webFetchHardMaxBytes    = 10_000_000
	webFetchDefaultTimeout  = 30_000
	webFetchHardTimeout     = 120_000
	webFetchRawLimitMin     = 16_000
	webFetchRawLimitMax     = webFetchHardMaxBytes
	webFetchFirecrawlMaxAge = 172_800_000
)

type WebFetchPayload struct {
	URL                 string               `json:"url"`
	FinalURL            string               `json:"final_url"`
	Status              int                  `json:"status"`
	ContentType         string               `json:"content_type,omitempty"`
	Redirected          bool                 `json:"redirected,omitempty"`
	RedirectCount       int                  `json:"redirect_count,omitempty"`
	RedirectChain       []string             `json:"redirect_chain,omitempty"`
	Extractor           string               `json:"extractor,omitempty"`
	Diagnostics         *WebFetchDiagnostics `json:"extractor_diagnostics,omitempty"`
	Handoff             *WebFetchHandoff     `json:"handoff,omitempty"`
	ExternalContent     *ExternalContentMeta `json:"external_content,omitempty"`
	ProviderDiagnostics *ProviderDiagnostics `json:"provider_diagnostics,omitempty"`
	SiteName            string               `json:"site_name,omitempty"`
	Title               string               `json:"title,omitempty"`
	Byline              string               `json:"byline,omitempty"`
	Excerpt             string               `json:"excerpt,omitempty"`
	Content             string               `json:"content"`
	ContentChars        int                  `json:"content_chars,omitempty"`
	WordCount           int                  `json:"word_count,omitempty"`
	ExtractMode         string               `json:"extract_mode,omitempty"`
	FallbackUsed        bool                 `json:"fallback_used,omitempty"`
	ReadabilityUsed     bool                 `json:"readability_used,omitempty"`
	ResponseBytes       int                  `json:"response_bytes,omitempty"`
	BodyTruncated       bool                 `json:"body_truncated,omitempty"`
	Warning             string               `json:"warning,omitempty"`
	Truncated           bool                 `json:"truncated"`
	TruncationReason    string               `json:"truncation_reason,omitempty"`
	FromCache           bool                 `json:"from_cache"`
	FetchedAt           int64                `json:"fetched_at"`
	TookMs              int64                `json:"took_ms,omitempty"`
}

type WebFetchDiagnostics struct {
	RequestedMode     string `json:"requested_mode,omitempty"`
	Extractor         string `json:"extractor,omitempty"`
	Redirected        bool   `json:"redirected,omitempty"`
	RedirectCount     int    `json:"redirect_count,omitempty"`
	CacheHit          bool   `json:"cache_hit,omitempty"`
	FallbackUsed      bool   `json:"fallback_used,omitempty"`
	ReadabilityUsed   bool   `json:"readability_used,omitempty"`
	BinaryUnsupported bool   `json:"binary_unsupported,omitempty"`
}

type WebFetchHandoff struct {
	Kind          string `json:"kind,omitempty"`
	PreferredTool string `json:"preferred_tool,omitempty"`
	Summary       string `json:"summary,omitempty"`
}

type WebFetchExtractResult struct {
	Content         string
	Title           string
	SiteName        string
	Byline          string
	Excerpt         string
	Mode            string
	WordCount       int
	FallbackUsed    bool
	ReadabilityUsed bool
}

type WebFetchFirecrawlConfig struct {
	Enabled         bool
	APIKey          string
	Endpoint        string
	TimeoutMs       int
	MaxAgeMs        int
	OnlyMainContent bool
}

type WebFetchFirecrawlRequest struct {
	URL              string
	FinalURLFallback string
	StatusFallback   int
	ToolName         string
	FallbackReason   string
	FallbackHint     string
	ExtractMode      string
	MaxChars         int
}

type WebFetchFirecrawlResponse struct {
	Success bool   `json:"success"`
	Warning string `json:"warning,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    struct {
		Markdown string `json:"markdown,omitempty"`
		Content  string `json:"content,omitempty"`
		Metadata struct {
			Title      string `json:"title,omitempty"`
			SourceURL  string `json:"sourceURL,omitempty"`
			StatusCode int    `json:"statusCode,omitempty"`
		} `json:"metadata,omitempty"`
	} `json:"data,omitempty"`
}

type WebFetchFirecrawlAttempt struct {
	Payload   WebFetchPayload
	Used      bool
	Attempted bool
	Detail    string
}

type WebFetchExecutionResult struct {
	Payload WebFetchPayload
	Links   []PageLink
}

type ExecuteWebFetchOptions struct {
	ToolName                string
	RawURL                  string
	DefaultExtractMode      string
	RequestedExtractMode    string
	UseCache                bool
	RequestMaxChars         int
	RequestMaxResponseBytes int
	RequestTimeoutMs        int
	MaxRedirects            int
	UserAgent               string
	RequestSignature        string
	CacheTTL                time.Duration
	Firecrawl               WebFetchFirecrawlConfig
	Prepare                 Preparer
	ClassifyNetworkError    NetworkErrorClassifier
}

func ExecuteWebFetch(ctx context.Context, opts ExecuteWebFetchOptions) (WebFetchExecutionResult, error) {
	toolName := strings.TrimSpace(opts.ToolName)
	if toolName == "" {
		toolName = "web_fetch"
	}
	rawURL := strings.TrimSpace(opts.RawURL)
	if rawURL == "" {
		return WebFetchExecutionResult{}, fmt.Errorf("%s: url is required", toolName)
	}
	startedAt := time.Now()
	extractMode := normalizeWebFetchExtractMode(opts.RequestedExtractMode)
	if extractMode == "auto" && strings.TrimSpace(opts.DefaultExtractMode) != "" {
		extractMode = normalizeWebFetchExtractMode(opts.DefaultExtractMode)
	}
	if opts.Prepare == nil {
		return WebFetchExecutionResult{}, fmt.Errorf("%s: request preparer is unavailable", toolName)
	}
	redirectChain := []string{evidenceURL(rawURL)}
	prepared, err := opts.Prepare(ctx, httprequest.PrepareInput{
		RawURL: rawURL, TimeoutMs: opts.RequestTimeoutMs, FollowRedirects: true, MaxRedirects: opts.MaxRedirects,
		OnRedirect: func(_ context.Context, raw string, _ int) {
			raw = evidenceURL(raw)
			if raw != "" {
				redirectChain = append(redirectChain, raw)
			}
		},
	})
	if err != nil {
		return WebFetchExecutionResult{}, &URLPrepareError{Tool: toolName, URL: rawURL, Cause: err}
	}
	if prepared.Close != nil {
		defer prepared.Close()
	}
	if prepared.URL == nil || prepared.Doer == nil {
		return WebFetchExecutionResult{}, fmt.Errorf("%s: request preparer returned an incomplete client", toolName)
	}
	requestURL := prepared.URL.String()
	evidenceRequestURL := evidenceURL(requestURL)
	redirectChain[0] = evidenceRequestURL
	cacheKey := WebFetchCacheKey(requestURL, opts.RequestMaxChars, opts.RequestMaxResponseBytes, normalizeWebFetchExtractMode(extractMode), opts.RequestSignature)
	if opts.UseCache && opts.CacheTTL > 0 {
		if cachedBlob, ok := ReadWebFetchCache(cacheKey); ok {
			var cached WebFetchPayload
			if err := json.Unmarshal(cachedBlob, &cached); err != nil {
				return WebFetchExecutionResult{}, err
			}
			cached.FromCache = true
			if cached.Diagnostics != nil {
				cached.Diagnostics.CacheHit = true
			}
			if cached.ProviderDiagnostics == nil {
				cached.ProviderDiagnostics = BuildFetchProviderDiagnostics(FetchProviderDiagnosticsInput{
					Tool:              toolName,
					EffectiveProvider: FetchProviderForExtractor(firstNonBlank(cached.Extractor, cached.ExtractMode)),
				})
			}
			return WebFetchExecutionResult{Payload: cached}, nil
		}
	}
	deadline := time.Now().Add(time.Duration(opts.RequestTimeoutMs) * time.Millisecond)

	runCtx, cancel := context.WithTimeout(ctx, time.Duration(opts.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(runCtx, http.MethodGet, requestURL, nil)
	if err != nil {
		return WebFetchExecutionResult{}, err
	}
	req.Header.Set("User-Agent", opts.UserAgent)
	req.Header.Set("Accept", "text/markdown, text/html;q=0.9, */*;q=0.1")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	res, err := prepared.Doer.Do(req)
	if err != nil {
		firecrawlAttempt := maybeWebFetchFirecrawlFallback(ctx, deadline, opts.Prepare, opts.ClassifyNetworkError, opts.Firecrawl, WebFetchFirecrawlRequest{
			URL:              requestURL,
			FinalURLFallback: requestURL,
			StatusFallback:   http.StatusOK,
			ToolName:         toolName,
			FallbackReason:   ProviderHealthRequestFailed,
			FallbackHint:     "direct HTTP request failed; external extractor recovered content",
			ExtractMode:      extractMode,
			MaxChars:         opts.RequestMaxChars,
		})
		if firecrawlAttempt.Used {
			firecrawlAttempt.Payload.TookMs = time.Since(startedAt).Milliseconds()
			if opts.UseCache && opts.CacheTTL > 0 {
				writeWebFetchCache(cacheKey, firecrawlAttempt.Payload, opts.CacheTTL)
			}
			return WebFetchExecutionResult{Payload: firecrawlAttempt.Payload}, nil
		}
		return WebFetchExecutionResult{}, formatWebFetchRequestError(toolName, evidenceRequestURL, opts.RequestTimeoutMs, err, firecrawlAttempt.Detail, opts.ClassifyNetworkError)
	}
	defer res.Body.Close()

	rawLimit := opts.RequestMaxResponseBytes
	if rawLimit < webFetchRawLimitMin {
		rawLimit = webFetchRawLimitMin
	}
	if rawLimit > webFetchRawLimitMax {
		rawLimit = webFetchRawLimitMax
	}
	body, rawTruncated, err := readBodyLimited(res.Body, rawLimit)
	if err != nil {
		return WebFetchExecutionResult{}, formatWebFetchRequestError(toolName, evidenceRequestURL, opts.RequestTimeoutMs, err, "", opts.ClassifyNetworkError)
	}
	contentType := strings.TrimSpace(res.Header.Get("Content-Type"))
	normalizedContentType := normalizeWebFetchContentType(contentType)
	warning := ""
	if rawTruncated {
		warning = fmt.Sprintf("Response body truncated after %d bytes.", rawLimit)
	}
	requestFinalURL := res.Request.URL.String()
	finalURL := evidenceURL(requestFinalURL)
	redirected := len(redirectChain) > 1
	siteName := resolveWebFetchSiteName(finalURL)
	links := extractPageLinks(string(body), normalizedContentType, requestFinalURL, 24)
	firecrawlStatusDetail := ""
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		firecrawlAttempt := maybeWebFetchFirecrawlFallback(ctx, deadline, opts.Prepare, opts.ClassifyNetworkError, opts.Firecrawl, WebFetchFirecrawlRequest{
			URL:              requestURL,
			FinalURLFallback: requestFinalURL,
			StatusFallback:   res.StatusCode,
			ToolName:         toolName,
			FallbackReason:   ProviderHealthStatusError,
			FallbackHint:     "direct HTTP status error; external extractor recovered content",
			ExtractMode:      extractMode,
			MaxChars:         opts.RequestMaxChars,
		})
		if firecrawlAttempt.Used {
			firecrawlAttempt.Payload.TookMs = time.Since(startedAt).Milliseconds()
			if opts.UseCache && opts.CacheTTL > 0 {
				writeWebFetchCache(cacheKey, firecrawlAttempt.Payload, opts.CacheTTL)
			}
			return WebFetchExecutionResult{Payload: firecrawlAttempt.Payload}, nil
		}
		firecrawlStatusDetail = firecrawlAttempt.Detail
	}
	if binaryWarning := webFetchBinaryWarning(toolName, contentType, body); binaryWarning != "" {
		warning = joinWebFetchWarnings(warning, binaryWarning)
		payload := WebFetchPayload{
			URL:           evidenceRequestURL,
			FinalURL:      finalURL,
			Status:        res.StatusCode,
			ContentType:   normalizedContentType,
			Redirected:    redirected,
			RedirectCount: len(redirectChain) - 1,
			RedirectChain: webFetchRedirectChainForPayload(redirectChain),
			Extractor:     "binary_unsupported",
			SiteName:      siteName,
			Content:       "",
			ContentChars:  0,
			WordCount:     0,
			ExtractMode:   "binary_unsupported",
			ResponseBytes: len(body),
			BodyTruncated: rawTruncated,
			Warning:       warning,
			Truncated:     false,
			FromCache:     false,
			FetchedAt:     time.Now().Unix(),
			TookMs:        time.Since(startedAt).Milliseconds(),
			Handoff:       webFetchBinaryHandoff(contentType, body),
		}
		finalizeWebFetchPayload(&payload, extractMode, toolName)
		if res.StatusCode < 200 || res.StatusCode >= 400 {
			return WebFetchExecutionResult{}, formatWebFetchStatusError(toolName, payload, firecrawlStatusDetail)
		}
		if opts.UseCache && opts.CacheTTL > 0 {
			writeWebFetchCache(cacheKey, payload, opts.CacheTTL)
		}
		return WebFetchExecutionResult{Payload: payload}, nil
	}
	extract := normalizeFetchContentWithURL(string(body), contentType, extractMode, requestFinalURL)
	if shouldTryWebFetchFirecrawlHTMLFallback(contentType, extract) {
		firecrawlAttempt := maybeWebFetchFirecrawlFallback(ctx, deadline, opts.Prepare, opts.ClassifyNetworkError, opts.Firecrawl, WebFetchFirecrawlRequest{
			URL:              requestURL,
			FinalURLFallback: requestFinalURL,
			StatusFallback:   res.StatusCode,
			ToolName:         toolName,
			FallbackReason:   ProviderHealthUnsupportedFeature,
			FallbackHint:     "local extraction returned low-value content; external extractor recovered content",
			ExtractMode:      extractMode,
			MaxChars:         opts.RequestMaxChars,
		})
		if firecrawlAttempt.Used {
			firecrawlAttempt.Payload.TookMs = time.Since(startedAt).Milliseconds()
			if opts.UseCache && opts.CacheTTL > 0 {
				writeWebFetchCache(cacheKey, firecrawlAttempt.Payload, opts.CacheTTL)
			}
			return WebFetchExecutionResult{Payload: firecrawlAttempt.Payload}, nil
		}
		warning = joinWebFetchWarnings(warning, formatWebFetchFirecrawlWarning(firecrawlAttempt.Detail))
	}
	content := strings.TrimSpace(extract.Content)
	if content == "" && len(body) > 0 {
		content = strings.TrimSpace(string(body))
		extract.Mode = "raw_fallback"
		extract.FallbackUsed = true
	}
	truncated := rawTruncated
	truncationReason := ""
	if rawTruncated {
		truncationReason = "raw_body_limit"
	}
	if trimmed, changed := trimToMaxChars(content, opts.RequestMaxChars); changed {
		content = trimmed
		truncated = true
		truncationReason = "max_chars"
	}
	payload := WebFetchPayload{
		URL:              evidenceRequestURL,
		FinalURL:         finalURL,
		Status:           res.StatusCode,
		ContentType:      normalizedContentType,
		Redirected:       redirected,
		RedirectCount:    len(redirectChain) - 1,
		RedirectChain:    webFetchRedirectChainForPayload(redirectChain),
		Extractor:        resolveWebFetchExtractor(normalizedContentType, extract.Mode),
		SiteName:         firstNonBlank(extract.SiteName, resolveWebFetchSiteName(finalURL)),
		Title:            extract.Title,
		Byline:           extract.Byline,
		Excerpt:          extract.Excerpt,
		Content:          content,
		ContentChars:     runeLen(content),
		WordCount:        extract.WordCount,
		ExtractMode:      extract.Mode,
		FallbackUsed:     extract.FallbackUsed,
		ReadabilityUsed:  extract.ReadabilityUsed,
		ResponseBytes:    len(body),
		BodyTruncated:    rawTruncated,
		Warning:          warning,
		Truncated:        truncated,
		TruncationReason: truncationReason,
		FromCache:        false,
		FetchedAt:        time.Now().Unix(),
		TookMs:           time.Since(startedAt).Milliseconds(),
	}
	finalizeWebFetchPayload(&payload, extractMode, toolName)
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return WebFetchExecutionResult{}, formatWebFetchStatusError(toolName, payload, firecrawlStatusDetail)
	}
	if opts.UseCache && opts.CacheTTL > 0 {
		writeWebFetchCache(cacheKey, payload, opts.CacheTTL)
	}
	return WebFetchExecutionResult{Payload: payload, Links: links}, nil
}

func finalizeWebFetchPayload(payload *WebFetchPayload, requestedMode string, toolName string) {
	if payload == nil {
		return
	}
	payload.Diagnostics = &WebFetchDiagnostics{
		RequestedMode:     normalizeWebFetchExtractMode(requestedMode),
		Extractor:         strings.TrimSpace(payload.Extractor),
		Redirected:        payload.Redirected,
		RedirectCount:     payload.RedirectCount,
		CacheHit:          payload.FromCache,
		FallbackUsed:      payload.FallbackUsed,
		ReadabilityUsed:   payload.ReadabilityUsed,
		BinaryUnsupported: strings.TrimSpace(payload.ExtractMode) == "binary_unsupported",
	}
	if payload.ProviderDiagnostics == nil {
		payload.ProviderDiagnostics = BuildFetchProviderDiagnostics(FetchProviderDiagnosticsInput{
			Tool:              firstNonBlank(toolName, "web_fetch"),
			EffectiveProvider: FetchProviderForExtractor(firstNonBlank(payload.Extractor, payload.ExtractMode)),
		})
	}
}

func webFetchRedirectChainForPayload(chain []string) []string {
	if len(chain) <= 1 {
		return nil
	}
	out := make([]string, 0, len(chain))
	seen := map[string]bool{}
	for _, item := range chain {
		item = evidenceURL(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	if len(out) <= 1 {
		return nil
	}
	return out
}

func maybeWebFetchFirecrawlFallback(
	ctx context.Context,
	deadline time.Time,
	prepare Preparer,
	classify NetworkErrorClassifier,
	cfg WebFetchFirecrawlConfig,
	req WebFetchFirecrawlRequest,
) WebFetchFirecrawlAttempt {
	if !cfg.Enabled || strings.TrimSpace(cfg.APIKey) == "" {
		return WebFetchFirecrawlAttempt{}
	}
	mode := normalizeWebFetchExtractMode(req.ExtractMode)
	if mode == "raw" {
		return WebFetchFirecrawlAttempt{}
	}
	attempt := WebFetchFirecrawlAttempt{Attempted: true}
	timeoutMs := webFetchDefaultTimeout
	if cfg.TimeoutMs > 0 {
		timeoutMs = cfg.TimeoutMs
	}
	if !deadline.IsZero() {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			attempt.Detail = "skipped: timeout budget exhausted before external fallback"
			return attempt
		}
		if remainingMs := int(remaining / time.Millisecond); remainingMs > 0 && remainingMs < timeoutMs {
			timeoutMs = remainingMs
		}
	}
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	if prepare == nil {
		attempt.Detail = "unavailable: request preparer is unavailable"
		return attempt
	}
	prepared, err := prepare(runCtx, httprequest.PrepareInput{RawURL: cfg.Endpoint, TimeoutMs: timeoutMs, FollowRedirects: true, MaxRedirects: 5, CredentialSensitive: true})
	if err != nil {
		attempt.Detail = "unavailable: " + truncateToolText(compactWhitespace(err.Error()), 180)
		return attempt
	}
	if prepared.Close != nil {
		defer prepared.Close()
	}
	if prepared.URL == nil || prepared.Doer == nil {
		attempt.Detail = "unavailable: request preparer returned an incomplete client"
		return attempt
	}
	endpoint := prepared.URL
	requestBody, err := json.Marshal(map[string]any{
		"url":             strings.TrimSpace(req.URL),
		"formats":         []string{"markdown"},
		"onlyMainContent": cfg.OnlyMainContent,
		"maxAge":          normalizeWebFetchFirecrawlMaxAgeMs(cfg.MaxAgeMs),
		"timeout":         timeoutMs,
	})
	if err != nil {
		attempt.Detail = "invalid_request"
		return attempt
	}
	httpReq, err := http.NewRequestWithContext(runCtx, http.MethodPost, endpoint.String(), bytes.NewReader(requestBody))
	if err != nil {
		attempt.Detail = "request_build_failed"
		return attempt
	}
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	res, err := prepared.Doer.Do(httpReq)
	if err != nil {
		attempt.Detail = "request_failed"
		if classify != nil && classify(err) == NetworkErrorProxyConfig {
			attempt.Detail = "proxy_config_invalid"
		}
		return attempt
	}
	defer res.Body.Close()

	body, _, err := readBodyLimited(res.Body, 1_000_000)
	if err != nil {
		attempt.Detail = "response_read_failed"
		return attempt
	}
	var decoded WebFetchFirecrawlResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		attempt.Detail = "invalid_response"
		return attempt
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 || !decoded.Success {
		reason := firstNonBlank(decoded.Error, decoded.Warning, http.StatusText(res.StatusCode))
		detail := fmt.Sprintf("status_error: status=%d", res.StatusCode)
		if strings.TrimSpace(reason) != "" {
			detail += " reason=" + truncateToolText(compactWhitespace(sanitizeWebFetchRequestURLText(reason, req.URL, req.FinalURLFallback)), 160)
		}
		attempt.Detail = detail
		return attempt
	}
	rawContent := firstNonBlank(decoded.Data.Markdown, decoded.Data.Content)
	if strings.TrimSpace(rawContent) == "" {
		attempt.Detail = "empty_content"
		return attempt
	}
	content := normalizeMarkdownWhitespace(rawContent)
	extractMode := "firecrawl_markdown"
	if mode != "markdown" {
		content = markdownToPlainText(content)
		extractMode = "firecrawl_text"
	}
	truncated := false
	truncationReason := ""
	if trimmed, changed := trimToMaxChars(content, req.MaxChars); changed {
		content = trimmed
		truncated = true
		truncationReason = "max_chars"
	}
	finalURL := evidenceURL(firstNonBlank(decoded.Data.Metadata.SourceURL, req.FinalURLFallback, req.URL))
	status := decoded.Data.Metadata.StatusCode
	if status <= 0 {
		status = req.StatusFallback
	}
	plainContent := content
	if extractMode == "firecrawl_markdown" {
		plainContent = markdownToPlainText(content)
	}
	payload := WebFetchPayload{
		URL:              evidenceURL(req.URL),
		FinalURL:         finalURL,
		Status:           status,
		ContentType:      "text/markdown",
		Extractor:        "firecrawl",
		SiteName:         resolveWebFetchSiteName(finalURL),
		Title:            strings.TrimSpace(decoded.Data.Metadata.Title),
		Excerpt:          truncateToolText(plainContent, 220),
		Content:          content,
		ContentChars:     runeLen(content),
		WordCount:        len(strings.Fields(plainContent)),
		ExtractMode:      extractMode,
		FallbackUsed:     true,
		ReadabilityUsed:  false,
		ResponseBytes:    len([]byte(rawContent)),
		BodyTruncated:    false,
		Warning:          sanitizeWebFetchRequestURLText(decoded.Warning, req.URL),
		Truncated:        truncated,
		TruncationReason: truncationReason,
		FromCache:        false,
		FetchedAt:        time.Now().Unix(),
	}
	finalizeWebFetchPayload(&payload, req.ExtractMode, req.ToolName)
	if payload.ProviderDiagnostics != nil {
		payload.ProviderDiagnostics.Fallback = &ProviderFallback{
			From:   ProviderDirectHTTP,
			To:     ProviderFirecrawl,
			Reason: NormalizeProviderHealthStatus(req.FallbackReason),
			Hint:   strings.TrimSpace(req.FallbackHint),
		}
	}
	attempt.Payload = payload
	attempt.Used = true
	return attempt
}

func formatWebFetchRequestError(toolName string, urlValue string, timeoutMs int, err error, firecrawlDetail string, classify NetworkErrorClassifier) error {
	if err == nil {
		return nil
	}
	if strings.TrimSpace(toolName) == "" {
		toolName = "web_fetch"
	}
	fallback := renderWebFetchFirecrawlDetail(firecrawlDetail)
	kind := WebFetchErrorRequest
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		kind = WebFetchErrorTimeout
	case webFetchTLSError(err):
		kind = WebFetchErrorTLS
	}
	if classify != nil {
		switch classify(err) {
		case NetworkErrorRedirectLimit:
			kind = WebFetchErrorRedirectLimit
		case NetworkErrorProxyConfig:
			kind = WebFetchErrorProxyConfig
		case NetworkErrorPolicyBlocked:
			kind = WebFetchErrorPolicyBlocked
		}
	}
	return &WebFetchError{
		Kind:      kind,
		ToolName:  toolName,
		URL:       urlValue,
		TimeoutMs: timeoutMs,
		fallback:  fallback,
		cause:     err,
	}
}

func webFetchTLSError(err error) bool {
	var verificationErr *tls.CertificateVerificationError
	var unknownAuthorityErr x509.UnknownAuthorityError
	var certificateInvalidErr x509.CertificateInvalidError
	return errors.As(err, &verificationErr) ||
		errors.As(err, &unknownAuthorityErr) ||
		errors.As(err, &certificateInvalidErr)
}

func sanitizeWebFetchRequestURLText(value string, requestURLs ...string) string {
	out := strings.TrimSpace(value)
	for _, raw := range requestURLs {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		out = strings.ReplaceAll(out, raw, evidenceURL(raw))
	}
	return out
}

func formatWebFetchStatusError(toolName string, payload WebFetchPayload, firecrawlDetail string) error {
	statusClass := "client_error"
	if payload.Status >= 500 {
		statusClass = "server_error"
	}
	if strings.TrimSpace(toolName) == "" {
		toolName = "web_fetch"
	}
	fallback := renderWebFetchFirecrawlDetail(firecrawlDetail)
	content := strings.ToLower(strings.TrimSpace(payload.Content))
	return &WebFetchError{
		Kind:        WebFetchErrorStatus,
		ToolName:    toolName,
		URL:         payload.FinalURL,
		Status:      payload.Status,
		StatusClass: statusClass,
		ContentType: payload.ContentType,
		Challenge: payload.Status == http.StatusUnauthorized ||
			payload.Status == http.StatusForbidden ||
			strings.Contains(content, "challenge") ||
			strings.Contains(content, "cf-mitigated"),
		fallback: fallback,
	}
}

func renderWebFetchFirecrawlDetail(detail string) string {
	trimmed := strings.TrimSpace(detail)
	if trimmed == "" {
		return ""
	}
	return fmt.Sprintf(` firecrawl_detail=%q`, truncateToolText(trimmed, 220))
}

func formatWebFetchFirecrawlWarning(detail string) string {
	trimmed := strings.TrimSpace(detail)
	if trimmed == "" {
		return ""
	}
	return "External extractor fallback failed; returning local HTML fallback content. Detail: " + truncateToolText(trimmed, 220)
}

func readBodyLimited(body io.Reader, maxBytes int) (content []byte, truncated bool, err error) {
	if maxBytes <= 0 {
		maxBytes = webFetchDefaultMaxChars
	}
	limited := io.LimitReader(body, int64(maxBytes+1))
	blob, err := io.ReadAll(limited)
	if err != nil {
		return nil, false, err
	}
	if len(blob) > maxBytes {
		return blob[:maxBytes], true, nil
	}
	return blob, false, nil
}

func writeWebFetchCache(key string, payload WebFetchPayload, ttl time.Duration) {
	blob, err := json.Marshal(payload)
	if err != nil {
		return
	}
	WriteWebFetchCache(key, blob, ttl)
}

func normalizeWebFetchFirecrawlMaxAgeMs(value int) int {
	if value <= 0 {
		return webFetchFirecrawlMaxAge
	}
	return value
}

func webFetchBinaryWarning(toolName string, contentType string, body []byte) string {
	if len(body) == 0 {
		return ""
	}
	if strings.TrimSpace(toolName) == "" {
		toolName = "web_fetch"
	}
	lowerType := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.Contains(lowerType, "application/pdf"):
		return "Binary PDF responses are not rendered by " + toolName + ". Use pdf for default PDF analysis, or pdf_read_pages/pdf_extract_structured for specialist follow-up."
	case strings.HasPrefix(lowerType, "image/"):
		return "Image responses are not rendered by " + toolName + ". Use a browser screenshot action or a media-specific tool instead."
	case strings.HasPrefix(lowerType, "audio/"), strings.HasPrefix(lowerType, "video/"):
		return "Audio/video responses are not rendered by " + toolName + ". Use a media-specific tool instead."
	case strings.Contains(lowerType, "application/zip"),
		strings.Contains(lowerType, "application/gzip"),
		strings.Contains(lowerType, "application/x-gzip"),
		strings.Contains(lowerType, "application/vnd."),
		strings.Contains(lowerType, "application/octet-stream"):
		return "Binary attachment responses are not rendered by " + toolName + "."
	}
	if !isLikelyTextBody(body) {
		return "Binary response bodies are not rendered by " + toolName + "."
	}
	return ""
}

func webFetchBinaryHandoff(contentType string, body []byte) *WebFetchHandoff {
	if len(body) == 0 {
		return nil
	}
	lowerType := strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case strings.Contains(lowerType, "application/pdf"):
		return &WebFetchHandoff{
			Kind:          "pdf",
			PreferredTool: "pdf",
			Summary:       "Binary PDF responses should be handed to the unified pdf tool.",
		}
	case strings.HasPrefix(lowerType, "image/"):
		return &WebFetchHandoff{
			Kind:          "image",
			PreferredTool: "image_analyze",
			Summary:       "Binary image responses should be handed to image_analyze for remote image inspection.",
		}
	case strings.HasPrefix(lowerType, "audio/"), strings.HasPrefix(lowerType, "video/"):
		return &WebFetchHandoff{
			Kind:          "media",
			PreferredTool: "browser",
			Summary:       "Binary audio/video responses need a media-capable or browser follow-up instead of text extraction.",
		}
	case strings.Contains(lowerType, "application/zip"),
		strings.Contains(lowerType, "application/gzip"),
		strings.Contains(lowerType, "application/x-gzip"),
		strings.Contains(lowerType, "application/vnd."),
		strings.Contains(lowerType, "application/octet-stream"):
		return &WebFetchHandoff{
			Kind:          "binary_attachment",
			PreferredTool: "browser",
			Summary:       "Binary attachment responses should be handed to browser or a file-specific tool for download/follow-up.",
		}
	}
	if !isLikelyTextBody(body) {
		return &WebFetchHandoff{
			Kind:          "binary_attachment",
			PreferredTool: "browser",
			Summary:       "Binary response bodies should be handed to browser or a file-specific tool instead of text extraction.",
		}
	}
	return nil
}

func isLikelyTextBody(body []byte) bool {
	if len(body) == 0 {
		return true
	}
	if bytes.IndexByte(body, 0) >= 0 {
		return false
	}
	sample := body
	if len(sample) > 4096 {
		sample = sample[:4096]
	}
	if utf8.Valid(sample) {
		return true
	}
	printable := 0
	for _, b := range sample {
		if b == '\n' || b == '\r' || b == '\t' || (b >= 32 && b < 127) {
			printable++
		}
	}
	return printable*100/len(sample) >= 85
}
