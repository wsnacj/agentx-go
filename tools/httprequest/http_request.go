// Package httprequest provides portable HTTP tool request/response coordination.
package httprequest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
)

// Name is the catalog name of the HTTP request tool.
const Name = "http_request"

const (
	defaultTimeoutMs    = 20_000
	defaultMaxChars     = 60_000
	defaultMaxRedirects = 4
	maxTimeoutMs        = 120_000
	maxRedirects        = 16
)

// HTTPDoer is the narrow Host-owned HTTP execution port.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// PrepareInput describes the policy-sensitive part of an HTTP request.
type PrepareInput struct {
	RawURL          string
	TimeoutMs       int
	FollowRedirects bool
	MaxRedirects    int
	// CredentialSensitive asks the Host to apply its credential-safe redirect
	// policy. It does not prescribe same-origin rules or header handling.
	CredentialSensitive bool
	// OnRedirect receives display-safe redirect observations from the Host.
	// The Host remains responsible for validating every redirect before it is
	// followed. Portable retrieval uses this hook only for diagnostics.
	OnRedirect func(context.Context, string, int)
}

// PreparedRequest contains a policy-validated URL and a Host-owned client.
type PreparedRequest struct {
	URL   *url.URL
	Doer  HTTPDoer
	Close func()
}

// Preparer validates the initial URL and creates a request-scoped HTTP client.
// Redirect validation, proxy policy and transport ownership remain in the Host.
type Preparer func(context.Context, PrepareInput) (PreparedRequest, error)

// Options configures portable request coordination.
type Options struct {
	Prepare      Preparer
	TimeoutMs    int
	MaxChars     int
	MaxRedirects int
	UserAgent    string
	Now          func() time.Time
}

// Request is the normalized request accepted by Run.
type Request struct {
	Method          string            `json:"method,omitempty"`
	URL             string            `json:"url"`
	Headers         map[string]string `json:"headers,omitempty"`
	Query           map[string]string `json:"query,omitempty"`
	Body            string            `json:"body,omitempty"`
	TimeoutMs       int               `json:"timeout_ms,omitempty"`
	MaxChars        int               `json:"max_chars,omitempty"`
	FollowRedirects *bool             `json:"follow_redirects,omitempty"`
	MaxRedirects    int               `json:"max_redirects,omitempty"`
}

type externalContentMeta struct {
	Untrusted bool     `json:"untrusted"`
	Source    string   `json:"source,omitempty"`
	Wrapped   bool     `json:"wrapped,omitempty"`
	Fields    []string `json:"fields,omitempty"`
}

type providerDiagnostics struct {
	Tool              string `json:"tool,omitempty"`
	EffectiveProvider string `json:"effective_provider,omitempty"`
}

type responsePayload struct {
	Method              string               `json:"method"`
	URL                 string               `json:"url"`
	FinalURL            string               `json:"final_url"`
	Status              int                  `json:"status"`
	Headers             map[string]string    `json:"headers,omitempty"`
	Body                string               `json:"body"`
	ExternalContent     *externalContentMeta `json:"external_content,omitempty"`
	ProviderDiagnostics *providerDiagnostics `json:"provider_diagnostics,omitempty"`
	Truncated           bool                 `json:"truncated"`
	DurationMs          int64                `json:"duration_ms"`
	ContentType         string               `json:"content_type,omitempty"`
}

// Register adds the HTTP request tool when a Host preparer is available.
func Register(reg toolcontract.Registrar, opts Options) {
	if reg == nil || opts.Prepare == nil {
		return
	}
	reg.Register(Definition(), NewHandler(opts))
}

// NewHandler returns the model-facing JSON handler.
func NewHandler(opts Options) toolcontract.Handler {
	return func(ctx context.Context, call toolcontract.Call) (toolcontract.Result, error) {
		var request Request
		if strings.TrimSpace(call.Arguments) != "" {
			if err := json.Unmarshal([]byte(call.Arguments), &request); err != nil {
				return "", fmt.Errorf("decode tool args: %w", err)
			}
		}
		return Run(ctx, request, opts)
	}
}

