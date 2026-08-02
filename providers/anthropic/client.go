// Package anthropic implements the Anthropic Messages HTTP protocol.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/providers"
	"github.com/wsnacj/agentx-go/providers/transport"
)

const (
	defaultBaseURL          = "https://api.anthropic.com"
	defaultMessagesPath     = "/v1/messages"
	defaultAnthropicVersion = "2023-06-01"
	defaultChatTimeout      = 2 * time.Minute
)

// Client is an immutable, concurrency-safe Anthropic Messages client.
type Client struct {
	name       string
	baseURL    string
	version    string
	transport  transport.Config
	authorize  Authorizer
	httpClient HTTPDoer
}

// New constructs a client without reading credentials or making network calls.
func New(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(firstNonEmpty(cfg.BaseURL, defaultBaseURL), "/")
	version := firstNonEmpty(cfg.Version, defaultAnthropicVersion)
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}
	return &Client{
		name:       cfg.Name,
		baseURL:    baseURL,
		version:    version,
		transport:  transport.Config{Mode: cfg.Transport.Mode, Headers: cloneHeaders(cfg.Transport.Headers)},
		authorize:  cfg.Authorize,
		httpClient: client,
	}, nil
}

// Chat executes one Anthropic Messages request.
func (c *Client) Chat(ctx context.Context, cfg ModelConfig, req llm.ChatRequest) (*llm.ChatResponse, *llm.Usage, error) {
	req = resolveChatRequest(cfg, req)
	payload, err := buildMessagesPayload(req)
	if err != nil {
		return nil, nil, err
	}
	settings := transport.Resolve(c.transport, llm.RequestOptionsFromMap(req.Options))
	raw, err := c.post(ctx, defaultMessagesPath, payload, defaultChatTimeout, settings)
	if err != nil {
		return nil, nil, err
	}
	return parseMessagesResponse(raw)
}

func (c *Client) post(ctx context.Context, requestPath string, payload any, timeout time.Duration, settings transport.Settings) ([]byte, error) {
	payloadAny, err := transport.ApplyPayloadHook(ctx, settings, payload)
	if err != nil {
		return nil, fmt.Errorf("apply payload hook: %w", err)
	}
	data, err := json.Marshal(payloadAny)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	endpoint := c.baseURL + requestPath
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("anthropic-version", c.version)
	transport.ApplyHeaders(request.Header, settings)
	if c.authorize != nil {
		if err := c.authorize(ctx, request.Header); err != nil {
			return nil, err
		}
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer response.Body.Close()
	if err := transport.ApplyResponseHook(ctx, settings, transport.ResponseMetadataFromHTTP(http.MethodPost, endpoint, response)); err != nil {
		return nil, fmt.Errorf("apply response hook: %w", err)
	}
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, &providers.APIError{StatusCode: response.StatusCode, Body: string(raw)}
	}
	return raw, nil
}

func resolveChatRequest(cfg ModelConfig, req llm.ChatRequest) llm.ChatRequest {
	out := req
	if out.Model == "" {
		out.Model = cfg.Model
	}
	if out.MaxTokens <= 0 {
		out.MaxTokens = cfg.MaxCompletion
	}
	if out.Temperature == 0 {
		if typed := llm.RequestOptionsFromMap(out.Options); typed.Temperature != nil {
			out.Temperature = float32(*typed.Temperature)
		} else {
			out.Temperature = cfg.Temperature
		}
	}
	return out
}

func cloneHeaders(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
