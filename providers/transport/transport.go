// Package transport provides provider-neutral request settings and hooks.
package transport

import (
	"context"
	"net/http"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

// Config contains host-supplied transport defaults. It contains no credentials.
type Config struct {
	Mode    string
	Headers map[string]string
}

// Settings is the normalized transport/session/cache contract consumed by providers.
type Settings struct {
	Mode            string
	Headers         map[string]string
	SessionID       string
	CacheControl    any
	MaxRetryDelayMs *int
	OnPayload       llm.PayloadHook
	OnResponse      llm.ResponseHook
}

// Resolve combines transport defaults with typed request overrides.
func Resolve(cfg Config, request llm.RequestOptions) Settings {
	headers := cloneStringMap(cfg.Headers)
	if len(request.Headers) > 0 {
		if headers == nil {
			headers = map[string]string{}
		}
		for key, value := range request.Headers {
			headers[key] = value
		}
	}
	mode := cfg.Mode
	if request.Transport != "" {
		mode = request.Transport
	}
	return Settings{
		Mode: mode, Headers: headers, SessionID: request.SessionID,
		CacheControl:    cloneValue(request.CacheControl),
		MaxRetryDelayMs: cloneIntPtr(request.MaxRetryDelayMs),
		OnPayload:       request.OnPayload, OnResponse: request.OnResponse,
	}
}

// ApplyPayload injects normalized session/cache fields without overriding explicit values.
func ApplyPayload(payload map[string]any, settings Settings) {
	if payload == nil {
		return
	}
	if settings.SessionID != "" {
		if _, exists := payload["session_id"]; !exists {
			payload["session_id"] = settings.SessionID
		}
	}
	if settings.CacheControl == nil {
		return
	}
	extraBody, _ := payload["extra_body"].(map[string]any)
	if extraBody == nil {
		extraBody = map[string]any{}
	}
	if _, exists := extraBody["cache_control"]; !exists {
		extraBody["cache_control"] = cloneValue(settings.CacheControl)
	}
	payload["extra_body"] = extraBody
}

// ApplyHeaders adds normalized headers without overriding provider-owned values.
func ApplyHeaders(headers http.Header, settings Settings) {
	if headers == nil {
		return
	}
	for key, value := range settings.Headers {
		if headers.Get(key) == "" {
			headers.Set(key, value)
		}
	}
}

// ApplyPayloadHook invokes the optional payload hook.
func ApplyPayloadHook(ctx context.Context, settings Settings, payload any) (any, error) {
	if settings.OnPayload == nil {
		return payload, nil
	}
	return settings.OnPayload(ctx, payload)
}

// ApplyResponseHook invokes the optional response hook.
func ApplyResponseHook(ctx context.Context, settings Settings, meta llm.ResponseMetadata) error {
	if settings.OnResponse == nil {
		return nil
	}
	return settings.OnResponse(ctx, meta)
}

// ResponseMetadataFromHTTP creates an immutable response metadata snapshot.
func ResponseMetadataFromHTTP(method, rawURL string, resp *http.Response) llm.ResponseMetadata {
	meta := llm.ResponseMetadata{Method: method, URL: rawURL}
	if resp != nil {
		meta.StatusCode = resp.StatusCode
		meta.Headers = cloneHeader(resp.Header)
	}
	return meta
}

type requestOptionsContextKey struct{}

// WithRequestOptions attaches a defensive copy of typed options to ctx.
func WithRequestOptions(ctx context.Context, options llm.RequestOptions) context.Context {
	return context.WithValue(ctx, requestOptionsContextKey{}, options.Clone())
}

// RequestOptionsFromContext returns a defensive copy of attached options.
func RequestOptionsFromContext(ctx context.Context) (llm.RequestOptions, bool) {
	if ctx == nil {
		return llm.RequestOptions{}, false
	}
	value, ok := ctx.Value(requestOptionsContextKey{}).(llm.RequestOptions)
	if !ok {
		return llm.RequestOptions{}, false
	}
	return value.Clone(), true
}

// ResolveFromContext merges context-bound request options with defaults.
func ResolveFromContext(ctx context.Context, cfg Config) Settings {
	options, _ := RequestOptionsFromContext(ctx)
	return Resolve(cfg, options)
}

func cloneHeader(headers http.Header) map[string][]string {
	if len(headers) == 0 {
		return nil
	}
	out := make(map[string][]string, len(headers))
	for key, values := range headers {
		out[key] = append([]string(nil), values...)
	}
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(value))
		for key, item := range value {
			out[key] = cloneValue(item)
		}
		return out
	case []any:
		out := make([]any, len(value))
		for i, item := range value {
			out[i] = cloneValue(item)
		}
		return out
	case []string:
		return append([]string(nil), value...)
	default:
		return value
	}
}

func cloneIntPtr(value *int) *int {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}