// Run executes the portable request lifecycle through the Host-owned Preparer.
func Run(ctx context.Context, request Request, opts Options) (toolcontract.Result, error) {
	rawURL := strings.TrimSpace(request.URL)
	if rawURL == "" {
		return "", agentxtoolerrors.NewMissingRequiredToolArgumentError(Name, []string{"url"}, Name+": url is required")
	}
	method := NormalizeMethod(request.Method)
	if method == "" {
		return "", fmt.Errorf("%s: unsupported method", Name)
	}
	if opts.Prepare == nil {
		return "", fmt.Errorf("%s: request preparer is unavailable", Name)
	}

	timeoutMs := clampPositive(request.TimeoutMs, opts.TimeoutMs, defaultTimeoutMs, maxTimeoutMs)
	maxChars := clampPositive(request.MaxChars, opts.MaxChars, defaultMaxChars, positiveOrMax(opts.MaxChars, defaultMaxChars))
	redirectLimit := clampPositive(request.MaxRedirects, opts.MaxRedirects, defaultMaxRedirects, maxRedirects)
	followRedirects := true
	if request.FollowRedirects != nil {
		followRedirects = *request.FollowRedirects
	}
	prepared, err := opts.Prepare(ctx, PrepareInput{
		RawURL: rawURL, TimeoutMs: timeoutMs, FollowRedirects: followRedirects, MaxRedirects: redirectLimit,
		CredentialSensitive: len(request.Headers) > 0 || strings.TrimSpace(request.Body) != "",
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", Name, err)
	}
	if prepared.Close != nil {
		defer prepared.Close()
	}
	if prepared.URL == nil || prepared.Doer == nil {
		return "", fmt.Errorf("%s: request preparer returned an incomplete client", Name)
	}
	requestURL := mergeQuery(prepared.URL, request.Query)
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	httpRequest, err := http.NewRequestWithContext(runCtx, method, requestURL.String(), strings.NewReader(request.Body))
	if err != nil {
		return "", err
	}
	userAgent := strings.TrimSpace(opts.UserAgent)
	if userAgent == "" {
		userAgent = "agentx-http-request/1.0"
	}
	httpRequest.Header.Set("User-Agent", userAgent)
	for key, value := range request.Headers {
		if name := strings.TrimSpace(key); name != "" {
			httpRequest.Header.Set(name, value)
		}
	}

	now := opts.Now
	if now == nil {
		now = time.Now
	}
	started := now()
	response, err := prepared.Doer.Do(httpRequest)
	if err != nil {
		return "", formatRequestError(requestURL.String(), timeoutMs, err)
	}
	if response == nil {
		return "", fmt.Errorf("%s: request_failed url=%s: empty response", Name, requestURL.String())
	}
	defer response.Body.Close()
	durationMs := now().Sub(started).Milliseconds()
	body, truncated, err := readBodyLimited(response.Body, maxChars)
	if err != nil {
		return "", err
	}
	if err := runCtx.Err(); err != nil {
		return "", err
	}
	finalURL := requestURL.String()
	if response.Request != nil && response.Request.URL != nil {
		finalURL = response.Request.URL.String()
	}
	payload := responsePayload{
		Method: method, URL: requestURL.String(), FinalURL: finalURL, Status: response.StatusCode,
		Headers: flattenHeaders(response.Header), Body: body, Truncated: truncated, DurationMs: durationMs,
		ContentType:         strings.TrimSpace(response.Header.Get("Content-Type")),
		ProviderDiagnostics: &providerDiagnostics{Tool: Name, EffectiveProvider: "direct_http"},
	}
	if strings.TrimSpace(body) != "" {
		payload.ExternalContent = &externalContentMeta{Untrusted: true, Source: Name, Fields: []string{"body"}}
	}
	blob, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	if response.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("%s: unexpected status %d: %s", Name, response.StatusCode, truncateText(body, 320))
	}
	return string(blob), nil
}

// Definition returns the stable HTTP tool schema.
func Definition() toolcontract.Definition {
	return toolcontract.Definition{Type: "function", Function: toolcontract.Function{
		Name:        Name,
		Description: "Protocol-level HTTP utility for explicit method/url/headers/body requests under outbound network policy. Prefer search/open_page/web_fetch for normal retrieval.",
		Parameters: map[string]any{
			"type": "object", "additionalProperties": false,
			"properties": map[string]any{
				"method":           map[string]any{"type": "string", "description": "HTTP method. Defaults to GET when omitted.", "enum": []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}},
				"url":              stringSchema("Absolute http(s) URL to request."),
				"headers":          map[string]any{"type": "object", "description": "Request headers to send. Values must be strings.", "additionalProperties": map[string]any{"type": "string"}},
				"query":            map[string]any{"type": "object", "description": "Query parameters to merge into the URL. Values must be strings.", "additionalProperties": map[string]any{"type": "string"}},
				"body":             stringSchema("Raw request body for methods that send a payload."),
				"timeout_ms":       intSchema("Maximum request runtime in milliseconds. Omit to use the configured default.", 1),
				"max_chars":        intSchema("Maximum response body characters to return. The runtime clamps this to its configured limit.", 256),
				"follow_redirects": boolSchema("Whether to follow redirects. Defaults to true."),
				"max_redirects":    intSchema("Maximum redirects to follow when redirects are enabled. The runtime clamps this to its configured limit.", 1),
			},
			"required": []string{"url"},
		},
		OutputSchema: outputSchema(),
	}}
}

