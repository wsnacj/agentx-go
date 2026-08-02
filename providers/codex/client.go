// Package codex implements the OpenAI Codex Responses HTTP/SSE protocol.
package codex

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
	defaultTimeout    = 2 * time.Minute
	responsesPath     = "/responses"
	defaultUserAgent  = "codex_cli_rs/0.0.0 (AgentX llmx)"
	defaultOriginator = "codex_cli_rs"
)

// Client is an immutable, concurrency-safe Codex Responses client.
type Client struct {
	name       string
	baseURL    string
	userAgent  string
	originator string
	transport  transport.Config
	authorize  Authorizer
	httpClient HTTPDoer
}

// New constructs a client without reading credentials or making network calls.
func New(cfg Config) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("codex: base URL is required")
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.Timeout}
	}
	return &Client{
		name: cfg.Name, baseURL: baseURL,
		userAgent:  firstNonEmpty(cfg.UserAgent, defaultUserAgent),
		originator: firstNonEmpty(cfg.Originator, defaultOriginator),
		transport:  transport.Config{Mode: cfg.Transport.Mode, Headers: cloneHeaders(cfg.Transport.Headers)},
		authorize:  cfg.Authorize, httpClient: client,
	}, nil
}

// Chat executes one streamed Responses request and returns the collected response.
func (c *Client) Chat(ctx context.Context, cfg ModelConfig, req llm.ChatRequest) (*llm.ChatResponse, *llm.Usage, error) {
	req = resolveChatRequest(cfg, req)
	payload, settings := c.buildChatPayload(cfg, req)
	payload["stream"] = true
	return c.postStream(ctx, responsesPath, payload, settings)
}

func (c *Client) postStream(ctx context.Context, path string, payload map[string]any, settings transport.Settings) (*llm.ChatResponse, *llm.Usage, error) {
	payloadAny, err := transport.ApplyPayloadHook(ctx, settings, payload)
	if err != nil {
		return nil, nil, fmt.Errorf("apply payload hook: %w", err)
	}
	data, err := json.Marshal(payloadAny)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal payload: %w", err)
	}
	target := c.baseURL + path
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "text/event-stream")
	transport.ApplyHeaders(request.Header, settings)
	setIfMissing(request.Header, "User-Agent", c.userAgent)
	setIfMissing(request.Header, "originator", c.originator)
	if c.authorize != nil {
		if err := c.authorize(ctx, request.Header); err != nil {
			return nil, nil, err
		}
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("send request: %w", err)
	}
	defer response.Body.Close()
	if err := transport.ApplyResponseHook(ctx, settings, transport.ResponseMetadataFromHTTP(http.MethodPost, target, response)); err != nil {
		return nil, nil, fmt.Errorf("apply response hook: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		raw, readErr := io.ReadAll(response.Body)
		if readErr != nil {
			return nil, nil, fmt.Errorf("read error response: %w", readErr)
		}
		return nil, nil, &providers.APIError{StatusCode: response.StatusCode, Body: string(raw)}
	}
	collector := newStreamCollector()
	if err := readServerSentEvents(response.Body, collector.handle); err != nil {
		return nil, nil, err
	}
	return collector.finish()
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

func setIfMissing(headers http.Header, key, value string) {
	if key == "" || value == "" || headers.Get(key) != "" {
		return
	}
	headers.Set(key, value)
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