// NormalizeMethod returns an allowed uppercase HTTP method, defaults an empty
// value to GET, and returns an empty string for unsupported methods.
func NormalizeMethod(raw string) string {
	method := strings.ToUpper(strings.TrimSpace(raw))
	if method == "" {
		return http.MethodGet
	}
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions:
		return method
	default:
		return ""
	}
}

func mergeQuery(base *url.URL, query map[string]string) *url.URL {
	if len(query) == 0 {
		return base
	}
	out := *base
	values := out.Query()
	for key, value := range query {
		if name := strings.TrimSpace(key); name != "" {
			values.Set(name, value)
		}
	}
	out.RawQuery = values.Encode()
	return &out
}

func readBodyLimited(body io.Reader, maxChars int) (string, bool, error) {
	if maxChars <= 0 {
		maxChars = defaultMaxChars
	}
	blob, err := io.ReadAll(io.LimitReader(body, int64(maxChars+1)))
	if err != nil {
		return "", false, err
	}
	text := string(blob)
	if !utf8.ValidString(text) {
		if len(text) <= maxChars {
			return text, false, nil
		}
		return text[:maxChars], true, nil
	}
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text, false, nil
	}
	return string(runes[:maxChars]), true, nil
}

func flattenHeaders(headers http.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	out := make(map[string]string, len(headers))
	for key, values := range headers {
		name := strings.ToLower(strings.TrimSpace(key))
		if name != "" && len(values) > 0 {
			out[name] = strings.TrimSpace(values[0])
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func formatRequestError(urlValue string, timeoutMs int, err error) error {
	lower := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case strings.Contains(lower, "context deadline exceeded"), strings.Contains(lower, "timeout awaiting response headers"):
		return fmt.Errorf("%s: timeout url=%s timeout_ms=%d hint=try reducing response size or increasing timeout_ms: %w", Name, urlValue, timeoutMs, err)
	case strings.Contains(lower, "too many redirects"):
		return fmt.Errorf("%s: redirect_limit url=%s hint=check redirect chain or max_redirects setting: %w", Name, urlValue, err)
	case strings.Contains(lower, "trusted env proxy invalid"), strings.Contains(lower, "trusted env proxy blocked"):
		return fmt.Errorf("%s: proxy_config_invalid url=%s hint=check tools.httpRequestTrustedEnvProxy and HTTP_PROXY/HTTPS_PROXY/NO_PROXY: %w", Name, urlValue, err)
	default:
		return fmt.Errorf("%s: request_failed url=%s: %w", Name, urlValue, err)
	}
}

func clampPositive(requested, configured, fallback, maximum int) int {
	value := requested
	if value <= 0 {
		value = configured
	}
	if value <= 0 {
		value = fallback
	}
	if maximum > 0 && value > maximum {
		value = maximum
	}
	return value
}

func positiveOrMax(configured, fallback int) int {
	if configured > 0 {
		return configured
	}
	return fallback
}

func truncateText(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if max <= 0 {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) <= max {
		return trimmed
	}
	return string(runes[:max])
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}
func boolSchema(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}
func intSchema(description string, minimum int) map[string]any {
	return map[string]any{"type": "integer", "description": description, "minimum": minimum}
}
func looseObjectSchema(description string) map[string]any {
	return map[string]any{"type": "object", "description": description, "additionalProperties": true}
}
func outputSchema() map[string]any {
	return map[string]any{"type": "object", "additionalProperties": false,
		"properties": map[string]any{
			"method": stringSchema("HTTP method used."), "url": stringSchema("Requested URL after query merge."),
			"final_url": stringSchema("Final URL after redirects."), "status": intSchema("HTTP status code.", 0),
			"headers": map[string]any{"type": "object", "description": "Response headers.", "additionalProperties": map[string]any{"type": "string"}},
			"body":    stringSchema("Response body after truncation."), "external_content": looseObjectSchema("Metadata marking externally fetched content as untrusted."),
			"provider_diagnostics": looseObjectSchema("Provider selection, capability, health, and fallback diagnostics."),
			"truncated":            boolSchema("True when the response body was truncated."), "duration_ms": intSchema("Request duration in milliseconds.", 0),
			"content_type": stringSchema("Response content type."),
		},
		"required": []string{"method", "url", "final_url", "status", "body", "truncated", "duration_ms"},
	}
}
